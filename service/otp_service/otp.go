package otp_service

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/yapkah/go-api/service/email_service"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/service/sms_service"
)

// OTP struct
type OTP struct {
	SendType   string `json:"send_type`
	ReceiverID string `json:"mobile_no"` // mobile_no/email
	OtpType    string `json:"otp_type"`  // REG: register | RP: Reset Password
	LangCode   string `json:"lang_code"`
}

// OtpSetting struct
type OtpSetting struct {
	MaxRequest     string `json:"max_request"`
	MaxRequestTime string `json:"max_request_time"`
	OtpExpiryTime  string `json:"otp_expiry_time"`
	FixedOtp       string `json:"fixed_otp"`
	SendOtp        bool   `json:"send_otp"`
}

// SendOTP send sms otp
func (otpData *OTP) SendOTP(tx *gorm.DB, msgData map[string]interface{}) (int, string) {
	var (
		errMsg string
	)

	// validate send type
	if helpers.StringInSlice(otpData.SendType, []string{"MOBILE", "EMAIL"}) == false {
		base.LogErrorLog("otpService:SendOTP()", "invalid_send_type", otpData, true)
		return 0, "something_went_wrong"
	}

	// validate otp type
	if helpers.StringInSlice(otpData.OtpType, []string{"REGISTER", "PASSWORD", "PIN", "MOBILE", "EMAIL", "WITHDRAW"}) == false {
		base.LogErrorLog("otpService:SendOTP()", e.GetMsg(e.SMS_INVALID_MODULE), "", true)
		return 0, "something_went_wrong"
	}

	// check otp limit
	errMsg = OtpLimitChecking(otpData)
	if errMsg != "" {
		return 0, errMsg
	}

	// generate otp
	otp, countDownSec, sendOtp, errMsg := Generate(tx, otpData)

	if errMsg != "" {
		return 0, errMsg
	}

	data := map[string]interface{}{"otp": otp.Otp}

	for k, d := range msgData {
		data[k] = d
	}

	if sendOtp {
		if otpData.SendType == "MOBILE" { // do send via mobile
			errMsg = sms_service.SendSmsByModules(otp, otpData.LangCode, data)
		} else if otpData.SendType == "EMAIL" { // do send via email
			errMsg = email_service.SendEmailByModules(otp, otpData.LangCode, data)
		}
	}

	return countDownSec, errMsg
}

// OtpLimitChecking func
func OtpLimitChecking(otpData *OTP) string {
	arrGeneralSetup, err := models.GetSysGeneralSetupByID("otp_setting")
	if err != nil {
		base.LogErrorLog("otpService:OtpLimitChecking()", err.Error(), "otp_setting", true)
		return "something_went_wrong"
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("otpService:OtpLimitChecking()", "GetSysGeneralSetupByID():1", e.OTP_SETTING_NOT_FOUND, true)
		return "something_went_wrong"
	}

	otpSetting := &OtpSetting{}
	err = json.Unmarshal([]byte(arrGeneralSetup.SettingValue1), otpSetting)
	if err != nil {
		base.LogErrorLog("otpService:OtpLimitChecking()", "Unmarshal():1", err.Error(), true)
		return "something_went_wrong"
	}

	maxReq, err := helpers.ValueToInt(otpSetting.MaxRequest)
	if err != nil {
		base.LogErrorLog("otpService:OtpLimitChecking()", "ValueToInt():1", err.Error(), true)
		return "something_went_wrong"
	}
	maxReqTime, err := helpers.ValueToDuration(otpSetting.MaxRequestTime)
	if err != nil {
		base.LogErrorLog("otpService:OtpLimitChecking()", "ValueToDuration():1", err.Error(), true)
		return "something_went_wrong"
	}

	date := base.GetCurrentDateTimeT().Add(-(maxReqTime * time.Minute))

	count, err := models.GetOtpByTimeCount(otpData.ReceiverID, otpData.OtpType, date)
	if err != nil {
		base.LogErrorLog("otpService:OtpLimitChecking()", "GetOtpByTimeCount():1", err.Error(), true)
		return "something_went_wrong"
	}

	if count >= maxReq {
		return e.GetMsg(e.OTP_EXCEED_MAX_REQUEST)
	}

	return ""
}

// Generate otp
func Generate(tx *gorm.DB, otpData *OTP) (*models.OTP, int, bool, string) {
	var (
		otp string
		err error
	)

	// generate otp
	arrGeneralSetup, err := models.GetSysGeneralSetupByID("otp_setting")
	if err != nil {
		base.LogErrorLog("otpService:Generate()", err.Error(), "otp_setting", true)
		return nil, 0, false, "something_went_wrong"
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("otpService:Generate()", "GetSysGeneralSetupByID():1", e.OTP_SETTING_NOT_FOUND, true)
		return nil, 0, false, "something_went_wrong"
	}
	otpSetting := &OtpSetting{}
	err = json.Unmarshal([]byte(arrGeneralSetup.SettingValue1), otpSetting)
	if err != nil {
		base.LogErrorLog("otpService:Generate()", "Unmarshal():1", err.Error(), true)
		return nil, 0, false, "something_went_wrong"
	}

	if otpSetting.FixedOtp != "" {
		otp = otpSetting.FixedOtp
	} else {
		rand.Seed(time.Now().UnixNano()) // to prevent same number every time restart
		otp = strconv.Itoa(rand.Intn(899999) + 100000)
	}

	// get expire time
	expiryTime, countDownSec, errMsg := getOTPExpireTime()
	if errMsg != "" {
		base.LogErrorLog("otpService:Generate()", "getOTPExpireTime():1", err.Error(), true)
		return nil, 0, false, "something_went_wrong"
	}

	arrSmsOtp := models.OTP{
		SendType:   otpData.SendType,
		ReceiverID: otpData.ReceiverID,
		OtpType:    otpData.OtpType,
		Otp:        string(otp),
		BValid:     1,
		ExpiredAt:  expiryTime,
	}

	addedSmsOtp, err := models.AddOTP(tx, arrSmsOtp)
	if err != nil {
		base.LogErrorLog("otpService:Generate()", "AddOTP():1", err.Error(), true)
		return nil, 0, false, "something_went_wrong"
	}

	// update otp.b_valid
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id != ?", CondValue: addedSmsOtp.ID},
		models.WhereCondFn{Condition: "receiver_id = ?", CondValue: otpData.ReceiverID},
		models.WhereCondFn{Condition: "otp_type = ?", CondValue: otpData.OtpType},
	)
	updateColumn := map[string]interface{}{"b_valid": 0}
	err = models.UpdatesFnTx(tx, "otp", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("otpService:Generate()", "UpdatesFnTx():1", err.Error(), true)
		return nil, 0, false, "something_went_wrong"
	}

	return addedSmsOtp, countDownSec, otpSetting.SendOtp, ""
}

// getOTPExpireTime get otp duration
func getOTPExpireTime() (time.Time, int, string) {
	arrGeneralSetup, err := models.GetSysGeneralSetupByID("otp_setting")
	if err != nil {
		base.LogErrorLog("otpService:getOTPExpireTime()", err.Error(), "otp_setting", true)
		return time.Time{}, 0, "something_went_wrong"
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("otpService:getOTPExpireTime()", "GetSysGeneralSetupByID():1", e.OTP_SETTING_NOT_FOUND, true)
		return time.Time{}, 0, "something_went_wrong"
	}

	otpSetting := &OtpSetting{}
	err = json.Unmarshal([]byte(arrGeneralSetup.SettingValue1), otpSetting)
	if err != nil {
		base.LogErrorLog("otpService:getOTPExpireTime()", "Unmarshal():1", err.Error(), true)
		return time.Time{}, 0, "something_went_wrong"
	}

	du, err := helpers.ValueToDuration(otpSetting.OtpExpiryTime)
	if err != nil {
		base.LogErrorLog("otpService:getOTPExpireTime()", "ValueToDuration():1", err.Error(), true)
		return time.Time{}, 0, "something_went_wrong"
	}

	countDownSec := int((du * time.Minute) / time.Second)

	return base.GetCurrentDateTimeT().Add(du * time.Minute), countDownSec, ""
}

// ValidateOTP func
func ValidateOTP(tx *gorm.DB, sendType, receiverID, inputOtp, otpType string) string {
	// validate otp
	inputOtp = strings.Trim(inputOtp, " ")

	arrSmsOtpCond := make([]models.WhereCondFn, 0)
	arrSmsOtpCond = append(arrSmsOtpCond,
		models.WhereCondFn{Condition: "otp.send_type = ?", CondValue: sendType},
		models.WhereCondFn{Condition: "otp.receiver_id = ?", CondValue: receiverID},
		models.WhereCondFn{Condition: "otp.otp_type = ?", CondValue: otpType},
		models.WhereCondFn{Condition: "otp.b_valid = ?", CondValue: 1},
	)
	otp, err := models.GetOtpFn(arrSmsOtpCond, "", false)

	if err != nil {
		base.LogErrorLog("otpService:ValidateOTP()", "GetOtpFn():1", err.Error(), true)
		return "something_went_wrong"
	}
	if otp == nil {
		return e.GetMsg(e.PLEASE_REQUEST_OTP)
	}

	// validate otp
	_, errMsg := otp.Validate(tx, inputOtp)
	if errMsg != "" {
		models.Rollback(tx)
		return errMsg
	}

	return ""
}
