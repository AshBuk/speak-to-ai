package hotkeys

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	evdev "github.com/gvalkov/golang-evdev"
)

// EvdevKeyboardProvider implements KeyboardEventProvider using Linux evdev
type EvdevKeyboardProvider struct {
	config        HotkeyConfig
	environment   EnvironmentType
	devices       []*evdev.InputDevice
	callbacks     map[string]func() error
	stopListening chan bool
	isListening   bool
	modifierState map[string]bool // For tracking modifier keys state
}

// NewEvdevKeyboardProvider creates a new EvdevKeyboardProvider instance
func NewEvdevKeyboardProvider(config HotkeyConfig, environment EnvironmentType) *EvdevKeyboardProvider {
	return &EvdevKeyboardProvider{
		config:        config,
		environment:   environment,
		callbacks:     make(map[string]func() error),
		stopListening: make(chan bool),
		isListening:   false,
		modifierState: make(map[string]bool),
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
		dev.File.Close()
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
			log.Printf("Warning: could not open input device %s: %v", path, err)
			continue
		}

		// Check if this device is a keyboard
		if strings.Contains(strings.ToLower(dev.Name), "keyboard") ||
			hasKeyEvents(dev) {
			devices = append(devices, dev)
		} else {
			dev.File.Close()
		}
	}

	return devices, nil
}

// hasKeyEvents checks if a device has key events
func hasKeyEvents(dev *evdev.InputDevice) bool {
	// Check if the device supports key events (EV_KEY)
	for evType := range dev.Capabilities {
		// EV_KEY is evdev type 1, but we need to compare by proper type
		if evType.Type == 1 { // EV_KEY
			return len(dev.Capabilities[evType]) > 0
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

				// Read events with a timeout
				events, err := p.devices[deviceIndex].Read()
				if err != nil {
					continue
				}

				// Process events
				for _, event := range events {
					// Only process key events (type 1)
					if event.Type == 1 {
						p.handleKeyEvent(deviceIndex, event)
					}
				}
			}
		}()
	}

	return nil
}

// handleKeyEvent processes a key event from a device
func (p *EvdevKeyboardProvider) handleKeyEvent(deviceIndex int, event evdev.InputEvent) {
	// Get key name
	keyCode := int(event.Code)
	keyName := getKeyName(keyCode)
	if keyName == "" {
		keyName = fmt.Sprintf("KEY_%d", keyCode)
	}

	// Track modifier key state
	if IsModifier(keyName) {
		// Value 1 = key down, 0 = key up
		p.modifierState[strings.ToLower(keyName)] = (event.Value == 1)
	}

	// Only process key down events (value 1)
	if event.Value != 1 {
		return
	}

	// Check for registered hotkeys
	for hotkeyStr, callback := range p.callbacks {
		hotkey := ParseHotkey(hotkeyStr)

		// Check if the pressed key matches the hotkey's main key
		if strings.EqualFold(hotkey.Key, keyName) {
			// Check if all required modifiers are pressed
			allModifiersPressed := true
			for _, mod := range hotkey.Modifiers {
				// Convert general modifier names to specific ones
				evdevModifier := ConvertModifierToEvdev(mod)
				if !p.modifierState[evdevModifier] {
					allModifiersPressed = false
					break
				}
			}

			// If all conditions met, trigger the callback
			if allModifiersPressed {
				go callback()
			}
		}
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
		dev.File.Close()
	}

	p.devices = nil
	p.isListening = false
}
