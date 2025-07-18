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

	// Close whisper engine to free resources
	if a.WhisperEngine != nil {
		if err := a.WhisperEngine.Close(); err != nil {
			a.Logger.Warning("Failed to close whisper engine: %v", err)
		}
	}

	a.Logger.Info("Daemon shutdown complete")
	return nil
}
