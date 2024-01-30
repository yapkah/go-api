package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// AddSysLoginLockedAccountLogStruct struct
type AddSysLoginLockedAccountLogStruct struct {
	MemberID  int    `gorm:"column:member_id" json:"member_id"`
	Username  string `gorm:"column:username" json:"username"`
	LoginType string `gorm:"column:login_type" json:"login_type"`
	ClientIP  string `gorm:"column:client_ip" json:"client_ip"`
}

// SysLoginLockedAccountLogStruct struct
type SysLoginLockedAccountLogStruct struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `gorm:"column:member_id"  json:"member_id"`
	Username  string    `gorm:"column:username" json:"username"`
	LoginType string    `gorm:"column:login_type" json:"login_type"`
	ClientIP  string    `gorm:"column:client_ip" json:"client_ip"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// GetSysLoginLockedAccountLogFn get sys_login_locked_account_log data with dynamic condition
func GetSysLoginLockedAccountLogFn(arrCond []WhereCondFn, debug bool) ([]*SysLoginLockedAccountLogStruct, error) {
	var result []*SysLoginLockedAccountLogStruct
	tx := db.Table("sys_login_locked_account_log")

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

// func AddSysLoginLockedAccountLog
func AddSysLoginLockedAccountLog(saveData AddSysLoginLockedAccountLogStruct) error {
	if err := db.Table("sys_login_locked_account_log").Create(&saveData).Error; err != nil {
		ErrorLog("AddSysLoginLockedAccountLog-AddSysLoginLockedAccountLogStruct", err.Error(), saveData)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}
