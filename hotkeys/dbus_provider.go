package hotkeys

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

// DbusKeyboardProvider implements KeyboardEventProvider using D-Bus portal
type DbusKeyboardProvider struct {
	config      HotkeyConfig
	environment EnvironmentType
	callbacks   map[string]func() error
	conn        *dbus.Conn
	isListening bool
	mutex       sync.Mutex
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

	for hotkey, callback := range p.callbacks {
		// Create shortcut options
		options := map[string]dbus.Variant{
			"description": dbus.MakeVariant(fmt.Sprintf("Speak-to-AI hotkey: %s", hotkey)),
		}

		// Register the shortcut
		call := obj.Call("org.freedesktop.portal.GlobalShortcuts.CreateShortcut", 0,
			"",     // parent_window (empty for no parent)
			hotkey, // shortcut_id
			hotkey, // shortcut
			options)

		if call.Err != nil {
			log.Printf("Warning: failed to register hotkey %s: %v", hotkey, call.Err)
			continue
		}

		// Start listening for this shortcut
		go p.listenForShortcut(hotkey, callback)
	}

	return nil
}

// listenForShortcut listens for a specific shortcut activation
func (p *DbusKeyboardProvider) listenForShortcut(shortcut string, callback func() error) {
	// Add signal match rule
	rule := "type='signal',interface='org.freedesktop.portal.GlobalShortcuts',member='Activated',path='/org/freedesktop/portal/desktop'"
	p.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)

	// Listen for signals
	c := make(chan *dbus.Signal, 10)
	p.conn.Signal(c)

	for sig := range c {
		if sig.Name == "org.freedesktop.portal.GlobalShortcuts.Activated" {
			if len(sig.Body) > 0 {
				if activatedShortcut, ok := sig.Body[0].(string); ok && activatedShortcut == shortcut {
					log.Printf("Hotkey activated: %s", shortcut)
					if err := callback(); err != nil {
						log.Printf("Error executing hotkey callback: %v", err)
					}
				}
			}
		}
	}
}
