// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/interfaces"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/whisper"
	"github.com/gorilla/websocket"
)

// WebSocketServer represents a WebSocket server
type WebSocketServer struct {
	config      *config.Config
	clients     map[*websocket.Conn]bool
	clientsLock sync.Mutex
	upgrader    websocket.Upgrader
	recorder    interfaces.AudioRecorder
	whisper     *whisper.WhisperEngine
	server      *http.Server
	started     bool
	retryCount  map[*websocket.Conn]int // Track retry attempts
	logger      logger.Logger
}

// Message represents a message for exchange via WebSocket
type Message struct {
	Type       string      `json:"type"`
	Payload    interface{} `json:"payload,omitempty"`
	APIVersion string      `json:"api_version,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	Timestamp  int64       `json:"timestamp,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// NewWebSocketServer creates a new instance of WebSocketServer
func NewWebSocketServer(config *config.Config, recorder interfaces.AudioRecorder, whisperEngine *whisper.WhisperEngine, logger logger.Logger) *WebSocketServer {
	return &WebSocketServer{
		config:  config,
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Check origin if CORS origins is set
				if config.WebServer.CORSOrigins != "*" {
					// TODO: Implement proper CORS check
					return true
				}
				return true
			},
		},
		recorder:   recorder,
		whisper:    whisperEngine,
		retryCount: make(map[*websocket.Conn]int),
		logger:     logger,
	}
}

// Start launches the WebSocket server
func (s *WebSocketServer) Start() error {
	if !s.config.WebServer.Enabled {
		return nil
	}

	// Handler for WebSocket connection setup
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)

	// API versioning - support multiple API versions
	apiVersion := s.config.WebServer.APIVersion
	if apiVersion != "" {
		mux.HandleFunc(fmt.Sprintf("/api/%s/ws", apiVersion), s.handleWebSocket)
	}

	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			s.logger.Debug("health write error: %v", err)
		}
	})

	// Create HTTP server with timeouts
	addr := fmt.Sprintf("%s:%d", s.config.WebServer.Host, s.config.WebServer.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		s.logger.Info("Starting WebSocket server on %s", addr)
		s.started = true
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("WebSocket server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the WebSocket server
func (s *WebSocketServer) Stop() {
	if s.server != nil && s.started {
		s.logger.Info("Stopping WebSocket server...")

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Close all client connections
		s.clientsLock.Lock()
		for client := range s.clients {
			_ = client.Close()
		}
		s.clients = make(map[*websocket.Conn]bool)
		s.clientsLock.Unlock()

		// Shutdown the server
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down WebSocket server: %v", err)
		} else {
			s.logger.Info("WebSocket server stopped")
		}

		s.started = false
	}
}

// handleWebSocket handles WebSocket connections
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check authentication first
	if !s.authenticate(r) {
		s.logger.Warning("Unauthorized WebSocket connection attempt from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if we're at max clients limit
	s.clientsLock.Lock()
	clientCount := len(s.clients)
	s.clientsLock.Unlock()

	if s.config.WebServer.MaxClients > 0 && clientCount >= s.config.WebServer.MaxClients {
		s.logger.Warning("Max clients limit reached, rejecting connection from %s", r.RemoteAddr)
		http.Error(w, "Too many connections", http.StatusServiceUnavailable)
		return
	}

	// Establish connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("Error upgrading to WebSocket: %v", err)
		return
	}

	// Configure connection
	conn.SetReadLimit(1024 * 1024) // 1MB message size limit
	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		s.logger.Debug("SetReadDeadline error: %v", err)
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	// Register new client
	s.clientsLock.Lock()
	s.clients[conn] = true
	s.clientsLock.Unlock()

	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Debug("conn close error: %v", err)
		}
		s.clientsLock.Lock()
		delete(s.clients, conn)
		delete(s.retryCount, conn)
		s.clientsLock.Unlock()
	}()

	// Send welcome message with version info
	s.sendMessage(conn, "connected", map[string]string{
		"server":      "Speak-to-AI",
		"api_version": s.config.WebServer.APIVersion,
	})

	// Start ping/pong
	go s.pingClient(conn)

	// Process messages from client
	s.processMessages(conn)
}

// pingClient sends periodic pings to keep connection alive
func (s *WebSocketServer) pingClient(conn *websocket.Conn) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			s.logger.Debug("Ping error: %v", err)
			return
		}
	}
}

// sendMessage sends a message to a client
func (s *WebSocketServer) sendMessage(conn *websocket.Conn, messageType string, payload interface{}, requestID ...string) {
	msg := Message{
		Type:       messageType,
		Payload:    payload,
		APIVersion: s.config.WebServer.APIVersion,
		Timestamp:  time.Now().Unix(),
	}

	// Set request ID if provided
	if len(requestID) > 0 && requestID[0] != "" {
		msg.RequestID = requestID[0]
	}

	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error("Error marshaling message: %v", err)
		return
	}

	// Log if enabled
	if s.config.WebServer.LogRequests {
		s.logger.Debug("Sending WebSocket message: %s", string(data))
	}

	// Send message
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		s.logger.Error("SetWriteDeadline error: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		s.logger.Error("Error sending message: %v", err)
	}
}

// BroadcastMessage sends a message to all connected clients
func (s *WebSocketServer) BroadcastMessage(messageType string, payload interface{}) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	for conn := range s.clients {
		s.sendMessage(conn, messageType, payload)
	}
}
