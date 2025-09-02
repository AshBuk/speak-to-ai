// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package hotkeys

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

// HotkeyAction represents a hotkey action callback
type HotkeyAction func() error

// HotkeyManager handles keyboard shortcuts
type HotkeyManager struct {
	config           HotkeyConfig
	isListening      bool
	isRecording      bool
	stopListening    chan bool
	recordingStarted func() error
	recordingStopped func() error
	hotkeyActions    map[string]HotkeyAction // Additional hotkey actions
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
		hotkeyActions: make(map[string]HotkeyAction),
	}

	// Initialize the appropriate keyboard provider based on environment and privileges
	manager.provider = selectProviderForEnvironment(manager.config, manager.environment)

	return manager
}

// selectProviderForEnvironment is defined per-OS (see manager_linux.go and manager_stub.go)

// RegisterCallbacks registers callback functions for hotkey actions
func (h *HotkeyManager) RegisterCallbacks(
	recordingStarted func() error,
	recordingStopped func() error,
) {
	h.recordingStarted = recordingStarted
	h.recordingStopped = recordingStopped
}

// RegisterHotkeyAction registers a custom hotkey action
func (h *HotkeyManager) RegisterHotkeyAction(hotkey string, action HotkeyAction) {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	h.hotkeyActions[hotkey] = action
}

// UnregisterHotkeyAction removes a hotkey action
func (h *HotkeyManager) UnregisterHotkeyAction(hotkey string) {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	delete(h.hotkeyActions, hotkey)
}

// GetRegisteredHotkeys returns all registered hotkeys
func (h *HotkeyManager) GetRegisteredHotkeys() []string {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()

	var hotkeys []string
	// Add recording hotkeys
	hotkeys = append(hotkeys, h.config.GetStartRecordingHotkey())

	// Add custom hotkeys
	for hotkey := range h.hotkeyActions {
		hotkeys = append(hotkeys, hotkey)
	}

	return hotkeys
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

	// Register additional hotkeys
	h.hotkeysMutex.Lock()
	for hotkey, action := range h.hotkeyActions {
		hotkeyToRegister := hotkey
		actionToExecute := action

		err := h.provider.RegisterHotkey(hotkeyToRegister, func() error {
			log.Printf("Custom hotkey detected: %s", hotkeyToRegister)
			if err := actionToExecute(); err != nil {
				log.Printf("Error executing hotkey action for %s: %v", hotkeyToRegister, err)
				return err
			}
			return nil
		})
		if err != nil {
			h.hotkeysMutex.Unlock()
			return fmt.Errorf("failed to register hotkey %s: %w", hotkeyToRegister, err)
		}
	}
	h.hotkeysMutex.Unlock()

	// Start the provider
	err = h.provider.Start()
	if err != nil {
		// Provider failed to start. Attempt fallback to evdev when available.
		log.Printf("Primary keyboard provider failed to start: %v", err)
		h.isListening = false // Reset state if provider failed to start

		// Fallback only makes sense on Linux where evdev may be available.
		// If current provider is DBus, try evdev as a backup.
		switch h.provider.(type) {
		case *DbusKeyboardProvider:
			fallback := NewEvdevKeyboardProvider(h.config, h.environment)
			if fallback != nil && fallback.IsSupported() {
				log.Println("Falling back to evdev keyboard provider")
				// Swap provider
				h.provider = fallback

				// Re-register hotkeys on the new provider
				if regErr := h.provider.RegisterHotkey(h.config.GetStartRecordingHotkey(), func() error {
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
				}); regErr != nil {
					return fmt.Errorf("failed to register start/stop hotkey on fallback provider: %w", regErr)
				}

				// Register additional hotkeys on the fallback provider
				h.hotkeysMutex.Lock()
				for hotkey, action := range h.hotkeyActions {
					hk := hotkey
					a := action
					if regErr := h.provider.RegisterHotkey(hk, func() error {
						log.Printf("Custom hotkey detected: %s", hk)
						if err := a(); err != nil {
							log.Printf("Error executing hotkey action for %s: %v", hk, err)
							return err
						}
						return nil
					}); regErr != nil {
						h.hotkeysMutex.Unlock()
						return fmt.Errorf("failed to register hotkey %s on fallback provider: %w", hk, regErr)
					}
				}
				h.hotkeysMutex.Unlock()

				// Start fallback provider
				if startErr := h.provider.Start(); startErr != nil {
					return fmt.Errorf("failed to start fallback keyboard provider: %w", startErr)
				}

				log.Println("Fallback keyboard provider started successfully")
				return nil
			}
		}

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
