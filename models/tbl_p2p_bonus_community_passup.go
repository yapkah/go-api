package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblP2PBonusCommunityPassup struct
type TblP2PBonusCommunityPassup struct {
	ID           int       `gorm:"t_bns_id" json:"id"`
	TBnsID       string    `gorm:"t_bns_id" json:"t_bns_id"`
	TMemberId    int       `json:"t_member_id"`
	TMemberLot   int       `json:"t_member_lot"`
	TDownlineLot int       `json:"t_downline_lot"`
	TCenterId    int       `json:"t_center_id"`
	TDocNo       string    `json:"t_doc_no"`
	TItemId      string    `json:"t_item_id"`
	ILvl         string    `json:"i_lvl"`
	ILvlPaid     string    `json:"i_lvl_paid"`
	FBv          string    `json:"f_bv"`
	FPerc        string    `json:"f_perc"`
	FBns         string    `json:"f_bns"`
	FBnsBurn     string    `json:"f_bns_burn"`
	DtCreated    time.Time `json:"dt_created"`
}

type TblP2PBonusCommunityPassupResult struct {
	// ID           int       `gorm:"t_bns_id" json:"id"`
	TBnsID     string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username   string    `json:"username"`
	DownlineId string    `json:"downline_id"`
	ILvl       string    `json:"i_lvl"`
	ILvlPaid   string    `json:"i_lvl_paid"`
	FBv        float64   `json:"f_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	FBnsBurn   float64   `json:"f_bns_burn"`
	FRate      float64   `json:"f_rate"`
	DtCreated  time.Time `json:"dt_created"`
}

// GetTblP2PBonusCommunityPassupFn get ewt_detail data with dynamic condition
func GetTblP2PBonusCommunityPassupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblP2PBonusCommunityPassup, error) {
	var result []*TblP2PBonusCommunityPassup
	tx := db.Table("tbl_p2p_bonus_community_passup")
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

//get Community p2p passup by memid
func GetP2PCommunityPassupBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblP2PBonusCommunityPassupResult, error) {
	var (
		rwd []*TblP2PBonusCommunityPassupResult
	)

	query := db.Table("tbl_p2p_bonus_community_passup as a").
		Select("a.t_bns_id as t_bns_id,b.nick_name as username,down.nick_name as downline_id,a.i_lvl,a.i_lvl_paid,a.f_bv,a.f_perc,a.f_bns,a.f_bns_burn,bonus.f_rate,a.dt_created").
		Joins("JOIN tbl_p2p_bonus as bonus ON bonus.t_bns_id = a.t_bns_id AND a.t_member_id = bonus.t_member_id").
		Joins("JOIN ent_member as b ON a.t_member_id = b.id").
		Joins("JOIN ent_member as down ON down.id = a.t_downline_id")

	if mem_id != 0 {
		query = query.Where("a.t_member_id = ?", mem_id)
	}

	if dateFrom != "" {
		// dateFrom = strings.Replace(dateFrom, "-", "", 2) + "0000"
		query = query.Where("a.t_bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		// dateTo = strings.Replace(dateTo, "-", "", 2) + "2359"
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