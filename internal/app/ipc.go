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

const (
	ipcTranscriptionTimeout = 45 * time.Second
)

// startIPCServer initializes the IPC server used by the CLI helper.
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

func (a *App) registerIPCHandlers(server *ipc.Server) {
	server.Register("start-recording", a.ipcHandleStartRecording)
	server.Register("stop-recording", a.ipcHandleStopRecording)
	server.Register("status", a.ipcHandleStatus)
	server.Register("last-transcript", a.ipcHandleLastTranscript)
}

func (a *App) ipcHandleStartRecording(ipc.Request) (ipc.Response, error) {
	if err := a.handleStartRecording(); err != nil {
		return ipc.Response{}, err
	}
	return ipc.NewSuccessResponse("recording started", map[string]any{
		"recording": true,
	}), nil
}

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
		// Still mark as success, but propagate warning to caller.
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

func (a *App) ipcHandleLastTranscript(ipc.Request) (ipc.Response, error) {
	return ipc.NewSuccessResponse("last transcript", map[string]any{
		"transcript": a.getLastTranscript(),
	}), nil
}

func (a *App) waitForTranscription(timeout time.Duration) (string, error) {
	if a.Services == nil || a.Services.IO == nil {
		return "", fmt.Errorf("io service not available")
	}
	if ioSvc, ok := a.Services.IO.(*services.IOService); ok && ioSvc != nil {
		return ioSvc.WaitForTranscription(timeout)
	}

	return "", fmt.Errorf("io service does not support transcription wait")
}

func (a *App) getLastTranscript() string {
	if a.Services == nil || a.Services.Audio == nil {
		return ""
	}
	if audioSvc, ok := a.Services.Audio.(*services.AudioService); ok && audioSvc != nil {
		return audioSvc.GetLastTranscript()
	}

	return ""
}
