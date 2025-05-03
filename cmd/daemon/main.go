package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/AshBuk/speak-to-ai/audio"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys"
	"github.com/AshBuk/speak-to-ai/output"
	"github.com/AshBuk/speak-to-ai/websocket"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// Command-line flags
var (
	configFile   string
	modelPath    string
	whisperPath  string
	quantizePath string
	debug        bool
)

// represents the display server type
type EnvironmentType string

const (
	// X11 display server
	EnvironmentX11 EnvironmentType = "X11"
	// Wayland display server
	EnvironmentWayland EnvironmentType = "Wayland"
	// Unknown display server
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
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override debug flag from command line if specified
	if debug {
		cfg.General.Debug = true
	}

	// Create directory for models and temporary files if it doesn't exist
	ensureDirectories(cfg)

	// Initialize model manager
	modelManager := whisper.NewModelManager(cfg)

	// Get model path from manager for immediate usage
	modelFilePath, err := modelManager.GetModelPath()
	if err != nil {
		log.Printf("Warning: Failed to get model path: %v", err)
		log.Println("Will continue startup and attempt to load model later")
	}

	// Initialize audio recorder
	recorder, err := audio.GetRecorder(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize audio recorder: %v", err)
	}

	// Initialize whisper engine
	whisperEngine := whisper.NewWhisperEngine(cfg, whisperPath, modelFilePath)

	// Initialize output manager based on environment
	outputEnv := convertEnvironmentType(env)
	outputter, err := output.GetOutputterFromConfig(cfg, outputEnv)
	if err != nil {
		log.Printf("Warning: Failed to initialize text outputter: %v", err)
	}

	// Convert string environment to hotkeys.EnvironmentType
	var hotkeyEnv hotkeys.EnvironmentType
	switch env {
	case "X11":
		hotkeyEnv = hotkeys.EnvironmentX11
	case "Wayland":
		hotkeyEnv = hotkeys.EnvironmentWayland
	default:
		hotkeyEnv = hotkeys.EnvironmentUnknown
	}

	// Create hotkey config adapter
	hotkeyConfig := hotkeys.NewConfigAdapter(
		cfg.Hotkeys.StartRecording,
		cfg.Hotkeys.CopyToClipboard,
		cfg.Hotkeys.PasteToActiveApp,
	)

	// Initialize hotkey manager with environment information
	hotkeyManager := hotkeys.NewHotkeyManager(hotkeyConfig, hotkeyEnv)

	// Initialize WebSocket server
	// Create a simple logger implementation
	logger := &simpleLogger{}
	wsServer := websocket.NewWebSocketServer(cfg, recorder, whisperEngine, logger)

	// Start WebSocket server if enabled
	if cfg.WebServer.Enabled {
		go wsServer.Start()
	}

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

				// Send transcript to WebSocket clients if server is enabled
				if cfg.WebServer.Enabled {
					wsServer.BroadcastMessage("transcript", r.transcript)
				}

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
	if cfg.WebServer.Enabled {
		wsServer.Stop()
	}

	// Cleanup any temp files using the interface
	if recorder != nil {
		if err := recorder.CleanupFile(); err != nil {
			log.Printf("Warning: failed to cleanup temporary file: %v", err)
		}
	}

	log.Println("Daemon shutdown complete")
}

// detectEnvironment detects the current display server environment
func detectEnvironment() EnvironmentType {
	// Check for Wayland
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return EnvironmentWayland
	}

	// Check for X11
	if os.Getenv("DISPLAY") != "" {
		return EnvironmentX11
	}

	// If neither is detected, assume X11 as fallback
	return EnvironmentUnknown
}

// ensureDirectories creates necessary directories for the application
func ensureDirectories(cfg *config.Config) {
	// Create model directory if it doesn't exist
	modelsDir := filepath.Dir(cfg.General.ModelPath)
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		log.Printf("Warning: Failed to create models directory: %v", err)
	}

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(cfg.General.TempAudioPath, 0755); err != nil {
		log.Printf("Warning: Failed to create temp directory: %v", err)
	}
}

// convertEnvironmentType converts EnvironmentType to output.EnvironmentType
func convertEnvironmentType(env EnvironmentType) output.EnvironmentType {
	switch env {
	case EnvironmentX11:
		return output.EnvironmentX11
	case EnvironmentWayland:
		return output.EnvironmentWayland
	default:
		return output.EnvironmentUnknown
	}
}

// simpleLogger implements the Logger interface for websocket
type simpleLogger struct{}

func (l *simpleLogger) Info(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *simpleLogger) Debug(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func (l *simpleLogger) Warning(format string, args ...interface{}) {
	log.Printf("[WARNING] "+format, args...)
}

func (l *simpleLogger) Error(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
