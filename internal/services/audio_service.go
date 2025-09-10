// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AshBuk/speak-to-ai/audio/factory"
	"github.com/AshBuk/speak-to-ai/audio/interfaces"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/constants"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/utils"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// AudioService implements AudioServiceInterface with full functionality
type AudioService struct {
	logger          logger.Logger
	config          *config.Config
	recorder        interfaces.AudioRecorder
	whisperEngine   *whisper.WhisperEngine
	streamingEngine *whisper.StreamingWhisperEngine
	modelManager    whisper.ModelManager

	// State management
	mu                       sync.RWMutex
	isRecording              bool
	lastTranscript           string
	audioRecorderNeedsReinit bool

	// Context for operations
	ctx    context.Context
	cancel context.CancelFunc

	// Dependencies
	ui  UIServiceInterface
	io  IOServiceInterface
	cfg ConfigServiceInterface
}

// NewAudioService creates a new AudioService instance
func NewAudioService(
	logger logger.Logger,
	config *config.Config,
	recorder interfaces.AudioRecorder,
	whisperEngine *whisper.WhisperEngine,
	streamingEngine *whisper.StreamingWhisperEngine,
	modelManager whisper.ModelManager,
) *AudioService {
	ctx, cancel := context.WithCancel(context.Background())

	return &AudioService{
		logger:          logger,
		config:          config,
		recorder:        recorder,
		whisperEngine:   whisperEngine,
		streamingEngine: streamingEngine,
		modelManager:    modelManager,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// SetDependencies sets service dependencies
func (as *AudioService) SetDependencies(ui UIServiceInterface, io IOServiceInterface) {
	as.ui = ui
	as.io = io
}

// SetConfig sets the config service dependency
func (as *AudioService) SetConfig(cfg ConfigServiceInterface) { as.cfg = cfg }

// HandleStartRecording starts audio recording
func (as *AudioService) HandleStartRecording() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.logger.Info("Starting recording...")

	// Ensure model is available
	if err := as.EnsureModelAvailable(); err != nil {
		as.logger.Error("Model not available: %v", err)
		as.setUIError(constants.MsgModelUnavailable)
		return fmt.Errorf("model not available: %w", err)
	}

	// Ensure audio recorder is available
	if err := as.EnsureAudioRecorderAvailable(); err != nil {
		as.logger.Error("Audio recorder not available: %v", err)
		as.setUIError(constants.MsgRecorderUnavailable)
		return fmt.Errorf("audio recorder not available: %w", err)
	}

	// Choose recording mode
	if as.config.Audio.EnableVAD && as.config.Audio.AutoStartStop {
		return as.HandleStartVADRecording()
	}

	if as.config.Audio.EnableStreaming && as.streamingEngine != nil {
		return as.HandleStartStreamingRecording()
	}

	// Standard recording
	return as.startStandardRecording()
}

// HandleStopRecording stops recording and starts transcription
func (as *AudioService) HandleStopRecording() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if !as.isRecording {
		return fmt.Errorf("no recording in progress")
	}

	as.logger.Info("Stopping recording and transcribing...")

	audioFile, err := as.recorder.StopRecording()
	if err != nil {
		as.logger.Warning("StopRecording returned error: %v", err)
		go as.handleRecordingError(err)

		// Auto-fallback to arecord if using ffmpeg
		if as.config.Audio.RecordingMethod == "ffmpeg" {
			// Persist change via ConfigService if available
			if as.cfg != nil {
				_ = as.cfg.UpdateRecordingMethod("arecord")
			}
			as.config.Audio.RecordingMethod = "arecord"
			as.audioRecorderNeedsReinit = true

			if as.ui != nil {
				as.ui.ShowNotification("Audio Fallback", "Switched to arecord due to ffmpeg capture error. Try recording again.")
				// Refresh tray to reflect new method
				as.ui.UpdateSettings(as.config)
			}
			as.logger.Info("Auto-fallback: switched to arecord due to ffmpeg failure")
		}

		// Ensure state is reset so the hotkey toggle can recover
		as.isRecording = false
		if as.ui != nil {
			as.ui.SetRecordingState(false)
		}

		// Swallow error to make stop idempotent and avoid being stuck
		return nil
	}

	as.isRecording = false

	// Update UI
	if as.ui != nil {
		as.ui.SetRecordingState(false)
		as.ui.ShowNotification(constants.NotifyRecordingStopped, constants.NotifyRecordingStopMsg)
	}

	// Start async transcription
	go as.transcribeAsync(audioFile)

	return nil
}

// IsRecording returns current recording state
func (as *AudioService) IsRecording() bool {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.isRecording
}

// GetLastTranscript returns the last transcription result
func (as *AudioService) GetLastTranscript() string {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.lastTranscript
}

// HandleStartStreamingRecording starts streaming transcription
func (as *AudioService) HandleStartStreamingRecording() error {
	as.logger.Info("Starting streaming recording...")

	if as.streamingEngine == nil {
		return fmt.Errorf("streaming engine not available")
	}

	// Start recording with streaming
	if err := as.recorder.StartRecording(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	as.isRecording = true

	// Update UI
	if as.ui != nil {
		as.ui.SetRecordingState(true)
		as.ui.SetTooltip("Streaming recording... Press hotkey to stop")
	}

	return nil
}

// HandleStreamingResult processes streaming transcription results
func (as *AudioService) HandleStreamingResult(text string, isFinal bool) {
	if as.io != nil {
		as.io.BroadcastTranscription(text, isFinal)
	}

	if isFinal {
		as.handleTranscriptionResult(text, nil)
	}
}

// HandleStartVADRecording starts VAD-based recording
func (as *AudioService) HandleStartVADRecording() error {
	as.logger.Info("Starting VAD recording...")

	// VAD implementation would go here
	// For now, fallback to standard recording
	return as.startStandardRecording()
}

// EnsureModelAvailable ensures whisper model is ready
func (as *AudioService) EnsureModelAvailable() error {
	if as.modelManager == nil {
		return fmt.Errorf("model manager not available")
	}

	// Try to get the model path, which will download if needed
	_, err := as.modelManager.GetModelPath()
	if err != nil {
		as.logger.Info("Model not found locally, checking download...")
		return fmt.Errorf("failed to ensure model available: %w", err)
	}

	return nil
}

// EnsureAudioRecorderAvailable ensures audio recorder is ready
func (as *AudioService) EnsureAudioRecorderAvailable() error {
	if as.audioRecorderNeedsReinit || as.recorder == nil {
		as.logger.Info("Reinitializing audio recorder...")

		recorder, err := factory.GetRecorder(as.config, as.logger)
		if err != nil {
			return fmt.Errorf("failed to reinitialize audio recorder: %w", err)
		}

		as.recorder = recorder
		as.audioRecorderNeedsReinit = false
	}

	return nil
}

// SwitchModel switches to a different model type
func (as *AudioService) SwitchModel(modelType string) error {
	as.logger.Info("Switching to model: %s", modelType)

	if as.modelManager == nil {
		return fmt.Errorf("model manager not available")
	}

	// Update config
	as.config.General.ModelType = modelType

	// Reinitialize engines with new model
	modelPath, err := as.modelManager.GetModelPath()
	if err != nil {
		return fmt.Errorf("failed to get model path: %w", err)
	}

	// Recreate whisper engine
	newEngine, err := whisper.NewWhisperEngine(as.config, modelPath)
	if err != nil {
		return fmt.Errorf("failed to create new whisper engine: %w", err)
	}

	as.whisperEngine = newEngine

	// Recreate streaming engine if needed
	if as.config.Audio.EnableStreaming {
		newStreamingEngine, err := whisper.NewStreamingWhisperEngine(as.config, modelPath)
		if err != nil {
			as.logger.Warning("Failed to reinitialize streaming engine: %v", err)
			as.streamingEngine = nil
		} else {
			as.streamingEngine = newStreamingEngine
		}
	}

	return nil
}

// Shutdown gracefully shuts down the audio service
func (as *AudioService) Shutdown() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.cancel()

	if as.isRecording && as.recorder != nil {
		if _, err := as.recorder.StopRecording(); err != nil {
			as.logger.Error("Error stopping recording during shutdown: %v", err)
		}
		as.isRecording = false
	}

	as.logger.Info("AudioService shutdown complete")
	return nil
}

// Private helper methods

// startStandardRecording starts standard recording mode
func (as *AudioService) startStandardRecording() error {
	// Set up audio level monitoring
	as.recorder.SetAudioLevelCallback(func(level float64) {
		if as.ui != nil {
			as.ui.UpdateRecordingUI(true, level)
		}
		as.logger.Debug("Audio level: %.2f", level)
	})

	// Start recording
	if err := as.recorder.StartRecording(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	as.isRecording = true

	// Update UI
	if as.ui != nil {
		as.ui.SetRecordingState(true)
		as.ui.SetTooltip("Recording... Press hotkey to stop")
		as.ui.ShowNotification(constants.NotifyRecordingStarted, "Speak now...")
	}

	return nil
}

// transcribeAsync performs async transcription
func (as *AudioService) transcribeAsync(audioFile string) {
	ctx, cancel := context.WithTimeout(as.ctx, 2*time.Minute)
	defer cancel()

	type result struct {
		transcript string
		err        error
	}

	resultChan := make(chan result, 1)

	go func() {
		transcript, err := as.whisperEngine.Transcribe(audioFile)
		select {
		case resultChan <- result{transcript: transcript, err: err}:
		case <-ctx.Done():
		}
	}()

	select {
	case res := <-resultChan:
		as.handleTranscriptionResult(res.transcript, res.err)
	case <-ctx.Done():
		as.handleTranscriptionCancellation(ctx.Err())
	}
}

// handleTranscriptionResult processes transcription results
func (as *AudioService) handleTranscriptionResult(transcript string, err error) {
	if err != nil {
		as.handleTranscriptionError(err)
		return
	}

	sanitized := utils.SanitizeTranscript(transcript)
	as.mu.Lock()
	as.lastTranscript = sanitized
	as.mu.Unlock()

	if sanitized == "" {
		as.handleEmptyTranscript()
		return
	}

	as.logger.Info("Transcription result: %s", sanitized)

	// Output text
	if as.io != nil {
		if err := as.io.OutputText(sanitized); err != nil {
			as.logger.Error("Failed to output text: %v", err)
			if as.ui != nil {
				as.ui.SetError("Output failed")
			}
			return
		}
	}

	// Update UI
	if as.ui != nil {
		as.ui.SetSuccess(constants.MsgTranscriptionComplete)
	}
}

// handleTranscriptionError handles transcription errors
func (as *AudioService) handleTranscriptionError(err error) {
	as.logger.Error("Transcription error: %v", err)
	if as.ui != nil {
		as.ui.SetError(constants.MsgTranscriptionFailed)
		as.ui.ShowNotification(constants.NotifyTranscriptionErr, err.Error())
	}
}

// handleEmptyTranscript handles empty transcription results
func (as *AudioService) handleEmptyTranscript() {
	as.logger.Info("Empty transcript received")
	if as.ui != nil {
		as.ui.SetError(constants.MsgNoSpeechDetected)
		as.ui.ShowNotification(constants.NotifyNoSpeech, constants.MsgTranscriptionEmpty)
	}
}

// handleRecordingError handles recording errors
func (as *AudioService) handleRecordingError(err error) {
	as.logger.Error("Recording error: %v", err)
	if as.ui != nil {
		as.ui.SetError("Recording error")
		as.ui.ShowNotification("Recording Error", err.Error())
	}
}

// handleTranscriptionCancellation handles transcription cancellation
func (as *AudioService) handleTranscriptionCancellation(err error) {
	as.logger.Warning("Transcription cancelled: %v", err)
	if as.ui != nil {
		as.ui.SetError("Transcription cancelled")
		as.ui.ShowNotification("Transcription Cancelled", "Operation timed out")
	}
}

// setUIError sets UI error state
func (as *AudioService) setUIError(message string) {
	if as.ui != nil {
		as.ui.SetError(message)
	}
}
