package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// BotFundingRate struct
type BotFundingRate struct {
	ID          int       `gorm:"primary_key" json:"id"`
	Symbol      string    `json:"symbol"`
	FundingTime int       `json:"fundingTime" gorm:"column:fundingTime"`
	FundingRate float64   `json:"fundingRate" gorm:"column:fundingRate"`
	BLatest     int       `json:"b_latest"`
	DtTimestamp time.Time `json:"dt_timestamp"`
}

// GetBotFundingRateFn
func GetBotFundingRateFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*BotFundingRate, error) {
	var result []*BotFundingRate
	tx := db.Table("bot_funding_rate")

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
