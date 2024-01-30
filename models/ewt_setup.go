package models

import (
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EwtSetup struct
type EwtSetup struct {
	ID                       int    `gorm:"primary_key" json:"id"`
	EwtTypeCode              string `json:"ewallet_type_code" gorm:"column:ewallet_type_code"`
	EwtTypeName              string `json:"ewallet_type_name" gorm:"column:ewallet_type_name"`
	EwtGroup                 string `json:"ewallet_group" gorm:"column:ewallet_group"`
	Control                  string `json:"control" gorm:"column:control"`
	CurrencyCode             string `json:"currency_code" gorm:"column:currency_code"`
	DecimalPoint             int    `json:"decimal_point" gorm:"column:decimal_point"`
	BlockchainCryptoTypeCode string `json:"blockchain_crypto_type_code" gorm:"column:blockchain_crypto_type_code"`
	BlockchainDecimalPoint   int    `json:"blockchain_decimal_point" gorm:"column:blockchain_decimal_point"`
	Status                   string `json:"status"`
	AdminShow                int    `json:"admin_show"`
	MemberShow               int    `json:"member_show"`
	Asset                    int    `json:"asset"`
	ShowAmt                  int    `json:"show_amt"`
	IncludeSpentBalance      int    `json:"include_spent_balance"`
	WithdrawalWithCrypto     int    `json:"withdrawal_with_crypto"` // Transfer To Exchange- to crypto
	Withdraw                 int    `json:"withdraw"`               // To Third Party Exchange
	Exchange                 int    `json:"exchange"`
	FinanceTrans             int    `json:"finance_trans"`
	BlockchainDepositSetting string `json:"blockchain_deposit_setting"`
	WalletTransactionSetting string `json:"wallet_transaction_setting"`
	ContractAddress          string `json:"contract_address"`
	ShowCryptoAddr           int    `json:"show_crypto_addr"`
	CryptoAddr               int    `json:"crypto_addr"`
	CryptoLength             int    `json:"crypto_length"`
	AppSettingList           string `json:"app_setting_list" gorm:"column:app_setting_list"`
	IsBase                   int    `json:"is_base"`
}

// MemberEwtSetupBalance struct
type MemberEwtSetupBalance struct {
	ID                       int     `gorm:"primary_key" json:"id"`
	EwtTypeCode              string  `json:"ewallet_type_code" gorm:"column:ewallet_type_code"`
	EwtTypeName              string  `json:"ewallet_type_name" gorm:"column:ewallet_type_name"`
	EwtGroup                 string  `json:"ewallet_group" gorm:"column:ewallet_group"`
	Control                  string  `json:"control" gorm:"column:control"`
	CurrencyCode             string  `json:"currency_code"`
	DecimalPoint             int     `json:"decimal_point" gorm:"column:decimal_point"`
	BlockchainCryptoTypeCode string  `json:"blockchain_crypto_type_code" gorm:"column:blockchain_crypto_type_code"`
	BlockchainDecimalPoint   int     `json:"blockchain_decimal_point" gorm:"column:blockchain_decimal_point"`
	Status                   string  `json:"status"`
	MemberShow               int     `json:"member_show"`
	Asset                    int     `json:"asset"`
	ShowAmt                  float64 `json:"show_amt"`
	IncludeSpentBalance      int     `json:"include_spent_balance"`
	WithdrawalWithCrypto     int     `json:"withdrawal_with_crypto"` // Transfer To Exchange- to crypto
	Withdraw                 int     `json:"withdraw"`
	Exchange                 int     `json:"exchange"`
	FinanceTrans             int     `json:"finance_trans"`
	BlockchainDepositSetting string  `json:"blockchain_deposit_setting"`
	WalletTransactionSetting string  `json:"wallet_transaction_setting"`
	ContractAddress          string  `json:"contract_address"`
	ShowCryptoAddr           int     `json:"show_crypto_addr"`
	CryptoAddr               int     `json:"crypto_addr"`
	CryptoLength             int     `json:"crypto_length"`
	AppSettingList           string  `json:"app_setting_list" gorm:"column:app_setting_list"`
	TotalIn                  float64 `json:"total_in" gorm:"column:total_in"`
	TotalOut                 float64 `json:"total_out" gorm:"column:total_out"`
	Balance                  float64 `json:"balance" gorm:"column:balance"`
	IsBase                   int     `json:"is_base"`
}

// GetEwtSetupFn get ewt_setup data with dynamic condition
func GetEwtSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*EwtSetup, error) {
	var result EwtSetup
	tx := db.Table("ewt_setup")
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

// GetMemberEwtSetupBalanceFn get ewt_setup data with dynamic condition
func GetMemberEwtSetupBalanceFn(entMemberID int, arrCond []WhereCondFn, selectColumn string, debug bool) ([]*MemberEwtSetupBalance, error) {
	var result []*MemberEwtSetupBalance
	tx := db.Table("ewt_setup")
	tx = tx.Select("ewt_setup.*, ewt_summary.total_in, ewt_summary.total_out, ewt_summary.balance" + selectColumn)
	tx = tx.Joins("LEFT JOIN ewt_summary ON ewt_setup.id = ewt_summary.ewallet_type_id AND ewt_summary.member_id = " + strconv.Itoa(entMemberID))
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

// GetEwtSetupListFn get ewt_setup data list with dynamic condition
func GetEwtSetupListFn(arrCond []WhereCondFn, debug bool) ([]*EwtSetup, error) {
	var result []*EwtSetup
	tx := db.Table("ewt_setup")

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
