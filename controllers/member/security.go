package member

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/mobile_service"
	"github.com/yapkah/go-api/service/otp_service"
)

// UpdatePasswordForm struct
type UpdatePasswordForm struct {
	CurrentPassword string `form:"current_password" json:"current_password" valid:"Required;"`
	Password        string `form:"password" json:"password" valid:"Required;"`
}

// UpdatePassword function
func UpdatePassword(c *gin.Context) {
	var (
		appG   = app.Gin{C: c}
		form   UpdatePasswordForm
		err    error
		errMsg string
	)

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// validate inputs
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	memberID := member.ID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.CurrentPassword)
	if err != nil {
		base.LogErrorLog("UpdatePassword-RsaDecryptPKCS1v15_CurrentPassword_Failed", err.Error(), form.CurrentPassword, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_current_password"}, nil)
		return
	}
	form.CurrentPassword = decryptedText

	decryptedText, err = util.RsaDecryptPKCS1v15(form.Password)
	if err != nil {
		base.LogErrorLog("UpdatePassword-RsaDecryptPKCS1v15_Password_Failed", err.Error(), form.Password, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_new_password_format"}, nil)
		return
	}
	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "password_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.Password = decryptedText

	// begin transaction
	tx := models.Begin()

	// update password
	memberPassword := member_service.MemberPassword{
		MemberID:            memberID,
		CurrentPassword:     form.CurrentPassword,
		Password:            form.Password,
		CurrentPasswordHash: member.Password,
	}

	// rollback if hit error
	errMsg = memberPassword.UpdateMemberPassword(tx, false)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("securityController:UpdatePassword()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	base.AddSysLog(memberID, make(map[string]interface{}), make(map[string]interface{}), "modify", "update-password", c)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "updated_successfully"}, nil)
	return
}

// UpdateSecondaryPasswordForm struct
type UpdateSecondaryPasswordForm struct {
	CurrentSecondaryPin string `form:"current_secondary_pin" json:"current_secondary_pin" valid:"Required;"`
	SecondaryPin        string `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

// UpdateSecondaryPassword function
func UpdateSecondaryPassword(c *gin.Context) {
	var (
		appG   = app.Gin{C: c}
		form   UpdateSecondaryPasswordForm
		err    error
		errMsg string
	)

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// validate inputs
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	memberID := member.ID
	decryptedText, err := util.RsaDecryptPKCS1v15(form.CurrentSecondaryPin)
	if err != nil {
		base.LogErrorLog("UpdateSecondaryPassword-RsaDecryptPKCS1v15_CurrentSecondaryPin_Failed", err.Error(), form.CurrentSecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_current_secondary_pin"}, nil)
		return
	}
	form.CurrentSecondaryPin = decryptedText

	decryptedText, err = util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("UpdateSecondaryPassword-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_new_secondary_pin_format"}, nil)
		return
	}
	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "SecondaryPin minimum character is 6"}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// begin transaction
	tx := models.Begin()

	// update secondary pin
	memberSecondaryPin := member_service.MemberSecondaryPin{
		MemberID:               memberID,
		CurrentSecondaryPin:    form.CurrentSecondaryPin,
		SecondaryPin:           form.SecondaryPin,
		CurrentSecondaryPinMd5: member.SecondaryPin,
	}

	// rollback if hit error
	errMsg = memberSecondaryPin.UpdateMemberSecondaryPin(tx, false)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("securityController:UpdateSecondaryPassword()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	base.AddSysLog(memberID, make(map[string]interface{}), make(map[string]interface{}), "modify", "update-secondary-pin", c)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "updated_successfully"}, nil)
	return
}

// ResetPasswordForm struct
type ResetPasswordForm struct {
	Email            string `form:"email" json:"email"`
	MobilePrefix     string `form:"mobile_prefix" json:"mobile_prefix"`
	MobileNo         string `form:"mobile_no" json:"mobile_no"`
	VerificationCode string `form:"verification_code" json:"verification_code" valid:"Required;MaxSize(20)"`
	Password         string `form:"password" json:"password" valid:"Required"`
}

// ResetPassword reset password
func ResetPassword(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form ResetPasswordForm
		// err    error
		receiverID, sendType, credentialNotFoundMsg, errMsg string
		allowMobile                                         bool = false
	)

	// validate inputs
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	ok = models.ExistLangague(c.GetHeader("Accept-Language"))
	// 	if ok {
	// 		langCode = c.GetHeader("Accept-Language")
	// 	}
	// }

	// check which flow to perform (phone/email)
	if allowMobile {
		if (form.MobileNo == "" && form.Email == "") || (form.MobileNo != "" && form.Email != "") {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_either_mobile_no_or_email"}, nil)
			return
		}
	} else if form.Email == "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_email"}, nil)
		return
	}

	// find member by mobile/email
	arrMemberFn := make([]models.WhereCondFn, 0)
	arrMemberFn = append(arrMemberFn,
		models.WhereCondFn{Condition: "members.status != ?", CondValue: "T"}, // exclude terminated login account
	)

	if allowMobile && form.MobileNo != "" {
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
			base.LogErrorLog("securityController:ResetPassword()", "GetSysTerritoryFn():1", err.Error(), true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if arrSysTerritory == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_mobile_prefix"}, nil)
			return
		}

		countryCode := arrSysTerritory.Code

		mobileNo := strings.Trim(form.MobileNo, " ")

		// validate mobile_no format
		num, errMsg := mobile_service.ParseMobileNo(mobileNo, countryCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		mobilePrefix := fmt.Sprintf("%v", *num.CountryCode)

		arrMemberFn = append(arrMemberFn,
			models.WhereCondFn{Condition: "members.mobile_prefix = ?", CondValue: mobilePrefix},
			models.WhereCondFn{Condition: "members.mobile_no = ?", CondValue: mobileNo},
		)

		// setting up for sendType, receiverID, successMSg
		receiverID = mobile_service.E164FormatWithouSymbol(num) // get international mobile no
		sendType = "MOBILE"
		credentialNotFoundMsg = "mobile_no_not_registered"

	} else if form.Email != "" {
		// validate email format
		valid := validation.Validation{}
		valid.Email(form.Email, "email")

		if valid.HasErrors() {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: valid.Errors[0].String()}, nil)
			return
		}

		arrMemberFn = append(arrMemberFn,
			models.WhereCondFn{Condition: "members.email = ?", CondValue: form.Email},
		)

		// setting up for sendType, receiverID, successMSg
		receiverID = form.Email
		sendType = "EMAIL"
		credentialNotFoundMsg = "email_not_registered"
	}

	arrMembers, err := models.GetMembersFn(arrMemberFn, false)
	if err != nil {
		base.LogErrorLog("securityController:ResetPassword()", "GetMembersFn():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	if arrMembers == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: credentialNotFoundMsg}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	// validate otp
	inputOtp := strings.Trim(form.VerificationCode, " ")

	errMsg = otp_service.ValidateOTP(tx, sendType, receiverID, inputOtp, "PASSWORD")
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	decryptedText, err := util.RsaDecryptPKCS1v15(form.Password)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("ResetPassword-RsaDecryptPKCS1v15_Password_Failed", err.Error(), form.Password, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_new_password_format"}, nil)
		return
	}
	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "Password minimum character is 6"}, nil)
		return
	}
	form.Password = decryptedText

	// update password
	memberPassword := member_service.MemberPassword{
		MemberID: arrMembers.ID,
		Password: form.Password,
	}

	// rollback if hit error
	errMsg = memberPassword.UpdateMemberPassword(tx, true)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("securityController:ResetPassword()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// member reset password log
	base.AddSysLog(arrMembers.ID, make(map[string]interface{}), make(map[string]interface{}), "modify", "reset-password", c)

	// login
	// get called platform
	// route := c.Request.URL.String()
	// platformCheckingRst := strings.Contains(route, "/api/app")
	// platform := "HTMLFIVE"
	// if platformCheckingRst {
	// 	platform = "APP"
	// }

	// find member login details
	// arrEntMemberMember, _ := models.GetEntMemberMemberFn(arrMemberFn, false)

	// db := models.GetDB() // no need transaction because if failed no need rollback
	// processRst, err := member_service.ProcessMemberLogin(db, arrEntMemberMember, langCode, platform, 0)
	// if err != nil {
	// 	base.LogErrorLog("securityController:ResetPassword()", "ProcessMemberLogin():1", err.Error(), true)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
	// 	return
	// }

	// tk := map[string]interface{}{
	// 	"access_token": processRst.AccessToken,
	// }

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// ResetSecondaryPinForm struct
type ResetSecondaryPinForm struct {
	Email            string `form:"email" json:"email"`
	MobilePrefix     string `form:"mobile_prefix" json:"mobile_prefix"`
	MobileNo         string `form:"mobile_no" json:"mobile_no"`
	VerificationCode string `form:"verification_code" json:"verification_code" valid:"Required;MaxSize(20)"`
	SecondaryPin     string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// ResetSecondaryPin reset password
func ResetSecondaryPin(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form ResetSecondaryPinForm
		// err    error
		receiverID, sendType, credentialNotFoundMsg, errMsg string
	)

	// validate inputs
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// check which flow to perform (phone/email)
	if (form.MobileNo == "" && form.Email == "") || (form.MobileNo != "" && form.Email != "") {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_either_mobile_no_or_email"}, nil)
		return
	}

	// find member by mobile/email
	arrMemberFn := make([]models.WhereCondFn, 0)
	arrMemberFn = append(arrMemberFn,
		models.WhereCondFn{Condition: "members.status != ?", CondValue: "T"}, // exclude terminated login account
	)

	if form.MobileNo != "" {
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
			base.LogErrorLog("securityController:ResetSecondaryPin()", "GetSysTerritoryFn():1", err.Error(), true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if arrSysTerritory == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_mobile_prefix"}, nil)
			return
		}

		countryCode := arrSysTerritory.Code

		mobileNo := strings.Trim(form.MobileNo, " ")

		// validate mobile_no format
		num, errMsg := mobile_service.ParseMobileNo(mobileNo, countryCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		mobilePrefix := fmt.Sprintf("%v", *num.CountryCode)

		arrMemberFn = append(arrMemberFn,
			models.WhereCondFn{Condition: "members.mobile_prefix = ?", CondValue: mobilePrefix},
			models.WhereCondFn{Condition: "members.mobile_no = ?", CondValue: mobileNo},
		)

		// setting up for sendType, receiverID, successMSg
		receiverID = mobile_service.E164FormatWithouSymbol(num) // get international mobile no
		sendType = "MOBILE"
		credentialNotFoundMsg = "mobile_no_not_registered"

	} else if form.Email != "" {
		// validate email format
		valid := validation.Validation{}
		valid.Email(form.Email, "email")

		if valid.HasErrors() {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: valid.Errors[0].String()}, nil)
			return
		}

		arrMemberFn = append(arrMemberFn,
			models.WhereCondFn{Condition: "members.email = ?", CondValue: form.Email},
		)

		// setting up for sendType, receiverID, successMSg
		receiverID = form.Email
		sendType = "EMAIL"
		credentialNotFoundMsg = "email_not_registered"
	}

	arrMembers, err := models.GetMembersFn(arrMemberFn, false)
	if err != nil {
		base.LogErrorLog("securityController:ResetSecondaryPin()", "GetMembersFn():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	if arrMembers == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: credentialNotFoundMsg}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	// validate otp
	inputOtp := strings.Trim(form.VerificationCode, " ")

	errMsg = otp_service.ValidateOTP(tx, sendType, receiverID, inputOtp, "PIN")
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("ResetSecondaryPin-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_new_secondary_pin_format"}, nil)
		return
	}
	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "SecondaryPin minimum character is 6"}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// update secondary pin
	memberPassword := member_service.MemberSecondaryPin{
		MemberID:     arrMembers.ID,
		SecondaryPin: form.SecondaryPin,
	}

	// rollback if hit error
	errMsg = memberPassword.UpdateMemberSecondaryPin(tx, true)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("securityController:ResetSecondaryPin()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// member reset transaction pin log
	base.AddSysLog(arrMembers.ID, make(map[string]interface{}), make(map[string]interface{}), "modify", "reset-secondary-pin", c)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "reset_successfully"}, nil)
	return
}

// CheckSecondaryPasswordForm struct
type CheckSecondaryPasswordForm struct {
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// CheckSecondaryPasswordv1 function
func CheckSecondaryPasswordv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form CheckSecondaryPasswordForm
	)

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// validate inputs
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("CheckSecondaryPasswordv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary pin
	pinValidation := base.SecondaryPin{
		MemId:              member.ID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// ResetPasswordWithHashedPrivateKeyForm struct
type ResetPasswordWithHashedPrivateKeyForm struct {
	PrivateKey string `form:"private_key" json:"private_key" valid:"Required;"`
	Password   string `form:"password" json:"password" valid:"Required"`
}

// ResetPasswordWithHashedPrivateKey reset password
func ResetPasswordWithHashedPrivateKey(c *gin.Context) {
	var (
		appG   = app.Gin{C: c}
		form   ResetPasswordWithHashedPrivateKeyForm
		err    error
		errMsg string
	)

	// validate inputs
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

	privateKey, err := util.RsaDecryptPKCS1v15(form.PrivateKey)
	if err != nil {
		base.LogErrorLog("securityController:ResetPasswordWithHashedPrivateKey()|RsaDecryptPKCS1v15():1", err.Error(), form.PrivateKey, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_private_key_format"}, nil)
		return
	}

	// get ent_member by hashed_private_key
	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.d_pk = ?", CondValue: privateKey},
	)
	arrEntMember, err := models.GetEntMemberFn(arrEntMemberFn, "", false)
	if err != nil {
		base.LogErrorLog("securityController:ResetPasswordWithHashedPrivateKey()", "GetEntMemberFn():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	if arrEntMember == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_private_key"}, nil)
		return
	}

	mainID := arrEntMember.MainID

	// get members by main_id
	arrMemberFn := make([]models.WhereCondFn, 0)
	arrMemberFn = append(arrMemberFn,
		models.WhereCondFn{Condition: "members.id = ?", CondValue: mainID}, // exclude terminated login account
	)
	arrMembers, err := models.GetMembersFn(arrMemberFn, false)
	if err != nil {
		base.LogErrorLog("securityController:ResetPasswordWithHashedPrivateKey()", "GetMembersFn():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	if arrMembers == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_private_key"}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	decryptedText, err := util.RsaDecryptPKCS1v15(form.Password)
	if err != nil {
		base.LogErrorLog("securityController:ResetPasswordWithHashedPrivateKey()|RsaDecryptPKCS1v15():2", err.Error(), form.Password, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_new_password_format"}, nil)
		return
	}
	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "Password minimum character is 6"}, nil)
		return
	}
	form.Password = decryptedText

	// update password
	memberPassword := member_service.MemberPassword{
		MemberID: mainID,
		Password: form.Password,
	}

	// rollback if hit error
	errMsg = memberPassword.UpdateMemberPassword(tx, true)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// switch current profile
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "main_id = ?", CondValue: mainID},
	)
	updateColumn := map[string]interface{}{"current_profile": 0, "updated_by": mainID}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		models.ErrorLog("securityController:ResetPasswordWithHashedPrivateKey()", "UpdatesFnTx():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	arrUpdCond2 := make([]models.WhereCondFn, 0)
	arrUpdCond2 = append(arrUpdCond2,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrEntMember.ID},
	)
	updateColumn2 := map[string]interface{}{"current_profile": 1, "updated_by": arrEntMember.ID}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond2, updateColumn2, false)
	if err != nil {
		models.ErrorLog("securityController:ResetPasswordWithHashedPrivateKey()", "UpdatesFnTx():2", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("securityController:ResetPasswordWithHashedPrivateKey()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// member reset password log
	base.AddSysLog(mainID, make(map[string]interface{}), make(map[string]interface{}), "modify", "reset-password", c)

	// login
	// get called platform
	route := c.Request.URL.String()
	platformCheckingRst := strings.Contains(route, "/api/app")
	platform := "HTMLFIVE"
	if platformCheckingRst {
		platform = "APP"
	}

	// get member login details
	arrEntMemberMember, _ := models.GetEntMemberMemberFn(arrMemberFn, false)

	db := models.GetDB() // no need transaction because if failed no need rollback
	processRst, err := member_service.ProcessMemberLogin(db, arrEntMemberMember, langCode, platform, 0)
	if err != nil {
		base.LogErrorLog("securityController:ResetPasswordWithHashedPrivateKey()", "ProcessMemberLogin():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	tk := map[string]interface{}{
		"access_token": processRst.AccessToken,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, tk)
	return
}

// RequestForgetPasswordForm struct
// type RequestForgetPasswordForm struct {
// 	Email        string `form:"email" json:"email"`
// 	MobilePrefix string `form:"mobile_prefix" json:"mobile_prefix"`
// 	MobileNo     string `form:"mobile_no" json:"mobile_no"`
// 	ReqType      string `form:"request_type" json:"request_type" valid:"Required;MaxSize(50)"`
// }

// // RequestForgetPassword function for
// func RequestForgetPassword(c *gin.Context) {
// 	var (
// 		appG                                     = app.Gin{C: c}
// 		form                                     RequestForgetPasswordForm
// 		err                                      error
// 		receiverID, sendType, successMsg, errMsg string
// 		membersID                                 int
// 	)

// 	ok, msg := app.BindAndValid(c, &form)
// 	if ok == false {
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
// 		return
// 	}

// 	// get lang code
// 	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
// 	if c.GetHeader("Accept-Language") != "" {
// 		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
// 		if ok {
// 			langCode = c.GetHeader("Accept-Language")
// 		}
// 	}

// 	if form.ReqType == "MOBILE" || form.ReqType == "EMAIL" {
// 		// checking for otp request to set mobile/email
// 		if form.ReqType == "MOBILE" && form.MobileNo == "" {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_mobile_no"}, nil)
// 			return
// 		}
// 		if form.ReqType == "EMAIL" && form.Email == "" {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_email"}, nil)
// 			return
// 		}
// 	} else {
// 		// check which flow to perform (phone/email)
// 		if (form.MobileNo == "" && form.Email == "") || (form.MobileNo != "" && form.Email != "") {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_either_mobile_no_or_email"}, nil)
// 			return
// 		}
// 	}

// 	if form.ReqType == "MOBILE" || (form.ReqType != "EMAIL" && form.MobileNo != "") {
// 		// validate mobile prefix
// 		if form.MobilePrefix == "" {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_mobile_prefix"}, nil)
// 			return
// 		}

// 		arrSysTerritoryFn := make([]models.WhereCondFn, 0)
// 		arrSysTerritoryFn = append(arrSysTerritoryFn,
// 			models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: form.MobilePrefix},
// 		)
// 		arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

// 		if err != nil {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 			return
// 		}

// 		if arrSysTerritory == nil {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_mobile_prefix"}, nil)
// 			return
// 		}

// 		countryCode := arrSysTerritory.Code

// 		mobileno := strings.Trim(form.MobileNo, " ")

// 		// validate mobile_no format
// 		num, errMsg := mobile_service.ParseMobileNo(mobileno, countryCode)
// 		if errMsg != "" {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
// 			return
// 		}

// 		if helpers.StringInSlice(form.ReqType, []string{"REGISTER", "MOBILE", "EMAIL"}) {
// 			mobilePrefix := fmt.Sprintf("%v", *num.CountryCode)

// 			// validate if unique
// 			ok, err := models.ExistsMemberByMobile(mobilePrefix, mobileno)
// 			if err != nil {
// 				models.ErrorLog("otpController:RequestOTP()", "ExistsMemberByMobile():1", err.Error())
// 				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 				return
// 			}
// 			if ok {
// 				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_MOBILE_EXISTS)}, nil)
// 				return
// 			}
// 		}

// 		// setting up for sendType, receiverID, successMSg
// 		receiverID = mobile_service.E164FormatWithouSymbol(num) // get international mobile no
// 		sendType = "MOBILE"
// 		successMsg = "please_kindly_check_mobile"

// 	} else if form.ReqType == "EMAIL" || (form.ReqType != "MOBILE" && form.Email != "") {
// 		// validate email format
// 		valid := validation.Validation{}
// 		valid.Email(form.Email, "email")

// 		if valid.HasErrors() {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: valid.Errors[0].String()}, nil)
// 			return
// 		}

// 		// validate if email is unique
// 		arrFn := make([]models.WhereCondFn, 1)
// 		arrFn = append(arrFn,
// 			models.WhereCondFn{Condition: "members.status IN ('A','I') AND members.email = ?", CondValue: form.Email},
// 		)
// 		members, err := models.GetEntMemberMemberFn(arrFn, false)
// 		if err != nil {
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 			return
// 		}
// 		if members == nil {
// 			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
// 			return
// 		}

// 		// setting up for sendType, receiverID, successMSg
// 		membersID = members.ID
// 		receiverID = form.Email
// 		sendType = "EMAIL"
// 		successMsg = "please_kindly_check_email"
// 	}

// 	if membersID < 0 {
// 		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
// 		return
// 	}

// 	// start generate random password
// 	tmpPassword := member_service.GenerateRandomLoginPassword()

// 	// update generated random password

// 	// encrypt password
// 	password, err := base.Bcrypt(tmpPassword)
// 	if err != nil {
// 		base.LogErrorLog("RequestForgetPassword-Bcrypt", err.Error(), tmpPassword, true)
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 		return
// 	}

// 	// being transaction
// 	tx := models.Begin()

// 	// update password
// 	arrUpdCond := make([]models.WhereCondFn, 0)
// 	arrUpdCond = append(arrUpdCond,
// 		models.WhereCondFn{Condition: "id = ?", CondValue: membersID},
// 	)
// 	updateColumn := map[string]interface{}{"tmp_password": password, "updated_by": membersID}
// 	err = models.UpdatesFnTx(tx, "members", arrUpdCond, updateColumn, false)
// 	if err != nil {
// 		tx.Rollback()
// 		arrErr := map[string]interface{}{"upd_condition":arrUpdCond,"updateColumn":updateColumn,}
// 		base.LogErrorLog("RequestForgetPassword-UpdatesFnTx", err.Error(), arrErr, true)
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 		return
// 	}
// 	// send email
// 	arrData := map[string]interface{}{"tmp_password": tmpPassword}

// 	// otp := models.OTP{
// 	// 	ReceiverID: ,
// 	// }
// 	if sendType == "MOBILE" { // do send via mobile
// 		errMsg = sms_service.SendSmsByModules(otp, langCode, arrData)
// 	} else if sendType == "EMAIL" { // do send via email
// 		errMsg = email_service.SendEmailByModules(otp, langCode, arrData)
// 	}

// 	// send otp via sms/email
// 	otpService := otp_service.OTP{
// 		SendType:   sendType,
// 		ReceiverID: receiverID,
// 		OtpType:    form.ReqType,
// 		LangCode:   langCode,
// 	}

// 	countDownSec, errMsg := otpService.SendOTP(tx, nil)
// 	if errMsg != "" {
// 		models.Rollback(tx)
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
// 		return
// 	}

// 	// commit transaction
// 	err = models.Commit(tx)
// 	if err != nil {
// 		models.ErrorLog("otpController:RequestOTP()", "Commit():1", err.Error())
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 		return
// 	}

// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: successMsg}, map[string]interface{}{"count_down_seconds": countDownSec})
// 	return
// }
