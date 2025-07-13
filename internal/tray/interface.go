package tray

// TrayManagerInterface defines the interface for tray managers
type TrayManagerInterface interface {
	Start()
	SetRecordingState(isRecording bool)
	SetTooltip(tooltip string)
	Stop()
}
