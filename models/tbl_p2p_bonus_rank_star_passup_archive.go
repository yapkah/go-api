package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblP2PBonusRankStarPassupArchive struct
type TblP2PBonusRankStarPassupArchive struct {
	Id                int       `json:"id" gorm:"column:id"`
	TBnsID            string    `json:"t_bns_id" gorm:"column:t_bns_id"`
	TBnsFrom          string    `json:"t_bns_fr" gorm:"column:t_bns_fr"`
	TMemberID         int       `json:"t_member_id" gorm:"column:t_member_id"`
	TMemberLot        string    `json:"t_member_lot" gorm:"column:t_member_lot"`
	TDirectSponsorID  int       `json:"t_direct_sponsor_id" gorm:"column:t_direct_sponsor_id"`
	TDirectSponsorLot string    `json:"t_direct_sponsor_lot" gorm:"column:t_direct_sponsor_lot"`
	TDownlineID       int       `json:"t_downline_id" gorm:"column:t_downline_id"`
	TDownlineLot      string    `json:"t_downline_lot" gorm:"column:t_downline_lot"`
	ILvl              int       `json:"i_lvl" gorm:"column:i_lvl"`
	FBv               float64   `json:"f_bv" gorm:"column:f_bv"`
	FQty              int       `json:"f_qty" gorm:"column:f_qty"`
	FBvAcc            float64   `json:"f_bv_acc" gorm:"column:f_bv_acc"`
	FQtyAcc           int       `json:"f_qty_acc" gorm:"column:f_qty_acc"`
	FBvCurrent        float64   `json:"f_bv_current" gorm:"column:f_bv_current"`
	FBvBig            float64   `json:"f_bv_big" gorm:"column:f_bv_big"`
	FQtyBig           int       `json:"f_qty_big" gorm:"column:f_qty_big"`
	FBvSmall          float64   `json:"f_bv_small" gorm:"column:f_bv_small"`
	FQtySmall         int       `json:"f_qty_small" gorm:"column:f_qty_small"`
	FBvDirect         float64   `json:"f_bv_direct" gorm:"column:f_bv_direct"`
	TRankEff          int       `json:"t_rank_eff" gorm:"column:t_rank_eff"`
	TRankQualify      int       `json:"t_rank_qualify" gorm:"column:t_rank_qualify"`
	TRankHighest      int       `json:"t_rank_highest" gorm:"column:t_rank_highest"`
	TRankGroup        int       `json:"t_rank_group" gorm:"column:t_rank_group"`
	TDownlineRank1    int       `json:"t_downline_rank_1" gorm:"column:t_downline_rank_1"`
	TDownlineRank2    int       `json:"t_downline_rank_2" gorm:"column:t_downline_rank_2"`
	TDownlineRank3    int       `json:"t_downline_rank_3" gorm:"column:t_downline_rank_3"`
	TDownlineRank4    int       `json:"t_downline_rank_4" gorm:"column:t_downline_rank_4"`
	TDownlineRank5    int       `json:"t_downline_rank_5" gorm:"column:t_downline_rank_5"`
	BUpdate           int       `json:"b_update" gorm:"column:b_update"`
	DtCreated         time.Time `json:"dt_created" gorm:"column:dt_created"`
}

// GetTblP2PBonusRankStarPassupArchiveFn
func GetTblP2PBonusRankStarPassupArchiveFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblP2PBonusRankStarPassupArchive, error) {
	var result []*TblP2PBonusRankStarPassupArchive
	tx := db.Table("tbl_p2p_bonus_rank_star_passup_archive")

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
	err := tx.Order("tbl_p2p_bonus_rank_star_passup_archive.t_bns_id DESC").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
