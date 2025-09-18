// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"os"
	"strings"
	"syscall"

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
		func() error { // toggle
			if container == nil || container.Audio == nil {
				return fmt.Errorf("audio service not available")
			}
			if container.Audio.IsRecording() {
				return container.Audio.HandleStopRecording()
			}
			return container.Audio.HandleStartRecording()
		},
		func() error { // show config
			if container == nil || container.UI == nil {
				return fmt.Errorf("UI service not available")
			}
			return container.UI.ShowConfigFile()
		},
		func() error { // show about
			if container == nil || container.UI == nil {
				return fmt.Errorf("UI service not available")
			}
			return container.UI.ShowAboutPage()
		},
		func() error { // reset to defaults
			if container == nil || container.Config == nil {
				return fmt.Errorf("config service not available")
			}

			if err := container.Config.ResetToDefaults(); err != nil {
				return err
			}

			// Hotkeys: reload manager to apply default combos immediately
			if container.Hotkeys != nil {
				cfgProvider := func() adapters.HotkeyConfig {
					if cfgSvc, ok := container.Config.(*ConfigService); ok && cfgSvc != nil {
						if cfg := cfgSvc.GetConfig(); cfg != nil {
							return adapters.NewConfigAdapter(cfg.Hotkeys.StartRecording, cfg.Hotkeys.Provider).
								WithAdditionalHotkeys(
									"",
									cfg.Hotkeys.ShowConfig,
									cfg.Hotkeys.ResetToDefaults,
								)
						}
					}
					return adapters.NewConfigAdapter("", "auto")
				}

				// Ensure callbacks preserved
				if err := container.Hotkeys.ReloadFromConfig(
					func() error { return container.Audio.HandleStartRecording() },
					func() error { return container.Audio.HandleStopRecording() },
					cfgProvider,
				); err != nil {
					return err
				}
			}

			return nil
		},
	)

	// Audio actions (recorder selection)
	components.TrayManager.SetAudioActions(
		func(method string) error {
			// Persist method and reinit on next start
			if container == nil || container.Config == nil {
				return fmt.Errorf("config service not available")
			}
			if err := container.Config.UpdateRecordingMethod(method); err != nil {
				return err
			}
			if audioSvc, ok := container.Audio.(*AudioService); ok {
				audioSvc.clearSession()
			}
			// Update tray to reflect new selection immediately
			if uiSvc, ok := container.UI.(*UIService); ok && uiSvc != nil {
				if cfgSvc, ok2 := container.Config.(*ConfigService); ok2 && cfgSvc != nil {
					if cfg := cfgSvc.GetConfig(); cfg != nil {
						uiSvc.UpdateSettings(cfg)
					}
				}
			}
			return nil
		},
	)

	// Settings actions (language, notifications)
	// TODO: Next feature - VAD implementation
	components.TrayManager.SetSettingsActions(
		// TODO: Next feature - VAD implementation
		// func(sensitivity string) error {
		//	if container == nil || container.Config == nil {
		//		return fmt.Errorf("config service not available")
		//	}
		//	return container.Config.UpdateVADSensitivity(sensitivity)
		// },
		func(language string) error {
			if container == nil || container.Config == nil {
				return fmt.Errorf("config service not available")
			}

			err := container.Config.UpdateLanguage(language)
			if err == nil && container.Audio != nil {
				if audioSvc, ok := container.Audio.(*AudioService); ok {
					audioSvc.clearSession()
				}
			}

			return err
		},
		func() error {
			if container == nil || container.Config == nil {
				return fmt.Errorf("config service not available")
			}
			if err := container.Config.ToggleWorkflowNotifications(); err != nil {
				return err
			}
			// Inform user about new state
			if container.UI != nil {
				enabled := "disabled"
				if cfgSvc, ok := container.Config.(*ConfigService); ok && cfgSvc != nil {
					if c := cfgSvc.GetConfig(); c != nil {
						if c.Notifications.EnableWorkflowNotifications {
							enabled = "enabled"
						}
					}
				}
				container.UI.ShowNotification("Workflow Notifications", "Now "+enabled)
			}
			return nil
		},
		func(mode string) error {
			if container == nil || container.Config == nil || container.IO == nil {
				return fmt.Errorf("services not available")
			}
			if err := container.Config.UpdateOutputMode(mode); err != nil {
				return err
			}
			// Apply the output method change to IOService
			if err := container.IO.SetOutputMethod(mode); err != nil {
				return err
			}
			// Update tray to reflect new selection immediately
			if uiSvc, ok := container.UI.(*UIService); ok && uiSvc != nil {
				if cfgSvc, ok2 := container.Config.(*ConfigService); ok2 && cfgSvc != nil {
					if cfg := cfgSvc.GetConfig(); cfg != nil {
						uiSvc.UpdateSettings(cfg)
					}
				}
			}
			return nil
		},
	)

	// Set callback for getting actual output tool names
	components.TrayManager.SetGetOutputToolsCallback(func() (clipboardTool, typeTool string) {
		if container == nil || container.IO == nil {
			return "unknown", "unknown"
		}
		if ioSvc, ok := container.IO.(*IOService); ok && ioSvc != nil {
			return ioSvc.GetOutputToolNames()
		}
		return "unknown", "unknown"
	})

	// Capture once support capability
	components.TrayManager.SetCaptureOnceSupport(func() bool {
		if container == nil || container.Hotkeys == nil {
			return false
		}
		return container.Hotkeys.SupportsCaptureOnce()
	})

	// Force UI update after callback is set
	if container != nil && container.Config != nil {
		if cfgSvc, ok := container.Config.(*ConfigService); ok && cfgSvc != nil {
			if cfg := cfgSvc.GetConfig(); cfg != nil {
				components.TrayManager.UpdateSettings(cfg)
			}
		}
	}

	// Hotkeys rebind wiring
	components.TrayManager.SetHotkeyRebindAction(func(action string) error {
		if container == nil || container.UI == nil || container.Config == nil || container.Hotkeys == nil {
			return fmt.Errorf("services not available")
		}

		// Show prompt and capture through provider
		if container.UI != nil {
			container.UI.ShowNotification("Rebind Hotkey", "Press new hotkey… (Esc to cancel)")
		}
		combo, err := container.Hotkeys.CaptureOnce(3000)
		if err != nil || strings.TrimSpace(combo) == "" {
			if container.UI != nil {
				container.UI.ShowNotification("Rebind Hotkey", "Cancelled or timeout")
			}
			return nil
		}

		// Persist in config
		if err := container.Config.UpdateHotkey(action, combo); err != nil {
			return err
		}

		// Reload hotkeys in manager from updated config
		cfgProvider := func() adapters.HotkeyConfig {
			if cfgSvc, ok := container.Config.(*ConfigService); ok && cfgSvc != nil {
				if cfg := cfgSvc.GetConfig(); cfg != nil {
					return adapters.NewConfigAdapter(cfg.Hotkeys.StartRecording, cfg.Hotkeys.Provider).
						WithAdditionalHotkeys(
							"",
							cfg.Hotkeys.ShowConfig,
							cfg.Hotkeys.ResetToDefaults,
						)
				}
			}
			return adapters.NewConfigAdapter("", "auto")
		}

		// Ensure callbacks preserved
		if err := container.Hotkeys.ReloadFromConfig(
			func() error { return container.Audio.HandleStartRecording() },
			func() error { return container.Audio.HandleStopRecording() },
			cfgProvider,
		); err != nil {
			return err
		}

		// Notify user about successful rebind
		if container.UI != nil {
			title := "Hotkey Updated"
			msg := fmt.Sprintf("%s -> %s", action, combo)
			container.UI.ShowNotification(title, msg)
		}
		return nil
	})

	// Ensure Quit exits app cleanly
	components.TrayManager.SetExitAction(func() {
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	})
}
