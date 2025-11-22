// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Package hotkeys provides a high-level facade for hotkey management
// It abstracts the underlying implementation of providers and event handling
//
// Subpackages:
//   - interfaces: Define contracts for hotkey providers
//   - providers:  Provide concrete implementations for different environments (D-Bus, evdev)
//   - manager:    Implement the core HotkeyManager logic, including provider selection and fallback
//   - adapters:   Provide adapters for configuration structs
//   - utils:      Offer utility functions for parsing and handling key combinations
//   - mocks:      Supply test doubles for end-to-end and integration scenarios
package hotkeys
