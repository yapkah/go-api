package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblP2PBonusSponsorPool struct
type TblP2PBonusSponsorPool struct {
	TBnsId    string  `json:"t_bns_id"`
	PoolCf    float64 `json:"pool_cf"`
	TotalPool float64 `json:"total_pool"`
}

// GetTblP2PBonusSponsorPoolFn get tbl_p2p_bonus_sponsor_pool data with dynamic condition
func GetTblP2PBonusSponsorPoolFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblP2PBonusSponsorPool, error) {
	var result []*TblP2PBonusSponsorPool
	tx := db.Table("tbl_p2p_bonus_sponsor_pool")
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
