package trading

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"

	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/trading_service"
)

// SellMemberTngForm struct
type SellMemberTradingForm struct {
	BuyID        int     `form:"buy_id" json:"buy_id" valid:"Required;"`
	Quantity     float64 `form:"quantity" json:"quantity" valid:"Required;"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func MemberSellTradingv1 function
func MemberSellTradingv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          SellMemberTradingForm
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

	if form.Quantity <= 0 {
		msgParams := map[string]string{}
		msgParams["q"] = "0"
		msg := app.MsgStruct{
			Msg:    "please_enter_more_than_:q_quantity",
			Params: msgParams,
		}
		appG.ResponseV2(0, http.StatusOK, msg, nil)
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
		base.LogErrorLog("MemberSellTradingv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
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
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: secondaryPinErr.Error()}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	arrData := trading_service.SellMemberTradingStruct{
		BuyID:    form.BuyID,
		Quantity: form.Quantity,
		MemberID: member.EntMemberID,
		LangCode: langCode,
	}

	err = trading_service.ProcessMemberSellTradingv1(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		msg := app.MsgStruct{
			Msg: err.Error(),
		}
		appG.ResponseV2(0, http.StatusOK, msg, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("MemberSellTradingv1-Commit Failed", err.Error(), "", true)
		message := app.MsgStruct{
			Msg: "sell_trading_failed",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// MemberSellTradingRequestForm struct
type MemberSellTradingRequestForm struct {
	CryptoCode   string  `form:"crypto_code" json:"crypto_code" valid:"Required;"`
	UnitPrice    float64 `form:"unit_price" json:"unit_price" valid:"Required;"`
	Quantity     float64 `form:"quantity" json:"quantity" valid:"Required;"`
	SigningKey   string  `form:"signing_key" json:"signing_key"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func MemberSellTradingRequestv1 function
func MemberSellTradingRequestv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberSellTradingRequestForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
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
		base.LogErrorLog("MemberSellTradingv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
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
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: secondaryPinErr.Error()}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	arrData := trading_service.SellMemberTradingRequestStruct{
		UnitPrice:   form.UnitPrice,
		Quantity:    form.Quantity,
		CryptoCode:  form.CryptoCode,
		EntMemberID: member.EntMemberID,
		SigningKey:  form.SigningKey,
	}

	totalAmount, err := trading_service.ProcessMemberSellTradingRequestv3(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("MemberSellTradingv1-Commit Failed", err.Error(), "", true)
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

// MemberSellTradingRequestv2Form struct
type MemberSellTradingRequestv2Form struct {
	CryptoCode   string  `form:"crypto_code" json:"crypto_code" valid:"Required;"`
	UnitPrice    float64 `form:"unit_price" json:"unit_price" valid:"Required;"`
	Quantity     float64 `form:"quantity" json:"quantity" valid:"Required;"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func MemberSellTradingRequestv2 function
func MemberSellTradingRequestv2(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberSellTradingRequestv2Form
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
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
		base.LogErrorLog("MemberSellTradingv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
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
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: secondaryPinErr.Error()}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()

	arrData := trading_service.SellMemberTradingRequestv2Struct{
		UnitPrice:   form.UnitPrice,
		Quantity:    form.Quantity,
		CryptoCode:  form.CryptoCode,
		EntMemberID: member.EntMemberID,
	}
	totalAmount, err := trading_service.ProcessMemberSellTradingRequestv4(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("MemberSellTradingv1-Commit Failed", err.Error(), "", true)
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
