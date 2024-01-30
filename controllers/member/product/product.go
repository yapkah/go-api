package product

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/product_service"
	"github.com/smartblock/gta-api/service/sales_service"
	"github.com/smartblock/gta-api/service/wallet_service"
)

type GetProductsForm struct {
	Type          string `json:"type" form:"type"`
	NftSeriesCode string `json:"nft_series_code" form:"nft_series_code"`
}

// GetProductsv1 func
func GetProductsv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetProductsForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

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
		groupType       = form.Type
		memberID        = member.EntMemberID
		prdCurrencyCode = ""
	)

	// validate input type
	if !helpers.StringInSlice(groupType, []string{"CONTRACT", "BOT"}) {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_type"}, "")
		return
	}

	// retrieve product group setting
	generalSetting, errMsg := product_service.GetProductsGroupSetting(memberID, groupType, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, "")
		return
	}

	// retrieve products
	products, errMsg := product_service.GetProductsv1(memberID, groupType, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, "")
		return
	}

	prdCurrencyCode = generalSetting["currency_code"].(string)

	// retrieve payment setting
	paymentSetting, paymentSettingErr := wallet_service.GetPaymentSettingByModule(memberID, groupType, "", prdCurrencyCode, langCode, true)
	if paymentSettingErr != "" {
		message := app.MsgStruct{
			Msg: paymentSettingErr,
		}
		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	var arrDataReturn = map[string]interface{}{
		"products":        products,
		"general_setting": generalSetting,
		"payment_setting": paymentSetting,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// PurchaseContractForm struct
type PurchaseContractForm struct {
	ContractCode  string  `form:"contract_code" json:"contract_code" valid:"Required"`
	Unit          float64 `form:"unit" json:"unit" valid:"Required"`
	PaymentType   string  `form:"payment_type" json:"payment_type" valid:"Required"`
	Payments      string  `form:"payments" json:"payments" valid:"Required"`
	MachineType   string  `form:"machine_type" json:"machine_type"`     // required if either t_price/p_price status is on
	SecPrice      float64 `form:"sec_price" json:"sec_price"`           // required if either t_price/p_price status is on
	FilecoinPrice float64 `form:"filecoin_price" json:"filecoin_price"` // required if machine type = "FILECOIN"
	ChiaPrice     float64 `form:"chia_price" json:"chia_price"`         // required if machine type = "CHIA"
	SecondaryPin  string  `form:"secondary_pin" json:"secondary_pin"`
}

// PurchaseContract function for verification without access token
func PurchaseContract(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form PurchaseContractForm
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
		base.LogErrorLog("productController:PurchaseContract():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
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

	sourceInterface, _ := c.Get("sourceName")
	sourceName := sourceInterface.(string)

	genTranxDataStatus := false
	if strings.ToLower(sourceName) == "htmlfive" {
		genTranxDataStatus = true
	}

	// perform purchase contract action
	msgStruct, arrData, docNo := product_service.PurchaseContract(tx, product_service.PurchaseContractStruct{
		MemberID:           entMemberID,
		ContractCode:       form.ContractCode,
		Unit:               form.Unit,
		PaymentType:        form.PaymentType,
		MachineType:        form.MachineType,
		SecPrice:           form.SecPrice,
		FilecoinPrice:      form.FilecoinPrice,
		ChiaPrice:          form.ChiaPrice,
		Payments:           form.Payments,
		GenTranxDataStatus: genTranxDataStatus,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("productController:PurchaseContract()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	if docNo != "" {
		// insert income cap
		db := models.GetDB() // no need set begin transaction
		errMsg := sales_service.InsertIncomeCapByDocNo(db, docNo)
		if errMsg != "" {
			base.LogErrorLog("productController:PurchaseContract():InsertIncomeCapByDocNo():1", err.Error(), map[string]interface{}{"docNo": docNo}, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		errMsg = sales_service.InsertNftAirdrop(db, docNo)
		if errMsg != "" {
			base.LogErrorLog("productController:PurchaseContract():InsertNftAirdrop():1", err.Error(), map[string]interface{}{"docNo": docNo}, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		errMsg = sales_service.InsertSlsMasterBnsQueue(db, docNo)
		if errMsg != "" {
			base.LogErrorLog("productController:PurchaseContract():InsertSlsMasterBnsQueue():1", err.Error(), map[string]interface{}{"docNo": docNo}, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}

// PostStakingForm struct
type PostStakingForm struct {
	PrdCode                 string  `form:"product_code" json:"product_code" valid:"Required"`
	Unit                    float64 `form:"unit" json:"unit" valid:"Required"`
	Payments                string  `form:"payments" json:"payments" valid:"Required"`
	ApprovedTransactionData string  `form:"approved_transaction_data" json:"approved_transaction_data"`
	SecondaryPin            string  `form:"secondary_pin" json:"secondary_pin"`
}

// PostStaking function for verification without access token
func PostStaking(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form PostStakingForm
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
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
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

	sourceInterface, _ := c.Get("sourceName")
	sourceName := sourceInterface.(string)

	genTranxDataStatus := false
	if strings.ToLower(sourceName) == "htmlfive" {
		genTranxDataStatus = true
	}

	// perform purchase contract action
	msgStruct, arrData := product_service.PostStaking(tx, product_service.PostStakingStruct{
		MemberID:                entMemberID,
		ProductCode:             form.PrdCode,
		Unit:                    form.Unit,
		Payments:                form.Payments,
		ApprovedTransactionData: form.ApprovedTransactionData,
		GenTranxDataStatus:      genTranxDataStatus,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("productController:PostStaking()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}

// PostUnstakeForm struct
type PostUnstakeForm struct {
	DocNo                  string  `form:"doc_no" json:"doc_no" valid:"Required"`
	Amount                 float64 `form:"amount" json:"amount" valid:"Required"`
	UnstakeTransactionData string  `form:"unstake_transaction_data" json:"unstake_transaction_data" valid:"Required"`
	SecondaryPin           string  `form:"secondary_pin" json:"secondary_pin"`
}

// PostUnstake function for verification without access token
func PostUnstake(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form PostUnstakeForm
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
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
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

	// perform purchase contract action
	msgStruct, arrData := product_service.PostUnstake(tx, product_service.PostUnstakeStruct{
		MemberID:               entMemberID,
		DocNo:                  form.DocNo,
		Amount:                 form.Amount,
		UnstakeTransactionData: form.UnstakeTransactionData,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("productController:PostUnstake()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}

// TopupContractForm struct
type TopupContractForm struct {
	DocNo        string  `form:"doc_no" json:"doc_no" valid:"Required"`
	Amount       float64 `form:"amount" json:"amount" valid:"Required"`
	PaymentType  string  `form:"payment_type" json:"payment_type" valid:"Required"`
	Payments     string  `form:"payments" json:"payments" valid:"Required"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin"`
}

// TopupContract function for verification without access token
func TopupContract(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form TopupContractForm
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
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
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
	sourceInterface, _ := c.Get("sourceName")
	sourceName := sourceInterface.(string)

	genTranxDataStatus := false
	if strings.ToLower(sourceName) == "htmlfive" {
		genTranxDataStatus = true
	}

	// perform purchase contract action
	msgStruct, arrData := product_service.TopupContract(tx, product_service.TopupContractStruct{
		MemberID:           entMemberID,
		DocNo:              form.DocNo,
		Amount:             form.Amount,
		PaymentType:        form.PaymentType,
		Payments:           form.Payments,
		GenTranxDataStatus: genTranxDataStatus,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("productController:TopupContract()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}

// TopupMiningNodeForm struct
type TopupMiningNodeForm struct {
	NodeID       int    `form:"node_id" json:"node_id" valid:"Required"`
	ContractCode string `form:"contract_code" json:"contract_code" valid:"Required"`
	PaymentType  string `form:"payment_type" json:"payment_type" valid:"Required"`
	Payments     string `form:"payments" json:"payments" valid:"Required"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin"`
}

// TopupMiningNode function for verification without access token
func TopupMiningNode(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form TopupMiningNodeForm
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
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
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
	sourceInterface, _ := c.Get("sourceName")
	sourceName := sourceInterface.(string)

	genTranxDataStatus := false
	if strings.ToLower(sourceName) == "htmlfive" {
		genTranxDataStatus = true
	}

	// perform purchase contract action
	msgStruct, arrData := product_service.TopupMiningNode(tx, product_service.TopupMiningNodeStruct{
		MemberID:           entMemberID,
		NodeID:             form.NodeID,
		ContractCode:       form.ContractCode,
		PaymentType:        form.PaymentType,
		Payments:           form.Payments,
		GenTranxDataStatus: genTranxDataStatus,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("productController:TopupMiningNode()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}

// GetNftSeries func
func GetNftSeries(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

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

	var memberID = member.EntMemberID

	// retrieve product group setting
	arrDataReturn, errMsg := product_service.GetNftSeries(memberID, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// GetSubscriptionCancellationSetup func
func GetSubscriptionCancellationSetup(c *gin.Context) {
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

	var memberID = member.EntMemberID

	// retrieve subscription cancellation setup
	arrDataReturn := product_service.GetSubscriptionCancellationSetup(memberID, langCode)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// PostSubscriptionCancellationForm struct
type PostSubscriptionCancellationForm struct {
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin"`
}

// PostSubscriptionCancellation function for verification without access token
func PostSubscriptionCancellation(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form PostSubscriptionCancellationForm
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
		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
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

	// perform purchase contract action
	msgStruct, arrData := product_service.PostSubscriptionCancellation(tx, product_service.PostSubscriptionCancellationStruct{
		MemberID: entMemberID,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		models.ErrorLog("productController:PostSubscriptionCancellation()", "Commit():1", err.Error())
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}
