package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EwtExchangeStruct struct
type EwtExchangeStruct struct {
	ID                   int       `gorm:"primary_key" json:"id"`
	MemberID             int       `json:"member_id"`
	DocNo                string    `json:"doc_no"`
	EwalletTypeID        int       `json:"ewallet_type_id"`
	EwalletTypeIDTo      int       `json:"ewallet_type_id_to"`
	Amount               float64   `json:"amount"`
	AdminFee             float64   `json:"admin_fee"`
	NettAmount           float64   `json:"nett_amount"`
	Rate                 float64   `json:"rate"`
	Status               string    `json:"status"`
	ConvertedTotalAmount float64   `json:"converted_total_amount"`
	CreatedAt            time.Time `json:"created_at"`
}

// GetEwtExchange func
func GetEwtExchange(arrCond []WhereCondFn, selectColumn string, debug bool) (*EwtExchangeStruct, error) {
	var ewtExchange EwtExchangeStruct
	tx := db.Table("ewt_exchange")
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
	err := tx.Find(&ewtExchange).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if ewtExchange.ID <= 0 {
		return nil, nil
	}

	return &ewtExchange, nil
}

// AddEwtExchange func
func AddEwtExchange(tx *gorm.DB, ewtExchange EwtExchangeStruct) (*EwtExchangeStruct, error) {
	if err := tx.Table("ewt_exchange").Create(&ewtExchange).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &ewtExchange, nil
}

type GetMemberTotalExchangedUsdtAmountStruct struct {
	TotalAmount float64 `json:"total_amount" gorm:"column:total_amount"`
}

func GetMemberTotalExchangedUsdtAmount(memID int, docCreatedAt string, debug bool) (*GetMemberTotalExchangedUsdtAmountStruct, error) {
	var ewtExchange GetMemberTotalExchangedUsdtAmountStruct

	query := db.Table("ewt_exchange").
		Select("SUM(ewt_detail.total_out) as total_amount").
		Joins("INNER JOIN ewt_detail ON ewt_detail.doc_no = ewt_exchange.doc_no").
		Where("ewt_exchange.member_id = ?", memID).
		Where("ewt_exchange.status = ? ", "PAID").
		Where("ewt_exchange.ewallet_type_id = ?", 6). // usds
		Where("ewt_detail.ewallet_type_id = ?", 1)    // usdt

	if docCreatedAt != "" {
		query = query.Where("ewt_exchange.paid_at <= ?", docCreatedAt)
	}

	if debug {
		query = query.Debug()
	}

	err := query.Find(&ewtExchange).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &ewtExchange, nil
}

// GetEwtExchangeFn func
func GetEwtExchangeFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtExchangeStruct, error) {
	var ewtExchange []*EwtExchangeStruct
	tx := db.Table("ewt_exchange")
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
	err := tx.Find(&ewtExchange).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return ewtExchange, nil
}

type GetTotalExchangedVolumeStruct struct {
	TotalAmount float64 `json:"total_amount" gorm:"column:total_amount"`
}

func GetTotalExchangedVolume(memID int, docCreatedAt string, debug bool) (*GetTotalExchangedVolumeStruct, error) {
	var ewtExchange GetTotalExchangedVolumeStruct

	query := db.Table("ewt_exchange").
		Select("SUM(ewt_exchange.amount) as total_amount").
		Joins("INNER JOIN ewt_detail ON ewt_detail.doc_no = ewt_exchange.doc_no")

	if docCreatedAt != "" {
		query = query.Where("ewt_exchange.paid_at <= ?", docCreatedAt)
	}

	if memID > 0 {
		query = query.Where("ewt_exchange.member_id = ?", memID)
	}

	if debug {
		query = query.Debug()
	}

	err := query.Find(&ewtExchange).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &ewtExchange, nil
}
