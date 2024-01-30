package util

import "encoding/base32"

// EncodeBase32 base64 encryption
func EncodeBase32(value string) string {
	return base32.StdEncoding.EncodeToString([]byte(value))
}
