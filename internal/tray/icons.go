// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"github.com/AshBuk/speak-to-ai/internal/logger"
)

// GetIconMicOff returns the binary data for the microphone-off icon
func GetIconMicOff(loggers ...logger.Logger) []byte {
	logSink := resolveLogger(loggers...)
	if data, ok := loadIconFromAppImage(); ok {
		return data
	}
	return mustDecodeIcon(iconMicOffBase64, logSink)
}

// GetIconMicOn returns the binary data for the microphone-on icon
func GetIconMicOn(loggers ...logger.Logger) []byte {
	logSink := resolveLogger(loggers...)
	if data, ok := loadIconFromAppImage(); ok {
		return data
	}
	return mustDecodeIcon(iconMicOnBase64, logSink)
}

// mustDecodeIcon decodes a base64-gzipped icon
func mustDecodeIcon(encoded string, logSink logger.Logger) []byte {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic("Failed to decode icon: " + err.Error())
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic("Failed to create gzip reader: " + err.Error())
	}
	defer func() {
		if err := gzipReader.Close(); err != nil {
			logSink.Warning("Failed to close gzip reader for icon: %v", err)
		}
	}()

	var buf bytes.Buffer
	// Limit decompressed size to 5MB to mitigate decompression bombs
	limited := io.LimitReader(gzipReader, 5*1024*1024)
	if _, err := io.Copy(&buf, limited); err != nil {
		panic("Failed to decompress icon: " + err.Error())
	}

	return buf.Bytes()
}

// loadIconFromAppImage tries to load an icon shipped inside the AppImage
func loadIconFromAppImage() ([]byte, bool) {
	appDir := os.Getenv("APPDIR")
	if appDir == "" {
		return nil, false
	}
	candidates := []string{
		filepath.Join(appDir, "speak-to-ai.png"),
		filepath.Join(appDir, "usr/share/icons/hicolor/256x256/apps/speak-to-ai.png"),
	}
	for _, p := range candidates {
		clean := filepath.Clean(p)
		if data, err := os.ReadFile(clean); err == nil && len(data) > 0 {
			return data, true
		}
	}
	return nil, false
}

func resolveLogger(loggers ...logger.Logger) logger.Logger {
	if len(loggers) > 0 && loggers[0] != nil {
		return loggers[0]
	}
	return logger.NewDefaultLogger(logger.WarningLevel)
}

// Base64-encoded gzipped PNG icons
// Generated with: cat icon.png | gzip -9 | base64 -w 0

// Microphone off icon (gray)
const iconMicOffBase64 = `H4sIAAAAAAACA+sM8HPn5ZLiYmBg4PX0cAkC0gIgzAEkGKxmLNgLpJiSvN1dGP6395/ZD+Sxl3j6urK/5BAVZTJYomVvDBQS9HRxDJG4nLwm2YHVh0c5ioFhehHDPK4+xtdASdUS14iSlMSSVKvkolQgxWBkYGSqa2Cha2QYYmRoZWBkZWKhbWBgZWCw24MhE0VDbn5KZlolbg2nRHdcBWrQgGsoycxNLS5JzC3ArWcuw0yQZxk8Xf1c1jklNAEAa1L7qgEBAAA=`

// Microphone on icon (red)
const iconMicOnBase64 = `H4sIAAAAAAACA+sM8HPn5ZLiYmBg4PX0cAkC0gIgzMgMJFVtc5WAlEKyR5AvA0OVGgNDQwsDwy+gUMMLBoZSAwaGVwkMDFYzGBjEC+bsCrQBSrAF+IS4/mdg+P//v6OsiSBQhDHJ292F8T+T7j0gh73E09eV/SWHqCiTwRIte2OgEI+ni2MIx/XkBAVeIM+AgfH4qtY+kOUlrhElKYklqVbJRalAisHIwMhU18BC18gwxMjQysDIysRC28DAysBgtwdDJoqG3PyUzLRK3BpOie64CtSgAddQkpmbWlySmFuAW89chpmgQGLwdPVzWeeU0AQA5nQkVjkBAAA=`
