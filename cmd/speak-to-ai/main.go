// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"os"
)

// Application entry point
// Strategy Pattern (Command Router)- routes execution between two modes:
//
//	→ CLI Mode (client):  start/stop/toggle/status/transcript → IPC
//	→ Daemon Mode (server): app.NewApp() → Initialize() → RunAndWait()
//
// Early Return - try CLI first, fall to daemon orchestrator
func main() {
	args := os.Args[1:]

	// Handle --version / -v before anything else
	for _, arg := range args {
		if isVersionFlag(arg) {
			printVersion()
		}
	}

	if handled, exitCode := maybeRunCLI(args); handled {
		os.Exit(exitCode)
	}
	if exitCode := runDaemon(args); exitCode != 0 {
		os.Exit(exitCode)
	}
}
