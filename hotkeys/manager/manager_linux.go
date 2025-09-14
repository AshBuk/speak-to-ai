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

type runtimeEnvironment int

const (
	envSystem runtimeEnvironment = iota
	envAppImage
	envFlatpak
)

func detectRuntimeEnvironment() runtimeEnvironment {
	if os.Getenv("APPIMAGE") != "" || os.Getenv("APPDIR") != "" {
		return envAppImage
	}
	if os.Getenv("FLATPAK_ID") != "" {
		return envFlatpak
	}
	return envSystem
}

func selectProviderForEnvironment(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	// Handle explicit provider override from config
	switch config.GetProvider() {
	case "evdev":
		logger.Info("Hotkeys provider override: evdev")
		return createEvdevProvider(config, environment, logger)
	case "dbus":
		logger.Info("Hotkeys provider override: dbus")
		return createDbusProvider(config, environment, logger)
	}

	// Auto-select provider based on runtime environment
	switch detectRuntimeEnvironment() {
	case envAppImage:
		return selectAppImageProvider(config, environment, logger)
	case envFlatpak:
		return selectFlatpakProvider(config, environment, logger)
	default:
		return selectSystemProvider(config, environment, logger)
	}
}

func selectAppImageProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	logger.Info("AppImage detected - checking evdev first for better compatibility")

	// Try evdev first (preferred for AppImage due to D-Bus portal issues)
	if evdevProvider := createEvdevProvider(config, environment, logger); evdevProvider.IsSupported() {
		logger.Info("Using evdev keyboard provider (AppImage mode)")
		return evdevProvider
	}

	logger.Info("evdev not available in AppImage, falling back to D-Bus")
	logger.Info("HOTKEY SETUP: For reliable hotkeys in AppImage, run:")
	logger.Info("  sudo usermod -a -G input $USER")
	logger.Info("  Then reboot or log out/in")

	// Fallback to D-Bus
	return createDbusProvider(config, environment, logger)
}

func selectFlatpakProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	logger.Info("Flatpak detected - using D-Bus provider only (evdev blocked by sandbox)")

	// Only D-Bus available in Flatpak sandbox
	return createDbusProvider(config, environment, logger)
}

func selectSystemProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	// Try D-Bus first (works without root permissions on modern DEs)
	if dbusProvider := createDbusProvider(config, environment, logger); dbusProvider.IsSupported() {
		logger.Info("Using D-Bus keyboard provider (GNOME/KDE)")
		return dbusProvider
	}

	logger.Info("D-Bus GlobalShortcuts portal not available, trying evdev...")

	// Fallback to evdev
	if evdevProvider := createEvdevProvider(config, environment, logger); evdevProvider.IsSupported() {
		logger.Info("Using evdev keyboard provider (requires root permissions)")
		return evdevProvider
	}

	logger.Info("evdev not available, hotkeys will be disabled")
	return createFallbackProvider(logger)
}

func createEvdevProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	return providers.NewEvdevKeyboardProvider(config, environment, logger)
}

func createDbusProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	return providers.NewDbusKeyboardProvider(config, environment, logger)
}

func createFallbackProvider(logger logger.Logger) interfaces.KeyboardEventProvider {
	logger.Warning("No supported keyboard provider available.")
	logger.Info("For hotkeys to work:")
	logger.Info("  - On GNOME/KDE: Ensure D-Bus session is running")
	logger.Info("  - On other DEs: Run with sudo or add user to 'input' group")
	logger.Info("  - Alternative: Use system-wide hotkey tools like sxhkd")
	return providers.NewDummyKeyboardProvider(logger)
}
