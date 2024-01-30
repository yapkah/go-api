package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// AddSlsMasterBnsQueueStruct struct
type AddSlsMasterBnsQueueStruct struct {
	ID       int       `gorm:"primary_key" json:"id"`
	DocNo    string    `json:"doc_no" gorm:"column:doc_no"`
	BStatus  string    `json:"b_status" gorm:"column:b_status"`
	DtCreate time.Time `json:"dt_create" gorm:"column:dt_create"`
}

// AddSlsMasterBnsQueue func
func AddSlsMasterBnsQueue(tx *gorm.DB, slsMasterBotSetting AddSlsMasterBnsQueueStruct) (*AddSlsMasterBnsQueueStruct, error) {
	if err := tx.Table("sls_master_bns_queue").Create(&slsMasterBotSetting).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMasterBotSetting, nil
}
