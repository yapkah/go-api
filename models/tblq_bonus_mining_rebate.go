package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

type TotalBonusMiningRebate struct {
	TotFBns float64 `json:"tot_f_bns"`
}

// GetTotalBonusMiningRebate fun
func GetTotalBonusMiningRebate(memberID int) (*TotalBonusMiningRebate, error) {
	var rwd TotalBonusMiningRebate

	err := db.Table("tblq_bonus_mining_rebate").
		Select("sum(tblq_bonus_mining_rebate.f_bns) as tot_f_bns").
		Where("tblq_bonus_mining_rebate.member_id = ?", memberID).
		Where("tblq_bonus_mining_rebate.mining_type = ?", "SEC").
		Group("tblq_bonus_mining_rebate.member_id").
		Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rwd, nil
}

type TblqBonusMiningRebateResult struct {
	TBnsId   string  `json:"t_bns_id"`
	NickName string  `json:"nick_name"`
	TDocNo   string  `json:"t_doc_no"`
	PPrice   float64 `json:"p_price"`
	PValue   float64 `json:"p_value"`
	// SecPrice    float64   `json:"sec_price"`
	RebatePerc  float64   `json:"rebate_perc"`
	FBv         float64   `json:"f_bv"`
	FPerc       float64   `json:"f_perc"`
	FBns        float64   `json:"f_bns"`
	DtPaid      time.Time `json:"dt_paid"`
	DtTimestamp time.Time `json:"dt_timestamp"`
}

//get bonus mining rebate by memid
func GetMiningRebateBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusMiningRebateResult, error) {
	var (
		rwd []*TblqBonusMiningRebateResult
	)

	query := db.Table("tblq_bonus_mining_rebate as a").
		Select("a.bns_id as t_bns_id ,a.doc_no as t_doc_no ,b.nick_name, a.p_price , a.p_value , a.rebate_perc , a.f_bv , a.f_perc , a.f_bns,a.dt_paid,a.dt_timestamp").
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
