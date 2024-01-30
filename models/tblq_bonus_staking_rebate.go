package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusStakingRebate struct
type TblqBonusStakingRebate struct {
	BnsId         string  `json:"bns_id"`
	MemberId      int     `json:"member_id"`
	Type          string  `json:"type"`
	FPerc         float64 `json:"f_perc"`
	FBns          float64 `json:"f_bns"`
	PersonalAsset float64 `json:"personal_asset"`
}

// TblqBonusStakingRebateFn get tblq_bonus_rebate data with dynamic condition
func TblqBonusStakingRebateFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusStakingRebate, error) {
	var result []*TblqBonusStakingRebate
	tx := db.Table("tblq_bonus_staking_rebate")
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
	err := tx.Order("bns_id ASC").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type TblqBonusStakingRebateResult struct {
	TBnsID              string    `gorm:"t_bns_id" json:"t_bns_id"`
	Username            string    `json:"username"`
	Rank                float64   `json:"rank"`
	PersonalGlobalValue float64   `json:"personal_global_value"`
	PersonalAsset       float64   `json:"personal_asset"`
	TotalGlobalAsset    float64   `json:"total_global_asset"`
	DailyRelease        float64   `json:"daily_release"`
	FPerc               float64   `json:"f_perc"`
	FBns                float64   `json:"f_bns"`
	DtCreated           time.Time `json:"dt_created"`
}

func GetStakingRebateBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusStakingRebateResult, error) {
	var (
		rwd []*TblqBonusStakingRebateResult
	)

	query := db.Table("tblq_bonus_staking_rebate as a").
		Select("a.bns_id as t_bns_id ,b.nick_name as username, a.rank, a.personal_global_value , a.personal_asset , a.total_global_asset , a.daily_release , a.f_perc, a.f_bns,a.dt_created").
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
