// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package recorders

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/interfaces"
	"github.com/AshBuk/speak-to-ai/audio/processing"
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
	buffer             *bytes.Buffer                 // For in-memory recording
	useBuffer          bool                          // Whether to use in-memory buffer instead of file
	tempManager        *processing.TempFileManager   // For managing temp files
	audioLevelCallback interfaces.AudioLevelCallback // Callback for audio level updates
	currentAudioLevel  float64                       // Current audio level (0.0 to 1.0)
	levelMutex         sync.RWMutex                  // Mutex for audio level access

	// Streaming support
	streamingEnabled bool
	pipeReader       *io.PipeReader
	pipeWriter       *io.PipeWriter
	stdoutPipe       io.ReadCloser
	chunkProcessor   *processing.ChunkProcessor
	audioChunks      chan []float32
	streamingActive  bool

	// Diagnostics
	stderrBuf bytes.Buffer

	// Process synchronization
	waitOnce   sync.Once
	processErr error
}

// NewBaseRecorder creates a new base recorder instance
func NewBaseRecorder(config *config.Config) BaseRecorder {
	// Calculate if we should use buffer based on recording settings
	// For small recordings (< 10 seconds), use memory buffer
	useBuffer := config.Audio.ExpectedDuration > 0 &&
		config.Audio.ExpectedDuration < 10 &&
		config.Audio.SampleRate <= 16000

	// Determine command timeout from configuration
	maxTime := time.Duration(config.Audio.MaxRecordingTime) * time.Second
	if maxTime <= 0 {
		maxTime = 5 * time.Minute
	}

	return BaseRecorder{
		config:           config,
		cmdTimeout:       maxTime,
		useBuffer:        useBuffer,
		buffer:           bytes.NewBuffer(nil),
		tempManager:      processing.GetTempFileManager(), // Use the global temp file manager
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
func (b *BaseRecorder) SetAudioLevelCallback(callback interfaces.AudioLevelCallback) {
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

	path, err := b.tempManager.CreateTempWav(b.config.General.TempAudioPath)
	if err != nil {
		return fmt.Errorf("failed to create temp audio file: %w", err)
	}
	b.outputFile = path
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
	processorConfig := processing.ChunkProcessorConfig{
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

	b.chunkProcessor = processing.NewChunkProcessor(processorConfig)
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

// waitForProcess safely waits for the command to finish (can be called multiple times)
func (b *BaseRecorder) waitForProcess() error {
	if b.cmd == nil {
		return fmt.Errorf("no command to wait for")
	}

	b.waitOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				b.processErr = fmt.Errorf("panic in Wait(): %v", r)
			}
		}()
		b.processErr = b.cmd.Wait()
	})

	return b.processErr
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
			done <- b.waitForProcess()
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

	// Clean up process state after successful termination
	b.cmd = nil
	b.cancel = nil

	// Log any stderr captured for diagnostics
	if b.stderrBuf.Len() > 0 {
		cmdName := "recorder"
		if b.cmd != nil && b.cmd.Path != "" {
			cmdName = b.cmd.Path
		}
		log.Printf("%s stderr: %s", cmdName, b.stderrBuf.String())
	}

	// If we're using a file and not streaming, verify it was created
	if !b.useBuffer && !b.streamingEnabled {
		// Small delay to ensure buffers are flushed to disk
		time.Sleep(50 * time.Millisecond)
		log.Printf("[DEBUG] Checking audio file: %s", b.outputFile)
		info, err := os.Stat(b.outputFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("[DEBUG] Audio file does not exist: %s", b.outputFile)
				return fmt.Errorf("audio file was not created")
			}
			log.Printf("[DEBUG] Failed to stat audio file: %v", err)
			return fmt.Errorf("failed to stat audio file: %w", err)
		}
		log.Printf("[DEBUG] Audio file size: %d bytes", info.Size())
		// Minimal valid WAV header is 44 bytes
		if info.Size() <= 44 {
			log.Printf("[AUDIO ERROR] Audio file empty (size=%d) - likely recording failed", info.Size())
			log.Printf("[AUDIO HINT] Check audio device availability and permissions")
			return fmt.Errorf("audio file is empty or invalid (size=%d)", info.Size())
		}
		log.Printf("[DEBUG] Audio file validation successful: %d bytes", info.Size())
	}

	return nil
}

// ExecuteRecordingCommand executes a recording command with proper setup for streaming or file output
// This is the main method that inheritors should use instead of StartProcessTemplate
func (b *BaseRecorder) ExecuteRecordingCommand(cmdName string, args []string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Check if recording is already in progress
	if b.cmd != nil && b.cmd.Process != nil {
		return fmt.Errorf("recording already in progress")
	}

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

	// Add output file to args if needed (for file output mode)
	finalArgs := args
	if !b.useBuffer && !b.streamingEnabled {
		finalArgs = append(args, b.outputFile)
	}

	// Security: validate command and sanitize args before execution
	if !config.IsCommandAllowed(b.config, cmdName) {
		return fmt.Errorf("command not allowed: %s", cmdName)
	}
	safeArgs := config.SanitizeCommandArgs(finalArgs)

	// Create command with context
	// The command name is validated against an allowlist and arguments are sanitized above.
	// #nosec G204 -- Safe: allowlisted cmdName and sanitized args mitigate command injection.
	b.cmd = exec.CommandContext(b.ctx, cmdName, safeArgs...)

	// Capture stderr for diagnostics
	b.stderrBuf.Reset()
	b.cmd.Stderr = &b.stderrBuf

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

		// Log stderr after start in background for visibility
		go func(name string) {
			// Wait for process end to flush stderr
			_ = b.waitForProcess()
			if b.stderrBuf.Len() > 0 {
				log.Printf("%s stderr: %s", name, b.stderrBuf.String())
			}
		}(cmdName)

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

	// Log stderr after start in background for visibility
	go func(name string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[AUDIO ERROR] Panic in %s monitoring: %v", name, r)
			}
		}()

		// Wait for process end to flush stderr
		if b.cmd == nil {
			return
		}

		err := b.waitForProcess()
		if b.stderrBuf.Len() > 0 {
			log.Printf("[AUDIO ERROR] %s stderr: %s", name, b.stderrBuf.String())
		}
		if err != nil {
			log.Printf("[AUDIO ERROR] %s exited with error: %v", name, err)
			// Provide specific troubleshooting hints
			switch name {
			case "ffmpeg":
				log.Printf("[AUDIO HINT] FFmpeg failed - try switching to arecord via tray menu")
				log.Printf("[AUDIO HINT] Common cause: PulseAudio sources SUSPENDED in PipeWire")
			case "arecord":
				log.Printf("[AUDIO HINT] arecord failed - check microphone permissions and hardware")
			}
		}
	}(cmdName)

	return nil
}
