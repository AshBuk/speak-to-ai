//go:build linux

package hotkeys

import "log"

func selectProviderForEnvironment(config HotkeyConfig, environment EnvironmentType) KeyboardEventProvider {
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
