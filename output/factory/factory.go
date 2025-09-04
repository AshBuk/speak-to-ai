// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package factory

import (
	"fmt"
	"os/exec"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/output/outputters"
)

// EnvironmentType represents the display server type
type EnvironmentType string

const (
	// EnvironmentX11 represents X11 display server
	EnvironmentX11 EnvironmentType = "X11"
	// EnvironmentWayland represents Wayland display server
	EnvironmentWayland EnvironmentType = "Wayland"
	// EnvironmentUnknown represents unknown display server
	EnvironmentUnknown EnvironmentType = "Unknown"
)

// Factory creates output managers based on environment and configuration
type Factory struct {
	config *config.Config
}

// NewFactory creates a new output factory
func NewFactory(config *config.Config) *Factory {
	return &Factory{
		config: config,
	}
}

// GetOutputter creates an appropriate outputter based on environment
func (f *Factory) GetOutputter(env EnvironmentType) (interfaces.Outputter, error) {
	// Choose clipboard tool based on environment
	clipboardTool := f.config.Output.ClipboardTool
	if clipboardTool == "auto" {
		switch env {
		case EnvironmentWayland:
			clipboardTool = "wl-copy"
		case EnvironmentX11:
			clipboardTool = "xclip"
		default:
			clipboardTool = "xclip" // Default to xclip
		}
	}

	// Choose type tool based on environment
	typeTool := f.config.Output.TypeTool
	if typeTool == "auto" {
		switch env {
		case EnvironmentWayland:
			// Try Wayland-compatible tools in order of preference
			if f.isToolAvailable("wtype") {
				typeTool = "wtype"
			} else if f.isToolAvailable("ydotool") {
				typeTool = "ydotool"
			} else {
				// Fallback: try xdotool (might work with XWayland)
				typeTool = "xdotool"
			}
		case EnvironmentX11:
			typeTool = "xdotool"
		default:
			// Auto-detect best available tool
			if f.isToolAvailable("xdotool") {
				typeTool = "xdotool"
			} else if f.isToolAvailable("wtype") {
				typeTool = "wtype"
			} else if f.isToolAvailable("ydotool") {
				typeTool = "ydotool"
			} else {
				typeTool = "xdotool" // Default fallback
			}
		}
	}

	// Security: Validate selected tool commands against allowlist
	if clipboardTool != "" && !config.IsCommandAllowed(f.config, clipboardTool) {
		return nil, fmt.Errorf("clipboard tool not allowed: %s", clipboardTool)
	}
	if typeTool != "" && !config.IsCommandAllowed(f.config, typeTool) {
		return nil, fmt.Errorf("type tool not allowed: %s", typeTool)
	}

	// Create appropriate outputter
	switch f.config.Output.DefaultMode {
	case config.OutputModeClipboard:
		return outputters.NewClipboardOutputter(clipboardTool, f.config)
	case config.OutputModeActiveWindow:
		// Use combined outputter even for ActiveWindow to enable seamless clipboard fallback
		// on compositors where typing tools are unsupported (e.g., GNOME on Wayland)
		return outputters.NewCombinedOutputter(clipboardTool, typeTool, f.config)
	case config.OutputModeCombined:
		return outputters.NewCombinedOutputter(clipboardTool, typeTool, f.config)
	default:
		return outputters.NewCombinedOutputter(clipboardTool, typeTool, f.config)
	}
}

// GetOutputterFromConfig is a convenience function to create an outputter directly from config
func GetOutputterFromConfig(config *config.Config, env EnvironmentType) (interfaces.Outputter, error) {
	factory := NewFactory(config)
	return factory.GetOutputter(env)
}

// isToolAvailable checks if a command-line tool is available in PATH
func (f *Factory) isToolAvailable(toolName string) bool {
	_, err := exec.LookPath(toolName)
	return err == nil
}
