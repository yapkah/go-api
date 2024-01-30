package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SlsMasterBotLogTrans struct
type SlsMasterBotLogTrans struct {
	ID     int    `gorm:"primary_key" json:"id"`
	Locale string `json:"locale" gorm:"column:locale"`
	Key    string `json:"key" gorm:"column:key"`
	Value  string `json:"value" gorm:"column:value"`
}

// GetSlsMasterBotLogTrans func
func GetSlsMasterBotLogTrans(arrCond []WhereCondFn, debug bool) ([]*SlsMasterBotLogTrans, error) {
	var result []*SlsMasterBotLogTrans
	tx := db.Table("sls_master_bot_log_trans")

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
