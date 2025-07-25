package audio

import (
	"math"
)

// VAD (Voice Activity Detection) provides simple energy-based voice detection
type VAD struct {
	threshold       float64 // Energy threshold for voice detection
	windowSize      int     // Window size for analysis (samples)
	energyHistory   []float64
	historySize     int
	silenceDuration int // Frames of silence before considering speech ended
	speechDuration  int // Frames of speech before considering speech started
	currentState    bool
	frameCounter    int
}

// NewVAD creates a new VAD instance
func NewVAD() *VAD {
	return &VAD{
		threshold:       0.001, // Adaptive threshold
		windowSize:      1024,  // 64ms at 16kHz
		historySize:     10,    // Keep last 10 energy values
		silenceDuration: 8,     // ~500ms of silence to end speech
		speechDuration:  3,     // ~200ms of speech to start
		energyHistory:   make([]float64, 0, 10),
		currentState:    false,
		frameCounter:    0,
	}
}

// IsSpeechActive analyzes audio chunk and returns if speech is detected
func (v *VAD) IsSpeechActive(audioChunk []float32) bool {
	if len(audioChunk) == 0 {
		return false
	}

	// Calculate energy for this chunk
	energy := v.calculateEnergy(audioChunk)

	// Update adaptive threshold
	v.updateThreshold(energy)

	// Determine if current frame has speech
	hasSpeech := energy > v.threshold

	// State machine for speech detection
	if v.currentState {
		// Currently in speech state
		if hasSpeech {
			v.frameCounter = 0 // Reset silence counter
		} else {
			v.frameCounter++
			if v.frameCounter >= v.silenceDuration {
				v.currentState = false // End speech
				v.frameCounter = 0
			}
		}
	} else {
		// Currently in silence state
		if hasSpeech {
			v.frameCounter++
			if v.frameCounter >= v.speechDuration {
				v.currentState = true // Start speech
				v.frameCounter = 0
			}
		} else {
			v.frameCounter = 0 // Reset speech counter
		}
	}

	return v.currentState
}

// calculateEnergy computes RMS energy of audio chunk
func (v *VAD) calculateEnergy(audioChunk []float32) float64 {
	var sum float64
	for _, sample := range audioChunk {
		sum += float64(sample * sample)
	}
	return math.Sqrt(sum / float64(len(audioChunk)))
}

// updateThreshold updates adaptive threshold based on energy history
func (v *VAD) updateThreshold(energy float64) {
	// Add to history
	v.energyHistory = append(v.energyHistory, energy)
	if len(v.energyHistory) > v.historySize {
		v.energyHistory = v.energyHistory[1:]
	}

	// Calculate adaptive threshold as mean + 2*std of recent energy
	if len(v.energyHistory) >= 3 {
		mean := v.calculateMean(v.energyHistory)
		std := v.calculateStd(v.energyHistory, mean)
		v.threshold = mean + 2*std

		// Ensure minimum threshold
		if v.threshold < 0.0005 {
			v.threshold = 0.0005
		}
	}
}

// calculateMean computes mean of energy history
func (v *VAD) calculateMean(values []float64) float64 {
	var sum float64
	for _, val := range values {
		sum += val
	}
	return sum / float64(len(values))
}

// calculateStd computes standard deviation
func (v *VAD) calculateStd(values []float64, mean float64) float64 {
	var sum float64
	for _, val := range values {
		diff := val - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(values)))
}

// Reset resets VAD state
func (v *VAD) Reset() {
	v.currentState = false
	v.frameCounter = 0
	v.energyHistory = v.energyHistory[:0]
}

// GetCurrentState returns current speech detection state
func (v *VAD) GetCurrentState() bool {
	return v.currentState
}
