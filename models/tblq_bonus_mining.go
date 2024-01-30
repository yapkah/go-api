package models

import (
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusMining struct
type TblqBonusMining struct {
	TBnsID       string    `gorm:"t_bns_id" json:"t_bns_id"`
	TMemberId    int       `json:"t_member_id"`
	TMemberLot   int       `json:"t_member_lot"`
	TUserId      string    `json:"t_user_id"`
	TFullName    string    `json:"t_full_name"`
	TRankStar    string    `json:"t_rank_star"`
	TRankEff     string    `json:"t_rank_eff"`
	TStatus      string    `json:"t_status"`
	FBnsRebate   float64   `json:"f_bns_rebate"`
	FBnsSharing  float64   `json:"f_bns_sharing"`
	FBnsGross    float64   `json:"f_bns_gross"`
	FBnsTot      float64   `json:"f_bns_tot"`
	FBnsTotLocal float64   `json:"f_bns_tot_local"`
	FRate        float64   `json:"f_rate"`
	DtCreated    time.Time `json:"dt_created"`
	Tnote        string    `json:"t_note"`
}

type MiningBonusSum struct {
	TBnsId       string  `json:"t_bns_id"`
	RebateBns    float64 `json:"rebate_bns"`
	SharingBns   float64 `json:"sharing_bns"`
	GlobalBns    float64 `json:"global_bns"`
	CommunityBns float64 `json:"community_bns"`
	SponsorBns   float64 `json:"sponsor_bns"`
	GlobalUsdBns float64 `json:"global_usd_bns"`
}

type TblqBonusMiningTotalRevenue struct {
	TotalBonusSec  float64 `json:"total_bonus_sec"`
	TotalBonusUsds float64 `json:"total_bonus_usds"`
}

// GetTblBonusFn get ewt_detail data with dynamic condition
func GetTblqBonusMiningFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusMining, error) {
	var result []*TblqBonusMining
	tx := db.Table("tblq_bonus_mining")
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

// GetGroupedMiningBnsIdRewardByMemId func
func GetGroupedMiningBnsIdRewardByMemId(memId int, dateFrom string, dateTo string) ([]*MiningBonusSum, error) {
	var rwd []*MiningBonusSum

	// query := db.Table("tblq_bonus_mining a").
	query := db.Table("rwd_period c").
		Select("c.batch_code as t_bns_id,SUM(a.f_bns_rebate) as rebate_bns,SUM(a.f_bns_sharing) as sharing_bns,SUM(a.f_bns_tot) as global_bns,SUM(b.f_bns_community) as community_bns,SUM(b.f_bns_sponsor) as sponsor_bns,SUM(b.f_bns_tot) as global_usd_bns").
		Joins("LEFT JOIN tblq_bonus_mining as a ON c.batch_code = a.t_bns_id AND a.t_member_id = ?", memId).
		Joins("LEFT JOIN tbl_mm_bonus as b ON c.batch_code = b.t_bns_id AND b.t_member_id=?", memId).
		Where("c.batch_code < CURDATE()")

	if dateFrom != "" {
		query = query.Where("c.batch_code >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("c.batch_code <= ?", dateTo)
	}

	err := query.Group("c.batch_code").Order("c.batch_code desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}

func GetMemberTotalMiningBns(memId int, rwdType string, month int, yesterday int) (*TblqBonusMiningTotalRevenue, error) {
	var rwd TblqBonusMiningTotalRevenue

	query := db.Table("tblq_bonus_mining a")

	switch strings.ToUpper(rwdType) {

	case "REBATE":
		query = query.Select("SUM(a.f_bns_rebate/a.f_rate) as total_bonus_sec,SUM(a.f_bns_rebate) as total_bonus_usds")

	case "REBATE_CRYPTO":
		query = query.Select("SUM(a.f_bns_rebate/a.f_rate) as total_bonus_sec,SUM(a.f_bns_rebate) as total_bonus_usds")

	case "SHARING":
		query = query.Select("SUM(a.f_bns_sharing/a.f_rate) as total_bonus_sec,SUM(a.f_bns_sharing) as total_bonus_usds")

	case "SHARING_PASSUP":
		query = query.Select("SUM(a.f_bns_sharing/a.f_rate) as total_bonus_sec,SUM(a.f_bns_sharing) as total_bonus_usds")

	case "COMMUNITY":
		query = db.Table("tbl_mm_bonus a")
		query = query.Select("SUM(a.f_bns_community/a.f_rate) as total_bonus_sec,SUM(a.f_bns_community) as total_bonus_usds")

	case "SPONSOR":
		query = db.Table("tbl_mm_bonus a")
		query = query.Select("SUM(a.f_bns_sponsor/a.f_rate) as total_bonus_sec,SUM(a.f_bns_sponsor) as total_bonus_usds")

	case "BZZ_RETURN":
		query = query.Select("SUM(a.f_bns_tot*0) as total_bonus_sec,SUM(a.f_bns_tot * 0) as total_bonus_usds") //show 0 for now

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
