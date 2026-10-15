// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"sync"
	"testing"

	"github.com/AshBuk/dabri/v2/audio/interfaces"
	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/internal/testutils"
)

type fakeRecorder struct{ cleanups int }

func (r *fakeRecorder) StartRecording() error                               { return nil }
func (r *fakeRecorder) StopRecording() (string, error)                      { return "", nil }
func (r *fakeRecorder) GetOutputFile() string                               { return "" }
func (r *fakeRecorder) CleanupFile() error                                  { r.cleanups++; return nil }
func (r *fakeRecorder) SetAudioLevelCallback(interfaces.AudioLevelCallback) {}
func (r *fakeRecorder) GetAudioLevel() float64                              { return 0 }

func TestAudioService_ClearSessionKeepsActiveRecording(t *testing.T) {
	rec := &fakeRecorder{}
	as := NewAudioService(testutils.NewMockLogger(), &config.Config{}, rec, nil, nil, nil)
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

func TestAudioService_ClearSessionConcurrentWithReads(t *testing.T) {
	as := NewAudioService(testutils.NewMockLogger(), &config.Config{}, &fakeRecorder{}, nil, nil, nil)
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(as.ClearSession)
		wg.Go(func() { _ = as.GetLastTranscript() })
	}
	wg.Wait()
}
