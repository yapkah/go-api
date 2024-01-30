package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// BotCryptoPairKc struct
type BotCryptoPairKc struct {
	ID            int       `gorm:"primary_key" json:"id"`
	Symbol        string    `json:"symbol"`
	BaseCurrency  string    `json:"baseCurrency" gorm:"column:baseCurrency"`
	QuoteCurrency string    `json:"quoteCurrency" gorm:"column:quoteCurrency"`
	OtherData     string    `json:"otherData" gorm:"column:otherData"`
	DtTimestamp   time.Time `json:"dt_timestamp"`
}

// GetBotCryptoPairKcFn
func GetBotCryptoPairKcFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*BotCryptoPairKc, error) {
	var result []*BotCryptoPairKc
	tx := db.Table("bot_cryptopair_kc")

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
