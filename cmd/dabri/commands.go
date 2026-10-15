// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AshBuk/dabri/v2/internal/ipc"
	"github.com/AshBuk/dabri/v2/internal/utils"
)

// addCLICommands registers all IPC-based subcommands on the root command.
// Each command sends a request to the running daemon via Unix socket.
func addCLICommands(root *cobra.Command) {
	root.AddCommand(
		newIPCCommand("start", "Start recording", "start-recording"),
		newIPCCommand("stop", "Stop recording and return transcript", "stop-recording"),
		newIPCCommand("toggle", "Toggle recording (start if stopped, stop if recording)", "toggle-recording"),
		newIPCCommand("status", "Show current state and configuration", "status"),
		newIPCCommand("transcript", "Show the last transcript", "last-transcript"),
	)
}

// newIPCCommand creates a cobra.Command that sends an IPC request to the daemon.
// Factory Pattern — shared flags and execution logic for all IPC commands.
func newIPCCommand(use, short, ipcCommand string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIPCCommand(cmd, ipcCommand)
		},
		SilenceUsage: true,
	}
	addCLIFlags(cmd)
	return cmd
}

func addCLIFlags(cmd *cobra.Command) {
	cmd.Flags().String("socket", "", "Path to IPC socket (defaults to user runtime path)")
	cmd.Flags().Bool("json", false, "Print responses as JSON")
	cmd.Flags().Int("timeout", 0, "Override timeout in seconds for the command")
}

// runIPCCommand executes a single IPC request to the daemon and formats the response.
// Execution flow: read flags → send IPC → format output (text or JSON)
func runIPCCommand(cmd *cobra.Command, ipcCommand string) error {
	socketPath, _ := cmd.Flags().GetString("socket")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	timeoutSec, _ := cmd.Flags().GetInt("timeout")

	if socketPath == "" {
		socketPath = utils.GetDefaultSocketPath()
	}

	name := cmd.Name()
	timeout := deriveTimeout(name, timeoutSec)

	req := ipc.Request{Command: ipcCommand}
	resp, err := ipc.SendRequest(socketPath, req, timeout)
	if err != nil {
		return err
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(resp)
	}
	printResponse(name, resp)
	return nil
}

const (
	defaultStatusTimeout = 5 * time.Second
	defaultStopTimeout   = 60 * time.Second
	defaultModelTimeout  = 5 * time.Minute
)

func deriveTimeout(command string, override int) time.Duration {
	if override > 0 {
		return time.Duration(override) * time.Second
	}

	switch command {
	case "stop", "toggle":
		return defaultStopTimeout
	case "set", "delete":
		return defaultModelTimeout
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
		if getBoolOr(data, "recording", false) {
			fmt.Println("Recording started.")
		} else {
			printTranscript(data, "Recording stopped.")
		}
	case "stop":
		if warning := getString(data, "warning"); warning != "" {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
		}
		printTranscript(data, "Recording stopped (no transcript available).")
	case "status":
		printStatusResponse(data)
	case "transcript":
		printTranscript(data, "No transcript available.")
	case "set":
		printField(data, "model", "Model switched to: %s\n")
	case "delete":
		printField(data, "model", "Model deleted: %s\n")
	default:
		if resp.Message != "" {
			fmt.Println(resp.Message)
		}
	}
}

func printStatusResponse(data map[string]any) {
	recording := getBoolOr(data, "recording", false)
	fmt.Printf("Recording: %t\n", recording)

	// Config section
	if cfg := getMap(data, "config"); cfg != nil {
		fmt.Println("\nConfig:")
		printField(cfg, "model", "  Model: %s\n")
		printField(cfg, "language", "  Language: %s\n")
		printField(cfg, "output_mode", "  Output mode: %s\n")
		printField(cfg, "audio_method", "  Audio method: %s\n")
	}
	// Hotkeys section
	if hotkeys := getMap(data, "hotkeys"); hotkeys != nil {
		fmt.Println("\nHotkeys:")
		printField(hotkeys, "start_stop", "  Start/Stop: %s\n")
		printField(hotkeys, "show_config", "  Show config: %s\n")
	}
	// WebSocket section
	if ws := getMap(data, "websocket"); ws != nil {
		enabled := getBoolOr(ws, "enabled", false)
		if enabled {
			host := getString(ws, "host")
			port := getIntOr(ws, "port", 0)
			fmt.Printf("\nWebSocket: enabled (%s:%d)\n", host, port)
		} else {
			fmt.Println("\nWebSocket: disabled")
		}
	}
}

// printTranscript prints the transcript, or fallback when there is none.
func printTranscript(data map[string]any, fallback string) {
	if !printField(data, "transcript", "%s\n") {
		fmt.Println(fallback)
	}
}

// printField prints the string at key with format when it is non-empty.
func printField(data map[string]any, key, format string) bool {
	value := getString(data, key)
	if value != "" {
		fmt.Printf(format, value)
	}
	return value != ""
}

// --- Response parsing helpers ---

func mapFromResponse(data any) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	if m, ok := data.(map[string]any); ok && m != nil {
		return m
	}
	return map[string]any{}
}

func getString(data map[string]any, key string) string {
	str, _ := data[key].(string)
	return str
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
