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
	copyToClipboard  func() error
	pasteToActiveApp func() error
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
	// Try evdev provider first (requires root permissions on Linux)
	evdevProvider := NewEvdevKeyboardProvider(h.config, h.environment)
	if evdevProvider.IsSupported() {
		log.Println("Using evdev keyboard provider")
		return evdevProvider
	}

	// If evdev is not available, try gohook (X11 environments)
	if h.environment == EnvironmentX11 {
		hookProvider := NewHookKeyboardProvider(h.config)
		if hookProvider.IsSupported() {
			log.Println("Using hook-based keyboard provider for X11")
			return hookProvider
		}
	}

	// Fallback to a dummy provider that logs warnings
	log.Println("Warning: No supported keyboard provider available. Hotkeys will not work.")
	return NewDummyKeyboardProvider()
}

// RegisterCallbacks registers callback functions for hotkey actions
func (h *HotkeyManager) RegisterCallbacks(
	recordingStarted func() error,
	recordingStopped func() error,
	copyToClipboard func() error,
	pasteToActiveApp func() error,
) {
	h.recordingStarted = recordingStarted
	h.recordingStopped = recordingStopped
	h.copyToClipboard = copyToClipboard
	h.pasteToActiveApp = pasteToActiveApp
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

	h.isListening = true

	log.Println("Starting hotkey manager...")
	log.Printf("- Start/Stop recording: %s", h.config.GetStartRecordingHotkey())
	log.Printf("- Copy to clipboard: %s", h.config.GetCopyToClipboardHotkey())
	log.Printf("- Paste to active app: %s", h.config.GetPasteToActiveAppHotkey())

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

	err = h.provider.RegisterHotkey(h.config.GetCopyToClipboardHotkey(), func() error {
		if h.copyToClipboard != nil {
			log.Println("Copy to clipboard hotkey detected")
			return h.copyToClipboard()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register copy to clipboard hotkey: %w", err)
	}

	err = h.provider.RegisterHotkey(h.config.GetPasteToActiveAppHotkey(), func() error {
		if h.pasteToActiveApp != nil {
			log.Println("Paste to active app hotkey detected")
			return h.pasteToActiveApp()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register paste to active app hotkey: %w", err)
	}

	// Start the provider
	return h.provider.Start()
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
	case "copy_to_clipboard":
		if h.copyToClipboard != nil {
			return h.copyToClipboard()
		}
	case "paste_to_active_app":
		if h.pasteToActiveApp != nil {
			return h.pasteToActiveApp()
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
	}

	if evdevName, ok := modifierMap[strings.ToLower(modifier)]; ok {
		return evdevName
	}

	return strings.ToLower(modifier)
}
