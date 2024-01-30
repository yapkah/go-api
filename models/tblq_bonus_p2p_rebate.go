package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusP2pRebate struct
type TblqBonusP2pRebate struct {
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

type TblqBonusP2PRebateResult struct {
	TBnsId      string    `json:"t_bns_id"`
	NickName    string    `json:"nick_name"`
	DocNo       string    `json:"doc_no"`
	FBv         float64   `json:"f_bv"`
	FPerc       float64   `json:"f_perc"`
	FBns        float64   `json:"f_bns"`
	FRate       float64   `json:"f_rate"`
	DtPaid      time.Time `json:"dt_paid"`
	DtTimestamp time.Time `json:"dt_timestamp"`
}

// TblqBonusRebateFn get tblq_bonus_rebate data with dynamic condition
func TblqBonusP2PRebateFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusP2pRebate, error) {
	var result []*TblqBonusP2pRebate
	tx := db.Table("tblq_bonus_p2p_rebate")
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

//get p2p rebate bonus by memid
func GetP2PRebateBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusP2PRebateResult, error) {
	var (
		rwd []*TblqBonusP2PRebateResult
	)

	query := db.Table("tblq_bonus_p2p_rebate as a").
		Select("a.bns_id as t_bns_id,b.nick_name,a.doc_no,a.f_bv,a.f_perc,a.f_bns,a.f_rate,a.dt_paid,a.dt_timestamp").
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
