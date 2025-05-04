package app

// RunAndWait starts all components and waits for shutdown signal
func (a *App) RunAndWait() error {
	// Start WebSocket server if enabled
	if a.Config.WebServer.Enabled {
		go a.WebSocketServer.Start()
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

	a.Logger.Info("Daemon shutdown complete")
	return nil
}

// registerCallbacks registers all the hotkey callbacks
func (a *App) registerCallbacks() {
	a.HotkeyManager.RegisterCallbacks(
		// Record start callback
		a.handleStartRecording,
		// Record stop and transcribe callback
		a.handleStopRecordingAndTranscribe,
		// Copy to clipboard callback
		a.handleCopyToClipboard,
		// Paste to active app callback
		a.handlePasteToActiveWindow,
	)
}
