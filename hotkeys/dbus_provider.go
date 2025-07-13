package hotkeys

import (
	"fmt"
	"log"
	"sync"

	"github.com/godbus/dbus/v5"
)

// DBusProvider implements KeyboardEventProvider using D-Bus Global Shortcuts Portal
type DBusProvider struct {
	conn      *dbus.Conn
	callbacks map[string]func() error
	mutex     sync.RWMutex
	running   bool
	stopChan  chan bool
	session   string
	shortcuts map[string]string // Maps hotkey to portal shortcut ID
}

// NewDBusProvider creates a new DBus provider instance
func NewDBusProvider() *DBusProvider {
	return &DBusProvider{
		callbacks: make(map[string]func() error),
		shortcuts: make(map[string]string),
		stopChan:  make(chan bool),
	}
}

// IsSupported checks if D-Bus Global Shortcuts Portal is available
func (p *DBusProvider) IsSupported() bool {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return false
	}
	defer conn.Close()

	// Check if the Global Shortcuts portal is available
	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	call := obj.Call("org.freedesktop.portal.GlobalShortcuts.CreateSession", 0, map[string]dbus.Variant{
		"session_handle_token": dbus.MakeVariant("test"),
	})

	return call.Err == nil
}

// Start initializes the D-Bus connection and creates a session
func (p *DBusProvider) Start() error {
	var err error
	p.conn, err = dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %v", err)
	}

	// Create session for global shortcuts
	obj := p.conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	call := obj.Call("org.freedesktop.portal.GlobalShortcuts.CreateSession", 0, map[string]dbus.Variant{
		"session_handle_token": dbus.MakeVariant("speak_to_ai"),
	})

	if call.Err != nil {
		return fmt.Errorf("failed to create global shortcuts session: %v", call.Err)
	}

	var sessionPath dbus.ObjectPath
	if err := call.Store(&sessionPath); err != nil {
		return fmt.Errorf("failed to get session path: %v", err)
	}

	p.session = string(sessionPath)
	p.running = true

	// Start listening for shortcut activations
	go p.listenForActivations()

	return nil
}

// Stop closes the D-Bus connection and cleans up
func (p *DBusProvider) Stop() {
	if !p.running {
		return
	}

	p.running = false
	p.stopChan <- true

	if p.conn != nil {
		p.conn.Close()
	}
}

// RegisterHotkey registers a hotkey with the portal
func (p *DBusProvider) RegisterHotkey(hotkey string, callback func() error) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return fmt.Errorf("DBus provider not started")
	}

	// Convert hotkey to portal format
	shortcutID := fmt.Sprintf("speak_to_ai_%s", hotkey)

	// Register with the portal
	obj := p.conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	shortcuts := map[string]dbus.Variant{
		shortcutID: dbus.MakeVariant(map[string]dbus.Variant{
			"description":       dbus.MakeVariant("Speak to AI hotkey"),
			"preferred_trigger": dbus.MakeVariant(hotkey),
		}),
	}

	call := obj.Call("org.freedesktop.portal.GlobalShortcuts.BindShortcuts", 0,
		dbus.ObjectPath(p.session), shortcuts, "", map[string]dbus.Variant{})

	if call.Err != nil {
		return fmt.Errorf("failed to bind shortcut %s: %v", hotkey, call.Err)
	}

	p.callbacks[shortcutID] = callback
	p.shortcuts[hotkey] = shortcutID

	log.Printf("Registered hotkey %s with ID %s", hotkey, shortcutID)
	return nil
}

// listenForActivations listens for shortcut activation signals
func (p *DBusProvider) listenForActivations() {
	if err := p.conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.portal.GlobalShortcuts"),
		dbus.WithMatchMember("Activated"),
	); err != nil {
		log.Printf("Failed to add match signal: %v", err)
		return
	}

	ch := make(chan *dbus.Signal, 10)
	p.conn.Signal(ch)

	for {
		select {
		case sig := <-ch:
			if sig.Name == "org.freedesktop.portal.GlobalShortcuts.Activated" {
				p.handleActivation(sig)
			}
		case <-p.stopChan:
			return
		}
	}
}

// handleActivation handles shortcut activation signals
func (p *DBusProvider) handleActivation(sig *dbus.Signal) {
	if len(sig.Body) < 2 {
		return
	}

	shortcutID, ok := sig.Body[1].(string)
	if !ok {
		return
	}

	p.mutex.RLock()
	callback, exists := p.callbacks[shortcutID]
	p.mutex.RUnlock()

	if exists && callback != nil {
		go func() {
			if err := callback(); err != nil {
				log.Printf("Error executing hotkey callback: %v", err)
			}
		}()
	}
}
