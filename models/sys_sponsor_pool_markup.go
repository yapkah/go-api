package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysSponsorPoolMarkup struct
type SysSponsorPoolMarkup struct {
	BnsDate    string  `json:"bns_date"`
	PoolAmount float64 `json:"pool_amount"`
}

// SysSponsorPoolMarkupFn get tblq_bonus_rebate data with dynamic condition
func GetSysSponsorPoolMarkupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SysSponsorPoolMarkup, error) {
	var result []*SysSponsorPoolMarkup
	tx := db.Table("sys_sponsor_pool_markup")

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
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
