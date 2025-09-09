// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
)

// handleToggleStreaming toggles streaming transcription on/off
func (a *App) handleToggleStreaming() error {
	a.Logger.Info("Toggling streaming transcription...")

	// Toggle the setting
	a.Config.Audio.EnableStreaming = !a.Config.Audio.EnableStreaming

	status := "disabled"
	if a.Config.Audio.EnableStreaming {
		status = "enabled"
	}

	// Show notification
	if a.NotifyManager != nil {
		a.notify("Streaming Mode", fmt.Sprintf("Streaming transcription %s", status))
	}

	// Update tray tooltip
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip(fmt.Sprintf("Streaming: %s", status))
	}

	a.Logger.Info("Streaming transcription %s", status)
	return nil
}

// handleSwitchModel cycles through available models
func (a *App) handleSwitchModel() error {
	a.Logger.Info("Switching model...")

	if a.ModelManager == nil {
		return fmt.Errorf("model manager not available")
	}

	// Get available models
	availableModels := a.ModelManager.GetAvailableModels()
	if len(availableModels) <= 1 {
		if a.NotifyManager != nil {
			a.notify("Model Switch", "Only one model available")
		}
		return nil
	}

	// Get current model
	currentModel := a.ModelManager.GetActiveModel()

	// Find next model in the list
	var modelNames []string
	for name := range availableModels {
		modelNames = append(modelNames, name)
	}

	// Sort for consistent ordering
	// Simple alphabetical sort
	for i := 0; i < len(modelNames)-1; i++ {
		for j := i + 1; j < len(modelNames); j++ {
			if modelNames[i] > modelNames[j] {
				modelNames[i], modelNames[j] = modelNames[j], modelNames[i]
			}
		}
	}

	// Find current index and switch to next
	currentIndex := -1
	for i, name := range modelNames {
		if name == currentModel {
			currentIndex = i
			break
		}
	}

	// Switch to next model (cycle around)
	nextIndex := (currentIndex + 1) % len(modelNames)
	nextModel := modelNames[nextIndex]

	// Switch the model
	if err := a.ModelManager.SwitchModel(nextModel); err != nil {
		a.Logger.Error("Failed to switch model: %v", err)
		if a.NotifyManager != nil {
			a.notify("Error", fmt.Sprintf("Failed to switch model: %v", err))
		}
		return err
	}

	// Get model info for notification
	modelInfo := availableModels[nextModel]

	// Show notification
	if a.NotifyManager != nil {
		a.notify("Model Switched", fmt.Sprintf("Now using: %s", modelInfo.Description))
	}

	// Update tray tooltip
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip(fmt.Sprintf("Model: %s", modelInfo.Type))
	}

	a.Logger.Info("Switched to model: %s", nextModel)
	return nil
}

// handleToggleVAD toggles Voice Activity Detection on/off
func (a *App) handleToggleVAD() error {
	a.Logger.Info("Toggling VAD...")

	// Toggle the setting
	a.Config.Audio.EnableVAD = !a.Config.Audio.EnableVAD

	status := "disabled"
	if a.Config.Audio.EnableVAD {
		status = "enabled"
	}

	// Show notification
	if a.NotifyManager != nil {
		a.notify("Voice Activity Detection", fmt.Sprintf("VAD %s", status))
	}

	// Update tray tooltip
	if a.TrayManager != nil {
		a.TrayManager.SetTooltip(fmt.Sprintf("VAD: %s", status))
	}

	a.Logger.Info("VAD %s", status)
	return nil
}
