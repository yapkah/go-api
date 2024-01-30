package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SysTradingCryptoPairSetup struct
type SysTradingCryptoPairSetup struct {
	ID     int    `gorm:"primary_key" json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status string `json:"status"`

	// Spot Grid Trading
	// GridQuantity int     `json:"grid_quantity"` // removed since 5th april 2022 and is calculated instead
	TakerRate float64 `json:"taker_rate"`

	// Martingale Trading
	// FirstOrderPrice      float64 `json:"first_order_price"`
	PriceScale float64 `json:"price_scale"`
	// SubsequentPriceScale int     `json:"subsequent_price_scale"` // removed since 5th april 2022
	TakeProfitRatio      float64 `json:"take_profit_ratio"`
	TakeProfitAdjustment float64 `json:"take_profit_adjustment"`
	// SafetyOrders         int     `json:"safety_orders"` // removed since 5th april 2022 and is calculated instead
	AddShares int `json:"add_shares"`
	// SubsequentAddShares  int     `json:"subsequent_add_shares"` // removed since 5th april 2022
	CircularTrans           int     `json:"circular_trans"`
	MtdPriceScale           float64 `json:"mtd_price_scale"`
	MtdTakeProfitRatio      float64 `json:"mtd_take_profit_ratio"`
	MtdTakeProfitAdjustment float64 `json:"mtd_take_profit_adjustment"`
	MtdAddShares            float64 `json:"mtd_add_shares"`
	MtdCircularTrans        int     `json:"mtd_circular_trans"`
}

// GetSysTradingCryptoPairSetupFn
func GetSysTradingCryptoPairSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SysTradingCryptoPairSetup, error) {
	var result []*SysTradingCryptoPairSetup
	tx := db.Table("sys_trading_crypto_pair_setup")

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

// GetSysTradingCryptoPairSetupByPlatformFn
func GetSysTradingCryptoPairSetupByPlatformFn(arrCond []WhereCondFn, selectColumn, platform string, debug bool) ([]*SysTradingCryptoPairSetup, error) {
	var result []*SysTradingCryptoPairSetup
	var table = "sys_trading_crypto_pair_setup"
	if platform == "KC" {
		table = "sys_trading_crypto_pair_setup_kc"
	}

	tx := db.Table(table)

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
