package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EwtSummaryStrategy struct
type EwtSummaryStrategy struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"member_id"`
	Platform  string    `json:"platform"`
	Coin      string    `json:"coin"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

// AddEwtSummaryBnStruct struct
type AddEwtSummaryStrategyStruct struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"member_id"`
	Platform  string    `json:"platform"`
	Coin      string    `json:"coin"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// GetEwtSummaryStrategyFn get ewt_summary_bn data with dynamic condition
func GetEwtSummaryStrategyFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtSummaryStrategy, error) {
	var result []*EwtSummaryStrategy
	tx := db.Table("ewt_summary_strategy")

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

// func AddEwtSummaryStrategy add ewt_summary_strategy records
func AddEwtSummaryStrategy(saveData AddEwtSummaryStrategyStruct) (*AddEwtSummaryStrategyStruct, error) {
	if err := db.Table("ewt_summary_strategy").Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtSummaryStrategy-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}
