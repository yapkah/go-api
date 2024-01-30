package models

import (
	"math"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TradingSetup struct
type TradingSetup struct {
	ID                         int     `gorm:"primary_key" json:"id"`
	CodeFrom                   string  `json:"code_from" gorm:"column:code_from"`
	NameFrom                   string  `json:"name_from" gorm:"column:name_from"`
	IDFrom                     int     `json:"id_from" gorm:"column:id_from"`
	IDTo                       int     `json:"id_to" gorm:"column:id_to"`
	CodeTo                     string  `json:"code_to" gorm:"column:code_to"`
	NameTo                     string  `json:"name_to" gorm:"column:name_to"`
	MinTrade                   float64 `json:"min_trade" gorm:"column:min_trade"`
	MaxTrade                   float64 `json:"max_trade" gorm:"column:max_trade"`
	DecimalPointFrom           int     `json:"decimal_point_from" gorm:"column:decimal_point_from"`
	MemberShow                 int     `json:"member_show" gorm:"column:member_show"`
	Fees                       float64 `json:"fees" gorm:"column:fees"`
	ContractAddrFrom           string  `json:"contract_address_from" gorm:"column:contract_address_from"`
	ContractAddrTo             string  `json:"contract_address_to" gorm:"column:contract_address_to"`
	ControlFrom                string  `json:"control" gorm:"column:control"`
	ControlTo                  string  `json:"control_to" gorm:"column:control_to"`
	DecimalPointTo             int     `json:"decimal_point_to" gorm:"column:decimal_point_to"`
	IsBaseFrom                 int     `json:"is_base_from" gorm:"column:is_base_from"`
	IsBaseTo                   int     `json:"is_base_to" gorm:"column:is_base_to"`
	BlockchainDecimalPointFrom int     `json:"blockchain_decimal_point_from" gorm:"column:blockchain_decimal_point_from"`
	BlockchainDecimalPointTo   int     `json:"blockchain_decimal_point_to" gorm:"column:blockchain_decimal_point_to"`
	Status                     string  `json:"status" gorm:"column:status"`
	AppSettingListFrom         string  `json:"app_setting_list_from" gorm:"column:app_setting_list_from"`
	AppSettingListTo           string  `json:"app_setting_list_to" gorm:"column:app_setting_list_to"`
	TradingBuyStatus           int     `json:"trading_buy_status" gorm:"column:trading_buy_status"`
	TradingSellStatus          int     `json:"trading_sell_status" gorm:"column:trading_sell_status"`
	TradingBuyOpenSponsorID    string  `json:"trading_buy_open_sponsor_id" gorm:"column:trading_buy_open_sponsor_id"`
	TradingSellOpenSponsorID   string  `json:"trading_sell_open_sponsor_id" gorm:"column:trading_sell_open_sponsor_id"`
}

// GetTradingSetupFn get trading_setup with dynamic condition
func GetTradingSetupFn(arrCond []WhereCondFn, debug bool) ([]*TradingSetup, error) {
	var result []*TradingSetup
	tx := db.Table("trading_setup").
		Joins("INNER JOIN ewt_setup ON trading_setup.code_from = ewt_setup.ewallet_type_code AND ewt_setup.status = 'A'").
		Joins("INNER JOIN ewt_setup AS ewt_setup_to ON trading_setup.code_to = ewt_setup_to.ewallet_type_code AND ewt_setup_to.status = 'A'").
		Select("trading_setup.*, ewt_setup.id AS 'id_from', ewt_setup.control, ewt_setup.ewallet_type_name AS 'name_from', ewt_setup.decimal_point AS 'decimal_point_from', ewt_setup.is_base AS 'is_base_from', ewt_setup.contract_address AS 'contract_address_from', ewt_setup.blockchain_decimal_point AS 'blockchain_decimal_point_from', " +
			"ewt_setup_to.id AS 'id_to', ewt_setup_to.ewallet_type_name AS 'name_to', ewt_setup_to.control AS 'control_to', ewt_setup_to.decimal_point AS 'decimal_point_to', ewt_setup_to.is_base AS 'is_base_to', ewt_setup_to.contract_address AS 'contract_address_to', ewt_setup_to.blockchain_decimal_point AS 'blockchain_decimal_point_to', " +
			"ewt_setup.app_setting_list AS 'app_setting_list_from', ewt_setup_to.app_setting_list AS 'app_setting_list_to'").
		Order("trading_setup.seq_no ASC")

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

// GetTradingSetupPaginateFn get trading_match with dynamic condition
func GetTradingSetupPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*TradingSetup, error) {
	var (
		result                []*TradingSetup
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("trading_setup").
		Joins("INNER JOIN ewt_setup ON trading_setup.code_from = ewt_setup.ewallet_type_code").
		Joins("INNER JOIN ewt_setup AS ewt_setup_to ON trading_setup.code_to = ewt_setup_to.ewallet_type_code").
		Select("trading_setup.*, ewt_setup.control, ewt_setup.ewallet_type_name AS 'name_from', ewt_setup.decimal_point AS 'decimal_point_from', " +
			"ewt_setup_to.id AS 'id_to', ewt_setup_to.ewallet_type_name AS 'name_to', ewt_setup_to.control AS 'control_to', ewt_setup_to.decimal_point AS 'decimal_point_to', " +
			"ewt_setup.app_setting_list AS 'app_setting_list_from', ewt_setup_to.app_setting_list AS 'app_setting_list_to'").
		Order("trading_setup.seq_no ASC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	arrLimitRowsSetting, _ := GetSysGeneralSetupByID("defaultlimitrow")
	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	// Total Records
	tx.Count(&totalRecord)
	oriPage := page
	if page != 0 {
		page--
	}

	newOffset := page * limit

	// Pagination and limit
	err := tx.Limit(limit).Offset(newOffset).Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	perPage = limit

	totalCurrentPageItems = int64(len(result))

	// return ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, nil
	arrPaginateData = SQLPaginateStdReturn{
		CurrentPage:           oriPage,
		PerPage:               perPage,
		TotalCurrentPageItems: totalCurrentPageItems,
		TotalPage:             totalPage,
		TotalPageItems:        totalRecord,
	}
	return arrPaginateData, result, nil
}
