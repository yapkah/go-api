package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblP2PBonusSponsor struct
type TblP2PBonusSponsor struct {
	ID        int       `gorm:"id" json:"id"`
	TBnsID    string    `gorm:"t_bns_id" json:"t_bns_id"`
	TMemberId int       `json:"t_member_id"`
	FBv       float64   `json:"f_bv"`
	FBvBf     float64   `json:"f_bv_bf"`
	NShare    int       `json:"n_share"`
	NDay      int       `json:"n_day"`
	FBns      float64   `json:"f_bns"`
	TStatus   string    `json:"t_status"`
	DtCreated time.Time `json:"dt_created"`
}

type TblP2PBonusSponsorResult struct {
	TBnsID    string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username  string    `json:"username"`
	FBv       float64   `json:"f_bv"`
	FBvBf     float64   `json:"f_bv_bf"`
	NShare    float64   `json:"n_share"`
	FBns      float64   `json:"f_bns"`
	FRate     float64   `json:"f_rate"`
	TStatus   string    `json:"t_status"`
	DtCreated time.Time `json:"dt_created"`
}

// GetTblP2PBonusSponsorFn get tbl_p2p_bonus_sponsor data with dynamic condition
func GetTblP2PBonusSponsorFn(arrCond []WhereCondFn, debug bool) ([]*TblP2PBonusSponsor, error) {
	var result []*TblP2PBonusSponsor
	tx := db.Table("tbl_p2p_bonus_sponsor")
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

//get P2P Sponsor Bonus by memid
func GetP2PSponsorBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblP2PBonusSponsorResult, error) {
	var (
		rwd []*TblP2PBonusSponsorResult
	)

	query := db.Table("tbl_p2p_bonus_sponsor as a").
		Select("a.t_bns_id as t_bns_id,b.nick_name as username,a.f_bv,a.f_bv_bf,a.n_share,a.f_bns,bonus.f_rate,a.t_status,a.dt_created").
		Joins("JOIN tbl_p2p_bonus as bonus ON a.t_bns_id = bonus.t_bns_id AND a.t_member_id = bonus.t_member_id").
		Joins("JOIN ent_member as b ON a.t_member_id = b.id")

	if mem_id != 0 {
		query = query.Where("a.t_member_id = ?", mem_id)
	}

	if dateFrom != "" {
		query = query.Where("a.t_bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("a.t_bns_id <= ?", dateTo)
	}

	err := query.Order("a.t_bns_id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}
