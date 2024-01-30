package wallet_service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/float"
	"github.com/yapkah/go-api/service/member_service"
)

// PaymentSetting struct
type PaymentSetting struct {
	Type       string                    `json:"type"`
	MainWallet []IndividualWalletSetting `json:"main_wallet"`
	SubWallet  []IndividualWalletSetting `json:"sub_wallet"`
}

// IndividualWalletSetting struct
type IndividualWalletSetting struct {
	EwalletTypeCode      string        `json:"ewallet_type_code"`
	EwalletTypeName      string        `json:"ewallet_type_name"`
	EwalletIconURL       string        `json:"ewallet_icon_url"`
	CurrencyCode         string        `json:"currency_code"`
	DecimalPoints        int           `json:"decimal_points"`
	AvailableBalance     float64       `json:"available_balance"`
	AvailableConvBalance float64       `json:"available_converted_balance"`
	ConversionRate       float64       `json:"conversion_rate"`
	Remark               string        `json:"remark"`
	MinPayPerc           int           `json:"min_pay_perc"`
	MaxPayPerc           int           `json:"max_pay_perc"`
	SigningKeyMethod     string        `json:"signing_key_method"`
	SigningKeySetting    []interface{} `json:"signing_key_setting"`
}

// GetPaymentSettingByModule func
func GetPaymentSettingByModule(memberID int, module, paymentType, prdCurrencyCode, langCode string, includeTransactionData bool) ([]PaymentSetting, string) {
	var (
		arrPaymentSetting                                      = make([]PaymentSetting, 0)
		arrMainWallet, arrSubWallet                            = make([]IndividualWalletSetting, 0), make([]IndividualWalletSetting, 0)
		availableBalance, availableConvBalance, conversionRate float64
		ewalletIconURL, signingKeyMethod                       string
		signingKeySetting                                      []interface{}
	)

	// get payment type under exchange module
	arrPaymentType, err := models.GetPaymentTypeByModules(module, paymentType)
	if err != nil {
		base.LogErrorLog("walletService:GetPaymentSettingByModule():GetPaymentTypeByModules():1", err.Error(), map[string]interface{}{"module": module, "paymentType": paymentType}, true)
		return nil, "something_went_wrong"
	}

	// foreach types and insert data
	for _, arrPaymentTypeVal := range arrPaymentType {
		arrMainWallet, arrSubWallet = make([]IndividualWalletSetting, 0), make([]IndividualWalletSetting, 0)

		arrSysGeneralPaymentSettingFn := make([]models.WhereCondFn, 0)
		arrSysGeneralPaymentSettingFn = append(arrSysGeneralPaymentSettingFn,
			models.WhereCondFn{Condition: "sys_general_payment_setting.module = ?", CondValue: module},
			models.WhereCondFn{Condition: "sys_general_payment_setting.type = ?", CondValue: arrPaymentTypeVal.Type},
			models.WhereCondFn{Condition: "sys_general_payment_setting.status = ?", CondValue: "A"},
		)
		arrSysGeneralPaymentSetting, err := models.GetGeneralPaymentSettingFn(arrSysGeneralPaymentSettingFn, "", false)
		if err != nil {
			base.LogErrorLog("walletService:GetPaymentSettingByModule():GetGeneralPaymentSettingFn():1", err.Error(), map[string]interface{}{"arrSysGeneralPaymentSettingFn": arrSysGeneralPaymentSettingFn}, true)
			return nil, "something_went_wrong"
		}

		// trim payment setting by condition
		arrSysGeneralPaymentSetting, errMsg := TrimPaymentSettingByCondition(arrSysGeneralPaymentSetting, memberID)
		if errMsg != "" {
			return nil, errMsg
		}

		nonce := -1
		for _, arrSysGeneralPaymentSettingVal := range arrSysGeneralPaymentSetting {
			availableBalance = 0
			availableConvBalance = 0
			conversionRate = 1
			signingKeySetting = nil

			// get balance base on ewt_setup.control
			if arrSysGeneralPaymentSettingVal.Control == "INTERNAL" { // get from internal db
				arrEwtSummaryFn := make([]models.WhereCondFn, 0)
				arrEwtSummaryFn = append(arrEwtSummaryFn,
					models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: memberID},
					models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrSysGeneralPaymentSettingVal.EwalletTypeID},
				)

				arrEwtSummary, err := models.GetEwtSummaryFn(arrEwtSummaryFn, "", false)
				if err != nil {
					base.LogErrorLog("walletService:GetPaymentSettingByModule():GetEwtSummaryFn():1", err.Error(), map[string]interface{}{"arrEwtSummaryFn": arrEwtSummaryFn}, true)
					return nil, "something_went_wrong"
				}

				if len(arrEwtSummary) > 0 {
					availableBalance = arrEwtSummary[0].Balance
					availableConvBalance = arrEwtSummary[0].Balance
				}
			} else if arrSysGeneralPaymentSettingVal.Control == "BLOCKCHAIN" { // get from blockchain
				// get wallet address
				db := models.GetDB() // no need set begin transaction
				cryptoAddr, err := member_service.ProcessGetMemAddress(db, memberID, arrSysGeneralPaymentSettingVal.EwalletTypeCode)
				if err != nil {
					base.LogErrorLog("walletService:GetPaymentSettingByModule():ProcessGetMemAddress():1", err.Error(), map[string]interface{}{"memberID": memberID, "cryptoType": arrSysGeneralPaymentSettingVal.EwalletTypeCode}, true)
					return nil, "something_went_wrong"
				}

				// get wallet balance
				BlkCWalBal := GetBlockchainWalletBalanceByAddressV1(arrSysGeneralPaymentSettingVal.EwalletTypeCode, cryptoAddr, memberID)

				availableBalance = BlkCWalBal.AvailableBalance
				availableConvBalance = BlkCWalBal.AvailableConvBalance

				// get signing key info by coin type and module
				if includeTransactionData {
					transferSigningKeySetting, errMsg := GetSigningKeySettingByModule(arrSysGeneralPaymentSettingVal.EwalletTypeCode, cryptoAddr, module)
					if errMsg != "" {
						return nil, errMsg
					}

					if nonce == -1 {
						nonce = transferSigningKeySetting["nonce"].(int)
					} else {
						nonce++
					}

					// append sign key setting get from other function
					transferSigningKeySetting["nonce"] = nonce
					transferSigningKeySetting["decimal_point"] = arrSysGeneralPaymentSettingVal.BlockchainDecimalPoint
					transferSigningKeySetting["contract_address"] = arrSysGeneralPaymentSettingVal.ContractAddress
					transferSigningKeySetting["is_base"] = false

					if arrSysGeneralPaymentSettingVal.IsBase == 1 {
						transferSigningKeySetting["is_base"] = true
					}

					// initialize the map before writing into it
					if signingKeySetting == nil {
						signingKeySetting = make([]interface{}, 0)
					}
					signingKeySetting = append(signingKeySetting, transferSigningKeySetting)
				} else {
					signingKeyMethod = "transfer"
				}
			}

			// get ewallet icon url
			ewalletIconURL = ""
			if arrSysGeneralPaymentSettingVal.AppSettingList != "" {
				appSettingList := &AppSettingList{}
				err = json.Unmarshal([]byte(arrSysGeneralPaymentSettingVal.AppSettingList), appSettingList)
				if err != nil {
					base.LogErrorLog("walletService:GetPaymentSettingByModule():Unmarshal():1", err.Error(), map[string]interface{}{"appSettingList": arrSysGeneralPaymentSettingVal.AppSettingList}, true)
					return nil, "something_went_wrong"
				}

				ewalletIconURL = appSettingList.EwalletIconURL
			}

			// get conversion rate
			conversionRate, errMsg = GetConversionRateByEwalletTypeCode(memberID, arrSysGeneralPaymentSettingVal.EwalletTypeCode, prdCurrencyCode)
			if errMsg != "" {
				base.LogErrorLog("walletService:GetPaymentSettingByModule():GetConversionRateByEwalletTypeCode()", errMsg, map[string]interface{}{"fromTokenTypeCode": arrSysGeneralPaymentSettingVal.EwalletTypeCode, "toTokenTypeCode": prdCurrencyCode}, true)
				return nil, "something_went_wrong"
			}

			if conversionRate > 0 {
				availableConvBalance = float.Div(availableConvBalance, conversionRate) // changed to mul for nft logic

				availableConvBalance, err = helpers.ValueToFloat(helpers.CutOffDecimal(availableConvBalance, 8, ".", "")) // cut off 8 decimal places
				if err != nil {
					base.LogErrorLog("walletService:GetPaymentSettingByModule():ValueToFloat():1", err.Error(), map[string]interface{}{"availableConvBalance": availableConvBalance}, true)
					return nil, "something_went_wrong"
				}
			}

			curIndividualWalletSetting := IndividualWalletSetting{
				EwalletTypeCode: arrSysGeneralPaymentSettingVal.EwalletTypeCode,
				EwalletTypeName: helpers.TranslateV2(arrSysGeneralPaymentSettingVal.EwalletTypeName, langCode, make(map[string]string)),
				EwalletIconURL:  ewalletIconURL,
				// CurrencyCode:    arrSysGeneralPaymentSettingVal.CurrencyCode,
				CurrencyCode:         helpers.TranslateV2(arrSysGeneralPaymentSettingVal.CurrencyCode, langCode, make(map[string]string)),
				DecimalPoints:        arrSysGeneralPaymentSettingVal.DecimalPoint,
				AvailableBalance:     availableBalance,
				AvailableConvBalance: availableConvBalance,
				ConversionRate:       conversionRate,
				MinPayPerc:           arrSysGeneralPaymentSettingVal.MinPayPerc,
				MaxPayPerc:           arrSysGeneralPaymentSettingVal.MaxPayPerc,
				SigningKeyMethod:     signingKeyMethod,
				SigningKeySetting:    signingKeySetting,
			}

			if arrSysGeneralPaymentSettingVal.Main == 1 {
				// set remark for main wallet with minimum payment
				if arrSysGeneralPaymentSettingVal.MinPayPerc > 0 {
					curIndividualWalletSetting.Remark = fmt.Sprintf("%s %s", helpers.TranslateV2("minimum_payment_:0%_of_:1", langCode, map[string]string{"0": fmt.Sprint(arrSysGeneralPaymentSettingVal.MinPayPerc), "1": helpers.TranslateV2(arrSysGeneralPaymentSettingVal.EwalletTypeName, langCode, nil)}), helpers.TranslateV2("main_payment_extra_info", langCode, map[string]string{}))
				}

				arrMainWallet = append(arrMainWallet, curIndividualWalletSetting)
			} else {
				if arrSysGeneralPaymentSettingVal.MaxPayPerc > 0 {
					curIndividualWalletSetting.Remark = helpers.TranslateV2("maximum_payment_:0%_of_:1", langCode, map[string]string{"0": fmt.Sprint(arrSysGeneralPaymentSettingVal.MaxPayPerc), "1": helpers.TranslateV2(arrSysGeneralPaymentSettingVal.EwalletTypeName, langCode, nil)})
				}

				arrSubWallet = append(arrSubWallet, curIndividualWalletSetting)
			}
		}

		arrPaymentSetting = append(arrPaymentSetting,
			PaymentSetting{Type: arrPaymentTypeVal.Type, MainWallet: arrMainWallet, SubWallet: arrSubWallet},
		)
	}

	return arrPaymentSetting, ""
}

// GetConversionRateByEwalletTypeCode func - (purchase contract eg: fromTokenTypeCode = SA/SP/SEAG, toTokenTypecode = USDT[product currency code])
func GetConversionRateByEwalletTypeCode(memberID int, fromTokenTypeCode, toTokenTypeCode string) (float64, string) {
	var (
		conversionRate        float64 = 1
		defaultConversionRate float64 = 1
		err                   error
	)

	if strings.HasPrefix(toTokenTypeCode, "NFT") && len(toTokenTypeCode) > 3 {
		if fromTokenTypeCode == "USDT" {
			conversionRate, err = base.GetLatestPriceMovementByTokenType(toTokenTypeCode)
			if err != nil {
				return defaultConversionRate, err.Error()
			}
		}
	}

	// fmt.Println("fromTokenTypeCode:", fromTokenTypeCode, "toTokenTypeCode:", toTokenTypeCode, "conversionRate:", conversionRate)
	return conversionRate, ""
}

// GetSigningKeySettingByModule func
func GetSigningKeySettingByModule(ewalletTypeCode, cryptoAddr, module string) (map[string]interface{}, string) {
	var arrSigningKeySetting map[string]interface{}

	arrModuleSettingID := map[string]interface{}{
		"CONTRACT":         "contract_sign_key_setting",
		"CONTRACT_TOPUP":   "contract_sign_key_setting",
		"EXCHANGE":         "exchange_sign_key_setting",
		"EXCHANGE_DEBIT":   "exchange_debit_sign_key_setting",
		"TRADING":          "trading_sign_key_setting",
		"STAKING":          "staking_sign_key_setting",
		"STAKING_APPROVED": "staking_approve_sign_key_setting",
		"TRANSFER":         "transfer_sign_key_setting",
		"UNSTAKE":          "unstake_sign_key_setting",
		"TRANSFER_BATCH":   "transfer_sign_key_setting",
		"WITHDRAW_POOL":    "withdraw_pool_sign_key_setting",
		"P2P":              "p2p_sign_key_setting",
		"STAKELIGA":        "stakeliga_sign_key_setting",
		"UNSTAKELALIGA":    "unstakelaliga_sign_key_setting",
		// "WITHDRAW":         "withdraw_sign_key_setting",
		"MINING":     "contract_sign_key_setting",
		"MINING_BZZ": "contract_sign_key_setting",
	}

	settingInterface := arrModuleSettingID[strings.ToUpper(module)]
	if settingInterface == nil {
		base.LogErrorLog("GetSigningKeySettingByModule-settingInterface_failed", "invalid_module_"+module, "", true)
		return nil, "something_went_wrong"
	}

	settingID := settingInterface.(string)

	arrGeneralSetup, err := models.GetSysGeneralSetupByID(settingID)
	if err != nil {
		base.LogErrorLog("GetSigningKeySettingByModule-GetSysGeneralSetupByID1_failed", err.Error(), settingID+"_not_found", true)
		return nil, "something_went_wrong"
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("GetSigningKeySettingByModule-GetSysGeneralSetupByID2_failed", settingID+"_not_found", arrGeneralSetup, true)
		return nil, "something_went_wrong"
	}

	if arrGeneralSetup.InputType1 == "0" {
		return nil, ""
	}

	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(arrGeneralSetup.InputValue1), &arrSigningKeySetting)

	if arrSigningKeySetting != nil {
		if helpers.IsArray(arrSigningKeySetting) == true {
			// is setting that will appear different for different coin.
			if arrSigningKeySetting[ewalletTypeCode] == nil {
				base.LogErrorLog("GetSigningKeySettingByModule-arrSigningKeySetting[ewalletTypeCode]_failed", settingID+"_setting_ewallet_type_code_"+ewalletTypeCode+"not_found", arrSigningKeySetting, true)
				return nil, "something_went_wrong"
			}

			arrSigningKeySetting = arrSigningKeySetting[ewalletTypeCode].(map[string]interface{})
		}

		if module != "TRANSFER_BATCH" {
			nonce, err := GetTransactionNonceViaAPI(cryptoAddr)
			if err != nil {
				base.LogErrorLog("GetSigningKeySettingByModule-GetTransactionNonceViaAPI_failed", err.Error(), cryptoAddr, true)
				return nil, "something_went_wrong"
			}
			arrSigningKeySetting["nonce"] = nonce
		}

		if module != "STAKING" && module != "STAKING_APPROVED" && module != "UNSTAKE" && module != "UNSTAKELALIGA" {
			toAddr, errMsg := GetCompanyAddress(ewalletTypeCode)
			if errMsg != "" {
				base.LogErrorLog("GetSigningKeySettingByModule-GetCompanyAddress-failed", errMsg, ewalletTypeCode, true)
				return nil, "something_went_wrong"
			}
			arrSigningKeySetting["to_address"] = toAddr
		}

	}

	return arrSigningKeySetting, ""
}

// PaymentStruct struct
type PaymentStruct struct {
	MainWallet []IndividualPaymentStruct `json:"main_wallet"`
	SubWallet  []IndividualPaymentStruct `json:"sub_wallet"`
}

// ConvertPaymentInputToStruct struct
func ConvertPaymentInputToStruct(payments string) (PaymentStruct, string) {
	paymentStruct := &PaymentStruct{}
	err := json.Unmarshal([]byte(payments), paymentStruct)
	if err != nil {
		return *paymentStruct, "invalid_payment_structure"
	}

	for arrMainWalletK, arrMainWalletV := range paymentStruct.MainWallet {
		// slice paid amount to x decimals points
		decimalPlaces := GetDecimalPlacesByEwalletTypeCode(arrMainWalletV.EwalletTypeCode)
		mainWalletAmount, err := helpers.ValueToFloat(helpers.CutOffDecimal(arrMainWalletV.Amount, uint(decimalPlaces), ".", ""))
		// mainWalletAmountStr := helpers.CutOffDecimal(arrMainWalletV.Amount, 8, ".", "")
		// fmt.Println("raw:", arrMainWalletV.Amount, "str:", mainWalletAmountStr, "float64:", mainWalletAmount)

		if err != nil {
			base.LogErrorLog("walletService:ConvertPaymentInputToStruct()", "ValueToFloat():1", err.Error(), true)
			return *paymentStruct, "something_went_wrong"
		}

		paymentStruct.MainWallet[arrMainWalletK].Amount = mainWalletAmount
	}

	for arrSubWalletK, arrSubWalletV := range paymentStruct.SubWallet {
		// slice paid amount to x decimals points
		decimalPlaces := GetDecimalPlacesByEwalletTypeCode(arrSubWalletV.EwalletTypeCode)
		subWalletAmount, err := helpers.ValueToFloat(helpers.CutOffDecimal(arrSubWalletV.Amount, uint(decimalPlaces), ".", ""))
		// subWalletAmountStr := helpers.CutOffDecimal(arrSubWalletV.Amount, 8, ".", "")
		// fmt.Println("raw:", arrSubWalletV.Amount, "str:", subWalletAmountStr, "float64:", subWalletAmount)
		if err != nil {
			base.LogErrorLog("walletService:ConvertPaymentInputToStruct()", "ValueToFloat():2", err.Error(), true)
			return *paymentStruct, "something_went_wrong"
		}

		paymentStruct.SubWallet[arrSubWalletK].Amount = subWalletAmount
	}

	return *paymentStruct, ""
}

// PaymentProcessStruct struct
type PaymentProcessStruct struct {
	MemberID                                     int
	PrdCurrencyCode, Module, Type, DocNo, Remark string
	Amount                                       float64
	Payments                                     PaymentStruct
}

// EwalletPaymentProcessStruct struct define payment for internal + blockchain payment
type EwalletPaymentProcessStruct struct {
	InternalWalletPayment   []SaveMemberWalletStruct
	BlockchainWalletPayment []BlockchainWalletStruct
}

// BlockchainWalletStruct struct define payment structure needed for blockchain payment
type BlockchainWalletStruct struct {
	EntMemberID, EwalletTypeID, DecimalPlaces                       int
	Amount, ConversionRate, ConvertedAmount                         float64
	EwalletTypeCode, CurrencyCode, TransactionType, TransactionData string
}

// PaymentProcess func to validate + deduct from ewallet
func PaymentProcess(tx *gorm.DB, paymentProcessData PaymentProcessStruct, conversionRate float64, langCode string) (app.MsgStruct, map[string]string) {
	// validate wallet payment
	msgStruct, ewalletPaymentProcessStruct := ValidatePayment(paymentProcessData, conversionRate, langCode)
	if msgStruct.Msg != "" {
		return msgStruct, nil
	}

	var arrData map[string]string
	var paymentType string

	// process internal wallet payment
	for _, internalWalletPayment := range ewalletPaymentProcessStruct.InternalWalletPayment {
		internalWalletPayment.DocNo = paymentProcessData.DocNo
		internalWalletPayment.Remark = paymentProcessData.Remark

		// fmt.Println("deduct_wallet:", internalWalletPayment)
		_, err := SaveMemberWallet(tx, internalWalletPayment)
		if err != nil {
			base.LogErrorLog("walletService:PaymentProcess()", "SaveMemberWallet():1", err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil
		}

		if paymentType != "" {
			paymentType += ", "
		}

		paymentType = fmt.Sprintf("%s%s %s", paymentType, helpers.CutOffDecimal(internalWalletPayment.TotalOut, uint(internalWalletPayment.DecimalPlaces), ".", ","), helpers.TranslateV2(internalWalletPayment.CurrencyCode, langCode, make(map[string]string)))
	}

	// process blockchain wallet payment
	for _, blockchainWalletPayment := range ewalletPaymentProcessStruct.BlockchainWalletPayment {
		// deduct wallet from blockchain
		errMsg, arrBlockchainWalletData := SaveMemberBlochchainWallet(SaveMemberBlochchainWalletStruct{
			EntMemberID:       blockchainWalletPayment.EntMemberID,
			EwalletTypeID:     blockchainWalletPayment.EwalletTypeID,
			DocNo:             paymentProcessData.DocNo,
			Status:            "P",
			TransactionType:   blockchainWalletPayment.TransactionType,
			TransactionData:   blockchainWalletPayment.TransactionData,
			TotalOut:          blockchainWalletPayment.Amount,
			ConversionRate:    blockchainWalletPayment.ConversionRate,
			ConvertedTotalOut: blockchainWalletPayment.ConvertedAmount,
			LogOnly:           0,
		})

		if errMsg != "" {
			return app.MsgStruct{Msg: errMsg}, nil
		}

		if paymentType != "" {
			paymentType += ", "
		}
		paymentType = fmt.Sprintf("%s%s %s", paymentType, helpers.CutOffDecimal(blockchainWalletPayment.Amount, uint(blockchainWalletPayment.DecimalPlaces), ".", ","), helpers.TranslateV2(blockchainWalletPayment.CurrencyCode, langCode, make(map[string]string)))

		// initialize the map before writing into it
		if arrData == nil {
			arrData = make(map[string]string)
		}
		arrData["hash_value"] = arrBlockchainWalletData["hashValue"]
	}

	// initialize the map before writing into it
	if arrData == nil {
		arrData = make(map[string]string)
	}
	arrData["payment_summary"] = paymentType

	return app.MsgStruct{Msg: ""}, arrData
}

// IndividualPaymentStruct struct
type IndividualPaymentStruct struct {
	EwalletTypeCode string  `json:"ewallet_type_code" form:"ewallet_type_code"`
	Amount          float64 `json:"amount" form:"amount"`
}

// ValidatePayment func that validate PaymentStruct
func ValidatePayment(paymentProcessData PaymentProcessStruct, conversionRate float64, langCode string) (app.MsgStruct, EwalletPaymentProcessStruct) {
	var (
		mainWallet                                              int = 0
		paidAmount, unpaidAmount                                float64
		toPaidAmount                                            float64 = paymentProcessData.Amount
		mainIndividualPaymentStruct, subIndividualPaymentStruct IndividualPaymentStruct
		prdCurrencyCode                                         string = paymentProcessData.PrdCurrencyCode
		module                                                  string = paymentProcessData.Module
		settingType                                             string = paymentProcessData.Type
		msgStruct                                               app.MsgStruct
		arrReturnData                                           EwalletPaymentProcessStruct
	)

	// validate main wallet
	for _, arrMainWallet := range paymentProcessData.Payments.MainWallet {
		mainIndividualPaymentStruct = arrMainWallet
		if mainIndividualPaymentStruct.EwalletTypeCode == "" {
			return app.MsgStruct{Msg: "main_wallet_cannot_be_empty"}, EwalletPaymentProcessStruct{}
		}

		msgStruct, arrReturnData = ValidatePaymentRules(mainIndividualPaymentStruct, prdCurrencyCode, module, settingType, paymentProcessData.MemberID, toPaidAmount, arrReturnData, conversionRate, "MAIN", langCode)
		if msgStruct.Msg != "" {
			return msgStruct, EwalletPaymentProcessStruct{}
		}
		// fmt.Println(mainIndividualPaymentStruct.Amount)
		mainWallet = 1
		paidAmount = float.Add(paidAmount, mainIndividualPaymentStruct.Amount)
	}

	// validate sub wallet
	for _, arrSubWallet := range paymentProcessData.Payments.SubWallet {
		subIndividualPaymentStruct = arrSubWallet
		if subIndividualPaymentStruct.EwalletTypeCode != "" { // validate if sub wallet provided, if yes then validate
			msgStruct, arrReturnData = ValidatePaymentRules(subIndividualPaymentStruct, prdCurrencyCode, module, settingType, paymentProcessData.MemberID, toPaidAmount, arrReturnData, conversionRate, "SUB", langCode)
			if msgStruct.Msg != "" {
				return msgStruct, EwalletPaymentProcessStruct{}
			}
			// fmt.Println(subIndividualPaymentStruct.Amount)
			paidAmount = float.Add(paidAmount, subIndividualPaymentStruct.Amount)
		}
	}

	// check total_paid amount
	// fmt.Println("toPaidAmt:", toPaidAmount, "paidAmt:", paidAmount)
	unpaidAmount = float.Sub(toPaidAmount, paidAmount)
	if unpaidAmount != 0 {
		if unpaidAmount < 0 {
			return app.MsgStruct{Msg: "excessive_payment"}, EwalletPaymentProcessStruct{}
		} else {
			return app.MsgStruct{Msg: "insufficient_payment"}, EwalletPaymentProcessStruct{}
		}
	}

	if mainWallet != 1 {
		return app.MsgStruct{Msg: "must_pay_with_main_wallet"}, EwalletPaymentProcessStruct{}
	}

	return app.MsgStruct{Msg: ""}, arrReturnData
}

// ValidatePaymentRules func that perform basic individual wallet payment rule checking
func ValidatePaymentRules(payment IndividualPaymentStruct, prdCurrencyCode, module, settingType string, memberID int, toPaidAmount float64, arrReturnData EwalletPaymentProcessStruct, definedConversionRate float64, paymentWalletType, langCode string) (app.MsgStruct, EwalletPaymentProcessStruct) {
	var (
		paidAmount                                            float64 = payment.Amount
		convertedPaidAmount                                   float64 = payment.Amount
		balance, conversionRate, minPaidAmount, maxPaidAmount float64
		ewalletTypeID                                         int
		control, ewalletTypeCode                              string
		main                                                  int = 1
	)

	if paymentWalletType == "SUB" {
		main = 0
	}

	// get payment setting
	arrSysGeneralPaymentSettingFn := make([]models.WhereCondFn, 0)
	arrSysGeneralPaymentSettingFn = append(arrSysGeneralPaymentSettingFn,
		models.WhereCondFn{Condition: "sys_general_payment_setting.module = ?", CondValue: module},
		models.WhereCondFn{Condition: "sys_general_payment_setting.type = ?", CondValue: settingType},
		models.WhereCondFn{Condition: "sys_general_payment_setting.main = ?", CondValue: main},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: payment.EwalletTypeCode},
		models.WhereCondFn{Condition: "sys_general_payment_setting.status = ?", CondValue: "A"},
	)

	arrSysGeneralPaymentSetting, err := models.GetGeneralPaymentSettingFn(arrSysGeneralPaymentSettingFn, "", false)
	if err != nil {
		base.LogErrorLog("walletService:ValidatePaymentRules():GetGeneralPaymentSettingFn():1", err.Error(), map[string]interface{}{"condition": arrSysGeneralPaymentSettingFn}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
	}

	// trim payment setting by condition
	arrSysGeneralPaymentSetting, errMsg := TrimPaymentSettingByCondition(arrSysGeneralPaymentSetting, memberID)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, EwalletPaymentProcessStruct{}
	}

	if len(arrSysGeneralPaymentSetting) <= 0 {
		return app.MsgStruct{Msg: "invalid_:0_wallet_ewallet_type_code_:1", Params: map[string]string{"0": helpers.Translate(strings.ToLower(paymentWalletType), langCode), "1": helpers.TranslateV2(payment.EwalletTypeCode, langCode, make(map[string]string))}}, EwalletPaymentProcessStruct{}
	}

	// validate mix max
	minPaidAmount = float.Mul(toPaidAmount, float.Div(float64(arrSysGeneralPaymentSetting[0].MinPayPerc), 100))
	if paidAmount < minPaidAmount {
		return app.MsgStruct{Msg: "please_pay_minimum_amount_of_:0_for_:1", Params: map[string]string{"0": helpers.CutOffDecimal(minPaidAmount, 8, ".", ","), "1": helpers.TranslateV2(payment.EwalletTypeCode, langCode, make(map[string]string))}}, EwalletPaymentProcessStruct{}
	}

	maxPaidAmount = float.Mul(toPaidAmount, float.Div(float64(arrSysGeneralPaymentSetting[0].MaxPayPerc), 100))
	if paidAmount > maxPaidAmount {
		return app.MsgStruct{Msg: "maximum_payment_for_:0_is_:1", Params: map[string]string{"0": helpers.TranslateV2(payment.EwalletTypeCode, langCode, make(map[string]string)), "1": helpers.CutOffDecimal(maxPaidAmount, 8, ".", ",")}}, EwalletPaymentProcessStruct{}
	}

	// validate balance
	control = arrSysGeneralPaymentSetting[0].Control
	ewalletTypeID = arrSysGeneralPaymentSetting[0].EwalletTypeID
	ewalletTypeCode = arrSysGeneralPaymentSetting[0].EwalletTypeCode

	if control == "INTERNAL" {
		arrEwtSummaryFn := make([]models.WhereCondFn, 0)
		arrEwtSummaryFn = append(arrEwtSummaryFn,
			models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: memberID},
			models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: ewalletTypeID},
		)

		arrEwtSummary, err := models.GetEwtSummaryFn(arrEwtSummaryFn, "", false)
		if err != nil {
			base.LogErrorLog("walletService:ValidatePaymentRules():GetEwtSummaryFn():1", err.Error(), map[string]interface{}{"condition": arrEwtSummaryFn}, true)
			return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		}

		if len(arrEwtSummary) > 0 {
			balance = arrEwtSummary[0].Balance
		}

		if paidAmount > 0 {
			// get conversion rate
			conversionRate, errMsg = GetConversionRateByEwalletTypeCode(memberID, arrSysGeneralPaymentSetting[0].EwalletTypeCode, prdCurrencyCode)
			if errMsg != "" {
				base.LogErrorLog("walletService:ValidatePaymentRules():GetConversionRateByEwalletTypeCode()", errMsg, map[string]interface{}{"fromTokenTypeCode": arrSysGeneralPaymentSetting[0].EwalletTypeCode, "toTokenTypeCode": prdCurrencyCode}, true)
				return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
			}

			// convertedPaidAmount = float.Div(paidAmount, conversionRate)
			convertedPaidAmount = float.Mul(paidAmount, conversionRate) // changed to mul for nft logic

			convertedPaidAmount, err = helpers.ValueToFloat(helpers.CutOffDecimal(convertedPaidAmount, 8, ".", "")) // cut off 8 decimal places
			if err != nil {
				base.LogErrorLog("walletService:ValidatePaymentRules():ValueToFloat():1", err.Error(), map[string]interface{}{"convertedPaidAmount": convertedPaidAmount}, true)
				return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
			}

			arrReturnData.InternalWalletPayment = append(arrReturnData.InternalWalletPayment,
				SaveMemberWalletStruct{
					EntMemberID:       memberID,
					EwalletTypeID:     ewalletTypeID,
					EwalletTypeCode:   ewalletTypeCode,
					CurrencyCode:      arrSysGeneralPaymentSetting[0].CurrencyCode,
					TotalOut:          convertedPaidAmount, // in wallet currency
					ConversionRate:    conversionRate,
					ConvertedTotalOut: paidAmount, // in payment currency
					TransactionType:   module,
					CreatedBy:         strconv.Itoa(memberID),
					DecimalPlaces:     arrSysGeneralPaymentSetting[0].DecimalPoint,
				},
			)
		}
	} else if control == "BLOCKCHAIN" {
		base.LogErrorLog("walletService:ValidatePaymentRules()", "invalid_control", map[string]interface{}{"paymentSetting": arrSysGeneralPaymentSetting[0]}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		// if paidAmount > 0 {
		// check if got transaction_data
		// if payment.ConvertedAmount == 0 {
		// 	base.LogErrorLog("walletService:ValidatePaymentRules()", "converted_amount_cannot_be_empty_for_blockchain_wallet", map[string]interface{}{"paymentStructure": payment}, true)
		// 	return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		// }

		// if payment.TransactionData == "" {
		// 	base.LogErrorLog("walletService:ValidatePaymentRules()", "transaction_data_cannot_be_empty_for_blockchain_wallet", map[string]interface{}{"paymentStructure": payment}, true)
		// 	return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		// }

		// db := models.GetDB() // no need set begin transaction
		// cryptoAddr, err := member_service.ProcessGetMemAddress(db, memberID, ewalletTypeCode)
		// if err != nil {
		// 	base.LogErrorLog("walletService:ValidatePaymentRules():ProcessGetMemAddress():1", err.Error(), map[string]interface{}{"memberID": memberID, "ewalletTypeCode": ewalletTypeCode}, true)
		// 	return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		// }

		// BlkCWalBal := GetBlockchainWalletBalanceByAddressV1(ewalletTypeCode, cryptoAddr, memberID)

		// balance = BlkCWalBal.AvailableBalance // balance in liga/sec

		// // get conversion rate
		// conversionRate, errMsg = GetConversionRateByEwalletTypeCode(memberID, arrSysGeneralPaymentSetting[0].EwalletTypeCode, prdCurrencyCode)
		// if errMsg != "" {
		// 	base.LogErrorLog("walletService:ValidatePaymentRules():GetConversionRateByEwalletTypeCode()", errMsg, map[string]interface{}{"fromTokenTypeCode": arrSysGeneralPaymentSetting[0].EwalletTypeCode, "toTokenTypeCode": prdCurrencyCode}, true)
		// 	return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		// }

		// if definedConversionRate > 0 {
		// 	conversionRate = definedConversionRate
		// }

		// // convertedPaidAmount = float.Div(paidAmount, conversionRate)
		// convertedPaidAmount, _ = decimal.NewFromFloat(paidAmount).Div(decimal.NewFromFloat(conversionRate)).Float64()
		// convertedPaidAmount, err = helpers.ValueToFloat(helpers.CutOffDecimal(convertedPaidAmount, 8, ".", "")) // cut off 8 decimal places
		// if err != nil {
		// 	base.LogErrorLog("walletService:ValidatePaymentRules():ValueToFloat():1", err.Error(), map[string]interface{}{"convertedPaidAmount": convertedPaidAmount}, true)
		// 	return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		// }

		// fmt.Println("paidAmount:", paidAmount, "conversionRate:", conversionRate, "before_cut:", float.Div(paidAmount, conversionRate), "calculated:", convertedPaidAmount, "passed_in:", payment.ConvertedAmount)
		// if convertedPaidAmount != payment.ConvertedAmount {
		// 	base.LogErrorLog("walletService:ValidatePaymentRules()", "invalid_converted_amount", map[string]interface{}{"paymentStructure": payment}, true)
		// 	return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
		// }

		// arrReturnData.BlockchainWalletPayment = append(arrReturnData.BlockchainWalletPayment,
		// 	BlockchainWalletStruct{
		// 		EntMemberID:     memberID,
		// 		EwalletTypeID:   ewalletTypeID,
		// 		EwalletTypeCode: ewalletTypeCode,
		// 		CurrencyCode:    arrSysGeneralPaymentSetting[0].CurrencyCode,
		// 		Amount:          convertedPaidAmount, // in wallet currency
		// 		ConversionRate:  conversionRate,
		// 		ConvertedAmount: paidAmount, // in payment currency
		// 		TransactionType: module,
		// 		TransactionData: payment.TransactionData,
		// 		DecimalPlaces:   arrSysGeneralPaymentSetting[0].DecimalPoint,
		// 	},
		// )
		// }
	} else {
		base.LogErrorLog("walletService:ValidatePaymentRules()", "invalid_control", map[string]interface{}{"paymentSetting": arrSysGeneralPaymentSetting[0]}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, EwalletPaymentProcessStruct{}
	}

	if balance < convertedPaidAmount {
		return app.MsgStruct{Msg: ":0_insufficient_balance", Params: map[string]string{"0": helpers.TranslateV2(payment.EwalletTypeCode, langCode, make(map[string]string))}}, EwalletPaymentProcessStruct{}
	}

	return app.MsgStruct{Msg: ""}, arrReturnData
}

// PaymentConditionFlag struct
type PaymentConditionFlag struct {
	FirstContract, NotFirstContract, NotBlockNetworkA, BlockNetworkA int // 0:not yet check,1:checked true,2:checked false
}

// TrimPaymentSettingByCondition func remove those payment option that does not meet condition for this member
func TrimPaymentSettingByCondition(arrSysGeneralPaymentSetting []models.SysGeneralPaymentSetting, memberID int) ([]models.SysGeneralPaymentSetting, string) {
	var (
		arrPaymentSetting    = make([]models.SysGeneralPaymentSetting, 0)
		paymentConditionFlag PaymentConditionFlag // to store condition checked status, if checked will not run checking logic again for the next payment setting (not suitable for condition that may affect by different ewallet type)
	)

	for _, arrSysGeneralPaymentSettingV := range arrSysGeneralPaymentSetting {
		arrCondition := helpers.Explode(arrSysGeneralPaymentSettingV.Condition, ",")

		// payment option only available for first purchase
		if helpers.StringInSlice("FIRST_CONTRACT", arrCondition) == true {
			if paymentConditionFlag.FirstContract == 0 { // not checked
				status, err := arrSysGeneralPaymentSettingV.FirstContract(memberID)
				if err != nil {
					base.LogErrorLog("walletService:TrimPaymentSettingByCondition()", "FirstContract():1", err.Error(), true)
					return nil, "something_went_wrong"
				}

				if !status {
					paymentConditionFlag.FirstContract = 2 // set to checked:false
					continue
				} else {
					paymentConditionFlag.FirstContract = 1 // set to checked:true
				}
			} else if paymentConditionFlag.FirstContract == 2 { // checked:false
				continue
			}
		}

		// payment option only available for not first purchase
		if helpers.StringInSlice("NOT_FIRST_CONTRACT", arrCondition) == true {
			if paymentConditionFlag.NotFirstContract == 0 { // not checked
				status, err := arrSysGeneralPaymentSettingV.NotFirstContract(memberID)
				if err != nil {
					base.LogErrorLog("walletService:TrimPaymentSettingByCondition()", "NotFirstContract():1", err.Error(), true)
					return nil, "something_went_wrong"
				}

				if !status {
					paymentConditionFlag.NotFirstContract = 2 // set to checked:false
					continue
				} else {
					paymentConditionFlag.NotFirstContract = 1 // set to checked:true
				}
			} else if paymentConditionFlag.NotFirstContract == 2 { // checked:false
				continue
			}
		}

		// payment option only available for those not in blocked network group LIGA_A
		if helpers.StringInSlice("NOT_IN_NETWORK_LIGA_A", arrCondition) == true {
			if paymentConditionFlag.NotBlockNetworkA == 0 { // not checked
				status := member_service.VerifyIfInNetwork(memberID, "LIGA_A")

				if !status {
					paymentConditionFlag.NotBlockNetworkA = 1 // set to checked:true
				} else {
					paymentConditionFlag.NotBlockNetworkA = 2 // set to checked:false
					continue
				}
			} else if paymentConditionFlag.NotBlockNetworkA == 2 { // checked:false
				continue
			}
		}

		// payment option only available for those in blocked network group LIGA_A
		if helpers.StringInSlice("IN_NETWORK_LIGA_A", arrCondition) == true {
			if paymentConditionFlag.BlockNetworkA == 0 && paymentConditionFlag.NotBlockNetworkA == 0 { // not checked
				status := member_service.VerifyIfInNetwork(memberID, "LIGA_A")

				if status {
					paymentConditionFlag.BlockNetworkA = 1 // set to checked:true
				} else {
					paymentConditionFlag.BlockNetworkA = 2 // set to checked:false
					continue
				}
			} else if paymentConditionFlag.BlockNetworkA == 2 || paymentConditionFlag.NotBlockNetworkA == 1 { // checked:false
				continue
			}
		}

		// payment option only available if got balance
		if helpers.StringInSlice("WITH_BALANCE", arrCondition) == true {
			if arrSysGeneralPaymentSettingV.Control == "INTERNAL" {
				arrEwtSummaryFn := make([]models.WhereCondFn, 0)
				arrEwtSummaryFn = append(arrEwtSummaryFn,
					models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: memberID},
					models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrSysGeneralPaymentSettingV.EwalletTypeID},
				)

				arrEwtSummary, err := models.GetEwtSummaryFn(arrEwtSummaryFn, "", false)
				if err != nil {
					return nil, err.Error()
				}

				availableBalance := 0.00
				if len(arrEwtSummary) > 0 {
					availableBalance = arrEwtSummary[0].Balance
				}

				if availableBalance <= 0 {
					continue
				}
			}
		}

		arrPaymentSetting = append(arrPaymentSetting, arrSysGeneralPaymentSettingV)
	}

	return arrPaymentSetting, ""
}

// TransactionData struct
type TransactionData struct {
	ConvertedAmount float64 `json:"converted_amount" form:"converted_amount"`
	TransactionData string  `json:"transaction_data" form:"transaction_data"`
}

// ConvertTransactionDataToStruct struct
func ConvertTransactionDataToStruct(transactionDataInput string) (TransactionData, string) {
	transactionData := &TransactionData{}
	err := json.Unmarshal([]byte(transactionDataInput), transactionData)
	if err != nil {
		return *transactionData, "invalid_transaction_data_structure"
	}

	return *transactionData, ""
}

type TradingSigningKeySettingRst struct {
	MethodID string `json:"method_id"`
	ChainID  string `json:"chain_id"`
	GasPrice string `json:"gas_price"`
	MaxGas   string `json:"max_gas"`
}

// ProcessGetTradingSigningKeySetting func
func ProcessGetTradingSigningKeySetting() (*TradingSigningKeySettingRst, error) {
	settingID := "trading_sign_key_setting"
	arrGeneralSetup, err := models.GetSysGeneralSetupByID(settingID)
	if err != nil {
		base.LogErrorLog("ProcessGetTradingSigningKeySetting_GetSysGeneralSetupByID_failed", err.Error(), settingID, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("ProcessGetTradingSigningKeySetting_missing_trading_sign_key_setting_failed", settingID+"_not_found", settingID, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrGeneralSetup.InputType1 == "0" {
		base.LogErrorLog("ProcessGetTradingSigningKeySetting_trading_sign_key_setting_is_off", "trading_sign_key_setting_for_trading_is_off", settingID, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrDataReturn := TradingSigningKeySettingRst{}

	// Unmarshal or Decode the JSON to the interface.
	err = json.Unmarshal([]byte(arrGeneralSetup.InputValue1), &arrDataReturn)
	if err != nil {
		base.LogErrorLog("ProcessGetTradingSigningKeySetting_json_decode_failed", err.Error(), arrGeneralSetup.InputValue1, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return &arrDataReturn, nil
}

type EwtTransactions struct {
	EwalletTypeID   int
	EwalletTypeName string
	CurrencyCode    string
	ConversionRate  float64
	PaidAmount      float64
	ReceivedAmount  float64
	Hash            string
}

// GetEwtTransactionsByDocNo func
func GetEwtTransactionsByDocNo(arrDocNo []string) []EwtTransactions {
	var (
		arrEwtTransactions []EwtTransactions
		arrEwtSetupLibrary = map[int]models.EwtSetup{}
	)

	// get from ewt_detail
	var arrEwtDetailsFn = make([]models.WhereCondFn, 0)
	arrEwtDetailsFn = append(arrEwtDetailsFn,
		models.WhereCondFn{Condition: " ewt_detail.doc_no IN(?) ", CondValue: arrDocNo},
	)
	arrEwtDetails, _ := models.GetEwtDetailFn(arrEwtDetailsFn, false)

	for _, arrEwtDetailsV := range arrEwtDetails {
		var (
			ewalletTypeID   = arrEwtDetailsV.EwalletTypeID
			ewalletTypeName = ""
			currencyCode    = ""
		)

		// get ewt_setup details if not grabbed before
		if _, ok := arrEwtSetupLibrary[ewalletTypeID]; !ok {
			arrEwtSetupFn := make([]models.WhereCondFn, 0)
			arrEwtSetupFn = append(arrEwtSetupFn,
				models.WhereCondFn{Condition: " ewt_setup.id = ? ", CondValue: ewalletTypeID},
			)
			arrEwtSetup, _ := models.GetEwtSetupListFn(arrEwtSetupFn, false)

			if len(arrEwtSetup) > 0 {
				arrEwtSetupLibrary[ewalletTypeID] = models.EwtSetup{
					EwtTypeName:  arrEwtSetup[0].EwtTypeName,
					CurrencyCode: arrEwtSetup[0].CurrencyCode,
				}
			}
		}

		if v, ok := arrEwtSetupLibrary[ewalletTypeID]; ok {
			ewalletTypeName = v.EwtTypeName
			currencyCode = v.CurrencyCode
		}

		arrEwtTransactions = append(arrEwtTransactions, EwtTransactions{
			EwalletTypeID:   ewalletTypeID,
			EwalletTypeName: ewalletTypeName,
			CurrencyCode:    currencyCode,
			ConversionRate:  1,
			PaidAmount:      arrEwtDetailsV.TotalOut,
			ReceivedAmount:  arrEwtDetailsV.TotalIn,
		})
	}

	// get from blockchain_trans
	var arrBlockchainTransFn = make([]models.WhereCondFn, 0)
	arrBlockchainTransFn = append(arrBlockchainTransFn,
		models.WhereCondFn{Condition: " blockchain_trans.doc_no IN(?) ", CondValue: arrDocNo},
	)
	arrBlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockchainTransFn, false)

	for _, arrBlockchainTransV := range arrBlockchainTrans {
		var (
			ewalletTypeID   = arrBlockchainTransV.EwalletTypeID
			ewalletTypeName = ""
			currencyCode    = ""
		)

		// get ewt_setup details if not grabbed before
		if _, ok := arrEwtSetupLibrary[ewalletTypeID]; !ok {
			arrEwtSetupFn := make([]models.WhereCondFn, 0)
			arrEwtSetupFn = append(arrEwtSetupFn,
				models.WhereCondFn{Condition: " ewt_setup.id = ? ", CondValue: ewalletTypeID},
			)
			arrEwtSetup, _ := models.GetEwtSetupListFn(arrEwtSetupFn, false)

			if len(arrEwtSetup) > 0 {
				arrEwtSetupLibrary[ewalletTypeID] = models.EwtSetup{
					EwtTypeName:  arrEwtSetup[0].EwtTypeName,
					CurrencyCode: arrEwtSetup[0].CurrencyCode,
				}
			}
		}

		if v, ok := arrEwtSetupLibrary[ewalletTypeID]; ok {
			ewalletTypeName = v.EwtTypeName
			currencyCode = v.CurrencyCode
		}

		arrEwtTransactions = append(arrEwtTransactions, EwtTransactions{
			EwalletTypeID:   ewalletTypeID,
			EwalletTypeName: ewalletTypeName,
			CurrencyCode:    currencyCode,
			ConversionRate:  arrBlockchainTransV.ConversionRate,
			PaidAmount:      arrBlockchainTransV.TotalOut,
			ReceivedAmount:  arrBlockchainTransV.TotalIn,
			Hash:            arrBlockchainTransV.HashValue,
		})
	}

	return arrEwtTransactions
}

// ExtraPaymentInfoStruct struct
type ExtraPaymentInfoStruct struct {
	GenTranxDataStatus bool
	EntMemberID        int
	Module             string
	ContractAddress    string
}

// ConvertPaymentInputToStructv2 struct
func ConvertPaymentInputToStructv2(payments string, extraPaymentInfo ExtraPaymentInfoStruct) (PaymentStruct, string) {
	paymentStruct := &PaymentStruct{}
	err := json.Unmarshal([]byte(payments), paymentStruct)
	if err != nil {
		return *paymentStruct, "invalid_payment_structure"
	}

	for arrMainWalletK, arrMainWalletV := range paymentStruct.MainWallet {
		// slice paid amount to x decimals points
		arrEwtSetupFn := make([]models.WhereCondFn, 0)
		arrEwtSetupFn = append(arrEwtSetupFn,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrMainWalletV.EwalletTypeCode},
			models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		)
		arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
		if err != nil {
			base.LogErrorLog("ConvertPaymentInputToStructv2-GetEwtSetupFn_arrMainWalletV_failed", err.Error(), arrEwtSetupFn, true)
			return *paymentStruct, "something_went_wrong"
		}
		if arrEwtSetup == nil {
			base.LogErrorLog("ConvertPaymentInputToStructv2-GetEwtSetupFn_arrMainWalletV_failed_2", "ewallet_type_code_"+arrMainWalletV.EwalletTypeCode+"_not_found", err.Error(), true)
			return *paymentStruct, "something_went_wrong"
		}
		decimalPlaces := arrEwtSetup.BlockchainDecimalPoint
		mainWalletAmount, err := helpers.ValueToFloat(helpers.CutOffDecimal(arrMainWalletV.Amount, uint(decimalPlaces), ".", ""))
		// mainWalletAmountStr := helpers.CutOffDecimal(arrMainWalletV.Amount, 8, ".", "")
		// fmt.Println("raw:", arrMainWalletV.Amount, "str:", mainWalletAmountStr, "float64:", mainWalletAmount)

		if err != nil {
			arrErr := map[string]interface{}{
				"Amount":        arrMainWalletV.Amount,
				"decimalPlaces": decimalPlaces,
			}
			base.LogErrorLog("ConvertPaymentInputToStructv2-ValueToFloat_1_failed", err.Error(), arrErr, true)
			return *paymentStruct, "something_went_wrong"
		}

		paymentStruct.MainWallet[arrMainWalletK].Amount = mainWalletAmount

		// if strings.ToLower(arrEwtSetup.Control) == "blockchain" {
		// 	extraPaymentInfo.ContractAddress = arrEwtSetup.ContractAddress

		// 	if extraPaymentInfo.GenTranxDataStatus {
		// 		tranxData, err := GenerateSalesTranxData(paymentStruct.MainWallet[arrMainWalletK], extraPaymentInfo)
		// 		if err != nil {
		// 			arrErr := map[string]interface{}{
		// 				"MainWallet":       paymentStruct.MainWallet,
		// 				"extraPaymentInfo": extraPaymentInfo,
		// 			}
		// 			base.LogErrorLog("ConvertPaymentInputToStructv2-GenerateSalesTranxData_main_wallet_failed", err.Error(), arrErr, true)
		// 			return *paymentStruct, "something_went_wrong"
		// 		}
		// 		paymentStruct.MainWallet[arrMainWalletK].TransactionData = tranxData
		// 	}
		// }
	}

	for arrSubWalletK, arrSubWalletV := range paymentStruct.SubWallet {
		// slice paid amount to x decimals points
		arrEwtSetupFn := make([]models.WhereCondFn, 0)
		arrEwtSetupFn = append(arrEwtSetupFn,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrSubWalletV.EwalletTypeCode},
			models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		)
		arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
		if err != nil {
			base.LogErrorLog("ConvertPaymentInputToStructv2-GetEwtSetupFn_arrSubWalletV_failed", err.Error(), arrEwtSetupFn, true)
			return *paymentStruct, "something_went_wrong"
		}
		if arrEwtSetup == nil {
			base.LogErrorLog("ConvertPaymentInputToStructv2-GetEwtSetupFn_arrSubWalletV_failed_2", "ewallet_type_code_"+arrSubWalletV.EwalletTypeCode+"_not_found", err.Error(), true)
			return *paymentStruct, "something_went_wrong"
		}
		decimalPlaces := arrEwtSetup.BlockchainDecimalPoint
		subWalletAmount, err := helpers.ValueToFloat(helpers.CutOffDecimal(arrSubWalletV.Amount, uint(decimalPlaces), ".", ""))
		// subWalletAmountStr := helpers.CutOffDecimal(arrSubWalletV.Amount, 8, ".", "")
		// fmt.Println("raw:", arrSubWalletV.Amount, "str:", subWalletAmountStr, "float64:", subWalletAmount)
		if err != nil {
			arrErr := map[string]interface{}{
				"Amount":        arrSubWalletV.Amount,
				"decimalPlaces": decimalPlaces,
			}
			base.LogErrorLog("ConvertPaymentInputToStructv2-ValueToFloat_2_failed", err.Error(), arrErr, true)
			return *paymentStruct, "something_went_wrong"
		}

		paymentStruct.SubWallet[arrSubWalletK].Amount = subWalletAmount

		// if strings.ToLower(arrEwtSetup.Control) == "blockchain" {
		// 	extraPaymentInfo.ContractAddress = arrEwtSetup.ContractAddress

		// 	if extraPaymentInfo.GenTranxDataStatus {
		// 		tranxData, err := GenerateSalesTranxData(paymentStruct.SubWallet[arrSubWalletK], extraPaymentInfo)
		// 		if err != nil {
		// 			arrErr := map[string]interface{}{
		// 				"MainWallet":       paymentStruct.MainWallet,
		// 				"extraPaymentInfo": extraPaymentInfo,
		// 			}
		// 			base.LogErrorLog("ConvertPaymentInputToStructv2-GenerateSalesTranxData_sub_wallet_failed", err.Error(), arrErr, true)
		// 			return *paymentStruct, "something_went_wrong"
		// 		}
		// 		paymentStruct.SubWallet[arrSubWalletK].TransactionData = tranxData
		// 	}
		// }
	}

	return *paymentStruct, ""
}

// GenerateSalesTranxData struct
func GenerateSalesTranxData(arrData IndividualPaymentStruct, arrExtraPaymentInfo ExtraPaymentInfoStruct) (string, error) {

	memberCryptoInfo, err := models.GetCustomMemberCryptoInfov2(arrExtraPaymentInfo.EntMemberID, arrData.EwalletTypeCode, true, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"entMemberID": arrExtraPaymentInfo.EntMemberID,
			"cryptoType":  arrData.EwalletTypeCode,
		}
		base.LogErrorLog("GenerateSalesTranxData-GetCustomMemberCryptoInfov2_main_failed", err.Error(), arrErr, true)
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	cryptoAddr := memberCryptoInfo.CryptoAddr
	privateKey := memberCryptoInfo.PrivateKey

	signingKeySetting, errMsg := GetSigningKeySettingByModule(arrData.EwalletTypeCode, cryptoAddr, arrExtraPaymentInfo.Module)
	if errMsg != "" {
		arrErr := map[string]interface{}{
			"entMemberID": arrExtraPaymentInfo.EntMemberID,
			"cryptoType":  arrData.EwalletTypeCode,
			"Module":      arrExtraPaymentInfo.Module,
		}
		base.LogErrorLog("GenerateSalesTranxData-GetSigningKeySettingByModule_failed", err.Error(), arrErr, true)
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
	}

	contractAddress := arrExtraPaymentInfo.ContractAddress
	toAddress := signingKeySetting["to_address"].(string)
	chainIDString := signingKeySetting["chain_id"].(string)
	chainID, _ := strconv.Atoi(chainIDString)
	chainIDInt64 := int64(chainID)
	maxGasString := signingKeySetting["max_gas"].(string)
	maxGas, _ := strconv.Atoi(maxGasString)
	maxGasUint64 := uint64(maxGas)

	// start generate signingKey
	arrProcecssGenerateSignTransaction := ProcecssGenerateSignTransactionStruct{
		TokenType:       arrData.EwalletTypeCode,
		PrivateKey:      privateKey,
		ContractAddress: contractAddress,
		ChainID:         chainIDInt64,
		FromAddr:        cryptoAddr, // member deduct wallet
		ToAddr:          toAddress,  // credit to company address
		Amount:          arrData.Amount,
		MaxGas:          maxGasUint64,
	}
	signingKey, err := ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
	if err != nil {
		base.LogErrorLog("GenerateSalesTranxData-ProcecssGenerateSignTransaction_failed", err.Error(), arrProcecssGenerateSignTransaction, true)
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
	}
	// end generate signingKey

	return signingKey, nil
}
