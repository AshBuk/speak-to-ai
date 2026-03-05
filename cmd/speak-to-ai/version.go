// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"runtime"

	appversion "github.com/AshBuk/speak-to-ai/internal/version"
)

// printVersion outputs version information and exits.
// Format: speak-to-ai version X.Y.Z (go1.XX.X linux/amd64)
func printVersion() {
	fmt.Printf("speak-to-ai version %s (%s %s/%s)\n",
		appversion.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	os.Exit(0)
}

// isVersionFlag checks if the argument is a version flag.
// Supports standard POSIX/GNU formats: -v (short) and --version (long).
func isVersionFlag(arg string) bool {
	return arg == "-v" || arg == "--version"
}
