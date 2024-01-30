package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SysFaq struct
type SysFaq struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Locale    string    `json:"locale"`
	Status    string    `json:"status"`
	SeqNo     int       `json:"seq_no"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// GetSysFaqFn
func GetSysFaqFn(arrCond []WhereCondFn, debug bool) ([]*SysFaq, error) {
	var result []*SysFaq

	tx := db.Table("sys_faq").
		Order("sys_faq.seq_no ASC")

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
