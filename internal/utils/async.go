// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"sync"
	"time"
)

// Track long-lived background goroutines for coordinated shutdown
var goroutineWg sync.WaitGroup

// Go launches a function in a goroutine and tracks it for shutdown coordination
func Go(fn func()) {
	goroutineWg.Add(1)
	go func() {
		defer goroutineWg.Done()
		fn()
	}()
}

// Wait for all tracked goroutines to complete or times out
// Returns true if all goroutines completed before the timeout, false otherwise
func WaitAll(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		goroutineWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
