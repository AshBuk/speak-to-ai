// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/manager"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// HotkeyManagerInterface defines the interface for hotkey managers
type HotkeyManagerInterface interface {
	Start() error
	Stop()
	RegisterCallbacks(startRecording, stopRecording func() error)
	RegisterHotkeyAction(action string, callback manager.HotkeyAction)
	ReloadConfig(newConfig adapters.HotkeyConfig) error
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
	// toggleVAD is the callback for toggling VAD
	// toggleVAD func() error,
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
	// TODO: Next feature - VAD implementation
	// hs.hotkeyManager.RegisterHotkeyAction("toggle_vad", toggleVAD)
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

// ReloadFromConfig rebuilds hotkey configuration in the underlying manager
func (hs *HotkeyService) ReloadFromConfig(startRecording, stopRecording func() error, configProvider func() adapters.HotkeyConfig) error {
	if hs.hotkeyManager == nil {
		return fmt.Errorf("hotkey manager not available")
	}
	cfg := configProvider()
	// ensure callbacks are set (in case of provider swap)
	hs.hotkeyManager.RegisterCallbacks(startRecording, stopRecording)
	return hs.hotkeyManager.ReloadConfig(cfg)
}

// CaptureOnce proxies one-shot capture to the underlying hotkey manager
func (hs *HotkeyService) CaptureOnce(timeoutMs int) (string, error) {
	if hs.hotkeyManager == nil {
		return "", fmt.Errorf("hotkey manager not available")
	}
	if timeoutMs <= 0 {
		timeoutMs = 3000
	}
	if hm, ok := hs.hotkeyManager.(*manager.HotkeyManager); ok {
		return hm.CaptureOnce(time.Duration(timeoutMs) * time.Millisecond)
	}
	return "", fmt.Errorf("capture not supported by manager")
}

// SupportsCaptureOnce reports whether the underlying manager supports capture once
func (hs *HotkeyService) SupportsCaptureOnce() bool {
	if hs.hotkeyManager == nil {
		return false
	}
	if hm, ok := hs.hotkeyManager.(*manager.HotkeyManager); ok {
		return hm.SupportsCaptureOnce()
	}
	return false
}
