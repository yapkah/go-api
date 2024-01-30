package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SupportTicketCategory struct
type SupportTicketCategory struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// GetSupportTicketCategoryFn func
func GetSupportTicketCategoryFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SupportTicketCategory, error) {
	var result []*SupportTicketCategory
	tx := db.Table("support_ticket_category")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	tx = tx.Order("id desc")
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
