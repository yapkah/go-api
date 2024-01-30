package util

import (
	"bytes"
	"encoding/ascii85"
)

// EncodeAscii85 ascii85 encode
func EncodeAscii85(value string) string {
	str := []byte(value)

	buffer := make([]byte, ascii85.MaxEncodedLen(len(str)))
	ascii85.Encode(buffer, str)

	return string(buffer)
}

// DecodeAscii85 ascii85 decode
func DecodeAscii85(value string) (string, error) {
	buffer := make([]byte, len(value))

	_, _, err := ascii85.Decode(buffer, []byte(value), true)
	if err != nil {
		return "", err
	}

	buffer = bytes.Trim(buffer, "\x00") // remove null bytes from output byte array
	return string(buffer), nil
}
