//go:build systray

package tray

import (
	"fmt"
	"log"

	"github.com/AshBuk/speak-to-ai/config"
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
	config         *config.Config

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
	outputMenu  *systray.MenuItem

	// Dynamic settings items
	hotkeyItems map[string]*systray.MenuItem
	audioItems  map[string]*systray.MenuItem
	modelItems  map[string]*systray.MenuItem
	outputItems map[string]*systray.MenuItem
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
		hotkeyItems:    make(map[string]*systray.MenuItem),
		audioItems:     make(map[string]*systray.MenuItem),
		modelItems:     make(map[string]*systray.MenuItem),
		outputItems:    make(map[string]*systray.MenuItem),
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
	tm.outputMenu = tm.settingsItem.AddSubMenuItem("📋 Output", "Output settings")

	// Populate with initial values if config is available
	if tm.config != nil {
		tm.populateSettingsMenus()
	}
}

// populateSettingsMenus populates the settings submenus with current config values
func (tm *TrayManager) populateSettingsMenus() {
	if tm.config == nil {
		return
	}

	// Populate Hotkeys menu
	tm.hotkeyItems["start_recording"] = tm.hotkeysMenu.AddSubMenuItem(
		fmt.Sprintf("Start Recording: %s", tm.config.Hotkeys.StartRecording),
		"Current start recording hotkey",
	)
	tm.hotkeyItems["start_recording"].Disable()

	// Populate Audio menu
	tm.audioItems["device"] = tm.audioMenu.AddSubMenuItem(
		fmt.Sprintf("Device: %s", tm.config.Audio.Device),
		"Current audio device",
	)
	tm.audioItems["device"].Disable()

	tm.audioItems["sample_rate"] = tm.audioMenu.AddSubMenuItem(
		fmt.Sprintf("Sample Rate: %d Hz", tm.config.Audio.SampleRate),
		"Current sample rate",
	)
	tm.audioItems["sample_rate"].Disable()

	tm.audioItems["method"] = tm.audioMenu.AddSubMenuItem(
		fmt.Sprintf("Method: %s", tm.config.Audio.RecordingMethod),
		"Current recording method",
	)
	tm.audioItems["method"].Disable()

	// Populate AI Model menu
	tm.modelItems["type"] = tm.aiModelMenu.AddSubMenuItem(
		fmt.Sprintf("Model: %s", tm.config.General.ModelType),
		"Current model type",
	)
	tm.modelItems["type"].Disable()

	tm.modelItems["language"] = tm.aiModelMenu.AddSubMenuItem(
		fmt.Sprintf("Language: %s", tm.config.General.Language),
		"Current language setting",
	)
	tm.modelItems["language"].Disable()

	// Populate Output menu
	tm.outputItems["mode"] = tm.outputMenu.AddSubMenuItem(
		fmt.Sprintf("Mode: %s", tm.config.Output.DefaultMode),
		"Current output mode",
	)
	tm.outputItems["mode"].Disable()

	tm.outputItems["clipboard_tool"] = tm.outputMenu.AddSubMenuItem(
		fmt.Sprintf("Clipboard Tool: %s", tm.config.Output.ClipboardTool),
		"Current clipboard tool",
	)
	tm.outputItems["clipboard_tool"].Disable()

	tm.outputItems["type_tool"] = tm.outputMenu.AddSubMenuItem(
		fmt.Sprintf("Type Tool: %s", tm.config.Output.TypeTool),
		"Current typing tool",
	)
	tm.outputItems["type_tool"].Disable()
}

// UpdateSettings updates the settings display with new configuration
func (tm *TrayManager) UpdateSettings(config *config.Config) {
	tm.config = config

	// Update existing menu items if they exist
	if tm.hotkeyItems["start_recording"] != nil {
		tm.hotkeyItems["start_recording"].SetTitle(fmt.Sprintf("Start Recording: %s", config.Hotkeys.StartRecording))
	}

	if tm.audioItems["device"] != nil {
		tm.audioItems["device"].SetTitle(fmt.Sprintf("Device: %s", config.Audio.Device))
	}

	if tm.audioItems["sample_rate"] != nil {
		tm.audioItems["sample_rate"].SetTitle(fmt.Sprintf("Sample Rate: %d Hz", config.Audio.SampleRate))
	}

	if tm.audioItems["method"] != nil {
		tm.audioItems["method"].SetTitle(fmt.Sprintf("Method: %s", config.Audio.RecordingMethod))
	}

	if tm.modelItems["type"] != nil {
		tm.modelItems["type"].SetTitle(fmt.Sprintf("Model: %s", config.General.ModelType))
	}

	if tm.modelItems["language"] != nil {
		tm.modelItems["language"].SetTitle(fmt.Sprintf("Language: %s", config.General.Language))
	}

	if tm.outputItems["mode"] != nil {
		tm.outputItems["mode"].SetTitle(fmt.Sprintf("Mode: %s", config.Output.DefaultMode))
	}

	if tm.outputItems["clipboard_tool"] != nil {
		tm.outputItems["clipboard_tool"].SetTitle(fmt.Sprintf("Clipboard Tool: %s", config.Output.ClipboardTool))
	}

	if tm.outputItems["type_tool"] != nil {
		tm.outputItems["type_tool"].SetTitle(fmt.Sprintf("Type Tool: %s", config.Output.TypeTool))
	}
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
