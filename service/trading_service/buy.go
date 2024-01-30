package trading_service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/service/wallet_service"

	"github.com/jinzhu/gorm"
)

type BuyMemberTradingStruct struct {
	SellID      int
	Quantity    float64
	EntMemberID int
	LangCode    string
}

// func ProcessMemberBuyTradingv1
func ProcessMemberBuyTradingv1(tx *gorm.DB, arrData BuyMemberTradingStruct) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.id = ? ", CondValue: arrData.SellID},
		models.WhereCondFn{Condition: " trading_sell.status = ? ", CondValue: "P"},
	)

	arrTradingSell, err := models.GetTradingSellFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1-failed_to_get_GetTradingSellFn", err.Error(), arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	if len(arrTradingSell) < 1 {
		base.LogErrorLog("ProcessMemberBuyTradingv1-invalid_sell_id", arrCond, nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrTradingSell[0].BalanceUnit < arrData.Quantity {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "quantity_not_enough_to_buy"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.code_from = ?", CondValue: arrTradingSell[0].CryptoCode},
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	)
	tradingSetup, err := models.GetTradingSetupFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1-invalid_trading_payment", err.Error(), arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	if len(tradingSetup) < 1 {
		base.LogErrorLog("ProcessMemberBuyTradingv1-wallet_trading_payment_missing", arrCond, nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	totalAmount := arrData.Quantity * arrTradingSell[0].UnitPrice

	docTypeB := "TRADB"
	docNoB, err := models.GetRunningDocNo(docTypeB, tx) //get doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1-GetRunningDocNo_TRADB_failed", err.Error(), docTypeB, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "buy_trading_failed"}
	}

	if strings.ToLower(tradingSetup[0].ControlTo) == "blockchain" {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDTo,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}

		// start check balance
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			TotalOut:        totalAmount,
			TransactionType: "TRADING",
			DocNo:           docNoB,
			Remark:          "#*buy_trading*# " + docNoB,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv1_failed_to_SaveMemberWallet", err.Error(), ewtOut, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end check balance
	}

	balanceUnit := arrTradingSell[0].BalanceUnit - arrData.Quantity
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.id = ? ", CondValue: arrTradingSell[0].ID},
	)

	updateColumn := map[string]interface{}{"balance_unit": balanceUnit, "updated_at": base.GetCurrentTime("2006-01-02 15:04:05")}
	err = models.UpdatesFn("trading_sell", arrCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1-failed_to_update_trading_sell", err.Error(), updateColumn, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "buy_trading_failed"}
	}

	arrCrtTradBuy := models.TradingBuy{
		CryptoCode:  arrTradingSell[0].CryptoCode,
		DocNo:       docNoB,
		MemberID:    arrData.EntMemberID,
		TotalUnit:   arrData.Quantity,
		UnitPrice:   arrTradingSell[0].UnitPrice,
		TotalAmount: totalAmount,
		Status:      "AP",
		CreatedBy:   strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:  base.GetCurrentDateTimeT(),
		ApprovedBy:  strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingBuy(tx, arrCrtTradBuy)
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1-failed_to_save_trading_buy", err.Error(), arrCrtTradBuy, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "buy_trading_failed"}
	}

	docTypeM := "TRADM"
	docNoM, err := models.GetRunningDocNo(docTypeM, tx) //get doc no

	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1-GetRunningDocNo_failed", err.Error(), docTypeM, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "buy_trading_failed"}
	}

	arrCrtTradMatch := models.TradingMatch{
		DocNo:          docNoM,
		CryptoCode:     arrTradingSell[0].CryptoCode,
		SellID:         arrTradingSell[0].ID,
		BuyID:          arrData.EntMemberID,
		SellerMemberID: arrTradingSell[0].MemberID,
		BuyerMemberID:  arrData.EntMemberID,
		TotalUnit:      arrData.Quantity,
		UnitPrice:      arrTradingSell[0].UnitPrice,
		TotalAmount:    totalAmount,
		Status:         "AP",
		CreatedBy:      strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:     base.GetCurrentDateTimeT(),
		ApprovedBy:     strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingMatch(tx, arrCrtTradMatch)
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1-failed_to_save_trading_match", err.Error(), arrCrtTradMatch, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "buy_trading_failed"}
	}

	err = models.UpdateRunningDocNo(docTypeB, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1_failed_in_UpdateRunningDocNo", err.Error(), docTypeB, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docTypeM, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingv1_failed_in_UpdateRunningDocNo", err.Error(), docTypeM, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}

type BuyMemberTradingRequestStruct struct {
	UnitPrice   float64
	Quantity    float64
	EntMemberID int
	CryptoCode  string
	SigningKey  string
}

// func ProcessMemberBuyTradingRequestv1
func ProcessMemberBuyTradingRequestv1(tx *gorm.DB, arrData BuyMemberTradingRequestStruct) error {

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
	docType := "TRADB"
	docNo, err := models.GetRunningDocNo(docType, tx) //get transfer doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv1_GetRunningDocNo_failed", err.Error(), docType, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if strings.ToLower(tradingSetup[0].ControlTo) == "blockchain" {
		// cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrData.MemberID, arrData.CryptoCode, true, false)
		// if err != nil {
		// 	arrErrData := map[string]interface{}{
		// 		"entMemberID": arrData.MemberID,
		// 		"cryptoType":  arrData.CryptoCode,
		// 	}
		// 	base.LogErrorLog("ProcessMemberSellTradingRequestv1_GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// }
		// bal, _, _ := wallet_service.GetBlockchainWalletBalanceByAddressV1(arrData.CryptoCode, cryptoAddr)

		// if bal < arrData.Quantity { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_sell"}
		// }

		// start send signing key to blockchain site
		// hashValue, errMsg := wallet_service.SignedTransaction(arrData.SigningKey)

		// if errMsg != "" {
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		// }
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

		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		// start check balance
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDTo,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}
		// end check balance

		// start deduct wallet
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			TotalOut:        totalAmount,
			TransactionType: "TRADING",
			DocNo:           docNo,
			Remark:          "#*buy_trading_request*# " + docNo,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv1_failed_to_SaveMemberWallet", err.Error(), ewtOut, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end deduct wallet
	}

	arrCrtTradBuy := models.TradingBuy{
		CryptoCode:  arrData.CryptoCode,
		DocNo:       docNo,
		MemberID:    arrData.EntMemberID,
		TotalUnit:   arrData.Quantity,
		UnitPrice:   arrData.UnitPrice,
		TotalAmount: totalAmount,
		Status:      "P",
		CreatedBy:   strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:  base.GetCurrentDateTimeT(),
		ApprovedBy:  strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingBuy(tx, arrCrtTradBuy)
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv1-failed_to_save_trading_buy", err.Error(), arrCrtTradBuy, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "buy_trading_failed"}
	}

	err = models.UpdateRunningDocNo(docType, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv1_failed_in_UpdateRunningDocNo", err.Error(), docType, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}

// func ProcessMemberBuyTradingRequestv2
func ProcessMemberBuyTradingRequestv2(tx *gorm.DB, arrData BuyMemberTradingRequestStruct) (string, error) {

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
	docType := "TRADB"
	docNo, err := models.GetRunningDocNo(docType, tx) //get transfer doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv2_GetRunningDocNo_TRADB_failed", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if strings.ToLower(tradingSetup[0].ControlFrom) == "blockchain" {
		if strings.ToLower(arrData.CryptoCode) == "sec" {
			//get price movement for sec
			tokenRate, err := models.GetLatestSecPriceMovement()

			if err != nil {
				base.LogErrorLog("ProcessMemberBuyTradingRequestv2_GetLatestSecPriceMovement_failed", err.Error(), tokenRate, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			if arrData.UnitPrice != tokenRate {
				arrErr := map[string]interface{}{
					"current_rate":   tokenRate,
					"front_end_rate": arrData.UnitPrice,
				}
				base.LogErrorLog("ProcessMemberBuyTradingRequestv2-front_end_price_not_tally_with_current_price", arrErr, arrData, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		} else if strings.ToLower(arrData.CryptoCode) == "liga" {
			//get price movement for LIGA
			tokenRate, err := models.GetLatestLigaPriceMovement()

			if err != nil {
				base.LogErrorLog("ProcessMemberBuyTradingRequestv2_GetLatestLigaPriceMovement_failed", err.Error(), tokenRate, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			if arrData.UnitPrice != tokenRate {
				arrErr := map[string]interface{}{
					"current_rate":   tokenRate,
					"front_end_rate": arrData.UnitPrice,
				}
				base.LogErrorLog("ProcessMemberBuyTradingRequestv2-front_end_price_not_tally_with_current_price", arrErr, arrData, true)
				return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		}
	}

	// totalAmount := arrData.Quantity * arrData.UnitPrice
	totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

	totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", "")
	totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
	if err != nil {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}
	totalAmount = totalAmountFloat

	var hashValue string
	var tradBuyStatus string

	if strings.ToLower(tradingSetup[0].ControlTo) == "blockchain" {
		if arrData.SigningKey == "" {
			arrErr := map[string]interface{}{
				"trading_setup": tradingSetup,
				"arrData":       arrData,
			}
			base.LogErrorLog("ProcessMemberBuyTradingRequestv2-signing_key_is_missing", arrErr, arrCond, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrData.EntMemberID, tradingSetup[0].CodeTo, true, false)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  tradingSetup[0].CodeTo,
			}
			base.LogErrorLog("ProcessMemberBuyTradingRequestv2-GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(tradingSetup[0].CodeTo, cryptoAddr, arrData.EntMemberID)
		bal := BlkCWalBal.Balance

		if totalAmount > bal { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_buy"}
		}

		// start send signing key to blockchain site
		// hashValue, errMsg := wallet_service.SignedTransaction(arrData.SigningKey)

		// if errMsg != "" {
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		// }
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			DocNo:           docNo,
			Status:          "P",
			TransactionType: "TRADING_BUY",
			TransactionData: arrData.SigningKey,
			TotalOut:        totalAmount,
			ConversionRate:  arrData.UnitPrice,
			Remark:          docNo,
			LogOnly:         0,
			// ConvertedTotalOut: totalAmount,
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site

		// if strings.ToLower(tradingSetup[0].CodeTo) == "usds" {
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingSetup[0].CodeTo + "H")},
		)
		holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

		if holdingEwtSetup != nil {
			// start add holding wallet for USDS holding wallet (bcz this trading transaction is not match yet)
			ewtIn := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     arrData.EntMemberID,
				EwalletTypeID:   holdingEwtSetup.ID,
				TotalIn:         totalAmount,
				TransactionType: "TRADING_BUY",
				DocNo:           docNo,
				Remark:          docNo,
				CreatedBy:       strconv.Itoa(arrData.EntMemberID),
				// Remark:          "#*buy_trading_request*#" + " " + docNo,
			}

			_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

			if err != nil {
				base.LogErrorLog("ProcessMemberBuyTradingRequestv2-SaveMemberWallet_failed", err.Error(), ewtIn, true)
				return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
			}
			// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
		}
		// }

		// return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		// start check balance
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDTo,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}
		// end check balance

		// start deduct wallet
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			TotalOut:        totalAmount,
			TransactionType: "TRADING_BUY",
			DocNo:           docNo,
			Remark:          docNo,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv2_failed_to_SaveMemberWallet", err.Error(), ewtOut, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end deduct wallet

		tradBuyStatus = "P"
	}

	arrCrtTradBuy := models.TradingBuy{
		CryptoCode:   tradingSetup[0].CodeTo,
		CryptoCodeTo: arrData.CryptoCode,
		DocNo:        docNo,
		MemberID:     arrData.EntMemberID,
		TotalUnit:    arrData.Quantity,
		UnitPrice:    arrData.UnitPrice,
		TotalAmount:  totalAmount,
		BalanceUnit:  arrData.Quantity,
		Status:       tradBuyStatus,
		SigningKey:   arrData.SigningKey,
		TransHash:    hashValue,
		CreatedBy:    strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:   base.GetCurrentDateTimeT(),
		ApprovedBy:   strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingBuy(tx, arrCrtTradBuy)
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv2-failed_to_save_trading_buy", err.Error(), arrCrtTradBuy, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docType, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv2_failed_in_UpdateRunningDocNo", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", ","), nil
}

// type CancelMemberTradingBuyRequestStruct
type CancelMemberTradingBuyRequestStruct struct {
	BuyID       int
	Quantity    float64
	EntMemberID int
}

// func ProcessMemberCancelTradingBuyRequestv1
func ProcessMemberCancelTradingBuyRequestv1(tx *gorm.DB, arrData CancelMemberTradingBuyRequestStruct) error {

	dtNow := base.GetCurrentTime("2006-01-02 15:04:05")
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: "P"},
		models.WhereCondFn{Condition: " trading_buy.id = ? ", CondValue: arrData.BuyID},
		models.WhereCondFn{Condition: " trading_buy.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_buy.balance_unit > ? ", CondValue: 0},
	)
	arrTrading, _ := models.GetTradingBuyFn(arrCond, false)

	if len(arrTrading) != 1 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrTrading[0].CryptoCode}, // usds
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
		base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-get_trading_cancel_doc_no_failed", err.Error(), nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	balanceUnit := arrTrading[0].BalanceUnit - arrData.Quantity
	// totalAmount := arrData.Quantity * arrTrading[0].UnitPrice
	totalAmount := float.Mul(arrData.Quantity, arrTrading[0].UnitPrice)

	var hashValue string
	var tradeCancelSigningKey string
	updateTradBuyColumn := map[string]interface{}{}
	tradCancelStatus := ""
	if strings.ToLower(tradingSetup.Control) == "blockchain" {

		// start get hotwallet
		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end get hotwallet

		// start check hotwallet balance
		balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(arrTrading[0].CryptoCode, hotWalletInfo.HotWalletAddress)
		if err != nil {
			base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-GetBlockchainWalletBalanceApiV1_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if balance < totalAmount {
			arrErr := map[string]interface{}{
				"cryptoType":       arrTrading[0].CryptoCode,
				"hotWalletBalance": balance,
				"quantityNeed":     arrData.Quantity,
				"buyID":            arrTrading[0].ID,
				"totalAmount":      totalAmount,
			}
			base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-hotwallet_balance_is_not_enough", err.Error(), arrErr, true)
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
				"buyID":      arrTrading[0].ID,
			}
			base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
			TokenType:       arrTrading[0].CryptoCode,
			PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
			ContractAddress: tradingSetup.ContractAddress,
			ChainID:         chainIDInt64,
			FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
			ToAddr:          cryptoAddr,                     // this is refer to the buyer address
			Amount:          totalAmount,
			MaxGas:          maxGasUint64,
		}
		signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-ProcecssGenerateSignTransaction_failed", err.Error(), arrProcecssGenerateSignTransaction, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		tradeCancelSigningKey = signingKey
		// end sign transaction for blockchain (from hotwallet to member account)

		// start send signed transaction to blockchain (from hotwallet to member account)
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:      arrTrading[0].MemberID,
			EwalletTypeID:    tradingSetup.ID,
			DocNo:            docNoC,
			Status:           "P",
			TransactionType:  "TRADING_CANCEL",
			TransactionData:  tradeCancelSigningKey,
			TotalIn:          totalAmount,
			ConversionRate:   arrTrading[0].UnitPrice,
			ConvertedTotalIn: totalAmount,
			LogOnly:          0, // take it as log only just in case error is happened in blockchain site.
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-SignedTransaction_failed", errMsg, signingKey+" sell_id:"+strconv.Itoa(arrTrading[0].ID), true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signed transaction to blockchain (from hotwallet to member account)
		// end send signing key to blockchain site
		// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}

	} else if strings.ToLower(tradingSetup.Control) == "internal" {
		// start return the directly
		arrEwtIn := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrTrading[0].MemberID,
			EwalletTypeID:   tradingSetup.ID,
			TotalIn:         totalAmount,
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

		// start add holding wallet for either holding usds [cancel partial / off prev trading buy amount] [this part move to trading callback function]
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingSetup.EwtTypeCode + "H"}, // usdsh
		)
		tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

		if tradingSetupHolding != nil {
			ewtOut := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     arrData.EntMemberID,
				EwalletTypeID:   tradingSetupHolding.ID,
				TotalOut:        totalAmount,
				TransactionType: "TRADING_CANCEL",
				DocNo:           docNoC,
				Remark:          docNoC,
				CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			}

			_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

			if err != nil {
				base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-failed_to_SaveMemberWallet_for_holding_wallet", err.Error(), ewtOut, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		}
		// start end add holding wallet for either holding usds [cancel partial / off prev trading buy amount] [this part move to trading callback function]

		tradCancelStatus = "AP"
	}

	arrCrtTradCancel := models.AddTradingCancelStruct{
		TradingID:       arrTrading[0].ID,
		MemberID:        arrData.EntMemberID,
		DocNo:           docNoC,
		TransactionType: "BUY",
		CryptoCode:      arrTrading[0].CryptoCode,
		TotalUnit:       arrData.Quantity,
		UnitPrice:       arrTrading[0].UnitPrice,
		TotalAmount:     totalAmount,
		SigningKey:      tradeCancelSigningKey,
		TransHash:       hashValue,
		Status:          tradCancelStatus,
		CreatedBy:       strconv.Itoa(arrData.EntMemberID),
	}

	// ApprovedAt: base.GetCurrentDateTimeT(),
	// ApprovedBy: strconv.Itoa(arrData.EntMemberID),

	_, err = models.AddTradingCancel(tx, arrCrtTradCancel)
	if err != nil {
		base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-failed_to_save_trading_cancel", err.Error(), arrCrtTradCancel, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docTypeC, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-failed_in_UpdateRunningDocNo", err.Error(), docTypeC, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrTrading[0].ID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "P"},
	)

	updateTradBuyColumn["balance_unit"] = balanceUnit
	updateTradBuyColumn["updated_by"] = arrData.EntMemberID

	err = models.UpdatesFnTx(tx, "trading_buy", arrUpdCond, updateTradBuyColumn, false)
	if err != nil {
		base.LogErrorLog("ProcessMemberCancelTradingBuyRequestv1-update_trading_sell_failed", "update_balance_unit_in_trading_sell", err.Error(), true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}

// func ProcessMemberBuyTradingRequestv3 - this will cover diff unit_price - with signing key
func ProcessMemberBuyTradingRequestv3(tx *gorm.DB, arrData BuyMemberTradingRequestStruct) (string, error) {

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
	docType := "TRADB"
	docNo, err := models.GetRunningDocNo(docType, tx) //get transfer doc no
	if err != nil {
		base.LogErrorLog("PProcessMemberBuyTradingRequestv3-_GetRunningDocNo_TRADB_failed", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

	totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", "")
	totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
	if err != nil {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}
	totalAmount = totalAmountFloat

	var hashValue string
	var tradBuyStatus string

	suggestedPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	if strings.ToLower(tradingSetup[0].ControlTo) == "blockchain" {
		if arrData.SigningKey == "" {
			arrErr := map[string]interface{}{
				"trading_setup": tradingSetup,
				"arrData":       arrData,
			}
			base.LogErrorLog("ProcessMemberBuyTradingRequestv3-signing_key_is_missing", arrErr, arrCond, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrData.EntMemberID, tradingSetup[0].CodeTo, true, false)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  tradingSetup[0].CodeTo,
			}
			base.LogErrorLog("ProcessMemberBuyTradingRequestv3-GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(tradingSetup[0].CodeTo, cryptoAddr, arrData.EntMemberID)
		bal := BlkCWalBal.Balance

		if totalAmount > bal { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_buy"}
		}

		// start send signing key to blockchain site
		// hashValue, errMsg := wallet_service.SignedTransaction(arrData.SigningKey)

		// if errMsg != "" {
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		// }
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			DocNo:           docNo,
			Status:          "P",
			TransactionType: "TRADING_BUY",
			TransactionData: arrData.SigningKey,
			TotalOut:        totalAmount,
			ConversionRate:  arrData.UnitPrice,
			Remark:          docNo,
			LogOnly:         0,
			// ConvertedTotalOut: totalAmount,
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site
		// return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		// start check balance
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDTo,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}
		// end check balance

		// start deduct wallet
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			TotalOut:        totalAmount,
			TransactionType: "TRADING_BUY",
			DocNo:           docNo,
			Remark:          docNo,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv3-failed_to_SaveMemberWallet2", err.Error(), ewtOut, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end deduct wallet

		tradBuyStatus = "P"
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingSetup[0].CodeTo + "H")},
	)
	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if holdingEwtSetup != nil {
		// start add holding wallet for holding wallet (bcz this trading transaction is not match yet)
		ewtIn := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   holdingEwtSetup.ID,
			TotalIn:         totalAmount,
			TransactionType: "TRADING_BUY",
			DocNo:           docNo,
			Remark:          docNo,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv3-SaveMemberWallet1_failed", err.Error(), ewtIn, true)
			return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
		}
		// end add holding wallet for holding wallet (bcz this trading transaction is not match yet)
	}

	arrCrtTradBuy := models.TradingBuy{
		CryptoCode:         tradingSetup[0].CodeTo,
		CryptoCodeTo:       arrData.CryptoCode,
		DocNo:              docNo,
		MemberID:           arrData.EntMemberID,
		TotalUnit:          arrData.Quantity,
		SuggestedUnitPrice: suggestedPrice,
		UnitPrice:          arrData.UnitPrice,
		TotalAmount:        totalAmount,
		BalanceUnit:        arrData.Quantity,
		Status:             tradBuyStatus,
		SigningKey:         arrData.SigningKey,
		TransHash:          hashValue,
		CreatedBy:          strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:         base.GetCurrentDateTimeT(),
		ApprovedBy:         strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingBuy(tx, arrCrtTradBuy)
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv3-failed_to_save_trading_buy", err.Error(), arrCrtTradBuy, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docType, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv3-failed_in_UpdateRunningDocNo", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", ","), nil
}

// func ProcessAutoTradingBuyRequestv1 - this will cover diff unit_price
func ProcessAutoTradingBuyRequestv1(tx *gorm.DB, arrData BuyMemberTradingRequestStruct) (string, error) {

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
	docType := "TRADB"
	docNo, err := models.GetRunningDocNo(docType, tx) //get transfer doc no
	if err != nil {
		base.LogErrorLog("ProcessAutoTradingBuyRequestv1-GetRunningDocNo_TRADB_failed", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

	totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", "")
	totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
	if err != nil {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}
	totalAmount = totalAmountFloat

	var hashValue string
	var tradBuyStatus string

	suggestedPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	if strings.ToLower(tradingSetup[0].ControlTo) == "blockchain" {

		tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

		chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
		chainIDInt64 := int64(chainID)
		maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
		maxGasUint64 := uint64(maxGas)

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.EntMemberID},
			models.WhereCondFn{Condition: " crypto_type = ? ", CondValue: tradingSetup[0].CodeTo},
		)

		cryptoAddr, err := models.GetEntMemberCryptoFn(arrCond, false)
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingBuyRequestv1-GetEntMemberCryptoFn_failed", err, arrCond, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(tradingSetup[0].CodeTo, cryptoAddr.CryptoAddress, arrData.EntMemberID)
		bal := BlkCWalBal.Balance

		if totalAmount > bal { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_buy"}
		}

		nonceApi, err := wallet_service.GetTransactionNonceViaAPI(cryptoAddr.CryptoAddress) // // this is refer to the hotwallet addr
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingBuyRequestv1-GetTransactionNonceViaAPI_failed", err.Error(), cryptoAddr.CryptoAddress, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "ProcessAutoTradingBuyRequestv1-GetTransactionNonceViaAPI_failed"}
		}
		// start get hotwallet
		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingSellRequestv1-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end get hotwallet

		arrGenerateSignTransaction := wallet_service.GenerateSignTransactionStruct{
			TokenType:       tradingSetup[0].CodeTo,
			PrivateKey:      cryptoAddr.PrivateKey,
			ContractAddress: tradingSetup[0].ContractAddrTo,
			ChainID:         chainIDInt64,
			Nonce:           uint64(nonceApi),
			ToAddr:          hotWalletInfo.HotWalletAddress,
			Amount:          totalAmount, // this is refer to amount for this transaction
			MaxGas:          maxGasUint64,
		}
		// base.LogErrorLog("ProcecssGenerateSignTransaction_net2", arrGenerateSignTransaction, nonce, true)
		signingKey, err := wallet_service.GenerateSignTransaction(arrGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessAutoTradingBuyRequestv1-GenerateSignTransaction_failed", err.Error(), arrGenerateSignTransaction, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "ProcessAutoTradingBuyRequestv1-GetTransactionNonceViaAPI_failed"}
		}
		// end sign transaction for blockchain (from hotwallet to member account)
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			DocNo:           docNo,
			Status:          "P",
			TransactionType: "TRADING_BUY",
			TransactionData: signingKey,
			TotalOut:        totalAmount,
			ConversionRate:  arrData.UnitPrice,
			Remark:          docNo,
			LogOnly:         0,
			// ConvertedTotalOut: totalAmount,
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site

		arrData.SigningKey = signingKey
		// return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		// start check balance
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDTo,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}
		// end check balance

		// start deduct wallet
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			TotalOut:        totalAmount,
			TransactionType: "TRADING_BUY",
			DocNo:           docNo,
			Remark:          docNo,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessAutoTradingBuyRequestv1-failed_to_SaveMemberWallet", err.Error(), ewtOut, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end deduct wallet

		tradBuyStatus = "P"
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingSetup[0].CodeTo + "H")},
	)
	holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if holdingEwtSetup != nil {
		// start add holding wallet for USDS holding wallet (bcz this trading transaction is not match yet)
		ewtIn := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   holdingEwtSetup.ID,
			TotalIn:         totalAmount,
			TransactionType: "TRADING_BUY",
			DocNo:           docNo,
			Remark:          docNo,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

		if err != nil {
			base.LogErrorLog("ProcessAutoTradingBuyRequestv1-SaveMemberWallet_failed", err.Error(), ewtIn, true)
			return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
		}
		// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
	}

	arrCrtTradBuy := models.TradingBuy{
		CryptoCode:         tradingSetup[0].CodeTo,
		CryptoCodeTo:       arrData.CryptoCode,
		DocNo:              docNo,
		MemberID:           arrData.EntMemberID,
		TotalUnit:          arrData.Quantity,
		SuggestedUnitPrice: suggestedPrice,
		UnitPrice:          arrData.UnitPrice,
		TotalAmount:        totalAmount,
		BalanceUnit:        arrData.Quantity,
		Status:             tradBuyStatus,
		SigningKey:         arrData.SigningKey,
		TransHash:          hashValue,
		CreatedBy:          strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:         base.GetCurrentDateTimeT(),
		ApprovedBy:         strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingBuy(tx, arrCrtTradBuy)
	if err != nil {
		base.LogErrorLog("ProcessAutoTradingBuyRequestv1-failed_to_save_trading_buy", err.Error(), arrCrtTradBuy, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docType, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessAutoTradingBuyRequestv1-failed_in_UpdateRunningDocNo", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", ","), nil
}

// MemberCancelTradingRequestForm struct
type ProcessCancelAutoTradingRequestForm struct {
	DocNo       string
	EntMemberID int
}

// func ProcessCancelAutoTradingBuyRequestv1
func ProcessCancelAutoTradingBuyRequestv1(tx *gorm.DB, arrData ProcessCancelAutoTradingRequestForm) error {

	dtNow := base.GetCurrentTime("2006-01-02 15:04:05")
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: "P"},
		models.WhereCondFn{Condition: " trading_buy.doc_no = ? ", CondValue: arrData.DocNo},
		models.WhereCondFn{Condition: " trading_buy.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_buy.balance_unit > ? ", CondValue: 0},
	)
	arrTrading, _ := models.GetTradingBuyFn(arrCond, false)

	if len(arrTrading) != 1 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrTrading[0].CryptoCode}, // usds
	)
	tradingSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingSetup == nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	quantity := arrTrading[0].BalanceUnit
	// start checking for cancel quantity
	// if arrTrading[0].BalanceUnit < arrData.Quantity {
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_cancel"}
	// }
	// end checking for cancel quantity

	docTypeC := "TRADC"
	docNoC, err := models.GetRunningDocNo(docTypeC, tx) //get doc no

	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-get_trading_cancel_doc_no_failed", err.Error(), nil, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	balanceUnit := arrTrading[0].BalanceUnit - quantity
	// totalAmount := quantity * arrTrading[0].UnitPrice
	totalAmount := float.Mul(quantity, arrTrading[0].UnitPrice)

	var hashValue string
	var tradeCancelSigningKey string
	updateTradBuyColumn := map[string]interface{}{}
	tradCancelStatus := ""
	if strings.ToLower(tradingSetup.Control) == "blockchain" {

		// start get hotwallet
		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end get hotwallet

		// start check hotwallet balance
		balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(arrTrading[0].CryptoCode, hotWalletInfo.HotWalletAddress)
		if err != nil {
			base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-GetBlockchainWalletBalanceApiV1_failed", err.Error(), hotWalletInfo, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if balance < totalAmount {
			arrErr := map[string]interface{}{
				"cryptoType":       arrTrading[0].CryptoCode,
				"hotWalletBalance": balance,
				"quantityNeed":     quantity,
				"buyID":            arrTrading[0].ID,
				"totalAmount":      totalAmount,
			}
			base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-hotwallet_balance_is_not_enough", err.Error(), arrErr, true)
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
				"buyID":      arrTrading[0].ID,
			}
			base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
			TokenType:       arrTrading[0].CryptoCode,
			PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
			ContractAddress: tradingSetup.ContractAddress,
			ChainID:         chainIDInt64,
			FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
			ToAddr:          cryptoAddr,                     // this is refer to the buyer address
			Amount:          totalAmount,
			MaxGas:          maxGasUint64,
		}
		signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-ProcecssGenerateSignTransaction_failed", err.Error(), arrProcecssGenerateSignTransaction, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		tradeCancelSigningKey = signingKey
		// end sign transaction for blockchain (from hotwallet to member account)

		// start send signed transaction to blockchain (from hotwallet to member account)
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:      arrTrading[0].MemberID,
			EwalletTypeID:    tradingSetup.ID,
			DocNo:            docNoC,
			Status:           "P",
			TransactionType:  "TRADING_CANCEL",
			TransactionData:  tradeCancelSigningKey,
			TotalIn:          totalAmount,
			ConversionRate:   arrTrading[0].UnitPrice,
			ConvertedTotalIn: totalAmount,
			LogOnly:          0, // take it as log only just in case error is happened in blockchain site.
		}
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-SignedTransaction_failed", errMsg, signingKey+" sell_id:"+strconv.Itoa(arrTrading[0].ID), true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signed transaction to blockchain (from hotwallet to member account)

		// start add holding wallet for either holding usds [cancel partial / off prev trading buy amount] [this part move to trading callback function]
		// arrCond = make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		// 	models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingSetup.EwtTypeCode + "H"}, // usdsh
		// )
		// tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

		// if tradingSetupHolding != nil {
		// 	ewtOut := wallet_service.SaveMemberWalletStruct{
		// 		EntMemberID:     arrData.EntMemberID,
		// 		EwalletTypeID:   tradingSetupHolding.ID,
		// 		TotalOut:        totalAmount,
		// 		TransactionType: "TRADING",
		// 		DocNo:           docNoC,
		// 		Remark:          "#*buy_trading_cancel*# " + docNoC,
		// 		CreatedBy:       strconv.Itoa(arrData.EntMemberID),
		// 	}

		// 	_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		// 	if err != nil {
		// 		base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1s-failed_to_SaveMemberWallet_for_holding_wallet", err.Error(), ewtOut, true)
		// 		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// 	}
		// }
		// start end add holding wallet for either holding usds [cancel partial / off prev trading buy amount] [this part move to trading callback function]

		// end send signing key to blockchain site
		// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}

	} else if strings.ToLower(tradingSetup.Control) == "internal" {
		// start return the directly
		arrEwtIn := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrTrading[0].MemberID,
			EwalletTypeID:   tradingSetup.ID,
			TotalIn:         totalAmount,
			TransactionType: "TRADING_CANCEL",
			DocNo:           docNoC,
			Remark:          docNoC,
			CreatedBy:       strconv.Itoa(arrTrading[0].MemberID),
		}

		_, err = wallet_service.SaveMemberWallet(tx, arrEwtIn)

		if err != nil {
			base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-SaveMemberWallet_failed", err.Error(), arrEwtIn, true)
			return err
		}
		if balanceUnit == 0 {
			updateTradBuyColumn["status"] = "C"
		}
		updateTradBuyColumn["approved_at"] = dtNow
		updateTradBuyColumn["approved_by"] = arrData.EntMemberID
		tradCancelStatus = "AP"
	}

	arrCrtTradCancel := models.AddTradingCancelStruct{
		TradingID:       arrTrading[0].ID,
		MemberID:        arrData.EntMemberID,
		DocNo:           docNoC,
		TransactionType: "BUY",
		CryptoCode:      arrTrading[0].CryptoCode,
		TotalUnit:       quantity,
		UnitPrice:       arrTrading[0].UnitPrice,
		TotalAmount:     totalAmount,
		SigningKey:      tradeCancelSigningKey,
		TransHash:       hashValue,
		Status:          tradCancelStatus,
		CreatedBy:       strconv.Itoa(arrData.EntMemberID),
	}

	// ApprovedAt: base.GetCurrentDateTimeT(),
	// ApprovedBy: strconv.Itoa(arrData.EntMemberID),

	_, err = models.AddTradingCancel(tx, arrCrtTradCancel)
	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-failed_to_save_trading_cancel", err.Error(), arrCrtTradCancel, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docTypeC, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-failed_in_UpdateRunningDocNo", err.Error(), docTypeC, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrTrading[0].ID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "P"},
	)

	updateTradBuyColumn["balance_unit"] = balanceUnit
	updateTradBuyColumn["updated_by"] = arrData.EntMemberID

	err = models.UpdatesFnTx(tx, "trading_buy", arrUpdCond, updateTradBuyColumn, false)
	if err != nil {
		base.LogErrorLog("ProcessCancelAutoTradingBuyRequestv1-update_trading_sell_failed", "update_balance_unit_in_trading_sell", err.Error(), true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}

type BuyMemberTradingRequestv2Struct struct {
	UnitPrice   float64
	Quantity    float64
	EntMemberID int
	CryptoCode  string
}

// func ProcessMemberBuyTradingRequestv4 - this will cover diff unit_price - without signing key
func ProcessMemberBuyTradingRequestv4(tx *gorm.DB, arrData BuyMemberTradingRequestv2Struct) (string, error) {
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
	docType := "TRADB"
	docNo, err := models.GetRunningDocNo(docType, tx) //get transfer doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv4-GetRunningDocNo_TRADB_failed", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	totalAmount := float.Mul(arrData.Quantity, arrData.UnitPrice)

	totalAmountString := helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", "")
	totalAmountFloat, err := strconv.ParseFloat(totalAmountString, 64)
	if err != nil {
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}
	totalAmount = totalAmountFloat

	var hashValue string
	var tradBuyStatus string
	var signingKey string

	suggestedPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	if strings.ToLower(tradingSetup[0].ControlTo) == "blockchain" {

		memberCryptoInfo, err := models.GetCustomMemberCryptoInfov2(arrData.EntMemberID, tradingSetup[0].CodeTo, true, false)
		if err != nil {
			arrErrData := map[string]interface{}{
				"entMemberID": arrData.EntMemberID,
				"cryptoType":  tradingSetup[0].CodeTo,
			}
			base.LogErrorLog("ProcessMemberBuyTradingRequestv4-GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		cryptoAddr := memberCryptoInfo.CryptoAddr

		BlkCWalBal := wallet_service.GetBlockchainWalletBalanceByAddressV1(tradingSetup[0].CodeTo, cryptoAddr, arrData.EntMemberID)
		bal := BlkCWalBal.AvailableBalance

		if totalAmount > bal { // expect front-end will pass unit in blockchain unit. eg: liga / sec. instead of currency trade to
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enough_to_buy"}
		}

		signedKeySetup, errMsg := wallet_service.GetSigningKeySettingByModule(tradingSetup[0].CodeTo, cryptoAddr, "TRADING")
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}
		signedKeySetup["decimal_point"] = tradingSetup[0].BlockchainDecimalPointTo

		hotWalletInfo, err := models.GetHotWalletInfo()
		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv4-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}

		chainIDString := signedKeySetup["chain_id"].(string)
		chainIDInt64, _ := strconv.ParseInt(chainIDString, 10, 64)

		maxGasString := signedKeySetup["max_gas"].(string)
		maxGasint64, _ := strconv.ParseInt(maxGasString, 10, 64)
		maxGasUint64 := uint64(maxGasint64)

		// start generate sign tranx
		arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
			TokenType:       tradingSetup[0].CodeTo,
			PrivateKey:      memberCryptoInfo.PrivateKey,
			ContractAddress: tradingSetup[0].ContractAddrTo,
			ChainID:         chainIDInt64,
			FromAddr:        cryptoAddr,                     // this is refer to the member addr
			ToAddr:          hotWalletInfo.HotWalletAddress, // this is refer to the hot wallet address
			Amount:          totalAmount,
			MaxGas:          maxGasUint64,
		}

		signingKeyRst, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv4-ProcecssGenerateSignTransaction_failed", err.Error(), arrProcecssGenerateSignTransaction, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}
		signingKey = signingKeyRst
		// end generate sign tranx

		// start deduct member blockchain wallet & send signing key to blockchain site
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			DocNo:           docNo,
			Status:          "P",
			TransactionType: "TRADING_BUY",
			TransactionData: signingKey,
			TotalOut:        totalAmount,
			ConversionRate:  arrData.UnitPrice,
			Remark:          docNo,
			LogOnly:         0,
			// ConvertedTotalOut: totalAmount,
		}

		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
		}

		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end deduct member blockchain wallet & send signing key to blockchain site

		// if strings.ToLower(tradingSetup[0].CodeTo) == "usds" {
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingSetup[0].CodeTo + "H")},
		)
		holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

		if holdingEwtSetup != nil {
			// start add holding wallet for USDS holding wallet (bcz this trading transaction is not match yet)
			ewtIn := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     arrData.EntMemberID,
				EwalletTypeID:   holdingEwtSetup.ID,
				TotalIn:         totalAmount,
				TransactionType: "TRADING_BUY",
				DocNo:           docNo,
				Remark:          docNo,
				CreatedBy:       strconv.Itoa(arrData.EntMemberID),
				// Remark:          "#*buy_trading_request*#" + " " + docNo,
			}

			_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

			if err != nil {
				base.LogErrorLog("ProcessMemberBuyTradingRequestv4-SaveMemberWallet_failed", err.Error(), ewtIn, true)
				return "0", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
			}
			// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
		}
		// }

		// return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	} else {
		// start check balance
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: arrData.EntMemberID,
			EwtTypeID:   tradingSetup[0].IDTo,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		if totalAmount > walletBalance.Balance {
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct"}
		}
		// end check balance

		// start deduct wallet
		ewtOut := wallet_service.SaveMemberWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   tradingSetup[0].IDTo,
			TotalOut:        totalAmount,
			TransactionType: "TRADING_BUY",
			DocNo:           docNo,
			Remark:          docNo,
			CreatedBy:       strconv.Itoa(arrData.EntMemberID),
			// Remark:          "#*buy_trading_request*#" + " " + docNo,
		}

		_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		if err != nil {
			base.LogErrorLog("ProcessMemberBuyTradingRequestv4-failed_to_SaveMemberWallet", err.Error(), ewtOut, true)
			return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end deduct wallet

		tradBuyStatus = "P"
	}

	arrCrtTradBuy := models.TradingBuy{
		CryptoCode:         tradingSetup[0].CodeTo,
		CryptoCodeTo:       arrData.CryptoCode,
		DocNo:              docNo,
		MemberID:           arrData.EntMemberID,
		TotalUnit:          arrData.Quantity,
		SuggestedUnitPrice: suggestedPrice,
		UnitPrice:          arrData.UnitPrice,
		TotalAmount:        totalAmount,
		BalanceUnit:        arrData.Quantity,
		Status:             tradBuyStatus,
		SigningKey:         signingKey,
		TransHash:          hashValue,
		CreatedBy:          strconv.Itoa(arrData.EntMemberID),
		ApprovedAt:         base.GetCurrentDateTimeT(),
		ApprovedBy:         strconv.Itoa(arrData.EntMemberID),
	}

	_, err = models.AddTradingBuy(tx, arrCrtTradBuy)
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv4-failed_to_save_trading_buy", err.Error(), arrCrtTradBuy, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	err = models.UpdateRunningDocNo(docType, tx) //update doc no
	if err != nil {
		base.LogErrorLog("ProcessMemberBuyTradingRequestv4-failed_in_UpdateRunningDocNo", err.Error(), docType, true)
		return "0", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return helpers.CutOffDecimal(totalAmount, uint(tradingSetup[0].DecimalPointTo), ".", ","), nil
}
