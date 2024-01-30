package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"golang.org/x/crypto/scrypt"
)

type PHCryReturnStruct struct {
	AddrInfo1 struct {
		Addr     string
		Scrypted string
	}
	AddrInfo2 struct {
		Addr     string
		Scrypted string
	}
}

// const (
// 	cpuCost         = 16384 // int CPU/memory cost parameter (logN)
// 	blockSize       = 8     // int block size parameter (octets)
// 	parallelisation = 1     // int parallelisation parameter (positive int)
// 	derivedKey      = 32    // int length of the derived key (octets)
// 	charSet         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
// )

// PHCry
func PHCry(value string) *PHCryReturnStruct {

	var (
		part1 string
		part2 string
	)
	wordCount := utf8.RuneCountInString(value)
	halfValue := wordCount / 2
	remainder := wordCount % 2

	if remainder > 0 {
		part1 = string(value[0 : halfValue+1])
		part2 = string(value[halfValue+1 : wordCount])
	} else {
		part1 = string(value[0:halfValue])
		part2 = string(value[halfValue:wordCount])
	}
	charSet := setting.Cfg.Section("custom").Key("CharSet").String()

	part1Len := utf8.RuneCountInString(part1)
	part1RString := base.GenerateRandomString(part1Len, charSet)

	part2Len := utf8.RuneCountInString(part2)
	part2RString := base.GenerateRandomString(part2Len, charSet)

	add1 := part2RString + part1
	add2 := part1RString + part2

	add1Byte := []byte(add1)
	cryptoSalt2 := setting.Cfg.Section("custom").Key("CryptoSalt2").String()
	generatedScryptedAdd1Byte, _ := GenerateScryptValue(add1Byte, cryptoSalt2)
	generatedScryptedAdd1String := string(generatedScryptedAdd1Byte)

	add2Byte := []byte(add2)
	cryptoSalt3 := setting.Cfg.Section("custom").Key("CryptoSalt3").String()
	generatedScryptedAdd2Byte, _ := GenerateScryptValue(add2Byte, cryptoSalt3)
	generatedScryptedAdd2String := string(generatedScryptedAdd2Byte)

	arrDataReturn := PHCryReturnStruct{}
	arrDataReturn.AddrInfo1.Addr = add1
	arrDataReturn.AddrInfo1.Scrypted = generatedScryptedAdd1String
	arrDataReturn.AddrInfo2.Addr = add2
	arrDataReturn.AddrInfo2.Scrypted = generatedScryptedAdd2String

	return &arrDataReturn

}

// func GenerateScrypt. Can refer https://github.com/elithrar/simple-scrypt/blob/master/scrypt.go
func GenerateScryptValue(value []byte, salt string) ([]byte, error) {

	// cryptoSalt := setting.Cfg.Section("custom").Key("CryptoSalt").String()
	cryptoSaltByte := []byte(salt)

	cpuCost, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("CPUCost").String())
	blockSize, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("BlockSize").String())
	parallelisation, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("Parallelisation").String())
	derivedKey, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("DerivedKey").String())
	// scrypt.Key returns the raw scrypt derived key.
	dk, err := scrypt.Key(value, cryptoSaltByte, cpuCost, blockSize, parallelisation, derivedKey)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("%x", dk)), nil
}

func GenerateHmacSHA256(secret, data, result string) string {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))

	// Write Data to it
	h.Write([]byte(data))

	sha := ""

	if result == "base64" {
		// Get result and encode as hexadecimal string
		sha = base64.StdEncoding.EncodeToString(h.Sum(nil))
	} else {
		// Get result and encode as hexadecimal string
		sha = hex.EncodeToString(h.Sum(nil))
	}

	return sha
}
