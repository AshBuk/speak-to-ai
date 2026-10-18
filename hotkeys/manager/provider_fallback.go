// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"
	"strings"

	"github.com/AshBuk/dabri/v2/hotkeys/interfaces"
	"github.com/AshBuk/dabri/v2/hotkeys/providers"
)

// actionToggleRecording is the stable id for the start/stop recording action.
const actionToggleRecording = "toggle_recording"

// Register all configured hotkeys on the given provider
func (h *HotkeyManager) registerAllHotkeysOn(provider interfaces.KeyboardEventProvider) error {
	// Register the primary start/stop recording hotkey
	if err := provider.RegisterHotkey(actionToggleRecording, h.config.GetStartRecordingHotkey(), func(pressed bool) error {
		// The callback owns the start/stop decision against the recorder's real
		// state, so this path holds no recording state of its own.
		if h.recording == nil {
			return nil
		}
		h.logger.Info("Recording hotkey detected: pressed=%t", pressed)
		if err := h.recording(pressed); err != nil {
			h.logger.Error("Error handling recording hotkey: %v", err)
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to register start/stop recording hotkey: %w", err)
	}
	// Register any additional custom hotkeys
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	for actionName, action := range h.hotkeyActions {
		hk := h.config.GetActionHotkey(actionName)
		if strings.TrimSpace(hk) == "" {
			// Skip actions that do not have a configured hotkey
			continue
		}

		act := action
		if err := provider.RegisterHotkey(actionName, hk, func(pressed bool) error {
			if !pressed {
				return nil
			}
			h.logger.Info("Custom hotkey detected: %s (%s)", actionName, hk)
			// Off the provider's listener: an action may reload hotkeys, which
			// stops the provider and waits for that listener to exit.
			go func() {
				if err := act(); err != nil {
					h.logger.Error("Error executing hotkey action for %s: %v", actionName, err)
				}
			}()
			return nil
		}); err != nil {
			return fmt.Errorf("failed to register hotkey %s for action %s: %w", hk, actionName, err)
		}
	}

	return nil
}

// Attempt to switch to the other provider (D-Bus <-> evdev) if the primary one fails
func startFallbackAfterRegistration(h *HotkeyManager, startErr error) error {
	h.logger.Error("Primary keyboard provider failed to start: %v", startErr)
	var fallback interfaces.KeyboardEventProvider
	var name string
	switch h.provider.(type) {
	case *providers.EvdevKeyboardProvider:
		fallback, name = providers.NewDbusKeyboardProvider(h.logger), "D-Bus"
	case *providers.DbusKeyboardProvider:
		fallback, name = providers.NewEvdevKeyboardProvider(h.logger), "evdev"
	}
	if fallback == nil || !fallback.IsSupported() {
		return fmt.Errorf("failed to start keyboard provider: %w", startErr)
	}
	h.logger.Info("Falling back to %s keyboard provider", name)
	h.provider = fallback
	if err := h.registerAllHotkeysOn(h.provider); err != nil {
		return fmt.Errorf("failed to register hotkeys on fallback provider: %w", err)
	}
	if err := h.provider.Start(); err != nil {
		return fmt.Errorf("failed to start fallback keyboard provider: %w", err)
	}
	h.isListening = true
	h.logger.Info("Fallback keyboard provider started successfully")
	return nil
}
