package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberReservedSponsor struct
type EntMemberReservedSponsor struct {
	ID        int `gorm:"primary_key" json:"id"`
	MemberID  int `json:"member_id" gorm:"column:member_id"`
	SponsorID int `json:"sponsor_id" gorm:"column:sponsor_id"`
	UplineID  int `json:"upline_id" gorm:"column:upline_id"`
	LegNo     int `json:"leg_no" gorm:"column:leg_no"`
}

// AddEntMemberReservedSponsor func
func AddEntMemberReservedSponsor(tx *gorm.DB, member EntMemberReservedSponsor) (*EntMemberReservedSponsor, error) {
	if err := tx.Table("ent_member_reserved_sponsor").Create(&member).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &member, nil
}

// GetEntMemberReservedSponsorFn func
func GetEntMemberReservedSponsorFn(arrCond []WhereCondFn, debug bool) (*EntMemberReservedSponsor, error) {
	var member EntMemberReservedSponsor

	tx := db.Table("ent_member_reserved_sponsor")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if debug {
		tx = tx.Debug()
	}

	err := tx.First(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if member.ID <= 0 {
		return nil, nil
	}
	return &member, nil
}
