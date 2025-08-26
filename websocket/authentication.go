// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"net/http"
	"strings"
)

// authenticate checks if the client is authenticated
func (s *WebSocketServer) authenticate(r *http.Request) bool {
	// If auth token is not set, all connections are allowed
	if s.config.WebServer.AuthToken == "" {
		return true
	}

	// Check for token in query params or headers
	queryToken := r.URL.Query().Get("token")
	headerToken := r.Header.Get("Authorization")

	// Extract bearer token if present
	if len(headerToken) > 7 && headerToken[:7] == "Bearer " {
		headerToken = headerToken[7:]
	}

	// Check if either token matches
	return queryToken == s.config.WebServer.AuthToken || headerToken == s.config.WebServer.AuthToken
}

// validateToken checks if a token is valid
func (s *WebSocketServer) validateToken(token string) bool {
	// If auth token is not set, all tokens are invalid
	if s.config.WebServer.AuthToken == "" {
		return false
	}

	// Trim whitespace
	token = strings.TrimSpace(token)

	// Compare with configured token
	return token == s.config.WebServer.AuthToken
}

// getClientIP gets the client's IP address
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP in the list
		return strings.Split(forwardedFor, ",")[0]
	}

	// Check for X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
