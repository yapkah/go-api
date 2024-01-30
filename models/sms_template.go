package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SmsTemplate struct
type SmsTemplate struct {
	ID        int    `gorm:"primary_key" json:"id"`
	Type      string `json:"type"`
	Locale    string `json:"locle"`
	Template  string `json:"template"`
	Status    string `json:"status"` // A: active | I: inactive
	UpdatedBy int    `json:"updated_by"`
}

// GetSmsTemplate get sms template
func GetSmsTemplate(arrCond []WhereCondFn, selectColumn string, debug bool) (*SmsTemplate, error) {
	var smsTemplate SmsTemplate
	tx := db.Table("sms_template")
	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&smsTemplate).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if smsTemplate.ID <= 0 {
		return nil, nil
	}

	return &smsTemplate, nil
}
