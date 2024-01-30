package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AddSysLoginAttemptsLogStruct struct
type AddSysLoginAttemptsLogStruct struct {
	MemberID  int    `gorm:"column:member_id" json:"member_id"`
	Username  string `gorm:"column:username" json:"username"`
	LoginType string `gorm:"column:login_type" json:"login_type"`
	ClientIP  string `gorm:"column:client_ip" json:"client_ip"`
	Attempts  int    `gorm:"column:attempts" json:"attempts"`
}

// SysLoginAttemptsLogStruct struct
type SysLoginAttemptsLogStruct struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `gorm:"column:member_id" json:"member_id"`
	Username  string    `gorm:"column:username" json:"username"`
	LoginType string    `gorm:"column:login_type" json:"login_type"`
	ClientIP  string    `gorm:"column:client_ip" json:"client_ip"`
	Attempts  int       `gorm:"column:attempts" json:"attempts"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// GetSysLoginAttemptsLogFn get wod_ticket data with dynamic condition
func GetSysLoginAttemptsLogFn(arrCond []WhereCondFn, debug bool) ([]*SysLoginAttemptsLogStruct, error) {
	var result []*SysLoginAttemptsLogStruct
	tx := db.Table("sys_login_attempts_log")

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

// func AddSysLoginAttemptsLog
func AddSysLoginAttemptsLog(saveData AddSysLoginAttemptsLogStruct) error {
	if err := db.Table("sys_login_attempts_log").Create(&saveData).Error; err != nil {
		ErrorLog("AddSysLoginAttemptsLog-AddSysLoginAttemptsLogStruct", err.Error(), saveData)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}
