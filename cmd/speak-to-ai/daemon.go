// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/internal/app"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

// Daemon orchestrator - initializes and runs the main application
// Execution flow:
//  1. Parse flags (--config, --debug)
//  2. Bootstrap logger (Info/Debug level)
//  3. Single-instance lock (prevent multiple daemon processes)
//  4. App lifecycle: NewApp() → Initialize() → RunAndWait()
func runDaemon(args []string) int {
	opts, err := parseDaemonOptions(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	// Bootstrap
	logLevel := logger.InfoLevel
	if opts.debug {
		logLevel = logger.DebugLevel
	}
	appLogger := logger.NewDefaultLogger(logLevel)
	// Log AppImage environment if detected
	if appDir := os.Getenv("APPDIR"); appDir != "" {
		appLogger.Info("Running inside AppImage, base path: %s", appDir)
	}
	configPath := opts.configFile
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

	// Resource cleanup:
	// defer guarantees unlock even on panic or early return
	defer func() {
		if err := lockFile.Unlock(); err != nil {
			appLogger.Warning("Failed to release lock: %v", err)
		}
	}()

	// App orchestration: delegate to App module
	// → see internal/app/app.go
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

// Daemon configuration options
type daemonOptions struct {
	configFile string // Path to YAML config file (default: config.yaml)
	debug      bool   // Enable debug logging level
}

// Parse daemon command-line flags (--config, --debug)
// Returns parsed options or flag.ErrHelp if --help was requested
func parseDaemonOptions(args []string) (*daemonOptions, error) {
	defaultConfigPath, err := config.ConfigFilePath()
	if err != nil {
		defaultConfigPath = "config.yaml" // fallback to current directory
	}
	opts := &daemonOptions{
		configFile: defaultConfigPath,
	}

	fs := flag.NewFlagSet("speak-to-ai", flag.ContinueOnError) // pls don't panic on parse error
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
