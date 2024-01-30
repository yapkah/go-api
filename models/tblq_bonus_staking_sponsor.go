package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqBonusStakingSponsor struct
type TblqBonusStakingSponsor struct {
	BnsID          string    `gorm:"bns_id" json:"bns_id"`
	MemberId       int       `json:"member_id"`
	CarryForwardBv float64   `json:"carryforward_bv"`
	TodayBv        float64   `json:"today_bv"`
	TotalBv        float64   `json:"total_bv"`
	GlobalBv       float64   `json:"global_bv"`
	ReleaseBv      float64   `json:"release_bv"`
	FPerc          float64   `json:"f_perc"`
	FBns           float64   `json:"f_bns"`
	DtCreated      time.Time `json:"dt_created"`
	DtPaid         time.Time `json:"dt_created"`
}

type TblqBonusStakingSponsorResult struct {
	TBnsID         string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username       string    `json:"username"`
	CarryForwardBv float64   `json:"carryforward_bv"`
	TodayBv        float64   `json:"today_bv"`
	TotalBv        float64   `json:"total_bv"`
	GlobalBv       float64   `json:"global_bv"`
	ReleaseBv      float64   `json:"release_bv"`
	FPerc          float64   `json:"f_perc"`
	FBns           float64   `json:"f_bns"`
	DtCreated      time.Time `json:"dt_created"`
}

func GetStakingSponsorBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusStakingSponsorResult, error) {
	var (
		rwd []*TblqBonusStakingSponsorResult
	)

	query := db.Table("tblq_bonus_staking_sponsor as a").
		Select("a.bns_id as t_bns_id ,b.nick_name,b.nick_name as username ,a.carryforward_bv, a.today_bv , a.total_bv , a.global_bv , a.release_bv , a.f_perc, a.f_bns ,a.dt_created").
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

// GetTblqBonusStakingSponsorFn get tblq_bonus_staking_sponsor data with dynamic condition
func GetTblqBonusStakingSponsorFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusStakingSponsor, error) {
	var result []*TblqBonusStakingSponsor
	tx := db.Table("tblq_bonus_staking_sponsor")
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
