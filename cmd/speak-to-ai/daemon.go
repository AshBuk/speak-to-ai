// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AshBuk/speak-to-ai/internal/app"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

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
