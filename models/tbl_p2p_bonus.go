package models

import (
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblP2PBonus struct
type TblP2PBonus struct {
	TBnsID          string    `gorm:"t_bns_id" json:"t_bns_id"`
	TMemberId       int       `json:"t_member_id"`
	TMemberLot      int       `json:"t_member_lot"`
	TUserId         string    `json:"t_user_id"`
	TFullName       string    `json:"t_full_name"`
	TRankStar       string    `json:"t_rank_star"`
	TRankEff        string    `json:"t_rank_eff"`
	TStatus         string    `json:"t_status"`
	FBnsRebate      float64   `json:"f_bns_rebate"`
	FBnsSponsor     float64   `json:"f_bns_sponsor"`
	FBnsDistributor float64   `json:"f_bns_distributor"`
	FBnsMatching    float64   `json:"f_bns_matching"`
	FBnsDividend    float64   `json:"f_bns_dividend"`
	FBnsDividend2   float64   `json:"f_bns_dividend2"`
	FBnsCommunity   float64   `json:"f_bns_community"`
	FBnsGross       float64   `json:"f_bns_gross"`
	FBnsAdj         float64   `json:"f_bns_adj"`
	FBnsTot         float64   `json:"f_bns_tot"`
	FBnsTotLocal    float64   `json:"f_bns_tot_local"`
	FRate           float64   `json:"f_rate"`
	DtCreated       time.Time `json:"dt_created"`
	Tnote           string    `json:"t_note"`
}

type TblP2PBonusSum struct {
	TBnsId string `json:"t_bns_id"`
	// Username     string  `json:"username"`
	RebateBns      float64 `json:"rebate_bns"`
	MatchingBns    float64 `json:"matching_bns"`
	CommunityBns   float64 `json:"community_bns"`
	SponsorBns     float64 `json:"sponsor_bns"`
	DistributorBns float64 `json:"distributor_bns"`
	GlobalBns      float64 `json:"global_bns"`
	// DtCreated    time.Time `json:"dt_created"`
}

type TblP2PBonusTotalRevenue struct {
	TotalBonusSec  float64 `json:"total_bonus_sec"`
	TotalBonusUsds float64 `json:"total_bonus_usds"`
}

// GetTblBonusFn get ewt_detail data with dynamic condition
func GetTblP2PBonusFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblP2PBonus, error) {
	var result []*TblP2PBonus
	tx := db.Table("tbl_p2p_bonus")
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

// GetGroupedP2PBnsIdRewardByMemId func
func GetGroupedP2PBnsIdRewardByMemId(memId int, dateFrom string, dateTo string) ([]*TblP2PBonusSum, error) {
	var rwd []*TblP2PBonusSum

	query := db.Table("tbl_p2p_bonus a").
		Select("a.t_bns_id,SUM(a.f_bns_rebate) as rebate_bns,SUM(a.f_bns_matching) as matching_bns, SUM(a.f_bns_community) as community_bns,SUM(a.f_bns_sponsor) as sponsor_bns,SUM(a.f_bns_distributor) as distributor_bns,SUM(a.f_bns_tot) as global_bns")
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

func GetMemberTotalP2PBns(memId int, rwdType string, month int, yesterday int) (*TblP2PBonusTotalRevenue, error) {
	var rwd TblP2PBonusTotalRevenue

	query := db.Table("tbl_p2p_bonus a")

	switch strings.ToUpper(rwdType) {

	case "REBATE":
		query = query.Select("SUM(a.f_bns_rebate/a.f_rate) as total_bonus_sec,SUM(a.f_bns_rebate) as total_bonus_usds")

	case "MATCHING":
		query = query.Select("SUM(a.f_bns_matching/a.f_rate) as total_bonus_sec,SUM(a.f_bns_matching) as total_bonus_usds")

	case "COMMUNITY":
		query = query.Select("SUM(a.f_bns_community/a.f_rate) as total_bonus_sec,SUM(a.f_bns_community)as total_bonus_usds")

	case "SPONSOR":
		query = query.Select("SUM(a.f_bns_sponsor/a.f_rate) as total_bonus_sec,SUM(a.f_bns_sponsor) as total_bonus_usds")

	case "DISTRIBUTOR":
		query = query.Select("SUM(a.f_bns_distributor/a.f_rate) as total_bonus_sec,SUM(a.f_bns_distributor) as total_bonus_usds")

	case "GLOBAL":
		query = query.Select("SUM(a.f_bns_tot/a.f_rate) as total_bonus_sec,SUM(a.f_bns_tot) as total_bonus_usds")

	default:
		query = query.Select("SUM(a.f_bns_tot/a.f_rate) as total_bonus_sec,SUM(a.f_bns_tot) as total_bonus_usds")
	}

	query = query.Where("a.t_member_id = ?", memId)

	if month == 1 {
		query = query.Where("month(a.t_bns_id)= month(current_date())")
	}

	if yesterday == 1 {
		query = query.Where("a.t_bns_id = subdate(current_date(), 1)")
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
