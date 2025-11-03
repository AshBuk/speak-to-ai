// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// Verify client credentials via token or open access policy
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

	// Check if either token matches using constant-time comparison to prevent timing attacks
	queryMatch := subtle.ConstantTimeCompare([]byte(queryToken), []byte(s.config.WebServer.AuthToken)) == 1
	headerMatch := subtle.ConstantTimeCompare([]byte(headerToken), []byte(s.config.WebServer.AuthToken)) == 1

	return queryMatch || headerMatch
}

// Confirm token matches configured authentication secret
func (s *WebSocketServer) validateToken(token string) bool { // nolint:unused // used in tests
	// If auth token is not set, all tokens are invalid
	if s.config.WebServer.AuthToken == "" {
		return false
	}

	// Trim whitespace
	token = strings.TrimSpace(token)

	// Compare with configured token using constant-time comparison
	return subtle.ConstantTimeCompare([]byte(token), []byte(s.config.WebServer.AuthToken)) == 1
}

// Extract real client IP considering proxy headers
func getClientIP(r *http.Request) string { // nolint:unused // used in tests
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
