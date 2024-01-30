package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SysEvents struct
type SysEvents struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Type      string    `json:"type"`
	BatchNo   int       `json:"batch_no"`
	TimeStart time.Time `json:"time_start"`
	TimeEnd   time.Time `json:"time_end"`
	Status    int       `json:"status"`
	Setting   string    `json:"setting"`
}

// GetSysEventsFn
func GetSysEventsFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SysEvents, error) {
	var result []*SysEvents
	tx := db.Table("sys_events")

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
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
