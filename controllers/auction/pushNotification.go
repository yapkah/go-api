package auction

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/service/notification_service"
)

//func ProcessSendIndPushNotification function
func ProcessSendIndPushNotification(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form notification_service.ProcessSendIndPushNotificationForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}
	sourceInterface, _ := c.Get("source")
	source := uint8(sourceInterface.(int))
	prjIDInterface, _ := c.Get("prjID")
	prjID := uint8(prjIDInterface.(int))

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if form.LangCode != "" {
		ok := models.ExistLangague(form.LangCode)
		if ok {
			langCode = form.LangCode
		}
	} else if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	tx := models.Begin()

	arrSendPushNotificationIndStruct := notification_service.ProcessSendPushNotificationIndFromApiReqStruct{
		CryptoAddress: form.CryptoAddress,
		Subject:       form.Subject,
		SubjectParams: form.SubjectParams,
		Msg:           form.Msg,
		MsgParams:     form.MsgParams,
		CustMsg:       form.CustMsg,
		LangCode:      langCode,
		Source:        source,
		PrjID:         prjID,
	}

	err := notification_service.ProcessSendPushNotificationIndFromApiReq(tx, arrSendPushNotificationIndStruct)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("ProcessSendIndPushNotification-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "stake_laliga_failed"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}

// ProcessSendBatchPushNotificationForm struct
type ProcessSendBatchPushNotificationForm struct {
	PNList []notification_service.ProcessSendIndPushNotificationForm `form:"pn_list" json:"pn_list" valid:"Required;"`
}

//func ProcessSendBatchPushNotification function
func ProcessSendBatchPushNotification(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form ProcessSendBatchPushNotificationForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}
	sourceInterface, _ := c.Get("source")
	source := uint8(sourceInterface.(int))
	prjIDInterface, _ := c.Get("prjID")
	prjID := uint8(prjIDInterface.(int))

	maxAllowPN := 100
	if len(form.PNList) > maxAllowPN {
		msgData := map[string]string{
			"amount": strconv.Itoa(maxAllowPN),
		}
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "max_pn_list_item_is_:amount", Params: msgData}, nil)
		return
	}

	tx := models.Begin()

	arrSendPushNotificationIndStruct := notification_service.ProcessPushNotificationBatchFromApiReqStruct{
		PNList: form.PNList,
		Source: source,
		PrjID:  prjID,
	}

	err := notification_service.ProcessPushNotificationBatchFromApiReq(tx, arrSendPushNotificationIndStruct)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("ProcessSendBatchPushNotification-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "stake_laliga_failed"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}

// MemberPushNotificationListForm struct
type MemberPushNotificationListForm struct {
	CryptoAddress string `form:"crypto_address" json:"crypto_address" valid:"Required;"`
	Page          int64  `form:"page" json:"page"`
	ViewType      string `form:"view_type" json:"view_type"`
	LangCode      string `form:"lang_code" json:"lang_code"`
}

//func GetMemberPushNotificationList function
func GetMemberPushNotificationList(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberPushNotificationListForm
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if form.LangCode != "" {
		ok := models.ExistLangague(form.LangCode)
		if ok {
			langCode = form.LangCode
		}
	} else if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	if form.Page < 1 {
		form.Page = 1
	}

	arrEntMemCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemCryptoFn = append(arrEntMemCryptoFn,
		models.WhereCondFn{Condition: " crypto_address = ?", CondValue: form.CryptoAddress},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEntMemCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemCryptoFn, false)

	if arrEntMemCrypto == nil {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	prjIDInterface, _ := c.Get("prjID")
	prjID := prjIDInterface.(int)

	if strings.ToLower(form.ViewType) == "land" {
		arrData := notification_service.MemberSysNotificationPaginateStruct{
			ApiKeyID: prjID,
			MemberID: arrEntMemCrypto.MemberID,
			LangCode: langCode,
		}

		arrDataReturn = notification_service.GetMemberSysNotificationLandListv1(arrData)

	} else {
		arrData := notification_service.MemberSysNotificationPaginateStruct{
			ApiKeyID: prjID,
			MemberID: arrEntMemCrypto.MemberID,
			LangCode: langCode,
			Page:     form.Page,
		}

		arrDataReturn = notification_service.GetMemberSysNotificationPaginateListv1(arrData)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// RemoveMemberPushNotificationForm struct
type RemoveMemberPushNotificationForm struct {
	ID            int    `form:"id" json:"id" valid:"Required;"`
	CryptoAddress string `form:"crypto_address" json:"crypto_address" valid:"Required;"`
	LangCode      string `form:"lang_code" json:"lang_code"`
}

//func RemoveMemberPushNotification function
func RemoveMemberPushNotification(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form RemoveMemberPushNotificationForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}
	// sourceInterface, _ := c.Get("source")
	// source := uint8(sourceInterface.(int))
	// prjIDInterface, _ := c.Get("prjID")
	// prjID := uint8(prjIDInterface.(int))

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if form.LangCode != "" {
	// 	ok := models.ExistLangague(form.LangCode)
	// 	if ok {
	// 		langCode = form.LangCode
	// 	}
	// } else if c.GetHeader("Accept-Language") != "" {
	// 	ok := models.ExistLangague(c.GetHeader("Accept-Language"))
	// 	if ok {
	// 		langCode = c.GetHeader("Accept-Language")
	// 	}
	// }
	arrEntMemCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemCryptoFn = append(arrEntMemCryptoFn,
		models.WhereCondFn{Condition: " crypto_address = ?", CondValue: form.CryptoAddress},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEntMemCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemCryptoFn, false)

	if arrEntMemCrypto == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_crypto_address"}, nil)
		return
	}

	tx := models.Begin()

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: " id = ?", CondValue: form.ID},
		models.WhereCondFn{Condition: " member_id = ?", CondValue: arrEntMemCrypto.MemberID},
	)

	updateColumn := map[string]interface{}{"b_show": 0, "updated_by": arrEntMemCrypto.MemberID}
	err := models.UpdatesFnTx(tx, "sys_notification", arrUpdCond, updateColumn, false)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("RemoveMemberPushNotification-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "stake_laliga_failed"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}
