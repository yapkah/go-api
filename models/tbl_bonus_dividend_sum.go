package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// BonusDividendSum struct
type BonusDividendSum struct {
	Eternity float64 `json:"total_eternity" gorm:"column:total_eternity"`
	Cullinan float64 `json:"total_cullinan" gorm:"column:total_cullinan"`
}

// GetTotalBonusPool get total bonus pool data with dynamic condition
func GetTotalBonusPool(arrCond []WhereCondFn, debug bool) (*BonusDividendSum, error) {

	var result BonusDividendSum
	tx := db.Table("tbl_bonus_dividend_sum").
		Select("admin_cf_20 AS 'total_eternity', admin_cf_80 AS 'total_cullinan'").
		Order("id DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
