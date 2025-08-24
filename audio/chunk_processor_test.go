package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"sync"
	"testing"
	"time"
)

func TestNewChunkProcessor(t *testing.T) {
	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		UseVAD:          true,
		VADSensitivity:  VADMedium,
	}

	processor := NewChunkProcessor(config)

	if processor == nil {
		t.Fatal("NewChunkProcessor returned nil")
	}

	expectedChunkSize := (16000 * 64) / 1000 // 1024 samples
	if processor.chunkSize != expectedChunkSize {
		t.Errorf("expected chunkSize %d, got %d", expectedChunkSize, processor.chunkSize)
	}

	expectedDuration := 64 * time.Millisecond
	if processor.chunkDuration != expectedDuration {
		t.Errorf("expected chunkDuration %v, got %v", expectedDuration, processor.chunkDuration)
	}

	if processor.vad == nil {
		t.Error("expected VAD to be initialized when UseVAD is true")
	}

	if processor.sampleRate != 16000 {
		t.Errorf("expected sampleRate 16000, got %d", processor.sampleRate)
	}
}

func TestNewChunkProcessor_WithoutVAD(t *testing.T) {
	config := ChunkProcessorConfig{
		ChunkDurationMs: 32,
		SampleRate:      8000,
		UseVAD:          false,
	}

	processor := NewChunkProcessor(config)

	if processor.vad != nil {
		t.Error("expected VAD to be nil when UseVAD is false")
	}

	expectedChunkSize := (8000 * 32) / 1000 // 256 samples
	if processor.chunkSize != expectedChunkSize {
		t.Errorf("expected chunkSize %d, got %d", expectedChunkSize, processor.chunkSize)
	}
}

func TestChunkProcessor_GetChunkDuration(t *testing.T) {
	config := ChunkProcessorConfig{
		ChunkDurationMs: 125,
		SampleRate:      16000,
	}

	processor := NewChunkProcessor(config)
	duration := processor.GetChunkDuration()

	expectedDuration := 125 * time.Millisecond
	if duration != expectedDuration {
		t.Errorf("expected duration %v, got %v", expectedDuration, duration)
	}
}

func TestChunkProcessor_GetChunkSize(t *testing.T) {
	config := ChunkProcessorConfig{
		ChunkDurationMs: 100,
		SampleRate:      8000,
	}

	processor := NewChunkProcessor(config)
	size := processor.GetChunkSize()

	expectedSize := (8000 * 100) / 1000 // 800 samples
	if size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, size)
	}
}

func TestChunkProcessor_bytesToFloat32(t *testing.T) {
	processor := NewChunkProcessor(ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
	})

	tests := []struct {
		name     string
		input    []byte
		expected []float32
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: []float32{},
		},
		{
			name:     "single float32",
			input:    float32ToBytes(1.5),
			expected: []float32{1.5},
		},
		{
			name:     "multiple float32",
			input:    append(float32ToBytes(1.0), float32ToBytes(-0.5)...),
			expected: []float32{1.0, -0.5},
		},
		{
			name:     "incomplete bytes",
			input:    []byte{0x00, 0x01, 0x02}, // Only 3 bytes, need 4 for float32
			expected: []float32{},              // Should return empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.bytesToFloat32(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if i < len(result) {
					// For the incomplete bytes test, we just check it doesn't panic
					if tt.name == "incomplete bytes" {
						continue
					}
					if result[i] != expected {
						t.Errorf("at index %d: expected %f, got %f", i, expected, result[i])
					}
				}
			}
		})
	}
}

func TestChunkProcessor_Reset(t *testing.T) {
	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		UseVAD:          true,
		VADSensitivity:  VADMedium,
	}

	processor := NewChunkProcessor(config)

	// Add some data to buffer
	processor.buffer = append(processor.buffer, 1.0, 2.0, 3.0)

	// Reset
	processor.Reset()

	if len(processor.buffer) != 0 {
		t.Errorf("expected empty buffer after reset, got %d items", len(processor.buffer))
	}
}

func TestChunkProcessor_processChunk_OnlyOnChunk(t *testing.T) {
	var receivedChunks [][]float32
	var mu sync.Mutex

	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		OnChunk: func(chunk []float32) error {
			mu.Lock()
			defer mu.Unlock()
			// Make a copy to avoid race conditions
			chunkCopy := make([]float32, len(chunk))
			copy(chunkCopy, chunk)
			receivedChunks = append(receivedChunks, chunkCopy)
			return nil
		},
		UseVAD: false,
	}

	processor := NewChunkProcessor(config)
	testChunk := []float32{1.0, 2.0, 3.0, 4.0}

	processor.processChunk(testChunk)

	// Give some time for goroutine to complete
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedChunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(receivedChunks))
		return
	}

	if len(receivedChunks[0]) != len(testChunk) {
		t.Errorf("expected chunk length %d, got %d", len(testChunk), len(receivedChunks[0]))
		return
	}

	for i, expected := range testChunk {
		if receivedChunks[0][i] != expected {
			t.Errorf("at index %d: expected %f, got %f", i, expected, receivedChunks[0][i])
		}
	}
}

func TestChunkProcessor_processChunk_WithVAD(t *testing.T) {
	var speechChunks [][]float32
	var mu sync.Mutex

	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		OnSpeech: func(chunk []float32) error {
			mu.Lock()
			defer mu.Unlock()
			chunkCopy := make([]float32, len(chunk))
			copy(chunkCopy, chunk)
			speechChunks = append(speechChunks, chunkCopy)
			return nil
		},
		UseVAD:         true,
		VADSensitivity: VADHigh, // Use high sensitivity for testing
	}

	processor := NewChunkProcessor(config)

	// Create loud audio that should trigger speech detection
	loudChunk := make([]float32, 1024)
	for i := range loudChunk {
		loudChunk[i] = 0.8
	}

	// Process multiple chunks to trigger VAD state changes
	for i := 0; i < 5; i++ {
		processor.processChunk(loudChunk)
		time.Sleep(5 * time.Millisecond) // Small delay for goroutines
	}

	// Give time for goroutines to complete
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Should have received some speech chunks
	if len(speechChunks) == 0 {
		t.Error("expected to receive speech chunks with loud audio")
	}
}

func TestChunkProcessor_processChunk_WithoutVAD(t *testing.T) {
	var speechChunks [][]float32
	var mu sync.Mutex

	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		OnSpeech: func(chunk []float32) error {
			mu.Lock()
			defer mu.Unlock()
			chunkCopy := make([]float32, len(chunk))
			copy(chunkCopy, chunk)
			speechChunks = append(speechChunks, chunkCopy)
			return nil
		},
		UseVAD: false, // No VAD - all chunks should be treated as speech
	}

	processor := NewChunkProcessor(config)
	testChunk := []float32{0.1, 0.2, 0.3}

	processor.processChunk(testChunk)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(speechChunks) != 1 {
		t.Errorf("expected 1 speech chunk without VAD, got %d", len(speechChunks))
	}
}

func TestChunkProcessor_ProcessStream(t *testing.T) {
	var receivedChunks [][]float32
	var mu sync.Mutex

	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		OnChunk: func(chunk []float32) error {
			mu.Lock()
			defer mu.Unlock()
			chunkCopy := make([]float32, len(chunk))
			copy(chunkCopy, chunk)
			receivedChunks = append(receivedChunks, chunkCopy)
			return nil
		},
		UseVAD: false,
	}

	processor := NewChunkProcessor(config)

	// Create test audio data
	testSamples := []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0}
	var buf bytes.Buffer
	for _, sample := range testSamples {
		binary.Write(&buf, binary.LittleEndian, sample)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := processor.ProcessStream(ctx, &buf)

	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("unexpected error: %v", err)
	}

	// Give time for processing
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedChunks) == 0 {
		t.Error("expected to receive at least one chunk")
	}
}

func TestChunkProcessor_ProcessStream_ContextCancellation(t *testing.T) {
	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		UseVAD:          false,
	}

	processor := NewChunkProcessor(config)

	// Create a stream that will never end
	reader := &infiniteReader{}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err := processor.ProcessStream(ctx, reader)

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestChunkProcessor_ProcessStream_EOF(t *testing.T) {
	var receivedChunks [][]float32
	var mu sync.Mutex

	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		OnChunk: func(chunk []float32) error {
			mu.Lock()
			defer mu.Unlock()
			chunkCopy := make([]float32, len(chunk))
			copy(chunkCopy, chunk)
			receivedChunks = append(receivedChunks, chunkCopy)
			return nil
		},
		UseVAD: false,
	}

	processor := NewChunkProcessor(config)

	// Create small buffer that will trigger EOF
	testSamples := []float32{1.0, 2.0}
	var buf bytes.Buffer
	for _, sample := range testSamples {
		binary.Write(&buf, binary.LittleEndian, sample)
	}

	ctx := context.Background()
	err := processor.ProcessStream(ctx, &buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Give time for flush processing
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Should have received the remaining data as a chunk
	if len(receivedChunks) == 0 {
		t.Error("expected to receive at least one chunk from flush")
	}
}

// Helper function to convert float32 to bytes (little-endian)
func float32ToBytes(f float32) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, f)
	return buf.Bytes()
}

// infiniteReader for testing context cancellation
type infiniteReader struct{}

func (r *infiniteReader) Read(p []byte) (int, error) {
	// Fill with some test data
	for i := range p {
		p[i] = byte(i % 256)
	}
	return len(p), nil
}

func TestChunkProcessor_flushBuffer(t *testing.T) {
	var receivedChunks [][]float32
	var mu sync.Mutex

	config := ChunkProcessorConfig{
		ChunkDurationMs: 64,
		SampleRate:      16000,
		OnChunk: func(chunk []float32) error {
			mu.Lock()
			defer mu.Unlock()
			chunkCopy := make([]float32, len(chunk))
			copy(chunkCopy, chunk)
			receivedChunks = append(receivedChunks, chunkCopy)
			return nil
		},
		UseVAD: false,
	}

	processor := NewChunkProcessor(config)

	// Add some data to buffer (less than chunk size)
	processor.buffer = []float32{1.0, 2.0, 3.0}

	processor.flushBuffer()

	// Give time for goroutine
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedChunks) != 1 {
		t.Errorf("expected 1 chunk from flush, got %d", len(receivedChunks))
		return
	}

	if len(receivedChunks[0]) != 3 {
		t.Errorf("expected flushed chunk length 3, got %d", len(receivedChunks[0]))
	}

	// Buffer should be empty after flush
	if len(processor.buffer) != 0 {
		t.Errorf("expected empty buffer after flush, got %d items", len(processor.buffer))
	}
}
