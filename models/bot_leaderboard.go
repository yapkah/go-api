package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// BotLeaderboard struct
type BotLeaderboard struct {
	ID            int       `gorm:"primary_key" json:"id"`
	Type          string    `gorm:"column:type" json:"string"`
	Symbol        string    `gorm:"column:symbol" json:"symbol"`
	TotalEarnings float64   `gorm:"column:totalEarnings" json:"totalEarnings"`
	Ratio         int       `gorm:"column:ratio" json:"ratio"`
	BLatest       int       `gorm:"column:b_latest" json:"b_latest"`
	DtTimestamp   time.Time `gorm:"column:dt_timestamp" json:"dt_timestamp"`
}

// GetBotLeaderboardFn get bot_leaderboard data with dynamic condition
func GetBotLeaderboardFn(arrCond []WhereCondFn, limit int, debug bool) ([]*BotLeaderboard, error) {
	var result []*BotLeaderboard
	tx := db.Table("bot_leaderboard").
		Order("bot_leaderboard.created_at DESC")

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

	return result, nil
}

func GetStrategyLeaderboardFn(arrCond []WhereCondFn, limit int, debug bool) ([]*BotLeaderboard, error) {
	var result []*BotLeaderboard
	tx := db.Table("bot_leaderboard").
		Order("bot_leaderboard.totalEarnings DESC")

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

	return result, nil
}
