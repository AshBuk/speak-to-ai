//go:build systray

package tray

import (
	"log"

	"github.com/getlantern/systray"
)

// TrayManager manages the system tray icon and menu
type TrayManager struct {
	isRecording    bool
	iconMicOff     []byte
	iconMicOn      []byte
	onExit         func()
	onToggle       func() error
	onShowConfig   func() error
	onReloadConfig func() error

	// Menu items
	toggleItem       *systray.MenuItem
	statusItem       *systray.MenuItem
	settingsItem     *systray.MenuItem
	showConfigItem   *systray.MenuItem
	reloadConfigItem *systray.MenuItem
	exitItem         *systray.MenuItem

	// Settings submenus
	hotkeysMenu *systray.MenuItem
	audioMenu   *systray.MenuItem
	aiModelMenu *systray.MenuItem
}

// NewTrayManager creates a new tray manager instance
func NewTrayManager(iconMicOff, iconMicOn []byte, onExit func(), onToggle func() error, onShowConfig func() error, onReloadConfig func() error) *TrayManager {
	return &TrayManager{
		isRecording:    false,
		iconMicOff:     iconMicOff,
		iconMicOn:      iconMicOn,
		onExit:         onExit,
		onToggle:       onToggle,
		onShowConfig:   onShowConfig,
		onReloadConfig: onReloadConfig,
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

	// Create main menu items
	tm.toggleItem = systray.AddMenuItem("🔴 Start Recording", "Start/Stop recording")
	tm.statusItem = systray.AddMenuItem("ℹ️  Status: Ready", "Current status")
	tm.statusItem.Disable() // Status item is just for display

	systray.AddSeparator()

	// Settings submenu
	tm.settingsItem = systray.AddMenuItem("⚙️  Settings", "Application settings")
	tm.createSettingsSubmenus()

	systray.AddSeparator()

	// Config actions
	tm.showConfigItem = systray.AddMenuItem("📄 Show Config File", "Open configuration file")
	tm.reloadConfigItem = systray.AddMenuItem("🔄 Reload Config", "Reload configuration from file")

	systray.AddSeparator()

	tm.exitItem = systray.AddMenuItem("❌ Quit", "Quit Speak-to-AI")

	// Handle menu item clicks
	go tm.handleMenuClicks()
}

// createSettingsSubmenus creates the settings submenus
func (tm *TrayManager) createSettingsSubmenus() {
	tm.hotkeysMenu = tm.settingsItem.AddSubMenuItem("🎹 Hotkeys", "Hotkey settings")
	tm.audioMenu = tm.settingsItem.AddSubMenuItem("🎤 Audio", "Audio settings")
	tm.aiModelMenu = tm.settingsItem.AddSubMenuItem("🤖 AI Model", "AI model settings")
}

// handleMenuClicks handles all menu item clicks
func (tm *TrayManager) handleMenuClicks() {
	for {
		select {
		case <-tm.toggleItem.ClickedCh:
			log.Println("Toggle recording clicked")
			if err := tm.onToggle(); err != nil {
				log.Printf("Error toggling recording: %v", err)
			}
		case <-tm.showConfigItem.ClickedCh:
			log.Println("Show config clicked")
			if tm.onShowConfig != nil {
				if err := tm.onShowConfig(); err != nil {
					log.Printf("Error showing config: %v", err)
				}
			}
		case <-tm.reloadConfigItem.ClickedCh:
			log.Println("Reload config clicked")
			if tm.onReloadConfig != nil {
				if err := tm.onReloadConfig(); err != nil {
					log.Printf("Error reloading config: %v", err)
				}
			}
		case <-tm.exitItem.ClickedCh:
			log.Println("Exit clicked")
			systray.Quit()
			return
		}
	}
}

// SetRecordingState updates the tray icon and menu to reflect recording state
func (tm *TrayManager) SetRecordingState(isRecording bool) {
	tm.isRecording = isRecording

	if isRecording {
		systray.SetIcon(tm.iconMicOn)
		systray.SetTooltip("Speak-to-AI: Recording...")
		tm.toggleItem.SetTitle("🔴 Stop Recording")
		tm.statusItem.SetTitle("ℹ️  Status: Recording")
	} else {
		systray.SetIcon(tm.iconMicOff)
		systray.SetTooltip("Speak-to-AI: Ready")
		tm.toggleItem.SetTitle("🔴 Start Recording")
		tm.statusItem.SetTitle("ℹ️  Status: Ready")
	}
}

// SetTooltip sets the tooltip text for the tray icon
func (tm *TrayManager) SetTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}

// Stop stops the tray manager
func (tm *TrayManager) Stop() {
	systray.Quit()
}
