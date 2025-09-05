//go:build !linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/providers"
)

// selectProviderForEnvironment returns a dummy provider on non-Linux to avoid pulling linux deps
func selectProviderForEnvironment(_ adapters.HotkeyConfig, _ interfaces.EnvironmentType) interfaces.KeyboardEventProvider {
	return providers.NewDummyKeyboardProvider()
}
