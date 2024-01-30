package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TradingSell struct
type TradingSell struct {
	ID                 int       `gorm:"primary_key" json:"id"`
	DocNo              string    `json:"doc_no" gorm:"column:doc_no"`
	CryptoCode         string    `json:"crypto_code" gorm:"column:crypto_code"`
	CryptoCodeTo       string    `json:"crypto_code_to" gorm:"column:crypto_code_to"`
	MemberID           int       `json:"member_id" gorm:"column:member_id"`
	SuggestedUnitPrice float64   `json:"suggested_unit_price" gorm:"column:suggested_unit_price"`
	UnitPrice          float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalUnit          float64   `json:"total_unit" gorm:"column:total_unit"`
	TotalAmount        float64   `json:"total_amount" gorm:"column:total_amount"`
	BalanceUnit        float64   `json:"balance_unit" gorm:"column:balance_unit"`
	Remark             string    `json:"remark" gorm:"column:remark"`
	Status             string    `json:"status" gorm:"column:status"`
	SigningKey         string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash          string    `json:"trans_hash" gorm:"column:trans_hash"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          string    `json:"created_by"`
	ApprovedAt         time.Time `json:"approved_at"`
	ApprovedBy         string    `json:"approved_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          string    `json:"updated_by"`
}

// GetTradingSellFn get ent_member_crypto with dynamic condition
func GetTradingSellFn(arrCond []WhereCondFn, debug bool) ([]*TradingSell, error) {
	var result []*TradingSell
	tx := db.Table("trading_sell").
		Joins("INNER JOIN ent_member ON trading_sell.member_id = ent_member.id").
		Select("trading_sell.*, ent_member.nick_name").
		Order("trading_sell.created_at DESC")

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

// TradingSellStruct struct
type TradingSellStruct struct {
	ID             int     `gorm:"primary_key" json:"id"`
	CryptoCode     string  `json:"crypto_code" gorm:"column:crypto_code"`
	CryptoFromName string  `json:"crypto_from_name" gorm:"column:crypto_from_name"`
	CryptoToName   string  `json:"crypto_to_name" gorm:"column:crypto_to_name"`
	MemberID       int     `json:"member_id" gorm:"column:member_id"`
	NickName       string  `json:"nick_name" gorm:"column:nick_name"`
	UnitPrice      float64 `json:"unit_price" gorm:"column:unit_price"`
	//UnitPrice              string `json:"unit_price" gorm:"column:unit_price"`
	TotalUnit              float64   `json:"total_unit" gorm:"column:total_unit"`
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

// GetTradingSellPaginateFn get trading_sell with dynamic condition
func GetTradingSellPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*TradingSellStruct, error) {
	var (
		result                []*TradingSellStruct
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("trading_sell").
		Joins("INNER JOIN ent_member ON trading_sell.member_id = ent_member.id").
		Joins("INNER JOIN ewt_setup crypto_from ON trading_sell.crypto_code = crypto_from.ewallet_type_code AND crypto_from.status = 'A'").
		Joins("INNER JOIN ewt_setup crypto_to ON trading_sell.crypto_code_to = crypto_to.ewallet_type_code AND crypto_to.status = 'A'").
		Select("trading_sell.*, ent_member.nick_name, crypto_from.ewallet_type_name AS 'crypto_from_name', crypto_to.ewallet_type_name AS 'crypto_to_name', crypto_from.decimal_point AS 'crypto_from_decimal_point', crypto_to.decimal_point AS 'crypto_to_decimal_point'").
		Order("trading_sell.created_at DESC")

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
	err := tx.Limit(limit).Offset(newOffset).Scan(&result).Error
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

// AddTradingSell func
func AddTradingSell(tx *gorm.DB, arrData TradingSell) (*TradingSell, error) {
	if err := tx.Table("trading_sell").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// AvailableTradingSellListStruct struct
type AvailableTradingSellListStruct struct {
	UnitPrice        float64 `json:"unit_price" gorm:"column:unit_price"`
	TotalBalanceUnit float64 `json:"total_balance_unit" gorm:"column:total_balance_unit"`
}

// GetAvailableTradingSellListFn get trading_sell with dynamic condition
func GetAvailableTradingSellListFn(arrCond []WhereCondFn, limit uint, debug bool) ([]*AvailableTradingSellListStruct, error) {
	var (
		result []*AvailableTradingSellListStruct
	)
	tx := db.Table("trading_sell").
		Select("trading_sell.unit_price, SUM(trading_sell.balance_unit) AS 'total_balance_unit'").
		Group("trading_sell.crypto_code, trading_sell.unit_price").
		Where("trading_sell.status = 'P'").
		Order("trading_sell.unit_price DESC")

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

type AutoMatchTrading struct {
	ID           int       `gorm:"primary_key" json:"id"`
	DocNo        string    `json:"doc_no" gorm:"column:doc_no"`
	CryptoCode   string    `json:"crypto_code" gorm:"column:crypto_code"`
	CryptoCodeTo string    `json:"crypto_code_to" gorm:"column:crypto_code_to"`
	MemberID     int       `json:"member_id" gorm:"column:member_id"`
	UnitPrice    float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalUnit    float64   `json:"total_unit" gorm:"column:total_unit"`
	TotalAmount  float64   `json:"total_amount" gorm:"column:total_amount"`
	BalanceUnit  float64   `json:"balance_unit" gorm:"column:balance_unit"`
	Remark       string    `json:"remark" gorm:"column:remark"`
	Status       string    `json:"status" gorm:"column:status"`
	SigningKey   string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash    string    `json:"trans_hash" gorm:"column:trans_hash"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	ApprovedAt   time.Time `json:"approved_at"`
	ApprovedBy   string    `json:"approved_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    string    `json:"updated_by"`
}

// GetAutoMatchTradingSellListFn get trading_sell with dynamic condition
func GetAutoMatchTradingSellListFn(arrCond []WhereCondFn, debug bool) ([]*AutoMatchTrading, error) {
	var result []*AutoMatchTrading
	tx := db.Table("trading_sell").
		Joins("INNER JOIN ent_member ON trading_sell.member_id = ent_member.id").
		Select("trading_sell.*, ent_member.nick_name").
		Order("trading_sell.created_at ASC")

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

type TradingBuySellPaginate struct {
	ID             int       `json:"id" gorm:"column:id"`
	ActionType     string    `json:"action_type" gorm:"column:action_type"`
	CryptoCode     string    `json:"crypto_code" gorm:"column:crypto_code"`
	CryptoCodeTo   string    `json:"crypto_code_to" gorm:"column:crypto_code_to"`
	DocNo          string    `json:"doc_no" gorm:"column:doc_no"`
	MemberID       int       `json:"member_id" gorm:"column:member_id"`
	TotalUnit      float64   `json:"total_unit" gorm:"column:total_unit"`
	UnitPrice      float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalAmount    float64   `json:"total_amount" gorm:"column:total_amount"`
	BalanceUnit    float64   `json:"balance_unit" gorm:"column:balance_unit"`
	Remark         string    `json:"remark" gorm:"column:remark"`
	Status         string    `json:"status" gorm:"column:status"`
	SigningKey     string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash      string    `json:"trans_hash" gorm:"column:trans_hash"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	CreatedBy      string    `json:"created_by" gorm:"column:created_by"`
	ApprovedAt     time.Time `json:"approved_at" gorm:"column:approved_at"`
	ApprovedBy     string    `json:"approved_by" gorm:"column:approved_by"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy      string    `json:"updated_by" gorm:"column:updated_by"`
	NickName       string    `json:"nick_name" gorm:"column:updated_by"`
	CryptoFromName string    `json:"crypto_from_name" gorm:"column:crypto_from_name"`
	CryptoToName   string    `json:"crypto_to_name" gorm:"column:crypto_to_name"`
}

type TotalQueryRecords struct {
	TotalRecords int64 `json:"total_record" gorm:"column:total_record"`
}

// GetTradingBuySellPaginateFn get trading_sell with dynamic condition
func GetTradingBuySellPaginateFn(arrUnionCond map[string][]ArrUnionRawCondText, page int64, debug bool) (SQLPaginateStdReturn, []*TradingBuySellPaginate, error) {
	var (
		result                []*TradingBuySellPaginate
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
		totalTradSellRecord   TotalQueryRecords
		totalTradBuyRecord    TotalQueryRecords
	)

	tradSellSelectSQL := "SELECT " +
		"trading_sell.id AS 'id', " +
		"'SELL' AS 'action_type', " +
		"trading_sell.crypto_code AS 'crypto_code',  " +
		"trading_sell.crypto_code_to AS 'crypto_code_to', " +
		"trading_sell.doc_no AS 'doc_no', " +
		"trading_sell.member_id AS 'member_id', " +
		"trading_sell.total_unit AS 'total_unit', " +
		"trading_sell.unit_price AS 'unit_price', " +
		"trading_sell.total_amount AS 'total_amount', " +
		"trading_sell.balance_unit AS 'balance_unit', " +
		"trading_sell.remark AS 'remark', " +
		"trading_sell.signing_key AS 'signing_key', " +
		"trading_sell.trans_hash AS 'trans_hash', " +
		"trading_sell.status AS 'status', " +
		"trading_sell.created_at AS 'created_at', " +
		"trading_sell.created_by AS 'created_by', " +
		"trading_sell.approved_at AS 'approved_at', " +
		"trading_sell.approved_by AS 'approved_by', " +
		"trading_sell.updated_at AS 'updated_at', " +
		"trading_sell.updated_by AS 'updated_by', " +
		"ent_member.nick_name,  " +
		"crypto_from.ewallet_type_name AS 'crypto_from_name',  " +
		"crypto_to.ewallet_type_name AS 'crypto_to_name' "

	tradSellTblSQL := "FROM trading_sell " +
		"INNER JOIN ent_member ON trading_sell.member_id = ent_member.id " +
		"INNER JOIN ewt_setup crypto_from ON trading_sell.crypto_code = crypto_from.ewallet_type_code AND crypto_from.status = 'A' " +
		"INNER JOIN ewt_setup crypto_to ON trading_sell.crypto_code_to = crypto_to.ewallet_type_code AND crypto_to.status = 'A' " +
		"WHERE 1=1 "

	tradSellWhereSql := ""

	if len(arrUnionCond["trading_sell"]) > 0 {
		for _, tradV := range arrUnionCond["trading_sell"] {
			tradSellWhereSql = tradSellWhereSql + " " + tradV.Cond
		}
	}

	tradSellSQL := tradSellSelectSQL + tradSellTblSQL + tradSellWhereSql

	tradBuySelectSQL := "SELECT " +
		"trading_buy.id AS 'id', " +
		"'BUY' AS 'action_type', " +
		"trading_buy.crypto_code AS 'crypto_code', " +
		"trading_buy.crypto_code_to AS 'crypto_code_to', " +
		"trading_buy.doc_no AS 'doc_no', " +
		"trading_buy.member_id AS 'member_id', " +
		"trading_buy.total_unit AS 'total_unit', " +
		"trading_buy.unit_price AS 'unit_price', " +
		"trading_buy.total_amount AS 'total_amount', " +
		"trading_buy.balance_unit AS 'balance_unit', " +
		"trading_buy.remark AS 'remark', " +
		"trading_buy.signing_key AS 'signing_key', " +
		"trading_buy.trans_hash AS 'trans_hash', " +
		"trading_buy.status AS 'status', " +
		"trading_buy.created_at AS 'created_at', " +
		"trading_buy.created_by AS 'created_by', " +
		"trading_buy.approved_at AS 'approved_at', " +
		"trading_buy.approved_by AS 'approved_by', " +
		"trading_buy.updated_at AS 'updated_at', " +
		"trading_buy.updated_by AS 'updated_by', " +
		"ent_member.nick_name AS 'nick_name', " +
		"crypto_from.ewallet_type_name AS 'crypto_from_name', " +
		"crypto_to.ewallet_type_name AS 'crypto_to_name' "

	tradBuyTblSQL := "FROM trading_buy " +
		"INNER JOIN ent_member ON trading_buy.member_id = ent_member.id " +
		"INNER JOIN ewt_setup crypto_from ON trading_buy.crypto_code = crypto_from.ewallet_type_code AND crypto_from.status = 'A' " +
		"INNER JOIN ewt_setup crypto_to ON trading_buy.crypto_code_to = crypto_to.ewallet_type_code AND crypto_to.status = 'A' " +
		"WHERE 1=1"

	tradBuyWhereSql := ""

	if len(arrUnionCond["trading_buy"]) > 0 {
		for _, tradV := range arrUnionCond["trading_buy"] {
			tradBuyWhereSql = tradBuyWhereSql + " " + tradV.Cond
		}
	}

	tradBuySQL := tradBuySelectSQL + tradBuyTblSQL + tradBuyWhereSql

	tradSql := tradSellSQL + " UNION " + tradBuySQL + " ORDER BY created_at DESC"

	tradSellCountSQL := "SELECT COUNT(*) AS 'total_record' " + tradSellTblSQL + tradSellWhereSql
	tx1 := db.Raw(tradSellCountSQL)
	err := tx1.Scan(&totalTradSellRecord).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	tradBuyCountSQL := "SELECT COUNT(*) AS 'total_record' " + tradBuyTblSQL + tradBuyWhereSql
	tx2 := db.Raw(tradBuyCountSQL)
	err = tx2.Scan(&totalTradBuyRecord).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalRecord = totalTradSellRecord.TotalRecords + totalTradBuyRecord.TotalRecords

	tx := db.Raw(tradSql)

	if debug {
		tx1 = tx1.Debug()
		tx = tx.Debug()
	}

	arrLimitRowsSetting, _ := GetSysGeneralSetupByID("defaultlimitrow")
	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	// Total Records
	// tx.Count(&totalRecord)
	oriPage := page
	if page != 0 {
		page--
	}

	// newOffset := page * limit
	// Pagination and limit
	// fmt.Println("limit:", limit)
	// fmt.Println("newOffset:", newOffset)
	err = tx.Find(&result).Error
	// err = tx.Limit(limit).Offset(newOffset).Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	processArr := make([]*TradingBuySellPaginate, 0)
	if len(result) > 0 {
		pageStart, pageEnd := paginate(int(page), int(limit), int(totalRecord))
		processArr = result[pageStart:pageEnd]
	}

	perPage = limit

	// totalCurrentPageItems = int64(len(result))
	totalCurrentPageItems = int64(len(processArr))

	arrPaginateData = SQLPaginateStdReturn{
		CurrentPage:           oriPage,
		PerPage:               perPage,
		TotalCurrentPageItems: totalCurrentPageItems,
		TotalPage:             totalPage,
		TotalPageItems:        totalRecord,
	}
	return arrPaginateData, processArr, nil
}

// paginate function
func paginate(pageNum int, pageSize int, sliceLength int) (int, int) {
	start := pageNum * pageSize

	if start > sliceLength {
		start = sliceLength
	}

	end := start + pageSize
	if end > sliceLength {
		end = sliceLength
	}

	return start, end
}

type AutoMatchTradingSellListV2 struct {
	Limit   uint8
	OrderBy string
}

// GetAutoMatchTradingSellListFnV2 get trading_sell with dynamic condition
func GetAutoMatchTradingSellListFnV2(arrCond []WhereCondFn, arrData AutoMatchTradingSellListV2, debug bool) ([]*AutoMatchTrading, error) {

	limit := arrData.Limit
	orderBy := arrData.OrderBy

	var result []*AutoMatchTrading
	tx := db.Table("trading_sell").
		Joins("INNER JOIN ent_member ON trading_sell.member_id = ent_member.id").
		Select("trading_sell.*, ent_member.nick_name")

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
