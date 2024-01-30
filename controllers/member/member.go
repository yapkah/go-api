package member

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"

	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/service/member_service"
	"github.com/smartblock/gta-api/service/mobile_service"
	"github.com/smartblock/gta-api/service/otp_service"
)

// GetProfile function
func GetProfile(c *gin.Context) {
	var (
		appG   = app.Gin{C: c}
		errMsg string
	)

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get profile by username
	profile, errMsg := member_service.GetProfile(member.NickName, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, profile)
}

// GetMemberSettingStatus function
func GetMemberSettingStatus(c *gin.Context) {
	var (
		appG   = app.Gin{C: c}
		errMsg string
	)

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberSettingStatus, errMsg := member_service.GetMemberSettingStatus(member.EntMemberID, langCode)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberSettingStatus)
	return
}

// GetMemberTreev1 function
func GetMemberTreev1(c *gin.Context) {
	var (
		appG             = app.Gin{C: c}
		form             member_service.MemberTreeFormStruct
		level            int
		downlineMemberID int
		incMem           int
		incDownMem       int
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

	if form.DownlineUsername != "" {
		if strings.ToLower(form.DownlineUsername) == strings.ToLower(member.NickName) {
			form.DownlineUsername = ""
		}
	}

	if form.DownlineUsername != "" {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member.status IN ('A') AND ent_member.nick_name = ? ", CondValue: form.DownlineUsername},
		)
		arrDownlineUsername, err := models.GetEntMemberFn(arrCond, "", false)
		if err != nil || arrDownlineUsername == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_member"}, nil)
			return
		}

		if arrDownlineUsername.ID == 0 {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_member"}, nil)
			return
		}

		// arrCond = make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: " ent_member.status IN ('A') AND ent_member.nick_name = ? ", CondValue: member.NickName},
		// )
		// entMemberLotSponsorRst, _ := models.GetEntMemberLotSponsorFn(arrCond, false)

		// sponsorLvl := 0
		// if len(entMemberLotSponsorRst) > 0 {
		// 	sponsorLvl = entMemberLotSponsorRst[0].Lvl
		// }

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " sponsor.nick_name = ? ", CondValue: member.NickName},
			models.WhereCondFn{Condition: " downline.nick_name != ? ", CondValue: member.NickName},
			// models.WhereCondFn{Condition: " downline_lot.i_lvl <= ? ", CondValue: sponsorLvl + form.Level},
			models.WhereCondFn{Condition: " downline.status IN  ('A') AND downline.nick_name = ? ", CondValue: form.DownlineUsername},
		)
		checkSponsorMemberRst, _ := models.GetEntMemberLotSponsorDetailFn(arrCond, false)

		if len(checkSponsorMemberRst) < 1 {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_member_search"}, nil)
			return
		}
		// checkSponsorMemberRst := member_service.CheckSponsorMember(member.EntMemberID, arrDownlineUsername.EntMemberID)
		// if !checkSponsorMemberRst {
		// 	appG.ResponseV2(0, http.StatusOK, "invalid_member_search", nil)
		// 	return
		// }
		downlineMemberID = arrDownlineUsername.ID
	}
	if form.Level > 0 {
		level = form.Level
	}
	if form.IncMem > 0 {
		incMem = form.IncMem
	}
	if form.IncDownMem > 0 {
		incDownMem = form.IncDownMem
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get profile by username
	arrData := member_service.ArrMemberTreeDataStruct{
		Level:            level,
		DownlineMemberID: downlineMemberID,
		IncMem:           incMem,
		IncDownMem:       incDownMem,
		EntMemberID:      member.EntMemberID,
		DataType:         form.DataType,
	}

	result, err := member_service.GetMemberTreev2(arrData, langCode)
	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, result)
}

// UpdateProfileForm struct
type UpdateProfileForm struct {
	FirstName   string `form:"first_name" json:"first_name"`
	CountryCode string `form:"country_code" json:"country_code"`
	GenderCode  string `form:"gender_code" json:"gender_code"`
	BirthDate   string `form:"birth_date" json:"birth_date"`
	// Username    string `form:"username" json:"username"`
	// CryptoType    string `form:"crypto_type" json:"crypto_type"`
	// CryptoAddress string `form:"crypto_address" json:"crypto_address"`
}

// UpdateProfile function
func UpdateProfile(c *gin.Context) {
	var (
		appG    = app.Gin{C: c}
		form    UpdateProfileForm
		err     error
		errMsg  string
		changes bool = false
	)

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	memberID := member.EntMemberID

	// begin transaction
	tx := models.Begin()

	// get current data for log
	errMsg, currentData := base.GetMemberLogData([]string{"ent_member", "ent_member_crypto"}, memberID)

	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// update username
	// if form.Username != "" {
	// 	// call update username func
	// 	memberUsername := member_service.MemberUsername{
	// 		MemberID: memberID,
	// 		Username: form.Username,
	// 	}

	// 	errMsg = memberUsername.UpdateMemberUsername(tx)
	// 	if errMsg != "" {
	// 		models.Rollback(tx)
	// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
	// 		return
	// 	}

	// 	if changes == false {
	// 		changes = true
	// 	}
	// }

	// update first_name
	if form.FirstName != "" {
		memberFirstName := member_service.MemberFirstName{
			MemberID:  memberID,
			FirstName: form.FirstName,
		}

		errMsg = memberFirstName.UpdateMemberFirstName(tx)
		if errMsg != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		if changes == false {
			changes = true
		}
	}

	// update country
	if form.CountryCode != "" {
		ok = models.ExistCountryCode(form.CountryCode)
		if !ok {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_country_code"}, nil)
			return
		}

		// add crypto address
		memberCountry := member_service.MemberCountry{
			MemberID:    memberID,
			CountryCode: form.CountryCode,
		}

		errMsg = memberCountry.UpdateMemberCountry(tx)
		if errMsg != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		if changes == false {
			changes = true
		}
	}

	// update gender
	if form.GenderCode != "" {
		memberGender := member_service.MemberGender{
			MemberID:   memberID,
			GenderCode: form.GenderCode,
		}

		errMsg = memberGender.UpdateMemberGender(tx)
		if errMsg != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		if changes == false {
			changes = true
		}
	}

	if form.BirthDate != "" {
		memberBirthDate := member_service.MemberBirthDate{
			MemberID:  memberID,
			BirthDate: form.BirthDate,
		}

		errMsg = memberBirthDate.UpdateMemberBirthDate(tx)
		if errMsg != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		if changes == false {
			changes = true
		}
	}

	// update crypto address
	// if form.CryptoAddress != "" {
	// 	if form.CryptoType == "" {
	// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "crypto_type_cannot_be_empty"}, nil)
	// 		return
	// 	}

	// 	// add crypto address
	// 	cryptoService := member_service.Crypto{
	// 		MemberID:      memberID,
	// 		CryptoType:    form.CryptoType,
	// 		CryptoAddress: form.CryptoAddress,
	// 	}

	// 	errMsg = cryptoService.Add(tx)
	// 	if errMsg != "" {
	// 		models.Rollback(tx)
	// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
	// 		return
	// 	}

	// 	if changes == false {
	// 		changes = true
	// 	}
	// }

	// verify if anything is updated
	if changes == false {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "nothing_is_updated"}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("memberController:UpdateProfile()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// get updated data for log
	errMsg, updatedData := base.GetMemberLogData([]string{"ent_member", "ent_member_crypto"}, memberID)

	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	base.AddSysLog(memberID, currentData, updatedData, "modify", "update-overview", c)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "updated_successfully"}, nil)
	return
}

// UpdateMobileForm struct
type UpdateMobileForm struct {
	MobilePrefix string `form:"mobile_prefix" json:"mobile_prefix" valid:"Required;MaxSize(5)"`
	MobileNo     string `form:"mobile_no" json:"mobile_no" valid:"Required;MinSize(8);MaxSize(15)"`
}

// UpdateMobile function
func UpdateMobile(c *gin.Context) {
	var (
		appG   = app.Gin{C: c}
		form   UpdateMobileForm
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

	// begin transaction
	tx := models.Begin()

	// get current data for log
	errMsg, currentData := base.GetMemberLogData([]string{"members"}, memberID)

	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// update secondary pin
	memberMobile := member_service.MemberMobile{
		MemberID:        memberID,
		EntMemberStatus: member.Status,
		MobilePrefix:    form.MobilePrefix,
		MobileNo:        form.MobileNo,
	}

	// rollback if hit error
	errMsg = memberMobile.UpdateMemberMobile(tx)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("memberController:UpdateMobile()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// get updated data for log
	errMsg, updatedData := base.GetMemberLogData([]string{"members"}, memberID)

	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	base.AddSysLog(memberID, currentData, updatedData, "modify", "update-overview", c)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "please_check_mobile_for_activation_code"}, nil)
	return
}

// ValidateReferralCodeStruct struct
type ValidateReferralCodeStruct struct {
	ReferralCode string `form:"referral_code" json:"referral_code" valid:"Required;MaxSize(50)"`
}

// ValidateReferralCode func
func ValidateReferralCode(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form ValidateReferralCodeStruct
	)

	// validate inputs
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	referralCode := strings.Trim(form.ReferralCode, " ")
	errMsg, arrEntMember := member_service.ValidateReferralCode(referralCode)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: errMsg}, map[string]string{"referral_username": arrEntMember.NickName})
	return
}

// GetRandUsername func
func GetRandUsername(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	username, errMsg := member_service.GetRandUsername()

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
	} else {
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]string{"username": username})
	}

	return
}

// BindMobileForm struct
type BindMobileForm struct {
	MobilePrefix string `form:"mobile_prefix" json:"mobile_prefix" valid:"Required;"`
	MobileNo     string `form:"mobile_no" json:"mobile_no" valid:"Required;"`
	// VerificationCode string `form:"verification_code" json:"verification_code" valid:"Required;MinSize(6);MaxSize(6)"`
	// SecondaryPin     string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// BindMobile func
func BindMobile(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form BindMobileForm
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

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)

	mainID := member.MainID
	entMemberID := member.EntMemberID

	// validate if mobile prefix exist
	arrSysTerritoryFn := make([]models.WhereCondFn, 0)
	arrSysTerritoryFn = append(arrSysTerritoryFn,
		models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: form.MobilePrefix},
	)
	arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

	if err != nil {
		base.LogErrorLog("memberController:BindMobile()", "GetSysTerritoryFn():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	if arrSysTerritory == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_mobile_prefix"}, nil)
		return
	}

	countryCode := arrSysTerritory.Code

	mobileNo := strings.Trim(form.MobileNo, " ")

	num, errMsg := mobile_service.ParseMobileNo(mobileNo, countryCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	mobilePrefix := fmt.Sprintf("%v", *num.CountryCode)

	// decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	// if err != nil {
	// 	base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
	// 	return
	// }

	// wordCount := utf8.RuneCountInString(decryptedText)
	// if wordCount < 6 {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
	// 	return
	// }
	// if wordCount > 6 {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
	// 	return
	// }
	// form.SecondaryPin = decryptedText

	// check secondary password
	// pinValidation := base.SecondaryPin{
	// 	MemId:              entMemberID,
	// 	SecondaryPin:       form.SecondaryPin,
	// 	MemberSecondaryPin: member.SecondaryPin,
	// 	LangCode:           langCode,
	// }

	// err = pinValidation.CheckSecondaryPin()
	// if err != nil {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
	// 	return
	// }

	// being transaction
	tx := models.Begin()

	// get current data for log
	errMsg, currentData := base.GetMemberLogData([]string{"members"}, mainID)

	// validate verification code
	// inputOtp := strings.Trim(form.VerificationCode, " ")

	// errMsg = otp_service.ValidateOTP(tx, "MOBILE", mobilePrefix+mobileNo, inputOtp, "MOBILE")
	// if errMsg != "" {
	// 	models.Rollback(tx)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
	// 	return
	// }

	// perform post exchange action
	bindMemberMobileEmailStruct := member_service.BindMemberMobileEmailStruct{
		BindType:     "MOBILE",
		MainID:       mainID,
		MobilePrefix: mobilePrefix,
		MobileNo:     mobileNo,
	}

	errMsg = member_service.BindMemberMobileEmail(tx, bindMemberMobileEmailStruct, langCode)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("memberController:BindMobile()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// get updated data for log
	errMsg, updatedData := base.GetMemberLogData([]string{"members"}, mainID)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	base.AddSysLog(entMemberID, currentData, updatedData, "modify", "update-user", c)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "mobile_no_binded_successfully"}, nil)
	return
}

// BindEmailForm struct
type BindEmailForm struct {
	Email            string `form:"email" json:"email"`
	VerificationCode string `form:"verification_code" json:"verification_code" valid:"Required;MinSize(6);MaxSize(6)"`
	SecondaryPin     string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// BindEmail func
func BindEmail(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form BindEmailForm
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

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)

	mainID := member.MainID
	entMemberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	// get current data for log
	errMsg, currentData := base.GetMemberLogData([]string{"members"}, mainID)

	// validate verification code
	inputOtp := strings.Trim(form.VerificationCode, " ")

	errMsg = otp_service.ValidateOTP(tx, "EMAIL", form.Email, inputOtp, "EMAIL")
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// perform post exchange action
	bindMemberMobileEmailStruct := member_service.BindMemberMobileEmailStruct{
		BindType: "EMAIL",
		MainID:   mainID,
		Email:    form.Email,
	}

	errMsg = member_service.BindMemberMobileEmail(tx, bindMemberMobileEmailStruct, langCode)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("memberController:BindEmail()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// get updated data for log
	errMsg, updatedData := base.GetMemberLogData([]string{"members"}, mainID)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	base.AddSysLog(entMemberID, currentData, updatedData, "modify", "update-user", c)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "email_binded_successfully"}, nil)
	return
}

type AddSupportTicketForm struct {
	CategoryCode string `form:"category_code" json:"category_code" valid:"Required;"`
	Title        string `form:"title" json:"title" valid:"Required;"`
	Msg          string `form:"msg" json:"msg" valid:"Required;"`
}

func PostSupportTicket(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AddSupportTicketForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	tx := models.Begin()

	supportTicket := member_service.SupportTicketStruct{
		MemberId:     entMemberID,
		CategoryCode: form.CategoryCode,
		Title:        form.Title,
		Msg:          form.Msg,
		LangCode:     langCode,
	}

	arrData, err := supportTicket.PostSupportTicket(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "ticket_submitted"}, arrData)
}

type SupportTicketListForm struct {
	Page int64 `form:"page" json:"page"`
}

func GetMemberSupportTicketList(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form SupportTicketListForm
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	if langCode == "zh-CN" {
		langCode = "zh"
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	suppTicketList := member_service.SupportTicketListStruct{
		MemberID: entMemberID,
		LangCode: langCode,
		Page:     form.Page,
	}
	arrSupportTicketList, err := suppTicketList.GetMemberSupportTicketListv1()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: "fail_to_get_support_ticket_list",
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrSupportTicketList)
}

type SupportTicketListHistoryForm struct {
	Page       int64  `form:"page" json:"page"`
	TicketCode string `form:"ticket_code" json:"ticket_code" valid:"Required"`
}

func GetMemberSupportTicketHistoryList(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form SupportTicketListHistoryForm
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	suppTicketList := member_service.SupportTicketHistoryListStruct{
		MemberID:   entMemberID,
		TicketCode: form.TicketCode,
		LangCode:   langCode,
		Page:       form.Page,
	}
	arrSupportTicketList, err := suppTicketList.GetMemberSupportTicketHistoryListv1()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: "fail_to_get_support_ticket_history_list",
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrSupportTicketList)
}

func GetMemberSupportTicketCategoryList(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	arrSupportTicketCategoryList, errMsg := member_service.GetMemberSupportTicketCategoryList(langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrSupportTicketCategoryList)
}

type AddSupportTickeReplyForm struct {
	TicketCode string `form:"ticket_code" json:"ticket_code" valid:"Required;"`
	Msg        string `form:"msg" json:"msg" valid:"Required;"`
}

func PostSupportTicketReply(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AddSupportTickeReplyForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	tx := models.Begin()

	supportTicket := member_service.SupportTicketReplyStruct{
		MemberId:   entMemberID,
		TicketCode: form.TicketCode,
		Msg:        form.Msg,
		LangCode:   langCode,
	}

	err := supportTicket.PostSupportTicketReply(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "ticket_replied"}, nil)
}

type AddSupportTickeCloseForm struct {
	TicketCode string `form:"ticket_code" json:"ticket_code" valid:"Required;"`
}

func PostSupportTicketClose(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form AddSupportTickeCloseForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	tx := models.Begin()

	supportTicket := member_service.SupportTicketCloseStruct{
		MemberId:   entMemberID,
		TicketCode: form.TicketCode,
		LangCode:   langCode,
	}

	err := supportTicket.PostSupportTicketClose(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "ticket_closed"}, nil)
}

type GetDashboardTypeForm struct {
	Type string `form:"type" json:"type"`
}

func GetDashboard(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetDashboardTypeForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	if form.Type == "" {
		form.Type = "MAIN"
	}

	dashboard := member_service.GetDashboardStruct{
		MemberID: entMemberID,
		LangCode: langCode,
		Type:     form.Type,
	}
	arrData, err := dashboard.GetMemberDashboard()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
}

func GetDashboardBanner(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		err  error
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	dashboard := member_service.GetDashboardStruct{
		MemberID: entMemberID,
		LangCode: langCode,
	}
	arrData, err := dashboard.GetMemberDashboardBanner()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
}

type PlacementSetupForm struct {
	PlacementCode string `form:"placement_code" json:"placement_code" valid:"Required;"`
}

func PlacementSetup(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form PlacementSetupForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}
	// retrieve langCode
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get placement group setting
	paymentGroup := member_service.GetPlacementLegOption(form.PlacementCode, langCode)

	var arrDataReturn = map[string]interface{}{
		"placement_group": paymentGroup,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

// BindPlacementForm struct
type BindPlacementForm struct {
	PlacementCode  string `form:"placement_code" json:"placement_code" valid:"Required"`
	PlacementGroup string `form:"placement_group" json:"placement_group" valid:"Required"`
	SecondaryPin   string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// BindPlacement func
func BindPlacement(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form BindPlacementForm
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

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)
	memID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("registerController:BindPlacement():RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// get tree info
	arrEntMemberSponsorFn := make([]models.WhereCondFn, 0)
	arrEntMemberSponsorFn = append(arrEntMemberSponsorFn,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.member_id = ?", CondValue: memID},
	)
	arrEntMemberSponsor, _ := models.GetMemberSponsorFn(arrEntMemberSponsorFn, false)

	// validate if already got placement
	if arrEntMemberSponsor.UplineID != 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "placement_already_set_before"}, nil)
		return
	}

	// validate placement code and leg with referral code
	var (
		referralID     int    = arrEntMemberSponsor.SponsorID
		placementID    int    = 0
		legNo          int    = 0
		placementCode  string = strings.Trim(form.PlacementCode, " ")
		placementGroup int    = 0
	)

	placementGroup, err = strconv.Atoi(form.PlacementGroup)
	if err != nil {
		base.LogErrorLog("registerController:BindPlacement():strconv.Atoi():1", err.Error(), map[string]interface{}{"value": form.PlacementGroup}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	errMsg, placementData := member_service.ValidatePlacementCode(memID, referralID, placementCode, placementGroup)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	if placementData == nil {
		base.LogErrorLog("registerController:BindPlacement():ValidatePlacementCode:1", "returned_data_is_empty", map[string]interface{}{"data": placementData}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	placementID = placementData.ID
	legNo = placementGroup

	// being transaction
	tx := models.Begin()

	// update sponsor tree
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: " member_id = ?", CondValue: memID},
	)

	updateColumn := map[string]interface{}{"upline_id": placementID, "leg_no": legNo}
	err = models.UpdatesFnTx(tx, "ent_member_tree_sponsor", arrUpdCond, updateColumn, false)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// save member lot queue
	arrAddEntMemberLotQueue := models.AddEntMemberLotQueueStruct{
		MemberID:   memID,
		MemberLot:  "01",
		SponsorID:  referralID,
		SponsorLot: "01",
		UplineID:   placementID,
		UplineLot:  "01",
		LegNo:      legNo,
		Type:       "REG",
		DtCreate:   base.GetCurrentTime("2006-01-02 15:04:05"),
	}
	_, err = models.AddEntMemberLotQueue(tx, arrAddEntMemberLotQueue)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("memberService:CreateProfile()", "AddEntMemberLotQueue():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("memberController:BindPlacement()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "placement_set_successfully"}, nil)
	return
}

// Update2FAForm struct
type Update2FAForm struct {
	Mode         string `form:"mode" json:"mode" valid:"Required"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// Update2FA func
func Update2FA(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form Update2FAForm
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

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)
	memID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("registerController:Update2FA():RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	update2FAParam := member_service.Update2FAParam{
		MemberID: memID,
		Mode:     form.Mode,
	}

	errMsg := member_service.Update2FA(tx, update2FAParam, langCode)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("memberController:Update2FA()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "set_successfully"}, nil)
	return
}

// Validate2FAForm struct
type Validate2FAForm struct {
	Passcode string `form:"passcode" json:"passcode" valid:"Required;MaxSize(6)"`
}

// Validate2FA function
func Validate2FA(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form Validate2FAForm
	)

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	memberID := member.EntMemberID

	// check if already got secret key
	arrEntMember2FaFn := make([]models.WhereCondFn, 0)
	arrEntMember2FaFn = append(arrEntMember2FaFn,
		models.WhereCondFn{Condition: "ent_member_2fa.member_id = ? ", CondValue: memberID},
		// models.WhereCondFn{Condition: "ent_member_2fa.b_enable = ? ", CondValue: 1},
	)
	arrEntMember2Fa, _ := models.GetEntMember2FA(arrEntMember2FaFn, false)

	if len(arrEntMember2Fa) > 0 {
		secret := util.EncodeBase32(arrEntMember2Fa[0].Secret)
		rst, err := totp.ValidateCustom(form.Passcode, secret, time.Now().UTC(), totp.ValidateOpts{
			Period:    30,
			Skew:      0,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil {
			base.LogErrorLog("Validate2FA():ValidateCustom()", err.Error(), map[string]interface{}{"passcode": form.Passcode, "secere": secret, "time": time.Now().UTC(), "setting": totp.ValidateOpts{
				Period:    30,
				Skew:      0,
				Digits:    otp.DigitsSix,
				Algorithm: otp.AlgorithmSHA1,
			}}, true)

			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}
		if rst {
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "valid_otp"}, nil)
			return
		}
	}

	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_otp"}, nil)
	return
}

type GetStrategyRankingForm struct {
	Type   string `form:"type" json:"type"`
	Market string `form:"market" json:"market"`
}

func GetStrategyRanking(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetStrategyRankingForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	if form.Type == "" {
		form.Type = "MAIN"
	}

	leaderboard := member_service.GetStrategyRankingStruct{
		MemberID: entMemberID,
		LangCode: langCode,
		Type:     form.Type,
		Market:   form.Market,
	}
	arrData, err := leaderboard.GetStrategyRanking()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
}

func GetCryptoPrice(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		err  error
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	crypto := member_service.GetCryptoPriceStruct{
		MemberID: entMemberID,
		LangCode: langCode,
	}
	arrData, err := crypto.GetCryptoPrice()

	// fmt.Println(arrData);
	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
}

// GetMemberPdfForm struct
type GetMemberPdfForm struct {
	DocType string `form:"doc_type" json:"doc_type" valid:"Required;"` // PACKAGE_A, PACKAGE_B
}

//func GetMemberPdf function
func GetMemberPdf(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberPdfForm
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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
	memberID := member.EntMemberID

	arrMemberPdf, errMsg := member_service.GetMemberPdf(memberID, form.DocType, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrMemberPdf)
	return
}
