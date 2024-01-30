package trading_service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/float"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

type MemberTradingBuySetupRstStruct struct {
	AvailableBalanceFrom        float64                  `json:"available_balance_from"`
	AvailableBalanceFromDisplay string                   `json:"available_balance_from_display"`
	CurrencyCodeFrom            string                   `json:"currency_code_from"`
	CurrencyCodeTo              string                   `json:"currency_code_to"`
	HandlingFeesRate            float64                  `json:"handling_fees_rate"`
	SuggestedPriceRate          float64                  `json:"suggested_price_rate"`
	SuggestedPriceRateDisplay   string                   `json:"suggested_price_rate_display"`
	QuantityAdjustment          string                   `json:"quantity_adjustment"`
	SigningKeySetting           []map[string]interface{} `json:"signing_key_setting"`
}

type MemberTradingSellSetupRstStruct struct {
	AvailableBalanceFrom        float64                  `json:"available_balance_from"`
	AvailableBalanceFromDisplay string                   `json:"available_balance_from_display"`
	CurrencyCodeFrom            string                   `json:"currency_code_from"`
	CurrencyCodeTo              string                   `json:"currency_code_to"`
	HandlingFeesRate            float64                  `json:"handling_fees_rate"`
	SuggestedPriceRate          float64                  `json:"suggested_price_rate"`
	SuggestedPriceRateDisplay   string                   `json:"suggested_price_rate_display"`
	QuantityAdjustment          string                   `json:"quantity_adjustment"`
	SigningKeySetting           []map[string]interface{} `json:"signing_key_setting"`
}

type MemberTradingSetupStruct struct {
	EntMemberID int
	LangCode    string
	CryptoCode  string
	RequestType string
}

// func GetMemberTradingSellSetupv1
func GetMemberTradingSellSetupv1(arrData MemberTradingSetupStruct) (*MemberTradingSellSetupRstStruct, error) {
	var arrDataReturn MemberTradingSellSetupRstStruct
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)

	arrTradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	db := models.GetDB() // no need set begin transaction
	if len(arrTradingSetup) > 0 {
		var bal float64
		// var convBal float64
		var priceRate float64
		// var paymentBal float64
		cryptoAddr, err := member_service.ProcessGetMemAddress(db, arrData.EntMemberID, arrTradingSetup[0].CodeFrom)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  arrTradingSetup[0].CodeFrom,
			}
			base.LogErrorLog("GetMemberTradingSellSetupv1_ProcessGetMemAddress_failed", err.Error(), arrErrData, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		arrSignedKeySetting := make([]map[string]interface{}, 0)
		if strings.ToLower(arrTradingSetup[0].ControlFrom) == "blockchain" {
			// bal, _, priceRate = wallet_service.GetBlockchainWalletBalanceByAddressV1(arrTradingSetup[0].CodeFrom, cryptoAddr)
			BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrTradingSetup[0].CodeFrom, cryptoAddr, arrData.EntMemberID)
			fmt.Println("BlkCWalBal:", BlkCWalBal)
			bal = BlkCWalBal.AvailableBalance
			fmt.Println("bal:", bal)

			tokenRate, err := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

			if err != nil {
				base.LogErrorLog("GetMemberTradingBuySetupv1_GetLatestSecPriceMovement_failed", err.Error(), tokenRate, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			if tokenRate > 0 {
				priceRate = tokenRate
			}

			// priceRate = BlkCWalBal.Rate
			signedKeySetup, errMsg := wallet_service.GetSigningKeySettingByModule(arrTradingSetup[0].CodeFrom, cryptoAddr, "TRADING")
			if errMsg != "" {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
			}
			signedKeySetup["decimal_point"] = arrTradingSetup[0].BlockchainDecimalPointFrom

			hotWalletInfo, err := models.GetHotWalletInfo()
			if err != nil {
				base.LogErrorLog("GetMemberTradingSellSetupv1_GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
			}
			var isBase bool
			if arrTradingSetup[0].IsBaseFrom == 1 {
				isBase = true
			}
			signedKeySetup["to_address"] = hotWalletInfo.HotWalletAddress
			signedKeySetup["is_base"] = isBase
			signedKeySetup["method"] = "transfer"
			if arrTradingSetup[0].ContractAddrFrom != "" {
				signedKeySetup["contract_address"] = arrTradingSetup[0].ContractAddrFrom
			}
			arrSignedKeySetting = append(arrSignedKeySetting,
				signedKeySetup,
			)
			// signedKeySetting = signedKeySetup
		} else {
			// arrEwtBal := wallet_service.GetWalletBalanceStruct{
			// 	EntMemberID: arrData.EntMemberID,
			// 	EwtTypeID:   arrTradingSetup[0].IDTo,
			// }
			// walletBalanceRst := wallet_service.GetWalletBalance(arrEwtBal)
			// if walletBalanceRst.Balance > 0 {
			// 	bal = walletBalanceRst.Balance
			// }
			base.LogErrorLog("GetMemberTradingSellSetupv1_arrTradingSetup[0].Control_error", "trade_control_setting_is_wrong", arrTradingSetup[0].ControlFrom, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		var quantityAdjustment float64
		quantityAdjustment = 1
		if arrTradingSetup[0].DecimalPointFrom > 0 {
			p := math.Pow(10, float64(arrTradingSetup[0].DecimalPointFrom))
			// quantityAdjustment = 1 / p
			quantityAdjustment = float.Div(1, p)
		}
		quantityAdjustmentString := helpers.CutOffDecimal(quantityAdjustment, uint(arrTradingSetup[0].DecimalPointFrom), ".", ",")

		nameToTranslated := helpers.TranslateV2(arrTradingSetup[0].NameTo, arrData.LangCode, nil)
		nameFromTranslated := helpers.TranslateV2(arrTradingSetup[0].NameFrom, arrData.LangCode, nil)
		fmt.Println("2 bal:", bal)
		arrDataReturn = MemberTradingSellSetupRstStruct{
			AvailableBalanceFrom:        bal,
			AvailableBalanceFromDisplay: helpers.CutOffDecimal(bal, uint(arrTradingSetup[0].DecimalPointFrom), ".", ","),
			CurrencyCodeFrom:            nameFromTranslated,
			CurrencyCodeTo:              nameToTranslated,
			HandlingFeesRate:            arrTradingSetup[0].Fees,
			SuggestedPriceRate:          priceRate,
			SuggestedPriceRateDisplay:   helpers.CutOffDecimal(priceRate, uint(arrTradingSetup[0].DecimalPointTo), ".", ",") + " " + nameToTranslated,
			QuantityAdjustment:          quantityAdjustmentString,
			SigningKeySetting:           arrSignedKeySetting,
		}
	}
	fmt.Println("last arrDataReturn", arrDataReturn)
	return &arrDataReturn, nil
}

// func GetMemberTradingBuySetupv1
func GetMemberTradingBuySetupv1(arrData MemberTradingSetupStruct) (*MemberTradingBuySetupRstStruct, error) {
	var arrDataReturn MemberTradingBuySetupRstStruct
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)

	arrTradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	if len(arrTradingSetup) > 0 {
		var bal float64
		priceRate := float64(1)

		arrSignedKeySetting := make([]map[string]interface{}, 0)

		if strings.ToLower(arrTradingSetup[0].ControlTo) == "internal" {
			arrEwtBal := wallet_service.GetWalletBalanceStruct{
				EntMemberID: arrData.EntMemberID,
				EwtTypeID:   arrTradingSetup[0].IDTo,
			}
			walletBalanceRst := wallet_service.GetWalletBalance(arrEwtBal)
			if walletBalanceRst.Balance > 0 {
				bal = walletBalanceRst.Balance
			}
		} else if strings.ToLower(arrTradingSetup[0].ControlTo) == "blockchain" {
			db := models.GetDB() // no need set begin transaction

			cryptoAddr, err := member_service.ProcessGetMemAddress(db, arrData.EntMemberID, arrTradingSetup[0].CodeTo)
			if err != nil {
				arrErrData := map[string]interface{}{
					"entMemberID": arrData.EntMemberID,
					"cryptoType":  arrTradingSetup[0].CodeTo,
				}
				base.LogErrorLog("GetMemberTradingBuySetupv1-ProcessGetMemAddress_failed", err.Error(), arrErrData, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrTradingSetup[0].CodeTo, cryptoAddr, arrData.EntMemberID)

			bal = BlkCWalBal.AvailableBalance
			priceRate = BlkCWalBal.Rate
			signedKeySetup, errMsg := wallet_service.GetSigningKeySettingByModule(arrTradingSetup[0].CodeTo, cryptoAddr, "TRADING")
			if errMsg != "" {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
			}
			signedKeySetup["decimal_point"] = arrTradingSetup[0].BlockchainDecimalPointTo

			hotWalletInfo, err := models.GetHotWalletInfo()
			if err != nil {
				base.LogErrorLog("GetMemberTradingBuySetupv1-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
			}
			var isBase bool
			// if arrTradingSetup[0].IsBaseFrom == 1 {
			// 	isBase = true
			// }
			signedKeySetup["to_address"] = hotWalletInfo.HotWalletAddress
			signedKeySetup["is_base"] = isBase
			signedKeySetup["method"] = "transfer"
			if arrTradingSetup[0].ContractAddrTo != "" {
				signedKeySetup["contract_address"] = arrTradingSetup[0].ContractAddrTo
			}
			arrSignedKeySetting = append(arrSignedKeySetting,
				signedKeySetup,
			)
			// base.LogErrorLog("GetMemberTradingBuySetupv1-arrTradingSetup[0].ControlTo", "payemnt_trade_control_setting_is_wrong", arrTradingSetup[0].ControlTo, true)
			// return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		tokenRate, err := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

		if err != nil {
			base.LogErrorLog("GetMemberTradingBuySetupv1_GetLatestSecPriceMovement_failed", err.Error(), tokenRate, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if tokenRate > 0 {
			priceRate = tokenRate
		}
		// if strings.ToLower(arrData.CryptoCode) == "sec" {
		// 	//get price movement for sec
		// 	tokenRate, err := models.GetLatestSecPriceMovement()

		// 	if err != nil {
		// 		base.LogErrorLog("GetMemberTradingBuySetupv1_GetLatestSecPriceMovement_failed", err.Error(), tokenRate, true)
		// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// 	}
		// 	priceRate = tokenRate
		// } else if strings.ToLower(arrData.CryptoCode) == "liga" {
		// 	//get price movement for LIGA
		// 	tokenRate, err := models.GetLatestLigaPriceMovement()

		// 	if err != nil {
		// 		base.LogErrorLog("GetMemberTradingBuySetupv1_GetLatestLigaPriceMovement_failed", err.Error(), tokenRate, true)
		// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// 	}
		// 	priceRate = tokenRate
		// }
		nameToTranslated := helpers.TranslateV2(arrTradingSetup[0].NameTo, arrData.LangCode, nil)
		nameFromTranslated := helpers.TranslateV2(arrTradingSetup[0].NameFrom, arrData.LangCode, nil)

		var quantityAdjustment float64
		quantityAdjustment = 1
		if arrTradingSetup[0].DecimalPointFrom > 0 {
			p := math.Pow(10, float64(arrTradingSetup[0].DecimalPointFrom))
			// quantityAdjustment = 1 / p
			quantityAdjustment = float.Div(1, p)
		}
		quantityAdjustmentString := helpers.CutOffDecimal(quantityAdjustment, uint(arrTradingSetup[0].DecimalPointFrom), ".", ",")

		arrDataReturn = MemberTradingBuySetupRstStruct{
			AvailableBalanceFrom:        bal,
			AvailableBalanceFromDisplay: helpers.CutOffDecimal(bal, uint(arrTradingSetup[0].DecimalPointFrom), ".", ","),
			CurrencyCodeFrom:            nameToTranslated,
			CurrencyCodeTo:              nameFromTranslated,
			HandlingFeesRate:            arrTradingSetup[0].Fees,
			SuggestedPriceRate:          priceRate,
			SuggestedPriceRateDisplay:   helpers.CutOffDecimal(priceRate, uint(arrTradingSetup[0].DecimalPointTo), ".", ",") + " " + nameToTranslated,
			QuantityAdjustment:          quantityAdjustmentString,
			SigningKeySetting:           arrSignedKeySetting,
		}
	}

	return &arrDataReturn, nil
}

type MemberTradingSelectionRstStruct struct {
	CryptoFromCode string `json:"crypto_from_code"`
	CryptoFromName string `json:"crypto_from_name"`
	CryptoToCode   string `json:"crypto_to_code"`
	CryptoToName   string `json:"crypto_to_name"`
	UnitPrice      string `json:"unit_price"`
}

// func GetMemberTradingSelectionListv1
func GetMemberTradingSelectionListv1(arrData MemberTradingSetupStruct) ([]MemberTradingSelectionRstStruct, error) {

	arrMemberTradingSelectionList := make([]MemberTradingSelectionRstStruct, 0)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
	)

	arrTradingSelection, _ := models.GetTradingSetupFn(arrCond, false)
	if len(arrTradingSelection) > 0 {
		unitPrice := "1.00"
		for _, arrTradingSelectionV := range arrTradingSelection {
			if strings.ToLower(arrTradingSelectionV.CodeFrom) == "liga" {
				arrMarketPrice, _ := models.GetLatestLigaPriceMovement()
				if arrMarketPrice > 0 {
					unitPrice = helpers.CutOffDecimal(arrMarketPrice, uint(arrTradingSelectionV.DecimalPointFrom), ".", ",")
				}
			} else if strings.ToLower(arrTradingSelectionV.CodeFrom) == "sec" {
				arrMarketPrice, _ := models.GetLatestSecPriceMovement()
				if arrMarketPrice > 0 {
					unitPrice = helpers.CutOffDecimal(arrMarketPrice, uint(arrTradingSelectionV.DecimalPointFrom), ".", ",")
				}
			}
			arrMemberTradingSelectionList = append(arrMemberTradingSelectionList,
				MemberTradingSelectionRstStruct{
					CryptoFromCode: arrTradingSelectionV.CodeFrom,
					CryptoFromName: helpers.TranslateV2(arrTradingSelectionV.NameFrom, arrData.LangCode, nil),
					CryptoToCode:   arrTradingSelectionV.CodeTo,
					CryptoToName:   helpers.TranslateV2(arrTradingSelectionV.NameTo, arrData.LangCode, nil),
					UnitPrice:      unitPrice,
				},
			)
		}
	}

	return arrMemberTradingSelectionList, nil
}

type PeriodicityStruct struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type TradingSelectionStruct struct {
	TradingSelectionCode string `json:"trading_selection_code"`
	TradingSelectionName string `json:"trading_selection_name"`
}

type MemberTradingViewSetupRstStruct struct {
	PeriodicityList      []PeriodicityStruct      `json:"periodicity_list"`
	TradingSelectionList []TradingSelectionStruct `json:"trading_selection_list"`
}

// func GetMemberTradingViewSetupv1
func GetMemberTradingViewSetupv1(arrData MemberTradingSetupStruct) (*MemberTradingViewSetupRstStruct, error) {

	var arrDataReturn MemberTradingViewSetupRstStruct

	arrTradingSelectionList := make([]TradingSelectionStruct, 0)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
	)

	arrTradingSelection, _ := models.GetTradingSetupFn(arrCond, false)
	if len(arrTradingSelection) > 0 {
		for _, arrTradingSelectionV := range arrTradingSelection {
			arrTradingSelectionList = append(arrTradingSelectionList,
				TradingSelectionStruct{
					TradingSelectionCode: arrTradingSelectionV.CodeFrom,
					TradingSelectionName: helpers.TranslateV2(arrTradingSelectionV.NameFrom, arrData.LangCode, nil),
				},
			)
		}
	}

	arrSetup, _ := models.GetSysGeneralSetupByID("trading_view_setup")

	var arrPeriodicityList []PeriodicityStruct
	err := json.Unmarshal([]byte(arrSetup.InputType1), &arrPeriodicityList)

	if err != nil {
		base.LogErrorLog("GetMemberTradingViewSetupv1-failed_to_decode_PeriodicityStruct", err.Error(), arrSetup.InputType1, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	for arrPeriodicityListK, arrPeriodicityListV := range arrPeriodicityList {
		arrPeriodicityList[arrPeriodicityListK].Name = helpers.TranslateV2(arrPeriodicityListV.Name, arrData.LangCode, nil)
	}

	arrDataReturn.PeriodicityList = arrPeriodicityList
	arrDataReturn.TradingSelectionList = arrTradingSelectionList

	return &arrDataReturn, nil
}

type MemberTradingBuySetupRstv2Struct struct {
	AvailableBalanceFrom        float64 `json:"available_balance_from"`
	AvailableBalanceFromDisplay string  `json:"available_balance_from_display"`
	CurrencyCodeFrom            string  `json:"currency_code_from"`
	CurrencyCodeTo              string  `json:"currency_code_to"`
	HandlingFeesRate            float64 `json:"handling_fees_rate"`
	SuggestedPriceRate          float64 `json:"suggested_price_rate"`
	SuggestedPriceRateDisplay   string  `json:"suggested_price_rate_display"`
	QuantityAdjustment          string  `json:"quantity_adjustment"`
}

// func GetMemberTradingBuySetupv2
func GetMemberTradingBuySetupv2(arrData MemberTradingSetupStruct) (*MemberTradingBuySetupRstv2Struct, error) {
	var arrDataReturn MemberTradingBuySetupRstv2Struct
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)

	arrTradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	if len(arrTradingSetup) > 0 {
		var bal float64
		priceRate := float64(1)

		if strings.ToLower(arrTradingSetup[0].ControlTo) == "internal" {
			arrEwtBal := wallet_service.GetWalletBalanceStruct{
				EntMemberID: arrData.EntMemberID,
				EwtTypeID:   arrTradingSetup[0].IDTo,
			}
			walletBalanceRst := wallet_service.GetWalletBalance(arrEwtBal)
			if walletBalanceRst.Balance > 0 {
				bal = walletBalanceRst.Balance
			}
		} else if strings.ToLower(arrTradingSetup[0].ControlTo) == "blockchain" {
			db := models.GetDB() // no need set begin transaction

			cryptoAddr, err := member_service.ProcessGetMemAddress(db, arrData.EntMemberID, arrTradingSetup[0].CodeTo)
			if err != nil {
				arrErrData := map[string]interface{}{
					"entMemberID": arrData.EntMemberID,
					"cryptoType":  arrTradingSetup[0].CodeTo,
				}
				base.LogErrorLog("GetMemberTradingBuySetupv1-ProcessGetMemAddress_failed", err.Error(), arrErrData, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrTradingSetup[0].CodeTo, cryptoAddr, arrData.EntMemberID)

			bal = BlkCWalBal.AvailableBalance
			priceRate = BlkCWalBal.Rate
		}

		tokenRate, err := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

		if err != nil {
			base.LogErrorLog("GetMemberTradingBuySetupv1_GetLatestSecPriceMovement_failed", err.Error(), tokenRate, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if tokenRate > 0 {
			priceRate = tokenRate
		}

		nameToTranslated := helpers.TranslateV2(arrTradingSetup[0].NameTo, arrData.LangCode, nil)
		nameFromTranslated := helpers.TranslateV2(arrTradingSetup[0].NameFrom, arrData.LangCode, nil)

		var quantityAdjustment float64
		quantityAdjustment = 1
		if arrTradingSetup[0].DecimalPointFrom > 0 {
			p := math.Pow(10, float64(arrTradingSetup[0].DecimalPointFrom))
			// quantityAdjustment = 1 / p
			quantityAdjustment = float.Div(1, p)
		}
		quantityAdjustmentString := helpers.CutOffDecimal(quantityAdjustment, uint(arrTradingSetup[0].DecimalPointFrom), ".", ",")

		arrDataReturn = MemberTradingBuySetupRstv2Struct{
			AvailableBalanceFrom:        bal,
			AvailableBalanceFromDisplay: helpers.CutOffDecimal(bal, uint(arrTradingSetup[0].DecimalPointFrom), ".", ","),
			CurrencyCodeFrom:            nameToTranslated,
			CurrencyCodeTo:              nameFromTranslated,
			HandlingFeesRate:            arrTradingSetup[0].Fees,
			SuggestedPriceRate:          priceRate,
			SuggestedPriceRateDisplay:   helpers.CutOffDecimal(priceRate, uint(arrTradingSetup[0].DecimalPointTo), ".", ",") + " " + nameToTranslated,
			QuantityAdjustment:          quantityAdjustmentString,
		}
	}

	return &arrDataReturn, nil
}

type MemberTradingSellSetupRstv2Struct struct {
	AvailableBalanceFrom        float64 `json:"available_balance_from"`
	AvailableBalanceFromDisplay string  `json:"available_balance_from_display"`
	CurrencyCodeFrom            string  `json:"currency_code_from"`
	CurrencyCodeTo              string  `json:"currency_code_to"`
	HandlingFeesRate            float64 `json:"handling_fees_rate"`
	SuggestedPriceRate          float64 `json:"suggested_price_rate"`
	SuggestedPriceRateDisplay   string  `json:"suggested_price_rate_display"`
	QuantityAdjustment          string  `json:"quantity_adjustment"`
}

// func GetMemberTradingSellSetupv2
func GetMemberTradingSellSetupv2(arrData MemberTradingSetupStruct) (*MemberTradingSellSetupRstv2Struct, error) {
	var arrDataReturn MemberTradingSellSetupRstv2Struct
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)

	arrTradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	db := models.GetDB() // no need set begin transaction
	if len(arrTradingSetup) > 0 {
		var bal float64
		// var convBal float64
		var priceRate float64
		// var paymentBal float64
		cryptoAddr, err := member_service.ProcessGetMemAddress(db, arrData.EntMemberID, arrTradingSetup[0].CodeFrom)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  arrTradingSetup[0].CodeFrom,
			}
			base.LogErrorLog("GetMemberTradingSellSetupv1_ProcessGetMemAddress_failed", err.Error(), arrErrData, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if strings.ToLower(arrTradingSetup[0].ControlFrom) == "blockchain" {
			// bal, _, priceRate = wallet_service.GetBlockchainWalletBalanceByAddressV1(arrTradingSetup[0].CodeFrom, cryptoAddr)
			BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrTradingSetup[0].CodeFrom, cryptoAddr, arrData.EntMemberID)
			bal = BlkCWalBal.AvailableBalance

			tokenRate, err := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

			if err != nil {
				base.LogErrorLog("GetMemberTradingBuySetupv1_GetLatestSecPriceMovement_failed", err.Error(), tokenRate, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			if tokenRate > 0 {
				priceRate = tokenRate
			}
		} else {
			// arrEwtBal := wallet_service.GetWalletBalanceStruct{
			// 	EntMemberID: arrData.EntMemberID,
			// 	EwtTypeID:   arrTradingSetup[0].IDTo,
			// }
			// walletBalanceRst := wallet_service.GetWalletBalance(arrEwtBal)
			// if walletBalanceRst.Balance > 0 {
			// 	bal = walletBalanceRst.Balance
			// }
			base.LogErrorLog("GetMemberTradingSellSetupv1_arrTradingSetup[0].Control_error", "trade_control_setting_is_wrong", arrTradingSetup[0].ControlFrom, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		var quantityAdjustment float64
		quantityAdjustment = 1
		if arrTradingSetup[0].DecimalPointFrom > 0 {
			p := math.Pow(10, float64(arrTradingSetup[0].DecimalPointFrom))
			// quantityAdjustment = 1 / p
			quantityAdjustment = float.Div(1, p)
		}
		quantityAdjustmentString := helpers.CutOffDecimal(quantityAdjustment, uint(arrTradingSetup[0].DecimalPointFrom), ".", ",")

		nameToTranslated := helpers.TranslateV2(arrTradingSetup[0].NameTo, arrData.LangCode, nil)
		nameFromTranslated := helpers.TranslateV2(arrTradingSetup[0].NameFrom, arrData.LangCode, nil)
		arrDataReturn = MemberTradingSellSetupRstv2Struct{
			AvailableBalanceFrom:        bal,
			AvailableBalanceFromDisplay: helpers.CutOffDecimal(bal, uint(arrTradingSetup[0].DecimalPointFrom), ".", ","),
			CurrencyCodeFrom:            nameFromTranslated,
			CurrencyCodeTo:              nameToTranslated,
			HandlingFeesRate:            arrTradingSetup[0].Fees,
			SuggestedPriceRate:          priceRate,
			SuggestedPriceRateDisplay:   helpers.CutOffDecimal(priceRate, uint(arrTradingSetup[0].DecimalPointTo), ".", ",") + " " + nameToTranslated,
			QuantityAdjustment:          quantityAdjustmentString,
		}
	}
	fmt.Println("last arrDataReturn", arrDataReturn)
	return &arrDataReturn, nil
}
