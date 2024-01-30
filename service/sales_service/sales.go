package sales_service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/float"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/crypto_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

type MemberSalesListPaymentStruct struct {
	Amount         string `json:"amount"`
	CurrencyCode   string `json:"currency_code"`
	AdditionalInfo string `json:"additional_info"`
}

type MemberSalesListStruct struct {
	StatusDesc                    string                         `json:"status_desc"`
	StatusCode                    string                         `json:"status_code"`
	StatusColorCode               string                         `json:"status_color_code"`
	PrdCode                       string                         `json:"prd_code"`
	PrdName                       string                         `json:"prd_name"`
	SalesType                     string                         `json:"sales_type"`
	DocNo                         string                         `json:"doc_no"`
	TotalUnit                     float64                        `json:"total_unit"`
	TotalAmount                   string                         `json:"total_amount"`
	Nft                           string                         `json:"nft"`
	NftAirdrop                    string                         `json:"nft_airdrop"`
	NftAirdropType                string                         `json:"nft_airdrop_type"`
	NftTier                       string                         `json:"nft_tier"`
	TokenRate                     string                         `json:"token_rate"`
	ExchangeRate                  string                         `json:"exchange_rate"`
	TotalTopupAmount              string                         `json:"total_topup_amount"`
	CurrencyCode                  string                         `json:"currency_code"`
	CreatedAt                     string                         `json:"created_at"`
	Payment                       []MemberSalesListPaymentStruct `json:"payment"`
	RebatePerc                    string                         `json:"rebate_perc"`
	TopupHistoryStatus            int                            `json:"topup_history_status"`
	CryptoPair                    string                         `json:"crypto_pair"`
	AutoTradingApiIcon            string                         `json:"auto_trading_api_icon"`
	AutoTradingSettingType        string                         `json:"auto_trading_setting_type"`
	AutoTradingColorCode          string                         `json:"auto_trading_color_code"`
	AutoTradingOperatingColorCode string                         `json:"auto_trading_operating_color_code"`
	AutoTradingSetting            []map[string]interface{}       `json:"auto_trading_setting"`
	TopupSetting                  map[string]interface{}         `json:"topup_setting"`
	RefundSetting                 map[string]interface{}         `json:"refund_setting"`
	PdfContract                   map[string]interface{}         `json:"pdf_contract"`
	ExpiredAt                     string                         `json:"expired_at"`
}

type MemberSalesStruct struct {
	MemberID int
	NickName string
	LangCode string
	Page     int64
	DateFrom string
	DateTo   string
	DocType  string
	PrdCode  string
}

// func GetMemberSalesListv1
func GetMemberSalesListv1(arrData MemberSalesStruct) (*app.ArrDataResponseList, string) {
	// arrNewMemberSalesList := make([]MemberSalesListStruct, 0)
	arrDataReturn := app.ArrDataResponseList{
		CurrentPageItems: make([]MemberSalesListStruct, 0),
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: arrData.MemberID},
		// models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "A"},
	)

	if arrData.DateFrom != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(sls_master.created_at) >= ? ", CondValue: arrData.DateFrom},
		)
	}

	if arrData.DateTo != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(sls_master.created_at) <= ? ", CondValue: arrData.DateTo},
		)
	}

	if arrData.DocType != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " sls_master.doc_type = ? ", CondValue: arrData.DocType},
		)

		if arrData.DocType == "BOT" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "AP"},
			)
		}
	}

	if arrData.PrdCode != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " prd_master.code = ? ", CondValue: arrData.PrdCode},
		)
	}

	sysLangCode, _ := models.GetLanguage(arrData.LangCode)

	if sysLangCode == nil || sysLangCode.ID == "" {
		return nil, "something_is_wrong"
	}

	// var sysLangID int
	// if sysLangCode.ID != "" {
	// 	sysLangIDInt, err := strconv.Atoi(sysLangCode.ID)
	// 	if err != nil {
	// 		models.ErrorLog("GetMemberSalesListv1_invalid_language_code", err.Error(), "invalid_langcode_is_received")
	// 		return arrDataReturn
	// 	}
	// 	sysLangID = sysLangIDInt
	// }

	arrPaginateData, arrMemberSalesList, _ := models.GetSlsMasterPaginateFn(arrCond, arrData.Page, false)
	arrNewMemberSalesList := make([]MemberSalesListStruct, 0)
	if len(arrMemberSalesList) > 0 {
		//check whether is first purchase
		arrSlsMasterFn := make([]models.WhereCondFn, 0)
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: arrData.MemberID},
			models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "sls_master.status IN(?,'P') ", CondValue: "AP"},
			models.WhereCondFn{Condition: "sls_master.grp_type = ? ", CondValue: "0"},
		)
		arrSlsMaster, _ := models.GetSlsMasterAscFn(arrSlsMasterFn, "", false)

		for _, arrMemberSalesListV := range arrMemberSalesList {
			// sliceOfLanguageID := helpers.StringToSliceInt(arrMemberSalesListV.LanguageID, ",")
			// intInSliceRst := helpers.IntInSlice(sysLangID, sliceOfLanguageID)
			// if !intInSliceRst {
			// 	continue
			// }
			// params := make(map[string]string)
			createdAtString := arrMemberSalesListV.CreatedAt.Format("2006-01-02 15:04:05")
			expiredAtString := arrMemberSalesListV.ExpiredAt.Format("2006-01-02 15:04:05")

			var arrMemberSalesListPayment = make([]MemberSalesListPaymentStruct, 0)

			rebatePerc := fmt.Sprintf("%g", float.RoundUp(arrMemberSalesListV.RebatePerc*100, 2)) + "%"

			var (
				nftAmt             float64
				nftAirdropAmt      float64
				nftTier            string
				topupStatus        int
				topupHistoryStatus int
				topupMultipleOf    float64
				refundStatus       int
				refundedAmount     float64
				availableRefundAmt float64
				MultipleOf         float64
				Min                float64
				// conversionRate        float64
				signingKeySetting     map[string]interface{}
				pdfContract           map[string]interface{}
				unstakePaymentSetting []wallet_service.PaymentSetting // preparing for future enhancement on listing performance issue
				topupPaymentSetting   []wallet_service.PaymentSetting
			)

			// get product group setting from prd master
			arrPrdMasterFn := make([]models.WhereCondFn, 0)
			arrPrdMasterFn = append(arrPrdMasterFn,
				models.WhereCondFn{Condition: " prd_master.id = ? ", CondValue: arrMemberSalesListV.PrdMasterID},
			)
			prdMaster, prdMasterErr := models.GetPrdMasterFn(arrPrdMasterFn, "", false)

			if prdMasterErr != nil {
				base.LogErrorLog("salesService:GetMemberSalesListv1()", "GetPrdMasterFn()", prdMasterErr.Error(), true)
				return nil, "something_went_wrong"
			}

			if len(prdMaster) > 0 {
				arrPrdGroupTypeSetup, errMsg := GetPrdGroupTypeSetup(prdMaster[0].PrdGroupSetting)
				if errMsg != "" {
					return nil, errMsg
				}

				Min = arrPrdGroupTypeSetup.KeyinMin
				MultipleOf = arrPrdGroupTypeSetup.KeyinMultipleOf
			}

			// if is staking
			// if arrMemberSalesListV.DocType == "STK" {
			// 	// get refund setting
			// 	refundSettingErrMsg, refundSetting := MapProductRefundSetting(arrMemberSalesListV.RefundSetting)
			// 	if refundSettingErrMsg != "" {
			// 		base.LogErrorLog("salesService:GetMemberSalesListv1()", "MapProductRefundSetting()", refundSettingErrMsg, true)
			// 		return nil, refundSettingErrMsg
			// 	}

			// 	if refundSetting.Status && arrMemberSalesListV.Status == "AP" {
			// 		refundStatus = 1

			// 		// get unstake signing key
			// 		db := models.GetDB() // no need set begin transaction
			// 		cryptoAddr, err := member_service.ProcessGetMemAddress(db, arrData.MemberID, arrMemberSalesListV.PrdMasterCode)
			// 		if err != nil {
			// 			base.LogErrorLog("salesService:GetMemberSalesListv1()", "ProcessGetMemAddress():1", err.Error(), true)
			// 			return nil, "something_went_wrong"
			// 		}

			// 		// unstakePaymentSetting, _ = wallet_service.GetPaymentSettingByModule(arrMemberSalesListV.MemberID, "UNSTAKE", arrMemberSalesListV.PrdMasterCode, arrData.LangCode, false)
			// 		unstakeSigningKeySetting, unstakeSigningKeySettingErrMsg := wallet_service.GetSigningKeySettingByModule(arrMemberSalesListV.PrdMasterCode, cryptoAddr, "UNSTAKE")
			// 		if unstakeSigningKeySettingErrMsg != "" {
			// 			return nil, unstakeSigningKeySettingErrMsg
			// 		}
			// 		signingKeySetting = unstakeSigningKeySetting

			// 		// get additional info for transaction data
			// 		arrCond := make([]models.WhereCondFn, 0)
			// 		arrCond = append(arrCond,
			// 			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrMemberSalesListV.PrdMasterCode},
			// 			models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
			// 		)
			// 		ewtSetup, ewtSetupErr := models.GetEwtSetupFn(arrCond, "", false)

			// 		if ewtSetupErr != nil {
			// 			base.LogErrorLog("salesService:GetMemberSalesListv1()", "GetEwtSetupFn():1", ewtSetupErr.Error(), true)
			// 			return nil, "something_went_wrong"
			// 		}

			// 		var isBase bool
			// 		if ewtSetup.IsBase == 1 {
			// 			isBase = true
			// 		}
			// 		signingKeySetting["decimal_point"] = ewtSetup.BlockchainDecimalPoint
			// 		signingKeySetting["is_base"] = isBase

			// 		// get latest conversion rate
			// 		// priceMovement, priceMovementErr := wallet_service.GetLatestPriceMovementByEwtTypeCode(arrMemberSalesListV.PrdMasterCode)
			// 		// if priceMovementErr != "" {
			// 		// 	return nil, "something_went_wrong"
			// 		// }
			// 		// conversionRate = priceMovement
			// 	}

			// 	// get refunded amount
			// 	if refundSetting.Status {
			// 		arrGetTotalRefundedAmount, getTotalRefundedAmountErr := models.GetTotalRefundedAmount(arrMemberSalesListV.ID)
			// 		if getTotalRefundedAmountErr != nil {
			// 			base.LogErrorLog("salesService:GetMemberSalesListv1()", "GetTotalRefundedAmount()", getTotalRefundedAmountErr.Error(), true)
			// 			return nil, "something_went_wrong"
			// 		}
			// 		refundedAmount = arrGetTotalRefundedAmount.TotRefundedAmount
			// 		availableRefundAmt = float.Sub(arrMemberSalesListV.TotalAmount, refundedAmount)
			// 	}

			// 	// deduct refund amount and convert the total amount to usds
			// 	arrMemberSalesListV.TotalAmount = float.Mul(float.Sub(arrMemberSalesListV.TotalAmount, refundedAmount), arrMemberSalesListV.TokenRate)

			// 	for _, payment := range arrMemberSalesListV.Payment {
			// 		amount := payment.Amount
			// 		if arrMemberSalesListV.CurrencyCode == payment.CurrencyCode {
			// 			amount = float.Sub(amount, refundedAmount)
			// 		}
			// 		arrMemberSalesListPayment = append(arrMemberSalesListPayment, MemberSalesListPaymentStruct{
			// 			Amount:       helpers.CutOffDecimal(amount, uint(payment.DecimalPoint), ".", ","),
			// 			CurrencyCode: payment.CurrencyCode,
			// 		})
			// 	}
			// } else {
			// 	for _, payment := range arrMemberSalesListV.Payment {
			// 		additionalInfo := ""
			// 		if arrMemberSalesListV.BatchNo != "" {
			// 			additionalInfo = "(" + helpers.TranslateV2("batch_no", arrData.LangCode, nil) + ": " + arrMemberSalesListV.BatchNo + ")"
			// 		}
			// 		arrMemberSalesListPayment = append(arrMemberSalesListPayment, MemberSalesListPaymentStruct{
			// 			Amount:         helpers.CutOffDecimal(payment.Amount, uint(payment.DecimalPoint), ".", ","),
			// 			CurrencyCode:   payment.CurrencyCode,
			// 			AdditionalInfo: additionalInfo,
			// 		})
			// 	}
			// }

			for _, payment := range arrMemberSalesListV.Payment {
				additionalInfo := ""
				if arrMemberSalesListV.BatchNo != "" {
					additionalInfo = "(" + helpers.TranslateV2("batch_no", arrData.LangCode, nil) + ": " + arrMemberSalesListV.BatchNo + ")"
				}
				arrMemberSalesListPayment = append(arrMemberSalesListPayment, MemberSalesListPaymentStruct{
					Amount:         helpers.CutOffDecimal(payment.Amount, uint(payment.DecimalPoint), ".", ","),
					CurrencyCode:   payment.CurrencyCode,
					AdditionalInfo: additionalInfo,
				})
			}

			// topup setting
			if arrMemberSalesListV.TopupSetting != "" {
				productTopupSetting, _ := GetProductTopupSetting(arrMemberSalesListV.TopupSetting)

				if productTopupSetting.HistoryStatus == true {
					topupHistoryStatus = 1
				}

				if productTopupSetting.Status == true {
					if arrMemberSalesListV.Status == "AP" {
						topupStatus = 1
						topupMultipleOf = productTopupSetting.MultipleOf
					}
				}
			}

			if topupStatus == 1 {
				topupPaymentSetting, _ = wallet_service.GetPaymentSettingByModule(arrMemberSalesListV.MemberID, "CONTRACT_TOPUP", "DEFAULT", arrMemberSalesListV.ProductCurrencyCode, arrData.LangCode, false)
			}

			statusColorCode := "#13B126"
			if arrMemberSalesListV.Status == "V" || arrMemberSalesListV.Status == "F" || arrMemberSalesListV.Status == "EP" {
				statusColorCode = "#F76464"
			} else if arrMemberSalesListV.Status == "P" {
				statusColorCode = "#FFA500"
			}

			var prdName = helpers.TranslateV2(arrMemberSalesListV.PrdMasterName, sysLangCode.Locale, nil)

			if arrMemberSalesListV.DocType == "NFT" {
				// get total nft
				var arrEwtDetailsFn = make([]models.WhereCondFn, 0)
				arrEwtDetailsFn = append(arrEwtDetailsFn,
					models.WhereCondFn{Condition: " ewt_detail.doc_no LIKE ?", CondValue: arrMemberSalesListV.DocNo},
					models.WhereCondFn{Condition: " ewt_detail.transaction_type LIKE ?", CondValue: "NFT"},
				)
				arrEwtDetails, _ := models.GetEwtDetailFn(arrEwtDetailsFn, false)

				for _, arrEwtDetailsV := range arrEwtDetails {
					nftAmt += arrEwtDetailsV.TotalIn
				}

				// get total nft airdrop
				var arrEwtDetailsFn2 = make([]models.WhereCondFn, 0)
				arrEwtDetailsFn2 = append(arrEwtDetailsFn2,
					models.WhereCondFn{Condition: " ewt_detail.doc_no LIKE ?", CondValue: arrMemberSalesListV.DocNo},
					models.WhereCondFn{Condition: " ewt_detail.transaction_type LIKE ?", CondValue: "NFT_AIRDROP"},
				)
				arrEwtDetails2, _ := models.GetEwtDetailFn(arrEwtDetailsFn2, false)

				for _, arrEwtDetailsV := range arrEwtDetails2 {
					nftAirdropAmt += arrEwtDetailsV.TotalIn
				}

				// get tier
				nftTier = "1"
				if arrMemberSalesListV.AirdropRate == 0.1 {
					nftTier = "2"
				} else if arrMemberSalesListV.AirdropRate == 0.2 {
					nftTier = "3"
				} else if arrMemberSalesListV.AirdropRate == 0.3 {
					nftTier = "4"
				}
			}

			// process mining machine data
			// if arrMemberSalesListV.DocType == "MM" {
			// 	// get sls_master_mining
			// 	arrSlsMasterMiningFn := make([]models.WhereCondFn, 0)
			// 	arrSlsMasterMiningFn = append(arrSlsMasterMiningFn,
			// 		models.WhereCondFn{Condition: "sls_master_mining.sls_master_id = ? ", CondValue: arrMemberSalesListV.ID},
			// 	)
			// 	arrSlsMasterMining, err := models.GetSlsMasterMiningFn(arrSlsMasterMiningFn, "", false)
			// 	if err != nil {
			// 		base.LogErrorLog("salesService:GetMemberSalesListv1():GetSlsMasterMiningFn()", err.Error(), map[string]interface{}{"cond": arrSlsMasterMiningFn}, true)
			// 		return nil, "something_went_wrong"
			// 	}

			// 	if len(arrSlsMasterMining) > 0 {
			// 		if arrSlsMasterMining[0].SecTib > 0 {
			// 			secMiningMachine = fmt.Sprintf("%s TiB", helpers.CutOffDecimal(arrSlsMasterMining[0].SecTib, 4, ".", ""))
			// 			secMiningMachineSubDetails = fmt.Sprintf("%s = %sU/TiB", helpers.TranslateV2("price", arrData.LangCode, nil), helpers.CutOffDecimal(arrSlsMasterMining[0].SecPrice, 0, ".", ""))
			// 		}

			// 		if arrSlsMasterMining[0].FilTib > 0 {
			// 			fileMiningMachine = fmt.Sprintf("%s TiB", helpers.CutOffDecimal(arrSlsMasterMining[0].FilTib, 4, ".", ""))
			// 			filMiningMachineSubDetails = fmt.Sprintf("%s = %sU/TiB", helpers.TranslateV2("price", arrData.LangCode, nil), helpers.CutOffDecimal(float.Mul(arrSlsMasterMining[0].FilPrice, arrSlsMasterMining[0].FilecoinPrice), 0, ".", ""))
			// 		}

			// 		if arrSlsMasterMining[0].XchTib > 0 {
			// 			chiaMiningMachine = fmt.Sprintf("%s TiB", helpers.CutOffDecimal(arrSlsMasterMining[0].XchTib, 4, ".", ""))
			// 			chiaMiningMachineSubDetails = fmt.Sprintf("%s = %sU/TiB", helpers.TranslateV2("price", arrData.LangCode, nil), helpers.CutOffDecimal(arrSlsMasterMining[0].XchPrice, 0, ".", ""))
			// 		}

			// 		if arrSlsMasterMining[0].BzzTib > 0 {
			// 			bzzMiningMachine = fmt.Sprintf("%s %s", helpers.CutOffDecimalv2(arrSlsMasterMining[0].BzzTib, 4, ".", ",", true), helpers.TranslateV2("per_nodes", arrData.LangCode, nil))
			// 			bzzMiningMachineSubDetails = fmt.Sprintf("%s = %sU", helpers.TranslateV2("price", arrData.LangCode, nil), helpers.CutOffDecimal(arrSlsMasterMining[0].BzzPrice, 0, ".", ""))

			// 			if arrMemberSalesListV.TotalBv <= 0 {
			// 				prdName = bzzMiningMachine
			// 			}
			// 		}

			// 		checkActiveMemberMiningNodeRst := CheckActiveMemberMiningNode(arrMemberSalesListV.MemberID)
			// 		// checkActiveMemberMiningNodeRst = false // temporary off pdf
			// 		if checkActiveMemberMiningNodeRst {

			// 			btnDownloadContractDisplay = 1

			// 			arrProcessGenerateBZZContractPDF := ProcessGenerateBZZContractPDFStruct{
			// 				NickName:     arrData.NickName,
			// 				MemberID:     arrMemberSalesListV.MemberID,
			// 				SlsMasterID:  arrMemberSalesListV.ID,
			// 				DocNo:        arrMemberSalesListV.DocNo,
			// 				LangCode:     arrData.LangCode,
			// 				TotalAmount:  helpers.CutOffDecimalv2(arrMemberSalesListV.TotalAmount, 4, ".", ",", true),
			// 				TotalNode:    helpers.CutOffDecimalv2(arrSlsMasterMining[0].BzzTib, 4, ".", ",", true),
			// 				SerialNumber: arrSlsMasterMining[0].SerialNumber,
			// 			}
			// 			err := ProcessGenerateBZZContractPDF(arrProcessGenerateBZZContractPDF)

			// 			if err != nil {
			// 				base.LogErrorLog("GetMemberSalesListv1-ProcessGenerateBZZContractPDF_failed", err.Error(), "", true)
			// 				btnDownloadContractDisplay = 0
			// 			}
			// 		}

			// 		apiServerDomain := setting.Cfg.Section("custom").Key("ApiServerDomain").String()
			// 		contractViewUrl := apiServerDomain + "/member/sales/view/node/" + arrMemberSalesListV.DocNo + "_node_contract_en.pdf"
			// 		contractDownloadUrl := apiServerDomain + "/member/sales/download/node/" + arrMemberSalesListV.DocNo + "_node_contract_en.pdf"

			// 		if strings.ToLower(arrData.LangCode) == "zh" {
			// 			contractViewUrl = apiServerDomain + "/member/sales/view/node/" + arrMemberSalesListV.DocNo + "_node_contract_zh.pdf"
			// 			contractDownloadUrl = apiServerDomain + "/member/sales/download/node/" + arrMemberSalesListV.DocNo + "_node_contract_zh.pdf"
			// 		}
			// 		pdfContract = map[string]interface{}{
			// 			"contract_view_url":             contractViewUrl,
			// 			"contract_download_url":         contractDownloadUrl,
			// 			"btn_download_contract_display": btnDownloadContractDisplay,
			// 		}
			// 	}
			// }

			var statusDesc = arrMemberSalesListV.StatusDesc

			if arrMemberSalesListV.DocType == "STK" {
				if arrMemberSalesListV.Status == "AP" {
					statusDesc = "staking"
				} else if arrMemberSalesListV.Status == "EP" {
					statusDesc = "complete"
				}
			}

			var (
				autoTradingApiIcon            = ""
				autoTradingSettingType        = ""
				cryptoPair                    = ""
				autoTradingColorCode          = ""
				autoTradingSetting            map[string]interface{}
				autoTradingSettingStruct      []map[string]interface{}
				autoTradingOperatingColorCode = "#FFA500"
				salesType                     = ""
			)
			if arrMemberSalesListV.DocType == "BOT" {
				// get auto bot related info
				arrSlsMasterBotSettingFn := make([]models.WhereCondFn, 0)
				arrSlsMasterBotSettingFn = append(arrSlsMasterBotSettingFn,
					models.WhereCondFn{Condition: " sls_master_bot_setting.sls_master_id = ?", CondValue: arrMemberSalesListV.ID},
				)
				arrSlsMasterBotSetting, err := models.GetSlsMasterBotSetting(arrSlsMasterBotSettingFn, "", false)
				if err != nil {
					base.LogErrorLog("tradingService:GetMemberSalesListv1():GetSlsMasterBotSetting():1", err.Error(), arrSlsMasterBotSettingFn, true)
				}
				if len(arrSlsMasterBotSetting) > 0 {
					autoTradingSettingType = arrSlsMasterBotSetting[0].SettingType
					cryptoPair = strings.Replace(arrSlsMasterBotSetting[0].CryptoPair, "USDT", "/USDT", 1) // in case sys_trading_crypto_pair_setup do not have this data

					arrSysTradingCryptoPairSetupFn := make([]models.WhereCondFn, 0)
					arrSysTradingCryptoPairSetupFn = append(arrSysTradingCryptoPairSetupFn,
						models.WhereCondFn{Condition: " code = ?", CondValue: arrSlsMasterBotSetting[0].CryptoPair},
					)
					arrSysTradingCryptoPairSetup, err := models.GetSysTradingCryptoPairSetupByPlatformFn(arrSysTradingCryptoPairSetupFn, "", arrSlsMasterBotSetting[0].Platform, false)
					if err != nil {
						base.LogErrorLog("tradingService:GetMemberSalesListv1():GetSysTradingCryptoPairSetupByPlatformFn():1", err.Error(), map[string]interface{}{"param": arrSysTradingCryptoPairSetupFn, "platform": arrSlsMasterBotSetting[0].Platform}, true)
					}
					if len(arrSysTradingCryptoPairSetup) > 0 {
						cryptoPair = arrSysTradingCryptoPairSetup[0].Name
					}

					err = json.Unmarshal([]byte(arrSlsMasterBotSetting[0].Setting), &autoTradingSetting)
					if err != nil {
						autoTradingSetting = nil
					}

					autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
						"label": helpers.TranslateV2("trading_amount", arrData.LangCode, map[string]string{}),
						"value": helpers.CutOffDecimal(arrMemberSalesListV.TotalAmount, uint(arrMemberSalesListV.DecimalPoint), ".", ","),
					})

					if len(autoTradingSetting) > 0 {
						for autoTradingSettingK, autoTradingSettingV := range autoTradingSetting {
							// Spot grid
							if autoTradingSettingK == "gridQuantity" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("grid_quantity", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							} else if autoTradingSettingK == "lowerPrice" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("lower_price", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							} else if autoTradingSettingK == "mode" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("mode", arrData.LangCode, map[string]string{}),
									"value": helpers.TranslateV2(autoTradingSettingV.(string)+"_mode", arrData.LangCode, map[string]string{}),
								})
							} else if autoTradingSettingK == "upperPrice" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("upper_price", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							}

							// Martingale/Reverse Martingale
							if autoTradingSettingK == "addShares" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("add_shares", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							} else if autoTradingSettingK == "firstOrderAmount" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("first_order_amount", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							} else if autoTradingSettingK == "firstOrderPrice" {
								value := fmt.Sprintf("%v", autoTradingSettingV)
								if value == "0" {
									value = helpers.TranslateV2("min", arrData.LangCode, map[string]string{})
								} else if value == "100000" {
									value = helpers.TranslateV2("max", arrData.LangCode, map[string]string{})
								}

								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("first_order_price", arrData.LangCode, map[string]string{}),
									"value": value,
								})
							} else if autoTradingSettingK == "priceScale" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("price_scale", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							} else if autoTradingSettingK == "safetyOrders" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("safety_orders", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							} else if autoTradingSettingK == "takeProfitAdjust" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("take_profit_adjustment", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							} else if autoTradingSettingK == "takeProfitRatio" {
								autoTradingSettingStruct = append(autoTradingSettingStruct, map[string]interface{}{
									"label": helpers.TranslateV2("take_profit_ratio", arrData.LangCode, map[string]string{}),
									"value": fmt.Sprintf("%v", autoTradingSettingV),
								})
							}
						}
					}
				}

				// color code
				autoTradingColorCode = helpers.AutoTradingColorCode(arrMemberSalesListV.PrdMasterCode)

				// api icon
				if arrSlsMasterBotSetting[0].Platform == "KC" {
					autoTradingApiIcon = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/KUCoin.png"
				} else {
					autoTradingApiIcon = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/Binance_Logo_x3.png"
				}

				// retrieve operating status
				arrEntMemberTradingTransactionFn := make([]models.WhereCondFn, 0)
				arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
					models.WhereCondFn{Condition: " 1=? AND ent_member_trading_transaction.doc_no In('" + arrMemberSalesListV.DocNo + "','" + arrMemberSalesListV.RefNo + "')", CondValue: 1},
				)
				arrEntMemberTradingTransaction, err := models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
				if err != nil {
					base.LogErrorLog("tradingService:GetMemberSalesListv1():GetEntMemberTradingTransactionFn():1", err.Error(), arrEntMemberTradingTransactionFn, true)
				}
				if len(arrEntMemberTradingTransaction) > 0 {
					autoTradingOperatingColorCode = "#309304"
				}
			}

			if arrMemberSalesListV.DocType == "CT" {
				salesType = "package_b_topup"

				if len(arrSlsMaster) > 0 {
					if arrSlsMaster[0].BatchNo == arrMemberSalesListV.DocNo {
						salesType = "package_b_subscription"
					}
				}
			}

			arrNewMemberSalesList = append(arrNewMemberSalesList,
				MemberSalesListStruct{
					PrdCode:            arrMemberSalesListV.PrdMasterCode,
					PrdName:            prdName,
					SalesType:          helpers.TranslateV2(salesType, sysLangCode.Locale, nil),
					RebatePerc:         rebatePerc,
					TopupHistoryStatus: topupHistoryStatus,
					TopupSetting: map[string]interface{}{
						"status":          topupStatus,
						"multiple_of":     topupMultipleOf,
						"payment_setting": topupPaymentSetting,
					},
					RefundSetting: map[string]interface{}{
						"status":                  refundStatus,
						"refunded_amount":         refundedAmount,
						"available_refund_amount": availableRefundAmt,
						"min":                     Min,
						"multiple_of":             MultipleOf,
						// "conversion_rate":         conversionRate,
						"transaction_data": signingKeySetting,
						"payment_setting":  unstakePaymentSetting,
					},
					StatusDesc:                    helpers.TranslateV2(statusDesc, sysLangCode.Locale, nil),
					StatusCode:                    arrMemberSalesListV.Status,
					StatusColorCode:               statusColorCode,
					DocNo:                         arrMemberSalesListV.DocNo,
					TotalUnit:                     arrMemberSalesListV.TotUnit,
					TotalAmount:                   helpers.CutOffDecimal(arrMemberSalesListV.TotalAmount, uint(arrMemberSalesListV.DecimalPoint), ".", ","),
					TokenRate:                     helpers.CutOffDecimal(arrMemberSalesListV.TokenRate, 8, ".", ","),
					ExchangeRate:                  helpers.CutOffDecimal(arrMemberSalesListV.ExchangeRate, 8, ".", ","),
					Nft:                           helpers.CutOffDecimal(nftAmt, 4, ".", ","),
					NftAirdrop:                    helpers.CutOffDecimal(nftAirdropAmt, 4, ".", ","),
					NftAirdropType:                fmt.Sprintf("%g%%", arrMemberSalesListV.AirdropRate*100),
					NftTier:                       nftTier,
					TotalTopupAmount:              helpers.CutOffDecimal(arrMemberSalesListV.TotalTv, 2, ".", ","),
					CurrencyCode:                  arrMemberSalesListV.ProductCurrencyCode,
					CryptoPair:                    cryptoPair,
					AutoTradingApiIcon:            autoTradingApiIcon,
					AutoTradingSettingType:        autoTradingSettingType,
					AutoTradingSetting:            autoTradingSettingStruct,
					AutoTradingColorCode:          autoTradingColorCode,
					AutoTradingOperatingColorCode: autoTradingOperatingColorCode,
					CreatedAt:                     createdAtString,
					Payment:                       arrMemberSalesListPayment,
					PdfContract:                   pdfContract,
					ExpiredAt:                     expiredAtString,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrNewMemberSalesList,
	}
	return &arrDataReturn, ""
}

// RefundSalesStruct struct
type RefundSalesStruct struct {
	BatchNo       string
	MemberID      int
	DocNo         string
	RequestAmount float64
}

type RefundSalesData struct {
	RefundAmount  float64
	PenaltyAmount float64
}

// RefundSales func
func RefundSales(tx *gorm.DB, sls RefundSalesStruct) (string, RefundSalesData) {
	var (
		memberID              int             = sls.MemberID
		docNo                 string          = sls.DocNo
		requestAmount         float64         = sls.RequestAmount
		refundAmount          float64         = 0
		refundEwalletTypeID   int             = 0
		penaltyAmount         float64         = 0
		penaltyEwalletTypeID  int             = 0
		returnRefundSalesData RefundSalesData = RefundSalesData{}
	)

	// cut keyin amount to 8 decimal places
	requestAmount, err := helpers.ValueToFloat(helpers.CutOffDecimal(requestAmount, 8, ".", ""))
	if err != nil {
		base.LogErrorLog("productService:RefundSales():ValueToFloat():1", err.Error(), map[string]interface{}{"value": requestAmount}, true)
		return "something_went_wrong", RefundSalesData{}
	}

	// validate sls_master.doc_no
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master.doc_no = ? ", CondValue: docNo},
	)
	arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("salesService:RefundSales():GetSlsMasterFn():1", err.Error(), map[string]interface{}{"condition": arrSlsMasterFn}, true)
		return "something_went_wrong", RefundSalesData{}
	}
	if len(arrSlsMaster) <= 0 {
		return "doc_no_not_found", RefundSalesData{}
	}

	// validate sls_master.status
	if arrSlsMaster[0].Status != "AP" {
		if arrSlsMaster[0].Status == "P" {
			return "doc_status_is_still_pending", RefundSalesData{}
		} else if arrSlsMaster[0].Status == "RF" {
			return "doc_already_fully_refunded", RefundSalesData{}
		} else {
			return "doc_status_not_allowed_to_refund", RefundSalesData{}
		}
	}

	slsMasterID := arrSlsMaster[0].ID
	prdMasterID := arrSlsMaster[0].PrdMasterID
	principalAmount := arrSlsMaster[0].TotalAmount

	// validate if got existing pending refund or not
	// arrSlsMasterRefundFn := make([]models.WhereCondFn, 0)
	// arrSlsMasterRefundFn = append(arrSlsMasterRefundFn,
	// 	models.WhereCondFn{Condition: "sls_master_refund.sls_master_id = ? ", CondValue: slsMasterID},
	// 	models.WhereCondFn{Condition: "sls_master_refund.status = ? ", CondValue: "P"},
	// )
	// arrSlsMasterRefund, err := models.GetSlsMasterRefundFn(arrSlsMasterRefundFn, "", false)
	// if err != nil {
	// 	base.LogErrorLog("salesService:RefundSales():GetSlsMasterRefundFn():1", err.Error(), map[string]interface{}{"condition": arrSlsMasterRefundFn}, true)
	// 	return "something_went_wrong", RefundSalesData{}
	// }
	// if len(arrSlsMasterRefund) > 0 {
	// 	return "please_wait_for_previous_refund_request_of_this_doc_to_approve_before_attempting_another_refund", RefundSalesData{}
	// }

	// get doc's product
	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.id = ? ", CondValue: prdMasterID},
	)
	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("salesService:RefundSales():GetPrdMasterFn():1", err.Error(), map[string]interface{}{"condition": arrPrdMasterFn}, true)
		return "something_went_wrong", RefundSalesData{}
	}
	if len(arrPrdMaster) <= 0 {
		return "doc_product_not_found", RefundSalesData{}
	}

	// validate prd_group_type.refund_setting
	var refundSetting, errMsg = GetPrdGroupTypeRefundSetup(arrPrdMaster[0].RefundSetting)
	if errMsg != "" {
		return errMsg, RefundSalesData{}
	}

	if !refundSetting.Status {
		return "doc_cannot_be_refunded", RefundSalesData{}
	}

	// validate refund amount
	if requestAmount <= 0 {
		return "please_enter_valid_amount", RefundSalesData{}
	}

	// validate available refund amount
	arrGetTotalRefundedAmount, err := models.GetTotalRequestedRefundAmount(slsMasterID)
	if err != nil {
		base.LogErrorLog("salesService:RefundSales():GetTotalRefundedAmount():1", err.Error(), map[string]interface{}{"slsMasterID": slsMasterID}, true)
		return "something_went_wrong", RefundSalesData{}
	}
	curRefundedAmt := arrGetTotalRefundedAmount.TotalRequestAmount
	availableRefundAmt := float.Sub(principalAmount, curRefundedAmt)

	// fmt.Println("principal", principalAmount, "I refunded", curRefundedAmt, "I want refund:", requestAmount, "But I can only refund:", availableRefundAmt)

	if requestAmount > availableRefundAmt {
		return "amount_exceed_available_refund_amount", RefundSalesData{}
	}

	// validate refund amount if refund type is full
	if strings.ToUpper(refundSetting.Type) == "FULL" && requestAmount != principalAmount {
		return "required_to_refund_full_principal_amount", RefundSalesData{}
	}

	// validate penalty
	// take latest doc_date of the same action type
	arrSlsMasterLatestFn := make([]models.WhereCondFn, 0)
	arrSlsMasterLatestFn = append(arrSlsMasterLatestFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: arrSlsMaster[0].Action},
		models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
	)
	arrSlsMasterLatest, _ := models.GetSlsMasterAscFn(arrSlsMasterLatestFn, "", false)

	curDate, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	latestSalesDate, _ := time.Parse("2006-01-02", arrSlsMasterLatest[0].CreatedAt.Format("2006-01-02"))

	days := int(curDate.Sub(latestSalesDate).Hours() / 24)

	penaltyPerc := refundSetting.PenaltyPercDef
	for _, penalty := range refundSetting.Penalty {
		if days >= penalty.Min && penaltyPerc > penalty.PenaltyPerc {
			penaltyPerc = penalty.PenaltyPerc
		}
	}

	if penaltyPerc > 0 {
		penaltyAmount = float.Mul(requestAmount, float.Div(float64(penaltyPerc), 100))
	}

	// if refundSetting.PenaltyPerc > 0 {
	// 	// set penalty amount if doc not yet expired
	// 	if helpers.CompareDateTime(time.Now(), "<", arrSlsMaster[0].ExpiredAt) {
	// 		penaltyAmount = float.Mul(requestAmount, float.Div(float64(refundSetting.PenaltyPerc), 100))
	// 	}
	// }

	refundAmount = requestAmount
	if penaltyAmount > 0 {
		refundAmount = refundAmount - penaltyAmount
	}

	if refundAmount <= 0 {
		base.LogErrorLog("salesService:RefundSales()", "invalid_refund_amount", map[string]interface{}{"refundAmount": refundAmount}, true)
		return "something_went_wrong", RefundSalesData{}
	}

	// if refundSetting.refund_ewallet_type_code != ""
	if refundSetting.RefundEwalletTypeCode != "" {
		arrEwtSetupFn := make([]models.WhereCondFn, 0)
		arrEwtSetupFn = append(arrEwtSetupFn,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: refundSetting.RefundEwalletTypeCode},
		)
		arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
		if err != nil {
			base.LogErrorLog("salesService:RefundSales():GetEwtSetupFn():1", err.Error(), map[string]interface{}{"condition": arrEwtSetupFn}, true)
			return "something_went_wrong", RefundSalesData{}
		}
		if arrEwtSetup == nil {
			base.LogErrorLog("salesService:RefundSales()", "refund_ewallet_type_code_not_found", map[string]interface{}{"refundEwtTypeCode": refundSetting.RefundEwalletTypeCode}, true)
			return "something_went_wrong", RefundSalesData{}
		}

		refundEwalletTypeID = arrEwtSetup.ID

		// refund to ewallet_type_id if no need approval then do refund
		// if refundAmount > 0 {
		// 	var refundWallet = wallet_service.SaveMemberWalletStruct{
		// 		EntMemberID:      memberID,
		// 		EwalletTypeID:    refundEwalletTypeID,
		// 		EwalletTypeCode:  refundSetting.RefundEwalletTypeCode,
		// 		CurrencyCode:     arrSlsMaster[0].CurrencyCode,
		// 		TotalIn:          refundAmount, // in wallet currency
		// 		ConversionRate:   1,
		// 		ConvertedTotalIn: refundAmount, // in payment currency
		// 		TransactionType:  "REFUND",
		// 		CreatedBy:        strconv.Itoa(memberID),
		// 		DecimalPlaces:    arrEwtSetup.DecimalPoint,
		// 	}

		// 	_, err = wallet_service.SaveMemberWallet(tx, refundWallet)
		// 	if err != nil {
		// 		base.LogErrorLog("salesService:RefundSales()", "SaveMemberWallet():1", err.Error(), true)
		// 		return "something_went_wrong", RefundSalesData{}
		// 	}
		// }
	}

	// insert to sls_master_refund
	var addSlsMasterRefund = models.SlsMasterRefund{
		BatchNo:              sls.BatchNo,
		SlsMasterID:          slsMasterID,
		MemberID:             memberID,
		RequestAmount:        requestAmount,
		RefundEwalletTypeID:  refundEwalletTypeID,
		RefundAmount:         refundAmount,
		PenaltyEwalletTypeID: penaltyEwalletTypeID,
		PenaltyPerc:          penaltyPerc,
		PenaltyAmount:        penaltyAmount,
		Status:               "P",
		CreatedBy:            fmt.Sprint(memberID),
		// RefundedAt:           base.GetCurrentDateTimeT(),
		// RefundedBy:           fmt.Sprint(memberID),
	}

	_, err = models.AddSlsMasterRefund(tx, addSlsMasterRefund)
	if err != nil {
		base.LogErrorLog("salesService:RefundSales():AddSlsMasterRefund():1", err.Error(), map[string]interface{}{"data": addSlsMasterRefund}, true)
		return "something_went_wrong", RefundSalesData{}
	}

	returnRefundSalesData.RefundAmount = refundAmount
	returnRefundSalesData.PenaltyAmount = penaltyAmount

	return "", returnRefundSalesData
}

// ApproveUnstakeCallback func for unstake callback
func ApproveUnstakeCallback(tx *gorm.DB, bcStatus, docNo string, totalIn float64) string {
	// get sls_master
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: "sls_master.doc_no = ? ", CondValue: docNo},
	)
	arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if err != nil {
		return err.Error()
	}
	if len(arrSlsMaster) <= 0 {
		return "doc_no_not_found_1"
	}

	slsMasterID := arrSlsMaster[0].ID
	principalAmount := arrSlsMaster[0].TotalAmount

	// get sls_master_refund
	arrSlsMasterRefundFn := make([]models.WhereCondFn, 0)
	arrSlsMasterRefundFn = append(arrSlsMasterRefundFn,
		models.WhereCondFn{Condition: "sls_master_refund.sls_master_id = ? ", CondValue: slsMasterID},
		models.WhereCondFn{Condition: "sls_master_refund.status = ? ", CondValue: "P"},
		models.WhereCondFn{Condition: "sls_master_refund.refund_amount = ? ", CondValue: totalIn},
	)
	arrSlsMasterRefund, err := models.GetSlsMasterRefundFn(arrSlsMasterRefundFn, "", false)
	if err != nil {
		return err.Error()
	}
	if len(arrSlsMasterRefund) <= 0 {
		arrSlsMasterRefundFn := make([]models.WhereCondFn, 0)
		arrSlsMasterRefundFn = append(arrSlsMasterRefundFn,
			models.WhereCondFn{Condition: "sls_master_refund.sls_master_id = ? ", CondValue: slsMasterID},
			models.WhereCondFn{Condition: "sls_master_refund.status = ? ", CondValue: bcStatus},
			models.WhereCondFn{Condition: "sls_master_refund.refund_amount = ? ", CondValue: totalIn},
		)
		arrSlsMasterRefund, err = models.GetSlsMasterRefundFn(arrSlsMasterRefundFn, "", false)
		if err != nil {
			return err.Error()
		}
		if len(arrSlsMasterRefund) <= 0 {
			return "no_pending_refund_request_found"
		}
	}

	// set sls_master_refund.status to AP/F
	updateSlsMasterRefundFn := make([]models.WhereCondFn, 0)
	updateSlsMasterRefundFn = append(updateSlsMasterRefundFn,
		models.WhereCondFn{Condition: "sls_master_refund.id = ?", CondValue: arrSlsMasterRefund[0].ID},
	)
	updateSlsMasterRefundCols := map[string]interface{}{"status": bcStatus}
	updateSlsMasterRefundRst := models.UpdatesFnTx(tx, "sls_master_refund", updateSlsMasterRefundFn, updateSlsMasterRefundCols, false)
	if updateSlsMasterRefundRst != nil {
		return err.Error()
	}

	if bcStatus == "AP" {
		// get total_refunded
		arrGetTotalRefundedAmount, err := models.GetTotalRequestedRefundAmount(slsMasterID)
		if err != nil {
			return err.Error()
		}
		curRefundedAmt := arrGetTotalRefundedAmount.TotalRequestAmount

		// if yes set sls_master.status to “RF” if fully refunded
		if curRefundedAmt >= principalAmount {
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: slsMasterID})
			updateColumn := map[string]interface{}{"status": "RF"}
			err = models.UpdatesFnTx(tx, "sls_master", arrUpdCond, updateColumn, false)
			if err != nil {
				return err.Error()
			}
		}
	}

	return ""
}

// SlsMasterCallback func
func SlsMasterCallback(tx *gorm.DB, docNo, hashValue, transactionType, bcStatus string) string {
	var (
		arrSlsMasterDetails []*models.SlsMasterDetailsByDocNo
		err                 error
	)

	if strings.HasPrefix(docNo, "BT") { // by batch
		arrSlsMasterDetails, err = models.GetSlsMasterDetailsByBatchNo(docNo)
		if err != nil {
			return err.Error()
		}
	} else { // single doc
		slsMasterDetails, err := models.GetSlsMasterDetailsByDocNo(docNo)
		if err != nil {
			return err.Error()
		}

		arrSlsMasterDetails = append(arrSlsMasterDetails, slsMasterDetails)
	}

	// return error if no batch_no/doc_no is found
	if len(arrSlsMasterDetails) <= 0 {
		return "invalid_doc_no"
	}

	// get blockchain trans record
	arrBlockCond2 := make([]models.WhereCondFn, 0)
	arrBlockCond2 = append(arrBlockCond2,
		models.WhereCondFn{Condition: "blockchain_trans.doc_no = ?", CondValue: docNo},
		models.WhereCondFn{Condition: "blockchain_trans.log_only = ? ", CondValue: 0},
		models.WhereCondFn{Condition: "blockchain_trans.hash_value != ?", CondValue: hashValue},
		models.WhereCondFn{Condition: "blockchain_trans.status = ?", CondValue: "P"},
	)
	blockchainTrans2, _ := models.GetBlockchainTransArrayFn(arrBlockCond2, false)

	// only approve doc if that there isn't any pending payment
	if len(blockchainTrans2) == 0 {
		for _, arrSlsMasterDetailsV := range arrSlsMasterDetails {
			curDocNo := arrSlsMasterDetailsV.DocNo

			// direct return success if contract is no longer pending
			if arrSlsMasterDetailsV.Status != "P" {
				// insert pool if it is approved for p2p (admin auto approve at 12am will not insert pool, so this callback part need to insert)
				if arrSlsMasterDetailsV.Status == "AP" && transactionType == "P2P" {
					// insert 100% to sec pool
					returnData := InsertSecPool(tx, curDocNo)
					if returnData != nil {
						return "[" + curDocNo + "] - InsertSecPool():" + fmt.Sprintf("%v", returnData)
					}
				}
				continue
			}

			// update sls_master.status
			arrSlsMasterUpdCond := make([]models.WhereCondFn, 0)
			arrSlsMasterUpdCond = append(arrSlsMasterUpdCond,
				models.WhereCondFn{Condition: "sls_master.doc_no = ?", CondValue: curDocNo},
			)

			updateSlsMasterColumn := map[string]interface{}{
				"status":      bcStatus,
				"approved_at": time.Now(),
				"approved_by": "AUTO",
				"updated_by":  "AUTO",
			}

			// only update doc_date when current time is after today's bonus run time
			todayDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")
			todayDateTime, _ := base.GetCurrentTimeV2("yyyy-mm-dd HH:MM:ss")
			bonusDateTime := todayDate + " 00:04:00"

			if todayDateTime >= bonusDateTime {
				updateSlsMasterColumn["doc_date"] = todayDate
				updateSlsMasterColumn["bns_batch"] = todayDate
			}

			err := models.UpdatesFnTx(tx, "sls_master", arrSlsMasterUpdCond, updateSlsMasterColumn, false)

			if err != nil {
				return "[" + curDocNo + "] - " + err.Error()
			}

			// calculate and insert to income cap wallet for contract purchase
			if transactionType == "CONTRACT" || transactionType == "P2P" {
				errMsg := InsertIncomeCapByDocNo(tx, curDocNo)
				if errMsg != "" {
					return "[" + curDocNo + "] - InsertIncomeCapByDocNo()" + errMsg
				}
			}

			if transactionType == "P2P" {
				// insert 100% to sec pool
				returnData := InsertSecPool(tx, curDocNo)
				if returnData != nil {
					return "[" + curDocNo + "] - InsertSecPool():" + fmt.Sprintf("%v", returnData)
				}
			}
		}
	}

	return ""
}

// InsertIncomeCapByDocNo func
func InsertIncomeCapByDocNo(tx *gorm.DB, docNo string) string {
	arrSlsMasterDetails, err := models.GetSlsMasterDetailsByDocNo(docNo)
	if err != nil {
		return err.Error()
	}

	if arrSlsMasterDetails == nil {
		return "invalid_doc_no"
	}

	if arrSlsMasterDetails.Action == "CONTRACT" {
		var leverage float64 = 0

		// Opt 1: leverage based on highest package
		// arrMemberHighestPackageInfo, err := models.GetMemberHighestPackageInfo(arrSlsMasterDetails.MemberID, arrSlsMasterDetails.Action, docNo)
		// if err != nil {
		// 	return err.Error()
		// }
		// if arrMemberHighestPackageInfo != nil {
		// 	leverage = arrMemberHighestPackageInfo.Leverage
		// }

		// OPt 2: leverage based on latest tier
		arrSlsTierFn := make([]models.WhereCondFn, 0)
		arrSlsTierFn = append(arrSlsTierFn,
			models.WhereCondFn{Condition: "sls_tier.member_id = ? ", CondValue: arrSlsMasterDetails.MemberID},
			// models.WhereCondFn{Condition: "sls_tier.created_by = ? ", CondValue: docNo},
			models.WhereCondFn{Condition: "sls_tier.status = ? ", CondValue: "A"},
		)
		arrSlsTier, err := models.GetSlsTierFn(arrSlsTierFn, "", false)
		if err != nil {
			base.LogErrorLog("salesService:InsertIncomeCapByDocNo():GetSlsTierFn():1", map[string]interface{}{"condition": arrSlsTierFn}, err.Error(), true)
			return "something_went_wrong"
		}
		var tier = arrSlsTier[0].Tier

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "prd_group_type.status = ?", CondValue: "A"},
		)
		arrGetPrdGroupType, err := models.GetPrdGroupTypeFn(arrCond, "", false)
		if err != nil {
			base.LogErrorLog("salesService:InsertIncomeCapByDocNo():GetPrdGroupTypeFn()", map[string]interface{}{"prdGroup": "CONTRACT", "arrCond": arrCond}, err.Error(), true)
			return "something_went_wrong"
		}
		if arrGetPrdGroupType[0].Setting == "" {
			base.LogErrorLog("salesService:InsertIncomeCapByDocNo():GetPrdGroupTypeFn()", map[string]interface{}{"data": arrGetPrdGroupType}, "prd_group_type.setting_is_empty", true)
			return "something_went_wrong"
		}

		arrPrdGroupTypeSetup, errMsg := GetPrdGroupTypeSetup(arrGetPrdGroupType[0].Setting)
		if errMsg != "" {
			base.LogErrorLog("salesService:InsertIncomeCapByDocNo():GetPrdGroupTypeFn()", map[string]interface{}{"raw_setting": arrGetPrdGroupType[0].Setting}, errMsg, true)
			return "something_went_wrong"
		}

		for _, arrPrdGroupTypeSetupTier := range arrPrdGroupTypeSetup.Tiers {
			if arrPrdGroupTypeSetupTier.Tier == tier {
				leverage = arrPrdGroupTypeSetupTier.Leverage
			}
		}

		if leverage > 0 {
			// if leverage > 0 && arrSlsMasterDetails.Status == "P" {
			incomeCapSetting, errMsg := MapProductIncomeCapSetting(arrSlsMasterDetails.IncomeCapSetting)
			if errMsg != "" {
				return errMsg
			}

			if incomeCapSetting.Status {
				incomeCap := float.Mul(arrSlsMasterDetails.TotalAmount, leverage)

				db := models.GetDB() // no need set begin transaction

				_, err := wallet_service.SaveMemberWallet(db, wallet_service.SaveMemberWalletStruct{
					EntMemberID:     arrSlsMasterDetails.MemberID,
					EwalletTypeID:   incomeCapSetting.EwalletTypeID,
					TotalIn:         incomeCap,
					TransactionType: "INCOME_CAP",
					DocNo:           docNo,
					Remark:          fmt.Sprintf("#*income_cap*# %g @ %g", float.RoundUp(arrSlsMasterDetails.TotalAmount, 0), leverage),
					CreatedBy:       "AUTO",
				})

				if err != nil {
					return "SaveMemberWallet_failed:" + err.Error()
				}

				// update sls_master.leverage
				updateSlsMasterFn := make([]models.WhereCondFn, 0)
				updateSlsMasterFn = append(updateSlsMasterFn,
					models.WhereCondFn{Condition: "sls_master.id = ?", CondValue: arrSlsMasterDetails.ID},
				)
				updateSlsMasterCols := map[string]interface{}{"leverage": leverage}
				_ = models.UpdatesFnTx(tx, "sls_master", updateSlsMasterFn, updateSlsMasterCols, false)
			}
		}
	}

	return ""
}

// SlsMasterTopupCallback func
func SlsMasterTopupCallback(tx *gorm.DB, docNo, hashValue, transactionType, bcStatus string) string {
	arrSlsMasterTopupFn := make([]models.WhereCondFn, 0)
	arrSlsMasterTopupFn = append(arrSlsMasterTopupFn,
		models.WhereCondFn{Condition: " sls_master_topup.doc_no = ? ", CondValue: docNo},
	)
	arrSlsMasterTopup, err := models.GetSlsMasterTopupFn(arrSlsMasterTopupFn, "", false)
	if err != nil {
		return err.Error()
	}

	// return error if no doc_no is found
	if len(arrSlsMasterTopup) <= 0 {
		return "invalid_doc_no"
	}

	// get blockchain trans record
	arrBlockCond2 := make([]models.WhereCondFn, 0)
	arrBlockCond2 = append(arrBlockCond2,
		models.WhereCondFn{Condition: "blockchain_trans.doc_no = ?", CondValue: docNo},
		models.WhereCondFn{Condition: "blockchain_trans.log_only = ? ", CondValue: 0},
		models.WhereCondFn{Condition: "blockchain_trans.hash_value != ?", CondValue: hashValue},
		models.WhereCondFn{Condition: "blockchain_trans.status = ?", CondValue: "P"},
	)
	blockchainTrans2, _ := models.GetBlockchainTransArrayFn(arrBlockCond2, false)

	// only approve doc if that there isn't any pending payment
	if len(blockchainTrans2) == 0 {
		var (
			slsMasterID                                  int     = arrSlsMasterTopup[0].SlsMasterID
			memID                                        int     = arrSlsMasterTopup[0].MemberID
			status                                       string  = arrSlsMasterTopup[0].Status
			topupAmount                                  float64 = arrSlsMasterTopup[0].TotalAmount
			slsTotalAmount, slsTotalTopupValue, leverage float64
		)

		// direct return success if contract topup is no longer pending
		if status != "P" {
			return ""
		}

		// get leverage
		arrSlsMasterDetails, err := models.GetSlsMasterDetailsByID(slsMasterID)
		if err != nil {
			return err.Error()
		}

		if arrSlsMasterDetails == nil {
			return "sales_not_found"
		}

		// calculate income cap for contract topup
		arrMemberHighestPackageInfo, err := models.GetMemberHighestPackageInfo(memID, arrSlsMasterDetails.Action, docNo)
		if err != nil {
			return "GetMemberHighestPackageInfo():" + err.Error()
		}

		if arrMemberHighestPackageInfo != nil {
			leverage = arrMemberHighestPackageInfo.Leverage
		}

		// insert to income cap wallet if leverage > 0
		if leverage > 0 {
			incomeCapSetting, errMsg := MapProductIncomeCapSetting(arrSlsMasterDetails.IncomeCapSetting)
			if errMsg != "" {
				return errMsg
			}

			if incomeCapSetting.Status {
				incomeCap := float.Mul(topupAmount, leverage)

				_, err := wallet_service.SaveMemberWallet(tx, wallet_service.SaveMemberWalletStruct{
					EntMemberID:     memID,
					EwalletTypeID:   incomeCapSetting.EwalletTypeID,
					TotalIn:         incomeCap,
					TransactionType: "INCOME_CAP",
					DocNo:           docNo,
					Remark:          fmt.Sprintf("#*income_cap*# %g @ %g", float.RoundUp(topupAmount, 0), leverage),
					CreatedBy:       "AUTO",
				})

				if err != nil {
					return "SaveMemberWallet():" + err.Error()
				}
			}
		}

		// update sls_master_topup.status, leverage
		arrSlsMasterTopupUpdCond := make([]models.WhereCondFn, 0)
		arrSlsMasterTopupUpdCond = append(arrSlsMasterTopupUpdCond,
			models.WhereCondFn{Condition: "sls_master_topup.doc_no = ?", CondValue: docNo},
		)
		updateSlsMasterTopupColumn := map[string]interface{}{
			"status":      bcStatus,
			"leverage":    leverage,
			"approved_at": time.Now(),
			"approved_by": "AUTO",
		}
		err = models.UpdatesFnTx(tx, "sls_master_topup", arrSlsMasterTopupUpdCond, updateSlsMasterTopupColumn, false)
		if err != nil {
			return "UpdatesFnTx()1:" + err.Error()
		}

		// recalculate sls_master.total_amount, total_sv and update sls_master
		slsTotalAmount = arrSlsMasterDetails.TotalAmount + topupAmount
		slsTotalTopupValue = arrSlsMasterDetails.TotalTv + topupAmount

		arrSlsMasterUpdCond := make([]models.WhereCondFn, 0)
		arrSlsMasterUpdCond = append(arrSlsMasterUpdCond,
			models.WhereCondFn{Condition: "sls_master.id = ?", CondValue: slsMasterID},
		)
		updateSlsMasterColumn := map[string]interface{}{
			"total_amount": slsTotalAmount,
			"total_tv":     slsTotalTopupValue,
			"updated_by":   "AUTO",
		}
		err = models.UpdatesFnTx(tx, "sls_master", arrSlsMasterUpdCond, updateSlsMasterColumn, false)
		if err != nil {
			return "UpdatesFnTx()2:" + err.Error()
		}
	}

	return ""
}

// MemberSalesTopupStruct struct
type MemberSalesTopupStruct struct {
	MemberID         int
	DocNo            string
	DateFrom, DateTo string
	LangCode         string
	Page             int64
}

// func GetMemberSalesTopupListv1
func GetMemberSalesTopupListv1(arrData MemberSalesTopupStruct) (interface{}, string) {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sls_master_topup.member_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " sls_master.doc_no = ? ", CondValue: arrData.DocNo},
	)

	if arrData.DateFrom != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(sls_master_topup.created_at) >= ? ", CondValue: arrData.DateFrom},
		)
	}

	if arrData.DateTo != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(sls_master_topup.created_at) <= ? ", CondValue: arrData.DateTo},
		)
	}

	sysLangCode, _ := models.GetLanguage(arrData.LangCode)

	if sysLangCode == nil || sysLangCode.ID == "" {
		return nil, "something_is_wrong"
	}

	arrMemberSalesTopupList, _ := models.GetSlsMasterTopupFn(arrCond, "", false)

	// convert to []interface{} for pagination
	var arrListingData []interface{}
	if len(arrMemberSalesTopupList) > 0 {
		for _, arrMemberSalesTopupListV := range arrMemberSalesTopupList {
			var (
				arrEwtTransactions = wallet_service.GetEwtTransactionsByDocNo([]string{arrMemberSalesTopupListV.DocNo})
				arrPayment         = map[string]interface{}{}
			)

			for _, v := range arrEwtTransactions {
				if v.PaidAmount > 0 {
					arrPayment["amount"] = helpers.CutOffDecimal(v.PaidAmount, 8, ".", ",")
					arrPayment["currency_code"] = v.CurrencyCode
					arrPayment["additional_info"] = ""
				}
			}

			arrListingData = append(arrListingData,
				map[string]interface{}{
					"total_amount":      helpers.CutOffDecimal(arrMemberSalesTopupListV.TotalAmount, 2, ".", ","),
					"status_desc":       helpers.TranslateV2(arrMemberSalesTopupListV.StatusDesc, arrData.LangCode, map[string]string{}),
					"status_color_code": helpers.GetStatusColorCodeByStatusCode(arrMemberSalesTopupListV.Status),
					"currency_code":     arrMemberSalesTopupListV.CurrencyCode,
					"payment":           arrPayment,
					"created_at":        arrMemberSalesTopupListV.CreatedAt.Format("2006-01-02 15:04:05"),
				},
			)
		}
	}

	page := base.Pagination{
		Page:    arrData.Page,
		DataArr: arrListingData,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, ""
}

// InsertSecPool func
func InsertSecPool(tx *gorm.DB, docNo string) map[string]interface{} {
	var (
		memberID         int
		ewalletTypeCode  string = "SEC"
		ewalletTypeID    int
		contractAddr     string
		fromCryptoAddr   string
		fromPrivateKey   string
		toCryptoAddr     string
		poolAmount       float64
		signingKeyModule string = "TRANSFER"
	)

	// get sls_master by docNo
	arrSlsMasterDetails, err := models.GetSlsMasterDetailsByDocNo(docNo)
	if err != nil {
		return map[string]interface{}{"error": "GetSlsMasterDetailsByDocNo():" + err.Error(), "docNo": docNo}
	}
	if arrSlsMasterDetails == nil {
		return map[string]interface{}{"error": "GetSlsMasterDetailsByDocNo():invalid_doc_no", "docNo": docNo}
	}
	memberID = arrSlsMasterDetails.MemberID
	poolAmount = arrSlsMasterDetails.TotalAmount

	// get wallet contract address
	arrEwtSetupFn := make([]models.WhereCondFn, 0)
	arrEwtSetupFn = append(arrEwtSetupFn,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: ewalletTypeCode},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
	if arrEwtSetup == nil {
		return map[string]interface{}{"error": "GetEwtSetupFn():ewt_setup_not_found_for_ewallet_type_code", "condition": arrEwtSetupFn}
	}
	ewalletTypeID = arrEwtSetup.ID
	contractAddr = arrEwtSetup.ContractAddress

	// get company crypto address and private key
	arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
		models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: 0},
		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
	)
	arrEntMemberCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)
	if arrEntMemberCrypto == nil {
		return map[string]interface{}{"error": "GetEntMemberCryptoFn():company_address_not_found", "condition": arrEntMemberCryptoFn}
	}
	fromCryptoAddr = arrEntMemberCrypto.CryptoAddress
	fromPrivateKey = arrEntMemberCrypto.PrivateKey

	// get sec pool address
	secP2PPoolWalletInfo, err := models.GetSECP2PPoolWalletInfo()
	if err != nil {
		return map[string]interface{}{"error": "GetSECP2PPoolWalletInfo():" + err.Error()}
	}
	toCryptoAddr = secP2PPoolWalletInfo.WalletAddress

	// get crypto address
	arrExchangeDebitSetting, errMsg := wallet_service.GetSigningKeySettingByModule(ewalletTypeCode, toCryptoAddr, signingKeyModule)
	if errMsg != "" {
		return map[string]interface{}{"error": "GetSigningKeySettingByModule():" + errMsg, "ewalletTypeCode": ewalletTypeCode, "toCryptoAddr": toCryptoAddr, "module": signingKeyModule}
	}

	chainID, _ := helpers.ValueToInt(arrExchangeDebitSetting["chain_id"].(string))
	maxGas, _ := helpers.ValueToInt(arrExchangeDebitSetting["max_gas"].(string))

	// generate signing key
	arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
		TokenType:       ewalletTypeCode,
		ContractAddress: contractAddr,
		FromAddr:        fromCryptoAddr,
		PrivateKey:      fromPrivateKey,
		ToAddr:          toCryptoAddr,
		Amount:          poolAmount,
		ChainID:         int64(chainID),
		MaxGas:          uint64(maxGas),
	}
	signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
	if err != nil {
		return map[string]interface{}{"error": "ProcecssGenerateSignTransaction():" + err.Error(), "arrProcecssGenerateSignTransaction": arrProcecssGenerateSignTransaction}
	}

	// call sign transaction + insert blockchain_trans
	var arrSaveMemberBlochchainWallet = wallet_service.SaveMemberBlochchainWalletStruct{
		EntMemberID:       memberID,
		EwalletTypeID:     ewalletTypeID,
		DocNo:             docNo,
		Status:            "P",
		TransactionType:   "P2P_POOL",
		TransactionData:   signingKey,
		TotalOut:          poolAmount,
		ConversionRate:    1,
		ConvertedTotalOut: poolAmount,
		LogOnly:           1,
	}
	errMsg, _ = wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlochchainWallet)
	if errMsg != "" {
		return map[string]interface{}{"error": "SaveMemberBlochchainWallet():" + errMsg, "arrSaveMemberBlochchainWallet": arrSaveMemberBlochchainWallet}
	}

	return nil
}

// GetMemberMiningNodeList struct
type GetMemberMiningNodeList struct {
	MemberID int
	LangCode string
	Page     int64
}

// func GetMemberMiningNodeListV1
func GetMemberMiningNodeListV1(arrData GetMemberMiningNodeList) (interface{}, string) {
	arrSlsMasterMiningNodeFn := make([]models.WhereCondFn, 0)
	arrSlsMasterMiningNodeFn = append(arrSlsMasterMiningNodeFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "AP"},
	)

	// if arrData.MemberID == 12246 {
	// 	base.LogErrorLogV2("start GetSlsMasterMiningNodeFn:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
	// }
	arrSlsMasterMiningNode, _ := models.GetSlsMasterMiningNodeFn(arrSlsMasterMiningNodeFn, "", false)
	// if arrData.MemberID == 12246 {
	// 	base.LogErrorLogV2("end GetSlsMasterMiningNodeFn:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
	// }
	// convert to []interface{} for pagination
	var arrListingData []interface{}
	if len(arrSlsMasterMiningNode) > 0 {
		// if arrData.MemberID == 12246 {
		// 	base.LogErrorLogV2("start looping:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
		// }

		var (
			nodeLabel                = helpers.TranslateV2("1_node", arrData.LangCode, map[string]string{})
			numberLabel              = helpers.TranslateV2("number", arrData.LangCode, map[string]string{})
			purchasedDateLabel       = helpers.TranslateV2("purchase_date", arrData.LangCode, map[string]string{})
			broadbandStartDateLabel  = helpers.TranslateV2("broadband_start_date", arrData.LangCode, map[string]string{})
			broadbandExpiryDateLabel = helpers.TranslateV2("broadband_expiry_date", arrData.LangCode, map[string]string{})
			statusLabel              = helpers.TranslateV2("node_status", arrData.LangCode, map[string]string{})
			contractAddressLabel     = helpers.TranslateV2("contract_address", arrData.LangCode, map[string]string{})
			walletAddressLabel       = helpers.TranslateV2("wallet_address", arrData.LangCode, map[string]string{})
			// todaySettlementsLabel    = helpers.TranslateV2("today_settlements", arrData.LangCode, map[string]string{})
			totalSettlementsLabel = helpers.TranslateV2("total_accumulated_settlements", arrData.LangCode, map[string]string{})
			totalMinedLabel       = helpers.TranslateV2("total_mined", arrData.LangCode, map[string]string{})
			okLabel               = helpers.TranslateV2("normal", arrData.LangCode, map[string]string{})
			notOkLabel            = helpers.TranslateV2("disconnected", arrData.LangCode, map[string]string{})
		)

		for key, arrSlsMasterMiningNodeV := range arrSlsMasterMiningNode {
			var (
				arrCardData     = []map[string]interface{}{}
				broadbandStatus = "0"
				curDate         = base.GetCurrentDateTimeT().Format("2006-01-02") + " 00:00:00"
				purchaseDate    = arrSlsMasterMiningNodeV.DocDate.Format("2006-01-02 15:04:05")
				startDate       = arrSlsMasterMiningNodeV.StartDate.Format("2006-01-02 15:04:05")
				expiryDate      = arrSlsMasterMiningNodeV.EndDate.Format("2006-01-02 15:04:05")
				expiryLastWeek  = arrSlsMasterMiningNodeV.EndDate.AddDate(0, 0, -7).Format("2006-01-02") + " 00:00:00"
			)

			// overcome problem when saving sls_master, created_at will be overwrite by CURRENT_TIMESTAMP
			var diffSecs = arrSlsMasterMiningNodeV.StartDate.Sub(arrSlsMasterMiningNodeV.DocDate)
			if diffSecs < 10 { // if differences is less than 10 seconds, updated sls_master.created_at to node.start_date
				purchaseDate = startDate

				arrUpdateSlsMasterFn := make([]models.WhereCondFn, 0)
				arrUpdateSlsMasterFn = append(arrUpdateSlsMasterFn,
					models.WhereCondFn{Condition: " sls_master.id = ? ", CondValue: arrSlsMasterMiningNodeV.SlsMasterID},
				)
				arrUpdateSlsMasterCols := map[string]interface{}{
					"created_at": purchaseDate,
				}
				_ = models.UpdatesFn("sls_master", arrUpdateSlsMasterFn, arrUpdateSlsMasterCols, false)
			}

			arrCardData = append(arrCardData, map[string]interface{}{
				"title": nodeLabel,
				"value": "",
				"copy":  0,
			}, map[string]interface{}{
				"title": numberLabel,
				"value": fmt.Sprint(key + 1),
				"copy":  0,
			}, map[string]interface{}{
				"title": purchasedDateLabel,
				"value": purchaseDate,
				"copy":  0,
			})

			if startDate != expiryDate {
				arrCardData = append(arrCardData,
					map[string]interface{}{
						"title": broadbandStartDateLabel,
						"value": startDate,
						"copy":  0,
					},
					map[string]interface{}{
						"title": broadbandExpiryDateLabel,
						"value": expiryDate,
						"copy":  0,
					})
			}

			if expiryLastWeek <= curDate {
				broadbandStatus = "1"
			}

			if arrSlsMasterMiningNodeV.IP != "" {
				var (
					ip               = arrSlsMasterMiningNodeV.IP
					status           = ""
					contractAddress  = ""
					walletAddress    = ""
					totalSettlements = ""
				)

				arrSwarmIPFn := make([]models.WhereCondFn, 0)
				arrSwarmIPFn = append(arrSwarmIPFn,
					models.WhereCondFn{Condition: " swarm_ip.ip = ? ", CondValue: ip},
				)
				arrSwarmIP, _ := models.GetSwarmIPFn(arrSwarmIPFn, "", false)

				if len(arrSwarmIP) > 0 { // get from swarm_ip
					status = arrSwarmIP[0].Status
					contractAddress = arrSwarmIP[0].ContractAddress
					walletAddress = arrSwarmIP[0].WalletAddress
					totalSettlements = arrSwarmIP[0].TotalSettlements
				} else { // call api
					// get ip health status
					arrHeathStatus, errMsg := crypto_service.GetHealthStatus(ip)
					if errMsg == "" {
						status = arrHeathStatus.Status
					}

					// get wallet address
					if walletAddress == "" {
						arrWalletAddress, errMsg := crypto_service.GetWalletAddress(ip)
						if errMsg == "" {
							walletAddress = arrWalletAddress.Ethereum
						}
					}

					// get contract address
					if contractAddress == "" {
						arrContractAddress, errMsg := crypto_service.GetContractAddress(ip)
						if errMsg == "" {
							contractAddress = arrContractAddress.Address
						}
					}

					// get settlements
					arrSettlements, errMsg := crypto_service.GetSettlements(ip)
					if errMsg == "" {
						totalSettlements = arrSettlements.TotalSent
					}

					// insert into swarm_ip
					db := models.GetDB()

					var addSwarmIPFn = models.AddSwarmIP{
						IP:               ip,
						WalletAddress:    walletAddress,
						ContractAddress:  contractAddress,
						Status:           status,
						TotalSettlements: totalSettlements,
						CreatedBy:        "AUTO",
					}

					_, err := models.AddSwarmIPFn(db, addSwarmIPFn)
					if err != nil {
						base.LogErrorLog("salesService:GetMemberMiningNodeListV1():AddSwarmIPFn()", map[string]interface{}{"addSwarmIPFn": addSwarmIPFn}, err.Error(), true)
					}
				}

				// append ip health status
				if status != "" {
					if status == "ok" {
						status = okLabel
					} else {
						status = notOkLabel
					}

					arrCardData = append(arrCardData, map[string]interface{}{
						"title": statusLabel,
						"value": status,
						"copy":  0,
					})
				}

				// append contract address
				if contractAddress != "" {
					arrCardData = append(arrCardData, map[string]interface{}{
						"title": contractAddressLabel,
						"value": contractAddress,
						"copy":  1,
					})
				}

				// append wallet address
				if walletAddress != "" {
					arrCardData = append(arrCardData, map[string]interface{}{
						"title": walletAddressLabel,
						"value": walletAddress,
						"copy":  1,
					})
				}

				// append settlements
				if totalSettlements != "" {
					arrCardData = append(arrCardData, map[string]interface{}{
						"title": totalSettlementsLabel,
						"value": totalSettlements,
						"copy":  0,
					})
				}

				// append total mined
				if walletAddress != "" {
					// grab total mined figure
					arrGetSwarmDataFn := make([]models.WhereCondFn, 0)
					arrGetSwarmDataFn = append(arrGetSwarmDataFn,
						models.WhereCondFn{Condition: " swarm_data.wallet_address = ? ", CondValue: walletAddress},
					)
					arrGetSwarmData, _ := models.GetSwarmDataFn(arrGetSwarmDataFn, "", false)
					if len(arrGetSwarmData) > 0 {
						arrCardData = append(arrCardData, map[string]interface{}{
							"title": totalMinedLabel,
							"value": helpers.CutOffDecimal(arrGetSwarmData[0].TotalMined, 8, ".", ","),
							"copy":  0,
						})
					}
				}
			}

			arrListingData = append(arrListingData,
				map[string]interface{}{
					"node_id":          arrSlsMasterMiningNodeV.ID,
					"doc_no":           arrSlsMasterMiningNodeV.DocNo,
					"start_date":       startDate,
					"end_date":         expiryDate,
					"card_data":        arrCardData,
					"broadband_status": broadbandStatus,
					"refresh_status":   "1",
					"history_status":   "1",
				},
			)
		}
		// if arrData.MemberID == 12246 {
		// 	base.LogErrorLogV2("end looping:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
		// }
	}

	page := base.Pagination{
		Page:    arrData.Page,
		DataArr: arrListingData,
	}

	// if arrData.MemberID == 12246 {
	// 	base.LogErrorLogV2("start PaginationInterfaceV1:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
	// }
	arrDataReturn := page.PaginationInterfaceV1()
	// if arrData.MemberID == 12246 {
	// 	base.LogErrorLogV2("end PaginationInterfaceV1:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
	// }

	return arrDataReturn, ""
}

// func GetMemberMiningNodeListV1
func GetMemberMiningNodeListUpdateV1(memberID, nodeID int, langCode string) []map[string]interface{} {
	var arrCardData = []map[string]interface{}{}

	arrSlsMasterMiningNodeFn := make([]models.WhereCondFn, 0)
	arrSlsMasterMiningNodeFn = append(arrSlsMasterMiningNodeFn,
		models.WhereCondFn{Condition: " sls_master_mining_node.id = ? ", CondValue: nodeID},
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "AP"},
	)
	arrSlsMasterMiningNode, _ := models.GetSlsMasterMiningNodeFn(arrSlsMasterMiningNodeFn, "", false)

	if len(arrSlsMasterMiningNode) <= 0 {
		return arrCardData
	}

	var (
		ip           = arrSlsMasterMiningNode[0].IP
		purchaseDate = arrSlsMasterMiningNode[0].DocDate.Format("2006-01-02 15:04:05")
		startDate    = arrSlsMasterMiningNode[0].StartDate.Format("2006-01-02 15:04:05")
		expiryDate   = arrSlsMasterMiningNode[0].EndDate.Format("2006-01-02 15:04:05")
	)

	arrCardData = append(arrCardData, map[string]interface{}{
		"title": helpers.TranslateV2("1_node", langCode, map[string]string{}),
		"value": "",
		"copy":  0,
	}, map[string]interface{}{
		"title": helpers.TranslateV2("purchase_date", langCode, map[string]string{}),
		"value": purchaseDate,
		"copy":  0,
	})

	if startDate != expiryDate {
		arrCardData = append(arrCardData,
			map[string]interface{}{
				"title": helpers.TranslateV2("broadband_start_date", langCode, map[string]string{}),
				"value": startDate,
				"copy":  0,
			},
			map[string]interface{}{
				"title": helpers.TranslateV2("broadband_expiry_date", langCode, map[string]string{}),
				"value": expiryDate,
				"copy":  0,
			})
	}

	if ip != "" {
		var (
			walletAddress    = ""
			contractAddress  = ""
			status           = ""
			totalSettlements = ""
		)

		arrSwarmIPFn := make([]models.WhereCondFn, 0)
		arrSwarmIPFn = append(arrSwarmIPFn,
			models.WhereCondFn{Condition: " swarm_ip.ip = ? ", CondValue: ip},
		)
		arrSwarmIP, _ := models.GetSwarmIPFn(arrSwarmIPFn, "", false)

		if len(arrSwarmIP) > 0 {
			walletAddress = arrSwarmIP[0].WalletAddress
			contractAddress = arrSwarmIP[0].ContractAddress
		}

		// get wallet address
		if walletAddress == "" {
			arrWalletAddress, errMsg := crypto_service.GetWalletAddress(ip)
			if errMsg == "" {
				walletAddress = arrWalletAddress.Ethereum
			}
		}

		// get contract address
		if contractAddress == "" {
			arrContractAddress, errMsg := crypto_service.GetContractAddress(ip)
			if errMsg == "" {
				contractAddress = arrContractAddress.Address
			}
		}

		// get ip health status
		arrHeathStatus, errMsg := crypto_service.GetHealthStatus(ip)
		if errMsg == "" {
			status = arrHeathStatus.Status
		}

		// get settlements
		arrSettlements, errMsg := crypto_service.GetSettlements(ip)
		if errMsg == "" {
			totalSettlements = arrSettlements.TotalSent
		}

		db := models.GetDB()
		if len(arrSwarmIP) > 0 { // update swarm_ip
			arrUpdateSwarmIpFn := make([]models.WhereCondFn, 0)
			arrUpdateSwarmIpFn = append(arrUpdateSwarmIpFn,
				models.WhereCondFn{Condition: " swarm_ip.id = ? ", CondValue: arrSwarmIP[0].ID},
			)
			arrUpdateSwarmIpCols := map[string]interface{}{
				"wallet_address":    walletAddress,
				"contract_address":  contractAddress,
				"status":            status,
				"total_settlements": totalSettlements,
				"updated_at":        base.GetCurrentTime("2006-01-02 15:04:05"),
				"updated_by":        "AUTO",
			}
			_ = models.UpdatesFn("swarm_ip", arrUpdateSwarmIpFn, arrUpdateSwarmIpCols, false)

		} else { // insert into swarm_ip
			var addSwarmIPFn = models.AddSwarmIP{
				IP:               ip,
				WalletAddress:    walletAddress,
				ContractAddress:  contractAddress,
				Status:           status,
				TotalSettlements: totalSettlements,
				CreatedBy:        "AUTO",
			}

			_, err := models.AddSwarmIPFn(db, addSwarmIPFn)
			if err != nil {
				base.LogErrorLog("salesService:GetMemberMiningNodeListUpdateV1():AddSwarmIPFn()", map[string]interface{}{"addSwarmIPFn": addSwarmIPFn}, err.Error(), true)
			}
		}

		if status == "ok" {
			status = helpers.TranslateV2("normal", langCode, map[string]string{})
		} else {
			status = helpers.TranslateV2("disconnected", langCode, map[string]string{})
		}

		arrCardData = append(arrCardData, map[string]interface{}{
			"title": helpers.TranslateV2("contract_address", langCode, map[string]string{}),
			"value": contractAddress,
			"copy":  1,
		}, map[string]interface{}{
			"title": helpers.TranslateV2("wallet_address", langCode, map[string]string{}),
			"value": walletAddress,
			"copy":  1,
		}, map[string]interface{}{
			"title": helpers.TranslateV2("node_status", langCode, map[string]string{}),
			"value": status,
			"copy":  0,
		}, map[string]interface{}{
			"title": helpers.TranslateV2("total_accumulated_settlements", langCode, map[string]string{}),
			"value": totalSettlements,
			"copy":  0,
		})

		if walletAddress != "" {
			// grab total mined figure
			arrGetSwarmDataFn := make([]models.WhereCondFn, 0)
			arrGetSwarmDataFn = append(arrGetSwarmDataFn,
				models.WhereCondFn{Condition: " swarm_data.wallet_address = ? ", CondValue: walletAddress},
			)
			arrGetSwarmData, _ := models.GetSwarmDataFn(arrGetSwarmDataFn, "", false)
			if len(arrGetSwarmData) > 0 {
				arrCardData = append(arrCardData, map[string]interface{}{
					"title": helpers.TranslateV2("total_mined", langCode, map[string]string{}),
					"value": helpers.CutOffDecimal(arrGetSwarmData[0].TotalMined, 8, ".", ","),
					"copy":  0,
				})
			}
		}
	}

	return arrCardData
}

func CheckActiveMemberMiningNode(memberID int) bool {

	dtNow := base.GetCurrentTime("2006-01-02 15:04:05")

	arrSlsMasterMiningNodeFn := make([]models.WhereCondFn, 0)
	arrSlsMasterMiningNodeFn = append(arrSlsMasterMiningNodeFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "AP"},
		models.WhereCondFn{Condition: " sls_master_mining_node.start_date <= ? ", CondValue: dtNow},
		models.WhereCondFn{Condition: " sls_master_mining_node.end_date >= ? ", CondValue: dtNow},
	)

	arrSlsMasterMiningNode, _ := models.GetSlsMasterMiningNodeFn(arrSlsMasterMiningNodeFn, "", false)

	var result bool
	if len(arrSlsMasterMiningNode) > 1 {
		result = true
	}

	return result
}

// GetMemberMiningNodeTopupList struct
type GetMemberMiningNodeTopupList struct {
	MemberID int
	NodeID   int
	LangCode string
	NickName string
	Page     int64
}

// func GetMemberMiningNodeTopupListV1
func GetMemberMiningNodeTopupListV1(arrData GetMemberMiningNodeTopupList) (interface{}, string) {
	arrSlsMasterMiningNodeTopupFn := make([]models.WhereCondFn, 0)
	arrSlsMasterMiningNodeTopupFn = append(arrSlsMasterMiningNodeTopupFn,
		models.WhereCondFn{Condition: " sls_master_mining_node_topup.member_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " sls_master_mining_node_topup.sls_master_mining_node_id = ? ", CondValue: arrData.NodeID},
	)

	arrPaginateData, arrSlsMasterMiningNodeTopup, _ := models.GetSlsMasterMiningNodeTopupPaginateFn(arrSlsMasterMiningNodeTopupFn, arrData.Page, false)

	var (
		arrListingData        []interface{}
		activeStatusColorCode = "#13B126"
		// pendingStatusColorCode = "#FFA500"
		expiredStatusColorCode     = "#F76464"
		btnDownloadContractDisplay int
		pdfContract                map[string]interface{}
	)

	if len(arrSlsMasterMiningNodeTopup) > 0 {
		for _, arrSlsMasterMiningNodeTopupV := range arrSlsMasterMiningNodeTopup {
			var (
				status          = helpers.TranslateV2(arrSlsMasterMiningNodeTopupV.Status, arrData.LangCode, map[string]string{})
				prdMasterName   = helpers.TranslateV2(arrSlsMasterMiningNodeTopupV.PrdName, arrData.LangCode, map[string]string{})
				createdAt       = arrSlsMasterMiningNodeTopupV.CreatedAt.Format("2006-01-02 15:04:05")
				statusColorCode = activeStatusColorCode
			)

			if arrSlsMasterMiningNodeTopupV.StatusCode != "AP" {
				statusColorCode = expiredStatusColorCode
			}

			arrProcessGenerateBZZContractPDF := ProcessGenerateBZZBroadbandContractPDFStruct{
				NickName:     arrData.NickName,
				MemberID:     arrData.MemberID,
				ID:           arrSlsMasterMiningNodeTopupV.ID,
				DocNo:        arrSlsMasterMiningNodeTopupV.DocNo,
				LangCode:     arrData.LangCode,
				Months:       strconv.Itoa(arrSlsMasterMiningNodeTopupV.Months),
				TotalNode:    "1",
				SerialNumber: strconv.Itoa(arrSlsMasterMiningNodeTopupV.SerialNumber),
			}
			err := ProcessGenerateBZZBroadbandContractPDF(arrProcessGenerateBZZContractPDF)
			btnDownloadContractDisplay = 1
			if err != nil {
				base.LogErrorLog("GetMemberSalesListv1-ProcessGenerateBZZBroadbandContractPDF_failed", err.Error(), "", true)
				btnDownloadContractDisplay = 0
			}

			apiServerDomain := setting.Cfg.Section("custom").Key("ApiServerDomain").String()
			contractViewUrl := apiServerDomain + "/member/sales/view/broadband/" + arrSlsMasterMiningNodeTopupV.DocNo + "_broadband_contract_en.pdf"
			contractDownloadUrl := apiServerDomain + "/member/sales/download/broadband/" + arrSlsMasterMiningNodeTopupV.DocNo + "_broadband_contract_en.pdf"

			if strings.ToLower(arrData.LangCode) == "zh" {
				contractViewUrl = apiServerDomain + "/member/sales/view/broadband/" + arrSlsMasterMiningNodeTopupV.DocNo + "_broadband_contract_zh.pdf"
				contractDownloadUrl = apiServerDomain + "/member/sales/download/broadband/" + arrSlsMasterMiningNodeTopupV.DocNo + "_broadband_contract_zh.pdf"
			}
			pdfContract = map[string]interface{}{
				"contract_view_url":             contractViewUrl,
				"contract_download_url":         contractDownloadUrl,
				"btn_download_contract_display": btnDownloadContractDisplay,
			}

			arrListingData = append(arrListingData,
				map[string]interface{}{
					"doc_no":            arrSlsMasterMiningNodeTopupV.DocNo,
					"prd_master_name":   prdMasterName,
					"status":            status,
					"status_color_code": statusColorCode,
					"created_at":        createdAt,
					"pdf_contract":      pdfContract,
				},
			)
		}
	}

	var arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrListingData,
	}
	return arrDataReturn, ""
}

type WalletData struct {
	TotalBalance     float64
	TotalMinedAmount float64
}

func UpdateWalletData(tx *gorm.DB, walletAddress string, walletData WalletData) string {
	// get member id with eth wallet address
	var (
		arrSalesDetails, _ = models.GetSalesDetailsByWalletAddress(walletAddress, false)
		memberID           = 0
	)

	if len(arrSalesDetails) > 0 {
		memberID = arrSalesDetails[0].MemberID
	}

	// insert to swarm_data table
	var arrSwarmData = models.AddSwarmData{
		MemberID:      memberID,
		WalletAddress: walletAddress,
		TotalBalance:  walletData.TotalBalance,
		TotalMined:    walletData.TotalMinedAmount,
	}

	var _, err = models.AddSwarmDataFn(tx, arrSwarmData)
	if err != nil {
		base.LogErrorLog("salesService:UpdateWalletData():AddSwarmDataFn()", map[string]interface{}{"arrSwarmData": arrSwarmData}, err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

type arrBallotGeneralSettingStruct struct {
	Min           int    `json:"min"`
	Max           int    `json:"max"`
	MultipleOf    int    `json:"multiple_of"`
	CurrencyFrom  string `json:"currency_from"`
	CurrencyTo    string `json:"currency_to"`
	Type          string `json:"type"`
	TicketNoRange int    `json:"ticket_no_range"`
}

type arrBallotGeneralSettingStructv2 struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type PostBallotStruct struct {
	Payments string
	MemberId int
	LangCode string
}

func (b *PostBallotStruct) PostBallot(tx *gorm.DB) (interface{}, error) {
	var (
		err                 error
		ticketNo            int
		totalAmount         float64
		arrGeneralSetting   arrBallotGeneralSettingStruct
		arrGeneralSettingv2 arrBallotGeneralSettingStructv2
	)

	type PaymentStruct struct {
		TypeCode string  `json:"type_code"`
		Amount   float64 `json:"amount"`
	}

	//get general setup
	arrGeneralSetup, err := models.GetSysGeneralSetupByID("ballot_setting")
	if err != nil {
		base.LogErrorLog("PostBallot-GetSysGeneralSetupByID1_failed", err.Error(), "ballot_setting", true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", b.LangCode), Data: err}
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("PostBallot-GetSysGeneralSetupByID2_failed", "ballot_setting_is_not_set", arrGeneralSetup, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", b.LangCode), Data: err}
	}

	json.Unmarshal([]byte(arrGeneralSetup.InputValue1), &arrGeneralSetting)

	json.Unmarshal([]byte(arrGeneralSetup.InputValue2), &arrGeneralSettingv2)

	if arrGeneralSettingv2.StartTime != "" && arrGeneralSettingv2.EndTime != "" {
		currTime := time.Now().Format("2006-01-02 15:04:05")

		if currTime < arrGeneralSettingv2.StartTime {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("ballot_event_not_yet_start", b.LangCode), Data: err}
		}

		if currTime > arrGeneralSettingv2.EndTime {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("ballot_event_ended", b.LangCode), Data: err}
		}

	}

	var paymentStruct map[string][]PaymentStruct
	err = json.Unmarshal([]byte(b.Payments), &paymentStruct)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_payment_struct", b.LangCode), Data: err}
	}

	//check if mem ballot before
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member_ballot.member_id = ? ", CondValue: b.MemberId},
		models.WhereCondFn{Condition: " ent_member_ballot.mem_type != ? ", CondValue: "ghost"},
	)
	arrEntMemberBallot, err := models.GetEntMemberBallotFn(arrCond, false)

	if err != nil {
		models.ErrorLog("PostBallot-GetEntMemberBallotFn()Fail", err.Error(), arrCond)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if len(arrEntMemberBallot) > 0 {
		ticketNo = arrEntMemberBallot[0].TicketNo
	}
	// else {
	// 	latestTickNo, _ := models.GetLatestBallotTicketNo(nil, false)
	// 	ticketNo = latestTickNo.TicketNo + arrGeneralSetting.TicketNoRange
	// }

	//check amount min max
	for _, v := range paymentStruct["payment_groups"] {
		totalAmount += v.Amount
	}

	//every time need at least 10
	if totalAmount < float64(arrGeneralSetting.Min) {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("min_amt_is", b.LangCode) + " " + strconv.Itoa(arrGeneralSetting.Min), Data: err}
	}

	if totalAmount > float64(arrGeneralSetting.Max) {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("max_amt_is", b.LangCode) + " " + strconv.Itoa(arrGeneralSetting.Max), Data: err}
	}

	// 1 acc maximum 1000
	//check member total ballot amt
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_ballot.member_id = ?", CondValue: b.MemberId},
	)
	latestVolume, _ := models.GetLatestBallotVolume(arrCond, false)
	memCurrVolume := latestVolume.Volume

	currMemBallotWithVolume, _ := decimal.NewFromFloat(memCurrVolume).Add(decimal.NewFromFloat(totalAmount)).Float64()

	if memCurrVolume >= float64(arrGeneralSetting.Max) {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("already_reach_max_ballot_amount", b.LangCode), Data: err}
	}

	if currMemBallotWithVolume > float64(arrGeneralSetting.Max) {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "ballot_amount_over_limit", Data: err}
	}

	for _, v := range paymentStruct["payment_groups"] {
		if v.Amount > 0 {

			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ballot_price_setting.type_code = ?", CondValue: v.TypeCode},
			)
			setting, err := models.GetBallotPriceSettingFn(arrCond, "", false)
			if err != nil {
				base.LogErrorLog("PostBallot - fail to get ballot price setting", err, b, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", b.LangCode), Data: err}
			}

			priceRate := setting.PriceRate
			convertedAmount, _ := decimal.NewFromFloat(v.Amount).Mul(decimal.NewFromFloat(priceRate)).Float64()

			if !helpers.IsMultipleOf(v.Amount, float64(arrGeneralSetting.MultipleOf)) {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "ballot_unit_must_be_multiple_of_:0", Data: err}
			}

			// //get current type code volume
			// arrCond = make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: "ent_member_ballot.type_code = ?", CondValue: v.TypeCode},
			// )
			// latestVolume, _ := models.GetLatestBallotVolume(arrCond, false)
			// curVolume := latestVolume.Volume
			// curBallotWithVolume, _ := decimal.NewFromFloat(curVolume).Add(decimal.NewFromFloat(v.Amount)).Float64()

			// if curVolume >= setting.MaxVolume {
			// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "already_reach_max_volume", Data: err}
			// }

			// if curBallotWithVolume > setting.MaxVolume {
			// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "ballot_amount_over_volume_limit", Data: err}
			// }

			// save ent_member_ballot
			arrBallot := models.EntMemberBallot{
				MemberId:        b.MemberId,
				TicketNo:        ticketNo,
				MemType:         "MEM",
				TransDate:       time.Now(),
				TypeCode:        v.TypeCode,
				CurrencyFrom:    arrGeneralSetting.CurrencyTo,
				Amount:          v.Amount,
				Price:           priceRate,
				ConvertedAmount: convertedAmount,
				CurrencyTo:      arrGeneralSetting.CurrencyFrom,
				CreatedAt:       time.Now(),
				CreatedBy:       strconv.Itoa(b.MemberId),
			}

			_, err = models.AddEntMemberBallot(tx, arrBallot)

			if err != nil {
				base.LogErrorLog("PostBallot - fail to save ent_member_ballot", err, arrBallot, true) //store error log
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", b.LangCode), Data: err}
			}
		}
	}

	arrData := make(map[string]interface{})
	arrData["amount"] = fmt.Sprint(totalAmount)
	arrData["trans_time"] = time.Now().Format("2006-01-02 15:04:05")

	return arrData, nil

}

type GetMemberBallotListStructv1 struct {
	TransType string `json:"trans_type"`
	TransDate string `json:"trans_date"`
}

type BallotTransactionStruct struct {
	MemberID int    `json:"member_id"`
	Page     int64  `json:"page"`
	LangCode string `json:"lang_code"`
}

func (s *BallotTransactionStruct) GetMemberBallotListv1() (interface{}, error) {

	arrBallotStatementList := make([]GetMemberBallotListStructv1, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_ballot.member_id = ?", CondValue: s.MemberID},
		models.WhereCondFn{Condition: "ent_member_ballot.mem_type != ?", CondValue: "ghost"},
	)

	arrEntMemberBallot, _ := models.GetEntMemberBallotFn(arrCond, false)

	if len(arrEntMemberBallot) > 0 {
		for _, v := range arrEntMemberBallot {

			arrBallotStatementList = append(arrBallotStatementList,
				GetMemberBallotListStructv1{
					TransType: helpers.CutOffDecimal(v.Price, uint(2), ".", ",") + " " + helpers.Translate(v.CurrencyTo, s.LangCode) + " " + "X" + " " + helpers.CutOffDecimal(v.Amount, 0, ".", ",") + " " + helpers.TranslateV2(v.CurrencyFrom, s.LangCode, nil),
					TransDate: v.TransDate.Format("2006-01-02 15:04:05"),
				})
		}
	}

	//start paginate
	sort.Slice(arrBallotStatementList, func(p, q int) bool {
		return arrBallotStatementList[q].TransDate < arrBallotStatementList[p].TransDate
	})

	arrDataReturn := app.ArrDataResponseList{}

	//general setup default limit rows
	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	curPage := s.Page

	if curPage == 0 {
		curPage = 1
	}

	if s.Page != 0 {
		s.Page--
	}

	totalRecord := len(arrBallotStatementList)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(s.Page), int(limit), totalRecord)

	processArr := arrBallotStatementList[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(curPage),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      processArr,
		TableHeaderList:       nil,
	}

	return arrDataReturn, nil
}

// GetBallotSetting func
func GetBallotSetting(memId int, langCode string) (map[string]interface{}, error) {

	type BallotSettingStruct struct {
		TypeCode     string  `json:"type_code"`
		PriceRate    float64 `json:"price_rate"`
		PriceRateStr string  `json:"price_rate_str"`
		Vesting      string  `json:"vesting"`
		CurrencyFrom string  `json:"currency_from"`
		CurrencyTo   string  `json:"currency_to"`
		MultipleOf   int     `json:"multiple_of"`
	}

	var (
		arrDataReturn       map[string]interface{}
		tokenInfoList       []BallotSettingStruct
		arrGeneralSetting   arrBallotGeneralSettingStruct
		arrGeneralSettingv2 arrBallotGeneralSettingStructv2
		err                 error
	)

	//get general setup
	arrGeneralSetup, err := models.GetSysGeneralSetupByID("ballot_setting")
	if err != nil {
		base.LogErrorLog("GetBallotSetting-GetSysGeneralSetupByID1_failed", err.Error(), "ballot_setting", true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", langCode), Data: err}
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("GetBallotSetting-GetSysGeneralSetupByID2_failed", "ballot_setting_is_not_set", arrGeneralSetup, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", langCode), Data: err}
	}

	json.Unmarshal([]byte(arrGeneralSetup.InputValue1), &arrGeneralSetting)
	json.Unmarshal([]byte(arrGeneralSetup.InputValue2), &arrGeneralSettingv2)

	setting, err := models.GetBallotPriceSettingListFn(nil, "", false)
	if err != nil {
		base.LogErrorLog("GetBallotSetting - fail to get ballot price setting", err, "", true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", langCode), Data: err}
	}

	for _, v := range setting {
		tokenInfoList = append(tokenInfoList,
			BallotSettingStruct{
				TypeCode:     v.TypeCode,
				PriceRate:    v.PriceRate,
				PriceRateStr: helpers.Translate(v.TypeCode, langCode) + " " + "(" + " " + fmt.Sprint(v.PriceRate) + " " + helpers.Translate(arrGeneralSetting.CurrencyFrom, langCode) + ")",
				Vesting:      v.VestingMonth + " " + helpers.Translate("Month", langCode) + " " + helpers.Translate("Vesting", langCode),
				CurrencyFrom: helpers.Translate(arrGeneralSetting.CurrencyFrom, langCode),
				CurrencyTo:   helpers.Translate(arrGeneralSetting.CurrencyTo, langCode),
				MultipleOf:   arrGeneralSetting.MultipleOf,
			})
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_ballot.member_id = ?", CondValue: memId},
	)
	latestVolume, _ := models.GetLatestBallotVolume(arrCond, false)
	memCurrVolume := latestVolume.Volume

	max := float64(arrGeneralSetting.Max)

	availableBalance, _ := decimal.NewFromFloat(max).Sub(decimal.NewFromFloat(memCurrVolume)).Float64()

	availableBalanceStr := helpers.CutOffDecimal(availableBalance, uint(2), ".", ",")

	arrDataReturn = map[string]interface{}{
		"min":                   arrGeneralSetting.Min,
		"max":                   arrGeneralSetting.Max,
		"start_date":            arrGeneralSettingv2.StartTime,
		"end_date":              arrGeneralSettingv2.EndTime,
		"token_info":            tokenInfoList,
		"available_balance_str": availableBalanceStr,
	}

	return arrDataReturn, nil
}

type PostBallotWinnerStruct struct {
	MemberId int
	Address  string
	LangCode string
}

func (b *PostBallotWinnerStruct) PostBallotWinner(tx *gorm.DB) (interface{}, error) {
	var (
		err      error
		ticketNo int
	)

	//get member ticket no
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member_ballot.member_id = ? ", CondValue: b.MemberId},
		models.WhereCondFn{Condition: " ent_member_ballot.mem_type != ? ", CondValue: "ghost"},
	)
	arrEntMemberBallot, err := models.GetEntMemberBallotFn(arrCond, false)

	if err != nil {
		models.ErrorLog("PostBallotWinner-GetEntMemberBallotFn()Fail", err.Error(), arrCond)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if len(arrEntMemberBallot) > 0 {
		ticketNo = arrEntMemberBallot[0].TicketNo
	}

	//check if submitted address before
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member_ballot_winner.ticket_no = ? ", CondValue: ticketNo},
	)
	arrEntMemberBallotWinner, _ := models.GetEntMemberBallotWinnerFn(arrCond, false)

	if len(arrEntMemberBallotWinner) > 0 {
		if arrEntMemberBallotWinner[0].BscAddress != "" {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("already_submit_address_before", b.LangCode), Data: err}
		}
	}

	// update to ent_member_ballot_winner
	updateFn := make([]models.WhereCondFn, 0)
	updateFn = append(updateFn,
		models.WhereCondFn{Condition: "ent_member_ballot_winner.ticket_no = ?", CondValue: ticketNo},
	)
	updateCols := map[string]interface{}{"bsc_address": b.Address}
	updateRst := models.UpdatesFnTx(tx, "ent_member_ballot_winner", updateFn, updateCols, false)
	if updateRst != nil {
		base.LogErrorLog("PostBallotWinner - fail to update winner address", b, ticketNo, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", b.LangCode), Data: err}
	}

	arrData := make(map[string]interface{})
	arrData["trans_time"] = time.Now().Format("2006-01-02 15:04:05")
	arrData["address"] = b.Address

	return arrData, nil

}

// InsertNftAirdrop func
func InsertNftAirdrop(tx *gorm.DB, docNo string) string {
	arrSlsMasterDetails, err := models.GetSlsMasterDetailsByDocNo(docNo)
	if err != nil {
		return err.Error()
	}

	if arrSlsMasterDetails == nil {
		return "invalid_doc_no"
	}

	if arrSlsMasterDetails.Action == "NFT" {
		// give nft airdrop bonus
		if arrSlsMasterDetails.TotalAirdropNft > 0 {
			// get nft airdrop amount
			arrPrdMasterFn := make([]models.WhereCondFn, 0)
			arrPrdMasterFn = append(arrPrdMasterFn,
				models.WhereCondFn{Condition: " prd_master.id = ? ", CondValue: arrSlsMasterDetails.PrdMasterID},
			)
			arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
			if err != nil {
				return "GetPrdMasterFn():" + err.Error()
			}
			if len(arrPrdMaster) <= 0 {
				return fmt.Sprintf("%s%d%s", "GetPrdMasterFn():prdMasterID:", arrSlsMasterDetails.PrdMasterID, "not found")
			}

			var (
				// nftSeriesCode      = arrSlsMasterDetails.NftSeriesCode
				nftEwalletTypeID   = arrSlsMasterDetails.AirdropEwalletTypeID
				nftAirdropBonusAmt = arrSlsMasterDetails.TotalAirdropNft
			)

			// get nft series setup by series code
			// arrNftSeriesSetupFn := make([]models.WhereCondFn, 0)
			// arrNftSeriesSetupFn = append(arrNftSeriesSetupFn,
			// 	models.WhereCondFn{Condition: "nft_series_setup.code = ?", CondValue: nftSeriesCode},
			// 	models.WhereCondFn{Condition: "nft_series_setup.status = ?", CondValue: "A"},
			// )
			// arrNftSeriesSetup, err := models.GetNftSeriesSetupFn(arrNftSeriesSetupFn, "", false)
			// if err != nil {
			// 	return "GetNftSeriesSetupFn():" + err.Error()
			// }
			// if len(arrNftSeriesSetup) > 0 {
			// 	nftEwalletTypeID = arrNftSeriesSetup[0].EwalletTypeID
			// }

			// convert usdt to nft
			// var nftAirdropBonusAmt = float.Div(arrSlsMasterDetails.TotalAirdrop, arrSlsMasterDetails.ExchangeRate)
			// nftAirdropBonusAmt = float.Div(nftAirdropBonusAmt, arrSlsMasterDetails.TokenRate)

			var saveMemberWalletParams = wallet_service.SaveMemberWalletStruct{
				EntMemberID:     arrSlsMasterDetails.MemberID,
				EwalletTypeID:   nftEwalletTypeID,
				TotalIn:         nftAirdropBonusAmt,
				TransactionType: "NFT_AIRDROP",
				DocNo:           docNo,
				Remark:          "#*nft_airdrop*#",
				CreatedBy:       "AUTO",
			}

			_, err = wallet_service.SaveMemberWallet(tx, saveMemberWalletParams)
			if err != nil {
				return "SaveMemberWallet():" + err.Error()
			}
		}
	}

	return ""
}

type MemberSalesListSummaryStruct struct {
	MemberID int
	DocType  string
	LangCode string
}

// func GetMemberSalesListSummary
func GetMemberSalesListSummary(input MemberSalesListSummaryStruct) (interface{}, string) {
	var arrData = map[string]interface{}{}

	if input.DocType == "" || input.DocType == "NFT" {
		var arrDataV = map[string]interface{}{}

		// get member total purchased nft package value (usdt)
		var arrSlsMasterFn = make([]models.WhereCondFn, 0)
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: input.MemberID},
			models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "NFT"},
			models.WhereCondFn{Condition: "sls_master.status = ? ", CondValue: "AP"},
		)
		arrSlsMaster, err := models.GetMemberTotalSalesFn(arrSlsMasterFn, false)
		if err != nil {
			base.LogErrorLog("GetMemberSalesListSummary():GetMemberTotalSalesFn", arrSlsMaster, map[string]interface{}{"condition": arrSlsMasterFn}, true)
			return nil, "something_went_wrong"
		}

		// get nft wallet id
		var arrEwtSetupFn = make([]models.WhereCondFn, 0)
		arrEwtSetupFn = append(arrEwtSetupFn,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "NFT"},
		)
		arrEwtSetup, _ := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
		if arrEwtSetup == nil {
			base.LogErrorLog("GetMemberSalesListSummary():GetEwtSetupFn", arrEwtSetup, map[string]interface{}{"condition": arrEwtSetupFn}, true)
			return nil, "something_went_wrong"
		}

		// get member purchased nft
		var (
			arrEwtDetailsFn = make([]models.WhereCondFn, 0)
			totalPurchased  float64
		)
		arrEwtDetailsFn = append(arrEwtDetailsFn,
			models.WhereCondFn{Condition: " ewt_detail.member_id = ?", CondValue: input.MemberID},
			models.WhereCondFn{Condition: " ewt_detail.ewallet_type_id = ?", CondValue: arrEwtSetup.ID},
			models.WhereCondFn{Condition: " ewt_detail.transaction_type LIKE ?", CondValue: "NFT"},
			models.WhereCondFn{Condition: " ewt_detail.total_in > ?", CondValue: 0},
		)
		arrEwtDetails, _ := models.GetEwtDetailFn(arrEwtDetailsFn, false)
		for _, arrEwtDetailsV := range arrEwtDetails {
			totalPurchased += arrEwtDetailsV.TotalIn
		}

		arrDataV["total_purchased"] = helpers.CutOffDecimal(totalPurchased, uint(4), ".", ",")
		arrDataV["total_purchased_currency_code"] = helpers.Translate("NFT", input.LangCode)
		arrDataV["total_pv"] = helpers.CutOffDecimal(arrSlsMaster.TotalAmount, uint(2), ".", ",")
		arrDataV["total_pv_currency_code"] = helpers.Translate("USDT", input.LangCode)

		if input.DocType == "" {
			arrData["NFT"] = arrDataV
		} else {
			arrData = arrDataV
		}
	}

	if input.DocType == "" || input.DocType == "STK" {
		// grab total purchase
		var arrDataV = map[string]interface{}{}

		// get member total stacked nft package value (nft)
		var arrSlsMasterFn = make([]models.WhereCondFn, 0)
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: input.MemberID},
			models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "STAKING"},
			models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
		)
		arrSlsMaster, err := models.GetMemberTotalSalesFn(arrSlsMasterFn, false)
		if err != nil {
			base.LogErrorLog("GetMemberSalesListSummary():GetMemberTotalSalesFn", arrSlsMaster, map[string]interface{}{"condition": arrSlsMasterFn}, true)
			return nil, "something_went_wrong"
		}

		arrDataV["total_stacked"] = helpers.CutOffDecimal(arrSlsMaster.TotalAmount, uint(4), ".", ",")
		arrDataV["total_stacked_currency_code"] = helpers.Translate("NFT", input.LangCode)

		if input.DocType == "" {
			arrData["STK"] = arrDataV
		} else {
			arrData = arrDataV
		}
	}

	return arrData, ""
}

// InsertSlsMasterBnsQueue func
func InsertSlsMasterBnsQueue(tx *gorm.DB, docNo string) string {
	arrSlsMasterDetails, err := models.GetSlsMasterDetailsByDocNo(docNo)
	if err != nil {
		return err.Error()
	}

	if arrSlsMasterDetails == nil {
		return "invalid_doc_no"
	}

	if arrSlsMasterDetails.Action == "CONTRACT" {
		addSlsMasterBnsQueue := models.AddSlsMasterBnsQueueStruct{
			DocNo:    docNo,
			BStatus:  "PENDING",
			DtCreate: time.Now(),
		}
		models.AddSlsMasterBnsQueue(tx, addSlsMasterBnsQueue)
	}

	return ""
}
