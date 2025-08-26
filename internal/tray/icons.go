// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package tray

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

// GetIconMicOff returns the binary data for the microphone-off icon
func GetIconMicOff() []byte {
	return mustDecodeIcon(iconMicOffBase64)
}

// GetIconMicOn returns the binary data for the microphone-on icon
func GetIconMicOn() []byte {
	return mustDecodeIcon(iconMicOnBase64)
}

// mustDecodeIcon decodes a base64-gzipped icon
func mustDecodeIcon(encoded string) []byte {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic("Failed to decode icon: " + err.Error())
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic("Failed to create gzip reader: " + err.Error())
	}
	defer gzipReader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gzipReader); err != nil {
		panic("Failed to decompress icon: " + err.Error())
	}

	return buf.Bytes()
}

// Base64-encoded gzipped PNG icons
// Generated with: cat icon.png | gzip -9 | base64 -w 0

// Microphone off icon (gray)
const iconMicOffBase64 = `H4sIAAAAAAACA+sM8HPn5ZLiYmBg4PX0cAkC0gIgzAEkGKxmLNgLpJiSvN1dGP6395/ZD+Sxl3j6urK/5BAVZTJYomVvDBQS9HRxDJG4nLwm2YHVh0c5ioFhehHDPK4+xtdASdUS14iSlMSSVKvkolQgxWBkYGSqa2Cha2QYYmRoZWBkZWKhbWBgZWCw24MhE0VDbn5KZlolbg2nRHdcBWrQgGsoycxNLS5JzC3ArWcuw0yQZxk8Xf1c1jklNAEAa1L7qgEBAAA=`

// Microphone on icon (red)
const iconMicOnBase64 = `H4sIAAAAAAACA+sM8HPn5ZLiYmBg4PX0cAkC0gIgzMgMJFVtc5WAlEKyR5AvA0OVGgNDQwsDwy+gUMMLBoZSAwaGVwkMDFYzGBjEC+bsCrQBSrAF+IS4/mdg+P//v6OsiSBQhDHJ292F8T+T7j0gh73E09eV/SWHqCiTwRIte2OgEI+ni2MIx/XkBAVeIM+AgfH4qtY+kOUlrhElKYklqVbJRalAisHIwMhU18BC18gwxMjQysDIysRC28DAysBgtwdDJoqG3PyUzLRK3BpOie64CtSgAddQkpmbWlySmFuAW89chpmgQGLwdPVzWeeU0AQA5nQkVjkBAAA=`
