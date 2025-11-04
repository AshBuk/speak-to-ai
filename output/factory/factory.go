// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package factory

import (
	"fmt"
	"os/exec"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/output/outputters"
)

// Defines the display server type
type EnvironmentType string

const (
	// EnvironmentX11 represents X11 display server
	EnvironmentX11 EnvironmentType = "X11"
	// EnvironmentWayland represents Wayland display server
	EnvironmentWayland EnvironmentType = "Wayland"
	// EnvironmentUnknown represents an unknown display server
	EnvironmentUnknown EnvironmentType = "Unknown"
)

// Creates output managers based on environment and configuration
type Factory struct {
	config *config.Config
}

// Create a new output factory
func NewFactory(config *config.Config) *Factory {
	return &Factory{
		config: config,
	}
}

// selectClipboardTool chooses an appropriate clipboard tool based on environment
func (f *Factory) selectClipboardTool(env EnvironmentType) string {
	clipboardTool := f.config.Output.ClipboardTool
	if clipboardTool != "auto" {
		return clipboardTool
	}

	switch env {
	case EnvironmentWayland:
		return "wl-copy"
	case EnvironmentX11:
		return "xsel"
	default:
		return "xsel"
	}
}

// selectTypeTool chooses an appropriate typing tool based on environment
func (f *Factory) selectTypeTool(env EnvironmentType) string {
	typeTool := f.config.Output.TypeTool
	if typeTool != "auto" {
		return typeTool
	}

	switch env {
	case EnvironmentWayland:
		return f.selectWaylandTypeTool()
	case EnvironmentX11:
		return "xdotool"
	default:
		return f.selectFallbackTypeTool()
	}
}

// selectWaylandTypeTool selects the best typing tool for Wayland
func (f *Factory) selectWaylandTypeTool() string {
	if platform.IsGNOMEWithWayland() {
		return f.selectGNOMEWaylandTypeTool()
	}
	return f.selectNonGNOMEWaylandTypeTool()
}

// selectGNOMEWaylandTypeTool selects typing tool for GNOME/Wayland
func (f *Factory) selectGNOMEWaylandTypeTool() string {
	if f.isToolAvailable("ydotool") {
		return "ydotool"
	}
	if f.isToolAvailable("wtype") {
		return "wtype"
	}
	return "xdotool"
}

// selectNonGNOMEWaylandTypeTool selects typing tool for non-GNOME Wayland
func (f *Factory) selectNonGNOMEWaylandTypeTool() string {
	if f.isToolAvailable("wtype") {
		return "wtype"
	}
	if f.isToolAvailable("ydotool") {
		return "ydotool"
	}
	return "xdotool"
}

// selectFallbackTypeTool auto-detects the best available typing tool
func (f *Factory) selectFallbackTypeTool() string {
	tools := []string{"xdotool", "wtype", "ydotool"}
	for _, tool := range tools {
		if f.isToolAvailable(tool) {
			return tool
		}
	}
	return "xdotool"
}

// Create an appropriate outputter based on the environment
func (f *Factory) GetOutputter(env EnvironmentType) (interfaces.Outputter, error) {
	clipboardTool := f.selectClipboardTool(env)
	typeTool := f.selectTypeTool(env)

	// Security: Validate selected tool commands against the allowlist
	if clipboardTool != "" && !config.IsCommandAllowed(f.config, clipboardTool) {
		return nil, fmt.Errorf("clipboard tool not allowed: %s", clipboardTool)
	}
	if typeTool != "" && !config.IsCommandAllowed(f.config, typeTool) {
		return nil, fmt.Errorf("type tool not allowed: %s", typeTool)
	}

	// Create the appropriate outputter based on the default mode
	switch f.config.Output.DefaultMode {
	case config.OutputModeClipboard:
		return outputters.NewClipboardOutputter(clipboardTool, f.config)
	case config.OutputModeActiveWindow:
		return outputters.NewTypeOutputter(typeTool, f.config)
	default:
		return outputters.NewClipboardOutputter(clipboardTool, f.config)
	}
}

// Create an outputter directly from a configuration
func GetOutputterFromConfig(config *config.Config, env EnvironmentType) (interfaces.Outputter, error) {
	factory := NewFactory(config)
	return factory.GetOutputter(env)
}

// Check if a command-line tool is available in the system's PATH
func (f *Factory) isToolAvailable(toolName string) bool {
	_, err := exec.LookPath(toolName)
	return err == nil
}
