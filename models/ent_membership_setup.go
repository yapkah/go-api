package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMembershipSetup struct
type EntMembershipSetup struct {
	ID            int       `gorm:"primary_key" json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	PeriodSetting string    `json:"period_setting"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
}

func GetEntMembershipSetup(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMembershipSetup, error) {
	var result []*EntMembershipSetup
	tx := db.Table("ent_membership_setup")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("ent_membership_setup.id asc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
