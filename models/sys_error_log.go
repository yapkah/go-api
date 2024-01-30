package models

import (
	"encoding/json"

	"github.com/yapkah/go-api/pkg/logging"
)

// SysErrorLog struct
type SysErrorLog struct {
	ID        int    `gorm:"primary_key" json:"id"`
	Data1     string `json:"data_1" gorm:"column:data_1"`
	Data2     string `json:"data_2" gorm:"column:data_2"`
	Data3     string `json:"data_3" gorm:"column:data_3"`
	CreatedBy string `json:"created_by"`
}

// ErrorLog func
func ErrorLog(data1 interface{}, data2 interface{}, data3 interface{}) {
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

	sys := SysErrorLog{
		Data1:     jdata1,
		Data2:     jdata2,
		Data3:     jdata3,
		CreatedBy: "AUTO",
	}

	err := db.Create(&sys).Error
	if err != nil {
		logging.Error("ErrorLog ERROR", err.Error(), jdata1, jdata2, jdata3)
	}

}

// ErrorLog func
func ErrorLogV2(data1 interface{}, data2 interface{}, data3 interface{}) *SysErrorLog {
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

	sys := SysErrorLog{
		Data1:     jdata1,
		Data2:     jdata2,
		Data3:     jdata3,
		CreatedBy: "AUTO",
	}

	err := db.Save(&sys).Error
	if err != nil {
		logging.Error("ErrorLog ERROR", err.Error(), jdata1, jdata2, jdata3)
	}

	return &sys
}
