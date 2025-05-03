package output

import (
	"github.com/AshBuk/speak-to-ai/config"
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
func (f *Factory) GetOutputter(env EnvironmentType) (Outputter, error) {
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
			typeTool = "wl-keyboard" // Placeholder, might be custom implementation
		case EnvironmentX11:
			typeTool = "xdotool"
		default:
			typeTool = "xdotool" // Default to xdotool
		}
	}

	// Create appropriate outputter
	switch f.config.Output.DefaultMode {
	case "clipboard":
		return NewClipboardOutputter(clipboardTool)
	case "active_window":
		return NewTypeOutputter(typeTool)
	case "combined":
		return NewCombinedOutputter(clipboardTool, typeTool)
	default:
		return NewCombinedOutputter(clipboardTool, typeTool)
	}
}

// GetOutputterFromConfig is a convenience function to create an outputter directly from config
func GetOutputterFromConfig(config *config.Config, env EnvironmentType) (Outputter, error) {
	factory := NewFactory(config)
	return factory.GetOutputter(env)
}
