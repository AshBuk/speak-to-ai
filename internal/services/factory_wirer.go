// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"os"
	"syscall"

	"github.com/AshBuk/dabri/v2/internal/ui/tray"
	"github.com/AshBuk/dabri/v2/internal/ui/window"
)

// wireUI connects tray menu and main window callbacks to the container's
// actions. Callbacks require the fully assembled container.
func wireUI(c *ServiceContainer, tm tray.Manager, win window.Manager) {
	if tm == nil {
		return
	}
	tm.SetCoreActions(c.ToggleRecording, c.UI.ShowConfigFile, c.UI.ShowAboutPage, c.ResetToDefaults)
	tm.SetShowWindowAction(win.Show)
	tm.SetAudioActions(c.setRecordingMethod)
	tm.SetModelAction(c.selectModel)
	tm.SetSettingsActions(c.setLanguage, c.toggleNotifications, c.setOutputMode)
	tm.OutputToolsCallback(c.IO.GetOutputToolNames)
	tm.SetCaptureOnceSupport(c.Hotkeys.SupportsCaptureOnce)
	tm.UpdateSettings(c.Config.GetConfig())
	tm.SetHotkeyRebindAction(c.rebindHotkey)
	tm.SetExitAction(quit)

	// The no-op window backend ignores these
	win.SetActions(window.Actions{
		OnToggleRecording: c.ToggleRecording,
		OnSelectModel:     c.selectModel,
		OnSelectLanguage:  c.setLanguage,
		OnSelectOutput:    c.setOutputMode,
		OnQuit:            quit,
	})
}

// quit requests a clean shutdown through the app's signal handler.
func quit() { _ = syscall.Kill(os.Getpid(), syscall.SIGTERM) }
