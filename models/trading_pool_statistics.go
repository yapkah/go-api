package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TradingPoolStatistic struct
type TradingPoolStatistic struct {
	ID            int       `gorm:"primary_key" json:"id"`
	Strategy      string    `json:"strategy"`
	TradeID       int       `json:"trade_id"`
	Symbol        string    `json:"symbol"`
	Num           float64   `json:"num"`
	Earnings      float64   `json:"earnings"`
	EarningsRatio float64   `json:"earnings_ratio"`
	BuyPrice      float64   `json:"buy_price"`
	SellPrice     float64   `json:"sell_price"`
	TradeAt       time.Time `json:"trade_at"`
	BStauts       int       `json:"b_status"`
	Timestamp     time.Time `json:"timestamp"`
}

// GetTradingPoolStatisticFn
func GetTradingPoolStatisticFn(arrCond []WhereCondFn, debug bool) ([]*TradingPoolStatistic, error) {
	var result []*TradingPoolStatistic

	tx := db.Table("trading_pool_statistics").
		Order("trading_pool_statistics.trade_at DESC")

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

// GetTradingPoolStatisticGroupFn
func GetTradingPoolStatisticGroupFn(arrCond []WhereCondFn, debug bool) ([]*TradingPoolStatistic, error) {
	var result []*TradingPoolStatistic

	tx := db.Table("trading_pool_statistics").
		Select("Date(trade_at) as date").
		Group("DATE(trade_at)")

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
