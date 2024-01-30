package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// BlockchainTransStruct struct
type BlockchainTransStruct struct {
	ID                int     `gorm:"primary_key" json:"id"`
	MemberID          int     `json:"member_id"`
	EwalletTypeID     int     `json:"ewallet_type_id"`
	DocNo             string  `json:"doc_no"`
	Status            string  `json:"status"`
	TransactionType   string  `json:"transaction_type"`
	TotalIn           float64 `json:"total_in"`
	TotalOut          float64 `json:"total_out"`
	ConversionRate    float64 `json:"conversion_rate"`
	ConvertedTotalIn  float64 `json:"converted_total_in"`
	ConvertedTotalOut float64 `json:"converted_total_out"`
	TransactionData   string  `json:"transaction_data"`
	HashValue         string  `json:"hash_value"`
	LogOnly           int     `json:"log_only"`
	Remark            string  `json:"remark"`
}

type BlockchainTransListStruct struct {
	ID                int       `gorm:"primary_key" json:"id"`
	MemberID          int       `json:"member_id"`
	EwalletTypeID     int       `json:"ewallet_type_id"`
	DocNo             string    `json:"doc_no"`
	Status            string    `json:"status"`
	TransactionType   string    `json:"transaction_type"`
	TotalIn           float64   `json:"total_in"`
	TotalOut          float64   `json:"total_out"`
	ConversionRate    float64   `json:"conversion_rate"`
	ConvertedTotalIn  float64   `json:"converted_total_in"`
	ConvertedTotalOut float64   `json:"converted_total_out"`
	TransactionData   string    `json:"transaction_data"`
	HashValue         string    `json:"hash_value"`
	LogOnly           int       `json:"log_only"`
	Remark            string    `json:"remark"`
	DtTimestamp       time.Time `json:"dt_timestamp"`
}

type BlockchainPendingAmount struct {
	TotalPendingAmount float64 `json:"total_pending_amount"`
}

// GetBlockchainTrans func
func GetBlockchainTrans(arrCond []WhereCondFn, selectColumn string, debug bool) (*BlockchainTransStruct, error) {
	var blockchainTrans BlockchainTransStruct
	tx := db.Table("blockchain_trans")
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
	err := tx.Find(&blockchainTrans).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if blockchainTrans.ID <= 0 {
		return nil, nil
	}

	return &blockchainTrans, nil
}

// AddBlockchainTransStruct struct
type AddBlockchainTransStruct struct {
	ID                int     `gorm:"primary_key" json:"id"`
	MemberID          int     `json:"member_id"`
	EwalletTypeID     int     `json:"ewallet_type_id"`
	DocNo             string  `json:"doc_no"`
	Status            string  `json:"status"`
	TransactionType   string  `json:"transaction_type"`
	TotalIn           float64 `json:"total_in"`
	TotalOut          float64 `json:"total_out"`
	ConversionRate    float64 `json:"conversion_rate"`
	ConvertedTotalIn  float64 `json:"converted_total_in"`
	ConvertedTotalOut float64 `json:"converted_total_out"`
	TransactionData   string  `json:"transaction_data"`
	HashValue         string  `json:"hash_value"`
	LogOnly           int     `json:"log_only"`
	Remark            string  `json:"remark"`
}

// AddBlockchainTrans func
func AddBlockchainTrans(tx *gorm.DB, blockchainTrans AddBlockchainTransStruct) (*AddBlockchainTransStruct, error) {
	if err := tx.Table("blockchain_trans").Create(&blockchainTrans).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &blockchainTrans, nil
}

func GetBlockchainTransArrayFn(arrCond []WhereCondFn, debug bool) ([]*BlockchainTransListStruct, error) {
	var result []*BlockchainTransListStruct
	tx := db.Table("blockchain_trans")
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

func GetTotalPendingBlockchainAmount(memID int, ewalletTypeCode string) (*BlockchainPendingAmount, error) {
	var result BlockchainPendingAmount

	query := db.Table("blockchain_trans a").
		Select("SUM(a.total_out) as total_pending_amount").
		Joins("inner join ewt_setup b ON a.ewallet_type_id = b.id").
		Where("a.member_id = ?", memID).
		Where("b.ewallet_type_code = ?", ewalletTypeCode).
		Where("a.status = ?", "P").
		Where("a.log_only = ?", 0)

	err := query.Find(&result).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

// BlockchainTransStruct struct
type AddBlockchainTransV2Struct struct {
	ID                int     `gorm:"primary_key" json:"id"`
	MemberID          int     `json:"member_id"`
	EwalletTypeID     int     `json:"ewallet_type_id"`
	DocNo             string  `json:"doc_no"`
	Status            string  `json:"status"`
	TransactionType   string  `json:"transaction_type"`
	TotalIn           float64 `json:"total_in"`
	TotalOut          float64 `json:"total_out"`
	ConversionRate    float64 `json:"conversion_rate"`
	ConvertedTotalIn  float64 `json:"converted_total_in"`
	ConvertedTotalOut float64 `json:"converted_total_out"`
	TransactionData   string  `json:"transaction_data"`
	HashValue         string  `json:"hash_value"`
	LogOnly           int     `json:"log_only"`
	Remark            string  `json:"remark"`
	DtTimestamp       string  `json:"dt_timestamp"`
}

// AddBlockchainTrans func
func AddBlockchainTransV2(tx *gorm.DB, blockchainTrans AddBlockchainTransV2Struct) (*AddBlockchainTransV2Struct, error) {
	if err := tx.Table("blockchain_trans").Create(&blockchainTrans).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &blockchainTrans, nil
}
