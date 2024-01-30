package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// GasFeeTron movement struct
type GasFeeTronMovement struct {
	ID         int       `gorm:"primary_key" json:"id"`
	TokenPrice float64   `gorm:"column:token_price" json:"token_price"`
	BLatest    int       `gorm:"column:b_latest" json:"b_latest"`
	CreatedBy  string    `gorm:"column:created_by" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedBy  string    `gorm:"column:updated_by" json:"updated_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func GetLatestGasFeeMovementTron() (float64, error) {
	var GasFeeTronMovement GasFeeTronMovement
	err := db.Table("gas_fee_movement_tron").
		Where("b_latest = 1").First(&GasFeeTronMovement).Error

	if err != nil {
		return float64(0), &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return GasFeeTronMovement.TokenPrice, nil
}

func GetGasFeeMovementTronFn(arrCond []WhereCondFn, limit int, debug bool) ([]*GasFeeTronMovement, error) {
	var result []*GasFeeTronMovement
	tx := db.Table("gas_fee_movement_tron").
		Order("gas_fee_movement_tron.created_at DESC")

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
