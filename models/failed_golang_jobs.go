package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// FailedGolangJobs struct
type FailedGolangJobs struct {
	ID         int    `gorm:"primary_key" json:"id"`
	Connection string `gorm:"column:connection" json:"connection"`
	Queue      string `gorm:"column:queue" json:"queue"`
	Payload    string `gorm:"column:payload" json:"payload"`
	Exception  string `gorm:"column:exception" json:"exception"`
}

// AddFailedGolangJobs add failed_golang_jobs
func AddFailedGolangJobs(tx *gorm.DB, saveData FailedGolangJobs) error {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddFailedGolangJobs-failed_golang_jobs", err.Error(), saveData)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}
