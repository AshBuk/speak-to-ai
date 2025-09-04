//go:build linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package hotkeys

import (
	"log"
	"os"
)

func selectProviderForEnvironment(config HotkeyConfig, environment EnvironmentType) KeyboardEventProvider {
	// AppImage: Prefer evdev due to potential D-Bus portal sandbox issues
	isAppImage := os.Getenv("APPIMAGE") != "" || os.Getenv("APPDIR") != ""
	if isAppImage {
		log.Println("AppImage detected - checking evdev first for better compatibility")
		evdevProvider := NewEvdevKeyboardProvider(config, environment)
		if evdevProvider.IsSupported() {
			log.Println("Using evdev keyboard provider (AppImage mode)")
			return evdevProvider
		}
		log.Println("evdev not available in AppImage, falling back to D-Bus")
		log.Println("HOTKEY SETUP: For reliable hotkeys in AppImage, run:")
		log.Println("  sudo usermod -a -G input $USER")
		log.Println("  Then reboot or log out/in")
	}

	// Try D-Bus provider first (works without root permissions on modern DEs)
	dbusProvider := NewDbusKeyboardProvider(config, environment)
	if dbusProvider.IsSupported() {
		log.Println("Using D-Bus keyboard provider (GNOME/KDE)")
		return dbusProvider
	}
	log.Println("D-Bus GlobalShortcuts portal not available, trying evdev...")

	// Fallback to evdev provider (requires root permissions but works everywhere)
	evdevProvider := NewEvdevKeyboardProvider(config, environment)
	if evdevProvider.IsSupported() {
		log.Println("Using evdev keyboard provider (requires root permissions)")
		return evdevProvider
	}
	log.Println("evdev not available, hotkeys will be disabled")

	// Final fallback to dummy provider with helpful instructions
	log.Println("Warning: No supported keyboard provider available.")
	log.Println("For hotkeys to work:")
	log.Println("  - On GNOME/KDE: Ensure D-Bus session is running")
	log.Println("  - On other DEs: Run with sudo or add user to 'input' group")
	log.Println("  - Alternative: Use system-wide hotkey tools like sxhkd")
	return NewDummyKeyboardProvider()
}
