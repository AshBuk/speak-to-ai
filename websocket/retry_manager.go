// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

// executeWithRetry attempts to execute a function with retries
func (s *WebSocketServer) executeWithRetry(fn func() error, conn *websocket.Conn) error {
	// Get current retry count for this connection
	s.clientsLock.Lock()
	currentRetries := s.retryCount[conn]
	s.clientsLock.Unlock()

	// Maximum number of retries
	maxRetries := 3

	// Execute the function
	err := fn()

	// If successful or reached max retries, reset counter and return
	if err == nil || currentRetries >= maxRetries {
		s.clientsLock.Lock()
		s.retryCount[conn] = 0
		s.clientsLock.Unlock()
		return err
	}

	// Increment retry counter
	s.clientsLock.Lock()
	s.retryCount[conn] = currentRetries + 1
	s.clientsLock.Unlock()

	// Retry with exponential backoff
	backoff := time.Duration(currentRetries+1) * 500 * time.Millisecond
	time.Sleep(backoff)

	s.logger.Debug("Retrying operation, attempt %d/%d", currentRetries+1, maxRetries)
	return s.executeWithRetry(fn, conn)
}

// resetRetryCount resets the retry counter for a connection
func (s *WebSocketServer) resetRetryCount(conn *websocket.Conn) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()
	s.retryCount[conn] = 0
}

// getRetryBackoff calculates retry backoff duration
func getRetryBackoff(attempt int) time.Duration {
	baseDelay := 500 * time.Millisecond
	maxDelay := 5 * time.Second

	// Calculate exponential backoff with jitter
	delay := time.Duration(attempt) * baseDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	// Add up to 20% jitter to avoid thundering herd problem
	jitter := time.Duration(float64(delay) * 0.2 * (float64(time.Now().UnixNano()%100) / 100.0))
	return delay + jitter
}
