package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusStaking struct
type TblqBonusStaking struct {
	BnsID         string    `gorm:"bns_id" json:"bns_id"`
	MemberId      int       `json:"member_id"`
	NickName      string    `json:"nick_name"`
	DocNo         string    `json:"doc_no"`
	GrpType       int       `json:"grp_type"`
	PrdMasterID   int       `json:"prd_master_id"`
	WalletTypeID  int       `json:"wallet_type_id"`
	StakingDate   time.Time `json:"staking_date"`
	StakingValue  float64   `json:"staking_value"`
	StakingPeriod int       `json:"staking_period"`
	FBv           float64   `json:"f_bv"`
	FPerc         float64   `json:"f_perc"`
	FBns          float64   `json:"f_bns"`
	DtPaid        time.Time `json:"dt_paid"`
	DtTimestamp   time.Time `json:"dt_timestamp"`
}

type TblqBonusStakingResult struct {
	TBnsID        string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username      string    `json:"username"`
	DocNo         string    `json:"doc_no"`
	StakingDate   time.Time `json:"staking_date"`
	StakingValue  float64   `json:"staking_value"`
	StakingPeriod int       `json:"staking_period"`
	FBv           float64   `json:"f_bv"`
	FPerc         float64   `json:"f_perc"`
	FBns          float64   `json:"f_bns"`
	DtPaid        time.Time `json:"dt_paid"`
	DtTimestamp   time.Time `json:"dt_timestamp"`
}

//get Staking Bonus by memid
func GetStakingBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusStakingResult, error) {
	var (
		rwd []*TblqBonusStakingResult
	)

	query := db.Table("tblq_bonus_staking as a").
		Select("a.bns_id as t_bns_id,b.nick_name as username,a.doc_no,a.staking_date,a.staking_value,a.staking_period,a.f_bv,a.f_perc,a.f_bns,a.dt_paid,a.dt_timestamp").
		Joins("JOIN ent_member as b ON a.member_id = b.id")

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

// GetTblqBonusStakingFn get tblq_bonus_staking data with dynamic condition
func GetTblqBonusStakingFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusStaking, error) {
	var result []*TblqBonusStaking
	tx := db.Table("tblq_bonus_staking")
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
