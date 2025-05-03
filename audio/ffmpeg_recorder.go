package audio

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/AshBuk/speak-to-ai/config"
)

// FFmpegRecorder implements AudioRecorder using ffmpeg
type FFmpegRecorder struct {
	BaseRecorder
	streamingEnabled bool
	pipeReader       *io.PipeReader
	pipeWriter       *io.PipeWriter
	stdoutPipe       io.ReadCloser
}

// NewFFmpegRecorder creates a new instance of FFmpegRecorder
func NewFFmpegRecorder(config *config.Config) *FFmpegRecorder {
	return &FFmpegRecorder{
		BaseRecorder:     NewBaseRecorder(config),
		streamingEnabled: config.Audio.EnableStreaming,
	}
}

// UseStreaming indicates if this recorder supports streaming mode
func (f *FFmpegRecorder) UseStreaming() bool {
	return f.streamingEnabled
}

// GetAudioStream returns the audio stream for streaming mode
func (f *FFmpegRecorder) GetAudioStream() (io.Reader, error) {
	if !f.streamingEnabled || f.pipeReader == nil {
		return nil, fmt.Errorf("streaming not enabled or recorder not started")
	}
	return f.pipeReader, nil
}

// StartRecording starts audio recording
func (f *FFmpegRecorder) StartRecording() error {
	// Setup streaming pipe if needed
	if f.streamingEnabled {
		f.pipeReader, f.pipeWriter = io.Pipe()
	}

	// Basic arguments
	args := []string{
		"-f", "alsa",
		"-i", f.config.Audio.Device,
		"-ar", fmt.Sprintf("%d", f.config.Audio.SampleRate),
		"-ac", fmt.Sprintf("%d", f.config.Audio.Channels),
	}

	// Add quality settings
	args = append(args, "-q:a", "0")

	// Configure output
	if f.streamingEnabled || f.useBuffer {
		// Output to pipe for streaming or buffer
		args = append(args, "-f", "wav", "-")
	} else {
		// Output to file
		args = append(args, f.outputFile)
	}

	// Create command but don't start it yet if we need to capture stdout
	if f.streamingEnabled {
		// Create context with timeout
		f.mutex.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		f.ctx = ctx
		f.cancel = cancel

		// Create command with context
		f.cmd = exec.CommandContext(f.ctx, "ffmpeg", args...)

		// Get stdout pipe before starting the command
		var err error
		f.stdoutPipe, err = f.cmd.StdoutPipe()
		if err != nil {
			f.mutex.Unlock()
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		// Start the command
		if err := f.cmd.Start(); err != nil {
			f.mutex.Unlock()
			return fmt.Errorf("failed to start recording: %w", err)
		}
		f.mutex.Unlock()

		// Setup goroutine to copy from stdout pipe to our pipe writer
		go func() {
			defer f.pipeWriter.Close()
			buf := make([]byte, 4096)
			for {
				n, err := f.stdoutPipe.Read(buf)
				if n > 0 {
					f.pipeWriter.Write(buf[:n])
				}
				if err != nil {
					break
				}
			}
		}()

		return nil
	}

	// For non-streaming mode, use the template method
	return f.StartProcessTemplate("ffmpeg", args, f.useBuffer)
}

// StopRecording stops recording and returns the path to the recorded file
func (f *FFmpegRecorder) StopRecording() (string, error) {
	// Close pipe if streaming
	if f.streamingEnabled && f.pipeWriter != nil {
		defer f.pipeWriter.Close()
	}

	// Stop the recording process
	if err := f.StopProcess(); err != nil {
		return "", err
	}

	return f.outputFile, nil
}
