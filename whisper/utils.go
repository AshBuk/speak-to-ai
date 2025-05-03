package whisper

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DetectWhisperBinary finds the whisper binary in the system
func DetectWhisperBinary() (string, error) {
	// List of possible binary names
	possibleNames := []string{
		"whisper",
		"whisper.cpp",
		"whisper-cpp",
		"whisper-server",
	}

	// Try to find in PATH
	for _, name := range possibleNames {
		path, err := exec.LookPath(name)
		if err == nil {
			// Verify it's actually a whisper binary
			if isWhisperBinary(path) {
				return path, nil
			}
		}
	}

	// Try common locations
	locations := []string{
		"./sources/core/whisper",
		"./whisper",
		"/usr/local/bin/whisper",
		"/usr/bin/whisper",
	}

	for _, path := range locations {
		if isValidExecutable(path) && isWhisperBinary(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("whisper binary not found")
}

// isWhisperBinary checks if a file is a whisper binary
func isWhisperBinary(path string) bool {
	// First check if it's executable
	if !isValidExecutable(path) {
		return false
	}

	// Try running with --help or --version to see if it's a whisper binary
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var outBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, path, "--help")
	cmd.Stdout = &outBuf

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Try running with --version instead
		var versionBuf bytes.Buffer
		versionCmd := exec.CommandContext(ctx, path, "--version")
		versionCmd.Stdout = &versionBuf

		if err := versionCmd.Run(); err != nil {
			return false
		}

		output := versionBuf.String()
		return strings.Contains(strings.ToLower(output), "whisper")
	}

	output := outBuf.String()
	return strings.Contains(strings.ToLower(output), "whisper")
}

// CreateTempWavFile creates a temporary WAV file from raw PCM data
func CreateTempWavFile(data []byte, sampleRate int) (string, error) {
	// Create a temporary file
	tmp, err := os.CreateTemp("", "whisper-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempPath := tmp.Name()

	// Close the file to allow ffmpeg to write to it
	tmp.Close()

	// Use ffmpeg to create a WAV file
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-f", "s16le", // Input format (signed 16-bit little-endian PCM)
		"-ar", fmt.Sprintf("%d", sampleRate), // Sample rate
		"-ac", "1", // Mono
		"-i", "-", // Input from stdin
		"-y",     // Overwrite output file
		tempPath, // Output file
	)

	// Create pipe for input
	cmd.Stdin = bytes.NewReader(data)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %v, stderr: %s", err, errBuf.String())
	}

	return tempPath, nil
}

// GetWhisperVersion gets the version of the whisper binary
func GetWhisperVersion(whisperBin string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var outBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, whisperBin, "--version")
	cmd.Stdout = &outBuf

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error getting whisper version: %w", err)
	}

	return strings.TrimSpace(outBuf.String()), nil
}
