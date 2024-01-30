package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusDistributorPassup struct
type TblBonusDistributorPassup struct {
	TBnsID       string    `json:"t_bns_id"`
	TMemberId    int       `json:"t_member_id"`
	TMemberLot   int       `json:"t_member_lot"`
	TDownlineId  int       `json:"t_downline_id"`
	TDownlineLot int       `json:"t_downline_lot"`
	TCenterId    int       `json:"t_center_id"`
	TDocNo       string    `json:"t_doc_no"`
	TItemId      int       `json:"t_item_id"`
	ILvl         int       `json:"i_lvl"`
	ILvlPaid     int       `json:"i_lvl_paid"`
	FBv          float64   `json:"f_bv"`
	FPerc        float64   `json:"f_perc"`
	FBns         float64   `json:"f_bns"`
	DtCreated    time.Time `json:"dt_created"`
	TNote        string    `json:"t_note"`
}

type TblBonusDistributorPassupResult struct {
	TBnsID     string    `json:"t_bns_id"`
	Username   string    `json:"username"`
	DownlineId string    `json:"downline_id"`
	ILvl       string    `json:"i_lvl"`
	ILvlPaid   string    `json:"i_lvl_paid"`
	FBv        float64   `json:"f_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	DtCreated  time.Time `json:"dt_created"`
}

//get Community passup by memid
func GetDistributorBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblBonusDistributorPassupResult, error) {
	var (
		rwd []*TblBonusDistributorPassupResult
	)

	query := db.Table("tbl_bonus_distributor_passup as a").
		Select("a.t_bns_id as t_bns_id,b.nick_name as username,down.nick_name as downline_id, a.i_lvl, a.i_lvl_paid,a.f_bv,a.f_perc as f_perc,a.f_bns,a.dt_created").
		Joins("JOIN ent_member as b ON a.t_member_id = b.id").
		Joins("JOIN ent_member as down ON down.id = a.t_downline_id")

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

// GetTblBonusDistributorFn get tbl_bonus_distributor_passup data with dynamic condition
func GetTblBonusDistributorFn(arrCond []WhereCondFn, debug bool) ([]*TblBonusDistributorPassup, error) {
	var result []*TblBonusDistributorPassup
	tx := db.Table("tbl_bonus_distributor_passup")
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
