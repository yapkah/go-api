package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqBonusPayout struct
type TblqBonusPayout struct {
	ID                  int       `gorm:"id" json:"id"`
	TBnsID              string    `gorm:"t_bns_id" json:"t_bns_id"`
	MemberId            int       `json:"member_id"`
	DeductEwalletTypeId int       `json:"deduct_ewallet_type_id"`
	BnsType             string    `json:"bns_type"`
	PaidEwalletTypeID   int       `json:"paid_ewallet_type_id"`
	DeductAmount        float64   `json:"deduct_amount"`
	PaidAmount          float64   `json:"paid_amount"`
	PriceRate           float64   `json:"price_rate"`
	AutoWithdrawal      int       `json:"auto_withdrawal"`
	Rate                float64   `json:"rate"`
	Remark              string    `json:"remark"`
	DtTimestamp         time.Time `json:"dt_timestamp"`
	PaidAt              time.Time `json:"paid_at"`
}

// GetTblBonusPayoutFn get tbl_bonus_payout data with dynamic condition
func GetTblBonusPayoutFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusPayout, error) {
	var result []*TblqBonusPayout
	tx := db.Table("tblq_bonus_payout")
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

type TblqBonusPayoutRecord struct {
	ID          string    `json:"id"`
	TBnsId      string    `json:"t_bns_id"`
	BnsType     string    `json:"bns_type"`
	PaidAmount  float64   `json:"paid_amount"`
	PaidEwallet string    `json:"paid_ewallet"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetBonusPayoutByMemId(memId int, dateFrom string, dateTo string) ([]*TblqBonusPayoutRecord, error) {
	var rwd []*TblqBonusPayoutRecord

	query := db.Table("tblq_bonus_payout a").
		Select("a.t_bns_id,a.bns_type,a.paid_amount,b.currency_code as paid_ewallet,a.remark,a.paid_at as created_at").
		Joins("JOIN ewt_setup as b ON a.paid_ewallet_type_id = b.id")

	if memId != 0 {
		query = query.Where("a.member_id = ?", memId)
	}

	if dateFrom != "" {
		query = query.Where("a.t_bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("a.t_bns_id <= ?", dateTo)
	}

	err := query.Order("a.t_bns_id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}

type SumBonusPayoutStruct struct {
	TotalBonusPayout float64 `json:"total_bonus_payout"`
}

func GetSumTotalBonusPayoutFn(arrCond []WhereCondFn, debug bool) (*SumBonusPayoutStruct, error) {
	var result SumBonusPayoutStruct
	tx := db.Table("tblq_bonus_payout").
		Select("SUM(tblq_bonus_payout.paid_amount) AS 'total_bonus_payout'")

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

	return &result, nil
}
