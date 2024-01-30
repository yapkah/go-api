package member

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/member_service"
)

//func GetMemberAccountListv1
func GetMemberAccountListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	arrData := member_service.MemberAccountListStruct{
		MemberID:    member.ID,
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}

	arrDataReturn := member_service.GetMemberAccountListv1(arrData)

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// SwitchCurrentActiveMemberAccountForm struct
type SwitchCurrentActiveMemberAccountForm struct {
	Username string `form:"username" json:"username"`
}

//func SwitchCurrentActiveMemberAccountv1
func SwitchCurrentActiveMemberAccountv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form SwitchCurrentActiveMemberAccountForm
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	ok := models.ExistLangague(c.GetHeader("Accept-Language"))
	// 	if ok {
	// 		langCode = c.GetHeader("Accept-Language")
	// 	}
	// }

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

	sourceInterface, _ := c.Get("source")
	source := sourceInterface.(int)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrTargetEntMember, _ := models.GetEntMemberFn(arrCond, "", false)
	if arrTargetEntMember == nil || arrTargetEntMember.ID < 1 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_account"}, nil)
		return
	}

	tx := models.Begin()

	_, currentActiveAccount := base.GetMemberLogData([]string{"ent_member"}, member.EntMemberID)

	arrCurrentData := map[string]interface{}{
		"current_active":  currentActiveAccount,
		"target_username": arrTargetEntMember,
	}
	arrData := member_service.SwitchCurrentActiveMemberAccountv2Struct{
		EntMemberID:  member.EntMemberID,
		MemberMainID: member.ID,
		UsernameTo:   form.Username,
		SourceID:     source,
	}
	fmt.Println("SwitchCurrentActiveMemberAccountv1 SourceID:", source)
	logData, err := member_service.SwitchCurrentActiveMemberAccountv2(tx, arrData)
	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "fail_to_switch_account"}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("SwitchCurrentActiveMemberAccountv1-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "fail_to_switch_account"}, nil)
		return
	}

	if logData {
		_, currentActiveUpdatedData := base.GetMemberLogData([]string{"ent_member"}, member.EntMemberID)
		_, TargetAccountUpdatedData := base.GetMemberLogData([]string{"ent_member"}, arrTargetEntMember.ID)

		arrUpdatedData := map[string]interface{}{
			"current_update_data": currentActiveUpdatedData,
			"target_update_data":  TargetAccountUpdatedData,
		}
		base.AddSysLog(member.EntMemberID, arrCurrentData, arrUpdatedData, "modify", "switch-account", c)
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}

// DeactivateMemberAccountForm struct
type DeactivateMemberAccountForm struct {
	Username     string `form:"username" json:"username"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func DeactivateMemberAccountv1
func DeactivateMemberAccountv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form DeactivateMemberAccountForm
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	ok := models.ExistLangague(c.GetHeader("Accept-Language"))
	// 	if ok {
	// 		langCode = c.GetHeader("Accept-Language")
	// 	}
	// }

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("DeactivateMemberAccountv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
		message := app.MsgStruct{
			Msg: "invalid_secondary_pin_format",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              member.EntMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrTargetEntMember, _ := models.GetEntMemberFn(arrCond, "", false)
	if arrTargetEntMember == nil || arrTargetEntMember.ID < 1 {
		message := app.MsgStruct{
			Msg: "invalid_account",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	if arrTargetEntMember.CurrentProfile == 1 {
		message := app.MsgStruct{
			Msg: "current_active_account_cannot_be_delete_please_choose_another_account_if_available",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	tx := models.Begin()

	arrCurrentData := map[string]interface{}{
		"ent_member": arrTargetEntMember,
	}
	arrData := member_service.DeactivateMemberAccountv1Struct{
		EntMemberID: member.EntMemberID,
		MemberID:    member.ID,
		UsernameTo:  form.Username,
	}
	logData, err := member_service.DeactivateMemberAccountv1(tx, arrData)
	if err != nil {
		models.Rollback(tx)
		message := app.MsgStruct{
			Msg: err.Error(),
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("DeactivateMemberAccountv1-Commit Failed", err.Error(), "", true)
		message := app.MsgStruct{
			Msg: "fail_to_switch_account",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	if logData {
		_, arrTargetAccountUpdatedData := base.GetMemberLogData([]string{"ent_member"}, arrTargetEntMember.ID)

		arrUpdatedData := map[string]interface{}{
			"ent_member": arrTargetAccountUpdatedData,
		}
		base.AddSysLog(member.EntMemberID, arrCurrentData, arrUpdatedData, "modify", "deactivate-account", c)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, nil)
}

// UnbindMemberAccountForm struct
type UnbindMemberAccountForm struct {
	Username     string `form:"username" json:"username" valid:"Required;"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func UnbindMemberAccountv1
func UnbindMemberAccountv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form UnbindMemberAccountForm
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	ok := models.ExistLangague(c.GetHeader("Accept-Language"))
	// 	if ok {
	// 		langCode = c.GetHeader("Accept-Language")
	// 	}
	// }

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	accessToken, ok := c.Get("access_token")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	at := accessToken.(*models.AccessToken)

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("UnbindMemberAccountv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
		message := app.MsgStruct{
			Msg: "invalid_secondary_pin_format",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              member.EntMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrTargetEntMember, _ := models.GetEntMemberFn(arrCond, "", false)
	if arrTargetEntMember == nil || arrTargetEntMember.ID < 1 {
		message := app.MsgStruct{
			Msg: "invalid_account",
		}
		appG.ResponseV2(1, http.StatusOK, message, nil)
		return
	}

	// if arrTargetEntMember.CurrentProfile == 1 {
	// 	message := app.MsgStruct{
	// 		Msg: "current_active_account_cannot_be_unbind_please_choose_another_account_if_available",
	// 	}
	// 	appG.ResponseV2(0, http.StatusOK, message, nil)
	// 	return
	// }

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "b_login = ?", CondValue: 1},
		models.WhereCondFn{Condition: "b_logout = ?", CondValue: 0},
		models.WhereCondFn{Condition: "t_token = ?", CondValue: at.ID},
	)
	appLoginLogRst, _ := models.GetAppLoginLogFn(arrCond, "", false)

	if appLoginLogRst != nil {
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " device_bind_log.t_push_noti_token = ? ", CondValue: appLoginLogRst.TPushNotiToken},
			models.WhereCondFn{Condition: " device_bind_log.member_id = ? ", CondValue: member.EntMemberID},
		)
		exsitingDeviceBindLogRst, _ := models.GetLatestDeviceBindLogFn(arrCond, false)

		if appLoginLogRst.TPushNotiToken != "" && exsitingDeviceBindLogRst.Bind == 1 {
			arrCrtData := models.DeviceBindLog{
				MemberID:       member.EntMemberID,
				TOs:            appLoginLogRst.TOs,
				TModel:         appLoginLogRst.TModel,
				TManufacturer:  appLoginLogRst.TManufacturer,
				TAppVersion:    appLoginLogRst.TAppVersion,
				TOsVersion:     appLoginLogRst.TOsVersion,
				TPushNotiToken: appLoginLogRst.TPushNotiToken,
				Bind:           0,
				CreatedBy:      strconv.Itoa(member.EntMemberID),
			}
			_ = models.AddDeviceBindLog(arrCrtData)
		}
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, nil)
}

// CheckImportMemberAccountForm struct
type CheckImportMemberAccountForm struct {
	PrivateKey string `form:"private_key" json:"private_key" valid:"Required;"`
	Username   string `form:"username" json:"username" valid:"Required;"`
}

//func CheckImportMemberAccountv1
func CheckImportMemberAccountv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form CheckImportMemberAccountForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: member.ID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		// models.WhereCondFn{Condition: "ent_member.private_key = ?", CondValue: form.PrivateKey}, // remove this in order to avoid sql injection
	)
	arrValidPrivateKey, _ := models.GetEntMemberListFn(arrCond, false)
	if len(arrValidPrivateKey) > 0 {
		decryptedText, err := util.RsaDecryptPKCS1v15(form.PrivateKey)
		if err != nil {
			base.LogErrorLog("CheckImportMemberAccountv1-form_PrivateKey_RsaDecryptPKCS1v15_Failed", err.Error(), form.PrivateKey, true)
			message := app.MsgStruct{
				Msg: "invalid_account",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		form.PrivateKey = decryptedText
		for _, arrValidPrivateKeyV := range arrValidPrivateKey {
			decryptedText2, err := util.RsaDecryptPKCS1v15(arrValidPrivateKeyV.PrivateKey)
			if err != nil {
				base.LogErrorLog("CheckImportMemberAccountv1-ent_member_PrivateKey_RsaDecryptPKCS1v15_Failed", err.Error(), arrValidPrivateKeyV.PrivateKey, true)
				message := app.MsgStruct{
					Msg: "invalid_account",
				}
				appG.ResponseV2(0, http.StatusOK, message, nil)
				return
			}
			if decryptedText2 == form.PrivateKey {
				if form.Username != arrValidPrivateKeyV.NickName {
					message := app.MsgStruct{
						Msg: "this_private_key_does_not_belongs_to_this_account",
					}
					appG.ResponseV2(0, http.StatusOK, message, nil)
					return
				}
				message := app.MsgStruct{
					Msg: "success",
				}
				appG.ResponseV2(1, http.StatusOK, message, nil)
				return
			}
		}
	}

	message := app.MsgStruct{
		Msg: "invalid_account",
	}
	appG.ResponseV2(0, http.StatusOK, message, nil)
}

// TagMemberAccountv1Form struct
type TagMemberAccountv1Form struct {
	Username       string `form:"username" json:"username" valid:"Required;"`
	TaggedUsername string `form:"tagged_username" json:"tagged_username"`
}

//func TagMemberAccountv1
func TagMemberAccountv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form TagMemberAccountv1Form
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	ok := models.ExistLangague(c.GetHeader("Accept-Language"))
	// 	if ok {
	// 		langCode = c.GetHeader("Accept-Language")
	// 	}
	// }

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	currentLoginMember := u.(*models.EntMemberMembers)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	member, _ := models.GetEntMemberFn(arrCond, "", false)

	if member == nil || member.ID < 1 {
		message := app.MsgStruct{
			Msg: "invalid_link_member_name",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	var tagEntMemberID int
	if form.TaggedUsername != "" {
		if form.Username == form.TaggedUsername {
			message := app.MsgStruct{
				Msg: "cannot_tag_same_account",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.TaggedUsername},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		arrTargetEntMember, _ := models.GetEntMemberFn(arrCond, "", false)
		if arrTargetEntMember == nil || arrTargetEntMember.ID < 1 {
			message := app.MsgStruct{
				Msg: "invalid_link_member_name",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		result := models.CheckTagMainAccount(arrTargetEntMember.ID, member.ID)
		if result != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "only_can_tag_to_main_account"}, nil)
			return
		}

		// start checking only can tagged main account for phase 1
		// if arrTargetEntMember.MainID != member.MainID { // checking on not allow others main ID
		// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "only_can_tag_to_main_account"}, nil)
		// 	return
		// }
		// if arrTargetEntMember.MainID != currentLoginMember.EntMemberID { // checking only can tag to main ID
		// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "only_can_tag_to_main_account"}, nil)
		// 	return
		// }
		// end checking only can tagged main account for phase 1
		tagEntMemberID = arrTargetEntMember.ID
	}

	tx := models.Begin()

	_, currentActiveAccount := base.GetMemberLogData([]string{"ent_member"}, member.ID)
	arrCurrentData := map[string]interface{}{
		"prev": currentActiveAccount,
	}

	arrData := member_service.TagMemberAccountv1Struct{
		CurrentLoginEntMemberID: currentLoginMember.EntMemberID,
		EntMemberID:             member.ID,
		TagEntMemberID:          tagEntMemberID,
	}
	logData, err := member_service.TagMemberAccountv1(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		message := app.MsgStruct{
			Msg: err.Error(),
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("TagMemberAccountv1-Commit Failed", err.Error(), "", true)
		message := app.MsgStruct{
			Msg: "something_went_wrong",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	if logData {
		_, currentActiveUpdatedData := base.GetMemberLogData([]string{"ent_member"}, member.ID)

		arrUpdatedData := map[string]interface{}{
			"current_update_data": currentActiveUpdatedData,
		}
		base.AddSysLog(member.ID, arrCurrentData, arrUpdatedData, "modify", "tag-account", c)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, nil)
}

// TagMemberAccountv1Form struct
// type MemberAccountTransferAllSetupForm struct {
// 	Username    string `form:"username" json:"username" valid:"Required;"`
// 	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin"`
// }

//func GetMemberAccountTransferExchangeBatchAssetsv1
func GetMemberAccountTransferExchangeBatchAssetsv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	arrData := member_service.MemberAccountTransferExchangeBatchAssetsStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}

	arrDataReturn := member_service.GetMemberAccountTransferExchangeBatchAssetsv1(arrData)

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}
