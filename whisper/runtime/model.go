// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package runtime

import (
	"fmt"
	"os"
	"runtime"

	// Low-level bindings
	whispercpp "github.com/ggerganov/whisper.cpp/bindings/go"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type model struct {
	path    string
	ctx     *whispercpp.Context
	backend BackendType
}

// Make sure model adheres to the interface
var _ Model = (*model)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New mirrors the upstream API but uses the default GPU-aware options.
func New(path string) (Model, error) {
	return NewWithOptions(path, DefaultOptions())
}

// NewWithOptions initialises a Whisper model with explicit runtime options.
func NewWithOptions(path string, opts Options) (Model, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	ctx, backend, err := initialiseContext(path, opts)
	if err != nil {
		return nil, err
	}

	model := &model{
		ctx:     ctx,
		path:    path,
		backend: backend,
	}

	// Return success.
	return model, nil
}

func (model *model) Close() error {
	if model.ctx != nil {
		model.ctx.Whisper_free()
	}

	// Release resources
	model.ctx = nil

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (model *model) String() string {
	str := "<whisper.model"
	if model.ctx != nil {
		str += fmt.Sprintf(" model=%q backend=%s", model.path, model.backend)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if model is multilingual (language and translation options are supported).
func (model *model) IsMultilingual() bool {
	return model.ctx.Whisper_is_multilingual() != 0
}

// Return all recognized languages. Initially it is set to auto-detect.
func (model *model) Languages() []string {
	result := make([]string, 0, whispercpp.Whisper_lang_max_id())
	for i := 0; i < whispercpp.Whisper_lang_max_id(); i++ {
		str := whispercpp.Whisper_lang_str(i)
		if model.ctx.Whisper_lang_id(str) >= 0 {
			result = append(result, str)
		}
	}
	return result
}

func (model *model) NewContext() (Context, error) {
	if model.ctx == nil {
		return nil, ErrInternalAppError
	}

	// Create new context
	params := model.ctx.Whisper_full_default_params(whispercpp.SAMPLING_GREEDY)
	params.SetTranslate(false)
	params.SetPrintSpecial(false)
	params.SetPrintProgress(false)
	params.SetPrintRealtime(false)
	params.SetPrintTimestamps(false)
	params.SetThreads(runtime.NumCPU())
	params.SetNoContext(true)

	// Return new context
	return newContext(model, params)
}

func (model *model) Backend() BackendType {
	return model.backend
}

///////////////////////////////////////////////////////////////////////////////
// HELPERS

func initialiseContext(path string, opts Options) (*whispercpp.Context, BackendType, error) {
	var lastErr error
	for _, backend := range orderedBackends(opts) {
		ctx, err := loadContext(path, backend, opts.GPUDevice)
		if err == nil {
			return ctx, backend, nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = ErrUnableToLoadModel
	}
	return nil, BackendCPU, lastErr
}

func orderedBackends(opts Options) []BackendType {
	switch opts.Backend {
	case BackendCPU:
		return []BackendType{BackendCPU}
	case BackendVulkan:
		if opts.AllowFallback {
			return []BackendType{BackendVulkan, BackendCPU}
		}
		return []BackendType{BackendVulkan}
	default: // BackendAuto and unknowns
		if opts.AllowFallback {
			return []BackendType{BackendVulkan, BackendCPU}
		}
		return []BackendType{BackendVulkan}
	}
}
