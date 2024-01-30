package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysLanguage struct
type SysLanguage struct {
	ID     string `gorm:"primary_key" json:"id"`
	Locale string `json:"locale"`
	Name   string `json:"name"`
}

// GetLanguage func
func GetLanguage(locale string) (*SysLanguage, error) {
	var sys SysLanguage
	err := db.Where("locale = ?", locale).First(&sys).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &sys, nil
}

// ExistLangague func
func ExistLangague(locale string) bool {
	sys, err := GetLanguage(locale)

	if err != nil {
		return false
	}

	if sys == nil {
		return false
	}

	return true
}

// GetLanguageList func
func GetLanguageList() ([]*SysLanguage, error) {
	var sys []*SysLanguage
	err := db.Order("id").Find(&sys).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return sys, nil
}
