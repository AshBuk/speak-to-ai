package whisper

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/AshBuk/speak-to-ai/config"
)

// WhisperEngine represents an interface for working with whisper
type WhisperEngine struct {
	config     *config.Config
	whisperBin string
	modelPath  string
}

// NewWhisperEngine creates a new instance of WhisperEngine
func NewWhisperEngine(config *config.Config, whisperBin, modelPath string) *WhisperEngine {
	return &WhisperEngine{
		config:     config,
		whisperBin: whisperBin,
		modelPath:  modelPath,
	}
}

// validatePaths checks that file paths are safe and exist
func (w *WhisperEngine) validatePaths() error {
	// Check whisper binary
	if !isValidExecutable(w.whisperBin) {
		return fmt.Errorf("whisper binary not found or not executable: %s", w.whisperBin)
	}

	// Check model path
	if !isValidFile(w.modelPath) {
		return fmt.Errorf("whisper model not found: %s", w.modelPath)
	}

	return nil
}

// Transcribe performs speech recognition from an audio file
func (w *WhisperEngine) Transcribe(audioFile string) (string, error) {
	// Validate paths
	if err := w.validatePaths(); err != nil {
		return "", err
	}

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

	// Prepare arguments for whisper
	args := []string{
		"-m", w.modelPath,
		"-f", audioFile,
		"--output-txt",
	}

	// If language is specified, add it
	if lang := w.config.General.Language; lang != "" && lang != "auto" {
		args = append(args, "-l", lang)
	}

	// Start the process
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(w.whisperBin, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running whisper: %w, stderr: %s", err, errBuf.String())
	}

	// Get the result
	transcript := outBuf.String()
	transcript = cleanTranscript(transcript)

	return transcript, nil
}

// cleanTranscript cleans the result from service information
func cleanTranscript(text string) string {
	// Remove timestamp markers in format [00:00:00.000 --> 00:00:00.000]
	lines := strings.Split(text, "\n")
	result := []string{}

	for _, line := range lines {
		// Skip empty lines and lines with timestamps
		if line == "" || strings.HasPrefix(line, "[") {
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, " ")
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

// isValidExecutable checks if a file exists and is executable
func isValidExecutable(path string) bool {
	if !isValidFile(path) {
		return false
	}

	// Try to get file info to check permissions
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if file has execute permission
	// 0100 in octal is the execute bit for user
	return info.Mode()&0100 != 0
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
