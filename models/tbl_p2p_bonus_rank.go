package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusP2PRank struct
type TblBonusP2PRank struct {
	TBnsFrom   string    `json:"t_bns_fr" gorm:"column:t_bns_fr"`
	TBnsTo     string    `json:"t_bns_to" gorm:"column:t_bns_to"`
	TMemberID  int       `json:"t_member_id" gorm:"column:t_member_id"`
	TMemberLot string    `json:"t_member_lot" gorm:"column:t_member_lot"`
	TRankOld   int       `json:"t_rank_old" gorm:"column:t_rank_old"`
	TRankEff   int       `json:"t_rank_eff" gorm:"column:t_rank_eff"`
	TPackageId int       `json:"t_package_id" gorm:"column:t_package_id"`
	TStatus    string    `json:"t_status" gorm:"column:t_status"`
	BLatest    int       `json:"b_latest" gorm:"column:b_latest"`
	DtCreate   time.Time `json:"dt_create"`
}

// GetTblP2PBonusRankFn get tbl_p2p_bonus_rank data with dynamic condition
func GetTblP2PBonusRankFn(arrCond []WhereCondFn, debug bool) ([]*TblBonusP2PRank, error) {
	var result []*TblBonusP2PRank
	tx := db.Table("tbl_p2p_bonus_rank").
		Order("tbl_p2p_bonus_rank.dt_create DESC")
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
