//go:build systray

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

// Create settings submenus
func (tm *TrayManager) createSettingsSubmenus() {
	tm.hotkeysMenu = tm.settingsItem.AddSubMenuItem("Hotkeys", "Hotkey settings")
	tm.audioRecorderMenu = tm.settingsItem.AddSubMenuItem("Audio Recorder", "Select audio recorder")
	tm.aiModelMenu = tm.settingsItem.AddSubMenuItem("Set Language", "Select recognition language")

	// VAD Sensitivity submenu
	// vadMenu := tm.settingsItem.AddSubMenuItem("VAD Sensitivity", "Select VAD sensitivity level")
	// tm.audioItems["vad_low"] = vadMenu.AddSubMenuItem("○ Low", "Low sensitivity")
	// tm.audioItems["vad_medium"] = vadMenu.AddSubMenuItem("○ Medium", "Medium sensitivity")
	// tm.audioItems["vad_high"] = vadMenu.AddSubMenuItem("○ High", "High sensitivity")

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

	// Populate Hotkeys menu (display + rebind)
	tm.updateHotkeysMenuUI()

	// Populate Audio Recorder submenu with selectable items
	tm.audioItems["recorder_arecord"] = tm.audioRecorderMenu.AddSubMenuItem(
		"○ arecord (ALSA)",
		"Use arecord (ALSA)",
	)
	tm.audioItems["recorder_ffmpeg"] = tm.audioRecorderMenu.AddSubMenuItem(
		"○ ffmpeg (PulseAudio)",
		"Use ffmpeg (PulseAudio)",
	)

	// Reflect current selections
	tm.updateRecorderRadioUI(tm.config.Audio.RecordingMethod)
	// TODO: Next feature - VAD implementation
	// tm.updateVADRadioUI(tm.config.Audio.VADSensitivity)

	// Handle clicks for recorder selection
	utils.Go(func() {
		for range tm.audioItems["recorder_arecord"].ClickedCh {
			tm.logger.Info("Audio recorder switched to arecord (UI)")
			tm.updateRecorderRadioUI("arecord")
			if tm.onSelectRecorder != nil {
				if err := tm.onSelectRecorder("arecord"); err != nil {
					tm.logger.Error("Error selecting recorder: %v", err)
				}
			}
		}
	})
	utils.Go(func() {
		for range tm.audioItems["recorder_ffmpeg"].ClickedCh {
			tm.logger.Info("Audio recorder switched to ffmpeg (UI)")
			tm.updateRecorderRadioUI("ffmpeg")
			if tm.onSelectRecorder != nil {
				if err := tm.onSelectRecorder("ffmpeg"); err != nil {
					tm.logger.Error("Error selecting recorder: %v", err)
				}
			}
		}
	})

	// Handle VAD sensitivity clicks
	// if tm.audioItems["vad_low"] != nil {
	// 	go func() {
	// 		for range tm.audioItems["vad_low"].ClickedCh {
	// 			tm.logger.Info("VAD sensitivity switched to low (UI)")
	// 			tm.updateVADRadioUI("low")
	// 			if tm.onSelectVADSens != nil {
	// 				if err := tm.onSelectVADSens("low"); err != nil {
	// 					tm.logger.Error("Error selecting VAD sensitivity: %v", err)
	// 				}
	// 			}
	// 		}
	// 	}()
	// }
	// if tm.audioItems["vad_medium"] != nil {
	// 	go func() {
	// 		for range tm.audioItems["vad_medium"].ClickedCh {
	// 			tm.logger.Info("VAD sensitivity switched to medium (UI)")
	// 			tm.updateVADRadioUI("medium")
	// 			if tm.onSelectVADSens != nil {
	// 				if err := tm.onSelectVADSens("medium"); err != nil {
	// 					tm.logger.Error("Error selecting VAD sensitivity: %v", err)
	// 				}
	// 			}
	// 		}
	// 	}()
	// }
	// if tm.audioItems["vad_high"] != nil {
	// 	go func() {
	// 		for range tm.audioItems["vad_high"].ClickedCh {
	// 			tm.logger.Info("VAD sensitivity switched to high (UI)")
	// 			tm.updateVADRadioUI("high")
	// 			if tm.onSelectVADSens != nil {
	// 				if err := tm.onSelectVADSens("high"); err != nil {
	// 					tm.logger.Error("Error selecting VAD sensitivity: %v", err)
	// 				}
	// 			}
	// 		}
	// 	}()
	// }

	// Handle workflow notifications toggle
	if tm.audioItems["workflow_notifications"] != nil {
		utils.Go(func() {
			for range tm.audioItems["workflow_notifications"].ClickedCh {
				tm.logger.Info("Workflow notifications toggle clicked")
				if tm.onToggleWorkflowNotify != nil {
					if err := tm.onToggleWorkflowNotify(); err != nil {
						tm.logger.Error("Error toggling workflow notifications: %v", err)
					}
					// Reflect new state in UI using updated config
					if tm.config != nil {
						tm.updateWorkflowNotificationUI(tm.config.Notifications.EnableWorkflowNotifications)
					}
				}
			}
		})
	}

	// Update workflow notifications toggle
	tm.updateWorkflowNotificationUI(tm.config.Notifications.EnableWorkflowNotifications)

	// Populate Language Selection menu
	langDefs := []struct{ key, title string }{
		{"en", "English"}, {"de", "German"}, {"fr", "French"}, {"es", "Spanish"}, {"he", "Hebrew"}, {"ru", "Russian"},
	}
	for _, l := range langDefs {
		indicator := "○ "
		if tm.config.General.Language == l.key {
			indicator = "● "
		}
		itm := tm.aiModelMenu.AddSubMenuItem(indicator+l.title, "")
		tm.modelItems["lang_"+l.key] = itm
	}

	// Language display (gray text)
	tm.modelItems["language"] = tm.aiModelMenu.AddSubMenuItem(
		"Current: "+tm.config.General.Language,
		"Current language setting",
	)
	tm.modelItems["language"].Disable()

	// Handle language clicks
	for _, k := range []string{"en", "de", "fr", "es", "he", "ru"} {
		if itm := tm.modelItems["lang_"+k]; itm != nil {
			key := k
			utils.Go(func() {
				for range itm.ClickedCh {
					tm.logger.Info("Language switched to %s (UI)", key)
					tm.updateLanguageRadioUI(key)
					if tm.onSelectLang != nil {
						if err := tm.onSelectLang(key); err != nil {
							tm.logger.Error("Error selecting language: %v", err)
						}
					}
				}
			})
		}
	}

	// Populate Output menu
	// Output mode selection
	modeDefs := []struct{ key, title string }{
		{"clipboard", "Clipboard"}, {"active_window", "Active Window"},
	}
	for _, m := range modeDefs {
		indicator := "○ "
		if tm.config.Output.DefaultMode == m.key {
			indicator = "● "
		}
		itm := tm.outputMenu.AddSubMenuItem(indicator+m.title, "")
		tm.outputItems["mode_"+m.key] = itm
	}

	// Mode display (gray text)
	tm.outputItems["mode"] = tm.outputMenu.AddSubMenuItem(
		tm.config.Output.DefaultMode,
		"Current output mode",
	)
	tm.outputItems["mode"].Disable()

	tm.outputItems["clipboard_tool"] = tm.outputMenu.AddSubMenuItem(
		"Clipboard Tool: "+tm.config.Output.ClipboardTool,
		"Current clipboard tool",
	)
	tm.outputItems["clipboard_tool"].Disable()

	tm.outputItems["type_tool"] = tm.outputMenu.AddSubMenuItem(
		"Type Tool: "+tm.config.Output.TypeTool,
		"Current typing tool",
	)
	tm.outputItems["type_tool"].Disable()

	// Handle output mode clicks
	for _, k := range []string{"clipboard", "active_window"} {
		if itm := tm.outputItems["mode_"+k]; itm != nil {
			key := k
			utils.Go(func() {
				for range itm.ClickedCh {
					tm.logger.Info("Output mode switched to %s (UI)", key)
					tm.updateOutputModeRadioUI(key)
					if tm.onSelectOutputMode != nil {
						if err := tm.onSelectOutputMode(key); err != nil {
							tm.logger.Error("Error selecting output mode: %v", err)
						}
					}
				}
			})
		}
	}
}

// updateHotkeysMenuUI ensures hotkeys items exist and reflect current config
func (tm *TrayManager) updateHotkeysMenuUI() {
	if tm.hotkeysMenu == nil || tm.config == nil {
		return
	}
	// Determine capture once support (default true if callback unset)
	supportsCaptureOnce := true
	if tm.getCaptureOnceSupport != nil {
		supportsCaptureOnce = tm.getCaptureOnceSupport()
	}
	// Start/Stop
	if tm.hotkeyItems["rebind_start_stop"] == nil {
		if supportsCaptureOnce {
			reb := tm.hotkeysMenu.AddSubMenuItem("Rebind Start/Stop…", "Change start/stop hotkey")
			tm.hotkeyItems["rebind_start_stop"] = reb
			utils.Go(func() {
				for range reb.ClickedCh {
					if tm.onRebindHotkey != nil {
						if err := tm.onRebindHotkey("start_recording"); err != nil {
							tm.logger.Error("Error rebinding start/stop: %v", err)
						}
					}
				}
			})
		}
	}
	if tm.hotkeyItems["display_start_stop"] == nil {
		tm.hotkeyItems["display_start_stop"] = tm.hotkeysMenu.AddSubMenuItem(
			"Start/Stop Recording: "+tm.config.Hotkeys.StartRecording,
			"Current start/stop recording hotkey",
		)
		tm.hotkeyItems["display_start_stop"].Disable()
	} else {
		tm.hotkeyItems["display_start_stop"].SetTitle("Start/Stop Recording: " + tm.config.Hotkeys.StartRecording)
	}
	// Show config
	if tm.hotkeyItems["rebind_show_config"] == nil {
		if supportsCaptureOnce {
			reb := tm.hotkeysMenu.AddSubMenuItem("Rebind Show Config…", "Change show config hotkey")
			tm.hotkeyItems["rebind_show_config"] = reb
			utils.Go(func() {
				for range reb.ClickedCh {
					if tm.onRebindHotkey != nil {
						if err := tm.onRebindHotkey("show_config"); err != nil {
							tm.logger.Error("Error rebinding show config: %v", err)
						}
					}
				}
			})
		}
	}
	if tm.hotkeyItems["display_show_config"] == nil {
		tm.hotkeyItems["display_show_config"] = tm.hotkeysMenu.AddSubMenuItem(
			"Show Config: "+tm.config.Hotkeys.ShowConfig,
			"Current show config hotkey",
		)
		tm.hotkeyItems["display_show_config"].Disable()
	} else {
		tm.hotkeyItems["display_show_config"].SetTitle("Show Config: " + tm.config.Hotkeys.ShowConfig)
	}
	// Reset defaults
	if tm.hotkeyItems["rebind_reset_defaults"] == nil {
		if supportsCaptureOnce {
			reb := tm.hotkeysMenu.AddSubMenuItem("Rebind Reset to Defaults…", "Change reset defaults hotkey")
			tm.hotkeyItems["rebind_reset_defaults"] = reb
			utils.Go(func() {
				for range reb.ClickedCh {
					if tm.onRebindHotkey != nil {
						if err := tm.onRebindHotkey("reset_to_defaults"); err != nil {
							tm.logger.Error("Error rebinding reset defaults: %v", err)
						}
					}
				}
			})
		}
	}
	if tm.hotkeyItems["display_reset_defaults"] == nil {
		tm.hotkeyItems["display_reset_defaults"] = tm.hotkeysMenu.AddSubMenuItem(
			"Reset to Defaults: "+tm.config.Hotkeys.ResetToDefaults,
			"Current reset to defaults hotkey",
		)
		tm.hotkeyItems["display_reset_defaults"].Disable()
	} else {
		tm.hotkeyItems["display_reset_defaults"].SetTitle("Reset to Defaults: " + tm.config.Hotkeys.ResetToDefaults)
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

// updateVADRadioUI updates VAD sensitivity radio titles
// func (tm *TrayManager) updateVADRadioUI(level string) {
// 	low := tm.audioItems["vad_low"]
// 	med := tm.audioItems["vad_medium"]
// 	high := tm.audioItems["vad_high"]
// 	if low == nil || med == nil || high == nil {
// 		return
// 	}
// 	switch level {
// 	case "low":
// 		low.SetTitle("● Low")
// 		med.SetTitle("○ Medium")
// 		high.SetTitle("○ High")
// 	case "high":
// 		high.SetTitle("● High")
// 		low.SetTitle("○ Low")
// 		med.SetTitle("○ Medium")
// 	default:
// 		med.SetTitle("● Medium")
// 		low.SetTitle("○ Low")
// 		high.SetTitle("○ High")
// 	}
// }

// updateLanguageRadioUI updates selection marks for language menu
func (tm *TrayManager) updateLanguageRadioUI(lang string) {
	langDefs := map[string]string{
		"en": "English", "de": "German", "fr": "French",
		"es": "Spanish", "he": "Hebrew", "ru": "Russian",
	}

	for key, title := range langDefs {
		if itm := tm.modelItems["lang_"+key]; itm != nil {
			if key == lang {
				itm.SetTitle("● " + title)
			} else {
				itm.SetTitle("○ " + title)
			}
		}
	}

	// Update the gray text display
	if langDisplay := tm.modelItems["language"]; langDisplay != nil {
		langDisplay.SetTitle(lang)
	}
}

// updateWorkflowNotificationUI updates the workflow notifications toggle UI
func (tm *TrayManager) updateWorkflowNotificationUI(enabled bool) {
	item := tm.audioItems["workflow_notifications"]
	if item == nil {
		return
	}

	if enabled {
		item.SetTitle("● Workflow Notifications")
	} else {
		item.SetTitle("○ Workflow Notifications")
	}
}

// updateOutputModeRadioUI updates selection marks for output mode menu
func (tm *TrayManager) updateOutputModeRadioUI(mode string) {
	modeDefs := map[string]string{
		"clipboard": "Clipboard", "active_window": "Active Window",
	}

	for key, title := range modeDefs {
		if itm := tm.outputItems["mode_"+key]; itm != nil {
			if key == mode {
				itm.SetTitle("● " + title)
			} else {
				itm.SetTitle("○ " + title)
			}
		}
	}

	// Update the gray text display
	if modeDisplay := tm.outputItems["mode"]; modeDisplay != nil {
		modeDisplay.SetTitle(mode)
	}
}

// updateOutputUI updates the output settings display
func (tm *TrayManager) updateOutputUI() {
	if tm.config == nil {
		return
	}

	// Update mode radio buttons and display
	tm.updateOutputModeRadioUI(tm.config.Output.DefaultMode)

	// Get actual tool names if callback is available
	clipboardTool := tm.config.Output.ClipboardTool
	typeTool := tm.config.Output.TypeTool
	if tm.onGetOutputTools != nil {
		clipboardTool, typeTool = tm.onGetOutputTools()
	}

	// Update clipboard tool display
	if clipboardItem := tm.outputItems["clipboard_tool"]; clipboardItem != nil {
		clipboardItem.SetTitle("Clipboard Tool: " + clipboardTool)
	}

	// Update type tool display
	if typeItem := tm.outputItems["type_tool"]; typeItem != nil {
		typeItem.SetTitle("Type Tool: " + typeTool)
	}
}
