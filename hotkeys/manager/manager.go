// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/providers"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// HotkeyAction represents a hotkey action callback
type HotkeyAction func() error

// Manages keyboard shortcuts, providers, and actions
type HotkeyManager struct {
	config           adapters.HotkeyConfig
	isListening      bool
	isRecording      bool
	stopListening    chan bool
	recordingStarted func() error
	recordingStopped func() error
	hotkeyActions    map[string]HotkeyAction // Maps hotkey actions to their callbacks
	hotkeysMutex     sync.Mutex
	environment      interfaces.EnvironmentType
	provider         interfaces.KeyboardEventProvider
	modifierState    map[string]bool // Tracks the state of modifier keys
	logger           logger.Logger
}

// Create a new instance of the HotkeyManager
func NewHotkeyManager(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) *HotkeyManager {
	manager := &HotkeyManager{
		config:        config,
		isListening:   false,
		isRecording:   false,
		stopListening: make(chan bool),
		environment:   environment,
		modifierState: make(map[string]bool),
		hotkeyActions: make(map[string]HotkeyAction),
		logger:        logger,
	}

	// Initialize the appropriate keyboard provider
	manager.provider = selectProviderForEnvironment(manager.config, manager.environment, manager.logger)

	return manager
}

// selectProviderForEnvironment is defined in OS-specific files (e.g., manager_linux.go)

// Register callbacks for recording start and stop events
func (h *HotkeyManager) RegisterCallbacks(
	recordingStarted func() error,
	recordingStopped func() error,
) {
	h.recordingStarted = recordingStarted
	h.recordingStopped = recordingStopped
}

// Register a custom hotkey action
func (h *HotkeyManager) RegisterHotkeyAction(hotkey string, action HotkeyAction) {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	h.hotkeyActions[hotkey] = action
}

// Unregister a custom hotkey action
func (h *HotkeyManager) UnregisterHotkeyAction(hotkey string) {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	delete(h.hotkeyActions, hotkey)
}

// Return all registered hotkeys
func (h *HotkeyManager) GetRegisteredHotkeys() []string {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()

	var hotkeys []string
	// Add the primary recording hotkeys
	hotkeys = append(hotkeys, h.config.GetStartRecordingHotkey())

	// Add any custom hotkeys
	for hotkey := range h.hotkeyActions {
		hotkeys = append(hotkeys, hotkey)
	}

	return hotkeys
}

// Start listening for hotkeys
func (h *HotkeyManager) Start() error {
	if h.isListening {
		return fmt.Errorf("hotkey manager is already running")
	}

	if h.provider == nil {
		return fmt.Errorf("no keyboard provider available - hotkeys will not work")
	}

	h.isListening = true

	h.logger.Info("Starting hotkey manager...")
	h.logger.Info("- Start/Stop recording: %s", h.config.GetStartRecordingHotkey())

	// Register all hotkeys on the selected provider
	if err := h.registerAllHotkeysOn(h.provider); err != nil {
		return err
	}

	// Start the provider and handle potential fallbacks
	err := h.provider.Start()
	if err != nil {
		h.isListening = false
		return startFallbackAfterRegistration(h, err)
	}

	return nil
}

// Stop the hotkey listener
func (h *HotkeyManager) Stop() {
	if h.isListening {
		h.provider.Stop()
		h.isListening = false
	}
}

// Return the current recording state
func (h *HotkeyManager) IsRecording() bool {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	return h.isRecording
}

// Forcefully set the recording state to false
func (h *HotkeyManager) ResetRecordingState() {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	h.isRecording = false
}

// Simulate a hotkey press for testing purposes
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

// Reload the configuration by stopping, updating the provider, and restarting
func (h *HotkeyManager) ReloadConfig(newConfig adapters.HotkeyConfig) error {
	if h.isListening && h.provider != nil {
		h.provider.Stop()
		h.isListening = false
	}

	h.config = newConfig
	h.provider = selectProviderForEnvironment(h.config, h.environment, h.logger)
	if h.provider == nil {
		return fmt.Errorf("no keyboard provider available - hotkeys will not work")
	}

	// Re-register all hotkeys on the new provider
	if err := h.registerAllHotkeysOn(h.provider); err != nil {
		return err
	}

	// Start the new provider
	if err := h.provider.Start(); err != nil {
		return startFallbackAfterRegistration(h, err)
	}

	h.isListening = true
	return nil
}

// Attempt to capture a single hotkey combination
// Fall back to a temporary evdev provider if the active one does not support capture
func (h *HotkeyManager) CaptureOnce(timeout time.Duration) (string, error) {
	if h.provider == nil {
		return "", fmt.Errorf("no keyboard provider available")
	}
	combo, err := h.provider.CaptureOnce(timeout)
	if err == nil && combo != "" {
		return combo, nil
	}
	// If the D-Bus provider fails, attempt a fallback to evdev for the capture
	if _, isDbus := h.provider.(*providers.DbusKeyboardProvider); isDbus {
		fallback := providers.NewEvdevKeyboardProvider(h.config, h.environment, h.logger)
		if fallback != nil && fallback.IsSupported() {
			return fallback.CaptureOnce(timeout)
		}
	}
	return "", err
}

// Check if the active provider supports the capture-once functionality
func (h *HotkeyManager) SupportsCaptureOnce() bool {
	if h.provider == nil {
		return false
	}
	// evdev supports capture-once, dbus does not
	if _, ok := h.provider.(*providers.EvdevKeyboardProvider); ok {
		return true
	}
	return false
}
