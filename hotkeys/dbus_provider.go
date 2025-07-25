package hotkeys

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

// DbusKeyboardProvider implements KeyboardEventProvider using D-Bus portal
type DbusKeyboardProvider struct {
	config        HotkeyConfig
	environment   EnvironmentType
	callbacks     map[string]func() error
	conn          *dbus.Conn
	sessionHandle string
	isListening   bool
	mutex         sync.Mutex
}

// NewDbusKeyboardProvider creates a new D-Bus keyboard provider
func NewDbusKeyboardProvider(config HotkeyConfig, environment EnvironmentType) *DbusKeyboardProvider {
	return &DbusKeyboardProvider{
		config:      config,
		environment: environment,
		callbacks:   make(map[string]func() error),
		isListening: false,
	}
}

// IsSupported checks if D-Bus portal GlobalShortcuts is available
func (p *DbusKeyboardProvider) IsSupported() bool {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Printf("D-Bus session bus not available: %v", err)
		return false
	}
	defer conn.Close()

	// Check if GlobalShortcuts portal is available
	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	call := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0)
	if call.Err != nil {
		log.Printf("D-Bus portal not available: %v", call.Err)
		return false
	}

	// Check if the introspection contains GlobalShortcuts interface
	var introspectData string
	if err := call.Store(&introspectData); err != nil {
		log.Printf("Failed to get introspection data: %v", err)
		return false
	}

	// Check for GlobalShortcuts interface in introspection data
	if len(introspectData) > 0 && containsGlobalShortcuts(introspectData) {
		log.Println("D-Bus portal GlobalShortcuts detected")
		return true
	}

	log.Println("D-Bus portal GlobalShortcuts not available")
	return false
}

// containsGlobalShortcuts checks if the introspection data contains GlobalShortcuts interface
func containsGlobalShortcuts(data string) bool {
	return strings.Contains(data, "GlobalShortcuts")
}

// Start begins listening for D-Bus hotkey events
func (p *DbusKeyboardProvider) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isListening {
		return fmt.Errorf("D-Bus keyboard provider already started")
	}

	var err error
	p.conn, err = dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus (D-Bus unavailable): %w", err)
	}

	// Register hotkeys using GlobalShortcuts portal
	if err := p.registerHotkeys(); err != nil {
		p.conn.Close()
		return fmt.Errorf("failed to register hotkeys (GlobalShortcuts portal unavailable): %w", err)
	}

	p.isListening = true
	log.Println("D-Bus hotkey provider started successfully")
	return nil
}

// Stop stops the D-Bus hotkey listener
func (p *DbusKeyboardProvider) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isListening {
		return
	}

	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}

	p.isListening = false
	log.Println("D-Bus hotkey provider stopped")
}

// RegisterHotkey registers a hotkey callback
func (p *DbusKeyboardProvider) RegisterHotkey(hotkey string, callback func() error) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.callbacks[hotkey]; exists {
		return fmt.Errorf("hotkey %s already registered", hotkey)
	}

	p.callbacks[hotkey] = callback
	log.Printf("D-Bus hotkey registered: %s", hotkey)
	return nil
}

// registerHotkeys registers all hotkeys using the GlobalShortcuts portal
func (p *DbusKeyboardProvider) registerHotkeys() error {
	obj := p.conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	// Step 1: Create a session using Request/Response pattern
	sessionOptions := map[string]dbus.Variant{
		"handle_token":         dbus.MakeVariant("speak_to_ai_session"),
		"session_handle_token": dbus.MakeVariant("speak_to_ai_session_handle"),
	}

	call := obj.Call("org.freedesktop.portal.GlobalShortcuts.CreateSession", 0, sessionOptions)
	if call.Err != nil {
		return fmt.Errorf("failed to create GlobalShortcuts session: %w", call.Err)
	}

	// Get the request handle from the call
	if len(call.Body) == 0 {
		return fmt.Errorf("no request handle returned from CreateSession")
	}

	requestHandle, ok := call.Body[0].(dbus.ObjectPath)
	if !ok {
		return fmt.Errorf("invalid request handle type from CreateSession")
	}

	// Wait for the Response signal to get the session handle
	sessionHandle, err := p.waitForSessionResponse(requestHandle)
	if err != nil {
		return fmt.Errorf("failed to get session handle: %w", err)
	}

	p.sessionHandle = sessionHandle

	// Step 2: Prepare shortcuts for binding
	shortcuts := make([]struct {
		ID   string
		Data map[string]dbus.Variant
	}, 0, len(p.callbacks))

	for hotkey := range p.callbacks {
		shortcutData := map[string]dbus.Variant{
			"description": dbus.MakeVariant(fmt.Sprintf("Speak-to-AI hotkey: %s", hotkey)),
		}
		shortcuts = append(shortcuts, struct {
			ID   string
			Data map[string]dbus.Variant
		}{
			ID:   hotkey,
			Data: shortcutData,
		})
	}

	// Step 3: Bind shortcuts to the session
	bindOptions := map[string]dbus.Variant{}
	call = obj.Call("org.freedesktop.portal.GlobalShortcuts.BindShortcuts", 0,
		dbus.ObjectPath(sessionHandle), shortcuts, "", bindOptions)
	if call.Err != nil {
		return fmt.Errorf("failed to bind shortcuts: %w", call.Err)
	}

	// Step 4: Start listening for shortcut activations
	go p.listenForShortcuts()

	return nil
}

// waitForSessionResponse waits for the Response signal from a CreateSession request
func (p *DbusKeyboardProvider) waitForSessionResponse(requestHandle dbus.ObjectPath) (string, error) {
	// Subscribe to the Response signal
	rule := fmt.Sprintf("type='signal',interface='org.freedesktop.portal.Request',member='Response',path='%s'", requestHandle)
	err := p.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule).Err
	if err != nil {
		return "", fmt.Errorf("failed to add match rule: %w", err)
	}

	// Create a channel to receive the signal
	c := make(chan *dbus.Signal, 1)
	p.conn.Signal(c)

	// Wait for the Response signal with a timeout
	timeout := time.After(5 * time.Second)
	select {
	case sig := <-c:
		if sig.Name == "org.freedesktop.portal.Request.Response" && sig.Path == requestHandle {
			if len(sig.Body) >= 2 {
				// Body[0] is response code, Body[1] is results
				responseCode, ok := sig.Body[0].(uint32)
				if !ok || responseCode != 0 {
					return "", fmt.Errorf("CreateSession request failed with code %v", responseCode)
				}

				results, ok := sig.Body[1].(map[string]dbus.Variant)
				if !ok {
					return "", fmt.Errorf("invalid results format in Response signal")
				}

				sessionHandleVariant, exists := results["session_handle"]
				if !exists {
					return "", fmt.Errorf("session_handle not found in Response results")
				}

				sessionHandle, ok := sessionHandleVariant.Value().(string)
				if !ok {
					return "", fmt.Errorf("invalid session_handle type in Response results")
				}

				return sessionHandle, nil
			}
		}
		return "", fmt.Errorf("unexpected signal received: %s", sig.Name)
	case <-timeout:
		return "", fmt.Errorf("timeout waiting for CreateSession response")
	}
}

// listenForShortcuts listens for shortcut activations from the GlobalShortcuts portal
func (p *DbusKeyboardProvider) listenForShortcuts() {
	// Add signal match rule for the session
	rule := fmt.Sprintf("type='signal',interface='org.freedesktop.portal.GlobalShortcuts',member='Activated',path='%s'", p.sessionHandle)
	p.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)

	// Listen for signals
	c := make(chan *dbus.Signal, 10)
	p.conn.Signal(c)

	for sig := range c {
		if sig.Name == "org.freedesktop.portal.GlobalShortcuts.Activated" {
			if len(sig.Body) >= 2 {
				// Body[0] is session handle, Body[1] is shortcut_id
				if sessionHandle, ok := sig.Body[0].(dbus.ObjectPath); ok && string(sessionHandle) == p.sessionHandle {
					if shortcutId, ok := sig.Body[1].(string); ok {
						if callback, exists := p.callbacks[shortcutId]; exists {
							log.Printf("Hotkey activated: %s", shortcutId)
							if err := callback(); err != nil {
								log.Printf("Error executing hotkey callback: %v", err)
							}
						}
					}
				}
			}
		}
	}
}
