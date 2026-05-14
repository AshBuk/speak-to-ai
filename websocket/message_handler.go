// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// Handle client requests with message validation and routing
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
		// Synchronous handling per-connection provides natural backpressure
		switch msg.Type {
		case "start-recording":
			s.handleStartRecording(conn, msg.RequestID)
		case "stop-recording":
			s.handleStopRecording(conn, msg.RequestID)
		case "ping":
			s.sendMessage(conn, "pong", nil)
		default:
			s.logger.Warning("Unknown message type: %s", msg.Type)
			s.sendError(conn, "unknown_type", fmt.Sprintf("Unknown message type: %s", msg.Type), msg.RequestID)
		}
	}
}

// Start recording through AudioService.
func (s *WebSocketServer) handleStartRecording(conn *websocket.Conn, requestID string) {
	audio := s.audioController()
	if audio == nil {
		s.sendError(conn, "recording_error", "audio controller not wired", requestID)
		return
	}
	err := s.executeWithRetry(audio.HandleStartRecording, conn)
	if err != nil {
		s.logger.Error("Error starting recording: %v", err)
		s.sendError(conn, "recording_error", fmt.Sprintf("Error starting recording: %v", err), requestID)
		return
	}
	s.sendMessage(conn, "recording-started", nil, requestID)
}

// Stop recording through AudioService and return the transcript.
func (s *WebSocketServer) handleStopRecording(conn *websocket.Conn, requestID string) {
	audio := s.audioController()
	if audio == nil {
		s.sendError(conn, "recording_error", "audio controller not wired", requestID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), transcriptionCtxTimeout)
	defer cancel()
	text, err := audio.HandleStopRecordingSync(ctx)
	if err != nil {
		if ctx.Err() != nil {
			s.logger.Error("Timeout transcribing audio")
			s.sendError(conn, "transcription_timeout", "Timeout transcribing audio", requestID)
			return
		}
		s.logger.Error("Error stopping/transcribing: %v", err)
		s.sendError(conn, "transcription_error", fmt.Sprintf("Error transcribing audio: %v", err), requestID)
		return
	}
	// Confirm stop after the operation actually succeeded, then deliver
	s.sendMessage(conn, "recording-stopped", nil, requestID)
	s.sendMessage(conn, "transcription", map[string]string{
		"text": text,
	}, requestID)
}

// Deliver structured error response for client debugging
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
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		s.logger.Error("SetWriteDeadline error: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		s.logger.Error("Error sending error message: %v", err)
	}
}
