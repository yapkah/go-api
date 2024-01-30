package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysEventsGhostList struct
type SysEventsGhostList struct {
	Username    string  `json:"username" gorm:"column:username"`
	TotalAmount float64 `json:"total_amount" gorm:"column:total_amount"`
}

// GetSysEventsGhostListFn
func GetSysEventsGhostListFn(eventType string, eventBatchNo, quota int, arrCond []WhereCondFn, debug bool) ([]*SysEventsGhostList, error) {
	var result []*SysEventsGhostList

	tx := db.Table("sys_events_ghost").
		Select("nick_name as username, SUM(amount) as total_amount").
		Where("event_type = ?", eventType).
		Where("event_batch_no = ?", eventBatchNo).
		Where("status = ?", "A")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	tx = tx.Group("nick_name, event_type, event_batch_no").
		Order("total_amount desc").
		Limit(quota)

	if debug {
		tx = tx.Debug()
	}

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
