package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/service/system_service"
)

// GetFaqListForm struct
type GetFaqListForm struct {
	Page int64 `form:"page" json:"page"`
}

// GetFaqList func
func GetFaqList(c *gin.Context) {
	var (
		appG     = app.Gin{C: c}
		form     GetFaqListForm
		langCode string = "en"
		page     int64  = 1
	)

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	if form.Page != 0 {
		page = form.Page
	}

	var getFaqListParam = system_service.GetFaqListParam{
		Page: page,
	}

	faqList, errMsg := system_service.GetFaqList(getFaqListParam, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, faqList)
}

// GetAboutUsDetails func
func GetAboutUsDetails(c *gin.Context) {
	var (
		appG            = app.Gin{C: c}
		langCode string = "en"
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	aboutUsDetails, errMsg := system_service.GetAboutUsDetails(langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, aboutUsDetails)
}

// GetCurrentServerTime func
func GetCurrentServerTime(c *gin.Context) {
	var appG = app.Gin{C: c}

	var serverTime = system_service.GetCurrentServerTime()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
		"timestamp": serverTime,
		"offset":    8,
		"format":    "EEEE d MMM y, HH:mm a",
	})
	return
}
