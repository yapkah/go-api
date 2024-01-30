package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SysGeneral struct
type SysGeneral struct {
	ID           int    `gorm:"primary_key" json:"id"`
	Type         string `json:"type"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	BDisplayCode string `json:"b_display_code"`
	Status       string `json:"status"`
}

// GetSysGeneralByType func
func GetSysGeneralByType(sysType string) ([]*SysGeneral, error) {
	var sys []*SysGeneral

	err := db.Where("type = ? AND status = 'A'", sysType).Find(&sys).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return sys, nil
}

// GetSysGeneralByCode func
func GetSysGeneralByCode(sysType string, code string) (*SysGeneral, error) {
	var sys SysGeneral

	err := db.Where("type = ? AND code = ? AND status = 'A'", sysType, code).First(&sys).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if sys.ID <= 0 {
		return nil, nil
	}

	return &sys, nil
}
