package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// Command-line flags
var (
	configFile   string
	modelPath    string
	whisperPath  string
	quantizePath string
	debug        bool
)

// EnvironmentType represents the display server type
type EnvironmentType string

const (
	// EnvironmentX11 represents X11 display server
	EnvironmentX11 EnvironmentType = "X11"
	// EnvironmentWayland represents Wayland display server
	EnvironmentWayland EnvironmentType = "Wayland"
	// EnvironmentUnknown represents unknown display server
	EnvironmentUnknown EnvironmentType = "Unknown"
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
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Speak-to-AI daemon...")

	// Detect environment (X11 or Wayland)
	env := detectEnvironment()
	log.Printf("Detected environment: %s", env)

	// Load configuration
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override debug flag from command line if specified
	if debug {
		config.General.Debug = true
	}

	// Create directory for models and temporary files if it doesn't exist
	ensureDirectories(config)

	// Initialize model manager
	modelsDir := filepath.Dir(config.General.ModelPath)
	modelManager := NewModelManager(config, modelsDir, quantizePath)

	// Preload models in a background goroutine to avoid delaying startup
	go func() {
		log.Println("Preloading models...")
		if err := modelManager.PreloadModels(); err != nil {
			log.Printf("Warning: model preloading failed: %v", err)
		}
	}()

	// Get model path from manager for immediate usage
	modelFilePath, err := modelManager.GetModelPath(config.General.ModelType, config.General.ModelPrecision)
	if err != nil {
		log.Printf("Warning: Failed to get model path: %v", err)
		log.Println("Will continue startup and attempt to load model later")
	}

	// Initialize audio manager (recorder)
	recorder, err := GetRecorder(config)
	if err != nil {
		log.Fatalf("Failed to initialize audio recorder: %v", err)
	}

	// Initialize whisper engine
	whisperEngine := NewWhisperEngine(config, whisperPath, modelFilePath)

	// Initialize output manager based on environment
	outputter, err := GetOutputter(config)
	if err != nil {
		log.Printf("Warning: Failed to initialize text outputter: %v", err)
	}

	// Initialize hotkey manager with environment information
	hotkeyManager := NewHotkeyManager(config, EnvironmentType(env))

	// Initialize WebSocket server
	wsServer := NewWebSocketServer(config, recorder, whisperEngine)

	// Start WebSocket server
	go wsServer.Start()

	// Helper variable to store last transcript for clipboard/paste
	var lastTranscript string

	// Register hotkey callbacks
	hotkeyManager.RegisterCallbacks(
		// Record start callback
		func() error {
			log.Println("Starting recording...")
			return recorder.StartRecording()
		},
		// Record stop callback
		func() error {
			log.Println("Stopping recording...")
			audioFile, err := recorder.StopRecording()
			if err != nil {
				return err
			}

			// Process audio with whisper
			log.Printf("Processing audio file: %s", audioFile)

			// Set a reasonable timeout for processing (e.g. 30 seconds)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Use a channel to collect the result or error
			type result struct {
				transcript string
				err        error
			}

			resultCh := make(chan result, 1)

			go func() {
				transcript, err := whisperEngine.Transcribe(audioFile)
				resultCh <- result{transcript, err}
			}()

			// Wait for result or timeout
			select {
			case r := <-resultCh:
				if r.err != nil {
					log.Printf("Error processing audio: %v", r.err)
					return r.err
				}

				log.Printf("Transcript: %s", r.transcript)

				// Send transcript to WebSocket clients
				wsServer.BroadcastMessage("transcript", r.transcript)

				// Store transcript for clipboard/paste operations
				lastTranscript = r.transcript

			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		},
		// Copy to clipboard callback
		func() error {
			if outputter == nil {
				log.Println("Output manager not initialized")
				return nil
			}

			if lastTranscript == "" {
				log.Println("No transcript available to copy")
				return nil
			}

			log.Println("Copying transcript to clipboard")
			return outputter.CopyToClipboard(lastTranscript)
		},
		// Paste to active app callback
		func() error {
			if outputter == nil {
				log.Println("Output manager not initialized")
				return nil
			}

			if lastTranscript == "" {
				log.Println("No transcript available to paste")
				return nil
			}

			log.Println("Pasting transcript to active application")
			return outputter.TypeToActiveWindow(lastTranscript)
		},
	)

	// Start hotkey manager
	err = hotkeyManager.Start()
	if err != nil {
		log.Printf("Warning: Failed to start hotkey manager: %v", err)
	}

	// Handle signals for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		log.Printf("Shutdown signal received: %s", sig)
		cancel()
	}()

	log.Println("Speak-to-AI is ready to use!")

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down...")

	// Cleanup
	hotkeyManager.Stop()

	// Stop WebSocket server
	wsServer.Stop()

	// Cleanup any temp files using the interface
	if recorder != nil {
		if err := recorder.CleanupFile(); err != nil {
			log.Printf("Warning: failed to cleanup temporary file: %v", err)
		}
	}

	log.Println("Daemon shutdown complete")
}

// detectEnvironment determines if the system is running X11 or Wayland
func detectEnvironment() EnvironmentType {
	// Check for Wayland
	wayland := os.Getenv("WAYLAND_DISPLAY")
	if wayland != "" {
		return EnvironmentWayland
	}

	// Check for X11
	x11 := os.Getenv("DISPLAY")
	if x11 != "" {
		return EnvironmentX11
	}

	// Try to detect using commands
	if commandExists("weston-info") || commandExists("swaybg") {
		return EnvironmentWayland
	}

	if commandExists("xdpyinfo") {
		return EnvironmentX11
	}

	return EnvironmentUnknown
}

// commandExists checks if a command exists in the system
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ensureDirectories creates necessary directories for the application
func ensureDirectories(config *Config) {
	// Create models directory
	modelsDir := filepath.Dir(config.General.ModelPath)
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		log.Printf("Warning: Failed to create models directory: %v", err)
	}

	// Create temp directory for audio recordings if specified
	tempDir := config.General.TempAudioPath
	if tempDir != "" {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			log.Printf("Warning: Failed to create temp directory: %v", err)
		}
	}
}
