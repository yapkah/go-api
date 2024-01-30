package email_service

import (
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/setting"
)

// SendEmailByModules func
func SendEmailByModules(otp *models.OTP, langCode string, msgData map[string]interface{}) string {
	var (
		msgTitle, msgContent string
		emailTemplate        *models.EmailTemplate
		appName              string
		err                  error
	)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "email_template.type = ?", CondValue: otp.OtpType},
		models.WhereCondFn{Condition: "email_template.locale = ?", CondValue: langCode},
		models.WhereCondFn{Condition: "email_template.status = ?", CondValue: "A"},
	)
	emailTemplate, err = models.GetEmailTemplate(arrCond, "", false)
	if err != nil {
		base.LogErrorLog("emailService:SendEmailByModules()", "GetEmailTemplate():1", err.Error(), true)
		return "something_went_wrong"
	}

	if emailTemplate == nil {
		defaultLangCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "email_template.type = ?", CondValue: otp.OtpType},
			models.WhereCondFn{Condition: "email_template.locale = ?", CondValue: defaultLangCode},
			models.WhereCondFn{Condition: "email_template.status = ?", CondValue: "A"},
		)
		emailTemplate, err = models.GetEmailTemplate(arrCond, "", false)
		if err != nil {
			base.LogErrorLog("emailService:SendEmailByModules()", "GetEmailTemplate():2", err.Error(), true)
			return "something_went_wrong"
		}
		if emailTemplate == nil {
			base.LogErrorLog("emailService:SendEmailByModules()", "GetEmailTemplate():2", "email_template_not_found", true)
			return "something_went_wrong"
		}

		appName = helpers.Translate("app_name_email", defaultLangCode)
	} else {
		appName = helpers.Translate("app_name_email", langCode)
	}

	msgTitle = emailTemplate.Title
	msgContent = emailTemplate.Template

	if len(msgData) > 0 {
		msgData["project_name"] = appName
		msgContent, err = base.TemplateReplace(msgContent, msgData)
		if err != nil {
			base.LogErrorLog("emailService:SendEmailByModules()", "TemplateReplace():1", err.Error(), true)
			return "something_went_wrong"
		}
	}

	sendmailData := base.CallSendMailApiStruct{
		Subject:  msgTitle,
		Message:  msgContent,
		Type:     "HTML",
		FromName: appName,
		ToEmail:  []string{otp.ReceiverID},
		ToName:   []string{otp.ReceiverID},
	}

	// err = sendmailData.SendMail("test@iscity.com.my", "iscity@123456789", "box.iscity.com.my", "587")
	// err = sendmailData.SendMail("noreply@xtlegends.com", "fZci3szaCWzm", "box01.securelayers.cloud", "587")
	err = sendmailData.CallSendMailApi()
	if err != nil {
		base.LogErrorLog("emailService:SendEmailByModules()", "SendMail():1", err.Error(), true)
		return "something_went_wrong"
	}

	// get environment
	// var env = "DEV"
	// if setting.Cfg.Section("app").Key("Environment").String() == "LIVE" {
	// 	env = "LIVE"
	// }

	// // send email base on environment
	// if env == "LIVE" {
	// 	fmt.Println("run mailgun")
	// 	// EMAIL SETTING
	// 	// username
	// 	// password
	// 	// email host
	// 	// email port
	// } else {
	// 	sendmailData := mail.SendMailData{
	// 		Subject:  msgTitle,
	// 		Message:  msgContent,
	// 		Type:     "HTML",
	// 		FromName: appName,
	// 		ToEmail:  []string{otp.ReceiverID},
	// 		ToName:   []string{otp.ReceiverID},
	// 	}

	// 	err = sendmailData.SendMail("sec@securelayers.cloud", "fZci3szaCWzm", "box01.securelayers.cloud", "587")
	// 	if err != nil {
	// 		base.LogErrorLog("emailService:SendEmailByModules()", "SendMail():1", err.Error(), true)
	// 		return "something_went_wrong"
	// 	}
	// }

	return ""
}

// Send func
// func send(message string, msgTemplateID int, mobilenos []string) error {
// 	var destination []string
// 	var response DigitalMediaResponse

// 	if len(mobilenos) == 0 {
// 		return nil, ""
// 	}

// 	arrGeneralSetup, err := models.GetSysGeneralSetupByID("sms_setting")
// 	if err != nil {
// 		base.LogErrorLog("smsService:send()", "GetSysGeneralSetupByID():1", err.Error(), true)
// 		return nil, "something_went_wrong"
// 	}
// 	if arrGeneralSetup == nil {
// 		base.LogErrorLog("smsService:send()", "GetSysGeneralSetupByID():2", "sms_setting_not_found", true)
// 		return nil, "something_went_wrong"
// 	}

// 	smsSetting := &SmsSetting{}
// 	err = json.Unmarshal([]byte(arrGeneralSetup.SettingValue1), smsSetting)
// 	if err != nil {
// 		base.LogErrorLog("smsService:send()", "Unmarshal():1", err.Error(), true)
// 		return nil, "something_went_wrong"
// 	}

// 	enable, err := helpers.ValueToInt(smsSetting.Valid)
// 	if err != nil {
// 		base.LogErrorLog("smsService:send()", "ValueToInt():1", err.Error(), true)
// 		return nil, "something_went_wrong"
// 	}
// 	if enable == 0 {
// 		return nil, ""
// 	}

// 	var smsInput []map[string]interface{}

// 	for _, mobileno := range mobilenos {
// 		smsInput = append(smsInput, map[string]interface{}{
// 			"to":     mobileno,
// 			"source": "golang",
// 			"body":   message,
// 		})
// 	}

// 	input := map[string]interface{}{"messages": smsInput}

// 	key := util.EncodeBase64(smsSetting.Username + ":" + smsSetting.PrivateKey)
// 	url := smsSetting.URL
// 	header := map[string]string{
// 		"Content-Type":  "application/json",
// 		"Accept":        "application/json",
// 		"Authorization": fmt.Sprintf("Basic %s", key),
// 	}

// 	res, err := base.RequestAPI("POST", url, header, input, &response)

// 	if err != nil {
// 		base.LogErrorLog("smsService:send()", "RequestAPI():1", err.Error(), true)
// 		return nil, "something_went_wrong"
// 	}

// 	for _, mobileno := range mobilenos {
// 		destination = append(destination, mobileno)
// 	}

// 	to, _ := json.Marshal(destination)
// 	returnValue, _ := json.Marshal(res)

// 	smsLog := models.SmsLog{
// 		MobileNo:    string(to),
// 		TemplateID:  msgTemplateID,
// 		MsgContent:  message,
// 		ReturnValue: string(returnValue),
// 		API:         url,
// 	}

// 	db := models.GetDB() // no need transaction because if failed no need rollback

// 	_, err = models.AddSmsLog(db, smsLog)
// 	if err != nil {
// 		base.LogErrorLog("smsService:send()", "AddSmsLog():1", err.Error(), true)
// 		return nil, "something_went_wrong"
// 	}

// 	return &response, ""
// }
