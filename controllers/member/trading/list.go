package trading

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/trading_service"
)

// MemberTradingBuyListForm struct
type MemberTradingBuyListForm struct {
	CryptoCode string `form:"crypto_code" json:"crypto_code"`
	Page       int64  `form:"page" json:"page"`
}

//func GetMemberTradingBuyListv1 function
func GetMemberTradingBuyListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingBuyListForm
		arrDataReturn interface{}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	if form.Page < 1 {
		form.Page = 1
	}

	// if strings.ToLower(form.ViewType) == "land" {
	// 	// member.EntMemberID = 7098
	// 	arrData := announcement_service.MemberAnnouncementLandStruct{
	// 		MemberID:   member.EntMemberID,
	// 		LangCode:   langCode,
	// 		CryptoCode: form.CryptoCode,
	// 	}

	// 	arrDataReturn = announcement_service.GetMemberAnnouncementLandv1(arrData)

	// } else {
	arrData := trading_service.MemberTradingBuyListPaginateStruct{
		EntMemberID: member.EntMemberID,
		CryptoCode:  form.CryptoCode,
		LangCode:    langCode,
		Page:        form.Page,
	}

	arrDataReturn = trading_service.GetMemberTradingBuyPaginateListv1(arrData)
	// }

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// MemberTradingSellListForm struct
type MemberTradingSellListForm struct {
	CryptoCode string `form:"crypto_code" json:"crypto_code"`
	Page       int64  `form:"page" json:"page"`
}

//func GetMemberTradingSellListv1 function
func GetMemberTradingSellListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingBuyListForm
		arrDataReturn interface{}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	if form.Page < 1 {
		form.Page = 1
	}

	// if strings.ToLower(form.ViewType) == "land" {
	// 	// member.EntMemberID = 7098
	// 	arrData := announcement_service.MemberAnnouncementLandStruct{
	// 		MemberID:   member.EntMemberID,
	// 		LangCode:   langCode,
	// 		CryptoCode: form.CryptoCode,
	// 	}

	// 	arrDataReturn = announcement_service.GetMemberAnnouncementLandv1(arrData)

	// } else {
	arrData := trading_service.MemberTradingSellListPaginateStruct{
		EntMemberID: member.EntMemberID,
		CryptoCode:  form.CryptoCode,
		LangCode:    langCode,
		Page:        form.Page,
	}

	arrDataReturn = trading_service.GetMemberTradingSellPaginateListv1(arrData)
	// }

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// MemberTradingTransactionListForm struct
type MemberTradingTransactionListForm struct {
	Type string `form:"type" json:"type"`
	Page int64  `form:"page" json:"page"`
}

//func GetMemberTradingTransListv1 function
func GetMemberTradingTransListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingTransactionListForm
		arrDataReturn interface{}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := trading_service.MemberTradingTransListPaginateStruct{
		MemberID: member.EntMemberID,
		Type:     form.Type,
		LangCode: langCode,
		Page:     form.Page,
	}

	arrDataReturn = trading_service.GetMemberTradingTransPaginateListv1(arrData)

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// MemberTradingMarketListForm struct
type MemberTradingMarketListForm struct {
	Page int64 `form:"page" json:"page"`
}

//func GetMemberTradingMarketListv1 function
func GetMemberTradingMarketListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingMarketListForm
		arrDataReturn interface{}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := trading_service.MemberTradingMarketListPaginateStruct{
		MemberID: member.EntMemberID,
		LangCode: langCode,
		Page:     form.Page,
	}

	arrDataReturn = trading_service.GetMemberTradingMarketPaginateListv1(arrData)

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// MemberTradingMarketPriceListForm struct
type MemberTradingMarketPriceListForm struct {
	CryptoCode   string `form:"crypto_code" json:"crypto_code" valid:"Required;"`
	Quantitative string `form:"quantitative" json:"quantitative" valid:"Required;"`
}

//func GetMemberAvailableTradingBuyListv1 function
func GetMemberAvailableTradingBuyListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingMarketPriceListForm
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
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

	if strings.ToLower(form.CryptoCode) == "sec" {
		arrDataReturn = trading_service.GetAvailableSecTradingBuyList(form.Quantitative, langCode)
	} else if strings.ToLower(form.CryptoCode) == "liga" {
		fmt.Println("here")
		arrDataReturn = trading_service.GetAvailableLigaTradingBuyList(form.Quantitative, langCode)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

//func GetMemberAvailableTradingSellListv1 function
func GetMemberAvailableTradingSellListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingMarketPriceListForm
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
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

	if strings.ToLower(form.CryptoCode) == "sec" {
		arrDataReturn = trading_service.GetAvailableSecTradingSellList(form.Quantitative, langCode)
	} else if strings.ToLower(form.CryptoCode) == "liga" {
		arrDataReturn = trading_service.GetAvailableLigaTradingSellList(form.Quantitative, langCode)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// MemberTradingHistoryTransListForm struct
type MemberTradingHistoryTransListForm struct {
	ActionType string `form:"action_type" json:"action_type"`
	Page       int    `form:"page" json:"page"`
	DateFrom   string `form:"date_from" json:"date_from"`
	DateTo     string `form:"date_to" json:"date_to"`
}

// func GetMemberTradingOpenOrderTransListv1
func GetMemberTradingOpenOrderTransListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingHistoryTransListForm
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := trading_service.MemberTradingHistoryTransListv1{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
		Page:        form.Page,
	}

	if form.DateFrom != "" {
		arrData.DateFrom = form.DateFrom
	}

	if form.DateTo != "" {
		arrData.DateTo = form.DateTo
	}

	if strings.ToLower(form.ActionType) == "buy" {
		arrDataReturn = trading_service.GetMemberTradingBuyOpenOrderTransListv1(arrData)
	} else if strings.ToLower(form.ActionType) == "sell" {
		arrDataReturn = trading_service.GetMemberTradingSellOpenOrderTransListv1(arrData)
	} else {
		arrDataReturn = trading_service.GetMemberTradingOpenOrderTransListv1(arrData)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// func GetMemberTradingOrderHistoryTransListv1
func GetMemberTradingOrderHistoryTransListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingHistoryTransListForm
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := trading_service.MemberTradingHistoryTransListv1{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
		Page:        form.Page,
	}

	if form.DateFrom != "" {
		arrData.DateFrom = form.DateFrom
	}

	if form.DateTo != "" {
		arrData.DateTo = form.DateTo
	}

	if strings.ToLower(form.ActionType) == "buy" {
		arrDataReturn = trading_service.GetMemberTradingBuyOrderHistoryTransListv1(arrData)
	} else if strings.ToLower(form.ActionType) == "sell" {
		arrDataReturn = trading_service.GetMemberTradingSellOrderHistoryTransListv1(arrData)
	} else {
		arrDataReturn = trading_service.GetMemberTradingOrderHistoryTransListv1(arrData)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// func GetMemberTradingHistoryTransListv1
func GetMemberTradingHistoryTransListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingHistoryTransListForm
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := trading_service.MemberTradingHistoryTransListv1{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
		Page:        form.Page,
	}

	if form.DateFrom != "" {
		arrData.DateFrom = form.DateFrom
	}

	if form.DateTo != "" {
		arrData.DateTo = form.DateTo
	}

	arrDataReturn = trading_service.GetMemberTradingHistoryTransListv1(arrData)

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// MemberTradingOrderHistoryTransDetailsForm struct
type MemberTradingOrderHistoryTransDetailsForm struct {
	ActionType string `form:"action_type" json:"action_type" valid:"Required;"`
	ID         int    `form:"id" json:"id" valid:"Required;"`
}

// func GetMemberTradingOrderHistoryTransDetailsv1
func GetMemberTradingOrderHistoryTransDetailsv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberTradingOrderHistoryTransDetailsForm
		arrDataReturn interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

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

	arrData := trading_service.MemberTradingOrderHistoryTransDetailsv1{
		EntMemberID: member.EntMemberID,
		ID:          form.ID,
		LangCode:    langCode,
	}

	if strings.ToLower(form.ActionType) == "buy" {
		arrDataReturn = trading_service.GetMemberTradingBuyOrderHistoryTransDetailsv1(arrData)
	} else if strings.ToLower(form.ActionType) == "sell" {
		arrDataReturn = trading_service.GetMemberTradingSellOrderHistoryTransDetailsv1(arrData)
	} else {

	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}
