// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package app

// RunAndWait starts all components and waits for shutdown signal
func (a *App) RunAndWait() error {
	// Start WebSocket server if enabled
	if a.Config.WebServer.Enabled {
		go func() {
			if err := a.WebSocketServer.Start(); err != nil {
				a.Logger.Info("Web interface disabled in desktop version. Integration in development: %v", err)
			}
		}()
	} else {
		a.Logger.Info("Web interface disabled. Use menu/config/hotkeys")
	}

	// Start the tray manager if available
	if a.TrayManager != nil {
		a.TrayManager.Start()
	}

	// Register hotkey callbacks
	a.registerCallbacks()

	// Start hotkey manager
	if err := a.HotkeyManager.Start(); err != nil {
		a.Logger.Warning("Failed to start hotkey manager: %v", err)
		if a.NotifyManager != nil {
			_ = a.NotifyManager.ShowNotification("Hotkeys Disabled", "Failed to initialize global hotkeys. On AppImage try 'evdev' provider and add user to 'input' group.")
		}
	}

	a.Logger.Info("Speak-to-AI is ready to use!")

	// Start goroutine to handle shutdown signal
	go func() {
		sig := <-a.ShutdownCh
		a.Logger.Info("Shutdown signal received: %s", sig)
		a.Cancel() // Cancel context to signal shutdown
	}()

	// Wait for shutdown signal
	<-a.Ctx.Done()

	return a.Shutdown()
}

// Shutdown gracefully shuts down all components
func (a *App) Shutdown() error {
	a.Logger.Info("Shutting down...")

	// Cleanup
	a.HotkeyManager.Stop()

	if a.Config.WebServer.Enabled {
		a.WebSocketServer.Stop()
	}

	// Stop tray manager if available
	if a.TrayManager != nil {
		a.TrayManager.Stop()
	}

	// Cleanup any temp files
	if a.Recorder != nil {
		if err := a.Recorder.CleanupFile(); err != nil {
			a.Logger.Warning("Failed to cleanup temporary file: %v", err)
		}
	}

	// Close whisper engines to free resources
	if a.WhisperEngine != nil {
		if err := a.WhisperEngine.Close(); err != nil {
			a.Logger.Warning("Failed to close whisper engine: %v", err)
		}
	}

	if a.StreamingEngine != nil {
		if err := a.StreamingEngine.Close(); err != nil {
			a.Logger.Warning("Failed to close streaming engine: %v", err)
		}
	}

	a.Logger.Info("Daemon shutdown complete")
	return nil
}
