//go:build linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/utils"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	evdev "github.com/holoplot/go-evdev"
)

// EvdevKeyboardProvider implements KeyboardEventProvider using Linux evdev
type EvdevKeyboardProvider struct {
	config        adapters.HotkeyConfig
	environment   interfaces.EnvironmentType
	devices       []*evdev.InputDevice
	callbacks     map[string]func() error
	stopListening chan bool
	isListening   bool
	modifierState map[string]bool // For tracking modifier keys state
	mutex         sync.RWMutex
	logger        logger.Logger
}

// NewEvdevKeyboardProvider creates a new EvdevKeyboardProvider instance
func NewEvdevKeyboardProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) *EvdevKeyboardProvider {
	return &EvdevKeyboardProvider{
		config:        config,
		environment:   environment,
		callbacks:     make(map[string]func() error),
		stopListening: make(chan bool),
		isListening:   false,
		modifierState: make(map[string]bool),
		logger:        logger,
	}
}

// IsSupported checks if evdev is supported on this system
func (p *EvdevKeyboardProvider) IsSupported() bool {
	// Try to find devices, but don't keep them open
	devices, err := p.findKeyboardDevices()
	if err != nil || len(devices) == 0 {
		return false
	}

	// Close devices since we're just testing
	for _, dev := range devices {
		if err := dev.Close(); err != nil {
			p.logger.Error("Failed to close evdev device: %v", err)
		}
	}

	return true
}

// findKeyboardDevices finds keyboard input devices
func (p *EvdevKeyboardProvider) findKeyboardDevices() ([]*evdev.InputDevice, error) {
	devices := []*evdev.InputDevice{}

	// List all input devices
	devicePaths, err := filepath.Glob("/dev/input/event*")
	if err != nil {
		return nil, fmt.Errorf("failed to list input devices: %w", err)
	}

	// Filter devices that are keyboards
	for _, path := range devicePaths {
		dev, err := evdev.Open(path)
		if err != nil {
			p.logger.Warning("Could not open input device %s: %v", path, err)
			continue
		}

		// Check if this device is a keyboard (avoid mice/gamepads)
		devName, _ := dev.Name()
		if strings.Contains(strings.ToLower(devName), "keyboard") ||
			isKeyboardDevice(dev) {
			devices = append(devices, dev)
		} else {
			if err := dev.Close(); err != nil {
				p.logger.Error("Failed to close evdev device: %v", err)
			}
		}
	}

	return devices, nil
}

// isKeyboardDevice checks if device exposes typical keyboard key codes (letters)
func isKeyboardDevice(dev *evdev.InputDevice) bool {
	types := dev.CapableTypes()
	isKey := false
	for _, evType := range types {
		if evType == evdev.EV_KEY {
			isKey = true
			break
		}
	}
	if !isKey {
		return false
	}
	// Presence of common letter key codes strongly indicates a keyboard
	events := dev.CapableEvents(evdev.EV_KEY)
	common := map[uint16]bool{16: true, 30: true, 44: true, 57: true} // q, a, z, space
	for _, code := range events {
		if common[uint16(code)] {
			return true
		}
	}
	return false
}

// Start begins listening for keyboard events
func (p *EvdevKeyboardProvider) Start() error {
	if p.isListening {
		return fmt.Errorf("evdev keyboard provider already started")
	}

	var err error
	p.devices, err = p.findKeyboardDevices()
	if err != nil {
		return fmt.Errorf("failed to find keyboard devices: %w", err)
	}

	if len(p.devices) == 0 {
		return fmt.Errorf("no keyboard devices found")
	}

	p.isListening = true

	// Start listening for events from each device
	for i := range p.devices {
		// Capture i in closure
		deviceIndex := i
		go func() {
			for {
				// Check if we should stop
				select {
				case <-p.stopListening:
					return
				default:
					// Continue with the event processing
				}

				// Read single event
				event, err := p.devices[deviceIndex].ReadOne()
				if err != nil {
					continue
				}

				// Only process key events
				if event.Type == evdev.EV_KEY {
					p.handleKeyEvent(deviceIndex, event)
				}
			}
		}()
	}

	return nil
}

// handleKeyEvent processes a key event from a device
func (p *EvdevKeyboardProvider) handleKeyEvent(_ int, event *evdev.InputEvent) {
	// Get key name
	keyCode := int(event.Code)
	keyName := getKeyName(keyCode)
	if keyName == "" {
		keyName = fmt.Sprintf("KEY_%d", keyCode)
	}

	// Track modifier key state
	if utils.IsModifier(keyName) || keyName == "leftmeta" || keyName == "rightmeta" {
		// Value 1 = key down, 0 = key up
		p.mutex.Lock()
		p.modifierState[strings.ToLower(keyName)] = (event.Value == 1)
		p.mutex.Unlock()
	}

	// Only process key down events (value 1)
	if event.Value != 1 {
		return
	}

	// Copy callbacks and modifier state under lock
	p.mutex.RLock()
	callbacksCopy := make(map[string]func() error, len(p.callbacks))
	for k, v := range p.callbacks {
		callbacksCopy[k] = v
	}
	modState := make(map[string]bool, len(p.modifierState))
	for k, v := range p.modifierState {
		modState[k] = v
	}
	p.mutex.RUnlock()

	// Check for registered hotkeys
	for hotkeyStr, callback := range callbacksCopy {
		hotkey := utils.ParseHotkey(hotkeyStr)

		// Check if the pressed key matches the hotkey's main key
		if strings.EqualFold(hotkey.Key, keyName) {
			// Check if all required modifiers are pressed
			allModifiersPressed := true
			for _, mod := range hotkey.Modifiers {
				if !modifierPressed(mod, modState) {
					allModifiersPressed = false
					break
				}
			}

			// If all conditions met, trigger the callback
			if allModifiersPressed {
				go func(cb func() error) {
					if err := cb(); err != nil {
						p.logger.Error("Hotkey callback error: %v", err)
					}
				}(callback)
			}
		}
	}
}

// modifierPressed determines if a modifier (generic or side-specific) is pressed
func modifierPressed(mod string, state map[string]bool) bool {
	m := strings.ToLower(mod)
	switch m {
	case "ctrl":
		return state["leftctrl"] || state["rightctrl"]
	case "leftctrl":
		return state["leftctrl"]
	case "rightctrl":
		return state["rightctrl"]
	case "alt":
		return state["leftalt"] || state["rightalt"]
	case "leftalt":
		return state["leftalt"]
	case "rightalt", "altgr":
		return state["rightalt"]
	case "shift":
		return state["leftshift"] || state["rightshift"]
	case "leftshift":
		return state["leftshift"]
	case "rightshift":
		return state["rightshift"]
	case "super", "meta", "win":
		return state["leftmeta"] || state["rightmeta"]
	case "leftmeta":
		return state["leftmeta"]
	case "rightmeta":
		return state["rightmeta"]
	default:
		// Fallback to simple mapping for any other names
		return state[utils.ConvertModifierToEvdev(m)]
	}
}

// Common key code to name mapping
func getKeyName(keyCode int) string {
	keyMap := map[int]string{
		1:   "esc",
		2:   "1",
		3:   "2",
		4:   "3",
		5:   "4",
		6:   "5",
		7:   "6",
		8:   "7",
		9:   "8",
		10:  "9",
		11:  "0",
		12:  "minus",
		13:  "equal",
		14:  "backspace",
		15:  "tab",
		16:  "q",
		17:  "w",
		18:  "e",
		19:  "r",
		20:  "t",
		21:  "y",
		22:  "u",
		23:  "i",
		24:  "o",
		25:  "p",
		26:  "leftbrace",
		27:  "rightbrace",
		28:  "enter",
		29:  "leftctrl",
		30:  "a",
		31:  "s",
		32:  "d",
		33:  "f",
		34:  "g",
		35:  "h",
		36:  "j",
		37:  "k",
		38:  "l",
		39:  "semicolon",
		40:  "apostrophe",
		41:  "grave",
		42:  "leftshift",
		43:  "backslash",
		44:  "z",
		45:  "x",
		46:  "c",
		47:  "v",
		48:  "b",
		49:  "n",
		50:  "m",
		51:  "comma",
		52:  "dot",
		53:  "slash",
		54:  "rightshift",
		55:  "kpasterisk",
		56:  "leftalt",
		57:  "space",
		58:  "capslock",
		59:  "f1",
		60:  "f2",
		61:  "f3",
		62:  "f4",
		63:  "f5",
		64:  "f6",
		65:  "f7",
		66:  "f8",
		67:  "f9",
		68:  "f10",
		69:  "numlock",
		70:  "scrolllock",
		71:  "kp7",
		72:  "kp8",
		73:  "kp9",
		74:  "kpminus",
		75:  "kp4",
		76:  "kp5",
		77:  "kp6",
		78:  "kpplus",
		79:  "kp1",
		80:  "kp2",
		81:  "kp3",
		82:  "kp0",
		83:  "kpdot",
		97:  "rightctrl",
		100: "rightalt",
		125: "leftmeta",
		126: "rightmeta",
	}

	if name, ok := keyMap[keyCode]; ok {
		return name
	}
	return ""
}

// RegisterHotkey registers a callback for a hotkey
func (p *EvdevKeyboardProvider) RegisterHotkey(hotkey string, callback func() error) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.callbacks[hotkey] = callback
	return nil
}

// Stop stops listening for keyboard events
func (p *EvdevKeyboardProvider) Stop() {
	if !p.isListening {
		return
	}

	// Signal all listeners to stop
	close(p.stopListening)

	// Close all devices
	for _, dev := range p.devices {
		if err := dev.Close(); err != nil {
			p.logger.Error("Failed to close evdev device: %v", err)
		}
	}

	p.devices = nil
	p.isListening = false
}

// CaptureOnce starts a short-lived capture session and returns a single normalized hotkey string
// or an error on timeout/cancel. It does not depend on Start/Stop lifecycle.
func (p *EvdevKeyboardProvider) CaptureOnce(timeout time.Duration) (string, error) {
	devices, err := p.findKeyboardDevices()
	if err != nil {
		return "", fmt.Errorf("failed to find keyboard devices: %w", err)
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("no keyboard devices found")
	}

	defer func() {
		for _, d := range devices {
			if closeErr := d.Close(); closeErr != nil {
				p.logger.Error("Failed to close evdev device: %v", closeErr)
			}
		}
	}()

	// Local modifier state independent from main listener
	modState := map[string]bool{}
	resultCh := make(chan string, 1)
	stopCh := make(chan struct{})

	for i := range devices {
		idx := i
		go func() {
			for {
				select {
				case <-stopCh:
					return
				default:
				}
				ev, err := devices[idx].ReadOne()
				if err != nil {
					return
				}
				if ev.Type != evdev.EV_KEY {
					continue
				}

				keyCode := int(ev.Code)
				keyName := getKeyName(keyCode)
				if keyName == "" {
					keyName = fmt.Sprintf("KEY_%d", keyCode)
				}

				// Ignore non-keyboard buttons (e.g., mouse BTN_* produce KEY_### fallback)
				if strings.HasPrefix(keyName, "KEY_") {
					continue
				}

				// Track modifiers (down/up)
				if utils.IsModifier(keyName) || keyName == "leftmeta" || keyName == "rightmeta" || keyName == "leftctrl" || keyName == "rightctrl" || keyName == "leftalt" || keyName == "rightalt" || keyName == "leftshift" || keyName == "rightshift" {
					modState[strings.ToLower(keyName)] = (ev.Value == 1)
					continue
				}

				// Only consider key down for main key
				if ev.Value != 1 {
					continue
				}

				// Cancel if Esc pressed without modifiers
				noMods := !modState["leftctrl"] && !modState["rightctrl"] &&
					!modState["leftshift"] && !modState["rightshift"] &&
					!modState["leftalt"] && !modState["rightalt"] &&
					!modState["leftmeta"] && !modState["rightmeta"]
				if strings.EqualFold(keyName, "esc") && noMods {
					select {
					case resultCh <- "":
					default:
					}
					return
				}

				mods := make([]string, 0, 5)
				if modState["leftctrl"] || modState["rightctrl"] {
					mods = append(mods, "ctrl")
				}
				if modState["leftshift"] || modState["rightshift"] {
					mods = append(mods, "shift")
				}
				if modState["leftalt"] || modState["rightalt"] {
					mods = append(mods, "alt")
				}
				if modState["rightalt"] {
					mods = append(mods, "altgr")
				}
				if modState["leftmeta"] || modState["rightmeta"] {
					mods = append(mods, "super")
				}

				combo := strings.ToLower(keyName)
				if len(mods) > 0 {
					combo = strings.Join(mods, "+") + "+" + combo
				}
				combo = utils.NormalizeHotkey(combo)
				// Basic validation; skip if invalid
				if err := utils.ValidateHotkey(combo); err != nil {
					continue
				}

				select {
				case resultCh <- combo:
				default:
				}
				return
			}
		}()
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var result string
	select {
	case result = <-resultCh:
		close(stopCh)
	case <-timer.C:
		close(stopCh)
		return "", fmt.Errorf("capture timeout")
	}

	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("capture cancelled")
	}
	return result, nil
}
