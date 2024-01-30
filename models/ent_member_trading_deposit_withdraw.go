package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberTradingDepositWithdraw struct
type EntMemberTradingDepositWithdraw struct {
	ID          int       `gorm:"primary_key" json:"id"`
	MemberID    int       `json:"member_id" gorm:"column:member_id"`
	DocNo       string    `json:"doc_no" gorm:"column:doc_no"`
	TotalAmount float64   `json:"total_amount" gorm:"column:total_amount"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

func GetEntMemberTradingDepositWithdraw(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberTradingDepositWithdraw, error) {
	var result []*EntMemberTradingDepositWithdraw
	tx := db.Table("ent_member_trading_deposit_withdraw")

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

// AddEntMemberTradingDepositWithdraw func
func AddEntMemberTradingDepositWithdraw(tx *gorm.DB, entMemberTradingDepositWithdraw EntMemberTradingDepositWithdraw) (*EntMemberTradingDepositWithdraw, error) {
	if err := tx.Table("ent_member_trading_deposit_withdraw").Create(&entMemberTradingDepositWithdraw).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &entMemberTradingDepositWithdraw, nil
}
