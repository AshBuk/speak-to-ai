//go:build systray

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"fmt"
	"log"
)

// createSettingsSubmenus creates the settings submenus
func (tm *TrayManager) createSettingsSubmenus() {
	tm.hotkeysMenu = tm.settingsItem.AddSubMenuItem("Hotkeys", "Hotkey settings")
	tm.audioRecorderMenu = tm.settingsItem.AddSubMenuItem("Audio Recorder", "Select audio recorder")

	tm.aiModelMenu = tm.settingsItem.AddSubMenuItem("Whisper Model", "AI model settings")

	// VAD Sensitivity submenu
	vadMenu := tm.settingsItem.AddSubMenuItem("VAD Sensitivity", "Select VAD sensitivity level")
	tm.audioItems["vad_low"] = vadMenu.AddSubMenuItem("○ Low", "Low sensitivity")
	tm.audioItems["vad_medium"] = vadMenu.AddSubMenuItem("○ Medium", "Medium sensitivity")
	tm.audioItems["vad_high"] = vadMenu.AddSubMenuItem("○ High", "High sensitivity")

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

	// Reflect current selections
	tm.updateRecorderRadioUI(tm.config.Audio.RecordingMethod)
	// TODO: Next feature - VAD implementation
	// tm.updateVADRadioUI(tm.config.Audio.VADSensitivity)

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

	// Handle VAD sensitivity clicks
	if tm.audioItems["vad_low"] != nil {
		go func() {
			for range tm.audioItems["vad_low"].ClickedCh {
				log.Println("VAD sensitivity switched to low (UI)")
				tm.updateVADRadioUI("low")
				if tm.onSelectVADSens != nil {
					if err := tm.onSelectVADSens("low"); err != nil {
						log.Printf("Error selecting VAD sensitivity: %v", err)
					}
				}
			}
		}()
	}
	if tm.audioItems["vad_medium"] != nil {
		go func() {
			for range tm.audioItems["vad_medium"].ClickedCh {
				log.Println("VAD sensitivity switched to medium (UI)")
				tm.updateVADRadioUI("medium")
				if tm.onSelectVADSens != nil {
					if err := tm.onSelectVADSens("medium"); err != nil {
						log.Printf("Error selecting VAD sensitivity: %v", err)
					}
				}
			}
		}()
	}
	if tm.audioItems["vad_high"] != nil {
		go func() {
			for range tm.audioItems["vad_high"].ClickedCh {
				log.Println("VAD sensitivity switched to high (UI)")
				tm.updateVADRadioUI("high")
				if tm.onSelectVADSens != nil {
					if err := tm.onSelectVADSens("high"); err != nil {
						log.Printf("Error selecting VAD sensitivity: %v", err)
					}
				}
			}
		}()
	}

	// Handle workflow notifications toggle
	if tm.audioItems["workflow_notifications"] != nil {
		go func() {
			for range tm.audioItems["workflow_notifications"].ClickedCh {
				log.Println("Workflow notifications toggle clicked")
				if tm.onToggleWorkflowNotify != nil {
					if err := tm.onToggleWorkflowNotify(); err != nil {
						log.Printf("Error toggling workflow notifications: %v", err)
					}
					// Reflect new state in UI using updated config
					if tm.config != nil {
						tm.updateWorkflowNotificationUI(tm.config.Notifications.EnableWorkflowNotifications)
					}
				}
			}
		}()
	}

	// Update workflow notifications toggle
	tm.updateWorkflowNotificationUI(tm.config.Notifications.EnableWorkflowNotifications)

	// Populate AI Model menu
	// Model type submenu
	modelMenu := tm.aiModelMenu.AddSubMenuItem("Model", "Select Whisper model type")
	for _, m := range []string{"tiny", "base", "small", "medium", "large"} {
		indicator := "○ "
		if tm.config.General.ModelType == m {
			indicator = "● "
		}
		itm := modelMenu.AddSubMenuItem(indicator+m, "")
		tm.modelItems["model_"+m] = itm
	}

	// Model type display (gray text under Model)
	tm.modelItems["type"] = tm.aiModelMenu.AddSubMenuItem(
		tm.config.General.ModelType,
		"Current model type",
	)
	tm.modelItems["type"].Disable()

	// Language selection submenu
	langMenu := tm.aiModelMenu.AddSubMenuItem("Set Language", "Select recognition language")
	langDefs := []struct{ key, title string }{
		{"auto", "Auto"}, {"en", "English"}, {"de", "German"}, {"fr", "French"}, {"es", "Spanish"}, {"he", "Hebrew"}, {"ru", "Russian"},
	}
	for _, l := range langDefs {
		indicator := "○ "
		if tm.config.General.Language == l.key {
			indicator = "● "
		}
		itm := langMenu.AddSubMenuItem(indicator+l.title, "")
		tm.modelItems["lang_"+l.key] = itm
	}

	// Language display (gray text under Set Language)
	tm.modelItems["language"] = tm.aiModelMenu.AddSubMenuItem(
		tm.config.General.Language,
		"Current language setting",
	)
	tm.modelItems["language"].Disable()

	// Handle language clicks
	for _, k := range []string{"auto", "en", "de", "fr", "es", "he", "ru"} {
		if itm := tm.modelItems["lang_"+k]; itm != nil {
			key := k
			go func() {
				for range itm.ClickedCh {
					log.Printf("Language switched to %s (UI)", key)
					tm.updateLanguageRadioUI(key)
					if tm.onSelectLang != nil {
						if err := tm.onSelectLang(key); err != nil {
							log.Printf("Error selecting language: %v", err)
						}
					}
				}
			}()
		}
	}

	// Handle model type clicks
	for _, m := range []string{"tiny", "base", "small", "medium", "large"} {
		if itm := tm.modelItems["model_"+m]; itm != nil {
			mm := m
			go func() {
				for range itm.ClickedCh {
					log.Printf("Model switched to %s (UI)", mm)
					tm.updateModelRadioUI(mm)
					if tm.onSelectModel != nil {
						if err := tm.onSelectModel(mm); err != nil {
							log.Printf("Error selecting model: %v", err)
						}
					}
				}
			}()
		}
	}

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
func (tm *TrayManager) updateVADRadioUI(level string) {
	low := tm.audioItems["vad_low"]
	med := tm.audioItems["vad_medium"]
	high := tm.audioItems["vad_high"]
	if low == nil || med == nil || high == nil {
		return
	}
	switch level {
	case "low":
		low.SetTitle("● Low")
		med.SetTitle("○ Medium")
		high.SetTitle("○ High")
	case "high":
		high.SetTitle("● High")
		low.SetTitle("○ Low")
		med.SetTitle("○ Medium")
	default:
		med.SetTitle("● Medium")
		low.SetTitle("○ Low")
		high.SetTitle("○ High")
	}
}

// updateLanguageRadioUI updates selection marks for language menu
func (tm *TrayManager) updateLanguageRadioUI(lang string) {
	langDefs := map[string]string{
		"auto": "Auto", "en": "English", "de": "German",
		"fr": "French", "es": "Spanish", "he": "Hebrew", "ru": "Russian",
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

// updateModelRadioUI updates selection marks for model type
func (tm *TrayManager) updateModelRadioUI(model string) {
	types := []string{"tiny", "base", "small", "medium", "large"}
	for _, t := range types {
		if itm := tm.modelItems["model_"+t]; itm != nil {
			if t == model {
				itm.SetTitle("● " + t)
			} else {
				itm.SetTitle("○ " + t)
			}
		}
	}

	// Update the gray text display
	if modelDisplay := tm.modelItems["type"]; modelDisplay != nil {
		modelDisplay.SetTitle(model)
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
