package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SEC struct
type SecPriceMovement struct {
	Id         int       `gorm:"primary_key" json:"id"`
	TokenPrice float64   `gorm:"token_price" json:"token_price"`
	BLatest    int       `gorm:"b_latest" json:"b_latest"`
	CreatedBy  string    `gorm:"created_by" json:"created_by"`
	CreatedAt  time.Time `gorm:"created_at" json:"created_at"`
	UpdatedBy  string    `gorm:"updated_by" json:"updated_at"`
	UpdatedAt  time.Time `gorm:"updated_at" json:"updated_at"`
}

// GetLatestSecPriceMovement
func GetLatestSecPriceMovement() (float64, error) {
	var SecPriceMovement SecPriceMovement
	err := db.Table("sec_price_movement").
		Where("b_latest = 1").First(&SecPriceMovement).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return SecPriceMovement.TokenPrice, nil
}

// GetSecPriceMovementFn get sec_price_movement data with dynamic condition
func GetSecPriceMovementFn(arrCond []WhereCondFn, limit int, debug bool) ([]*SecPriceMovement, error) {
	var result []*SecPriceMovement
	tx := db.Table("sec_price_movement").
		Order("sec_price_movement.created_at DESC")
		// Order("sec_price_movement.created_at ASC")

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

// Sec struct
type MinMaxSecPriceMovement struct {
	MinTokenPrice float64 `gorm:"column:min_token_price" json:"min_token_price"`
	MaxTokenPrice float64 `gorm:"column:max_token_price" json:"max_token_price"`
}

// GetMinMaxSecPriceMovementFn get sec_price_movement min & max data with dynamic condition
func GetMinMaxSecPriceMovementFn(arrCond []WhereCondFn, limit int, debug bool) (*MinMaxSecPriceMovement, error) {
	var result MinMaxSecPriceMovement
	tx := db.Table("sec_price_movement").
		Select("MIN(token_price) AS min_token_price, MAX(token_price) AS max_token_price")

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

	return &result, nil
}
