package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberTradingWalletLimit struct
type EntMemberTradingWalletLimit struct {
	ID          int       `gorm:"primary_key" json:"id"`
	MemberID    int       `json:"member_id" gorm:"column:member_id"`
	Module      string    `json:"module" gorm:"column:module"`
	TotalAmount float64   `json:"total_amount" gorm:"column:total_amount"`
	Status      string    `json:"status" gorm:"column:status"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

func GetEntMemberTradingWalletLimit(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberTradingWalletLimit, error) {
	var result []*EntMemberTradingWalletLimit
	tx := db.Table("ent_member_trading_wallet_limit").
		Order("created_at DESC")

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

// AddEntMemberTradingWalletLimit func
func AddEntMemberTradingWalletLimit(tx *gorm.DB, entMemberTradingDeposit EntMemberTradingWalletLimit) (*EntMemberTradingWalletLimit, error) {
	if err := tx.Table("ent_member_trading_wallet_limit").Create(&entMemberTradingDeposit).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &entMemberTradingDeposit, nil
}
