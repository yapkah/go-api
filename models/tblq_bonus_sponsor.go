package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusSponsor struct
type TblqBonusSponsor struct {
	ID        int       `gorm:"id" json:"id"`
	TBnsID    string    `gorm:"t_bns_id" json:"t_bns_id"`
	TMemberId int       `json:"t_member_id"`
	FBv       float64   `json:"f_bv"`
	FPerc     float64   `json:"f_perc"`
	FBvBf     float64   `json:"f_bv_bf"`
	NShare    int       `json:"n_share"`
	NDay      int       `json:"n_day"`
	FBns      float64   `json:"f_bns"`
	TStatus   string    `json:"t_status"`
	DtCreated time.Time `json:"dt_created"`
}

type TblqBonusSponsorResult struct {
	TBnsID     string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username   string    `json:"username"`
	DocNo      string    `json:"doc_no"`
	FBv        float64   `json:"f_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	FBnsBurn   float64   `json:"f_bns_burn"`
	DownlineID string    `json:"downline_id"`
	DtCreated  time.Time `json:"dt_created"`
}

//get Sponsor Bonus by memid
func GetSponsorBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusSponsorResult, error) {
	var (
		rwd []*TblqBonusSponsorResult
	)

	query := db.Table("tblq_bonus_sponsor as a").
		Select("a.bns_id as t_bns_id,b.nick_name as username,a.doc_no,a.f_bv,a.f_perc,a.f_bns,a.f_bns_burn,down.nick_name as downline_id,a.dt_created").
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

// GetTblqBonusSponsorFn get tblq_bonus_sponsor data with dynamic condition
func GetTblqBonusSponsorFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusSponsor, error) {
	var result []*TblqBonusSponsor
	tx := db.Table("tblq_bonus_sponsor")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("tblq_bonus_sponsor.bns_id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// ContractMiningActionRankingList struct
type ContractMiningActionRankingList struct {
	ID       int     `json:"id" gorm:"primary_key"`
	TBnsID   string  `json:"t_bns_id" gorm:"column:t_bns_id"`
	Username string  `json:"username" gorm:"column:username"`
	FBns     float64 `json:"f_bns" gorm:"column:f_bns"`
}

// GetContractMiningActionRankingListFn get ent_member_crypto with dynamic condition
func GetContractMiningActionRankingListFn(date string, maxNumber int, debug bool) ([]*ContractMiningActionRankingList, error) {
	var result []*ContractMiningActionRankingList

	tx := db.Raw("SELECT * FROM (SELECT tbl_bonus_sponsor.*, ent_member.nick_name as username "+
		"FROM tbl_bonus_sponsor "+
		"INNER JOIN ent_member ON tbl_bonus_sponsor.t_member_id = ent_member.id "+
		"WHERE date(tbl_bonus_sponsor.t_bns_id) = ? "+
		"AND tbl_bonus_sponsor.f_bns > ? "+
		"AND tbl_bonus_sponsor.t_status = ? "+
		"ORDER BY tbl_bonus_sponsor.f_bns DESC "+
		"LIMIT ? ) as a", date, 0, "A", maxNumber)

	if debug {
		tx = tx.Debug()
	}

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
