package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberMembership struct
type EntMemberMembership struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"member_id" gorm:"column:member_id"`
	BValid    int       `json:"b_valid" gorm:"column:b_valid"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	ExpiredAt time.Time `json:"expired_at"`
}

func GetEntMemberMembership(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberMembership, error) {
	var result []*EntMemberMembership
	tx := db.Table("ent_member_membership")

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

// AddEntMemberMembershipStruct struct
type AddEntMemberMembershipStruct struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"member_id" gorm:"column:member_id"`
	BValid    int       `json:"b_valid" gorm:"column:b_valid"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	ExpiredAt time.Time `json:"expired_at"`
}

// AddEntMemberMembership func
func AddEntMemberMembership(tx *gorm.DB, entMemberMembership AddEntMemberMembershipStruct) (*AddEntMemberMembershipStruct, error) {
	if err := tx.Table("ent_member_membership").Create(&entMemberMembership).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &entMemberMembership, nil
}
