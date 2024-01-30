package models

import "time"

// ApiKeyAccessEvents struct
type ApiKeyAccessEvents struct {
	ID        int       `gorm:"primary_key" json:"id"`
	ApiKeyID  int       `json:"api_key_id"`
	IpAddress string    `json:"ip_address"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// AddApiKeyAccessEvents add api key access events
func AddApiKeyAccessEvents(arrData ApiKeyAccessEvents) {
	if err := db.Create(&arrData).Error; err != nil {
		ErrorLog("AddApiKeyAccessEvents-ApiKeyAccessEvents", err.Error(), arrData)
	}
}
