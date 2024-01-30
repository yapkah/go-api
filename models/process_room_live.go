package models

import (
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/smartblock/gta-api/pkg/e"
)

// ProcessStruct struct
type ProcessStruct struct {
	ThreadID int `gorm:"column:thread_id" json:"thread_id"`
}

// func GetProcecss
func GetProcess() *ProcessStruct {
	var result ProcessStruct
	db.Raw("select connection_id() as thread_id").Scan(&result)

	return &result
}

// ProcessQueueStruct struct
type ProcessQueueStruct struct {
	ID           int       `gorm:"primary_key" gorm:"column:id" json:"id"`
	ProcessID    string    `json:"process_id" gorm:"column:process_id"`
	Status       string    `json:"status" gorm:"column:status"`
	DtProcess    time.Time `json:"dt_process" gorm:"column:dt_process"`
	DtUpdated    time.Time `json:"dt_updated" gorm:"column:dt_updated"`
	MinutePassed int       `json:"minute_passed" gorm:"column:minute_passed"`
}

// func GetProcessQueue
func GetProcessQueue() []*ProcessQueueStruct {
	var result []*ProcessQueueStruct
	db.Raw("SELECT TIMESTAMPDIFF(MINUTE , dt_updated, NOW()) AS 'minute_passed', wod_process_queue.id, wod_process_queue.status, wod_process_queue.process_id FROM wod_process_queue ORDER BY id DESC LIMIT 1 ").Scan(&result)

	return result
}

// InfoSchemaProcessLisStruct struct
type InfoSchemaProcessLisStruct struct {
	ID big.Int `gorm:"column:ID" json:"ID"`
}

// func GetInfoSchemaProcessList
func GetInfoSchemaProcessList(pid string, debug bool) []*InfoSchemaProcessLisStruct {
	var result []*InfoSchemaProcessLisStruct
	tx := db.Raw("SELECT * FROM information_schema.processlist").
		Where("ID = ?", pid)
	if debug {
		tx = tx.Debug()
		os.Exit(0)
	}
	tx = tx.Scan(&result)

	return result
}

// AddWodProcessQueueStruct struct
type AddWodProcessQueueStruct struct {
	ID        int `gorm:"primary_key" gorm:"column:id" json:"id"`
	ProcessID int `gorm:"column:process_id" json:"process_id"`
}

// func AddWodProcessQueue
func AddWodProcessQueue(arrSaveData AddWodProcessQueueStruct) (*AddWodProcessQueueStruct, error) {
	if err := db.Table("wod_process_queue").Create(&arrSaveData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrSaveData, nil
}
