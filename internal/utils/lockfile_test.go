// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestTryLockIgnoresStalePIDFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultLockFileName)
	// A crashed instance leaves its PID behind; inside a Flatpak sandbox that PID is often our own
	if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		t.Fatal(err)
	}

	first := NewLockFile(path)
	if err := first.TryLock(); err != nil {
		t.Fatalf("stale lock file blocked startup: %v", err)
	}
	defer func() { _ = first.Unlock() }()

	if err := NewLockFile(path).TryLock(); err == nil {
		t.Fatal("second instance acquired a held lock")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != strconv.Itoa(os.Getpid()) {
		t.Fatalf("lock file PID = %q, held lock's PID was overwritten", data)
	}
}
