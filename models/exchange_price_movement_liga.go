package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// ExchangePriceMovementLiga struct
type ExchangePriceMovementLiga struct {
	Id         int     `gorm:"primary_key" json:"id"`
	TokenPrice float64 `gorm:"column:token_price" json:"token_price"`
}

// GetLatestExchangePriceMovementLiga func
func GetLatestExchangePriceMovementLiga() (float64, error) {
	var ExchangePriceMovementLiga ExchangePriceMovementLiga
	err := db.Table("exchange_price_movement_liga").
		Order("id desc").
		First(&ExchangePriceMovementLiga).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return ExchangePriceMovementLiga.TokenPrice, nil
}

// AddExchangePriceMovementLigaStruct struct
type AddExchangePriceMovementLigaStruct struct {
	ID         int     `gorm:"primary_key" json:"id"`
	TokenPrice float64 `gorm:"column:token_price" json:"token_price"`
	CreatedBy  int     `gorm:"column:created_by" json:"created_by"`
}

// func AddExchangePriceMovementLigaTx add exchange_price_liga records
func AddExchangePriceMovementLigaTx(tx *gorm.DB, arrData AddExchangePriceMovementLigaStruct) (*AddExchangePriceMovementLigaStruct, error) {
	if err := tx.Table("exchange_price_movement_liga").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// func AddExchangePriceMovementLiga add exchange_price_liga records
func AddExchangePriceMovementLiga(arrData AddExchangePriceMovementLigaStruct) (*AddExchangePriceMovementLigaStruct, error) {
	if err := db.Table("exchange_price_movement_liga").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// GetExchangePriceMovementLigaFn get exchange_price_movement_liga data with dynamic condition
func GetExchangePriceMovementLigaFn(arrCond []WhereCondFn, limit int, debug bool) ([]*ExchangePriceMovementLiga, error) {
	var result []*ExchangePriceMovementLiga
	tx := db.Table("exchange_price_movement_liga").
		Order("exchange_price_movement_liga.created_at DESC")
		// Order("exchange_price_movement_liga.created_at ASC")

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
