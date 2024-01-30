package models

import (
	"net/http"

	"github.com/yapkah/go-api/pkg/e"
)

// ApiLog struct
type ApiLog struct {
	ID          int    `gorm:"primary_key" json:"id"`
	UserID      int    `json:"user_id"`
	UserType    string `json:"user_type"`
	TokenID     string `json:"token_id"`
	RouteName   string `json:"route_name"`
	Method      string `json:"method"`
	Header      string `json:"header"`
	IPAddress   string `json:"ip_address"`
	Input       string `json:"input"`
	Output      string `json:"output"`
	RunningTime int    `json:"running_time"`
}

// AddAPILog add api log
func AddAPILog(route, method, header, ipaddress, input string) (*ApiLog, error) {
	log := ApiLog{
		RouteName: route,
		Method:    method,
		Header:    header,
		IPAddress: ipaddress,
		Input:     input,
	}

	if err := db.Create(&log).Error; err != nil {
		ErrorLog("ApiLog-AddAPILog", err.Error(), map[string]interface{}{"route": route, "method": method, "header": header, "ipaddress": ipaddress, "input": input})
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &log, nil
}

// Update add api log
func (a *ApiLog) Update(output string, runtime int) error {
	a.Output = output
	a.RunningTime = runtime
	err := save(a)
	if err != nil {
		ErrorLog("ApiLog-Update", err.Error(), map[string]interface{}{"output": output, "runtime": runtime})
		return err
	}
	return nil
}

// UpdateUser update user data
func (a *ApiLog) UpdateUser(userid int, usertype, tokenid string) error {
	a.UserID = userid
	a.UserType = usertype
	a.TokenID = tokenid
	err := save(a)
	if err != nil {
		ErrorLog("ApiLog-UpdateUser", err.Error(), map[string]interface{}{"userid": userid, "usertype": usertype, "tokenid": tokenid})
		return err
	}
	return nil
}

// UpdateOutput update user data
func (a *ApiLog) UpdateOutput(output string) error {
	a.Output = output
	err := save(a)
	if err != nil {
		ErrorLog("ApiLog-UpdateOutput", err.Error(), map[string]interface{}{"output": output})
		return err
	}
	return nil
}
