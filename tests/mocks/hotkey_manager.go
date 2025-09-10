// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package mocks

import "github.com/AshBuk/speak-to-ai/hotkeys/manager"

// HotkeyManagerInterface defines the interface that MockHotkeyManager implements
type HotkeyManagerInterface interface {
	Start() error
	Stop()
	RegisterCallbacks(startRecording, stopRecording func() error)
	RegisterHotkeyAction(action string, callback manager.HotkeyAction)
}

// MockHotkeyManager implements HotkeyManagerInterface for testing
type MockHotkeyManager struct {
	startCalled                bool
	stopCalled                 bool
	registerCallbacksCalled    bool
	registerHotkeyActionCalled map[string]bool
	startError                 error
	callbacks                  struct {
		startRecording  func() error
		stopRecording   func() error
		toggleStreaming manager.HotkeyAction
		toggleVAD       manager.HotkeyAction
		switchModel     manager.HotkeyAction
		showConfig      manager.HotkeyAction
		reloadConfig    manager.HotkeyAction
	}
}

func (m *MockHotkeyManager) Start() error {
	m.startCalled = true
	return m.startError
}

func (m *MockHotkeyManager) Stop() {
	m.stopCalled = true
}

func (m *MockHotkeyManager) RegisterCallbacks(startRecording, stopRecording func() error) {
	m.registerCallbacksCalled = true
	m.callbacks.startRecording = startRecording
	m.callbacks.stopRecording = stopRecording
}

func (m *MockHotkeyManager) RegisterHotkeyAction(action string, callback manager.HotkeyAction) {
	if m.registerHotkeyActionCalled == nil {
		m.registerHotkeyActionCalled = make(map[string]bool)
	}
	m.registerHotkeyActionCalled[action] = true

	switch action {
	case "toggle_streaming":
		m.callbacks.toggleStreaming = callback
	case "toggle_vad":
		m.callbacks.toggleVAD = callback
	case "switch_model":
		m.callbacks.switchModel = callback
	case "show_config":
		m.callbacks.showConfig = callback
	case "reload_config":
		m.callbacks.reloadConfig = callback
	}
}

// Test helper methods
func (m *MockHotkeyManager) WasStartCalled() bool {
	return m.startCalled
}

func (m *MockHotkeyManager) WasStopCalled() bool {
	return m.stopCalled
}

func (m *MockHotkeyManager) WereCallbacksRegistered() bool {
	return m.registerCallbacksCalled
}

func (m *MockHotkeyManager) WasHotkeyActionRegistered(action string) bool {
	return m.registerHotkeyActionCalled[action]
}

func (m *MockHotkeyManager) TriggerCallback(callbackType string) error {
	switch callbackType {
	case "startRecording":
		if m.callbacks.startRecording != nil {
			return m.callbacks.startRecording()
		}
	case "toggleStreaming":
		if m.callbacks.toggleStreaming != nil {
			return m.callbacks.toggleStreaming()
		}
	}
	return nil
}

// SetStartError sets the error that will be returned by Start method
func (m *MockHotkeyManager) SetStartError(err error) {
	m.startError = err
}
