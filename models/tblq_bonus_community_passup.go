package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusCommunityPassup struct
type TblqBonusCommunityPassup struct {
	ID         int       `gorm:"t_bns_id" json:"id"`
	BnsID      string    `gorm:"bns_id" json:"t_bns_id"`
	MemberId   int       `json:"member_id"`
	DownlineId int       `json:"downline_id"`
	TDocNo     string    `json:"t_doc_no"`
	ILvl       string    `json:"i_lvl"`
	ILvlPaid   string    `json:"i_lvl_paid"`
	FBv        string    `json:"f_bv"`
	FPerc      string    `json:"f_perc"`
	FBns       string    `json:"f_bns"`
	FBnsBurn   string    `json:"f_bns_burn"`
	DtCreated  time.Time `json:"dt_created"`
}

type TblqBonusCommunityPassupResult struct {
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
	DtCreated  time.Time `json:"dt_created"`
}

// GetTblqBonusCommunityPassupFn get ewt_detail data with dynamic condition
func GetTblqBonusCommunityPassupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusCommunityPassup, error) {
	var result []*TblqBonusCommunityPassup
	tx := db.Table("tblq_bonus_community_passup")
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

//get Community passup by memid
func GetTblQCommunityBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusCommunityPassupResult, error) {
	var (
		rwd []*TblqBonusCommunityPassupResult
	)

	query := db.Table("tblq_bonus_community_passup as a").
		Select("a.bns_id as t_bns_id,b.nick_name as username,down.nick_name as downline_id,a.i_lvl,a.i_lvl_paid,a.f_bv,a.f_perc,a.f_bns,a.f_bns_burn,a.dt_created").
		Joins("JOIN ent_member as b ON a.member_id = b.id").
		Joins("JOIN ent_member as down ON down.id = a.downline_id")

	if mem_id != 0 {
		query = query.Where("a.member_id = ?", mem_id)
	}

	if dateFrom != "" {
		// dateFrom = strings.Replace(dateFrom, "-", "", 2) + "0000"
		query = query.Where("a.bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		// dateTo = strings.Replace(dateTo, "-", "", 2) + "2359"
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
