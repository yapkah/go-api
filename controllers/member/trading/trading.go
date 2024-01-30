package trading

import (
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/trading_service"
	"github.com/smartblock/gta-api/service/wallet_service"
)

// GetMemberTradingStatusForm struct
type GetMemberTradingStatusForm struct {
	Module string `form:"module" json:"module"`
}

// GetMemberTradingStatus func
func GetMemberTradingStatus(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberTradingStatusForm
	)

	app.BindAndValid(c, &form)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	var (
		member          = u.(*models.EntMemberMembers)
		memID           = member.EntMemberID
		username        = member.NickName
		module   string = form.Module
		langCode string = ""
		arrRst          = map[string]interface{}{}
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// validate module
	if module != "" {
		if !helpers.StringInSlice(module, []string{"TNC", "API", "MEMBERSHIP", "DEPOSIT", "LIMIT", "WITHDRAWAL"}) {
			appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_module"}, nil)
			return
		}
	}

	// retrieve strategy list
	if module == "" || module == "STRATEGY" {
		memberTradingStrategy, errMsg := trading_service.GetMemberCurrentTradingStrategyList(memID, username, langCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		arrRst["strategies"] = memberTradingStrategy
	}

	// retrieve tnc status
	if module == "" || module == "TNC" {
		memberTradingTnc, errMsg := trading_service.GetMemberCurrentTradingTncStatus(memID, langCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		arrRst["tnc_url"] = memberTradingTnc.TncUrl
		arrRst["tnc_status"] = memberTradingTnc.Status
	}

	// retrieve api status
	if module == "" || module == "API" {
		memberTradingStatus, errMsg := trading_service.GetMemberCurrentTradingApiStatus(memID, langCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		arrRst["api_management_popup_reminder"] = memberTradingStatus.PopupReminder
		arrRst["api_management_status"] = memberTradingStatus.Status
		arrRst["api_management_reset_status"] = memberTradingStatus.ResetStatus
		arrRst["api_management"] = memberTradingStatus.MemberTradingApi

		tradingApiPlatform, errMsg := trading_service.GetTradingApiPlatform(langCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		arrRst["api_platform_options"] = tradingApiPlatform
	}

	// retrieve membership status
	if module == "" || module == "MEMBERSHIP" {
		tradingMembership, errMsg := trading_service.GetMemberMembershipStatus(memID, langCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		arrRst["membership_status"] = tradingMembership.Status
		arrRst["membership_expiry_date"] = tradingMembership.ExpiryDate
		arrRst["membership_renew_status"] = tradingMembership.RenewStatus
		arrRst["membership_expiring"] = tradingMembership.MembershipExpiring
	}

	// retrieve deposit status
	if module == "" || module == "DEPOSIT" {
		tradingDeposit, errMsg := trading_service.GetMemberDepositStatus(memID, langCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		arrRst["deposit_status"] = tradingDeposit.Status
		arrRst["deposit_min"] = tradingDeposit.DepositMin
		arrRst["deposit_max"] = tradingDeposit.DepositMax
		arrRst["deposit_options"] = tradingDeposit.DepositOptions
		arrRst["current_deposit_low"] = tradingDeposit.DepositLow
		arrRst["current_deposit_low_and_with_bot"] = tradingDeposit.DepositLowAndWithBot
		arrRst["current_deposit_amount"] = helpers.CutOffDecimal(tradingDeposit.CurrentDepositAmount, 2, ".", ",")
	}

	// retrieve wallet limit status
	if module == "" || module == "LIMIT" {
		tradingLimit, errMsg := trading_service.GetMemberTradingLimitStatus(memID, langCode)
		if errMsg != "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
			return
		}

		arrRst["wallet_limit_status"] = tradingLimit.Status
		arrRst["wallet_limit_min"] = tradingLimit.SpotWalletLimitMin // will be removed
		arrRst["wallet_limit_max"] = tradingLimit.SpotWalletLimitMax // will be removed
		arrRst["spot_wallet_limit_min"] = tradingLimit.SpotWalletLimitMin
		arrRst["spot_wallet_limit_max"] = tradingLimit.SpotWalletLimitMax
		arrRst["future_wallet_limit_min"] = tradingLimit.FutureWalletLimitMin
		arrRst["future_wallet_limit_max"] = tradingLimit.FutureWalletLimitMax
		arrRst["wallet_limit_options"] = tradingLimit.WalletLimitOptions
		arrRst["current_wallet_limit"] = helpers.CutOffDecimal(tradingLimit.CurrentSpotWalletLimit, 2, ".", ",") // will be removed
		arrRst["current_spot_wallet_limit"] = helpers.CutOffDecimal(tradingLimit.CurrentSpotWalletLimit, 2, ".", ",")
		arrRst["current_future_wallet_limit"] = helpers.CutOffDecimal(tradingLimit.CurrentFutureWalletLimit, 2, ".", ",")
	}

	// retrieve deposit withdrawal status
	if module == "" || module == "WITHDRAWAL" {
		arrRst["deposit_withdrawal_min"] = 100
		arrRst["deposit_withdrawal_max"] = 100000
		arrRst["deposit_withdrawal_options"] = []map[string]interface{}{
			{
				"code": "USDT",
				"name": helpers.TranslateV2("usdt_wallet", langCode, map[string]string{}),
			},
			// {
			// 	"code": "USDC",
			// 	"name": helpers.TranslateV2("usdc_wallet", langCode, map[string]string{}),
			// },
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrRst)
}

// MemberTradingApi struct
type MemberTradingApi struct {
	PlatformCode string `form:"platform_code" json:"platform_code" valid:"Required"`
	ApiKey       string `form:"api_key" json:"api_key" valid:"Required"`
	Secret       string `form:"secret" json:"secret" valid:"Required"`
	Passphrase   string `form:"passphrase" json:"passphrase"` // required if platformCode = "KC"
	// Module       string `form:"module" json:"module"`         // required if platformCode = "KC"
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// UpdateMemberTradingApi
func UpdateMemberTradingApi(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form MemberTradingApi
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	platform := form.PlatformCode
	memberTradingApiDetails := []trading_service.UpdateMemberTradingApiDetails{}

	if platform == "BN" {
		memberTradingApiDetails = append(memberTradingApiDetails,
			trading_service.UpdateMemberTradingApiDetails{
				Module: "FUTURE",
				ApiKey: form.ApiKey,
				Secret: form.Secret,
			},
			trading_service.UpdateMemberTradingApiDetails{
				Module: "SPOT",
				ApiKey: form.ApiKey,
				Secret: form.Secret,
			},
		)
	} else if platform == "KC" {
		// validate passphrase, apikey2, secret2, passphrase2
		if form.Passphrase == "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "passphrase_is_required"}, nil)
			return
		}

		// if !helpers.StringInSlice(form.Module, []string{"FUTURE", "SPOT"}) {
		// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_module"}, nil)
		// 	return
		// }

		memberTradingApiDetails = append(memberTradingApiDetails,
			trading_service.UpdateMemberTradingApiDetails{
				Module:     "FUTURE",
				ApiKey:     form.ApiKey,
				Secret:     form.Secret,
				Passphrase: form.Passphrase,
			},
			trading_service.UpdateMemberTradingApiDetails{
				Module:     "SPOT",
				ApiKey:     form.ApiKey,
				Secret:     form.Secret,
				Passphrase: form.Passphrase,
			},
		)
	}

	// being transaction
	tx := models.Begin()

	// update trading api
	memberTradingApi := trading_service.UpdateMemberTradingApiParam{
		MemberID:                memberID,
		PlatformCode:            form.PlatformCode,
		MemberTradingApiDetails: memberTradingApiDetails,
	}
	errMsg := trading_service.UpdateMemberTradingApi(tx, memberTradingApi)
	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// MemberTradingTnc struct
type MemberTradingTnc struct {
	Signature string `form:"signature" json:"signature" valid:"Required"` // img file
}

// UpdateMemberTradingTnc
func UpdateMemberTradingTnc(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		// form MemberTradingTnc
		err error
	)

	// ok, msg := app.BindAndValid(c, &form)
	// if ok == false {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
	// 	return
	// }

	file, header, err := c.Request.FormFile("signature")
	if err != nil {
		message := app.MsgStruct{
			Msg: "please_upload_signature",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	// being transaction
	tx := models.Begin()

	memberTradingTnc := trading_service.MemberTradingTnc{
		MemberID:        memberID,
		SignatureFile:   file,
		SignatureHeader: header,
		LangCode:        langCode,
	}
	errMsg := memberTradingTnc.UpdateMemberTradingTnc(tx)

	if errMsg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingTnc()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// GetTradingDepositSetup func
func GetTradingDepositSetup(c *gin.Context) {
	var appG = app.Gin{C: c}

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	// retrieve langCode
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	var (
		memberID = member.EntMemberID
		module   = "TRADING_DEPOSIT"
	)

	// retrieve payment setting
	paymentSetting, paymentSettingErr := wallet_service.GetPaymentSettingByModule(memberID, module, "", "USDT", langCode, true)
	if paymentSettingErr != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: paymentSettingErr}, "")
		return
	}

	var arrDataReturn = map[string]interface{}{
		"payment_setting": paymentSetting,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// AddTradingDepositForm struct
type AddTradingDepositForm struct {
	Amount       float64 `form:"amount" json:"amount" valid:"Required"`
	Payments     string  `form:"payments" json:"payments" valid:"Required"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin"`
}

// AddTradingDeposit function for verification without access token
func AddTradingDeposit(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AddTradingDepositForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)

	entMemberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:AddTradingDeposit():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	// perform add trading deposit
	msgStruct, arrData := trading_service.AddTradingDeposit(tx, trading_service.AddTradingDepositParam{
		MemberID: entMemberID,
		Amount:   form.Amount,
		Payments: form.Payments,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:AddTradingDeposit()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}

// WithdrawTradingDepositForm struct
type WithdrawTradingDepositForm struct {
	Amount          float64 `form:"amount" json:"amount" valid:"Required"`
	EwalletTypeCode string  `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required"`
	SecondaryPin    string  `form:"secondary_pin" json:"secondary_pin"`
}

// WithdrawTradingDeposit function for verification without access token
func WithdrawTradingDeposit(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form WithdrawTradingDepositForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)

	entMemberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:WithdrawTradingDeposit():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	// perform withdraw trading deposit
	msgStruct := trading_service.WithdrawTradingDeposit(tx, trading_service.WithdrawTradingDepositParam{
		MemberID:        entMemberID,
		Amount:          form.Amount,
		EwalletTypeCode: form.EwalletTypeCode,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:WithdrawTradingDeposit()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// UpdateTradingWalletLimitForm struct
type UpdateTradingWalletLimitForm struct {
	Module       string  `form:"module" json:"module" valid:"Required"` // SPOT/FUTURE
	Amount       float64 `form:"amount" json:"amount" valid:"Required"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// UpdateTradingWalletLimit function for verification without access token
func UpdateTradingWalletLimit(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form UpdateTradingWalletLimitForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)

	entMemberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateTradingWalletLimit():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	// perform purchase membership action
	msgStruct := trading_service.UpdateTradingWalletLimit(tx, trading_service.UpdateTradingWalletLimitParam{
		MemberID: entMemberID,
		Module:   form.Module,
		Amount:   form.Amount,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateTradingWalletLimit()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// GetMemberAutoTradingSetupForm struct
type GetMemberAutoTradingSetupForm struct {
	StrategyCode string `form:"strategy_code" json:"strategy_code" valid:"Required"` //prd_master.code
}

// GetMemberAutoTradingSetup func
func GetMemberAutoTradingSetup(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingSetupForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member              = u.(*models.EntMemberMembers)
		memID               = member.EntMemberID
		strategyCode        = form.StrategyCode
		langCode     string = ""
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingSetup, errMsg := trading_service.GetMemberAutoTradingSetup(memID, strategyCode, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingSetup)
}

// GetMemberAutoTradingFundingRateForm struct
type GetMemberAutoTradingFundingRateForm struct {
	CryptoPair string `form:"crypto_pair" json:"crypto_pair" valid:"Required"` //prd_master.code
}

// GetMemberAutoTradingFundingRate func
func GetMemberAutoTradingFundingRate(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingFundingRateForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member            = u.(*models.EntMemberMembers)
		memID             = member.EntMemberID
		cryptoPair        = form.CryptoPair
		langCode   string = ""
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingFundingRate, errMsg := trading_service.GetMemberAutoTradingFundingRate(memID, cryptoPair, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingFundingRate)
}

// GetMemberAutoTradingGridDataForm struct
type GetMemberAutoTradingGridDataForm struct {
	Mode       string `form:"mode" json:"mode" valid:"Required"`               // ARITHMETIC/GEOMETRIC
	CryptoPair string `form:"crypto_pair" json:"crypto_pair" valid:"Required"` //prd_master.code
}

// GetMemberAutoTradingGridData func
func GetMemberAutoTradingGridData(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingGridDataForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member            = u.(*models.EntMemberMembers)
		memID             = member.EntMemberID
		cryptoPair        = form.CryptoPair
		mode              = form.Mode
		langCode   string = ""
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingGridData, errMsg := trading_service.GetMemberAutoTradingGridData(memID, cryptoPair, mode, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingGridData)
}

// GetMemberAutoTradingMartingaleDataForm struct
type GetMemberAutoTradingMartingaleDataForm struct {
	CryptoPair string `form:"crypto_pair" json:"crypto_pair" valid:"Required"` //prd_master.code
}

// GetMemberAutoTradingMartingaleData func
func GetMemberAutoTradingMartingaleData(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingMartingaleDataForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member            = u.(*models.EntMemberMembers)
		memID             = member.EntMemberID
		cryptoPair        = form.CryptoPair
		langCode   string = ""
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingMartingaleData, errMsg := trading_service.GetMemberAutoTradingMartingaleData(memID, cryptoPair, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingMartingaleData)
}

// GetMemberAutoTradingReverseMartingaleDataForm struct
type GetMemberAutoTradingReverseMartingaleDataForm struct {
	CryptoPair string `form:"crypto_pair" json:"crypto_pair" valid:"Required"` //prd_master.code
}

// GetMemberAutoTradingReverseMartingaleData func
func GetMemberAutoTradingReverseMartingaleData(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingReverseMartingaleDataForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member            = u.(*models.EntMemberMembers)
		memID             = member.EntMemberID
		cryptoPair        = form.CryptoPair
		langCode   string = ""
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingMartingaleData, errMsg := trading_service.GetMemberAutoTradingReverseMartingaleData(memID, cryptoPair, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingMartingaleData)
}

// MemberAutoTradingCFRA struct
type MemberAutoTradingCFRA struct {
	Type         string  `form:"type" json:"type" valid:"Required"` // AI/PROF
	Amount       float64 `form:"amount" json:"amount" valid:"Required"`
	CryptoPair   string  `form:"crypto_pair" json:"crypto_pair"` // only required if is PROF setting
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// AddMemberAutoTradingCFRA
func AddMemberAutoTradingCFRA(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form MemberAutoTradingCFRA
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	// validate on api management, tnc signature, membership, deposit, wallet_limit
	errMsg := trading_service.ValidateMemberAutoTradingStatus(memberID)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// validate secondary pin
	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	memberAutoTrading := trading_service.MemberAutoTradingCFRA{
		MemberID:   memberID,
		Type:       form.Type,
		Amount:     form.Amount,
		CryptoPair: form.CryptoPair,
		LangCode:   langCode,
	}
	msgStruct := trading_service.AddMemberAutoTradingCFRA(tx, memberAutoTrading)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:AddMemberAutoTradingCFRA()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// MemberAutoTradingCIFRA struct
type MemberAutoTradingCIFRA struct {
	Amount       float64 `form:"amount" json:"amount" valid:"Required"`
	CryptoPair   string  `form:"crypto_pair" json:"crypto_pair" valid:"Required"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// AddMemberAutoTradingCIFRA
func AddMemberAutoTradingCIFRA(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form MemberAutoTradingCIFRA
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	// validate on api management, tnc signature, membership, deposit, wallet_limit
	errMsg := trading_service.ValidateMemberAutoTradingStatus(memberID)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// validate secondary pin
	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	memberAutoTrading := trading_service.MemberAutoTradingCIFRA{
		MemberID:   memberID,
		Type:       "AI",
		Amount:     form.Amount,
		CryptoPair: form.CryptoPair,
		LangCode:   langCode,
	}
	msgStruct := trading_service.AddMemberAutoTradingCIFRA(tx, memberAutoTrading)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:AddMemberAutoTradingCIFRA()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// MemberAutoTradingSGT struct
type MemberAutoTradingSGT struct {
	Type                     string  `form:"type" json:"type" valid:"Required"` // AI/PROF
	Amount                   float64 `form:"amount" json:"amount" valid:"Required"`
	CryptoPair               string  `form:"crypto_pair" json:"crypto_pair" valid:"Required"`
	UpperPrice               float64 `form:"upper_price" json:"upper_price"`                                 // required if type = PROF
	LowerPrice               float64 `form:"lower_price" json:"lower_price"`                                 // required if type = PROF
	CryptoPricePercentage    float64 `form:"crypto_price_percentage" json:"crypto_price_percentage"`         // required if type = PROF | mode = GEOMETRIC only
	CalCryptoPricePercentage float64 `form:"cal_crypto_price_percentage" json:"cal_crypto_price_percentage"` // required if type = PROF | mode = ARITHMETIC only
	Mode                     string  `form:"mode" json:"mode"`                                               // required if type = PROF | ARITHMETIC/GEOMETRIC
	SecondaryPin             string  `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// AddMemberAutoTradingSGT
func AddMemberAutoTradingSGT(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form MemberAutoTradingSGT
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	// validate on api management, tnc signature, membership, deposit, wallet_limit
	errMsg := trading_service.ValidateMemberAutoTradingStatus(memberID)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// validate secondary pin
	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	memberAutoTrading := trading_service.MemberAutoTradingSGT{
		MemberID:                 memberID,
		Type:                     form.Type,
		Amount:                   form.Amount,
		CryptoPair:               form.CryptoPair,
		LowerPrice:               form.LowerPrice,
		UpperPrice:               form.UpperPrice,
		CryptoPricePercentage:    form.CryptoPricePercentage,
		CalCryptoPricePercentage: form.CalCryptoPricePercentage,
		Mode:                     form.Mode,
		LangCode:                 langCode,
	}
	msgStruct := trading_service.AddMemberAutoTradingSGT(tx, memberAutoTrading)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:AddMemberAutoTradingSGT()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// MemberAutoTradingMT struct
type MemberAutoTradingMT struct {
	Type               string  `form:"type" json:"type" valid:"Required"` // AI/PROF
	Amount             float64 `form:"amount" json:"amount" valid:"Required"`
	CryptoPair         string  `form:"crypto_pair" json:"crypto_pair" valid:"Required"`
	FirstOrderAmount   float64 `form:"first_order_amount" json:"first_order_amount" valid:"Required"`
	FirstOrderPrice    float64 `form:"first_order_price" json:"first_order_price"`       // required if type = PROF
	PriceScale         float64 `form:"price_scale" json:"price_scale"`                   // required if type = PROF
	TakeProfitCallback float64 `form:"take_profit_callback" json:"take_profit_callback"` // required if type = PROF
	TakeProfit         float64 `form:"take_profit" json:"take_profit"`                   // required if type = PROF
	// AddShares          float64 `form:"add_shares" json:"add_shares"`                     // required if type = PROF
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// AddMemberAutoTradingMT
func AddMemberAutoTradingMT(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form MemberAutoTradingMT
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	// validate on api management, tnc signature, membership, deposit, wallet_limit
	errMsg := trading_service.ValidateMemberAutoTradingStatus(memberID)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// validate secondary pin
	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	memberAutoTrading := trading_service.MemberAutoTradingMT{
		MemberID:           memberID,
		Type:               form.Type,
		Amount:             form.Amount,
		CryptoPair:         form.CryptoPair,
		FirstOrderAmount:   form.FirstOrderAmount,
		FirstOrderPrice:    form.FirstOrderPrice,
		PriceScale:         form.PriceScale,
		TakeProfitCallback: form.TakeProfitCallback,
		TakeProfit:         form.TakeProfit,
		// AddShares:          form.AddShares,
		LangCode: langCode,
	}
	msgStruct := trading_service.AddMemberAutoTradingMT(tx, memberAutoTrading)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:AddMemberAutoTradingMT()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// MemberAutoTradingMTD struct
type MemberAutoTradingMTD struct {
	Type               string  `form:"type" json:"type" valid:"Required"` // AI/PROF
	Amount             float64 `form:"amount" json:"amount" valid:"Required"`
	CryptoPair         string  `form:"crypto_pair" json:"crypto_pair" valid:"Required"`
	FirstOrderAmount   float64 `form:"first_order_amount" json:"first_order_amount" valid:"Required"`
	FirstOrderPrice    float64 `form:"first_order_price" json:"first_order_price"`       // required if type = PROF
	PriceScale         float64 `form:"price_scale" json:"price_scale"`                   // required if type = PROF
	TakeProfitCallback float64 `form:"take_profit_callback" json:"take_profit_callback"` // required if type = PROF
	TakeProfit         float64 `form:"take_profit" json:"take_profit"`                   // required if type = PROF
	// AddShares          float64 `form:"add_shares" json:"add_shares"`                     // required if type = PROF
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// AddMemberAutoTradingMTD
func AddMemberAutoTradingMTD(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form MemberAutoTradingMTD
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	// validate on api management, tnc signature, membership, deposit, wallet_limit
	errMsg := trading_service.ValidateMemberAutoTradingStatus(memberID)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	// validate secondary pin
	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	memberAutoTrading := trading_service.MemberAutoTradingMTD{
		MemberID:           memberID,
		Type:               form.Type,
		Amount:             form.Amount,
		CryptoPair:         form.CryptoPair,
		FirstOrderAmount:   form.FirstOrderAmount,
		FirstOrderPrice:    form.FirstOrderPrice,
		PriceScale:         form.PriceScale,
		TakeProfitCallback: form.TakeProfitCallback,
		TakeProfit:         form.TakeProfit,
		// AddShares:          form.AddShares,
		LangCode: langCode,
	}
	msgStruct := trading_service.AddMemberAutoTradingMTD(tx, memberAutoTrading)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:AddMemberAutoTradingMTD()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// AutoTradingLiquidationParam struct
type AutoTradingLiquidationParam struct {
	DocNo        string `form:"doc_no" json:"doc_no"`
	StrategyCode string `form:"strategy_code" json:"strategy_code" valid:"Required"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required"`
}

// PostAutoTradingLiquidation
func PostAutoTradingLiquidation(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AutoTradingLiquidationParam
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// get member info from access token
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	memberID := member.EntMemberID

	// validate secondary pin
	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("tradingController:UpdateMemberTradingApi():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
		return
	}

	wordCount := utf8.RuneCountInString(decryptedText)
	if wordCount < 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	if wordCount > 6 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
		return
	}
	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              memberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// being transaction
	tx := models.Begin()

	memberAutoTradingLiquidation := trading_service.AutoTradingLiquidationParam{
		MemberID:     memberID,
		DocNo:        form.DocNo,
		StrategyCode: form.StrategyCode,
		LangCode:     langCode,
	}
	msgStruct := trading_service.PostAutoTradingLiquidation(tx, memberAutoTradingLiquidation)
	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("tradingController:PostAutoTradingLiquidation()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

// GetMemberAutoTradingTransactionForm struct
type GetMemberAutoTradingTransactionForm struct {
	Strategy   string `form:"strategy" json:"strategy"`
	CryptoPair string `form:"crypto_pair" json:"crypto_pair"`
	DateFrom   string `form:"date_from" json:"date_from"`
	DateTo     string `form:"date_to" json:"date_to"`
	Page       int64  `form:"page" json:"page"`
}

// GetMemberAutoTradingTransaction func
func GetMemberAutoTradingTransaction(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingTransactionForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member          = u.(*models.EntMemberMembers)
		memID           = member.EntMemberID
		langCode string = ""
		page     int64  = 1
	)

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

	var getMemberAutoTradingTransactionParam = trading_service.GetMemberAutoTradingTransactionParam{
		Strategy:   form.Strategy,
		CryptoPair: form.CryptoPair,
		DateFrom:   form.DateFrom,
		DateTo:     form.DateTo,
		Page:       page,
	}

	memberAutoTradingTransaction, errMsg := trading_service.GetMemberAutoTradingTransaction(memID, getMemberAutoTradingTransactionParam, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingTransaction)
}

// GetMemberAutoTradingProfitForm struct
type GetMemberAutoTradingProfitForm struct {
	Type             string `form:"type" json:"type" valid:"Required"`
	Strategy         string `form:"strategy" json:"strategy"`
	CryptoPair       string `form:"crypto_pair" json:"crypto_pair"`
	DateFrom         string `form:"date_from" json:"date_from"`
	DateTo           string `form:"date_to" json:"date_to"`
	DownlineUsername string `form:"downline_username" json:"downline_username"`
	Page             int64  `form:"page" json:"page"`
}

// GetMemberAutoTradingProfit func
func GetMemberAutoTradingProfit(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingProfitForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if !ok {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member          = u.(*models.EntMemberMembers)
		memID           = member.EntMemberID
		dataType        = form.Type
		strategy        = form.Strategy
		page     int64  = 1
		langCode string = ""
	)

	// validate input page number
	if form.Page != 0 {
		page = form.Page
	}

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingProfit, errMsg := trading_service.GetMemberAutoTradingProfit(memID, dataType, strategy, form.CryptoPair, form.DateFrom, form.DateTo, form.DownlineUsername, page, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingProfit)
}

// GetMemberAutoTradingProfitGraphForm struct
type GetMemberAutoTradingProfitGraphForm struct {
	Type     string `form:"type" json:"type" valid:"Required"`
	DataType string `form:"data_type" json:"data_type" valid:"Required"`
}

// GetMemberAutoTradingProfitGraph func
func GetMemberAutoTradingProfitGraph(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingProfitGraphForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member          = u.(*models.EntMemberMembers)
		memID           = member.EntMemberID
		langCode string = ""
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingProfitGraph, errMsg := trading_service.GetMemberAutoTradingProfitGraph(memID, form.Type, form.DataType, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingProfitGraph)
}

// GetMemberAutoTradingReportsForm struct
type GetMemberAutoTradingReportsForm struct {
	PoolType string `form:"pool_type" json:"pool_type"`
	Strategy string `form:"strategy" json:"strategy"`
	DateFrom string `form:"date_from" json:"date_from"`
	DateTo   string `form:"date_to" json:"date_to"`
	Page     int64  `form:"page" json:"page"`
}

// GetMemberAutoTradingReports func
func GetMemberAutoTradingReports(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingReportsForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member          = u.(*models.EntMemberMembers)
		memID           = member.EntMemberID
		langCode string = ""
		page     int64  = 1
	)

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

	var getMemberAutoTradingReportsParam = trading_service.GetMemberAutoTradingReportsParam{
		PoolType: form.PoolType,
		Strategy: form.Strategy,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		Page:     page,
	}

	memberAutoTradingTransaction, errMsg := trading_service.GetMemberAutoTradingReports(memID, getMemberAutoTradingReportsParam, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingTransaction)
}

// GetMemberAutoTradingLogsForm struct
type GetMemberAutoTradingLogsForm struct {
	Page int64 `form:"page" json:"page"`
}

// GetMemberAutoTradingLogs func
func GetMemberAutoTradingLogs(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingLogsForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member          = u.(*models.EntMemberMembers)
		memID           = member.EntMemberID
		langCode string = ""
		page     int64  = 1
	)

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

	var getMemberAutoTradingLogsParam = trading_service.GetMemberAutoTradingLogsParam{
		Page: page,
	}

	memberAutoTradingTransaction, errMsg := trading_service.GetMemberAutoTradingLogs(memID, getMemberAutoTradingLogsParam, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingTransaction)
}

// GetMemberAutoTradingSafetyOrdersForm struct
type GetMemberAutoTradingSafetyOrdersForm struct {
	FirstOrderAmount float64 `form:"first_order_amount" json:"first_order_amount" valid:"Required"`
	Amount           float64 `form:"amount" json:"amount" valid:"Required"`
}

// GetMemberAutoTradingSafetyOrders func
func GetMemberAutoTradingSafetyOrders(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberAutoTradingSafetyOrdersForm
	)

	// validate access user
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	// validate param
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var (
		member                  = u.(*models.EntMemberMembers)
		memID                   = member.EntMemberID
		firstOrderAmount        = form.FirstOrderAmount
		amount                  = form.Amount
		langCode         string = ""
	)

	// retrieve langCode
	langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	memberAutoTradingMartingaleData, errMsg := trading_service.GetMemberAutoTradingSafetyOrders(memID, firstOrderAmount, amount, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, memberAutoTradingMartingaleData)
}
