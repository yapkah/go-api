package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusRankBlockLeg struct
type TblBonusRankBlockLeg struct {
	ID           int     `json:"id" gorm:"column:id"`
	TBnsID       string  `json:"t_bns_id" gorm:"column:t_bns_id"`
	MemberID     int     `json:"t_member_id" gorm:"column:t_member_id"`
	TMemberLot   int     `json:"t_member_lot" gorm:"column:t_member_lot"`
	TDownlineID  int     `json:"t_downline_id" gorm:"column:t_downline_id"`
	TDownlineLot string  `json:"t_downline_lot" gorm:"column:t_downline_lot"`
	FAmt         float64 `json:"f_amt" gorm:"column:f_amt"`
	FAmtDownline float64 `json:"f_amt_downline" gorm:"column:f_amt_downline"`
	DtCreated    float64 `json:"dt_created" gorm:"column:dt_created"`
}

// GetTblBonusRankStarPassupFn
func GetTblBonusRankBlockLegFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblBonusRankBlockLeg, error) {
	var result []*TblBonusRankBlockLeg
	tx := db.Table("tbl_bonus_rank_block_leg")

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
	err := tx.Order("tbl_bonus_rank_block_leg.t_bns_id DESC").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type SumTotalDirectDownlineAmt struct {
	TotalAmount float64 `json:"total_amount" gorm:"column:total_amount"`
}

func GetTotalDirectSponsorBlockAmount(memID int) (*SumTotalDirectDownlineAmt, error) {
	var result SumTotalDirectDownlineAmt

	query := db.Table("tbl_bonus_rank_block_leg").
		Select("SUM(f_amt_downline) as total_amount").
		Where("t_member_id = ?", memID)

	err := query.Find(&result).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
