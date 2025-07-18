package whisper

import (
	"os"
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
)

func TestCleanTranscript(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean text without timestamps",
			input:    "Hello world\nThis is a test",
			expected: "Hello world\nThis is a test",
		},
		{
			name:     "text with timestamps",
			input:    "[00:00:00.000 --> 00:00:02.000]\nHello world\n[00:00:02.000 --> 00:00:04.000]\nThis is a test",
			expected: "[00:00:00.000 --> 00:00:02.000]\nHello world\n[00:00:02.000 --> 00:00:04.000]\nThis is a test",
		},
		{
			name:     "text with empty lines",
			input:    "Hello world\n\nThis is a test\n\n",
			expected: "Hello world\n\nThis is a test\n\n",
		},
		{
			name:     "mixed timestamps and empty lines",
			input:    "[00:00:00.000 --> 00:00:02.000]\nHello world\n\n[00:00:02.000 --> 00:00:04.000]\n\nThis is a test\n",
			expected: "[00:00:00.000 --> 00:00:02.000]\nHello world\n\n[00:00:02.000 --> 00:00:04.000]\n\nThis is a test\n",
		},
		{
			name:     "only timestamps",
			input:    "[00:00:00.000 --> 00:00:02.000]\n[00:00:02.000 --> 00:00:04.000]",
			expected: "[00:00:00.000 --> 00:00:02.000]\n[00:00:02.000 --> 00:00:04.000]",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only empty lines",
			input:    "\n\n\n",
			expected: "\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanTranscript(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewWhisperEngine(t *testing.T) {
	config := &config.Config{}
	modelPath := "/non/existent/model.bin"

	// Test that NewWhisperEngine returns error for non-existent model
	engine, err := NewWhisperEngine(config, modelPath)
	if err == nil {
		t.Fatalf("Expected error for non-existent model, got nil")
	}

	if engine != nil {
		t.Errorf("Expected nil engine when model doesn't exist")
	}

	// Test error message
	expectedError := "whisper model not found: /non/existent/model.bin"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestIsValidFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_file")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing file",
			path:     tempFile.Name(),
			expected: true,
		},
		{
			name:     "non-existing file",
			path:     "/non/existing/file.txt",
			expected: false,
		},
		{
			name:     "path traversal attempt",
			path:     "../../../etc/passwd",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFile(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsValidExecutable removed - function no longer exists in new implementation

// TestWhisperEngine_validatePaths removed - method no longer exists in new implementation

func TestGetFileSize(t *testing.T) {
	// Create a temporary file with known content
	tempFile, err := os.CreateTemp("", "test_size")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := "Hello, World!"
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	tests := []struct {
		name         string
		path         string
		expectedSize int64
		expectError  bool
	}{
		{
			name:         "existing file",
			path:         tempFile.Name(),
			expectedSize: int64(len(content)),
			expectError:  false,
		},
		{
			name:         "non-existing file",
			path:         "/non/existing/file.txt",
			expectedSize: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := getFileSize(tt.path)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && size != tt.expectedSize {
				t.Errorf("expected size %d, got %d", tt.expectedSize, size)
			}
		})
	}
}

func TestCheckDiskSpace(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid path",
			path:        "/tmp",
			expectError: false,
		},
		{
			name:        "invalid path",
			path:        "/non/existing/path",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkDiskSpace(tt.path)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
