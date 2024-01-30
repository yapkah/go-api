package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysGeneralPaymentSetting struct
type SysGeneralPaymentSetting struct {
	ID                     int    `gorm:"primary_key" json:"id"`
	EwalletTypeID          int    `json:"ewallet_type_id"`
	EwalletTypeCode        string `json:"ewallet_type_code"`
	EwalletTypeName        string `json:"ewallet_type_name"`
	AppSettingList         string `json:"app_setting_list"`
	CurrencyCode           string `json:"currency_code"`
	Control                string `json:"control"`
	Module                 string `json:"module"`
	Type                   string `json:"type"`
	DecimalPoint           int    `json:"decimal_point"`
	BlockchainDecimalPoint int    `json:"blockchain_decimal_point"`
	Status                 string `json:"status"`
	Main                   int    `json:"main"`
	MinPayPerc             int    `json:"min_pay_perc"`
	MaxPayPerc             int    `json:"max_pay_perc"`
	Condition              string `json:"condition"`
	SeqNo                  int    `json:"seq_no"`
	ContractAddress        string `json:"contract_address"`
	IsBase                 int    `json:"is_base"`
}

// GetGeneralPaymentSettingFn func
func GetGeneralPaymentSettingFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]SysGeneralPaymentSetting, error) {
	var sysGeneralPaymentSetting []SysGeneralPaymentSetting
	tx := db.Table("sys_general_payment_setting").
		Joins("JOIN ewt_setup ON ewt_setup.id = sys_general_payment_setting.ewallet_type_id AND ewt_setup.status = 'A'").
		Select("sys_general_payment_setting.*, ewt_setup.ewallet_type_code, ewt_setup.ewallet_type_name, ewt_setup.currency_code, ewt_setup.control, ewt_setup.decimal_point, ewt_setup.contract_address, ewt_setup.app_setting_list, ewt_setup.is_base, ewt_setup.blockchain_decimal_point").
		Order("sys_general_payment_setting.seq_no ASC")

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
	err := tx.Find(&sysGeneralPaymentSetting).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return sysGeneralPaymentSetting, nil
}

// SysGeneralPaymentType struct
type SysGeneralPaymentType struct {
	Type string `json:"type"`
}

// GetPaymentTypeByModules func
func GetPaymentTypeByModules(module, paymentType string) ([]*SysGeneralPaymentType, error) {
	var sysGeneralPaymentType []*SysGeneralPaymentType
	tx := db.Table("sys_general_payment_setting").
		Select("sys_general_payment_setting.type").
		Where("module = ?", module).
		Where("status = ?", "A")

	if paymentType != "" {
		tx = tx.Where("type = ?", paymentType)
	}

	tx = tx.Group("sys_general_payment_setting.type")

	err := tx.Find(&sysGeneralPaymentType).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return sysGeneralPaymentType, nil
}

// PAYMENT VALIDATION START BELOW

// FirstContract func to validate if it is first purchase
func (sysGeneralPaymentSetting *SysGeneralPaymentSetting) FirstContract(memberID int) (bool, error) {
	var slsMaster SlsMaster
	err := db.Table("sls_master").
		Where("sls_master.member_id = ?", memberID).
		Where("sls_master.status IN(?,?)", "A", "P").
		Where("sls_master.doc_type IN(?)", "CT").
		First(&slsMaster).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if slsMaster.ID <= 0 {
		return true, nil
	}

	return false, nil
}

// NotFirstContract func to validate if it is first purchase
func (sysGeneralPaymentSetting *SysGeneralPaymentSetting) NotFirstContract(memberID int) (bool, error) {
	var slsMaster SlsMaster
	err := db.Table("sls_master").
		Where("sls_master.member_id = ?", memberID).
		Where("sls_master.status IN(?,?)", "A", "P").
		Where("sls_master.doc_type IN(?)", "CT").
		First(&slsMaster).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if slsMaster.ID <= 0 {
		return false, nil
	}

	return true, nil
}
