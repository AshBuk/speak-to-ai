// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func printCLIUsage(w io.Writer, fs *flag.FlagSet) {
	name := filepath.Base(os.Args[0])
	if _, err := fmt.Fprintf(w, "Usage: %s [flags] <command>\n\n", name); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "Commands:"); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "  start        Start recording"); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "  stop         Stop recording and return transcript"); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "  status       Show current recording status"); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "  transcript   Show the last transcript"); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "Flags:"); err != nil {
		reportUsageError(err)
		return
	}
	originalOutput := fs.Output()
	fs.SetOutput(w)
	fs.PrintDefaults()
	fs.SetOutput(originalOutput)
}

func reportUsageError(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Failed to write usage information: %v\n", err)
}

func printCombinedUsage(w io.Writer, daemonFS *flag.FlagSet) {
	name := filepath.Base(os.Args[0])
	if _, err := fmt.Fprintf(w, "Usage:\n  %s [daemon flags]\n  %s [CLI flags] <start|stop|status|transcript>\n  %s --version\n\n", name, name, name); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "Daemon Flags:"); err != nil {
		reportUsageError(err)
		return
	}
	originalOutput := daemonFS.Output()
	daemonFS.SetOutput(w)
	daemonFS.PrintDefaults()
	daemonFS.SetOutput(originalOutput)

	if _, err := fmt.Fprintln(w); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "CLI Commands:"); err != nil {
		reportUsageError(err)
		return
	}
	commandRows := []string{
		"  start        Start recording",
		"  stop         Stop recording and return transcript",
		"  status       Show current recording status",
		"  transcript   Show the last transcript",
	}
	for _, row := range commandRows {
		if _, err := fmt.Fprintln(w, row); err != nil {
			reportUsageError(err)
			return
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "CLI Flags:"); err != nil {
		reportUsageError(err)
		return
	}

	var (
		cliSocket  string
		cliJSON    bool
		cliTimeout int
	)
	cliFS := flag.NewFlagSet("speak-to-ai-cli", flag.ContinueOnError)
	cliFS.SetOutput(w)
	setupCLIFlags(cliFS, &cliSocket, &cliJSON, &cliTimeout)
	cliFS.PrintDefaults()
}
