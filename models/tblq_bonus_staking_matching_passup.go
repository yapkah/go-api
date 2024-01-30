package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusStakingMatchingPassup struct
type TblqBonusStakingMatchingPassup struct {
	BnsID        string    `gorm:"bns_id" json:"bns_id"`
	MemberId     int       `json:"member_id"`
	DownlineId   int       `json:"downline_id"`
	WalletTypeId int       `json:"wallet_type_id"`
	ILvl         string    `json:"i_lvl"`
	ILvlPaid     string    `json:"i_lvl_paid"`
	FBv          float64   `json:"f_bv"`
	FPerc        float64   `json:"f_perc"`
	FBns         float64   `json:"f_bns"`
	BurnBv       float64   `json:"burn_bv"`
	DtCreated    time.Time `json:"dt_created"`
	DtPaid       time.Time `json:"dt_paid"`
}

type TblqBonusStakingMatchingPassupResult struct {
	TBnsID     string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username   string    `json:"username"`
	DownlineId string    `json:"downline_id"`
	ILvl       string    `json:"i_lvl"`
	ILvlPaid   string    `json:"i_lvl_paid"`
	FBv        float64   `json:"f_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	BurnBv     float64   `json:"burn_bv"`
	DtCreated  time.Time `json:"dt_created"`
}

func GetStakingMatchingPassupBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusStakingMatchingPassupResult, error) {
	var (
		rwd []*TblqBonusStakingMatchingPassupResult
	)

	query := db.Table("tblq_bonus_staking_matching_passup as a").
		Select("a.bns_id as t_bns_id ,b.nick_name,b.nick_name as username , down.nick_name as downline_id , a.i_lvl , a.i_lvl_paid , a.f_bv , a.f_perc, a.f_bns, a.burn_bv").
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

// GetTblqBonusStakingMatchingPassupFn get tblq_bonus_staking_matching_passup data with dynamic condition
func GetTblqBonusStakingMatchingPassupFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusStakingMatchingPassup, error) {
	var result []*TblqBonusStakingMatchingPassup
	tx := db.Table("tblq_bonus_staking_matching_passup")
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
