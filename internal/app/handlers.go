// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
)

// Defines hotkey handler methods that delegate to the appropriate services

// Handle the start recording hotkey
func (a *App) handleStartRecording() error {
	if a.Services == nil || a.Services.Audio == nil {
		return fmt.Errorf("audio service not available")
	}
	return a.Services.Audio.HandleStartRecording()
}

// Handle the stop recording hotkey
func (a *App) handleStopRecordingAndTranscribe() error {
	if a.Services == nil || a.Services.Audio == nil {
		return fmt.Errorf("audio service not available")
	}
	return a.Services.Audio.HandleStopRecording()
}

// handleToggleVAD handles VAD toggle hotkey
// TODO: Next feature - VAD implementation
// func (a *App) handleToggleVAD() error {
//	if a.Services == nil || a.Services.Config == nil {
//		return fmt.Errorf("config service not available")
//	}
//	return a.Services.Config.ToggleVAD()
// }

// Handle the show config hotkey
func (a *App) handleShowConfig() error {
	if a.Services == nil || a.Services.UI == nil {
		return fmt.Errorf("UI service not available")
	}
	return a.Services.UI.ShowConfigFile()
}

// Handle the reset to defaults hotkey
func (a *App) handleResetToDefaults() error {
	if a.Services == nil || a.Services.Config == nil {
		return fmt.Errorf("config service not available")
	}
	if err := a.Services.Config.ResetToDefaults(); err != nil {
		return err
	}
	// Reload hotkeys to apply the new default bindings immediately
	if a.Services.Hotkeys != nil {
		cfgProvider := func() adapters.HotkeyConfig {
			if cfg := a.Services.Config.GetConfig(); cfg != nil {
				return adapters.NewConfigAdapter(cfg.Hotkeys.StartRecording, cfg.Hotkeys.Provider).
					WithAdditionalHotkeys(
						"",
						cfg.Hotkeys.ShowConfig,
						cfg.Hotkeys.ResetToDefaults,
					)
			}
			return adapters.NewConfigAdapter("", "auto")
		}
		if err := a.Services.Hotkeys.ReloadFromConfig(
			a.handleStartRecording,
			a.handleStopRecordingAndTranscribe,
			cfgProvider,
		); err != nil {
			return err
		}
	}
	return nil
}
