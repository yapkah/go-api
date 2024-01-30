package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SlsMasterBotLog struct
type SlsMasterBotLog struct {
	ID         int       `gorm:"primary_key" json:"id"`
	MemberID   int       `json:"member_id" gorm:"column:member_id"`
	DocNo      string    `json:"doc_no" gorm:"column:doc_no"`
	Status     string    `json:"status" gorm:"column:status"`
	RemarkType string    `json:"remark_type" gorm:"column:remark_type"`
	Remark     string    `json:"remark" gorm:"column:remark"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
}

// GetSlsMasterBotLog func
func GetSlsMasterBotLog(arrCond []WhereCondFn, debug bool) ([]*SlsMasterBotLog, error) {
	var result []*SlsMasterBotLog
	tx := db.Table("sls_master_bot_log").
		Order("sls_master_bot_log.id DESC").
		Limit(500)

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

// AddSlsMasterBotLog func
func AddSlsMasterBotLog(tx *gorm.DB, slsMasterBotLog SlsMasterBotLog) (*SlsMasterBotLog, error) {
	if err := tx.Table("sls_master_bot_log").Create(&slsMasterBotLog).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMasterBotLog, nil
}
