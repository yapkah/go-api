package product_service

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/service/wallet_service"
)

type ExchangeSetting struct {
	// Code string `json:"code"`
	Min            float64                         `json:"min"`
	MultipleOf     float64                         `json:"multiple_of"`
	PaymentSetting []wallet_service.PaymentSetting `json:"payment_setting"`
}

// GetExchangeSetting func
func GetExchangeSetting(memberID int, langCode string) (map[string]interface{}, string) {
	var (
		module                           string = "EXCHANGE"
		prdCurrencyCode, prdGroupGetting string
		paymentType                      string  = "DEFAULT"
		keyinMin                         float64 = 1
		keyinMultipleOf                  float64 = 1
		arrDataReturn                    map[string]interface{}
	)

	// get prd_group_type setting
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: module},
		models.WhereCondFn{Condition: "prd_group_type.status = ?", CondValue: "A"},
	)
	arrGetPrdGroupType, err := models.GetPrdGroupTypeFn(arrCond, "", false)
	if err != nil {
		base.LogErrorLog("GetProductsGroupSetting:GetPrdGroupTypeFn()", map[string]interface{}{"arrCond": arrCond}, err.Error(), true)
		return nil, "something_went_wrong"
	}

	if len(arrGetPrdGroupType) <= 0 {
		base.LogErrorLog("GetProductsGroupSetting:GetPrdGroupTypeFn()", map[string]interface{}{"arrCond": arrCond}, "prd_group_type_not_found", true)
		return nil, "something_went_wrong"
	}

	prdCurrencyCode = arrGetPrdGroupType[0].CurrencyCode
	prdGroupGetting = arrGetPrdGroupType[0].Setting

	if prdGroupGetting != "" {
		arrPrdGroupTypeSetup, errMsg := GetPrdGroupTypeSetup(prdGroupGetting)
		if errMsg != "" {
			return nil, errMsg
		}

		keyinMin = arrPrdGroupTypeSetup.KeyinMin
		keyinMultipleOf = arrPrdGroupTypeSetup.KeyinMultipleOf
	}

	paymentSetting, errMsg := wallet_service.GetPaymentSettingByModule(memberID, module, paymentType, prdCurrencyCode, langCode, true)
	if errMsg != "" {
		return nil, errMsg
	}

	arrDataReturn = map[string]interface{}{
		"products": map[string]interface{}{
			"min":             keyinMin,
			"multiple_of":     keyinMultipleOf,
			"payment_setting": paymentSetting,
		},
	}

	return arrDataReturn, ""
}

// PostExchangeStruct struct
type PostExchangeStruct struct {
	GenTranxDataStatus bool
	Type               string
	MemberID           int
	Amount             float64
	Payments           string
}

// ExchangeCallbackStatus struct
type ExchangeCallbackStatus struct {
	Callback bool
	DocNo    string
}

// // PostExchange func
// func PostExchange(tx *gorm.DB, postStakingStruct PostExchangeStruct, langCode string) (app.MsgStruct, map[string]string, ExchangeCallbackStatus) {
// 	var (
// 		err                              error
// 		errMsg                           string
// 		docType                          string = "EX"
// 		docNo                            string
// 		memberID                         int    = postStakingStruct.MemberID
// 		module                           string = "EXCHANGE"
// 		paymentType                      string = postStakingStruct.Type
// 		prdCurrencyCode, prdGroupGetting string
// 		keyinMin, keyinMultipleOf        float64
// 		exchangeAmount                   float64 = postStakingStruct.Amount
// 		delayTime                        time.Duration
// 	)

// 	// get prd_group_type setting
// 	arrCond := make([]models.WhereCondFn, 0)
// 	arrCond = append(arrCond,
// 		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: module},
// 		models.WhereCondFn{Condition: "prd_group_type.status = ?", CondValue: "A"},
// 	)
// 	arrGetPrdGroupType, err := models.GetPrdGroupTypeFn(arrCond, "", false)
// 	if err != nil {
// 		base.LogErrorLog("PostExchange:GetPrdGroupTypeFn()", err.Error(), map[string]interface{}{"arrCond": arrCond}, true)
// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 	}
// 	if len(arrGetPrdGroupType) <= 0 {
// 		base.LogErrorLog("PostExchange:GetPrdGroupTypeFn()", "prd_group_type_not_found", map[string]interface{}{"arrCond": arrCond}, true)
// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 	}

// 	prdCurrencyCode = arrGetPrdGroupType[0].CurrencyCode
// 	prdGroupGetting = arrGetPrdGroupType[0].Setting
// 	docType = arrGetPrdGroupType[0].DocType

// 	if prdGroupGetting != "" {
// 		arrPrdGroupTypeSetup, errMsg := GetPrdGroupTypeSetup(prdGroupGetting)
// 		if errMsg != "" {
// 			base.LogErrorLog("PostExchange:GetPrdGroupTypeSetup()", map[string]interface{}{"prdGroupGetting": prdGroupGetting}, errMsg, true)
// 			return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 		}

// 		keyinMin = arrPrdGroupTypeSetup.KeyinMin
// 		keyinMultipleOf = arrPrdGroupTypeSetup.KeyinMultipleOf
// 		delayTime = arrPrdGroupTypeSetup.DelayTime
// 	}

// 	// validate if exchange amount is positive
// 	if exchangeAmount <= 0 {
// 		return app.MsgStruct{Msg: "please_enter_valid_amount"}, nil, ExchangeCallbackStatus{}
// 	}

// 	// validate keyinMin
// 	if exchangeAmount < keyinMin {
// 		return app.MsgStruct{Msg: "minimum_amount_must_be_:0", Params: map[string]string{"0": helpers.CutOffDecimal(keyinMin, 8, ".", ",")}}, nil, ExchangeCallbackStatus{}
// 	}

// 	// validate keyinMultipleOf
// 	if !helpers.IsMultipleOf(exchangeAmount, keyinMultipleOf) {
// 		return app.MsgStruct{Msg: "amount_must_be_multiple_of_:0", Params: map[string]string{"0": helpers.CutOffDecimal(keyinMultipleOf, 8, ".", ",")}}, nil, ExchangeCallbackStatus{}
// 	}

// 	extraPaymentInfo := wallet_service.ExtraPaymentInfoStruct{
// 		GenTranxDataStatus: postStakingStruct.GenTranxDataStatus,
// 		EntMemberID:        memberID,
// 		Module:             module,
// 	}

// 	// map wallet payment structure format
// 	paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStructv2(postStakingStruct.Payments, extraPaymentInfo)
// 	if errMsg != "" {
// 		return app.MsgStruct{Msg: errMsg}, nil, ExchangeCallbackStatus{}
// 	}

// 	// check compny balance
// 	arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
// 	arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
// 		models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: 0},
// 		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
// 		models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
// 	)
// 	arrCompanyCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)
// 	if arrCompanyCrypto == nil {
// 		base.LogErrorLog("walletService:PostExchange():GetEntMemberCryptoFn():1", "company_address_not_found", map[string]interface{}{"condition": arrEntMemberCryptoFn}, true)
// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 	}
// 	companyAddress := arrCompanyCrypto.CryptoAddress
// 	companyblkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1("USDS", companyAddress, 0)

// 	curCompanyBalance := companyblkCWalBal.AvailableBalance

// 	if curCompanyBalance < exchangeAmount {
// 		base.LogErrorLog("walletService:PostExchange():GetEntMemberCryptoFn():1", "insufficient_company_balance", map[string]interface{}{"company_balance": companyblkCWalBal}, true)
// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 	}

// 	// tx passed into running doc need to be not affected by beginTransaction becoz blockchain_trans will still insert even if sales generate failed
// 	db := models.GetDB()
// 	docNo, err = models.GetRunningDocNo(docType, db) //get transfer doc no
// 	if err != nil {
// 		base.LogErrorLog("walletService:PostExchange():GetRunningDocNo():1", err.Error(), map[string]interface{}{"docType": docType}, true)
// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 	}
// 	err = models.UpdateRunningDocNo(docType, db) //update transfer doc no
// 	if err != nil {
// 		base.LogErrorLog("walletService:PostExchange():UpdateRunningDocNo():1", err.Error(), map[string]interface{}{"docType": docType}, true)
// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 	}

// 	// insert queue for debit money of USDS
// 	payableAt := base.GetCurrentDateTimeT().Add((delayTime * time.Minute))

// 	// calculate total spent usdt amount
// 	var totalSpentUSDTAmount = 0.00
// 	for _, arrMainWallet := range paymentStruct.MainWallet {
// 		if arrMainWallet.EwalletTypeCode == "USDT" {
// 			totalSpentUSDTAmount += arrMainWallet.Amount
// 		}
// 	}

// 	for _, arrSubWallet := range paymentStruct.SubWallet {
// 		if arrSubWallet.EwalletTypeCode == "USDT" {
// 			totalSpentUSDTAmount += arrSubWallet.Amount
// 		}
// 	}

// 	var arrAddEwtExchange = models.EwtExchangeStruct{
// 		MemberID: memberID,
// 		DocNo:    docNo,
// 		Status:   "PENDING",
// 		Amount:   exchangeAmount,
// 		// EwalletTypeID:  6, // USDS
// 		EwalletTypeID:  24, // internal USDS
// 		TotalUsdtSpent: totalSpentUSDTAmount,
// 		PayableAt:      payableAt,
// 	}

// 	_, err = models.AddEwtExchange(tx, arrAddEwtExchange)
// 	if err != nil {
// 		base.LogErrorLog("walletService:PostExchange():AddEwtExchange():1", err.Error(), map[string]interface{}{"arrAddEwtExchange": arrAddEwtExchange}, true)
// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ExchangeCallbackStatus{}
// 	}

// 	// validate payment with pay amount + deduct wallet
// 	msgStruct, arrData := wallet_service.PaymentProcess(tx, wallet_service.PaymentProcessStruct{
// 		MemberID:        memberID,
// 		PrdCurrencyCode: prdCurrencyCode,
// 		Module:          "EXCHANGE",
// 		Type:            paymentType,
// 		DocNo:           docNo,
// 		Amount:          exchangeAmount,
// 		Payments:        paymentStruct,
// 	}, 0, langCode)
// 	if msgStruct.Msg != "" {
// 		return msgStruct, nil, ExchangeCallbackStatus{}
// 	}

// 	// if no use blockchain wallet, straight call ExchangeCallback() to generate and send exchange_debit signed transaction
// 	exchangeCallbackStatus := ExchangeCallbackStatus{}

// 	directCallback := true // default is allow callback
// 	// loop main and sub wallet, if found got wallet with transaction data then set direct callback to false
// 	for _, arrMainWallet := range paymentStruct.MainWallet {
// 		if arrMainWallet.ConvertedAmount > 0 && arrMainWallet.TransactionData != "" {
// 			directCallback = false
// 		}
// 	}

// 	for _, arrSubWallet := range paymentStruct.SubWallet {
// 		if arrSubWallet.ConvertedAmount > 0 && arrSubWallet.TransactionData != "" {
// 			directCallback = false
// 		}
// 	}

// 	if directCallback {
// 		exchangeCallbackStatus = ExchangeCallbackStatus{Callback: true, DocNo: docNo}
// 	}

// 	return app.MsgStruct{Msg: ""}, arrData, exchangeCallbackStatus
// }

// ExchangeCallback func
// func ExchangeCallback(tx *gorm.DB, docNo string) string {
// 	// get ewt_exchange
// 	arrCond := make([]models.WhereCondFn, 0)
// 	arrCond = append(arrCond,
// 		models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: docNo},
// 		models.WhereCondFn{Condition: "ewt_exchange.status = ?", CondValue: "PENDING"},
// 	)
// 	arrEwtExchange, _ := models.GetEwtExchange(arrCond, "", false)
// 	if arrEwtExchange == nil {
// 		return "ewt_exchange_not_found"
// 	}

// 	var (
// 		memberID      = arrEwtExchange.MemberID
// 		amount        = arrEwtExchange.Amount
// 		ewalletTypeID = arrEwtExchange.EwalletTypeID
// 	)

// 	// insert into ewt_exchange.ewallet_type_id
// 	var saveMemberWalletArrData = wallet_service.SaveMemberWalletStruct{
// 		EntMemberID:      memberID,
// 		EwalletTypeID:    ewalletTypeID,
// 		TotalIn:          amount,
// 		ConversionRate:   1,
// 		ConvertedTotalIn: amount,
// 		TransactionType:  "EXCHANGE",
// 		DocNo:            docNo,
// 		CreatedBy:        strconv.Itoa(memberID),
// 	}
// 	_, err := wallet_service.SaveMemberWallet(tx, saveMemberWalletArrData)
// 	if err != nil {
// 		base.LogErrorLog("walletService:ExchangeCallback():SaveMemberWallet():1", err.Error(), map[string]interface{}{"saveMemberWalletArrData": saveMemberWalletArrData}, true)
// 		return "something_went_wrong"
// 	}

// 	// approve ewt_exchange
// 	errMsg := ExchangeApproveCallback(tx, docNo, true)
// 	if errMsg != "" {
// 		base.LogErrorLog("walletService:ExchangeCallback():ExchangeApproveCallback()", errMsg, map[string]interface{}{"docNo": docNo, "status": true}, true)
// 		return "something_went_wrong"
// 	}

// 	// var ewalletTypeCode string = "USDS"
// 	// get member crypto address
// 	// db := models.GetDB() // no need set begin transaction
// 	// cryptoAddr, err := member_service.ProcessGetMemAddress(db, memberID, "USDS")
// 	// if err != nil {
// 	// 	return "ProcessGetMemAddress()" + err.Error()
// 	// }

// 	// // get company crypto address and private key
// 	// arrCond = make([]models.WhereCondFn, 0)
// 	// arrCond = append(arrCond,
// 	// 	models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: 0},
// 	// 	models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
// 	// 	models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
// 	// )
// 	// arrCompanyCrypto, _ := models.GetEntMemberCryptoFn(arrCond, false)
// 	// if arrCompanyCrypto == nil {
// 	// 	return "company_address_not_found"
// 	// }
// 	// companyAddress := arrCompanyCrypto.CryptoAddress
// 	// companyPrivateKey := arrCompanyCrypto.PrivateKey

// 	// // get wallet contract address
// 	// arrCond = make([]models.WhereCondFn, 0)
// 	// arrCond = append(arrCond,
// 	// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: ewalletTypeCode},
// 	// )
// 	// arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
// 	// if arrEwtSetup == nil {
// 	// 	return "ewt_setup_not_found_for_ewallet_type_code_" + ewalletTypeCode
// 	// }
// 	// contractAddress := arrEwtSetup.ContractAddress

// 	// // get member crypto address
// 	// arrExchangeDebitSetting, errMsg := wallet_service.GetSigningKeySettingByModule("USDS", cryptoAddr, "EXCHANGE_DEBIT")
// 	// if errMsg != "" {
// 	// 	return "GetSigningKeySettingByModule():" + errMsg
// 	// }

// 	// chainID, _ := helpers.ValueToInt(arrExchangeDebitSetting["chain_id"].(string))
// 	// maxGas, _ := helpers.ValueToInt(arrExchangeDebitSetting["max_gas"].(string))
// 	// // send exchange-debit signing key
// 	// arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
// 	// 	TokenType:       ewalletTypeCode,
// 	// 	PrivateKey:      companyPrivateKey,
// 	// 	ContractAddress: contractAddress,
// 	// 	ChainID:         int64(chainID),
// 	// 	FromAddr:        companyAddress,
// 	// 	ToAddr:          cryptoAddr, // this is refer to the buyer address
// 	// 	Amount:          amount,
// 	// 	MaxGas:          uint64(maxGas),
// 	// }

// 	// signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
// 	// if err != nil {
// 	// 	return "ProcecssGenerateSignTransaction():" + err.Error()
// 	// }

// 	// // call sign transaction + insert blockchain_trans
// 	// errMsg, _ = wallet_service.SaveMemberBlochchainWallet(wallet_service.SaveMemberBlochchainWalletStruct{
// 	// 	EntMemberID:      memberID,
// 	// 	EwalletTypeID:    arrEwtSetup.ID,
// 	// 	DocNo:            docNo,
// 	// 	Status:           "P",
// 	// 	TransactionType:  "EXCHANGE",
// 	// 	TransactionData:  signingKey,
// 	// 	TotalIn:          amount,
// 	// 	ConversionRate:   1,
// 	// 	ConvertedTotalIn: amount,
// 	// 	LogOnly:          0,
// 	// })
// 	// if errMsg != "" {
// 	// 	return "SaveMemberBlochchainWallet():" + errMsg
// 	// }

// 	return ""
// }

// ExchangeApproveCallback func
func ExchangeApproveCallback(tx *gorm.DB, docNo string, status bool) string {
	// update ewt_exchange record
	arrEwtExchangeUpdCond := make([]models.WhereCondFn, 0)
	arrEwtExchangeUpdCond = append(arrEwtExchangeUpdCond,
		models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: docNo},
	)

	ewtExchangeStatus := "PAID"
	if !status {
		ewtExchangeStatus = "FAILED"
	}

	updateEwtExchangeColumn := map[string]interface{}{"paid_at": time.Now(), "status": ewtExchangeStatus}
	ewtExchangeErr := models.UpdatesFnTx(tx, "ewt_exchange", arrEwtExchangeUpdCond, updateEwtExchangeColumn, false)
	if ewtExchangeErr != nil {
		return ewtExchangeErr.Error()
	}

	return ""
}
