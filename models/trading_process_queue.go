package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TradingProcessQueue struct
type TradingProcessQueue struct {
	ID        int       `gorm:"primary_key" json:"id"`
	ProcessID string    `json:"process_id" gorm:"column:process_id"`
	Status    string    `json:"status" gorm:"column:status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetTradingProcessQueueFn get trading_cancel with dynamic condition
func GetTradingProcessQueueFn(arrCond []WhereCondFn, debug bool) ([]*TradingProcessQueue, error) {
	var result []*TradingProcessQueue
	tx := db.Table("trading_process_queue").
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

// AddTradingProcessQueueStruct struct
type AddTradingProcessQueueStruct struct {
	ID        int    `gorm:"primary_key" json:"id"`
	ProcessID string `json:"process_id" gorm:"column:process_id"`
	Status    string `json:"status" gorm:"column:status"`
}

// AddTradingProcessQueue func
func AddTradingProcessQueue(tx *gorm.DB, arrData AddTradingProcessQueueStruct) (*AddTradingProcessQueueStruct, error) {
	if err := tx.Table("trading_process_queue").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}
