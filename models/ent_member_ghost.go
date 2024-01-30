package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberGhost struct
type EntMemberGhost struct {
	ID           int       `gorm:"primary_key" json:"id"`
	Username     string    `json:"username"`
	BscAddress   string    `json:"bsc_address"`
	CreatedBy    string    `json:"created_by"`
	dt_timestamp time.Time `json:"dt_timestamp"`
}

// GetEntMemberGhostFn get ent_member_ghost data with dynamic condition
func GetEntMemberGhostFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*EntMemberGhost, error) {
	var entMember EntMemberGhost
	tx := db.Table("ent_member_ghost").
		Select("ent_member_ghost.*")

	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&entMember).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if entMember.ID <= 0 {
		return nil, nil
	}

	return &entMember, nil
}

type TotalGhostMember struct {
	TotalMember int `json:"total_member"`
}

func GetTotalGhostMemberFn(arrCond []WhereCondFn, debug bool) (*TotalGhostMember, error) {
	var result TotalGhostMember
	tx := db.Table("ent_member_ghost")
	tx = tx.Select("COUNT(DISTINCT(ent_member_ghost.id)) AS 'total_member'")

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

	return &result, nil
}
