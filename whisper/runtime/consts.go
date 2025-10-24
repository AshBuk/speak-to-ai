// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package runtime

import (
	"errors"

	// Low-level bindings
	whispercpp "github.com/ggerganov/whisper.cpp/bindings/go"
)

///////////////////////////////////////////////////////////////////////////////
// ERRORS

var (
	ErrUnableToLoadModel    = errors.New("unable to load model")
	ErrInternalAppError     = errors.New("internal application error")
	ErrProcessingFailed     = errors.New("processing failed")
	ErrUnsupportedLanguage  = errors.New("unsupported language")
	ErrModelNotMultilingual = errors.New("model is not multilingual")
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

// SampleRate is the sample rate of the audio data.
const SampleRate = whispercpp.SampleRate

// SampleBits is the number of bytes per sample.
const SampleBits = whispercpp.SampleBits
