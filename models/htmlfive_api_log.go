package models

import (
	"net/http"

	"github.com/smartblock/gta-api/pkg/e"
)

// HtmlfiveApiLog struct
type HtmlfiveApiLog struct {
	ID            int    `gorm:"primary_key" json:"id"`
	ApiType       string `json:"api_type"`
	TServer       string `json:"t_server"`
	TRequest      string `json:"t_request"`
	PrjConfigCode string `json:"prj_config_code"`
	Method        string `json:"method"`
	UrlLink       string `json:"url_link"`
	DataSent      string `json:"data_sent"`
}

// AddHtmlfiveApiLog add htmlfive api log
func AddHtmlfiveApiLog(route, method, header, ipaddress, input string) (*HtmlfiveApiLog, error) {
	log := HtmlfiveApiLog{
		UrlLink:       route,
		Method:        method,
		TServer:       header,
		ApiType:       ipaddress,
		TRequest:      input,
		PrjConfigCode: "htmlfive",
	}

	if err := db.Create(&log).Error; err != nil {
		ErrorLog("AddHtmlfiveApiLog-failed_to_save_api_log", err.Error(), map[string]interface{}{"route": route, "method": method, "header": header, "ipaddress": ipaddress, "input": input})
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &log, nil
}

// Update add api log
func (a *HtmlfiveApiLog) UpdateHtmlfiveApiLog(output string) error {
	a.DataSent = output
	err := save(a)
	if err != nil {
		ErrorLog("ApiLog-Update", err.Error(), map[string]interface{}{"output": output})
		return err
	}
	return nil
}

// UpdateUser update user data
// func (a *HtmlfiveApiLog) UpdateUser(userid int, usertype, tokenid string) error {
// 	a.UserID = userid
// 	a.UserType = usertype
// 	a.TokenID = tokenid
// 	err := save(a)
// 	if err != nil {
// 		ErrorLog("ApiLog-UpdateUser", err.Error(), map[string]interface{}{"userid": userid, "usertype": usertype, "tokenid": tokenid})
// 		return err
// 	}
// 	return nil
// }

// UpdateHtmlfiveOutput update user data
func (a *HtmlfiveApiLog) UpdateHtmlfiveOutput(output string) error {
	a.DataSent = output
	err := save(a)
	if err != nil {
		ErrorLog("ApiLog-UpdateHtmlfiveOutput", err.Error(), map[string]interface{}{"output": output})
		return err
	}
	return nil
}
