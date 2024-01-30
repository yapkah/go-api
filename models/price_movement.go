package models

import (
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// priceMovement struct
type PriceMovement struct {
	Id         int       `gorm:"primary_key" json:"id"`
	TokenPrice float64   `gorm:"token_price" json:"token_price"`
	BLatest    int       `gorm:"b_latest" json:"b_latest"`
	CreatedBy  string    `gorm:"created_by" json:"created_by"`
	CreatedAt  time.Time `gorm:"created_at" json:"created_at"`
	UpdatedBy  string    `gorm:"updated_by" json:"updated_at"`
	UpdatedAt  time.Time `gorm:"updated_at" json:"updated_at"`
}

// GetLatestPriceMovementByTokenType
func GetLatestPriceMovementByTokenType(TokenType string) (float64, error) {
	var PriceMovement PriceMovement
	TableName := strings.ToLower(TokenType) + "_price_movement"

	err := db.Table(TableName).
		Where("b_latest = 1").First(&PriceMovement).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return PriceMovement.TokenPrice, nil
}

// GetPriceMovementFn get price_movement data with dynamic condition
func GetPriceMovementByTokenTypeFn(TokenType string, arrCond []WhereCondFn, limit int, debug bool) ([]*PriceMovement, error) {
	var result []*PriceMovement
	TableName := strings.ToLower(TokenType) + "_price_movement"

	tx := db.Table(TableName).
		Order("created_at DESC")

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
