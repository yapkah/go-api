package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberTradingTnc struct
type EntMemberTradingTnc struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"member_id"`
	Signature string    `json:"signature"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// GetEntMemberTradingTncFn
func GetEntMemberTradingTncFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberTradingTnc, error) {
	var result []*EntMemberTradingTnc
	tx := db.Table("ent_member_trading_tnc")

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

// AddEntMemberTradingTnc
func AddEntMemberTradingTnc(tx *gorm.DB, saveData EntMemberTradingTnc) (*EntMemberTradingTnc, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddEntMemberTradingTnc-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}
