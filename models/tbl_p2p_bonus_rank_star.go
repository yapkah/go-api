package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblP2PBonusRankStar struct
type TblP2PBonusRankStar struct {
	TBnsFrom string `json:"t_bns_fr" gorm:"column:t_bns_fr"`
	TBnsTo   string `json:"t_bns_to" gorm:"column:t_bns_to"`
	MemberID int    `json:"t_member_id" gorm:"column:t_member_id"`
	TRankEff int    `json:"t_rank_eff" gorm:"column:t_rank_eff"`
	TType    string `json:"t_type" gorm:"column:t_type"`
	BLatest  int    `json:"b_latest" gorm:"column:b_latest"`
}

// GetTblP2PBonusRankStarFn get wod_member_rank data with dynamic condition
func GetTblP2PBonusRankStarFn(arrCond []WhereCondFn, debug bool) ([]*TblP2PBonusRankStar, error) {
	var result []*TblP2PBonusRankStar
	tx := db.Table("tbl_p2p_bonus_rank_star").
		Order("tbl_p2p_bonus_rank_star.dt_create DESC")
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
