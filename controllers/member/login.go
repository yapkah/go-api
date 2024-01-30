package member

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"strings"

	// "github.com/smartblock/gta-api/helpers"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/member_service"
)

// LoginMemberForm struct
type LoginMemberForm struct {
	LoginType    string `form:"login_type" json:"login_type" valid:"Required"`
	Password     string `form:"password" json:"password" valid:"Required"`
	LangCode     string `form:"lang_code" json:"lang_code"`
	Email        string `form:"email" json:"email"`
	MobilePrefix string `form:"mobile_prefix" json:"mobile_prefix"`
	MobileNo     string `form:"mobile_no" json:"mobile_no"`
	Username     string `form:"username" json:"username"`
}

// Login function
func Login(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form LoginMemberForm
	)

	route := c.Request.URL.String()
	platformCheckingRst := strings.Contains(route, "/api/app")
	platform := "HTMLFIVE"
	if platformCheckingRst {
		platform = "APP"
	}

	// validate input
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	// check language
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if form.LangCode != "" {
		ok = models.ExistLangague(form.LangCode)
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	} else if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	arrCond := make([]models.WhereCondFn, 0)
	if strings.ToLower(form.LoginType) == "email" {
		if form.Email == "" {
			message := app.MsgStruct{
				Msg: "please_enter_email",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "members.email = ?", CondValue: form.Email},
			models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		)
	} else if strings.ToLower(form.LoginType) == "mobile" {
		if form.MobileNo == "" {
			message := app.MsgStruct{
				Msg: "please_enter_mobile_number",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		if form.MobilePrefix == "" {
			message := app.MsgStruct{
				Msg: "please_enter_mobile_prefix",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		// validate if mobile prefix exist
		arrSysTerritoryFn := make([]models.WhereCondFn, 0)
		arrSysTerritoryFn = append(arrSysTerritoryFn,
			models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: form.MobilePrefix},
		)
		arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

		if err != nil {
			base.LogErrorLog("Login_failed_to_get_GetSysTerritoryFn", err.Error(), arrSysTerritoryFn, true)
			message := app.MsgStruct{
				Msg: "something_went_wrong",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		if arrSysTerritory == nil {
			message := app.MsgStruct{
				Msg: "invalid_mobile_prefix",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		callingNoPrefix := arrSysTerritory.CallingNoPrefix

		mobilePrefix := strings.Replace(callingNoPrefix, "+", "", -1)

		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "members.mobile_prefix = ?", CondValue: mobilePrefix},
			models.WhereCondFn{Condition: "members.mobile_no = ?", CondValue: form.MobileNo},
			models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		)
	} else if strings.ToLower(form.LoginType) == "username" {
		if form.Username == "" {
			message := app.MsgStruct{
				Msg: "please_enter_username",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
			// models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
	} else {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}

	sourceInterface, ok := c.Get("source")
	if ok == false {
		base.LogErrorLog("Login-invalid_source", "", "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	source := sourceInterface.(int)

	var entMemberID int
	var member *models.EntMemberMembers
	arrCurrentActiveProfileMemberData := models.CurrentActiveProfileMemberStruct{
		SourceID: source,
	}
	if strings.ToLower(form.LoginType) == "email" || strings.ToLower(form.LoginType) == "mobile" {
		member, _ = models.GetCurrentActiveProfileMemberFn(arrCond, arrCurrentActiveProfileMemberData, true)
		if member == nil || member.ID < 1 {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
			return
		}
		entMemberID = member.EntMemberID
	} else {
		arrEntMem, _ := models.GetEntMemberFn(arrCond, "", false)
		if arrEntMem == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
			return
		}
		arrMembersFn := make([]models.WhereCondFn, 0)
		arrMembersFn = append(arrMembersFn,
			models.WhereCondFn{Condition: "members.id = ?", CondValue: arrEntMem.MainID},
			models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		)
		arrMembers, _ := models.GetMembersFn(arrMembersFn, false)
		if arrMembers == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
			return
		}
		arrEntMemMembersFn := make([]models.WhereCondFn, 0)
		arrEntMemMembersFn = append(arrEntMemMembersFn,
			models.WhereCondFn{Condition: "members.id = ?", CondValue: arrEntMem.MainID},
			models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		)
		member, _ = models.GetCurrentActiveProfileMemberFn(arrEntMemMembersFn, arrCurrentActiveProfileMemberData, false)
		if member == nil || member.ID < 1 {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
			return
		}
		entMemberID = arrEntMem.ID
	}

	// if member == nil || member.ID < 1 {
	// 	message := app.MsgStruct{
	// 		Msg: "invalid_login_info",
	// 	}
	// 	appG.ResponseV2(0, http.StatusOK, message, nil)
	// 	return
	// }
	// find member

	if member != nil && member.Status == "T" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}

	verificationRst := member_service.CheckMemberAccessPermission(entMemberID)
	if !verificationRst {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}

	arrLoginAttemptsLog := member_service.LoginAttemptsLogData{
		MemberID: entMemberID,
		ClientIP: c.ClientIP(),
	}

	// check password
	decryptedText, err := util.RsaDecryptPKCS1v15(form.Password)
	if err != nil {
		base.LogErrorLog("Login-RsaDecryptPKCS1v15_Failed", err.Error(), form.Password, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}
	form.Password = decryptedText
	err = base.CheckBcryptPassword(member.Password, form.Password)
	if err != nil {
		member_service.LoginAttemptsLog(arrLoginAttemptsLog, "member", "F")
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}

	isLockedAccountRst, _ := member_service.IsLockedAccount(arrLoginAttemptsLog, "member")
	if isLockedAccountRst.IsLockedAccount {
		arrReplace := make(map[string]string)
		arrReplace["hours"] = fmt.Sprintf("%g", isLockedAccountRst.Hours)
		arrReplace["minutes"] = fmt.Sprintf("%g", isLockedAccountRst.Minutes)
		arrReplace["seconds"] = fmt.Sprintf("%g", isLockedAccountRst.Seconds)
		arrMsg := app.MsgStruct{
			Msg:    "account_is_locked_for_:hours_hours_:minutes_minutes_:seconds_seconds",
			Params: arrReplace,
		}
		appG.ResponseV2(0, http.StatusOK, arrMsg, nil)
		return
	}

	// start update current member login
	tx := models.Begin()
	err = member_service.UpdateCurrentProfileWithLoginMember(tx, member.MainID, entMemberID, source)
	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}
	models.Commit(tx)
	// end update current member login

	// start get the correct current member info
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: entMemberID},
		models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
	)
	member, _ = models.GetCurrentActiveProfileMemberFn(arrCond, arrCurrentActiveProfileMemberData, false)
	// end get the correct current member info

	tx = models.Begin()
	processRst, err := member_service.ProcessMemberLogin(tx, member, langCode, platform, uint8(source))

	if err != nil {
		models.Rollback(tx)
		member_service.LoginAttemptsLog(arrLoginAttemptsLog, "member", "F")
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("Login-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}

	member_service.LoginAttemptsLog(arrLoginAttemptsLog, "member", "S")

	dataReturn := map[string]interface{}{
		// "username":        member.NickName,
		"access_token": processRst.AccessToken,
		// "refresh_token":   processRst.RefreshToken,
		// "expired_at_unix": processRst.ExpiredAtUnix,
		// "expired_at":      processRst.ExpiredAt,
	}
	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, dataReturn)
}

// AdminLoginMemberFrom struct
type AdminLoginMemberFrom struct {
	Username string `form:"username" json:"username" valid:"Required;MaxSize(100)"`
	LangCode string `form:"lang_code" json:"lang_code" valid:"Required;MinSize(2)"`
}

// AdminGenerateMemberAccess function
func AdminGenerateMemberAccess(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AdminLoginMemberFrom
	)
	platform := "HTMLFIVE"

	// validate input
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	sourceInterface, ok := c.Get("source")
	if ok == false {
		base.LogErrorLog("Login-invalid_source", "", "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	source := sourceInterface.(int)

	username := strings.Trim(form.Username, " ")

	// find member
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
	)

	arrAdminEntMemberMemberData := map[string]string{
		"nick_name": username,
	}
	member, _ := models.GetAdminEntMemberMemberFn(arrCond, arrAdminEntMemberMemberData, true)
	if member == nil || member.ID < 1 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}

	tx := models.Begin()

	processRst, err := member_service.ProcessMemberLogin(tx, member, form.LangCode, platform, uint8(source))
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("LoginByAdmin-ProcessMemberLogin", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("LoginByAdmin-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}

	dataReturn := map[string]interface{}{
		"access_token": processRst.AccessToken,
		// "refresh_token": processRst.RefreshToken,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, dataReturn)
}

// AdminLoginGatewayFrom struct
type AdminLoginGatewayFrom struct {
	Token string `form:"token" json:"token" valid:"Required"`
}

// AdminLoginGateway function
func AdminLoginGateway(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AdminLoginGatewayFrom
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	// find member
	accessToken, _ := models.GetAccessTokenByID(form.Token)

	var accessTokenScope []string
	json.Unmarshal([]byte(accessToken.Scope), &accessTokenScope)

	refreshToken, _ := models.GetRefreshTokenByAccessTokenID(form.Token)
	refreshTokenScope := append([]string{"REFRESH"}, accessTokenScope...)

	at, _ := util.GenerateToken(form.Token, accessToken.SubID, accessToken.ExpiresAt, accessTokenScope, "LOGIN", nil)        // generate access token
	rt, _ := util.GenerateToken(refreshToken.ID, accessToken.SubID, refreshToken.ExpiresAt, refreshTokenScope, "LOGIN", nil) // generate refresh token

	dataReturn := map[string]interface{}{
		"access_token":  at.Token,
		"refresh_token": rt.Token,
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, dataReturn)
}

// AdminLoginGatewayTmpPasswordForm struct
type AdminLoginGatewayTmpPasswordForm struct {
	Username string `form:"username" json:"username" valid:"Required"`
	Password string `form:"password" json:"password" valid:"Required"`
	LangCode string `form:"lang_code" json:"lang_code"`
}

// AdminLoginGateway function
func AdminLoginGatewayTmpPassword(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AdminLoginGatewayTmpPasswordForm
	)

	route := c.Request.URL.String()
	platformCheckingRst := strings.Contains(route, "/api/app")
	platform := "HTMLFIVE"
	if platformCheckingRst {
		platform = "APP"
	}

	sourceInterface, _ := c.Get("source")
	source := sourceInterface.(int)

	// validate input
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	// check language
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if form.LangCode != "" {
		ok = models.ExistLangague(form.LangCode)
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	} else if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// check password
	decryptedText, err := util.RsaDecryptPKCS1v15(form.Password)
	if err != nil {
		base.LogErrorLog("AdminLoginGatewayTmpPassword-RsaDecryptPKCS1v15_Failed", err.Error(), form.Password, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_password"}, nil)
		return
	}

	form.Password = decryptedText

	// start get the correct current member info
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
	)
	entMember, _ := models.GetEntMemberFn(arrCond, "", true)
	// emd get the correct current member info

	if entMember == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}
	if entMember.Status == "T" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "member_no_bind_mnemonic_yet"}, nil)
		return
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "  member_id = ?", CondValue: entMember.ID},
		models.WhereCondFn{Condition: " expired_at >= NOW() AND tmp_pw = ?", CondValue: form.Password},
	)
	validTmpPw, err := models.GetEntMemberTmpPwFn(arrCond, false)
	if err != nil {
		base.LogErrorLog("AdminLoginGatewayTmpPassword-GetEntMemberTmpPwFn_Failed", err.Error(), arrCond, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_password"}, nil)
		return
	}

	if len(validTmpPw) < 1 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_password"}, nil)
		return
	}

	// start get the correct current member info
	arrAdminEntMemberMemberData := map[string]string{
		"member_id": strconv.Itoa(entMember.ID),
	}
	arrCond = make([]models.WhereCondFn, 0)
	member, _ := models.GetAdminEntMemberMemberFn(arrCond, arrAdminEntMemberMemberData, false)
	// emd get the correct current member info

	tx := models.Begin()
	// start update current member login
	err = member_service.UpdateCurrentProfileWithLoginMember(tx, member.MainID, member.EntMemberID, source)
	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_login_info"}, nil)
		return
	}
	models.Commit(tx)
	// end update current member login

	tx = models.Begin()
	processRst, err := member_service.ProcessMemberLogin(tx, member, langCode, platform, uint8(source))

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: validTmpPw[0].ID},
	)
	updateColumn := map[string]interface{}{"t_token": processRst.ATUUID}
	err = models.UpdatesFnTx(tx, "ent_member_tmp_pw", arrUpdCond, updateColumn, false)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("AdminLoginGatewayTmpPassword-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}
	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("AdminLoginGatewayTmpPassword-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}

	dataReturn := map[string]interface{}{
		// "username":        member.NickName,
		"access_token": processRst.AccessToken,
		// "refresh_token":   processRst.RefreshToken,
		// "expired_at_unix": processRst.ExpiredAtUnix,
		// "expired_at":      processRst.ExpiredAt,
	}
	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, dataReturn)
}

type AddressLoginMemberFrom struct {
	Address  string `form:"address" json:"address" valid:"Required"`
	LangCode string `form:"lang_code" json:"lang_code" valid:"Required;MinSize(2)"`
}

// AdminGenerateMemberAccess function
func AddressGenerateMemberAccess(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AddressLoginMemberFrom
	)
	platform := "HTMLFIVE"

	// validate input
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

	if langCode == "zh-CN" {
		langCode = "zh"
	}
	if form.LangCode == "zh-CN" {
		form.LangCode = "zh"
	}

	sourceInterface, ok := c.Get("source")
	if ok == false {
		base.LogErrorLog("Login-invalid_source", "", "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	source := sourceInterface.(int)

	address := strings.Trim(form.Address, " ")

	// find member
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ent_member.source = ?", CondValue: "addr"},
	)

	arrAdminEntMemberMemberData := map[string]string{
		"nick_name": address,
	}
	member, _ := models.GetAdminEntMemberMemberFn(arrCond, arrAdminEntMemberMemberData, false)

	if member == nil || member.ID < 1 {
		//create member if not exists
		tx := models.Begin()

		subID, err := models.GenerateMemberSubID()
		if err != nil {
			tx.Rollback()
			base.LogErrorLog("LoginByAddress-GenerateMemberSubID Failed", err.Error(), "", true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
			return
		}

		// add member
		arrMemberFn := models.Members{
			SubID:       subID,
			UserTypeID:  66,
			UserGroupID: 3,
			Status:      "A",
		}

		arrMember, err := models.AddMember(tx, arrMemberFn)

		if err != nil {
			tx.Rollback()
			base.LogErrorLog("LoginByAddress-AddMember Failed", err.Error(), "", true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
			return
		}

		mainID := arrMember.ID

		memCode := member_service.GenRandomMemberCode()
		curDate, err := base.GetCurrentTimeV2("yyyy-mm-dd")
		arrAddEntMemberFn := models.AddEntMemberStruct{
			CompanyID:          1,
			MainID:             mainID,
			MemberType:         "MEM",
			Source:             "addr",
			NickName:           address,
			Code:               memCode,
			CurrentProfile:     1,
			Status:             "A",
			JoinDate:           curDate,
			PreferLanguageCode: langCode,
		}

		entMember, err := models.AddEntMember(tx, arrAddEntMemberFn)

		if err != nil {
			tx.Rollback()
			base.LogErrorLog("LoginByAddress-AddEntMember Failed", err.Error(), "", true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
			return
		}

		arrEntMemberTreeSponsorFn := models.EntMemberTreeSponsor{
			MemberID:   entMember.ID,
			MemberLot:  "01",
			UplineID:   1,
			UplineLot:  "01",
			SponsorID:  1,
			SponsorLot: "01",
			Lvl:        2,
			CreatedBy:  entMember.ID,
		}

		_, err = models.AddEntMemberTreeSponsor(tx, arrEntMemberTreeSponsorFn)
		if err != nil {
			tx.Rollback()
			base.LogErrorLog("LoginByAddress-AddEntMemberTreeSponsor Failed", err.Error(), "", true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
			return
		}

		tx.Commit()

		member, _ = models.GetAdminEntMemberMemberFn(arrCond, arrAdminEntMemberMemberData, false)

	}

	tx := models.Begin()

	processRst, err := member_service.ProcessMemberLogin(tx, member, form.LangCode, platform, uint8(source))
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("LoginByAddress-ProcessMemberLogin", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("LoginByAddress-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "login_failed"}, nil)
		return
	}

	serverDomain := setting.Cfg.Section("custom").Key("MemberServerDomain").String()

	dataReturn := map[string]interface{}{
		"access_token": processRst.AccessToken,
		"url":          serverDomain + "/gateway/" + processRst.AccessToken + "/" + form.LangCode,
		// "refresh_token": processRst.RefreshToken,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, dataReturn)
}
