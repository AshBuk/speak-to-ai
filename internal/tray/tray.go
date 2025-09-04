//go:build systray

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"fmt"
	"log"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/getlantern/systray"
)

// TrayManager manages the system tray icon and menu
type TrayManager struct {
	isRecording    bool
	iconMicOff     []byte
	iconMicOn      []byte
	onExit         func()
	onToggle       func() error
	onShowConfig   func() error
	onReloadConfig func() error
	config         *config.Config

	// Menu items
	toggleItem       *systray.MenuItem
	statusItem       *systray.MenuItem
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
}

// NewTrayManager creates a new tray manager instance
func NewTrayManager(iconMicOff, iconMicOn []byte, onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) *TrayManager {
	return &TrayManager{
		isRecording:    false,
		iconMicOff:     iconMicOff,
		iconMicOn:      iconMicOn,
		onExit:         onExit,
		onToggle:       onToggle,
		onShowConfig:   onShowConfig,
		onReloadConfig: onReloadConfig,
		hotkeyItems:    make(map[string]*systray.MenuItem),
		audioItems:     make(map[string]*systray.MenuItem),
		modelItems:     make(map[string]*systray.MenuItem),
		outputItems:    make(map[string]*systray.MenuItem),
	}
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
	tm.toggleItem = systray.AddMenuItem("🔴 Start Recording", "Start/Stop recording")
	tm.statusItem = systray.AddMenuItem("ℹ️  Status: Ready", "Current status")
	tm.statusItem.Disable() // Status item is just for display

	systray.AddSeparator()

	// Settings submenu
	tm.settingsItem = systray.AddMenuItem("⚙️  Settings", "Application settings")
	tm.createSettingsSubmenus()

	systray.AddSeparator()

	// Config actions
	tm.showConfigItem = systray.AddMenuItem("📄 Show Config File", "Open configuration file")
	tm.reloadConfigItem = systray.AddMenuItem("🔄 Reload Config", "Reload configuration from file")

	systray.AddSeparator()

	tm.exitItem = systray.AddMenuItem("❌ Quit", "Quit Speak-to-AI")

	// Handle menu item clicks
	go tm.handleMenuClicks()
}

// createSettingsSubmenus creates the settings submenus
func (tm *TrayManager) createSettingsSubmenus() {
	tm.hotkeysMenu = tm.settingsItem.AddSubMenuItem("Hotkeys", "Hotkey settings")
	tm.audioRecorderMenu = tm.settingsItem.AddSubMenuItem("Audio Recorder", "Select audio recorder")

	// Device and Sample Rate as top-level items, placed after Audio Recorder
	tm.audioItems["device"] = tm.settingsItem.AddSubMenuItem(
		"Device: -",
		"Current audio device",
	)
	tm.audioItems["device"].Disable()

	tm.audioItems["sample_rate"] = tm.settingsItem.AddSubMenuItem(
		"Sample Rate: -",
		"Current sample rate",
	)
	tm.audioItems["sample_rate"].Disable()

	tm.aiModelMenu = tm.settingsItem.AddSubMenuItem("Whisper Model", "AI model settings")
	tm.outputMenu = tm.settingsItem.AddSubMenuItem("Output", "Output settings")

	// Populate with initial values if config is available
	if tm.config != nil {
		tm.populateSettingsMenus()
	}
}

// populateSettingsMenus populates the settings submenus with current config values
func (tm *TrayManager) populateSettingsMenus() {
	if tm.config == nil {
		return
	}

	// Populate Hotkeys menu
	tm.hotkeyItems["start_recording"] = tm.hotkeysMenu.AddSubMenuItem(
		fmt.Sprintf("Start Recording: %s", tm.config.Hotkeys.StartRecording),
		"Current start recording hotkey",
	)
	tm.hotkeyItems["start_recording"].Disable()

	// Populate Audio Recorder submenu with selectable items
	tm.audioItems["recorder_arecord"] = tm.audioRecorderMenu.AddSubMenuItem(
		"○ arecord (ALSA)",
		"Use arecord (ALSA)",
	)
	tm.audioItems["recorder_ffmpeg"] = tm.audioRecorderMenu.AddSubMenuItem(
		"○ ffmpeg (PulseAudio)",
		"Use ffmpeg (PulseAudio)",
	)

	// Add test recording item
	tm.audioItems["recorder_test"] = tm.audioRecorderMenu.AddSubMenuItem(
		"Test Recording",
		"Record 3s sample to validate settings",
	)

	// Reflect current selection
	tm.updateRecorderRadioUI(tm.config.Audio.RecordingMethod)

	// Handle clicks for recorder selection
	go func() {
		for range tm.audioItems["recorder_arecord"].ClickedCh {
			log.Println("Audio recorder switched to arecord (UI)")
			tm.updateRecorderRadioUI("arecord")
			if tm.onSelectRecorder != nil {
				if err := tm.onSelectRecorder("arecord"); err != nil {
					log.Printf("Error selecting recorder: %v", err)
				}
			}
		}
	}()
	go func() {
		for range tm.audioItems["recorder_ffmpeg"].ClickedCh {
			log.Println("Audio recorder switched to ffmpeg (UI)")
			tm.updateRecorderRadioUI("ffmpeg")
			if tm.onSelectRecorder != nil {
				if err := tm.onSelectRecorder("ffmpeg"); err != nil {
					log.Printf("Error selecting recorder: %v", err)
				}
			}
		}
	}()

	// Handle test recording
	go func() {
		for range tm.audioItems["recorder_test"].ClickedCh {
			log.Println("Test recording clicked")
			if tm.onTestRecording != nil {
				if err := tm.onTestRecording(); err != nil {
					log.Printf("Test recording failed: %v", err)
				}
			}
		}
	}()

	// Update Device and Sample Rate placeholders created in createSettingsSubmenus
	if tm.audioItems["device"] != nil {
		tm.audioItems["device"].SetTitle(fmt.Sprintf("Device: %s", tm.config.Audio.Device))
	}
	if tm.audioItems["sample_rate"] != nil {
		tm.audioItems["sample_rate"].SetTitle(fmt.Sprintf("Sample Rate: %d Hz", tm.config.Audio.SampleRate))
	}

	// Populate AI Model menu
	tm.modelItems["type"] = tm.aiModelMenu.AddSubMenuItem(
		fmt.Sprintf("Model: %s", tm.config.General.ModelType),
		"Current model type",
	)
	tm.modelItems["type"].Disable()

	tm.modelItems["language"] = tm.aiModelMenu.AddSubMenuItem(
		fmt.Sprintf("Language: %s", tm.config.General.Language),
		"Current language setting",
	)
	tm.modelItems["language"].Disable()

	// Populate Output menu
	tm.outputItems["mode"] = tm.outputMenu.AddSubMenuItem(
		fmt.Sprintf("Mode: %s", tm.config.Output.DefaultMode),
		"Current output mode",
	)
	tm.outputItems["mode"].Disable()

	tm.outputItems["clipboard_tool"] = tm.outputMenu.AddSubMenuItem(
		fmt.Sprintf("Clipboard Tool: %s", tm.config.Output.ClipboardTool),
		"Current clipboard tool",
	)
	tm.outputItems["clipboard_tool"].Disable()

	tm.outputItems["type_tool"] = tm.outputMenu.AddSubMenuItem(
		fmt.Sprintf("Type Tool: %s", tm.config.Output.TypeTool),
		"Current typing tool",
	)
	tm.outputItems["type_tool"].Disable()
}

// UpdateSettings updates the settings display with new configuration
func (tm *TrayManager) UpdateSettings(config *config.Config) {
	tm.config = config

	// Update existing menu items if they exist
	if tm.hotkeyItems["start_recording"] != nil {
		tm.hotkeyItems["start_recording"].SetTitle(fmt.Sprintf("Start Recording: %s", config.Hotkeys.StartRecording))
	}

	if tm.audioItems["device"] != nil {
		tm.audioItems["device"].SetTitle(fmt.Sprintf("Device: %s", config.Audio.Device))
	}

	if tm.audioItems["sample_rate"] != nil {
		tm.audioItems["sample_rate"].SetTitle(fmt.Sprintf("Sample Rate: %d Hz", config.Audio.SampleRate))
	}

	// Update recorder selection UI
	tm.updateRecorderRadioUI(config.Audio.RecordingMethod)

	if tm.modelItems["type"] != nil {
		tm.modelItems["type"].SetTitle(fmt.Sprintf("Model: %s", config.General.ModelType))
	}

	if tm.modelItems["language"] != nil {
		tm.modelItems["language"].SetTitle(fmt.Sprintf("Language: %s", config.General.Language))
	}

	if tm.outputItems["mode"] != nil {
		tm.outputItems["mode"].SetTitle(fmt.Sprintf("Mode: %s", config.Output.DefaultMode))
	}

	if tm.outputItems["clipboard_tool"] != nil {
		tm.outputItems["clipboard_tool"].SetTitle(fmt.Sprintf("Clipboard Tool: %s", config.Output.ClipboardTool))
	}

	if tm.outputItems["type_tool"] != nil {
		tm.outputItems["type_tool"].SetTitle(fmt.Sprintf("Type Tool: %s", config.Output.TypeTool))
	}
}

// updateRecorderRadioUI updates titles to emulate radio selection
func (tm *TrayManager) updateRecorderRadioUI(method string) {
	arecordItem := tm.audioItems["recorder_arecord"]
	ffmpegItem := tm.audioItems["recorder_ffmpeg"]
	if arecordItem == nil || ffmpegItem == nil {
		return
	}

	switch method {
	case "arecord":
		arecordItem.SetTitle("● arecord (ALSA)")
		ffmpegItem.SetTitle("○ ffmpeg (PulseAudio)")
	case "ffmpeg":
		ffmpegItem.SetTitle("● ffmpeg (PulseAudio)")
		arecordItem.SetTitle("○ arecord (ALSA)")
	default:
		arecordItem.SetTitle("○ arecord (ALSA)")
		ffmpegItem.SetTitle("○ ffmpeg (PulseAudio)")
	}
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
			log.Println("Reload config clicked")
			if tm.onReloadConfig != nil {
				if err := tm.onReloadConfig(); err != nil {
					log.Printf("Error reloading config: %v", err)
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

	if isRecording {
		systray.SetIcon(tm.iconMicOn)
		systray.SetTooltip("Speak-to-AI: Recording...")
		tm.toggleItem.SetTitle("🔴 Stop Recording")
		tm.statusItem.SetTitle("ℹ️  Status: Recording")
	} else {
		systray.SetIcon(tm.iconMicOff)
		systray.SetTooltip("Speak-to-AI: Ready")
		tm.toggleItem.SetTitle("🔴 Start Recording")
		tm.statusItem.SetTitle("ℹ️  Status: Ready")
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
