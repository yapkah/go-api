package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqBonusRebate struct
type TblqBonusRebate struct {
	BnsId        string    `json:"t_bns_id"`
	MemberId     int       `json:"member_id"`
	NickName     string    `json:"nick_name"`
	DocNo        string    `json:"doc_no"`
	GrpType      int       `json:"grp_type"`
	PrdMasterId  int       `json:"prd_master_id"`
	WalletTypeId int       `json:"wallet_type_id"`
	BnsDays      int       `json:"bns_days"`
	PackageValue float64   `json:"package_value"`
	FBv          float64   `json:"f_bv"`
	FPerc        float64   `json:"f_perc"`
	FBns         float64   `json:"f_bns"`
	FRate        float64   `json:"f_rate"`
	DtPaid       time.Time `json:"dt_paid"`
	DtTimestamp  time.Time `json:"dt_timestamp"`
}

type TblqBonusRebateResult struct {
	TBnsId      string    `json:"t_bns_id"`
	NickName    string    `json:"nick_name"`
	MemberTier  string    `json:"member_tier"`
	FBv         float64   `json:"f_bv"`
	FPerc       float64   `json:"f_perc"`
	FBns        float64   `json:"f_bns"`
	FRate       float64   `json:"f_rate"`
	DtPaid      time.Time `json:"dt_paid"`
	DtTimestamp time.Time `json:"dt_timestamp"`
}

// TblqBonusRebateFn get tblq_bonus_rebate data with dynamic condition
func TblqBonusRebateFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusRebate, error) {
	var result []*TblqBonusRebate
	tx := db.Table("tblq_bonus_rebate")
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

//get rebate bonus by memid
func GetRebateBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusRebateResult, error) {
	var (
		rwd []*TblqBonusRebateResult
	)

	query := db.Table("tblq_bonus_rebate as a").
		Select("a.bns_id as t_bns_id,b.nick_name,a.member_tier,a.f_bv,a.f_perc,a.f_bns,a.f_rate,a.dt_paid,a.dt_timestamp").
		Joins("JOIN ent_member as b ON a.member_id = b.id")

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

type TblqBonusRebateGroup struct {
	BnsID    string  `json:"bns_id"`
	TotalBns float64 `json:"total_bns"`
}

// TblqBonusRebateGroupFn
func TblqBonusRebateGroupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusRebateGroup, error) {
	var result []*TblqBonusRebateGroup
	tx := db.Table("tblq_bonus_rebate").
		Select("bns_id, sum(f_bns) as total_bns").
		Group("bns_id").
		Order("bns_id desc")

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
