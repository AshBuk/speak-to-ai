// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/hotkeys/manager"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// HotkeyManagerInterface defines the interface for hotkey managers
type HotkeyManagerInterface interface {
	Start() error
	Stop()
	RegisterCallbacks(startRecording, stopRecording func() error)
	RegisterHotkeyAction(action string, callback manager.HotkeyAction)
}

// HotkeyService implements HotkeyServiceInterface
type HotkeyService struct {
	logger        logger.Logger
	hotkeyManager HotkeyManagerInterface
}

// NewHotkeyService creates a new HotkeyService instance
func NewHotkeyService(
	logger logger.Logger,
	hotkeyManager HotkeyManagerInterface,
) *HotkeyService {
	return &HotkeyService{
		logger:        logger,
		hotkeyManager: hotkeyManager,
	}
}

// SetupHotkeyCallbacks configures hotkey callbacks with handler functions
func (hs *HotkeyService) SetupHotkeyCallbacks(
	startRecording func() error,
	stopRecording func() error,
	toggleStreaming func() error,
	toggleVAD func() error,
	switchModel func() error,
	showConfig func() error,
	resetToDefaults func() error,
) error {
	if hs.hotkeyManager == nil {
		return fmt.Errorf("hotkey manager not available")
	}

	hs.logger.Info("Setting up hotkey callbacks...")

	// Register the main recording callbacks
	hs.hotkeyManager.RegisterCallbacks(startRecording, stopRecording)

	// Register additional hotkey actions
	hs.hotkeyManager.RegisterHotkeyAction("toggle_streaming", toggleStreaming)
	hs.hotkeyManager.RegisterHotkeyAction("toggle_vad", toggleVAD)
	hs.hotkeyManager.RegisterHotkeyAction("switch_model", switchModel)
	hs.hotkeyManager.RegisterHotkeyAction("show_config", showConfig)
	hs.hotkeyManager.RegisterHotkeyAction("reset_to_defaults", resetToDefaults)

	hs.logger.Info("Hotkey callbacks configured successfully")
	return nil
}

// RegisterHotkeys implements HotkeyServiceInterface
func (hs *HotkeyService) RegisterHotkeys() error {
	if hs.hotkeyManager == nil {
		return fmt.Errorf("hotkey manager not available")
	}

	hs.logger.Info("Registering hotkeys...")

	return hs.hotkeyManager.Start()
}

// UnregisterHotkeys implements HotkeyServiceInterface
func (hs *HotkeyService) UnregisterHotkeys() error {
	if hs.hotkeyManager == nil {
		return nil
	}

	hs.logger.Info("Unregistering hotkeys...")

	hs.hotkeyManager.Stop()
	return nil
}

// Shutdown implements HotkeyServiceInterface
func (hs *HotkeyService) Shutdown() error {
	// Unregister hotkeys
	if err := hs.UnregisterHotkeys(); err != nil {
		hs.logger.Error("Error unregistering hotkeys: %v", err)
		return err
	}

	hs.logger.Info("HotkeyService shutdown complete")
	return nil
}
