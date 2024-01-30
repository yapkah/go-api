package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusRankStarPassup struct
type TblBonusRankStarPassup struct {
	ID                int     `json:"id" gorm:"column:id"`
	TBnsID            string  `json:"t_bns_id" gorm:"column:t_bns_id"`
	TBnsFr            string  `json:"t_bns_fr" gorm:"column:t_bns_fr"`
	MemberID          int     `json:"t_member_id" gorm:"column:t_member_id"`
	TMemberLot        int     `json:"t_member_lot" gorm:"column:t_member_lot"`
	TDirectSponsorID  int     `json:"t_direct_sponsor_id" gorm:"column:t_direct_sponsor_id"`
	TDirectSponsorLot int     `json:"t_direct_sponsor_lot" gorm:"column:t_direct_sponsor_lot"`
	TDownlineID       int     `json:"t_downline_id" gorm:"column:t_downline_id"`
	TDownlineLot      string  `json:"t_downline_lot" gorm:"column:t_downline_lot"`
	ILvl              int     `json:"i_lvl" gorm:"column:i_lvl"`
	FBV               float64 `json:"f_bv" gorm:"column:f_bv"`
	FQty              int     `json:"f_qty" gorm:"column:f_qty"`
	FBVAcc            float64 `json:"f_bv_acc" gorm:"column:f_bv_acc"`
	FQtyAcc           float64 `json:"f_qty_acc" gorm:"column:f_qty_acc"`
	FBVCurrent        float64 `json:"f_bv_current" gorm:"column:f_bv_current"`
	FBVBig            float64 `json:"f_bv_big" gorm:"column:f_bv_big"`
	FQtyBig           float64 `json:"f_qty_big" gorm:"column:f_qty_big"`
	FBvSmall          float64 `json:"f_bv_small" gorm:"column:f_bv_small"`
	FQtySmall         float64 `json:"f_qty_small" gorm:"column:f_qty_small"`
	FBvDirect         float64 `json:"f_bv_direct" gorm:"column:f_bv_direct"`
	TRankBlock        int     `json:"t_rank_block" gorm:"column:t_rank_block"`
	TotalDownline     int     `json:"total_downline" gorm:"column:total_downline"`
}

// GetTblBonusRankStarPassupFn
func GetTblBonusRankStarPassupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblBonusRankStarPassup, error) {
	var result []*TblBonusRankStarPassup
	tx := db.Table("tbl_bonus_rank_star_passup")

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
	err := tx.Order("tbl_bonus_rank_star_passup.t_bns_id DESC").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// TblBonusRankStarPassup struct
type NumberOfDownline struct {
	NumberOfDownline int `json:"number_of_downline" gorm:"column:number_of_downline"`
}

// GetTblBonusRankStarPassupFn
func GetNumberOfDownline(arrCond []WhereCondFn, debug bool) (*NumberOfDownline, error) {
	var result NumberOfDownline
	tx := db.Table("tbl_bonus_rank_star_passup").Select("COUNT(*) as number_of_downline").Where("t_bns_id = DATE_ADD(CURRENT_DATE(), INTERVAL ? DAY)", -1)

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
