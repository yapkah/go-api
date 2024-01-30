package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblMiningBonusRankStarPassup struct
type TblMiningBonusRankStarPassup struct {
	TBnsID           string  `json:"t_bns_id" gorm:"column:t_bns_id"`
	TBnsFrom         string  `json:"t_bns_fr" gorm:"column:t_bns_fr"`
	MemberID         int     `json:"t_member_id" gorm:"column:t_member_id"`
	FBvSmall         float64 `json:"f_bv_small" gorm:"column:f_bv_small"`
	FQty             float64 `json:"f_qty" gorm:"column:f_qty"`
	FQtyAcc          float64 `json:"f_qty_acc" gorm:"column:f_qty_acc"`
	TDirectSponsorID string  `json:"t_direct_sponsor_id" gorm:"column:t_direct_sponsor_id"`
	TDownlineID      int     `json:"t_downline_id" gorm:"column:t_downline_id"`
	TotalDownline    int     `json:"total_downline" gorm:"column:total_downline"`
}

// GetTblMiningBonusRankStarPassupFn
func GetTblMiningBonusRankStarPassupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblMiningBonusRankStarPassup, error) {
	var result []*TblMiningBonusRankStarPassup
	tx := db.Table("tbl_mm_bonus_rank_star_passup")

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
	err := tx.Order("tbl_mm_bonus_rank_star_passup.t_bns_id DESC").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
