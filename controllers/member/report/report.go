package report

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/service/report_service"
)

// GetReportSetup func
func GetReportSetup(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	// retrieve langCode
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// retrieve product group setting
	generalSetting, errMsg := report_service.GetEntMemberReport(langCode, entMemberID)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, generalSetting)
}

// GetReportList func
func GetReportList(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form report_service.GetReportListForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	member := u.(*models.EntMemberMembers)
	form.MemberID = member.EntMemberID

	// retrieve langCode
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}
	form.LangCode = langCode

	var arrDataReturn = report_service.ReportHandler(form)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}
