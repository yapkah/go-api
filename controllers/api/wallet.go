package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/wallet_service"
)

// WalletBalanceListApiv1Form struct
type WalletBalanceListApiv1Form struct {
	Username string `form:"username" json:"username" valid:"Required;"`
}

//func GetWalletBalanceListApiv1 function
func GetWalletBalanceListApiv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form WalletBalanceListApiv1Form
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

	if form.Username != "" {
		form.Username = strings.ToLower(form.Username)
	}

	// if form.Username != "trader_sell" && form.Username != "trader_buy" {
	// 	base.LogErrorLog("GetWalletBalanceListApiv1-invalid_username", form.Username, nil, true)
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_username"}, nil)
	// 	return
	// }

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	member, _ := models.GetEntMemberFn(arrCond, "", false)
	if member == nil {
		base.LogErrorLog("GetWalletBalanceListApiv1-GetEntMemberFn_failed", arrCond, nil, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_username"}, nil)
		return
	}

	ewtTypeCode := "'USDT', 'LIGA', 'SEC'"
	// if form.Username == "trader_sell" {
	// 	ewtTypeCode = "'LIGA', 'SEC'"
	// } else if form.Username == "trader_buy" {
	// 	ewtTypeCode = "'USDT'"
	// }

	arrDataReturn, err := wallet_service.GetWalletBalanceListApiv1(member.ID, ewtTypeCode, langCode)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}
