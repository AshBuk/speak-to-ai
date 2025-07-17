package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/AshBuk/speak-to-ai/audio"
	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/hotkeys"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	"github.com/AshBuk/speak-to-ai/internal/notify"
	"github.com/AshBuk/speak-to-ai/internal/platform"
	"github.com/AshBuk/speak-to-ai/internal/tray"
	"github.com/AshBuk/speak-to-ai/output"
	"github.com/AshBuk/speak-to-ai/websocket"
	"github.com/AshBuk/speak-to-ai/whisper"
)

// App represents the application with all its components
type App struct {
	Logger          logger.Logger
	Config          *config.Config
	ConfigFile      string // Path to the configuration file
	Environment     platform.EnvironmentType
	ModelManager    *whisper.ModelManager
	Recorder        audio.AudioRecorder
	WhisperEngine   *whisper.WhisperEngine
	OutputManager   output.Outputter
	HotkeyManager   *hotkeys.HotkeyManager
	WebSocketServer *websocket.WebSocketServer
	TrayManager     tray.TrayManagerInterface
	NotifyManager   *notify.NotificationManager
	LastTranscript  string
	ShutdownCh      chan os.Signal
	Ctx             context.Context
	Cancel          context.CancelFunc
}

// NewApp creates a new application instance with the specified configuration
func NewApp(configFile string, debug bool, modelPath, quantizePath string) *App {
	app := &App{
		ShutdownCh: make(chan os.Signal, 1),
	}

	// Create context with cancellation for clean shutdown
	app.Ctx, app.Cancel = context.WithCancel(context.Background())

	// Set up signal handling
	signal.Notify(app.ShutdownCh, os.Interrupt, syscall.SIGTERM)

	return app
}
