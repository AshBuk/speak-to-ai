// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"runtime"
)

// Version information - injected at build time via ldflags:
//
//	go build -ldflags "-X main.version=1.5.0"
//
// Defaults to "dev" for development builds.
var version = "dev"

// printVersion outputs version information and exits.
// Format: speak-to-ai version X.Y.Z (go1.XX.X linux/amd64)
func printVersion() {
	fmt.Printf("speak-to-ai version %s (%s %s/%s)\n",
		version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	os.Exit(0)
}

// isVersionFlag checks if the argument is a version flag.
// Supports standard POSIX/GNU formats: -v (short) and --version (long).
func isVersionFlag(arg string) bool {
	return arg == "-v" || arg == "--version"
}
