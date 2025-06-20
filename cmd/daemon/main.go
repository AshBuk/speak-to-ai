package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AshBuk/speak-to-ai/internal/app"
)

// Command-line flags
var (
	configFile   string
	modelPath    string
	whisperPath  string
	quantizePath string
	debug        bool
)

func init() {
	// Parse command-line arguments
	flag.StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&modelPath, "model", "", "Path to whisper model file")
	flag.StringVar(&whisperPath, "whisper", "sources/core/whisper", "Path to whisper binary")
	flag.StringVar(&quantizePath, "quantize", "sources/core/quantize", "Path to quantize binary")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.Parse()
}

func main() {
	// Adjust paths for AppImage and Flatpak environments
	adjustPathsForAppImage()
	adjustPathsForFlatpak()

	// Create application instance
	application := app.NewApp(configFile, debug, whisperPath, modelPath, quantizePath)

	// Initialize the application
	if err := application.Initialize(configFile, debug, whisperPath, modelPath, quantizePath); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// Run the application and wait for shutdown
	if err := application.RunAndWait(); err != nil {
		log.Fatalf("Application error: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// adjustPathsForAppImage detects if running inside an AppImage and adjusts paths accordingly
func adjustPathsForAppImage() {
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
		log.Println("Warning: Running in AppImage but could not detect AppDir")
		return
	}

	log.Printf("Running inside AppImage, base path: %s", appDir)

	// Adjust paths for AppImage
	if whisperPath == "sources/core/whisper" {
		whisperPath = filepath.Join(appDir, "sources/core/whisper")
		log.Printf("Adjusted whisper path: %s", whisperPath)
	}

	if quantizePath == "sources/core/quantize" {
		quantizePath = filepath.Join(appDir, "sources/core/quantize")
		log.Printf("Adjusted quantize path: %s", quantizePath)
	}

	// If no model path specified, check built-in model
	if modelPath == "" {
		builtinModelPath := filepath.Join(appDir, "sources/language-models/base.bin")
		if _, err := os.Stat(builtinModelPath); err == nil {
			modelPath = builtinModelPath
			log.Printf("Using built-in model: %s", modelPath)
		}
	}
}

// adjustPathsForFlatpak detects if running inside a Flatpak and adjusts paths accordingly
func adjustPathsForFlatpak() {
	// Check for Flatpak environment
	flatpakInfo := os.Getenv("FLATPAK_ID")
	if flatpakInfo == "" {
		// Not running in Flatpak, use default paths
		return
	}

	log.Printf("Running inside Flatpak: %s", flatpakInfo)

	// Adjust paths for Flatpak
	if whisperPath == "sources/core/whisper" {
		whisperPath = "/app/bin/whisper"
		log.Printf("Adjusted whisper path: %s", whisperPath)
	}

	if quantizePath == "sources/core/quantize" {
		quantizePath = "/app/bin/quantize"
		log.Printf("Adjusted quantize path: %s", quantizePath)
	}

	// If no model path specified, check built-in model
	if modelPath == "" {
		builtinModelPath := "/app/share/speak-to-ai/models/base.bin"
		if _, err := os.Stat(builtinModelPath); err == nil {
			modelPath = builtinModelPath
			log.Printf("Using built-in model: %s", modelPath)
		}
	}

	// Adjust config file path for Flatpak if not specified
	if configFile == "config.yaml" {
		flatpakConfigPath := "/app/share/speak-to-ai/config.yaml"
		if _, err := os.Stat(flatpakConfigPath); err == nil {
			configFile = flatpakConfigPath
			log.Printf("Using Flatpak config: %s", configFile)
		}
	}
}
