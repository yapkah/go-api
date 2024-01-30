package member

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/otp_service"
	"github.com/yapkah/go-api/service/trading_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

// AddWalletForm struct
// type AddWalletForm struct {
// 	WalletType string `gorm:"column(wallet_type)" json:"wallet_type" valid:"Required;"`
// 	//Email 			string 		`gorm:"column(Email)" json:"Email" valid:"Required;"`
// 	TransType string `gorm:"column(trans_type)" json:"trans_type" valid:"Required;"`
// 	Amount    string `gorm:"column(amount)" json:"amount" valid:"Required;"`
// }

// AddWithdrawalForm struct
// type AddWithdrawalForm struct {
// 	//WalletType string `gorm:"column(wallet_type)" json:"wallet_type" valid:"Required;"`
// 	//Email 			string 		`gorm:"column(email)" json:"email" valid:"Required;"`
// 	DocNo string `gorm:"column(doc_no)" json:"doc_no" valid:"Required;MaxSize(20);"`
// 	//Amount float64 `gorm:"column(amount)" json:"amount" valid:"Required;"`
// }

// AddDebitForm struct
// type AddDebitForm struct {
// 	WalletType string `gorm:"column(wallet_type)" json:"wallet_type" valid:"Required;"`
// 	//TransType  string `gorm:"column(trans_type)" json:"trans_type" valid:"Required;"`
// 	WithdrawAddr      string `json:"withdraw_addr" valid:"Required;"`
// 	SecondaryPassword string `json:"secondary_password" valid:"Required;"`
// 	Amount            string `gorm:"column(amount)" json:"amount" valid:"Required;"`
// }

// AddTransferForm struct
type AddTransferForm struct {
	Amount            float64 `form:"amount" json:"amount" valid:"Required;"`
	EwalletTypeCode   string  `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required;"`
	EwalletTypeCodeTo string  `form:"ewallet_type_code_to" json:"ewallet_type_code_to" valid:"Required;"`
	To                string  `form:"to" json:"to" valid:"Required;"`
	Remark            string  `form:"remark" json:"remark"`
	SecondaryPin      string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

// AddTransferExchangeForm struct
type AddTransferExchangeForm struct {
	SigningKey   string  `form:"signing_key" json:"signing_key"`
	Amount       float64 `form:"amount" json:"amount" valid:"Required;"`
	EwalletType  string  `form:"ewallet_type" json:"ewallet_type" valid:"Required;"`
	To           string  `form:"to" json:"to" valid:"Required;"`
	Remark       string  `form:"remark" json:"remark"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

// AddTransferExchangeBatchForm struct
type AddTransferExchangeBatchForm struct {
	TransactionData string `form:"transaction_data" json:"transaction_data" valid:"Required;"`
	// EwalletType     string `form:"ewallet_type" json:"ewallet_type" valid:"Required;"`
	// AddrTo          string `form:"address_to" json:"address_to" valid:"Required;"`
}

// Get Setting struct
type GetSettingForm struct {
	WalletType string `json:"wallet_type" valid:"MaxSize(25)"`
}

//Get transaction statement
type TransactionStatements struct {
	Page           int64  `form:"page" json:"page"`
	DateFrom       string `form:"date_from" json:"date_from" valid:"MaxSize(100)"`
	DateTo         string `form:"date_to" json:"date_to" valid:"MaxSize(100)"`
	WalletTypeCode string `form:"wallet_type_code" json:"wallet_type_code"`
}

// get Reward Statements
type RewardStatements struct {
	RewardType string `form:"reward_type" json:"reward_type" valid:"Required;MaxSize(100)"`
	BnsId      string `form:"bns_id" json:"bns_id" valid:"Required;MaxSize(100)"`
	Limit      string `form:"limit" json:"limit" valid:"Required;"`
	Page       string `form:"page" json:"page" valid:"Required;"`
	//OrderField string `form:"order_field" json:"order_field" valid:"Required;MaxSize(20)"`
	//OrderType  string `form:"order_type" json:"order_type" valid:"Required;MaxSize(20)"`
}

// Crypto Deposit
type GetCryptoAddressStruct struct {
	WalletType string `json:"wallet_type"`
}

// Balance Struc
type BalanceStruc struct {
	WalletBalance interface{} `json:"wallet_balance"`
	Achievement   interface{} `json:"achievement"`
}

// ConvertForm struct
type ConvertForm struct {
	ConvertFrom string `json:"convert_from" valid:"Required;"`
	ConvertTo   string `json:"convert_to" valid:"Required;"`
	Amount      string `json:"amount" valid:"Required;"`
}

//Get withdraw transaction fee
type WithdrawTransactionFee struct {
	Address         string `form:"address" json:"address" valid:"Required;"`
	Amount          string `form:"amount" json:"amount" valid:"Required;"`
	EwalletTypeCode string `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required;"`
}

//Get transfer exchange transaction fee
type TransferExchangeTransactionFee struct {
	EwalletTypeCode string `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required;"`
	FromAddress     string `form:"from_address" json:"from_address" valid:"Required;"`
	ToAddress       string `form:"to_address" json:"to_address" valid:"Required;"`
	Amount          string `form:"amount" json:"amount" valid:"Required;"`
}

// func PostCredit(c *gin.Context) {
// 	var (
// 		appG = app.Gin{C: c}
// 		form AddWalletForm
// 	)

// 	ok, msg := app.BindAndValid(c, &form)

// 	if ok == false {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	if !helpers.IsNumeric(form.Amount) {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, "amount_nan", "")
// 		return
// 	}

// 	u, ok := c.Get("access_user")
// 	member := u.(*models.Members)

// 	tx := models.BeginTx(&sql.TxOptions{})

// 	memService := member_service.Wallet{
// 		WalletType: form.WalletType,
// 		SubId:      member.SubID,
// 		Amount:     helpers.NumberFormat2Dec(form.Amount),
// 		TransType:  form.TransType,
// 		Tx:         tx,
// 	}

// 	member, err := models.GetMemberBySubID(member.SubID)

// 	if err != nil {
// 		tx.Rollback()
// 		appG.ResponseError(err)
// 		return
// 	}
// 	if member == nil {
// 		tx.Rollback()
// 		appG.Response("error", http.StatusBadRequest, e.ERROR, "member_not_found")
// 		return
// 	}

// 	_, err = memService.Credit()

// 	if err != nil {
// 		tx.Rollback()
// 		appG.ResponseError(err)
// 		return
// 	}

// 	tx.Commit()
// 	appG.Response("success", http.StatusOK, e.SUCCESS, nil)
// }

// func PostDebit(c *gin.Context) {
// 	var (
// 		appG = app.Gin{C: c}
// 		form AddDebitForm
// 	)

// 	ok, msg := app.BindAndValid(c, &form)

// 	if ok == false {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	if !helpers.IsNumeric(form.Amount) {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, "amount_nan", "")
// 		return
// 	}

// 	ewt_setup, err := models.GetEwtSetup(form.WalletType)

// 	if err != nil {
// 		appG.ResponseError(err)
// 		return
// 	}

// 	if math.Mod(helpers.NumberFormat2Dec(form.Amount), ewt_setup.WithdrawBlk) > 0 {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, "not_block_of_10", "")
// 		return
// 	}

// 	u, ok := c.Get("access_user")
// 	member := u.(*models.Members)

// 	tx := models.BeginTx(&sql.TxOptions{})

// 	memService := member_service.Wallet{
// 		WalletType:        form.WalletType,
// 		SubId:             member.SubID,
// 		Amount:            helpers.NumberFormat2Dec(form.Amount),
// 		EWalletTypeId:     strconv.Itoa(ewt_setup.Id),
// 		TransType:         strings.ToUpper("withdrawal"),
// 		SecondaryPassword: form.SecondaryPassword,
// 		TxAddr:            form.WithdrawAddr,
// 		Tx:                tx,
// 	}

// 	mem, err := models.GetMemberBySubID(member.SubID)

// 	if mem == nil {
// 		tx.Rollback()
// 		appG.Response("error", http.StatusBadRequest, e.MEMBER_NOT_FOUND, "member_not_found")
// 		return
// 	}

// 	_, err = memService.Debit()

// 	if err != nil {
// 		tx.Rollback()
// 		appG.ResponseError(err)
// 		return
// 	}

// 	tx.Commit()
// 	appG.Response("success", http.StatusOK, e.SUCCESS, nil)
// }

// func PostCancel(c *gin.Context) {
// 	var (
// 		appG = app.Gin{C: c}
// 		form AddWithdrawalForm
// 	)

// 	ok, msg := app.BindAndValid(c, &form)

// 	if ok == false {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	u, ok := c.Get("access_user")
// 	member := u.(*models.Members)

// 	tx := models.BeginTx(&sql.TxOptions{})

// 	memService := member_service.Wallet{
// 		MemberId: strconv.Itoa(member.ID),
// 		DocNo:    form.DocNo,
// 		Tx:       tx,
// 	}

// 	withdraw, err := memService.CancelWithdrawal()

// 	if err != nil {
// 		tx.Rollback()
// 		appG.ResponseError(err)
// 		return
// 	}

// 	ewt_setup, ewt_err := models.GetEwtSetupById(helpers.NumberFormatInt(withdraw.EwalletTypeId))

// 	if ewt_err != nil {
// 		tx.Rollback()
// 		appG.ResponseError(ewt_err)
// 		return
// 	}

// 	memService = member_service.Wallet{
// 		WalletType: ewt_setup.CurrencyCode,
// 		SubId:      member.SubID,
// 		Amount:     (withdraw.Amount * -1) + withdraw.AdminFee,
// 		TransType:  strings.ToUpper("withdrawal_refund"),
// 		Tx:         tx,
// 	}

// 	_, err = memService.Credit()

// 	if err != nil {
// 		tx.Rollback()
// 		appG.ResponseError(err)
// 		return
// 	}

// 	tx.Commit()
// 	appG.Response("success", http.StatusOK, e.SUCCESS, nil)
// }

func GetMemberBalanceListv1(c *gin.Context) {
	type GetMemberBalanceForm struct {
		EwalletTypeCode string `form:"ewallet_type_code" json:"ewallet_type_code"`
	}

	var (
		appG = app.Gin{C: c}
		form GetMemberBalanceForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	balance, err := wallet_service.GetMemberBalanceListv1(entMemberID, form.EwalletTypeCode, langCode)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	var arrDataReturn interface{}

	if balance != nil {
		arrDataReturn = balance
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

func GetWalletTransactionv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form TransactionStatements
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	members := u.(*models.EntMemberMembers)
	memId := members.EntMemberID

	walStatement := wallet_service.WalletTransactionStruct{
		MemberID: memId,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		Page:     form.Page,
		LangCode: langCode,
	}

	rst, err := walStatement.WalletStatement("")

	if err != nil {
		err = errors.New("failed_to_get_wallet_statement")
		appG.ResponseError(err)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, rst)
}

func GetCryptoAddress(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetCryptoAddressStruct
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	// u, ok := c.Get("access_user")
	// members := u.(*models.EntMemberMembers)
	// memId := members.EntMemberID

	// if form.WalletType == "USDT" {
	// 	form.WalletType = "ETH"
	// }

	// address, err := models.GetMemberCryptoByMemID(memId, "")

	// if address == nil {
	// 	// err = errors.New("empty_member_crypto")
	// 	// appG.ResponseError(err)
	// 	// return

	// 	arrDataReturn := make([]interface{}, 0) //return empty
	// message := app.MsgStruct{
	// 	Msg: "success",
	// }

	// appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
	// 	return
	// }

	// if err != nil {
	// 	err = errors.New("failed_to_get_address")
	// 	appG.ResponseError(err)
	// 	return
	// }

	address, err := models.GetDepositWalletAddress()

	if address == nil {
		arrDataReturn := make([]interface{}, 0) //return empty
		message := app.MsgStruct{
			Msg: "success",
		}

		appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
		return
	}

	if err != nil {
		err = errors.New("failed_to_get_address")
		appG.ResponseError(err)
		return
	}

	//check hash
	// hash := base.SHA256(strings.ToLower(address.CryptoAddr + "" + "USDT" + "WOD"))

	// if hash != address.CryptoEncryptAddr {
	// 	err = errors.New("wrong_key")
	// 	appG.ResponseError(err)
	// 	return
	// }

	arrDataReturn := map[string]interface{}{
		// "crypto_type":    address.CryptoType,
		"crypto_address": address.CryptoAddr,
		// "private_key":    address.PrivateKey,
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)

}

// func GetWalletSetting(c *gin.Context) {
// 	var (
// 		appG           = app.Gin{C: c}
// 		form           GetSettingForm
// 		final_settings []interface{}
// 	)

// 	ok, msg := app.BindAndValid(c, &form)

// 	if ok == false {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	u, ok := c.Get("access_user")
// 	member := u.(*models.Members)

// 	memService := member_service.Wallet{
// 		SubId:      member.SubID,
// 		WalletType: form.WalletType,
// 	}

// 	// post Transfer
// 	settings, err := memService.PostSetting()

// 	if err != nil {
// 		appG.ResponseError(err)
// 		return
// 	}

// 	for _, v := range settings {
// 		setting_from, setting_to, err := models.GetConvertSetting(v.Id)
// 		var sf_settings []interface{}
// 		var st_settings []interface{}

// 		if err != nil {
// 			appG.ResponseError(err)
// 			return
// 		}

// 		for _, sf := range setting_from {
// 			sf.WalletName = appG.Trans(strings.Replace(strings.ToLower(sf.WalletName), " ", "_", -1), nil)
// 			sf_settings = append(sf_settings, sf)
// 		}

// 		for _, st := range setting_to {
// 			st.WalletName = appG.Trans(strings.Replace(strings.ToLower(st.WalletName), " ", "_", -1), nil)
// 			st_settings = append(st_settings, st)
// 		}

// 		//if len(sf_settings) > 0{
// 		//	v.ConvertFromSetting.ConvertFrom = 1
// 		//}else {
// 		//	v.ConvertFromSetting.ConvertFrom = 0
// 		//}
// 		//if len(st_settings) > 0{
// 		//	v.ConvertToSetting.ConvertTo = 1
// 		//}else {
// 		//	v.ConvertToSetting.ConvertTo = 0
// 		//}
// 		v.ConvertFromSetting.WalletList = sf_settings
// 		v.ConvertToSetting.WalletList = st_settings
// 		v.BDisplayName = appG.Trans(strings.Replace(strings.ToLower(v.BDisplayName), " ", "_", -1), nil)

// 		final_settings = append(final_settings, v)
// 	}

// 	appG.Response("success", http.StatusOK, e.SUCCESS, final_settings)
// }
// func PostConvert(c *gin.Context) {
// 	var (
// 		appG = app.Gin{C: c}
// 		form ConvertForm
// 	)

// 	ok, msg := app.BindAndValid(c, &form)

// 	if ok == false {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	u, ok := c.Get("access_user")
// 	member := u.(*models.Members)

// 	balance, err_balance := models.GetMemberBalance(form.ConvertFrom, member.ID)
// 	var con_balnace int = int(balance)
// 	if helpers.NumberFormatInt(form.Amount) > con_balnace {
// 		appG.Response("error", http.StatusBadRequest, e.INSUFFICIENT_BALANCE, nil)
// 		return
// 	}

// 	if err_balance != nil {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	tx := models.BeginTx(&sql.TxOptions{Isolation: sql.LevelSerializable})
// 	memService := member_service.Convert{
// 		MemberId:    member.ID,
// 		ConvertFrom: form.ConvertFrom,
// 		ConvertTo:   form.ConvertTo,
// 		Amount:      helpers.NumberFormat2Dec(form.Amount),
// 		//LangCode: 		form.LangCode,
// 		Tx: tx,
// 	}

// 	err := memService.PostConvert()

// 	if err != nil {
// 		tx.Rollback()
// 		appG.ResponseError(err)
// 		return
// 	}

// 	tx.Commit()
// 	appG.Response("success", http.StatusOK, e.SUCCESS, nil)
// }
// func GetRewardStatements(c *gin.Context) {
// 	var (
// 		appG = app.Gin{C: c}
// 		form RewardStatements
// 	)
// 	u, ok := c.Get("access_user")
// 	// validate input
// 	ok, msg := app.BindAndValid(c, &form)
// 	if ok == false {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	member := u.(*models.Members)

// 	memService := member_service.RewardStatement{
// 		UserId:     member.UserID,
// 		MemberId:   member.ID,
// 		RewardType: form.RewardType,
// 		BnsId:      form.BnsId,
// 		Page:       helpers.NumberFormatInt(form.Page),
// 		Limit:      helpers.NumberFormatInt(form.Limit),
// 	}

// 	var response app.JsonResponse

// 	transaction, total_records, total_pages, end_flag := memService.GetRewardStatements()

// 	response.Data = transaction
// 	response.TotalPages = int64(total_pages)
// 	response.TotalRecords = total_records
// 	response.EndFlag = end_flag

// 	appG.ResponseList(http.StatusOK, http.StatusOK, response.Data, helpers.NumberFormatInt(form.Page), response.TotalPages, response.TotalRecords, response.EndFlag, nil)
// }

func GetMemberStatementListv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}
	// u, _ := c.Get("access_user")
	// member := u.(*models.Members)

	arrStatementListSetting, _ := models.GetSysGeneralSetupByID("statement_list_api_setting")
	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	var arrDataReturn []arrStatementListSettingListStruct
	if arrStatementListSetting != nil {
		var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
		json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
		arrDataReturn = arrStatementListSettingList["statement_list"]
		for _, v1 := range arrStatementListSettingList["statement_list"] {
			v1.Name = helpers.Translate(v1.Name, langCode)
		}
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

func GetWithdrawStatement(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form TransactionStatements
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")

		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	walStatement := wallet_service.WalletTransactionStructV2{
		MemberID:       entMemberID,
		DateFrom:       form.DateFrom,
		DateTo:         form.DateTo,
		Page:           form.Page,
		LangCode:       langCode,
		TransType:      "WITHDRAW",
		WalletTypeCode: strings.ToUpper(form.WalletTypeCode),
	}

	rst, err := walStatement.WithdrawStatement()

	if err != nil {
		err = errors.New("failed_to_get_withdraw_statement")
		appG.ResponseError(err)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, rst)
}

func GetTransferStatement(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form TransactionStatements
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	members := u.(*models.EntMemberMembers)
	memId := members.EntMemberID

	walStatement := wallet_service.WalletTransactionStruct{
		MemberID: memId,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		Page:     form.Page,
		LangCode: langCode,
	}

	rst, err := walStatement.WalletStatement("TRANSFER")

	if err != nil {
		err = errors.New("failed_to_get_transfer_statement")
		appG.ResponseError(err)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, rst)
}

func GetWithdrawTransactionFee(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form WithdrawTransactionFee
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	u, ok := c.Get("access_user")
	members := u.(*models.EntMemberMembers)
	memId := members.EntMemberID

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	withdrawFee := wallet_service.WithdrawTransactionFeeStruct{
		MemId:           memId,
		Address:         form.Address,
		Amount:          form.Amount,
		EwalletTypeCode: form.EwalletTypeCode,
		LangCode:        langCode,
	}

	rst, err := withdrawFee.GetWithdrawTransactionFee()

	if err != nil {
		// err = errors.New("failed_to_get_withdraw_transaction_fee")
		appG.ResponseError(err)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, rst)
}

//for ui-v2
func GetMemberStatementListv2(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}
	// u, _ := c.Get("access_user")
	// member := u.(*models.Members)

	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Code string `json:"code"`
	}

	type arrStatementTypeStruct struct {
		WalletStatement []arrStatementListSettingListStruct `json:"wallet_statement"`
		BonusStatement  []arrStatementListSettingListStruct `json:"bonus_statement"`
	}

	arrStatementListSetting, _ := models.GetSysGeneralSetupByID("statement_list_api_setting_v2")

	var arrDataReturn []arrStatementTypeStruct
	if arrStatementListSetting != nil {
		var arrStatementListSettingList map[string][]arrStatementTypeStruct
		json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
		arrDataReturn = arrStatementListSettingList["statement_list"]

		for k, v1 := range arrStatementListSettingList["statement_list"] {

			for kWal, vWal := range v1.WalletStatement {
				vWal.Name = helpers.Translate(vWal.Name, langCode)
				arrDataReturn[k].WalletStatement[kWal] = vWal
			}

			for kBns, vBns := range v1.BonusStatement {
				vBns.Name = helpers.Translate(vBns.Name, langCode)
				arrDataReturn[k].BonusStatement[kBns] = vBns
			}
		}
	}

	message := app.MsgStruct{
		Msg: "suceess",
	}

	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

type WalletTypeStatementStruct struct {
	Page           int64  `form:"page" json:"page"`
	DateFrom       string `form:"date_from" json:"date_from" valid:"MaxSize(100)"`
	DateTo         string `form:"date_to" json:"date_to" valid:"MaxSize(100)"`
	WalletTypeCode string `form:"wallet_type_code" json:"wallet_type_code"`
	TransType      string `form:"trans_type" json:"trans_type"`
	RewardTypeCode string `form:"reward_type_code" json:"reward_type_code"`
}

func GetWalletStatement(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form WalletTypeStatementStruct
	)

	//validate input
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
	memId := members.EntMemberID

	walStatement := wallet_service.WalletTransactionStructV2{
		MemberID:       memId,
		DateFrom:       form.DateFrom,
		DateTo:         form.DateTo,
		Page:           form.Page,
		LangCode:       langCode,
		WalletTypeCode: strings.ToUpper(form.WalletTypeCode),
		TransType:      form.TransType,
		RewardTypeCode: strings.ToUpper(form.RewardTypeCode),
	}

	rst, err := walStatement.WalletStatementV4()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "failed_to_get_wallet_statement"}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, rst)
	return
}

func GetWithdrawDetail(c *gin.Context) {

	type WithdrawDetailStruct struct {
		DocNo string `form:"doc_no" json:"doc_no"`
	}

	type WithdrawDetailReturnStruct struct {
		GasFee        string `json:"gas_fee"`
		NetAmount     string `json:"net_amount"`
		CryptoAddress string `json:"crypto_address"`
		CreatedAt     string `json:"created_at"`
	}

	var (
		appG = app.Gin{C: c}
		form WithdrawDetailStruct
	)

	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	rst, err := models.GetEwtWithdrawDetailByDocNo(form.DocNo)

	arrWalletStatementList := make([]WithdrawDetailReturnStruct, 0)

	if rst != nil {
		arrWalletStatementList = append(arrWalletStatementList,
			WithdrawDetailReturnStruct{
				GasFee:        fmt.Sprintf("%.6f", rst.GasFee),
				NetAmount:     fmt.Sprintf("%.6f", rst.NetAmount),
				CryptoAddress: rst.CryptoAddrTo,
				CreatedAt:     rst.CreatedAt.Format("2006-01-02 15:04:05"),
			})
	}

	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}

	var arrTableHeaderList []arrStatementListSettingListStruct

	arrStatementListSetting, _ := models.GetSysGeneralSetupByID("withdraw_detail_api_setting")
	if arrStatementListSetting != nil {
		var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
		json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
		arrTableHeaderList = arrStatementListSettingList["table_header_list"]
		for k, v1 := range arrStatementListSettingList["table_header_list"] {
			v1.Name = helpers.Translate(v1.Name, langCode)
			arrTableHeaderList[k] = v1
		}
	}

	if err != nil {
		err = errors.New("failed_to_get_withdraw_detail")
		appG.ResponseError(err)
		return
	}

	arrDataReturn := map[string]interface{}{
		"data":              arrWalletStatementList,
		"table_header_list": arrTableHeaderList,
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)

}

func GetTransferDetail(c *gin.Context) {

	type TransferDetailStruct struct {
		DocNo string `form:"doc_no" json:"doc_no"`
	}

	type TransferDetailReturnStruct struct {
		MemberFrom string `json:"member_from"`
		MemberTo   string `json:"member_to"`
		WalletFrom string `json:"wallet_from"`
		WalletTo   string `json:"wallet_to"`
		CreatedAt  string `json:"created_at"`
	}

	var (
		appG = app.Gin{C: c}
		form TransferDetailStruct
	)

	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	rst, err := models.GetEwtTransferDetailByDocNo(form.DocNo)

	arrWalletStatementList := make([]TransferDetailReturnStruct, 0)

	if rst != nil {
		arrWalletStatementList = append(arrWalletStatementList,
			TransferDetailReturnStruct{
				MemberFrom: rst.MemberFrom,
				MemberTo:   rst.MemberTo,
				WalletFrom: rst.WalletFrom,
				WalletTo:   rst.WalletTo,
				CreatedAt:  rst.CreatedAt.Format("2006-01-02 15:04:05"),
			})
	}

	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}

	var arrTableHeaderList []arrStatementListSettingListStruct

	arrStatementListSetting, _ := models.GetSysGeneralSetupByID("transfer_detail_api_setting")
	if arrStatementListSetting != nil {
		var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
		json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
		arrTableHeaderList = arrStatementListSettingList["table_header_list"]
		for k, v1 := range arrStatementListSettingList["table_header_list"] {
			v1.Name = helpers.Translate(v1.Name, langCode)
			arrTableHeaderList[k] = v1
		}
	}

	if err != nil {
		err = errors.New("failed_to_get_transfer_detail")
		appG.ResponseError(err)
		return
	}

	arrDataReturn := map[string]interface{}{
		"data":              arrWalletStatementList,
		"table_header_list": arrTableHeaderList,
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)

}

//post Transfer func
func PostTransfer(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form AddTransferForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	tx := models.Begin()

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("PostTransfer-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: members.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	//get wallet setting
	arrWalCond := make([]models.WhereCondFn, 0)
	arrWalCond = append(arrWalCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: form.EwalletTypeCode},
	)
	walSetup, err := models.GetEwtSetupFn(arrWalCond, "", false)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	if walSetup == nil {
		base.LogErrorLog("PostTransfer - empty wallet setup", walSetup, arrWalCond, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_wallet"}, nil)
		return
	}

	transfer := wallet_service.PostTransferStruct{
		MemberId:      entMemberID,
		EwtTypeCode:   strings.ToUpper(form.EwalletTypeCode),
		EwtTypeCodeTo: strings.ToUpper(form.EwalletTypeCodeTo),
		Amount:        form.Amount,
		MemberTo:      form.To,
		Remark:        form.Remark,
		LangCode:      langCode,
	}

	arrData, err := transfer.PostTransfer(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "transfer_successful"}, arrData)
	return

}

//post Transfer Exchange func
func PostTransferExchange(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form AddTransferExchangeForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	tx := models.Begin()

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("PostTransferExchange-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: members.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	//get wallet setting
	arrWalCond := make([]models.WhereCondFn, 0)
	arrWalCond = append(arrWalCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: form.EwalletType},
	)
	walSetup, err := models.GetEwtSetupFn(arrWalCond, "", false)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	if walSetup == nil {
		base.LogErrorLog("PostTransferExchange - empty wallet setup", walSetup, arrWalCond, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_wallet"}, nil)
		return
	}

	if walSetup.Control == "BLOCKCHAIN" {
		if form.SigningKey == "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "signing_key_cannot_be_empty"}, nil)
			return
		}

		transfer := wallet_service.PostTransferExchangeStruct{
			SigningKey:    form.SigningKey,
			MemberId:      entMemberID,
			EwtTypeCode:   strings.ToUpper(form.EwalletType),
			Amount:        form.Amount,
			WalletAddress: form.To,
			Remark:        form.Remark,
			LangCode:      langCode,
		}

		arrData, err := transfer.PostTransferExchange(tx)

		if err != nil {
			tx.Rollback()
			appG.ResponseError(err)
			return
		}

		tx.Commit()

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "transfer_successful"}, arrData)
	} else {
		transfer := wallet_service.PostTransferStruct{
			MemberId:    entMemberID,
			EwtTypeCode: strings.ToUpper(form.EwalletType),
			Amount:      form.Amount,
			MemberTo:    form.To,
			Remark:      form.Remark,
			LangCode:    langCode,
		}

		arrData, err := transfer.PostTransfer(tx)

		if err != nil {
			tx.Rollback()
			appG.ResponseError(err)
			return
		}

		tx.Commit()

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "transfer_successful"}, arrData)
		return
	}

}

type GetCryptoDetailStruct struct {
	EwtTypeCode string `json:"ewallet_type_code" form:"ewallet_type_code"`
	CryptoCode  string `json:"crypto_code" form:"crypto_code"`
}

type CryptoDetailSetupStruct struct {
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	DecimalFrom  int     `json:"decimal_from"`
	DecimalTo    int     `json:"decimal_to"`
}

type ActiveOrderStruct struct {
	Amount          float64 `json:"amount"`
	ConvertedAmount float64 `json:"converted_amount"`
	ExpiryAt        int64   `json:"expiry_at"`
	CryptoAddress   string  `json:"crypto_address"`
}

type GetCryptoDetailResponseStruct struct {
	Rate        float64                 `json:"rate"`
	Setup       CryptoDetailSetupStruct `json:"setup"`
	ActiveOrder *ActiveOrderStruct      `json:"active_order"`
	Lock        int                     `json:"lock"`
}

// type BlockchainSettingStruct struct {

// }

func GetCryptoDetail(c *gin.Context) {
	var (
		appG        = app.Gin{C: c}
		form        GetCryptoDetailStruct
		activeOrder *ActiveOrderStruct
		lock        int
		// blockchainSetting BlockchainSettingStruct
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: msg[0],
		}, nil)
		return
	}

	u, _ := c.Get("access_user")
	members := u.(*models.EntMemberMembers)
	memID := members.EntMemberID

	// json decode BlockchainDepositSetting
	// json.Unmarshal([]byte(wallet.BlockchainDepositSetting), &arrDataReturn.Object)

	// begin transaction
	tx := models.Begin()

	data, depositInfoErr := wallet_service.GetDepositInfo(tx, wallet_service.GetDepositInfoStruct{
		CryptoCode:  form.CryptoCode,
		EwtTypeCode: form.EwtTypeCode,
		MemberID:    memID,
	}, true)

	if depositInfoErr != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: depositInfoErr}, nil)
		return
	}

	// redeeem address if cant lock price
	if !data.CanLock {
		cryptoAddress, cryptoAddrErr := member_service.ProcessGetMemAddress(tx, memID, data.WalletFrom.EwtTypeCode)

		if cryptoAddrErr != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		activeOrder = &ActiveOrderStruct{
			CryptoAddress: cryptoAddress,
		}
	} else {
		lock = 1
		// Check and return the EwtTopup Record
		EwtTopup, EwtTopupErr := wallet_service.CheckCryptoPurchase(tx, wallet_service.CheckCryptoPurchaseStruct{
			WalletTo:   data.WalletTo,
			WalletFrom: data.WalletFrom,
			MemberID:   memID,
		})

		if EwtTopupErr != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: EwtTopupErr}, nil)
			return
		}

		if EwtTopup != nil {
			cryptoAddress, cryptoAddrErr := member_service.ProcessGetMemAddress(tx, memID, form.CryptoCode)

			if cryptoAddrErr != nil {
				models.Rollback(tx)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}

			activeOrder = &ActiveOrderStruct{
				Amount:          EwtTopup.TotalIn,
				CryptoAddress:   cryptoAddress,
				ConvertedAmount: EwtTopup.ConvertedTotalAmount,
				ExpiryAt:        EwtTopup.ExpiryAt.Unix(),
			}
			data.Rate = EwtTopup.ConversionRate
		} else {
			activeOrder = nil
		}
	}

	err := models.Commit(tx)

	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("walletController:GetCryptoDetail()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	settingValue1, _ := strconv.ParseFloat(data.Setting.SettingValue1, 64)
	settingValue2, _ := strconv.ParseFloat(data.Setting.SettingValue2, 64)

	arrDataReturn := GetCryptoDetailResponseStruct{
		Rate: data.Rate,
		Setup: CryptoDetailSetupStruct{
			Min:          settingValue1,
			Max:          settingValue2,
			FromCurrency: data.WalletFrom.CurrencyCode,
			DecimalFrom:  data.WalletFrom.DecimalPoint,
			ToCurrency:   data.WalletTo.CurrencyCode,
			DecimalTo:    data.WalletTo.DecimalPoint,
		},
		ActiveOrder: activeOrder,
		Lock:        lock,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)

}

type AddCryptoPurchaseStruct struct {
	Amount     float64 `json:"amount" form:"amount"`
	CryptoCode string  `json:"crypto_code" form:"crypto_code"`
}

func AddCryptoPurchase(c *gin.Context) {
	var (
		appG        = app.Gin{C: c}
		form        AddCryptoPurchaseStruct
		activeOrder *ActiveOrderStruct
		lock        int
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: msg[0],
		}, nil)
		return
	}

	u, _ := c.Get("access_user")

	members := u.(*models.EntMemberMembers)
	memID := members.EntMemberID

	// begin transaction
	tx := models.Begin()

	data, depositInfoErr := wallet_service.GetDepositInfo(tx, wallet_service.GetDepositInfoStruct{
		CryptoCode: form.CryptoCode,
		MemberID:   memID,
	}, true)

	if depositInfoErr != "" {
		models.Rollback(tx)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: depositInfoErr}, nil)
		return
	}

	if !data.CanLock {
		cryptoAddress, cryptoAddrErr := member_service.ProcessGetMemAddress(tx, memID, form.CryptoCode)

		if cryptoAddrErr != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		activeOrder = &ActiveOrderStruct{
			CryptoAddress: cryptoAddress,
		}
	} else {
		lock = 1
		// Check and return the EwtTopup Record
		EwtTopup, EwtTopupErr := wallet_service.CheckCryptoPurchase(tx, wallet_service.CheckCryptoPurchaseStruct{
			WalletTo:   data.WalletTo,
			WalletFrom: data.WalletFrom,
			MemberID:   memID,
		})

		if EwtTopupErr != "" {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: EwtTopupErr}, nil)
			return
		}

		// redeem address
		cryptoAddress, cryptoAddrErr := member_service.ProcessGetMemAddress(tx, memID, form.CryptoCode)
		if cryptoAddrErr != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		// Add Crypto purchase if no record found
		if EwtTopup == nil {

			AddCryptoPurchase, AddCryptoPurchaseErr := wallet_service.AddCryptoPurchase(tx, wallet_service.CryptoPurchaseStruct{
				Rate:          data.Rate,
				WalletTo:      data.WalletTo,
				WalletFrom:    data.WalletFrom,
				Amount:        form.Amount,
				CryptoCode:    form.CryptoCode,
				MemberID:      memID,
				CryptoAddress: cryptoAddress,
			})

			if AddCryptoPurchaseErr != "" {
				models.Rollback(tx)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: AddCryptoPurchaseErr}, nil)
				return
			}
			EwtTopup = AddCryptoPurchase
		} else {
			data.Rate = EwtTopup.ConversionRate
		}

		activeOrder = &ActiveOrderStruct{
			CryptoAddress:   cryptoAddress,
			Amount:          EwtTopup.TotalIn,
			ConvertedAmount: EwtTopup.ConvertedTotalAmount,
			ExpiryAt:        EwtTopup.ExpiryAt.Unix(),
		}
	}

	err := models.Commit(tx)

	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("walletController:AddCryptoPurchase()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	settingValue1, _ := strconv.ParseFloat(data.Setting.SettingValue1, 64)
	settingValue2, _ := strconv.ParseFloat(data.Setting.SettingValue2, 64)

	arrDataReturn := GetCryptoDetailResponseStruct{
		Rate:        data.Rate,
		ActiveOrder: activeOrder,
		Setup: CryptoDetailSetupStruct{
			Min:          settingValue1,
			Max:          settingValue2,
			ToCurrency:   data.WalletTo.CurrencyCode,
			DecimalTo:    data.WalletTo.DecimalPoint,
			FromCurrency: data.WalletFrom.CurrencyCode,
			DecimalFrom:  data.WalletFrom.DecimalPoint,
		},
		Lock: lock,
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)

}

type CancelCryptoPurchaseStruct struct {
	CryptoCode string `json:"crypto_code" form:"crypto_code"`
}

func CancelCryptoPurchase(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form CancelCryptoPurchaseStruct
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: msg[0],
		}, nil)
		return
	}

	u, _ := c.Get("access_user")

	members := u.(*models.EntMemberMembers)
	memID := members.EntMemberID

	// begin transaction
	tx := models.Begin()

	data, depositInfoErr := wallet_service.GetDepositInfo(tx, wallet_service.GetDepositInfoStruct{
		CryptoCode: form.CryptoCode,
		MemberID:   memID,
	}, false)

	if depositInfoErr != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: depositInfoErr}, nil)
		return
	}

	// Cancel Crypto Purchase
	CancelCryptoPurchaseErr := wallet_service.CancelCryptoPurchase(tx, wallet_service.CancelCryptoPurchaseStruct{
		WalletTo:   data.WalletTo,
		WalletFrom: data.WalletFrom,
		MemberID:   memID,
	})

	if CancelCryptoPurchaseErr != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: CancelCryptoPurchaseErr}, nil)
		return
	}

	err := models.Commit(tx)

	if err != nil {
		base.LogErrorLog("walletController:CancelCryptoPurchase()", "Commit():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "cancelled_successfully"}, nil)

}

func GetWalletSummaryDetails(c *gin.Context) {

	type WalletSummaryDetailStruct struct {
		Page           int64  `form:"page" json:"page"`
		DateFrom       string `form:"date_from" json:"date_from" valid:"MaxSize(100)"`
		DateTo         string `form:"date_to" json:"date_to" valid:"MaxSize(100)"`
		WalletTypeCode string `form:"wallet_type_code" valid:"Required;"`
		TransType      string `form:"trans_type"`
	}

	var (
		appG = app.Gin{C: c}
		form WalletSummaryDetailStruct
	)

	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}

		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	members := u.(*models.EntMemberMembers)
	memId := members.EntMemberID

	walSumDet := wallet_service.WalletTransactionStructV2{ //share with wallet transaction
		MemberID:       memId,
		DateFrom:       form.DateFrom,
		DateTo:         form.DateTo,
		Page:           form.Page,
		LangCode:       langCode,
		WalletTypeCode: strings.ToUpper(form.WalletTypeCode),
		TransType:      form.TransType,
	}

	rst, err := walSumDet.WalletSummaryDetail()

	if err != nil {
		err = errors.New("failed_to_get_wallet_summary_detail")
		appG.ResponseError(err)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, rst)

}

//post Transfer Exchange func
func PostTransferExchangeBatch(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form AddTransferExchangeBatchForm
		err  error
	)

	// TransData struct
	type TransData struct {
		SigningKey      string `json:"signing_key"`
		Amount          string `json:"amount"`
		EwalletTypeCode string `json:"ewallet_type"`
		To              string `json:"to"`
		AccountName     string `json:"account_name"`
	}

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	// u, ok := c.Get("access_user")
	// if ok == false {
	// 	appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
	// 	return
	// }

	// members := u.(*models.EntMemberMembers)
	// entMemberID := members.EntMemberID

	arrTransData := make([]TransData, 0)
	err = json.Unmarshal([]byte(form.TransactionData), &arrTransData)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	//foreach call transfer
	for _, v := range arrTransData {
		tx := models.Begin()
		//get wallet setting
		arrWalCond := make([]models.WhereCondFn, 0)
		arrWalCond = append(arrWalCond,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: v.EwalletTypeCode},
		)
		walSetup, err := models.GetEwtSetupFn(arrWalCond, "", false)

		if err != nil {
			tx.Rollback()
			base.LogErrorLog("PostTransferExchangeBatch - fail to get wallet setup", err, arrWalCond, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		if walSetup == nil {
			tx.Rollback()
			base.LogErrorLog("PostTransferExchangeBatch - empty wallet setup", v, arrWalCond, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_wallet"}, nil)
			return
		}

		//get current member info
		arrMemFromCond := make([]models.WhereCondFn, 0)
		arrMemFromCond = append(arrMemFromCond,
			models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: v.AccountName},
		)
		memberFrom, err := models.GetEntMemberFn(arrMemFromCond, "", false) //get member details

		if err != nil {
			tx.Rollback()
			base.LogErrorLog("PostTransferExchangeBatch - fail to get member info", err, v, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_member_info"}, nil)
			return

		}

		if memberFrom == nil {
			tx.Rollback()
			base.LogErrorLog("PostTransferExchangeBatch - empty member info", arrMemFromCond, v, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		amount, err := strconv.ParseFloat(v.Amount, 64)

		if err != nil {
			tx.Rollback()
			base.LogErrorLog("PostTransferExchangeBatch - fail_to_convert_amount_to_float64", err, v, true)
			appG.ResponseError(err)
			return
		}

		memToId := 0
		memToName := v.To
		if walSetup.Control == "BLOCKCHAIN" {
			if v.SigningKey == "" {
				tx.Rollback()
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "signing_key_cannot_be_empty"}, nil)
				return
			}

			//check if tagged to this member
			arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
			arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
				models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
				models.WhereCondFn{Condition: "ent_member_crypto.crypto_address = ?", CondValue: v.To},
				models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
			)
			arrEntMemberCrypto, err := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)

			if err != nil {
				tx.Rollback()
				base.LogErrorLog("PostTransferExchangeBatch - fail_to_check_address_to", err, v, true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "fail_to_check_address_to"}, nil)
				return
			}

			if arrEntMemberCrypto == nil {
				tx.Rollback()
				base.LogErrorLog("PostTransferExchangeBatch - invalid_address_to", arrEntMemberCryptoFn, v, true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_address_to"}, nil)
				return
			}

			memToId = arrEntMemberCrypto.MemberID
		} else {
			//get member to info
			arrMemToCond := make([]models.WhereCondFn, 0)
			arrMemToCond = append(arrMemToCond,
				models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: v.To},
			)
			memberTo, err := models.GetEntMemberFn(arrMemToCond, "", false) //get member to details

			if err != nil {
				tx.Rollback()
				base.LogErrorLog("PostTransferExchangeBatch - fail to get member to info", err, v, true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_member_info"}, nil)
				return

			}

			if memberTo == nil {
				tx.Rollback()
				base.LogErrorLog("PostTransferExchangeBatch - empty member info", arrMemToCond, v, true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}

			memToId = memberTo.ID
			memToName = memberTo.NickName
		}

		if memberFrom.TaggedMemberID != memToId {
			tx.Rollback()
			base.LogErrorLog("PostTransferExchangeBatch - transfer_to_member_is_not_tagged", memberFrom, v, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transfer_to_member_is_not_tagged"}, nil)
			return
		}

		if walSetup.Control == "BLOCKCHAIN" {
			transfer := wallet_service.PostTransferExchangeStruct{
				SigningKey:    v.SigningKey,
				MemberId:      memberFrom.ID,
				EwtTypeCode:   strings.ToUpper(v.EwalletTypeCode),
				Amount:        amount,
				WalletAddress: v.To,
				LangCode:      langCode,
			}

			_, err = transfer.PostTransferExchange(tx)

			if err != nil {
				tx.Rollback()
				appG.ResponseError(err)
				return
			}
		} else {
			transfer := wallet_service.PostTransferStruct{
				MemberId:    memberFrom.ID,
				EwtTypeCode: strings.ToUpper(v.EwalletTypeCode),
				Amount:      amount,
				MemberTo:    memToName,
				LangCode:    langCode,
			}

			_, err := transfer.PostTransfer(tx)

			if err != nil {
				tx.Rollback()
				appG.ResponseError(err)
				return
			}
		}

		tx.Commit()
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "transfer_successful"}, nil)
}

// AddTransferExchangeForm struct
type MemberAccountTransferExchangeBatchSetupForm struct {
	EwalletType  string `form:"ewallet_type" json:"ewallet_type" valid:"Required;"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

func GetMemberAccountTransferExchangeBatchSetupv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberAccountTransferExchangeBatchSetupForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv1-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: members.SecondaryPin,
	}

	err = pinValidation.CheckSecondaryPin()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	arrData := wallet_service.MemberAccountTransferExchangeBatchSetupStruct{
		MemberID:    members.ID,
		EntMemberID: entMemberID,
		LangCode:    langCode,
		EwalletType: form.EwalletType,
	}

	arrDataReturn, err := wallet_service.GetMemberAccountTransferExchangeBatchSetupv2(arrData)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}

func GetPendingTransferOut(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		// form GetPendingTransferOutForm
	)

	// ok, msg := app.BindAndValid(c, &form)

	// if ok == false {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
	// 	return
	// }

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	// if c.PostForm("lang_code") != "" {
	// 	langCode = c.PostForm("lang_code")
	// } else if c.GetHeader("Accept-Language") != "" {
	// 	langCode = c.GetHeader("Accept-Language")
	// }

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	pendingTransctions, pendingTransctionsErr := wallet_service.GetPendingTransferOut(entMemberID)

	if pendingTransctionsErr != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: pendingTransctionsErr}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
		"pending_list": pendingTransctions,
	})
	return
}

type AdjustOutForm struct {
	TransactionData string `form:"transaction_data" json:"transaction_data"`
	TransactionType string `form:"transaction_type" json:"transaction_type"`
	TokenType       string `form:"token_type" json:"token_type"`
	TransactionIds  string `form:"transaction_ids" json:"transaction_ids"`
}

func AdjustOut(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AdjustOutForm
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	adjustOutErr := wallet_service.AdjustOut(entMemberID, wallet_service.AdjustOutData{
		TransactionData: form.TransactionData,
		TransactionType: form.TransactionType,
		TokenType:       form.TokenType,
		TransactionIds:  form.TransactionIds,
	})

	if adjustOutErr != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: adjustOutErr}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return

}

type SigningKeyForm struct {
	EwalletTypeCode string `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required;"`
	Method          string `form:"method" json:"method" valid:"Required;"`
}

func GetWalletSigningKey(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form SigningKeyForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)

	setting, err := wallet_service.GetWalletSigningKeySetting(members.EntMemberID, form.EwalletTypeCode, strings.ToUpper(form.Method))

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, setting)
	return
}

// AddWithdrawForm struct
type AddWithdrawForm struct {
	Amount            float64 `form:"amount" json:"amount" valid:"Required;"`
	EwalletTypeCode   string  `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required;"`
	Address           string  `form:"address" json:"address" valid:"Required;"`
	SecondaryPin      string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
	Remark            string  `form:"remark" json:"remark"`
	ChargesType       string  `form:"charges_type" json:"charges_type"`
	EwalletTypeCodeTo string  `form:"ewallet_type_code_to" json:"ewallet_type_code_to" valid:"Required;"`
	VerificationCode  string  `form:"verification_code" json:"verification_code" valid:"Required;"`
}

func PostWithdraw(c *gin.Context) {
	var (
		appG   = app.Gin{C: c}
		form   AddWithdrawForm
		errMsg string
		err    error
	)

	tx := models.Begin()

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID
	receiverID := member.Email

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		tx.Rollback()
		base.LogErrorLog("PostWithdraw-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
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
		tx.Rollback()
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	//check block all withdrawal
	withdrawAllblockStatus := member_service.VerifyIfInNetwork(entMemberID, "WD_BLK_ALL")

	if withdrawAllblockStatus {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "network_is_being_blocked_from_withdraw"}, nil)
		return
	}

	//verify otp
	inputOtp := strings.Trim(form.VerificationCode, " ")

	errMsg = otp_service.ValidateOTP(tx, "EMAIL", receiverID, inputOtp, "WITHDRAW")
	if errMsg != "" {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	withdraw := wallet_service.PostWithdrawStruct{
		MemberId:          entMemberID,
		Address:           form.Address,
		Amount:            form.Amount,
		EwalletTypeCode:   form.EwalletTypeCode,
		LangCode:          langCode,
		Remark:            form.Remark,
		ChargesType:       form.ChargesType,
		EwalletTypeCodeTo: form.EwalletTypeCodeTo,
	}
	arrData, err := withdraw.PostWithdraw(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "withdraw_successful"}, arrData)
	return
}

// AddTransferExchangeFormV2 struct
type AddTransferExchangeFormV2 struct {
	Amount       float64 `form:"amount" json:"amount" valid:"Required;"`
	EwalletType  string  `form:"ewallet_type" json:"ewallet_type" valid:"Required;"`
	AddrTo       string  `form:"address_to" json:"address_to" valid:"Required;"`
	Remark       string  `form:"remark" json:"remark"`
	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

//post Transfer Exchange func
func PostTransferExchangeV2(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form AddTransferExchangeForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	tx := models.Begin()

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("PostTransferExchange-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: members.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	//get wallet setting
	arrWalCond := make([]models.WhereCondFn, 0)
	arrWalCond = append(arrWalCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: form.EwalletType},
	)
	walSetup, err := models.GetEwtSetupFn(arrWalCond, "", false)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	if walSetup == nil {
		base.LogErrorLog("PostTransferExchange - empty wallet setup", walSetup, arrWalCond, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_wallet"}, nil)
		return
	}

	if walSetup.Control == "BLOCKCHAIN" {
		// if form.SigningKey == "" {
		// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "signing_key_cannot_be_empty"}, nil)
		// 	return
		// }

		//begin gen signing key
		signKey := wallet_service.GenerateSigningKeyByModuleStruct{
			MemberId:       entMemberID,
			WalletTypeCode: strings.ToUpper(form.EwalletType),
			Module:         "TRANSFER",
			Amount:         form.Amount,
		}

		signingKey, err := signKey.GenerateSigningKeyByModule()
		//end gen signing key

		if err != nil {
			// base.LogErrorLog("PostTransferExchange - gen signing key error", err, signKey, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		if signingKey == "" {
			// base.LogErrorLog("PostTransferExchange - empty signing key", signingKey, signKey, true)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		transfer := wallet_service.PostTransferExchangeStruct{
			SigningKey:    signingKey,
			MemberId:      entMemberID,
			EwtTypeCode:   strings.ToUpper(form.EwalletType),
			Amount:        form.Amount,
			WalletAddress: form.To,
			Remark:        form.Remark,
			LangCode:      langCode,
		}

		arrData, err := transfer.PostTransferExchange(tx)

		if err != nil {
			tx.Rollback()
			appG.ResponseError(err)
			return
		}

		tx.Commit()

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "transfer_successful"}, arrData)
	} else {
		transfer := wallet_service.PostTransferStruct{
			MemberId:    entMemberID,
			EwtTypeCode: strings.ToUpper(form.EwalletType),
			Amount:      form.Amount,
			MemberTo:    form.To,
			Remark:      form.Remark,
			LangCode:    langCode,
		}

		arrData, err := transfer.PostTransfer(tx)

		if err != nil {
			tx.Rollback()
			appG.ResponseError(err)
			return
		}

		tx.Commit()

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "transfer_successful"}, arrData)
		return
	}

}

type GetWalletSettingForm struct {
	EwalletTypeCode string `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required;"`
}

func GetWithdrawSetting(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form GetWalletSettingForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	withdrawSetting := wallet_service.GetWithdrawSettingStruct{
		MemberID:    entMemberID,
		EwtTypeCode: form.EwalletTypeCode,
		LangCode:    langCode,
	}

	arrDataReturn, err := withdrawSetting.GetMemberWithdrawSettingv1()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

func GetTransferSetting(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form GetWalletSettingForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	transferSetting := wallet_service.GetTransferSettingStruct{
		MemberID:    entMemberID,
		EwtTypeCode: form.EwalletTypeCode,
		LangCode:    langCode,
	}

	arrDataReturn, err := transferSetting.GetMemberTransferSettingv1()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

func GetTransferExchangeSetting(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form GetWalletSettingForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	transferSetting := wallet_service.GetTransferExchangeSettingStruct{
		MemberID:    entMemberID,
		EwtTypeCode: form.EwalletTypeCode,
		LangCode:    langCode,
	}

	arrDataReturn, err := transferSetting.GetMemberTransferExchangeSettingv1()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

func GetExchangeSetting(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form GetWalletSettingForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	exchangeSetting := wallet_service.GetExchangeSettingStruct{
		MemberID:    entMemberID,
		EwtTypeCode: form.EwalletTypeCode,
		LangCode:    langCode,
	}

	arrDataReturn, err := exchangeSetting.GetMemberExchangeSettingv1()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

// CancelWithdrawForm struct
type CancelWithdrawForm struct {
	DocNo        string `form:"doc_no" json:"doc_no" valid:"Required;"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

func CancelWithdraw(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form CancelWithdrawForm
		err  error
	)

	tx := models.Begin()

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("CancelWithdraw-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
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

	cancelWithdraw := wallet_service.CancelWithdrawStruct{
		MemberId: entMemberID,
		DocNo:    form.DocNo,
	}
	arrData, err := cancelWithdraw.CancelWithdraw(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "cancel_successful"}, arrData)
	return
}

type AddExchangeForm struct {
	Amount            float64 `form:"amount" json:"amount" valid:"Required;"`
	EwalletTypeCode   string  `form:"ewallet_type_code" json:"ewallet_type_code" valid:"Required;"`
	SecondaryPin      string  `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
	EwalletTypeCodeTo string  `form:"ewallet_type_code_to" json:"ewallet_type_code_to" valid:"Required;"`
}

func PostExchange(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AddExchangeForm
		err  error
	)

	tx := models.Begin()

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		tx.Rollback()
		base.LogErrorLog("PostExchange-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
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
		tx.Rollback()
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	//check block all exchange
	exchangeAllblockStatus := member_service.VerifyIfInNetwork(entMemberID, "EX_BLK_ALL")

	if exchangeAllblockStatus {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "network_is_being_blocked_from_exchange"}, nil)
		return
	}

	exchange := wallet_service.PostExchangeStruct{
		MemberId:          entMemberID,
		Amount:            form.Amount,
		EwalletTypeCode:   form.EwalletTypeCode,
		LangCode:          langCode,
		EwalletTypeCodeTo: form.EwalletTypeCodeTo,
	}
	arrData, err := exchange.PostExchange(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "exchange_successful"}, arrData)
	return
}

func GetMemberStrategyBalancev1(c *gin.Context) {
	type GetMemberStrategyBalanceForm struct {
		Platform string `form:"platform" json:"platform" valid:"Required;"`
	}

	var (
		appG = app.Gin{C: c}
		form GetMemberStrategyBalanceForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	strategyBalance := trading_service.GetMemberStrategyBalanceStruct{
		MemberID: entMemberID,
		Platform: form.Platform,
		LangCode: langCode,
	}

	arrDataReturn, err := strategyBalance.GetMemberStrategyBalancev1()

	if err != "" {
		message := app.MsgStruct{
			Msg: err,
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

func GetWalletStatementStrategy(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form WalletTypeStatementStruct
	)

	//validate input
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
	memId := members.EntMemberID

	walStatement := wallet_service.WalletTransactionStructV2{
		MemberID:       memId,
		DateFrom:       form.DateFrom,
		DateTo:         form.DateTo,
		Page:           form.Page,
		LangCode:       langCode,
		WalletTypeCode: strings.ToUpper(form.WalletTypeCode),
		TransType:      form.TransType,
	}

	rst, err := walStatement.WalletStatementStrategyV1()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "failed_to_get_wallet_statement_strategy"}, "")
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, rst)
	return
}

func GetMemberStrategyFuturesBalancev1(c *gin.Context) {
	type GetMemberStrategyBalanceForm struct {
		Platform string `form:"platform" json:"platform"`
	}

	var (
		appG = app.Gin{C: c}
		form GetMemberStrategyBalanceForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	members := u.(*models.EntMemberMembers)
	entMemberID := members.EntMemberID

	if form.Platform == "" {
		form.Platform = "BN" //default as binance
	}

	strategyBalance := trading_service.GetMemberStrategyBalanceStruct{
		MemberID: entMemberID,
		Platform: form.Platform,
		LangCode: langCode,
	}

	arrDataReturn, err := strategyBalance.GetMemberStrategyFuturesBalancev1()

	if err != "" {
		message := app.MsgStruct{
			Msg: err,
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}
