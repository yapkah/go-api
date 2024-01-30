package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

type GolangJobsFnStruct struct {
	JobName string
}

// GolangJobs struct
type GolangJobs struct {
	ID          int    `gorm:"primary_key" json:"id"`
	Queue       string `gorm:"column:queue" json:"queue"`
	Payload     string `gorm:"column:payload" json:"payload"`
	Attempts    int    `gorm:"column:attempts" json:"attempts"`
	ReservedAt  uint64 `gorm:"column:reserved_at" json:"reserved_at"`
	CreatedAt   uint64 `gorm:"column:created_at" json:"created_at"`
	AvailableAt uint64 `gorm:"column:available_at" json:"available_at"`
}

// GetGolangJobsFn get jobs with dynamic condition
func GetGolangJobsFn(arrCond []WhereCondFn, debug bool) ([]*GolangJobs, error) {
	var result []*GolangJobs
	tx := db.Table("golang_jobs").
		Order("golang_jobs.id ASC")

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
