package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntCurrentProfileStruct struct
type EntCurrentProfileStruct struct {
	ID        int       `gorm:"primary_key" json:"id"`
	SourceID  int       `gorm:"column:source_id" json:"source_id"`
	MainID    int       `gorm:"column:main_id" json:"main_id"`
	MemberID  int       `gorm:"column:member_id" json:"member_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy string    `gorm:"column:created_by" json:"created_by"`
}

// AddEntCurrentProfileStruct struct
type AddEntCurrentProfileStruct struct {
	ID       int `gorm:"primary_key" json:"id"`
	SourceID int `gorm:"column:source_id" json:"source_id"`
	MainID   int `gorm:"column:main_id" json:"main_id"`
	MemberID int `gorm:"column:member_id" json:"member_id"`
}

// AddEntCurrentProfile add member
func AddEntCurrentProfile(tx *gorm.DB, arrData AddEntCurrentProfileStruct) (*AddEntCurrentProfileStruct, error) {
	if err := tx.Table("ent_current_profile").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// GetEntCurrentProfileFn get ent_current_profile with dynamic condition
func GetEntCurrentProfileFn(arrCond []WhereCondFn, debug bool) ([]*EntCurrentProfileStruct, error) {
	var result []*EntCurrentProfileStruct
	tx := db.Table("ent_current_profile").
		Order("ent_current_profile.id DESC")

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
