// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package runtime

// BackendType identifies the preferred execution backend for the Whisper model.
type BackendType string

const (
	// BackendAuto prefers GPU execution when available and falls back to CPU.
	BackendAuto BackendType = "auto"
	// BackendCPU forces CPU execution.
	BackendCPU BackendType = "cpu"
	// BackendVulkan requests the Vulkan backend explicitly.
	BackendVulkan BackendType = "vulkan"
)

// Options configure how a Whisper model should be initialised.
type Options struct {
	// Backend selects the preferred backend. Defaults to BackendAuto.
	Backend BackendType
	// GPUDevice allows selecting a GPU device (for CUDA style backends). Use -1 for auto.
	GPUDevice int
	// AllowFallback toggles automatic fallback to CPU when the preferred backend is unavailable.
	AllowFallback bool
}

// DefaultOptions returns the recommended default options.
func DefaultOptions() Options {
	return Options{
		Backend:       BackendAuto,
		GPUDevice:     -1,
		AllowFallback: true,
	}
}
