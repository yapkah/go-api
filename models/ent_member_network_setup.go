package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberNetworkSetup struct
type EntMemberNetworkSetup struct {
	ID       int    `gorm:"primary_key" json:"id"`
	Type     string `json:"type"`
	MemberID string `json:"member_id"`
}

// GetEntMemberNetworkSetupFn
func GetEntMemberNetworkSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberNetworkSetup, error) {
	var result []*EntMemberNetworkSetup
	tx := db.Table("ent_member_network_setup")

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
