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

// CLI configuration options
type cliOptions struct {
	command    string        // CLI command verb (start/stop/status/transcript)
	socketPath string        // Path to IPC Unix socket
	jsonOutput bool          // Output response as JSON
	timeout    time.Duration // IPC request timeout
}

// Parse CLI command-line flags and command
// Returns parsed options or error if invalid
func parseCLIOptions(args []string) (*cliOptions, error) {
	var (
		socketPath string
		jsonOutput bool
		timeoutSec int
	)

	fs := flag.NewFlagSet("speak-to-ai", flag.ContinueOnError) // pls don't panic on parse error
	var parseOutput strings.Builder
	fs.SetOutput(&parseOutput)

	setupCLIFlags(fs, &socketPath, &jsonOutput, &timeoutSec)
	fs.Usage = func() {
		printCLIUsage(os.Stderr, fs)
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, flag.ErrHelp
		}
		if parseOutput.Len() > 0 {
			fmt.Fprint(os.Stderr, parseOutput.String())
		}
		fs.Usage()
		return nil, err
	}
	remaining := fs.Args()
	if len(remaining) == 0 {
		fs.Usage()
		return nil, fmt.Errorf("no command specified")
	}

	command := strings.ToLower(remaining[0])
	if !isCLIVerb(command) {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", remaining[0])
		fs.Usage()
		return nil, fmt.Errorf("unknown command: %s", remaining[0])
	}
	timeout := deriveTimeout(command, timeoutSec)
	if socketPath == "" {
		socketPath = utils.GetDefaultSocketPath()
	}

	return &cliOptions{
		command:    command,
		socketPath: socketPath,
		jsonOutput: jsonOutput,
		timeout:    timeout,
	}, nil
}

// CLI orchestrator - coordinates the entire CLI workflow
// Facade Pattern - simplified interface to CLI subsystem
// Execution flow:
//  1. Detect CLI command (start/stop/status/transcript)
//  2. Parse flags (--socket, --json, --timeout)
//  3. Send IPC request → daemon via Unix socket
//  4. Format response (text or JSON)
//
// Returns: (handled=true if CLI command detected, exitCode)
func maybeRunCLI(args []string) (bool, int) {
	if !isCLICommandRequested(args) {
		return false, 0
	}

	opts, err := parseCLIOptions(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return true, 0
		}
		return true, 2
	}

	resp, err := executeCLICommand(opts.command, opts.socketPath, opts.timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return true, 1
	}
	if opts.jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode response: %v\n", err)
			return true, 1
		}
		return true, 0
	}
	printResponse(opts.command, resp)
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
	case "start", "stop", "toggle", "status", "transcript":
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
	case "toggle":
		req = ipc.Request{Command: "toggle-recording"}
		resp, err = ipc.SendRequest(socketPath, req, timeout)
	case "transcript":
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
	case "stop", "toggle":
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
	case "toggle":
		recording := getBoolOr(data, "recording", false)
		if recording {
			fmt.Println("Recording started.")
		} else {
			if transcript, ok := getString(data, "transcript"); ok && transcript != "" {
				fmt.Println(transcript)
			} else {
				fmt.Println("Recording stopped.")
			}
		}
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
		printStatusResponse(data)
	case "transcript":
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

func getMap(data map[string]any, key string) map[string]any {
	if value, ok := data[key]; ok {
		if m, ok := value.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func getIntOr(data map[string]any, key string, fallback int) int {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return fallback
}

func printStatusResponse(data map[string]any) {
	recording := getBoolOr(data, "recording", false)
	fmt.Printf("Recording: %t\n", recording)

	// Config section
	if cfg := getMap(data, "config"); cfg != nil {
		fmt.Println("\nConfig:")
		if model, ok := getString(cfg, "model"); ok && model != "" {
			fmt.Printf("  Model: %s\n", model)
		}
		if lang, ok := getString(cfg, "language"); ok && lang != "" {
			fmt.Printf("  Language: %s\n", lang)
		}
		if mode, ok := getString(cfg, "output_mode"); ok && mode != "" {
			fmt.Printf("  Output mode: %s\n", mode)
		}
		if method, ok := getString(cfg, "audio_method"); ok && method != "" {
			fmt.Printf("  Audio method: %s\n", method)
		}
	}
	// Hotkeys section
	if hotkeys := getMap(data, "hotkeys"); hotkeys != nil {
		fmt.Println("\nHotkeys:")
		if startStop, ok := getString(hotkeys, "start_stop"); ok && startStop != "" {
			fmt.Printf("  Start/Stop: %s\n", startStop)
		}
		if showConfig, ok := getString(hotkeys, "show_config"); ok && showConfig != "" {
			fmt.Printf("  Show config: %s\n", showConfig)
		}
	}
	// WebSocket section
	if ws := getMap(data, "websocket"); ws != nil {
		enabled := getBoolOr(ws, "enabled", false)
		if enabled {
			host, _ := getString(ws, "host")
			port := getIntOr(ws, "port", 0)
			fmt.Printf("\nWebSocket: enabled (%s:%d)\n", host, port)
		} else {
			fmt.Println("\nWebSocket: disabled")
		}
	}
}
