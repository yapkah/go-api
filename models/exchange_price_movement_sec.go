package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// ExchangePriceMovementSec struct
type ExchangePriceMovementSec struct {
	Id         int     `gorm:"primary_key" json:"id"`
	TokenPrice float64 `gorm:"column:token_price" json:"token_price"`
}

// GetLatestExchangePriceMovementSec func
func GetLatestExchangePriceMovementSec() (float64, error) {
	var ExchangePriceMovementSec ExchangePriceMovementSec
	err := db.Table("exchange_price_movement_sec").
		Order("id desc").
		First(&ExchangePriceMovementSec).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return ExchangePriceMovementSec.TokenPrice, nil
}

// AddExchangePriceMovementSecStruct struct
type AddExchangePriceMovementSecStruct struct {
	ID         int     `gorm:"primary_key" json:"id"`
	TokenPrice float64 `gorm:"column:token_price" json:"token_price"`
	CreatedBy  int     `gorm:"column:created_by" json:"created_by"`
}

// func AddExchangePriceMovementSecTx add exchange_price_liga records
func AddExchangePriceMovementSecTx(tx *gorm.DB, arrData AddExchangePriceMovementSecStruct) (*AddExchangePriceMovementSecStruct, error) {
	if err := tx.Table("exchange_price_movement_sec").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// func AddExchangePriceMovementSecTx add exchange_price_liga records
func AddExchangePriceMovementSec(arrData AddExchangePriceMovementSecStruct) (*AddExchangePriceMovementSecStruct, error) {
	if err := db.Table("exchange_price_movement_sec").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// GetExchangePriceMovementSecFn get exchange_price_movement_sec data with dynamic condition
func GetExchangePriceMovementSecFn(arrCond []WhereCondFn, limit int, debug bool) ([]*ExchangePriceMovementSec, error) {
	var result []*ExchangePriceMovementSec
	tx := db.Table("exchange_price_movement_sec").
		Order("exchange_price_movement_sec.created_at DESC")
		// Order("exchange_price_movement_sec.created_at ASC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			if v.CondValue != nil {
				tx = tx.Where(v.Condition, v.CondValue)
			} else {
				tx = tx.Where(v.Condition)
			}
		}
	}
	if debug {
		tx = tx.Debug()
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
