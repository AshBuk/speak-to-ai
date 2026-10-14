// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package factory

import (
	"testing"

	"github.com/AshBuk/dabri/v2/config"
)

func TestIsToolAvailable(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	factory := NewFactory(cfg)

	// Test with a tool that should always be available
	available := factory.isToolAvailable("sh")
	if !available {
		t.Error("sh should be available on Linux systems")
	}

	// Test with a tool that likely doesn't exist
	notAvailable := factory.isToolAvailable("non-existent-tool-12345")
	if notAvailable {
		t.Error("non-existent tool should not be available")
	}
}

func TestSecurityAllowedCommands(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	// Test that new Wayland tools are in allowed commands
	waylandTools := []string{"wtype", "ydotool"}

	for _, tool := range waylandTools {
		if !config.IsCommandAllowed(cfg, tool) {
			t.Errorf("Wayland tool %s should be in allowed commands list", tool)
		}
	}

	// Test that traditional tools are still allowed
	traditionalTools := []string{"xdotool", "xsel", "wl-copy"}

	for _, tool := range traditionalTools {
		if !config.IsCommandAllowed(cfg, tool) {
			t.Errorf("Traditional tool %s should be in allowed commands list", tool)
		}
	}
}
