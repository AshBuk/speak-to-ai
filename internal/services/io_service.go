// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	outputFactory "github.com/AshBuk/speak-to-ai/output/factory"
	outputInterfaces "github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/websocket"
)

// IOService implements IOServiceInterface
type IOService struct {
	logger          logger.Logger
	config          *config.Config
	outputManager   outputInterfaces.Outputter
	webSocketServer *websocket.WebSocketServer
	// Dependencies
	ui  UIServiceInterface
	cfg ConfigServiceInterface
}

// NewIOService creates a new IOService instance
func NewIOService(
	logger logger.Logger,
	config *config.Config,
	outputManager outputInterfaces.Outputter,
	webSocketServer *websocket.WebSocketServer,
) *IOService {
	return &IOService{
		logger:          logger,
		config:          config,
		outputManager:   outputManager,
		webSocketServer: webSocketServer,
	}
}

// OutputText implements IOServiceInterface
func (ios *IOService) OutputText(text string) error {
	if ios.outputManager == nil {
		return fmt.Errorf("output manager not available")
	}

	ios.logger.Info("Outputting text: %s", text)

	switch ios.config.Output.DefaultMode {
	case "clipboard":
		if err := ios.outputManager.CopyToClipboard(text); err != nil {
			return ios.fallbackToTyping(text, err)
		}
		ios.logger.Debug("Successfully copied text to clipboard")
	case "active_window":
		if err := ios.outputManager.TypeToActiveWindow(text); err != nil {
			return ios.fallbackToClipboard(text, err)
		}
		ios.logger.Debug("Successfully typed text to active window")
	default:
		ios.logger.Warning("Unknown output mode '%s', using typing with clipboard fallback", ios.config.Output.DefaultMode)
		if err := ios.outputManager.TypeToActiveWindow(text); err != nil {
			return ios.fallbackToClipboard(text, err)
		}
	}

	return nil
}

// OutputToClipboard outputs text to clipboard specifically
func (ios *IOService) OutputToClipboard(text string) error {
	if ios.outputManager == nil {
		return fmt.Errorf("output manager not available")
	}

	if err := ios.outputManager.CopyToClipboard(text); err != nil {
		ios.logger.Error("Failed to copy to clipboard: %v", err)
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// OutputByTyping outputs text by typing
func (ios *IOService) OutputByTyping(text string) error {
	if ios.outputManager == nil {
		return fmt.Errorf("output manager not available")
	}

	if err := ios.outputManager.TypeToActiveWindow(text); err != nil {
		ios.logger.Error("Failed to type to active window: %v", err)
		return fmt.Errorf("failed to type to active window: %w", err)
	}

	return nil
}

// SetOutputMethod implements IOServiceInterface
func (ios *IOService) SetOutputMethod(method string) error {
	ios.logger.Info("Setting output method to: %s", method)

	// Validate method
	if method != "clipboard" && method != "active_window" {
		return fmt.Errorf("invalid output method: %s (must be 'clipboard' or 'active_window')", method)
	}

	// Persist via ConfigService if available
	if ios.cfg != nil {
		if err := ios.cfg.UpdateOutputMode(method); err != nil {
			return err
		}
	} else {
		ios.config.Output.DefaultMode = method
	}

	// Recreate output manager immediately to reflect new mode
	env := ios.detectOutputEnvironment()
	out, err := outputFactory.GetOutputterFromConfig(ios.config, env)
	if err != nil {
		return fmt.Errorf("failed to reinitialize output manager: %w", err)
	}
	ios.outputManager = out

	// Notify UI to refresh settings display
	if ios.ui != nil {
		ios.ui.UpdateSettings(ios.config)
	}

	ios.logger.Info("Output method set to: %s", method)
	return nil
}

// StartWebSocketServer implements IOServiceInterface
func (ios *IOService) StartWebSocketServer() error {
	if ios.webSocketServer == nil {
		return fmt.Errorf("WebSocket server not configured")
	}

	ios.logger.Info("Starting WebSocket server...")

	// Start the WebSocket server
	return ios.webSocketServer.Start()
}

// StopWebSocketServer implements IOServiceInterface
func (ios *IOService) StopWebSocketServer() error {
	if ios.webSocketServer == nil {
		return nil
	}

	ios.logger.Info("Stopping WebSocket server...")

	// Stop the WebSocket server
	ios.webSocketServer.Stop()
	return nil
}

// GetOutputToolNames returns the actual tool names being used
func (ios *IOService) GetOutputToolNames() (clipboardTool, typeTool string) {
	if ios.outputManager == nil {
		return "unknown", "unknown"
	}
	return ios.outputManager.GetToolNames()
}

// Shutdown implements IOServiceInterface
func (ios *IOService) Shutdown() error {
	var lastErr error

	// Stop WebSocket server
	if err := ios.StopWebSocketServer(); err != nil {
		ios.logger.Error("Error stopping WebSocket server: %v", err)
		lastErr = err
	}

	ios.logger.Info("IOService shutdown complete")
	return lastErr
}

// detectOutputEnvironment maps platform environment to output factory environment type
func (ios *IOService) detectOutputEnvironment() outputFactory.EnvironmentType {
	switch platform.DetectEnvironment() {
	case platform.EnvironmentWayland:
		return outputFactory.EnvironmentWayland
	case platform.EnvironmentX11:
		return outputFactory.EnvironmentX11
	default:
		return outputFactory.EnvironmentUnknown
	}
}

// SetUIService sets UI service dependency
func (ios *IOService) SetUIService(ui UIServiceInterface) { ios.ui = ui }

// SetConfigService sets Config service dependency
func (ios *IOService) SetConfigService(cfg ConfigServiceInterface) { ios.cfg = cfg }

// fallbackToTyping switches to typing mode and executes the output
func (ios *IOService) fallbackToTyping(text string, originalErr error) error {
	ios.logger.Warning("Clipboard method failed: %v. Switching to active_window mode.", originalErr)

	if err := ios.switchOutputMode(config.OutputModeActiveWindow, "clipboard failed"); err != nil {
		return fmt.Errorf("failed to switch output mode: %w", err)
	}

	if err := ios.outputManager.TypeToActiveWindow(text); err != nil {
		return fmt.Errorf("both clipboard and typing failed - clipboard: %w, typing: %v", originalErr, err)
	}

	return nil
}

// fallbackToClipboard switches to clipboard mode and executes the output
func (ios *IOService) fallbackToClipboard(text string, originalErr error) error {
	ios.logger.Warning("Active window method failed: %v. Switching to clipboard mode.", originalErr)

	if err := ios.switchOutputMode(config.OutputModeClipboard, "typing failed"); err != nil {
		return fmt.Errorf("failed to switch output mode: %w", err)
	}

	if err := ios.outputManager.CopyToClipboard(text); err != nil {
		return fmt.Errorf("both typing and clipboard failed - typing: %w, clipboard: %v", originalErr, err)
	}

	ios.logger.Debug("Successfully copied text to clipboard (after switch)")
	return nil
}

// switchOutputMode switches the output mode and persists it in the config
func (ios *IOService) switchOutputMode(mode string, reason string) error {
	ios.logger.Warning("Switching output mode to '%s' due to '%s'", mode, reason)

	// Update and persist config if service is available
	if ios.cfg != nil {
		if err := ios.cfg.UpdateOutputMode(mode); err != nil {
			ios.logger.Error("Failed to persist output mode: %v", err)
			return fmt.Errorf("failed to persist output mode: %w", err)
		}
	} else {
		old := ios.config.Output.DefaultMode
		ios.config.Output.DefaultMode = mode
		ios.logger.Warning("ConfigService not available; switched output mode in memory from %s to %s", old, mode)
	}

	// Recreate output manager for the new mode based on current environment
	env := ios.detectOutputEnvironment()
	out, err := outputFactory.GetOutputterFromConfig(ios.config, env)
	if err != nil {
		ios.logger.Error("Failed to reinitialize output manager after mode switch: %v", err)
		return fmt.Errorf("failed to reinitialize output manager after mode switch: %w", err)
	}
	ios.outputManager = out

	// Notify user
	if ios.ui != nil {
		msg := fmt.Sprintf("Switched to %s due to %s", mode, reason)
		ios.ui.ShowNotification("Output Fallback", msg)
		ios.ui.UpdateSettings(ios.config)
	}

	ios.logger.Info("Output mode switched to '%s' and persisted", mode)
	return nil
}
