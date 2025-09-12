// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"fmt"
	"os"
	"strings"

	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/providers"
)

// registerAllHotkeysOn registers recording and custom hotkeys on the given provider
func (h *HotkeyManager) registerAllHotkeysOn(provider interfaces.KeyboardEventProvider) error {
	// Register start/stop recording hotkey
	if err := provider.RegisterHotkey(h.config.GetStartRecordingHotkey(), func() error {
		h.hotkeysMutex.Lock()
		defer h.hotkeysMutex.Unlock()

		if !h.isRecording && h.recordingStarted != nil {
			h.logger.Info("Start recording hotkey detected")
			if err := h.recordingStarted(); err != nil {
				h.logger.Error("Error starting recording: %v", err)
				return err
			}
			h.isRecording = true
		} else if h.isRecording && h.recordingStopped != nil {
			h.logger.Info("Stop recording hotkey detected")
			if err := h.recordingStopped(); err != nil {
				h.logger.Error("Error stopping recording: %v", err)
				return err
			}
			h.isRecording = false
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to register start/stop recording hotkey: %w", err)
	}

	// Register additional hotkeys (action -> configured hotkey string)
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()
	for actionName, action := range h.hotkeyActions {
		hk := h.config.GetActionHotkey(actionName)
		if strings.TrimSpace(hk) == "" {
			// Skip actions without configured hotkey
			continue
		}

		act := action
		if err := provider.RegisterHotkey(hk, func() error {
			h.logger.Info("Custom hotkey detected: %s (%s)", actionName, hk)
			if err := act(); err != nil {
				h.logger.Error("Error executing hotkey action for %s: %v", actionName, err)
				return err
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to register hotkey %s for action %s: %w", hk, actionName, err)
		}
	}

	return nil
}

// startFallbackAfterRegistration attempts to switch to a fallback provider (evdev)
// Allows fallback on GNOME/KDE when running in AppImage due to portal sandboxing issues.
func startFallbackAfterRegistration(h *HotkeyManager, startErr error) error {
	h.logger.Error("Primary keyboard provider failed to start: %v", startErr)

	// Check if running in AppImage - allow fallback even on GNOME/KDE
	de := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	isAppImage := os.Getenv("APPIMAGE") != "" || os.Getenv("APPDIR") != ""

	if (strings.Contains(de, "gnome") || strings.Contains(de, "kde")) && !isAppImage {
		h.logger.Info("Skipping evdev fallback on GNOME/KDE; please check portal permissions")
		return fmt.Errorf("failed to start keyboard provider: %w", startErr)
	}

	if isAppImage {
		h.logger.Info("AppImage detected - allowing evdev fallback for better hotkey compatibility")
	}

	// Only fallback when current provider is DBus and evdev is supported
	switch h.provider.(type) {
	case *providers.DbusKeyboardProvider:
		fallback := providers.NewEvdevKeyboardProvider(h.config, h.environment, h.logger)
		if fallback != nil && fallback.IsSupported() {
			h.logger.Info("Falling back to evdev keyboard provider")
			// Swap provider
			h.provider = fallback

			// Re-register hotkeys on fallback provider
			if err := h.registerAllHotkeysOn(h.provider); err != nil {
				return fmt.Errorf("failed to register hotkeys on fallback provider: %w", err)
			}

			// Start fallback provider
			if err := h.provider.Start(); err != nil {
				return fmt.Errorf("failed to start fallback keyboard provider: %w", err)
			}

			h.logger.Info("Fallback keyboard provider started successfully")
			if isAppImage {
				h.logger.Info("AppImage hint: add user to 'input' group for evdev hotkeys:")
				h.logger.Info("sudo usermod -a -G input $USER && reboot")
			}
			return nil
		}
	}

	return fmt.Errorf("failed to start keyboard provider: %w", startErr)
}
