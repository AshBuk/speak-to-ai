//go:build systray

package tray

import (
	"log"

	"github.com/getlantern/systray"
)

// TrayManager manages the system tray icon and menu
type TrayManager struct {
	isRecording bool
	iconMicOff  []byte
	iconMicOn   []byte
	onExit      func()
	onToggle    func() error
	toggleItem  *systray.MenuItem
	statusItem  *systray.MenuItem
	exitItem    *systray.MenuItem
}

// NewTrayManager creates a new tray manager instance
func NewTrayManager(iconMicOff, iconMicOn []byte, onExit func(), onToggle func() error) *TrayManager {
	return &TrayManager{
		isRecording: false,
		iconMicOff:  iconMicOff,
		iconMicOn:   iconMicOn,
		onExit:      onExit,
		onToggle:    onToggle,
	}
}

// Start initializes and starts the system tray icon and menu
func (tm *TrayManager) Start() {
	go systray.Run(tm.onReady, tm.onExit)
}

// onReady sets up the system tray when it's ready
func (tm *TrayManager) onReady() {
	systray.SetIcon(tm.iconMicOff)
	systray.SetTitle("Speak-to-AI")
	systray.SetTooltip("Speak-to-AI: Offline Speech-to-Text")

	// Create menu items
	tm.toggleItem = systray.AddMenuItem("Start Recording", "Start/Stop recording")
	tm.statusItem = systray.AddMenuItem("Status: Ready", "Current status")
	tm.statusItem.Disable() // Status item is just for display
	systray.AddSeparator()
	tm.exitItem = systray.AddMenuItem("Quit", "Quit Speak-to-AI")

	// Handle menu item clicks
	go func() {
		for {
			select {
			case <-tm.toggleItem.ClickedCh:
				log.Println("Toggle recording clicked")
				if err := tm.onToggle(); err != nil {
					log.Printf("Error toggling recording: %v", err)
				}
			case <-tm.exitItem.ClickedCh:
				log.Println("Exit clicked")
				systray.Quit()
				return
			}
		}
	}()
}

// SetRecordingState updates the tray icon and menu to reflect recording state
func (tm *TrayManager) SetRecordingState(isRecording bool) {
	tm.isRecording = isRecording

	if isRecording {
		systray.SetIcon(tm.iconMicOn)
		systray.SetTooltip("Speak-to-AI: Recording...")
		tm.toggleItem.SetTitle("Stop Recording")
		tm.statusItem.SetTitle("Status: Recording")
	} else {
		systray.SetIcon(tm.iconMicOff)
		systray.SetTooltip("Speak-to-AI: Ready")
		tm.toggleItem.SetTitle("Start Recording")
		tm.statusItem.SetTitle("Status: Ready")
	}
}

// Stop stops the tray manager by quitting systray
func (tm *TrayManager) Stop() {
	systray.Quit()
}
