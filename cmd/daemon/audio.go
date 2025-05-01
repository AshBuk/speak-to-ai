package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// AudioRecorder interface for audio recording
type AudioRecorder interface {
	StartRecording() error
	StopRecording() (string, error)
	GetOutputFile() string
	CleanupFile() error
}

// BaseRecorder implements common functionality for audio recorders
type BaseRecorder struct {
	config     *Config
	cmd        *exec.Cmd
	outputFile string
}

// GetOutputFile returns the path to the recorded audio file
func (b *BaseRecorder) GetOutputFile() string {
	return b.outputFile
}

// CleanupFile removes the temporary audio file
func (b *BaseRecorder) CleanupFile() error {
	if b.outputFile == "" {
		return nil // Nothing to clean up
	}

	if _, err := os.Stat(b.outputFile); os.IsNotExist(err) {
		return nil // File doesn't exist
	}

	return os.Remove(b.outputFile)
}

// createTempFile creates a temporary file for audio recording
func (b *BaseRecorder) createTempFile() error {
	// Create temporary file for recording
	tempDir := b.config.General.TempAudioPath
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Create unique filename
	timestamp := time.Now().Format("20060102-150405")
	b.outputFile = filepath.Join(tempDir, fmt.Sprintf("audio_%s.wav", timestamp))

	return nil
}

// stopProcess stops the recording process
func (b *BaseRecorder) stopProcess() error {
	if b.cmd == nil || b.cmd.Process == nil {
		return fmt.Errorf("recording not started")
	}

	// Send signal to terminate process
	err := b.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return fmt.Errorf("failed to stop recording: %w", err)
	}

	// Wait for process to complete
	err = b.cmd.Wait()
	if err != nil {
		// Ignore error as we deliberately interrupted the process
	}

	// Check that file was created
	if _, err := os.Stat(b.outputFile); os.IsNotExist(err) {
		return fmt.Errorf("audio file was not created")
	}

	return nil
}

// ArecordRecorder implements AudioRecorder using arecord
type ArecordRecorder struct {
	BaseRecorder
}

// NewArecordRecorder creates a new instance of ArecordRecorder
func NewArecordRecorder(config *Config) *ArecordRecorder {
	return &ArecordRecorder{
		BaseRecorder: BaseRecorder{
			config: config,
		},
	}
}

// StartRecording starts audio recording
func (a *ArecordRecorder) StartRecording() error {
	err := a.createTempFile()
	if err != nil {
		return err
	}

	// Prepare arecord command
	args := []string{
		"-D", a.config.Audio.Device,
		"-f", a.config.Audio.Format,
		"-r", fmt.Sprintf("%d", a.config.Audio.SampleRate),
		"-c", fmt.Sprintf("%d", a.config.Audio.Channels),
		a.outputFile,
	}

	// Start command
	a.cmd = exec.Command("arecord", args...)
	return a.cmd.Start()
}

// StopRecording stops audio recording and returns path to the file
func (a *ArecordRecorder) StopRecording() (string, error) {
	err := a.stopProcess()
	if err != nil {
		return "", err
	}

	return a.outputFile, nil
}

// FFmpegRecorder implements AudioRecorder using ffmpeg
type FFmpegRecorder struct {
	BaseRecorder
}

// NewFFmpegRecorder creates a new instance of FFmpegRecorder
func NewFFmpegRecorder(config *Config) *FFmpegRecorder {
	return &FFmpegRecorder{
		BaseRecorder: BaseRecorder{
			config: config,
		},
	}
}

// StartRecording starts audio recording
func (f *FFmpegRecorder) StartRecording() error {
	err := f.createTempFile()
	if err != nil {
		return err
	}

	// Prepare ffmpeg command
	args := []string{
		"-f", "pulse",
		"-i", f.config.Audio.Device,
		"-ac", fmt.Sprintf("%d", f.config.Audio.Channels),
		"-ar", fmt.Sprintf("%d", f.config.Audio.SampleRate),
		"-y", // Overwrite existing file
		f.outputFile,
	}

	f.cmd = exec.Command("ffmpeg", args...)
	return f.cmd.Start()
}

// StopRecording stops audio recording and returns path to the file
func (f *FFmpegRecorder) StopRecording() (string, error) {
	err := f.stopProcess()
	if err != nil {
		return "", err
	}

	return f.outputFile, nil
}

// GetRecorder returns the appropriate recorder based on configuration
func GetRecorder(config *Config) (AudioRecorder, error) {
	switch config.Audio.RecordingMethod {
	case "arecord":
		return NewArecordRecorder(config), nil
	case "ffmpeg":
		return NewFFmpegRecorder(config), nil
	default:
		return nil, fmt.Errorf("unsupported recording method: %s", config.Audio.RecordingMethod)
	}
}
