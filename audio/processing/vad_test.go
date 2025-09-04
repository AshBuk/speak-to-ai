// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package processing

import (
	"math"
	"testing"
)

func TestParseVADSensitivity(t *testing.T) {
	tests := []struct {
		input    string
		expected VADSensitivity
	}{
		{"low", VADLow},
		{"high", VADHigh},
		{"medium", VADMedium},
		{"invalid", VADMedium},
		{"", VADMedium},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseVADSensitivity(tt.input)
			if result != tt.expected {
				t.Errorf("ParseVADSensitivity(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewVAD(t *testing.T) {
	vad := NewVAD()

	if vad == nil {
		t.Fatal("NewVAD() returned nil")
	}

	// Check default values
	if vad.windowSize != 1024 {
		t.Errorf("expected windowSize 1024, got %d", vad.windowSize)
	}
	if vad.historySize != 10 {
		t.Errorf("expected historySize 10, got %d", vad.historySize)
	}
	if vad.currentState {
		t.Error("expected initial state to be false")
	}
	if vad.frameCounter != 0 {
		t.Errorf("expected initial frameCounter 0, got %d", vad.frameCounter)
	}
}

func TestNewVADWithSensitivity(t *testing.T) {
	tests := []struct {
		sensitivity       VADSensitivity
		expectedThreshold float64
		expectedSilence   int
		expectedSpeech    int
	}{
		{VADLow, 0.002, 12, 5},
		{VADMedium, 0.001, 8, 3},
		{VADHigh, 0.0005, 5, 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.sensitivity), func(t *testing.T) {
			vad := NewVADWithSensitivity(tt.sensitivity)

			if vad.threshold != tt.expectedThreshold {
				t.Errorf("expected threshold %f, got %f", tt.expectedThreshold, vad.threshold)
			}
			if vad.silenceDuration != tt.expectedSilence {
				t.Errorf("expected silenceDuration %d, got %d", tt.expectedSilence, vad.silenceDuration)
			}
			if vad.speechDuration != tt.expectedSpeech {
				t.Errorf("expected speechDuration %d, got %d", tt.expectedSpeech, vad.speechDuration)
			}
		})
	}
}

func TestVAD_IsSpeechActive_EmptyChunk(t *testing.T) {
	vad := NewVAD()
	result := vad.IsSpeechActive([]float32{})

	if result {
		t.Error("expected false for empty chunk")
	}
}

func TestVAD_IsSpeechActive_SilentAudio(t *testing.T) {
	vad := NewVAD()

	// Create silent audio (all zeros)
	silentChunk := make([]float32, 1024)

	// Test multiple frames to ensure consistent silence detection
	for i := 0; i < 10; i++ {
		result := vad.IsSpeechActive(silentChunk)
		if result {
			t.Errorf("frame %d: expected false for silent audio", i)
		}
	}
}

func TestVAD_IsSpeechActive_LoudAudio(t *testing.T) {
	vad := NewVADWithSensitivity(VADHigh) // Use high sensitivity for more reliable detection

	// Create loud audio
	loudChunk := make([]float32, 1024)
	for i := range loudChunk {
		loudChunk[i] = 0.5 // High amplitude
	}

	// Feed multiple frames to trigger speech detection
	var speechDetected bool
	for i := 0; i < 15; i++ { // Increase iterations to account for VAD state machine
		result := vad.IsSpeechActive(loudChunk)
		if result {
			speechDetected = true
			break
		}
	}

	if !speechDetected {
		t.Error("expected speech to be detected for loud audio")
	}
}

func TestVAD_calculateEnergy(t *testing.T) {
	vad := NewVAD()

	tests := []struct {
		name     string
		chunk    []float32
		expected float64
	}{
		{
			name:     "silent audio",
			chunk:    []float32{0, 0, 0, 0},
			expected: 0,
		},
		{
			name:     "unit amplitude",
			chunk:    []float32{1, 1, 1, 1},
			expected: 1,
		},
		{
			name:     "mixed amplitude",
			chunk:    []float32{0.5, -0.5, 0.5, -0.5},
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			energy := vad.calculateEnergy(tt.chunk)
			if math.Abs(energy-tt.expected) > 1e-10 {
				t.Errorf("calculateEnergy() = %f, want %f", energy, tt.expected)
			}
		})
	}
}

func TestVAD_calculateMean(t *testing.T) {
	vad := NewVAD()

	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "single value",
			values:   []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "multiple values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "negative values",
			values:   []float64{-2.0, -1.0, 0.0, 1.0, 2.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean := vad.calculateMean(tt.values)
			if math.Abs(mean-tt.expected) > 1e-10 {
				t.Errorf("calculateMean() = %f, want %f", mean, tt.expected)
			}
		})
	}
}

func TestVAD_calculateStd(t *testing.T) {
	vad := NewVAD()

	tests := []struct {
		name     string
		values   []float64
		mean     float64
		expected float64
	}{
		{
			name:     "single value",
			values:   []float64{5.0},
			mean:     5.0,
			expected: 0.0,
		},
		{
			name:     "constant values",
			values:   []float64{3.0, 3.0, 3.0},
			mean:     3.0,
			expected: 0.0,
		},
		{
			name:     "varying values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			mean:     3.0,
			expected: math.Sqrt(2.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std := vad.calculateStd(tt.values, tt.mean)
			if math.Abs(std-tt.expected) > 1e-10 {
				t.Errorf("calculateStd() = %f, want %f", std, tt.expected)
			}
		})
	}
}

func TestVAD_updateThreshold(t *testing.T) {
	vad := NewVAD()

	// Initially no history
	if len(vad.energyHistory) != 0 {
		t.Error("expected empty energy history initially")
	}

	// Add first energy value
	vad.updateThreshold(0.1)
	if len(vad.energyHistory) != 1 {
		t.Errorf("expected 1 energy value, got %d", len(vad.energyHistory))
	}

	// Add more values to trigger adaptive threshold
	vad.updateThreshold(0.2)
	vad.updateThreshold(0.3)

	if len(vad.energyHistory) != 3 {
		t.Errorf("expected 3 energy values, got %d", len(vad.energyHistory))
	}

	// Threshold should be updated with minimum constraint
	if vad.threshold < 0.0005 {
		t.Errorf("threshold %f should not be less than minimum 0.0005", vad.threshold)
	}

	// Fill up to history size limit
	for i := 0; i < 10; i++ {
		vad.updateThreshold(0.1)
	}

	if len(vad.energyHistory) > vad.historySize {
		t.Errorf("energy history size %d exceeds limit %d", len(vad.energyHistory), vad.historySize)
	}
}

func TestVAD_Reset(t *testing.T) {
	vad := NewVAD()

	// Set up some state
	vad.currentState = true
	vad.frameCounter = 5
	vad.energyHistory = []float64{0.1, 0.2, 0.3}

	// Reset
	vad.Reset()

	// Check state is cleared
	if vad.currentState {
		t.Error("expected currentState to be false after reset")
	}
	if vad.frameCounter != 0 {
		t.Errorf("expected frameCounter 0 after reset, got %d", vad.frameCounter)
	}
	if len(vad.energyHistory) != 0 {
		t.Errorf("expected empty energy history after reset, got %d items", len(vad.energyHistory))
	}
}

func TestVAD_GetCurrentState(t *testing.T) {
	vad := NewVAD()

	// Initial state should be false
	if vad.GetCurrentState() {
		t.Error("expected initial state to be false")
	}

	// Manually set state and verify getter
	vad.currentState = true
	if !vad.GetCurrentState() {
		t.Error("expected state to be true")
	}
}

func TestVAD_StateMachine(t *testing.T) {
	vad := NewVADWithSensitivity(VADHigh) // Use high sensitivity for faster transitions

	// Create loud audio that should trigger speech
	loudChunk := make([]float32, 1024)
	for i := range loudChunk {
		loudChunk[i] = 0.8
	}

	// Create silent audio
	silentChunk := make([]float32, 1024)

	// Should start in silence
	if vad.GetCurrentState() {
		t.Error("should start in silence state")
	}

	// Feed loud audio until speech is detected
	speechDetected := false
	for i := 0; i < 10; i++ {
		if vad.IsSpeechActive(loudChunk) {
			speechDetected = true
			break
		}
	}

	if !speechDetected {
		t.Error("speech should be detected after feeding loud audio")
	}

	// Feed silent audio until speech ends
	speechEnded := false
	for i := 0; i < 10; i++ {
		if !vad.IsSpeechActive(silentChunk) {
			speechEnded = true
			break
		}
	}

	if !speechEnded {
		t.Error("speech should end after feeding silent audio")
	}
}
