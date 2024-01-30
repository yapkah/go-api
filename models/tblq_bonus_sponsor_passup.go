package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqBonusSponsorPassup struct
type TblqBonusSponsorPassup struct {
	BnsID      string    `gorm:"bns_id" json:"bns_id"`
	MemberId   int       `json:"member_id"`
	DocNo      string    `json:"doc_no"`
	DownlineID int       `json:"downline_id"`
	ILvl       int       `json:"i_lvl"`
	ILvlPaid   int       `json:"i_lvl_paid"`
	FBv        float64   `json:"f_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	DtCreated  time.Time `json:"dt_created"`
	DtFlush    time.Time `json:"dt_flush"`
}

type TblqBonusSponsorPassupResult struct {
	TBnsID     string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username   string    `json:"username"`
	DocNo      string    `json:"doc_no"`
	FBv        float64   `json:"f_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	DownlineID string    `json:"downline_id"`
	DtCreated  time.Time `json:"dt_created"`
	DtFlush    time.Time `json:"dt_flush"`
}

//get Sponsor Passup Bonus Passup by memid
func GetSponsorBonusPassupByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusSponsorPassupResult, error) {
	var (
		rwd []*TblqBonusSponsorPassupResult
	)

	query := db.Table("tblq_bonus_sponsor_passup as a").
		Select("a.bns_id as t_bns_id,b.nick_name as username,a.doc_no,a.f_bv,a.f_perc,a.f_bns,down.nick_name as downline_id,a.dt_created,a.dt_flush").
		Joins("JOIN ent_member as b ON a.member_id = b.id").
		Joins("JOIN ent_member as down ON down.id = a.downline_id")

	if mem_id != 0 {
		query = query.Where("a.member_id = ?", mem_id)
	}

	if dateFrom != "" {
		query = query.Where("a.bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("a.bns_id <= ?", dateTo)
	}

	err := query.Order("a.bns_id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}

// GetTblqBonusSponsorPassupFn get tblq_bonus_sponsor_passup data with dynamic condition
func GetTblqBonusSponsorPassupFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusSponsorPassup, error) {
	var result []*TblqBonusSponsorPassup
	tx := db.Table("tblq_bonus_sponsor_passup")
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
