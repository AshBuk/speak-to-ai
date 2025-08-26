//go:build !linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package hotkeys

// selectProviderForEnvironment returns a dummy provider on non-Linux to avoid pulling linux deps
func selectProviderForEnvironment(_ HotkeyConfig, _ EnvironmentType) KeyboardEventProvider {
	return NewDummyKeyboardProvider()
}
