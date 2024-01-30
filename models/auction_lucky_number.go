package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AuctionLuckyNumber struct
type AuctionLuckyNumber struct {
	ID          int       `gorm:"primary_key" json:"id"`
	RarityCode  string    `json:"rarity_code"`
	LuckyNumber int       `json:"lucky_number"`
	DateStart   time.Time `json:"date_start"`
	DateEnd     time.Time `json:"date_end"`
	Status      int       `json:"status"`
}

// GetAuctionLuckyNumber
func GetAuctionLuckyNumber(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*AuctionLuckyNumber, error) {
	var result []*AuctionLuckyNumber
	tx := db.Table("auction_lucky_number").
		Order("id desc")

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
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
