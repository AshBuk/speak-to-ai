// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package notify

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/AshBuk/dabri/v2/config"
	"github.com/AshBuk/dabri/v2/internal/constants"
	"github.com/AshBuk/dabri/v2/internal/testutils"
)

// Helper function to create test config
func createTestConfig() *config.Config {
	cfg := &config.Config{}
	config.SetDefaultConfig(cfg)
	return cfg
}

// installFakeNotifySend puts a notify-send stub alone on PATH, so tests never
// reach the real desktop. Each run appends its argv, one per line.
func installFakeNotifySend(t *testing.T, exitCode int) (argsFile string) {
	t.Helper()
	dir := t.TempDir()
	argsFile = filepath.Join(dir, "args")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" >> %q\nexit %d\n", argsFile, exitCode)
	if err := os.WriteFile(filepath.Join(dir, "notify-send"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	return argsFile
}

// recordedArgs returns every argv line recorded by the stub, nil if it never ran.
func recordedArgs(t *testing.T, argsFile string) []string {
	t.Helper()
	data, err := os.ReadFile(argsFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func TestNewNotificationManager(t *testing.T) {
	appName := "TestApp"
	cfg := createTestConfig()
	nm := NewNotificationManager(appName, testutils.StaticConfig(cfg))
	if nm == nil {
		t.Fatal("NewNotificationManager returned nil")
	}

	if nm.appName != appName {
		t.Errorf("Expected appName %q, got %q", appName, nm.appName)
	}
}

func TestShowNotification_PassesTextAfterOptionTerminator(t *testing.T) {
	args := installFakeNotifySend(t, 0)
	nm := NewNotificationManager("TestApp", testutils.StaticConfig(createTestConfig()))

	// A summary that looks like a flag must not be parsed as one
	if err := nm.ShowNotification("--urgency=critical", "body"); err != nil {
		t.Fatal(err)
	}
	want := []string{"--app-name", "TestApp", "--icon", "dialog-information-symbolic", "--", "--urgency=critical", "body"}
	if got := recordedArgs(t, args); !slices.Equal(got, want) {
		t.Errorf("args:\n got  %q\n want %q", got, want)
	}
}

func TestWorkflowNotifications_FollowSetting(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint("enabled=", enabled), func(t *testing.T) {
			args := installFakeNotifySend(t, 0)
			cfg := createTestConfig()
			cfg.Notifications.EnableWorkflowNotifications = enabled
			nm := NewNotificationManager("TestApp", testutils.StaticConfig(cfg))

			for _, notify := range []func() error{nm.NotifyStartRecording, nm.NotifyStopRecording, nm.NotifyTranscriptionComplete} {
				if err := notify(); err != nil {
					t.Fatal(err)
				}
			}
			titles := 0
			for _, a := range recordedArgs(t, args) {
				if a == "--" {
					titles++
				}
			}
			if want := map[bool]int{false: 0, true: 3}[enabled]; titles != want {
				t.Errorf("sent %d notifications, want %d", titles, want)
			}
		})
	}
}

func TestNotifyTranscriptionComplete_BodyFollowsOutputMode(t *testing.T) {
	for mode, body := range map[string]string{
		config.OutputModeActiveWindow: constants.NotifyTranscriptionTypedMsg,
		config.OutputModeClipboard:    constants.NotifyTranscriptionMsg,
	} {
		t.Run(mode, func(t *testing.T) {
			args := installFakeNotifySend(t, 0)
			cfg := createTestConfig()
			cfg.Notifications.EnableWorkflowNotifications = true
			cfg.Output.DefaultMode = mode

			if err := NewNotificationManager("TestApp", testutils.StaticConfig(cfg)).NotifyTranscriptionComplete(); err != nil {
				t.Fatal(err)
			}
			if got := recordedArgs(t, args); len(got) == 0 || got[len(got)-1] != body {
				t.Errorf("body: got %q, want %q", got, body)
			}
		})
	}
}

func TestSendNotification_Errors(t *testing.T) {
	t.Run("command not allowed", func(t *testing.T) {
		args := installFakeNotifySend(t, 0)
		cfg := createTestConfig()
		cfg.Security.AllowedCommands = []string{"xsel"}

		err := NewNotificationManager("TestApp", testutils.StaticConfig(cfg)).NotifyError("boom")
		if err == nil || err.Error() != "notify-send command not allowed" {
			t.Errorf("got %v", err)
		}
		if got := recordedArgs(t, args); got != nil {
			t.Errorf("disallowed command ran with %q", got)
		}
	})
	t.Run("notify-send fails", func(t *testing.T) {
		installFakeNotifySend(t, 1)
		err := NewNotificationManager("TestApp", testutils.StaticConfig(createTestConfig())).NotifyError("boom")
		if err == nil || !strings.HasPrefix(err.Error(), "failed to send notification:") {
			t.Errorf("got %v", err)
		}
	})
}

func TestNotificationManager_IsAvailable(t *testing.T) {
	nm := NewNotificationManager("TestApp", testutils.StaticConfig(createTestConfig()))

	installFakeNotifySend(t, 0)
	if !nm.IsAvailable() {
		t.Error("notify-send on PATH must be reported available")
	}
	t.Setenv("PATH", t.TempDir())
	if nm.IsAvailable() {
		t.Error("notify-send missing from PATH must be reported unavailable")
	}
}
