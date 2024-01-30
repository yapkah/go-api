package sms_service

import (
	"encoding/json"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
)

// DigitalMediaResponse struct
type DigitalMediaResponse struct {
	DigitalMedia interface{} `json:"DigitalMedia"`
}

// SmsSetting struct
type SmsSetting struct {
	Valid      string `json:"valid"`
	URL        string `json:"url"`
	Username   string `json:"username"`
	PrivateKey string `json:"private_key"`
}

// SMSResponse struct
type SMSResponse struct {
	MediaType      string `json:"MediaType"`
	Message        string `json:"Message"`
	ResultID       string `json:"ResultID"`
	TotalRecipient string `json:"TotalRecipient"`
	TotalPage      string `json:"TotalPage"`
	Currency       string `json:"Currency"`
	AmountSpend    string `json:"AmountSpend"`
	SubmissionID   string `json:"SubmissionID"`
	Result         string `json:"Result"`
}

// Send func
func send(appName, message string, msgTemplateID int, mobilenos []string) (*DigitalMediaResponse, string) {
	var response DigitalMediaResponse

	if len(mobilenos) == 0 {
		return nil, ""
	}

	arrGeneralSetup, err := models.GetSysGeneralSetupByID("sms_setting")
	if err != nil {
		base.LogErrorLog("smsService:send()", err.Error(), "sms_setting", true)
		return nil, "something_went_wrong"
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("smsService:send()", "GetSysGeneralSetupByID():2", "sms_setting_not_found", true)
		return nil, "something_went_wrong"
	}

	smsSetting := &SmsSetting{}
	err = json.Unmarshal([]byte(arrGeneralSetup.SettingValue1), smsSetting)
	if err != nil {
		base.LogErrorLog("smsService:send()", "Unmarshal():1", err.Error(), true)
		return nil, "something_went_wrong"
	}

	enable, err := helpers.ValueToInt(smsSetting.Valid)
	if err != nil {
		base.LogErrorLog("smsService:send()", "ValueToInt():1", err.Error(), true)
		return nil, "something_went_wrong"
	}
	if enable == 0 {
		return nil, ""
	}

	for _, mobileno := range mobilenos {
		input := map[string]interface{}{
			"api_key":    smsSetting.Username,
			"api_secret": smsSetting.PrivateKey,
			"type":       "unicode",
			"to":         mobileno,
			"from":       appName,
			"text":       message,
		}

		url := smsSetting.URL
		header := map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		}

		res, err := base.RequestAPI("GET", url, header, input, &response)

		if err != nil {
			base.LogErrorLog("smsService:send()", "RequestAPI():1", err.Error(), true)
			return nil, "something_went_wrong"
		}

		returnValue, _ := json.Marshal(res)

		smsLog := models.SmsLog{
			MobileNo:    mobileno,
			TemplateID:  msgTemplateID,
			MsgContent:  message,
			ReturnValue: string(returnValue),
			API:         url,
		}

		db := models.GetDB() // no need transaction because if failed no need rollback

		_, err = models.AddSmsLog(db, smsLog)
		if err != nil {
			base.LogErrorLog("smsService:send()", "AddSmsLog():1", err.Error(), true)
			return nil, "something_went_wrong"
		}
	}

	return &response, ""
}
