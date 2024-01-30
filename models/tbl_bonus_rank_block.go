package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusRankBlock struct
type TblBonusRankBlock struct {
	TBnsFr       string    `json:"t_bns_fr"`
	TBnsTo       string    `json:"t_bns_to"`
	MemberId     int       `json:"t_member_id"`
	MemberLot    int       `json:"t_member_lot"`
	TRankEff     int       `json:"t_rank_eff"`
	TRankQualify int       `json:"t_rank_qualify"`
	TRankGroup   int       `json:"t_rank_group"`
	TType        int       `json:"t_type"`
	TStatus      string    `json:"t_status"`
	BLatest      int       `json:"b_latest"`
	DtCreate     time.Time `json:"dt_create"`
}

// TblBonusRankBlockFn get tblq_bonus_matching data with dynamic condition
func GetTblBonusRankBlockFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblBonusRankBlock, error) {
	var result []*TblBonusRankBlock
	tx := db.Table("tbl_bonus_rank_block")
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
