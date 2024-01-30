package member

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/mobile_service"
	"github.com/yapkah/go-api/service/otp_service"
)

// RequestOTPForm struct
type RequestOTPForm struct {
	Email        string `form:"email" json:"email"`
	MobilePrefix string `form:"mobile_prefix" json:"mobile_prefix"`
	MobileNo     string `form:"mobile_no" json:"mobile_no"`
	OtpType      string `form:"otp_type" json:"otp_type" valid:"Required;MaxSize(50)"`
}

// RequestOTP function for verification without access token
func RequestOTP(c *gin.Context) {
	var (
		appG                                     = app.Gin{C: c}
		form                                     RequestOTPForm
		err                                      error
		receiverID, sendType, successMsg, errMsg string
		allowMobile                              bool = false
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// check for supported otp type
	if helpers.StringInSlice(form.OtpType, []string{"REGISTER", "PASSWORD", "PIN", "MOBILE", "EMAIL", "WITHDRAW"}) == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_otp_type"}, nil)
		return
	}

	if allowMobile && (form.OtpType == "MOBILE" || (form.OtpType != "EMAIL" && form.MobileNo != "")) {
		// validate mobile prefix
		if form.MobilePrefix == "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_mobile_prefix"}, nil)
			return
		}

		arrSysTerritoryFn := make([]models.WhereCondFn, 0)
		arrSysTerritoryFn = append(arrSysTerritoryFn,
			models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: form.MobilePrefix},
		)
		arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

		if err != nil {
			base.LogErrorLog("otpController:RequestOTP()", "GetSysTerritoryFn():1", err.Error(), true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if arrSysTerritory == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_mobile_prefix"}, nil)
			return
		}

		countryCode := arrSysTerritory.Code

		mobileno := strings.Trim(form.MobileNo, " ")

		// validate mobile_no format
		num, errMsg := mobile_service.ParseMobileNo(mobileno, countryCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		if helpers.StringInSlice(form.OtpType, []string{"REGISTER", "MOBILE", "EMAIL"}) {
			mobilePrefix := fmt.Sprintf("%v", *num.CountryCode)

			// validate if unique
			ok, err := models.ExistsMemberByMobile(mobilePrefix, mobileno)
			if err != nil {
				models.ErrorLog("otpController:RequestOTP()", "ExistsMemberByMobile():1", err.Error())
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
			if ok {
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_MOBILE_EXISTS)}, nil)
				return
			}
		}

		// setting up for sendType, receiverID, successMSg
		receiverID = mobile_service.E164FormatWithouSymbol(num) // get international mobile no
		sendType = "MOBILE"
		successMsg = "please_check_mobile_for_verification_code"

	} else if form.OtpType == "EMAIL" || (form.OtpType != "MOBILE" && form.Email != "") {
		// validate email format
		valid := validation.Validation{}
		valid.Email(form.Email, "email")

		if valid.HasErrors() {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: valid.Errors[0].String()}, nil)
			return
		}

		if helpers.StringInSlice(form.OtpType, []string{"REGISTER", "MOBILE", "EMAIL"}) {
			// validate if email is unique
			ok, err := models.ExistsMemberByEmail(form.Email)
			if err != nil {
				models.ErrorLog("otpController:RequestOTP()", "ExistsMemberByMobile():1", err.Error())
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
			if ok {
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_EMAIL_EXISTS)}, nil)
				return
			}
		}

		// setting up for sendType, receiverID, successMSg
		receiverID = form.Email
		sendType = "EMAIL"
		successMsg = "please_check_email_for_verification_code"
	} else {
		if allowMobile {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_either_mobile_no_or_email"}, nil)
			return
		} else {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_email"}, nil)
			return
		}
	}

	// being transaction
	tx := models.Begin()

	// send otp via sms/email
	otpService := otp_service.OTP{
		SendType:   sendType,
		ReceiverID: receiverID,
		OtpType:    form.OtpType,
		LangCode:   langCode,
	}

	countDownSec, errMsg := otpService.SendOTP(tx, nil)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("otpController:RequestOTP()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: successMsg}, map[string]interface{}{"count_down_seconds": countDownSec})
	return
}
