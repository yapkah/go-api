package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EwtTopup struct
type EwtTopupStruct struct {
	ID                    int       `gorm:"primary_key" json:"id"`
	MemberID              int       `json:"member_id"`
	EwalletTypeID         int       `json:"ewallet_type_id"`
	DocNo                 string    `json:"doc_no"`
	Type                  string    `json:"type"`
	CurrencyCode          string    `json:"currency_code"`
	TransDate             time.Time `json:"trans_date"`
	Status                string    `json:"status"`
	TotalIn               float64   `json:"total_in"`
	Charges               float64   `json:"charges"`
	ConvertedCurrencyCode string    `json:"converted_currency_code"`
	ConversionRate        float64   `json:"conversion_rate"`        // To Bank
	ConvertedTotalAmount  float64   `json:"converted_total_amount"` // To Bank
	AdditionalMsg         string    `json:"additional_msg"`
	Remark                string    `json:"remark"`
	CreatedBy             int       `json:"created_by"`
	ExpiryAt              time.Time `json:"expiry_at"`
	ExtraAmount           float64   `json:"extra_amount"`
	ExtraPerc             float64   `json:"extra_perc"`
	FromAddr              string    `json:"from_addr"`
}

type EwtTopupDetailStruct struct {
	ID                    int       `gorm:"primary_key" json:"id"`
	MemberID              int       `json:"member_id"`
	EwalletTypeID         int       `json:"ewallet_type_id"`
	DocNo                 string    `json:"doc_no"`
	Type                  string    `json:"type"`
	CurrencyCode          string    `json:"currency_code"`
	TransDate             time.Time `json:"trans_date"`
	Status                string    `json:"status"`
	StatusDesc            string    `json:"status_desc"`
	TotalIn               float64   `json:"total_in"`
	Charges               float64   `json:"charges"`
	ConvertedCurrencyCode string    `json:"converted_currency_code"`
	ConversionRate        float64   `json:"conversion_rate"`        // To Bank
	ConvertedTotalAmount  float64   `json:"converted_total_amount"` // To Bank
	AdditionalMsg         string    `json:"additional_msg"`
	Remark                string    `json:"remark"`
	CreatedBy             string    `json:"created_by"`
	ExpiryAt              time.Time `json:"expiry_at"`
	ExtraAmount           float64   `json:"extra_amount"`
	ExtraPerc             float64   `json:"extra_perc"`
	FromAddr              string    `json:"from_addr"`
}

// GetEwtTopupFn get ewt_setup data with dynamic condition
func GetEwtTopupFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*EwtTopupStruct, error) {
	var result EwtTopupStruct
	tx := db.Table("ewt_topup")
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

	if (err != nil && err != gorm.ErrRecordNotFound) || result.ID <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

// AddEwtTopup add member
func AddEwtTopup(tx *gorm.DB, entEwtTopup EwtTopupStruct) (*EwtTopupStruct, error) {
	if err := tx.Table("ewt_topup").Create(&entEwtTopup).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &entEwtTopup, nil
}

func GetEwtTopupArrayFn(arrCond []WhereCondFn, debug bool) ([]*EwtTopupDetailStruct, error) {
	var result []*EwtTopupDetailStruct
	tx := db.Table("ewt_topup a").
		Select("a.*,b.name as status_desc").
		Joins("inner join sys_general b ON a.status = b.code and b.type='deposit-status'")
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
