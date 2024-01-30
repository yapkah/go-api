package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EwtSummary struct
type EwtSummary struct {
	ID            int       `gorm:"primary_key" json:"id"`
	MemberID      int       `json:"member_id"`
	EwalletTypeID int       `json:"ewallet_type_id"`
	CurrencyCode  string    `json:"currency_code"`
	TotalIn       float64   `json:"total_in"`
	TotalOut      float64   `json:"total_out"`
	Balance       float64   `json:"balance"`
	TempBalance   float64   `json:"temp_balance"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     string    `json:"updated_by"`
}

// EwtSummarySetup struct
type EwtSummarySetup struct {
	ID              int       `gorm:"primary_key" json:"id"`
	MemberID        int       `json:"member_id"`
	EwalletTypeID   int       `json:"ewallet_type_id"`
	EwalletTypeCode string    `json:"ewallet_type_code"`
	EwalletTypeName string    `json:"ewallet_type_name"`
	CurrencyCode    string    `json:"currency_code"`
	DecimalPoint    int       `json:"decimal_point"`
	TotalIn         float64   `json:"total_in"`
	TotalOut        float64   `json:"total_out"`
	Balance         float64   `json:"balance"`
	TempBalance     float64   `json:"temp_balance"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       string    `json:"created_by"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       string    `json:"updated_by"`
}

// AddEwtSummaryStruct struct
type AddEwtSummaryStruct struct {
	ID            int       `gorm:"primary_key" json:"id"`
	MemberID      int       `json:"member_id"`
	EwalletTypeID int       `json:"ewallet_type_id"`
	CurrencyCode  string    `json:"currency_code"`
	TotalIn       float64   `json:"total_in"`
	TotalOut      float64   `json:"total_out"`
	Balance       float64   `json:"balance"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
}

// GetEwtSummaryFn get ewt_summary data with dynamic condition
func GetEwtSummaryFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtSummary, error) {
	var result []*EwtSummary
	tx := db.Table("ewt_summary")

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

// GetEwtSummaryFn get ewt_summary data with dynamic condition
func GetEwtSummaryFnTx(tx *gorm.DB, arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtSummary, error) {
	var result []*EwtSummary
	tx = tx.Table("ewt_summary")

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

// GetEwtSummarySetupFn get ewt_summary data with dynamic condition
func GetEwtSummarySetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtSummarySetup, error) {
	var result []*EwtSummarySetup
	tx := db.Table("ewt_summary")
	tx = tx.Select("ewt_setup.*, ewt_summary.total_in, ewt_summary.total_out, ewt_summary.balance" + selectColumn)
	tx = tx.Joins("INNER JOIN ewt_setup ON ewt_summary.ewallet_type_id = ewt_setup.id")

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

// GetEwtSummarySetupFn get ewt_summary data with dynamic condition
func GetEwtSummarySetupFnTx(tx *gorm.DB, arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtSummarySetup, error) {
	var result []*EwtSummarySetup
	tx = tx.Table("ewt_summary")
	tx = tx.Select("ewt_setup.*, ewt_summary.total_in, ewt_summary.total_out, ewt_summary.balance" + selectColumn)
	tx = tx.Joins("INNER JOIN ewt_setup ON ewt_summary.ewallet_type_id = ewt_setup.id")

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

// func AddEwtSummary add ewt_detail records
func AddEwtSummary(tx *gorm.DB, saveData AddEwtSummaryStruct) (*AddEwtSummaryStruct, error) {
	if err := tx.Table("ewt_summary").Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtSummary-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

// func AddEwtSummaryWithoutTx add ewt_detail records
func AddEwtSummaryWithoutTx(saveData AddEwtSummaryStruct) (*AddEwtSummaryStruct, error) {
	if err := db.Table("ewt_summary").Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtSummary-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

type TotalBalanceStruct struct {
	Balance float64 `json:"balance"`
}

func GetTotalBalance(arrCond []WhereCondFn, debug bool) (*TotalBalanceStruct, error) {
	var result TotalBalanceStruct
	tx := db.Table("ewt_summary")
	tx = tx.Select("SUM(ewt_summary.balance) AS 'balance'")
	tx = tx.Joins("INNER JOIN ewt_setup ON ewt_summary.ewallet_type_id = ewt_setup.id")

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
