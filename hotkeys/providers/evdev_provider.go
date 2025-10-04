//go:build linux

// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AshBuk/speak-to-ai/hotkeys/utils"
	"github.com/AshBuk/speak-to-ai/internal/logger"
	evdev "github.com/holoplot/go-evdev"
)

// Implements KeyboardEventProvider using the Linux evdev interface
type EvdevKeyboardProvider struct {
	devices       []*evdev.InputDevice
	callbacks     map[string]func() error
	stopListening chan bool
	isListening   bool
	modifierState map[string]bool // Tracks the state of modifier keys
	mutex         sync.RWMutex
	logger        logger.Logger
	wg            sync.WaitGroup // Tracks device listener goroutines
	stopping      int32          // atomic flag to soften expected errors during shutdown
}

// Create a new EvdevKeyboardProvider instance
func NewEvdevKeyboardProvider(logger logger.Logger) *EvdevKeyboardProvider {
	return &EvdevKeyboardProvider{
		callbacks:     make(map[string]func() error),
		stopListening: make(chan bool),
		isListening:   false,
		modifierState: make(map[string]bool),
		logger:        logger,
	}
}

// Check if the evdev provider is supported on the current system
func (p *EvdevKeyboardProvider) IsSupported() bool {
	// Attempt to find keyboard devices without keeping them open
	devices, err := p.findKeyboardDevices()
	if err != nil || len(devices) == 0 {
		return false
	}

	// Close the devices after the check
	for _, dev := range devices {
		if err := dev.Close(); err != nil {
			p.logger.Error("Failed to close evdev device: %v", err)
		}
	}

	return true
}

// Find all input devices that are identified as keyboards
func (p *EvdevKeyboardProvider) findKeyboardDevices() ([]*evdev.InputDevice, error) {
	devices := []*evdev.InputDevice{}

	devicePaths, err := filepath.Glob("/dev/input/event*")
	if err != nil {
		return nil, fmt.Errorf("failed to list input devices: %w", err)
	}

	// Filter for devices that are keyboards
	for _, path := range devicePaths {
		dev, err := evdev.Open(path)
		if err != nil {
			p.logger.Warning("Could not open input device %s: %v", path, err)
			continue
		}

		// Check if the device is a keyboard to avoid capturing mice, etc
		devName, _ := dev.Name()
		if strings.Contains(strings.ToLower(devName), "keyboard") ||
			isKeyboardDevice(dev) {
			devices = append(devices, dev)
		} else {
			if err := dev.Close(); err != nil {
				p.logger.Error("Failed to close evdev device: %v", err)
			}
		}
	}

	return devices, nil
}

// Check if a device exposes typical keyboard-related key codes
func isKeyboardDevice(dev *evdev.InputDevice) bool {
	// Check for EV_KEY capability
	types := dev.CapableTypes()
	hasKeyType := false
	for _, evType := range types {
		if evType == evdev.EV_KEY {
			hasKeyType = true
			break
		}
	}
	if !hasKeyType {
		return false
	}

	// The presence of common letter keys strongly indicates a keyboard
	events := dev.CapableEvents(evdev.EV_KEY)
	common := map[uint16]bool{16: true, 30: true, 44: true, 57: true} // q, a, z, space
	for _, code := range events {
		if common[uint16(code)] {
			return true
		}
	}
	return false
}

// Start listening for keyboard events from all found devices
func (p *EvdevKeyboardProvider) Start() error {
	if p.isListening {
		return fmt.Errorf("evdev keyboard provider already started")
	}

	var err error
	// (Re)initialize stop channel on each start to avoid using a closed channel
	p.stopListening = make(chan bool)
	// clear stopping flag
	atomic.StoreInt32(&p.stopping, 0)
	p.devices, err = p.findKeyboardDevices()
	if err != nil {
		return fmt.Errorf("failed to find keyboard devices: %w", err)
	}

	if len(p.devices) == 0 {
		return fmt.Errorf("no keyboard devices found")
	}

	p.isListening = true

	// Start a listener goroutine for each device
	for i := range p.devices {
		idx := i
		p.wg.Add(1)
		go func(deviceIdx int) {
			defer p.wg.Done()
			p.listenDevice(deviceIdx)
		}(idx)
	}

	return nil
}

// listenDevice listens to one device events and exits on stop signal or critical read error
func (p *EvdevKeyboardProvider) listenDevice(idx int) {
	for {
		select {
		case <-p.stopListening:
			return
		default:
		}

		event, err := p.devices[idx].ReadOne()
		if err != nil {
			// Exit on device read error to avoid infinite loops
			// During shutdown, the device is expected to be closed; downgrade to warning
			if atomic.LoadInt32(&p.stopping) == 1 {
				p.logger.Warning("Device read ended: %v", err)
			} else {
				p.logger.Error("Device read error: %v", err)
			}
			return
		}

		if event.Type == evdev.EV_KEY {
			p.handleKeyEvent(idx, event)
		}
	}
}

// Process a key event from a device
func (p *EvdevKeyboardProvider) handleKeyEvent(_ int, event *evdev.InputEvent) {
	keyCode := int(event.Code)
	keyName := utils.GetKeyName(keyCode)
	if keyName == "" {
		keyName = fmt.Sprintf("KEY_%d", keyCode)
	}

	// Track the state of modifier keys
	if utils.IsModifierKey(keyName) {
		// Value 1 = key down, 0 = key up
		p.mutex.Lock()
		p.modifierState[strings.ToLower(keyName)] = (event.Value == 1)
		p.mutex.Unlock()
	}

	// Only process key-down events for hotkey triggers
	if event.Value != 1 {
		return
	}

	// Create a thread-safe copy of callbacks and modifier state
	p.mutex.RLock()
	callbacksCopy := make(map[string]func() error, len(p.callbacks))
	for k, v := range p.callbacks {
		callbacksCopy[k] = v
	}
	modState := make(map[string]bool, len(p.modifierState))
	for k, v := range p.modifierState {
		modState[k] = v
	}
	p.mutex.RUnlock()

	// Check for registered hotkeys
	for hotkeyStr, callback := range callbacksCopy {
		hotkey := utils.ParseHotkey(hotkeyStr)

		// Check if the pressed key matches the hotkey's main key
		if strings.EqualFold(hotkey.Key, keyName) {
			// Check if all required modifiers are also pressed
			allModifiersPressed := true
			for _, mod := range hotkey.Modifiers {
				if !utils.IsModifierPressed(mod, modState) {
					allModifiersPressed = false
					break
				}
			}

			if allModifiersPressed {
				go func(cb func() error) {
					if err := cb(); err != nil {
						p.logger.Error("Hotkey callback error: %v", err)
					}
				}(callback)
			}
		}
	}
}

// Register a callback for a hotkey
func (p *EvdevKeyboardProvider) RegisterHotkey(hotkey string, callback func() error) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.callbacks[hotkey] = callback
	return nil
}

// Stop listening for keyboard events and close all devices
func (p *EvdevKeyboardProvider) Stop() {
	if !p.isListening {
		return
	}

	// Signal all listener goroutines to stop
	if p.stopListening != nil {
		// make Stop idempotent: close only once and nil the channel
		close(p.stopListening)
		p.stopListening = nil
	}

	// Close all device handles to unblock any pending reads
	atomic.StoreInt32(&p.stopping, 1)
	for _, dev := range p.devices {
		if err := dev.Close(); err != nil {
			// If already closed, this is expected during shutdown
			p.logger.Warning("Evdev device close (ignored): %v", err)
		}
	}

	// Wait for all device listeners to exit
	p.wg.Wait()

	p.devices = nil
	p.isListening = false
	atomic.StoreInt32(&p.stopping, 0)
}

// Start a short-lived capture session to get a single hotkey combination
func (p *EvdevKeyboardProvider) CaptureOnce(timeout time.Duration) (string, error) {
	devices, err := p.findKeyboardDevices()
	if err != nil {
		return "", fmt.Errorf("failed to find keyboard devices: %w", err)
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("no keyboard devices found")
	}

	// Use a local modifier state for this capture session with mutex protection
	var modStateMutex sync.Mutex
	modState := map[string]bool{}
	resultCh := make(chan string, 1)
	stopCh := make(chan struct{})
	var captureWg sync.WaitGroup

	for i := range devices {
		idx := i
		captureWg.Add(1)
		go func() {
			defer captureWg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
				}
				ev, err := devices[idx].ReadOne()
				if err != nil {
					return
				}
				if ev.Type != evdev.EV_KEY {
					continue
				}

				keyCode := int(ev.Code)
				keyName := utils.GetKeyName(keyCode)
				if keyName == "" {
					keyName = fmt.Sprintf("KEY_%d", keyCode)
				}

				// Ignore non-keyboard buttons (e.g., mouse buttons)
				if strings.HasPrefix(keyName, "KEY_") {
					continue
				}

				// Track modifier state with mutex protection
				if utils.IsModifierKey(keyName) {
					modStateMutex.Lock()
					modState[strings.ToLower(keyName)] = (ev.Value == 1)
					modStateMutex.Unlock()
					continue
				}

				// Only consider key-down for the main hotkey
				if ev.Value != 1 {
					continue
				}

				// Cancel if Esc is pressed without modifiers
				modStateMutex.Lock()
				isCancelled := utils.CheckCancelCondition(keyName, modState)
				modStateMutex.Unlock()

				if isCancelled {
					select {
					case resultCh <- "":
					default:
					}
					return
				}

				modStateMutex.Lock()
				mods := utils.BuildModifierState(modState)
				modStateMutex.Unlock()

				combo := utils.BuildHotkeyString(mods, keyName)
				if err := utils.ValidateHotkey(combo); err != nil {
					continue
				}

				select {
				case resultCh <- combo:
				default:
				}
				return
			}
		}()
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var result string
	select {
	case result = <-resultCh:
		close(stopCh)
	case <-timer.C:
		close(stopCh)
		result = ""
	}

	// Wait for all capture goroutines to finish BEFORE closing devices
	captureWg.Wait()

	// Now safe to close devices after all goroutines stopped
	for _, d := range devices {
		if closeErr := d.Close(); closeErr != nil {
			// Already closed or error - not critical
			p.logger.Warning("Evdev device close after capture: %v", closeErr)
		}
	}

	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("capture cancelled")
	}
	return result, nil
}
