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

// Detect the current runtime environment (system, AppImage, or Flatpak)
func detectRuntimeEnvironment() runtimeEnvironment {
	if os.Getenv("APPIMAGE") != "" || os.Getenv("APPDIR") != "" {
		return envAppImage
	}
	if os.Getenv("FLATPAK_ID") != "" {
		return envFlatpak
	}
	return envSystem
}

// Select the most appropriate hotkey provider based on configuration and environment
func selectProviderForEnvironment(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	// Handle an explicit provider override from the configuration
	switch config.GetProvider() {
	case "evdev":
		logger.Info("Hotkeys provider override: evdev")
		return createEvdevProvider(config, environment, logger)
	case "dbus":
		logger.Info("Hotkeys provider override: dbus")
		return createDbusProvider(config, environment, logger)
	}

	// Auto-select the provider based on the runtime environment
	switch detectRuntimeEnvironment() {
	case envAppImage:
		return selectAppImageProvider(config, environment, logger)
	case envFlatpak:
		return selectFlatpakProvider(config, environment, logger)
	default:
		return selectSystemProvider(config, environment, logger)
	}
}

// Select the provider for an AppImage environment
func selectAppImageProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	logger.Info("AppImage detected - checking evdev first for better compatibility")

	// Try evdev first, as it is often more reliable in AppImage contexts
	if evdevProvider := createEvdevProvider(config, environment, logger); evdevProvider.IsSupported() {
		logger.Info("Using evdev keyboard provider (AppImage mode)")
		return evdevProvider
	}

	logger.Info("evdev not available in AppImage, falling back to D-Bus")
	logger.Info("HOTKEY SETUP: For reliable hotkeys in AppImage, run:")
	logger.Info("  sudo usermod -a -G input $USER")
	logger.Info("  Then reboot or log out/in")

	// Fallback to D-Bus if evdev is not available
	return createDbusProvider(config, environment, logger)
}

// Select the provider for a Flatpak environment
func selectFlatpakProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	logger.Info("Flatpak detected - using D-Bus provider only (evdev blocked by sandbox)")

	// Only D-Bus is available within the Flatpak sandbox
	return createDbusProvider(config, environment, logger)
}

// Select the provider for a standard system environment
func selectSystemProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	// Try D-Bus first, as it works without root permissions on modern desktops
	if dbusProvider := createDbusProvider(config, environment, logger); dbusProvider.IsSupported() {
		logger.Info("Using D-Bus keyboard provider (GNOME/KDE)")
		return dbusProvider
	}

	logger.Info("D-Bus GlobalShortcuts portal not available, trying evdev...")

	// Fallback to evdev if D-Bus is not available
	if evdevProvider := createEvdevProvider(config, environment, logger); evdevProvider.IsSupported() {
		logger.Info("Using evdev keyboard provider (requires root permissions)")
		return evdevProvider
	}

	logger.Info("evdev not available, hotkeys will be disabled")
	return createFallbackProvider(logger)
}

// Create an evdev provider instance
func createEvdevProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	return providers.NewEvdevKeyboardProvider(config, environment, logger)
}

// Create a D-Bus provider instance
func createDbusProvider(config adapters.HotkeyConfig, environment interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	return providers.NewDbusKeyboardProvider(config, environment, logger)
}

// Create a dummy provider as a last resort
func createFallbackProvider(logger logger.Logger) interfaces.KeyboardEventProvider {
	logger.Warning("No supported keyboard provider available")
	logger.Info("For hotkeys to work:")
	logger.Info("  - On GNOME/KDE: Ensure D-Bus session is running")
	logger.Info("  - On other DEs: Run with sudo or add user to 'input' group")
	logger.Info("  - Alternative: Use system-wide hotkey tools like sxhkd")
	return providers.NewDummyKeyboardProvider(logger)
}
