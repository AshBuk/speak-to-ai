//go:build integration
// +build integration

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package integration

import (
	"testing"
	"time"

	"github.com/AshBuk/speak-to-ai/hotkeys/providers"
	"github.com/AshBuk/speak-to-ai/internal/app"
	"github.com/AshBuk/speak-to-ai/internal/testutils"
	"go.uber.org/goleak"
)

// TestMain enables goroutine leak detection for integration tests
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// Ignore known long-lived goroutines from dependencies
		goleak.IgnoreTopFunction("github.com/godbus/dbus/v5.(*Conn).inWorker"),
		goleak.IgnoreTopFunction("github.com/godbus/dbus/v5.(*Conn).outWorker"),
	)
}

// TestAppShutdownNoGoroutineLeaks verifies graceful shutdown without leaks
func TestAppShutdownNoGoroutineLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("github.com/godbus/dbus/v5.(*Conn).inWorker"),
		goleak.IgnoreTopFunction("github.com/godbus/dbus/v5.(*Conn).outWorker"),
	)

	mockLogger := testutils.NewMockLogger()
	app := app.NewApp(mockLogger)

	// Initialize without starting long-lived services
	err := app.Initialize("../../config.yaml", false)
	if err != nil {
		t.Logf("Initialize failed (expected in test environment): %v", err)
	}

	// Shutdown should complete without goroutine leaks
	err = app.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Give goroutines time to exit
	time.Sleep(100 * time.Millisecond)
}

// TestGracefulShutdownTimeout verifies shutdown completes within timeout
func TestGracefulShutdownTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockLogger := testutils.NewMockLogger()
	app := app.NewApp(mockLogger)

	err := app.Initialize("../../config.yaml", false)
	if err != nil {
		t.Logf("Initialize failed: %v", err)
	}

	// Measure shutdown time
	start := time.Now()
	err = app.Shutdown()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Shutdown should complete quickly (< 6 seconds including 5s timeout)
	if duration > 6*time.Second {
		t.Errorf("Shutdown took too long: %v (expected < 6s)", duration)
	}

	t.Logf("Shutdown completed in %v", duration)
}

// TestEvdevProviderLifecycle verifies evdev provider goroutine cleanup
func TestEvdevProviderLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	defer goleak.VerifyNone(t)

	mockLogger := testutils.NewMockLogger()

	provider := providers.NewEvdevKeyboardProvider(mockLogger)

	// Check if evdev is supported
	if !provider.IsSupported() {
		t.Skip("Evdev not supported in this environment")
	}

	// Start provider
	err := provider.Start()
	if err != nil {
		t.Skipf("Could not start evdev provider: %v", err)
	}

	// Give goroutines time to start
	time.Sleep(50 * time.Millisecond)

	// Stop should wait for all device listener goroutines
	start := time.Now()
	provider.Stop()
	duration := time.Since(start)

	// Should complete quickly (< 1 second)
	if duration > 1*time.Second {
		t.Errorf("Stop() took too long: %v (expected < 1s)", duration)
	}

	// Give goroutines time to fully exit
	time.Sleep(100 * time.Millisecond)

	t.Logf("Evdev provider stopped in %v", duration)
}

// TestDbusProviderLifecycle verifies dbus provider goroutine cleanup
func TestDbusProviderLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("github.com/godbus/dbus/v5.(*Conn).inWorker"),
		goleak.IgnoreTopFunction("github.com/godbus/dbus/v5.(*Conn).outWorker"),
	)

	mockLogger := testutils.NewMockLogger()

	provider := providers.NewDbusKeyboardProvider(mockLogger)

	// Check if dbus is supported
	if !provider.IsSupported() {
		t.Skip("DBus not supported in this environment")
	}

	// Start provider
	err := provider.Start()
	if err != nil {
		t.Skipf("Could not start dbus provider: %v", err)
	}

	// Give listener time to start
	time.Sleep(50 * time.Millisecond)

	// Stop should wait for listener goroutine
	start := time.Now()
	provider.Stop()
	duration := time.Since(start)

	// Should complete quickly
	if duration > 1*time.Second {
		t.Errorf("Stop() took too long: %v (expected < 1s)", duration)
	}

	// Give goroutine time to exit
	time.Sleep(100 * time.Millisecond)

	t.Logf("DBus provider stopped in %v", duration)
}
