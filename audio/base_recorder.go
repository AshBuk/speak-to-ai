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
	config      *config.Config
	cmd         *exec.Cmd
	outputFile  string
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.Mutex
	cmdTimeout  time.Duration
	buffer      *bytes.Buffer    // For in-memory recording
	useBuffer   bool             // Whether to use in-memory buffer instead of file
	tempManager *TempFileManager // For managing temp files
}

// NewBaseRecorder creates a new base recorder instance
func NewBaseRecorder(config *config.Config) BaseRecorder {
	// Calculate if we should use buffer based on recording settings
	// For small recordings (< 10 seconds), use memory buffer
	useBuffer := config.Audio.ExpectedDuration > 0 &&
		config.Audio.ExpectedDuration < 10 &&
		config.Audio.SampleRate <= 16000

	return BaseRecorder{
		config:      config,
		cmdTimeout:  60 * time.Second, // Default timeout
		useBuffer:   useBuffer,
		buffer:      bytes.NewBuffer(nil),
		tempManager: GetTempFileManager(), // Use the global temp file manager
	}
}

// GetOutputFile returns the path to the recorded audio file
func (b *BaseRecorder) GetOutputFile() string {
	return b.outputFile
}

// UseStreaming indicates if this recorder supports streaming mode
func (b *BaseRecorder) UseStreaming() bool {
	return false // Base implementation doesn't support streaming
}

// GetAudioStream returns the audio stream for streaming mode
func (b *BaseRecorder) GetAudioStream() (io.Reader, error) {
	return nil, fmt.Errorf("streaming not supported by this recorder")
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

	// Create command with context
	b.cmd = exec.CommandContext(b.ctx, cmdName, args...)

	// Set up output redirection if using buffer
	if b.useBuffer && outputPipe {
		b.cmd.Stdout = b.buffer
	}

	// Start the process
	if err := b.cmd.Start(); err != nil {
		b.cancel()
		return fmt.Errorf("failed to start recording: %w", err)
	}

	return nil
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
			break
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
