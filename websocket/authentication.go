// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"crypto/subtle"
	"net/http"
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
