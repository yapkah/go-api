package models

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// BotPremiumIndex struct
type BotPremiumIndex struct {
	ID              int       `gorm:"primary_key" json:"id"`
	Symbol          string    `json:"symbol"`
	BaseCurrency    string    `json:"baseCurrency" gorm:"column:baseCurrency"`
	QuoteCurrency   string    `json:"quoteCurrency" gorm:"column:quoteCurrency"`
	MarkPrice       float64   `json:"markPrice" gorm:"column:markPrice"`
	IndexPrice      float64   `json:"indexPrice" gorm:"column:indexPrice"`
	LastFundingRate float64   `json:"lastFundingRate" gorm:"column:lastFundingRate"`
	DtTimestamp     time.Time `json:"dt_timestamp"`
}

// GetBotPremiumIndexFn
func GetBotPremiumIndexFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*BotPremiumIndex, error) {
	var result []*BotPremiumIndex
	tx := db.Table("bot_premium_index")

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

// GetBotPremiumIndexRankedFundingRate
func GetBotPremiumIndexRankedFundingRate(rankNo int, platform string, debug bool) ([]*BotPremiumIndex, error) {
	var table = "bot_premium_index"
	if platform == "KC" {
		table = "bot_premium_index_kc"
	}

	var result []*BotPremiumIndex
	tx := db.Table(table).
		Where(table+".b_latest = ?", 1).
		Where(table + ".symbol NOT LIKE '1000%' ").
		Where(table + ".symbol NOT LIKE '%BUSD' ").
		Where(table + ".symbol NOT LIKE '%DEFIUSDT%' ").
		Where(table + ".symbol NOT LIKE '%DEFIUSDTM%' ").
		Where(table + ".symbol NOT LIKE '%BTCDOWNUSDT%' ").
		Where(table + ".symbol NOT LIKE '%BTCDOMUSDT%' ").
		Where(table + ".symbol NOT LIKE '%BTCSTUSDT%' ").
		Where(table + ".symbol NOT LIKE '%BIGTIME%' ").
		Where(table + ".symbol NOT LIKE '%\\_%' ")

	if platform == "KC" {
		tx = tx.Where(table + ".quoteCurrency LIKE 'USDT' ")

		// get usdt pairs that is with isMarginEnabled = false
		arrBotCryptoPairFn := []WhereCondFn{}
		arrBotCryptoPairFn = append(arrBotCryptoPairFn,
			WhereCondFn{Condition: "quoteCurrency = ?", CondValue: "USDT"},
		)
		arrBotCryptoPair, err := GetBotCryptoPairKcFn(arrBotCryptoPairFn, "", false)
		if err != nil {
			return nil, err
		}

		whereIn := ""

		for _, arrBotCryptoPairV := range arrBotCryptoPair {
			// find those isMarginEnabled = false
			type DataFormat struct {
				Symbol          string `json:"symbol"`
				Name            string `json:"name"`
				BaseCurrency    string `json:"baseCurrency"`
				QuoteCurrency   string `json:"quoteCurrency"`
				IsMarginEnabled bool   `json:"isMarginEnabled"`
				EnableTrading   bool   `json:"enableTrading"`
			}

			// mapping event setting into struct
			dataFormat := &DataFormat{}
			err := json.Unmarshal([]byte(arrBotCryptoPairV.OtherData), dataFormat)
			if err != nil {
				return nil, err
			}

			if !dataFormat.IsMarginEnabled {
				continue
			}

			if whereIn != "" {
				whereIn = whereIn + ","
			}
			whereIn = whereIn + "'" + arrBotCryptoPairV.BaseCurrency + "'"
		}

		if whereIn != "" {
			tx = tx.Where("baseCurrency IN(" + whereIn + ")")
		}

	} else {
		tx = tx.Where(table + ".symbol LIKE '%USDT' ")
	}

	tx = tx.Order(table + ".lastFundingRate DESC").
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

// GetBotPremiumIndexFundingRateHistory
func GetBotPremiumIndexFundingRateHistory(cryptoPair, platform string, limit int, debug bool) ([]*BotPremiumIndex, error) {
	var table = "bot_premium_index"
	if platform == "KC" {
		table = "bot_premium_index_kc"
	}

	var result []*BotPremiumIndex
	tx := db.Table(table).
		Where(table+".symbol = ?", cryptoPair).
		Order(table + ".dt_timestamp DESC").
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
