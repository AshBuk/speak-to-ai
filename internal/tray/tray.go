//go:build systray

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"fmt"
	"log"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/constants"
	"github.com/getlantern/systray"
)

// TrayManager manages the system tray icon and menu
type TrayManager struct {
	isRecording       bool
	iconMicOff        []byte
	iconMicOn         []byte
	onExit            func()
	onToggle          func() error
	onShowConfig      func() error
	onResetToDefaults func() error
	config            *config.Config

	// Menu items
	toggleItem       *systray.MenuItem
	settingsItem     *systray.MenuItem
	showConfigItem   *systray.MenuItem
	reloadConfigItem *systray.MenuItem
	exitItem         *systray.MenuItem

	// Settings submenus
	hotkeysMenu       *systray.MenuItem
	audioRecorderMenu *systray.MenuItem
	aiModelMenu       *systray.MenuItem
	outputMenu        *systray.MenuItem

	// Dynamic settings items
	hotkeyItems map[string]*systray.MenuItem
	audioItems  map[string]*systray.MenuItem
	modelItems  map[string]*systray.MenuItem
	outputItems map[string]*systray.MenuItem

	// Audio action callbacks
	onSelectRecorder func(method string) error
	onTestRecording  func() error
	// Settings callbacks
	onSelectVADSens        func(sensitivity string) error
	onSelectLang           func(language string) error
	onSelectModel          func(modelType string) error
	onToggleWorkflowNotify func() error
}

// NewTrayManager creates a new tray manager instance
func NewTrayManager(iconMicOff, iconMicOn []byte, onExit func(), onToggle func() error, onShowConfig func() error, onResetToDefaults func() error) *TrayManager {
	return &TrayManager{
		isRecording:       false,
		iconMicOff:        iconMicOff,
		iconMicOn:         iconMicOn,
		onExit:            onExit,
		onToggle:          onToggle,
		onShowConfig:      onShowConfig,
		onResetToDefaults: onResetToDefaults,
		hotkeyItems:       make(map[string]*systray.MenuItem),
		audioItems:        make(map[string]*systray.MenuItem),
		modelItems:        make(map[string]*systray.MenuItem),
		outputItems:       make(map[string]*systray.MenuItem),
	}
}

// SetCoreActions allows wiring core menu callbacks after construction
func (tm *TrayManager) SetCoreActions(onToggle func() error, onShowConfig func() error, onResetToDefaults func() error) {
	tm.onToggle = onToggle
	tm.onShowConfig = onShowConfig
	tm.onResetToDefaults = onResetToDefaults
}

// Start initializes and starts the system tray icon and menu
func (tm *TrayManager) Start() {
	go systray.Run(tm.onReady, tm.onExit)
}

// onReady sets up the system tray when it's ready
func (tm *TrayManager) onReady() {
	systray.SetIcon(tm.iconMicOff)
	systray.SetTitle("Speak-to-AI")
	systray.SetTooltip("Speak-to-AI: Offline Speech-to-Text")

	// Create main menu items
	tm.toggleItem = systray.AddMenuItem(fmt.Sprintf("%s Start Recording", constants.IconRecording), "Start/Stop recording")

	// Workflow notifications toggle
	tm.audioItems["workflow_notifications"] = systray.AddMenuItem(
		"○ Workflow Notifications",
		"Toggle workflow notifications (recording, transcription)",
	)

	systray.AddSeparator()

	// Settings submenu
	tm.settingsItem = systray.AddMenuItem(fmt.Sprintf("%s  Settings", constants.TraySettings), "Application settings")
	tm.createSettingsSubmenus()

	systray.AddSeparator()

	// Config actions
	tm.showConfigItem = systray.AddMenuItem(fmt.Sprintf("%s Show Config File", constants.TrayShowConfig), "Open configuration file")
	tm.reloadConfigItem = systray.AddMenuItem(fmt.Sprintf("%s Reset to Defaults", constants.IconConfig), "Reset all settings to default values")

	systray.AddSeparator()

	tm.exitItem = systray.AddMenuItem(fmt.Sprintf("%s Quit", constants.IconError), "Quit Speak-to-AI")

	// Handle menu item clicks
	go tm.handleMenuClicks()

	// Apply the current recording state once menu items are ready
	// This avoids missing early state updates before systray initialization
	tm.SetRecordingState(tm.isRecording)
}

// createSettingsSubmenus is implemented in settings_menu.go

// UpdateSettings updates the settings display with new configuration
func (tm *TrayManager) UpdateSettings(config *config.Config) {
	// Store and delegate all UI updates to helpers in settings_menu.go
	tm.config = config
	// The helpers gracefully no-op if items are not yet created
	tm.updateRecorderRadioUI(config.Audio.RecordingMethod)
	tm.updateVADRadioUI(config.Audio.VADSensitivity)
	tm.updateLanguageRadioUI(config.General.Language)
	tm.updateModelRadioUI(config.General.ModelType)
	tm.updateWorkflowNotificationUI(config.Notifications.EnableWorkflowNotifications)
}

// handleMenuClicks handles all menu item clicks
func (tm *TrayManager) handleMenuClicks() {
	for {
		select {
		case <-tm.toggleItem.ClickedCh:
			log.Println("Toggle recording clicked")
			if err := tm.onToggle(); err != nil {
				log.Printf("Error toggling recording: %v", err)
			}
		case <-tm.showConfigItem.ClickedCh:
			log.Println("Show config clicked")
			if tm.onShowConfig != nil {
				if err := tm.onShowConfig(); err != nil {
					log.Printf("Error showing config: %v", err)
				}
			}
		case <-tm.reloadConfigItem.ClickedCh:
			log.Println("Reset to defaults clicked")
			if tm.onResetToDefaults != nil {
				if err := tm.onResetToDefaults(); err != nil {
					log.Printf("Error resetting to defaults: %v", err)
				}
			}
		case <-tm.exitItem.ClickedCh:
			log.Println("Exit clicked")
			// Quit systray first to ensure UI responds
			systray.Quit()
			// Then trigger application shutdown
			if tm.onExit != nil {
				tm.onExit() // Call directly, not in goroutine to ensure cleanup
			}
			return
		}
	}
}

// SetRecordingState updates the tray icon and menu to reflect recording state
func (tm *TrayManager) SetRecordingState(isRecording bool) {
	tm.isRecording = isRecording

	// Guard against early calls before onReady creates menu items
	if tm.toggleItem == nil {
		return
	}

	if isRecording {
		systray.SetIcon(tm.iconMicOn)
		systray.SetTooltip("Speak-to-AI: Recording...")
		tm.toggleItem.SetTitle(fmt.Sprintf("%s Stop Recording", constants.IconStop))
	} else {
		systray.SetIcon(tm.iconMicOff)
		systray.SetTooltip("Speak-to-AI: Ready")
		tm.toggleItem.SetTitle(fmt.Sprintf("%s Start Recording", constants.IconRecording))
	}
}

// SetTooltip sets the tooltip text for the tray icon
func (tm *TrayManager) SetTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}

// Stop stops the tray manager
func (tm *TrayManager) Stop() {
	systray.Quit()
}

// SetAudioActions sets callbacks for audio-related actions (recorder selection, test recording)
func (tm *TrayManager) SetAudioActions(onSelectRecorder func(method string) error, onTestRecording func() error) {
	tm.onSelectRecorder = onSelectRecorder
	tm.onTestRecording = onTestRecording
}

// SetExitAction allows overriding the exit callback (useful once services are wired)
func (tm *TrayManager) SetExitAction(onExit func()) {
	tm.onExit = onExit
}

// SetSettingsActions sets callbacks for general settings
func (tm *TrayManager) SetSettingsActions(
	onSelectVADSensitivity func(sensitivity string) error,
	onSelectLanguage func(language string) error,
	onSelectModelType func(modelType string) error,
	onToggleWorkflowNotifications func() error,
) {
	tm.onSelectVADSens = onSelectVADSensitivity
	tm.onSelectLang = onSelectLanguage
	tm.onSelectModel = onSelectModelType
	tm.onToggleWorkflowNotify = onToggleWorkflowNotifications
}
