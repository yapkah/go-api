package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberReport struct
type EntMemberReport struct {
	ID          int    `gorm:"primary_key" json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	FilterParam string `json:"filter_param"`
	Header      string `json:"header"`
}

// GetEntMemberReportFn
func GetEntMemberReportFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberReport, error) {
	var result []*EntMemberReport
	tx := db.Table("ent_member_report")

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
