package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

type TotalBonusMiningRebateCrypto struct {
	TotFBns float64 `json:"tot_f_bns"`
}

// GetTotalBonusMiningRebateCrypto fun
func GetTotalBonusMiningRebateCrypto(memberID int, cryptoType string) (*TotalBonusMiningRebateCrypto, error) {
	var rwd TotalBonusMiningRebateCrypto

	err := db.Table("tblq_bonus_mining_rebate_crypto").
		Select("sum(tblq_bonus_mining_rebate_crypto.f_bns) as tot_f_bns").
		Where("tblq_bonus_mining_rebate_crypto.member_id = ?", memberID).
		Where("tblq_bonus_mining_rebate_crypto.mining_type = ?", cryptoType).
		Group("tblq_bonus_mining_rebate_crypto.member_id").
		Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rwd, nil
}

type TblqBonusMiningRebateCryptoResult struct {
	TBnsId      string    `json:"t_bns_id"`
	NickName    string    `json:"nick_name"`
	DocNo       string    `json:"doc_no"`
	MiningType  string    `json:"mining_type"`
	PriceRate   float64   `json:"price_rate"`
	PriceValue  float64   `json:"price_value"`
	OwnPrice    float64   `json:"own_price"`
	MarketRate  float64   `json:"market_rate"`
	FBv         float64   `json:"f_bv"`
	FPerc       float64   `json:"f_perc"`
	FBns        float64   `json:"f_bns"`
	DtPaid      time.Time `json:"dt_paid"`
	DtTimestamp time.Time `json:"dt_timestamp"`
}

//get bonus mining rebate crypto by memid
func GetMiningRebateCryptoBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusMiningRebateCryptoResult, error) {
	var (
		rwd []*TblqBonusMiningRebateCryptoResult
	)

	query := db.Table("tblq_bonus_mining_rebate_crypto as a").
		Select("a.bns_id as t_bns_id ,b.nick_name,a.doc_no, a.mining_type , a.price_rate ,  a.price_value , a.own_price , a.market_rate , a.f_bv , a.f_perc ,a.f_bns, a.dt_paid,a.dt_timestamp").
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
