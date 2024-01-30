package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TradingBuy struct
type TradingBuy struct {
	ID                 int       `gorm:"primary_key" json:"id"`
	CryptoCode         string    `json:"crypto_code" gorm:"column:crypto_code"`
	CryptoCodeTo       string    `json:"crypto_code_to" gorm:"column:crypto_code_to"`
	DocNo              string    `json:"doc_no" gorm:"column:doc_no"`
	MemberID           int       `json:"member_id" gorm:"column:member_id"`
	TotalUnit          float64   `json:"total_unit" gorm:"column:total_unit"`
	SuggestedUnitPrice float64   `json:"suggested_unit_price" gorm:"column:suggested_unit_price"`
	UnitPrice          float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalAmount        float64   `json:"total_amount" gorm:"column:total_amount"`
	BalanceUnit        float64   `json:"balance_unit" gorm:"column:balance_unit"`
	Remark             string    `json:"remark" gorm:"column:remark"`
	SigningKey         string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash          string    `json:"trans_hash" gorm:"column:trans_hash"`
	Status             string    `json:"status" gorm:"column:status"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          string    `json:"created_by"`
	ApprovedAt         time.Time `json:"approved_at"`
	ApprovedBy         string    `json:"approved_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          string    `json:"updated_by"`
}

// GetTradingBuyFn get trading_buy with dynamic condition
func GetTradingBuyFn(arrCond []WhereCondFn, debug bool) ([]*TradingBuy, error) {
	var result []*TradingBuy
	tx := db.Table("trading_buy").
		Joins("INNER JOIN ent_member ON trading_buy.member_id = ent_member.id").
		Select("trading_buy.*, ent_member.nick_name").
		Order("trading_buy.created_at DESC")

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

// TradingBuyStruct struct
type TradingBuyStruct struct {
	ID                     int       `gorm:"primary_key" json:"id"`
	CryptoCode             string    `json:"crypto_code" gorm:"column:crypto_code"`
	CryptoFromName         string    `json:"crypto_from_name" gorm:"column:crypto_from_name"`
	CryptoToName           string    `json:"crypto_to_name" gorm:"column:crypto_to_name"`
	MemberID               int       `json:"-" gorm:"column:member_id"`
	NickName               string    `json:"nick_name" gorm:"column:nick_name"`
	TotalUnit              float64   `json:"total_unit" gorm:"column:total_unit"`
	UnitPrice              float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalAmount            float64   `json:"total_amount" gorm:"column:total_amount"`
	BalanceUnit            float64   `json:"balance_unit" gorm:"column:balance_unit"`
	Remark                 string    `json:"remark" gorm:"column:remark"`
	Status                 string    `json:"status" gorm:"column:status"`
	CreatedAt              time.Time `json:"created_at"`
	CreatedBy              string    `json:"created_by"`
	ApprovedAt             time.Time `json:"approved_at"`
	ApprovedBy             string    `json:"approved_by"`
	UpdatedAt              time.Time `json:"updated_at"`
	UpdatedBy              string    `json:"updated_by"`
	CryptoFromDecimalPoint int       `json:"crypto_from_decimal_point" gorm:"column:crypto_from_decimal_point"`
	CryptoToDecimalPoint   int       `json:"crypto_to_decimal_point" gorm:"column:crypto_to_decimal_point"`
}

// GetTradingBuyPaginateFn get trading_buy with dynamic condition
func GetTradingBuyPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*TradingBuyStruct, error) {
	var (
		result                []*TradingBuyStruct
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("trading_buy").
		Joins("INNER JOIN ent_member ON trading_buy.member_id = ent_member.id").
		Joins("INNER JOIN ewt_setup crypto_from ON trading_buy.crypto_code = crypto_from.ewallet_type_code AND crypto_from.status = 'A'").
		Joins("INNER JOIN ewt_setup crypto_to ON trading_buy.crypto_code_to = crypto_to.ewallet_type_code AND crypto_to.status = 'A'").
		Select("trading_buy.*, ent_member.nick_name, crypto_from.ewallet_type_name AS 'crypto_from_name', crypto_to.ewallet_type_name AS 'crypto_to_name', crypto_from.decimal_point AS 'crypto_from_decimal_point', crypto_to.decimal_point AS 'crypto_to_decimal_point'").
		Order("trading_buy.created_at DESC")

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

// AddTradingBuy func
func AddTradingBuy(tx *gorm.DB, arrData TradingBuy) (*TradingBuy, error) {
	if err := tx.Table("trading_buy").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// AvailableTradingBuyListStruct struct
type AvailableTradingBuyListStruct struct {
	UnitPrice        float64 `json:"unit_price" gorm:"column:unit_price"`
	TotalBalanceUnit float64 `json:"total_balance_unit" gorm:"column:total_balance_unit"`
}

// GetAvailableTradingBuyListFn get trading_buy with dynamic condition
func GetAvailableTradingBuyListFn(arrCond []WhereCondFn, limit uint, debug bool) ([]*AvailableTradingBuyListStruct, error) {
	var (
		result []*AvailableTradingBuyListStruct
	)
	tx := db.Table("trading_buy").
		Select("trading_buy.unit_price, SUM(trading_buy.balance_unit) AS 'total_balance_unit'").
		Group("trading_buy.crypto_code, trading_buy.unit_price").
		Where("trading_buy.status = 'P'").
		Order("trading_buy.unit_price DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	// Pagination and limit
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetAutoMatchTradingBuyListFn get ent_member_crypto with dynamic condition
func GetAutoMatchTradingBuyListFn(arrCond []WhereCondFn, limit int, debug bool) ([]*AutoMatchTrading, error) {
	var result []*AutoMatchTrading
	tx := db.Table("trading_buy").
		Joins("INNER JOIN ent_member ON trading_buy.member_id = ent_member.id").
		Select("trading_buy.*, ent_member.nick_name").
		Order("trading_buy.created_at ASC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type AutoMatchTradingBuyListV2 struct {
	Limit   uint8
	OrderBy string
}

// GetAutoMatchTradingBuyListFnV2 get ent_member_crypto with dynamic condition
func GetAutoMatchTradingBuyListFnV2(arrCond []WhereCondFn, arrData AutoMatchTradingBuyListV2, debug bool) ([]*AutoMatchTrading, error) {

	limit := arrData.Limit
	orderBy := arrData.OrderBy

	var result []*AutoMatchTrading
	tx := db.Table("trading_buy").
		Joins("INNER JOIN ent_member ON trading_buy.member_id = ent_member.id").
		Select("trading_buy.*, ent_member.nick_name")

	if orderBy != "" {
		tx = tx.Order(orderBy)
	}

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if limit > 0 {
		tx = tx.Limit(limit)
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
