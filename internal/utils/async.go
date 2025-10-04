// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

// generation-based tracker to isolate batches between waits
var (
	currentGen int64
	genMu      sync.Mutex
	genCounts  = map[int64]int64{}
	lastGoTsNs int64
)

// Go launches a function in a goroutine and tracks it for shutdown coordination
func Go(fn func()) {
	gen := atomic.LoadInt64(&currentGen)
	atomic.StoreInt64(&lastGoTsNs, time.Now().UnixNano())
	genMu.Lock()
	genCounts[gen] = genCounts[gen] + 1
	genMu.Unlock()

	go func(g int64) {
		defer func() {
			genMu.Lock()
			genCounts[g] = genCounts[g] - 1
			if genCounts[g] == 0 {
				delete(genCounts, g)
			}
			genMu.Unlock()
		}()
		fn()
	}(gen)
}

// Wait for all tracked goroutines in the current generation to complete or time out
// Returns true if all goroutines completed before the timeout, false otherwise
func WaitAll(timeout time.Duration) bool {
	gen := atomic.LoadInt64(&currentGen)
	deadline := time.Now().Add(timeout)
	start := time.Now()
	const minSettle = 50 * time.Millisecond
	for {
		genMu.Lock()
		count := genCounts[gen]
		genMu.Unlock()
		if count == 0 {
			// Ensure we don't return immediately before late registrations
			if time.Since(start) < minSettle {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			// Additional stabilization window to capture last-second registrations
			last := time.Unix(0, atomic.LoadInt64(&lastGoTsNs))
			if time.Since(last) > minSettle/2 {
				return true
			}
		}
		if time.Now().After(deadline) {
			// Advance generation so future Go() calls are tracked separately
			atomic.AddInt64(&currentGen, 1)
			return false
		}
		time.Sleep(1 * time.Millisecond)
	}
}
