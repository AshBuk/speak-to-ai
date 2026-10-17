// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"fmt"
	"sync"
	"testing"

	"github.com/AshBuk/dabri/v2/audio/interfaces"
	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/internal/testutils"
)

type fakeRecorder struct {
	cleanups int
	stopErr  error
}

func (r *fakeRecorder) StartRecording() error          { return nil }
func (r *fakeRecorder) StopRecording() (string, error) { return "", r.stopErr }
func (r *fakeRecorder) GetOutputFile() string          { return "" }
func (r *fakeRecorder) CleanupFile() error             { r.cleanups++; return nil }

func TestAudioService_ClearSessionKeepsActiveRecording(t *testing.T) {
	rec := &fakeRecorder{}
	as := NewAudioService(testutils.NewMockLogger(), testutils.StaticConfig(&config.Config{}), rec, nil, nil, nil)
	as.isRecording = true

	as.ClearSession()
	if rec.cleanups != 0 {
		t.Error("ClearSession removed the file of an active recording")
	}
	if !as.audioRecorderNeedsReinit {
		t.Error("recorder should be reinitialized on the next start")
	}

	as.isRecording = false
	as.ClearSession()
	if rec.cleanups != 1 {
		t.Errorf("cleanups = %d, want 1", rec.cleanups)
	}
}

func TestAudioService_StopDiscardsTooShortRecording(t *testing.T) {
	rec := &fakeRecorder{stopErr: fmt.Errorf("%w (payload=0 < 1600)", interfaces.ErrRecordingTooShort)}
	cfg := &config.Config{}
	cfg.Audio.RecordingMethod = "ffmpeg"
	as := NewAudioService(testutils.NewMockLogger(), testutils.StaticConfig(cfg), rec, nil, nil, nil)
	as.isRecording = true

	if err := as.HandleStopRecording(); err != nil {
		t.Fatalf("HandleStopRecording() = %v, want nil", err)
	}
	if as.isRecording || rec.cleanups != 1 || as.audioRecorderNeedsReinit {
		t.Errorf("isRecording=%t cleanups=%d needsReinit=%t, want false 1 false", as.isRecording, rec.cleanups, as.audioRecorderNeedsReinit)
	}
}

func TestAudioService_ClearSessionConcurrentWithReads(t *testing.T) {
	as := NewAudioService(testutils.NewMockLogger(), testutils.StaticConfig(&config.Config{}), &fakeRecorder{}, nil, nil, nil)
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(as.ClearSession)
		wg.Go(func() { _ = as.GetLastTranscript() })
	}
	wg.Wait()
}
