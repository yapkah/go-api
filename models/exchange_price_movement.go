package models

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// ExchangePriceMovement struct
type ExchangePriceMovement struct {
	Id         int       `gorm:"primary_key" json:"id"`
	TokenPrice float64   `gorm:"token_price" json:"token_price"`
	CreatedBy  string    `gorm:"created_by" json:"created_by"`
	CreatedAt  time.Time `gorm:"created_at" json:"created_at"`
	UpdatedBy  string    `gorm:"updated_by" json:"updated_at"`
	UpdatedAt  time.Time `gorm:"updated_at" json:"updated_at"`
}

// GetLatestExchangePriceMovementByTokenType
func GetLatestExchangePriceMovementByTokenType(TokenType string) (float64, error) {
	var result ExchangePriceMovement
	tableName := "exchange_price_movement_" + strings.ToLower(TokenType)

	err := db.Table(tableName).
		Order("id DESC").First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result.TokenPrice, nil
}

type ArrFnStruct struct {
	ArrCond []WhereCondFn
	Limit   int
	OrderBy string
}

// GetExchangePriceMovementByTokenTypeFn get exchange_price_movement data with dynamic condition
func GetExchangePriceMovementByTokenTypeFn(TokenType string, arrFn ArrFnStruct, debug bool) ([]*ExchangePriceMovement, error) {
	var result []*ExchangePriceMovement
	tableName := "exchange_price_movement_" + strings.ToLower(TokenType)

	tx := db.Table(tableName)

	if len(arrFn.ArrCond) > 0 {
		for _, v := range arrFn.ArrCond {
			if v.CondValue != nil {
				tx = tx.Where(v.Condition, v.CondValue)
			} else {
				tx = tx.Where(v.Condition)
			}
		}
	}

	if arrFn.OrderBy != "" {
		tx = tx.Order(arrFn.OrderBy)
	}
	if arrFn.Limit > 0 {
		tx = tx.Limit(arrFn.Limit)
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

// MinMaxExchangePriceMovement struct
type MinMaxExchangePriceMovementPerDay struct {
	MinPrice  float64 `gorm:"column:min_price" json:"min_price"`
	MaxPrice  float64 `gorm:"column:max_price" json:"max_price"`
	TimeSlice string  `gorm:"column:timeslice" json:"timeslice"`
	DTUnix    int     `gorm:"column:dt_unix" json:"dt_unix"`
}

// MinMaxExchangePriceMovement struct
type MinMaxExchangePriceMovementByMinute struct {
	MinPrice  float64   `gorm:"column:min_price" json:"min_price"`
	MaxPrice  float64   `gorm:"column:max_price" json:"max_price"`
	TimeSlice time.Time `gorm:"column:timeslice" json:"timeslice"`
	DTUnix    int       `gorm:"column:dt_unix" json:"dt_unix"`
}

// GetMinMaxExchangePriceMovementByTokenTypePerDayFn get exchange_price_movement data with dynamic condition
func GetMinMaxExchangePriceMovementByTokenTypePerDayFn(TokenType string, arrFn []WhereCondFn, debug bool) ([]*MinMaxExchangePriceMovementPerDay, error) {
	var result []*MinMaxExchangePriceMovementPerDay
	tableName := "exchange_price_movement_" + strings.ToLower(TokenType)

	tx := db.Table(tableName).
		Select("MIN(token_price) AS 'min_price', MAX(token_price) AS 'max_price', DATE_FORMAT(created_at, \"%Y-%m-%d\") AS 'timeslice', UNIX_TIMESTAMP(DATE_FORMAT(created_at, \"%Y-%m-%d\")) AS 'dt_unix'").
		Group("timeslice, dt_unix").
		Order("timeslice ASC").
		Limit(200)

	if len(arrFn) > 0 {
		for _, v := range arrFn {
			if v.CondValue != nil {
				tx = tx.Where(v.Condition, v.CondValue)
			} else {
				tx = tx.Where(v.Condition)
			}
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

// GetMinMaxExchangePriceMovementByTokenTypeByMinuteFn get exchange_price_movement data with dynamic condition
func GetMinMaxExchangePriceMovementByTokenTypeByMinuteFn(TokenType string, minutes int, arrFn []WhereCondFn, debug bool) ([]*MinMaxExchangePriceMovementByMinute, error) {
	var result []*MinMaxExchangePriceMovementByMinute
	tableName := "exchange_price_movement_" + strings.ToLower(TokenType)
	min := minutes * 60
	minString := strconv.Itoa(min)
	timeSlice := "FROM_UNIXTIME(FLOOR(UNIX_TIMESTAMP(created_at)/" + minString + ")*" + minString + ")"
	tx := db.Table(tableName).
		Select("MIN(token_price) AS 'min_price', MAX(token_price) AS 'max_price', " + timeSlice + " AS 'timeslice', UNIX_TIMESTAMP(" + timeSlice + ") AS 'dt_unix'").
		Group("timeslice, dt_unix").
		Order("timeslice ASC").
		Limit(200)

	if len(arrFn) > 0 {
		for _, v := range arrFn {
			if v.CondValue != nil {
				tx = tx.Where(v.Condition, v.CondValue)
			} else {
				tx = tx.Where(v.Condition)
			}
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

type MinMaxExchangePriceMovementByTokenType struct {
	MinTokenPrice float64 `gorm:"column:min_token_price" json:"min_token_price"`
	MaxTokenPrice float64 `gorm:"column:max_token_price" json:"max_token_price"`
}

// GetMinMaxExchangePriceMovementByTokenTypeFn get exchange_price_movement by token type min & max data with dynamic condition
func GetMinMaxExchangePriceMovementByTokenTypeFn(TokenType string, arrCond []WhereCondFn, limit int, debug bool) (*MinMaxExchangePriceMovementByTokenType, error) {
	var result MinMaxExchangePriceMovementByTokenType
	tableName := "exchange_price_movement_" + strings.ToLower(TokenType)

	tx := db.Table(tableName).
		Select("MIN(token_price) AS min_token_price, MAX(token_price) AS max_token_price")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			if v.CondValue != nil {
				tx = tx.Where(v.Condition, v.CondValue)
			} else {
				tx = tx.Where(v.Condition)
			}
		}
	}
	if debug {
		tx = tx.Debug()
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
