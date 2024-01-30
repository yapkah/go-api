package models

import (
	"net/http"

	"github.com/smartblock/gta-api/pkg/e"
)

// AuctionApiLog struct
type AuctionApiLog struct {
	ID            int    `gorm:"primary_key" json:"id"`
	ApiType       string `json:"api_type"`
	TServer       string `json:"t_server"`
	TRequest      string `json:"t_request"`
	PrjConfigCode string `json:"prj_config_code"`
	Method        string `json:"method"`
	UrlLink       string `json:"url_link"`
	DataSent      string `json:"data_sent"`
}

// AddAuctionApiLog add blockchain api log
func AddAuctionApiLog(route, method, header, ipaddress, input string) (*AuctionApiLog, error) {
	log := AuctionApiLog{
		UrlLink:       route,
		Method:        method,
		TServer:       header,
		ApiType:       ipaddress,
		TRequest:      input,
		PrjConfigCode: "blockchain",
	}

	if err := db.Create(&log).Error; err != nil {
		ErrorLog("AddAuctionApiLog-failed_to_save_api_log", err.Error(), map[string]interface{}{"route": route, "method": method, "header": header, "ipaddress": ipaddress, "input": input})
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &log, nil
}

// Update add api log
func (a *AuctionApiLog) UpdateAuctionApiLog(output string) error {
	a.DataSent = output
	err := save(a)
	if err != nil {
		ErrorLog("ApiLog-Update", err.Error(), map[string]interface{}{"output": output})
		return err
	}
	return nil
}
