package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// BotPremiumIndexKc struct
type BotPremiumIndexKc struct {
	ID              int       `gorm:"primary_key" json:"id"`
	Symbol          string    `json:"symbol"`
	BaseCurrency    string    `json:"baseCurrency"`
	QuoteCurrency   string    `json:"quoteCurrency"`
	MarkPrice       float64   `json:"markPrice" gorm:"column:markPrice"`
	IndexPrice      float64   `json:"indexPrice" gorm:"column:indexPrice"`
	LastFundingRate float64   `json:"lastFundingRate" gorm:"column:lastFundingRate"`
	DtTimestamp     time.Time `json:"dt_timestamp"`
}

// GetBotPremiumIndexKcFn
func GetBotPremiumIndexKcFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*BotPremiumIndexKc, error) {
	var result []*BotPremiumIndexKc
	tx := db.Table("bot_premium_index_kc")

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

// GetBotPremiumIndexKcRankedFundingRate
func GetBotPremiumIndexKcRankedFundingRate(rankNo int, debug bool) ([]*BotPremiumIndexKc, error) {
	var result []*BotPremiumIndexKc
	tx := db.Table("bot_premium_index_kc").
		Where("bot_premium_index_kc.b_latest = ?", 1).
		Where("bot_premium_index_kc.symbol LIKE '%USDT' ").
		Where("bot_premium_index_kc.symbol NOT LIKE '1000%' ").
		Where("bot_premium_index_kc.symbol NOT LIKE '%BUSD' ").
		Where("bot_premium_index_kc.symbol NOT LIKE '%DEFIUSDT%' ").
		Where("bot_premium_index_kc.symbol NOT LIKE '%BTCDOWNUSDT%' ").
		Where("bot_premium_index_kc.symbol NOT LIKE '%\\_%' ").
		Order("bot_premium_index_kc.lastFundingRate DESC").
		Limit(rankNo)

	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetBotPremiumIndexKcFundingRateHistory
func GetBotPremiumIndexKcFundingRateHistory(cryptoPair string, limit int, debug bool) ([]*BotPremiumIndexKc, error) {
	var result []*BotPremiumIndexKc
	tx := db.Table("bot_premium_index_kc").
		Where("bot_premium_index_kc.symbol = ?", cryptoPair).
		Order("bot_premium_index_kc.dt_timestamp DESC").
		Limit(limit)

	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
