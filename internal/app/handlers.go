// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import "fmt"

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

// handleToggleStreaming handles streaming toggle hotkey
func (a *App) handleToggleStreaming() error {
	if a.Services == nil || a.Services.Config == nil {
		return fmt.Errorf("config service not available")
	}
	return a.Services.Config.ToggleStreaming()
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
	if a.Services == nil || a.Services.Audio == nil {
		return fmt.Errorf("audio service not available")
	}

	// Cycle through available models (simplified implementation)
	currentModel := "base" // Default
	nextModel := "small"
	if currentModel == "small" {
		nextModel = "base"
	}

	return a.Services.Audio.SwitchModel(nextModel)
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
