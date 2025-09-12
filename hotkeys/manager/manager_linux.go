//go:build linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"os"

	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/providers"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

func selectProviderForEnvironment(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	// Respect provider override from config
	switch config.GetProvider() {
	case "evdev":
		logger.Info("Hotkeys provider override: evdev")
		return providers.NewEvdevKeyboardProvider(config, environment, logger)
	case "dbus":
		logger.Info("Hotkeys provider override: dbus")
		return providers.NewDbusKeyboardProvider(config, environment, logger)
	}
	// AppImage: Prefer evdev due to potential D-Bus portal sandbox issues
	isAppImage := os.Getenv("APPIMAGE") != "" || os.Getenv("APPDIR") != ""
	if isAppImage {
		logger.Info("AppImage detected - checking evdev first for better compatibility")
		evdevProvider := providers.NewEvdevKeyboardProvider(config, environment, logger)
		if evdevProvider.IsSupported() {
			logger.Info("Using evdev keyboard provider (AppImage mode)")
			return evdevProvider
		}
		logger.Info("evdev not available in AppImage, falling back to D-Bus")
		logger.Info("HOTKEY SETUP: For reliable hotkeys in AppImage, run:")
		logger.Info("  sudo usermod -a -G input $USER")
		logger.Info("  Then reboot or log out/in")
	}

	// Try D-Bus provider first (works without root permissions on modern DEs)
	dbusProvider := providers.NewDbusKeyboardProvider(config, environment, logger)
	if dbusProvider.IsSupported() {
		logger.Info("Using D-Bus keyboard provider (GNOME/KDE)")
		return dbusProvider
	}
	logger.Info("D-Bus GlobalShortcuts portal not available, trying evdev...")

	// Fallback to evdev provider (requires root permissions but works everywhere)
	evdevProvider := providers.NewEvdevKeyboardProvider(config, environment, logger)
	if evdevProvider.IsSupported() {
		logger.Info("Using evdev keyboard provider (requires root permissions)")
		return evdevProvider
	}
	logger.Info("evdev not available, hotkeys will be disabled")

	// Final fallback to dummy provider with helpful instructions
	logger.Warning("No supported keyboard provider available.")
	logger.Info("For hotkeys to work:")
	logger.Info("  - On GNOME/KDE: Ensure D-Bus session is running")
	logger.Info("  - On other DEs: Run with sudo or add user to 'input' group")
	logger.Info("  - Alternative: Use system-wide hotkey tools like sxhkd")
	return providers.NewDummyKeyboardProvider(logger)
}
