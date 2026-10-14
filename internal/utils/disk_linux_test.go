//go:build linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package utils

import "testing"

func TestCheckDiskSpace(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid path",
			path:        "/tmp",
			expectError: false,
		},
		{
			name:        "invalid path",
			path:        "/non/existing/path",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDiskSpace(tt.path)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
