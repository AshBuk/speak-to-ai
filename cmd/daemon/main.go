// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/internal/app"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/utils"
)

var (
	configFile string
	debug      bool
)

func init() {
	// Parse command-line arguments
	flag.StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.Parse()
}

func main() {
	// Create logger early for consistent logging
	logLevel := logger.InfoLevel
	if debug {
		logLevel = logger.DebugLevel
	}
	appLogger := logger.NewDefaultLogger(logLevel)

	// Adjust paths for AppImage and Flatpak environments to find bundled assets.
	adjustPathsForAppImage(appLogger)
	adjustPathsForFlatpak(appLogger)

	// Single-instance protection
	lockFile := utils.NewLockFile(utils.GetDefaultLockPath())
	if isRunning, pid, err := lockFile.CheckExistingInstance(); err != nil {
		appLogger.Warning("Failed to check existing instance: %v", err)
	} else if isRunning {
		fmt.Fprintf(os.Stderr, "Another instance of speak-to-ai is already running (PID: %d)\n", pid)
		fmt.Fprintf(os.Stderr, "If you're sure no other instance is running, remove the lock file: %s\n", lockFile.GetLockFilePath())
		os.Exit(1)
	}

	if err := lockFile.TryLock(); err != nil {
		appLogger.Error("Failed to acquire application lock: %v", err)
		os.Exit(1)
	}

	// Ensure lock is released on exit
	defer func() {
		if err := lockFile.Unlock(); err != nil {
			appLogger.Warning("Failed to release lock: %v", err)
		}
	}()

	// Create application instance with service-based architecture
	application := app.NewApp(appLogger)

	if err := application.Initialize(configFile, debug); err != nil {
		appLogger.Error("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	if err := application.RunAndWait(); err != nil {
		appLogger.Error("Application error: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// Check for an AppImage environment and, if detected,
// modify the config file path to use the bundled configuration if available.
func adjustPathsForAppImage(logger logger.Logger) {
	appImagePath := os.Getenv("APPIMAGE")
	if appImagePath == "" {
		return
	}

	appDir := os.Getenv("APPDIR")
	if appDir == "" {
		// Try to detect AppDir from ARGV0 as a fallback.
		argv0 := os.Getenv("ARGV0")
		if argv0 != "" && strings.HasSuffix(argv0, "/AppRun") {
			appDir = filepath.Dir(argv0)
		}
	}

	if appDir == "" {
		logger.Warning("Running in AppImage but could not detect AppDir")
		return
	}

	logger.Info("Running inside AppImage, base path: %s", appDir)

	// If the default config path is in use, prefer the bundled config if it exists.
	if configFile == "config.yaml" {
		bundledConfig := filepath.Join(appDir, "config.yaml")
		if _, err := os.Stat(bundledConfig); err == nil {
			configFile = bundledConfig
			logger.Info("Using AppImage bundled config: %s", configFile)
		}
	}
}

// Check for a Flatpak environment and, if detected,
// modify the config file path to use the standard Flatpak configuration path.
func adjustPathsForFlatpak(logger logger.Logger) {
	flatpakInfo := os.Getenv("FLATPAK_ID")
	if flatpakInfo == "" {
		return
	}

	logger.Info("Running inside Flatpak: %s", flatpakInfo)

	// If the default config path is in use, switch to the standard Flatpak path.
	if configFile == "config.yaml" {
		flatpakConfigPath := "/app/share/speak-to-ai/config.yaml"
		if _, err := os.Stat(flatpakConfigPath); err == nil {
			configFile = flatpakConfigPath
			logger.Info("Using Flatpak config: %s", configFile)
		}
	}
}
