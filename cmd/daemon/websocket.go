package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketServer represents a WebSocket server
type WebSocketServer struct {
	config      *Config
	clients     map[*websocket.Conn]bool
	clientsLock sync.Mutex
	upgrader    websocket.Upgrader
	recorder    AudioRecorder
	whisper     *WhisperEngine
	server      *http.Server
	started     bool
}

// Message represents a message for exchange via WebSocket
type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
}

// NewWebSocketServer creates a new instance of WebSocketServer
func NewWebSocketServer(config *Config, recorder AudioRecorder, whisper *WhisperEngine) *WebSocketServer {
	return &WebSocketServer{
		config:  config,
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, we should check origin
				return true
			},
		},
		recorder: recorder,
		whisper:  whisper,
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

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.WebServer.Host, s.config.WebServer.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting WebSocket server on %s", addr)
		s.started = true
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the WebSocket server
func (s *WebSocketServer) Stop() {
	if s.server != nil && s.started {
		log.Println("Stopping WebSocket server...")

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Close all client connections
		s.clientsLock.Lock()
		for client := range s.clients {
			client.Close()
		}
		s.clients = make(map[*websocket.Conn]bool)
		s.clientsLock.Unlock()

		// Shutdown the server
		if err := s.server.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down WebSocket server: %v", err)
		} else {
			log.Println("WebSocket server stopped")
		}

		s.started = false
	}
}

// handleWebSocket handles WebSocket connections
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Establish connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Register new client
	s.clientsLock.Lock()
	s.clients[conn] = true
	s.clientsLock.Unlock()

	// Remove client when connection closes
	defer func() {
		s.clientsLock.Lock()
		delete(s.clients, conn)
		s.clientsLock.Unlock()
	}()

	// Process messages from client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing WebSocket message: %v", err)
			continue
		}

		// Process message based on type
		switch msg.Type {
		case "start-recording":
			go s.handleStartRecording(conn)
		case "stop-recording":
			go s.handleStopRecording(conn)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// handleStartRecording handles request to start recording
func (s *WebSocketServer) handleStartRecording(conn *websocket.Conn) {
	err := s.recorder.StartRecording()
	if err != nil {
		s.sendError(conn, fmt.Sprintf("Failed to start recording: %v", err))
		return
	}

	s.sendMessage(conn, "recording-started", "")
}

// handleStopRecording handles request to stop recording
func (s *WebSocketServer) handleStopRecording(conn *websocket.Conn) {
	// Stop recording and get path to audio file
	audioFile, err := s.recorder.StopRecording()
	if err != nil {
		s.sendError(conn, fmt.Sprintf("Failed to stop recording: %v", err))
		return
	}

	// Send message that recording stopped
	s.sendMessage(conn, "recording-stopped", "")

	// Perform transcription
	s.sendMessage(conn, "transcribing", "")
	transcript, err := s.whisper.Transcribe(audioFile)
	if err != nil {
		s.sendError(conn, fmt.Sprintf("Failed to transcribe: %v", err))
		return
	}

	// Send result
	s.sendMessage(conn, "transcript", transcript)
}

// sendMessage sends a message to client
func (s *WebSocketServer) sendMessage(conn *websocket.Conn, messageType, payload string) {
	msg := Message{
		Type:    messageType,
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// sendError sends an error message to client
func (s *WebSocketServer) sendError(conn *websocket.Conn, errMsg string) {
	s.sendMessage(conn, "error", errMsg)
}

// BroadcastMessage sends a message to all connected clients
func (s *WebSocketServer) BroadcastMessage(messageType, payload string) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	for client := range s.clients {
		s.sendMessage(client, messageType, payload)
	}
}
