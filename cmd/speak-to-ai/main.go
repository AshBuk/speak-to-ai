// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"os"
)

func main() {
	args := os.Args[1:]

	if handled, exitCode := maybeRunCLI(args); handled {
		os.Exit(exitCode)
	}

	if exitCode := runDaemon(args); exitCode != 0 {
		os.Exit(exitCode)
	}
}
