package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TradingCancel struct
type TradingCancel struct {
	ID              int       `gorm:"primary_key" json:"id"`
	TradingID       int       `json:"trading_id" gorm:"column:trading_id"`
	MemberID        int       `json:"member_id" gorm:"column:member_id"`
	DocNo           string    `json:"doc_no" gorm:"column:doc_no"`
	TransactionType string    `json:"transaction_type" gorm:"column:transaction_type"`
	CryptoCode      string    `json:"crypto_code" gorm:"column:crypto_code"`
	TotalUnit       float64   `json:"total_unit" gorm:"column:total_unit"`
	UnitPrice       float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalAmount     float64   `json:"total_amount" gorm:"column:total_amount"`
	Remark          string    `json:"remark" gorm:"column:remark"`
	SigningKey      string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash       string    `json:"trans_hash" gorm:"column:trans_hash"`
	Status          string    `json:"status" gorm:"column:status"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       string    `json:"created_by"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       string    `json:"updated_by"`
	ApprovedAt      time.Time `json:"approved_at"`
	ApprovedBy      string    `json:"approved_by"`
}

// GetTradingCancelFn get trading_cancel with dynamic condition
func GetTradingCancelFn(arrCond []WhereCondFn, debug bool) ([]*TradingCancel, error) {
	var result []*TradingCancel
	tx := db.Table("trading_cancel").
		Joins("INNER JOIN ent_member ON trading_cancel.member_id = ent_member.id").
		Select("trading_cancel.*, ent_member.nick_name").
		Order("trading_cancel.created_at DESC")

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

// AddTradingCancelStruct struct
type AddTradingCancelStruct struct {
	ID              int       `gorm:"primary_key" json:"id"`
	TradingID       int       `json:"trading_id" gorm:"column:trading_id"`
	MemberID        int       `json:"member_id" gorm:"column:member_id"`
	DocNo           string    `json:"doc_no" gorm:"column:doc_no"`
	TransactionType string    `json:"transaction_type" gorm:"column:transaction_type"`
	CryptoCode      string    `json:"crypto_code" gorm:"column:crypto_code"`
	TotalUnit       float64   `json:"total_unit" gorm:"column:total_unit"`
	UnitPrice       float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalAmount     float64   `json:"total_amount" gorm:"column:total_amount"`
	Remark          string    `json:"remark" gorm:"column:remark"`
	SigningKey      string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash       string    `json:"trans_hash" gorm:"column:trans_hash"`
	Status          string    `json:"status" gorm:"column:status"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       string    `json:"created_by"`
}

// AddTradingCancel func
func AddTradingCancel(tx *gorm.DB, arrData AddTradingCancelStruct) (*AddTradingCancelStruct, error) {
	if err := tx.Table("trading_cancel").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// TotalTradingCancel struct
type TotalTradingCancel struct {
	TotalCancelUnit float64 `json:"total_cancel_unit" gorm:"column:total_cancel_unit"`
}

// GetTotalTradingCancelFn get trading_cancel with dynamic condition
func GetTotalTradingCancelFn(arrCond []WhereCondFn, debug bool) (*TotalTradingCancel, error) {
	var result TotalTradingCancel
	tx := db.Table("trading_cancel").
		Select("SUM(trading_cancel.total_unit) AS 'total_cancel_unit'")

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
