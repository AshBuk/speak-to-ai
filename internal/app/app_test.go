// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"testing"
	"time"

	"github.com/AshBuk/dabri/v2/internal/services"
	"github.com/AshBuk/dabri/v2/internal/testutils"
)

func TestRuntimeContext_NewRuntimeContext(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	runtime := NewRuntimeContext(mockLogger)
	if runtime == nil {
		t.Fatal("NewRuntimeContext returned nil")
	}
	if runtime.Logger != mockLogger {
		t.Error("Logger not set correctly")
	}
	if runtime.Ctx == nil {
		t.Error("Context not initialized")
	}
	if runtime.Cancel == nil {
		t.Error("Cancel function not initialized")
	}
	if runtime.ShutdownCh == nil {
		t.Error("Shutdown channel not initialized")
	}

	// Test that context is not cancelled initially
	select {
	case <-runtime.Ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Expected behavior
	}
	// Test cancellation
	runtime.Cancel()
	select {
	case <-runtime.Ctx.Done():
		// Expected behavior
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after calling Cancel()")
	}
}

func TestRuntimeContext_ShutdownChannel(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	runtime := NewRuntimeContext(mockLogger)
	// Test that shutdown channel is buffered
	if cap(runtime.ShutdownCh) == 0 {
		t.Error("Shutdown channel should be buffered")
	}
}

func TestApp_NewApp(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	app := NewApp(mockLogger)
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.Services != nil {
		t.Error("Services must stay nil until Initialize")
	}
	if app.Runtime == nil {
		t.Error("Runtime context not initialized")
	}
	if app.Runtime.Logger != mockLogger {
		t.Error("Runtime logger not set correctly")
	}
}

func TestApp_Shutdown_NilServices(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	app := NewApp(mockLogger)
	app.Services = nil

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown should not error with nil services: %v", err)
	}
	// Verify context was cancelled
	select {
	case <-app.Runtime.Ctx.Done():
		// Expected behavior
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after Shutdown()")
	}
}

func TestApp_Integration_Shutdown_WithServices(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	app := NewApp(mockLogger)
	// Create a mock service container with our mock service
	mockServices := &services.ServiceContainer{}

	// We'll test that the service container's Shutdown is called
	// Since we can't easily mock the entire service container,
	// we'll test that the method handles the nil services gracefully
	app.Services = mockServices

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify context was cancelled
	select {
	case <-app.Runtime.Ctx.Done():
		// Expected behavior
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after Shutdown()")
	}
}

func TestApp_Integration_AppLifecycle(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	app := NewApp(mockLogger)
	// Test initial state
	if app.Runtime.Ctx.Err() != nil {
		t.Error("Context should not be cancelled initially")
	}

	// Test shutdown
	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify context was cancelled
	if app.Runtime.Ctx.Err() == nil {
		t.Error("Context should be cancelled after shutdown")
	}
}
