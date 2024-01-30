package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// NftTimer struct
type NftTimer struct {
	ID          int       `gorm:"primary_key" json:"id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Title       string    `json:"title"`
	CustCurrNft float64   `json:"cust_curr_nft"`
	TotalNft    float64   `json:"total_nft"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

// GetNftTimerFn
func GetNftTimerFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*NftTimer, error) {
	var result []*NftTimer
	tx := db.Table("nft_timer")

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
