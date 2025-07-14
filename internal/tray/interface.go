package tray

import "github.com/AshBuk/speak-to-ai/config"

// TrayManagerInterface defines the interface for tray managers
type TrayManagerInterface interface {
	Start()
	SetRecordingState(isRecording bool)
	SetTooltip(tooltip string)
	UpdateSettings(config *config.Config)
	Stop()
}
