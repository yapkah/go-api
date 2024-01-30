package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

type EwtWithdrawSetup struct {
	ID                int     `gorm:"column:id" json:"id"`
	WithdrawType      string  `gorm:"column:withdraw_type" json:"withdraw_type"`
	ChargesType       string  `gorm:"column:charges_type" json:"charges_type"`
	EwalletTypeId     int     `gorm:"column:ewallet_type_id" json:"ewallet_type_id"`
	EwalletTypeIdTo   int     `gorm:"column:ewallet_type_id_to" json:"ewallet_type_id_to"`
	Main              int     `gorm:"column:main" json:"main"`
	Min               float64 `gorm:"column:min" json:"min"`
	Max               float64 `gorm:"column:max" json:"max"`
	ProcessingFee     float64 `gorm:"column:processing_fee" json:"processing_fee"`
	AdminFee          float64 `gorm:"column:admin_fee" json:"admin_fee"`
	GasFee            float64 `gorm:"column:gas_fee" json:"gas_fee"`
	Markup            float64 `gorm:"column:markup" json:"markup"`
	MultipleOf        int     `gorm:"column:multiple_of" json:"multiple_of"`
	CountdownDays     int     `gorm:"column:countdown_days" json:"countdown_days"`
	WaitPreviousDone  int     `gorm:"column:wait_previous_done" json:"wait_previous_done"`
	Remark            string  `gorm:"column:remark" json:"remark"`
	Comments          string  `gorm:"column:comments" json:"comments"`
	EwalletTypeCode   string  `gorm:"column:ewt_type_code" json:"ewallet_type_code"`
	EwalletTypeName   string  `gorm:"column:ewt_type_name" json:"ewallet_type_name"`
	CurrencyCode      string  `gorm:"column:currency_code" json:"currency_code"`
	EwalletTypeCodeTo string  `gorm:"column:ewt_type_code_to" json:"ewallet_type_code_to"`
	EwalletTypeNameTo string  `gorm:"column:ewt_type_name_to" json:"ewallet_type_name_to"`
	CurrencyCodeTo    string  `gorm:"column:currency_code_to" json:"currency_code_to"`
	EwalletToBlkCCode string  `gorm:"column:ewt_to_blockchain_code" json:"ewt_to_blockchain_code"`
}

// GetEwtWithdrawSetupFn get ewt_withdraw_setup data with dynamic condition
func GetEwtWithdrawSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtWithdrawSetup, error) {
	var result []*EwtWithdrawSetup
	tx := db.Table("ewt_withdraw_setup").
		Joins("INNER JOIN ewt_setup ewt_from ON ewt_withdraw_setup.ewallet_type_id = ewt_from.id").
		Joins("LEFT JOIN ewt_setup ewt_to ON ewt_withdraw_setup.ewallet_type_id_to = ewt_to.id").
		Select("ewt_withdraw_setup.*, ewt_from.ewallet_type_code AS 'ewt_type_code', ewt_from.ewallet_type_name AS 'ewt_type_name', ewt_from.currency_code AS 'currency_code', " +
			"ewt_to.ewallet_type_code AS 'ewt_type_code_to', ewt_to.ewallet_type_name AS 'ewt_type_name_to', ewt_to.currency_code AS 'currency_code_to', ewt_to.blockchain_crypto_type_code AS 'ewt_to_blockchain_code' " + selectColumn)
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

type EwtWithdrawSetupV2 struct {
	ID               int     `gorm:"column:id" json:"id"`
	WithdrawType     string  `gorm:"column:withdraw_type" json:"withdraw_type"`
	ChargesType      string  `gorm:"column:charges_type" json:"charges_type"`
	EwalletTypeId    int     `gorm:"column:ewallet_type_id" json:"ewallet_type_id"`
	EwalletTypeIdTo  int     `gorm:"column:ewallet_type_id_to" json:"ewallet_type_id_to"`
	Main             int     `gorm:"column:main" json:"main"`
	Min              float64 `gorm:"column:min" json:"min"`
	Max              float64 `gorm:"column:max" json:"max"`
	ProcessingFee    float64 `gorm:"column:processing_fee" json:"processing_fee"`
	AdminFee         float64 `gorm:"column:admin_fee" json:"admin_fee"`
	GasFee           float64 `gorm:"column:gas_fee" json:"gas_fee"`
	Markup           float64 `gorm:"column:markup" json:"markup"`
	MultipleOf       int     `gorm:"column:multiple_of" json:"multiple_of"`
	CountdownDays    int     `gorm:"column:countdown_days" json:"countdown_days"`
	WaitPreviousDone int     `gorm:"column:wait_previous_done" json:"wait_previous_done"`
	Remark           string  `gorm:"column:remark" json:"remark"`
	Comments         string  `gorm:"column:comments" json:"comments"`
}

func GetEwtWithdrawSetupFnV2(arrCond []WhereCondFn, selectColumn string, debug bool) (*EwtWithdrawSetupV2, error) {
	var result EwtWithdrawSetupV2
	tx := db.Table("ewt_withdraw_setup")
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
