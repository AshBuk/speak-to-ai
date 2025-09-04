// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

// Package audio provides audio recording functionality with support for multiple recording methods,
// audio processing, and factory patterns for creating recorders.
//
// Subpackages:
//   - interfaces: AudioRecorder interface and related types
//   - recorders: Implementation of audio recorders (arecord, ffmpeg, base)
//   - factory: Factory for creating audio recorders based on configuration
//   - processing: Audio processing utilities (VAD, chunk processing, temp file management)
//   - mocks: Mock implementations for testing
package audio
