// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package testutils

import "github.com/AshBuk/dabri/v2/config"

// StaticConfig returns a config getter that always yields cfg.
func StaticConfig(cfg *config.Config) func() *config.Config {
	return func() *config.Config { return cfg }
}
