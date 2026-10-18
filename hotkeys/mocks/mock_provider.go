// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package mocks

import (
	"errors"
	"fmt"
	"sync"
)

// MockHotkeyProvider implements KeyboardEventProvider interface for testing
type MockHotkeyProvider struct {
	mu                sync.RWMutex
	isStarted         bool
	isSupported       bool
	registeredHotkeys map[string]func(bool) error
	startError        error
	stopCalled        bool
}

// NewMockHotkeyProvider creates a new mock hotkey provider
func NewMockHotkeyProvider() *MockHotkeyProvider {
	return &MockHotkeyProvider{
		isSupported:       true,
		registeredHotkeys: make(map[string]func(bool) error),
	}
}

// Start simulates starting the hotkey provider
func (m *MockHotkeyProvider) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.startError != nil {
		return m.startError
	}
	if m.isStarted {
		return errors.New("hotkey provider already started")
	}
	m.isStarted = true
	return nil
}

// Stop simulates stopping the hotkey provider
func (m *MockHotkeyProvider) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isStarted {
		m.isStarted = false
		m.stopCalled = true
	}
}

// RegisterHotkey simulates registering an action, keyed by its stable id.
func (m *MockHotkeyProvider) RegisterHotkey(id, hotkey string, callback func(pressed bool) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isSupported {
		return errors.New("hotkey registration not supported")
	}
	if callback == nil {
		return errors.New("callback cannot be nil")
	}
	m.registeredHotkeys[id] = callback
	return nil
}

// IsSupported returns whether the provider is supported
func (m *MockHotkeyProvider) IsSupported() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isSupported
}

// Test helper methods

// SetSupported configures whether the provider is supported
func (m *MockHotkeyProvider) SetSupported(supported bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isSupported = supported
}

// SetStartError configures the mock to return an error on Start
func (m *MockHotkeyProvider) SetStartError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startError = err
}

// IsStarted returns whether the provider is started
func (m *MockHotkeyProvider) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isStarted
}

// WasStopCalled returns whether Stop was called
func (m *MockHotkeyProvider) WasStopCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stopCalled
}

// SimulateHotkeyPress simulates pressing a registered hotkey
func (m *MockHotkeyProvider) SimulateHotkeyPress(hotkey string) error {
	m.mu.RLock()
	callback, exists := m.registeredHotkeys[hotkey]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("hotkey %s not registered", hotkey)
	}
	return callback(true)
}
