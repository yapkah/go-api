package models

import (
	"net/http"

	"github.com/smartblock/gta-api/pkg/e"
)

// AddGeneralApiLogStruct struct
type AddGeneralApiLogStruct struct {
	ID            int         `gorm:"primary_key" json:"id"`
	PrjConfigCode string      `gorm:"column:prj_config_code" json:"prj_config_code"`
	URLLink       string      `gorm:"column:url_link" json:"url_link"`
	ApiType       string      `gorm:"column:api_type" json:"api_type"`
	Method        string      `gorm:"column:method" json:"method"`
	DataSent      interface{} `gorm:"column:data_sent" json:"data_sent"`
	DataReceived  interface{} `gorm:"column:data_received" json:"data_received"`
	ServerData    string      `gorm:"column:server_data" json:"server_data"`
}

// func AddGeneralApiLog add general_api_log records
func AddGeneralApiLog(saveData AddGeneralApiLogStruct) (*AddGeneralApiLogStruct, error) {
	if err := db.Table("general_api_log").Create(&saveData).Error; err != nil {
		ErrorLog("AddGeneralApiLog-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}
