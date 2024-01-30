package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysSubscPercStruct struct
type SysSubscPercStruct struct {
	ID         int       `gorm:"primary_key" json:"id"`
	Percentage float64   `gorm:"column:percentage" json:"percentage"`
	BLatest    int       `gorm:"column:b_latest" json:"b_latest"`
	CreatedBy  string    `gorm:"column:created_by" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func GetLatestSysSubscPerc() (*SysSubscPercStruct, error) {
	var sysSubscPerc SysSubscPercStruct
	err := db.Table("sys_subsc_perc").
		Where("b_latest = 1").First(&sysSubscPerc).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &sysSubscPerc, nil
}
