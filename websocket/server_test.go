// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/mocks"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/whisper"
	"github.com/gorilla/websocket"
)

// Mock implementations for testing
type MockLogger struct {
	messages []string
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.messages = append(m.messages, "DEBUG: "+format)
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.messages = append(m.messages, "INFO: "+format)
}

func (m *MockLogger) Warning(format string, args ...interface{}) {
	m.messages = append(m.messages, "WARNING: "+format)
}

func (m *MockLogger) Error(format string, args ...interface{}) {
	m.messages = append(m.messages, "ERROR: "+format)
}

type MockWhisperEngine struct {
	transcribeResult string
	transcribeErr    error
}

func (m *MockWhisperEngine) Transcribe(audioFile string) (string, error) {
	return m.transcribeResult, m.transcribeErr
}

func createTestConfig() *config.Config {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	cfg.WebServer.Enabled = true
	cfg.WebServer.Port = 8080
	cfg.WebServer.Host = "localhost"
	cfg.WebServer.AuthToken = ""
	cfg.WebServer.APIVersion = "v1"
	cfg.WebServer.LogRequests = false
	cfg.WebServer.CORSOrigins = "*"
	cfg.WebServer.MaxClients = 10
	return cfg
}

func createTestWhisperEngine(cfg *config.Config) *whisper.WhisperEngine {
	engine, _ := whisper.NewWhisperEngine(cfg, "/dev/null")
	return engine
}

func TestNewWebSocketServer(t *testing.T) {
	cfg := createTestConfig()
	recorder := &mocks.MockAudioRecorder{}
	whisperEngine, _ := whisper.NewWhisperEngine(cfg, "/dev/null")
	logger := &MockLogger{}

	server := NewWebSocketServer(cfg, recorder, whisperEngine, logger)

	if server == nil {
		t.Fatal("NewWebSocketServer returned nil")
	}

	if server.config != cfg {
		t.Error("Config not set correctly")
	}

	if server.logger != logger {
		t.Error("Logger not set correctly")
	}

	if server.clients == nil {
		t.Error("Clients map should be initialized")
	}

	if server.retryCount == nil {
		t.Error("Retry count map should be initialized")
	}
}

func TestWebSocketServer_Start_Disabled(t *testing.T) {
	cfg := createTestConfig()
	cfg.WebServer.Enabled = false

	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	err := server.Start()

	if err != nil {
		t.Errorf("Expected no error when server is disabled, got %v", err)
	}

	if server.started {
		t.Error("Server should not be started when disabled")
	}
}

func TestWebSocketServer_Start_Enabled(t *testing.T) {
	cfg := createTestConfig()
	cfg.WebServer.Port = 0 // Use random port for testing

	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	err := server.Start()

	if err != nil {
		t.Errorf("Expected no error when starting server, got %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Clean up
	server.Stop()
}

func TestWebSocketServer_Stop(t *testing.T) {
	cfg := createTestConfig()
	cfg.WebServer.Port = 0 // Use random port for testing

	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	// Start server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	server.Stop()

	if server.started {
		t.Error("Server should not be started after Stop()")
	}
}

func TestWebSocketServer_Authentication_NoToken(t *testing.T) {
	cfg := createTestConfig()
	cfg.WebServer.AuthToken = ""

	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	// Create test request
	req := httptest.NewRequest("GET", "/ws", nil)

	result := server.authenticate(req)

	if !result {
		t.Error("Expected authentication to pass when no token is required")
	}
}

func TestWebSocketServer_Authentication_WithToken(t *testing.T) {
	cfg := createTestConfig()
	cfg.WebServer.AuthToken = "test-token"

	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	tests := []struct {
		name       string
		setupReq   func(*http.Request)
		expectAuth bool
	}{
		{
			name: "valid query token",
			setupReq: func(req *http.Request) {
				q := req.URL.Query()
				q.Set("token", "test-token")
				req.URL.RawQuery = q.Encode()
			},
			expectAuth: true,
		},
		{
			name: "valid header token",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer test-token")
			},
			expectAuth: true,
		},
		{
			name: "invalid token",
			setupReq: func(req *http.Request) {
				q := req.URL.Query()
				q.Set("token", "wrong-token")
				req.URL.RawQuery = q.Encode()
			},
			expectAuth: false,
		},
		{
			name: "no token",
			setupReq: func(req *http.Request) {
				// No token provided
			},
			expectAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			tt.setupReq(req)

			result := server.authenticate(req)

			if result != tt.expectAuth {
				t.Errorf("Expected authentication result %v, got %v", tt.expectAuth, result)
			}
		})
	}
}

func TestWebSocketServer_ValidateToken(t *testing.T) {
	cfg := createTestConfig()
	cfg.WebServer.AuthToken = "test-token"

	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "valid token",
			token:    "test-token",
			expected: true,
		},
		{
			name:     "invalid token",
			token:    "wrong-token",
			expected: false,
		},
		{
			name:     "empty token",
			token:    "",
			expected: false,
		},
		{
			name:     "token with whitespace",
			token:    "  test-token  ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.validateToken(tt.token)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWebSocketServer_ValidateToken_NoAuthToken(t *testing.T) {
	cfg := createTestConfig()
	cfg.WebServer.AuthToken = ""

	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	result := server.validateToken("any-token")

	if result {
		t.Error("Expected validation to fail when no auth token is set")
	}
}

func TestWebSocketServer_SendMessage(t *testing.T) {
	cfg := createTestConfig()
	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	// Create test WebSocket connection
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}
		defer conn.Close()

		// Send test message
		server.sendMessage(conn, "test", map[string]string{"key": "value"}, "req-123")

		// Read the message
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		// Parse the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		// Verify message fields
		if msg.Type != "test" {
			t.Errorf("Expected message type 'test', got %q", msg.Type)
		}

		if msg.RequestID != "req-123" {
			t.Errorf("Expected request ID 'req-123', got %q", msg.RequestID)
		}

		if msg.APIVersion != cfg.WebServer.APIVersion {
			t.Errorf("Expected API version %q, got %q", cfg.WebServer.APIVersion, msg.APIVersion)
		}

		if msg.Timestamp == 0 {
			t.Error("Expected timestamp to be set")
		}
	}))
	defer testServer.Close()

	// Connect to test server
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
}

func TestWebSocketServer_ExecuteWithRetry_Success(t *testing.T) {
	cfg := createTestConfig()
	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	// Create mock connection
	conn := &websocket.Conn{}
	server.clients[conn] = true
	server.retryCount[conn] = 0

	// Test successful execution
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := server.executeWithRetry(fn, conn)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d calls", callCount)
	}

	// Verify retry count was reset
	if server.retryCount[conn] != 0 {
		t.Errorf("Expected retry count to be reset to 0, got %d", server.retryCount[conn])
	}
}

func TestWebSocketServer_ExecuteWithRetry_MaxRetries(t *testing.T) {
	cfg := createTestConfig()
	server := NewWebSocketServer(cfg, &mocks.MockAudioRecorder{}, createTestWhisperEngine(cfg), &MockLogger{})

	// Create mock connection
	conn := &websocket.Conn{}
	server.clients[conn] = true
	server.retryCount[conn] = 0

	// Test function that always fails
	callCount := 0
	testErr := fmt.Errorf("test error")
	fn := func() error {
		callCount++
		return testErr
	}

	err := server.executeWithRetry(fn, conn)

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}

	// Should be called 4 times (initial + 3 retries)
	if callCount != 4 {
		t.Errorf("Expected function to be called 4 times, got %d calls", callCount)
	}

	// Verify retry count was reset after max retries
	if server.retryCount[conn] != 0 {
		t.Errorf("Expected retry count to be reset to 0, got %d", server.retryCount[conn])
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func(*http.Request)
		expected string
	}{
		{
			name: "X-Forwarded-For header",
			setupReq: func(req *http.Request) {
				req.Header.Set("X-Forwarded-For", "192.168.1.1,10.0.0.1")
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			setupReq: func(req *http.Request) {
				req.Header.Set("X-Real-IP", "192.168.1.2")
			},
			expected: "192.168.1.2",
		},
		{
			name: "RemoteAddr fallback",
			setupReq: func(req *http.Request) {
				req.RemoteAddr = "192.168.1.3:12345"
			},
			expected: "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			tt.setupReq(req)

			result := getClientIP(req)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetRetryBackoff(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  1,
			minDelay: 500 * time.Millisecond,
			maxDelay: 1000 * time.Millisecond,
		},
		{
			name:     "second attempt",
			attempt:  2,
			minDelay: 1000 * time.Millisecond,
			maxDelay: 1500 * time.Millisecond,
		},
		{
			name:     "high attempt (capped)",
			attempt:  20,
			minDelay: 4000 * time.Millisecond,
			maxDelay: 6000 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := getRetryBackoff(tt.attempt)

			if delay < tt.minDelay {
				t.Errorf("Expected delay >= %v, got %v", tt.minDelay, delay)
			}

			if delay > tt.maxDelay {
				t.Errorf("Expected delay <= %v, got %v", tt.maxDelay, delay)
			}
		})
	}
}
