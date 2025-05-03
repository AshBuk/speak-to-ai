package main

import (
	"flag"
	"log"
	"os"

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