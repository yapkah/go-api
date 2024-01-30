package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberTmpPw struct
type EntMemberTmpPw struct {
	ID           int       `gorm:"primary_key" json:"id"`
	MemberID     int       `gorm:"column:member_id" json:"member_id"`
	MemberMainID int       `gorm:"column:member_main_id" json:"member_main_id"`
	TmpPW        string    `gorm:"column:tmp_pw" json:"tmp_pw"`
	TToken       string    `gorm:"column:t_token" json:"t_token"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy    string    `gorm:"column:created_by" json:"created_by"`
	ExpiredAt    time.Time `gorm:"column:expired_at" json:"expired_at"`
}

// GetEntMemberTmpPwFn get ent_member data with dynamic condition
func GetEntMemberTmpPwFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberTmpPw, error) {
	var result []*EntMemberTmpPw
	tx := db.Table("ent_member_tmp_pw")

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
