// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AshBuk/speak-to-ai/internal/ipc"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

func maybeRunCLI(args []string) (bool, int) {
	if !isCLICommandRequested(args) {
		return false, 0
	}

	var (
		socketPath string
		jsonOutput bool
		timeoutSec int
	)

	fs := flag.NewFlagSet("speak-to-ai", flag.ContinueOnError)
	var parseOutput strings.Builder
	fs.SetOutput(&parseOutput)

	setupCLIFlags(fs, &socketPath, &jsonOutput, &timeoutSec)
	fs.Usage = func() {
		printCLIUsage(os.Stderr, fs)
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return true, 0
		}
		if parseOutput.Len() > 0 {
			fmt.Fprint(os.Stderr, parseOutput.String())
		}
		fs.Usage()
		return true, 2
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fs.Usage()
		return true, 2
	}

	command := strings.ToLower(remaining[0])
	if !isCLIVerb(command) {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", remaining[0])
		fs.Usage()
		return true, 2
	}

	timeout := deriveTimeout(command, timeoutSec)

	if socketPath == "" {
		socketPath = utils.GetDefaultSocketPath()
	}

	resp, err := executeCLICommand(command, socketPath, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return true, 1
	}

	if jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode response: %v\n", err)
			return true, 1
		}
		return true, 0
	}

	printResponse(command, resp)
	return true, 0
}

func setupCLIFlags(fs *flag.FlagSet, socketPath *string, jsonOutput *bool, timeoutSec *int) {
	fs.StringVar(socketPath, "socket", "", "Path to IPC socket (defaults to user runtime path)")
	fs.BoolVar(jsonOutput, "json", false, "Print responses as JSON")
	fs.IntVar(timeoutSec, "timeout", 0, "Override timeout in seconds for the command")
}

func isCLICommandRequested(args []string) bool {
	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--":
			if i+1 < len(args) {
				return isCLIVerb(strings.ToLower(args[i+1]))
			}
			return false
		case !strings.HasPrefix(arg, "-") || arg == "-":
			return isCLIVerb(strings.ToLower(arg))
		default:
			name, inlineValue := splitFlagNameAndValue(arg)
			switch name {
			case "json":
				i++
			case "socket", "timeout":
				if inlineValue {
					i++
				} else {
					if i+1 >= len(args) {
						return true
					}
					i += 2
				}
			default:
				return false
			}
		}
	}
	return false
}

func splitFlagNameAndValue(flagArg string) (name string, hasValue bool) {
	name = strings.TrimLeft(flagArg, "-")
	if idx := strings.Index(name, "="); idx != -1 {
		return name[:idx], true
	}
	return name, false
}

func isCLIVerb(command string) bool {
	switch command {
	case "start", "stop", "status", "transcript", "last-transcript":
		return true
	default:
		return false
	}
}

func executeCLICommand(command, socketPath string, timeout time.Duration) (ipc.Response, error) {
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
	}

	return resp, err
}

const (
	defaultStatusTimeout = 5 * time.Second
	defaultStopTimeout   = 60 * time.Second
)

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
