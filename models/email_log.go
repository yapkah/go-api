package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EmailLog struct
type EmailLog struct {
	ID       int    `gorm:"primary_key" json:"id"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
	Data     string `json:"data"`
}

// AddEmailLog add email log
func AddEmailLog(tx *gorm.DB, emailLog EmailLog) (*EmailLog, error) {
	if err := tx.Table("email_log").Create(&emailLog).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &emailLog, nil
}
