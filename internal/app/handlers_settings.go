// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// handleSelectVADSensitivity updates VAD sensitivity and persists config
func (a *App) handleSelectVADSensitivity(sensitivity string) error {
	s := strings.ToLower(sensitivity)
	switch s {
	case "low", "medium", "high":
	default:
		return fmt.Errorf("invalid VAD sensitivity: %s", sensitivity)
	}

	if a.Config.Audio.VADSensitivity == s {
		return nil
	}

	old := a.Config.Audio.VADSensitivity
	a.Config.Audio.VADSensitivity = s

	// Persist
	if a.ConfigFile != "" {
		if err := config.SaveConfig(a.ConfigFile, a.Config); err != nil {
			// rollback on error
			a.Config.Audio.VADSensitivity = old
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	// Update tray UI
	if a.TrayManager != nil {
		go a.TrayManager.UpdateSettings(a.Config)
	}

	// Notify
	if a.NotifyManager != nil {
		title := strings.ToUpper(s[:1]) + s[1:] // Manual title case
		a.notify("VAD Sensitivity", fmt.Sprintf("Set to %s", title))
	}
	a.Logger.Info("VAD sensitivity changed to %s", s)
	return nil
}

// handleSelectLanguage updates recognition language and persists config
func (a *App) handleSelectLanguage(language string) error {
	lang := strings.ToLower(language)
	switch lang {
	case "auto", "en", "de", "fr", "es", "he", "ru":
	default:
		return fmt.Errorf("invalid language: %s", language)
	}

	if a.Config.General.Language == lang {
		return nil
	}

	old := a.Config.General.Language
	a.Config.General.Language = lang

	// Persist
	if a.ConfigFile != "" {
		if err := config.SaveConfig(a.ConfigFile, a.Config); err != nil {
			// rollback on error
			a.Config.General.Language = old
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	// Update tray UI
	if a.TrayManager != nil {
		go a.TrayManager.UpdateSettings(a.Config)
	}

	// Notify
	if a.NotifyManager != nil {
		name := map[string]string{
			"auto": "Auto",
			"en":   "English",
			"de":   "German",
			"fr":   "French",
			"es":   "Spanish",
			"he":   "Hebrew",
			"ru":   "Russian",
		}[lang]
		if name == "" {
			name = strings.ToUpper(lang)
		}
		a.notify("Language", fmt.Sprintf("Set to %s", name))
	}
	a.Logger.Info("Language changed to %s", lang)
	return nil
}

// handleSelectModelType switches Whisper model type, auto-downloading if needed
func (a *App) handleSelectModelType(modelType string) error {
	t := strings.ToLower(modelType)
	switch t {
	case "tiny", "base", "small", "medium", "large":
	default:
		return fmt.Errorf("invalid model type: %s", modelType)
	}

	if a.Config.General.ModelType == t {
		return nil
	}

	// Update config first (config-first approach)
	oldType := a.Config.General.ModelType
	a.Config.General.ModelType = t

	// Notify starting
	if a.NotifyManager != nil {
		a.notify("Speak-to-AI", fmt.Sprintf("Switching model to %s...", t))
	}

	// Progress callback updates tray tooltip
	progress := func(downloaded, total int64, percentage float64) {
		a.setUIProcessing(fmt.Sprintf("Downloading %s: %.1f%%", t, percentage))
	}

	// Ensure model path (download if missing)
	modelPath, err := a.ModelManager.GetModelPathWithProgress(progress)
	if err != nil {
		// rollback model type
		a.Config.General.ModelType = oldType
		a.setUIError("Model switch failed")
		return fmt.Errorf("failed to prepare model: %w", err)
	}

	// Validate model file
	if err := a.ModelManager.ValidateModel(modelPath); err != nil {
		a.Config.General.ModelType = oldType
		return fmt.Errorf("invalid model file: %w", err)
	}

	// Reinitialize whisper engines with the new model
	if a.WhisperEngine != nil {
		_ = a.WhisperEngine.Close()
	}
	engine, err := whisper.NewWhisperEngine(a.Config, modelPath)
	if err != nil {
		a.Config.General.ModelType = oldType
		return fmt.Errorf("failed to init whisper engine: %w", err)
	}
	a.WhisperEngine = engine

	// Streaming engine
	if a.Config.Audio.EnableStreaming {
		streaming, sErr := whisper.NewStreamingWhisperEngine(a.Config, modelPath)
		if sErr != nil {
			a.Logger.Warning("Failed to init streaming engine for new model: %v", sErr)
			a.StreamingEngine = nil
		} else {
			a.StreamingEngine = streaming
		}
	}

	// Persist: remember active model path as well
	oldActive := a.Config.General.ActiveModel
	a.Config.General.ActiveModel = modelPath
	if a.ConfigFile != "" {
		if err := config.SaveConfig(a.ConfigFile, a.Config); err != nil {
			// rollback on failure
			a.Config.General.ActiveModel = oldActive
			a.Config.General.ModelType = oldType
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	// Update tray UI
	a.setUIReady()
	if a.TrayManager != nil {
		go a.TrayManager.UpdateSettings(a.Config)
	}

	if a.NotifyManager != nil {
		title := strings.ToUpper(t[:1]) + t[1:] // Manual title case
		a.notify("Speak-to-AI", fmt.Sprintf("Model switched to %s", title))
	}
	a.Logger.Info("Model switched to type=%s, path=%s", t, modelPath)
	return nil
}

// handleToggleWorkflowNotifications toggles workflow notifications and persists config
func (a *App) handleToggleWorkflowNotifications() error {
	current := a.Config.Notifications.EnableWorkflowNotifications
	a.Config.Notifications.EnableWorkflowNotifications = !current

	// Persist
	if a.ConfigFile != "" {
		if err := config.SaveConfig(a.ConfigFile, a.Config); err != nil {
			// rollback on error
			a.Config.Notifications.EnableWorkflowNotifications = current
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	// Update tray UI
	if a.TrayManager != nil {
		go a.TrayManager.UpdateSettings(a.Config)
	}

	// Notify (always show this system notification)
	if a.NotifyManager != nil {
		status := "enabled"
		if !a.Config.Notifications.EnableWorkflowNotifications {
			status = "disabled"
		}
		a.notify("Workflow Notifications", fmt.Sprintf("Workflow notifications %s", status))
	}
	a.Logger.Info("Workflow notifications toggled to %t", a.Config.Notifications.EnableWorkflowNotifications)
	return nil
}
