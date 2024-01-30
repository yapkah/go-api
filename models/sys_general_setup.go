package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
)

// SysGeneralSetup struct
type SysGeneralSetup struct {
	ID            int    `gorm:"primary_key" json:"id"`
	SettingID     string `gorm:"primary_key" json:"setting_id"`
	ParentID      string `json:"parent_id" gorm:"column:parent_id"`
	SettingTitle  string `json:"setting_title" gorm:"column:setting_title"`
	SettingDesc   string `json:"setting_desc" gorm:"column:setting_desc"`
	InputType1    string `json:"input_type_1" gorm:"column:input_type_1"`
	InputValue1   string `json:"input_value_1" gorm:"column:input_value_1"`
	SettingValue1 string `json:"setting_value_1" gorm:"column:setting_value_1"`
	InputType2    string `json:"input_type_2" gorm:"column:input_type_2"`
	InputValue2   string `json:"input_value_2" gorm:"column:input_value_2"`
	SettingValue2 string `json:"setting_value_2" gorm:"column:setting_value_2"`
	InputType3    string `json:"input_type_3" gorm:"column:input_type_3"`
	InputValue3   string `json:"input_value_3" gorm:"column:input_value_3"`
	SettingValue3 string `json:"setting_value_3" gorm:"column:setting_value_3"`
	InputType4    string `json:"input_type_4" gorm:"column:input_type_4"`
	InputValue4   string `json:"input_value_4" gorm:"column:input_value_4"`
	SettingValue4 string `json:"setting_value_4" gorm:"column:setting_value_4"`
	InputType5    string `json:"input_type_5" gorm:"column:input_type_5"`
	InputValue5   string `json:"input_value_5" gorm:"column:input_value_5"`
	SettingValue5 string `json:"setting_value_5" gorm:"column:setting_value_5"`
}

// GetSysGeneralSetupByID func
func GetSysGeneralSetupByID(settingID string) (*SysGeneralSetup, error) {
	var sys SysGeneralSetup
	err := db.Where("setting_id = ?", settingID).First(&sys).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: err.Error(), Data: map[string]string{"settingID": settingID}}
		}
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return &sys, nil
}

// GetHotWalletInfoRst struct
type GetHotWalletInfoRst struct {
	HotWalletAddress    string
	HotWalletPrivateKey string
}

// GetHotWalletInfo func
func GetHotWalletInfo() (*GetHotWalletInfoRst, error) {
	var result SysGeneralSetup

	err := db.Where("setting_id = 'hotwallet_info'").First(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: "hotwallet_info_no_set"}
		}
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	// start checking for the addr
	scryptText := result.InputType1
	generatedScryptedCryptoAddrString := result.InputValue1
	addrByte := []byte(scryptText)
	cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt1").String()

	err = CompareHashAndScryptedValue(generatedScryptedCryptoAddrString, addrByte, cryptoSalt1)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: "invalid_crypto_address_info"}
	}
	// end checking for the addr

	// start checking for the pk
	scryptText = result.SettingValue1
	generatedScryptedCryptoAddrString = result.InputType2
	addrByte = []byte(scryptText)

	err = CompareHashAndScryptedValue(generatedScryptedCryptoAddrString, addrByte, cryptoSalt1)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: "invalid_pk_info"}
	}
	// end checking for the pk

	arrDataReturn := GetHotWalletInfoRst{
		HotWalletAddress:    result.InputType1,
		HotWalletPrivateKey: result.SettingValue1,
	}

	return &arrDataReturn, nil
}

// SECP2PPoolWalletInfoRst struct
type SECP2PPoolWalletInfoRst struct {
	WalletAddress    string
	WalletPrivateKey string
}

// GetSECP2PPoolWalletInfo func
func GetSECP2PPoolWalletInfo() (*SECP2PPoolWalletInfoRst, error) {
	var result SysGeneralSetup

	err := db.Where("setting_id = 'sec_p2p_pool_wallet_info'").First(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: "sec_p2p_pool_wallet_info_no_set"}
		}
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	// start checking for the addr
	scryptText := result.InputType1
	generatedScryptedCryptoAddrString := result.InputValue1
	addrByte := []byte(scryptText)
	cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt1").String()

	err = CompareHashAndScryptedValue(generatedScryptedCryptoAddrString, addrByte, cryptoSalt1)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: "invalid_crypto_address_info"}
	}
	// end checking for the addr

	// start checking for the pk
	scryptText = result.SettingValue1
	generatedScryptedCryptoAddrString = result.InputType2
	addrByte = []byte(scryptText)

	err = CompareHashAndScryptedValue(generatedScryptedCryptoAddrString, addrByte, cryptoSalt1)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: "invalid_pk_info"}
	}
	// end checking for the pk

	arrDataReturn := SECP2PPoolWalletInfoRst{
		WalletAddress:    result.InputType1,
		WalletPrivateKey: result.SettingValue1,
	}

	return &arrDataReturn, nil
}
