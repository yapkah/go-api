package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblMiningBonusRankStar struct
type TblMiningBonusRankStar struct {
	TBnsFrom string `json:"t_bns_fr" gorm:"column:t_bns_fr"`
	TBnsTo   string `json:"t_bns_to" gorm:"column:t_bns_to"`
	MemberID int    `json:"t_member_id" gorm:"column:t_member_id"`
	TRankEff int    `json:"t_rank_eff" gorm:"column:t_rank_eff"`
	TType    string `json:"t_type" gorm:"column:t_type"`
	BLatest  int    `json:"b_latest" gorm:"column:b_latest"`
}

// GetTblMiningBonusRankStarFn
func GetTblMiningBonusRankStarFn(arrCond []WhereCondFn, debug bool) ([]*TblMiningBonusRankStar, error) {
	var result []*TblMiningBonusRankStar
	tx := db.Table("tbl_mm_bonus_rank_star").
		Order("tbl_mm_bonus_rank_star.dt_create DESC")
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
