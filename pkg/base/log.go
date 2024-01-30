package base

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/e"
)

// GetMemberLogData get member log data
func GetMemberLogData(arrTable []string, memberID int) (string, map[string]interface{}) {
	var (
		value      = make(map[string]interface{})
		validTable = []string{"members", "ent_member", "ent_member_crypto"}
	)

	// value := map[string]interface{}{
	// 	"ent_member":        arrEntMember,
	// 	"ent_member_crypto": arrEntMemberCrypto,
	// }

	for _, table := range arrTable {
		if helpers.StringInSlice(table, validTable) == true {
			switch table {
			case "members":
				arrEntMemberFn := make([]models.WhereCondFn, 0)
				arrEntMemberFn = append(arrEntMemberFn,
					models.WhereCondFn{Condition: "members.id = ?", CondValue: memberID},
				)
				arrEntMember, err := models.GetMembersFn(arrEntMemberFn, false)

				if err != nil {
					models.ErrorLog("base:GetMemberLogData()", "GetMembersFn():1", err.Error())
					return "something_went_wrong", nil
				}
				if arrEntMember == nil {
					return e.GetMsg(e.INVALID_MEMBER), nil
				}

				value["members"] = arrEntMember

			case "ent_member":
				arrEntMemberFn := make([]models.WhereCondFn, 0)
				arrEntMemberFn = append(arrEntMemberFn,
					models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: memberID},
				)
				arrEntMember, err := models.GetEntMemberFn(arrEntMemberFn, "", false)

				if err != nil {
					models.ErrorLog("base:GetMemberLogData()", "GetEntMemberFn():1", err.Error())
					return "something_went_wrong", nil
				}
				if arrEntMember == nil {
					return e.GetMsg(e.INVALID_MEMBER), nil
				}

				value["ent_member"] = arrEntMember

			case "ent_member_crypto":
				arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
				arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
					models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: memberID},
					models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
				)
				arrEntMemberCrypto, err := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)

				if err != nil {
					models.ErrorLog("base:GetMemberLogData()", "GetEntMemberCryptoFn():1", err.Error())
					return "something_went_wrong", nil
				}
				if arrEntMemberCrypto == nil {
					// insert empty array
					// return e.GetMsg(e.INVALID_MEMBER), nil
				}

				value["ent_member_crypto"] = arrEntMemberCrypto
			}
		}
	}

	return "", value
}

// AddSysLog retrieve server data and add to sys log
func AddSysLog(memberID int, currentData, updatedData map[string]interface{}, logType, event string, c *gin.Context) string {
	var (
		req                           *http.Request
		ipAddress, ipLocation, device string
		arrServerData                 = make(map[string]interface{})
	)

	req = c.Request
	marshalCurrentData, err := json.Marshal(currentData)
	if err != nil {
		models.ErrorLog("base:AddSysLog()", "Marshal():1", err.Error())
		return "something_went_wrong"
	}

	marshalUpdatedData, err := json.Marshal(updatedData)
	if err != nil {
		models.ErrorLog("base:AddSysLog()", "Marshal():2", err.Error())
		return "something_went_wrong"
	}

	ipAddress, err = getIP(req)
	if err != nil {
		models.ErrorLog("base:AddSysLog()", "getIP():1", err.Error())
		return "something_went_wrong"
	}

	// ipLocation = req.Header.Get("HTTP_USER_AGENT") // need to get location from ip
	device = req.Header.Get("User-Agent")

	arrServerData["header"] = req.Header
	arrServerData["path"] = req.URL.String()
	marshalServerData, err := json.Marshal(arrServerData)
	if err != nil {
		models.ErrorLog("base:AddSysLog()", "Marshal():3", err.Error())
		return "something_went_wrong"
	}

	var arrSysLog = models.AddSysLogStruct{
		UserID:     memberID,
		MemberID:   memberID,
		Type:       logType,
		Event:      event,
		Status:     "S",
		OldValue:   string(marshalCurrentData),
		NewValue:   string(marshalUpdatedData),
		IPAddress:  ipAddress,
		IPLocation: ipLocation,
		Device:     device,
		ServerData: string(marshalServerData),
	}

	_, err = models.AddSysLog(arrSysLog)

	if err != nil {
		models.ErrorLog("base:AddSysLog()", "AddSysLog():1", err.Error())
		return "something_went_wrong"
	}

	return ""
}

func getIP(r *http.Request) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}
