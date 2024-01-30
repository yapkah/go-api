package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TradingMatch struct
type TradingMatch struct {
	ID             int       `gorm:"primary_key" json:"id"`
	DocNo          string    `json:"doc_no" gorm:"column:doc_no"`
	CryptoCode     string    `json:"crypto_code" gorm:"column:crypto_code"`
	SellID         int       `json:"sell_id" gorm:"column:sell_id"`
	BuyID          int       `json:"buy_id" gorm:"column:buy_id"`
	SellerMemberID int       `json:"seller_member_id" gorm:"column:seller_member_id"`
	BuyerMemberID  int       `json:"buyer_member_id" gorm:"column:buyer_member_id"`
	TotalUnit      float64   `json:"total_unit" gorm:"column:total_unit"`
	UnitPrice      float64   `json:"unit_price" gorm:"column:unit_price"`
	ExchangePrice  float64   `json:"exchange_price" gorm:"column:exchange_price"`
	TotalAmount    float64   `json:"total_amount" gorm:"column:total_amount"`
	SigningKey     string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash      string    `json:"trans_hash" gorm:"column:trans_hash"`
	Remark         string    `json:"remark" gorm:"column:remark"`
	Status         string    `json:"status" gorm:"column:status"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
	ApprovedAt     time.Time `json:"approved_at"`
	ApprovedBy     string    `json:"approved_by"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      string    `json:"updated_by"`
}

// TradingMatchStruct struct
type TradingMatchStruct struct {
	ID                int       `gorm:"primary_key" json:"id"`
	DocNo             string    `json:"doc_no" gorm:"column:doc_no"`
	TradSellID        string    `json:"sell_id" gorm:"column:sell_id"`
	TradBuyID         string    `json:"buy_id" gorm:"column:buy_id"`
	BuyerEntMemberID  int       `json:"buyer_member_id" gorm:"column:buyer_member_id"`
	SellerEntMemberID int       `json:"seller_member_id" gorm:"column:seller_member_id"`
	BuyerNickName     string    `json:"buyer_username" gorm:"column:buyer_username"`
	SellerNickName    string    `json:"seller_username" gorm:"column:seller_username"`
	CryptoCode        string    `json:"crypto_code" gorm:"column:crypto_code"`
	TotalUnit         float64   `json:"total_unit" gorm:"column:total_unit"`
	UnitPrice         float64   `json:"unit_price" gorm:"column:unit_price"`
	ExchangePrice     float64   `json:"exchange_price" gorm:"column:exchange_price"`
	TotalAmount       float64   `json:"total_amount" gorm:"column:total_amount"`
	Remark            string    `json:"remark" gorm:"column:remark"`
	Status            string    `json:"status" gorm:"column:status"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
	ApprovedAt        time.Time `json:"approved_at"`
	ApprovedBy        string    `json:"approved_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         string    `json:"updated_by"`
	CryptoCodeTo      string    `json:"crypto_code_to" gorm:"column:crypto_code_to"`
	CryptoNameFrom    string    `json:"crypto_name_from" gorm:"column:crypto_name_from"`
	CryptoNameTo      string    `json:"crypto_name_to" gorm:"column:crypto_name_to"`
	BuyUnitPrice      float64   `json:"buy_unit_price" gorm:"column:buy_unit_price"`
}

// GetTradingBuyFn get trading_match with dynamic condition
func GetTradingMatchFn(arrCond []WhereCondFn, debug bool) ([]*TradingMatchStruct, error) {
	var result []*TradingMatchStruct
	tx := db.Table("trading_match").
		Joins("INNER JOIN trading_buy ON trading_match.buy_id = trading_buy.id").
		Joins("INNER JOIN trading_sell ON trading_match.sell_id = trading_sell.id").
		Joins("INNER JOIN ewt_setup ewt_to ON trading_sell.crypto_code_to = ewt_to.ewallet_type_code AND ewt_to.status = 'A'").
		Joins("INNER JOIN ewt_setup ewt_from ON trading_match.crypto_code = ewt_from.ewallet_type_code AND ewt_from.status = 'A'").
		Joins("INNER JOIN ent_member buyer ON trading_match.buyer_member_id = buyer.id").
		Joins("INNER JOIN ent_member seller ON trading_match.seller_member_id = seller.id").
		Select("trading_match.*, buyer.nick_name AS 'buyer_username', seller.nick_name AS 'seller_username', ewt_from.ewallet_type_name AS 'crypto_name_from', ewt_to.ewallet_type_code AS 'crypto_code_to', ewt_to.ewallet_type_name AS 'crypto_name_to', trading_buy.unit_price AS 'buy_unit_price'").
		Order("trading_match.created_at DESC")

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

// GetTradingMatchPaginateFn get trading_match with dynamic condition
func GetTradingMatchPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*TradingMatchStruct, error) {
	var (
		result                []*TradingMatchStruct
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("trading_match").
		Joins("INNER JOIN trading_buy ON trading_match.buy_id = trading_buy.id").
		Joins("INNER JOIN trading_sell ON trading_match.sell_id = trading_sell.id").
		Joins("INNER JOIN ewt_setup ewt_to ON trading_sell.crypto_code_to = ewt_to.ewallet_type_code").
		Joins("INNER JOIN ewt_setup ewt_from ON trading_match.crypto_code = ewt_from.ewallet_type_code").
		Joins("INNER JOIN ent_member buyer ON trading_match.buyer_member_id = buyer.id").
		Joins("INNER JOIN ent_member seller ON trading_match.seller_member_id = seller.id").
		Select("trading_match.*, buyer.nick_name AS 'buyer_username', seller.nick_name AS 'seller_username', ewt_from.ewallet_type_name AS 'crypto_name_from', ewt_to.ewallet_type_code AS 'crypto_code_to', ewt_to.ewallet_type_name AS 'crypto_name_to', trading_buy.unit_price AS 'buy_unit_price'").
		Order("trading_match.created_at DESC")

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

// AddTradingMatch func
func AddTradingMatch(tx *gorm.DB, arrData TradingMatch) (*TradingMatch, error) {
	if err := tx.Table("trading_match").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

type TradingDetails struct {
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	ActionType string    `json:"action_type" gorm:"column:action_type"`
	UnitPrice  float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalUnit  float64   `json:"total_unit" gorm:"column:total_unit"`
}

// GetTradingDetailsListFn func
func GetTradingDetailsListFn(arrUnionCond map[string][]ArrUnionRawCondText, debug bool) ([]*TradingDetails, error) {

	var result []*TradingDetails

	tradMatchSelectSQL := "SELECT " +
		"'MATCH' AS 'action_type', " +
		"trading_match.unit_price AS 'unit_price',  " +
		"trading_match.created_at AS 'created_at', " +
		"trading_match.total_unit AS 'total_unit' "
	tradMatchTblSQL := " FROM trading_match " +
		"WHERE trading_match.status IN ('AP','M') "

	tradMatchWhereSql := ""

	if len(arrUnionCond["trading_match"]) > 0 {
		for _, tradV := range arrUnionCond["trading_match"] {
			tradMatchWhereSql = tradMatchWhereSql + " " + tradV.Cond
		}
	}

	tradMatchSQL := tradMatchSelectSQL + tradMatchTblSQL + tradMatchWhereSql

	tradCancelSelectSQL := "SELECT " +
		"'CANCEL' AS 'action_type', " +
		"trading_cancel.unit_price AS 'unit_price', " +
		"trading_cancel.created_at AS 'created_at', " +
		"trading_cancel.total_unit AS 'total_unit' "

	tradCancelTblSQL := "FROM trading_cancel " +
		"WHERE trading_cancel.status = 'AP'"

	tradCancelWhereSql := ""

	if len(arrUnionCond["trading_cancel"]) > 0 {
		for _, tradV := range arrUnionCond["trading_cancel"] {
			tradCancelWhereSql = tradCancelWhereSql + " " + tradV.Cond
		}
	}

	tradCancelSQL := tradCancelSelectSQL + tradCancelTblSQL + tradCancelWhereSql

	tradSql := tradMatchSQL + " UNION " + tradCancelSQL + " ORDER BY created_at DESC"

	tx := db.Raw(tradSql)

	if debug {
		tx = tx.Debug()
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// TradingMatchMarketDetails struct
type TradingMatchMarketDetails struct {
	Volume float64 `json:"total_matched_unit" gorm:"column:total_matched_unit"`
}

// GetTradingMatchMarketDetails func
func GetTradingMatchMarketDetails(arrCond []WhereCondFn, debug bool) (*TradingMatchMarketDetails, error) {
	var result TradingMatchMarketDetails
	tx := db.Table("trading_match").
		Select("SUM(trading_match.total_unit) AS 'total_matched_unit'").
		Group("trading_match.crypto_code")

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

type TotalTradingMatchStruct struct {
	TotalAmount float64 `json:"total_amount" gorm:"column:total_amount"`
	TotalUnit   float64 `json:"total_unit" gorm:"column:total_unit"`
}

// GetTotalTradingMatchFn func
func GetTotalTradingMatchFn(arrCond []WhereCondFn, debug bool) (*TotalTradingMatchStruct, error) {
	var result TotalTradingMatchStruct
	tx := db.Table("trading_match").
		Select("SUM(trading_match.total_amount) AS 'total_amount', SUM(trading_match.total_unit) AS 'total_unit'")

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
