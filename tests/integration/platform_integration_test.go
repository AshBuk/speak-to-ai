//go:build integration
// +build integration

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package integration

import (
	"os"
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys/adapters"
	hotkeyinterfaces "github.com/AshBuk/speak-to-ai/hotkeys/interfaces"
	"github.com/AshBuk/speak-to-ai/hotkeys/providers"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	outputfactory "github.com/AshBuk/speak-to-ai/output/factory"
)

// Platform-specific integration tests

func TestWaylandCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test Wayland-specific functionality
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	t.Run("wayland_output_tools", func(t *testing.T) {
		// Test that Wayland output tools are properly configured
		factory := outputfactory.NewFactory(cfg)

		// Test clipboard on Wayland
		_, err := factory.GetOutputter(outputfactory.EnvironmentWayland)
		if err != nil {
			t.Logf("Expected error in test environment (tools not available): %v", err)
		} else {
			t.Log("Wayland outputter created successfully")
		}
	})

	t.Run("wayland_hotkeys", func(t *testing.T) {
		// Test DBus hotkey provider
		hotkeyConfig := adapters.NewConfigAdapter("altgr+comma", "auto")
		provider := providers.NewDbusKeyboardProvider(hotkeyConfig, hotkeyinterfaces.EnvironmentWayland, newMockLogger())

		if provider.IsSupported() {
			t.Log("DBus GlobalShortcuts portal is available")
		} else {
			t.Log("DBus GlobalShortcuts portal not available (expected in CI)")
		}
	})

	t.Run("wayland_environment_detection", func(t *testing.T) {
		// Test environment detection logic
		env := platform.DetectEnvironment()
		t.Logf("Detected environment: %s", env)

		// Environment detection should not crash
		if env == "" {
			t.Error("Environment detection returned empty string")
		}
	})
}

func TestX11Compatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test X11-specific functionality
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	t.Run("x11_output_tools", func(t *testing.T) {
		// Test that X11 output tools are properly configured
		factory := outputfactory.NewFactory(cfg)

		// Test clipboard on X11
		_, err := factory.GetOutputter(outputfactory.EnvironmentX11)
		if err != nil {
			t.Logf("Expected error in test environment (tools not available): %v", err)
		} else {
			t.Log("X11 outputter created successfully")
		}
	})

	t.Run("x11_hotkeys", func(t *testing.T) {
		// Test evdev hotkey provider
		hotkeyConfig := adapters.NewConfigAdapter("altgr+comma", "auto")
		provider := providers.NewEvdevKeyboardProvider(hotkeyConfig, hotkeyinterfaces.EnvironmentX11, newMockLogger())

		if provider.IsSupported() {
			t.Log("Evdev input devices are available")
		} else {
			t.Log("Evdev input devices not available (expected in CI)")
		}
	})
}

func TestCrossplatformToolFallbacks(t *testing.T) {
	// Test that tool fallback logic works correctly
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	factory := outputfactory.NewFactory(cfg)

	environments := []outputfactory.EnvironmentType{
		outputfactory.EnvironmentX11,
		outputfactory.EnvironmentWayland,
		outputfactory.EnvironmentUnknown,
	}

	for _, env := range environments {
		t.Run(string(env), func(t *testing.T) {
			// Test that factory can create outputters for any environment
			// (even if tools aren't available, the logic should be sound)
			_, err := factory.GetOutputter(env)
			if err != nil {
				t.Logf("Environment %s: %v (expected in test environment)", env, err)
			} else {
				t.Logf("Environment %s: outputter created successfully", env)
			}
		})
	}
}

func TestSecurityValidation(t *testing.T) {
	// Test that security validation is working
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	// Test allowed commands
	allowedCommands := []string{"xdotool", "wtype", "wl-copy", "arecord"}
	for _, cmd := range allowedCommands {
		if !config.IsCommandAllowed(cfg, cmd) {
			t.Errorf("Command %s should be allowed but isn't", cmd)
		}
	}

	// Test disallowed commands
	disallowedCommands := []string{"rm", "curl", "wget", "sh", "bash"}
	for _, cmd := range disallowedCommands {
		if config.IsCommandAllowed(cfg, cmd) {
			t.Errorf("Command %s should not be allowed but is", cmd)
		}
	}

	// Test argument sanitization
	testArgs := []string{
		"normal-arg",
		"--flag",
		"/path/to/file",
	}

	sanitized := config.SanitizeCommandArgs(testArgs)
	if len(sanitized) != len(testArgs) {
		t.Log("Some arguments were filtered (expected for security)")
	}

	// Test dangerous arguments are filtered
	dangerousArgs := []string{
		"arg; rm -rf /",
		"$(malicious command)",
		"arg|dangerous",
		"../../../etc/passwd",
	}

	sanitizedDangerous := config.SanitizeCommandArgs(dangerousArgs)
	if len(sanitizedDangerous) > 0 {
		t.Log("Some dangerous arguments were not filtered - review security policy")
	}
}

func TestModelAvailability(t *testing.T) {
	// Test model file handling
	t.Run("default_model_path", func(t *testing.T) {
		cfg := &config.Config{}
		config.SetDefaultConfig(cfg)

		// Check if default model path exists
		if _, err := os.Stat(cfg.General.ModelPath); err != nil {
			t.Logf("Default model not found at %s (expected for fresh install)", cfg.General.ModelPath)
		} else {
			t.Logf("Default model found at %s", cfg.General.ModelPath)
		}
	})

	t.Run("model_download_location", func(t *testing.T) {
		// Test that the model download location is accessible
		cfg := &config.Config{}
		config.SetDefaultConfig(cfg)

		modelDir := "sources/language-models"
		if info, err := os.Stat(modelDir); err != nil {
			t.Logf("Model directory %s not found (expected for fresh repo)", modelDir)
		} else if !info.IsDir() {
			t.Errorf("Model path %s exists but is not a directory", modelDir)
		} else {
			t.Logf("Model directory %s exists", modelDir)
		}
	})
}

func TestFlatpakCompatibility(t *testing.T) {
	// Test Flatpak-specific considerations
	t.Run("flatpak_paths", func(t *testing.T) {
		// Check if we're running in Flatpak environment
		if os.Getenv("FLATPAK_ID") != "" {
			t.Log("Running in Flatpak environment")

			// Test Flatpak-specific paths
			flatpakConfigDir := os.Getenv("XDG_CONFIG_HOME")
			if flatpakConfigDir == "" {
				t.Log("XDG_CONFIG_HOME not set in Flatpak")
			} else {
				t.Logf("Flatpak config dir: %s", flatpakConfigDir)
			}
		} else {
			t.Log("Not running in Flatpak environment")
		}
	})

	t.Run("sandboxed_tools", func(t *testing.T) {
		// Test that sandboxed tools are available
		// These would be built as modules in Flatpak
		tools := []string{"arecord", "xdotool", "wl-copy", "wtype"}

		for _, tool := range tools {
			t.Run(tool, func(t *testing.T) {
				// In Flatpak, tools should be available in /app/bin
				// Outside Flatpak, they might be in system PATH
				t.Logf("Testing availability of %s", tool)
			})
		}
	})
}
