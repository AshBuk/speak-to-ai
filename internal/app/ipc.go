// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/AshBuk/speak-to-ai/internal/ipc"
	"github.com/AshBuk/speak-to-ai/internal/services"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

// IPC Handlers - Protocol adapter between Unix socket IPC and Business Services
// Request-response IPC commands (sync, wait-for-result with timeout)
//
// Architecture Flow:
//   CLI client (speak-to-ai start-recording)
//       ↓
//   Unix Socket (utils.GetDefaultSocketPath())
//       ↓
//   IPC Server (internal/ipc package)
//       ↓
//   IPC Handlers (this file) ← protocol adapter layer
//       ↓
//   Handler methods (handlers.go) ← business logic adapter
//       ↓
//   Business Services (AudioService/IOService)

const (
	ipcTranscriptionTimeout = 45 * time.Second
)

// startIPCServer Initializes Unix socket IPC server for CLI client communication
// Creates server on default socket path, registers command handlers, starts listening
func (a *App) startIPCServer() error {
	if a.Services == nil {
		return fmt.Errorf("services not initialized")
	}
	socketPath := utils.GetDefaultSocketPath()
	server := ipc.NewServer(socketPath, a.Runtime.Logger)
	a.registerIPCHandlers(server)
	if err := server.Start(); err != nil {
		return err
	}
	a.ipcServer = server
	return nil
}

// registerIPCHandlers Registers 4 IPC command handlers with server
// Maps command names to handler functions for request routing
func (a *App) registerIPCHandlers(server *ipc.Server) {
	server.Register("start-recording", a.ipcHandleStartRecording)
	server.Register("stop-recording", a.ipcHandleStopRecording)
	server.Register("status", a.ipcHandleStatus)
	server.Register("last-transcript", a.ipcHandleLastTranscript)
}

// ipcHandleStartRecording Command handler - starts audio recording via handlers.go
// Returns JSON response with recording state
func (a *App) ipcHandleStartRecording(ipc.Request) (ipc.Response, error) {
	if err := a.handleStartRecording(); err != nil {
		return ipc.Response{}, err
	}
	return ipc.NewSuccessResponse("recording started", map[string]any{
		"recording": true,
	}), nil
}

// ipcHandleStopRecording Command handler - stops recording and waits for transcript
// Synchronous: blocks CLI until transcription ready (45s timeout) or returns last transcript
// Graceful errors: ErrNoRecordingInProgress returns success (not error) for idempotency
func (a *App) ipcHandleStopRecording(ipc.Request) (ipc.Response, error) {
	if err := a.handleStopRecordingAndTranscribe(); err != nil {
		if errors.Is(err, services.ErrNoRecordingInProgress) {
			return ipc.NewSuccessResponse("recording already stopped", map[string]any{
				"recording":  false,
				"transcript": "",
			}), nil
		}
		return ipc.Response{}, err
	}
	transcript, err := a.waitForTranscription(ipcTranscriptionTimeout)
	if err != nil {
		return ipc.NewSuccessResponse("recording stopped", map[string]any{
			"recording":  false,
			"transcript": a.getLastTranscript(),
			"warning":    err.Error(),
		}), nil
	}
	return ipc.NewSuccessResponse("recording stopped", map[string]any{
		"recording":  false,
		"transcript": transcript,
	}), nil
}

// ipcHandleStatus Command handler - returns current recording state + last transcript
// Non-blocking: returns immediately without waiting for transcription
func (a *App) ipcHandleStatus(ipc.Request) (ipc.Response, error) {
	recording := false
	if a.Services != nil && a.Services.Audio != nil {
		recording = a.Services.Audio.IsRecording()
	}
	return ipc.NewSuccessResponse("status", map[string]any{
		"recording":       recording,
		"last_transcript": a.getLastTranscript(),
	}), nil
}

// ipcHandleLastTranscript Command handler - returns last saved transcript
// Non-blocking: returns cached transcript without waiting
func (a *App) ipcHandleLastTranscript(ipc.Request) (ipc.Response, error) {
	return ipc.NewSuccessResponse("last transcript", map[string]any{
		"transcript": a.getLastTranscript(),
	}), nil
}

// waitForTranscription Blocks until transcription ready or timeout expires
// Used by ipcHandleStopRecording to provide synchronous CLI response
func (a *App) waitForTranscription(timeout time.Duration) (string, error) {
	if a.Services == nil || a.Services.IO == nil {
		return "", fmt.Errorf("io service not available")
	}
	return a.Services.IO.WaitForTranscription(timeout)
}

// getLastTranscript Returns cached transcript without waiting
// Used by status/last-transcript commands for immediate response
func (a *App) getLastTranscript() string {
	if a.Services == nil || a.Services.Audio == nil {
		return ""
	}
	return a.Services.Audio.GetLastTranscript()
}
