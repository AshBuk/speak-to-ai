//go:build !linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package manager

import (
	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	"github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/providers"
	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// Return a dummy provider on non-Linux systems to avoid build errors
func selectProviderForEnvironment(_ adapters.HotkeyConfig, _ interfaces.EnvironmentType, logger logger.Logger) interfaces.KeyboardEventProvider {
	return providers.NewDummyKeyboardProvider(logger)
}
