// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AshBuk/speak-to-ai/config"
	"github.com/AshBuk/speak-to-ai/config/models"
	"github.com/AshBuk/speak-to-ai/internal/testutils"
)

func TestConfigService(t *testing.T) {
	// Create mock logger
	mockLogger := testutils.NewMockLogger()

	// Create test config
	testConfig := &models.Config{}
	testConfig.General.Language = "en"
	testConfig.General.ModelType = "base"
	// TODO: Next feature - VAD implementation
	// testConfig.Audio.VADSensitivity = "medium"
	// testConfig.Audio.EnableVAD = false
	testConfig.Notifications.EnableWorkflowNotifications = true

	t.Run("NewConfigService", func(t *testing.T) {
		service := NewConfigService(mockLogger, testConfig, "/test/config.yaml")

		if service == nil {
			t.Fatal("NewConfigService returned nil")
		}
		if service.logger != mockLogger {
			t.Error("Logger not set correctly")
		}
		if service.config != testConfig {
			t.Error("Config not set correctly")
		}
		if service.configFile != "/test/config.yaml" {
			t.Error("Config file path not set correctly")
		}
	})

	t.Run("GetConfig", func(t *testing.T) {
		service := NewConfigService(mockLogger, testConfig, "/test/config.yaml")

		result := service.GetConfig()
		if result != testConfig {
			t.Error("GetConfig did not return the correct config")
		}
	})

	// 	t.Run("UpdateVADSensitivity", func(t *testing.T) {
	// 		// Create temporary config file
	// 		tempDir := t.TempDir()
	// 		configPath := filepath.Join(tempDir, "test_config.yaml")
	//
	// 		// Create initial config file
	// 		err := config.SaveConfig(configPath, testConfig)
	// 		if err != nil {
	// 			t.Fatalf("Failed to create test config file: %v", err)
	// 		}
	//
	// 		service := NewConfigService(mockLogger, testConfig, configPath)
	//
	// 		// Test valid sensitivity values
	// 		testCases := []string{"low", "medium", "high"}
	// 		for _, sensitivity := range testCases {
	// 			err := service.UpdateVADSensitivity(sensitivity)
	// 			if err != nil {
	// 				t.Errorf("UpdateVADSensitivity(%s) failed: %v", sensitivity, err)
	// 			}
	// 			if service.config.Audio.VADSensitivity != sensitivity {
	// 				t.Errorf("VAD sensitivity not updated to %s", sensitivity)
	// 			}
	// 		}
	//
	// 		// Test invalid sensitivity
	// 		err = service.UpdateVADSensitivity("invalid")
	// 		if err == nil {
	// 			t.Error("UpdateVADSensitivity should fail with invalid sensitivity")
	// 		}
	// 	})

	t.Run("UpdateLanguage", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.yaml")

		err := config.SaveConfig(configPath, testConfig)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		service := NewConfigService(mockLogger, testConfig, configPath)

		err = service.UpdateLanguage("ru")
		if err != nil {
			t.Errorf("UpdateLanguage failed: %v", err)
		}
		if service.config.General.Language != "ru" {
			t.Error("Language not updated correctly")
		}

		// Test same language (should not error)
		err = service.UpdateLanguage("ru")
		if err != nil {
			t.Errorf("UpdateLanguage with same language should not error: %v", err)
		}
	})

	t.Run("UpdateModelType", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.yaml")

		err := config.SaveConfig(configPath, testConfig)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		service := NewConfigService(mockLogger, testConfig, configPath)

		err = service.UpdateModelType("small")
		if err != nil {
			t.Errorf("UpdateModelType failed: %v", err)
		}
		if service.config.General.ModelType != "small" {
			t.Error("Model type not updated correctly")
		}
	})

	t.Run("ToggleWorkflowNotifications", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.yaml")

		err := config.SaveConfig(configPath, testConfig)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		service := NewConfigService(mockLogger, testConfig, configPath)

		// Initial state should be true
		if !service.config.Notifications.EnableWorkflowNotifications {
			t.Error("Initial workflow notifications state should be true")
		}

		// Toggle to false
		err = service.ToggleWorkflowNotifications()
		if err != nil {
			t.Errorf("ToggleWorkflowNotifications failed: %v", err)
		}
		if service.config.Notifications.EnableWorkflowNotifications {
			t.Error("Workflow notifications should be toggled to false")
		}

		// Toggle back to true
		err = service.ToggleWorkflowNotifications()
		if err != nil {
			t.Errorf("ToggleWorkflowNotifications failed: %v", err)
		}
		if !service.config.Notifications.EnableWorkflowNotifications {
			t.Error("Workflow notifications should be toggled to true")
		}
	})

	// 	t.Run("ToggleVAD", func(t *testing.T) {
	// 		tempDir := t.TempDir()
	// 		configPath := filepath.Join(tempDir, "test_config.yaml")
	//
	// 		err := config.SaveConfig(configPath, testConfig)
	// 		if err != nil {
	// 			t.Fatalf("Failed to create test config file: %v", err)
	// 		}
	//
	// 		service := NewConfigService(mockLogger, testConfig, configPath)
	//
	// 		// Initial state should be false
	// 		if service.config.Audio.EnableVAD {
	// 			t.Error("Initial VAD state should be false")
	// 		}
	//
	// 		// Toggle to true
	// 		err = service.ToggleVAD()
	// 		if err != nil {
	// 			t.Errorf("ToggleVAD failed: %v", err)
	// 		}
	// 		if !service.config.Audio.EnableVAD {
	// 			t.Error("VAD should be toggled to true")
	// 		}
	// 	})

	t.Run("LoadConfig", func(t *testing.T) {
		service := NewConfigService(mockLogger, testConfig, "")

		err := service.LoadConfig("/new/path/config.yaml")
		if err != nil {
			t.Errorf("LoadConfig failed: %v", err)
		}
		if service.configFile != "/new/path/config.yaml" {
			t.Error("Config file path not updated correctly")
		}
	})

	t.Run("SaveConfig_NoPath", func(t *testing.T) {
		service := NewConfigService(mockLogger, testConfig, "")

		err := service.SaveConfig()
		if err == nil {
			t.Error("SaveConfig should fail when no config file path is set")
		}
	})

	t.Run("ResetToDefaults", func(t *testing.T) {
		// Create temp config file
		tempFile, err := os.CreateTemp("", "test-config-*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		service := NewConfigService(mockLogger, testConfig, tempFile.Name())

		// Modify some settings first
		service.config.General.Language = "fr"
		// TODO: Next feature - VAD implementation
		// service.config.Audio.EnableVAD = true

		err = service.ResetToDefaults()
		if err != nil {
			t.Errorf("ResetToDefaults failed: %v", err)
		}

		// Verify settings were reset to defaults
		if service.config.General.Language != "en" {
			t.Error("Language should be reset to default 'en'")
		}
		// TODO: Next feature - VAD implementation
		// if service.config.Audio.EnableVAD != false {
		//	t.Error("EnableVAD should be reset to default false")
		// }
	})

	t.Run("Shutdown", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.yaml")

		err := config.SaveConfig(configPath, testConfig)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		service := NewConfigService(mockLogger, testConfig, configPath)

		err = service.Shutdown()
		if err != nil {
			t.Errorf("Shutdown failed: %v", err)
		}

		// Verify config was saved (file should exist)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file should exist after shutdown")
		}
	})
}
