package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// AddSlsMasterBatchStruct struct
type AddSlsMasterBatchStruct struct {
	ID        int       `gorm:"primary_key" json:"id"`
	BatchNo   string    `json:"batch_no" gorm:"column:batch_no"`
	Quantity  float64   `json:"quantity" gorm:"column:quantity"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// AddSlsMasterBatch func
func AddSlsMasterBatch(tx *gorm.DB, slsMaster AddSlsMasterBatchStruct) (*AddSlsMasterBatchStruct, error) {
	if err := tx.Table("sls_master_batch").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}
