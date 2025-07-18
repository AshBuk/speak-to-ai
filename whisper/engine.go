package whisper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	// "github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// WhisperEngine represents an interface for working with whisper
type WhisperEngine struct {
	config    *config.Config
	model     whisper.Model
	modelPath string
}

// NewWhisperEngine creates a new instance of WhisperEngine
func NewWhisperEngine(config *config.Config, modelPath string) (*WhisperEngine, error) {
	// Validate model path
	if !isValidFile(modelPath) {
		return nil, fmt.Errorf("whisper model not found: %s", modelPath)
	}

	// Load the model with go-whisper bindings
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load whisper model: %w", err)
	}

	return &WhisperEngine{
		config:    config,
		model:     model,
		modelPath: modelPath,
	}, nil
}

// Close closes the whisper model and releases resources
func (w *WhisperEngine) Close() error {
	// Close the model with go-whisper bindings
	if w.model != nil {
		return w.model.Close()
	}
	return nil
}

// Transcribe performs speech recognition from an audio file
func (w *WhisperEngine) Transcribe(audioFile string) (string, error) {
	// Validate the audio file
	if !isValidFile(audioFile) {
		return "", fmt.Errorf("audio file not found or invalid: %s", audioFile)
	}

	// Check file size
	fileSize, err := getFileSize(audioFile)
	if err != nil {
		return "", fmt.Errorf("error checking audio file size: %w", err)
	}

	// Set a reasonable size limit
	const maxFileSize int64 = 50 * 1024 * 1024
	if fileSize > maxFileSize {
		return "", fmt.Errorf("audio file too large (%d bytes), max allowed is %d bytes", fileSize, maxFileSize)
	}

	// Check available disk space for output
	if err := checkDiskSpace(audioFile); err != nil {
		return "", fmt.Errorf("insufficient disk space: %w", err)
	}

	// Load audio data
	audioData, err := w.loadAudioData(audioFile)
	if err != nil {
		return "", fmt.Errorf("failed to load audio data: %w", err)
	}

	// Create context for transcription
	context, err := w.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("failed to create whisper context: %w", err)
	}
	// defer context.Close() // API may not have Close method

	// Set language if specified
	if lang := w.config.General.Language; lang != "" && lang != "auto" {
		if err := context.SetLanguage(lang); err != nil {
			return "", fmt.Errorf("failed to set language: %w", err)
		}
	}

	// Process audio data
	if err := context.Process(audioData, nil, nil, nil); err != nil {
		return "", fmt.Errorf("failed to process audio: %w", err)
	}

	// Extract transcript
	var transcript strings.Builder
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}
		transcript.WriteString(segment.Text)
		transcript.WriteString(" ")
	}

	result := strings.TrimSpace(transcript.String())
	return result, nil
}

// loadAudioData loads audio data from file and converts it to float32 samples
func (w *WhisperEngine) loadAudioData(audioFile string) ([]float32, error) {
	// Open the WAV file
	file, err := os.Open(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create WAV decoder
	decoder := wav.NewDecoder(file)
	if decoder == nil {
		return nil, fmt.Errorf("failed to create WAV decoder")
	}

	// Read the audio buffer
	audioBuffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to read audio buffer: %w", err)
	}

	// Convert to float32 samples
	// For now, convert manually from IntBuffer to float32
	samples := make([]float32, audioBuffer.NumFrames())
	for i := 0; i < audioBuffer.NumFrames(); i++ {
		// Convert int to float32 normalized to [-1.0, 1.0]
		intSample := audioBuffer.Data[i]
		samples[i] = float32(intSample) / 32768.0
	}
	return samples, nil
}

// isValidFile checks if a file exists and is accessible
func isValidFile(path string) bool {
	// Check for path traversal attempts
	clean := filepath.Clean(path)
	if clean != path {
		return false
	}

	// Check file existence and access
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// cleanTranscript cleans the transcript text (legacy function for tests)
func cleanTranscript(text string) string {
	// Simple implementation for backward compatibility with tests
	return text
}


// getFileSize returns the size of a file in bytes
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// checkDiskSpace ensures there's enough disk space available
func checkDiskSpace(path string) error {
	// Get directory stats
	dir := filepath.Dir(path)
	var stat syscall.Statfs_t
	err := syscall.Statfs(dir, &stat)
	if err != nil {
		return err
	}

	// Calculate available space
	available := stat.Bavail * uint64(stat.Bsize)

	// Require at least 100MB free
	const requiredSpace uint64 = 100 * 1024 * 1024
	if available < requiredSpace {
		return fmt.Errorf("insufficient disk space: %d bytes available, %d required", available, requiredSpace)
	}

	return nil
}
