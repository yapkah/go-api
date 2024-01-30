package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EmailTemplate struct
type EmailTemplate struct {
	ID        int    `gorm:"primary_key" json:"id"`
	Type      string `json:"type"`
	Locale    string `json:"locle"`
	Title     string `json:"title"`
	Template  string `json:"template"`
	Status    string `json:"status"` // A: active | I: inactive
	UpdatedBy int    `json:"updated_by"`
}

// GetEmailTemplate get email template
func GetEmailTemplate(arrCond []WhereCondFn, selectColumn string, debug bool) (*EmailTemplate, error) {
	var emailTemplate EmailTemplate
	tx := db.Table("email_template")
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
	err := tx.Find(&emailTemplate).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if emailTemplate.ID <= 0 {
		return nil, nil
	}

	return &emailTemplate, nil
}
