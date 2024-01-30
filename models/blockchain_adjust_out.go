package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// BlockchainAdjustOut struct
type BlockchainAdjustOut struct {
	ID                int     `gorm:"primary_key" json:"id"`
	MemberID          int     `json:"member_id"`
	EwalletTypeID     int     `json:"ewallet_type_id"`
	EwalletTypeCode   string  `json:"ewallet_type_code"`
	Status            string  `json:"status"`
	TransactionType   string  `json:"transaction_type"`
	TotalIn           float64 `json:"total_in"`
	TotalOut          float64 `json:"total_out"`
	ConversionRate    float64 `json:"conversion_rate"`
	ConvertedTotalIn  float64 `json:"converted_total_in"`
	ConvertedTotalOut float64 `json:"converted_total_out"`
	Remark            string  `json:"remark"`
}

type BlockchainAdjustOutSum struct {
	EwalletTypeCode string  `json:"ewallet_type_code"`
	TotalOut        float64 `json:"total_out"`
	TotalIn         float64 `json:"total_in"`
	TransactionType string  `json:"transaction_type"`
	TransactionIds  string  `json:"transaction_ids"`
}

// GetBlockchainAdjustOutFn get wod_member_rank data with dynamic condition
func GetBlockchainAdjustOutFn(arrCond []WhereCondFn, debug bool) ([]*BlockchainAdjustOut, error) {
	var result []*BlockchainAdjustOut
	tx := db.Table("blockchain_adjust_out")

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

// GetBlockchainAdjustOutFn get wod_member_rank data with dynamic condition
func GetBlockchainAdjustOutSumFn(arrCond []WhereCondFn, debug bool) ([]*BlockchainAdjustOutSum, error) {
	var result []*BlockchainAdjustOutSum
	tx := db.Table("blockchain_adjust_out").Joins("JOIN ewt_setup ON ewt_setup.id = blockchain_adjust_out.ewallet_type_id").Select("SUM(blockchain_adjust_out.total_out) as total_out, SUM(blockchain_adjust_out.total_in) as total_in, ewt_setup.ewallet_type_code, blockchain_adjust_out.transaction_type, GROUP_CONCAT(blockchain_adjust_out.id SEPARATOR ',') as transaction_ids")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	tx = tx.Group("ewt_setup.ewallet_type_code, blockchain_adjust_out.transaction_type")

	if debug {
		tx = tx.Debug()
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type BlockchainPendingAdjustOutAmount struct {
	TotalPendingAmount float64 `json:"total_pending_amount"`
}

func GetTotalPendingBlockchainAdjustOutAmount(memID int, ewalletTypeCode string) (*BlockchainPendingAdjustOutAmount, error) {
	var result BlockchainPendingAdjustOutAmount

	query := db.Table("blockchain_adjust_out a").
		Select("SUM(a.total_out) as total_pending_amount").
		Joins("inner join ewt_setup b ON a.ewallet_type_id = b.id").
		Where("a.member_id = ?", memID).
		Where("b.ewallet_type_code = ?", ewalletTypeCode).
		Where("a.status = ?", "P").
		Where("a.transaction_type = ?", "ADJUST")

	err := query.Find(&result).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
