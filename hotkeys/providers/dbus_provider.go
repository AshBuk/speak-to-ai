//go:build linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/utils"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	dbus "github.com/godbus/dbus/v5"
)

// Implements KeyboardEventProvider using the D-Bus GlobalShortcuts portal
type DbusKeyboardProvider struct {
	config        adapters.HotkeyConfig
	environment   interfaces.EnvironmentType
	callbacks     map[string]func() error
	conn          *dbus.Conn
	sessionHandle string
	isListening   bool
	mutex         sync.Mutex
	logger        logger.Logger
	wg            sync.WaitGroup // Tracks listener goroutine
}

// Create a new D-Bus keyboard provider
func NewDbusKeyboardProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) *DbusKeyboardProvider {
	return &DbusKeyboardProvider{
		config:      config,
		environment: environment,
		callbacks:   make(map[string]func() error),
		isListening: false,
		logger:      logger,
	}
}

// Check if the D-Bus GlobalShortcuts portal is available
func (p *DbusKeyboardProvider) IsSupported() bool {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		p.logger.Warning("D-Bus session bus not available: %v", err)
		return false
	}
	defer func() {
		if err := conn.Close(); err != nil {
			p.logger.Error("Failed to close D-Bus connection: %v", err)
		}
	}()

	// Check if the GlobalShortcuts portal is available by introspecting the desktop portal
	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	call := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0)
	if call.Err != nil {
		p.logger.Warning("D-Bus portal not available: %v", call.Err)
		return false
	}

	var introspectData string
	if err := call.Store(&introspectData); err != nil {
		p.logger.Error("Failed to get introspection data: %v", err)
		return false
	}

	if len(introspectData) > 0 && containsGlobalShortcuts(introspectData) {
		p.logger.Info("D-Bus portal GlobalShortcuts detected")
		return true
	}

	p.logger.Info("D-Bus portal GlobalShortcuts not available")
	return false
}

// Check if the introspection data contains the GlobalShortcuts interface
func containsGlobalShortcuts(data string) bool {
	return strings.Contains(data, "GlobalShortcuts")
}

// Start listening for D-Bus hotkey events
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

	// Register hotkeys via the GlobalShortcuts portal
	if err := p.registerHotkeys(); err != nil {
		if closeErr := p.conn.Close(); closeErr != nil {
			p.logger.Error("Failed to close D-Bus connection: %v", closeErr)
		}
		p.logger.Error("DBus GlobalShortcuts binding failed: %v", err)
		p.logger.Info("Hint: In Flatpak/AppImage, global shortcuts may require user consent")
		return fmt.Errorf("failed to register hotkeys (GlobalShortcuts portal unavailable): %w", err)
	}

	p.isListening = true
	p.logger.Info("D-Bus hotkey provider started successfully")
	return nil
}

// Stop the D-Bus hotkey listener
func (p *DbusKeyboardProvider) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isListening {
		return
	}

	// Close D-Bus connection to unblock signal channel
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			p.logger.Error("Failed to close D-Bus connection: %v", err)
		}
		p.conn = nil
	}

	p.isListening = false

	// Wait for listener goroutine to exit
	p.wg.Wait()

	p.logger.Info("D-Bus hotkey provider stopped")
}

// Register a hotkey and its callback
func (p *DbusKeyboardProvider) RegisterHotkey(hotkey string, callback func() error) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.callbacks[hotkey]; exists {
		return fmt.Errorf("hotkey %s already registered", hotkey)
	}

	p.callbacks[hotkey] = callback
	p.logger.Info("D-Bus hotkey registered: %s", hotkey)
	return nil
}

// Return an error as this provider does not support capture-once functionality
func (p *DbusKeyboardProvider) CaptureOnce(timeout time.Duration) (string, error) {
	return "", fmt.Errorf("captureOnce not supported in dbus provider")
}

// Convert a hotkey string to a desktop-portal accelerator string
// e.g., "ctrl+shift+a" -> "<Ctrl><Shift>a"
func convertHotkeyToAccelerator(hotkey string) string {
	combo := utils.ParseHotkey(hotkey)
	var prefix strings.Builder
	for _, m := range combo.Modifiers {
		switch strings.ToLower(m) {
		case "ctrl", "leftctrl", "rightctrl":
			prefix.WriteString("<Ctrl>")
		case "alt", "leftalt":
			prefix.WriteString("<Alt>")
		case "rightalt", "altgr":
			prefix.WriteString("<AltGr>")
		case "shift", "leftshift", "rightshift":
			prefix.WriteString("<Shift>")
		case "super", "meta", "win", "leftmeta", "rightmeta":
			prefix.WriteString("<Super>")
		}
	}

	// Map special key names to the standard accelerator format
	key := combo.Key
	switch strings.ToLower(key) {
	case "comma":
		key = "comma"
	case "period":
		key = "period"
	case "space":
		key = "space"
	case "enter", "return":
		key = "Return"
	case "tab":
		key = "Tab"
	case "escape", "esc":
		key = "Escape"
	case "backspace":
		key = "BackSpace"
	case "delete", "del":
		key = "Delete"
	}

	return prefix.String() + key
}

// Register all hotkeys using the GlobalShortcuts portal
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

	// Step 2: Prepare shortcuts for binding (include accelerator)
	// Format according to GlobalShortcuts portal spec: a(sa{sv})
	shortcuts := make([]struct {
		ID   string
		Data map[string]dbus.Variant
	}, 0, len(p.callbacks))

	for hotkey := range p.callbacks {
		accel := convertHotkeyToAccelerator(hotkey)
		p.logger.Info("DBus: Converting hotkey '%s' to accelerator '%s'", hotkey, accel)

		shortcutData := map[string]dbus.Variant{
			"description":       dbus.MakeVariant(fmt.Sprintf("Speak-to-AI hotkey: %s", hotkey)),
			"preferred_trigger": dbus.MakeVariant(accel),
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
	p.logger.Info("DBus: Binding %d shortcuts to session %s", len(shortcuts), sessionHandle)

	call = obj.Call("org.freedesktop.portal.GlobalShortcuts.BindShortcuts", 0,
		dbus.ObjectPath(sessionHandle), shortcuts, "", bindOptions)
	if call.Err != nil {
		return fmt.Errorf("failed to bind shortcuts: %w", call.Err)
	}

	p.logger.Info("DBus: Successfully bound shortcuts")

	// Step 4: Start listening for shortcut activations in tracked goroutine
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.listenForShortcuts()
	}()

	return nil
}

// Wait for the Response signal from a CreateSession request
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

// Listen for shortcut activation signals from the GlobalShortcuts portal
func (p *DbusKeyboardProvider) listenForShortcuts() {
	// Add a signal match rule for the session
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
							p.logger.Info("Hotkey activated: %s", shortcutId)
							if err := callback(); err != nil {
								p.logger.Error("Error executing hotkey callback: %v", err)
							}
						}
					}
				}
			}
		}
	}
}
