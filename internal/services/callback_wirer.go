// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// CallbackWirer wires the callbacks to the tray manager
type CallbackWirer struct {
	logger logger.Logger
}

func NewCallbackWirer(logger logger.Logger) *CallbackWirer {
	return &CallbackWirer{logger: logger}
}

func (cw *CallbackWirer) Wire(container *ServiceContainer, components *Components) {
	if components == nil || components.TrayManager == nil {
		return
	}

	// Core actions (toggle, show config, reset defaults)
	components.TrayManager.SetCoreActions(
		cw.makeToggleCallback(container),
		cw.makeShowConfigCallback(container),
		cw.makeShowAboutCallback(container),
		cw.makeResetToDefaultsCallback(container),
	)

	// Audio actions (recorder selection)
	components.TrayManager.SetAudioActions(cw.makeRecorderSelectionCallback(container))

	// Settings actions (language, notifications)
	components.TrayManager.SetSettingsActions(
		cw.makeLanguageCallback(container),
		cw.makeToggleNotificationsCallback(container),
		cw.makeOutputModeCallback(container),
	)

	// Set callback for getting actual output tool names
	components.TrayManager.SetGetOutputToolsCallback(cw.makeGetOutputToolsCallback(container))

	// Capture once support capability
	components.TrayManager.SetCaptureOnceSupport(cw.makeCaptureOnceSupportCallback(container))

	// Force UI update after callback is set
	cw.updateTraySettings(container, components.TrayManager)

	// Hotkeys rebind wiring
	components.TrayManager.SetHotkeyRebindAction(cw.makeHotkeyRebindCallback(container))

	// Ensure Quit exits app cleanly
	components.TrayManager.SetExitAction(func() {
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	})
}

func (cw *CallbackWirer) makeToggleCallback(container *ServiceContainer) func() error {
	return func() error {
		if container == nil || container.Audio == nil {
			return fmt.Errorf("audio service not available")
		}
		if container.Audio.IsRecording() {
			return container.Audio.HandleStopRecording()
		}
		return container.Audio.HandleStartRecording()
	}
}

func (cw *CallbackWirer) makeShowConfigCallback(container *ServiceContainer) func() error {
	return func() error {
		if container == nil || container.UI == nil {
			return fmt.Errorf("UI service not available")
		}
		return container.UI.ShowConfigFile()
	}
}

func (cw *CallbackWirer) makeShowAboutCallback(container *ServiceContainer) func() error {
	return func() error {
		if container == nil || container.UI == nil {
			return fmt.Errorf("UI service not available")
		}
		return container.UI.ShowAboutPage()
	}
}

func (cw *CallbackWirer) makeResetToDefaultsCallback(container *ServiceContainer) func() error {
	return func() error {
		if container == nil || container.Config == nil {
			return fmt.Errorf("config service not available")
		}

		if err := container.Config.ResetToDefaults(); err != nil {
			return err
		}

		return cw.reloadHotkeysFromConfig(container)
	}
}

func (cw *CallbackWirer) makeRecorderSelectionCallback(container *ServiceContainer) func(string) error {
	return func(method string) error {
		if container == nil || container.Config == nil {
			return fmt.Errorf("config service not available")
		}
		if err := container.Config.UpdateRecordingMethod(method); err != nil {
			return err
		}
		if audioSvc, ok := container.Audio.(*AudioService); ok {
			audioSvc.clearSession()
		}
		cw.updateUISettings(container)
		return nil
	}
}

func (cw *CallbackWirer) makeLanguageCallback(container *ServiceContainer) func(string) error {
	return func(language string) error {
		if container == nil || container.Config == nil {
			return fmt.Errorf("config service not available")
		}
		return container.Config.UpdateLanguage(language)
	}
}

func (cw *CallbackWirer) makeToggleNotificationsCallback(container *ServiceContainer) func() error {
	return func() error {
		if container == nil || container.Config == nil {
			return fmt.Errorf("config service not available")
		}
		if err := container.Config.ToggleWorkflowNotifications(); err != nil {
			return err
		}

		if container.UI != nil {
			enabled := cw.getNotificationStatus(container)
			container.UI.ShowNotification("Workflow Notifications", "Now "+enabled)
		}
		return nil
	}
}

func (cw *CallbackWirer) makeOutputModeCallback(container *ServiceContainer) func(string) error {
	return func(mode string) error {
		if container == nil || container.Config == nil || container.IO == nil {
			return fmt.Errorf("services not available")
		}
		if err := container.Config.UpdateOutputMode(mode); err != nil {
			return err
		}
		if err := container.IO.SetOutputMethod(mode); err != nil {
			return err
		}
		cw.updateUISettings(container)
		return nil
	}
}

func (cw *CallbackWirer) makeGetOutputToolsCallback(container *ServiceContainer) func() (string, string) {
	return func() (clipboardTool, typeTool string) {
		if container == nil || container.IO == nil {
			return "unknown", "unknown"
		}
		if ioSvc, ok := container.IO.(*IOService); ok && ioSvc != nil {
			return ioSvc.GetOutputToolNames()
		}
		return "unknown", "unknown"
	}
}

func (cw *CallbackWirer) makeCaptureOnceSupportCallback(container *ServiceContainer) func() bool {
	return func() bool {
		if container == nil || container.Hotkeys == nil {
			return false
		}
		return container.Hotkeys.SupportsCaptureOnce()
	}
}

func (cw *CallbackWirer) makeHotkeyRebindCallback(container *ServiceContainer) func(string) error {
	return func(action string) error {
		if container == nil || container.UI == nil || container.Config == nil || container.Hotkeys == nil {
			return fmt.Errorf("services not available")
		}

		container.UI.ShowNotification("Rebind Hotkey", "Press new hotkey… (Esc to cancel)")

		combo, err := container.Hotkeys.CaptureOnce(3000)
		if err != nil || strings.TrimSpace(combo) == "" {
			container.UI.ShowNotification("Rebind Hotkey", "Cancelled or timeout")
			return nil
		}

		if err := container.Config.UpdateHotkey(action, combo); err != nil {
			return err
		}

		if err := cw.reloadHotkeysFromConfig(container); err != nil {
			return err
		}

		container.UI.ShowNotification("Hotkey Updated", fmt.Sprintf("%s -> %s", action, combo))
		return nil
	}
}

// Helper methods to reduce duplication
func (cw *CallbackWirer) reloadHotkeysFromConfig(container *ServiceContainer) error {
	if container.Hotkeys == nil {
		return nil
	}

	cfgProvider := cw.makeConfigProvider(container)
	return container.Hotkeys.ReloadFromConfig(
		func() error { return container.Audio.HandleStartRecording() },
		func() error { return container.Audio.HandleStopRecording() },
		cfgProvider,
	)
}

func (cw *CallbackWirer) makeConfigProvider(container *ServiceContainer) func() adapters.HotkeyConfig {
	return func() adapters.HotkeyConfig {
		cfgSvc, ok := container.Config.(*ConfigService)
		if !ok || cfgSvc == nil {
			return adapters.NewConfigAdapter("", "auto")
		}

		cfg := cfgSvc.GetConfig()
		if cfg == nil {
			return adapters.NewConfigAdapter("", "auto")
		}

		return adapters.NewConfigAdapter(cfg.Hotkeys.StartRecording, cfg.Hotkeys.Provider).
			WithAdditionalHotkeys("", cfg.Hotkeys.ShowConfig, cfg.Hotkeys.ResetToDefaults)
	}
}

func (cw *CallbackWirer) updateUISettings(container *ServiceContainer) {
	uiSvc, ok := container.UI.(*UIService)
	if !ok || uiSvc == nil {
		return
	}

	cfgSvc, ok := container.Config.(*ConfigService)
	if !ok || cfgSvc == nil {
		return
	}

	if cfg := cfgSvc.GetConfig(); cfg != nil {
		uiSvc.UpdateSettings(cfg)
	}
}

func (cw *CallbackWirer) updateTraySettings(container *ServiceContainer, trayManager interface{ UpdateSettings(*config.Config) }) {
	if container == nil || container.Config == nil {
		return
	}

	cfgSvc, ok := container.Config.(*ConfigService)
	if !ok || cfgSvc == nil {
		return
	}

	if cfg := cfgSvc.GetConfig(); cfg != nil {
		trayManager.UpdateSettings(cfg)
	}
}

func (cw *CallbackWirer) getNotificationStatus(container *ServiceContainer) string {
	cfgSvc, ok := container.Config.(*ConfigService)
	if !ok || cfgSvc == nil {
		return "disabled"
	}

	cfg := cfgSvc.GetConfig()
	if cfg != nil && cfg.Notifications.EnableWorkflowNotifications {
		return "enabled"
	}
	return "disabled"
}
