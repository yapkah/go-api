package models

import (
	"encoding/json"

	"github.com/yapkah/go-api/pkg/logging"
)

// SysErrorLog struct
type EwtTransferExchangeLog struct {
	ID        string `gorm:"primary_key" json:"id"`
	MemID     int    `json:"member_id" gorm:"column:member_id"`
	Data1     string `json:"data_1" gorm:"column:frontend_receive_data"`
	Data2     string `json:"data_2" gorm:"column:blockchain_api_sent_data"`
	Data3     string `json:"data_3" gorm:"column:blockchain_api_return"`
	CreatedBy string `json:"created_by"`
}

// ErrorLog func
func SaveEwtTransferExchangeLog(memID int, data1 interface{}, data2 interface{}, data3 interface{}) {
	var jdata1, jdata2, jdata3 string

	if data1 != nil && data1 != "" {
		a, err := json.Marshal(data1)
		if err == nil {
			jdata1 = string(a)
		}
	}
	if data2 != nil && data2 != "" {
		b, err := json.Marshal(data2)
		if err == nil {
			jdata2 = string(b)
		}
	}
	if data3 != nil && data3 != "" {
		c, err := json.Marshal(data3)
		if err == nil {
			jdata3 = string(c)
		}
	}

	ewt := EwtTransferExchangeLog{
		MemID:     memID,
		Data1:     jdata1,
		Data2:     jdata2,
		Data3:     jdata3,
		CreatedBy: "AUTO",
	}

	err := db.Create(&ewt).Error
	if err != nil {
		logging.Error("ErrorLog ERROR", err.Error(), memID, jdata1, jdata2, jdata3)
	}

}
