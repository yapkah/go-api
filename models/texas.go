package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// Texas struct
type Texas struct {
	ID        int       `gorm:"primary_key" json:"id"`
	TexasID   int       `json:"texas_id" gorm:"column:texas_id"`
	Texas     string    `json:"texas" gorm:"column:texas"`
	EnTexas   string    `json:"en_texas" gorm:"column:en_texas"`
	CreatedAt time.Time `json:"created_at"`
}

// GetTexasFn get sc_hash data with dynamic condition
func GetTexasFn(arrCond []WhereCondFn, debug bool) ([]*Texas, error) {
	var result []*Texas
	tx := db.Table("texas")
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

// func AddTexas
func AddTexas(saveData Texas) (*Texas, error) {
	if err := db.Table("texas").Create(&saveData).Error; err != nil {
		ErrorLog("AddTexas-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}
