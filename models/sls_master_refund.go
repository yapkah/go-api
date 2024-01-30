package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SlsMasterRefund struct
type SlsMasterRefund struct {
	ID                   int       `gorm:"primary_key" json:"id"`
	BatchNo              string    `json:"batch_no" gorm:"column:batch_no"`
	SlsMasterID          int       `json:"sls_master_id" gorm:"column:sls_master_id"`
	MemberID             int       `json:"member_id" gorm:"column:member_id"`
	RequestAmount        float64   `json:"request_amount" gorm:"column:request_amount"`
	RefundEwalletTypeID  int       `json:"refund_ewallet_type_id" gorm:"column:refund_ewallet_type_id"`
	RefundAmount         float64   `json:"refund_amount" gorm:"column:refund_amount"`
	PenaltyEwalletTypeID int       `json:"penalty_ewallet_type_id" gorm:"column:penalty_ewallet_type_id"`
	PenaltyPerc          int       `json:"penalty_perc" gorm:"column:penalty_perc"`
	PenaltyAmount        float64   `json:"penalty_amount" gorm:"column:penalty_amount"`
	Status               string    `json:"status" gorm:"status"`
	CreatedAt            time.Time `json:"created_at"`
	CreatedBy            string    `json:"created_by"`
	// RefundedAt           time.Time `json:"updated_at"`
	// RefundedBy           string    `json:"updated_by"`
}

// GetSlsMasterRefundFn func
func GetSlsMasterRefundFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMasterRefund, error) {
	var result []*SlsMasterRefund
	tx := db.Table("sls_master_refund")

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

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// AddSlsMasterRefund func
func AddSlsMasterRefund(tx *gorm.DB, slsMaster SlsMasterRefund) (*SlsMasterRefund, error) {
	if err := tx.Table("sls_master_refund").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}

// TotalRequestedRefundAmount struct
type TotalRequestedRefundAmount struct {
	TotalRequestAmount float64 `json:"request_amount" gorm:"column:request_amount"`
}

// GetTotalRequestedRefundAmount func
func GetTotalRequestedRefundAmount(slsMasterID int) (*TotalRequestedRefundAmount, error) {
	var totalRequestedRefundAmount TotalRequestedRefundAmount

	query := db.Table("sls_master_refund").
		Select("SUM(sls_master_refund.request_amount) as request_amount").
		Where("sls_master_refund.sls_master_id = ?", slsMasterID).
		Where("sls_master_refund.status != ?", "R") // include pending and approved refund request

	err := query.Find(&totalRequestedRefundAmount).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &totalRequestedRefundAmount, nil
}
