package membership

import (
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/membership_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

// GetMembershipSetup func
func GetMembershipSetup(c *gin.Context) {
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
		module   = "MEMBERSHIP"
	)

	// retrieve products
	packageSetting, errMsg := membership_service.GetMembershipSetup(memberID, langCode)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, "")
		return
	}

	// retrieve payment setting
	paymentSetting, paymentSettingErr := wallet_service.GetPaymentSettingByModule(memberID, module, "", "USDT", langCode, true)
	if paymentSettingErr != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: paymentSettingErr}, "")
		return
	}

	var arrDataReturn = map[string]interface{}{
		"package_setting": packageSetting,
		"payment_setting": paymentSetting,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

// GetMembershipUpdateHistoryListFilter struct
type GetMembershipUpdateHistoryListFilter struct {
	DateFrom string `form:"date_from" json:"date_from"`
	DateTo   string `form:"date_to" json:"date_to"`
	Page     int64  `form:"page" json:"page"`
}

// func GetMembershipUpdateHistoryList function
func GetMembershipUpdateHistoryList(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMembershipUpdateHistoryListFilter
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := membership_service.MembershipUpdateHistoryListFilter{
		MemberID: member.EntMemberID,
		LangCode: langCode,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		Page:     form.Page,
	}

	packageSetting, errMsg := membership_service.GetMembershipUpdateHistoryList(arrData)
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, packageSetting)
}

// PurchaseMembershipForm struct
type PurchaseMembershipForm struct {
	PackageCode  string `form:"package_code" json:"package_code" valid:"Required"`
	Payments     string `form:"payments" json:"payments"`
	PromoCode    string `form:"promo_code" json:"promo_code"`
	PinCode      string `form:"pin_code" json:"pin_code"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin"`
}

// PurchaseMembership function for verification without access token
func PurchaseMembership(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form PurchaseMembershipForm
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
		base.LogErrorLog("productController:PurchaseMembership():RsaDecryptPKCS1v15()", err.Error(), map[string]interface{}{"secondary_pin": form.SecondaryPin}, true)
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
	msgStruct, arrData := membership_service.PurchaseMembership(tx, membership_service.PurchaseMembershipParam{
		MemberID:    entMemberID,
		PackageCode: form.PackageCode,
		Payments:    form.Payments,
		PromoCode:   form.PromoCode,
		PinCode:     form.PinCode,
	}, langCode)

	if msgStruct.Msg != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
		return
	}

	// commit transaction
	err = models.Commit(tx)
	if err != nil {
		base.LogErrorLog("productController:PurchaseMembership()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}
