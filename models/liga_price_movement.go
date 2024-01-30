package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// LIGA struct
type LigaPriceMovement struct {
	ID         int       `gorm:"primary_key" json:"id"`
	TokenPrice float64   `gorm:"column:token_price" json:"token_price"`
	BLatest    int       `gorm:"column:b_latest" json:"b_latest"`
	CreatedBy  string    `gorm:"column:created_by" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedBy  string    `gorm:"column:updated_by" json:"updated_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// GetLatestLigaPriceMovement
func GetLatestLigaPriceMovement() (float64, error) {
	var LigaPriceMovement LigaPriceMovement
	err := db.Table("liga_price_movement").
		Where("b_latest = 1").First(&LigaPriceMovement).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return LigaPriceMovement.TokenPrice, nil
}

// GetLigaPriceMovementFn get liga_price_movement data with dynamic condition
func GetLigaPriceMovementFn(arrCond []WhereCondFn, limit int, debug bool) ([]*LigaPriceMovement, error) {
	var result []*LigaPriceMovement
	tx := db.Table("liga_price_movement").
		Order("liga_price_movement.created_at DESC")
		// Order("liga_price_movement.created_at ASC")

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

// LIGA struct
type MinMaxLigaPriceMovement struct {
	MinTokenPrice float64 `gorm:"column:min_token_price" json:"min_token_price"`
	MaxTokenPrice float64 `gorm:"column:max_token_price" json:"max_token_price"`
}

// GetMinMaxLigaPriceMovementFn get liga_price_movement min & max data with dynamic condition
func GetMinMaxLigaPriceMovementFn(arrCond []WhereCondFn, limit int, debug bool) (*MinMaxLigaPriceMovement, error) {
	var result MinMaxLigaPriceMovement
	tx := db.Table("liga_price_movement").
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

// GetDefLigaPriceMovementFn get liga_price_movement data with dynamic condition
func GetDefLigaPriceMovementFn(dateString []string, limit int, debug bool) ([]*LigaPriceMovement, error) {
	arrDataReturn := make([]*LigaPriceMovement, 0)
	for _, dateStringV := range dateString {
		arrCond := make([]WhereCondFn, 0)
		arrCond = append(arrCond,
			// WhereCondFn{Condition: " liga_price_movement.created_at <= (SELECT created_at FROM liga_price_movement where b_latest = 1)"},
			WhereCondFn{Condition: " DATE(liga_price_movement.created_at) <= ?", CondValue: dateStringV},
		)
		ligaPriceMovement, _ := GetLigaPriceMovementFn(arrCond, limit, debug)
		if len(ligaPriceMovement) > 0 {
			arrDataReturn = append(arrDataReturn, ligaPriceMovement[0])
		}
	}
	return arrDataReturn, nil
}
