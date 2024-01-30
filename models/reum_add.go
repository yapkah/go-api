package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// ReumAddStruct struct
type ReumAddStruct struct {
	ID                int    `gorm:"primary_key" json:"id"`
	CryptoAddr        string `json:"crypto_addr"`
	CryptoEncryptAddr string `json:"crypto_encrypt_addr"`
	PrivateKey        string `json:"private_key"`
	MemberID          string `json:"member_id"`
	CryptoType        string `json:"crypto_type"`
	WalletTypeId      string `json:"wallet_type_id"`
}

func GetDepositWalletAddress() (*ReumAddStruct, error) {
	var crypto ReumAddStruct

	query := db.Table("reum_add a")

	err := query.Find(&crypto).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &crypto, nil
}

// AddReumAddr add reum addr records
func AddReumAddr(tx *gorm.DB, arrData ReumAddStruct) (*ReumAddStruct, error) {
	if err := tx.Table("reum_add").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// GetReumAddrFn get reum_add with dynamic condition
func GetReumAddFn(arrCond []WhereCondFn, debug bool) ([]*ReumAddStruct, error) {
	var result []*ReumAddStruct
	tx := db.Table("reum_add")

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

type CompletedAddStruct struct {
	ID              int    `json:"id" gorm:"column:id"`
	Part1Addr       string `json:"part_1_addr" gorm:"column:part_1_addr"`
	Part1Encryptedd string `json:"part_1_encrypted_addr" gorm:"column:part_1_encrypted_addr"`
	Part2Addr       string `json:"part_2_addr" gorm:"column:part_2_addr"`
	Part2Encryptedd string `json:"part_2_encrypted_addr" gorm:"column:part_2_encrypted_addr"`
}

// GetCompletedAddFn get reum_add with dynamic condition
func GetCompletedAddFn(arrCond []WhereCondFn, debug bool) ([]*CompletedAddStruct, error) {
	var result []*CompletedAddStruct
	tx := db.Table("reum_add").
		Joins("INNER JOIN sc_hash ON reum_add.id = sc_hash.sc_id").
		Select("reum_add.id, crypto_addr AS 'part_1_addr', crypto_encrypt_addr AS 'part_1_encrypted_addr', " +
			"sc_part AS 'part_2_addr', sc_encrypt AS 'part_2_encrypted_addr'")

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
