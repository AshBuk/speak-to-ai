// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestGoNoGoroutineLeak verifies that utils.Go() goroutines complete
func TestGoNoGoroutineLeak(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var counter int32

	// Launch multiple tracked goroutines
	for i := 0; i < 10; i++ {
		Go(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	// Wait for all goroutines
	if !WaitAll(1 * time.Second) {
		t.Error("WaitAll timed out")
	}

	// Verify all goroutines executed
	if atomic.LoadInt32(&counter) != 10 {
		t.Errorf("Expected 10 goroutines, got %d", counter)
	}

	// goleak will fail if any goroutines leaked
}

// TestWaitAllTimeout verifies timeout behavior
func TestWaitAllTimeout(t *testing.T) {
	// Launch a long-running goroutine
	Go(func() {
		time.Sleep(5 * time.Second)
	})

	// WaitAll with short timeout should return false
	start := time.Now()
	result := WaitAll(100 * time.Millisecond)
	duration := time.Since(start)

	if result {
		t.Error("WaitAll should have timed out")
	}

	// Should timeout around 100ms
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Timeout duration unexpected: %v (expected ~100ms)", duration)
	}

	// Note: Long-running goroutine will complete eventually
	// This is acceptable for shutdown timeout scenario
}

// TestWaitAllSuccess verifies successful completion
func TestWaitAllSuccess(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var completed atomic.Bool

	Go(func() {
		time.Sleep(50 * time.Millisecond)
		completed.Store(true)
	})

	// Wait with sufficient timeout
	start := time.Now()
	result := WaitAll(500 * time.Millisecond)
	duration := time.Since(start)

	if !result {
		t.Error("WaitAll should have completed successfully")
	}

	if !completed.Load() {
		t.Error("Goroutine did not complete")
	}

	// Should complete around 50ms, not wait full timeout
	if duration > 150*time.Millisecond {
		t.Errorf("WaitAll took too long: %v (goroutine finished in ~50ms)", duration)
	}

	t.Logf("WaitAll completed in %v", duration)
}

// TestConcurrentGoLaunches verifies thread-safety
func TestConcurrentGoLaunches(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var counter int32

	// Launch goroutines from multiple goroutines (stress test WaitGroup)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				Go(func() {
					atomic.AddInt32(&counter, 1)
					time.Sleep(5 * time.Millisecond)
				})
			}
		}()
	}

	// Wait for all 50 goroutines
	if !WaitAll(2 * time.Second) {
		t.Error("WaitAll timed out")
	}

	// Verify count
	if atomic.LoadInt32(&counter) != 50 {
		t.Errorf("Expected 50 goroutines, got %d", counter)
	}
}
