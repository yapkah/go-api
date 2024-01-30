package trading_service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/float"
	"github.com/yapkah/go-api/service/wallet_service"

	"github.com/jinzhu/gorm"
)

type SellMemberTradingStruct struct {
	BuyID    int
	Quantity float64
	MemberID int
	LangCode string
}

// func ProcessMemberSellTradingv1
func ProcessMemberSellTradingv1(tx *gorm.DB, arrData SellMemberTradingStruct) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.id = ? ", CondValue: arrData.BuyID},
		models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: "P"},
	)

	arrTradingBuy, err := models.GetTradingBuyFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingv1-failed_to_get_GetTradingBuyFn", err.Error(), arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	if len(arrTradingBuy) < 1 {
		base.LogErrorLog("ProcessMemberSellTradingv1-invalid_buy_id", arrCond, nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	if arrTradingBuy[0].TotalUnit < arrData.Quantity {
		msgParams := map[string]string{}
		msgParams["q"] = "0"
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_quantity_equal_or_less_than_:q", Data: msgParams}
	}

	// start get trading info
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.code_from = ?", CondValue: arrTradingBuy[0].CryptoCode},
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	)
	arrTradeSetup, err := models.GetTradingSetupFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingv1-failed_to_GetTradingSetupFn", err.Error(), arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	if len(arrTradeSetup) < 1 {
		base.LogErrorLog("ProcessMemberSellTradingv1-not_valid_to_sell", arrCond, nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}
	// end get trading info

	// arrCond = make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrTradeSetup[0].CodeTo},
	// )
	// ewtSetup, err := models.GetEwtSetupFn(arrCond, "", false)

	// if err != nil {
	// 	base.LogErrorLog("ProcessMemberSellTradingv1-invalid_trading_payment", err.Error(), arrCond, true)
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
	// }
	// if ewtSetup.ID < 1 {
	// 	base.LogErrorLog("ProcessMemberSellTradingv1-wallet_trading_payment_missing", arrCond, nil, true)
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	// }

	cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrTradingBuy[0].MemberID, arrTradingBuy[0].CryptoCode, true, false)
	if err != nil {
		arrErrData := map[string]interface{}{
			"entMemberID": arrData.MemberID,
			"cryptoType":  arrTradingBuy[0].CryptoCode,
		}
		base.LogErrorLog("ProcessMemberSellTradingRequestv1_GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	totalAmount := arrData.Quantity * arrTradingBuy[0].UnitPrice

	// start generate sign transaction from company to member
	signedKeySetup, errMsg := wallet_service.GetSigningKeySettingByModule(arrTradingBuy[0].CryptoCode, cryptoAddr, "TRADING")

	if errMsg != "" {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "member_id = ?", CondValue: 0},
	)
	arrComPK, err := models.GetReumAddFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingv1_failed_to_GetReumAddFn", err.Error(), arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	if len(arrComPK) < 1 {
		base.LogErrorLog("ProcessMemberSellTradingv1_missing_company_private_key", arrCond, nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	arrSignTranx := wallet_service.SignTransactionViaApiStruct{
		TokenType:  arrTradingBuy[0].CryptoCode,
		PrivateKey: arrComPK[0].PrivateKey,
		ToAddr:     cryptoAddr,
		Value:      fmt.Sprintf("%.8f", arrData.Quantity),
		Gas:        fmt.Sprintf("%v", signedKeySetup["max_gas"]),
		GasPrice:   fmt.Sprintf("%v", signedKeySetup["gas_price"]),
	}
	signTranxRst, err := wallet_service.GenerateSignTransactionViaApi(arrSignTranx)

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
	}
	// end generate sign transaction from company to member

	// start send signing key to blockchain site
	hashValue, errMsg := wallet_service.SignedTransaction(signTranxRst.TransactionHash)

	if errMsg != "" {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
	}
	// end send signing key to blockchain site

	// start check balance
	// ewtOut := SaveMemberWalletStruct{
	// 	EntMemberID:     arrData.MemberID,
	// 	EwalletTypeID:   ewtSetup.ID,
	// 	TotalOut:        w.WithdrawAmt,
	// 	TransactionType: "TRADING",
	// 	DocNo:           docs,
	// 	Remark:          "#*withdraw*#" + " " + docs,
	// 	CreatedBy:       strconv.Itoa(w.MemberId),
	// }

	// _, err_out := SaveMemberWallet(tx, ewtOut)

	// if err_out != nil {
	// 	models.ErrorLog(err_out, "ewt-withdraw-errCode09", ewtOut) //store error log
	// 	return nil, "", "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: helpers.Translate("withdrawal error"+" "+"code09", w.LangCode), Data: ""}
	// }
	// end check balance

	arrCrtTradSell := models.TradingSell{
		CryptoCode:  arrTradingBuy[0].CryptoCode,
		MemberID:    arrData.MemberID,
		TotalUnit:   arrData.Quantity,
		UnitPrice:   arrTradingBuy[0].UnitPrice,
		TotalAmount: totalAmount,
		Status:      "AP",
		CreatedBy:   strconv.Itoa(arrData.MemberID),
		ApprovedAt:  base.GetCurrentDateTimeT(),
		ApprovedBy:  strconv.Itoa(arrData.MemberID),
	}

	_, err = models.AddTradingSell(tx, arrCrtTradSell)
	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingv1-failed_to_save_trading_sell", err.Error(), arrCrtTradSell, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	docTypeM := "TRADM"
	docNoM, err := models.GetRunningDocNo(docTypeM, tx) //get withdraw doc no

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingv1-get_trading_doc_no_failed", err.Error(), nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	arrCrtTradMatch := models.TradingMatch{
		DocNo:          docNoM,
		CryptoCode:     arrTradingBuy[0].CryptoCode,
		SellID:         arrTradingBuy[0].ID,
		BuyID:          arrData.MemberID,
		SellerMemberID: arrTradingBuy[0].MemberID,
		BuyerMemberID:  arrData.MemberID,
		TotalUnit:      arrData.Quantity,
		UnitPrice:      arrTradingBuy[0].UnitPrice,
		TotalAmount:    totalAmount,
		SigningKey:     signTranxRst.TransactionHash,
		TransHash:      hashValue,
		Status:         "AP",
		CreatedBy:      strconv.Itoa(arrData.MemberID),
		ApprovedAt:     base.GetCurrentDateTimeT(),
		ApprovedBy:     strconv.Itoa(arrData.MemberID),
	}

	_, err = models.AddTradingMatch(tx, arrCrtTradMatch)
	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingv1-failed_to_save_trading_match", err.Error(), arrCrtTradMatch, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_failed"}
	}

	err = models.UpdateRunningDocNo(docTypeM, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingRequestv1_failed_in_UpdateRunningDocNo", err.Error(), docTypeM, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}

type SellMemberTradingRequestStruct struct {
	UnitPrice   float64
	Quantity    float64
	EntMemberID int
	CryptoCode  string
	SigningKey  string
}

// func ProcessMemberSellTradingRequestv1
func ProcessMemberSellTradingRequestv1(tx *gorm.DB, arrData SellMemberTradingRequestStruct) error {
	totalAmount := arrData.Quantity * arrData.UnitPrice

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)
	tradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	if len(tradingSetup) < 1 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	docTypeS := "TRADS"
	docNoS, err := models.GetRunningDocNo(docTypeS, tx) //get withdraw doc no

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingRequestv1-get_trading_doc_no_failed", err.Error(), nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	fmt.Println("docNoS:", docNoS)
	if strings.ToLower(tradingSetup[0].ControlFrom) == "blockchain" {
		cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrData.EntMemberID, arrData.CryptoCode, true, false)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  arrData.CryptoCode,
			}
			base.LogErrorLog("ProcessMemberSellTradingRequestv1_GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr, arrData.EntMemberID)
		bal := BlkCWalBal.Balance

		// bal, _, _ := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr)
		fmt.Println("bal:", bal)

		// start checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]
		// if strings.ToLower(arrData.CryptoCode) == "sec" || strings.ToLower(arrData.CryptoCode) == "liga" {

		// 	arrCond = make([]models.WhereCondFn, 0)
		// 	arrCond = append(arrCond,
		// 		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		// 		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode+"H")},
		// 	)
		// 	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

		// 	if holdingEwtSetup != nil{
		// 		arrHoldingEwtBal := wallet_service.GetWalletBalanceStruct{
		// 			EntMemberID: arrData.EntMemberID,
		// 			EwtTypeID:   holdingEwtSetup.ID,
		// 		}
		// 		holdingWalletBalance := wallet_service.GetWalletBalance(arrHoldingEwtBal)
		// 		if holdingWalletBalance.Balance > 0 {
		// 			bal = bal - holdingWalletBalance.Balance
		// 		}
		// 	}
		// }
		// end checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]

		if bal < arrData.Quantity { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_sell"}
		}

		// start send signing key to blockchain site
		// hashValue, errMsg := wallet_service.SignedTransaction(arrData.SigningKey)

		// if errMsg != "" {
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		// }
		// end send signing key to blockchain site

		if strings.ToLower(arrData.CryptoCode) == "sec" || strings.ToLower(arrData.CryptoCode) == "liga" {
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
				models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode + "H")},
			)
			holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

			if holdingEwtSetup != nil {
				// start add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
				ewtIn := wallet_service.SaveMemberWalletStruct{
					EntMemberID:     arrData.EntMemberID,
					EwalletTypeID:   holdingEwtSetup.ID,
					TotalIn:         arrData.Quantity,
					TransactionType: "TRADING",
					DocNo:           docNoS,
					Remark:          "#*sell_trading_request*#" + " " + docNoS,
					CreatedBy:       strconv.Itoa(arrData.EntMemberID),
				}

				_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

				if err != nil {
					base.LogErrorLog("ProcessMemberSellTradingRequestv1_SaveMemberWallet_failed", err.Error(), ewtIn, true)
					return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
				}
				// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
			}
		}

		arrCrtTradSell := models.TradingSell{
			CryptoCode:  arrData.CryptoCode,
			DocNo:       docNoS,
			MemberID:    arrData.EntMemberID,
			TotalUnit:   arrData.Quantity,
			UnitPrice:   arrData.UnitPrice,
			TotalAmount: totalAmount,
			BalanceUnit: arrData.Quantity,
			Status:      "P",
			SigningKey:  arrData.SigningKey,
			// TransHash:   hashValue,
			CreatedBy:  strconv.Itoa(arrData.EntMemberID),
			ApprovedAt: base.GetCurrentDateTimeT(),
			ApprovedBy: strconv.Itoa(arrData.EntMemberID),
		}

		_, err = models.AddTradingSell(tx, arrCrtTradSell)
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv1_failed_to_save_trading_sell", err.Error(), arrCrtTradSell, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		err = models.UpdateRunningDocNo(docTypeS, tx) //update doc no
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv1_failed_in_UpdateRunningDocNo", err.Error(), docTypeS, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
	} else {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_not_available"}
	}

	return nil
}

// func ProcessMemberSellTradingRequestv2
func ProcessMemberSellTradingRequestv2(tx *gorm.DB, arrData SellMemberTradingRequestStruct) (string, error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)
	tradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	if len(tradingSetup) < 1 {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	docTypeS := "TRADS"
	docNoS, err := models.GetRunningDocNo(docTypeS, tx) //get withdraw doc no

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingRequestv2-get_trading_selsl_doc_no_failed", err.Error(), nil, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if strings.ToLower(tradingSetup[0].ControlFrom) == "blockchain" {

		if arrData.SigningKey == "" {
			arrErr := map[string]interface{}{
				"trading_setup": tradingSetup,
				"arrData":       arrData,
			}
			base.LogErrorLog("ProcessMemberSellTradingRequestv2-signing_key_is_missing", arrErr, arrCond, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if strings.ToLower(arrData.CryptoCode) == "sec" {
			//get price movement for sec
			tokenRate, err := models.GetLatestSecPriceMovement()

			if err != nil {
				base.LogErrorLog("ProcessMemberSellTradingRequestv2_GetLatestSecPriceMovement_failed", err.Error(), tokenRate, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			if arrData.UnitPrice != tokenRate {
				arrErr := map[string]interface{}{
					"current_rate":   tokenRate,
					"front_end_rate": arrData.UnitPrice,
				}
				base.LogErrorLog("ProcessMemberSellTradingRequestv2_front_end_price_not_tally_with_current_price", arrErr, arrData, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		} else if strings.ToLower(arrData.CryptoCode) == "liga" {
			//get price movement for LIGA
			tokenRate, err := models.GetLatestLigaPriceMovement()

			if err != nil {
				base.LogErrorLog("ProcessMemberSellTradingRequestv2_GetLatestLigaPriceMovement_failed", err.Error(), tokenRate, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			if arrData.UnitPrice != tokenRate {
				arrErr := map[string]interface{}{
					"current_rate":   tokenRate,
					"front_end_rate": arrData.UnitPrice,
				}
				base.LogErrorLog("ProcessMemberSellTradingRequestv2_front_end_price_not_tally_with_current_price", arrErr, arrData, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		}

		cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrData.EntMemberID, arrData.CryptoCode, true, false)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  arrData.CryptoCode,
			}
			base.LogErrorLog("ProcessMemberSellTradingRequestv2_GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr, arrData.EntMemberID)
		bal := BlkCWalBal.Balance

		// totalAmount := arrData.Quantity * arrData.UnitPrice
		totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

		totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", "")
		totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
		if err != nil {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
		totalAmount = totalAmountFloat

		// bal, _, _ := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr)
		// fmt.Println("bal:", bal)

		// start checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]
		// if strings.ToLower(arrData.CryptoCode) == "sec" || strings.ToLower(arrData.CryptoCode) == "liga" {

		// 	arrCond = make([]models.WhereCondFn, 0)
		// 	arrCond = append(arrCond,
		// 		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		// 		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode+"H")},
		// 	)
		// 	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

		// 	if holdingEwtSetup != nil{
		// 		arrHoldingEwtBal := wallet_service.GetWalletBalanceStruct{
		// 			EntMemberID: arrData.EntMemberID,
		// 			EwtTypeID:   holdingEwtSetup.ID,
		// 		}
		// 		holdingWalletBalance := wallet_service.GetWalletBalance(arrHoldingEwtBal)
		// 		if holdingWalletBalance.Balance > 0 {
		// 			bal = bal - holdingWalletBalance.Balance
		// 		}
		// 	}
		// }
		// end checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]

		if bal < arrData.Quantity { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_sell"}
		}

		// start send signing key to blockchain site
		// hashValue, errMsg := wallet_service.SignedTransaction(arrData.SigningKey)

		// if errMsg != "" {
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		// }
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:       arrData.EntMemberID,
			EwalletTypeID:     tradingSetup[0].IDFrom,
			DocNo:             docNoS,
			Status:            "P",
			TransactionType:   "TRADING_SELL",
			TransactionData:   arrData.SigningKey,
			TotalOut:          arrData.Quantity,
			ConversionRate:    arrData.UnitPrice,
			ConvertedTotalOut: totalAmount,
			LogOnly:           0,
			Remark:            docNoS,
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		hashValue := saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site

		if strings.ToLower(arrData.CryptoCode) == "sec" || strings.ToLower(arrData.CryptoCode) == "liga" {
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
				models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode + "H")},
			)
			holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

			if holdingEwtSetup != nil {
				// start add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
				ewtIn := wallet_service.SaveMemberWalletStruct{
					EntMemberID:     arrData.EntMemberID,
					EwalletTypeID:   holdingEwtSetup.ID,
					TotalIn:         arrData.Quantity,
					TransactionType: "TRADING_SELL",
					DocNo:           docNoS,
					Remark:          docNoS,
					CreatedBy:       strconv.Itoa(arrData.EntMemberID),
					// Remark:          "#*sell_trading_request*#" + " " + docNoS,
				}

				_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

				if err != nil {
					base.LogErrorLog("ProcessMemberSellTradingRequestv2_SaveMemberWallet_failed", err.Error(), ewtIn, true)
					return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
				}
				// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
			}
		}

		arrCrtTradSell := models.TradingSell{
			CryptoCode:         arrData.CryptoCode,
			CryptoCodeTo:       tradingSetup[0].CodeTo,
			DocNo:              docNoS,
			MemberID:           arrData.EntMemberID,
			TotalUnit:          arrData.Quantity,
			SuggestedUnitPrice: arrData.UnitPrice,
			UnitPrice:          arrData.UnitPrice,
			TotalAmount:        totalAmount,
			BalanceUnit:        arrData.Quantity,
			Status:             "",
			SigningKey:         arrData.SigningKey,
			TransHash:          hashValue,
			CreatedBy:          strconv.Itoa(arrData.EntMemberID),
			ApprovedAt:         base.GetCurrentDateTimeT(),
			ApprovedBy:         strconv.Itoa(arrData.EntMemberID),
		}

		_, err = models.AddTradingSell(tx, arrCrtTradSell)
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv2_failed_to_save_trading_sell", err.Error(), arrCrtTradSell, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		err = models.UpdateRunningDocNo(docTypeS, tx) //update doc no
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv2_failed_in_UpdateRunningDocNo", err.Error(), docTypeS, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
	} else {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_not_available"}
	}

	return helpers.CutOffDecimal(arrData.Quantity, uint(tradingSetup[0].DecimalPointFrom), ".", ","), nil
}

// type CancelMemberTradingSellRequestStruct
type CancelMemberTradingSellRequestStruct struct {
	SellID      int
	Quantity    float64
	EntMemberID int
}

// func ProcessMemberCancelTradingSellRequestv1
func ProcessMemberCancelTradingSellRequestv1(tx *gorm.DB, arrData CancelMemberTradingSellRequestStruct) error {

	dtNow := base.GetCurrentTime("2006-01-02 15:04:05")
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.status = ? ", CondValue: "P"},
		models.WhereCondFn{Condition: " trading_sell.id = ? ", CondValue: arrData.SellID},
		models.WhereCondFn{Condition: " trading_sell.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_sell.balance_unit > ? ", CondValue: 0},
	)
	arrTrading, _ := models.GetTradingSellFn(arrCond, false)

	if len(arrTrading) != 1 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrTrading[0].CryptoCode},
	)
	tradingSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingSetup == nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	// start checking for cancel quantity
	if arrTrading[0].BalanceUnit < arrData.Quantity {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_cancel"}
	}
	// end checking for cancel quantity

	docTypeC := "TRADC"
	docNoC, err := models.GetRunningDocNo(docTypeC, tx) //get doc no

	if err != nil {
		base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-get_trading_cancel_doc_no_failed", err.Error(), nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	// totalAmount := arrData.Quantity * arrTrading[0].UnitPrice
	totalAmount := float.Mul(arrData.Quantity, arrTrading[0].UnitPrice)
	balanceUnit := arrTrading[0].BalanceUnit - arrData.Quantity

	var hashValue string
	var tradeCancelSigningKey string
	updateTradBuyColumn := map[string]interface{}{}

	if strings.ToLower(tradingSetup.Control) == "blockchain" {

		// start get hotwallet
		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end get hotwallet

		// start check hotwallet balance
		balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(arrTrading[0].CryptoCode, hotWalletInfo.HotWalletAddress)
		if err != nil {
			base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-GetBlockchainWalletBalanceApiV1_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if balance < arrData.Quantity {
			arrErr := map[string]interface{}{
				"cryptoType":       arrTrading[0].CryptoCode,
				"hotWalletBalance": balance,
				"quantityNeed":     arrData.Quantity,
				"sellID":           arrTrading[0].ID,
			}
			base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-hotwallet_balance_is_not_enough", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end check hotwallet balance

		// start sign transaction for blockchain (from hotwallet to member account)
		tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

		chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
		chainIDInt64 := int64(chainID)
		maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
		maxGasUint64 := uint64(maxGas)

		cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrTrading[0].MemberID, arrTrading[0].CryptoCode, true, false)
		if err != nil {
			arrErr := map[string]interface{}{
				"MemberID":   arrTrading[0].MemberID,
				"CryptoCode": arrTrading[0].CryptoCode,
				"sellID":     arrTrading[0].ID,
			}
			base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
			TokenType:       arrTrading[0].CryptoCode,
			PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
			ContractAddress: tradingSetup.ContractAddress,
			ChainID:         chainIDInt64,
			FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
			ToAddr:          cryptoAddr,                     // this is refer to the buyer address
			Amount:          arrData.Quantity,
			MaxGas:          maxGasUint64,
		}
		signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-ProcecssGenerateSignTransaction_failed", err.Error(), arrProcecssGenerateSignTransaction, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		tradeCancelSigningKey = signingKey
		// end sign transaction for blockchain (from hotwallet to member account)
		// for debug purposes
		// fmt.Println("arrData:", arrData)
		// fmt.Println("arrProcecssGenerateSignTransaction:", arrProcecssGenerateSignTransaction)
		// base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-ProcecssGenerateSignTransaction_debug", arrData, arrProcecssGenerateSignTransaction, true)

		// start send signed transaction to blockchain (from hotwallet to member account)
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:      arrTrading[0].MemberID,
			EwalletTypeID:    tradingSetup.ID,
			DocNo:            docNoC,
			Status:           "P",
			TransactionType:  "TRADING_CANCEL",
			TransactionData:  tradeCancelSigningKey,
			TotalIn:          arrData.Quantity,
			ConversionRate:   arrTrading[0].UnitPrice,
			ConvertedTotalIn: totalAmount,
			LogOnly:          0, // take it as log only just in case error is happened in blockchain site.
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-SignedTransaction_failed", errMsg, signingKey+" sell_id:"+strconv.Itoa(arrTrading[0].ID), true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signed transaction to blockchain (from hotwallet to member account)
		// end send signing key to blockchain site
	} else if strings.ToLower(tradingSetup.Control) == "internal" {

		arrEwtIn := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrTrading[0].MemberID,
			EwalletTypeID:   tradingSetup.ID,
			TotalIn:         arrData.Quantity,
			TransactionType: "TRADING_CANCEL",
			DocNo:           docNoC,
			Remark:          docNoC,
			CreatedBy:       strconv.Itoa(arrTrading[0].MemberID),
		}

		_, err = wallet_service.SaveMemberWallet(tx, arrEwtIn)

		if err != nil {
			base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-SaveMemberWallet_failed", err.Error(), arrEwtIn, true)
			return err
		}
		if balanceUnit == 0 {
			updateTradBuyColumn["status"] = "C"
		}
		updateTradBuyColumn["approved_at"] = dtNow
		updateTradBuyColumn["approved_by"] = arrData.EntMemberID

		// start add holding wallet for either holding sec / holding liga [cancel partial / off prev trading sell amount] [this part move to trading callback function]
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingSetup.EwtTypeCode + "H"},
		)
		tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

		if tradingSetupHolding != nil {
			ewtOut := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     arrData.EntMemberID,
				EwalletTypeID:   tradingSetupHolding.ID,
				TotalOut:        arrData.Quantity,
				TransactionType: "TRADING_CANCEL",
				DocNo:           docNoC,
				Remark:          docNoC,
				CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			}

			_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

			if err != nil {
				base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-failed_to_SaveMemberWallet_for_holding_wallet", err.Error(), ewtOut, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		}
		// end holding wallet for either holding sec / holding liga [cancel partial / off prev trading sell amount] [this part move to trading callback function]

		base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-invalid_tradingSetup.Control", "sell_id:"+strconv.Itoa(arrTrading[0].ID), tradingSetup.Control, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1-invalid_tradingSetup.Control", "sell_id:"+strconv.Itoa(arrTrading[0].ID), tradingSetup.Control, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCrtTradCancel := models.AddTradingCancelStruct{
		TradingID:       arrTrading[0].ID,
		MemberID:        arrData.EntMemberID,
		DocNo:           docNoC,
		TransactionType: "SELL",
		CryptoCode:      arrTrading[0].CryptoCode,
		TotalUnit:       arrData.Quantity,
		UnitPrice:       arrTrading[0].UnitPrice,
		TotalAmount:     totalAmount,
		SigningKey:      tradeCancelSigningKey,
		TransHash:       hashValue,
		CreatedBy:       strconv.Itoa(arrData.EntMemberID),
	}

	// ApprovedAt: base.GetCurrentDateTimeT(),
	// ApprovedBy: strconv.Itoa(arrData.EntMemberID),

	_, err = models.AddTradingCancel(tx, arrCrtTradCancel)
	if err != nil {
		base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1_failed_to_save_trading_cancel", err.Error(), arrCrtTradCancel, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docTypeC, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberCancelTradingSellRequestv1_failed_in_UpdateRunningDocNo", err.Error(), docTypeC, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrTrading[0].ID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "P"},
	)

	updateTradBuyColumn["balance_unit"] = balanceUnit
	updateTradBuyColumn["updated_by"] = arrData.EntMemberID

	err = models.UpdatesFnTx(tx, "trading_sell", arrUpdCond, updateTradBuyColumn, false)
	if err != nil {
		base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-update_trading_sell_failed", "update_balance_unit_in_trading_sell", err.Error(), true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}

// func ProcessMemberSellTradingRequestv3 - this will cover diff unit_price
func ProcessMemberSellTradingRequestv3(tx *gorm.DB, arrData SellMemberTradingRequestStruct) (string, error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)
	tradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	if len(tradingSetup) < 1 {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	// start checking on close whole network
	if tradingSetup[0].TradingSellStatus != 1 {
		var openTradingSellStatus bool
		// start checking on only open certain network
		if tradingSetup[0].TradingSellOpenSponsorID != "" {
			arrTargetStringID := strings.Split(tradingSetup[0].TradingSellOpenSponsorID, ",")
			arrTargetID := make([]int, 0)
			for _, arrTargetStringIDV := range arrTargetStringID {
				targetID, _ := strconv.Atoi(arrTargetStringIDV)
				arrTargetID = append(arrTargetID, targetID)
			}
			arrNearestUpline, err := models.GetNearestUplineByMemId(arrData.EntMemberID, arrTargetID, "", false)

			if err != nil {
				arrErr := map[string]interface{}{
					"MemberID":    arrData.EntMemberID,
					"arrTargetID": arrTargetID,
				}
				base.LogErrorLog("ProcessMemberSellTradingRequestv3-GetNearestUplineByMemId_failed", err.Error(), arrErr, true)
			}

			if arrNearestUpline != nil {
				openTradingSellStatus = true
			}
		}

		if !openTradingSellStatus {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "trading_is_not_available"}
		}
		// end checking on only open certain network
	}
	// end checking on close whole network

	docTypeS := "TRADS"
	docNoS, err := models.GetRunningDocNo(docTypeS, tx) //get withdraw doc no

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingRequestv3-get_trading_sell_doc_no_failed", err.Error(), nil, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	var hashValue string
	var tradSellStatus string

	suggestedPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	// totalAmount := arrData.Quantity * arrData.UnitPrice
	totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

	totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", "")
	totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
	if err != nil {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}
	totalAmount = totalAmountFloat

	if strings.ToLower(tradingSetup[0].ControlFrom) == "blockchain" {

		if arrData.SigningKey == "" {
			arrErr := map[string]interface{}{
				"trading_setup": tradingSetup,
				"arrData":       arrData,
			}
			base.LogErrorLog("ProcessMemberSellTradingRequestv3-signing_key_is_missing", arrErr, arrCond, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrData.EntMemberID, arrData.CryptoCode, true, false)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  arrData.CryptoCode,
			}
			base.LogErrorLog("ProcessMemberSellTradingRequestv3-GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr, arrData.EntMemberID)
		bal := BlkCWalBal.Balance

		// bal, _, _ := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr)
		// fmt.Println("bal:", bal)

		// start checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]
		// if strings.ToLower(arrData.CryptoCode) == "sec" || strings.ToLower(arrData.CryptoCode) == "liga" {

		// 	arrCond = make([]models.WhereCondFn, 0)
		// 	arrCond = append(arrCond,
		// 		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		// 		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode+"H")},
		// 	)
		// 	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

		// 	if holdingEwtSetup != nil{
		// 		arrHoldingEwtBal := wallet_service.GetWalletBalanceStruct{
		// 			EntMemberID: arrData.EntMemberID,
		// 			EwtTypeID:   holdingEwtSetup.ID,
		// 		}
		// 		holdingWalletBalance := wallet_service.GetWalletBalance(arrHoldingEwtBal)
		// 		if holdingWalletBalance.Balance > 0 {
		// 			bal = bal - holdingWalletBalance.Balance
		// 		}
		// 	}
		// }
		// end checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]

		if bal < arrData.Quantity { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_sell"}
		}

		// start send signing key to blockchain site
		// hashValue, errMsg := wallet_service.SignedTransaction(arrData.SigningKey)

		// if errMsg != "" {
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		// }
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:       arrData.EntMemberID,
			EwalletTypeID:     tradingSetup[0].IDFrom,
			DocNo:             docNoS,
			Status:            "P",
			TransactionType:   "TRADING_SELL",
			TransactionData:   arrData.SigningKey,
			TotalOut:          arrData.Quantity,
			ConversionRate:    arrData.UnitPrice,
			ConvertedTotalOut: totalAmount,
			LogOnly:           0,
			Remark:            docNoS,
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site
	} else {
		// start check balance
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDFrom,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}
		// end check balance

		// start deduct wallet
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDFrom,
			TotalOut:        arrData.Quantity,
			TransactionType: "TRADING_SELL",
			DocNo:           docNoS,
			Remark:          docNoS,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv3-failed_to_SaveMemberWallet2", err.Error(), ewtOut, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end deduct wallet

		tradSellStatus = "P"
	}

	// start credit in holding wallet if it is exist in ewt_setup
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingSetup[0].CodeFrom + "H")},
	)
	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if holdingEwtSetup != nil {
		// start add holding wallet for holding wallet (bcz this trading transaction is not match yet)
		ewtIn := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   holdingEwtSetup.ID,
			TotalIn:         arrData.Quantity,
			TransactionType: "TRADING_SELL",
			DocNo:           docNoS,
			Remark:          docNoS,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*sell_trading_request*#" + " " + docNoS,
		}

		_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv3-SaveMemberWallet2_failed", err.Error(), ewtIn, true)
			return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
		}
		// end add holding wallet for holding wallet (bcz this trading transaction is not match yet)
	}
	// end credit in holding wallet if it is exist in ewt_setup

	arrCrtTradSell := models.TradingSell{
		CryptoCode:         arrData.CryptoCode,
		CryptoCodeTo:       tradingSetup[0].CodeTo,
		DocNo:              docNoS,
		MemberID:           arrData.EntMemberID,
		TotalUnit:          arrData.Quantity,
		SuggestedUnitPrice: suggestedPrice,
		UnitPrice:          arrData.UnitPrice,
		TotalAmount:        totalAmount,
		BalanceUnit:        arrData.Quantity,
		Status:             tradSellStatus,
		SigningKey:         arrData.SigningKey,
		TransHash:          hashValue,
		CreatedBy:          strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:         base.GetCurrentDateTimeT(),
		ApprovedBy:         strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingSell(tx, arrCrtTradSell)
	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingRequestv3-failed_to_save_trading_sell", err.Error(), arrCrtTradSell, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docTypeS, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingRequestv3-failed_in_UpdateRunningDocNo", err.Error(), docTypeS, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return helpers.CutOffDecimal(arrData.Quantity, uint(tradingSetup[0].DecimalPointFrom), ".", ","), nil
}

// func ProcessAutoTradingSellRequestv1 - this will cover diff unit_price
func ProcessAutoTradingSellRequestv1(tx *gorm.DB, arrData SellMemberTradingRequestStruct) (string, error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)
	tradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	if len(tradingSetup) < 1 {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	docTypeS := "TRADS"
	docNoS, err := models.GetRunningDocNo(docTypeS, tx) //get withdraw doc no

	if err != nil {
		base.LogErrorLog("ProcessAutoTradingSellRequestv1-get_trading_selsl_doc_no_failed", err.Error(), nil, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	var hashValue string
	var tradSellStatus string

	suggestedPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	// totalAmount := arrData.Quantity * arrData.UnitPrice
	totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

	totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", "")
	totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
	if err != nil {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}
	totalAmount = totalAmountFloat

	if strings.ToLower(tradingSetup[0].ControlFrom) == "blockchain" {

		tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

		chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
		chainIDInt64 := int64(chainID)
		maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
		maxGasUint64 := uint64(maxGas)

		entMemberCrypto := tradingSetup[0].CodeFrom
		if strings.ToLower(tradingSetup[0].CodeFrom) == "liga" || strings.ToLower(tradingSetup[0].CodeFrom) == "sec" {
			entMemberCrypto = "sec"
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.EntMemberID},
			models.WhereCondFn{Condition: " crypto_type = ? ", CondValue: entMemberCrypto},
		)

		cryptoAddr, err := models.GetEntMemberCryptoFn(arrCond, false)
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingSellRequestv1-GetEntMemberCryptoFn_failed", err, arrCond, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(tradingSetup[0].CodeFrom, cryptoAddr.CryptoAddress, arrData.EntMemberID)
		bal := BlkCWalBal.Balance

		if bal < arrData.Quantity { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_sell"}
		}

		nonceApi, err := wallet_service.GetTransactionNonceViaAPI(cryptoAddr.CryptoAddress) // // this is refer to the hotwallet addr
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingSellRequestv1-GetTransactionNonceViaAPI_failed", err.Error(), cryptoAddr.CryptoAddress, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "ProcessAutoTradingSellRequestv1-GetTransactionNonceViaAPI_failed"}
		}
		// start get hotwallet
		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingSellRequestv1-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end get hotwallet
		arrGenerateSignTransaction := wallet_service.GenerateSignTransactionStruct{
			TokenType:       tradingSetup[0].CodeFrom,
			PrivateKey:      cryptoAddr.PrivateKey,
			ContractAddress: tradingSetup[0].ContractAddrFrom,
			ChainID:         chainIDInt64,
			Nonce:           uint64(nonceApi),
			ToAddr:          hotWalletInfo.HotWalletAddress,
			Amount:          arrData.Quantity, // this is refer to amount for this transaction
			MaxGas:          maxGasUint64,
		}
		// base.LogErrorLog("ProcecssGenerateSignTransaction_net2", arrGenerateSignTransaction, nonce, true)
		signingKey, err := wallet_service.GenerateSignTransaction(arrGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingSellRequestv1-GenerateSignTransaction_failed", err.Error(), arrGenerateSignTransaction, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "ProcessAutoTradingSellRequestv1-GetTransactionNonceViaAPI_failed"}
		}
		// end sign transaction for blockchain (from hotwallet to member account)

		arrData.SigningKey = signingKey
		// start send signing key to blockchain site
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:       arrData.EntMemberID,
			EwalletTypeID:     tradingSetup[0].IDFrom,
			DocNo:             docNoS,
			Status:            "P",
			TransactionType:   "TRADING_SELL",
			TransactionData:   signingKey,
			TotalOut:          arrData.Quantity,
			ConversionRate:    arrData.UnitPrice,
			ConvertedTotalOut: totalAmount,
			LogOnly:           0,
			Remark:            docNoS,
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site
	} else {
		// start check balance
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDFrom,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}
		// end check balance

		// start deduct wallet
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDFrom,
			TotalOut:        arrData.Quantity,
			TransactionType: "TRADING_SELL",
			DocNo:           docNoS,
			Remark:          docNoS,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessAutoTradingSellRequestv1-failed_to_SaveMemberWallet2", err.Error(), ewtOut, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end deduct wallet

		tradSellStatus = "P"

	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode + "H")},
	)
	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if holdingEwtSetup != nil {
		// start add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
		ewtIn := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   holdingEwtSetup.ID,
			TotalIn:         arrData.Quantity,
			TransactionType: "TRADING_SELL",
			DocNo:           docNoS,
			Remark:          docNoS,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*sell_trading_request*#" + " " + docNoS,
		}

		_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

		if err != nil {
			base.LogErrorLog("ProcessAutoTradingSellRequestv1-SaveMemberWallet_failed", err.Error(), ewtIn, true)
			return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
		}
		// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
	}

	arrCrtTradSell := models.TradingSell{
		CryptoCode:         arrData.CryptoCode,
		CryptoCodeTo:       tradingSetup[0].CodeTo,
		DocNo:              docNoS,
		MemberID:           arrData.EntMemberID,
		TotalUnit:          arrData.Quantity,
		SuggestedUnitPrice: suggestedPrice,
		UnitPrice:          arrData.UnitPrice,
		TotalAmount:        totalAmount,
		BalanceUnit:        arrData.Quantity,
		Status:             tradSellStatus,
		SigningKey:         arrData.SigningKey,
		TransHash:          hashValue,
		CreatedBy:          strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:         base.GetCurrentDateTimeT(),
		ApprovedBy:         strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingSell(tx, arrCrtTradSell)
	if err != nil {
		base.LogErrorLog("ProcessAutoTradingSellRequestv1-failed_to_save_trading_sell", err.Error(), arrCrtTradSell, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docTypeS, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessAutoTradingSellRequestv1-failed_in_UpdateRunningDocNo", err.Error(), docTypeS, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	return helpers.CutOffDecimal(arrData.Quantity, uint(tradingSetup[0].DecimalPointFrom), ".", ","), nil
}

// func ProcessCancelAutoTradingSellRequestv1
func ProcessCancelAutoTradingSellRequestv1(tx *gorm.DB, arrData ProcessCancelAutoTradingRequestForm) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.status = ? ", CondValue: "P"},
		models.WhereCondFn{Condition: " trading_sell.doc_no = ? ", CondValue: arrData.DocNo},
		models.WhereCondFn{Condition: " trading_sell.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_sell.balance_unit > ? ", CondValue: 0},
	)
	arrTrading, _ := models.GetTradingSellFn(arrCond, false)

	if len(arrTrading) != 1 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrTrading[0].CryptoCode},
	)
	tradingSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingSetup == nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	quantity := arrTrading[0].BalanceUnit
	// start checking for cancel quantity
	// if arrTrading[0].BalanceUnit < quantity {
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_cancel"}
	// }
	// end checking for cancel quantity

	docTypeC := "TRADC"
	docNoC, err := models.GetRunningDocNo(docTypeC, tx) //get doc no

	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-get_trading_cancel_doc_no_failed", err.Error(), nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	// totalAmount := quantity * arrTrading[0].UnitPrice
	totalAmount := float.Mul(quantity, arrTrading[0].UnitPrice)
	balanceUnit := arrTrading[0].BalanceUnit - quantity

	var hashValue string
	var tradeCancelSigningKey string
	updateTradBuyColumn := map[string]interface{}{}

	if strings.ToLower(tradingSetup.Control) == "blockchain" {

		// start get hotwallet
		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end get hotwallet

		// start check hotwallet balance
		balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(arrTrading[0].CryptoCode, hotWalletInfo.HotWalletAddress)
		if err != nil {
			base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-GetBlockchainWalletBalanceApiV1_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if balance < quantity {
			arrErr := map[string]interface{}{
				"cryptoType":       arrTrading[0].CryptoCode,
				"hotWalletBalance": balance,
				"quantityNeed":     quantity,
				"sellID":           arrTrading[0].ID,
			}
			base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-hotwallet_balance_is_not_enough", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end check hotwallet balance

		// start sign transaction for blockchain (from hotwallet to member account)
		tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

		chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
		chainIDInt64 := int64(chainID)
		maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
		maxGasUint64 := uint64(maxGas)

		cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrTrading[0].MemberID, arrTrading[0].CryptoCode, true, false)
		if err != nil {
			arrErr := map[string]interface{}{
				"MemberID":   arrTrading[0].MemberID,
				"CryptoCode": arrTrading[0].CryptoCode,
				"sellID":     arrTrading[0].ID,
			}
			base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
			TokenType:       arrTrading[0].CryptoCode,
			PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
			ContractAddress: tradingSetup.ContractAddress,
			ChainID:         chainIDInt64,
			FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
			ToAddr:          cryptoAddr,                     // this is refer to the buyer address
			Amount:          quantity,
			MaxGas:          maxGasUint64,
		}
		signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-ProcecssGenerateSignTransaction_failed", err.Error(), arrProcecssGenerateSignTransaction, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		tradeCancelSigningKey = signingKey
		// end sign transaction for blockchain (from hotwallet to member account)
		// for debug purposes
		// fmt.Println("arrData:", arrData)
		// fmt.Println("arrProcecssGenerateSignTransaction:", arrProcecssGenerateSignTransaction)
		// base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-ProcecssGenerateSignTransaction_debug", arrData, arrProcecssGenerateSignTransaction, true)

		// start send signed transaction to blockchain (from hotwallet to member account)
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:      arrTrading[0].MemberID,
			EwalletTypeID:    tradingSetup.ID,
			DocNo:            docNoC,
			Status:           "P",
			TransactionType:  "TRADING_CANCEL",
			TransactionData:  tradeCancelSigningKey,
			TotalIn:          quantity,
			ConversionRate:   arrTrading[0].UnitPrice,
			ConvertedTotalIn: totalAmount,
			LogOnly:          0, // take it as log only just in case error is happened in blockchain site.
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-SignedTransaction_failed", errMsg, signingKey+" sell_id:"+strconv.Itoa(arrTrading[0].ID), true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signed transaction to blockchain (from hotwallet to member account)

		// start add holding wallet for either holding sec / holding liga [cancel partial / off prev trading sell amount] [this part move to trading callback function]
		// arrCond = make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		// 	models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingSetup.EwtTypeCode + "H"},
		// )
		// tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

		// if tradingSetupHolding != nil {
		// 	ewtOut := wallet_service.SaveMemberWalletStruct{
		// 		EntMemberID:     arrData.EntMemberID,
		// 		EwalletTypeID:   tradingSetupHolding.ID,
		// 		TotalOut:        quantity,
		// 		TransactionType: "TRADING",
		// 		DocNo:           docNoC,
		// 		Remark:          "#*sell_trading_cancel*# " + docNoC,
		// 		CreatedBy:       strconv.Itoa(arrData.EntMemberID),
		// 	}

		// 	_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		// 	if err != nil {
		// 		base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-failed_to_SaveMemberWallet_for_holding_wallet", err.Error(), ewtOut, true)
		// 		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// 	}
		// }
		// start end holding wallet for either holding sec / holding liga [cancel partial / off prev trading sell amount] [this part move to trading callback function]

		// end send signing key to blockchain site
	} else if strings.ToLower(tradingSetup.Control) == "internal" {
		// start so far no such thing yet
		base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-invalid_tradingSetup.Control", "sell_id:"+strconv.Itoa(arrTrading[0].ID), tradingSetup.Control, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-invalid_tradingSetup.Control", "sell_id:"+strconv.Itoa(arrTrading[0].ID), tradingSetup.Control, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCrtTradCancel := models.AddTradingCancelStruct{
		TradingID:       arrTrading[0].ID,
		MemberID:        arrData.EntMemberID,
		DocNo:           docNoC,
		TransactionType: "SELL",
		CryptoCode:      arrTrading[0].CryptoCode,
		TotalUnit:       quantity,
		UnitPrice:       arrTrading[0].UnitPrice,
		TotalAmount:     totalAmount,
		SigningKey:      tradeCancelSigningKey,
		TransHash:       hashValue,
		CreatedBy:       strconv.Itoa(arrData.EntMemberID),
	}

	// ApprovedAt: base.GetCurrentDateTimeT(),
	// ApprovedBy: strconv.Itoa(arrData.EntMemberID),

	_, err = models.AddTradingCancel(tx, arrCrtTradCancel)
	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-failed_to_save_trading_cancel", err.Error(), arrCrtTradCancel, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docTypeC, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-failed_in_UpdateRunningDocNo", err.Error(), docTypeC, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrTrading[0].ID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "P"},
	)

	updateTradBuyColumn["balance_unit"] = balanceUnit
	updateTradBuyColumn["updated_by"] = arrData.EntMemberID

	err = models.UpdatesFnTx(tx, "trading_sell", arrUpdCond, updateTradBuyColumn, false)
	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingSellRequestv1-update_trading_sell_failed", "update_balance_unit_in_trading_sell", err.Error(), true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}

type SellMemberTradingRequestv2Struct struct {
	UnitPrice   float64
	Quantity    float64
	EntMemberID int
	CryptoCode  string
}

// func ProcessMemberSellTradingRequestv4 - this will cover diff unit_price - without signing key
func ProcessMemberSellTradingRequestv4(tx *gorm.DB, arrData SellMemberTradingRequestv2Struct) (string, error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " trading_setup.code_from = ? ", CondValue: arrData.CryptoCode},
	)
	tradingSetup, _ := models.GetTradingSetupFn(arrCond, false)

	if len(tradingSetup) < 1 {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	docTypeS := "TRADS"
	docNoS, err := models.GetRunningDocNo(docTypeS, tx) //get trading sell doc no

	if err != nil {
		base.LogErrorLog("ProcessMemberSellTradingRequestv4-get_trading_sell_doc_no_failed", err.Error(), nil, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	suggestedPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	var signingKey string

	if strings.ToLower(tradingSetup[0].ControlFrom) == "blockchain" {
		memberCryptoInfo, err := models.GetCustomMemberCryptoInfov2(arrData.EntMemberID, arrData.CryptoCode, true, false)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  arrData.CryptoCode,
			}
			base.LogErrorLog("ProcessMemberSellTradingRequestv4-GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		cryptoAddr := memberCryptoInfo.CryptoAddr

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr, arrData.EntMemberID)
		bal := BlkCWalBal.AvailableBalance

		if bal < arrData.Quantity { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_sell"}
		}

		// totalAmount := arrData.Quantity * arrData.UnitPrice
		totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

		totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointFrom), ".", "")
		totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
		if err != nil {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}

		totalAmount = totalAmountFloat

		signedKeySetup, errMsg := wallet_service.GetSigningKeySettingByModule(tradingSetup[0].CodeFrom, cryptoAddr, "TRADING")
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		signedKeySetup["decimal_point"] = tradingSetup[0].BlockchainDecimalPointFrom

		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv4-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}

		chainIDString := signedKeySetup["chain_id"].(string)
		chainIDInt64, _ := strconv.ParseInt(chainIDString, 10, 64)

		maxGasString := signedKeySetup["max_gas"].(string)
		maxGasint64, _ := strconv.ParseInt(maxGasString, 10, 64)
		maxGasUint64 := uint64(maxGasint64)

		// start generate sign tranx
		arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
			TokenType:       tradingSetup[0].CodeFrom,
			PrivateKey:      memberCryptoInfo.PrivateKey,
			ContractAddress: tradingSetup[0].ContractAddrFrom,
			ChainID:         chainIDInt64,
			FromAddr:        cryptoAddr,                     // this is refer to the member addr
			ToAddr:          hotWalletInfo.HotWalletAddress, // this is refer to the hot wallet address
			Amount:          arrData.Quantity,
			MaxGas:          maxGasUint64,
		}
		signingKeyRst, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv4-ProcecssGenerateSignTransaction_failed", err.Error(), arrProcecssGenerateSignTransaction, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}
		signingKey = signingKeyRst

		// bal, _, _ := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr)
		// fmt.Println("bal:", bal)

		// start checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]
		// if strings.ToLower(arrData.CryptoCode) == "sec" || strings.ToLower(arrData.CryptoCode) == "liga" {

		// 	arrCond = make([]models.WhereCondFn, 0)
		// 	arrCond = append(arrCond,
		// 		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		// 		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode+"H")},
		// 	)
		// 	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

		// 	if holdingEwtSetup != nil{
		// 		arrHoldingEwtBal := wallet_service.GetWalletBalanceStruct{
		// 			EntMemberID: arrData.EntMemberID,
		// 			EwtTypeID:   holdingEwtSetup.ID,
		// 		}
		// 		holdingWalletBalance := wallet_service.GetWalletBalance(arrHoldingEwtBal)
		// 		if holdingWalletBalance.Balance > 0 {
		// 			bal = bal - holdingWalletBalance.Balance
		// 		}
		// 	}
		// }
		// end checking on holding wallet. open this if GetBlockchainWalletBalanceByAddressV1 does not have data [available balance]

		// start send signing key to blockchain site
		// hashValue, errMsg := wallet_service.SignedTransaction(arrData.SigningKey)

		// if errMsg != "" {
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		// }
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:       arrData.EntMemberID,
			EwalletTypeID:     tradingSetup[0].IDFrom,
			DocNo:             docNoS,
			Status:            "P",
			TransactionType:   "TRADING_SELL",
			TransactionData:   signingKey,
			TotalOut:          arrData.Quantity,
			ConversionRate:    arrData.UnitPrice,
			ConvertedTotalOut: totalAmount,
			LogOnly:           0,
			Remark:            docNoS,
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		hashValue := saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site

		if strings.ToLower(arrData.CryptoCode) == "sec" || strings.ToLower(arrData.CryptoCode) == "liga" {
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
				models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(arrData.CryptoCode + "H")},
			)
			holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

			if holdingEwtSetup != nil {
				// start add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
				ewtIn := wallet_service.SaveMemberWalletStruct{
					EntMemberID:     arrData.EntMemberID,
					EwalletTypeID:   holdingEwtSetup.ID,
					TotalIn:         arrData.Quantity,
					TransactionType: "TRADING_SELL",
					DocNo:           docNoS,
					Remark:          docNoS,
					CreatedBy:       strconv.Itoa(arrData.EntMemberID),
					// Remark:          "#*sell_trading_request*#" + " " + docNoS,
				}

				_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

				if err != nil {
					base.LogErrorLog("ProcessMemberSellTradingRequestv4-SaveMemberWallet_failed", err.Error(), ewtIn, true)
					return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
				}
				// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
			}
		}

		arrCrtTradSell := models.TradingSell{
			CryptoCode:         arrData.CryptoCode,
			CryptoCodeTo:       tradingSetup[0].CodeTo,
			DocNo:              docNoS,
			MemberID:           arrData.EntMemberID,
			TotalUnit:          arrData.Quantity,
			SuggestedUnitPrice: suggestedPrice,
			UnitPrice:          arrData.UnitPrice,
			TotalAmount:        totalAmount,
			BalanceUnit:        arrData.Quantity,
			Status:             "",
			SigningKey:         signingKey,
			TransHash:          hashValue,
			CreatedBy:          strconv.Itoa(arrData.EntMemberID),
			ApprovedAt:         base.GetCurrentDateTimeT(),
			ApprovedBy:         strconv.Itoa(arrData.EntMemberID),
		}

		_, err = models.AddTradingSell(tx, arrCrtTradSell)
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv4-failed_to_save_trading_sell", err.Error(), arrCrtTradSell, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		err = models.UpdateRunningDocNo(docTypeS, tx) //update doc no
		if err != nil {
			base.LogErrorLog("ProcessMemberSellTradingRequestv4-failed_in_UpdateRunningDocNo", err.Error(), docTypeS, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
	} else {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sell_trading_not_available"}
	}

	return helpers.CutOffDecimal(arrData.Quantity, uint(tradingSetup[0].DecimalPointFrom), ".", ","), nil
}
