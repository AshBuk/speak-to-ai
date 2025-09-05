// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Package hotkeys provides hotkey management functionality with support for multiple providers
// and desktop environments. It handles global hotkey registration and event processing.
//
// Subpackages:
//   - interfaces: KeyboardEventProvider interface and related types
//   - providers: Provider implementations (dbus, evdev, dummy)
//   - manager: HotkeyManager and provider fallback logic
//   - adapters: Configuration adapters and utilities
//   - mocks: Mock implementations for testing
//   - utils: Utility functions for hotkeys
package hotkeys
