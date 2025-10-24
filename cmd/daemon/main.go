// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AshBuk/speak-to-ai/internal/app"
	"github.com/AshBuk/speak-to-ai/internal/ipc"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/utils"
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

func runDaemon(args []string) int {
	opts, err := parseDaemonOptions(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	// Create logger early for consistent logging
	logLevel := logger.InfoLevel
	if opts.debug {
		logLevel = logger.DebugLevel
	}
	appLogger := logger.NewDefaultLogger(logLevel)

	configPath := adjustPathsForAppImage(appLogger, opts.configFile)
	configPath = adjustPathsForFlatpak(appLogger, configPath)

	// Single-instance protection
	lockFile := utils.NewLockFile(utils.GetDefaultLockPath())
	if isRunning, pid, err := lockFile.CheckExistingInstance(); err != nil {
		appLogger.Warning("Failed to check existing instance: %v", err)
	} else if isRunning {
		fmt.Fprintf(os.Stderr, "Another instance of speak-to-ai is already running (PID: %d)\n", pid)
		fmt.Fprintf(os.Stderr, "If you're sure no other instance is running, remove the lock file: %s\n", lockFile.GetLockFilePath())
		return 1
	}

	if err := lockFile.TryLock(); err != nil {
		appLogger.Error("Failed to acquire application lock: %v", err)
		return 1
	}

	// Ensure lock is released on exit
	defer func() {
		if err := lockFile.Unlock(); err != nil {
			appLogger.Warning("Failed to release lock: %v", err)
		}
	}()

	// Create application instance with service-based architecture
	application := app.NewApp(appLogger)

	if err := application.Initialize(configPath, opts.debug); err != nil {
		appLogger.Error("Failed to initialize application: %v", err)
		return 1
	}

	if err := application.RunAndWait(); err != nil {
		appLogger.Error("Application error: %v", err)
		return 1
	}

	return 0
}

type daemonOptions struct {
	configFile string
	debug      bool
}

func parseDaemonOptions(args []string) (*daemonOptions, error) {
	opts := &daemonOptions{
		configFile: "config.yaml",
	}

	fs := flag.NewFlagSet("speak-to-ai", flag.ContinueOnError)
	var parseOutput strings.Builder
	fs.SetOutput(&parseOutput)

	fs.StringVar(&opts.configFile, "config", opts.configFile, "Path to configuration file")
	fs.BoolVar(&opts.debug, "debug", false, "Enable debug mode")

	fs.Usage = func() {
		printCombinedUsage(os.Stderr, fs)
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

	if remaining := fs.Args(); len(remaining) > 0 {
		fmt.Fprintf(os.Stderr, "Unknown arguments: %v\n", remaining)
		fs.Usage()
		return nil, fmt.Errorf("unexpected arguments")
	}

	return opts, nil
}

func printDaemonUsage(w io.Writer, fs *flag.FlagSet) {
	name := filepath.Base(os.Args[0])
	if _, err := fmt.Fprintf(w, "Usage: %s [daemon flags]\n\n", name); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "Daemon Flags:"); err != nil {
		reportUsageError(err)
		return
	}
	originalOutput := fs.Output()
	fs.SetOutput(w)
	fs.PrintDefaults()
	fs.SetOutput(originalOutput)
	if _, err := fmt.Fprintln(w); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintf(w, "CLI mode: %s [CLI flags] <start|stop|status|transcript>\n", name); err != nil {
		reportUsageError(err)
		return
	}
	if _, err := fmt.Fprintln(w, "For CLI flags, run with the desired command and --help."); err != nil {
		reportUsageError(err)
		return
	}
}

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
	if _, err := fmt.Fprintf(w, "Usage:\n  %s [daemon flags]\n  %s [CLI flags] <start|stop|status|transcript>\n\n", name, name); err != nil {
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

// Check for an AppImage environment and, if detected,
// modify the config file path to use the bundled configuration if available.
func adjustPathsForAppImage(logger logger.Logger, configPath string) string {
	appImagePath := os.Getenv("APPIMAGE")
	if appImagePath == "" {
		return configPath
	}

	appDir := os.Getenv("APPDIR")
	if appDir == "" {
		argv0 := os.Getenv("ARGV0")
		if argv0 != "" && strings.HasSuffix(argv0, "/AppRun") {
			appDir = filepath.Dir(argv0)
		}
	}

	if appDir == "" {
		logger.Warning("Running in AppImage but could not detect AppDir")
		return configPath
	}

	logger.Info("Running inside AppImage, base path: %s", appDir)

	if configPath == "config.yaml" {
		bundledConfig := filepath.Join(appDir, "config.yaml")
		if _, err := os.Stat(bundledConfig); err == nil {
			logger.Info("Using AppImage bundled config: %s", bundledConfig)
			return bundledConfig
		}
	}

	return configPath
}

// Check for a Flatpak environment and, if detected,
// modify the config file path to use the standard Flatpak configuration path.
func adjustPathsForFlatpak(logger logger.Logger, configPath string) string {
	flatpakInfo := os.Getenv("FLATPAK_ID")
	if flatpakInfo == "" {
		return configPath
	}

	logger.Info("Running inside Flatpak: %s", flatpakInfo)

	if configPath == "config.yaml" {
		flatpakConfigPath := "/app/share/speak-to-ai/config.yaml"
		if _, err := os.Stat(flatpakConfigPath); err == nil {
			logger.Info("Using Flatpak config: %s", flatpakConfigPath)
			return flatpakConfigPath
		}
	}

	return configPath
}
