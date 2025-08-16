//go:build !linux

package hotkeys

// selectProviderForEnvironment returns a dummy provider on non-Linux to avoid pulling linux deps
func selectProviderForEnvironment(_ HotkeyConfig, _ EnvironmentType) KeyboardEventProvider {
	return NewDummyKeyboardProvider()
}
