// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"sync"
	"testing"

	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/internal/testutils"
)

type fakeOutputter struct{}

func (fakeOutputter) CopyToClipboard(string) error            { return nil }
func (fakeOutputter) TypeToActiveWindow(string) error         { return nil }
func (fakeOutputter) GetToolNames() (clipboard, typer string) { return "fake", "fake" }

// A mode switch swaps the outputter from the tray goroutine while a
// transcription goroutine writes through it.
func TestIOService_SwapOutputterDuringOutput(t *testing.T) {
	cfg := &config.Config{}
	cfg.Output.DefaultMode = config.OutputModeClipboard
	ios := NewIOService(testutils.NewMockLogger(), testutils.StaticConfig(cfg), fakeOutputter{}, nil)

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() { ios.setOutputter(fakeOutputter{}) })
		wg.Go(func() {
			if err := ios.OutputText("hello"); err != nil {
				t.Errorf("OutputText failed: %v", err)
			}
		})
		wg.Go(func() { _, _ = ios.GetOutputToolNames() })
	}
	wg.Wait()
}
