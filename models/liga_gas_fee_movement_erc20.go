package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// LigaGasFeeErc20Movement struct
type LigaGasFeeErc20Movement struct {
	ID         int       `gorm:"primary_key" json:"id"`
	TokenPrice float64   `gorm:"column:token_price" json:"token_price"`
	BLatest    int       `gorm:"column:b_latest" json:"b_latest"`
	CreatedBy  int       `gorm:"column:created_by" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedBy  int       `gorm:"column:updated_by" json:"updated_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func GetLatestLigaGasFeeMovementErc20() (float64, error) {
	var LigaGasFeeErc20Movement LigaGasFeeErc20Movement
	err := db.Table("liga_gas_fee_movement_erc20").
		Where("b_latest = 1").First(&LigaGasFeeErc20Movement).Error

	if err != nil {
		return float64(20), &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return LigaGasFeeErc20Movement.TokenPrice, nil
}

// GetLigaPriceMovementFn get liga_price_movement data with dynamic condition
func GetLigaGasFeeMovementErc20Fn(arrCond []WhereCondFn, limit int, debug bool) ([]*LigaGasFeeErc20Movement, error) {
	var result []*LigaGasFeeErc20Movement
	tx := db.Table("liga_gas_fee_movement_erc20").
		Order("liga_gas_fee_movement_erc20.created_at DESC")

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
