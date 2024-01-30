package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMembershipSetupPrice struct
type EntMembershipSetupPrice struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Code      string    `json:"code"`
	Status    string    `json:"status"`
	UnitPrice float64   `json:"unit_price"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

func GetEntMembershipSetupPrice(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMembershipSetupPrice, error) {
	var result []*EntMembershipSetupPrice
	tx := db.Table("ent_membership_setup_price")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("ent_membership_setup_price.id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
