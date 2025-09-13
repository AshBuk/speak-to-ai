// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

import (
	"testing"
	"time"

	"github.com/AshBuk/speak-to-ai/internal/services"
	"github.com/AshBuk/speak-to-ai/internal/testutils"
)

func TestRuntimeContext(t *testing.T) {
	mockLogger := testutils.NewMockLogger()

	t.Run("NewRuntimeContext", func(t *testing.T) {
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
	})

	t.Run("ShutdownChannel", func(t *testing.T) {
		runtime := NewRuntimeContext(mockLogger)

		// Test that shutdown channel is buffered
		if cap(runtime.ShutdownCh) == 0 {
			t.Error("Shutdown channel should be buffered")
		}
	})
}

func TestApp(t *testing.T) {
	mockLogger := testutils.NewMockLogger()

	t.Run("NewApp", func(t *testing.T) {
		app := NewApp(mockLogger)

		if app == nil {
			t.Fatal("NewApp returned nil")
		}
		if app.Services == nil {
			t.Error("Services container not initialized")
		}
		if app.Runtime == nil {
			t.Error("Runtime context not initialized")
		}
		if app.Runtime.Logger != mockLogger {
			t.Error("Runtime logger not set correctly")
		}
	})

	t.Run("ServiceAccessors", func(t *testing.T) {
		app := NewApp(mockLogger)

		// Test accessor methods with nil services
		if app.Audio() != nil {
			t.Error("Audio() should return nil when service is not set")
		}
		if app.UI() != nil {
			t.Error("UI() should return nil when service is not set")
		}
		if app.IO() != nil {
			t.Error("IO() should return nil when service is not set")
		}
		if app.Config() != nil {
			t.Error("Config() should return nil when service is not set")
		}
		if app.Hotkeys() != nil {
			t.Error("Hotkeys() should return nil when service is not set")
		}

		// Test accessor methods with nil Services container
		app.Services = nil
		if app.Audio() != nil {
			t.Error("Audio() should return nil when Services container is nil")
		}
		if app.UI() != nil {
			t.Error("UI() should return nil when Services container is nil")
		}
		if app.IO() != nil {
			t.Error("IO() should return nil when Services container is nil")
		}
		if app.Config() != nil {
			t.Error("Config() should return nil when Services container is nil")
		}
		if app.Hotkeys() != nil {
			t.Error("Hotkeys() should return nil when Services container is nil")
		}
	})

	t.Run("Shutdown_NilServices", func(t *testing.T) {
		app := NewApp(mockLogger)
		app.Services = nil

		err := app.Shutdown()
		if err != nil {
			t.Errorf("Shutdown should not error with nil services: %v", err)
		}

		// Verify context was cancelled
		select {
		case <-app.Runtime.Ctx.Done():
			// Expected behavior
		case <-time.After(100 * time.Millisecond):
			t.Error("Context should be cancelled after Shutdown()")
		}
	})
}

// Mock services for testing
type mockServiceWithShutdown struct {
	shutdownCalled bool
	shutdownError  error
}

func (m *mockServiceWithShutdown) Shutdown() error {
	m.shutdownCalled = true
	return m.shutdownError
}

func TestApp_Integration(t *testing.T) {
	mockLogger := testutils.NewMockLogger()

	t.Run("Shutdown_WithServices", func(t *testing.T) {
		app := NewApp(mockLogger)

		// Create a mock service container with our mock service
		mockServices := &services.ServiceContainer{}

		// We'll test that the service container's Shutdown is called
		// Since we can't easily mock the entire service container,
		// we'll test that the method handles the nil services gracefully
		app.Services = mockServices

		err := app.Shutdown()
		if err != nil {
			t.Errorf("Shutdown failed: %v", err)
		}

		// Verify context was cancelled
		select {
		case <-app.Runtime.Ctx.Done():
			// Expected behavior
		case <-time.After(100 * time.Millisecond):
			t.Error("Context should be cancelled after Shutdown()")
		}
	})

	t.Run("AppLifecycle", func(t *testing.T) {
		app := NewApp(mockLogger)

		// Test initial state
		if app.Runtime.Ctx.Err() != nil {
			t.Error("Context should not be cancelled initially")
		}

		// Test shutdown
		err := app.Shutdown()
		if err != nil {
			t.Errorf("Shutdown failed: %v", err)
		}

		// Verify context was cancelled
		if app.Runtime.Ctx.Err() == nil {
			t.Error("Context should be cancelled after shutdown")
		}
	})
}

func TestApp_HandlerMethods(t *testing.T) {
	mockLogger := testutils.NewMockLogger()

	t.Run("Handlers_NilServices", func(t *testing.T) {
		app := NewApp(mockLogger)
		app.Services = nil

		// Test all handler methods with nil services
		err := app.handleStartRecording()
		if err == nil {
			t.Error("handleStartRecording should fail with nil services")
		}

		err = app.handleStopRecordingAndTranscribe()
		if err == nil {
			t.Error("handleStopRecordingAndTranscribe should fail with nil services")
		}

		// TODO: Next feature - VAD implementation
		// err = app.handleToggleVAD()
		// if err == nil {
		//	t.Error("handleToggleVAD should fail with nil services")
		// }

		err = app.handleSwitchModel()
		if err == nil {
			t.Error("handleSwitchModel should fail with nil services")
		}

		err = app.handleShowConfig()
		if err == nil {
			t.Error("handleShowConfig should fail with nil services")
		}

		err = app.handleResetToDefaults()
		if err == nil {
			t.Error("handleResetToDefaults should fail with nil services")
		}
	})

	t.Run("Handlers_NilSpecificServices", func(t *testing.T) {
		app := NewApp(mockLogger)
		// Services container exists but individual services are nil

		err := app.handleStartRecording()
		if err == nil {
			t.Error("handleStartRecording should fail with nil audio service")
		}

		err = app.handleStopRecordingAndTranscribe()
		if err == nil {
			t.Error("handleStopRecordingAndTranscribe should fail with nil audio service")
		}

		// TODO: Next feature - VAD implementation
		// err = app.handleToggleVAD()
		// if err == nil {
		//	t.Error("handleToggleVAD should fail with nil config service")
		// }

		err = app.handleSwitchModel()
		if err == nil {
			t.Error("handleSwitchModel should fail with nil audio service")
		}

		err = app.handleShowConfig()
		if err == nil {
			t.Error("handleShowConfig should fail with nil UI service")
		}

		err = app.handleResetToDefaults()
		if err == nil {
			t.Error("handleResetToDefaults should fail with nil config service")
		}
	})
}
