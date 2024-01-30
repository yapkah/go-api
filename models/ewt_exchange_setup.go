package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

type EwtExchangeSetup struct {
	ID                int     `gorm:"column:id" json:"id"`
	EwalletTypeId     int     `gorm:"column:ewallet_type_id" json:"ewallet_type_id"`
	EwalletTypeIdTo   int     `gorm:"column:ewallet_type_id_to" json:"ewallet_type_id_to"`
	Main              int     `gorm:"column:main" json:"main"`
	Min               float64 `gorm:"column:min" json:"min"`
	Max               float64 `gorm:"column:max" json:"max"`
	ProcessingFee     float64 `gorm:"column:processing_fee" json:"processing_fee"`
	AdminFee          float64 `gorm:"column:admin_fee" json:"admin_fee"`
	Markup            float64 `gorm:"column:markup" json:"markup"`
	MultipleOf        int     `gorm:"column:multiple_of" json:"multiple_of"`
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

// GetEwtExchangeSetupFn get ewt_exchange_setup data with dynamic condition
func GetEwtExchangeSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtExchangeSetup, error) {
	var result []*EwtExchangeSetup
	tx := db.Table("ewt_exchange_setup").
		Joins("INNER JOIN ewt_setup ewt_from ON ewt_exchange_setup.ewallet_type_id = ewt_from.id").
		Joins("LEFT JOIN ewt_setup ewt_to ON ewt_exchange_setup.ewallet_type_id_to = ewt_to.id").
		Select("ewt_exchange_setup.*, ewt_from.ewallet_type_code AS 'ewt_type_code', ewt_from.ewallet_type_name AS 'ewt_type_name', ewt_from.currency_code AS 'currency_code', " +
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

type EwtExchangeSetupV2 struct {
	ID              int     `gorm:"column:id" json:"id"`
	EwalletTypeId   int     `gorm:"column:ewallet_type_id" json:"ewallet_type_id"`
	EwalletTypeIdTo int     `gorm:"column:ewallet_type_id_to" json:"ewallet_type_id_to"`
	Main            int     `gorm:"column:main" json:"main"`
	Min             float64 `gorm:"column:min" json:"min"`
	Max             float64 `gorm:"column:max" json:"max"`
	ProcessingFee   float64 `gorm:"column:processing_fee" json:"processing_fee"`
	AdminFee        float64 `gorm:"column:admin_fee" json:"admin_fee"`
	Markup          float64 `gorm:"column:markup" json:"markup"`
	MultipleOf      int     `gorm:"column:multiple_of" json:"multiple_of"`
	Remark          string  `gorm:"column:remark" json:"remark"`
	Comments        string  `gorm:"column:comments" json:"comments"`
}

func GetEwtExchangeSetupFnV2(arrCond []WhereCondFn, selectColumn string, debug bool) (*EwtExchangeSetupV2, error) {
	var result EwtExchangeSetupV2
	tx := db.Table("ewt_exchange_setup")
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
