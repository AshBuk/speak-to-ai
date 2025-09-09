// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/internal/constants"
)

// UI helper methods to consolidate tray and notification management

// setUIReady sets the UI to ready state
func (a *App) setUIReady() {
	a.updateUIState(false, constants.IconReady+" "+constants.MsgReady)
}

// setUIError sets the UI to error state with message
func (a *App) setUIError(message string) {
	a.updateUIState(false, constants.IconError+" "+message)
}

// setUIProcessing sets the UI to processing state with message
func (a *App) setUIProcessing(message string) {
	a.updateUIState(false, constants.IconProcessing+" "+message)
}

// setUIWarning sets the UI to warning state with message
func (a *App) setUIWarning(message string) {
	a.updateUIState(false, constants.IconWarning+" "+message)
}

// setUIRecording sets the UI to recording state with audio level
func (a *App) setUIRecording(level int) {
	levelBar := a.generateLevelBar(level)
	tooltip := fmt.Sprintf("%s %s Level: %s %d%%", constants.IconRecording, constants.MsgRecording, levelBar, level)
	a.updateUIState(true, tooltip)
}

// notifyError shows error notification
func (a *App) notifyError(err error) {
	a.notify("Error", err.Error())
}

// notifyInfo shows info notification
func (a *App) notifyInfo(title, message string) {
	a.notify(title, message)
}

// generateLevelBar creates a visual audio level bar
func (a *App) generateLevelBar(level int) string {
	if level > 100 {
		level = 100
	}

	var levelBar string
	bars := level / 10
	for i := 0; i < 10; i++ {
		if i < bars {
			levelBar += "█"
		} else {
			levelBar += "░"
		}
	}
	return levelBar
}
