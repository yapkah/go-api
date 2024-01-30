package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SmsLog struct
type SmsLog struct {
	ID          int    `gorm:"primary_key" json:"id"`
	MobileNo    string `json:"mobile_no"`
	TemplateID  int    `json:"template_id"`
	MsgContent  string `json:"msg_content"`
	ReturnValue string `json:"return_value"`
	API         string `json:"api"`
}

// AddSmsLog add sms log
func AddSmsLog(tx *gorm.DB, smsLog SmsLog) (*SmsLog, error) {
	if err := tx.Table("sms_log").Create(&smsLog).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &smsLog, nil
}
