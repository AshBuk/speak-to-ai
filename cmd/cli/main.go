// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AshBuk/speak-to-ai/internal/ipc"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

const (
	defaultStatusTimeout = 5 * time.Second
	defaultStopTimeout   = 60 * time.Second
)

func main() {
	var (
		socketPath string
		jsonOutput bool
		timeoutSec int
	)

	flag.StringVar(&socketPath, "socket", "", "Path to IPC socket (defaults to user runtime path)")
	flag.BoolVar(&jsonOutput, "json", false, "Print responses as JSON")
	flag.IntVar(&timeoutSec, "timeout", 0, "Override timeout in seconds for the command")
	flag.Usage = func() {
		usageWriter := flag.CommandLine.Output()
		writeUsage := func(format string, args ...any) {
			if _, err := fmt.Fprintf(usageWriter, format, args...); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Failed to write usage information: %v\n", err)
			}
		}

		writeUsage("Usage: %s [flags] <command>\n", os.Args[0])
		writeUsage("\n")
		writeUsage("Commands:\n")
		writeUsage("  start        Start recording\n")
		writeUsage("  stop         Stop recording and return transcript\n")
		writeUsage("  status       Show current recording status\n")
		writeUsage("  transcript   Show the last transcript\n")
		writeUsage("\n")
		writeUsage("Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	if socketPath == "" {
		socketPath = utils.GetDefaultSocketPath()
	}

	command := strings.ToLower(args[0])
	timeout := deriveTimeout(command, timeoutSec)

	var (
		req  ipc.Request
		resp ipc.Response
		err  error
	)

	switch command {
	case "start":
		req = ipc.Request{Command: "start-recording"}
		resp, err = ipc.SendRequest(socketPath, req, timeout)
	case "stop":
		req = ipc.Request{Command: "stop-recording"}
		resp, err = ipc.SendRequest(socketPath, req, timeout)
	case "status":
		req = ipc.Request{Command: "status"}
		resp, err = ipc.SendRequest(socketPath, req, timeout)
	case "transcript", "last-transcript":
		req = ipc.Request{Command: "last-transcript"}
		resp, err = ipc.SendRequest(socketPath, req, timeout)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode response: %v\n", err)
			os.Exit(1)
		}
		return
	}

	printResponse(command, resp)
}

func deriveTimeout(command string, override int) time.Duration {
	if override > 0 {
		return time.Duration(override) * time.Second
	}

	switch command {
	case "stop":
		return defaultStopTimeout
	default:
		return defaultStatusTimeout
	}
}

func printResponse(command string, resp ipc.Response) {
	data := mapFromResponse(resp.Data)

	switch command {
	case "start":
		fmt.Println("Recording started.")
	case "stop":
		if warning, ok := getString(data, "warning"); ok && warning != "" {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
		}
		if transcript, ok := getString(data, "transcript"); ok && transcript != "" {
			fmt.Println(transcript)
		} else {
			fmt.Println("Recording stopped (no transcript available).")
		}
	case "status":
		recording := getBoolOr(data, "recording", false)
		fmt.Printf("Recording: %t\n", recording)
		if transcript, ok := getString(data, "last_transcript"); ok && transcript != "" {
			fmt.Printf("Last transcript: %s\n", transcript)
		}
	case "transcript", "last-transcript":
		if transcript, ok := getString(data, "transcript"); ok && transcript != "" {
			fmt.Println(transcript)
		} else {
			fmt.Println("No transcript available.")
		}
	default:
		// Fallback: print message field if present.
		if resp.Message != "" {
			fmt.Println(resp.Message)
		}
	}
}

func mapFromResponse(data any) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	if m, ok := data.(map[string]any); ok && m != nil {
		return m
	}
	return map[string]any{}
}

func getString(data map[string]any, key string) (string, bool) {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

func getBoolOr(data map[string]any, key string, fallback bool) bool {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case bool:
			return v
		case string:
			return strings.EqualFold(v, "true")
		}
	}
	return fallback
}
