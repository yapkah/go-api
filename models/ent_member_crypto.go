package models

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
	"golang.org/x/crypto/scrypt"
)

// AddEntMemberCryptoStruct struct
type AddEntMemberCryptoStruct struct {
	ID                int       `gorm:"primary_key" json:"id"`
	MemberID          int       `json:"member_id"`
	CryptoType        string    `json:"crypto_type"`
	CryptoAddress     string    `json:"crypto_address"`
	CryptoEncryptAddr string    `json:"crypto_encrypt_addr"`
	PrivateKey        string    `json:"-" gorm:"column:private_key"`
	Mnemonic          string    `json:"-" gorm:"column:mn"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         int       `json:"created_by"`
}

// AddEntMemberCrypto add member
func AddEntMemberCrypto(tx *gorm.DB, entMemberCrypto AddEntMemberCryptoStruct) (*AddEntMemberCryptoStruct, error) {
	if err := tx.Table("ent_member_crypto").Create(&entMemberCrypto).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &entMemberCrypto, nil
}

// EntMemberCrypto struct
type EntMemberCrypto struct {
	ID                int       `gorm:"primary_key" json:"id"`
	MemberID          int       `json:"member_id"`
	CryptoType        string    `json:"crypto_type"`
	CryptoAddress     string    `json:"crypto_address"`
	CryptoEncryptAddr string    `json:"crypto_encrypt_addr"`
	PrivateKey        string    `json:"-" gorm:"column:private_key"`
	Mnemonic          string    `json:"-" gorm:"column:mn"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         string    `json:"updated_by"`
}

// GetEntMemberCryptoFn get ent_member_crypto with dynamic condition
func GetEntMemberCryptoFn(arrCond []WhereCondFn, debug bool) (*EntMemberCrypto, error) {
	var entMemberCrypto EntMemberCrypto
	tx := db.Table("ent_member_crypto")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&entMemberCrypto).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if entMemberCrypto.ID <= 0 {
		return nil, nil
	}

	return &entMemberCrypto, nil
}

func GetMemberCryptoByMemID(Id int, WalletType string) (*EntMemberCrypto, error) {
	var crypto EntMemberCrypto

	query := db.Where("member_id = ? AND status= 'A'", Id)

	if WalletType != "" {
		query = query.Where("crypto_type = ?", WalletType)
	}

	err := query.Find(&crypto).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &crypto, nil
}

// func GetCustomMemberCryptoAddr using member_id and crypto type
func GetCustomMemberCryptoAddr(entMemberID int, cryptoType string, checkAddrHash bool, debug bool) (cryptoAddr string, err error) {

	if entMemberID > 0 {
		arrCond := make([]WhereCondFn, 0)
		arrCond = append(arrCond,
			WhereCondFn{Condition: "ent_member.id = ?", CondValue: entMemberID},
		)
		arrEntMem, _ := GetEntMemberFn(arrCond, "", false)
		if arrEntMem == nil {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member_id"}
		}
	}

	arrCond := make([]WhereCondFn, 0)
	arrCond = append(arrCond,
		WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: entMemberID},
		WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)
	if cryptoType != "" {
		if strings.ToLower(cryptoType) == "usdt_erc20" || strings.ToLower(cryptoType) == "usdc_erc20" {
			cryptoType = "ETH"
		}
		if strings.ToLower(cryptoType) == "usdc" {
			cryptoType = "USDT"
		}
		if strings.ToLower(cryptoType) == "bep" {
			cryptoType = "BUSD"
		}
		// if strings.ToLower(cryptoType) == "usdt" || strings.ToLower(cryptoType) == "eth" {
		// 	cryptoType = "ETH"
		// }
		arrCond = append(arrCond,
			WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: cryptoType},
		)
	}
	arrExistingMemCrypto, _ := GetEntMemberCryptoFn(arrCond, debug)

	if arrExistingMemCrypto == nil {
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address"}
	}

	if checkAddrHash {

		// start checking for the main encrypt addr
		addrByte := []byte(arrExistingMemCrypto.CryptoAddress)
		cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt1").String()
		err := CompareHashAndScryptedValue(arrExistingMemCrypto.CryptoEncryptAddr, addrByte, cryptoSalt1)
		if err != nil {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_1"}
		}
		// end checking for the main encrypt addr

		// start checking for the sub
		arrCond := make([]WhereCondFn, 0)
		arrCond = append(arrCond,
			WhereCondFn{Condition: "reum_add.member_id = ?", CondValue: entMemberID},
		)
		if cryptoType != "" {
			arrCond = append(arrCond,
				WhereCondFn{Condition: "reum_add.crypto_type = ?", CondValue: cryptoType},
			)
		}

		arrExistingCompletedCryptoAdd, _ := GetCompletedAddFn(arrCond, debug)

		if len(arrExistingCompletedCryptoAdd) < 1 {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_2"}
		}

		wordCount := utf8.RuneCountInString(arrExistingMemCrypto.CryptoAddress)
		halfValue := wordCount / 2
		remainder := wordCount % 2

		part1 := string(arrExistingMemCrypto.CryptoAddress[0:halfValue])
		part2 := string(arrExistingMemCrypto.CryptoAddress[halfValue:wordCount])

		if remainder > 0 {
			part1 = string(arrExistingMemCrypto.CryptoAddress[0 : halfValue+1])
			part2 = string(arrExistingMemCrypto.CryptoAddress[halfValue+1 : wordCount])
		}

		var part1Status bool
		var part2Status bool
		if strings.Contains(arrExistingMemCrypto.CryptoAddress, part1) {
			part1Status = true
		}

		if !part1Status {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_3"}
		}

		if strings.Contains(arrExistingMemCrypto.CryptoAddress, part2) {
			part2Status = true
		}

		if !part2Status {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_4"}
		}

		// start checking for the sub encrypt addr1
		addr2Byte := []byte(arrExistingCompletedCryptoAdd[0].Part1Addr)
		cryptoSalt2 := setting.Cfg.Section("custom").Key("CryptoSalt2").String()
		err = CompareHashAndScryptedValue(arrExistingCompletedCryptoAdd[0].Part1Encryptedd, addr2Byte, cryptoSalt2)
		if err != nil {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_5"}
		}
		// end checking for the sub encrypt addr1

		// start checking for the sub encrypt addr2
		addr3Byte := []byte(arrExistingCompletedCryptoAdd[0].Part2Addr)
		cryptoSalt3 := setting.Cfg.Section("custom").Key("CryptoSalt3").String()
		err = CompareHashAndScryptedValue(arrExistingCompletedCryptoAdd[0].Part2Encryptedd, addr3Byte, cryptoSalt3)
		if err != nil {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_6"}
		}
		// end checking for the sub encrypt addr2
		// end checking for the sub
	}

	return arrExistingMemCrypto.CryptoAddress, nil
}

// func CompareHashAndScryptedValue. The comparison performed by this function is constant-time. It returns nil on success, and an error if the derived keys do not match. Can refer https://github.com/elithrar/simple-scrypt/blob/master/scrypt.go
func CompareHashAndScryptedValue(hash string, password []byte, salt string) error {
	// Decode existing hash, retrieve params and salt.
	dk, err := hex.DecodeString(hash)
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	// cryptoSalt := setting.Cfg.Section("custom").Key("CryptoSalt").String()
	cryptoSaltByte := []byte(salt)

	// scrypt the cleartext password with the same parameters and salt
	cpuCost, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("CPUCost").String())
	blockSize, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("BlockSize").String())
	parallelisation, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("Parallelisation").String())
	derivedKey, _ := strconv.Atoi(setting.Cfg.Section("custom").Key("DerivedKey").String())
	other, err := scrypt.Key(password, cryptoSaltByte, cpuCost, blockSize, parallelisation, derivedKey)
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	// Constant time comparison
	if subtle.ConstantTimeCompare(dk, other) == 1 {
		return nil
	}

	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "scrypt: the hashed password does not match the hash of the given password"}
}

// GetEntMemberCryptoListFn get ent_member_crypto with dynamic condition
func GetEntMemberCryptoListFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberCrypto, error) {
	var entMemberCrypto []*EntMemberCrypto
	tx := db.Table("ent_member_crypto")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&entMemberCrypto).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return entMemberCrypto, nil
}

// TaggedMemberCryptoAddr struct
type TaggedMemberCryptoAddr struct {
	MemberID          int    `json:"ent_member_id" gorm:"column:ent_member_id"`
	NickName          string `json:"nick_name" gorm:"column:nick_name"`
	CryptoAddress     string `json:"crypto_address" gorm:"column:crypto_address"`
	CryptoEncryptAddr string `json:"crypto_encrypt_addr" gorm:"column:crypto_encrypt_addr"`
	TaggedMemberID    int    `json:"tagged_member_id" gorm:"column:tagged_member_id"`
	// TaggedCryptoAddress     string `json:"tagged_crypto_address" gorm:"column:tagged_crypto_address"`
	// TaggedCryptoEncryptAddr string `json:"tagged_crypto_encrypt_addr" gorm:"column:tagged_crypto_encrypt_addr"`
}

func GetTaggedMemberCryptoAddrFn(arrCond []WhereCondFn, debug bool) ([]*TaggedMemberCryptoAddr, error) {
	var result []*TaggedMemberCryptoAddr
	tx := db.Table("ent_member_crypto AS member_crypto").
		Joins("INNER JOIN ent_member ON member_crypto.member_id = ent_member.id AND ent_member.tagged_member_id > 0").
		// Joins("INNER JOIN ent_member_crypto AS tagged_member_crypto ON ent_member.tagged_member_id = tagged_member_crypto.member_id").
		// Select("ent_member.id AS ent_member_id, ent_member.nick_name, member_crypto.crypto_address, member_crypto.crypto_encrypt_addr, ent_member.tagged_member_id, tagged_member_crypto.crypto_address AS tagged_crypto_address, tagged_member_crypto.crypto_encrypt_addr AS tagged_crypto_encrypt_addr")
		Select("ent_member.id AS ent_member_id, ent_member.nick_name, member_crypto.crypto_address, member_crypto.crypto_encrypt_addr, ent_member.tagged_member_id")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type CustomMemberCryptoInfov2Rst struct {
	CryptoAddr string
	PrivateKey string
}

// func GetCustomMemberCryptoInfov2 using member_id and crypto type
func GetCustomMemberCryptoInfov2(entMemberID int, cryptoType string, counterChecking bool, debug bool) (*CustomMemberCryptoInfov2Rst, error) {
	arrCond := make([]WhereCondFn, 0)
	arrCond = append(arrCond,
		WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: entMemberID},
		WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)
	if cryptoType != "" {
		if strings.ToLower(cryptoType) == "usdt_erc20" {
			cryptoType = "ETH"
		}
		// if strings.ToLower(cryptoType) == "usdt" || strings.ToLower(cryptoType) == "eth" {
		// 	cryptoType = "ETH"
		// }
		if strings.ToLower(cryptoType) == "liga" || strings.ToLower(cryptoType) == "sec" || strings.ToLower(cryptoType) == "usds" {
			cryptoType = "SEC"
		}
		arrCond = append(arrCond,
			WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: cryptoType},
		)
	}
	arrExistingMemCrypto, _ := GetEntMemberCryptoFn(arrCond, debug)

	if arrExistingMemCrypto == nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address"}
	}

	if counterChecking {

		// start checking for the main encrypt addr
		addrByte := []byte(arrExistingMemCrypto.CryptoAddress)
		cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt1").String()
		err := CompareHashAndScryptedValue(arrExistingMemCrypto.CryptoEncryptAddr, addrByte, cryptoSalt1)
		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_1"}
		}
		// end checking for the main encrypt addr

		// start checking for the sub
		arrCond := make([]WhereCondFn, 0)
		arrCond = append(arrCond,
			WhereCondFn{Condition: "reum_add.member_id = ?", CondValue: entMemberID},
		)
		if cryptoType != "" {
			arrCond = append(arrCond,
				WhereCondFn{Condition: "reum_add.crypto_type = ?", CondValue: cryptoType},
			)
		}

		arrExistingCompletedCryptoAdd, _ := GetCompletedAddFn(arrCond, debug)

		if len(arrExistingCompletedCryptoAdd) < 1 {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_2"}
		}

		wordCount := utf8.RuneCountInString(arrExistingMemCrypto.CryptoAddress)
		halfValue := wordCount / 2
		remainder := wordCount % 2

		part1 := string(arrExistingMemCrypto.CryptoAddress[0:halfValue])
		part2 := string(arrExistingMemCrypto.CryptoAddress[halfValue:wordCount])

		if remainder > 0 {
			part1 = string(arrExistingMemCrypto.CryptoAddress[0 : halfValue+1])
			part2 = string(arrExistingMemCrypto.CryptoAddress[halfValue+1 : wordCount])
		}

		var part1Status bool
		var part2Status bool
		if strings.Contains(arrExistingMemCrypto.CryptoAddress, part1) {
			part1Status = true
		}

		if !part1Status {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_3"}
		}

		if strings.Contains(arrExistingMemCrypto.CryptoAddress, part2) {
			part2Status = true
		}

		if !part2Status {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_4"}
		}

		// start checking for the sub encrypt addr1
		addr2Byte := []byte(arrExistingCompletedCryptoAdd[0].Part1Addr)
		cryptoSalt2 := setting.Cfg.Section("custom").Key("CryptoSalt2").String()
		err = CompareHashAndScryptedValue(arrExistingCompletedCryptoAdd[0].Part1Encryptedd, addr2Byte, cryptoSalt2)
		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_5"}
		}
		// end checking for the sub encrypt addr1

		// start checking for the sub encrypt addr2
		addr3Byte := []byte(arrExistingCompletedCryptoAdd[0].Part2Addr)
		cryptoSalt3 := setting.Cfg.Section("custom").Key("CryptoSalt3").String()
		err = CompareHashAndScryptedValue(arrExistingCompletedCryptoAdd[0].Part2Encryptedd, addr3Byte, cryptoSalt3)
		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address.hacker_is_suspected_6"}
		}
		// end checking for the sub encrypt addr2
		// end checking for the sub
	}

	privateKey, err := GetCustomMemberPKInfo(entMemberID, cryptoType, counterChecking, debug)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
	}

	arrDataReturn := CustomMemberCryptoInfov2Rst{
		CryptoAddr: arrExistingMemCrypto.CryptoAddress,
		PrivateKey: privateKey,
	}

	return &arrDataReturn, nil
}

func GetCustomMemberPKInfo(entMemberID int, cryptoType string, counterChecking bool, debug bool) (privateKey string, err error) {
	arrCond := make([]WhereCondFn, 0)
	arrCond = append(arrCond,
		WhereCondFn{Condition: " ent_member.id = ?", CondValue: entMemberID},
		WhereCondFn{Condition: " ent_member.status = ?", CondValue: "A"},
	)
	entMember, _ := GetEntMemberFn(arrCond, "", debug)

	if entMember == nil {
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member"}
	}

	arrCond = make([]WhereCondFn, 0)
	arrCond = append(arrCond,
		WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: entMemberID},
		WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)

	if cryptoType != "" {
		if strings.ToLower(cryptoType) == "usdt_erc20" {
			cryptoType = "ETH"
		}
		// if strings.ToLower(cryptoType) == "usdt" || strings.ToLower(cryptoType) == "eth" {
		// 	cryptoType = "ETH"
		// }
		if strings.ToLower(cryptoType) == "liga" || strings.ToLower(cryptoType) == "sec" || strings.ToLower(cryptoType) == "usds" {
			cryptoType = "SEC"
		}
		arrCond = append(arrCond,
			WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: cryptoType},
		)
	}
	arrExistingMemCrypto, _ := GetEntMemberCryptoFn(arrCond, debug)

	if arrExistingMemCrypto == nil {
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "member_no_related_crypto_address"}
	}

	if counterChecking {
		pkSalt := setting.Cfg.Section("custom").Key("PKSalt").String()

		key := arrExistingMemCrypto.PrivateKey + pkSalt
		keyByte := []byte(key)
		hasher := sha256.New()
		hasher.Write(keyByte)
		sha256Value := hex.EncodeToString(hasher.Sum(nil))

		if debug {
			fmt.Println("pkSalt:", pkSalt)
			fmt.Println("key:", key)
			fmt.Println("DPK:", entMember.DPK)
			fmt.Println("sha256Value:", sha256Value)
		}

		if entMember.DPK != sha256Value {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member_private_key_hacker_suspected_1"}
		}
	}

	return arrExistingMemCrypto.PrivateKey, nil
}
