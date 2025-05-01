package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// TextOutputter interface for text output methods
type TextOutputter interface {
	CopyToClipboard(text string) error
	TypeToActiveWindow(text string) error
}

// Environment represents the display server type
type Environment string

const (
	// X11 display server
	X11 Environment = "x11"
	// Wayland display server
	Wayland Environment = "wayland"
	// Unknown environment
	UnknownEnv Environment = "unknown"
)

// OutputHelper provides common functionality for running commands
type OutputHelper struct {
	// Common fields could be added here if needed
}

// RunCommand executes a command with standardized error handling
func (h *OutputHelper) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// RunCommandWithInput executes a command with stdin input
func (h *OutputHelper) RunCommandWithInput(input string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	return cmd.Run()
}

// CheckCommandExists verifies a command is available
func (h *OutputHelper) CheckCommandExists(cmd string) (bool, error) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return false, fmt.Errorf("%s not found: %w", cmd, err)
	}
	return true, nil
}

// LinuxOutputter implements TextOutputter for Linux
type LinuxOutputter struct {
	config *Config
	env    Environment
	helper OutputHelper
}

// NewLinuxOutputter creates a new instance of LinuxOutputter
func NewLinuxOutputter(config *Config) *LinuxOutputter {
	env := DetectEnvironment()
	log.Printf("Detected environment: %s", env)
	return &LinuxOutputter{
		config: config,
		env:    env,
		helper: OutputHelper{},
	}
}

// CopyToClipboard copies text to clipboard
func (o *LinuxOutputter) CopyToClipboard(text string) error {
	// Determine command based on settings or auto-detection
	clipboardTool := o.config.Output.ClipboardTool
	if clipboardTool == "auto" {
		if o.env == Wayland {
			clipboardTool = "wl-copy"
		} else {
			clipboardTool = "xclip"
		}
	}

	log.Printf("Using clipboard tool: %s", clipboardTool)

	switch clipboardTool {
	case "wl-copy":
		exists, err := o.helper.CheckCommandExists("wl-copy")
		if !exists {
			return err
		}
		return o.helper.RunCommand("wl-copy", text)
	case "xclip":
		exists, err := o.helper.CheckCommandExists("xclip")
		if !exists {
			return err
		}
		return o.helper.RunCommandWithInput(text, "xclip", "-selection", "clipboard", "-i")
	default:
		return fmt.Errorf("unsupported clipboard tool: %s", clipboardTool)
	}
}

// TypeToActiveWindow types text in the active window
func (o *LinuxOutputter) TypeToActiveWindow(text string) error {
	// Determine command based on settings or auto-detection
	typeTool := o.config.Output.TypeTool
	if typeTool == "auto" {
		if o.env == Wayland {
			typeTool = "wl-clipboard"
		} else {
			typeTool = "xdotool"
		}
	}

	log.Printf("Using type tool: %s", typeTool)

	switch typeTool {
	case "xdotool":
		exists, err := o.helper.CheckCommandExists("xdotool")
		if !exists {
			return err
		}
		return o.helper.RunCommand("xdotool", "type", "--clearmodifiers", text)
	case "wl-clipboard":
		// Check for required tools
		exists, err := o.helper.CheckCommandExists("wl-copy")
		if !exists {
			return err
		}
		exists, err = o.helper.CheckCommandExists("wtype")
		if !exists {
			return err
		}

		// Two step process: copy to clipboard then simulate paste
		if err := o.CopyToClipboard(text); err != nil {
			return err
		}

		// Simulate Ctrl+V with wtype
		return o.helper.RunCommand("sh", "-c", "sleep 0.5 && wtype -M ctrl -k v -m ctrl")
	case "dbus":
		// Some applications can use D-Bus
		return fmt.Errorf("dbus input method not fully implemented")
	default:
		return fmt.Errorf("unsupported type tool: %s", typeTool)
	}
}

// DetectEnvironment detects the current display server environment
func DetectEnvironment() Environment {
	// Check XDG_SESSION_TYPE environment variable
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	if sessionType != "" {
		sessionType = strings.ToLower(sessionType)
		if sessionType == "wayland" {
			return Wayland
		} else if sessionType == "x11" {
			return X11
		}
	}

	// Check WAYLAND_DISPLAY environment variable
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return Wayland
	}

	// Check if running under X11
	if os.Getenv("DISPLAY") != "" {
		// Try to run xprop to confirm X11
		helper := OutputHelper{}
		if _, err := helper.CheckCommandExists("xprop"); err == nil {
			if helper.RunCommand("xprop", "-root") == nil {
				return X11
			}
		}
	}

	// Try to detect using common tools
	helper := OutputHelper{}
	if _, err := helper.CheckCommandExists("wayland-info"); err == nil {
		return Wayland
	}

	if _, err := helper.CheckCommandExists("xdpyinfo"); err == nil {
		if helper.RunCommand("xdpyinfo") == nil {
			return X11
		}
	}

	log.Println("Warning: Could not determine display server environment, defaulting to X11")
	return X11
}

// GetOutputter returns the appropriate output handler based on OS
func GetOutputter(config *Config) (TextOutputter, error) {
	switch runtime.GOOS {
	case "linux":
		return NewLinuxOutputter(config), nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
