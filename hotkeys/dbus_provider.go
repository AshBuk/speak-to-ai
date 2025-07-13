package hotkeys

import (
	"fmt"
	"log"
	"sync"

	"github.com/godbus/dbus/v5"
)

// DbusKeyboardProvider implements KeyboardEventProvider using D-Bus
type DbusKeyboardProvider struct {
	config        HotkeyConfig
	environment   EnvironmentType
	callbacks     map[string]func() error
	conn          *dbus.Conn
	isListening   bool
	stopListening chan bool
	mutex         sync.Mutex
	registeredIds []uint32 // Track registered hotkey IDs
}

// NewDbusKeyboardProvider creates a new D-Bus keyboard provider
func NewDbusKeyboardProvider(config HotkeyConfig, environment EnvironmentType) *DbusKeyboardProvider {
	return &DbusKeyboardProvider{
		config:        config,
		environment:   environment,
		callbacks:     make(map[string]func() error),
		isListening:   false,
		stopListening: make(chan bool),
		registeredIds: make([]uint32, 0),
	}
}

// IsSupported checks if D-Bus hotkey registration is supported
func (p *DbusKeyboardProvider) IsSupported() bool {
	// Try to connect to session bus
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Printf("D-Bus session bus not available: %v", err)
		return false
	}
	defer conn.Close()

	// Check if GNOME Shell is available (supports global hotkeys)
	if p.checkGnomeShell(conn) {
		log.Println("GNOME Shell detected - D-Bus hotkeys supported")
		return true
	}

	// Check if KDE Plasma is available
	if p.checkKDEPlasma(conn) {
		log.Println("KDE Plasma detected - D-Bus hotkeys supported")
		return true
	}

	// Check if other desktop environments support global hotkeys
	if p.checkGenericHotkeySupport(conn) {
		log.Println("Generic D-Bus hotkey support detected")
		return true
	}

	log.Println("No D-Bus hotkey support detected")
	return false
}

// checkGnomeShell checks if GNOME Shell is available
func (p *DbusKeyboardProvider) checkGnomeShell(conn *dbus.Conn) bool {
	obj := conn.Object("org.gnome.Shell", "/org/gnome/Shell")
	call := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0)
	return call.Err == nil
}

// checkKDEPlasma checks if KDE Plasma is available
func (p *DbusKeyboardProvider) checkKDEPlasma(conn *dbus.Conn) bool {
	obj := conn.Object("org.kde.kglobalaccel", "/kglobalaccel")
	call := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0)
	return call.Err == nil
}

// checkGenericHotkeySupport checks for generic hotkey support
func (p *DbusKeyboardProvider) checkGenericHotkeySupport(conn *dbus.Conn) bool {
	// Check for freedesktop.org standard hotkey service
	obj := conn.Object("org.freedesktop.impl.portal.desktop.gtk", "/org/freedesktop/portal/desktop")
	call := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0)
	return call.Err == nil
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
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	// Register hotkeys based on desktop environment
	if p.checkGnomeShell(p.conn) {
		err = p.registerGnomeHotkeys()
	} else if p.checkKDEPlasma(p.conn) {
		err = p.registerKDEHotkeys()
	} else {
		err = p.registerGenericHotkeys()
	}

	if err != nil {
		p.conn.Close()
		return fmt.Errorf("failed to register hotkeys: %w", err)
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

	// Unregister all hotkeys
	p.unregisterAllHotkeys()

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

// registerGnomeHotkeys registers hotkeys for GNOME Shell
func (p *DbusKeyboardProvider) registerGnomeHotkeys() error {
	obj := p.conn.Object("org.gnome.Shell", "/org/gnome/Shell")

	for hotkey, callback := range p.callbacks {
		// Convert hotkey format to GNOME format
		gnomeHotkey := p.convertToGnomeFormat(hotkey)

		// Register the hotkey
		call := obj.Call("org.gnome.Shell.Eval", 0, fmt.Sprintf(`
			Main.wm.addKeybinding('%s', new Gio.Settings(), Meta.KeyBindingFlags.NONE, 
				Shell.ActionMode.NORMAL, function() {
					imports.dbus.session.emit_signal(null, '/org/freedesktop/speak_to_ai', 
						'org.freedesktop.speak_to_ai', 'HotkeyPressed', 
						new GLib.Variant('(s)', ['%s']));
				});
		`, gnomeHotkey, hotkey))

		if call.Err != nil {
			return fmt.Errorf("failed to register GNOME hotkey %s: %w", hotkey, call.Err)
		}

		// Store callback for later use
		go p.listenForHotkeySignal(hotkey, callback)
	}

	return nil
}

// registerKDEHotkeys registers hotkeys for KDE Plasma
func (p *DbusKeyboardProvider) registerKDEHotkeys() error {
	obj := p.conn.Object("org.kde.kglobalaccel", "/kglobalaccel")

	for hotkey, callback := range p.callbacks {
		// Convert hotkey format to KDE format
		kdeHotkey := p.convertToKDEFormat(hotkey)

		// Register the hotkey
		call := obj.Call("org.kde.KGlobalAccel.setShortcut", 0,
			"speak-to-ai", hotkey, kdeHotkey, "speak-to-ai", hotkey, 0)

		if call.Err != nil {
			return fmt.Errorf("failed to register KDE hotkey %s: %w", hotkey, call.Err)
		}

		// Store callback for later use
		go p.listenForHotkeySignal(hotkey, callback)
	}

	return nil
}

// registerGenericHotkeys registers hotkeys using generic D-Bus methods
func (p *DbusKeyboardProvider) registerGenericHotkeys() error {
	// Use freedesktop.org portal for generic hotkey support
	obj := p.conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	for hotkey, callback := range p.callbacks {
		// Register via portal
		call := obj.Call("org.freedesktop.portal.GlobalShortcuts.CreateShortcut", 0,
			map[string]interface{}{
				"shortcut_id": hotkey,
				"description": fmt.Sprintf("Speak-to-AI hotkey: %s", hotkey),
			})

		if call.Err != nil {
			log.Printf("Warning: failed to register generic hotkey %s: %v", hotkey, call.Err)
			continue
		}

		// Store callback for later use
		go p.listenForHotkeySignal(hotkey, callback)
	}

	return nil
}

// listenForHotkeySignal listens for hotkey activation signals
func (p *DbusKeyboardProvider) listenForHotkeySignal(hotkey string, callback func() error) {
	// Add match rule for the signal
	rule := fmt.Sprintf("type='signal',interface='org.freedesktop.speak_to_ai',member='HotkeyPressed',path='/org/freedesktop/speak_to_ai'")
	p.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)

	// Listen for signals
	c := make(chan *dbus.Signal, 10)
	p.conn.Signal(c)

	go func() {
		for {
			select {
			case sig := <-c:
				if sig.Name == "org.freedesktop.speak_to_ai.HotkeyPressed" {
					if len(sig.Body) > 0 {
						if pressedHotkey, ok := sig.Body[0].(string); ok && pressedHotkey == hotkey {
							if err := callback(); err != nil {
								log.Printf("Error executing hotkey callback for %s: %v", hotkey, err)
							}
						}
					}
				}
			case <-p.stopListening:
				return
			}
		}
	}()
}

// convertToGnomeFormat converts hotkey string to GNOME format
func (p *DbusKeyboardProvider) convertToGnomeFormat(hotkey string) string {
	// Convert from our format to GNOME format
	// Example: "ctrl+shift+c" -> "<Control><Shift>c"
	combo := ParseHotkey(hotkey)
	result := ""

	for _, mod := range combo.Modifiers {
		switch mod {
		case "ctrl":
			result += "<Control>"
		case "alt":
			result += "<Alt>"
		case "shift":
			result += "<Shift>"
		case "super", "meta", "win":
			result += "<Super>"
		case "altgr":
			result += "<Mod5>"
		}
	}

	result += combo.Key
	return result
}

// convertToKDEFormat converts hotkey string to KDE format
func (p *DbusKeyboardProvider) convertToKDEFormat(hotkey string) string {
	// Convert from our format to KDE format
	// Example: "ctrl+shift+c" -> "Ctrl+Shift+C"
	combo := ParseHotkey(hotkey)
	result := ""

	for i, mod := range combo.Modifiers {
		if i > 0 {
			result += "+"
		}
		switch mod {
		case "ctrl":
			result += "Ctrl"
		case "alt":
			result += "Alt"
		case "shift":
			result += "Shift"
		case "super", "meta", "win":
			result += "Meta"
		case "altgr":
			result += "AltGr"
		}
	}

	if len(combo.Modifiers) > 0 {
		result += "+"
	}
	result += combo.Key
	return result
}

// unregisterAllHotkeys unregisters all registered hotkeys
func (p *DbusKeyboardProvider) unregisterAllHotkeys() {
	if p.conn == nil {
		return
	}

	// Signal all listeners to stop
	close(p.stopListening)
	p.stopListening = make(chan bool)

	// Unregister based on desktop environment
	if p.checkGnomeShell(p.conn) {
		p.unregisterGnomeHotkeys()
	} else if p.checkKDEPlasma(p.conn) {
		p.unregisterKDEHotkeys()
	} else {
		p.unregisterGenericHotkeys()
	}

	p.registeredIds = p.registeredIds[:0]
}

// unregisterGnomeHotkeys unregisters GNOME hotkeys
func (p *DbusKeyboardProvider) unregisterGnomeHotkeys() {
	obj := p.conn.Object("org.gnome.Shell", "/org/gnome/Shell")

	for hotkey := range p.callbacks {
		gnomeHotkey := p.convertToGnomeFormat(hotkey)
		obj.Call("org.gnome.Shell.Eval", 0, fmt.Sprintf(`
			Main.wm.removeKeybinding('%s');
		`, gnomeHotkey))
	}
}

// unregisterKDEHotkeys unregisters KDE hotkeys
func (p *DbusKeyboardProvider) unregisterKDEHotkeys() {
	obj := p.conn.Object("org.kde.kglobalaccel", "/kglobalaccel")

	for hotkey := range p.callbacks {
		obj.Call("org.kde.KGlobalAccel.unregister", 0, "speak-to-ai", hotkey)
	}
}

// unregisterGenericHotkeys unregisters generic hotkeys
func (p *DbusKeyboardProvider) unregisterGenericHotkeys() {
	obj := p.conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	for hotkey := range p.callbacks {
		obj.Call("org.freedesktop.portal.GlobalShortcuts.DeleteShortcut", 0, hotkey)
	}
}
