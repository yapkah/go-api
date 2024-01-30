package trading_service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/service/wallet_service"
)

func TestDBTranx() {

	arrList1 := []int{1, 2, 3, 4}
	arrList2 := []int{11, 22, 33, 44}

	for _, arrList1V := range arrList1 {
		for _, arrList2V := range arrList2 {

			if arrList2V == 33 {
				continue
			} else {

				fmt.Println("arrList1V:", arrList1V)
				fmt.Println("arrList2V:", arrList2V)
			}
		}
	}

	// err = models.Commit(tx)
	// if err != nil {
	// 	models.Rollback(tx)
	// 	base.LogErrorLog("ProcessAutoMatchTrading-Commit_failed", err.Error(), nil, true)
	// 	// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-commit_failed"}
	// 	continue
	// }
}

// func ProcessAutoMatchTrading
func ProcessAutoMatchTrading(manual bool) {
	settingID := "process_auto_match_trading_setting"
	arrSettingRst, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil || arrSettingRst == nil {
		fmt.Println("no process_auto_match_trading_setting setting")
		return
	}

	if arrSettingRst.InputType1 != "1" && !manual {
		fmt.Println("process_auto_match_trading_setting is off")
		return
	}

	tradingProcessQueuePendingStatus := GetTradingProcessQueuePendingStatus()
	if tradingProcessQueuePendingStatus {
		fmt.Println("TradingProcessQueuePendingStatus:", tradingProcessQueuePendingStatus)
		return
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.status = ? ", CondValue: "P"},
		models.WhereCondFn{Condition: " trading_sell.balance_unit > ? ", CondValue: 0},
	)
	arrAutoMatchTradingSellListV2 := models.AutoMatchTradingSellListV2{
		OrderBy: "unit_price ASC",
	}
	tradingSellList, _ := models.GetAutoMatchTradingSellListFnV2(arrCond, arrAutoMatchTradingSellListV2, false)

	tradSellRecordStatus := false
	tradBuyRecordStatus := false

	if len(tradingSellList) > 0 {
		tradSellRecordStatus = true
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: "P"},
		models.WhereCondFn{Condition: " trading_buy.balance_unit > ? ", CondValue: 0},
	)
	arrAutoMatchTradingBuyListV2 := models.AutoMatchTradingBuyListV2{
		OrderBy: "unit_price ASC",
	}
	tradingBuyList, _ := models.GetAutoMatchTradingBuyListFnV2(arrCond, arrAutoMatchTradingBuyListV2, false)

	if len(tradingBuyList) > 0 {
		tradBuyRecordStatus = true
	}

	// run auto match if both table got records
	if tradSellRecordStatus && tradBuyRecordStatus {
		tradActionPririoty := "sell" // set default is sell
		sellDtString := base.TimeFormat(tradingSellList[0].CreatedAt, "2006-01-02 15:04:05")
		buyDtString := base.TimeFormat(tradingBuyList[0].CreatedAt, "2006-01-02 15:04:05")

		if sellDtString < buyDtString {
			tradActionPririoty = "sell"
		} else if sellDtString > buyDtString {
			tradActionPririoty = "buy"
		}

		if tradActionPririoty == "sell" {
			// ProcessMatchingSellDetails(tradingSellList)
			ProcessMatchingSellDetailsV2(tradingSellList)
		} else { // buy
			// ProcessMatchingBuyDetails()
			ProcessMatchingBuyDetailsV2(tradingBuyList)
		}
		fmt.Println("auto-match-success")
	}

	return
}

// func ProcessMatchingSellDetails - will auto match only if buy unit price = sell unit price
func ProcessMatchingSellDetails(tradingSellList []*models.AutoMatchTrading) {
	for _, tradingSellListV := range tradingSellList {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.unit_price = ? ", CondValue: tradingSellListV.UnitPrice},      // same price
			models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: "P"},                                 // trading is still in progress
			models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: tradingSellListV.CryptoCode}, // same crypto code
			models.WhereCondFn{Condition: " trading_buy.member_id <> ? ", CondValue: tradingSellListV.MemberID},       // no self match
		)
		tradingBuyList, _ := models.GetAutoMatchTradingBuyListFn(arrCond, 0, false)

		if len(tradingBuyList) > 0 {
			for _, tradingBuyListV := range tradingBuyList {
				tx := models.Begin()

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCode},
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				sellCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if sellCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-GetEwtSetupFn_tradingSellListV_failed", "setting_tradingSellListV_missing", arrCond, true)
					return
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCodeTo}, // example is liga
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				buyCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if buyCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-GetEwtSetupFn_tradingBuyListV_failed", "setting_tradingBuyListV_missing", arrCond, true)
					return
				}

				docTypeM := "TRADM"
				docNoM, err := models.GetRunningDocNo(docTypeM, tx) //get doc no
				dtNowT := base.GetCurrentDateTimeT()

				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-GetRunningDocNo_TRADM_failed", err.Error(), docTypeM, true)
					return
				}
				var tradMatchTranxHash string
				var tradMatchSigningKey string
				var tradMatchTotalUnit float64
				var tradMatchTotalAmount float64
				var nextTradingSell bool
				var addTradingProcessQueue bool
				var stopProcessNextTradingMatching bool

				if tradingSellListV.BalanceUnit == tradingBuyListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched fully.
					// fmt.Println("1")
					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * tradingSellListV.UnitPrice
					err = AutoMatchingSellEqualBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)

					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
				} else if tradingSellListV.BalanceUnit > tradingBuyListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched partially.
					// fmt.Println("2")
					tradMatchTotalUnit = tradingBuyListV.BalanceUnit
					tradMatchTotalAmount = tradingBuyListV.BalanceUnit * tradingBuyListV.UnitPrice
					err = AutoMatchingSellGreatBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradingSellListV.BalanceUnit = tradingSellListV.BalanceUnit - tradingBuyListV.BalanceUnit
					nextTradingSell = false
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = false
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
				} else if tradingSellListV.BalanceUnit < tradingBuyListV.BalanceUnit {
					// trading sell is matched fully.
					// trading buy is matched partially.
					// fmt.Println("3")
					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * tradingSellListV.UnitPrice
					err = AutoMatchingSellLessBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
				}

				tradMatchTotalAmountString := helpers.CutOffDecimal(tradMatchTotalAmount, uint(sellCryptoCode.DecimalPoint), ".", "")
				tradMatchTotalAmountFloat, err := strconv.ParseFloat(tradMatchTotalAmountString, 64)
				if err != nil {
					models.Rollback(tx)
					arrErr := map[string]interface{}{
						"tradMatchTotalAmount":       tradMatchTotalAmount,
						"tradMatchTotalAmountString": tradMatchTotalAmountString,
						"tradMatchTotalAmountFloat":  tradMatchTotalAmountFloat,
					}
					base.LogErrorLog("ProcessAutoMatchTrading-helpers_CutOffDecimal_tradMatchTotalAmount_failed", err.Error(), arrErr, true)
					return
				}
				tradMatchTotalAmount = tradMatchTotalAmountFloat

				if strings.ToLower(buyCryptoCode.Control) == "blockchain" {
					// start get hotwallet
					hotWalletInfo, err := models.GetHotWalletInfo()
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
						return
					}
					// end get hotwallet

					// start check hotwallet balance
					balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(tradingBuyListV.CryptoCodeTo, hotWalletInfo.HotWalletAddress)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}

					if balance < tradMatchTotalUnit {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"hotWalletBalance": balance,
							"cryptoType":       tradingBuyListV.CryptoCodeTo,
							"quantityNeed":     tradMatchTotalUnit,
							"sellID":           tradingSellListV.ID,
							"buyID":            tradingBuyListV.ID,
						}
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						}

						base.LogErrorLog("ProcessAutoMatchTrading-hotwallet_balance_is_not_enough", errMsg, arrErr, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}
					// end check hotwallet balance

					// start sign transaction for blockchain (from hotwallet to member account)
					tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

					chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
					chainIDInt64 := int64(chainID)
					maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
					maxGasUint64 := uint64(maxGas)

					cryptoAddr, err := models.GetCustomMemberCryptoAddr(tradingBuyListV.MemberID, tradingBuyListV.CryptoCodeTo, true, false)
					if err != nil {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"MemberID":   tradingBuyListV.MemberID,
							"CryptoCode": tradingBuyListV.CryptoCodeTo,
						}
						base.LogErrorLog("ProcessAutoMatchTrading-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
						return
					}

					arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
						TokenType:       tradingBuyListV.CryptoCodeTo,
						PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
						ContractAddress: buyCryptoCode.ContractAddress,
						ChainID:         chainIDInt64,
						FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
						ToAddr:          cryptoAddr,                     // this is refer to the buyer address
						Amount:          tradMatchTotalUnit,
						MaxGas:          maxGasUint64,
					}
					signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradMatchSigningKey = signingKey
					// end sign transaction for blockchain (from hotwallet to member account)

					// start send signed transaction to blockchain (from hotwallet to member account)
					arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
						EntMemberID:      tradingBuyListV.MemberID,
						EwalletTypeID:    buyCryptoCode.ID,
						DocNo:            docNoM,
						Status:           "P",
						TransactionType:  "TRADING_MATCH",
						TransactionData:  tradMatchSigningKey,
						TotalIn:          tradMatchTotalUnit,
						ConversionRate:   tradingBuyListV.UnitPrice,
						ConvertedTotalIn: tradMatchTotalAmount,
						LogOnly:          0,
						Remark:           docNoM,
					}
					errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
					if errMsg != "" {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-SignedTransaction_failed", errMsg, signingKey, true)
						base.LogErrorLog("ProcessAutoMatchTrading-net", signingKey, arrSaveMemberBlockchainWallet, true)
						return
					}

					tradMatchTranxHash = saveMemberBlochchainWalletRst["hashValue"]
					// end send signed transaction to blockchain (from hotwallet to member account)
				} else {
					// so far no this yet
				}

				// start create trading match record
				arrCrtTradMatch := models.TradingMatch{
					DocNo:          docNoM,
					CryptoCode:     tradingSellListV.CryptoCode,
					SellID:         tradingSellListV.ID,
					BuyID:          tradingBuyListV.ID,
					SellerMemberID: tradingSellListV.MemberID,
					BuyerMemberID:  tradingBuyListV.MemberID,
					TotalUnit:      tradMatchTotalUnit,
					UnitPrice:      tradingSellListV.UnitPrice,
					TotalAmount:    tradMatchTotalAmount,
					SigningKey:     tradMatchSigningKey,
					TransHash:      tradMatchTranxHash,
					Status:         "AP",
					CreatedBy:      "AUTO",
					ApprovedAt:     dtNowT,
					ApprovedBy:     "AUTO",
				}

				_, err = models.AddTradingMatch(tx, arrCrtTradMatch)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-failed_to_save_trading_match", err.Error(), arrCrtTradMatch, true)
					return
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
				}
				// end create trading match record

				err = models.UpdateRunningDocNo(docTypeM, tx) //update doc no
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-failed_in_UpdateRunningDocNo_docTypeM", err.Error(), docTypeM, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
					return
				}

				// start add record to stop the next auto matching start
				if addTradingProcessQueue {
					arrCrtTradProcessQueue := models.AddTradingProcessQueueStruct{
						ProcessID: docNoM,
						Status:    "P",
					}
					_, err = models.AddTradingProcessQueue(tx, arrCrtTradProcessQueue)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-failed_to_save_trading_process_queue", err.Error(), arrCrtTradProcessQueue, true)
						return
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
					}
				}
				// end add record to stop the next auto matching start

				// start commit if everything is success
				err = models.Commit(tx)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-Commit_failed", err.Error(), nil, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-commit_failed"}
					return
				}

				// start prevent the code flow being continue flow
				if stopProcessNextTradingMatching {
					return
				}
				// end prevent the code flow being continue flow

				if nextTradingSell {
					break
				}

				// run finish for matching, this is needed bcz matching need to b 1 by 1.
				return
			}
		}
	}
}

// func ProcessMatchingBuyDetails - will auto match only if buy unit price = sell unit price
func ProcessMatchingBuyDetails() {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: "P"}, // trading is still in progress
		models.WhereCondFn{Condition: " trading_buy.balance_unit > ? ", CondValue: 0},
	)
	tradingBuyList, _ := models.GetAutoMatchTradingBuyListFn(arrCond, 0, false)

	for _, tradingBuyListV := range tradingBuyList {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.unit_price = ? ", CondValue: tradingBuyListV.UnitPrice},      // same price
			models.WhereCondFn{Condition: " trading_sell.status = ? ", CondValue: "P"},                                // trading is still in progress
			models.WhereCondFn{Condition: " trading_sell.crypto_code_to = ? ", CondValue: tradingBuyListV.CryptoCode}, // same crypto code
			models.WhereCondFn{Condition: " trading_sell.member_id <> ? ", CondValue: tradingBuyListV.MemberID},       // no self match
		)
		tradingSellList, _ := models.GetAutoMatchTradingSellListFn(arrCond, false)

		if len(tradingSellList) > 0 {
			for _, tradingSellListV := range tradingSellList {
				tx := models.Begin()

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCode},
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				sellCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if sellCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-GetEwtSetupFn_tradingSellListV_failed", "setting_tradingSellListV_missing", arrCond, true)
					return
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCodeTo},
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				buyCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if buyCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-GetEwtSetupFn_tradingBuyListV_failed", "setting_tradingBuyListV_missing", arrCond, true)
					return
				}

				docTypeM := "TRADM"
				docNoM, err := models.GetRunningDocNo(docTypeM, tx) //get doc no
				dtNowT := base.GetCurrentDateTimeT()

				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-GetRunningDocNo_TRADM_failed", err.Error(), docTypeM, true)
					return
				}
				var tradMatchTranxHash string
				var tradMatchSigningKey string
				var tradMatchTotalUnit float64
				var tradMatchTotalAmount float64
				var nextTradingSell bool
				var addTradingProcessQueue bool
				var stopProcessNextTradingMatching bool
				// fmt.Println("sell balance unit:", tradingSellListV.BalanceUnit)
				// fmt.Println("buy balance unit:", tradingBuyListV.BalanceUnit)
				if tradingSellListV.BalanceUnit == tradingBuyListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched fully.
					// fmt.Println("1")
					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * tradingSellListV.UnitPrice
					err = AutoMatchingSellEqualBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)

					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
				} else if tradingBuyListV.BalanceUnit > tradingSellListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched partially.
					// fmt.Println("2")
					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * tradingSellListV.UnitPrice
					err = AutoMatchingSellLessBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradingBuyListV.BalanceUnit = tradingBuyListV.BalanceUnit - tradingSellListV.BalanceUnit
					nextTradingSell = false
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = false
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
				} else if tradingBuyListV.BalanceUnit < tradingSellListV.BalanceUnit {
					// trading sell is matched fully.
					// trading buy is matched partially.
					// fmt.Println("3")
					tradMatchTotalUnit = tradingBuyListV.BalanceUnit
					tradMatchTotalAmount = tradingBuyListV.BalanceUnit * tradingBuyListV.UnitPrice
					err = AutoMatchingSellGreatBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
				}

				tradMatchTotalAmountString := helpers.CutOffDecimal(tradMatchTotalAmount, uint(sellCryptoCode.DecimalPoint), ".", "")
				tradMatchTotalAmountFloat, err := strconv.ParseFloat(tradMatchTotalAmountString, 64)
				if err != nil {
					models.Rollback(tx)
					arrErr := map[string]interface{}{
						"tradMatchTotalAmount":       tradMatchTotalAmount,
						"tradMatchTotalAmountString": tradMatchTotalAmountString,
						"tradMatchTotalAmountFloat":  tradMatchTotalAmountFloat,
					}
					base.LogErrorLog("ProcessAutoMatchTrading-helpers_CutOffDecimal_tradMatchTotalAmount_failed", err.Error(), arrErr, true)
					return
				}

				tradMatchTotalAmount = tradMatchTotalAmountFloat

				if strings.ToLower(buyCryptoCode.Control) == "blockchain" {
					// start get hotwallet
					hotWalletInfo, err := models.GetHotWalletInfo()
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
						return
					}
					// end get hotwallet

					// start check hotwallet balance
					balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(tradingBuyListV.CryptoCodeTo, hotWalletInfo.HotWalletAddress)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-GetBlockchainWalletBalanceApiV1_failed", err.Error(), hotWalletInfo, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}

					if balance < tradMatchTotalUnit {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"hotWalletBalance": balance,
							"cryptoType":       tradingBuyListV.CryptoCodeTo,
							"quantityNeed":     tradMatchTotalUnit,
							"sellID":           tradingSellListV.ID,
							"buyID":            tradingBuyListV.ID,
						}
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						}

						base.LogErrorLog("ProcessAutoMatchTrading-hotwallet_balance_is_not_enough", errMsg, arrErr, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}
					// end check hotwallet balance

					// start sign transaction for blockchain (from hotwallet to member account)
					tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

					chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
					chainIDInt64 := int64(chainID)
					maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
					maxGasUint64 := uint64(maxGas)

					cryptoAddr, err := models.GetCustomMemberCryptoAddr(tradingBuyListV.MemberID, tradingBuyListV.CryptoCodeTo, true, false)
					if err != nil {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"MemberID":   tradingBuyListV.MemberID,
							"CryptoCode": tradingBuyListV.CryptoCodeTo,
						}
						base.LogErrorLog("ProcessAutoMatchTrading-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
						return
					}

					arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
						TokenType:       tradingBuyListV.CryptoCodeTo,
						PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
						ContractAddress: buyCryptoCode.ContractAddress,
						ChainID:         chainIDInt64,
						FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
						ToAddr:          cryptoAddr,                     // this is refer to the buyer address
						Amount:          tradMatchTotalUnit,
						MaxGas:          maxGasUint64,
					}
					signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradMatchSigningKey = signingKey
					// end sign transaction for blockchain (from hotwallet to member account)

					// start send signed transaction to blockchain (from hotwallet to member account)
					arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
						EntMemberID:       tradingBuyListV.MemberID,
						EwalletTypeID:     buyCryptoCode.ID,
						DocNo:             docNoM,
						Status:            "P",
						TransactionType:   "TRADING_MATCH",
						TransactionData:   tradMatchSigningKey,
						TotalIn:           tradMatchTotalUnit,
						ConversionRate:    tradingBuyListV.UnitPrice,
						ConvertedTotalOut: tradMatchTotalAmount,
						LogOnly:           0,
						Remark:            docNoM,
					}
					errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
					if errMsg != "" {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-SignedTransaction_failed", errMsg, signingKey, true)
						base.LogErrorLog("ProcessAutoMatchTrading-net", signingKey, arrSaveMemberBlockchainWallet, true)
						return
					}

					tradMatchTranxHash = saveMemberBlochchainWalletRst["hashValue"]
					// end send signed transaction to blockchain (from hotwallet to member account)
				} else {
					// so far no this yet
				}

				// start create trading match record
				arrCrtTradMatch := models.TradingMatch{
					DocNo:          docNoM,
					CryptoCode:     tradingSellListV.CryptoCode,
					SellID:         tradingSellListV.ID,
					BuyID:          tradingBuyListV.ID,
					SellerMemberID: tradingSellListV.MemberID,
					BuyerMemberID:  tradingBuyListV.MemberID,
					TotalUnit:      tradMatchTotalUnit,
					UnitPrice:      tradingSellListV.UnitPrice,
					TotalAmount:    tradMatchTotalAmount,
					SigningKey:     tradMatchSigningKey,
					TransHash:      tradMatchTranxHash,
					Status:         "AP",
					CreatedBy:      "AUTO",
					ApprovedAt:     dtNowT,
					ApprovedBy:     "AUTO",
				}

				_, err = models.AddTradingMatch(tx, arrCrtTradMatch)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-failed_to_save_trading_match", err.Error(), arrCrtTradMatch, true)
					return
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
				}
				// end create trading match record

				err = models.UpdateRunningDocNo(docTypeM, tx) //update doc no
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-failed_in_UpdateRunningDocNo_docTypeM", err.Error(), docTypeM, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
					return
				}

				// start add record to stop the next auto matching start
				if addTradingProcessQueue {
					arrCrtTradProcessQueue := models.AddTradingProcessQueueStruct{
						ProcessID: docNoM,
						Status:    "P",
					}
					_, err = models.AddTradingProcessQueue(tx, arrCrtTradProcessQueue)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessAutoMatchTrading-failed_to_save_trading_process_queue", err.Error(), arrCrtTradProcessQueue, true)
						return
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
					}
				}
				// end add record to stop the next auto matching start

				// start commit if everything is success
				err = models.Commit(tx)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessAutoMatchTrading-Commit_failed", err.Error(), nil, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-commit_failed"}
					return
				}

				// start prevent the code flow being continue flow
				if stopProcessNextTradingMatching {
					return
				}
				// end prevent the code flow being continue flow

				if nextTradingSell {
					break
				}

				// run finish for matching, this is needed bcz matching need to b 1 by 1.
				return
			}
		}
	}
}

type AutoMatchingSellEqualBuyRst struct {
	TradMatchTotalUnit   float64
	TradMatchTotalAmount float64
}

func AutoMatchingSellEqualBuy(tx *gorm.DB, tradingSellListV *models.AutoMatchTrading, tradingBuyListV *models.AutoMatchTrading, sellCryptoCode *models.EwtSetup, buyCryptoCode *models.EwtSetup) error {

	tradBuyStatus := "AP"
	tradSellStatus := "AP"
	tradBuyBalUnit := 0
	tradSellBalUnit := 0

	// start update trading sell
	updateTradSellColumn := map[string]interface{}{}
	updateTradSellColumn["status"] = tradSellStatus
	updateTradSellColumn["balance_unit"] = tradSellBalUnit

	arrUpdTradSellCond := make([]models.WhereCondFn, 0)
	arrUpdTradSellCond = append(arrUpdTradSellCond,
		models.WhereCondFn{Condition: " trading_sell.id = ? ", CondValue: tradingSellListV.ID},
	)

	err := models.UpdatesFnTx(tx, "trading_sell", arrUpdTradSellCond, updateTradSellColumn, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"trading_sell_upd_cond": arrUpdTradSellCond,
			"trading_sell_upd_data": updateTradSellColumn,
		}
		base.LogErrorLog("ProcessAutoMatchTrading-UpdatesFn_trading_sell_failed", err.Error(), arrErr, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-UpdatesFn_trading_buy_failed"}
	}
	// end update trading sell

	// start update trading buy
	updateTradBuyColumn := map[string]interface{}{}
	updateTradBuyColumn["status"] = tradBuyStatus
	updateTradBuyColumn["balance_unit"] = tradBuyBalUnit

	arrUpdTradBuyCond := make([]models.WhereCondFn, 0)
	arrUpdTradBuyCond = append(arrUpdTradBuyCond,
		models.WhereCondFn{Condition: " trading_buy.id = ? ", CondValue: tradingBuyListV.ID},
	)

	err = models.UpdatesFnTx(tx, "trading_buy", arrUpdTradBuyCond, updateTradBuyColumn, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"trading_buy_upd_cond": arrUpdTradBuyCond,
			"trading_buy_upd_data": updateTradBuyColumn,
		}
		base.LogErrorLog("ProcessAutoMatchTrading-UpdatesFn_trading_buy_failed", err.Error(), arrErr, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-UpdatesFn_trading_buy_failed"}
	}
	// end update trading buy
	return nil
}

func AutoMatchingSellGreatBuy(tx *gorm.DB, tradingSellListV *models.AutoMatchTrading, tradingBuyListV *models.AutoMatchTrading, sellCryptoCode *models.EwtSetup, buyCryptoCode *models.EwtSetup) error {

	tradBuyStatus := "AP"
	tradSellStatus := "P"
	tradBuyBalUnit := 0
	tradSellBalUnit := tradingSellListV.BalanceUnit - tradingBuyListV.BalanceUnit

	// start update trading sell
	updateTradSellColumn := map[string]interface{}{}
	updateTradSellColumn["status"] = tradSellStatus
	updateTradSellColumn["balance_unit"] = tradSellBalUnit

	arrUpdTradSellCond := make([]models.WhereCondFn, 0)
	arrUpdTradSellCond = append(arrUpdTradSellCond,
		models.WhereCondFn{Condition: " trading_sell.id = ? ", CondValue: tradingSellListV.ID},
	)

	err := models.UpdatesFnTx(tx, "trading_sell", arrUpdTradSellCond, updateTradSellColumn, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"trading_sell_upd_cond": arrUpdTradSellCond,
			"trading_sell_upd_data": updateTradSellColumn,
		}
		base.LogErrorLog("ProcessAutoMatchTrading-UpdatesFn_trading_sell_failed", err.Error(), arrErr, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-UpdatesFn_trading_buy_failed"}
	}
	// end update trading sell

	// start update trading buy
	updateTradBuyColumn := map[string]interface{}{}
	updateTradBuyColumn["status"] = tradBuyStatus
	updateTradBuyColumn["balance_unit"] = tradBuyBalUnit

	arrUpdTradBuyCond := make([]models.WhereCondFn, 0)
	arrUpdTradBuyCond = append(arrUpdTradBuyCond,
		models.WhereCondFn{Condition: " trading_buy.id = ? ", CondValue: tradingBuyListV.ID},
	)

	err = models.UpdatesFnTx(tx, "trading_buy", arrUpdTradBuyCond, updateTradBuyColumn, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"trading_buy_upd_cond": arrUpdTradBuyCond,
			"trading_buy_upd_data": updateTradBuyColumn,
		}
		base.LogErrorLog("ProcessAutoMatchTrading-UpdatesFn_trading_buy_failed", err.Error(), arrErr, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-UpdatesFn_trading_buy_failed"}
	}
	// end update trading buy
	return nil
}

func AutoMatchingSellLessBuy(tx *gorm.DB, tradingSellListV *models.AutoMatchTrading, tradingBuyListV *models.AutoMatchTrading, sellCryptoCode *models.EwtSetup, buyCryptoCode *models.EwtSetup) error {

	tradSellStatus := "AP"
	tradBuyStatus := "P"
	tradBuyBalUnit := tradingBuyListV.BalanceUnit - tradingSellListV.BalanceUnit
	tradSellBalUnit := 0

	// start update trading sell
	updateTradSellColumn := map[string]interface{}{}
	updateTradSellColumn["status"] = tradSellStatus
	updateTradSellColumn["balance_unit"] = tradSellBalUnit

	arrUpdTradSellCond := make([]models.WhereCondFn, 0)
	arrUpdTradSellCond = append(arrUpdTradSellCond,
		models.WhereCondFn{Condition: " trading_sell.id = ? ", CondValue: tradingSellListV.ID},
	)

	err := models.UpdatesFnTx(tx, "trading_sell", arrUpdTradSellCond, updateTradSellColumn, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"trading_sell_upd_cond": arrUpdTradSellCond,
			"trading_sell_upd_data": updateTradSellColumn,
		}
		base.LogErrorLog("ProcessAutoMatchTrading-UpdatesFn_trading_sell_failed", err.Error(), arrErr, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-UpdatesFn_trading_buy_failed"}
	}
	// end update trading sell

	// start update trading buy
	updateTradBuyColumn := map[string]interface{}{}
	updateTradBuyColumn["status"] = tradBuyStatus
	updateTradBuyColumn["balance_unit"] = tradBuyBalUnit

	arrUpdTradBuyCond := make([]models.WhereCondFn, 0)
	arrUpdTradBuyCond = append(arrUpdTradBuyCond,
		models.WhereCondFn{Condition: " trading_buy.id = ? ", CondValue: tradingBuyListV.ID},
	)

	err = models.UpdatesFnTx(tx, "trading_buy", arrUpdTradBuyCond, updateTradBuyColumn, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"trading_buy_upd_cond": arrUpdTradBuyCond,
			"trading_buy_upd_data": updateTradBuyColumn,
		}
		base.LogErrorLog("ProcessAutoMatchTrading-UpdatesFn_trading_buy_failed", err.Error(), arrErr, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-UpdatesFn_trading_buy_failed"}
	}
	// end update trading buy
	return nil
}

func GetTradingProcessQueuePendingStatus() bool {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " status = ? ", CondValue: "P"},
	)
	tradingProcessQueue, _ := models.GetTradingProcessQueueFn(arrCond, false)

	if len(tradingProcessQueue) > 0 {
		// prev last trading match is not successfully matched
		return true
	}
	return false
}

// func ProcessMatchingSellDetailsV2 - will auto match only if buy unit price >= sell unit price
func ProcessMatchingSellDetailsV2(tradingSellList []*models.AutoMatchTrading) {
	for _, tradingSellListV := range tradingSellList {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.unit_price >= ? ", CondValue: tradingSellListV.UnitPrice},     // same price
			models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: "P"},                                 // trading is still in progress
			models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: tradingSellListV.CryptoCode}, // same crypto code
			models.WhereCondFn{Condition: " trading_buy.crypto_code = ? ", CondValue: tradingSellListV.CryptoCodeTo},  // same crypto code
			models.WhereCondFn{Condition: " trading_buy.member_id <> ? ", CondValue: tradingSellListV.MemberID},       // no self match
		)
		arrAutoMatchTradingBuyListV2 := models.AutoMatchTradingBuyListV2{
			Limit:   1,
			OrderBy: " unit_price ASC",
		}
		tradingBuyList, _ := models.GetAutoMatchTradingBuyListFnV2(arrCond, arrAutoMatchTradingBuyListV2, false)

		if len(tradingBuyList) > 0 {
			for _, tradingBuyListV := range tradingBuyList {
				tx := models.Begin()

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCode}, // eg: usds
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				sellCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if sellCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingSellDetailsV2-GetEwtSetupFn_tradingSellListV_failed", "setting_tradingSellListV_missing", arrCond, true)
					return
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCodeTo}, // eg: liga
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				buyCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if buyCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingSellDetailsV2-GetEwtSetupFn_tradingBuyListV_failed", "setting_tradingBuyListV_missing", arrCond, true)
					return
				}

				docTypeM := "TRADM"
				docNoM, err := models.GetRunningDocNo(docTypeM, tx) //get doc no
				dtNowT := base.GetCurrentDateTimeT()

				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingSellDetailsV2-GetRunningDocNo_TRADM_failed", err.Error(), docTypeM, true)
					return
				}
				var tradMatchTranxHash string
				var tradMatchSigningKey string
				var tradMatchTotalUnit float64
				var tradMatchTotalAmount float64
				var nextTradingSell bool
				var addTradingProcessQueue bool
				var stopProcessNextTradingMatching bool

				// base.LogErrorLog("ProcessMatchingSellDetailsV2", nil, nil, true)
				// base.LogErrorLog(tradingSellListV.CreatedAt, tradingBuyListV.CreatedAt, tradingSellListV.CreatedAt.Before(tradingBuyListV.CreatedAt), true)
				exchangePrice := tradingSellListV.UnitPrice // this exchange price need to be sell price. - trading buy is asking price
				exchangePriceCreatedBy := tradingSellListV.MemberID
				if tradingBuyListV.CreatedAt.Before(tradingSellListV.CreatedAt) {
					exchangePrice = tradingBuyListV.UnitPrice
					exchangePriceCreatedBy = tradingBuyListV.MemberID
					// base.LogErrorLog(tradingBuyListV, tradingSellListV, exchangePrice, true)
				}
				// base.LogErrorLog(tradingBuyListV, tradingSellListV, exchangePrice, true)

				if tradingSellListV.BalanceUnit == tradingBuyListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched fully.
					// fmt.Println("ProcessMatchingSellDetailsV2-1")
					base.LogErrorLog("ProcessMatchingSellDetailsV2-1", tradingSellListV, tradingBuyListV, true)
					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * tradingSellListV.UnitPrice
					err = AutoMatchingSellEqualBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)

					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
					// fmt.Println("tradMatchTotalAmount:", tradMatchTotalAmount)
				} else if tradingSellListV.BalanceUnit > tradingBuyListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched partially.
					// fmt.Println("ProcessMatchingSellDetailsV2-2")
					// base.LogErrorLog("ProcessMatchingSellDetailsV2-2", tradingSellListV, tradingBuyListV, true)
					unitPrice := tradingBuyListV.UnitPrice
					if tradingBuyListV.UnitPrice > tradingSellListV.UnitPrice {
						unitPrice = tradingSellListV.UnitPrice
					}

					tradMatchTotalUnit = tradingBuyListV.BalanceUnit
					tradMatchTotalAmount = tradingBuyListV.BalanceUnit * unitPrice
					err = AutoMatchingSellGreatBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradingSellListV.BalanceUnit = tradingSellListV.BalanceUnit - tradingBuyListV.BalanceUnit
					nextTradingSell = false
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = false
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
					// fmt.Println("tradMatchTotalAmount:", tradMatchTotalAmount)
				} else if tradingSellListV.BalanceUnit < tradingBuyListV.BalanceUnit {
					// trading sell is matched fully.
					// trading buy is matched partially.
					// base.LogErrorLog("ProcessMatchingSellDetailsV2-3", tradingSellListV, tradingBuyListV, true)
					unitPrice := tradingSellListV.UnitPrice
					if tradingBuyListV.UnitPrice > tradingSellListV.UnitPrice {
						unitPrice = tradingSellListV.UnitPrice
					}

					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * unitPrice
					err = AutoMatchingSellLessBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
					// fmt.Println("tradMatchTotalAmount:", tradMatchTotalAmount)
				}

				tradMatchTotalAmountString := helpers.CutOffDecimal(tradMatchTotalAmount, uint(sellCryptoCode.DecimalPoint), ".", "")
				tradMatchTotalAmountFloat, err := strconv.ParseFloat(tradMatchTotalAmountString, 64)
				if err != nil {
					models.Rollback(tx)
					arrErr := map[string]interface{}{
						"tradMatchTotalAmount":       tradMatchTotalAmount,
						"tradMatchTotalAmountString": tradMatchTotalAmountString,
						"tradMatchTotalAmountFloat":  tradMatchTotalAmountFloat,
					}
					base.LogErrorLog("ProcessMatchingSellDetailsV2-helpers_CutOffDecimal_tradMatchTotalAmount_failed", err.Error(), arrErr, true)
					return
				}
				tradMatchTotalAmount = tradMatchTotalAmountFloat

				if strings.ToLower(buyCryptoCode.Control) == "blockchain" {
					// start get hotwallet
					hotWalletInfo, err := models.GetHotWalletInfo()
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
						return
					}
					// end get hotwallet

					// start check hotwallet balance
					balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(tradingBuyListV.CryptoCodeTo, hotWalletInfo.HotWalletAddress)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}

					if balance < tradMatchTotalUnit {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"hotWalletBalance": balance,
							"cryptoType":       tradingBuyListV.CryptoCodeTo,
							"quantityNeed":     tradMatchTotalUnit,
							"sellID":           tradingSellListV.ID,
							"buyID":            tradingBuyListV.ID,
						}
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						}

						base.LogErrorLog("ProcessMatchingSellDetailsV2-hotwallet_balance_is_not_enough", errMsg, arrErr, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}
					// end check hotwallet balance

					// start sign transaction for blockchain (from hotwallet to member account)
					tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

					chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
					chainIDInt64 := int64(chainID)
					maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
					maxGasUint64 := uint64(maxGas)

					cryptoAddr, err := models.GetCustomMemberCryptoAddr(tradingBuyListV.MemberID, tradingBuyListV.CryptoCodeTo, true, false)
					if err != nil {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"MemberID":   tradingBuyListV.MemberID,
							"CryptoCode": tradingBuyListV.CryptoCodeTo,
						}
						base.LogErrorLog("ProcessMatchingSellDetailsV2-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
						return
					}

					arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
						TokenType:       tradingBuyListV.CryptoCodeTo,
						PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
						ContractAddress: buyCryptoCode.ContractAddress,
						ChainID:         chainIDInt64,
						FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
						ToAddr:          cryptoAddr,                     // this is refer to the buyer address
						Amount:          tradMatchTotalUnit,
						MaxGas:          maxGasUint64,
					}
					signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradMatchSigningKey = signingKey
					// end sign transaction for blockchain (from hotwallet to member account)

					// start send signed transaction to blockchain (from hotwallet to member account)
					arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
						EntMemberID:      tradingBuyListV.MemberID,
						EwalletTypeID:    buyCryptoCode.ID,
						DocNo:            docNoM,
						Status:           "P",
						TransactionType:  "TRADING_MATCH",
						TransactionData:  tradMatchSigningKey,
						TotalIn:          tradMatchTotalUnit,
						ConversionRate:   tradingBuyListV.UnitPrice,
						ConvertedTotalIn: tradMatchTotalAmount,
						LogOnly:          0,
						Remark:           docNoM,
					}
					errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
					if errMsg != "" {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"signingKey":                    signingKey,
							"arrSaveMemberBlockchainWallet": arrSaveMemberBlockchainWallet,
						}
						base.LogErrorLog("ProcessMatchingSellDetailsV2-SignedTransaction_failed", errMsg, arrErr, true)
						return
					}

					tradMatchTranxHash = saveMemberBlochchainWalletRst["hashValue"]
					// end send signed transaction to blockchain (from hotwallet to member account)
				} else {
					// start credit to member internal wallet
					// start add holding wallet for holding wallet (cancel off the matching amount)
					ewtIn := wallet_service.SaveMemberWalletStruct{
						EntMemberID:     tradingBuyListV.MemberID,
						EwalletTypeID:   buyCryptoCode.ID,
						TotalIn:         tradMatchTotalUnit,
						TransactionType: "TRADING_MATCH",
						DocNo:           docNoM,
						Remark:          docNoM,
						CreatedBy:       strconv.Itoa(tradingBuyListV.MemberID),
						// Remark:          "#*sell_trading_request*#" + " " + docNoS,
					}

					_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-SaveMemberWallet_wallet_failed", err.Error(), ewtIn, true)
						return
						// return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
					}
					// end add holding wallet for holding wallet (cancel off the matching amount)
					// end credit to member internal wallet
				}

				// start credit in holding wallet if it is exist in ewt_setup
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
					models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(sellCryptoCode.EwtTypeCode + "H")},
				)
				holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

				if holdingEwtSetup != nil {
					// start add holding wallet for holding wallet (cancel off the matching amount)
					ewtIn := wallet_service.SaveMemberWalletStruct{
						EntMemberID:     tradingBuyListV.MemberID,
						EwalletTypeID:   holdingEwtSetup.ID,
						TotalOut:        tradMatchTotalAmount,
						TransactionType: "TRADING_MATCH",
						DocNo:           docNoM,
						Remark:          docNoM,
						CreatedBy:       strconv.Itoa(tradingBuyListV.MemberID),
						// Remark:          "#*sell_trading_request*#" + " " + docNoS,
					}

					_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-SaveMemberWallet_holding_wallet_failed", err.Error(), ewtIn, true)
						return
						// return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
					}
					// end add holding wallet for holding wallet (cancel off the matching amount)
				}
				// end credit in holding wallet if it is exist in ewt_setup

				// start create trading match record
				arrCrtTradMatch := models.TradingMatch{
					DocNo:          docNoM,
					CryptoCode:     tradingSellListV.CryptoCode,
					SellID:         tradingSellListV.ID,
					BuyID:          tradingBuyListV.ID,
					SellerMemberID: tradingSellListV.MemberID,
					BuyerMemberID:  tradingBuyListV.MemberID,
					TotalUnit:      tradMatchTotalUnit,
					UnitPrice:      tradingSellListV.UnitPrice,
					ExchangePrice:  exchangePrice,
					TotalAmount:    tradMatchTotalAmount,
					SigningKey:     tradMatchSigningKey,
					TransHash:      tradMatchTranxHash,
					Status:         "AP",
					CreatedBy:      "AUTO",
					ApprovedAt:     dtNowT,
					ApprovedBy:     "AUTO",
				}

				_, err = models.AddTradingMatch(tx, arrCrtTradMatch)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingSellDetailsV2-failed_to_save_trading_match", err.Error(), arrCrtTradMatch, true)
					return
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
				}
				// end create trading match record

				err = models.UpdateRunningDocNo(docTypeM, tx) //update doc no
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingSellDetailsV2-failed_in_UpdateRunningDocNo_docTypeM", err.Error(), docTypeM, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
					return
				}

				// start save new exhange price rate - this one need to ignore commit.
				if strings.ToLower(tradingSellListV.CryptoCode) == "sec" {
					arrExchangePrice := models.AddExchangePriceMovementSecStruct{
						TokenPrice: exchangePrice,
						CreatedBy:  exchangePriceCreatedBy,
					}
					_, err := models.AddExchangePriceMovementSec(arrExchangePrice)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-failed_in_AddExchangePriceMovementSec", err.Error(), arrExchangePrice, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}

				} else if strings.ToLower(tradingSellListV.CryptoCode) == "liga" {
					arrExchangePrice := models.AddExchangePriceMovementLigaStruct{
						TokenPrice: exchangePrice,
						CreatedBy:  exchangePriceCreatedBy,
					}
					_, err := models.AddExchangePriceMovementLiga(arrExchangePrice)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-failed_in_AddExchangePriceMovementLiga", err.Error(), arrExchangePrice, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}
				}
				// end save new exhange price rate - this one need to ignore commit.

				// start add record to stop the next auto matching start
				if addTradingProcessQueue {
					arrCrtTradProcessQueue := models.AddTradingProcessQueueStruct{
						ProcessID: docNoM,
						Status:    "P",
					}
					_, err = models.AddTradingProcessQueue(tx, arrCrtTradProcessQueue)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-failed_to_save_trading_process_queue", err.Error(), arrCrtTradProcessQueue, true)
						return
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
					}
				}
				// end add record to stop the next auto matching start

				// start commit if everything is success
				err = models.Commit(tx)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingSellDetailsV2-Commit_failed", err.Error(), nil, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-commit_failed"}
					return
				}

				// start prevent the code flow being continue flow
				if stopProcessNextTradingMatching {
					return
				}
				// end prevent the code flow being continue flow

				if nextTradingSell {
					break
				}

				// run finish for matching, this is needed bcz matching need to b 1 by 1.
				return
			}
		}
	}
}

// func ProcessMatchingBuyDetailsV2 - will auto match only if buy unit price >= sell unit price
func ProcessMatchingBuyDetailsV2(tradingBuyList []*models.AutoMatchTrading) {
	for _, tradingBuyListV := range tradingBuyList {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.unit_price <= ? ", CondValue: tradingBuyListV.UnitPrice},     // same price
			models.WhereCondFn{Condition: " trading_sell.status = ? ", CondValue: "P"},                                // trading is still in progress
			models.WhereCondFn{Condition: " trading_sell.crypto_code_to = ? ", CondValue: tradingBuyListV.CryptoCode}, // same crypto code
			models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: tradingBuyListV.CryptoCodeTo},  // same crypto code
			models.WhereCondFn{Condition: " trading_sell.member_id <> ? ", CondValue: tradingBuyListV.MemberID},       // no self match
		)
		arrAutoMatchTradingSellListV2 := models.AutoMatchTradingSellListV2{
			Limit:   1,
			OrderBy: "unit_price ASC",
		}
		tradingSellList, _ := models.GetAutoMatchTradingSellListFnV2(arrCond, arrAutoMatchTradingSellListV2, false)

		if len(tradingSellList) > 0 {
			for _, tradingSellListV := range tradingSellList {
				tx := models.Begin()

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCode}, // eg: usds
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				sellCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if sellCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingBuyDetailsV2-GetEwtSetupFn_tradingBuyListV_failed", "setting_tradingBuyListV_missing", arrCond, true)
					return
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: tradingBuyListV.CryptoCodeTo}, // eg: liga
					models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
				)
				buyCryptoCode, _ := models.GetEwtSetupFn(arrCond, "", false)

				if buyCryptoCode == nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingBuyDetailsV2-GetEwtSetupFn_tradingBuyListV_failed", "setting_tradingBuyListV_missing", arrCond, true)
					return
				}

				docTypeM := "TRADM"
				docNoM, err := models.GetRunningDocNo(docTypeM, tx) //get doc no
				dtNowT := base.GetCurrentDateTimeT()

				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingBuyDetailsV2-GetRunningDocNo_TRADM_failed", err.Error(), docTypeM, true)
					return
				}
				var tradMatchTranxHash string
				var tradMatchSigningKey string
				var tradMatchTotalUnit float64
				var tradMatchTotalAmount float64
				var nextTradingSell bool
				var addTradingProcessQueue bool
				var stopProcessNextTradingMatching bool

				// base.LogErrorLog("ProcessMatchingBuyDetailsV2", nil, nil, true)
				// base.LogErrorLog(tradingSellListV.CreatedAt, tradingBuyListV.CreatedAt, tradingSellListV.CreatedAt.Before(tradingBuyListV.CreatedAt), true)
				exchangePrice := tradingBuyListV.UnitPrice // this exchange price need to be buy price. - trading sell is asking price
				exchangePriceCreatedBy := tradingBuyListV.MemberID
				if tradingSellListV.CreatedAt.Before(tradingBuyListV.CreatedAt) { // if sell is before buy
					exchangePrice = tradingSellListV.UnitPrice
					exchangePriceCreatedBy = tradingSellListV.MemberID
					// base.LogErrorLog(tradingBuyListV, tradingSellListV, exchangePrice, true)
				}
				// base.LogErrorLog(tradingBuyListV, tradingSellListV, exchangePrice, true)

				// fmt.Println("sell balance unit:", tradingSellListV.BalanceUnit)
				// fmt.Println("buy balance unit:", tradingBuyListV.BalanceUnit)
				if tradingSellListV.BalanceUnit == tradingBuyListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched fully.
					// fmt.Println("ProcessMatchingBuyDetailsV2-1")
					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * tradingSellListV.UnitPrice
					err = AutoMatchingSellEqualBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)

					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradingSellListV.BalanceUnit:", tradingSellListV.BalanceUnit)
					// fmt.Println("tradMatchTotalAmount:", tradMatchTotalAmount)
					// base.LogErrorLog("ProcessMatchingBuyDetailsV2-1", tradMatchTotalUnit, tradMatchTotalAmount, true)
				} else if tradingBuyListV.BalanceUnit > tradingSellListV.BalanceUnit {
					// trading buy is matched fully.
					// trading sell is matched partially.
					// fmt.Println("ProcessMatchingBuyDetailsV2-2")
					unitPrice := tradingSellListV.UnitPrice
					if tradingBuyListV.UnitPrice < tradingSellListV.UnitPrice {
						unitPrice = tradingBuyListV.UnitPrice
					}
					tradMatchTotalUnit = tradingSellListV.BalanceUnit
					tradMatchTotalAmount = tradingSellListV.BalanceUnit * unitPrice
					err = AutoMatchingSellLessBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradingBuyListV.BalanceUnit = tradingBuyListV.BalanceUnit - tradingSellListV.BalanceUnit
					nextTradingSell = false
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = false
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
					// fmt.Println("tradMatchTotalAmount:", tradMatchTotalAmount)
					// base.LogErrorLog("ProcessMatchingBuyDetailsV2-2", tradMatchTotalUnit, tradMatchTotalAmount, true)
				} else if tradingBuyListV.BalanceUnit < tradingSellListV.BalanceUnit {
					// trading sell is matched fully.
					// trading buy is matched partially.
					// fmt.Println("ProcessMatchingBuyDetailsV2-3")
					unitPrice := tradingBuyListV.UnitPrice
					if tradingBuyListV.UnitPrice > tradingSellListV.UnitPrice {
						unitPrice = tradingSellListV.UnitPrice
					}
					tradMatchTotalUnit = tradingBuyListV.BalanceUnit
					tradMatchTotalAmount = tradingBuyListV.BalanceUnit * unitPrice
					err = AutoMatchingSellGreatBuy(tx, tradingSellListV, tradingBuyListV, sellCryptoCode, buyCryptoCode)
					if err != nil {
						models.Rollback(tx)
						return
					}
					nextTradingSell = true
					addTradingProcessQueue = true
					stopProcessNextTradingMatching = true
					// fmt.Println("tradMatchTotalUnit:", tradMatchTotalUnit)
					// fmt.Println("tradMatchTotalAmount:", tradMatchTotalAmount)
					// base.LogErrorLog("ProcessMatchingBuyDetailsV2-3", tradMatchTotalUnit, tradMatchTotalAmount, true)
				}

				tradMatchTotalAmountString := helpers.CutOffDecimal(tradMatchTotalAmount, uint(sellCryptoCode.DecimalPoint), ".", "")
				tradMatchTotalAmountFloat, err := strconv.ParseFloat(tradMatchTotalAmountString, 64)
				if err != nil {
					models.Rollback(tx)
					arrErr := map[string]interface{}{
						"tradMatchTotalAmount":       tradMatchTotalAmount,
						"tradMatchTotalAmountString": tradMatchTotalAmountString,
						"tradMatchTotalAmountFloat":  tradMatchTotalAmountFloat,
					}
					base.LogErrorLog("ProcessMatchingBuyDetailsV2-helpers_CutOffDecimal_tradMatchTotalAmount_failed", err.Error(), arrErr, true)
					return
				}

				tradMatchTotalAmount = tradMatchTotalAmountFloat

				if strings.ToLower(buyCryptoCode.Control) == "blockchain" {
					// start get hotwallet
					hotWalletInfo, err := models.GetHotWalletInfo()
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingBuyDetailsV2-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
						return
					}
					// end get hotwallet

					// start check hotwallet balance
					balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(tradingBuyListV.CryptoCodeTo, hotWalletInfo.HotWalletAddress)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingBuyDetailsV2-GetBlockchainWalletBalanceApiV1_failed", err.Error(), hotWalletInfo, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}

					if balance < tradMatchTotalUnit {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"hotWalletBalance": balance,
							"cryptoType":       tradingBuyListV.CryptoCodeTo,
							"quantityNeed":     tradMatchTotalUnit,
							"sellID":           tradingSellListV.ID,
							"buyID":            tradingBuyListV.ID,
						}
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						}

						base.LogErrorLog("ProcessMatchingBuyDetailsV2-hotwallet_balance_is_not_enough", errMsg, arrErr, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}
					// end check hotwallet balance

					// start sign transaction for blockchain (from hotwallet to member account)
					tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

					chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
					chainIDInt64 := int64(chainID)
					maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
					maxGasUint64 := uint64(maxGas)

					cryptoAddr, err := models.GetCustomMemberCryptoAddr(tradingBuyListV.MemberID, tradingBuyListV.CryptoCodeTo, true, false)
					if err != nil {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"MemberID":   tradingBuyListV.MemberID,
							"CryptoCode": tradingBuyListV.CryptoCodeTo,
						}
						base.LogErrorLog("ProcessMatchingBuyDetailsV2-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
						return
					}

					arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
						TokenType:       tradingBuyListV.CryptoCodeTo,
						PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
						ContractAddress: buyCryptoCode.ContractAddress,
						ChainID:         chainIDInt64,
						FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
						ToAddr:          cryptoAddr,                     // this is refer to the buyer address
						Amount:          tradMatchTotalUnit,
						MaxGas:          maxGasUint64,
					}
					signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
					if err != nil {
						models.Rollback(tx)
						return
					}
					tradMatchSigningKey = signingKey
					// end sign transaction for blockchain (from hotwallet to member account)

					// start send signed transaction to blockchain (from hotwallet to member account)
					arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
						EntMemberID:       tradingBuyListV.MemberID,
						EwalletTypeID:     buyCryptoCode.ID,
						DocNo:             docNoM,
						Status:            "P",
						TransactionType:   "TRADING_MATCH",
						TransactionData:   tradMatchSigningKey,
						TotalIn:           tradMatchTotalUnit,
						ConversionRate:    tradingBuyListV.UnitPrice,
						ConvertedTotalOut: tradMatchTotalAmount,
						LogOnly:           0,
						Remark:            docNoM,
					}
					errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
					if errMsg != "" {
						models.Rollback(tx)
						arrErr := map[string]interface{}{
							"signingKey":                    signingKey,
							"arrSaveMemberBlockchainWallet": arrSaveMemberBlockchainWallet,
						}
						base.LogErrorLog("ProcessMatchingBuyDetailsV2-SignedTransaction_failed", errMsg, arrErr, true)
						return
					}

					tradMatchTranxHash = saveMemberBlochchainWalletRst["hashValue"]
					// end send signed transaction to blockchain (from hotwallet to member account)
				} else {
					// so far no this yet
				}

				// start credit in holding wallet if it is exist in ewt_setup
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
					models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(sellCryptoCode.EwtTypeCode + "H")},
				)
				holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

				if holdingEwtSetup != nil {
					// start add holding wallet for holding wallet (cancel off the matching amount)
					ewtIn := wallet_service.SaveMemberWalletStruct{
						EntMemberID:     tradingBuyListV.MemberID,
						EwalletTypeID:   holdingEwtSetup.ID,
						TotalOut:        tradMatchTotalAmount,
						TransactionType: "TRADING_MATCH",
						DocNo:           docNoM,
						Remark:          docNoM,
						CreatedBy:       strconv.Itoa(tradingBuyListV.MemberID),
						// Remark:          "#*sell_trading_request*#" + " " + docNoS,
					}

					_, err = wallet_service.SaveMemberWallet(tx, ewtIn)

					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingSellDetailsV2-SaveMemberWallet_holding_wallet_failed", err.Error(), ewtIn, true)
						return
						// return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
					}
					// end add holding wallet for holding wallet (cancel off the matching amount)
				}
				// end credit in holding wallet if it is exist in ewt_setup

				// start create trading match record
				arrCrtTradMatch := models.TradingMatch{
					DocNo:          docNoM,
					CryptoCode:     tradingSellListV.CryptoCode,
					SellID:         tradingSellListV.ID,
					BuyID:          tradingBuyListV.ID,
					SellerMemberID: tradingSellListV.MemberID,
					BuyerMemberID:  tradingBuyListV.MemberID,
					TotalUnit:      tradMatchTotalUnit,
					UnitPrice:      tradingSellListV.UnitPrice,
					ExchangePrice:  exchangePrice,
					TotalAmount:    tradMatchTotalAmount,
					SigningKey:     tradMatchSigningKey,
					TransHash:      tradMatchTranxHash,
					Status:         "AP",
					CreatedBy:      "AUTO",
					ApprovedAt:     dtNowT,
					ApprovedBy:     "AUTO",
				}

				_, err = models.AddTradingMatch(tx, arrCrtTradMatch)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingBuyDetailsV2-failed_to_save_trading_match", err.Error(), arrCrtTradMatch, true)
					return
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
				}
				// end create trading match record

				err = models.UpdateRunningDocNo(docTypeM, tx) //update doc no
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingBuyDetailsV2-failed_in_UpdateRunningDocNo_docTypeM", err.Error(), docTypeM, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
					return
				}

				// start save new exhange price rate - this one need to ignore commit.
				if strings.ToLower(tradingSellListV.CryptoCode) == "sec" {
					arrExchangePrice := models.AddExchangePriceMovementSecStruct{
						TokenPrice: exchangePrice,
						CreatedBy:  exchangePriceCreatedBy,
					}
					_, err := models.AddExchangePriceMovementSec(arrExchangePrice)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingBuyDetailsV2-AddExchangePriceMovementSec", err.Error(), arrExchangePrice, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}
				} else if strings.ToLower(tradingSellListV.CryptoCode) == "liga" {
					arrExchangePrice := models.AddExchangePriceMovementLigaStruct{
						TokenPrice: exchangePrice,
						CreatedBy:  exchangePriceCreatedBy,
					}
					_, err := models.AddExchangePriceMovementLiga(arrExchangePrice)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingBuyDetailsV2-failed_in_AddExchangePriceMovementLiga", err.Error(), arrExchangePrice, true)
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
						return
					}

				}
				// end save new exhange price rate - this one need to ignore commit.

				// start add record to stop the next auto matching start
				if addTradingProcessQueue {
					arrCrtTradProcessQueue := models.AddTradingProcessQueueStruct{
						ProcessID: docNoM,
						Status:    "P",
					}
					_, err = models.AddTradingProcessQueue(tx, arrCrtTradProcessQueue)
					if err != nil {
						models.Rollback(tx)
						base.LogErrorLog("ProcessMatchingBuyDetailsV2-failed_to_save_trading_process_queue", err.Error(), arrCrtTradProcessQueue, true)
						return
						// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match_failed_to_save_trading_match"}
					}
				}
				// end add record to stop the next auto matching start

				// start commit if everything is success
				err = models.Commit(tx)
				if err != nil {
					models.Rollback(tx)
					base.LogErrorLog("ProcessMatchingBuyDetailsV2-Commit_failed", err.Error(), nil, true)
					// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "auto_match-commit_failed"}
					return
				}

				// start prevent the code flow being continue flow
				if stopProcessNextTradingMatching {
					return
				}
				// end prevent the code flow being continue flow

				if nextTradingSell {
					break
				}

				// run finish for matching, this is needed bcz matching need to b 1 by 1.
				return
			}
		}
	}
}
