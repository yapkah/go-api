package sms_service

import (
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
)

// SendSmsByModules func
func SendSmsByModules(smsOtp *models.OTP, langCode string, msgData map[string]interface{}) string {
	var (
		msgContent  string
		smsTemplate *models.SmsTemplate
		appName     string
		err         error
		errMsg      string
	)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "sms_template.type = ?", CondValue: smsOtp.OtpType},
		models.WhereCondFn{Condition: "sms_template.locale = ?", CondValue: langCode},
		models.WhereCondFn{Condition: "sms_template.status = ?", CondValue: "A"},
	)
	smsTemplate, err = models.GetSmsTemplate(arrCond, "", false)
	if err != nil {
		base.LogErrorLog("smsservice:SendSmsByModules()", "GetSmsTemplate():1", err.Error(), true)
		return "something_went_wrong"
	}
	if smsTemplate == nil {
		defaultLangCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "sms_template.type = ?", CondValue: smsOtp.OtpType},
			models.WhereCondFn{Condition: "sms_template.locale = ?", CondValue: defaultLangCode},
			models.WhereCondFn{Condition: "sms_template.status = ?", CondValue: "A"},
		)
		smsTemplate, err = models.GetSmsTemplate(arrCond, "", false)
		if err != nil {
			base.LogErrorLog("smsservice:SendSmsByModules()", "GetSmsTemplate():2", err.Error(), true)
			return "something_went_wrong"
		}
		if smsTemplate == nil {
			base.LogErrorLog("smsservice:SendSmsByModules()", "GetSmsTemplate():2", e.SMS_LABEL_NOT_FOUND, true)
			return "something_went_wrong"
		}

		appName = helpers.Translate("app_name_sms", defaultLangCode)
	} else {
		appName = helpers.Translate("app_name_sms", langCode)
	}

	msgContent = appName + " " + smsTemplate.Template

	if len(msgData) > 0 {
		msgContent, err = base.TemplateReplace(msgContent, msgData)
		if err != nil {
			base.LogErrorLog("smsservice:SendSmsByModules()", "TemplateReplace():1", err.Error(), true)
			return "something_went_wrong"
		}
	}

	_, errMsg = send(appName, msgContent, smsTemplate.ID, []string{smsOtp.ReceiverID})

	return errMsg
}
