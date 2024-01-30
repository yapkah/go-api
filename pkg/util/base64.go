package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncodeBase64 base64 encryption
// example output: ZmVmZHNjZmZmZg==
func EncodeBase64(value string) string {
	// return base64.URLEncoding.EncodeToString([]byte(value))
	return base64.StdEncoding.EncodeToString([]byte(value))
}

// DecodeBase64 base64 encryption
func DecodeBase64(value string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var key = []byte("fj6rsy8ydsh53fji") // 16 bytes

// Encrypt function
// example output: 31595d35a8f1fe9a3a10712b725e37d370e66ab824f3effe71c03935db68127f
func Encrypt(value string) (string, error) {
	text := []byte(value)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	b := base64.StdEncoding.EncodeToString(text)
	cipherText := make([]byte, aes.BlockSize+len(b))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(b))

	return string(cipherText), nil
}

// Decrypt function
func Decrypt(value string) (string, error) {
	text := []byte(value)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(text) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	decryptedText, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return "", err
	}
	return string(decryptedText), nil
}
