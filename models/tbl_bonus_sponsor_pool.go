package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusSponsorPool struct
type TblBonusSponsorPool struct {
	TBnsId    string  `json:"t_bns_id"`
	PoolCf    float64 `json:"pool_cf"`
	TotalPool float64 `json:"total_pool"`
}

// TblBonusSponsorPoolFn get tblq_bonus_rebate data with dynamic condition
func GetTblBonusSponsorPoolFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblBonusSponsorPool, error) {
	var result []*TblBonusSponsorPool
	tx := db.Table("tbl_bonus_sponsor_pool")
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
