package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberSetting struct
type EntMemberSetting struct {
	ID            int    `gorm:"primary_key" json:"id" gorm:"column:id"`
	MemberID      int    `json:"member_id" gorm:"column:member_id"`
	Type          string `json:"current_rank" gorm:"column:type"`
	EwalletTypeID int    `json:"grade_id" gorm:"column:ewallet_type_id"`
	Mode          string `json:"total_star" gorm:"column:mode"`
	Flag          int    `json:"d_last_star" gorm:"column:flag"`
}

// GetEntMemberSetting get ent_member_setting data with dynamic condition
func GetEntMemberSetting(arrCond []WhereCondFn, orderBy []OrderByFn, debug bool) ([]*EntMemberSetting, error) {
	var result []*EntMemberSetting
	tx := db.Table("ent_member_setting")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if len(orderBy) > 0 {
		for _, o := range orderBy {
			tx = tx.Order(o.Condition)
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
