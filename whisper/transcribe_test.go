package whisper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
)

func TestWhisperEngine_Transcribe_ValidPaths(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create mock whisper binary
	whisperBin := filepath.Join(tempDir, "whisper")
	whisperContent := `#!/bin/bash
echo "Hello, world!"
`
	err := os.WriteFile(whisperBin, []byte(whisperContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock whisper binary: %v", err)
	}

	// Create mock model file
	modelPath := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelPath, []byte("mock model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	// Create mock audio file
	audioFile := filepath.Join(tempDir, "audio.wav")
	err = os.WriteFile(audioFile, []byte("mock audio data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock audio file: %v", err)
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	cfg.General.Language = "en"

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	result, err := engine.Transcribe(audioFile)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "Hello, world!" {
		t.Errorf("Expected 'Hello, world!', got %q", result)
	}
}

func TestWhisperEngine_Transcribe_InvalidWhisperBinary(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	engine := NewWhisperEngine(cfg, "/non/existent/whisper", "/non/existent/model.bin")

	result, err := engine.Transcribe("/non/existent/audio.wav")

	if err == nil {
		t.Error("Expected error for invalid whisper binary")
	}

	if result != "" {
		t.Errorf("Expected empty result on error, got %q", result)
	}
}

func TestWhisperEngine_Transcribe_InvalidModelPath(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	// Use a real binary that exists
	engine := NewWhisperEngine(cfg, "/bin/echo", "/non/existent/model.bin")

	result, err := engine.Transcribe("/non/existent/audio.wav")

	if err == nil {
		t.Error("Expected error for invalid model path")
	}

	if result != "" {
		t.Errorf("Expected empty result on error, got %q", result)
	}
}

func TestWhisperEngine_Transcribe_InvalidAudioFile(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create mock whisper binary
	whisperBin := filepath.Join(tempDir, "whisper")
	whisperContent := `#!/bin/bash
echo "test"
`
	err := os.WriteFile(whisperBin, []byte(whisperContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock whisper binary: %v", err)
	}

	// Create mock model file
	modelPath := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelPath, []byte("mock model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	result, err := engine.Transcribe("/non/existent/audio.wav")

	if err == nil {
		t.Error("Expected error for invalid audio file")
	}

	if result != "" {
		t.Errorf("Expected empty result on error, got %q", result)
	}
}

func TestWhisperEngine_Transcribe_LargeAudioFile(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create mock whisper binary
	whisperBin := filepath.Join(tempDir, "whisper")
	whisperContent := `#!/bin/bash
echo "test"
`
	err := os.WriteFile(whisperBin, []byte(whisperContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock whisper binary: %v", err)
	}

	// Create mock model file
	modelPath := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelPath, []byte("mock model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	// Create large audio file (over 50MB)
	audioFile := filepath.Join(tempDir, "large_audio.wav")
	largeData := make([]byte, 51*1024*1024) // 51MB
	err = os.WriteFile(audioFile, largeData, 0644)
	if err != nil {
		t.Fatalf("Failed to create large audio file: %v", err)
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	result, err := engine.Transcribe(audioFile)

	if err == nil {
		t.Error("Expected error for large audio file")
	}

	if result != "" {
		t.Errorf("Expected empty result on error, got %q", result)
	}
}

func TestWhisperEngine_Transcribe_WithLanguage(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create mock whisper binary that echoes its arguments
	whisperBin := filepath.Join(tempDir, "whisper")
	whisperContent := `#!/bin/bash
echo "Arguments: $@"
echo "Test transcription"
`
	err := os.WriteFile(whisperBin, []byte(whisperContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock whisper binary: %v", err)
	}

	// Create mock model file
	modelPath := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelPath, []byte("mock model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	// Create mock audio file
	audioFile := filepath.Join(tempDir, "audio.wav")
	err = os.WriteFile(audioFile, []byte("mock audio data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock audio file: %v", err)
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	cfg.General.Language = "ru"

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	result, err := engine.Transcribe(audioFile)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should contain the transcription text
	if result != "Arguments: -m "+modelPath+" -f "+audioFile+" --output-txt -l ru Test transcription" {
		t.Errorf("Unexpected result: %q", result)
	}
}

func TestWhisperEngine_Transcribe_AutoLanguage(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create mock whisper binary that echoes its arguments
	whisperBin := filepath.Join(tempDir, "whisper")
	whisperContent := `#!/bin/bash
echo "Arguments: $@"
echo "Test transcription"
`
	err := os.WriteFile(whisperBin, []byte(whisperContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock whisper binary: %v", err)
	}

	// Create mock model file
	modelPath := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelPath, []byte("mock model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	// Create mock audio file
	audioFile := filepath.Join(tempDir, "audio.wav")
	err = os.WriteFile(audioFile, []byte("mock audio data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock audio file: %v", err)
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	cfg.General.Language = "auto"

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	result, err := engine.Transcribe(audioFile)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should not contain -l flag for auto language
	if result != "Arguments: -m "+modelPath+" -f "+audioFile+" --output-txt Test transcription" {
		t.Errorf("Unexpected result: %q", result)
	}
}

func TestWhisperEngine_Transcribe_WhisperFailure(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create mock whisper binary that fails
	whisperBin := filepath.Join(tempDir, "whisper")
	whisperContent := `#!/bin/bash
echo "Error: Failed to process" >&2
exit 1
`
	err := os.WriteFile(whisperBin, []byte(whisperContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock whisper binary: %v", err)
	}

	// Create mock model file
	modelPath := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelPath, []byte("mock model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	// Create mock audio file
	audioFile := filepath.Join(tempDir, "audio.wav")
	err = os.WriteFile(audioFile, []byte("mock audio data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock audio file: %v", err)
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	result, err := engine.Transcribe(audioFile)

	if err == nil {
		t.Error("Expected error when whisper binary fails")
	}

	if result != "" {
		t.Errorf("Expected empty result on error, got %q", result)
	}
}

func TestWhisperEngine_ValidatePaths_Success(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create mock whisper binary
	whisperBin := filepath.Join(tempDir, "whisper")
	err := os.WriteFile(whisperBin, []byte("#!/bin/bash\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock whisper binary: %v", err)
	}

	// Create mock model file
	modelPath := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelPath, []byte("mock model data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	err = engine.validatePaths()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestWhisperEngine_ValidatePaths_InvalidBinary(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	engine := NewWhisperEngine(cfg, "/non/existent/whisper", "/non/existent/model.bin")

	err := engine.validatePaths()

	if err == nil {
		t.Error("Expected error for invalid whisper binary")
	}
}

func TestWhisperEngine_ValidatePaths_InvalidModel(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	engine := NewWhisperEngine(cfg, "/bin/echo", "/non/existent/model.bin")

	err := engine.validatePaths()

	if err == nil {
		t.Error("Expected error for invalid model path")
	}
}

func TestWhisperEngine_Constructor(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	whisperBin := "/usr/bin/whisper"
	modelPath := "/path/to/model.bin"

	engine := NewWhisperEngine(cfg, whisperBin, modelPath)

	if engine == nil {
		t.Fatal("NewWhisperEngine returned nil")
	}

	if engine.config != cfg {
		t.Error("Config not set correctly")
	}

	if engine.whisperBin != whisperBin {
		t.Errorf("Expected whisperBin %q, got %q", whisperBin, engine.whisperBin)
	}

	if engine.modelPath != modelPath {
		t.Errorf("Expected modelPath %q, got %q", modelPath, engine.modelPath)
	}
}

func TestWhisperEngine_PathTraversal(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	// Test path traversal in whisper binary
	engine := NewWhisperEngine(cfg, "../../../bin/whisper", "/path/to/model.bin")

	err := engine.validatePaths()

	if err == nil {
		t.Error("Expected error for path traversal in whisper binary")
	}

	// Test path traversal in model path
	engine = NewWhisperEngine(cfg, "/bin/echo", "../../../etc/passwd")

	err = engine.validatePaths()

	if err == nil {
		t.Error("Expected error for path traversal in model path")
	}
}

func TestWhisperEngine_EmptyPaths(t *testing.T) {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)

	// Test empty whisper binary
	engine := NewWhisperEngine(cfg, "", "/path/to/model.bin")

	err := engine.validatePaths()

	if err == nil {
		t.Error("Expected error for empty whisper binary")
	}

	// Test empty model path
	engine = NewWhisperEngine(cfg, "/bin/echo", "")

	err = engine.validatePaths()

	if err == nil {
		t.Error("Expected error for empty model path")
	}
}

func TestWhisperEngine_Integration(t *testing.T) {
	// Skip integration test in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// This test requires actual whisper binary and model
	// Skip if not available
	if !isValidExecutable("whisper") {
		t.Skip("whisper binary not available, skipping integration test")
	}

	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	cfg.General.Language = "en"

	engine := NewWhisperEngine(cfg, "whisper", "model.bin")

	// This would require actual model file and audio file
	// For now, just test that the engine can be created
	if engine == nil {
		t.Error("Expected engine to be created")
	}
}
