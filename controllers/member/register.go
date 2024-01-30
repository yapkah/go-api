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
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/mobile_service"
	"github.com/yapkah/go-api/service/otp_service"
)

// RegisterForm struct
type RegisterForm struct {
	Username         string `form:"username" json:"username" valid:"Required;MinSize(4);MaxSize(19)"`
	FirstName        string `form:"first_name" json:"first_name"`
	Email            string `form:"email" json:"email" valid:"Required"`
	MobilePrefix     string `form:"mobile_prefix" json:"mobile_prefix"`
	MobileNo         string `form:"mobile_no" json:"mobile_no"`
	CountryCode      string `form:"country_code" json:"country_code" valid:"Required"`
	VerificationCode string `form:"verification_code" json:"verification_code" valid:"Required;MinSize(6);MaxSize(6)"`
	Password         string `form:"password" json:"password" valid:"Required"`
	SecondaryPin     string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
	ReferralCode     string `form:"referral_code" json:"referral_code" valid:"Required"`
}

// Register function
func Register(c *gin.Context) {
	var (
		appG                           = app.Gin{C: c}
		form                           RegisterForm
		err                            error
		mobilePrefix, mobileNo, errMsg string
		// mobilePrefix, mobileNo, sendType, receiverID, errMsg string
		referralID int
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// check which flow to perform (phone/email)
	// if (form.MobileNo == "" && form.Email == "") || (form.MobileNo != "" && form.Email != "") {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_either_mobile_no_or_email"}, nil)
	// 	return
	// }

	// validate if username is unique
	// check username format
	username := strings.Trim(form.Username, " ")
	errMsg = base.UsernameChecking(username)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// validate if username is unique
	ok, err = member_service.ExistsMemberByUsername(username)
	if err != nil {
		base.LogErrorLog("registerController:Register()", "ExistsMemberByUsername():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	if ok {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_USERNAME_ALREADY_EXISTS)}, nil)
		return
	}

	// establish where condition to find members info later for login process
	arrEntMemberMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		// models.WhereCondFn{Condition: "members.username = ?", CondValue: username},
	)

	// validate mobile_prefix+mobile_no/email
	if form.MobileNo != "" {
		// validate mobile prefix
		if form.MobilePrefix == "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_mobile_prefix"}, nil)
			return
		}

		// validate if mobile prefix exist
		arrSysTerritoryFn := make([]models.WhereCondFn, 0)
		arrSysTerritoryFn = append(arrSysTerritoryFn,
			models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: form.MobilePrefix},
		)
		arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

		if err != nil {
			base.LogErrorLog("registerController:Register()", "GetSysTerritoryFn():1", err.Error(), true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if arrSysTerritory == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_mobile_prefix"}, nil)
			return
		}

		countryCode := arrSysTerritory.Code

		mobileNo = strings.Trim(form.MobileNo, " ")

		num, errMsg := mobile_service.ParseMobileNo(mobileNo, countryCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		mobilePrefix = fmt.Sprintf("%v", *num.CountryCode)

		// validate if mobile no is unique
		ok, err := models.ExistsMemberByMobile(mobilePrefix, mobileNo)
		if err != nil {
			base.LogErrorLog("registerController:Register()", "ExistsMemberByMobile():1", err.Error(), true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}
		if ok {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_MOBILE_EXISTS)}, nil)
			return
		}

		// sendType = "MOBILE"
		// receiverID = mobilePrefix + mobileNo

		// arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		// 	models.WhereCondFn{Condition: "members.mobile_prefix = ?", CondValue: mobilePrefix},
		// 	models.WhereCondFn{Condition: "members.mobile_no = ?", CondValue: mobileNo},
		// )
	} else if form.Email != "" {
		// validate email format
		valid := validation.Validation{}
		valid.Email(form.Email, "email")

		if valid.HasErrors() {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: valid.Errors[0].String()}, nil)
			return
		}

		// validate if email is unique
		ok, err := models.ExistsMemberByEmail(form.Email)
		if err != nil {
			base.LogErrorLog("otpController:Register()", "ExistsMemberByEmail():1", err.Error(), true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}
		if ok {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_EMAIL_EXISTS)}, nil)
			return
		}

		// sendType = "EMAIL"
		// receiverID = form.Email

		// arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		// 	models.WhereCondFn{Condition: "members.email = ?", CondValue: form.Email},
		// 	models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		// )
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}
	sourceInterface, ok := c.Get("source")
	source := sourceInterface.(int)

	sourceNameInterface, _ := c.Get("sourceName")
	sourceName := sourceNameInterface.(string)
	if ok == false {
		base.LogErrorLog("Login-invalid_source", "", "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	decryptedText, err := util.RsaDecryptPKCS1v15(form.Password)
	if err != nil {
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.Password, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_password_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "password_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.Password = decryptedText

	decryptedText, err = util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount = utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// begin transaction
	tx := models.Begin()

	// validate verification code
	// inputOtp := strings.Trim(form.VerificationCode, " ")

	// errMsg = otp_service.ValidateOTP(tx, sendType, receiverID, inputOtp, "REGISTER")
	// if errMsg != "" {
	// 	models.Rollback(tx)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
	// 	return
	// }

	// validate referral code
	referralID = 1
	if form.ReferralCode != "" {
		referralCode := strings.Trim(form.ReferralCode, " ")
		errMsg, sponsorData := member_service.ValidateReferralCode(referralCode)

		if errMsg != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		referralID = sponsorData.ID
	}

	addMember := member_service.Member{
		// Username:     username,
		Email:        form.Email,
		MobilePrefix: mobilePrefix,
		MobileNo:     mobileNo,
		Password:     form.Password,
		SecondaryPin: form.SecondaryPin,
		ReferralID:   referralID,
		LangCode:     langCode,
	}

	// add member
	errMsg, mainID := addMember.Add(tx)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction - must commit first else can't get reserved referral code for create profile steps
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("registerController:Register()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		models.WhereCondFn{Condition: "members.id = ?", CondValue: mainID},
	)

	// add ent_member
	countryID := 1
	addEntMember := member_service.EntMember{
		MainID:       mainID,
		Username:     username,
		CountryID:    countryID,
		ReferralCode: form.ReferralCode,
		LangCode:     langCode,
		Source:       sourceName,
	}

	// begin transaction
	tx = models.Begin()

	// create profile process (create ent_member + ent_member_sponsor_tree)
	errMsg, entMemberID := addEntMember.CreateProfile(tx)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("registerController:Register()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// login
	// get called platform
	route := c.Request.URL.String()
	platformCheckingRst := strings.Contains(route, "/api/app")
	platform := "HTMLFIVE"
	if platformCheckingRst {
		platform = "APP"
	}

	// start get latest ent_member info. this method might not accurate
	// arrLatestEntMemberFn := make([]models.WhereCondFn, 0)
	// arrLatestEntMemberFn = append(arrLatestEntMemberFn,
	// 	models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: mainID},
	// 	// models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	// )
	// arrLatestEntMember, _ := models.GetLatestEntMemberFn(arrLatestEntMemberFn, false)
	// end get latest ent_member info. this method might not accurate

	// start update current member login
	tx = models.Begin()
	err = member_service.UpdateCurrentProfileWithLoginMember(tx, mainID, entMemberID, source)
	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}
	models.Commit(tx)
	// end update current member login

	// find member login details
	arrEntMemberMember, _ := models.GetEntMemberMemberFn(arrEntMemberMemberFn, false)

	db := models.GetDB() // no need transaction because if failed no need rollback
	processRst, err := member_service.ProcessMemberLogin(db, arrEntMemberMember, langCode, platform, 0)
	if err != nil {
		models.ErrorLog("registerController:Register()", "ProcessMemberLogin():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	tk := map[string]interface{}{
		"access_token": processRst.AccessToken,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "user_registered_successfully"}, tk)
	return
}

// CreateAccountForm struct
type CreateAccountForm struct {
	Username     string `form:"username" json:"username" valid:"Required;MinSize(4);MaxSize(19)"`
	ReferralCode string `form:"referral_code" json:"referral_code"`
}

// CreateAccount func
func CreateAccount(c *gin.Context) {
	var (
		appG              = app.Gin{C: c}
		form              CreateAccountForm
		mainID, countryID int
		errMsg            string
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	mainID = member.ID

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get country code from mobile_prefix
	countryID = 1
	if member.MobilePrefix != "" && member.MobileNo != "" {
		arrSysTerritoryFn := make([]models.WhereCondFn, 0)
		arrSysTerritoryFn = append(arrSysTerritoryFn,
			models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: "+" + member.MobilePrefix},
		)
		arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

		if err != nil {
			base.LogErrorLog("registerController:CreateProfile()", "GetSysTerritoryFn():1", err.Error(), true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if arrSysTerritory != nil {
			countryID = arrSysTerritory.ID
		}
	}

	// begin transaction
	tx := models.Begin()

	memService := member_service.EntMember{
		MainID:       mainID,
		Username:     form.Username,
		CountryID:    countryID,
		ReferralCode: form.ReferralCode,
		LangCode:     langCode,
	}

	// create profile process (create ent_member + ent_member_sponsor_tree)
	errMsg, _ = memService.CreateProfile(tx)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	err := models.Commit(tx)
	if err != nil {
		base.LogErrorLog("registerController:CreateProfile()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "registered_successfully"}, nil)
	return
}

// Registerv2 function
func Registerv2(c *gin.Context) {
	// type arrBallotGeneralSettingStructv2 struct {
	// 	StartTime string `json:"start_time"`
	// 	EndTime   string `json:"end_time"`
	// }

	var (
		appG                                                 = app.Gin{C: c}
		form                                                 RegisterForm
		err                                                  error
		mobilePrefix, mobileNo, sendType, receiverID, errMsg string
		referralID                                           int
		// arrGeneralSettingv2 arrBallotGeneralSettingStructv2
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// //check ballot session
	// arrGeneralSetup, err := models.GetSysGeneralSetupByID("ballot_setting")
	// if err != nil {
	// 	base.LogErrorLog("Registerv2-GetSysGeneralSetupByID1_failed", err.Error(), "ballot_setting", true)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
	// 	return
	// }
	// if arrGeneralSetup == nil {
	// 	base.LogErrorLog("Registerv2-GetSysGeneralSetupByID2_failed", "ballot_setting_is_not_set", arrGeneralSetup, true)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
	// 	return
	// }

	// json.Unmarshal([]byte(arrGeneralSetup.InputValue2), &arrGeneralSettingv2)

	// if arrGeneralSettingv2.StartTime != "" && arrGeneralSettingv2.EndTime != "" {
	// 	currTime := time.Now().Format("2006-01-02 15:04:05")

	// 	if currTime > arrGeneralSettingv2.EndTime {
	// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "register_session_ended"}, nil)
	// 		return
	// 	}

	// }

	// check which flow to perform (phone/email)
	// if (form.MobileNo == "" && form.Email == "") || (form.MobileNo != "" && form.Email != "") {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_either_mobile_no_or_email"}, nil)
	// 	return
	// }

	// validate if username is unique
	// check username format
	username := strings.Trim(form.Username, " ")
	errMsg = base.UsernameChecking(username)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// validate if username is unique
	ok, err = member_service.ExistsMemberByUsername(username)
	if err != nil {
		base.LogErrorLog("Registerv2-ExistsMemberByUsername", err.Error(), username, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	if ok {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_USERNAME_ALREADY_EXISTS)}, nil)
		return
	}

	// validate country code
	sys, err := models.GetCountryByCode(form.CountryCode)

	if err != nil {
		base.LogErrorLog("Registerv2:GetCountryByCode", err.Error(), map[string]interface{}{"countryCode": form.CountryCode}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	if sys == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_country_code"}, nil)
		return
	}

	var countryID = sys.ID

	// establish where condition to find members info later for login process
	arrEntMemberMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		// models.WhereCondFn{Condition: "members.username = ?", CondValue: username},
	)

	// validate mobile_prefix+mobile_no/email
	if form.Email != "" {
		// validate email format
		valid := validation.Validation{}
		valid.Email(form.Email, "email")

		if valid.HasErrors() {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: valid.Errors[0].String()}, nil)
			return
		}

		// validate if email is unique
		ok, err := models.ExistsMemberByEmail(form.Email)
		if err != nil {
			base.LogErrorLog("Registerv2-ExistsMemberByEmail", err.Error(), form.Email, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}
		if ok {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_EMAIL_EXISTS)}, nil)
			return
		}

		sendType = "EMAIL"
		receiverID = form.Email

		// arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		// 	models.WhereCondFn{Condition: "members.email = ?", CondValue: form.Email},
		// 	models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		// )
	} else if form.MobileNo != "" {
		// validate mobile prefix
		if form.MobilePrefix == "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_provide_mobile_prefix"}, nil)
			return
		}

		// validate if mobile prefix exist
		arrSysTerritoryFn := make([]models.WhereCondFn, 0)
		arrSysTerritoryFn = append(arrSysTerritoryFn,
			models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: form.MobilePrefix},
		)
		arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

		if err != nil {
			base.LogErrorLog("Registerv2-GetSysTerritoryFn", err.Error(), arrSysTerritoryFn, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if arrSysTerritory == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_mobile_prefix"}, nil)
			return
		}

		countryCode := arrSysTerritory.Code

		mobileNo = strings.Trim(form.MobileNo, " ")

		num, errMsg := mobile_service.ParseMobileNo(mobileNo, countryCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		mobilePrefix = fmt.Sprintf("%v", *num.CountryCode)

		// validate if mobile no is unique
		ok, err := models.ExistsMemberByMobile(mobilePrefix, mobileNo)
		if err != nil {
			base.LogErrorLog("Registerv2-ExistsMemberByMobile_failed", err.Error(), map[string]interface{}{"mobilePrefix": mobilePrefix, "mobileNo": mobileNo}, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}
		if ok {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: e.GetMsg(e.MEMBER_MOBILE_EXISTS)}, nil)
			return
		}

		// sendType = "MOBILE"
		// receiverID = mobilePrefix + mobileNo

		// arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		// 	models.WhereCondFn{Condition: "members.mobile_prefix = ?", CondValue: mobilePrefix},
		// 	models.WhereCondFn{Condition: "members.mobile_no = ?", CondValue: mobileNo},
		// )
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	sourceNameInterface, _ := c.Get("sourceName")
	sourceName := sourceNameInterface.(string)

	// sourceInterface, ok := c.Get("source")
	// source := sourceInterface.(int)

	// if !ok {
	// 	base.LogErrorLog("Registerv2-invalid_source", "", sourceInterface, true)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
	// 	return
	// }

	decryptedText, err := util.RsaDecryptPKCS1v15(form.Password)
	if err != nil {
		base.LogErrorLog("Registerv2-RsaDecryptPKCS1v15_Failed", err.Error(), form.Password, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_password_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "password_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.Password = decryptedText

	decryptedText, err = util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("Registerv2-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount = utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// begin transaction
	tx := models.Begin()

	// validate verification code
	if sendType == "EMAIL" || sendType == "MOBILE" {
		inputOtp := strings.Trim(form.VerificationCode, " ")

		errMsg = otp_service.ValidateOTP(tx, sendType, receiverID, inputOtp, "REGISTER")
		if errMsg != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}
	}

	// validate referral code
	referralID = 1 // system default sponsor id
	if form.ReferralCode != "" {
		referralCode := strings.Trim(form.ReferralCode, " ")
		errMsg, entMember := member_service.ValidateReferralCode(referralCode)

		if errMsg != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		referralID = entMember.ID
	}

	addMember := member_service.Member{
		// Username:     username,
		Email:        form.Email,
		MobilePrefix: mobilePrefix,
		MobileNo:     mobileNo,
		Password:     form.Password,
		SecondaryPin: form.SecondaryPin,
		ReferralID:   referralID,
		LangCode:     langCode,
	}

	// add member
	errMsg, mainID := addMember.Add(tx)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction - must commit first else can't get reserved referral code for create profile steps
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("Registerv2:Register()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	arrEntMemberMemberFn = append(arrEntMemberMemberFn,
		models.WhereCondFn{Condition: "members.id = ?", CondValue: mainID},
	)

	// add ent_member
	// countryID := 1
	addEntMember := member_service.EntMember{
		MainID:       mainID,
		Username:     username,
		FirstName:    form.FirstName,
		CountryID:    countryID,
		ReferralCode: form.ReferralCode,
		LangCode:     langCode,
		Source:       sourceName,
	}

	// begin transaction
	tx = models.Begin()

	// create profile process (create ent_member + ent_member_sponsor_tree)
	errMsg, _ = addEntMember.CreateProfilev2(tx)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("Registerv2-commit_failed", err.Error(), addEntMember, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// login
	// get called platform
	// route := c.Request.URL.String()
	// platformCheckingRst := strings.Contains(route, "/api/app")
	// platform := "HTMLFIVE"
	// if platformCheckingRst {
	// 	platform = "APP"
	// }

	// start get latest ent_member info. this method might not accurate
	// arrLatestEntMemberFn := make([]models.WhereCondFn, 0)
	// arrLatestEntMemberFn = append(arrLatestEntMemberFn,
	// 	models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: mainID},
	// 	// models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	// )
	// arrLatestEntMember, _ := models.GetLatestEntMemberFn(arrLatestEntMemberFn, false)
	// end get latest ent_member info. this method might not accurate

	// start update current member login
	// tx = models.Begin()
	// err = member_service.UpdateCurrentProfileWithLoginMember(tx, mainID, entMemberID, source)
	// if err != nil {
	// 	models.Rollback(tx)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
	// 	return
	// }
	// models.Commit(tx)
	// end update current member login

	// find member login details
	// arrEntMemberMember, _ := models.GetEntMemberMemberFn(arrEntMemberMemberFn, false)

	// db := models.GetDB() // no need transaction because if failed no need rollback
	// processRst, err := member_service.ProcessMemberLogin(db, arrEntMemberMember, langCode, platform, uint8(source))
	// if err != nil {
	// 	base.LogErrorLog("Registerv2-ProcessMemberLogin_failed", err.Error(), map[string]interface{}{"arrEntMemberMember": arrEntMemberMember, "langCode": langCode, "platform": platform, "source": uint8(source)}, true)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
	// 	return
	// }

	// tk := map[string]interface{}{
	// 	"access_token": processRst.AccessToken,
	// }

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "user_registered_successfully"}, nil)
	return
}

// CreateAccountv2 func
func CreateAccountv2(c *gin.Context) {
	var (
		appG              = app.Gin{C: c}
		form              CreateAccountForm
		mainID, countryID int
		errMsg            string
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	mainID = member.ID

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get country code from mobile_prefix
	countryID = 1
	if member.MobilePrefix != "" && member.MobileNo != "" {
		arrSysTerritoryFn := make([]models.WhereCondFn, 0)
		arrSysTerritoryFn = append(arrSysTerritoryFn,
			models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: "+" + member.MobilePrefix},
		)
		arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

		if err != nil {
			base.LogErrorLog("CreateAccountv2-GetSysTerritoryFn_failed", err.Error(), arrSysTerritoryFn, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if arrSysTerritory != nil {
			countryID = arrSysTerritory.ID
		}
	}

	// begin transaction
	tx := models.Begin()

	memService := member_service.EntMember{
		MainID:       mainID,
		Username:     form.Username,
		CountryID:    countryID,
		ReferralCode: form.ReferralCode,
		LangCode:     langCode,
	}

	// create profile process (create ent_member + ent_member_sponsor_tree)
	errMsg, _ = memService.CreateProfilev2(tx)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	err := models.Commit(tx)
	if err != nil {
		base.LogErrorLog("CreateAccountv2-CreateProfilev2", "Commit_failed", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "registered_successfully"}, nil)
	return
}
