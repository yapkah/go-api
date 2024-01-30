package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EwtLimiter struct
type EwtLimiter struct {
	ID            int     `gorm:"primary_key" json:"id"`
	MemberID      int     `json:"member_id"`
	EwalletTypeID int     `json:"ewallet_type_id"`
	LimitAmount   float64 `json:"limit_amount"`
	CreatedBy     string  `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
}

type MemberTotalEwtLimiter struct {
	TotalLimitAmount float64 `json:"total_limit_amount"`
}

// EwtLimiterFn get ewt_limiter data with dynamic condition
func EwtLimiterFn(arrCond []WhereCondFn, debug bool) ([]*EwtLimiter, error) {
	var result []*EwtLimiter
	tx := db.Table("ewt_limiter")
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

func GetMemAdminLimiterByEwtTypeID(member_id int, wallet_id int) (*MemberTotalEwtLimiter, error) {
	var result MemberTotalEwtLimiter
	tx := db.Table("ewt_limiter").Select("sum(limit_amount) as total_limit_amount").Where("member_id = ? AND ewallet_type_id = ?", member_id, wallet_id)

	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
