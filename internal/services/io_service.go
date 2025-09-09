// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/internal/logger"
	outputInterfaces "github.com/AshBuk/speak-to-ai/output/interfaces"
	"github.com/AshBuk/speak-to-ai/websocket"
)

// IOService implements IOServiceInterface
type IOService struct {
	logger          logger.Logger
	outputManager   outputInterfaces.Outputter
	webSocketServer *websocket.WebSocketServer
}

// NewIOService creates a new IOService instance
func NewIOService(
	logger logger.Logger,
	outputManager outputInterfaces.Outputter,
	webSocketServer *websocket.WebSocketServer,
) *IOService {
	return &IOService{
		logger:          logger,
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

	// Use the appropriate output method based on configuration
	// For now, try clipboard first, then fallback to typing
	if err := ios.outputManager.CopyToClipboard(text); err != nil {
		// Fallback to typing if clipboard fails
		if typeErr := ios.outputManager.TypeToActiveWindow(text); typeErr != nil {
			return fmt.Errorf("both clipboard and typing failed - clipboard: %w, typing: %v", err, typeErr)
		}
	} else {
		return nil
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

	// TODO: Implement output method switching logic
	// This will depend on how output manager handles method switching

	return nil
}

// BroadcastTranscription implements IOServiceInterface
func (ios *IOService) BroadcastTranscription(text string, isFinal bool) {
	if ios.webSocketServer == nil {
		ios.logger.Debug("WebSocket server not available, skipping broadcast")
		return
	}

	ios.logger.Debug("Broadcasting transcription: %s (final: %t)", text, isFinal)

	// Send transcription update via WebSocket
	payload := map[string]interface{}{
		"text":    text,
		"isFinal": isFinal,
	}
	ios.webSocketServer.BroadcastMessage("transcription", payload)
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
