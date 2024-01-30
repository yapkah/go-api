package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// AddEntMemberCryptoStruct struct
type RwdPeriod struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Type      string    `json:"type"`
	BatchCode string    `json:"batch_code"`
	DateFrom  time.Time `json:"date_from"`
	DateTo    time.Time `json:"date_to"`
	Active    int       `json:"active"`
	Paid      int       `json:"paid"`
	Dividend  int       `json:"dividend"`
}

func CheckRewardPeriod(batch_code string) (*RwdPeriod, error) {
	var period RwdPeriod

	query := db.Table("rwd_period a")

	if batch_code != "" {
		query = query.Where("a.batch_code = ?", batch_code)
	}

	err := query.Find(&period).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &period, nil
}
