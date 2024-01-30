package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// customGasFee struct
type CustomGasFee struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Percent   float64   `gorm:"percent" json:"percent"`
	CreatedBy string    `gorm:"created_by" json:"created_by"`
	CreatedAt time.Time `gorm:"created_at" json:"created_at"`
}

// GetLastestCustomGasFee
func GetLastestCustomGasFee(arrCond []WhereCondFn, selectColumn string, debug bool) (*CustomGasFee, error) {
	var customGasFee CustomGasFee

	tx := db.Table("custom_gas_fee")

	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if debug {
		tx = tx.Debug()
	}

	err := tx.Order("id desc").
		Limit(1).Find(&customGasFee).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &customGasFee, nil
}
