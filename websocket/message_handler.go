// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// processMessages reads and processes messages from a client
func (s *WebSocketServer) processMessages(conn *websocket.Conn) {
	for {
		_, rawMessage, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Debug("WebSocket error: %v", err)
			}
			break
		}

		// Log request if enabled
		if s.config.WebServer.LogRequests {
			s.logger.Debug("Received WebSocket message: %s", string(rawMessage))
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			s.logger.Error("Error parsing WebSocket message: %v", err)
			s.sendError(conn, "invalid_message", "Could not parse message", msg.RequestID)
			continue
		}

		// Process message based on type
		switch msg.Type {
		case "start-recording":
			go s.handleStartRecording(conn, msg.RequestID)
		case "stop-recording":
			go s.handleStopRecording(conn, msg.RequestID)
		case "ping":
			s.sendMessage(conn, "pong", nil)
		default:
			s.logger.Warning("Unknown message type: %s", msg.Type)
			s.sendError(conn, "unknown_type", fmt.Sprintf("Unknown message type: %s", msg.Type), msg.RequestID)
		}
	}
}

// handleStartRecording starts audio recording
func (s *WebSocketServer) handleStartRecording(conn *websocket.Conn, requestID string) {
	// Start recording
	err := s.executeWithRetry(func() error {
		return s.recorder.StartRecording()
	}, conn)

	if err != nil {
		s.logger.Error("Error starting recording: %v", err)
		s.sendError(conn, "recording_error", fmt.Sprintf("Error starting recording: %v", err), requestID)
		return
	}

	s.sendMessage(conn, "recording-started", nil, requestID)
}

// handleStopRecording stops recording and transcribes the audio
func (s *WebSocketServer) handleStopRecording(conn *websocket.Conn, requestID string) {
	// Stop recording
	audioFile, err := s.recorder.StopRecording()
	if err != nil {
		s.logger.Error("Error stopping recording: %v", err)
		s.sendError(conn, "recording_error", fmt.Sprintf("Error stopping recording: %v", err), requestID)
		return
	}

	// Notify client that recording has stopped
	s.sendMessage(conn, "recording-stopped", nil, requestID)

	// Send the audio file to Whisper for transcription
	type transcriptionResult struct {
		text string
		err  error
	}

	// Create channel for receiving transcription result
	resultCh := make(chan transcriptionResult, 1)

	// Start transcription in a goroutine
	go func() {
		text, err := s.whisper.Transcribe(audioFile)
		resultCh <- transcriptionResult{text: text, err: err}
	}()

	// Wait for result with timeout
	select {
	case result := <-resultCh:
		if result.err != nil {
			s.logger.Error("Error transcribing audio: %v", result.err)
			s.sendError(conn, "transcription_error", fmt.Sprintf("Error transcribing audio: %v", result.err), requestID)
		} else {
			// Success - send transcribed text
			s.sendMessage(conn, "transcription", map[string]string{
				"text": result.text,
			}, requestID)
		}
	case <-time.After(30 * time.Second):
		s.logger.Error("Timeout transcribing audio")
		s.sendError(conn, "transcription_timeout", "Timeout transcribing audio", requestID)
	}

	// Clean up audio file
	if err := s.recorder.CleanupFile(); err != nil {
		s.logger.Error("Error cleaning up audio file: %v", err)
	}
}

// sendError sends an error message to a client
func (s *WebSocketServer) sendError(conn *websocket.Conn, errorType string, errorMsg string, requestID string) {
	msg := Message{
		Type:       "error",
		Error:      errorType,
		Payload:    errorMsg,
		APIVersion: s.config.WebServer.APIVersion,
		RequestID:  requestID,
		Timestamp:  time.Now().Unix(),
	}

	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error("Error marshaling error message: %v", err)
		return
	}

	// Send message
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		s.logger.Error("SetWriteDeadline error: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		s.logger.Error("Error sending error message: %v", err)
	}
}
