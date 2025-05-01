package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	evdev "github.com/gvalkov/golang-evdev"
)

// KeyCombination represents a hotkey combination
type KeyCombination struct {
	Modifiers []string // Modifier keys like "ctrl", "alt", "shift"
	Key       string   // Main key
}

// HotkeyManager handles keyboard shortcuts
type HotkeyManager struct {
	config           *Config
	isListening      bool
	isRecording      bool
	stopListening    chan bool
	recordingStarted func() error
	recordingStopped func() error
	copyToClipboard  func() error
	pasteToActiveApp func() error
	hotkeysMutex     sync.Mutex
	environment      EnvironmentType
	devices          []*evdev.InputDevice // For evdev support
}

// NewHotkeyManager creates a new instance of HotkeyManager
func NewHotkeyManager(config *Config, environment EnvironmentType) *HotkeyManager {
	return &HotkeyManager{
		config:        config,
		isListening:   false,
		isRecording:   false,
		stopListening: make(chan bool),
		environment:   environment,
	}
}

// RegisterCallbacks registers callback functions for hotkey actions
func (h *HotkeyManager) RegisterCallbacks(
	recordingStarted func() error,
	recordingStopped func() error,
	copyToClipboard func() error,
	pasteToActiveApp func() error,
) {
	h.recordingStarted = recordingStarted
	h.recordingStopped = recordingStopped
	h.copyToClipboard = copyToClipboard
	h.pasteToActiveApp = pasteToActiveApp
}

// parseHotkey converts string representation to KeyCombination
func parseHotkey(hotkeyStr string) KeyCombination {
	combo := KeyCombination{}
	parts := strings.Split(hotkeyStr, "+")

	// If there's only one part, it's just a key
	if len(parts) == 1 {
		combo.Key = strings.TrimSpace(parts[0])
		return combo
	}

	// Last part is the key, the rest are modifiers
	combo.Key = strings.TrimSpace(parts[len(parts)-1])
	for i := 0; i < len(parts)-1; i++ {
		modifier := strings.ToLower(strings.TrimSpace(parts[i]))
		combo.Modifiers = append(combo.Modifiers, modifier)
	}

	return combo
}

// Start begins listening for hotkeys
func (h *HotkeyManager) Start() error {
	if h.isListening {
		return fmt.Errorf("hotkey manager is already running")
	}

	h.isListening = true

	log.Println("Starting hotkey manager...")
	log.Printf("- Start/Stop recording: %s", h.config.Hotkeys.StartRecording)
	log.Printf("- Copy to clipboard: %s", h.config.Hotkeys.CopyToClipboard)
	log.Printf("- Paste to active app: %s", h.config.Hotkeys.PasteToActiveApp)

	// Find keyboard devices
	err := h.findKeyboardDevices()
	if err != nil {
		log.Printf("Warning: failed to find keyboard devices: %v", err)
		log.Println("Hotkey support requires access to input devices")
		log.Println("Try running the application with elevated permissions")
		return err
	}

	// Start listening on all keyboard devices
	for i, device := range h.devices {
		deviceIndex := i
		go func(dev *evdev.InputDevice) {
			log.Printf("Starting evdev listener on device: %s", dev.Name)

			for {
				select {
				case <-h.stopListening:
					log.Printf("Stopping hotkey listener for device %s", dev.Name)
					return
				default:
					// Read events
					events, err := dev.Read()
					if err != nil {
						log.Printf("Error reading from device %s: %v", dev.Name, err)
						// Try to reopen the device if it was disconnected
						time.Sleep(1 * time.Second)
						continue
					}

					// Process events
					for _, event := range events {
						if event.Type == evdev.EV_KEY && event.Value == 1 { // Key press
							h.handleEvdevEvent(deviceIndex, event)
						}
					}
				}
			}
		}(device)
	}

	log.Println("Hotkey manager started")
	return nil
}

// findKeyboardDevices finds all keyboard input devices available
func (h *HotkeyManager) findKeyboardDevices() error {
	// Find all input devices
	devicesPath := "/dev/input/event*"
	devices, err := filepath.Glob(devicesPath)
	if err != nil {
		return fmt.Errorf("failed to find input devices: %w", err)
	}

	// Check each device to see if it's a keyboard
	for _, devicePath := range devices {
		device, err := evdev.Open(devicePath)
		if err != nil {
			log.Printf("Warning: cannot open device %s: %v", devicePath, err)
			continue
		}

		// Check if device has any key capabilities
		hasKeys := false
		for _, keyCode := range device.Capabilities {
			if len(keyCode) > 0 {
				hasKeys = true
				break
			}
		}

		if hasKeys {
			log.Printf("Found keyboard device: %s (%s)", device.Name, devicePath)
			h.devices = append(h.devices, device)
		}
	}

	if len(h.devices) == 0 {
		return fmt.Errorf("no keyboard devices found")
	}

	return nil
}

// handleEvdevEvent processes key events from evdev
func (h *HotkeyManager) handleEvdevEvent(deviceIndex int, event evdev.InputEvent) {
	// Convert evdev keycode to string representation
	keyCode := int(event.Code)
	keyName := evdev.KEY[keyCode]

	// Debug logging
	if h.config.General.Debug {
		log.Printf("Evdev key event: %s", keyName)
	}

	// Check if it matches any of our hotkeys
	// For simplicity, just checking exact match, not combinations yet
	if keyName == h.config.Hotkeys.StartRecording {
		h.hotkeysMutex.Lock()
		defer h.hotkeysMutex.Unlock()

		if !h.isRecording && h.recordingStarted != nil {
			log.Println("Start recording hotkey detected")
			if err := h.recordingStarted(); err != nil {
				log.Printf("Error starting recording: %v", err)
			} else {
				h.isRecording = true
			}
		}
	} else if keyName == h.config.Hotkeys.StopRecording {
		h.hotkeysMutex.Lock()
		defer h.hotkeysMutex.Unlock()

		if h.isRecording && h.recordingStopped != nil {
			log.Println("Stop recording hotkey detected")
			if err := h.recordingStopped(); err != nil {
				log.Printf("Error stopping recording: %v", err)
			} else {
				h.isRecording = false
			}
		}
	} else if keyName == h.config.Hotkeys.CopyToClipboard {
		if h.copyToClipboard != nil {
			log.Println("Copy to clipboard hotkey detected")
			if err := h.copyToClipboard(); err != nil {
				log.Printf("Error copying to clipboard: %v", err)
			}
		}
	} else if keyName == h.config.Hotkeys.PasteToActiveApp {
		if h.pasteToActiveApp != nil {
			log.Println("Paste to active app hotkey detected")
			if err := h.pasteToActiveApp(); err != nil {
				log.Printf("Error pasting to active app: %v", err)
			}
		}
	}
}

// Stop stops the hotkey listener
func (h *HotkeyManager) Stop() {
	if h.isListening {
		h.stopListening <- true
		h.isListening = false

		// The evdev library doesn't provide a Close method
		// Devices will be closed by the system when the program exits
		h.devices = nil
	}
}

// IsRecording returns the current recording state
func (h *HotkeyManager) IsRecording() bool {
	return h.isRecording
}

// SimulateHotkeyPress simulates a hotkey press for testing
func (h *HotkeyManager) SimulateHotkeyPress(hotkeyName string) error {
	h.hotkeysMutex.Lock()
	defer h.hotkeysMutex.Unlock()

	switch hotkeyName {
	case "startRecording":
		if !h.isRecording {
			log.Println("Simulating start recording hotkey")
			if h.recordingStarted != nil {
				if err := h.recordingStarted(); err != nil {
					return err
				}
				h.isRecording = true
			}
		}
	case "stopRecording":
		if h.isRecording {
			log.Println("Simulating stop recording hotkey")
			if h.recordingStopped != nil {
				if err := h.recordingStopped(); err != nil {
					return err
				}
				h.isRecording = false
			}
		}
	case "copyToClipboard":
		log.Println("Simulating copy to clipboard hotkey")
		if h.copyToClipboard != nil {
			return h.copyToClipboard()
		}
	case "pasteToActiveApp":
		log.Println("Simulating paste to active app hotkey")
		if h.pasteToActiveApp != nil {
			return h.pasteToActiveApp()
		}
	default:
		return fmt.Errorf("unknown hotkey: %s", hotkeyName)
	}
	return nil
}
