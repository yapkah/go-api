package trading

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"

	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/trading_service"
)

// MemberCancelTradingRequestForm struct
type MemberCancelTradingRequestForm struct {
	ID            int     `form:"id" json:"id" valid:"Required;"`
	TradingAction string  `form:"trading_action" json:"trading_action" valid:"Required;"`
	Quantity      float64 `form:"quantity" json:"quantity" valid:"Required;"`
	SecondaryPin  string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func MemberCancelTradingRequestv1 function
func MemberCancelTradingRequestv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberCancelTradingRequestForm
	)

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

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("MemberCancelTradingRequestv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}
	// wordCount := utf8.RuneCountInString(decryptedText)
	// if wordCount < 6 {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "SecondaryPin minimum character is 6"}, nil)
	// 	return
	// }
	form.SecondaryPin = decryptedText

	// check secondary password
	secondaryPin := base.SecondaryPin{
		MemId:              member.ID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
	}

	secondaryPinErr := secondaryPin.CheckSecondaryPin()

	if secondaryPinErr != nil {
		// base.LogErrorLog("MemberCancelTradingRequestv1-CheckSecondaryPin", secondaryPinErr.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: secondaryPinErr.Error()}, nil)
		return
	}
	// start add this bcz of deicmal problem
	quantityString := helpers.CutOffDecimalv2(form.Quantity, 10, ".", "", true)
	qtyFloat64, err := strconv.ParseFloat(quantityString, 64)
	if err != nil {
		arrErr := map[string]interface{}{
			"form_quantity":  form.Quantity,
			"quantityString": quantityString,
		}
		base.LogErrorLog("MemberCancelTradingRequestv1-ParseFloat_Failed", err.Error(), arrErr, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}
	form.Quantity = qtyFloat64
	// end add this bcz of deicmal problem
	if strings.ToLower(form.TradingAction) == "buy" {
		// begin transaction
		tx := models.Begin()

		arrData := trading_service.CancelMemberTradingBuyRequestStruct{
			BuyID:       form.ID,
			Quantity:    form.Quantity,
			EntMemberID: member.EntMemberID,
		}

		err = trading_service.ProcessMemberCancelTradingBuyRequestv1(tx, arrData)

		if err != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		err = models.Commit(tx)
		if err != nil {
			models.Rollback(tx)
			base.LogErrorLog("MemberCancelTradingRequestv1-Commit Failed", err.Error(), "", true)
			message := app.MsgStruct{
				Msg: "something_went_wrong",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
	} else if strings.ToLower(form.TradingAction) == "sell" {
		// begin transaction
		tx := models.Begin()

		arrData := trading_service.CancelMemberTradingSellRequestStruct{
			SellID:      form.ID,
			Quantity:    form.Quantity,
			EntMemberID: member.EntMemberID,
		}

		err = trading_service.ProcessMemberCancelTradingSellRequestv1(tx, arrData)

		if err != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		err = models.Commit(tx)
		if err != nil {
			models.Rollback(tx)
			base.LogErrorLog("MemberCancelTradingRequestv1-Commit Failed", err.Error(), "", true)
			message := app.MsgStruct{
				Msg: "something_went_wrong",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
	} else {
		base.LogErrorLog("MemberCancelTradingRequestv1-invalid_action_type", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}
