package models

import (
	"time"
)

// AddApiKeyAdminEventsStruct struct
type AddApiKeyAdminEventsStruct struct {
	ID        int       `gorm:"primary_key" json:"id"`
	ApiKeyID  int       `json:"api_key_id"`
	IpAddress string    `json:"ip_address"`
	Event     string    `json:"event"`
	CreatedAt time.Time `json:"created_at"`
}

// AddApiKeyAdminEvents add api_key_admin_events
func AddApiKeyAdminEvents(arrData AddApiKeyAdminEventsStruct) {
	if err := db.Table("api_key_admin_events").Create(&arrData).Error; err != nil {
		ErrorLog("AddApiKeyAdminEvents-AddApiKeyAdminEventsStruct", err.Error(), arrData)
	}
}
