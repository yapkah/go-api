package trading_service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/service/wallet_service"
)

// func UpdateTradingSellTranxCallback (blockchain call back)
func UpdateTradingSellTranxCallback(tx *gorm.DB, docNo string, status bool) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.status = '' AND doc_no = ?", CondValue: docNo},
	)
	tradingSellTranx, _ := models.GetTradingSellFn(arrCond, false)

	if len(tradingSellTranx) != 1 {
		// no such records. skip it. just let it pass
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingSellTranx[0].CryptoCode},
		models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
	)
	tradingCryptoCodeFrom, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingCryptoCodeFrom == nil {
		// no such records. skip it. just let it pass
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
	}

	if status {
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "doc_no = ?", CondValue: docNo},
		)

		updateColumn := map[string]interface{}{"status": "P"}
		err := models.UpdatesFnTx(tx, "trading_sell", arrUpdCond, updateColumn, false)

		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_sell", Data: err}
		}

		return nil
	} else {
		// return back money back to seller
		if strings.ToLower(tradingCryptoCodeFrom.Control) == "blockchain" {
			if strings.ToLower(tradingSellTranx[0].CryptoCode) == "sec" || strings.ToLower(tradingSellTranx[0].CryptoCode) == "liga" {
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
					models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingSellTranx[0].CryptoCode + "H")},
				)
				holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

				if holdingEwtSetup != nil {
					// start add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
					ewtOut := wallet_service.SaveMemberWalletStruct{
						EntMemberID:     tradingSellTranx[0].MemberID,
						EwalletTypeID:   holdingEwtSetup.ID,
						TotalOut:        tradingSellTranx[0].TotalUnit,
						TransactionType: "TRADING_SELL_REJECT",
						DocNo:           tradingSellTranx[0].DocNo,
						Remark:          tradingSellTranx[0].DocNo,
						CreatedBy:       strconv.Itoa(tradingSellTranx[0].MemberID),
						// Remark:          "#*reject_sell_trading_request*#" + " " + tradingSellTranx[0].DocNo,
					}

					_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

					if err != nil {
						return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "SaveMemberWallet_failed", Data: err}
					}
					// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
				}
			}
		} else {
			// so far no such process yet. [even will have, it will 100% pass bcz it is internal wallet]
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
		}
	}

	return nil
}

// func UpdateTradingBuyTranxCallback (blockchain call back -  [bcz so far only involve usds wallet])
func UpdateTradingBuyTranxCallback(tx *gorm.DB, docNo string, status bool) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.status = '' AND doc_no = ?", CondValue: docNo},
	)
	tradingBuyTranx, _ := models.GetTradingBuyFn(arrCond, false)

	if len(tradingBuyTranx) != 1 {
		// no such records. skip it. just let it pass
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records_1"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingBuyTranx[0].CryptoCode},
		models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
	)
	tradingCryptoCodeFrom, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingCryptoCodeFrom == nil {
		// no such records. skip it. just let it pass
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records_2"}
	}

	if status {
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "doc_no = ?", CondValue: docNo},
		)

		updateColumn := map[string]interface{}{"status": "P"}
		err := models.UpdatesFnTx(tx, "trading_buy", arrUpdCond, updateColumn, false)

		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_buy", Data: err}
		}
	} else {
		// perform return back the money back to the buyer

		// return back money back to seller
		if strings.ToLower(tradingCryptoCodeFrom.Control) == "blockchain" {
			if strings.ToLower(tradingBuyTranx[0].CryptoCode) == "usds" {
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
					models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingBuyTranx[0].CryptoCode + "H")},
				)
				holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

				if holdingEwtSetup != nil {
					// start add holding wallet for either usds holding wallet (bcz this trading transaction is not match yet)
					ewtIn := wallet_service.SaveMemberWalletStruct{
						EntMemberID:     tradingBuyTranx[0].MemberID,
						EwalletTypeID:   holdingEwtSetup.ID,
						TotalIn:         tradingBuyTranx[0].TotalAmount,
						TransactionType: "TRADING_BUY_REJECT",
						DocNo:           tradingBuyTranx[0].DocNo,
						Remark:          tradingBuyTranx[0].DocNo,
						CreatedBy:       strconv.Itoa(tradingBuyTranx[0].MemberID),
						// Remark:          "#*reject_buy_trading_request*#" + " " + tradingBuyTranx[0].DocNo,
					}

					_, err := wallet_service.SaveMemberWallet(tx, ewtIn)

					if err != nil {
						return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "SaveMemberWallet_failed", Data: err}
					}
					// end add holding wallet for either usds holding wallet (bcz this trading transaction is not match yet)
				}
			}
		} else {
			// so far no such process yet. [even will have, it will 100% pass bcz it is internal wallet]
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
		}
	}

	return nil
}

// func UpdateTradingMatchTranxCallback (blockchain call back)
func UpdateTradingMatchTranxCallback(tx *gorm.DB, docNo string, status bool) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_match.doc_no = ?", CondValue: docNo},
		models.WhereCondFn{Condition: " trading_match.status = ?", CondValue: "AP"},
	)
	tradingTranx, _ := models.GetTradingMatchFn(arrCond, false)

	if len(tradingTranx) != 1 {
		// no such records. skip it. just let it pass
		base.LogErrorLog("debug1", tradingTranx, arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.id = ?", CondValue: tradingTranx[0].TradSellID},
	)
	tradingSellTranx, _ := models.GetTradingSellFn(arrCond, false)

	if len(tradingSellTranx) != 1 {
		// no such records. skip it. just let it pass
		base.LogErrorLog("debug2", tradingSellTranx, arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingSellTranx[0].CryptoCodeTo}, // usds / usdt
		models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
	)
	tradingCryptoCodeTo, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingCryptoCodeTo == nil {
		// no such records. skip it. just let it pass
		base.LogErrorLog("debug3", tradingCryptoCodeTo, arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingSellTranx[0].CryptoCode}, // liga / sec
		models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
	)
	tradingCryptoCodeFrom, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingCryptoCodeFrom == nil {
		// no such records. skip it. just let it pass
		base.LogErrorLog("debug4", tradingCryptoCodeFrom, arrCond, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
	}

	if status {
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "doc_no = ?", CondValue: docNo},
		)

		updateColumn := map[string]interface{}{"status": "M"}
		err := models.UpdatesFnTx(tx, "trading_match", arrUpdCond, updateColumn, false)

		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_match", Data: err}
		}

		var ewtTotalIn float64
		var totalAmount float64
		// start process after trading
		if strings.ToLower(tradingCryptoCodeTo.Control) == "internal" {
			ewtTotalIn = tradingTranx[0].TotalAmount // set this as default.
			// start adjusting ewtTotalIn for wallet if decimal problem existed. eg. 169.999999998 [incorrect vers.] and 170 [correct vers.]
			if tradingSellTranx[0].BalanceUnit == 0 {
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " trading_match.sell_id = ? AND trading_match.status IN ('M', 'AP')", CondValue: tradingSellTranx[0].ID},
				)
				totalTradingMatchRst, _ := models.GetTotalTradingMatchFn(arrCond, false)
				if totalTradingMatchRst.TotalAmount > 0 {
					missingMatchedTotalAmount := tradingSellTranx[0].TotalAmount - totalTradingMatchRst.TotalAmount
					previousAmount := totalTradingMatchRst.TotalAmount - ewtTotalIn // find previous match amount
					fmt.Println("internal previousAmount:", previousAmount)
					fmt.Println("internal ewtTotalIn:", ewtTotalIn)
					//ewtTotalIn = ewtTotalIn + missingMatchedTotalAmount
					ewtTotalIn = ewtTotalIn + missingMatchedTotalAmount // previous match amount
					totalAmount = ewtTotalIn + previousAmount           // overall total amount
					// fmt.Println("internal totalAmount:", totalAmount)
				}
			}
			// end adjusting ewtTotalIn for wallet if decimal problem existed. eg. 169.999999998 [incorrect vers.] and 170 [correct vers.]

			// start seller receive money for internal wallet after trad match is confirm
			ewtIn := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     tradingSellTranx[0].MemberID,
				EwalletTypeID:   tradingCryptoCodeTo.ID,
				TotalIn:         ewtTotalIn,
				TransactionType: "TRADING_MATCH",
				DocNo:           tradingTranx[0].DocNo,
				Remark:          tradingTranx[0].DocNo,
				CreatedBy:       "AUTO",
				// Remark:          "#*trading_match*#" + " " + tradingTranx[0].DocNo,
			}

			_, err := wallet_service.SaveMemberWallet(tx, ewtIn)
			if err != nil {
				base.LogErrorLog("UpdateTradingMatchTranxCallback-SaveMemberWallet_failed_received_money_for_internal_wallet_afterTrad_match_is_confirm", err.Error(), ewtIn, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-SaveMemberWallet_failed_received_money_for_internal_wallet_afterTrad_match_is_confirm", Data: err}
			}
			// end seller receive money for internal wallet after trad match is confirm

			// start return back buyer if buyer price is greater than seller price
			if tradingTranx[0].BuyUnitPrice > tradingTranx[0].UnitPrice {
				priceDiff := float.Sub(tradingTranx[0].BuyUnitPrice, tradingTranx[0].UnitPrice)
				totalReturnAmount := float.Mul(priceDiff, tradingTranx[0].TotalUnit)
				// start receive money for internal wallet after trad match is confirm
				ewtIn := wallet_service.SaveMemberWalletStruct{
					EntMemberID:     tradingTranx[0].BuyerEntMemberID,
					EwalletTypeID:   tradingCryptoCodeTo.ID,
					TotalIn:         totalReturnAmount,
					TransactionType: "TRADING_MATCH",
					DocNo:           tradingTranx[0].DocNo,
					Remark:          "#*return_extra_paid_for_matching*# " + tradingTranx[0].DocNo,
					CreatedBy:       "AUTO",
				}

				_, err := wallet_service.SaveMemberWallet(tx, ewtIn)
				if err != nil {
					base.LogErrorLog("UpdateTradingMatchTranxCallback-SaveMemberWallet_return_extra_paid_for_matching_failed", err.Error(), ewtIn, true)
					return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-SaveMemberWallet_return_extra_paid_for_matching_failed", Data: err}
				}
				// end receive money for internal wallet after trad match is confirm
			}
			// end return back buyer if buyer price is greater than seller price
		} else {
			// so far happened in usds
			// start need to handle for decimal problem
			ewtTotalIn = tradingTranx[0].TotalAmount // set this as default.
			fmt.Println("ewtTotalIn:", ewtTotalIn)
			var missingMatchedTotalAmount float64
			// start adjusting ewtTotalIn for wallet if decimal problem existed. eg. 169.999999998 [incorrect vers.] and 170 [correct vers.]
			if tradingSellTranx[0].BalanceUnit == 0 {
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " trading_match.sell_id = ? AND trading_match.status IN ('M', 'AP')", CondValue: tradingSellTranx[0].ID},
				)
				totalTradingMatchRst, _ := models.GetTotalTradingMatchFn(arrCond, false)
				if totalTradingMatchRst.TotalAmount > 0 {
					fmt.Println("blockchain missingMatchedTotalAmount:", tradingSellTranx[0].TotalAmount, totalTradingMatchRst.TotalAmount)
					missingMatchedTotalAmount = tradingSellTranx[0].TotalAmount - totalTradingMatchRst.TotalAmount
					previousAmount := totalTradingMatchRst.TotalAmount - ewtTotalIn //find previous match amount
					fmt.Println("blockchain previousAmount:", previousAmount)
					fmt.Println("blockchain ewtTotalIn:", ewtTotalIn)
					ewtTotalIn = ewtTotalIn + missingMatchedTotalAmount
					totalAmount = ewtTotalIn + previousAmount // overall total amount
				}
			}
			// end adjusting ewtTotalIn for wallet if decimal problem existed. eg. 169.999999998 [incorrect vers.] and 170 [correct vers.]
			fmt.Println("ewtTotalIn 2:", ewtTotalIn)

			// start hotwallet transfer money to trading sell member
			arrHotWalletCreditToMemberData := ArrHotWalletCreditToMemberStruct{
				EntMemberID:  tradingSellTranx[0].MemberID,
				EwtTypeCode:  tradingCryptoCodeTo.EwtTypeCode,
				EwtTypeID:    tradingCryptoCodeTo.ID,
				ContractAddr: tradingCryptoCodeTo.ContractAddress,
				Amount:       ewtTotalIn,
				DocNo:        tradingTranx[0].DocNo,
				UnitPrice:    tradingSellTranx[0].UnitPrice,
				Remark:       tradingTranx[0].DocNo,
			}

			HotWalletCreditToMemberRst, err := HotWalletCreditToMember(tx, arrHotWalletCreditToMemberData)
			if err != nil {
				base.LogErrorLog("UpdateTradingMatchTranxCallback-HotWalletCreditToMember_to_seller_failed", err.Error(), arrHotWalletCreditToMemberData, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-HotWalletCreditToMember_to_seller_failed", Data: err}
			}
			// end hotwallet transfer money to trading sell member

			// start add holding wallet for holding usds [bcz this trading transaction is matched]
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
				models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingCryptoCodeTo.EwtTypeCode + "H"},
			)
			tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

			fmt.Println("ewtTotalIn 3:", ewtTotalIn)
			if tradingSetupHolding != nil {
				ewtIn := wallet_service.SaveMemberWalletStruct{
					EntMemberID:     tradingTranx[0].BuyerEntMemberID,
					EwalletTypeID:   tradingSetupHolding.ID,
					TotalOut:        ewtTotalIn,
					TransactionType: "TRADING_MATCH",
					DocNo:           tradingTranx[0].DocNo,
					Remark:          tradingTranx[0].DocNo,
					CreatedBy:       "AUTO",
					// AdditionalMsg:   tradMatchTranxHash,
					// Remark:          "#*trading_match*# " + tradingTranx[0].DocNo,
				}

				_, err := wallet_service.SaveMemberWallet(tx, ewtIn)

				if err != nil {
					base.LogErrorLog("UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_to_buyer_failed", err.Error(), ewtIn, true)
					return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_to_buyer_failed", Data: err}
				}
			}
			// end add holding wallet for holding usds [bcz this trading transaction is matched]

			nonce := HotWalletCreditToMemberRst.Nonce + 1
			// start return back buyer if buyer price is greater than seller price
			if tradingTranx[0].BuyUnitPrice > tradingTranx[0].UnitPrice {
				priceDiff := float.Sub(tradingTranx[0].BuyUnitPrice, tradingTranx[0].UnitPrice)
				totalReturnAmount := float.Mul(priceDiff, tradingTranx[0].TotalUnit)
				// start receive money for internal wallet after trad match is confirm
				arrHotWalletCreditToMemberData := ArrHotWalletCreditToMemberStruct{
					EntMemberID:  tradingTranx[0].BuyerEntMemberID,
					EwtTypeCode:  tradingCryptoCodeTo.EwtTypeCode,
					EwtTypeID:    tradingCryptoCodeTo.ID,
					ContractAddr: tradingCryptoCodeTo.ContractAddress,
					Amount:       totalReturnAmount,
					DocNo:        tradingTranx[0].DocNo,
					UnitPrice:    tradingTranx[0].UnitPrice,
					Nonce:        nonce,
					Remark:       "#*return_extra_paid_for_matching*# " + tradingTranx[0].DocNo,
				}

				_, err := HotWalletCreditToMember(tx, arrHotWalletCreditToMemberData)
				if err != nil {
					base.LogErrorLog("UpdateTradingMatchTranxCallback-HotWalletCreditToMember_to_buyer_failed", err.Error(), arrHotWalletCreditToMemberData, true)
					return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-HotWalletCreditToMember_to_buyer_failed", Data: err}
				}
				// end receive money for internal wallet after trad match is confirm

				// start credit out member holding wallet
				if tradingSetupHolding != nil {
					ewtOut := wallet_service.SaveMemberWalletStruct{
						EntMemberID:     tradingTranx[0].BuyerEntMemberID,
						EwalletTypeID:   tradingSetupHolding.ID,
						TotalOut:        totalReturnAmount,
						TransactionType: "TRADING_MATCH",
						DocNo:           tradingTranx[0].DocNo,
						Remark:          "#return_extra_paid_for_matching# " + tradingTranx[0].DocNo,
						CreatedBy:       "AUTO",
						AdditionalMsg:   "give back buyer extra paid holding",
						// Remark:          "#trading_match# " + tradingTranx[0].DocNo,
					}

					_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

					if err != nil {
						base.LogErrorLog("UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_to_buyer_failed2", err.Error(), ewtOut, true)
						return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_to_buyer_failed2", Data: err}
					}
				}
				// end credit out member holding wallet

			}
			// end return back buyer if buyer price is greater than seller price
		}

		if strings.ToLower(tradingCryptoCodeFrom.Control) == "blockchain" {
			// start add holding wallet for either holding sec / holding liga [bcz this trading transaction is matched]
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
				models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingTranx[0].CryptoCode + "H"},
			)
			tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

			if tradingSetupHolding != nil {
				ewtOut := wallet_service.SaveMemberWalletStruct{
					EntMemberID:     tradingSellTranx[0].MemberID,
					EwalletTypeID:   tradingSetupHolding.ID,
					TotalOut:        tradingTranx[0].TotalUnit,
					TransactionType: "TRADING_MATCH",
					DocNo:           tradingTranx[0].DocNo,
					Remark:          tradingTranx[0].DocNo,
					CreatedBy:       "AUTO",
					// Remark:          "#*trading_match*# " + tradingTranx[0].DocNo,
				}

				_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

				if err != nil {
					base.LogErrorLog("UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_failed", err.Error(), ewtOut, true)
					return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_failed", Data: err}
				}
			}
			// end add holding wallet for either holding sec / holding liga [bcz this trading transaction is matched]
		}

		// start update trading_sell.status = "M" if it is fully matched (balance unit = 0)
		arrUpdCond = make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "trading_sell.id = ?", CondValue: tradingTranx[0].TradSellID},
			models.WhereCondFn{Condition: "trading_sell.balance_unit = ?", CondValue: 0},
		)
		updateColumn = map[string]interface{}{"status": "M"}
		if totalAmount > 0 {
			updateColumn["total_amount"] = totalAmount // need to update to latest total amount bcz wallet decimal problem existed. eg. 169.999999998 [incorrect vers.] and 170 [correct vers.]
		}
		err = models.UpdatesFnTx(tx, "trading_sell", arrUpdCond, updateColumn, false)

		if err != nil {
			arrErr := map[string]interface{}{
				"arrUpdCond":   arrUpdCond,
				"updateColumn": updateColumn,
			}
			base.LogErrorLog("UpdateTradingMatchTranxCallback-update_fail_in_trading_sell_for_fully_match", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_sell_for_fully_match", Data: err}
		}
		// end update trading_sell.status = "M" if it is fully matched (balance unit = 0)
		// start update trading_buy.status = "M" if it is fully matched (balance unit = 0)
		arrUpdCond = make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "trading_buy.id = ?", CondValue: tradingTranx[0].TradBuyID},
			models.WhereCondFn{Condition: "trading_buy.balance_unit = ?", CondValue: 0},
		)

		updateColumn = map[string]interface{}{"status": "M"}
		err = models.UpdatesFnTx(tx, "trading_buy", arrUpdCond, updateColumn, false)

		if err != nil {
			arrErr := map[string]interface{}{
				"arrUpdCond":   arrUpdCond,
				"updateColumn": updateColumn,
			}
			base.LogErrorLog("UpdateTradingMatchTranxCallback-update_fail_in_trading_buy_for_fully_match", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_buy_for_fully_match", Data: err}
		}
		// end update trading_buy.status = "M" if it is fully matched (balance unit = 0)
		// start update trading_process_queue.status = "AP" if the process is fully completed. so that next trading_match can be start
		arrUpdCond = make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "trading_process_queue.process_id = ?", CondValue: tradingTranx[0].DocNo},
		)
		updateColumn = map[string]interface{}{"status": "AP"}
		err = models.UpdatesFnTx(tx, "trading_process_queue", arrUpdCond, updateColumn, false)

		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_process_queue_for_fully_match", Data: err}
		}
		// end update trading_process_queue.status = "AP" if the process is fully completed. so that next trading_match can be start

		// start fix add back trading_match.total_amount for decimal problem. eg. 169.999999998 [incorrect vers.] and 170 [correct vers.]
		if ewtTotalIn > 0 {
			arrUpdCond = make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: "doc_no = ?", CondValue: docNo},
			)

			updateColumn = map[string]interface{}{"total_amount": ewtTotalIn}
			err = models.UpdatesFnTx(tx, "trading_match", arrUpdCond, updateColumn, false)

			if err != nil {
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_match", Data: err}
			}
		}
		// end fix add back trading_match.total_amount for decimal problem. eg. 169.999999998 [incorrect vers.] and 170 [correct vers.]
	} else {
		// match is failed - need to do this one
	}

	return nil
}

// func UpdateTradingCancelTranxCallback (blockchain call back)
func UpdateTradingCancelTranxCallback(tx *gorm.DB, docNo string, status bool) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_cancel.status = '' AND doc_no = ?", CondValue: docNo},
	)
	tradingTranx, _ := models.GetTradingCancelFn(arrCond, false)

	if len(tradingTranx) != 1 {
		// no such records. skip it. just let it pass
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records-tradingTranx"}
	}

	if strings.ToLower(tradingTranx[0].TransactionType) == "sell" {
		err := UpdateTradingCancelSellTranxCallback(tx, tradingTranx[0], status)
		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
	} else if strings.ToLower(tradingTranx[0].TransactionType) == "buy" {
		err := UpdateTradingCancelBuyTranxCallback(tx, tradingTranx[0], status)
		if err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
	} else {
		// no such records. skip it. just let it pass
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records-else"}
	}
	return nil
}

// func Sub func for UpdateTradingCancelBuyTranxCallback (blockchain call back) - UpdateTradingCancelBuyTranxCallback
func UpdateTradingCancelBuyTranxCallback(tx *gorm.DB, arrData *models.TradingCancel, status bool) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.id = ?", CondValue: arrData.TradingID},
		models.WhereCondFn{Condition: " trading_buy.status = ?", CondValue: "P"},
	)
	tradingTranx, _ := models.GetTradingBuyFn(arrCond, false)

	if len(tradingTranx) != 1 {
		// no such records. skip it. just let it pass
		// base.LogErrorLog("UpdateTradingCancelBuyTranxCallback-GetTradingSellFn_failed", arrCond, tradingTranx, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records_UpdateTradingCancelBuyTranxCallback-GetTradingBuyFn_failed"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
	)
	tradingCryptoCodeFrom, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingCryptoCodeFrom == nil {
		// no such records. skip it. just let it pass
		// base.LogErrorLog("UpdateTradingCancelBuyTranxCallback-GetEwtSetupFn_failed", arrCond, tradingCryptoCodeFrom, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_crypto_code_records_UpdateTradingCancelBuyTranxCallback-GetEwtSetupFn_failed"}
	}

	if status {
		// start update status for trading_buy. // fully cancel
		if tradingTranx[0].BalanceUnit <= 0 {
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: " trading_buy.id = ?", CondValue: arrData.TradingID},
				models.WhereCondFn{Condition: " trading_buy.status = ?", CondValue: "P"},
			)

			updateColumn := map[string]interface{}{"status": "C"}
			err := models.UpdatesFnTx(tx, "trading_buy", arrUpdCond, updateColumn, false)

			if err != nil {
				arrErr := map[string]interface{}{
					"arrUpdCond":   arrUpdCond,
					"updateColumn": updateColumn,
				}
				base.LogErrorLog("UpdateTradingCancelBuyTranxCallback-UpdatesFnTx_trading_buy_failed", err.Error(), arrErr, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingCancelBuyTranxCallback-UpdatesFnTx_trading_buy_failed"}
			}
		}
		// end update status for trading_buy // fully cancel

		// start add holding wallet for blockchain wallet holding usds [cancel partial / off prev trading sell amount]
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrData.CryptoCode + "H"},
		)
		tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

		if tradingSetupHolding != nil {
			ewtIn := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     arrData.MemberID,
				EwalletTypeID:   tradingSetupHolding.ID,
				TotalOut:        arrData.TotalAmount,
				TransactionType: "TRADING_CANCEL",
				DocNo:           arrData.DocNo,
				Remark:          arrData.DocNo,
				CreatedBy:       strconv.Itoa(arrData.MemberID),
			}

			_, err := wallet_service.SaveMemberWallet(tx, ewtIn)

			if err != nil {
				base.LogErrorLog("UpdateTradingCancelBuyTranxCallback-failed_to_SaveMemberWallet_for_holding_wallet", err.Error(), ewtIn, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingCancelBuyTranxCallback-failed_to_SaveMemberWallet_for_holding_wallet"}
			}
		}
		// end add holding wallet for blockchain wallet holding usds [cancel partial / off prev trading sell amount]

		// start update status for trading_cancel
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "id = ?", CondValue: arrData.ID},
		)

		updateColumn := map[string]interface{}{"status": "AP"}
		err := models.UpdatesFnTx(tx, "trading_cancel", arrUpdCond, updateColumn, false)

		if err != nil {
			arrErr := map[string]interface{}{
				"arrUpdCond":   arrUpdCond,
				"updateColumn": updateColumn,
			}
			base.LogErrorLog("UpdateTradingCancelBuyTranxCallback-UpdatesFnTx_trading_cancel_failed", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingCancelBuyTranxCallback-UpdatesFnTx_trading_cancel_failed", Data: err}
		}
		// end update status for trading_cancel
	} else {
		// so far no action need to be taken
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_action_for_this"}
	}

	return nil
}

// func Sub func for UpdateTradingCancelTranxCallback (blockchain call back) - UpdateTradingCancelSellTranxCallback
func UpdateTradingCancelSellTranxCallback(tx *gorm.DB, arrData *models.TradingCancel, status bool) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.id = ?", CondValue: arrData.TradingID},
		models.WhereCondFn{Condition: " trading_sell.status = ?", CondValue: "P"},
	)
	tradingTranx, _ := models.GetTradingSellFn(arrCond, false)

	if len(tradingTranx) != 1 {
		// no such records. skip it. just let it pass
		// base.LogErrorLog("UpdateTradingCancelSellTranxCallback-GetTradingSellFn_failed", arrCond, tradingTranx, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records_UpdateTradingCancelSellTranxCallback-GetTradingSellFn_failed"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
	)
	tradingCryptoCodeFrom, _ := models.GetEwtSetupFn(arrCond, "", false)

	if tradingCryptoCodeFrom == nil {
		// no such records. skip it. just let it pass
		// base.LogErrorLog("UpdateTradingCancelSellTranxCallback-GetEwtSetupFn_failed", arrCond, tradingCryptoCodeFrom, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_crypto_code_records_UpdateTradingCancelSellTranxCallback-GetEwtSetupFn_failed"}
	}

	if status {
		// start update status for trading_sell. // fully cancel
		if tradingTranx[0].BalanceUnit <= 0 {
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: " trading_sell.id = ?", CondValue: arrData.TradingID},
				models.WhereCondFn{Condition: " trading_sell.status = ?", CondValue: "P"},
			)

			updateColumn := map[string]interface{}{"status": "C"}
			err := models.UpdatesFnTx(tx, "trading_sell", arrUpdCond, updateColumn, false)

			if err != nil {
				arrErr := map[string]interface{}{
					"arrUpdCond":   arrUpdCond,
					"updateColumn": updateColumn,
				}
				base.LogErrorLog("UpdateTradingCancelSellTranxCallback-UpdatesFnTx_trading_sell_failed", err.Error(), arrErr, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_sell_UpdateTradingCancelSellTranxCallback-UpdatesFnTx_trading_sell_failed"}
			}
		}
		// end update status for trading_sell // fully cancel

		// start add holding wallet for either holding sec / holding liga [cancel partial / off prev trading sell amount]
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: arrData.CryptoCode + "H"},
		)
		tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

		if tradingSetupHolding != nil {
			ewtIn := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     arrData.MemberID,
				EwalletTypeID:   tradingSetupHolding.ID,
				TotalOut:        arrData.TotalUnit,
				TransactionType: "TRADING_CANCEL",
				DocNo:           arrData.DocNo,
				Remark:          arrData.DocNo,
				CreatedBy:       strconv.Itoa(arrData.MemberID),
			}

			_, err := wallet_service.SaveMemberWallet(tx, ewtIn)

			if err != nil {
				base.LogErrorLog("UpdateTradingCancelSellTranxCallback-failed_to_SaveMemberWallet_for_holding_wallet", err.Error(), ewtIn, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingCancelSellTranxCallback-failed_to_SaveMemberWallet_for_holding_wallet"}
			}
		}
		// end add holding wallet for either holding sec / holding liga [cancel partial / off prev trading sell amount]

		// start update status for trading_cancel
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "id = ?", CondValue: arrData.ID},
		)

		updateColumn := map[string]interface{}{"status": "AP"}
		err := models.UpdatesFnTx(tx, "trading_cancel", arrUpdCond, updateColumn, false)

		if err != nil {
			arrErr := map[string]interface{}{
				"arrUpdCond":   arrUpdCond,
				"updateColumn": updateColumn,
			}
			base.LogErrorLog("UpdateTradingCancelSellTranxCallback-UpdatesFnTx_trading_cancel_failed", err.Error(), arrErr, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "update_fail_in_trading_cancel_UpdateTradingCancelSellTranxCallback-UpdatesFnTx_trading_cancel_failed", Data: err}
		}
		// end update status for trading_cancel
	} else {
		// so far no action need to be taken
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_action_for_this"}

		// return back money back to seller
		// if strings.ToLower(tradingCryptoCodeFrom.Control) == "blockchain" {
		// 	if strings.ToLower(tradingTranx[0].CryptoCode) == "sec" || strings.ToLower(tradingTranx[0].CryptoCode) == "liga" {
		// 		arrCond = make([]models.WhereCondFn, 0)
		// 		arrCond = append(arrCond,
		// 			models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		// 			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: strings.ToUpper(tradingSellTranx[0].CryptoCode + "H")},
		// 		)
		// 		holdingEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

		// 		if holdingEwtSetup != nil {
		// 			// start add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
		// 			ewtOut := wallet_service.SaveMemberWalletStruct{
		// 				EntMemberID:     tradingSellTranx[0].MemberID,
		// 				EwalletTypeID:   holdingEwtSetup.ID,
		// 				TotalOut:        tradingSellTranx[0].TotalUnit,
		// 				TransactionType: "TRADING",
		// 				DocNo:           tradingSellTranx[0].DocNo,
		// 				Remark:          "#*reject_sell_trading_request*#" + " " + tradingSellTranx[0].DocNo,
		// 				CreatedBy:       strconv.Itoa(tradingSellTranx[0].MemberID),
		// 			}

		// 			_, err := wallet_service.SaveMemberWallet(tx, ewtOut)

		// 			if err != nil {
		// 				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "SaveMemberWallet_failed", Data: err}
		// 			}
		// 			// end add holding wallet for either sec / liga holding wallet (bcz this trading transaction is not match yet)
		// 		}
		// 	}
		// } else {
		// 	// so far no such process yet. [even will have, it will 100% pass bcz it is internal wallet]
		// }
	}

	return nil
}

func UpdateTradingAfterMatchTranxCallback(tx *gorm.DB, docNo string, status bool) error {
	return nil
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "blockchain_trans.doc_no = ?", CondValue: docNo},
		models.WhereCondFn{Condition: "blockchain_trans.transaction_type = ? ", CondValue: "TRADING_AFTER_MATCH"},
		models.WhereCondFn{Condition: "blockchain_trans.log_only = ? ", CondValue: 0},
		models.WhereCondFn{Condition: "blockchain_trans.status = ?", CondValue: "P"},
	)
	blockchainTrans, _ := models.GetBlockchainTransArrayFn(arrCond, false)

	if len(blockchainTrans) > 0 {

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.id = ?", CondValue: blockchainTrans[0].EwalletTypeID},
			models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
		)
		tradingCryptoCodeTo, _ := models.GetEwtSetupFn(arrCond, "", false)
		fmt.Println("tradingCryptoCodeTo", tradingCryptoCodeTo.EwtTypeName)
		if tradingCryptoCodeTo == nil {
			// no such records. skip it. just let it pass
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_records"}
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingCryptoCodeTo.EwtTypeCode + "H"},
			models.WhereCondFn{Condition: " ewt_setup.status = ?", CondValue: "A"},
		)
		tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)
		fmt.Println("tradingSetupHolding", tradingSetupHolding.EwtTypeName)

		if tradingSetupHolding != nil {
			// start add holding wallet for holding usds [bcz this trading transaction after sales is finish]
			ewtIn := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     blockchainTrans[0].MemberID,
				EwalletTypeID:   tradingSetupHolding.ID,
				TotalIn:         blockchainTrans[0].TotalIn,
				TransactionType: "TRADING",
				DocNo:           blockchainTrans[0].DocNo,
				Remark:          "#*trading_after_match*# " + blockchainTrans[0].DocNo,
				AdditionalMsg:   blockchainTrans[0].HashValue,
				CreatedBy:       "AUTO",
			}

			_, err := wallet_service.SaveMemberWallet(tx, ewtIn)

			if err != nil {
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingAfterMatchTranxCallback-SaveMemberWallet_holding_1_failed", Data: err}
			}
			// end add holding wallet for holding usds [bcz this trading transaction after sales is finish]
		}
	}
	return nil
}

type ArrHotWalletCreditToMemberStruct struct {
	EntMemberID  int
	EwtTypeCode  string
	EwtTypeID    int
	ContractAddr string
	Amount       float64
	DocNo        string
	UnitPrice    float64
	Nonce        uint64
	Remark       string
}

type HotWalletCreditToMemberRstStruct struct {
	Nonce uint64
}

func HotWalletCreditToMember(tx *gorm.DB, arrData ArrHotWalletCreditToMemberStruct) (*HotWalletCreditToMemberRstStruct, error) {
	// start get hotwallet
	hotWalletInfo, err := models.GetHotWalletInfo()
	if err != nil {
		base.LogErrorLog("HotWalletCreditToMember-GetHotWalletInfo_failed", err.Error(), hotWalletInfo, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-GetHotWalletInfo_sql_failed", Data: err}
	}
	// end get hotwallet

	// start check hotwallet balance
	balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(arrData.EwtTypeCode, hotWalletInfo.HotWalletAddress)
	if err != nil {
		arrErr := map[string]interface{}{
			"EwtTypeCode":      arrData.EwtTypeCode,
			"HotWalletAddress": hotWalletInfo.HotWalletAddress,
		}
		base.LogErrorLog("HotWalletCreditToMember-GetBlockchainWalletBalanceApiV1_failed", err.Error(), arrErr, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-GetBlockchainWalletBalanceApiV1_return_failed"}
	}
	if balance < arrData.Amount {
		arrErr := map[string]interface{}{
			"hotWalletBalance": balance,
			"ewtTotalIn":       arrData.Amount,
		}
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		base.LogErrorLog("HotWalletCreditToMember-hotwallet_balance_is_not_enough", errMsg, arrErr, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-hotwallet_balance_is_not_enough"}
	}
	// end check hotwallet balance

	// start sign transaction for blockchain (from hotwallet to member account)
	tradingSigningKeySetting, err := wallet_service.ProcessGetTradingSigningKeySetting()

	chainID, _ := strconv.Atoi(tradingSigningKeySetting.ChainID)
	chainIDInt64 := int64(chainID)
	maxGas, _ := strconv.Atoi(tradingSigningKeySetting.MaxGas)
	maxGasUint64 := uint64(maxGas)

	cryptoAddr, err := models.GetCustomMemberCryptoAddr(arrData.EntMemberID, arrData.EwtTypeCode, true, false)
	if err != nil {
		arrErr := map[string]interface{}{
			"MemberID":   arrData.EntMemberID,
			"CryptoCode": arrData.EwtTypeCode,
		}
		base.LogErrorLog("HotWalletCreditToMember-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-GetCustomMemberCryptoAddr_failed"}
	}

	// arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
	// 	TokenType:       arrData.EwtTypeCode,
	// 	PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
	// 	ContractAddress: arrData.ContractAddr,
	// 	ChainID:         chainIDInt64,
	// 	FromAddr:        hotWalletInfo.HotWalletAddress, // this is refer to the hotwallet addr
	// 	ToAddr:          cryptoAddr,                     // this is refer to the seller address
	// 	Amount:          arrData.Amount,
	// 	MaxGas:          maxGasUint64,
	// }
	// signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
	// if err != nil {
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-ProcecssGenerateSignTransaction_failed"}
	// }
	nonce := arrData.Nonce
	if arrData.Nonce == 0 {
		nonceApi, err := wallet_service.GetTransactionNonceViaAPI(hotWalletInfo.HotWalletAddress) // // this is refer to the hotwallet addr
		if err != nil {
			base.LogErrorLog("HotWalletCreditToMember-GetTransactionNonceViaAPI_failed", err.Error(), hotWalletInfo.HotWalletAddress, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-GetTransactionNonceViaAPI_failed"}
		}
		nonce = uint64(nonceApi)
	}

	arrGenerateSignTransaction := wallet_service.GenerateSignTransactionStruct{
		TokenType:       arrData.EwtTypeCode,
		PrivateKey:      hotWalletInfo.HotWalletPrivateKey,
		ContractAddress: arrData.ContractAddr,
		ChainID:         chainIDInt64,
		Nonce:           uint64(nonce),
		ToAddr:          cryptoAddr,
		Amount:          arrData.Amount, // this is refer to amount for this transaction
		MaxGas:          maxGasUint64,
	}
	// base.LogErrorLog("ProcecssGenerateSignTransaction_net2", arrGenerateSignTransaction, nonce, true)
	signingKey, err := wallet_service.GenerateSignTransaction(arrGenerateSignTransaction)
	if err != nil {
		base.LogErrorLog("HotWalletCreditToMember-GenerateSignTransaction_failed", err.Error(), arrGenerateSignTransaction, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-GetTransactionNonceViaAPI_failed"}
	}
	// end sign transaction for blockchain (from hotwallet to member account)

	// start send signed transaction to blockchain (from hotwallet to member account)
	arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
		EntMemberID:     arrData.EntMemberID,
		EwalletTypeID:   arrData.EwtTypeID,
		DocNo:           arrData.DocNo,
		Status:          "P",
		TransactionType: "TRADING_AFTER_MATCH",
		TransactionData: signingKey,
		TotalIn:         arrData.Amount,
		ConversionRate:  arrData.UnitPrice,
		LogOnly:         0,
		Remark:          arrData.Remark,
		// ConvertedTotalOut: ewtTotalIn,
	}
	// errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
	errMsg, _ := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
	if errMsg != "" {
		// models.Rollback(tx)
		base.LogErrorLog("HotWalletCreditToMember-SaveMemberBlochchainWallet_failed", errMsg, arrSaveMemberBlockchainWallet, true)
		// base.LogErrorLog("HotWalletCreditToMember-net", signingKey, arrSaveMemberBlockchainWallet, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "HotWalletCreditToMember-SignedTransaction_failed", Data: err}
	}

	// tradMatchTranxHash := saveMemberBlochchainWalletRst["hashValue"]
	// start send signed transaction to blockchain (from hotwallet to member account)
	// end receive money for blockchain wallet after trad match is confirm

	// start add holding wallet for holding usds [bcz this trading transaction is matched]
	// arrCond := make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
	// 	models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ?", CondValue: tradingCryptoCodeTo.EwtTypeCode + "H"},
	// )
	// tradingSetupHolding, _ := models.GetEwtSetupFn(arrCond, "", false)

	// if tradingSetupHolding != nil {
	// 	ewtIn := wallet_service.SaveMemberWalletStruct{
	// 		EntMemberID:     tradingSellTranx[0].MemberID,
	// 		EwalletTypeID:   tradingSetupHolding.ID,
	// 		TotalOut:        ewtTotalIn,
	// 		TransactionType: "TRADING",
	// 		DocNo:           tradingTranx[0].DocNo,
	// 		Remark:          "#*trading_after_match*# " + tradingTranx[0].DocNo,
	// 		AdditionalMsg:   tradMatchTranxHash,
	// 		CreatedBy:       "AUTO",
	// 	}

	// 	_, err := wallet_service.SaveMemberWallet(tx, ewtIn)

	// 	if err != nil {
	// 		base.LogErrorLog("UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_failed", err.Error(), ewtIn, true)
	// 		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-SaveMemberWallet_holding_1_failed", Data: err}
	// 	}
	// }
	// end add holding wallet for holding usds [bcz this trading transaction is matched]
	// end need to handle for decimal problem

	arrDataReturn := HotWalletCreditToMemberRstStruct{
		Nonce: uint64(nonce),
	}
	return &arrDataReturn, nil
}

// ewtIn := wallet_service.SaveMemberWalletStruct{
// 	EntMemberID:     tradingSellTranx[0].MemberID,
// 	EwalletTypeID:   tradingCryptoCodeTo.ID,
// 	TotalIn:         ewtTotalIn,
// 	TransactionType: "TRADING",
// 	DocNo:           tradingTranx[0].DocNo,
// 	Remark:          "#*trading_match*#" + " " + tradingTranx[0].DocNo,
// 	CreatedBy:       "AUTO",
// }
// _, err := wallet_service.SaveMemberWallet(tx, ewtIn)
// if err != nil {
// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "UpdateTradingMatchTranxCallback-SaveMemberWallet_failed", Data: err}
// }
