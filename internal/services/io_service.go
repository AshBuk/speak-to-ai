// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	outputInterfaces "github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/websocket"
)

// IOService implements IOServiceInterface
type IOService struct {
	logger          logger.Logger
	config          *config.Config
	outputManager   outputInterfaces.Outputter
	webSocketServer *websocket.WebSocketServer
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

	// Use the method specified in config
	switch ios.config.Output.DefaultMode {
	case "clipboard":
		if err := ios.outputManager.CopyToClipboard(text); err != nil {
			ios.logger.Warning("Clipboard method failed, trying typing fallback: %v", err)
			return ios.HandleTypingFallback(text)
		}
		ios.logger.Debug("Successfully copied text to clipboard")
	case "active_window":
		if err := ios.outputManager.TypeToActiveWindow(text); err != nil {
			ios.logger.Warning("Active window method failed, trying clipboard fallback: %v", err)
			if clipErr := ios.outputManager.CopyToClipboard(text); clipErr != nil {
				return fmt.Errorf("both active window and clipboard failed - typing: %w, clipboard: %v", err, clipErr)
			}
			ios.logger.Debug("Successfully copied text to clipboard (fallback)")
		} else {
			ios.logger.Debug("Successfully typed text to active window")
		}
	default:
		// Fallback to old behavior for unknown modes
		ios.logger.Warning("Unknown output mode '%s', using clipboard with typing fallback", ios.config.Output.DefaultMode)
		if err := ios.outputManager.CopyToClipboard(text); err != nil {
			// Fallback to typing if clipboard fails
			if typeErr := ios.outputManager.TypeToActiveWindow(text); typeErr != nil {
				return fmt.Errorf("both clipboard and typing failed - clipboard: %w, typing: %v", err, typeErr)
			}
		}
		ios.logger.Debug("Successfully output text with fallback")
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

	// Method is stored in config and used by OutputText
	ios.logger.Info("Output method switched to: %s", method)
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

// HandleTypingFallback implements IOServiceInterface
func (ios *IOService) HandleTypingFallback(text string) error {
	ios.logger.Info("Handling typing fallback for: %s", text)

	// Try typing when clipboard fails
	if err := ios.OutputByTyping(text); err != nil {
		ios.logger.Error("Typing fallback also failed: %v", err)
		return fmt.Errorf("both clipboard and typing failed: %w", err)
	}

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
