// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"
	"log"
	"sync"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
)

// HotkeyAction represents a hotkey action callback
type HotkeyAction func() error

// HotkeyManager handles keyboard shortcuts
type HotkeyManager struct {
	config           adapters.HotkeyConfig
	isListening      bool
	isRecording      bool
	stopListening    chan bool
	recordingStarted func() error
	recordingStopped func() error
	hotkeyActions    map[string]HotkeyAction // Additional hotkey actions
	hotkeysMutex     sync.Mutex
	environment      interfaces.EnvironmentType
	provider         interfaces.KeyboardEventProvider
	modifierState    map[string]bool // Track state of modifier keys
}

// NewHotkeyManager creates a new instance of HotkeyManager
func NewHotkeyManager(config adapters.HotkeyConfig, environment interfaces.EnvironmentType) *HotkeyManager {
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

	// Register all hotkeys using helper
	if err := h.registerAllHotkeysOn(h.provider); err != nil {
		return err
	}

	// Start the provider
	err := h.provider.Start()
	if err != nil {
		// Delegate fallback logic to helper
		h.isListening = false
		return startFallbackAfterRegistration(h, err)
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

// ResetRecordingState forcefully sets recording state to false
func (h *HotkeyManager) ResetRecordingState() {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	h.isRecording = false
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
		if h.recordingStopped != nil {
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
