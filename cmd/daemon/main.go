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

// Command-line flags
var (
	configFile   string
	modelPath    string
	quantizePath string
	debug        bool
)

func init() {
	// Parse command-line arguments
	flag.StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&modelPath, "model", "", "Path to whisper model file")
	flag.StringVar(&quantizePath, "quantize", "sources/core/quantize", "Path to quantize binary")
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

	// Adjust paths for AppImage and Flatpak environments
	adjustPathsForAppImage(appLogger)
	adjustPathsForFlatpak(appLogger)

	// Single-instance protection
	lockFile := utils.NewLockFile(utils.GetDefaultLockPath())

	// Check if another instance is already running
	if isRunning, pid, err := lockFile.CheckExistingInstance(); err != nil {
		appLogger.Warning("Failed to check existing instance: %v", err)
	} else if isRunning {
		fmt.Fprintf(os.Stderr, "Another instance of speak-to-ai is already running (PID: %d)\n", pid)
		fmt.Fprintf(os.Stderr, "If you're sure no other instance is running, remove the lock file: %s\n", lockFile.GetLockFilePath())
		os.Exit(1)
	}

	// Try to acquire the lock
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

	// Initialize the application
	if err := application.Initialize(configFile, debug, modelPath, quantizePath); err != nil {
		appLogger.Error("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// Run the application and wait for shutdown
	if err := application.RunAndWait(); err != nil {
		appLogger.Error("Application error: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// adjustPathsForAppImage detects if running inside an AppImage and adjusts paths accordingly
func adjustPathsForAppImage(logger logger.Logger) {
	// Check for AppImage environment
	appImagePath := os.Getenv("APPIMAGE")
	if appImagePath == "" {
		// Not running in AppImage, use default paths
		return
	}

	// Get AppImage directory
	appDir := os.Getenv("APPDIR")
	if appDir == "" {
		// Try to detect AppDir from ARGV0
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

	// Adjust paths for AppImage
	if quantizePath == "sources/core/quantize" {
		quantizePath = filepath.Join(appDir, "sources/core/quantize")
		logger.Info("Adjusted quantize path: %s", quantizePath)
	}

	// If no model path specified, check built-in model
	if modelPath == "" {
		builtinModelPath := filepath.Join(appDir, "sources/language-models/small-q5_1.bin")
		if _, err := os.Stat(builtinModelPath); err == nil {
			modelPath = builtinModelPath
			logger.Info("Using built-in model: %s", modelPath)
		}
	}

	// If default config path is used, prefer bundled config if present
	if configFile == "config.yaml" {
		bundledConfig := filepath.Join(appDir, "config.yaml")
		if _, err := os.Stat(bundledConfig); err == nil {
			configFile = bundledConfig
			logger.Info("Using AppImage bundled config: %s", configFile)
		}
	}
}

// adjustPathsForFlatpak detects if running inside a Flatpak and adjusts paths accordingly
func adjustPathsForFlatpak(logger logger.Logger) {
	// Check for Flatpak environment
	flatpakInfo := os.Getenv("FLATPAK_ID")
	if flatpakInfo == "" {
		// Not running in Flatpak, use default paths
		return
	}

	logger.Info("Running inside Flatpak: %s", flatpakInfo)

	// Adjust paths for Flatpak
	if quantizePath == "sources/core/quantize" {
		quantizePath = "/app/bin/quantize"
		logger.Info("Adjusted quantize path: %s", quantizePath)
	}

	// If no model path specified, check built-in model
	if modelPath == "" {
		builtinModelPath := "/app/share/speak-to-ai/models/small-q5_1.bin"
		if _, err := os.Stat(builtinModelPath); err == nil {
			modelPath = builtinModelPath
			logger.Info("Using built-in model: %s", modelPath)
		}
	}

	// Adjust config file path for Flatpak if not specified
	if configFile == "config.yaml" {
		flatpakConfigPath := "/app/share/speak-to-ai/config.yaml"
		if _, err := os.Stat(flatpakConfigPath); err == nil {
			configFile = flatpakConfigPath
			logger.Info("Using Flatpak config: %s", configFile)
		}
	}
}
