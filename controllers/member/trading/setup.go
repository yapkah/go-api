package trading

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/service/trading_service"
)

type MemberTradingSetupForm struct {
	CryptoCode  string `form:"crypto_code" json:"crypto_code" valid:"Required"`
	RequestType string `form:"request_type" json:"request_type" valid:"Required"`
}

//func GetMemberTradingSetupv1
func GetMemberTradingSetupv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		arrDataReturn interface{}
		form          MemberTradingSetupForm
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
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

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	arrData := trading_service.MemberTradingSetupStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
		CryptoCode:  form.CryptoCode,
	}

	if strings.ToLower(form.RequestType) == "buy" {
		rst, err := trading_service.GetMemberTradingBuySetupv1(arrData)
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}
		arrDataReturn = rst
	} else if strings.ToLower(form.RequestType) == "sell" {
		rst, err := trading_service.GetMemberTradingSellSetupv1(arrData)
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}
		arrDataReturn = rst
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

//func GetMemberTradingSelectionListv1
func GetMemberTradingSelectionListv1(c *gin.Context) {
	var (
		appG          = app.Gin{C: c}
		arrDataReturn interface{}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
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

	arrData := trading_service.MemberTradingSetupStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}

	arrDataReturn, err := trading_service.GetMemberTradingSelectionListv1(arrData)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

//func GetMemberTradingViewSetupv1
func GetMemberTradingViewSetupv1(c *gin.Context) {
	var (
		appG          = app.Gin{C: c}
		arrDataReturn interface{}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	// u, ok := c.Get("access_user")
	// if !ok {
	// 	message := app.MsgStruct{
	// 		Msg: "invalid_member",
	// 	}
	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
	// 	return
	// }

	// member := u.(*models.EntMemberMembers)

	arrData := trading_service.MemberTradingSetupStruct{
		LangCode: langCode,
	}

	arrDataReturn, err := trading_service.GetMemberTradingViewSetupv1(arrData)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

//func GetMemberTradingSetupv2
func GetMemberTradingSetupv2(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		arrDataReturn interface{}
		form          MemberTradingSetupForm
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
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

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	arrData := trading_service.MemberTradingSetupStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
		CryptoCode:  form.CryptoCode,
	}

	if strings.ToLower(form.RequestType) == "buy" {
		rst, err := trading_service.GetMemberTradingBuySetupv2(arrData)
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}
		arrDataReturn = rst
	} else if strings.ToLower(form.RequestType) == "sell" {
		rst, err := trading_service.GetMemberTradingSellSetupv2(arrData)
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}
		arrDataReturn = rst
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}
