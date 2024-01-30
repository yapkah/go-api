package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// LaligaProcessQueue struct
type LaligaProcessQueue struct {
	ID        int       `gorm:"primary_key" json:"id"`
	ProcessID string    `json:"process_id" gorm:"column:process_id"`
	Status    string    `json:"status" gorm:"column:status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetLaligaProcessQueueFn get trading_cancel with dynamic condition
func GetLaligaProcessQueueFn(arrCond []WhereCondFn, debug bool) ([]*LaligaProcessQueue, error) {
	var result []*LaligaProcessQueue
	tx := db.Table("laliga_process_queue").
		Order("created_at ASC")

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

// AddLaligaProcessQueueStruct struct
type AddLaligaProcessQueueStruct struct {
	ID        int    `gorm:"primary_key" json:"id"`
	ProcessID string `json:"process_id" gorm:"column:process_id"`
	Status    string `json:"status" gorm:"column:status"`
}

// AddLaligaProcessQueue func
func AddLaligaProcessQueue(tx *gorm.DB, arrData AddLaligaProcessQueueStruct) (*AddLaligaProcessQueueStruct, error) {
	if err := tx.Table("laliga_process_queue").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}
