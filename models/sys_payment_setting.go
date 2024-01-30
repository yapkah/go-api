package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysPaymentSetting struct
type SysPaymentSetting struct {
	ID            string `gorm:"primary_key" json:"id"`
	PlanPurchase  string `gorm:"primary_key" json:"plan_purchase"`
	EwalletTypeID int    `json:"ewallet_type_id"`
	MinPayPerc    int    `json:"min_pay_perc" gorm:"column:min_pay_perc"`
	MaxPayPerc    int    `json:"max_pay_perc" gorm:"column:max_pay_perc"`
	SSType        int    `json:"ss_type" gorm:"column:ss_type"`
	SSGroup       int    `json:"ss_group" gorm:"column:ss_group"`
	SSPosition    string `json:"ss_position" gorm:"column:ss_position"`
	Status        string `json:"value"`
}

// GetSysPaymentSettingFn func
func GetSysPaymentSettingFn(arrWhereFn []WhereCondFn, selectColumn string, debug bool) ([]*SysPaymentSetting, error) {
	var result []*SysPaymentSetting
	tx := db.Table("sys_payment_setting").
		Select("sys_payment_setting.*" + selectColumn)

	if len(arrWhereFn) > 0 {
		for _, v := range arrWhereFn {
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

// GetSysPaymentSettingFn func
func GetSysPaymentSettingFnV2(arrFn ArrModelFn, selectColumn string, debug bool) ([]*SysPaymentSetting, error) {
	var result []*SysPaymentSetting
	tx := db.Table("sys_payment_setting").
		Select("sys_payment_setting.*" + selectColumn)

	if len(arrFn.Join) > 0 {
		for _, v := range arrFn.Join {
			if v.JoinValue != nil {
				tx = tx.Joins(v.JoinTable, v.JoinValue)
			} else {
				tx = tx.Joins(v.JoinTable)
			}
		}
	}
	if len(arrFn.Where) > 0 {
		for _, v := range arrFn.Where {
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

type WalletSysPaymentSettingStruct struct {
	ID              int     `gorm:"primary_key" json:"id"`
	PlanPurchase    string  `gorm:"primary_key" json:"plan_purchase"`
	EwalletTypeID   int     `json:"ewallet_type_id"`
	MinPayPerc      int     `json:"min_pay_perc" gorm:"column:min_pay_perc"`
	MaxPayPerc      int     `json:"max_pay_perc" gorm:"column:max_pay_perc"`
	SSType          int     `json:"ss_type" gorm:"column:ss_type"`
	SSGroup         int     `json:"ss_group" gorm:"column:ss_group"`
	SSPosition      string  `json:"ss_position" gorm:"column:ss_position"`
	Status          string  `json:"value"`
	Amount          float64 `json:"amount"`
	BFree           int     `json:"b_free"`
	Balance         float64 `json:"balance"`
	EwalletTypeCode string  `json:"ewallet_type_code"`
	EwalletTypeName string  `json:"ewallet_type_name"`
	DecimalPoint    int     `json:"decimal_point"`
}

// GetWalletSysPaymentSettingFn func
func GetWalletSysPaymentSettingFn(arrWhereFn []WhereCondFn, selectColumn string, debug bool) ([]*WalletSysPaymentSettingStruct, error) {
	var result []*WalletSysPaymentSettingStruct
	tx := db.Table("sys_payment_setting").
		Select("sys_payment_setting.id, sys_payment_setting.plan_purchase, sys_payment_setting.ewallet_type_id, sys_payment_setting.min_pay_perc, sys_payment_setting.max_pay_perc, " +
			"sys_payment_setting.ss_type, sys_payment_setting.ss_group, sys_payment_setting.ss_position, sys_payment_setting.status, " +
			"wod_room_type.amount, wod_room_type.b_free, ewt_summary.balance, ewt_setup.ewallet_type_code AS 'ewallet_type_code', ewt_setup.ewallet_type_name AS 'ewallet_type_name', ewt_setup.decimal_point" + selectColumn).
		Joins("INNER JOIN ewt_setup ON sys_payment_setting.ewallet_type_id = ewt_setup.id AND ewt_setup.status = 'A'").
		Joins("INNER JOIN wod_room_type ON sys_payment_setting.plan_purchase = wod_room_type.code").
		Joins("LEFT JOIN ewt_summary ON sys_payment_setting.ewallet_type_id = ewt_summary.ewallet_type_id").
		Where("sys_payment_setting.status = 'A'")

	if len(arrWhereFn) > 0 {
		for _, v := range arrWhereFn {
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
