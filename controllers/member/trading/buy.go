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

// BuyMemberTngForm struct
type BuyMemberTradingForm struct {
	SellID       int     `form:"sell_id" json:"sell_id" valid:"Required;"`
	Quantity     float64 `form:"quantity" json:"quantity" valid:"Required;"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func MemberBuyTradingv1 function
func MemberBuyTradingv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          BuyMemberTradingForm
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

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("MemberBuyTradingv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
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

	arrData := trading_service.BuyMemberTradingStruct{
		SellID:      form.SellID,
		Quantity:    form.Quantity,
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}

	err = trading_service.ProcessMemberBuyTradingv1(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("MemberBuyTradingv1-Commit Failed", err.Error(), "", true)
		message := app.MsgStruct{
			Msg: "buy_trading_failed",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// MemberBuyTradingRequestForm struct
type MemberBuyTradingRequestForm struct {
	CryptoCode   string  `form:"crypto_code" json:"crypto_code"`
	UnitPrice    float64 `form:"unit_price" json:"unit_price"`
	Quantity     float64 `form:"quantity" json:"quantity" valid:"Required;"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
	SigningKey   string  `form:"signing_key" json:"signing_key"`
}

//func MemberBuyTradingRequestv1 function
func MemberBuyTradingRequestv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberBuyTradingRequestForm
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

	if form.Quantity <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "quantity_should_greater_than_0"}, nil)
		return
	}
	if form.UnitPrice <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "unit_price_should_greater_than_0"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("MemberBuyTradingv1-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
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

	arrData := trading_service.BuyMemberTradingRequestStruct{
		UnitPrice:   form.UnitPrice,
		Quantity:    form.Quantity,
		CryptoCode:  form.CryptoCode,
		EntMemberID: member.EntMemberID,
		SigningKey:  form.SigningKey,
	}

	totalAmount, err := trading_service.ProcessMemberBuyTradingRequestv3(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("MemberBuyTradingv1-Commit Failed", err.Error(), "", true)
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

// MemberBuyTradingRequestv2Form struct
type MemberBuyTradingRequestv2Form struct {
	CryptoCode   string  `form:"crypto_code" json:"crypto_code"`
	UnitPrice    float64 `form:"unit_price" json:"unit_price"`
	Quantity     float64 `form:"quantity" json:"quantity" valid:"Required;"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//func MemberBuyTradingRequestv2 function
func MemberBuyTradingRequestv2(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberBuyTradingRequestv2Form
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

	if form.Quantity <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "quantity_should_greater_than_0"}, nil)
		return
	}
	if form.UnitPrice <= 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "unit_price_should_greater_than_0"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("MemberBuyTradingRequestv2-RsaDecryptPKCS1v15_SecondaryPin_Failed", err.Error(), form.SecondaryPin, true)
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

	arrData := trading_service.BuyMemberTradingRequestv2Struct{
		UnitPrice:   form.UnitPrice,
		Quantity:    form.Quantity,
		CryptoCode:  form.CryptoCode,
		EntMemberID: member.EntMemberID,
	}
	totalAmount, err := trading_service.ProcessMemberBuyTradingRequestv4(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("MemberBuyTradingRequestv2-Commit Failed", err.Error(), "", true)
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
