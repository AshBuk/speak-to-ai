// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
)

// BaseRecorder implements common functionality for audio recorders
type BaseRecorder struct {
	config             *config.Config
	cmd                *exec.Cmd
	outputFile         string
	ctx                context.Context
	cancel             context.CancelFunc
	mutex              sync.Mutex
	cmdTimeout         time.Duration
	buffer             *bytes.Buffer      // For in-memory recording
	useBuffer          bool               // Whether to use in-memory buffer instead of file
	tempManager        *TempFileManager   // For managing temp files
	audioLevelCallback AudioLevelCallback // Callback for audio level updates
	currentAudioLevel  float64            // Current audio level (0.0 to 1.0)
	levelMutex         sync.RWMutex       // Mutex for audio level access

	// Streaming support
	streamingEnabled bool
	pipeReader       *io.PipeReader
	pipeWriter       *io.PipeWriter
	stdoutPipe       io.ReadCloser
	chunkProcessor   *ChunkProcessor
	audioChunks      chan []float32
	streamingActive  bool
}

// NewBaseRecorder creates a new base recorder instance
func NewBaseRecorder(config *config.Config) BaseRecorder {
	// Calculate if we should use buffer based on recording settings
	// For small recordings (< 10 seconds), use memory buffer
	useBuffer := config.Audio.ExpectedDuration > 0 &&
		config.Audio.ExpectedDuration < 10 &&
		config.Audio.SampleRate <= 16000

	return BaseRecorder{
		config:           config,
		cmdTimeout:       60 * time.Second, // Default timeout
		useBuffer:        useBuffer,
		buffer:           bytes.NewBuffer(nil),
		tempManager:      GetTempFileManager(), // Use the global temp file manager
		streamingEnabled: config.Audio.EnableStreaming,
		audioChunks:      make(chan []float32, 10), // Buffer for 10 chunks
		streamingActive:  false,
	}
}

// GetOutputFile returns the path to the recorded audio file
func (b *BaseRecorder) GetOutputFile() string {
	return b.outputFile
}

// UseStreaming indicates if this recorder supports streaming mode
func (b *BaseRecorder) UseStreaming() bool {
	return b.streamingEnabled
}

// GetAudioStream returns the audio stream for streaming mode
func (b *BaseRecorder) GetAudioStream() (io.Reader, error) {
	if b.streamingEnabled && b.pipeReader != nil {
		return b.pipeReader, nil
	}
	return b.buffer, nil
}

// SetAudioLevelCallback sets the callback for audio level monitoring
func (b *BaseRecorder) SetAudioLevelCallback(callback AudioLevelCallback) {
	b.levelMutex.Lock()
	defer b.levelMutex.Unlock()
	b.audioLevelCallback = callback
}

// GetAudioLevel returns the current audio level (0.0 to 1.0)
func (b *BaseRecorder) GetAudioLevel() float64 {
	b.levelMutex.RLock()
	defer b.levelMutex.RUnlock()
	return b.currentAudioLevel
}

// updateAudioLevel updates the current audio level and calls callback if set
func (b *BaseRecorder) updateAudioLevel(level float64) {
	b.levelMutex.Lock()
	b.currentAudioLevel = level
	callback := b.audioLevelCallback
	b.levelMutex.Unlock()

	if callback != nil {
		callback(level)
	}
}

// calculateAudioLevel calculates audio level from PCM data
func (b *BaseRecorder) calculateAudioLevel(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Assume 16-bit PCM data
	var sum int64 = 0
	samples := len(data) / 2

	for i := 0; i < len(data)-1; i += 2 {
		// Convert bytes to 16-bit signed integer
		sample := int16(data[i]) | int16(data[i+1])<<8
		sum += int64(sample) * int64(sample)
	}

	if samples == 0 {
		return 0.0
	}

	// Calculate RMS (Root Mean Square)
	rms := float64(sum) / float64(samples)
	rms = rms / (32768.0 * 32768.0) // Normalize to 0-1 range

	// Apply square root for RMS
	if rms > 0 {
		return rms * 10.0 // Amplify for better visibility
	}

	return 0.0
}

// CleanupFile removes the temporary audio file
func (b *BaseRecorder) CleanupFile() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.useBuffer {
		// Clear the buffer
		b.buffer.Reset()
		return nil
	}

	if b.outputFile == "" {
		return nil // Nothing to clean up
	}

	// Use temp file manager for cleanup
	b.tempManager.RemoveFile(b.outputFile, true)
	return nil
}

// createTempFile creates a temporary file for audio recording
func (b *BaseRecorder) createTempFile() error {
	if b.useBuffer {
		// Using in-memory buffer, no file needed
		b.buffer.Reset()
		return nil
	}

	// Create temporary file for recording
	tempDir := b.config.General.TempAudioPath
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create unique filename based on timestamp and random suffix
	timestamp := time.Now().Format("20060102-150405")
	b.outputFile = filepath.Join(tempDir, fmt.Sprintf("audio_%s.wav", timestamp))

	// Register with temp file manager
	b.tempManager.AddFile(b.outputFile)

	return nil
}

// StartProcessTemplate is a template method for starting recording processes
func (b *BaseRecorder) StartProcessTemplate(cmdName string, args []string, outputPipe bool) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Create context with timeout
	b.ctx, b.cancel = context.WithTimeout(context.Background(), b.cmdTimeout)

	if b.useBuffer {
		// Reset buffer
		b.buffer.Reset()
	} else {
		// Create temporary file if not using in-memory buffer
		if err := b.createTempFile(); err != nil {
			return err
		}
	}

	// Security: validate command and sanitize args before execution
	if !b.config.IsCommandAllowed(cmdName) {
		return fmt.Errorf("command not allowed: %s", cmdName)
	}
	safeArgs := config.SanitizeCommandArgs(args)

	// Create command with context
	b.cmd = exec.CommandContext(b.ctx, cmdName, safeArgs...)

	// Set up output redirection and audio level monitoring
	if b.useBuffer && outputPipe {
		// Create a pipe to monitor audio data
		stdout, err := b.cmd.StdoutPipe()
		if err != nil {
			b.cancel()
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		// Start goroutine to read data and monitor audio levels
		go b.monitorAudioLevel(stdout)
	}

	// Start the process
	if err := b.cmd.Start(); err != nil {
		b.cancel()
		return fmt.Errorf("failed to start recording: %w", err)
	}

	return nil
}

// StartStreamingRecording starts streaming recording and returns audio chunks channel
func (b *BaseRecorder) StartStreamingRecording() (<-chan []float32, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.streamingActive {
		return nil, fmt.Errorf("streaming recording already active")
	}

	if !b.streamingEnabled {
		return nil, fmt.Errorf("streaming not enabled for this recorder")
	}

	// Initialize chunk processor
	processorConfig := ChunkProcessorConfig{
		ChunkDurationMs: 1000, // 1 second chunks
		SampleRate:      b.config.Audio.SampleRate,
		UseVAD:          true,
		OnSpeech: func(chunk []float32) error {
			select {
			case b.audioChunks <- chunk:
				return nil
			default:
				// Channel full, skip this chunk
				return nil
			}
		},
	}

	b.chunkProcessor = NewChunkProcessor(processorConfig)
	b.streamingActive = true

	// Start processing audio stream
	go func() {
		defer func() {
			b.mutex.Lock()
			b.streamingActive = false
			close(b.audioChunks)
			b.mutex.Unlock()
		}()

		// Get audio stream
		stream, err := b.GetAudioStream()
		if err != nil {
			return
		}

		// Process stream
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := b.chunkProcessor.ProcessStream(ctx, stream); err != nil {
			log.Printf("Error processing audio stream: %v", err)
		}
	}()

	return b.audioChunks, nil
}

// StopStreamingRecording stops streaming recording
func (b *BaseRecorder) StopStreamingRecording() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if !b.streamingActive {
		return nil // Already stopped
	}

	if b.chunkProcessor != nil {
		b.chunkProcessor.Reset()
	}

	b.streamingActive = false
	return nil
}

// StopRecording provides common StopRecording implementation for all recorders
func (b *BaseRecorder) StopRecording() (string, error) {
	// Close streaming pipe if enabled
	if b.streamingEnabled && b.pipeWriter != nil {
		defer func() {
			if err := b.pipeWriter.Close(); err != nil {
				log.Printf("failed to close pipe writer: %v", err)
			}
		}()
	}

	// Stop the recording process
	if err := b.StopProcess(); err != nil {
		return "", err
	}

	return b.outputFile, nil
}

// monitorAudioLevel reads audio data and calculates levels
func (b *BaseRecorder) monitorAudioLevel(reader io.Reader) {
	buf := make([]byte, 4096) // Buffer for reading audio data

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading audio data: %v", err)
				}
				return
			}

			if n > 0 {
				// Write to buffer
				b.buffer.Write(buf[:n])

				// Calculate and update audio level
				level := b.calculateAudioLevel(buf[:n])
				b.updateAudioLevel(level)
			}
		}
	}
}

// StopProcess stops the recording process with proper cleanup and retries
func (b *BaseRecorder) StopProcess() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.cmd == nil || b.cmd.Process == nil {
		return fmt.Errorf("recording not started")
	}

	// Cancel the context to initiate graceful shutdown
	if b.cancel != nil {
		b.cancel()
	}

	// Try to terminate the process gracefully
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("Retry %d to stop recording process", i)
		}

		// Send signal to terminate process
		err = b.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			log.Printf("Warning: failed to interrupt process: %v", err)
			// Try with more force
			err = b.cmd.Process.Kill()
			if err != nil {
				log.Printf("Warning: failed to kill process: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		// Wait for process with timeout
		done := make(chan error, 1)
		waitDone := false

		go func() {
			done <- b.cmd.Wait()
		}()

		select {
		case err := <-done:
			waitDone = true
			if err != nil {
				log.Printf("Process exited with error: %v", err)
			}
		case <-time.After(500 * time.Millisecond):
			// Process still running, continue to next retry
			continue
		}

		if waitDone {
			break
		}
	}

	// If we're using a file, verify it was created
	if !b.useBuffer {
		if _, err := os.Stat(b.outputFile); os.IsNotExist(err) {
			return fmt.Errorf("audio file was not created")
		}
	}

	return nil
}

// ExecuteRecordingCommand executes a recording command with proper setup for streaming or file output
// This is the main method that inheritors should use instead of StartProcessTemplate
func (b *BaseRecorder) ExecuteRecordingCommand(cmdName string, args []string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Create context with timeout
	b.ctx, b.cancel = context.WithTimeout(context.Background(), b.cmdTimeout)

	// Setup streaming pipe if needed
	if b.streamingEnabled {
		b.pipeReader, b.pipeWriter = io.Pipe()
	}

	// Setup output file or buffer
	if b.useBuffer {
		b.buffer.Reset()
	} else {
		if err := b.createTempFile(); err != nil {
			return err
		}
	}

	// Security: validate command and sanitize args before execution
	if !b.config.IsCommandAllowed(cmdName) {
		return fmt.Errorf("command not allowed: %s", cmdName)
	}
	safeArgs := config.SanitizeCommandArgs(args)

	// Create command with context
	b.cmd = exec.CommandContext(b.ctx, cmdName, safeArgs...)

	// Handle streaming mode
	if b.streamingEnabled {
		// Get stdout pipe before starting the command
		var err error
		b.stdoutPipe, err = b.cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		// Start the command
		if err := b.cmd.Start(); err != nil {
			return fmt.Errorf("failed to start recording: %w", err)
		}

		// Setup goroutine to copy from stdout pipe to our pipe writer
		go func() {
			defer func() {
				if err := b.pipeWriter.Close(); err != nil {
					log.Printf("failed to close pipe writer in goroutine: %v", err)
				}
			}()
			buf := make([]byte, 4096)
			for {
				n, err := b.stdoutPipe.Read(buf)
				if n > 0 {
					if _, err := b.pipeWriter.Write(buf[:n]); err != nil {
						log.Printf("audio stream pipe write error: %v", err)
						break
					}
				}
				if err != nil {
					break
				}
			}
		}()

		return nil
	}

	// Handle buffer mode
	if b.useBuffer {
		// Create a pipe to monitor audio data
		stdout, err := b.cmd.StdoutPipe()
		if err != nil {
			b.cancel()
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		// Start goroutine to read data and monitor audio levels
		go b.monitorAudioLevel(stdout)
	}

	// Start the process
	if err := b.cmd.Start(); err != nil {
		b.cancel()
		return fmt.Errorf("failed to start recording: %w", err)
	}

	return nil
}
