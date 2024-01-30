package models

import (
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// ExchangePriceSetting struct
type BallotPriceSetting struct {
	ID           int     `gorm:"primary_key" json:"id"`
	TypeCode     string  `gorm:"type_code" json:"type_code"`
	PriceRate    float64 `json:"price_rate"`
	ReleasePerc  float64 `json:"release_perc"`
	MaxVolume    float64 `json:"max_volume"`
	VestingMonth string  `json:"vesting_month"`
}

// GetBallotPriceSettingFn get ballot_price_setting data with dynamic condition
func GetBallotPriceSettingFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*BallotPriceSetting, error) {
	var result BallotPriceSetting
	tx := db.Table("ballot_price_setting")
	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
		os.Exit(1)
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

func GetBallotPriceSettingListFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*BallotPriceSetting, error) {
	var result []*BallotPriceSetting
	tx := db.Table("ballot_price_setting")
	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
		os.Exit(1)
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
