// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/internal/testutils"
	"github.com/AshBuk/dabri/v2/tests/mocks"
)

type actionFixture struct {
	log     *mocks.CallLog
	audio   *mocks.MockAudioService
	cfg     *mocks.MockConfigService
	hotkeys *mocks.MockHotkeyService
	c       *ServiceContainer
	w       *FactoryWirer
}

func newActionFixture() *actionFixture {
	log := &mocks.CallLog{}
	f := &actionFixture{
		log:     log,
		audio:   &mocks.MockAudioService{Log: log},
		cfg:     &mocks.MockConfigService{Log: log, Cfg: &config.Config{}},
		hotkeys: &mocks.MockHotkeyService{Log: log},
	}
	f.cfg.Cfg.General.WhisperModel = "base"
	f.cfg.Cfg.Hotkeys.StartRecording = "alt+r"
	f.c = &ServiceContainer{
		Audio:   f.audio,
		UI:      &mocks.MockUIService{Log: log},
		IO:      &mocks.MockIOService{Log: log},
		Config:  f.cfg,
		Hotkeys: f.hotkeys,
	}
	f.w = NewFactoryWirer(testutils.NewMockLogger())
	return f
}

func (f *actionFixture) expect(t *testing.T, want ...string) {
	t.Helper()
	if !slices.Equal(f.log.Calls, want) {
		t.Errorf("calls:\n got  %q\n want %q", f.log.Calls, want)
	}
}

func TestToggleRecording(t *testing.T) {
	f := newActionFixture()
	_ = f.w.makeToggleCallback(f.c)()
	f.audio.Recording = true
	_ = f.w.makeToggleCallback(f.c)()
	f.expect(t, "Audio.Start", "Audio.Stop")
}

func TestResetToDefaults(t *testing.T) {
	f := newActionFixture()
	if err := f.w.makeResetToDefaultsCallback(f.c)(); err != nil {
		t.Fatal(err)
	}
	f.expect(t, "Config.Reset", "Hotkeys.Reload:alt+r")

	f = newActionFixture()
	f.cfg.ResetErr = errors.New("boom")
	if err := f.w.makeResetToDefaultsCallback(f.c)(); err == nil {
		t.Fatal("expected error")
	}
	f.expect(t, "Config.Reset")
}

func TestSwitchModel(t *testing.T) {
	f := newActionFixture()
	if err := f.w.makeModelSelectionCallback(f.c)(context.Background(), "small"); err != nil {
		t.Fatal(err)
	}
	f.expect(t, "Audio.SwitchModel:small", "Config.Model:small", "UI.UpdateSettings", "UI.Notify:Model Switch")
}

func TestSwitchModel_SwitchFails(t *testing.T) {
	f := newActionFixture()
	f.audio.SwitchModelErr = map[string]error{"small": errors.New("boom")}
	if err := f.w.makeModelSelectionCallback(f.c)(context.Background(), "small"); err == nil {
		t.Fatal("expected error")
	}
	f.expect(t, "Audio.SwitchModel:small", "UI.Notify:Model Switch Failed")
}

func TestSwitchModel_CancelledIsSilent(t *testing.T) {
	f := newActionFixture()
	f.audio.SwitchModelErr = map[string]error{"small": context.Canceled}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := f.w.makeModelSelectionCallback(f.c)(ctx, "small"); err == nil {
		t.Fatal("expected error")
	}
	f.expect(t, "Audio.SwitchModel:small")
}

func TestSwitchModel_PersistFailsRollsBack(t *testing.T) {
	f := newActionFixture()
	f.cfg.UpdateModelErr = errors.New("disk full")
	if err := f.w.makeModelSelectionCallback(f.c)(context.Background(), "small"); err == nil {
		t.Fatal("expected error")
	}
	f.expect(t, "Audio.SwitchModel:small", "Config.Model:small", "Audio.SwitchModel:base")
}

func TestSettingsActions(t *testing.T) {
	f := newActionFixture()
	_ = f.w.makeLanguageCallback(f.c)("ru")
	_ = f.w.makeOutputModeCallback(f.c)("clipboard")
	_ = f.w.makeRecorderSelectionCallback(f.c)("ffmpeg")
	_ = f.w.makeToggleNotificationsCallback(f.c)()
	f.expect(t,
		"Config.Language:ru", "UI.UpdateSettings",
		"Config.OutputMode:clipboard", "IO.SetOutputMethod:clipboard", "UI.UpdateSettings",
		"Config.RecordingMethod:ffmpeg", "Audio.ClearSession", "UI.UpdateSettings",
		"Config.ToggleNotifications", "UI.Notify:Workflow Notifications",
	)
}

func TestRebindHotkey(t *testing.T) {
	f := newActionFixture()
	if err := f.w.makeHotkeyRebindCallback(f.c)("start_recording"); err != nil {
		t.Fatal(err)
	}
	f.expect(t, "UI.Notify:Rebind Hotkey", "Hotkeys.Capture", "UI.Notify:Rebind Hotkey")

	f = newActionFixture()
	f.hotkeys.Captured = "ctrl+space"
	if err := f.w.makeHotkeyRebindCallback(f.c)("start_recording"); err != nil {
		t.Fatal(err)
	}
	f.expect(t, "UI.Notify:Rebind Hotkey", "Hotkeys.Capture",
		"Config.Hotkey:start_recording=ctrl+space", "Hotkeys.Reload:alt+r", "UI.Notify:Hotkey Updated")
}
