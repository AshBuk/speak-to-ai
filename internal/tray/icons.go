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
const iconMicOffBase64 = `
H4sIAAAAAAAA/3SPweqbMBCF9zzF4P1/c8HgRgSvVIRuuvHFE39GbGuUkdpe+vRFTUIJhZ6BmTPnzAzr
mmqIMLDyxIc5pTX0eCn0CnpK+qj5X+8UPQx8gS+pEvR41PgClQpK1YEL3lqFlmvU7RRCQKdbgUbV9vht
Fa2Ij6p9ROiHp5Hn/q7zFCf0uY7jaE40/XUQsEo3pIVOEUqoQlLdOeE1R8+CL+xPYvv9tJ+uLVJAH0XE
IVfhbFcO4qTZk5M23qXhVDaWLRs4PH0tDKPEv5TXh0NpY1PSzuU1WrDUGYtFt90IXjZYfrIbQ27Bl0KX
oD9Sjy1rNTYNj22rYJcnjtdpOdCbhXbJlW3uMSQx0l4uM9RNd1KwVi1j+YGJpnVa4kNXGFM2aMZZGiYp
l9mAb6mW/TK/AgAA//9MnxM3jgEAAA==
`

// Microphone on icon (red)
const iconMicOnBase64 = `
H4sIAAAAAAAA/3SOweraQBCG9zzF4PmvveBPJCJXWkI33MQXTzRZsdkaZWSbhjx9MU0ohW56YL7hm38G
pgnVMFqzcHSqD2lKa7RoS/kKLSV1LPJ1J8FgUK95TqVAi8csvqYlclC8VsG8MmjKNch2DiGyLi8ZjODs
+MsJ3hG9KfaI2HfPI89tX+taPqGNdRhGfaLpr4eIRXRHXvKxVDElh2a/HNfkWVVqFfQnse1+2k/XFsmj
JRFxyCM4W8tB7BV9dMpEl4ZThVk1GRg8PJJLRolP22Q6jUqKTUq7PK3RgqVUUm1e7zRfNpg/dLBu3YK9
yExm+JZ6bJnRWDccN0ZwLbpwvE7rgTZLtUupbHPvIYmROsvQQDlpG5U19mPzfp6muO+h3ykpzZRX+wn9
NKtmmj8BAAD//9qlc2+OAQAA
`
