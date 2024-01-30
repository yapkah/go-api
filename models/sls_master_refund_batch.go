package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SlsMasterRefundBatch struct
type SlsMasterRefundBatch struct {
	ID        int       `gorm:"primary_key" json:"id"`
	DocNo     string    `json:"doc_no" gorm:"column:doc_no"`
	MemberID  int       `json:"member_id" gorm:"column:member_id"`
	Status    string    `json:"status" gorm:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

// GetSlsMasterRefundBatchFn func
func GetSlsMasterRefundBatchFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMasterRefundBatch, error) {
	var result []*SlsMasterRefundBatch
	tx := db.Table("sls_master_refund_batch").
		Order("created_at DESC")

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

// AddSlsMasterRefundBatch func
func AddSlsMasterRefundBatch(tx *gorm.DB, slsMaster SlsMasterRefundBatch) (*SlsMasterRefundBatch, error) {
	if err := tx.Table("sls_master_refund_batch").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}
