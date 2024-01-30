package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMember2Fa struct
type EntMember2Fa struct {
	ID       int    `gorm:"primary_key" json:"id"`
	MemberID int    `json:"member_id"`
	Secret   string `json:"secret"`
	CodeUrl  string `json:"codeurl"`
	BEnable  int    `json:"b_enable"`
}

func GetEntMember2FA(arrCond []WhereCondFn, debug bool) ([]*EntMember2Fa, error) {
	var result []*EntMember2Fa
	tx := db.Table("ent_member_2fa").
		Select("ent_member_2fa.*").
		Joins("inner join ent_member ON ent_member.id = ent_member_2fa.member_id").
		Order("id DESC")

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

// AddEntMember2FaParam struct
type AddEntMember2FaParam struct {
	ID       int    `gorm:"primary_key" json:"id"`
	MemberID int    `json:"member_id"`
	Secret   string `json:"secret"`
	CodeUrl  string `json:"codeurl" gorm:"column:codeurl"`
	BEnable  int    `json:"b_enable"`
}

// AddEntMember2FA add ent_member_lot_queue
func AddEntMember2FA(tx *gorm.DB, tree AddEntMember2FaParam) (*AddEntMember2FaParam, error) {
	if err := tx.Table("ent_member_2fa").Create(&tree).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &tree, nil
}
