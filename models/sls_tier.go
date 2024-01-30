package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SlsTier struct
type SlsTier struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"country_id"`
	Tier      string    `json:"tier"`
	Status    string    `json:"status"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// GetSlsTierFn
func GetSlsTierFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsTier, error) {
	var result []*SlsTier
	tx := db.Table("sls_tier").Order("sls_tier.created_at DESC")

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

// AddSlsTier func
func AddSlsTier(tx *gorm.DB, slsNftTier SlsTier) (*SlsTier, error) {
	if err := tx.Table("sls_tier").Create(&slsNftTier).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsNftTier, nil
}
