package member

import (
	"fmt"
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

// RegisterByMobileForm struct
type MemberNotificationForm struct {
	Popup    int    `form:"pop_up" json:"pop_up"`
	Scenario string `form:"scenario" json:"scenario"`
}

//func GetMemberNotificationPopUpListv1 function
func GetMemberNotificationListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberNotificationForm
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	// u, ok := c.Get("access_user")
	// if !ok {
	// 	message := app.MsgStruct{
	// 		Msg: "invalid_member",
	// 	}
	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, "")
	// 	return
	// }

	// member := u.(*models.EntMemberMembers)

	// member.EntMemberID = 7098
	arrData := notification_service.MemberNotificationStruct{
		// MemberID: member.EntMemberID,
		LangCode: langCode,
		PopUp:    form.Popup,
		Scenario: form.Scenario,
	}
	arrMemberNotificationPopUpList := notification_service.GetMemberNotificationListv1(arrData)

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrMemberNotificationPopUpList)
}

// MemberPushNotificationListForm struct
type MemberPushNotificationListv1Form struct {
	Page     int64  `form:"page" json:"page"`
	ViewType string `form:"view_type" json:"view_type"`
}

//func GetMemberPushNotificationListv1 function
func GetMemberPushNotificationListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberPushNotificationListv1Form
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

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

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	prjIDInterface, _ := c.Get("prjID")
	prjID := prjIDInterface.(int)

	if strings.ToLower(form.ViewType) == "land" {
		arrData := notification_service.MemberSysNotificationPaginateStruct{
			ApiKeyID: prjID,
			MemberID: member.EntMemberID,
			LangCode: langCode,
		}
		arrDataReturn = notification_service.GetMemberSysNotificationLandListv1(arrData)

	} else {
		arrData := notification_service.MemberSysNotificationPaginateStruct{
			ApiKeyID: prjID,
			MemberID: member.EntMemberID,
			LangCode: langCode,
			Page:     form.Page,
		}
		fmt.Println("MemberSysNotificationPaginateStruct arrData:", arrData)
		arrDataReturn = notification_service.GetMemberSysNotificationPaginateListv1(arrData)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

type ProcessRemovePushNotificationForm struct {
	PNIDList []int `form:"pn_id_list" json:"pn_id_list" valid:"Required;"`
}

//func ProcessRemoveMemberPushNotificationv1 function
func ProcessRemoveMemberPushNotificationv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form ProcessRemovePushNotificationForm
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

	if len(form.PNIDList) < 1 {
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
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

	tx := models.Begin()

	pnIDListStr := ""
	for _, pNIDListV := range form.PNIDList {
		pnIDString := strconv.Itoa(pNIDListV)
		if pnIDListStr != "" {
			pnIDListStr = pnIDListStr + "," + pnIDString
		} else {
			pnIDListStr = pnIDString
		}
	}

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: " sys_notification.id IN (" + pnIDListStr + ") AND sys_notification.member_id = ?", CondValue: member.EntMemberID},
	)

	updateColumn := map[string]interface{}{"b_show": 0, "updated_by": member.EntMemberID}
	err := models.UpdatesFnTx(tx, "sys_notification", arrUpdCond, updateColumn, false)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("ProcessRemoveMemberPushNotificationv1-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}
