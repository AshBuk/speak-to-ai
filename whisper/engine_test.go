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
			expected: "Hello world This is a test",
		},
		{
			name:     "text with timestamps",
			input:    "[00:00:00.000 --> 00:00:02.000]\nHello world\n[00:00:02.000 --> 00:00:04.000]\nThis is a test",
			expected: "Hello world This is a test",
		},
		{
			name:     "text with empty lines",
			input:    "Hello world\n\nThis is a test\n\n",
			expected: "Hello world This is a test",
		},
		{
			name:     "mixed timestamps and empty lines",
			input:    "[00:00:00.000 --> 00:00:02.000]\nHello world\n\n[00:00:02.000 --> 00:00:04.000]\n\nThis is a test\n",
			expected: "Hello world This is a test",
		},
		{
			name:     "only timestamps",
			input:    "[00:00:00.000 --> 00:00:02.000]\n[00:00:02.000 --> 00:00:04.000]",
			expected: "",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only empty lines",
			input:    "\n\n\n",
			expected: "",
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
	whisperBin := "/usr/bin/whisper"
	modelPath := "/path/to/model.bin"

	engine := NewWhisperEngine(config, whisperBin, modelPath)

	if engine.config != config {
		t.Errorf("expected config to be set correctly")
	}
	if engine.whisperBin != whisperBin {
		t.Errorf("expected whisperBin to be %s, got %s", whisperBin, engine.whisperBin)
	}
	if engine.modelPath != modelPath {
		t.Errorf("expected modelPath to be %s, got %s", modelPath, engine.modelPath)
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

func TestIsValidExecutable(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "common executable",
			path:     "/bin/ls",
			expected: true,
		},
		{
			name:     "non-existing executable",
			path:     "/non/existing/executable",
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
			result := isValidExecutable(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWhisperEngine_validatePaths(t *testing.T) {
	// Create temporary files for testing
	tempModel, err := os.CreateTemp("", "model*.bin")
	if err != nil {
		t.Fatalf("failed to create temp model file: %v", err)
	}
	defer os.Remove(tempModel.Name())
	tempModel.Close()

	tests := []struct {
		name        string
		whisperBin  string
		modelPath   string
		expectError bool
	}{
		{
			name:        "valid paths",
			whisperBin:  "/bin/ls", // Use ls as a valid executable for testing
			modelPath:   tempModel.Name(),
			expectError: false,
		},
		{
			name:        "invalid whisper binary",
			whisperBin:  "/non/existing/whisper",
			modelPath:   tempModel.Name(),
			expectError: true,
		},
		{
			name:        "invalid model path",
			whisperBin:  "/bin/ls",
			modelPath:   "/non/existing/model.bin",
			expectError: true,
		},
		{
			name:        "both paths invalid",
			whisperBin:  "/non/existing/whisper",
			modelPath:   "/non/existing/model.bin",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewWhisperEngine(&config.Config{}, tt.whisperBin, tt.modelPath)
			err := engine.validatePaths()

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

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
