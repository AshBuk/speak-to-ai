// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package audio

import (
	"context"
	"io"
	"sync"
	"time"
	"unsafe"
)

// ChunkProcessor handles real-time audio chunk processing
type ChunkProcessor struct {
	chunkSize     int                   // Size of each chunk in samples
	chunkDuration time.Duration         // Duration of each chunk
	vad           *VAD                  // Voice activity detection
	onChunk       func([]float32) error // Callback for each audio chunk
	onSpeech      func([]float32) error // Callback for speech chunks only
	buffer        []float32             // Internal buffer for incomplete chunks
	mutex         sync.Mutex            // Protects buffer access
	sampleRate    int                   // Sample rate for timing calculations
}

// ChunkProcessorConfig configuration for chunk processor
type ChunkProcessorConfig struct {
	ChunkDurationMs int                   // Chunk duration in milliseconds
	SampleRate      int                   // Audio sample rate
	OnChunk         func([]float32) error // Called for every chunk
	OnSpeech        func([]float32) error // Called only for speech chunks
	UseVAD          bool                  // Enable voice activity detection
	VADSensitivity  VADSensitivity        // VAD sensitivity level
}

// NewChunkProcessor creates a new chunk processor
func NewChunkProcessor(config ChunkProcessorConfig) *ChunkProcessor {
	chunkSize := (config.SampleRate * config.ChunkDurationMs) / 1000

	processor := &ChunkProcessor{
		chunkSize:     chunkSize,
		chunkDuration: time.Duration(config.ChunkDurationMs) * time.Millisecond,
		onChunk:       config.OnChunk,
		onSpeech:      config.OnSpeech,
		buffer:        make([]float32, 0, chunkSize*2),
		sampleRate:    config.SampleRate,
	}

	if config.UseVAD {
		processor.vad = NewVADWithSensitivity(config.VADSensitivity)
	}

	return processor
}

// ProcessStream processes audio stream in real-time chunks
func (cp *ChunkProcessor) ProcessStream(ctx context.Context, stream io.Reader) error {
	// Buffer for reading from stream
	readBuffer := make([]byte, cp.chunkSize*4) // 4 bytes per float32

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read from stream
		n, err := stream.Read(readBuffer)
		if err != nil {
			if err == io.EOF {
				// Process remaining buffer
				cp.flushBuffer()
				return nil
			}
			return err
		}

		if n == 0 {
			continue
		}

		// Convert bytes to float32 samples
		samples := cp.bytesToFloat32(readBuffer[:n])

		// Add to internal buffer
		cp.mutex.Lock()
		cp.buffer = append(cp.buffer, samples...)
		cp.mutex.Unlock()

		// Process complete chunks
		cp.processBufferedChunks()
	}
}

// processBufferedChunks processes all complete chunks in buffer
func (cp *ChunkProcessor) processBufferedChunks() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	for len(cp.buffer) >= cp.chunkSize {
		// Extract chunk
		chunk := make([]float32, cp.chunkSize)
		copy(chunk, cp.buffer[:cp.chunkSize])

		// Remove processed samples from buffer
		cp.buffer = cp.buffer[cp.chunkSize:]

		// Process chunk asynchronously to avoid blocking
		go cp.processChunk(chunk)
	}
}

// processChunk processes a single audio chunk
func (cp *ChunkProcessor) processChunk(chunk []float32) {
	// Always call onChunk if provided
	if cp.onChunk != nil {
		if err := cp.onChunk(chunk); err != nil {
			// Log error but continue processing
			return
		}
	}

	// Check for speech if VAD is enabled
	if cp.vad != nil && cp.onSpeech != nil {
		if cp.vad.IsSpeechActive(chunk) {
			if err := cp.onSpeech(chunk); err != nil {
				// Log error but continue processing
				return
			}
		}
	} else if cp.onSpeech != nil {
		// If no VAD, treat all chunks as speech
		cp.onSpeech(chunk)
	}
}

// flushBuffer processes any remaining samples in buffer
func (cp *ChunkProcessor) flushBuffer() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if len(cp.buffer) > 0 {
		// Process remaining samples as final chunk
		chunk := make([]float32, len(cp.buffer))
		copy(chunk, cp.buffer)
		cp.buffer = cp.buffer[:0]

		go cp.processChunk(chunk)
	}
}

// bytesToFloat32 converts byte buffer to float32 samples
// Assumes little-endian 32-bit float format
func (cp *ChunkProcessor) bytesToFloat32(data []byte) []float32 {
	samples := make([]float32, len(data)/4)

	for i := 0; i < len(samples); i++ {
		offset := i * 4
		if offset+3 < len(data) {
			// Convert 4 bytes to float32 (little-endian)
			bits := uint32(data[offset]) |
				uint32(data[offset+1])<<8 |
				uint32(data[offset+2])<<16 |
				uint32(data[offset+3])<<24

			samples[i] = *(*float32)(unsafe.Pointer(&bits))
		}
	}

	return samples
}

// Reset resets the processor state
func (cp *ChunkProcessor) Reset() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	cp.buffer = cp.buffer[:0]
	if cp.vad != nil {
		cp.vad.Reset()
	}
}

// GetChunkDuration returns the duration of each chunk
func (cp *ChunkProcessor) GetChunkDuration() time.Duration {
	return cp.chunkDuration
}

// GetChunkSize returns the size of each chunk in samples
func (cp *ChunkProcessor) GetChunkSize() int {
	return cp.chunkSize
}
