package models

import (
	"net/http"

	"github.com/smartblock/gta-api/pkg/e"
)

// AddSysLogStruct struct
type AddSysLogStruct struct {
	ID         int    `gorm:"primary_key" json:"id"`
	UserID     int    `json:"user_id"`
	MemberID   int    `json:"member_id"`
	Type       string `json:"type"`
	Event      string `json:"event"`
	Status     string `json:"status"`
	OldValue   string `json:"old_value"`
	NewValue   string `json:"new_value"`
	IPAddress  string `json:"ip_address"` // A: active | I : inactive | T: terminate | S: suspend
	IPLocation string `json:"ip_location"`
	Device     string `json:"device"`
	ServerData string `json:"server_data"`
}

// AddSysLog add member
func AddSysLog(sysLog AddSysLogStruct) (*AddSysLogStruct, error) {
	if err := db.Table("sys_log").Create(&sysLog).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sysLog, nil
}
