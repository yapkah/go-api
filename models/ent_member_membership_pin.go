package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberMembershipPin struct
type EntMemberMembershipPin struct {
	ID             int       `gorm:"primary_key" json:"id"`
	MemberID       int       `json:"member_id" gorm:"column:member_id"`
	Status         string    `json:"status" gorm:"column:status"`
	MembershipType string    `json:"membership_type" gorm:"column:membership_type"`
	PinCode        string    `json:"pin_code" gorm:"column:pin_code"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
}

func GetEntMemberMembershipPin(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberMembershipPin, error) {
	var result []*EntMemberMembershipPin
	tx := db.Table("ent_member_membership_pin")

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
