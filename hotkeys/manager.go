package hotkeys

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

// HotkeyManager handles keyboard shortcuts
type HotkeyManager struct {
	config           HotkeyConfig
	isListening      bool
	isRecording      bool
	stopListening    chan bool
	recordingStarted func() error
	recordingStopped func() error
	hotkeysMutex     sync.Mutex
	environment      EnvironmentType
	provider         KeyboardEventProvider
	modifierState    map[string]bool // Track state of modifier keys
}

// NewHotkeyManager creates a new instance of HotkeyManager
func NewHotkeyManager(config HotkeyConfig, environment EnvironmentType) *HotkeyManager {
	manager := &HotkeyManager{
		config:        config,
		isListening:   false,
		isRecording:   false,
		stopListening: make(chan bool),
		environment:   environment,
		modifierState: make(map[string]bool),
	}

	// Initialize the appropriate keyboard provider based on environment and privileges
	manager.provider = manager.selectKeyboardProvider()

	return manager
}

// selectKeyboardProvider chooses the most appropriate keyboard provider based on
// environment and availability
func (h *HotkeyManager) selectKeyboardProvider() KeyboardEventProvider {
	// Try D-Bus provider first (works without root permissions on modern DEs)
	dbusProvider := NewDbusKeyboardProvider(h.config, h.environment)
	if dbusProvider.IsSupported() {
		log.Println("Using D-Bus keyboard provider (GNOME/KDE)")
		return dbusProvider
	}
	log.Println("D-Bus GlobalShortcuts portal not available, trying evdev...")

	// Fallback to evdev provider (requires root permissions but works everywhere)
	evdevProvider := NewEvdevKeyboardProvider(h.config, h.environment)
	if evdevProvider.IsSupported() {
		log.Println("Using evdev keyboard provider (requires root permissions)")
		return evdevProvider
	}
	log.Println("evdev not available, hotkeys will be disabled")

	// Final fallback to dummy provider with helpful instructions
	log.Println("Warning: No supported keyboard provider available.")
	log.Println("For hotkeys to work:")
	log.Println("  - On GNOME/KDE: Ensure D-Bus session is running")
	log.Println("  - On other DEs: Run with sudo or add user to 'input' group")
	log.Println("  - Alternative: Use system-wide hotkey tools like sxhkd")
	return NewDummyKeyboardProvider()
}

// RegisterCallbacks registers callback functions for hotkey actions
func (h *HotkeyManager) RegisterCallbacks(
	recordingStarted func() error,
	recordingStopped func() error,
) {
	h.recordingStarted = recordingStarted
	h.recordingStopped = recordingStopped
}

// ParseHotkey converts string representation to KeyCombination
func ParseHotkey(hotkeyStr string) KeyCombination {
	combo := KeyCombination{}
	parts := strings.Split(hotkeyStr, "+")

	// If there's only one part, it's just a key
	if len(parts) == 1 {
		combo.Key = strings.TrimSpace(parts[0])
		return combo
	}

	// Last part is the key, the rest are modifiers
	combo.Key = strings.TrimSpace(parts[len(parts)-1])
	for i := 0; i < len(parts)-1; i++ {
		modifier := strings.ToLower(strings.TrimSpace(parts[i]))
		combo.Modifiers = append(combo.Modifiers, modifier)
	}

	return combo
}

// Start begins listening for hotkeys
func (h *HotkeyManager) Start() error {
	if h.isListening {
		return fmt.Errorf("hotkey manager is already running")
	}

	// Check if provider is available
	if h.provider == nil {
		return fmt.Errorf("no keyboard provider available - hotkeys will not work")
	}

	h.isListening = true

	log.Println("Starting hotkey manager...")
	log.Printf("- Start/Stop recording: %s", h.config.GetStartRecordingHotkey())

	// Register hotkeys with the provider
	err := h.provider.RegisterHotkey(h.config.GetStartRecordingHotkey(), func() error {
		h.hotkeysMutex.Lock()
		defer h.hotkeysMutex.Unlock()

		if !h.isRecording && h.recordingStarted != nil {
			log.Println("Start recording hotkey detected")
			if err := h.recordingStarted(); err != nil {
				log.Printf("Error starting recording: %v", err)
				return err
			}
			h.isRecording = true
		} else if h.isRecording && h.recordingStopped != nil {
			log.Println("Stop recording hotkey detected")
			if err := h.recordingStopped(); err != nil {
				log.Printf("Error stopping recording: %v", err)
				return err
			}
			h.isRecording = false
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register start/stop recording hotkey: %w", err)
	}

	// Start the provider
	err = h.provider.Start()
	if err != nil {
		h.isListening = false // Reset state if provider failed to start
		return fmt.Errorf("failed to start keyboard provider: %w", err)
	}

	return nil
}

// Stop stops the hotkey listener
func (h *HotkeyManager) Stop() {
	if h.isListening {
		h.provider.Stop()
		h.isListening = false
	}
}

// IsRecording returns the current recording state
func (h *HotkeyManager) IsRecording() bool {
	return h.isRecording
}

// SimulateHotkeyPress simulates a hotkey press for testing
func (h *HotkeyManager) SimulateHotkeyPress(hotkeyName string) error {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()

	switch hotkeyName {
	case "start_recording":
		if !h.isRecording && h.recordingStarted != nil {
			if err := h.recordingStarted(); err != nil {
				return err
			}
			h.isRecording = true
		}
	case "stop_recording":
		if h.isRecording && h.recordingStopped != nil {
			if err := h.recordingStopped(); err != nil {
				return err
			}
			h.isRecording = false
		}
	default:
		return fmt.Errorf("unknown hotkey: %s", hotkeyName)
	}

	return nil
}

// IsModifier returns true if the key name is a modifier key
func IsModifier(keyName string) bool {
	modifiers := map[string]bool{
		"ctrl":       true,
		"alt":        true,
		"shift":      true,
		"super":      true,
		"meta":       true,
		"win":        true,
		"altgr":      true, // AltGr modifier for international keyboards
		"hyper":      true, // Hyper modifier
		"leftctrl":   true,
		"rightctrl":  true,
		"leftalt":    true,
		"rightalt":   true,
		"leftshift":  true,
		"rightshift": true,
	}

	return modifiers[strings.ToLower(keyName)]
}

// ConvertModifierToEvdev converts common modifier names to evdev key names
func ConvertModifierToEvdev(modifier string) string {
	modifierMap := map[string]string{
		"ctrl":  "leftctrl",
		"alt":   "leftalt",
		"shift": "leftshift",
		"super": "leftmeta",
		"meta":  "leftmeta",
		"win":   "leftmeta",
		"altgr": "rightalt",
	}

	if evdevName, ok := modifierMap[strings.ToLower(modifier)]; ok {
		return evdevName
	}

	return strings.ToLower(modifier)
}
