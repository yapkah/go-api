package member

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/reward_service"
)

//Get reward summary
type GetRewardSummaryForm struct {
	Page     int64  `form:"page" json:"page"`
	DateFrom string `form:"date_from" json:"date_from" valid:"MaxSize(100)"`
	DateTo   string `form:"date_to" json:"date_to" valid:"MaxSize(100)"`
	RwdType  string `form:"reward_type_code"`
}

func GetRewardSummary(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetRewardSummaryForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	rwdSummaryDet := reward_service.RewardStatementPostStruct{
		MemberID: entMemberID,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		RwdType:  strings.ToUpper(form.RwdType),
		Page:     form.Page,
		LangCode: langCode,
	}

	rst, err := rwdSummaryDet.RewardSummary()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "failed_to_get_reward_summary"}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, rst)
	return
}

//Get reward statement
type GetRewardStatementForm struct {
	Page     int64  `form:"page" json:"page"`
	DateFrom string `form:"date_from" json:"date_from" valid:"MaxSize(100)"`
	DateTo   string `form:"date_to" json:"date_to" valid:"MaxSize(100)"`
	RwdType  string `form:"reward_type_code" valid:"Required"`
	// WalletTypeCode string `form:"wallet_type_code"`
}

func GetRewardStatement(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetRewardStatementForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	rwdStatement := reward_service.RewardStatementPostStruct{
		MemberID: entMemberID,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		RwdType:  strings.ToUpper(form.RwdType),
		Page:     form.Page,
		LangCode: langCode,
		// WalletTypeCode: form.WalletTypeCode,
	}

	rst, err := rwdStatement.RewardStatement()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "failed_to_get_reward_statement"}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, rst)
	return
}

//Get reward history
type GetRewardHistoryForm struct {
	Page     int64  `form:"page" json:"page"`
	DateFrom string `form:"date_from" json:"date_from" valid:"MaxSize(100)"`
	DateTo   string `form:"date_to" json:"date_to" valid:"MaxSize(100)"`
}

func GetRewardHistory(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetRewardHistoryForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	rwdHistory := reward_service.RewardHistoryPostStruct{
		MemberID: entMemberID,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		Page:     form.Page,
		LangCode: langCode,
	}

	rst, err := rwdHistory.RewardHistory()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "failed_to_get_reward_history"}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, rst)
	return
}

type GetRewardGraphForm struct {
	DateFrom string `form:"date_from" json:"date_from" valid:"MaxSize(100)"`
	DateTo   string `form:"date_to" json:"date_to" valid:"MaxSize(100)"`
	RwdType  string `form:"reward_type_code" valid:"Required"`
	Type     string `form:"type" valid:"Required"`
}

func GetRewardGraph(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetRewardGraphForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	rwdGraphDet := reward_service.RewardGraphPostStruct{
		MemberID: entMemberID,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		RwdType:  strings.ToUpper(form.RwdType),
		Type:     strings.ToUpper(form.Type),
		LangCode: langCode,
	}

	rst, err := rwdGraphDet.RewardGraph()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "failed_to_get_reward_graph"}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, rst)
	return
}
