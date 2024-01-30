package models

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EwtDetail struct
type EwtDetail struct {
	ID                int       `gorm:"primary_key" json:"id"`
	MemberID          int       `json:"member_id"`
	EwalletTypeID     int       `json:"ewallet_type_id"`
	CurrencyCode      string    `json:"currency_code"`
	TransactionType   string    `json:"transaction_type"`
	TransDate         time.Time `json:"trans_date"`
	TotalIn           float64   `json:"total_in"`
	TotalOut          float64   `json:"total_out"`
	ConversionRate    float64   `json:"conversion_rate"`
	ConvertedTotalIn  float64   `json:"converted_total_in"`
	ConvertedTotalOut float64   `json:"converted_total_out"`
	Balance           float64   `json:"balance"`
	DocNo             string    `json:"doc_no"`
	AdditionalMsg     string    `json:"additional_msg"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         string    `json:"updated_by"`
}

// EwtDetailTransaction struct

type EwtDetailTransaction struct {
	ID              int    `gorm:"primary_key" json:"id"`
	MemberID        int    `json:"member_id"`
	EwalletTypeID   int    `json:"ewallet_type_id"`
	EwalletTypeName string `json:"ewallet_type_name"`
	CurrencyCode    string `json:"currency_code"`
	TransactionType string `json:"transaction_type"`
	// TransType       string    `json:"trans_type"`
	Type          string    `json:"type"` // receive /transfer
	TransDate     time.Time `json:"trans_date"`
	TotalIn       float64   `json:"total_in"`
	TotalOut      float64   `json:"total_out"`
	Balance       float64   `json:"balance"`
	DocNo         string    `json:"doc_no"`
	AdditionalMsg string    `json:"additional_msg"`
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     string    `json:"updated_by"`
}

type EwtDetailWithSetup struct {
	ID                int       `gorm:"primary_key" json:"id"`
	MemberID          int       `json:"member_id"`
	EwalletTypeID     int       `json:"ewallet_type_id"`
	CurrencyCode      string    `json:"currency_code"`
	DecimalPoint      int       `json:"decimal_point"`
	TransactionType   string    `json:"transaction_type"`
	TransDate         time.Time `json:"trans_date"`
	TotalIn           float64   `json:"total_in"`
	TotalOut          float64   `json:"total_out"`
	ConversionRate    float64   `json:"conversion_rate"`
	ConvertedTotalIn  float64   `json:"converted_total_in"`
	ConvertedTotalOut float64   `json:"converted_total_out"`
	Balance           float64   `json:"balance"`
	DocNo             string    `json:"doc_no"`
	AdditionalMsg     string    `json:"additional_msg"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         string    `json:"updated_by"`
	EwalletTypeName   string    `json:"ewallet_type_name"`
}

type MaxSummaryDetail struct {
	ID int `json:"id"`
}

// GetEwtDetailFn get ewt_detail data with dynamic condition
func GetEwtDetailFn(arrCond []WhereCondFn, debug bool) ([]*EwtDetail, error) {
	var result []*EwtDetail
	tx := db.Table("ewt_detail")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

func GetEwtDetailWithSetup(arrCond []WhereCondFn, debug bool) ([]*EwtDetailWithSetup, error) {
	var result []*EwtDetailWithSetup
	tx := db.Table("ewt_detail").
		Select("ewt_setup.currency_code,ewt_detail.*, ewt_setup.decimal_point,ewt_setup.ewallet_type_name").
		Joins("inner join ewt_setup ON ewt_detail.ewallet_type_id = ewt_setup.id")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Limit(400).Order("ewt_detail.id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// func AddEwtDetail add ewt_detail records`
func AddEwtDetail(tx *gorm.DB, saveData EwtDetail) (*EwtDetail, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtDetail-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

// func AddEwtDetailWithoutTx add ewt_detail records`
func AddEwtDetailWithoutTx(saveData EwtDetail) (*EwtDetail, error) {
	if err := db.Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtDetail-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

//get ewtdetail for transaction list with paginate
func GetEwtDetailForTransactionList(page int64, mem_id int, transType string, dateFrom string, dateTo string) ([]*EwtDetailTransaction, int64, int64, int64, float64, error) {
	var (
		ewt                   []*EwtDetailTransaction
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		// endFlag               int64 = page
	)

	//general setup default limit rows
	arrLimitRowsSetting, _ := GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)
	query := db.Table("ewt_detail a").
		// Select("a.*,b.ewallet_type_name").
		Select("a.id,a.member_id,a.ewallet_type_id,a.trans_date,a.total_in,a.total_out,a.balance,a.doc_no,a.created_at,a.created_by,b.ewallet_type_name, CASE WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'GAME_RESULT%' THEN 'REWARD' WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'DEPOSIT%' THEN 'TOPUP' WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'WITHDRAW%' THEN 'WITHDRAW' WHEN a.transaction_type = 'MIGRATE' AND (a.remark LIKE 'TRANSFER_TO%' OR a.remark LIKE 'TRANSFER_FROM%') THEN 'TRANSFER' WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'GAME_TICKET%' THEN 'GAME_PAYMENT'  WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'MIGRATE%'AND a.total_in >0  THEN CONCAT(b.ewallet_type_name,' IN') WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'MIGRATE%'AND a.total_out >0  THEN CONCAT(b.ewallet_type_name,' OUT') ELSE a.transaction_type END AS 'transaction_type', CASE WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'GAME_RESULT%' THEN '#*GAME_RESULT*#' WHEN a.transaction_type = 'MIGRATE' AND a.remark LIKE 'MIGRATE%' THEN '#*successful*#' ELSE a.remark END AS 'remark'").
		Joins("inner join ewt_setup b ON a.ewallet_type_id = b.id")

	if transType != "" {
		query = query.Where("a.transaction_type = ?", transType)
	}

	if dateFrom != "" {
		query = query.Where("date(a.trans_date) >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("date(a.trans_date) <= ?", dateTo)
	}

	query = query.Where("a.member_id = ?", mem_id)

	// Total Records
	query.Count(&totalRecord)

	if page != 0 {
		page--
	}

	newOffset := page * limit

	// Pagination and limit
	err := query.Order("id desc").Limit(limit).Offset(newOffset).Find(&ewt).Error
	if err != nil {
		return nil, 0, 0, 0, 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	// if int64(endFlag) >= int64(totalPage) {
	// 	endFlag = 1
	// } else {
	// 	endFlag = 0
	// }

	perPage = limit

	totalCurrentPageItems = int64(len(ewt))

	return ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, nil
}

//get ewtdetail for wallet statement v2 list with paginate
func GetEwtDetailByWalletTypeForStatementList(page int64, mem_id int, transType string, dateFrom string, dateTo string, WalletTypeCode string) ([]*EwtDetailTransaction, int64, int64, int64, float64, error) {
	var (
		ewt                   []*EwtDetailTransaction
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
	)

	//general setup default limit rows
	arrLimitRowsSetting, _ := GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)
	query := db.Table("ewt_detail a").
		Select("a.id,a.member_id,a.transaction_type,a.ewallet_type_id,a.trans_date,a.total_in,a.total_out,a.balance,a.doc_no,a.created_at,a.created_by,b.ewallet_type_name, CASE WHEN a.total_in > 0 THEN 'receive' WHEN a.total_out > 0 THEN 'transfer' END AS 'type'").
		Joins("inner join ewt_setup b ON a.ewallet_type_id = b.id").
		Where("b.ewallet_type_code = ?", WalletTypeCode)

	// if transType != "" {
	// 	var transType2 = strings.Split(transType, ",")
	// 	data := []string{}

	// 	for _, v := range transType2 {
	// 		if v == "OTHERS" {
	// 			data = append(data, "ADJUSTMENT")
	// 			data = append(data, "REFUND")
	// 		} else if v == "BONUS" {
	// 			data = append(data, v)
	// 			data = append(data, "REWARD")
	// 		} else {
	// 			data = append(data, v)
	// 		}
	// 	}
	// 	query = query.Where("a.transaction_type IN (?)", data)
	// }

	if transType != "" {
		if transType == strings.ToLower("in") {
			query = query.Where("a.total_in > 0")
		} else if transType == strings.ToLower("out") {
			query = query.Where("a.total_out > 0")
		}

	}

	if dateFrom != "" {
		query = query.Where("date(a.trans_date) >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("date(a.trans_date) <= ?", dateTo)
	}

	query = query.Where("a.member_id = ?", mem_id)

	// Total Records
	query.Count(&totalRecord)

	if page != 0 {
		page--
	}

	newOffset := page * limit

	// Pagination and limit
	err := query.Order("id asc").Limit(limit).Offset(newOffset).Find(&ewt).Error
	if err != nil {
		return nil, 0, 0, 0, 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	perPage = limit

	totalCurrentPageItems = int64(len(ewt))

	return ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, nil
}

func GetEwtDetailByWalletTypeForSummaryDetail(page int64, mem_id int, transType string, dateFrom string, dateTo string, WalletTypeCode string) ([]*MaxSummaryDetail, int64, int64, int64, float64, error) {
	var (
		ewt                   []*MaxSummaryDetail
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
	)

	//general setup default limit rows
	arrLimitRowsSetting, _ := GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)
	query := db.Table("ewt_detail a").
		Select("Max(a.id) as id").
		Joins("inner join ewt_setup b ON a.ewallet_type_id = b.id").
		Where("b.ewallet_type_code = ?", WalletTypeCode)

	if transType != "" {
		if transType == strings.ToLower("in") {
			query = query.Where("a.total_in > 0")
		} else if transType == strings.ToLower("out") {
			query = query.Where("a.total_out > 0")
		}

	}

	if dateFrom != "" {
		query = query.Where("date(a.trans_date) >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("date(a.trans_date) <= ?", dateTo)
	}

	query = query.Where("a.member_id = ?", mem_id).Group("date(a.trans_date)")

	// Total Records
	query.Count(&totalRecord)

	if page != 0 {
		page--
	}

	newOffset := page * limit

	// Pagination and limit
	err := query.Order("id asc").Limit(limit).Offset(newOffset).Find(&ewt).Error
	if err != nil {
		return nil, 0, 0, 0, 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	perPage = limit

	totalCurrentPageItems = int64(len(ewt))

	return ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, nil
}

type EwtDetailStrategyWithSetup struct {
	ID                int       `gorm:"primary_key" json:"id"`
	MemberID          int       `json:"member_id"`
	EwalletTypeID     int       `json:"ewallet_type_id"`
	CurrencyCode      string    `json:"currency_code"`
	DecimalPoint      int       `json:"decimal_point"`
	TransactionType   string    `json:"transaction_type"`
	TransDate         time.Time `json:"trans_date"`
	TotalIn           float64   `json:"total_in"`
	TotalOut          float64   `json:"total_out"`
	ConversionRate    float64   `json:"conversion_rate"`
	ConvertedTotalIn  float64   `json:"converted_total_in"`
	ConvertedTotalOut float64   `json:"converted_total_out"`
	Balance           float64   `json:"balance"`
	DocNo             string    `json:"doc_no"`
	AdditionalMsg     string    `json:"additional_msg"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         string    `json:"updated_by"`
	EwalletName       string    `json:"ewallet_name"`
}

func GetEwtDetailStrategyWithSetup(arrCond []WhereCondFn, debug bool) ([]*EwtDetailStrategyWithSetup, error) {
	var result []*EwtDetailStrategyWithSetup
	tx := db.Table("ewt_detail a").
		Select("b.currency_code,a.*, b.decimal_point,b.ewallet_type_name as ewallet_name").
		Joins("inner join ewt_setup b ON a.ewallet_type_id = b.id")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("a.id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
