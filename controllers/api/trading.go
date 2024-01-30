package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/service/trading_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

// ProcessAutoTradingBuyRequestv1Form struct
type ProcessAutoTradingBuyRequestv1Form struct {
	CryptoCode string  `form:"crypto_code" json:"crypto_code"`
	UnitPrice  float64 `form:"unit_price" json:"unit_price"`
	Quantity   float64 `form:"quantity" json:"quantity" valid:"Required;"`
}

//func ProcessAutoTradingBuyRequestv1 function
func ProcessAutoTradingBuyRequestv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form ProcessAutoTradingBuyRequestv1Form
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: "trader_buy"},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	member, _ := models.GetEntMemberFn(arrCond, "", false)
	if member == nil {
		base.LogErrorLog("ProcessAutoTradingBuyRequestv1-missing_trader_buy_member_in_system", arrCond, nil, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_nick_name"}, nil)
		return
	}

	if form.Quantity <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "quantity_should_greater_than_0"}, nil)
		return
	}
	if form.UnitPrice <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "unit_price_should_greater_than_0"}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	arrData := trading_service.BuyMemberTradingRequestStruct{
		UnitPrice:   form.UnitPrice,
		Quantity:    form.Quantity,
		CryptoCode:  form.CryptoCode,
		EntMemberID: member.ID,
	}

	totalAmount, err := trading_service.ProcessAutoTradingBuyRequestv1(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("ProcessAutoTradingBuyRequestv1-Commit Failed", err.Error(), "", true)
		message := app.MsgStruct{
			Msg: "buy_trading_failed",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	arrDataReturn := map[string]interface{}{
		"total_payment": totalAmount,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// ProcessAutoTradingSellRequestv1Form struct
type ProcessAutoTradingSellRequestv1Form struct {
	CryptoCode string  `form:"crypto_code" json:"crypto_code" valid:"Required;"`
	UnitPrice  float64 `form:"unit_price" json:"unit_price" valid:"Required;"`
	Quantity   float64 `form:"quantity" json:"quantity" valid:"Required;"`
}

//func ProcessAutoTradingSellRequestv1 function
func ProcessAutoTradingSellRequestv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form ProcessAutoTradingSellRequestv1Form
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: "trader_sell"},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	member, _ := models.GetEntMemberFn(arrCond, "", false)
	if member == nil {
		base.LogErrorLog("ProcessAutoTradingSellRequestv1-missing_trader_sell_member_in_system", arrCond, nil, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_nick_name"}, nil)
		return
	}

	if form.Quantity <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "quantity_should_greater_than_0"}, nil)
		return
	}
	if form.UnitPrice <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "unit_price_should_greater_than_0"}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	arrData := trading_service.SellMemberTradingRequestStruct{
		UnitPrice:   form.UnitPrice,
		Quantity:    form.Quantity,
		CryptoCode:  form.CryptoCode,
		EntMemberID: member.ID,
	}

	totalAmount, err := trading_service.ProcessAutoTradingSellRequestv1(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("ProcessAutoTradingSellRequestv1-Commit Failed", err.Error(), "", true)
		message := app.MsgStruct{
			Msg: "sell_trading_failed",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	arrDataReturn := map[string]interface{}{
		"total_payment": totalAmount,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

//func GetPriceListApiv1 function
func GetPriceListApiv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	arrDataReturn := wallet_service.GetPriceListApiv1()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// OpenOrderListForm struct
type OpenOrderListForm struct {
	CryptoType string `form:"crypto_type" json:"crypto_type" valid:"Required;"`
	Action     string `form:"action" json:"action" valid:"Required;"`
}

//func GetOpenOrderListv1
func GetOpenOrderListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form OpenOrderListForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	if strings.ToLower(form.Action) == "buy" {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: form.CryptoType},
		)

		arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingBuyListFn(arrCond, 0, true)
		// tradingKeyName := strings.ToLower(form.Action) + "_open_order"

		arrDataReturn := map[string]interface{}{
			"open_order": arrAvailableTradingPriceListRst,
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)

	} else {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: form.CryptoType},
		)

		arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingSellListFn(arrCond, 0, false)

		// tradingKeyName := strings.ToLower(form.Action) + "_open_order"

		arrDataReturn := map[string]interface{}{
			"open_order": arrAvailableTradingPriceListRst,
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	}

	return

}

//func GetAutoTradingListv1
func GetAutoTradingListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form trading_service.AutoTradingListForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	var statusList = []string{"p", "c", "m"}
	statusStatus := false
	if form.Status != "" {
		statusStatus = helpers.StringInSlice(strings.ToLower(form.Status), statusList)
	}

	if !statusStatus {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_status"}, nil)
		return
	}

	arrTradingList := trading_service.GetAutoTradingListv1(form)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrTradingList)
}

// MemberCancelTradingRequestForm struct
type ProcessCancelAutoTradingRequestForm struct {
	DocNo    string `form:"doc_no" json:"doc_no" valid:"Required;"`
	Username string `form:"username" json:"username"`
	Action   string `form:"action" json:"action" valid:"Required;"`
}

//func ProcessCancelAutoTradingRequestv1 function
func ProcessCancelAutoTradingRequestv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form ProcessCancelAutoTradingRequestForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	// end add this bcz of deicmal problem
	if strings.ToLower(form.Action) == "buy" {
		username := "trader_buy"
		if form.Username != "" {
			username = form.Username
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ent_member.nick_name = ? ", CondValue: username},
		)
		entMember, _ := models.GetEntMemberFn(arrCond, "", false)
		if entMember == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_username"}, nil)
			return
		}
		// begin transaction
		tx := models.Begin()

		arrData := trading_service.ProcessCancelAutoTradingRequestForm{
			DocNo:       form.DocNo,
			EntMemberID: entMember.ID,
		}

		err := trading_service.ProcessCancelAutoTradingBuyRequestv1(tx, arrData)

		if err != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		err = models.Commit(tx)
		if err != nil {
			models.Rollback(tx)
			base.LogErrorLog("ProcessCancelAutoTradingRequestv1-Commit Failed", err.Error(), "", true)
			message := app.MsgStruct{
				Msg: "something_went_wrong",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
	} else if strings.ToLower(form.Action) == "sell" {
		username := "trader_sell"
		if form.Username != "" {
			username = form.Username
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ent_member.nick_name = ? ", CondValue: username},
		)
		entMember, _ := models.GetEntMemberFn(arrCond, "", false)
		if entMember == nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_username"}, nil)
			return
		}
		// begin transaction
		tx := models.Begin()

		arrData := trading_service.ProcessCancelAutoTradingRequestForm{
			DocNo:       form.DocNo,
			EntMemberID: entMember.ID,
		}

		err := trading_service.ProcessCancelAutoTradingSellRequestv1(tx, arrData)

		if err != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		err = models.Commit(tx)
		if err != nil {
			models.Rollback(tx)
			base.LogErrorLog("ProcessCancelAutoTradingRequestv1-Commit Failed", err.Error(), "", true)
			message := app.MsgStruct{
				Msg: "something_went_wrong",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
	} else {
		base.LogErrorLog("ProcessCancelAutoTradingRequestv1-invalid_action_type", form.Action, nil, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}
