package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// Jobs struct
type Jobs struct {
	ID          int    `gorm:"primary_key" json:"id"`
	Queue       string `gorm:"column:queue" json:"queue"`
	Payload     string `gorm:"column:payload" json:"payload"`
	Attempts    int    `gorm:"column:attempts" json:"attempts"`
	ReservedAt  uint64 `gorm:"column:reserved_at" json:"reserved_at"`
	CreatedAt   uint64 `gorm:"column:created_at" json:"created_at"`
	AvailableAt uint64 `gorm:"column:available_at" json:"available_at"`
}

// GetJobsFn get jobs with dynamic condition
func GetJobsFn(arrCond []WhereCondFn, debug bool) ([]*Jobs, error) {
	var result []*Jobs
	tx := db.Table("jobs").
		Order("jobs.id DESC")

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
