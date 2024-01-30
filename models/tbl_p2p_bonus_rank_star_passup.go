package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblP2PBonusRankStarPassup struct
type TblP2PBonusRankStarPassup struct {
	TBnsID           string  `json:"t_bns_id" gorm:"column:t_bns_id"`
	TBnsFrom         string  `json:"t_bns_fr" gorm:"column:t_bns_fr"`
	MemberID         int     `json:"t_member_id" gorm:"column:t_member_id"`
	FBvSmall         float64 `json:"f_bv_small" gorm:"column:f_bv_small"`
	TDirectSponsorID string  `json:"t_direct_sponsor_id" gorm:"column:t_direct_sponsor_id"`
	TDownlineID      int     `json:"t_downline_id" gorm:"column:t_downline_id"`
	TotalDownline    int     `json:"total_downline" gorm:"column:total_downline"`
}

// GetTblP2PBonusRankStarPassupFn
func GetTblP2PBonusRankStarPassupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblP2PBonusRankStarPassup, error) {
	var result []*TblP2PBonusRankStarPassup
	tx := db.Table("tbl_p2p_bonus_rank_star_passup")

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
	err := tx.Order("tbl_p2p_bonus_rank_star_passup.t_bns_id DESC").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// P2PNumberOfDownline struct
type P2PNumberOfDownline struct {
	P2PNumberOfDownline int `json:"number_of_downline" gorm:"column:number_of_downline"`
}

// GetP2PNumberOfDownline
func GetP2PNumberOfDownline(arrCond []WhereCondFn, debug bool) (*P2PNumberOfDownline, error) {
	var result P2PNumberOfDownline
	tx := db.Table("tbl_p2p_bonus_rank_star_passup").Select("COUNT(*) as number_of_downline").Where("t_bns_id = DATE_ADD(CURRENT_DATE(), INTERVAL ? DAY)", -1)

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
