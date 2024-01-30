package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// DeviceBindLog struct
type DeviceBindLog struct {
	ID             int       `gorm:"primary_key" gorm:"column:id" json:"id"`
	MemberID       int       `gorm:"column:member_id" json:"member_id"`
	TOs            string    `gorm:"column:t_os" json:"t_os"`
	TModel         string    `gorm:"column:t_model" json:"t_model"`
	TManufacturer  string    `gorm:"column:t_manufacturer" json:"t_manufacturer"`
	TAppVersion    string    `gorm:"column:t_app_version" json:"t_app_version"`
	TOsVersion     string    `gorm:"column:t_os_version" json:"t_os_version"`
	TPushNotiToken string    `gorm:"column:t_push_noti_token" json:"t_push_noti_token"`
	Bind           int       `gorm:"column:bind" json:"bind"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy      string    `gorm:"column:created_by" json:"created_by"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy      string    `gorm:"column:updated_by" json:"updated_by"`
}

// AddDeviceBindLog add device bind log
func AddDeviceBindLog(saveData DeviceBindLog) error {
	if err := db.Table("device_bind_log").Create(&saveData).Error; err != nil {
		ErrorLog("AddDeviceBindLog", err.Error(), saveData)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GetLatestDeviceBindLogFn get device_bind_log data with dynamic condition
func GetLatestDeviceBindLogFn(arrCond []WhereCondFn, debug bool) (*DeviceBindLog, error) {
	var result DeviceBindLog
	tx := db.Table("device_bind_log").
		Order("created_at DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
