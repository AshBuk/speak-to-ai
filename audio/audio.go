// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Package audio provides a high-level facade for audio recording functionality.
// It abstracts the underlying implementation details of recorder creation and management.
//
// Subpackages:
//   - interfaces: Define contracts (interfaces) for audio recording components
//   - recorders:  Provide concrete implementations of audio recorders (e.g., arecord, ffmpeg)
//   - factory:    Implement a factory for creating recorder instances based on configuration
//   - processing: Offer utilities for audio processing, such as temporary file management
//   - mocks:      Supply mock implementations for testing purposes
package audio
