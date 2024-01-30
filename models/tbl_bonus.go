package models

import (
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonus struct
type TblBonus struct {
	TBnsID            string    `gorm:"t_bns_id" json:"t_bns_id"`
	TMemberId         int       `json:"t_member_id"`
	TMemberLot        int       `json:"t_member_lot"`
	TUserId           string    `json:"t_user_id"`
	TFullName         string    `json:"t_full_name"`
	TRankStar         string    `json:"t_rank_star"`
	TRankEff          string    `json:"t_rank_eff"`
	TStatus           string    `json:"t_status"`
	FBnsSponsor       float64   `json:"f_bns_sponsor"`
	FBnsSponsorAnnual float64   `json:"f_bns_sponsor_annual"`
	FBnsStaking       float64   `json:"f_bns_staking"`
	FBnsCommunity     float64   `json:"f_bns_community"`
	FBnsGross         float64   `json:"f_bns_gross"`
	FBnsAdj           float64   `json:"f_bns_adj"`
	FBnsTot           float64   `json:"f_bns_tot"`
	FBnsTotLocal      float64   `json:"f_bns_tot_local"`
	FRate             float64   `json:"f_rate"`
	DtCreated         time.Time `json:"dt_created"`
	Tnote             string    `json:"t_note"`
	BnsIDMax          string    `json:"bns_id_max"`
	BnsIDMin          string    `json:"bns_id_min"`
}

type TblBonusSum struct {
	TBnsId string `json:"t_bns_id"`
	// Username     string  `json:"username"`
	// RebateBns          float64 `json:"rebate_bns"`
	// MatchingBns        float64 `json:"matching_bns"`
	CommunityBns     float64 `json:"community_bns"`
	SponsorBns       float64 `json:"sponsor_bns"`
	SponsorAnnualBns float64 `json:"sponsor_annual_bns"`
	// DistributorBns     float64 `json:"distributor_bns"`
	// StakingSponsorBns  float64 `json:"staking_sponsor_bns"`
	// StakingRebateBns   float64 `json:"staking_rebate_bns"`
	// StakingMatchingBns float64 `json:"staking_matching_bns"`
	StakingBns    float64 `json:"staking_bns"`
	GlobalBns     float64 `json:"global_bns"`
	GlobalBnsConv float64 `json:"global_bns_conv"`
	// DtCreated    time.Time `json:"dt_created"`
}

type TblBonusTotalRevenue struct {
	TotalBonus     float64 `json:"total_bonus"`
	TotalBonusConv float64 `json:"total_bonus_conv"`
}

// GetTblBonusFn get ewt_detail data with dynamic condition
func GetTblBonusFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblBonus, error) {
	var result []*TblBonus
	tx := db.Table("tbl_bonus")
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
	err := tx.Order("tbl_bonus.t_bns_id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetGroupedBnsIdRewardByMemId func
func GetGroupedBnsIdRewardByMemId(memId int, dateFrom, dateTo string) ([]*TblBonusSum, error) {
	var rwd []*TblBonusSum

	query := db.Table("tbl_bonus a").
		Select("a.t_bns_id,SUM(a.f_bns_community) as community_bns,SUM(a.f_bns_sponsor) as sponsor_bns,SUM(a.f_bns_staking) as staking_bns,SUM(a.f_bns_sponsor_annual) as sponsor_annual_bns, SUM(a.f_bns_tot) as global_bns, SUM(a.f_bns_tot/f_rate) as global_bns_conv")
		// Joins("JOIN ent_member as b ON a.t_member_id = b.id")

	if memId != 0 {
		query = query.Where("a.t_member_id = ?", memId)
	}

	if dateFrom != "" {
		query = query.Where("a.t_bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("a.t_bns_id <= ?", dateTo)
	}

	err := query.Group("a.t_bns_id").Order("a.t_bns_id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}

func GetMemberTotalBns(memId int, rwdType string, month, yesterday, untilToday int) (*TblBonusTotalRevenue, error) {
	var rwd TblBonusTotalRevenue

	query := db.Table("tbl_bonus a")

	switch strings.ToUpper(rwdType) {

	case "REBATE":
		query = query.Select("SUM(a.f_bns_rebate) as total_bonus,SUM(a.f_bns_rebate/a.f_rate) as total_bonus_conv")

	case "SPONSOR":
		query = query.Select("SUM(a.f_bns_sponsor) as total_bonus,SUM(a.f_bns_sponsor/a.f_rate) as total_bonus_conv")

	case "COMMUNITY":
		query = query.Select("SUM(a.f_bns_community) as total_bonus,SUM(a.f_bns_community/a.f_rate)as total_bonus_conv")

	case "GENERATION":
		query = query.Select("SUM(a.f_bns_block) as total_bonus,SUM(a.f_bns_block/a.f_rate)as total_bonus_conv")

	case "PAIR":
		query = query.Select("SUM(a.f_bns_pair) as total_bonus,SUM(a.f_bns_pair/a.f_rate)as total_bonus_conv")

	default:
		query = query.Select("SUM(a.f_bns_tot) as total_bonus,SUM(a.f_bns_tot/a.f_rate) as total_bonus_conv")
	}

	query = query.Where("a.t_member_id = ?", memId)

	if month == 1 {
		query = query.Where("month(a.t_bns_id)= month(current_date())")
	}

	if yesterday == 1 {
		query = query.Where("a.t_bns_id = subdate(current_date(), 1)")
	}

	if untilToday == 1 {
		query = query.Where("a.t_bns_id <= subdate(current_date(), 1)")
	}

	// query = query.Debug()

	err := query.Group("a.t_member_id").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rwd, nil
}

type TblBonusTotalRevenueList struct {
	TBnsID         string  `json:"t_bns_id"`
	TotalBonus     float64 `json:"total_bonus"`
	TotalBonusConv float64 `json:"total_bonus_conv"`
	Year           string  `json:"year"`
	Month          string  `json:"month"`
	Week           string  `json:"week"`
}

func GetMemberTotalBnsList(memId int, rwdType string) ([]*TblBonusTotalRevenueList, error) {
	var rwd []*TblBonusTotalRevenueList

	query := db.Table("tbl_bonus a")

	switch strings.ToUpper(rwdType) {

	case "REBATE":
		query = query.Select("a.t_bns_id,SUM(a.f_bns_rebate) as total_bonus,SUM(a.f_bns_rebate/a.f_rate) as total_bonus_conv")

	case "SPONSOR":
		query = query.Select("a.t_bns_id,SUM(a.f_bns_sponsor) as total_bonus,SUM(a.f_bns_sponsor/a.f_rate) as total_bonus_conv")

	case "COMMUNITY":
		query = query.Select("a.t_bns_id,SUM(a.f_bns_community) as total_bonus,SUM(a.f_bns_community/a.f_rate)as total_bonus_conv")

	case "GENERATION":
		query = query.Select("a.t_bns_id,SUM(a.f_bns_block) as total_bonus,SUM(a.f_bns_block/a.f_rate)as total_bonus_conv")

	case "PAIR":
		query = query.Select("a.t_bns_id,SUM(a.f_bns_pair) as total_bonus,SUM(a.f_bns_pair/a.f_rate)as total_bonus_conv")

	default:
		query = query.Select("a.t_bns_id,SUM(a.f_bns_tot) as total_bonus,SUM(a.f_bns_tot/a.f_rate) as total_bonus_conv")
	}

	query = query.Where("a.t_member_id = ?", memId).Group("a.t_member_id,a.t_bns_id").Order("a.t_bns_id asc")

	// query = query.Debug()

	err := query.Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}

func GetMemberTotalBnsTimeFrameList(memId int, rwdType, timeFrame string) ([]*TblBonusTotalRevenueList, error) {
	var rwd []*TblBonusTotalRevenueList

	query := db.Table("tbl_bonus")

	if strings.ToUpper(timeFrame) == "MONTHLY" {
		switch strings.ToUpper(rwdType) {

		case "REBATE":
			query = query.Select("extract(month from t_bns_id) as month,extract(year from t_bns_id) as year,SUM(f_bns_rebate) as total_bonus,SUM(f_bns_rebate/f_rate) as total_bonus_conv")

		case "SPONSOR":
			query = query.Select("extract(month from t_bns_id) as month,extract(year from t_bns_id) as year,SUM(f_bns_sponsor) as total_bonus,SUM(f_bns_sponsor/f_rate) as total_bonus_conv")

		case "COMMUNITY":
			query = query.Select("extract(month from t_bns_id) as month,extract(year from t_bns_id) as year,SUM(f_bns_community) as total_bonus,SUM(f_bns_community/f_rate)as total_bonus_conv")

		case "GENERATION":
			query = query.Select("extract(month from t_bns_id) as month,extract(year from t_bns_id) as year,SUM(f_bns_block) as total_bonus,SUM(f_bns_block/f_rate)as total_bonus_conv")

		case "PAIR":
			query = query.Select("extract(month from t_bns_id) as month,extract(year from t_bns_id) as year,SUM(f_bns_pair) as total_bonus,SUM(f_bns_pair/f_rate)as total_bonus_conv")

		default:
			query = query.Select("extract(month from t_bns_id) as month,extract(year from t_bns_id) as year,SUM(f_bns_tot) as total_bonus,SUM(f_bns_tot/f_rate) as total_bonus_conv")
		}

		query = query.Where("t_member_id = ?", memId).
			Group("extract(month from t_bns_id),extract(year from t_bns_id)").
			Order("extract(month from t_bns_id),extract(year from t_bns_id) asc")

	} else if strings.ToUpper(timeFrame) == "YEARLY" {
		switch strings.ToUpper(rwdType) {

		case "REBATE":
			query = query.Select("extract(year from t_bns_id) as t_bns_id,SUM(f_bns_rebate) as total_bonus,SUM(f_bns_rebate/f_rate) as total_bonus_conv")

		case "SPONSOR":
			query = query.Select("extract(year from t_bns_id) as t_bns_id,SUM(f_bns_sponsor) as total_bonus,SUM(f_bns_sponsor/f_rate) as total_bonus_conv")

		case "COMMUNITY":
			query = query.Select("extract(year from t_bns_id) as t_bns_id,SUM(f_bns_community) as total_bonus,SUM(f_bns_community/f_rate)as total_bonus_conv")

		case "GENERATION":
			query = query.Select("extract(year from t_bns_id) as t_bns_id,SUM(f_bns_block) as total_bonus,SUM(f_bns_block/f_rate)as total_bonus_conv")

		case "PAIR":
			query = query.Select("extract(year from t_bns_id) as t_bns_id,SUM(f_bns_pair) as total_bonus,SUM(f_bns_pair/f_rate)as total_bonus_conv")

		default:
			query = query.Select("extract(year from t_bns_id) as t_bns_id,SUM(f_bns_tot) as total_bonus,SUM(f_bns_tot/f_rate) as total_bonus_conv")
		}

		query = query.Where("t_member_id = ?", memId).
			Group("extract(year from t_bns_id)").
			Order("extract(year from t_bns_id) asc")
	}

	// query = query.Debug()

	err := query.Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}

type TblBonusSumByBnsType struct {
	TotalBonus     float64 `json:"total_bonus"`
	TotalBonusConv float64 `json:"total_bonus_conv"`
}

// GetMemberSumBnsByBnsType func
func GetMemberSumBnsByBnsType(memId int, rwdType, dateFrom, dateTo string) (*TblBonusSumByBnsType, error) {
	var rwd TblBonusSumByBnsType

	query := db.Table("tbl_bonus")

	switch strings.ToUpper(rwdType) {

	case "REBATE":
		query = query.Select("SUM(f_bns_rebate) as total_bonus,SUM(f_bns_rebate/f_rate) as total_bonus_conv")

	case "SPONSOR":
		query = query.Select("SUM(f_bns_sponsor) as total_bonus,SUM(f_bns_sponsor/f_rate) as total_bonus_conv")

	case "COMMUNITY":
		query = query.Select("SUM(f_bns_community) as total_bonus,SUM(f_bns_community/f_rate)as total_bonus_conv")

	case "GENERATION":
		query = query.Select("SUM(f_bns_block) as total_bonus,SUM(f_bns_block/f_rate)as total_bonus_conv")

	case "PAIR":
		query = query.Select("SUM(f_bns_pair) as total_bonus,SUM(f_bns_pair/f_rate)as total_bonus_conv")

	default:
		query = query.Select("SUM(f_bns_tot) as total_bonus,SUM(f_bns_tot/f_rate) as total_bonus_conv")
	}

	if memId != 0 {
		query = query.Where("t_member_id = ?", memId)
	}

	if dateFrom != "" {
		query = query.Where("t_bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("t_bns_id <= ?", dateTo)
	}

	err := query.Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &rwd, nil
}
