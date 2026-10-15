// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"io"
	"os"
	"testing"

	"github.com/AshBuk/dabri/v2/internal/ipc"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	os.Stdout = orig
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestPrintResponse(t *testing.T) {
	tests := []struct {
		command string
		data    map[string]any
		want    string
	}{
		{"toggle", map[string]any{"recording": true}, "Recording started.\n"},
		{"toggle", map[string]any{"transcript": "hello"}, "hello\n"},
		{"toggle", nil, "Recording stopped.\n"},
		{"stop", map[string]any{"transcript": ""}, "Recording stopped (no transcript available).\n"},
		{"transcript", nil, "No transcript available.\n"},
		{"set", map[string]any{"model": "base"}, "Model switched to: base\n"},
		{"delete", nil, ""},
		{"status", map[string]any{
			"recording": false,
			"config":    map[string]any{"model": "small", "language": ""},
			"websocket": map[string]any{"enabled": true, "host": "localhost", "port": float64(8080)},
		}, "Recording: false\n\nConfig:\n  Model: small\n\nWebSocket: enabled (localhost:8080)\n"},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := captureStdout(t, func() { printResponse(tt.command, ipc.Response{Data: tt.data}) })
			if got != tt.want {
				t.Errorf("printResponse(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}
