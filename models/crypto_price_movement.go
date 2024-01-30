package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// CryptoPriceMovement struct
type CryptoPriceMovement struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Code      string    `json:"code"`
	Price     float64   `json:"price"`
	BLatest   int       `json:"b_latest"`
	CreatedAt time.Time `json:"created_at"`
}

// GetCryptoPriceMovementFn
func GetCryptoPriceMovementFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*CryptoPriceMovement, error) {
	var result []*CryptoPriceMovement
	tx := db.Table("crypto_price_movement").
		Order("crypto_price_movement.seq_no asc").
		Order("crypto_price_movement.created_at DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	//testing
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetCurrentCryptoPriceMovementFn
func GetCurrentCryptoPriceMovementFn(code, date string, debug bool) (*CryptoPriceMovement, error) {
	var result CryptoPriceMovement

	tx := db.Table("crypto_price_movement").
		Where("date(crypto_price_movement.created_at) LIKE (select date(max(created_at)) FROM crypto_price_movement WHERE `code` = ? AND date(`created_at`) < ?)", code, date).
		Where("crypto_price_movement.code LIKE ?", code).
		// Where("crypto_price_movement.b_latest = ?", 1).
		Order("crypto_price_movement.price DESC").
		Limit(1)

	if debug {
		tx = tx.Debug()
	}

	err := tx.First(&result).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

// AddCryptoPriceMovementStruct struct
type AddCryptoPriceMovementStruct struct {
	Code    string  `json:"code"`
	Price   float64 `json:"price"`
	BLatest int     `json:"b_latest"`
}

// AddEntMemberCrypto add member
func AddCryptoPriceMovement(tx *gorm.DB, params AddCryptoPriceMovementStruct) (*AddCryptoPriceMovementStruct, error) {
	if err := tx.Table("crypto_price_movement").Create(&params).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &params, nil
}

// // EntMemberCrypto struct
// type EntMemberCrypto struct {
// 	ID                int       `gorm:"primary_key" json:"id"`
// 	MemberID          int       `json:"member_id"`
// 	CryptoType        string    `json:"crypto_type"`
// 	CryptoAddress     string    `json:"crypto_address"`
// 	CryptoEncryptAddr string    `json:"crypto_encrypt_addr"`
// 	PrivateKey        string    `json:"private_key"`
// 	Status            string    `json:"status"`
// 	CreatedAt         time.Time `json:"created_at"`
// 	CreatedBy         string    `json:"created_by"`
// 	UpdatedAt         time.Time `json:"updated_at"`
// 	UpdatedBy         string    `json:"updated_by"`
// }

// // GetEntMemberCryptoFn get ent_member_crypto with dynamic condition
// func GetEntMemberCryptoFn(arrCond []WhereCondFn, debug bool) (*EntMemberCrypto, error) {
// 	var entMemberCrypto EntMemberCrypto
// 	tx := db.Table("ent_member_crypto")

// 	if len(arrCond) > 0 {
// 		for _, v := range arrCond {
// 			tx = tx.Where(v.Condition, v.CondValue)
// 		}
// 	}
// 	if debug {
// 		tx = tx.Debug()
// 	}
// 	err := tx.Find(&entMemberCrypto).Error

// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}

// 	if entMemberCrypto.ID <= 0 {
// 		return nil, nil
// 	}

// 	return &entMemberCrypto, nil
// }

// func GetMemberCryptoByMemID(Id int, WalletType string) (*EntMemberCrypto, error) {
// 	var crypto EntMemberCrypto

// 	query := db.Where("member_id = ? AND status= 'A'", Id)

// 	if WalletType != "" {
// 		query = query.Where("crypto_type = ?", WalletType)
// 	}

// 	err := query.Find(&crypto).Error

// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}

// 	if err == gorm.ErrRecordNotFound {
// 		return nil, nil
// 	}

// 	return &crypto, nil
// }

// GetCurrentCryptoPriceMovementFn
func GetLatestCryptoPriceMovementFn(code string, debug bool) (*CryptoPriceMovement, error) {
	var result CryptoPriceMovement

	tx := db.Table("crypto_price_movement").
		Where("crypto_price_movement.code LIKE ?", code).
		Where("crypto_price_movement.b_latest = ?", 1).
		Limit(1)

	if debug {
		tx = tx.Debug()
	}

	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
