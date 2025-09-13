// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
)

// Hotkey handler methods that delegate to services

// handleStartRecording handles the start recording hotkey
func (a *App) handleStartRecording() error {
	if a.Services == nil || a.Services.Audio == nil {
		return fmt.Errorf("audio service not available")
	}
	return a.Services.Audio.HandleStartRecording()
}

// handleStopRecordingAndTranscribe handles the stop recording hotkey
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

// handleSwitchModel handles model switching hotkey
func (a *App) handleSwitchModel() error {
	if a.Services == nil || a.Services.Config == nil || a.Services.Audio == nil {
		return fmt.Errorf("services not available")
	}

	cfg := a.Services.Config.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not available")
	}

	order := []string{"tiny", "base", "small", "medium", "large"}
	current := cfg.General.ModelType
	idx := 0
	for i, m := range order {
		if m == current {
			idx = i
			break
		}
	}
	next := order[(idx+1)%len(order)]

	// Persist in config first
	if err := a.Services.Config.UpdateModelType(next); err != nil {
		return err
	}
	// Switch model in audio service (ensures model availability/init)
	return a.Services.Audio.SwitchModel(next)
}

// handleShowConfig handles show config hotkey
func (a *App) handleShowConfig() error {
	if a.Services == nil || a.Services.UI == nil {
		return fmt.Errorf("UI service not available")
	}
	return a.Services.UI.ShowConfigFile()
}

// handleResetToDefaults handles reset to defaults hotkey
func (a *App) handleResetToDefaults() error {
	if a.Services == nil || a.Services.Config == nil {
		return fmt.Errorf("config service not available")
	}
	return a.Services.Config.ResetToDefaults()
}
