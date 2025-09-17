// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import "github.com/AshBuk/speak-to-ai/config"

// TrayManagerInterface defines the interface for tray managers
type TrayManagerInterface interface {
	Start()
	SetRecordingState(isRecording bool)
	SetTooltip(tooltip string)
	UpdateSettings(config *config.Config)
	// SetExitAction sets the callback invoked when Quit is clicked
	SetExitAction(onExit func())
	// SetCoreActions sets core menu callbacks (toggle, show config, show about, reset to defaults)
	SetCoreActions(onToggle func() error, onShowConfig func() error, onShowAbout func() error, onResetToDefaults func() error)
	// SetAudioActions sets callbacks for audio-related actions
	SetAudioActions(onSelectRecorder func(method string) error)
	// SetHotkeyRebindAction sets callback to rebind hotkeys by action name
	SetHotkeyRebindAction(onRebind func(action string) error)
	// SetSettingsActions sets callbacks for general settings from tray (VAD, Language, Model, Notifications)
	SetSettingsActions(
		// onSelectVADSensitivity func(sensitivity string) error,
		onSelectLanguage func(language string) error,
		onSelectModelType func(modelType string) error,
		onToggleWorkflowNotifications func() error,
		onSelectOutputMode func(mode string) error,
	)
	// SetGetOutputToolsCallback sets the callback for getting actual output tool names
	SetGetOutputToolsCallback(callback func() (clipboardTool, typeTool string))
	// SetCaptureOnceSupport sets a callback indicating whether capture once is supported
	SetCaptureOnceSupport(callback func() bool)
	Stop()
}
