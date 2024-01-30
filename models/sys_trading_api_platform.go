package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysTradingApiPlatform struct
type SysTradingApiPlatform struct {
	ID      int    `gorm:"primary_key" json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	ImgUrl  string `json:"img_url"`
	Setting string `json:"setting"`
}

// GetSysTradingApiPlatformFn
func GetSysTradingApiPlatformFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SysTradingApiPlatform, error) {
	var result []*SysTradingApiPlatform
	tx := db.Table("sys_trading_api_platform")

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
