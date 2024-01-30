package event

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/event_service"
)

// GetEventSponsorRankingSettingForm struct
type GetEventSponsorRankingSettingForm struct {
	Type    string `form:"type" json:"type" valid:"Required"` // "ND", "ALL"
	BatchNo int    `form:"batch_no" json:"batch_no" valid:"Required;"`
}

// GetEventSponsorRankingSetting function
func GetEventSponsorRankingSetting(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetEventSponsorRankingSettingForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	arrEventSponsorRankingSetting, errMsg := event_service.GetEventSponsorRankingSetting(form.Type, form.BatchNo, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrEventSponsorRankingSetting)
}

// GetEventSponsorRankingListForm struct
type GetEventSponsorRankingListForm struct {
	Type     string `form:"type" json:"type" valid:"Required"` // "ND", "ALL"
	BatchNo  int    `form:"batch_no" json:"batch_no" valid:"Required;"`
	DateFrom string `form:"date_from" json:"date_from"` // if non given, take latest list
	DateTo   string `form:"date_to" json:"date_to"`
	Page     int64  `form:"page" json:"page"`
}

// GetEventSponsorRankingList function
func GetEventSponsorRankingList(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetEventSponsorRankingListForm
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	langCode = c.GetHeader("Accept-Language")
	// }

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := event_service.GetEventSponsorRankingListStruct{
		Type:     form.Type,
		BatchNo:  form.BatchNo,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		Page:     form.Page,
	}

	arrEventSponsorRankingList, errMsg := event_service.GetEventSponsorRankingList(arrData)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrEventSponsorRankingList)
}

// GetAuctionLuckyNumberList function
func GetAuctionLuckyNumberList(c *gin.Context) {
	var (
		appG                      = app.Gin{C: c}
		status                    = 0
		title                     string
		arrAuctionLuckyNumberList interface{}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	arrData := event_service.GetAuctionLuckyNumberListStruct{
		LangCode: langCode,
	}

	var (
		curDateTime    = time.Now()
		setDateTime, _ = base.StrToDateTime("2021-05-01", "2006-01-02")
	)

	if helpers.CompareDateTime(curDateTime, ">=", setDateTime) {
		status = 1
	}

	status = 0
	if status == 1 {
		title, arrAuctionLuckyNumberList = event_service.GetAuctionLuckyNumberList(arrData)
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{"status": status, "history": status, "title": title, "list": arrAuctionLuckyNumberList})
}

// GetAuctionLuckyNumberHistoryListForm struct
type GetAuctionLuckyNumberHistoryListForm struct {
	DateFrom string `form:"date_from" json:"date_from"` // if non given, take latest list
	DateTo   string `form:"date_to" json:"date_to"`
}

// GetAuctionLuckyNumberHistoryList function
func GetAuctionLuckyNumberHistoryList(c *gin.Context) {
	var (
		appG                      = app.Gin{C: c}
		status                    = 1
		arrAuctionLuckyNumberList interface{}
		form                      GetAuctionLuckyNumberHistoryListForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	arrData := event_service.GetAuctionLuckyNumberListStruct{
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		LangCode: langCode,
	}

	if status == 1 {
		_, arrAuctionLuckyNumberList = event_service.GetAuctionLuckyNumberList(arrData)
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrAuctionLuckyNumberList)
}
