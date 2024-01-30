package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EwtWithdraw struct
type EwtWithdrawPool struct {
	ID                    int       `gorm:"primary_key" json:"id"`
	EwtWithdrawId         int       `json:"ewt_withdraw_id"`
	DocNo                 string    `json:"doc_no"`
	MemberId              int       `json:"member_id`
	EwalletTypeId         int       `json:"ewallet_type_id"`
	CurrencyCode          string    `json:"currency_code"`
	Type                  string    `json:"type"`
	TransDate             time.Time `json:"trans_date"`
	Amount                float64   `json:"amount"`
	ConvertedCurrencyCode string    `json:"converted_currency_code"`
	ConversionRate        float64   `json:"conversion_rate"`
	ConvertedTotalAmount  float64   `json:"converted_total_amount"`
	Remark                string    `json:"remark"`
	TransactionData       string    `json:"transaction_data"`
	TranHash              string    `json:"tran_hash"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	CreatedBy             int       `json:"created_by"`
}

type EwtWithdrawPoolTotal struct {
	LigaPool float64 `json:"liga_pool"`
	UsdPool  float64 `json:"usd_pool"`
}

// GetEwtWithdrawPoolFn get ewt_withdraw data with dynamic condition
func GetEwtWithdrawPoolFn(arrCond []WhereCondFn, debug bool) ([]*EwtWithdrawPool, error) {
	var result []*EwtWithdrawPool
	tx := db.Table("ewt_withdraw_pool")
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

// AddEwtDetail add ewt_detail records`
func AddEwtWithdrawPool(tx *gorm.DB, saveData EwtWithdrawPool) (*EwtWithdrawPool, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtWithdrawPool-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

// get total pool amount
func GetEwtWithdrawPoolTotal(arrCond []WhereCondFn, debug bool) (*EwtWithdrawPoolTotal, error) {
	var result EwtWithdrawPoolTotal
	tx := db.Table("ewt_withdraw_pool").Select("IFNULL(SUM(amount), 0) as liga_pool, IFNULL(SUM(converted_total_amount), 0) as usd_pool")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
