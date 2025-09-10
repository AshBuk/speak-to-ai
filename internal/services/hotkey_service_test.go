// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"errors"
	"testing"

	"github.com/AshBuk/speak-to-ai/tests/mocks"
)

func TestHotkeyService(t *testing.T) {
	mockLogger := &mocks.MockLogger{}

	t.Run("NewHotkeyService", func(t *testing.T) {
		mockManager := &mocks.MockHotkeyManager{}
		service := NewHotkeyService(mockLogger, mockManager)

		if service == nil {
			t.Fatal("NewHotkeyService returned nil")
		}
		if service.logger != mockLogger {
			t.Error("Logger not set correctly")
		}
		if service.hotkeyManager != mockManager {
			t.Error("Hotkey manager not set correctly")
		}
	})

	t.Run("RegisterHotkeys_Success", func(t *testing.T) {
		mockManager := &mocks.MockHotkeyManager{}
		service := NewHotkeyService(mockLogger, mockManager)

		err := service.RegisterHotkeys()
		if err != nil {
			t.Errorf("RegisterHotkeys failed: %v", err)
		}
		if !mockManager.WasStartCalled() {
			t.Error("Hotkey manager Start method was not called")
		}
	})

	t.Run("RegisterHotkeys_Error", func(t *testing.T) {
		expectedError := errors.New("hotkey registration failed")
		mockManager := &mocks.MockHotkeyManager{}
		mockManager.SetStartError(expectedError)
		service := NewHotkeyService(mockLogger, mockManager)

		err := service.RegisterHotkeys()
		if err != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}
	})

	t.Run("RegisterHotkeys_NoManager", func(t *testing.T) {
		service := NewHotkeyService(mockLogger, nil)

		err := service.RegisterHotkeys()
		if err == nil {
			t.Error("RegisterHotkeys should fail when hotkey manager is nil")
		}
	})

	t.Run("UnregisterHotkeys_Success", func(t *testing.T) {
		mockManager := &mocks.MockHotkeyManager{}
		service := NewHotkeyService(mockLogger, mockManager)

		err := service.UnregisterHotkeys()
		if err != nil {
			t.Errorf("UnregisterHotkeys failed: %v", err)
		}
		if !mockManager.WasStopCalled() {
			t.Error("Hotkey manager Stop method was not called")
		}
	})

	t.Run("UnregisterHotkeys_NoManager", func(t *testing.T) {
		service := NewHotkeyService(mockLogger, nil)

		err := service.UnregisterHotkeys()
		if err != nil {
			t.Errorf("UnregisterHotkeys should not fail when hotkey manager is nil: %v", err)
		}
	})

	t.Run("SetupHotkeyCallbacks_Success", func(t *testing.T) {
		mockManager := &mocks.MockHotkeyManager{}
		service := NewHotkeyService(mockLogger, mockManager)

		// Create test callback functions
		var callbacksCalled []string
		startRecording := func() error { callbacksCalled = append(callbacksCalled, "startRecording"); return nil }
		stopRecording := func() error { callbacksCalled = append(callbacksCalled, "stopRecording"); return nil }
		toggleStreaming := func() error { callbacksCalled = append(callbacksCalled, "toggleStreaming"); return nil }
		toggleVAD := func() error { callbacksCalled = append(callbacksCalled, "toggleVAD"); return nil }
		switchModel := func() error { callbacksCalled = append(callbacksCalled, "switchModel"); return nil }
		showConfig := func() error { callbacksCalled = append(callbacksCalled, "showConfig"); return nil }
		reloadConfig := func() error { callbacksCalled = append(callbacksCalled, "reloadConfig"); return nil }

		err := service.SetupHotkeyCallbacks(
			startRecording, stopRecording, toggleStreaming, toggleVAD,
			switchModel, showConfig, reloadConfig,
		)
		if err != nil {
			t.Errorf("SetupHotkeyCallbacks failed: %v", err)
		}

		// Verify RegisterCallbacks was called
		if !mockManager.WereCallbacksRegistered() {
			t.Error("RegisterCallbacks was not called")
		}

		// Verify all hotkey actions were registered
		expectedActions := []string{"toggle_streaming", "toggle_vad", "switch_model", "show_config", "reset_to_defaults"}
		for _, action := range expectedActions {
			if !mockManager.WasHotkeyActionRegistered(action) {
				t.Errorf("RegisterHotkeyAction was not called for action: %s", action)
			}
		}

		// Test that callbacks work
		mockManager.TriggerCallback("startRecording")
		mockManager.TriggerCallback("toggleStreaming")

		if len(callbacksCalled) != 2 {
			t.Errorf("Expected 2 callbacks to be called, got %d", len(callbacksCalled))
		}
		if callbacksCalled[0] != "startRecording" || callbacksCalled[1] != "toggleStreaming" {
			t.Errorf("Unexpected callbacks called: %v", callbacksCalled)
		}
	})

	t.Run("SetupHotkeyCallbacks_NoManager", func(t *testing.T) {
		service := NewHotkeyService(mockLogger, nil)

		err := service.SetupHotkeyCallbacks(
			func() error { return nil },
			func() error { return nil },
			func() error { return nil },
			func() error { return nil },
			func() error { return nil },
			func() error { return nil },
			func() error { return nil },
		)
		if err == nil {
			t.Error("SetupHotkeyCallbacks should fail when hotkey manager is nil")
		}
	})

	t.Run("Shutdown_Success", func(t *testing.T) {
		mockManager := &mocks.MockHotkeyManager{}
		service := NewHotkeyService(mockLogger, mockManager)

		err := service.Shutdown()
		if err != nil {
			t.Errorf("Shutdown failed: %v", err)
		}
		if !mockManager.WasStopCalled() {
			t.Error("Hotkey manager Stop method was not called during shutdown")
		}
	})

	t.Run("Shutdown_NoManager", func(t *testing.T) {
		service := NewHotkeyService(mockLogger, nil)

		err := service.Shutdown()
		if err != nil {
			t.Errorf("Shutdown should not fail when hotkey manager is nil: %v", err)
		}
	})
}
