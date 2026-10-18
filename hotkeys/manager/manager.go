// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/AshBuk/dabri/v2/hotkeys/adapters"
	"github.com/AshBuk/dabri/v2/hotkeys/interfaces"
	"github.com/AshBuk/dabri/v2/hotkeys/providers"
	"github.com/AshBuk/dabri/v2/internal/logger"
)

// HotkeyAction represents a hotkey action callback
type HotkeyAction func() error

// HotkeyManager manages keyboard shortcuts, providers, and actions
type HotkeyManager struct {
	config        adapters.HotkeyConfig
	isListening   bool
	recording     func(pressed bool) error
	hotkeyActions map[string]HotkeyAction // Maps hotkey actions to their callbacks
	hotkeysMutex  sync.Mutex
	environment   interfaces.EnvironmentType
	provider      interfaces.KeyboardEventProvider
	logger        logger.Logger
}

// NewHotkeyManager creates a new instance of the HotkeyManager
func NewHotkeyManager(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) *HotkeyManager {
	manager := &HotkeyManager{
		config:        config,
		isListening:   false,
		environment:   environment,
		hotkeyActions: make(map[string]HotkeyAction),
		logger:        logger,
	}
	// Initialize the appropriate keyboard provider
	manager.provider = selectProviderForEnvironment(manager.config, manager.environment, manager.logger)

	return manager
}

// selectProviderForEnvironment is defined in OS-specific files (e.g., manager_linux.go)

// RegisterRecording registers the callback for the recording hotkey's press
// (pressed=true) and release (pressed=false). The callback owns the start/stop
// decision (it consults the recorder's real state), so the hotkey keeps no
// recording state of its own and never desyncs from the window or tray.
func (h *HotkeyManager) RegisterRecording(recording func(pressed bool) error) {
	h.recording = recording
}

// RegisterHotkeyAction registers a custom hotkey action
func (h *HotkeyManager) RegisterHotkeyAction(hotkey string, action HotkeyAction) {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	h.hotkeyActions[hotkey] = action
}

// UnregisterHotkeyAction unregisters a custom hotkey action
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
	h.logger.Info("Starting hotkey manager...")
	h.logger.Info("- Start/Stop recording: %s", h.config.GetStartRecordingHotkey())
	// Register all hotkeys on the selected provider
	if err := h.registerAllHotkeysOn(h.provider); err != nil {
		return err
	}
	// Start the provider and handle potential fallbacks
	if err := h.provider.Start(); err != nil {
		return startFallbackAfterRegistration(h, err)
	}
	h.isListening = true
	return nil
}

// Stop the hotkey listener
func (h *HotkeyManager) Stop() {
	if h.isListening {
		h.provider.Stop()
		h.isListening = false
	}
}

// SimulateHotkeyPress simulates a hotkey press for testing purposes
func (h *HotkeyManager) SimulateHotkeyPress(hotkeyName string) error {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()

	switch hotkeyName {
	case "toggle_recording":
		if h.recording != nil {
			return h.recording(true)
		}
	default:
		return fmt.Errorf("unknown hotkey: %s", hotkeyName)
	}
	return nil
}

// ReloadConfig reloads the configuration by stopping, updating the provider, and restarting
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
	if err := h.provider.Start(); err != nil {
		return startFallbackAfterRegistration(h, err)
	}
	h.isListening = true
	return nil
}

// CaptureOnce captures a single hotkey combination. Only evdev reads keys
// directly; with the portal the desktop owns the bindings. The listener stops
// for the capture, so the pressed keys trigger no action.
func (h *HotkeyManager) CaptureOnce(timeout time.Duration) (string, error) {
	evdev, ok := h.provider.(*providers.EvdevKeyboardProvider)
	if !ok {
		return "", fmt.Errorf("capture not supported")
	}
	evdev.Stop()
	result, err := evdev.CaptureOnce(timeout)
	if h.isListening {
		if startErr := evdev.Start(); startErr != nil {
			h.logger.Error("Failed to restart hotkeys after capture: %v", startErr)
		}
	}
	return result, err
}

// PushToTalk reports whether the recording hotkey records while held: the
// config asks for it and the active provider reports key release (evdev only).
func (h *HotkeyManager) PushToTalk() bool {
	return h.SupportsCaptureOnce() && h.config.GetMode() == "ptt"
}

// SupportsCaptureOnce reports whether the active provider reads keys directly
// (evdev), which both hotkey capture and push-to-talk need.
func (h *HotkeyManager) SupportsCaptureOnce() bool {
	_, ok := h.provider.(*providers.EvdevKeyboardProvider)
	return ok
}
