package wallet_service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/member_service"
)

var blockchainBatchLimitApi = 100

type WithdrawalCryptoSetup struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	// CryptoAddr    string  `json:"crypto_addr"`
	// GasFeePercent   float64 `json:"gas_fee_percent"`
	// AdminFeePercent float64 `json:"admin_fee_percent"`
}

type TransferSetup struct {
	EwtTypeCodeFrom string  `json:"ewallet_type_code_from"`
	EwtTypeNameFrom string  `json:"ewallet_type_name_from"`
	EwtTypeCodeTo   string  `json:"ewallet_type_code_to"`
	EwtTypeNameTo   string  `json:"ewallet_type_name_to"`
	Min             float64 `json:"min"`
	Max             float64 `json:"max"`
	// AdminFee float64 `json:"admin_fee"`
}

// type ExchangeSetup struct {
// 	Min             float64 `json:"min"`
// 	Price           float64 `json:"price"`
// 	EwalletTypeCode string  `json:"ewallet_type_code"`
// 	EwalletImgUrl   string  `json:"ewallet_img_url"`
// 	Balance         float64 `json:"balance"`
// }

type WalletBalanceSetup struct {
	// Asset               int `json:"asset"`
	// IncludeSpentBalance int `json:"include_spent_balance"`
	// WithdrawalCryptoStatus int `json:"withdrawal_crypto_status"` // To Crypto
	BlockchainDepositSetting interface{} `json:"blockchain_deposit"`
	// WalletTransactionSetting interface{} `json:"wallet_transaction_setting"`
	// SigningKeySetting        interface{} `json:"signing_key_setting"`
	// ShowCryptoAddr           int         `json:"show_crypto_addr"`
	// CryptoAddr               string      `json:"crypto_addr"`
	AppSettingList         interface{} `json:"app_setting_list"`
	TransferExchangeStatus int         `json:"transfer_exchange_status"`
	TransferStatus         int         `json:"transfer_status"`
	WithdrawStatus         int         `json:"withdraw_status"`
	ExchangeStatus         int         `json:"exchange_status"`
	RewardTypeList         interface{} `json:"reward_type_list"`
}

type WalletBalance struct {
	EwtTypeCode  string `json:"ewallet_type_code"`
	EwtTypeName  string `json:"ewallet_type_name"`
	EwtGroup     string `json:"ewallet_group"`
	CurrencyCode string `json:"currency_code"`
	// WalletAddress string `json:"wallet_address"`
	// Percentage                   float64            `json:"percentage"`
	Balance    float64 `json:"balance"`
	BalanceStr string  `json:"balance_str"`
	// LiveRate                     float64 `json:"live_rate"`
	// LiveRateStr                  string  `json:"live_rate_str"`
	ConvertedBalance    float64 `json:"converted_balance"`
	ConvertedBalanceStr string  `json:"converted_balance_str"`
	WithdrawBalance     float64 `json:"withdraw_balance"`
	WithdrawBalanceStr  string  `json:"withdraw_balance_str"`
	// AvailableBalance    float64 `json:"available_balance"`
	// AvailableBalanceStr string  `json:"available_balance_str"`
	// AvailableConvertedBalance    float64 `json:"available_converted_balance"`
	// AvailableConvertedBalanceStr string  `json:"available_converted_balance_str"`
	// HoldingBalance    float64 `json:"holding_balance"`
	// HoldingBalanceStr string  `json:"holding_balance_str"`
	// WithholdingBalance           float64            `json:"withholding_balance"`
	// WithholdingBalanceStr        string             `json:"withholding_balance_str"`
	DecimalPoint int                `json:"decimal_point"`
	Setup        WalletBalanceSetup `json:"setup"`
}

type WalletBalance2 struct {
	TotalBalance    float64     `json:"total_balance"`
	TotalBalanceStr string      `json:"total_balance_str"`
	CurrencyCode    string      `json:"currency_code"`
	Wallet          interface{} `json:"wallet_data"`
}

func GetMemberBalanceListv1(entMemberID int, ewtTypeCode string, langCode string) (interface{}, error) {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	)
	if ewtTypeCode != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: ewtTypeCode},
		)
	}
	result, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberBalanceListv1_GetMemberEwtSetupBalanceFn", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if len(result) > 0 {
		var decimalPoint uint
		if ewtTypeCode != "" {
			for _, v := range result {
				translatedWalletName := helpers.Translate(v.EwtTypeName, langCode)
				type arrBlockchainDepositSettingStruct struct {
					DispCryptoAddrStatus string `json:"disp_crypto_addr_status"`
					EwtTypeCode          string `json:"ewallet_type_code"`
					EwtTypeName          string `json:"ewallet_type_name"`
					// WalletTypeImageUrl   string `json:"wallet_type_image_url"`
				}
				var blockchainDepositSetting []arrBlockchainDepositSettingStruct
				if v.BlockchainDepositSetting != "" {
					json.Unmarshal([]byte(v.BlockchainDepositSetting), &blockchainDepositSetting)
					for k, v1 := range blockchainDepositSetting {
						v1.EwtTypeName = helpers.Translate(v1.EwtTypeName, langCode)
						blockchainDepositSetting[k] = v1
					}
				} else {
					blockchainDepositSetting = make([]arrBlockchainDepositSettingStruct, 0)
				}

				type arrWalletTransactionSettingListStruct struct {
					RwdCode string `json:"reward_code"`
					RwdName string `json:"reward_name"`
				}

				var walletTransactionSetting []arrWalletTransactionSettingListStruct
				if v.WalletTransactionSetting != "" {
					json.Unmarshal([]byte(v.WalletTransactionSetting), &walletTransactionSetting)

					for k, v1 := range walletTransactionSetting {
						v1.RwdName = helpers.Translate(v1.RwdName, langCode)
						walletTransactionSetting[k] = v1
					}

				} else {
					walletTransactionSetting = make([]arrWalletTransactionSettingListStruct, 0)
				}

				var appSettingList interface{}
				if v.AppSettingList != "" {
					json.Unmarshal([]byte(v.AppSettingList), &appSettingList)
				} else {
					appSettingList = nil
				}

				//check block all withdrawal
				status3 := member_service.VerifyIfInNetwork(entMemberID, "WD_BLK_ALL")

				if status3 {
					v.Withdraw = 0
				}

				arrTransferSetupCond := make([]models.WhereCondFn, 0)
				arrTransferSetupCond = append(arrTransferSetupCond,
					models.WhereCondFn{Condition: "ewt_transfer_setup.ewallet_type_id_from = ?", CondValue: v.ID},
					models.WhereCondFn{Condition: "ewt_transfer_setup.ewt_transfer_type = ?", CondValue: "Internal"},
					models.WhereCondFn{Condition: "ewt_transfer_setup.member_show = ?", CondValue: 1},
				)

				arrTransferSetupRst, _ := models.GetEwtTransferSetupFn(arrTransferSetupCond, "", false)
				transferStatus := 0
				if len(arrTransferSetupRst) > 0 {
					transferStatus = 1
				}

				bal := v.Balance
				convBal := v.Balance
				rate := float64(1) //default rate
				withdrawBal := float64(0)

				if v.Withdraw == 1 {
					withdrawBal = v.Balance

					if v.EwtTypeCode == "TP" || v.EwtTypeCode == "RPA" || v.EwtTypeCode == "RPB" {
						withdrawBal, _ = GetMemberWithdrawBalance(entMemberID, v.ID)
					}

				}

				convBal, _ = decimal.NewFromFloat(bal).Mul(decimal.NewFromFloat(rate)).Float64()
				// availConvBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()

				//get holding wallet
				// arrHoldCond := make([]models.WhereCondFn, 0)
				// arrHoldCond = append(arrHoldCond,
				// 	models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
				// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(v.EwtTypeCode) + "H"},
				// )

				// HoldResult, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrHoldCond, "", false)

				// if err != nil {
				// 	base.LogErrorLog("GetMemberBalanceListv1 -get Holding Wallet error", err, arrHoldCond, true)
				// }

				// if len(HoldResult) > 0 {
				// 	holdingBal = HoldResult[0].Balance
				// 	bal, _ = decimal.NewFromFloat(bal).Add(decimal.NewFromFloat(HoldResult[0].Balance)).Float64()
				// 	if bal < 0 {
				// 		bal = float64(0)
				// 	}
				// 	// convBal, _ = decimal.NewFromFloat(bal).Mul(decimal.NewFromFloat(rate)).Float64()
				// }
				// }

				// wal_addr, wal_addr_err := models.GetCustomMemberCryptoAddr(entMemberID, v.EwtTypeCode, true, false)

				// if wal_addr_err != nil {
				// 	wal_addr = ""
				// }

				// if v.EwtTypeCode == "USDT" {
				// 	wal_addr = ""
				// }

				decimalPoint = 2
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: v.EwtTypeCode},
				)
				ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
				if ewtSetup != nil {
					decimalPoint = uint(ewtSetup.DecimalPoint)
				}
				BalStr := helpers.CutOffDecimal(bal, decimalPoint, ".", ",")
				// RateStr := helpers.CutOffDecimal(rate, decimalPoint, ".", ",")
				ConvBalStr := helpers.CutOffDecimal(convBal, decimalPoint, ".", ",")
				// AvailBalStr := helpers.CutOffDecimal(availBal, decimalPoint, ".", ",")
				// AvailConvBalStr := helpers.CutOffDecimal(availConvBal, decimalPoint, ".", ",")
				// HoldingBalStr := helpers.CutOffDecimal(holdingBal, decimalPoint, ".", ",")
				// WithholdingBalStr := helpers.CutOffDecimal(withholdingBal, decimalPoint, ".", ",")
				WithdrawBalStr := helpers.CutOffDecimal(withdrawBal, decimalPoint, ".", ",")
				if v.EwtTypeCode == "CAP" {
					v.EwtTypeCode = "EC"
				}

				arrDataReturn := WalletBalance{
					EwtTypeCode:  v.EwtTypeCode,
					EwtTypeName:  translatedWalletName,
					EwtGroup:     v.EwtGroup,
					CurrencyCode: v.CurrencyCode,
					Balance:      bal,
					BalanceStr:   BalStr,
					// LiveRate:                     rate,
					// LiveRateStr:                  RateStr,
					ConvertedBalance:    convBal,
					ConvertedBalanceStr: ConvBalStr,
					WithdrawBalance:     withdrawBal,
					WithdrawBalanceStr:  WithdrawBalStr,
					// AvailableBalance:    availBal,
					// AvailableBalanceStr: AvailBalStr,
					// AvailableConvertedBalance:    availConvBal,
					// AvailableConvertedBalanceStr: AvailConvBalStr,
					// HoldingBalance:    holdingBal,
					// HoldingBalanceStr: HoldingBalStr,
					// WithholdingBalance:           withholdingBal,
					// WithholdingBalanceStr:        WithholdingBalStr,
					DecimalPoint: v.DecimalPoint,
					Setup: WalletBalanceSetup{
						BlockchainDepositSetting: blockchainDepositSetting,
						AppSettingList:           appSettingList,
						TransferExchangeStatus:   v.WithdrawalWithCrypto, //transfer exchange
						TransferStatus:           transferStatus,         // internal transfer
						WithdrawStatus:           v.Withdraw,
						ExchangeStatus:           v.Exchange,
						RewardTypeList:           walletTransactionSetting,
					},
				}

				decimalPoint = 2
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "USDT"},
				)
				ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
				if ewtSetup != nil {
					decimalPoint = uint(ewtSetup.DecimalPoint)
				}
				TotalBalStr := helpers.CutOffDecimal(bal, decimalPoint, ".", ",")

				arrDataReturn2 := WalletBalance2{
					TotalBalance:    bal,
					TotalBalanceStr: TotalBalStr,
					CurrencyCode:    helpers.Translate("USDT", langCode),
					Wallet:          arrDataReturn,
				}

				return arrDataReturn2, nil
			}
		} else {
			arrDataReturn := make([]WalletBalance, 0)
			totalBalance := float64(0)
			for _, v := range result {
				translatedWalletName := helpers.Translate(v.EwtTypeName, langCode)
				type arrBlockchainDepositSettingStruct struct {
					DispCryptoAddrStatus string `json:"disp_crypto_addr_status"`
					EwtTypeCode          string `json:"ewallet_type_code"`
					EwtTypeName          string `json:"ewallet_type_name"`
					// WalletTypeImageUrl   string `json:"wallet_type_image_url"`
				}
				var blockchainDepositSetting []arrBlockchainDepositSettingStruct
				if v.BlockchainDepositSetting != "" {
					json.Unmarshal([]byte(v.BlockchainDepositSetting), &blockchainDepositSetting)
					for k, v1 := range blockchainDepositSetting {
						v1.EwtTypeName = helpers.Translate(v1.EwtTypeName, langCode)
						blockchainDepositSetting[k] = v1
					}
				} else {
					blockchainDepositSetting = make([]arrBlockchainDepositSettingStruct, 0)
				}

				type arrWalletTransactionSettingListStruct struct {
					RwdCode string `json:"reward_code"`
					RwdName string `json:"reward_name"`
				}

				var walletTransactionSetting []arrWalletTransactionSettingListStruct
				if v.WalletTransactionSetting != "" {
					json.Unmarshal([]byte(v.WalletTransactionSetting), &walletTransactionSetting)

					for k, v1 := range walletTransactionSetting {
						v1.RwdName = helpers.Translate(v1.RwdName, langCode)
						walletTransactionSetting[k] = v1
					}

				} else {
					walletTransactionSetting = make([]arrWalletTransactionSettingListStruct, 0)
				}

				var appSettingList interface{}
				if v.AppSettingList != "" {
					json.Unmarshal([]byte(v.AppSettingList), &appSettingList)
				} else {
					appSettingList = nil
				}

				//check block all withdrawal
				status3 := member_service.VerifyIfInNetwork(entMemberID, "WD_BLK_ALL")

				if status3 {
					v.Withdraw = 0
				}

				arrTransferSetupCond := make([]models.WhereCondFn, 0)
				arrTransferSetupCond = append(arrTransferSetupCond,
					models.WhereCondFn{Condition: "ewt_transfer_setup.ewallet_type_id_from = ?", CondValue: v.ID},
					models.WhereCondFn{Condition: "ewt_transfer_setup.ewt_transfer_type = ?", CondValue: "Internal"},
					models.WhereCondFn{Condition: "ewt_transfer_setup.member_show = ?", CondValue: 1},
				)

				arrTransferSetupRst, _ := models.GetEwtTransferSetupFn(arrTransferSetupCond, "", false)
				transferStatus := 0
				if len(arrTransferSetupRst) > 0 {
					transferStatus = 1
				}

				bal := v.Balance
				convBal := v.Balance
				rate := float64(1) // default rate
				withdrawBal := float64(0)

				//get withdraw balance

				if v.Withdraw == 1 {
					withdrawBal = v.Balance

					if v.EwtTypeCode == "TP" || v.EwtTypeCode == "RPA" || v.EwtTypeCode == "RPB" {
						withdrawBal, _ = GetMemberWithdrawBalance(entMemberID, v.ID)
					}

				}
				// availBal := v.Balance
				// availConvBal := v.Balance
				// withholdingBal := float64(0)
				// holdingBal := float64(0)
				// if v.Control == "BLOCKCHAIN" {
				// for _, v2 := range allBlkCWalBal {
				// 	if v.EwtTypeCode == v2.WalletTypeCode {
				// 		bal = v2.Balance
				// 		convBal = v2.ConvertedBalance
				// 		rate = v2.Rate
				// 		availBal = v2.AvailableBalance
				// 		availConvBal = v2.ConvertedAvailableBalance
				// 		withholdingBal = v2.WithHoldingBalance
				// 	}
				// }

				// } else {
				//get rate
				// rate, err = base.GetLatestPriceMovementByTokenType(v.EwtTypeCode, entMemberID, 0)
				// if err != nil {
				// 	base.LogErrorLog("GetMemberBalanceListv1 -GetLatestPriceMovementByTokenType error", err, map[string]interface{}{"wallet_type_code": v.EwtTypeCode}, true)
				// }

				convBal, _ = decimal.NewFromFloat(bal).Mul(decimal.NewFromFloat(rate)).Float64()
				// availConvBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()

				//get holding wallet
				// arrHoldCond := make([]models.WhereCondFn, 0)
				// arrHoldCond = append(arrHoldCond,
				// 	models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
				// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(v.EwtTypeCode) + "H"},
				// )

				// HoldResult, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrHoldCond, "", false)

				// if err != nil {
				// 	base.LogErrorLog("GetMemberBalanceListv1 -get Holding Wallet error", err, arrHoldCond, true)
				// }

				// if len(HoldResult) > 0 {
				// 	holdingBal = HoldResult[0].Balance
				// 	bal, _ = decimal.NewFromFloat(bal).Add(decimal.NewFromFloat(HoldResult[0].Balance)).Float64()
				// 	if bal < 0 {
				// 		bal = float64(0)
				// 	}
				// 	// convBal, _ = decimal.NewFromFloat(bal).Mul(decimal.NewFromFloat(rate)).Float64()
				// }
				// }

				totalBalance, _ = decimal.NewFromFloat(totalBalance).Add(decimal.NewFromFloat(convBal)).Float64()

				// wal_addr, wal_addr_err := models.GetCustomMemberCryptoAddr(entMemberID, v.EwtTypeCode, true, false)

				// if wal_addr_err != nil {
				// 	wal_addr = ""
				// }

				decimalPoint = 2
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: v.EwtTypeCode},
				)
				ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
				if ewtSetup != nil {
					decimalPoint = uint(ewtSetup.DecimalPoint)
				}
				BalStr := helpers.CutOffDecimal(bal, decimalPoint, ".", ",")
				// RateStr := helpers.CutOffDecimal(rate, decimalPoint, ".", ",")
				ConvBalStr := helpers.CutOffDecimal(convBal, decimalPoint, ".", ",")
				// AvailBalStr := helpers.CutOffDecimal(availBal, decimalPoint, ".", ",")
				// AvailConvBalStr := helpers.CutOffDecimal(availConvBal, decimalPoint, ".", ",")
				// HoldingBalStr := helpers.CutOffDecimal(holdingBal, decimalPoint, ".", ",")
				WithdrawBalStr := helpers.CutOffDecimal(withdrawBal, decimalPoint, ".", ",")

				// WithholdingBalStr := helpers.CutOffDecimal(withholdingBal, decimalPoint, ".", ",")

				if v.EwtTypeCode == "CAP" {
					v.EwtTypeCode = "EC"
				}

				if v.ShowAmt == 1 {
					//check blockchain_trans && ewt_details

					// arrEwtDetCond := make([]models.WhereCondFn, 0)
					// arrEwtDetCond = append(arrEwtDetCond,
					// 	models.WhereCondFn{Condition: " ewt_detail.member_id = ? ", CondValue: entMemberID},
					// 	models.WhereCondFn{Condition: " ewt_detail.ewallet_type_id = ?", CondValue: v.ID},
					// )
					// arrEwtDetail, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

					// arrBlockTransCond := make([]models.WhereCondFn, 0)
					// arrBlockTransCond = append(arrBlockTransCond,
					// 	models.WhereCondFn{Condition: " blockchain_trans.member_id = ? ", CondValue: entMemberID},
					// 	models.WhereCondFn{Condition: " blockchain_trans.ewallet_type_id = ?", CondValue: v.ID},
					// )
					// arrBlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockTransCond, false)

					// if len(arrEwtDetail) > 0 || len(arrBlockchainTrans) > 0 {
					if bal > 0 {
						arrDataReturn = append(arrDataReturn, WalletBalance{
							EwtTypeCode:  v.EwtTypeCode,
							EwtTypeName:  translatedWalletName,
							EwtGroup:     v.EwtGroup,
							CurrencyCode: v.CurrencyCode,
							// WalletAddress: wal_addr,
							Balance:    bal,
							BalanceStr: BalStr,
							// LiveRate:                     rate,
							// LiveRateStr:                  RateStr,
							ConvertedBalance:    convBal,
							ConvertedBalanceStr: ConvBalStr,
							WithdrawBalance:     withdrawBal,
							WithdrawBalanceStr:  WithdrawBalStr,
							// AvailableBalance:    availBal,
							// AvailableBalanceStr: AvailBalStr,
							// AvailableConvertedBalance:    availConvBal,
							// AvailableConvertedBalanceStr: AvailConvBalStr,
							// HoldingBalance:    holdingBal,
							// HoldingBalanceStr: HoldingBalStr,
							// WithholdingBalance:           withholdingBal,
							// WithholdingBalanceStr:        WithholdingBalStr,
							DecimalPoint: v.DecimalPoint,
							Setup: WalletBalanceSetup{
								BlockchainDepositSetting: blockchainDepositSetting,
								AppSettingList:           appSettingList,
								TransferExchangeStatus:   v.WithdrawalWithCrypto, //transfer exchange
								TransferStatus:           transferStatus,         // internal transfer
								WithdrawStatus:           v.Withdraw,             //withdraw (80%+20%)
								ExchangeStatus:           v.Exchange,
								RewardTypeList:           walletTransactionSetting,
							},
						},
						)
					}
				} else {
					arrDataReturn = append(arrDataReturn, WalletBalance{
						EwtTypeCode:  v.EwtTypeCode,
						EwtTypeName:  translatedWalletName,
						EwtGroup:     v.EwtGroup,
						CurrencyCode: v.CurrencyCode,
						// WalletAddress: wal_addr,
						Balance:    bal,
						BalanceStr: BalStr,
						// LiveRate:                     rate,
						// LiveRateStr:                  RateStr,
						ConvertedBalance:    convBal,
						ConvertedBalanceStr: ConvBalStr,
						WithdrawBalance:     withdrawBal,
						WithdrawBalanceStr:  WithdrawBalStr,
						// AvailableBalance:    availBal,
						// AvailableBalanceStr: AvailBalStr,
						// AvailableConvertedBalance:    availConvBal,
						// AvailableConvertedBalanceStr: AvailConvBalStr,
						// HoldingBalance:    holdingBal,
						// HoldingBalanceStr: HoldingBalStr,
						// WithholdingBalance:           withholdingBal,
						// WithholdingBalanceStr:        WithholdingBalStr,
						DecimalPoint: v.DecimalPoint,
						Setup: WalletBalanceSetup{
							BlockchainDepositSetting: blockchainDepositSetting,
							AppSettingList:           appSettingList,
							TransferExchangeStatus:   v.WithdrawalWithCrypto, //transfer exchange
							TransferStatus:           transferStatus,         // internal transfer
							WithdrawStatus:           v.Withdraw,
							ExchangeStatus:           v.Exchange,
							RewardTypeList:           walletTransactionSetting,
						},
					},
					)
				}
			}

			// perc := float64(0)
			// //to calculate wallet percentage based on total balance (btc and eth ... will in 0)
			// for k, v := range arrDataReturn {
			// 	perc = (math.Round(v.ConvertedBalance*1000) / 1000) / (math.Round(totalBalance*1000) / 1000) * 100
			// 	v.Percentage = float64(0)
			// 	if perc > 0 {
			// 		v.Percentage = math.Round(perc*1000) / 1000
			// 	}
			// 	arrDataReturn[k] = v
			// }

			decimalPoint = 2
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "USDT"},
			)
			ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
			if ewtSetup != nil {
				decimalPoint = uint(ewtSetup.DecimalPoint)
			}
			TotalBalStr := helpers.CutOffDecimal(totalBalance, decimalPoint, ".", ",")

			arrDataReturn2 := WalletBalance2{
				TotalBalance:    totalBalance,
				TotalBalanceStr: TotalBalStr,
				CurrencyCode:    helpers.Translate("USDT", langCode),
				Wallet:          arrDataReturn,
			}

			return arrDataReturn2, nil
		}
	}
	return nil, nil
}

type WalletBalanceStruct struct {
	EntMemberID int     `json:"ent_member_id"`
	EwtTypeID   int     `json:"ewallet_type_id"`
	EwtTypeCode string  `json:"ewallet_type_code"`
	EwtTypeName string  `json:"ewallet_type_name"`
	Balance     float64 `json:"balance"`
}

type GetWalletBalanceStruct struct {
	Tx          *gorm.DB
	EntMemberID int
	EwtTypeID   int
	EwtTypeCode string
}

func GetWalletBalance(arrData GetWalletBalanceStruct) WalletBalanceStruct {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: arrData.EntMemberID},
	)

	if arrData.EwtTypeID > 0 {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrData.EwtTypeID},
		)
	}

	if arrData.EwtTypeCode != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.EwtTypeCode},
		)
	}

	result, _ := models.GetEwtSummarySetupFn(arrCond, "", false)

	arrDataReturn := WalletBalanceStruct{
		EntMemberID: arrData.EntMemberID,
		EwtTypeID:   arrData.EwtTypeID,
		EwtTypeCode: "",
		EwtTypeName: "",
		Balance:     0,
	}

	if result != nil {
		if len(result) > 0 {
			arrDataReturn = WalletBalanceStruct{
				EntMemberID: arrData.EntMemberID,
				EwtTypeID:   arrData.EwtTypeID,
				EwtTypeCode: result[0].EwalletTypeCode,
				EwtTypeName: result[0].EwalletTypeName,
				Balance:     result[0].Balance,
			}
		}
	}

	return arrDataReturn
}

func GetWalletBalanceTx(arrData GetWalletBalanceStruct) WalletBalanceStruct {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: arrData.EntMemberID},
	)

	if arrData.EwtTypeID > 0 {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrData.EwtTypeID},
		)
	}

	if arrData.EwtTypeCode != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.EwtTypeCode},
		)
	}

	result, _ := models.GetEwtSummarySetupFnTx(arrData.Tx, arrCond, "", false)

	arrDataReturn := WalletBalanceStruct{
		EntMemberID: arrData.EntMemberID,
		EwtTypeID:   arrData.EwtTypeID,
		EwtTypeCode: "",
		EwtTypeName: "",
		Balance:     0,
	}

	if result != nil {
		if len(result) > 0 {
			arrDataReturn = WalletBalanceStruct{
				EntMemberID: arrData.EntMemberID,
				EwtTypeID:   arrData.EwtTypeID,
				EwtTypeCode: result[0].EwalletTypeCode,
				EwtTypeName: result[0].EwalletTypeName,
				Balance:     result[0].Balance,
			}
		}
	}

	return arrDataReturn
}

type SaveMemberWalletStruct struct {
	EntMemberID       int
	EwalletTypeID     int
	EwalletTypeCode   string
	CurrencyCode      string
	DecimalPlaces     int
	AllowNegative     int
	TotalIn           float64
	TotalOut          float64
	ConversionRate    float64
	ConvertedTotalIn  float64
	ConvertedTotalOut float64
	TransactionType   string
	DocNo             string
	AdditionalMsg     string
	Remark            string
	CreatedBy         string
}

func SaveMemberWallet(tx *gorm.DB, arrData SaveMemberWalletStruct) (int, error) {

	/* start example data. can refer type SaveMemberWalletStruct
	ewtIn := wallet_service.SaveMemberWalletStruct{
		EntMemberID:     6,
		EwalletTypeID:   form.WalletTypeID,
		TotalIn:         form.TotalIn,(float64) // if this is pass, TotalOut must empty / 0
		TotalOut:        form.TotalOut,(float64) // if this is pass, TotalIn must empty / 0
		TransactionType: "ADJUSTMENT",
		DocNo:           "ADJ00000001",
		Remark:          "ADJUSMENT FOR TESTING",
		CreatedBy:       "6",
	}
	start example data */
	if arrData.EntMemberID == 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_EntMemberID", Data: arrData}
	}
	if arrData.EwalletTypeID == 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_EwalletTypeID", Data: arrData}
	}

	flagAmount := false

	if arrData.TotalIn > 0 || arrData.TotalOut > 0 {
		flagAmount = true
	}
	if arrData.TotalIn > 0 && arrData.TotalOut > 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_either_total_in_or_total_out", Data: arrData}
	}

	createdBy := "AUTO"
	if arrData.CreatedBy != "" {
		createdBy = arrData.CreatedBy
	}

	if !flagAmount {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_total_in_or_total_out_to_procecss", Data: arrData}
	}

	if arrData.TotalIn > 0 {
		totalInString := helpers.CutOffDecimal(arrData.TotalIn, 8, ".", "")
		totalInFloat, err := strconv.ParseFloat(totalInString, 64)
		if err != nil {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
		arrData.TotalIn = totalInFloat
		// arrData.TotalIn = float.RoundDown(arrData.TotalIn, 8) // set only allow 8 decimal point to be process and save
	}
	if arrData.TotalOut > 0 {
		totalOutString := helpers.CutOffDecimal(arrData.TotalOut, 8, ".", "")
		totalOutFloat, err := strconv.ParseFloat(totalOutString, 64)
		if err != nil {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
		arrData.TotalOut = totalOutFloat
		// arrData.TotalOut = float.RoundDown(arrData.TotalOut, 8) // set only allow 8 decimal point to be process and save
	}

	arrEwtBal := GetWalletBalanceStruct{
		EntMemberID: arrData.EntMemberID,
		EwtTypeID:   arrData.EwalletTypeID,
	}
	walletBalance := GetWalletBalance(arrEwtBal)
	if arrData.TotalOut > 0 {
		if arrData.AllowNegative != 1 && arrData.TotalOut > walletBalance.Balance {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct", Data: arrData}
		}
	}

	curDateTime := time.Now()
	latestbal := walletBalance.Balance + arrData.TotalIn - arrData.TotalOut

	arrEwtDetail := models.EwtDetail{
		MemberID:          arrData.EntMemberID,
		EwalletTypeID:     arrData.EwalletTypeID,
		TransactionType:   arrData.TransactionType,
		TransDate:         curDateTime,
		TotalIn:           arrData.TotalIn,
		TotalOut:          arrData.TotalOut,
		ConversionRate:    arrData.ConversionRate,
		ConvertedTotalIn:  arrData.ConvertedTotalIn,
		ConvertedTotalOut: arrData.ConvertedTotalOut,
		Balance:           latestbal,
		DocNo:             arrData.DocNo,
		AdditionalMsg:     arrData.AdditionalMsg,
		Remark:            arrData.Remark,
		CreatedBy:         createdBy,
	}

	AddEwtDetailRst, err := models.AddEwtDetail(tx, arrEwtDetail)

	if err != nil {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}

	arrExisingWalletCond := make([]models.WhereCondFn, 0)
	arrExisingWalletCond = append(arrExisingWalletCond,
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrData.EwalletTypeID},
	)

	arrExisingWallet, err := models.GetEwtSummaryFn(arrExisingWalletCond, "", false)
	if err != nil {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}

	if len(arrExisingWallet) > 0 {
		updateColumn := map[string]interface{}{"balance": latestbal, "updated_by": createdBy}
		if arrData.TotalIn > 0 {
			// updateColumn["total_in"] = float.Add(arrData.TotalIn, arrExisingWallet[0].TotalIn) // this will cut down decimal. can not use this. tested
			updateColumn["total_in"] = arrData.TotalIn + arrExisingWallet[0].TotalIn
		}
		if arrData.TotalOut > 0 {
			// updateColumn["total_out"] = float.Add(arrData.TotalOut, arrExisingWallet[0].TotalOut) // this will cut down decimal. can not use this. tested
			updateColumn["total_out"] = arrData.TotalOut + arrExisingWallet[0].TotalOut
		}

		err := models.UpdatesFnTx(tx, "ewt_summary", arrExisingWalletCond, updateColumn, false)

		if err != nil {
			return 0, err
		}

	} else {
		arrEwtSummary := models.AddEwtSummaryStruct{
			MemberID:      arrData.EntMemberID,
			EwalletTypeID: arrData.EwalletTypeID,
			TotalIn:       arrData.TotalIn,
			TotalOut:      arrData.TotalOut,
			Balance:       latestbal,
			CreatedBy:     "AUTO",
		}
		_, err := models.AddEwtSummary(tx, arrEwtSummary)
		if err != nil {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
	}

	return AddEwtDetailRst.ID, nil
}

func SaveMemberWalletTx(tx *gorm.DB, arrData SaveMemberWalletStruct) (int, error) {

	/* start example data. can refer type SaveMemberWalletStruct
	ewtIn := wallet_service.SaveMemberWalletStruct{
		EntMemberID:     6,
		EwalletTypeID:   form.WalletTypeID,
		TotalIn:         form.TotalIn,(float64) // if this is pass, TotalOut must empty / 0
		TotalOut:        form.TotalOut,(float64) // if this is pass, TotalIn must empty / 0
		TransactionType: "ADJUSTMENT",
		DocNo:           "ADJ00000001",
		Remark:          "ADJUSMENT FOR TESTING",
		CreatedBy:       "6",
	}
	start example data */
	if arrData.EntMemberID == 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_EntMemberID", Data: arrData}
	}
	if arrData.EwalletTypeID == 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_EwalletTypeID", Data: arrData}
	}

	flagAmount := false

	if arrData.TotalIn > 0 || arrData.TotalOut > 0 {
		flagAmount = true
	}
	if arrData.TotalIn > 0 && arrData.TotalOut > 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_either_total_in_or_total_out", Data: arrData}
	}

	createdBy := "AUTO"
	if arrData.CreatedBy != "" {
		createdBy = arrData.CreatedBy
	}

	if !flagAmount {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_total_in_or_total_out_to_procecss", Data: arrData}
	}

	if arrData.TotalIn > 0 {
		totalInString := helpers.CutOffDecimal(arrData.TotalIn, 8, ".", "")
		totalInFloat, err := strconv.ParseFloat(totalInString, 64)
		if err != nil {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
		arrData.TotalIn = totalInFloat
		// arrData.TotalIn = float.RoundDown(arrData.TotalIn, 8) // set only allow 8 decimal point to be process and save
	}
	if arrData.TotalOut > 0 {
		totalOutString := helpers.CutOffDecimal(arrData.TotalOut, 8, ".", "")
		totalOutFloat, err := strconv.ParseFloat(totalOutString, 64)
		if err != nil {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
		arrData.TotalOut = totalOutFloat
		// arrData.TotalOut = float.RoundDown(arrData.TotalOut, 8) // set only allow 8 decimal point to be process and save
	}

	arrEwtBal := GetWalletBalanceStruct{
		Tx:          tx,
		EntMemberID: arrData.EntMemberID,
		EwtTypeID:   arrData.EwalletTypeID,
	}
	walletBalance := GetWalletBalanceTx(arrEwtBal)
	if arrData.TotalOut > 0 {
		if arrData.AllowNegative != 1 && arrData.TotalOut > walletBalance.Balance {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct", Data: arrData}
		}
	}

	curDateTime := time.Now()
	latestbal := walletBalance.Balance + arrData.TotalIn - arrData.TotalOut

	arrEwtDetail := models.EwtDetail{
		MemberID:          arrData.EntMemberID,
		EwalletTypeID:     arrData.EwalletTypeID,
		TransactionType:   arrData.TransactionType,
		TransDate:         curDateTime,
		TotalIn:           arrData.TotalIn,
		TotalOut:          arrData.TotalOut,
		ConversionRate:    arrData.ConversionRate,
		ConvertedTotalIn:  arrData.ConvertedTotalIn,
		ConvertedTotalOut: arrData.ConvertedTotalOut,
		Balance:           latestbal,
		DocNo:             arrData.DocNo,
		AdditionalMsg:     arrData.AdditionalMsg,
		Remark:            arrData.Remark,
		CreatedBy:         createdBy,
	}

	AddEwtDetailRst, err := models.AddEwtDetail(tx, arrEwtDetail)

	if err != nil {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}

	arrExisingWalletCond := make([]models.WhereCondFn, 0)
	arrExisingWalletCond = append(arrExisingWalletCond,
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrData.EwalletTypeID},
	)

	arrExisingWallet, err := models.GetEwtSummaryFnTx(tx, arrExisingWalletCond, "", false)
	if err != nil {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}

	if len(arrExisingWallet) > 0 {
		updateColumn := map[string]interface{}{"balance": latestbal, "updated_by": createdBy}
		if arrData.TotalIn > 0 {
			// updateColumn["total_in"] = float.Add(arrData.TotalIn, arrExisingWallet[0].TotalIn) // this will cut down decimal. can not use this. tested
			updateColumn["total_in"] = arrData.TotalIn + arrExisingWallet[0].TotalIn
		}
		if arrData.TotalOut > 0 {
			// updateColumn["total_out"] = float.Add(arrData.TotalOut, arrExisingWallet[0].TotalOut) // this will cut down decimal. can not use this. tested
			updateColumn["total_out"] = arrData.TotalOut + arrExisingWallet[0].TotalOut
		}

		err := models.UpdatesFnTx(tx, "ewt_summary", arrExisingWalletCond, updateColumn, false)

		if err != nil {
			return 0, err
		}

	} else {
		arrEwtSummary := models.AddEwtSummaryStruct{
			MemberID:      arrData.EntMemberID,
			EwalletTypeID: arrData.EwalletTypeID,
			TotalIn:       arrData.TotalIn,
			TotalOut:      arrData.TotalOut,
			Balance:       latestbal,
			CreatedBy:     "AUTO",
		}
		_, err := models.AddEwtSummary(tx, arrEwtSummary)
		if err != nil {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
	}

	return AddEwtDetailRst.ID, nil
}

func SaveMemberWalletWithoutTx(arrData SaveMemberWalletStruct) (int, error) {

	/* start example data. can refer type SaveMemberWalletStruct
	ewtIn := wallet_service.SaveMemberWalletStruct{
		EntMemberID:     6,
		EwalletTypeID:   form.WalletTypeID,
		TotalIn:         form.TotalIn,(float64) // if this is pass, TotalOut must empty / 0
		TotalOut:        form.TotalOut,(float64) // if this is pass, TotalIn must empty / 0
		TransactionType: "ADJUSTMENT",
		DocNo:           "ADJ00000001",
		Remark:          "ADJUSMENT FOR TESTING",
		CreatedBy:       "6",
	}
	start example data */
	if arrData.EntMemberID == 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_EntMemberID", Data: arrData}
	}
	if arrData.EwalletTypeID == 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_EwalletTypeID", Data: arrData}
	}

	flagAmount := false

	if arrData.TotalIn > 0 || arrData.TotalOut > 0 {
		flagAmount = true
	}
	if arrData.TotalIn > 0 && arrData.TotalOut > 0 {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_either_total_in_or_total_out", Data: arrData}
	}

	createdBy := "AUTO"
	if arrData.CreatedBy != "" {
		createdBy = arrData.CreatedBy
	}

	if !flagAmount {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "please_enter_total_in_or_total_out_to_procecss", Data: arrData}
	}

	if arrData.TotalIn > 0 {
		arrData.TotalIn = float.RoundDown(arrData.TotalIn, 8) // set only allow 8 decimal point to be process and save
	}
	if arrData.TotalOut > 0 {
		arrData.TotalOut = float.RoundDown(arrData.TotalOut, 8) // set only allow 8 decimal point to be process and save
	}

	arrEwtBal := GetWalletBalanceStruct{
		EntMemberID: arrData.EntMemberID,
		EwtTypeID:   arrData.EwalletTypeID,
	}
	walletBalance := GetWalletBalance(arrEwtBal)

	if arrData.TotalOut > walletBalance.Balance {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "not_enu_bal_deduct", Data: arrData}
	}

	curDateTime := time.Now()
	latestbal := walletBalance.Balance + arrData.TotalIn - arrData.TotalOut

	arrEwtDetail := models.EwtDetail{
		MemberID:        arrData.EntMemberID,
		EwalletTypeID:   arrData.EwalletTypeID,
		TransactionType: arrData.TransactionType,
		TransDate:       curDateTime,
		TotalIn:         arrData.TotalIn,
		TotalOut:        arrData.TotalOut,
		Balance:         latestbal,
		DocNo:           arrData.DocNo,
		AdditionalMsg:   arrData.AdditionalMsg,
		Remark:          arrData.Remark,
		CreatedBy:       createdBy,
	}

	AddEwtDetailRst, err := models.AddEwtDetailWithoutTx(arrEwtDetail)

	if err != nil {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}

	arrExisingWalletCond := make([]models.WhereCondFn, 0)
	arrExisingWalletCond = append(arrExisingWalletCond,
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrData.EwalletTypeID},
	)

	arrExisingWallet, err := models.GetEwtSummaryFn(arrExisingWalletCond, "", false)
	if err != nil {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	}

	if len(arrExisingWallet) > 0 {
		updateColumn := map[string]interface{}{"balance": latestbal, "updated_by": createdBy}
		if arrData.TotalIn > 0 {
			// updateColumn["total_in"] = float.Add(arrData.TotalIn, arrExisingWallet[0].TotalIn) // this will cut down decimal. can not use this. tested
			updateColumn["total_in"] = arrData.TotalIn + arrExisingWallet[0].TotalIn
		}
		if arrData.TotalOut > 0 {
			// updateColumn["total_out"] = float.Add(arrData.TotalOut, arrExisingWallet[0].TotalOut) // this will cut down decimal. can not use this. tested
			updateColumn["total_out"] = arrData.TotalOut + arrExisingWallet[0].TotalOut
		}

		err := models.UpdatesFn("ewt_summary", arrExisingWalletCond, updateColumn, false)

		if err != nil {
			models.ErrorLog("SaveMemberWallet-Update ewt_summary balance", arrExisingWalletCond, updateColumn)
			return 0, err
		}

	} else {
		arrEwtSummary := models.AddEwtSummaryStruct{
			MemberID:      arrData.EntMemberID,
			EwalletTypeID: arrData.EwalletTypeID,
			TotalIn:       arrData.TotalIn,
			TotalOut:      arrData.TotalOut,
			Balance:       latestbal,
			CreatedBy:     "AUTO",
		}
		_, err := models.AddEwtSummaryWithoutTx(arrEwtSummary)
		if err != nil {
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
		}
	}

	return AddEwtDetailRst.ID, nil
}

type PostTransferStruct struct {
	MemberId      int
	EwtTypeCode   string
	EwtTypeCodeTo string
	Amount        float64
	MemberTo      string
	Remark        string
	LangCode      string
}

func (t *PostTransferStruct) PostTransfer(tx *gorm.DB) (interface{}, error) {

	var (
		err error
	)

	//check member wallet balance - save member wallet function will handle

	//get wallet setting
	arrWalCond := make([]models.WhereCondFn, 0)
	arrWalCond = append(arrWalCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: t.EwtTypeCode},
	)
	walSetup, err := models.GetEwtSetupFn(arrWalCond, "", false)

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to get wallet setup", err, arrWalCond, true)
		return nil, err
	}

	if walSetup == nil {
		base.LogErrorLog("PostTransfer - empty wallet setup returned", walSetup, arrWalCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("empty_wallet_setup", t.LangCode), Data: t}
	}

	eWalletId := walSetup.ID

	//get wallet to setting
	arrWalToCond := make([]models.WhereCondFn, 0)
	arrWalToCond = append(arrWalToCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: t.EwtTypeCodeTo},
	)
	walSetupTo, err := models.GetEwtSetupFn(arrWalToCond, "", false)

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to get wallet to setup", err, arrWalToCond, true)
		return nil, err
	}

	if walSetupTo == nil {
		base.LogErrorLog("PostTransfer - empty wallet to setup returned", walSetupTo, arrWalToCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("empty_wallet_to_setup", t.LangCode), Data: t}
	}

	eWalletIdTo := walSetupTo.ID

	//get member info
	arrMemInfoCond := make([]models.WhereCondFn, 0)
	arrMemInfoCond = append(arrMemInfoCond,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: t.MemberId},
	)
	memberInfo, err := models.GetEntMemberFn(arrMemInfoCond, "", false) //get member to details

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to get member info", err, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}

	if memberInfo == nil {
		base.LogErrorLog("PostTransfer - invalid member info", err, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_member_info", t.LangCode), Data: t}
	}

	//get member to info
	arrMemToCond := make([]models.WhereCondFn, 0)
	arrMemToCond = append(arrMemToCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: t.MemberTo},
	)
	memberTo, err := models.GetEntMemberFn(arrMemToCond, "", false) //get member to details

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to get member to info", err, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_member_to", t.LangCode), Data: t}
	}

	if memberTo == nil {
		// base.LogErrorLog("PostTransfer - empty member to info", err, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_member_to", t.LangCode), Data: t}
	}

	memToId := memberTo.ID

	//check amt
	if t.Amount <= 0 {
		// base.LogErrorLog("PostTransfer - amount cannot negative", "t.Amount", t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("amount_must_more_than_0", t.LangCode), Data: t}
	}

	//check setting
	arrTransferSetupCond := make([]models.WhereCondFn, 0)
	arrTransferSetupCond = append(arrTransferSetupCond,
		models.WhereCondFn{Condition: "ewt_transfer_setup.ewallet_type_id_from = ?", CondValue: eWalletId},
		models.WhereCondFn{Condition: "ewt_transfer_setup.transfer_type_id_to = ?", CondValue: eWalletIdTo},
	)

	arrTransferSetupRst, err := models.GetEwtTransferSetupFn(arrTransferSetupCond, "", false)

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to get transfer setup", err, arrTransferSetupCond, true)
		return nil, err
	}

	if len(arrTransferSetupRst) < 0 {
		base.LogErrorLog("PostTransfer - empty transfer setup", arrTransferSetupRst, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}

	adminFee := float64(0)
	for _, transfSettingVal := range arrTransferSetupRst {
		if t.Amount < transfSettingVal.TransferMin {
			strAmt := helpers.CutOffDecimal(transfSettingVal.TransferMin, 2, ".", "")
			// base.LogErrorLog("PostTransfer - minimum transfer amount is"+" "+strAmt, "transferSetup-TransferMin"+" "+":"+strAmt, t, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("minimum_transfer_amount_is"+" "+strAmt, t.LangCode), Data: t}
		}

		//check transfer from wallet
		if transfSettingVal.EwalletTypeIdFrom != eWalletId {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("cannot_transfer_from_this_wallet", t.LangCode), Data: t}
		}

		//check transfer to wallet
		if transfSettingVal.EwalletTypeIdTo != eWalletIdTo {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("cannot_transfer_to_this_wallet", t.LangCode), Data: t}
		}

		//check whether is to same member
		if transfSettingVal.TransferSameMember == 1 {
			if memToId != t.MemberId {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("cannot_transfer_to_other_account", t.LangCode), Data: t}
			}
		} else {
			//check if is same member
			if memToId == t.MemberId {
				// base.LogErrorLog("PostTransfer - cannot transfer to same account", memToId, t.MemberId, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("cannot_transfer_to_same_account", t.LangCode), Data: t}
			}
		}

		//check within newtwork
		if transfSettingVal.TransferSponsorTree == 1 {
			memberNetw1 := member_service.CheckSponsorMember(t.MemberId, memToId) // koo func- sponsor_id, downline_id
			memberNetw2 := member_service.CheckSponsorMember(memToId, t.MemberId) // koo func- sponsor_id, downline_id

			if memberNetw1 == false && memberNetw2 == false {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_within_network", t.LangCode), Data: t}
			}

		}

		adminFee = float64(transfSettingVal.AdminFee)

	}

	adminFeeAmount, _ := decimal.NewFromFloat(t.Amount).Mul(decimal.NewFromFloat(adminFee)).Float64()
	adminFeeAmount, _ = decimal.NewFromFloat(adminFeeAmount).Div(decimal.NewFromFloat(100)).Float64()
	transferAmount, _ := decimal.NewFromFloat(t.Amount).Sub(decimal.NewFromFloat(adminFeeAmount)).Float64()

	//check wallet lock
	wallet_lock, err := models.GetEwtLockByMemberId(t.MemberId, eWalletId)

	if wallet_lock.InternalTransfer == 1 {
		// base.LogErrorLog("PostTransfer - wallet is locked from perform transfer", wallet_lock, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("wallet_is_locked_from_performing_transfer", t.LangCode), Data: t}
	}

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to check wallet lock", err, wallet_lock, true)
		return nil, err
	}

	docs, err := models.GetRunningDocNo("WT", tx) //get transfer doc no

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to get WT doc no", err, t, true)
		return nil, err
	}

	//deduct balance
	ewtOut := SaveMemberWalletStruct{
		EntMemberID:     t.MemberId,
		EwalletTypeID:   eWalletId,
		TotalOut:        t.Amount,
		TransactionType: "TRANSFER",
		DocNo:           docs,
		Remark:          "#*transfer_to*#" + " " + memberTo.NickName,
		// CreatedBy:       strconv.Itoa(t.MemberId),
	}

	_, err = SaveMemberWallet(tx, ewtOut)

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to save ewt out", err, ewtOut, true)
		return nil, err
	}

	ewtIn := SaveMemberWalletStruct{
		EntMemberID:   memToId,
		EwalletTypeID: eWalletIdTo,
		TotalIn:       transferAmount,
		// TotalIn:         t.Amount,
		TransactionType: "TRANSFER",
		DocNo:           docs,
		Remark:          "#*transfer_from*#" + " " + memberInfo.NickName,
		// CreatedBy:       strconv.Itoa(t.MemberId),
	}

	_, err = SaveMemberWallet(tx, ewtIn)
	if err != nil {
		base.LogErrorLog("PostTransfer - fail to save ewt in", err, ewtIn, true)
		return nil, err
	}

	//store ewt_transfer
	arrEwtTransfer := models.EwtTransfer{
		MemberIdFrom:   t.MemberId,
		MemberIdTo:     memToId,
		DocNo:          docs,
		EwtTypeFrom:    eWalletId,
		EwtTypeTo:      eWalletIdTo,
		TransferAmount: t.Amount,
		AdminFee:       adminFeeAmount,
		NettAmount:     transferAmount,
		Remark:         t.Remark,
		Status:         "AP",
		CreatedAt:      time.Now(),
		CreatedBy:      t.MemberId,
	}

	_, err = models.AddEwtTransfer(tx, arrEwtTransfer) //store transfer

	if err != nil {
		base.LogErrorLog("PostTransfer - fail to save ewt_transfer", err, arrEwtTransfer, true)
		return nil, err
	}

	err = models.UpdateRunningDocNo("WT", tx) //update transfer doc no

	if err != nil {
		base.LogErrorLog("PostTransfer -fail to update WT doc no", err, t, true)
		return nil, err
	}

	arrData := make(map[string]interface{})
	arrData["wallet_from"] = walSetup.EwtTypeCode
	arrData["wallet_to"] = walSetupTo.EwtTypeCode
	arrData["member_to"] = t.MemberTo
	arrData["remark"] = t.Remark
	arrData["amount"] = fmt.Sprintf("%f", t.Amount)
	if adminFeeAmount > 0 {
		arrData["nett_amount"] = fmt.Sprintf("%f", transferAmount)
		arrData["admin_fee_amount"] = fmt.Sprintf("%f", adminFeeAmount)
	}
	arrData["trans_time"] = time.Now().Format("2006-01-02 15:04:05")

	return arrData, nil

}

type WalletPaymentListStruct struct {
	EwtTypeCode   string  `json:"ewallet_type_code"`
	EwtTypeName   string  `json:"ewallet_type_name"`
	Balance       float64 `json:"balance"`
	DecimalPoint  int     `json:"decimal_point"`
	PayID         int     `json:"pay_id"`
	MinPercentage int     `json:"min_percentage"`
	MaxPercentage int     `json:"max_percentage"`
}

type WalletPaymentByRoomStruct struct {
	RoomCost          float64                   `json:"room_cost"`
	WalletPaymentList []WalletPaymentListStruct `json:"wallet_payment_list"`
}

func GetWalletPaymentInfoByRoomTypev1(entMemberID int, roomTypeCode string, langCode string) (*WalletPaymentByRoomStruct, error) {
	arrWalletPaymentList := make([]WalletPaymentListStruct, 0)
	arrDataReturn := WalletPaymentByRoomStruct{
		WalletPaymentList: arrWalletPaymentList,
	}
	return &arrDataReturn, nil
}

type PostWithdrawStruct struct {
	MemberId          int
	Address           string
	LangCode          string
	Amount            float64
	EwalletTypeCode   string
	Remark            string
	ChargesType       string
	EwalletTypeCodeTo string
}

type ChainPayResponse struct {
	StatusCode int                    `json:"statusCode"`
	Status     string                 `json:"status"`
	Message    []string               `json:"message"`
	Msg        string                 `json:"msg"`
	Data       map[string]interface{} `json:"data"`
}

func (w *PostWithdrawStruct) PostWithdraw(tx *gorm.DB) (interface{}, error) {
	var (
		err                  error
		decimalPoint         uint    = 2
		gasFee               float64 = 0
		adminFee             float64 = 0
		rate                 float64 = 1
		toWalletCurrency     string
		toWalletDecimalPoint uint    = 2
		netAmount            float64 = 0
		convertedNetAmount   float64 = 0
		convertedTotalAmount float64 = 0
		expDate              time.Time
	)

	//check member wallet balance - save member wallet function will handle

	//get wallet setup
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(w.EwalletTypeCode)},
	)
	WalSetup, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("PostWithdraw - fail to get wallet setup", err, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: err}
	}

	if WalSetup == nil {
		base.LogErrorLog("PostWithdraw - empty wallet setup", w, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: w}
	}

	decimalPoint = uint(WalSetup.DecimalPoint)
	eWalletId := WalSetup.ID

	//get wallet to setup
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(w.EwalletTypeCodeTo)},
	)
	WalSetupTo, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("PostWithdraw - fail to get wallet setup", err, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: err}
	}

	if WalSetupTo == nil {
		base.LogErrorLog("PostWithdraw - empty wallet setup to", w, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: w}
	}

	toWalletCurrency = WalSetupTo.CurrencyCode
	toWalletDecimalPoint = uint(WalSetupTo.DecimalPoint)
	eWalletIdTo := WalSetupTo.ID

	//get withdraw setup
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_withdraw_setup.ewallet_type_id = ?", CondValue: WalSetup.ID},
		models.WhereCondFn{Condition: "ewt_withdraw_setup.ewallet_type_id_to = ?", CondValue: WalSetupTo.ID},
		models.WhereCondFn{Condition: "ewt_withdraw_setup.withdraw_type = ?", CondValue: "CRYPTO"},
	)

	arrWithdrawSetup, err := models.GetEwtWithdrawSetupFnV2(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("PostWithdraw - GetEwtWithdrawSetupFnV2-General Setup", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if arrWithdrawSetup == nil {
		base.LogErrorLog("PostWithdraw - empty withdraw setup", arrWithdrawSetup, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: w}
	}

	//get withdraw setup - charges type

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_withdraw_setup.ewallet_type_id = ?", CondValue: WalSetup.ID},
		models.WhereCondFn{Condition: "ewt_withdraw_setup.withdraw_type = ?", CondValue: "CRYPTO"},
		models.WhereCondFn{Condition: "ewt_withdraw_setup.charges_type = ?", CondValue: w.ChargesType},
	)

	arrWithdrawSetupCharges, err := models.GetEwtWithdrawSetupFnV2(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("PostWithdraw - GetEwtWithdrawSetupFnV2 - Charges", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if arrWithdrawSetupCharges == nil {
		base.LogErrorLog("PostWithdraw - empty withdraw setup for charges", arrWithdrawSetup, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: w}
	}

	adminFee = float64(arrWithdrawSetupCharges.AdminFee)

	//begin checking

	if WalSetup.EwtTypeCode == "TP" || WalSetup.EwtTypeCode == "RPB" || WalSetup.EwtTypeCode == "RPA" {
		//check limiter balance
		withdrawBal, err := GetMemberWithdrawBalance(w.MemberId, eWalletId)
		if err != nil {
			base.LogErrorLog("PostWithdraw - GetMemberWithdrawBalance Err", err, withdrawBal, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: w}
		}

		if w.Amount > withdrawBal {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_enough_balance_to_deduct", w.LangCode), Data: ""}
		}
	}

	//check member wallet lock
	walletLock, err := models.GetEwtLockByMemberId(w.MemberId, WalSetup.ID)

	if err != nil {
		base.LogErrorLog("PostWithdraw - fail to get ewtLock Setup", err, w, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: w}
	}

	if walletLock.Withdrawal == 1 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("wallet_is_being_locked_from_withdraw", w.LangCode), Data: w}
	}

	//check amt
	if w.Amount <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("amount_must_more_than_0", w.LangCode), Data: ""}
	}

	// check min
	if w.Amount < arrWithdrawSetup.Min {
		strAmt := helpers.CutOffDecimal(arrWithdrawSetup.Min, 2, ".", "")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("minimum_withdraw_amount_is"+" "+strAmt, w.LangCode), Data: w}
	}

	//check multiple of
	if arrWithdrawSetup.MultipleOf > 0 {
		multipleOf := float64(arrWithdrawSetup.MultipleOf)
		if !helpers.IsMultipleOf(w.Amount, multipleOf) {
			strAmt := helpers.CutOffDecimal(multipleOf, 2, ".", "")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("amount_must_be_multiple_of"+" "+strAmt, w.LangCode), Data: w}
		}
	}

	//check wait previous
	if arrWithdrawSetup.WaitPreviousDone == 1 {
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_withdraw.status IN ('W','PR') AND ewt_withdraw.member_id = ? ", CondValue: w.MemberId},
		)
		checkPendingWithdraw, err := models.GetEwtWithdrawFn(arrCond, false)

		if err != nil {
			base.LogErrorLog("PostWithdraw - checkPendingWithdraw", err.Error(), arrCond, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if len(checkPendingWithdraw) > 0 {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("please_wait_previous_withdraw_to_complete", w.LangCode), Data: ""}
		}
	}

	//check address
	if WalSetupTo.BlockchainCryptoTypeCode == "TRX" {
		checkAddr := base.IsValidTRXAddress(w.Address)
		if checkAddr == false {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_trx_address", w.LangCode), Data: w}
		}
	} else {
		//check eth
		checkAddr := base.IsValidETHAddress(w.Address)

		if checkAddr == false {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_eth_address", w.LangCode), Data: w}
		}
	}

	//check member whether is first time withdraw
	// arrCond = make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: " ewt_withdraw.member_id = ? ", CondValue: w.MemberId},
	// 	models.WhereCondFn{Condition: " ewt_withdraw.status = ? ", CondValue: "AP"},
	// )
	// checkFirstWithdraw, err := models.GetEwtWithdrawFn(arrCond, false)

	// if err != nil {
	// 	base.LogErrorLog("PostWithdraw - GetEwtWithdrawFn", err.Error(), arrCond, true)
	// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	// }

	// if len(checkFirstWithdraw) < 1 {
	// 	//first time withdraw
	// 	if w.ChargesType != "FREE" {
	// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_withdraw_charges", w.LangCode), Data: w}
	// 	}
	// } else {
	// 	//check whether countdown end
	// 	arrCond = make([]models.WhereCondFn, 0)
	// 	arrCond = append(arrCond,
	// 		models.WhereCondFn{Condition: "ewt_withdraw.member_id = ?", CondValue: w.MemberId},
	// 		models.WhereCondFn{Condition: "ewt_withdraw.charges_type = ?", CondValue: "FREE"},
	// 		models.WhereCondFn{Condition: "ewt_withdraw.status = ?", CondValue: "AP"},
	// 	)

	// 	arrCheckCountdown, err := models.GetEwtWithdrawFn(arrCond, false) //get latest free record
	// 	if err != nil {
	// 		base.LogErrorLog("PostWithdraw - check countdown status err", err.Error(), arrCond, true)
	// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	// 	}

	// 	if len(arrCheckCountdown) > 0 {
	// 		//not yet expired
	// 		if helpers.CompareDateTime(time.Now(), "<", arrCheckCountdown[0].ExpiredAt) {
	// 			if w.ChargesType == "FREE" {
	// 				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_aplicable_for_free_withdrawal", w.LangCode), Data: w}
	// 			}
	// 		}
	// 	}
	// }

	//check whether countdown end
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_withdraw.member_id = ?", CondValue: w.MemberId},
		models.WhereCondFn{Condition: "ewt_withdraw.charges_type = ?", CondValue: "FREE"},
		models.WhereCondFn{Condition: "ewt_withdraw.status = ?", CondValue: "AP"},
	)

	arrCheckCountdown, err := models.GetEwtWithdrawFn(arrCond, false) //get latest free record
	if err != nil {
		base.LogErrorLog("PostWithdraw - check countdown status err", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if len(arrCheckCountdown) > 0 {
		//not yet expired
		if helpers.CompareDateTime(time.Now(), "<", arrCheckCountdown[0].ExpiredAt) {
			if w.ChargesType == "FREE" {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_aplicable_for_free_withdrawal", w.LangCode), Data: w}
			}
		}
	}

	//end checking

	//get rate
	if WalSetup.CurrencyCode != WalSetupTo.CurrencyCode { //if to wallet currency not same get need rate
		rate, err = base.GetLatestPriceMovementByTokenType(w.EwalletTypeCode)
		if err != nil {
			base.LogErrorLog("PostWithdraw - GetLatestPriceMovementByTokenType Error", err, w, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: w}
		}
	}

	netAmount = w.Amount

	//get gas fee
	if WalSetupTo.BlockchainCryptoTypeCode == "TRX" {
		gasFee, _ = models.GetLatestGasFeeMovementTron()
	}

	if WalSetupTo.BlockchainCryptoTypeCode == "ETH" {
		gasFee, _ = models.GetLatestGasFeeMovementErc20()
	}

	//deduct gas fee
	netAmount, _ = decimal.NewFromFloat(netAmount).Sub(decimal.NewFromFloat(gasFee)).Float64()
	convertedNetAmount, _ = decimal.NewFromFloat(netAmount).Mul(decimal.NewFromFloat(rate)).Float64()
	convertedTotalAmount, _ = decimal.NewFromFloat(w.Amount).Mul(decimal.NewFromFloat(rate)).Float64()

	if netAmount <= 0 || convertedNetAmount <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_enough_amount_to_deduct_gas_fee", w.LangCode), Data: w}
	}

	adminFeeAmount, _ := decimal.NewFromFloat(w.Amount).Mul(decimal.NewFromFloat(adminFee)).Float64()
	adminFeeAmount, _ = decimal.NewFromFloat(adminFeeAmount).Div(decimal.NewFromFloat(100)).Float64()
	convertedAdminFee, _ := decimal.NewFromFloat(adminFeeAmount).Mul(decimal.NewFromFloat(rate)).Float64()

	//deduct admin fee
	netAmount, _ = decimal.NewFromFloat(netAmount).Sub(decimal.NewFromFloat(adminFeeAmount)).Float64()
	convertedNetAmount, _ = decimal.NewFromFloat(netAmount).Mul(decimal.NewFromFloat(rate)).Float64()
	convertedTotalAmount, _ = decimal.NewFromFloat(netAmount).Mul(decimal.NewFromFloat(rate)).Float64()

	if netAmount <= 0 || convertedNetAmount <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_enough_amount_to_deduct_admin_fee", w.LangCode), Data: w}
	}

	docs, err := models.GetRunningDocNo("WD", tx) //get withdraw doc no

	if err != nil {
		base.LogErrorLog("PostWithdraw - fail to get WD doc no", err, w, true) //store error log
		return nil, err
	}

	//deduct balance
	ConvertedTotalOut, _ := decimal.NewFromFloat(w.Amount).Mul(decimal.NewFromFloat(rate)).Float64()
	ewtOut := SaveMemberWalletStruct{
		EntMemberID:       w.MemberId,
		EwalletTypeID:     eWalletId,
		TotalOut:          w.Amount,
		ConversionRate:    rate,
		ConvertedTotalOut: ConvertedTotalOut,
		TransactionType:   "WITHDRAW",
		DocNo:             docs,
		Remark:            docs,
	}

	_, err = SaveMemberWallet(tx, ewtOut)

	if err != nil {
		base.LogErrorLog("PostWithdraw - fail to save ewt_detail", err, ewtOut, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: err}
	}

	if w.ChargesType == "FREE" {
		expDate = time.Now().AddDate(0, 0, arrWithdrawSetup.CountdownDays)
	} else {
		expDate = time.Now()
	}

	transType := strings.ToUpper(WalSetupTo.EwtTypeCode)

	if transType == "USDT" {
		transType = "USDT_TRC20"
	}
	if transType == "USDC" {
		transType = "USDC_TRC20"
	}

	//save ewt_withdraw
	arrEwtWithdraw := models.AddEwtWithdrawStruct{
		MemberId:              w.MemberId,
		DocNo:                 docs,
		EwalletTypeId:         eWalletId,
		EwalletTypeIdTo:       eWalletIdTo,
		CurrencyCode:          strings.ToUpper(WalSetup.CurrencyCode),
		Type:                  "CRYPTO",
		TransactionType:       strings.ToUpper(WalSetupTo.EwtTypeCode),
		TransDate:             time.Now(),
		TotalOut:              w.Amount,
		NetAmount:             netAmount,
		ChargesType:           w.ChargesType,
		AdminFee:              adminFeeAmount,
		ConversionRate:        rate,
		ConvertedTotalAmount:  convertedTotalAmount,
		ConvertedNetAmount:    convertedNetAmount,
		ConvertedAdminFee:     convertedAdminFee,
		ConvertedCurrencyCode: toWalletCurrency,
		CryptoAddrTo:          w.Address,
		GasFee:                gasFee,
		Status:                "W",
		CreatedAt:             time.Now(),
		CreatedBy:             w.MemberId,
		ExpiredAt:             expDate,
	}

	_, err = models.AddEwtWithdraw(tx, arrEwtWithdraw) //store withdraw

	if err != nil {
		base.LogErrorLog("PostWithdraw - fail to save ewt_withdraw", err, arrEwtWithdraw, true) //store error log
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: err}
	}

	err = models.UpdateRunningDocNo("WD", tx) //update withdraw doc no

	if err != nil {
		base.LogErrorLog("PostWithdraw - fail to update WD doc no", err, w, true) //store error log
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: err}
	}

	convertedNetAmountStr := helpers.CutOffDecimal(convertedNetAmount, toWalletDecimalPoint, ".", ",")
	gasFeeAmountStr := helpers.CutOffDecimal(gasFee, decimalPoint, ".", ",")
	adminFeeAmountStr := helpers.CutOffDecimal(adminFeeAmount, decimalPoint, ".", ",")

	arrData := make(map[string]interface{})
	arrData["ewallet_type"] = helpers.Translate(w.EwalletTypeCode, w.LangCode)
	arrData["address_to"] = w.Address
	arrData["remark"] = w.Remark
	arrData["amount"] = w.Amount
	arrData["payment"] = convertedNetAmountStr + " " + toWalletCurrency
	if gasFee > 0 {
		arrData["gas_fee"] = gasFeeAmountStr + " " + strings.ToUpper(w.EwalletTypeCode)
	}

	if adminFee > 0 {
		arrData["admin_fee"] = adminFeeAmountStr
	}

	arrData["trans_time"] = time.Now().Format("2006-01-02 15:04:05")
	arrData["ewallet_type_to"] = helpers.Translate(WalSetupTo.EwtTypeName, w.LangCode)

	return arrData, nil

}

// type PaymentStruct struct {
// 	PayID  int     `json:"pay_id"`
// 	Amount float64 `json:"amount"`
// }

type WalletStatementListStruct struct {
	ID              string `gorm:"primary_key" json:"id"`
	TransDate       string `json:"trans_date"`
	EwalletTypeName string `json:"ewallet_type_name"`
	TransType       string `json:"trans_type"`
	CreditIn        string `json:"credit_in"`
	CreditOut       string `json:"credit_out"`
	Status          string `json:"status"`
	Remark          string `json:"remark"`
}

type WithdrawStatementListStruct struct {
	ID              string `gorm:"primary_key" json:"id"`
	DocNo           string `json:"doc_no"`
	TransDate       string `json:"trans_date"`
	EwalletTypeName string `json:"ewallet_type_name"`
	// WithdrawType    string `json:"withdraw_type"`
	TransType string `json:"trans_type"`
	// CreditOut       string `json:"credit_out"`
	// GasFee          string `json:"gas_fee"`
	// NetAmount       string `json:"net_amount"`
	Amount          string `json:"amount"`
	Address         string `json:"address"`
	Hash            string `json:"hash"`
	Status          string `json:"status"`
	StatusColorCode string `json:"status_color_code"`
	Remark          string `json:"remark"`
	CancelStatus    int    `json:"cancel_status"`
}

type TransferStatementListStruct struct {
	ID              string `gorm:"primary_key" json:"id"`
	MemberFrom      string `json:"member_from"`
	MemberTo        string `json:"member_to"`
	DocNo           string `json:"doc_no"`
	EwalletTypeFrom string `json:"ewallet_type_from"`
	EwalletTypeTo   string `json:"ewallet_type_to"`
	CreditOut       string `json:"credit_out"`
	TransType       string `json:"trans_type"`
	TransDate       string `json:"trans_date"`
	Status          string `json:"status"`
	Remark          string `json:"remark"`
}

type TransferStatementListToStruct struct {
	ID              string `gorm:"primary_key" json:"id"`
	MemberFrom      string `json:"member_from"`
	MemberTo        string `json:"member_to"`
	DocNo           string `json:"doc_no"`
	EwalletTypeFrom string `json:"ewallet_type_from"`
	EwalletTypeTo   string `json:"ewallet_type_to"`
	CreditIn        string `json:"credit_in"`
	TransType       string `json:"trans_type"`
	TransDate       string `json:"trans_date"`
	Status          string `json:"status"`
	Remark          string `json:"remark"`
}

type WalletTransactionStruct struct {
	MemberID int    `json:"member_id"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
	Page     int64  `json:"page"`
	LangCode string `json:"lang_code"`
}

/* this func is shared by wallet statement, transfer statement and withdraw statement*/
func (s *WalletTransactionStruct) WalletStatement(listType string) (app.ArrDataResponseList, error) {

	transType := ""
	if listType != "" {
		transType = listType
	}

	ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, err := models.GetEwtDetailForTransactionList(s.Page, s.MemberID, transType, s.DateFrom, s.DateTo)

	if s.Page == 0 {
		s.Page = 1
	}

	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}

	var arrTableHeaderList []arrStatementListSettingListStruct

	// var arrWalletStatementList []interface{}
	// arrWalletStatementList := make([]TransferStatementListStruct, 0)

	arrWalletStatementList := make([]interface{}, 0)

	//if not empty result
	if ewt != nil {

		switch listType {
		case "WITHDRAW":
			arrStatementListSetting, _ := models.GetSysGeneralSetupByID("withdraw_statement_api_setting")
			if arrStatementListSetting != nil {
				var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
				json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
				arrTableHeaderList = arrStatementListSettingList["table_header_list"]
				for k, v1 := range arrStatementListSettingList["table_header_list"] {
					v1.Name = helpers.Translate(v1.Name, s.LangCode)
					arrTableHeaderList[k] = v1
				}
			}

			for _, v := range ewt {
				status := helpers.Translate("completed", s.LangCode)
				// ewtTypeName := helpers.Translate(v.EwalletTypeName, s.LangCode)
				// transType := helpers.Translate(v.TransactionType, s.LangCode)

				withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

				if withdrawDet != nil {
					// withdrawType := helpers.Translate(withdrawDet.Type, s.LangCode)
					status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
					// remark := helpers.TransRemark(withdrawDet.Remark, s.LangCode)
					// remark := helpers.Translate(withdrawDet.Remark, s.LangCode)

					arrWalletStatementList = append(arrWalletStatementList,
						WithdrawStatementListStruct{
							ID:        strconv.Itoa(withdrawDet.ID),
							DocNo:     withdrawDet.DocNo,
							TransDate: withdrawDet.TransDate.Format("2006-01-02 15:04:05"),
							// EwalletTypeName: ewtTypeName,
							// WithdrawType:    withdrawType,
							// TransType:       transType,
							// CreditOut:       fmt.Sprintf("%.2f", withdrawDet.TotalOut),
							// GasFee:          fmt.Sprintf("%.6f", withdrawDet.GasFee),
							// NetAmount:       fmt.Sprintf("%.6f", withdrawDet.NetAmount),
							Amount:  fmt.Sprintf("%.6f", withdrawDet.NetAmount),
							Address: withdrawDet.CryptoAddrTo,
							Status:  status,
							// Remark:       remark,
						})
				}
			}

		case "TRANSFER":
			arrStatementListSetting, _ := models.GetSysGeneralSetupByID("transfer_statement_api_setting")
			if arrStatementListSetting != nil {
				var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
				json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
				arrTableHeaderList = arrStatementListSettingList["table_header_list"]
				for k, v1 := range arrStatementListSettingList["table_header_list"] {
					v1.Name = helpers.Translate(v1.Name, s.LangCode)
					arrTableHeaderList[k] = v1
				}
			}

			for _, v := range ewt {
				status := helpers.Translate("completed", s.LangCode)
				transType := helpers.Translate(v.TransactionType, s.LangCode)
				transferDet, _ := models.GetEwtTransferDetailByDocNo(v.DocNo)
				remark := helpers.TransRemark(v.Remark, s.LangCode)

				if transferDet != nil {
					ewtTypeFrom := helpers.Translate(transferDet.WalletFrom, s.LangCode)
					ewtTypeTo := helpers.Translate(transferDet.WalletTo, s.LangCode)
					// status = helpers.Translate(transferDet.StatusDesc, s.LangCode)
					// remark := helpers.TransRemark(transferDet.Reason, s.LangCode)

					if transferDet.MemberIdFrom == s.MemberID { //transfer from member view
						arrWalletStatementList = append(arrWalletStatementList,
							TransferStatementListStruct{
								ID:              strconv.Itoa(transferDet.ID),
								DocNo:           transferDet.DocNo,
								TransDate:       transferDet.CreatedAt.Format("2006-01-02 15:04:05"),
								MemberFrom:      transferDet.MemberFrom,
								MemberTo:        transferDet.MemberTo,
								EwalletTypeFrom: ewtTypeFrom,
								EwalletTypeTo:   ewtTypeTo,
								TransType:       transType,
								CreditOut:       fmt.Sprintf("%.2f", transferDet.TransferAmount),
								Status:          status,
								Remark:          remark,
							})
					} else if transferDet.MemberIdTo == s.MemberID {
						arrWalletStatementList = append(arrWalletStatementList,
							TransferStatementListToStruct{
								ID:              strconv.Itoa(transferDet.ID),
								DocNo:           transferDet.DocNo,
								TransDate:       transferDet.CreatedAt.Format("2006-01-02 15:04:05"),
								MemberFrom:      transferDet.MemberFrom,
								MemberTo:        transferDet.MemberTo,
								EwalletTypeFrom: ewtTypeFrom,
								EwalletTypeTo:   ewtTypeTo,
								TransType:       transType,
								CreditIn:        fmt.Sprintf("%.2f", transferDet.TransferAmount),
								Status:          status,
								Remark:          remark,
							})
					}
				}
			}

		default: //wallet statement

			arrStatementListSetting, _ := models.GetSysGeneralSetupByID("wallet_statement_api_setting")
			if arrStatementListSetting != nil {
				var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
				json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
				arrTableHeaderList = arrStatementListSettingList["table_header_list"]
				for k, v1 := range arrStatementListSettingList["table_header_list"] {
					v1.Name = helpers.Translate(v1.Name, s.LangCode)
					arrTableHeaderList[k] = v1
				}
			}

			for _, v := range ewt {
				status := helpers.Translate("completed", s.LangCode)
				ewtTypeName := helpers.Translate(v.EwalletTypeName, s.LangCode)
				transType := helpers.Translate(v.TransactionType, s.LangCode)
				remark := helpers.TransRemark(v.Remark, s.LangCode)

				if v.TransactionType == "WITHDRAW" {
					//get withdraw detail
					withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

					if withdrawDet != nil {
						status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
					}
				}
				// else if v.TransactionType == "TRANSFER" {
				// 	//get transfer detail
				// 	transferDet, _ := models.GetEwtTransferDetailByDocNo(v.DocNo)

				// 	if transferDet != nil {
				// 		status = helpers.Translate(transferDet.StatusDesc, s.LangCode)
				// 	}
				// }

				TotalIn := fmt.Sprintf("%.2f", v.TotalIn)
				if TotalIn == "0.00" {
					TotalIn = ""
				}

				TotalOut := fmt.Sprintf("%.2f", v.TotalOut)
				if TotalOut == "0.00" {
					TotalOut = ""
				}

				arrWalletStatementList = append(arrWalletStatementList,
					WalletStatementListStruct{
						ID:              strconv.Itoa(v.ID),
						TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
						EwalletTypeName: ewtTypeName,
						TransType:       transType,
						CreditIn:        TotalIn,
						CreditOut:       TotalOut,
						Status:          status,
						Remark:          remark,
					})

			}

		}
	}

	if err != nil {
		models.ErrorLog("failed_to_get_ewt_detail", err, s) //store error log
		return app.ArrDataResponseList{}, err
	}

	arrDataReturn := app.ArrDataResponseList{
		CurrentPage:           int(s.Page),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      arrWalletStatementList,
		TableHeaderList:       arrTableHeaderList,
	}

	return arrDataReturn, nil
}

type WithdrawTransactionFeeStruct struct {
	MemId           int    `json:"member_id"`
	Address         string `json:"address"`
	Amount          string `json:"amount"`
	EwalletTypeCode string `json:"ewallet_type_code"`
	CryptoType      string `json:"crypto_type"`
	LangCode        string `json:"lang_code"`
}

type WithdrawTransactionFeeReturnStruct struct {
	EthEstimateFee        float64 `json:"eth_estimate_fee"`
	ConvertRate           float64 `json:"convert_rate"`
	Markup                float64 `json:"markup"`
	ConvertRateWithMarkup float64 `json:"convert_rate_with_markup"`
	GasPrice              string  `json:"gas_price"`
	USDTGasFee            float64 `json:"usdt_gas_fee"`
}

type ConvertRateResponse struct {
	USDT float64 `json:"USDT"`
}

func (w *WithdrawTransactionFeeStruct) GetWithdrawTransactionFee() (WithdrawTransactionFeeReturnStruct, error) {

	var response ChainPayResponse     //for chainpay return response
	var response2 ConvertRateResponse //for convert return response

	//call yeejia site for transaction fee
	apiSetting, _ := models.GetSysGeneralSetupByID("chainpay_api_setting")

	data := map[string]interface{}{
		"Address":     w.Address,
		"Amount":      w.Amount,
		"Asset_code":  w.EwalletTypeCode, //current only USDT
		"Config_code": "SEC_ETH",
		"Crypto_type": w.CryptoType,
	}

	api_key := apiSetting.InputValue2

	hash_key := base.HashInput(data, api_key)

	url := apiSetting.InputValue1 + "/estimateTransactionFee"
	header := map[string]string{
		"Content-Type": "application/json",
		"Hash-Key":     hash_key,
	}

	res, err_api := base.RequestAPI("POST", url, header, data, &response)

	if err_api != nil {
		base.LogErrorLogV2("get_estimate_transaction_fee_failed", err_api.Error(), map[string]interface{}{"err": err_api, "data": data}, true, "blockchain")
		models.EwtWithdrawalLog(w.MemId, "transaction-fee-api", data, err_api, nil, nil)
		return WithdrawTransactionFeeReturnStruct{}, err_api
	}

	if res.StatusCode != 200 {
		errMsg, _ := json.Marshal(response.Msg)

		base.LogErrorLogV2("estimate_transaction_fee_SEC", "", res.Body, true, "blockchain")
		models.EwtWithdrawalLog(w.MemId, "transaction-fee-api", data, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: string(errMsg)}, nil, nil) //store withdrawal log

		return WithdrawTransactionFeeReturnStruct{}, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: helpers.Translate(strings.Replace(string(errMsg), "\"", "", 2), w.LangCode)}
	}

	api_return_result, _ := json.Marshal(response.Data["estimate_fee"])

	fee := string(api_return_result)
	fee = strings.Replace(fee, "\"", "", 2)
	eth_estimate_fee, _ := strconv.ParseFloat(fee, 64)

	api_return_result2, _ := json.Marshal(response.Data["gas_price"]) //for gas_price

	gas_price := string(api_return_result2)
	gas_price = strings.Replace(gas_price, "\"", "", 2)

	//get markup from setting

	// arrWalCond := make([]models.WhereCondFn, 0)
	// arrWalCond = append(arrWalCond,
	// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: w.EwalletTypeCode},
	// )
	// walSetup, walSetupErr := models.GetEwtSetupFn(arrWalCond, "", false)

	// if walSetupErr != nil {
	// 	base.LogErrorLog(walSetupErr, "withdraw-transaction-fee", walSetupErr, true)
	// 	models.EwtWithdrawalLog(w.MemId, "transaction-fee-api", data, walSetupErr, nil, nil)
	// 	return WithdrawTransactionFeeReturnStruct{}, walSetupErr
	// }

	//get convert api
	url2 := "https://min-api.cryptocompare.com/data/price"

	header2 := map[string]string{
		"Content-Type": "application/json",
	}

	data2 := map[string]interface{}{
		"fsym":  "ETH",             //from
		"tsyms": w.EwalletTypeCode, //to
	}

	_, err_api2 := base.RequestAPI("GET", url2, header2, data2, &response2)

	if err_api2 != nil {
		models.ErrorLog("get_convert_rate_fail", err_api2.Error(), map[string]interface{}{"err": err_api2, "data": data2})
		models.EwtWithdrawalLog(w.MemId, "transaction-fee-api", data, err_api2, nil, nil)
		return WithdrawTransactionFeeReturnStruct{}, err_api2
	}

	api2_return_result, _ := json.Marshal(response2.USDT)

	rate := string(api2_return_result)

	floatRate, _ := strconv.ParseFloat(rate, 64)

	rate2 := floatRate

	// if walSetup.WithdrawMarkup != 0 {
	// 	floatRate = floatRate * (1 + walSetup.WithdrawMarkup) //already in percentage
	// }

	USDTGasFee := eth_estimate_fee * floatRate

	USDTGasFeeDecimal := fmt.Sprintf("%.6f", USDTGasFee)

	USDTGasFeeFloat, _ := strconv.ParseFloat(USDTGasFeeDecimal, 64)

	arrDataReturn := WithdrawTransactionFeeReturnStruct{
		EthEstimateFee: eth_estimate_fee,
		ConvertRate:    rate2,
		// Markup:                walSetup.WithdrawMarkup,
		ConvertRateWithMarkup: floatRate,
		GasPrice:              gas_price,
		USDTGasFee:            USDTGasFeeFloat,
	}

	return arrDataReturn, nil
}

func (w *WithdrawTransactionFeeStruct) GetWithdrawTransactionFeeV2() (string, error) {

	var response ChainPayResponse //for chainpay return response

	//call yeejia site for transaction fee
	apiSetting, _ := models.GetSysGeneralSetupByID("chainpay_api_setting")

	data := map[string]interface{}{
		"Address":     w.Address,
		"Amount":      w.Amount,
		"Asset_code":  w.EwalletTypeCode,
		"Config_code": "SEC_ETH",
		"Crypto_type": w.CryptoType,
	}

	api_key := apiSetting.InputValue2

	hash_key := base.HashInput(data, api_key)

	url := apiSetting.InputValue1 + "/estimateTransactionFee"
	header := map[string]string{
		"Content-Type": "application/json",
		"Hash-Key":     hash_key,
	}

	res, err_api := base.RequestAPI("POST", url, header, data, &response)

	if err_api != nil {
		base.LogErrorLogV2("get_estimate_transaction_fee_failed", err_api.Error(), map[string]interface{}{"err": err_api, "data": data}, true, "blockchain")
		models.EwtWithdrawalLog(w.MemId, "transaction-fee-api", data, err_api, nil, nil)
		return "", err_api
	}

	if res.StatusCode != 200 {
		errMsg, _ := json.Marshal(response.Msg)

		base.LogErrorLogV2("estimate_transaction_fee_SEC", "", res.Body, true, "blockchain")
		models.EwtWithdrawalLog(w.MemId, "transaction-fee-api", data, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: string(errMsg)}, nil, nil) //store withdrawal log

		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: helpers.Translate(strings.Replace(string(errMsg), "\"", "", 2), w.LangCode)}
	}

	api_return_result2, _ := json.Marshal(response.Data["gas_price"]) //for gas_price

	gas_price := string(api_return_result2)
	gas_price = strings.Replace(gas_price, "\"", "", 2)

	return gas_price, nil
}

// ProcessUpdateCryptoWithdrawalv1Struct struct
type ProcessUpdateCryptoWithdrawalv1Struct struct {
	BatchID      int
	ConfigCode   string
	AssetCode    string
	DocNo        string
	ToAddress    string
	TxHash       string
	Value        string
	Gas          string
	GasPrice     string
	RunStatus    string // either "COMPLETE" or "REJECTED"
	RunAttempts  int
	RunStatusAt  string
	ApproverName string
	ApproveAt    string
	Remark       string
}

// func ProcessUpdateCryptoWithdrawalv1
func ProcessUpdateCryptoWithdrawalv1(tx *gorm.DB, arrData ProcessUpdateCryptoWithdrawalv1Struct) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_withdraw.doc_no = ? ", CondValue: arrData.DocNo},
		models.WhereCondFn{Condition: " ewt_withdraw.chainpay_batch_id = ? ", CondValue: arrData.BatchID},
		models.WhereCondFn{Condition: " ewt_withdraw.status IN ('W','PR') AND ewt_withdraw.crypto_addr_to = ? ", CondValue: arrData.ToAddress},
	)
	arrEwtWithdraw, err := models.GetEwtWithdrawFn(arrCond, false)

	if err != nil {
		models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-failed_to_get_arrEwtWithdraw", err.Error(), arrCond)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if len(arrEwtWithdraw) < 1 {
		// models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-failed_to_get_arrEwtWithdraw", "invalid_record", arrCond)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_crypto_witdhrawal_record", Data: arrCond}
	}

	if strings.ToLower(arrData.RunStatus) == "complete" && arrData.TxHash != "" { // perform update transaction to succeess
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_withdraw.tran_hash = ? ", CondValue: arrData.TxHash},
		)
		arrExistingEwtWithdrawTxHash, _ := models.GetEwtWithdrawFn(arrCond, false)

		if len(arrExistingEwtWithdrawTxHash) > 0 {
			models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-invalid_tx_hash", "invalid_record", arrCond)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_tx_hash", Data: arrCond}
		}

		if arrEwtWithdraw[0].TranHash != "" {
			// models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-failed_to_get_arrEwtWithdraw", "invalid_record", arrCond)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_tx_hash_crypto_witdhrawal_record", Data: arrCond}
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_withdraw.doc_no = ? ", CondValue: arrData.DocNo},
			models.WhereCondFn{Condition: " ewt_withdraw.chainpay_batch_id = ? ", CondValue: arrData.BatchID},
			models.WhereCondFn{Condition: " ewt_withdraw.crypto_addr_to = ? ", CondValue: arrData.ToAddress},
		)
		updateColumn := map[string]interface{}{
			"tran_hash":   arrData.TxHash,
			"status":      "CP",
			"approved_at": arrData.ApproveAt,
			"approved_by": arrData.ApproverName,
			"updated_at":  arrData.RunStatusAt,
			"updated_by":  "BLOCKCHAIN",
		}

		if arrData.Remark != "" {
			updateColumn["remark"] = arrData.Remark
		}
		err := models.UpdatesFnTx(tx, "ewt_withdraw", arrCond, updateColumn, false)

		if err != nil {
			models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-update_ewt_withdraw_success_failed", err.Error(), arrCond)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}
		return nil
	} else if strings.ToLower(arrData.RunStatus) == "rejected" {
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_withdraw.doc_no = ? ", CondValue: arrData.DocNo},
			models.WhereCondFn{Condition: " ewt_withdraw.chainpay_batch_id = ? ", CondValue: arrData.BatchID},
			models.WhereCondFn{Condition: " ewt_withdraw.crypto_addr_to = ? ", CondValue: arrData.ToAddress},
		)
		updateColumn := map[string]interface{}{
			"status":      "R",
			"rejected_at": arrData.ApproveAt,
			"rejected_by": arrData.ApproverName,
			"updated_at":  arrData.RunStatusAt,
			"updated_by":  "BLOCKCHAIN",
		}

		if arrData.Remark != "" {
			updateColumn["remark"] = arrData.Remark
		}
		err := models.UpdatesFnTx(tx, "ewt_withdraw", arrCond, updateColumn, false)

		if err != nil {
			models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-update_reject_ewt_withdraw_failed", err.Error(), arrCond)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_detail.doc_no = ? ", CondValue: arrData.DocNo},
			models.WhereCondFn{Condition: " ewt_detail.total_out > 0 AND ewt_detail.transaction_type = ? ", CondValue: "WITHDRAW"},
		)
		arrEwtDetail, err := models.GetEwtDetailFn(arrCond, false)

		if err != nil {
			models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-failed_to_get_arrEwtDetail", err.Error(), arrCond)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if len(arrEwtDetail) > 0 {
			// start return money based on withdraw deduction
			for _, v1 := range arrEwtDetail {
				ewtIn := SaveMemberWalletStruct{
					EntMemberID:     v1.MemberID,
					EwalletTypeID:   v1.EwalletTypeID,
					TotalIn:         v1.TotalOut,
					TransactionType: "WITHDRAW",
					DocNo:           v1.DocNo,
					Remark:          "#*reject_withdraw*# " + v1.DocNo,
					CreatedBy:       "AUTO",
				}

				_, err := SaveMemberWallet(tx, ewtIn)
				if err != nil {
					models.ErrorLog("ProcessUpdateCryptoWithdrawalv1-save_wallet_failed", err.Error(), ewtIn)
					return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
				}
			}
		}
		return nil
	}

	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_action"}
}

type WalletTransactionStructV2 struct {
	MemberID       int    `json:"member_id"`
	DateFrom       string `json:"date_from"`
	DateTo         string `json:"date_to"`
	Page           int64  `json:"page"`
	LangCode       string `json:"lang_code"`
	WalletTypeCode string `json:"wallet_type_code"`
	TransType      string `json:"trans_type"`
	RewardTypeCode string `json:"reward_type_code"`
}

type WalletTransactionResultStructV2 struct {
	ID              string `json:"id"`
	DocNo           string `json:"doc_no"`
	EwalletTypeName string `json:"ewallet_type_name"`
	// TransactionType string `json:"transaction_type"`
	TransDate string `json:"trans_date"`
	TransType string `json:"type"`
	TotalIn   string `json:"total_in"`
	TotalOut  string `json:"total_out"`
	// Balance         string `json:"balance"`
	AdditionalMsg string `json:"additional_msg"`
	Remark        string `json:"remark"`
	Status        string `json:"status"`
	// CreatedAt       string `json:"created_at"`
	// CreatedBy       string `json:"created_by"`
}

/* this func is shared by wallet statement, transfer statement and withdraw statement*/
func (s *WalletTransactionStructV2) WalletStatementV2() (app.ArrDataResponseList, error) {

	ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, err := models.GetEwtDetailByWalletTypeForStatementList(s.Page, s.MemberID, s.TransType, s.DateFrom, s.DateTo, s.WalletTypeCode)

	if s.Page == 0 {
		s.Page = 1
	}

	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		// Popup int    `json:"popup"`
	}

	var (
		decimalPoint uint
	)

	var arrTableHeaderList []arrStatementListSettingListStruct

	arrStatementListSetting, _ := models.GetSysGeneralSetupByID("wallet_statement_api_v2_setting")
	if arrStatementListSetting != nil {
		var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
		json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
		arrTableHeaderList = arrStatementListSettingList["table_header_list"]
		for k, v1 := range arrStatementListSettingList["table_header_list"] {
			v1.Name = helpers.Translate(v1.Name, s.LangCode)
			arrTableHeaderList[k] = v1
		}
	}

	arrWalletStatementList := make([]WalletTransactionResultStructV2, 0)

	for _, v := range ewt {
		status := helpers.Translate("completed", s.LangCode)
		remark := helpers.TransRemark(v.Remark, s.LangCode)
		// balance := fmt.Sprintf("%.2f", v.Balance)
		docNo := v.DocNo
		trans_type := helpers.Translate(v.Type, s.LangCode)

		// if s.TransType != "" {
		// 	balance = ""
		// }

		decimalPoint = 2
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: v.EwalletTypeID},
		)
		ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
		if ewtSetup != nil {
			decimalPoint = uint(ewtSetup.DecimalPoint)
		}
		TotalIn := helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")

		if TotalIn == "0.00" {
			TotalIn = ""
		}

		TotalOut := helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
		if TotalOut == "0.00" {
			TotalOut = ""
		}

		// if v.TransType == "WITHDRAW" {
		// 	//get withdraw detail
		// 	withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

		// 	if withdrawDet != nil {
		// 		status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
		// 		CryptoAddress = withdrawDet.CryptoAddrTo
		// 		GasFee = fmt.Sprintf("%.6f", withdrawDet.GasFee)
		// 		NetAmount = fmt.Sprintf("%.6f", withdrawDet.NetAmount)
		// 	}
		// } else if v.TransType == "TRANSFER" {
		// 	//get transfer detail
		// 	transferDet, _ := models.GetEwtTransferDetailByDocNo(v.DocNo)

		// 	if transferDet != nil {
		// 		FromMember = transferDet.MemberFrom
		// 		ToMember = transferDet.MemberTo
		// 	}
		// } else if v.TransType == "BONUS" {
		// 	docNo = helpers.TransRemark(v.DocNo, s.LangCode)
		// }

		arrWalletStatementList = append(arrWalletStatementList,
			WalletTransactionResultStructV2{
				ID:              strconv.Itoa(v.ID),
				DocNo:           docNo,
				EwalletTypeName: v.EwalletTypeName,
				// TransactionType: v.TransactionType,
				TransDate: v.TransDate.Format("2006-01-02 15:04:05"),
				TransType: trans_type,
				TotalIn:   TotalIn,
				TotalOut:  TotalOut,
				// Balance:         balance,
				AdditionalMsg: v.AdditionalMsg,
				Remark:        remark,
				Status:        status,
			})
	}

	if err != nil {
		base.LogErrorLog("WalletStatementV2 -failed to get ewt detail", err.Error(), map[string]interface{}{"err": err, "data": s}, true)
		return app.ArrDataResponseList{}, err
	}

	//get wallet balance
	// arrCond := make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: s.WalletTypeCode},
	// )
	// bal, wal_bal_err := models.GetMemberEwtSetupBalanceFn(s.MemberID, arrCond, "", false)

	// if wal_bal_err != nil {
	// 	base.LogErrorLog("WalletStatementV2 -failed to get wallet balance", wal_bal_err.Error(), map[string]interface{}{"err": wal_bal_err, "data": s}, true)
	// 	return app.ArrWalletDataResponseList{}, wal_bal_err
	// }

	// walbal := fmt.Sprintf("%.2f", 0)

	// if len(bal) > 0 {
	// 	walbal = fmt.Sprintf("%.2f", bal[0].Balance)
	// }

	arrDataReturn := app.ArrDataResponseList{
		CurrentPage:           int(s.Page),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      arrWalletStatementList,
		TableHeaderList:       arrTableHeaderList,
	}

	return arrDataReturn, nil
}

type BlockchainWalletReturnStruct struct {
	Balance              float64
	ConvertedBalance     float64
	Rate                 float64
	AvailableBalance     float64
	AvailableConvBalance float64
	WithholdingBalance   float64
}

// for sec & liga this func will return converted val from price movement table
func GetBlockchainWalletBalanceByAddressV1(wallet_type string, address string, entMemberID int) BlockchainWalletReturnStruct {
	var response app.ApiResponse

	secApiSetting, _ := models.GetSysGeneralSetupByID("sec_api_setting")

	data := map[string]interface{}{
		"token_type": wallet_type,
		"address":    address,
	}

	api_key := secApiSetting.InputValue2

	url := secApiSetting.InputValue1 + "api/account/balance"
	header := map[string]string{
		"Content-Type":    "application/json",
		"X-Authorization": api_key,
	}

	res, err_api := base.RequestAPI("POST", url, header, data, &response)

	if err_api != nil {
		base.LogErrorLogV2("GetBlockchainWalletBalanceByAddressV1 -fail to call blockchain balance api", err_api.Error(), map[string]interface{}{"err": err_api, "data": data}, true, "blockchain")
		return BlockchainWalletReturnStruct{
			Rate: 1,
		}
	}

	if res.StatusCode != 200 {
		errMsg, _ := json.Marshal(response.Msg)
		errMsgStr := string(errMsg)
		errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		base.LogErrorLogV2("GetBlockchainWalletBalanceByAddressV1 -fail to get blockchain wallet balance", errMsgStr, map[string]interface{}{"err": res.Body, "data": data}, true, "blockchain")
		return BlockchainWalletReturnStruct{
			Rate: 1,
		}
	}

	if response.Data["balance"] == nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -empty balance", response.Data, res.Body, true)
		return BlockchainWalletReturnStruct{
			Rate: 1,
		}
	}

	return_balance, err := json.Marshal(response.Data["balance"])

	if err != nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -fail to process returned balance", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return BlockchainWalletReturnStruct{
			Rate: 1,
		}
	}

	strBal := string(return_balance)
	strBal = strings.Replace(strBal, "\"", "", 2)

	var decimalPoint uint
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(wallet_type)},
	)
	ewtSetup, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -fail to get wallet setup", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return BlockchainWalletReturnStruct{
			Rate: 1,
		}
	}

	if ewtSetup != nil {
		decimalPoint = uint(ewtSetup.DecimalPoint)
	}

	strBal = helpers.CutOffStringsDecimal(strBal, decimalPoint, '.')
	bal := float64(0)
	bal, err = strconv.ParseFloat(strBal, 64)

	if err != nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -fail to convert balance to float64", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return BlockchainWalletReturnStruct{
			Rate: 1,
		}
	}

	token_rate, err := base.GetLatestPriceMovementByTokenType(wallet_type)
	if err != nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -get price movement error", err, map[string]interface{}{"wallet_type_code": wallet_type}, true)
	}

	balance := bal
	availBal := bal
	rate := token_rate
	convBal, _ := decimal.NewFromFloat(bal).Mul(decimal.NewFromFloat(rate)).Float64()
	availConvBal, _ := decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
	withholdingBal := float64(0)

	//get holding wallet
	arrHoldCond := make([]models.WhereCondFn, 0)
	arrHoldCond = append(arrHoldCond,
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(wallet_type) + "H"},
	)

	HoldResult, Hold_err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrHoldCond, "", false)

	if Hold_err != nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -get Holding Wallet error", Hold_err, HoldResult, true)
	}

	if len(HoldResult) > 0 {
		balance, _ = decimal.NewFromFloat(bal).Add(decimal.NewFromFloat(HoldResult[0].Balance)).Float64()
		if balance < 0 {
			balance = float64(0)
		}
		convBal, _ = decimal.NewFromFloat(balance).Mul(decimal.NewFromFloat(rate)).Float64()
	}

	// //get withholding wallet
	// arrWithHoldCond := make([]models.WhereCondFn, 0)
	// arrWithHoldCond = append(arrWithHoldCond,
	// 	models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "W" + strings.ToUpper(wallet_type)},
	// )

	// WithHoldResult, WithHold_err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrWithHoldCond, "", false)

	// if WithHold_err != nil {
	// 	base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -get WithHolding Wallet error", err, arrWithHoldCond, true)
	// }

	// if len(WithHoldResult) > 0 {
	// 	withholdingBal = WithHoldResult[0].Balance
	// 	balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(WithHoldResult[0].Balance)).Float64()
	// 	if balance < 0 {
	// 		balance = float64(0)
	// 	}
	// 	convBal, _ = decimal.NewFromFloat(balance).Mul(decimal.NewFromFloat(rate)).Float64()
	// }

	//get blockchain_trans -pending & log_only = 0
	PendingAmt, PendingAmtError := models.GetTotalPendingBlockchainAmount(entMemberID, strings.ToUpper(wallet_type))
	if PendingAmtError != nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -get Blockchain Trans Pending Amount error", PendingAmtError, PendingAmt, true)
	}

	if PendingAmt != nil {
		availBal, _ = decimal.NewFromFloat(availBal).Sub(decimal.NewFromFloat(PendingAmt.TotalPendingAmount)).Float64()
		availConvBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
	}

	//get blockchain_adjust_out where pending
	PendingAdjOutAmt, err := models.GetTotalPendingBlockchainAdjustOutAmount(entMemberID, strings.ToUpper(wallet_type))
	if err != nil {
		base.LogErrorLog("GetBlockchainWalletBalanceByAddressV1 -get Blockchain Adjust Out Pending Amount error", err, wallet_type, true)
	}

	if PendingAdjOutAmt != nil {
		if PendingAdjOutAmt.TotalPendingAmount != 0 {
			availBal, _ = decimal.NewFromFloat(availBal).Sub(decimal.NewFromFloat(PendingAdjOutAmt.TotalPendingAmount)).Float64()
			availConvBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
		}
	}

	if availBal < 0 {
		availBal = float64(0)
		availConvBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
	}

	if availBal > balance {
		availBal = balance
		availConvBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
	}

	strBalance := helpers.CutOffDecimal(balance, decimalPoint, ".", "")
	balance, _ = strconv.ParseFloat(strBalance, 64)

	strConvBal := helpers.CutOffDecimal(convBal, decimalPoint, ".", "")
	convBal, _ = strconv.ParseFloat(strConvBal, 64)

	strAvailBal := helpers.CutOffDecimal(availBal, decimalPoint, ".", "")
	availBal, _ = strconv.ParseFloat(strAvailBal, 64)

	strAvailConvBal := helpers.CutOffDecimal(availConvBal, decimalPoint, ".", "")
	availConvBal, _ = strconv.ParseFloat(strAvailConvBal, 64)

	strWithholdingBal := helpers.CutOffDecimal(withholdingBal, decimalPoint, ".", "")
	withholdingBal, _ = strconv.ParseFloat(strWithholdingBal, 64)

	return BlockchainWalletReturnStruct{
		Balance:              balance,
		ConvertedBalance:     convBal,
		Rate:                 rate,
		AvailableBalance:     availBal,
		AvailableConvBalance: availConvBal,
		WithholdingBalance:   withholdingBal,
	}
}

// type CoinbaseApiResponseStruct map[string]float64
type CoinbaseApiResponseStruct struct {
	Data struct {
		Currency string            `json:"currency"`
		Rates    map[string]string `json:"rates"`
	} `json:"data"`
}

func GetLatestCryptoLiveRateFromTo(tx *gorm.DB, from_crypto string, to_currency string, markup float64) (float64, string) {
	var (
		rate     float64
		response CoinbaseApiResponseStruct
	)

	if to_currency == "" {
		to_currency = "USDT"
	}
	// url := "https://min-api.cryptocompare.com/data/price?fsym=" + from_crypto + "&tsyms=" + to_currency
	url := "https://api.coinbase.com/v2/exchange-rates?currency=" + from_crypto
	res, err_api := base.RequestAPI("GET", url, nil, nil, nil)

	if err_api != nil {
		base.LogErrorLog("GetLatestCryptoLiveRateFromTo_error_in_api_call_before_call", err_api.Error(), nil, true)
		return 0.00, "something_went_wrong"
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetLatestCryptoLiveRateFromTo_error_in_api_call_after_call", res.Body, nil, true)
		return 0.00, "something_went_wrong"
	}

	jsonDecodeErr := json.Unmarshal([]byte(res.Body), &response)
	if jsonDecodeErr != nil {
		base.LogErrorLog("GetLatestCryptoLiveRateFromTo_error_in_json_decode_api_result", jsonDecodeErr.Error(), res.Body, true)
		return 0.00, "something_went_wrong"
	}

	rate = 1
	for key, value := range response.Data.Rates {
		if key == to_currency {
			newRate, _ := strconv.ParseFloat(value, 64)
			rate = newRate
		}
	}

	// update previous price to not latest
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "b_latest = ?", CondValue: 1},
		models.WhereCondFn{Condition: "code = ?", CondValue: from_crypto + to_currency},
	)
	updateColumn := map[string]interface{}{"b_latest": 0}
	err := models.UpdatesFnTx(tx, "crypto_price_movement", arrUpdCond, updateColumn, false)

	if err != nil {
		base.LogErrorLog("wallet_service:GetLatestCryptoLiveRateFromTo()", "UpdatesFnTx()", err.Error(), true)
		return 0.00, "something_went_wrong"
	}

	cryptoPriceMovement := models.AddCryptoPriceMovementStruct{
		Code:    from_crypto + to_currency,
		Price:   rate,
		BLatest: 1,
	}
	models.AddCryptoPriceMovement(tx, cryptoPriceMovement)

	if markup > 0 {
		rate = rate * markup
	}
	return rate, ""
}

type PostTransferExchangeStruct struct {
	SigningKey    string
	MemberId      int
	EwtTypeCode   string
	Amount        float64
	WalletAddress string
	Remark        string
	LangCode      string
}

func (t *PostTransferExchangeStruct) PostTransferExchange(tx *gorm.DB) (interface{}, error) {
	var (
		err      error
		response app.ApiResponse
	)

	//check member wallet balance - handle & deduct balance by blockchain site

	//get wallet setting
	arrWalCond := make([]models.WhereCondFn, 0)
	arrWalCond = append(arrWalCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: t.EwtTypeCode},
	)
	walSetup, err := models.GetEwtSetupFn(arrWalCond, "", false)

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail to get wallet setup", err, arrWalCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}

	if walSetup == nil {
		base.LogErrorLog("PostTransferExchange - empty wallet setup returned", walSetup, arrWalCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}

	eWalletId := walSetup.ID

	// if t.Amount < walSetup.WithdrawMin {
	// 	strAmt := helpers.CutOffDecimal(walSetup.WithdrawMin, 2, ".", "")
	// 	// base.LogErrorLog("PostTransferExchange - minimum transfer amount is"+" "+strAmt, "ewtSetup-withdrawMin"+" "+":"+strAmt, t, true)
	// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("minimum_transfer_amount_is"+" "+strAmt, t.LangCode), Data: t}
	// }

	//check amt
	if t.Amount <= 0 {
		// base.LogErrorLog("PostTransferExchange - amount cannot negative", t.Amount, t, true) //store error log
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("amount_must_more_than_0", t.LangCode), Data: t}
	}

	//start check on member wallet lock
	wallet_lock, err := models.GetEwtLockByMemberId(t.MemberId, eWalletId)

	if wallet_lock.InternalTransfer == 1 {
		// base.LogErrorLog("PostTransferExchange - wallet is locked from being transfer", wallet_lock, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("wallet_is_locked_from_being_transfer", t.LangCode), Data: t}
	}

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail to get ewtLock Setup", err, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}
	//end check on member wallet lock

	//check to addr
	arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
		models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
		models.WhereCondFn{Condition: "ent_member_crypto.crypto_address = ?", CondValue: t.WalletAddress},
		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)
	arrEntMemberCrypto, err := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail_to_check_address_to", err, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}

	if arrEntMemberCrypto == nil {
		// base.LogErrorLog("PostTransferExchange - invalid_address_to", arrEntMemberCryptoFn, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_address_to", t.LangCode), Data: t}
	}

	memToId := arrEntMemberCrypto.MemberID

	//get member to info
	arrMemToCond := make([]models.WhereCondFn, 0)
	arrMemToCond = append(arrMemToCond,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: memToId},
	)
	memberTo, err := models.GetEntMemberFn(arrMemToCond, "", false) //get member to details

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail to get member to info", err, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}

	if memberTo == nil {
		base.LogErrorLog("PostTransferExchange - empty member to info", arrMemToCond, t, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_member_to", t.LangCode), Data: t}
	}

	//check if is same member
	if memToId == t.MemberId {
		// base.LogErrorLog("PostTransferExchange - cannot transfer to same account", memToId, t.MemberId, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("cannot_transfer_to_same_account", t.LangCode), Data: t}
	}

	//check transfer setup
	arrTransferSetupCond := make([]models.WhereCondFn, 0)
	arrTransferSetupCond = append(arrTransferSetupCond,
		models.WhereCondFn{Condition: "ewt_transfer_setup.ewallet_type_id_from = ?", CondValue: eWalletId},
		models.WhereCondFn{Condition: "ewt_transfer_setup.ewt_transfer_type = ?", CondValue: "Blockchain"},
	)

	arrTransferSetupRst, err := models.GetEwtTransferSetupFn(arrTransferSetupCond, "", false)

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail to get transfer setup", err, arrTransferSetupCond, true)
		return nil, err
	}

	if len(arrTransferSetupRst) > 0 {
		if arrTransferSetupRst[0].TransferSponsorTree == 1 {
			//check within newtwork
			memberNetw1 := member_service.CheckSponsorMember(t.MemberId, memToId) // koo func- sponsor_id, downline_id
			memberNetw2 := member_service.CheckSponsorMember(memToId, t.MemberId) // koo func- sponsor_id, downline_id

			if memberNetw1 == false && memberNetw2 == false {
				base.LogErrorLog("PostTransferExchange - not_within_network", map[string]interface{}{"memberNetw1": memberNetw1, "memberNetw2": memberNetw2}, t, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_within_network", t.LangCode), Data: t}
			}

		}
	}

	//check available bal
	// MemAddr, err := models.GetCustomMemberCryptoAddr(t.MemberId, t.EwtTypeCode, true, false)
	// if err != nil {
	// 	base.LogErrorLog("PostTransferExchange - fail_to_get_member_address", err, t, true)
	// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	// }
	// BlkCWalBal := GetBlockchainWalletBalanceByAddressV1(t.EwtTypeCode, MemAddr, t.MemberId)
	// availableBalance := BlkCWalBal.AvailableBalance

	// if t.Amount > availableBalance {
	// 	submitAmt := helpers.CutOffDecimal(t.Amount, 2, ".", "")
	// 	availableBal := helpers.CutOffDecimal(availableBalance, 2, ".", "")
	// 	base.LogErrorLog("PostTransferExchange - not_enough_balance", "amount_submitted:"+" "+submitAmt, "avail_bal:"+" "+availableBal, true)
	// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_enough_balance", t.LangCode), Data: t}
	// }

	//get WT doc no
	docs, err := models.GetRunningDocNo("WT", tx) //get transfer doc no

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail to get WT doc no", err, t, true)
		return nil, err
	}

	//call yeejia signed transaction api
	secApiSetting, _ := models.GetSysGeneralSetupByID("sec_api_setting")

	data := map[string]interface{}{
		"transaction_data": t.SigningKey,
	}

	api_key := secApiSetting.InputValue2

	url := secApiSetting.InputValue1 + "api/transaction/send/signedTransaction"
	header := map[string]string{
		"Content-Type":    "application/json",
		"X-Authorization": api_key,
	}

	res, err := base.RequestAPI("POST", url, header, data, &response)

	if err != nil {
		models.SaveEwtTransferExchangeLog(t.MemberId, t, data, err)
		base.LogErrorLogV2("PostTransferExchange - signedTransaction api call error", err.Error(), map[string]interface{}{"err": err, "data": data}, true, "blockchain")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: t}
	}

	if res.StatusCode != 200 {
		errMsg, _ := json.Marshal(response.Msg)
		errMsgStr := string(errMsg)
		errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		base.LogErrorLogV2("PostTransferExchange - signedTransaction api error returned", errMsgStr, map[string]interface{}{"err": res.Body, "data": data}, true, "blockchain")
		models.SaveEwtTransferExchangeLog(t.MemberId, t, data, res)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate(errMsgStr, t.LangCode), Data: t}
	}

	rtn_hash, err := json.Marshal(response.Data["hash"])
	if err != nil {
		base.LogErrorLog("PostTransferExchange - process returned hash failed", err, res, true)
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: ""}
	}
	return_hash := string(rtn_hash)
	return_hash = strings.Replace(return_hash, "\"", "", 2)

	if return_hash == "" {
		base.LogErrorLog("PostTransferExchange - empty signedTransactionHash", return_hash, res, true)
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: ""}
	}
	//end call transaction api

	//store blockchain_trans
	arrBlockchainTran := models.AddBlockchainTransStruct{
		MemberID:          t.MemberId,
		EwalletTypeID:     eWalletId,
		DocNo:             docs,
		Status:            "P",
		TransactionType:   "TRANSFER_TO_EXCHANGE",
		TotalOut:          t.Amount,
		ConversionRate:    float64(1),
		ConvertedTotalOut: t.Amount,
		TransactionData:   t.SigningKey,
		HashValue:         return_hash,
		Remark:            "#*transfer_to*#" + " " + memberTo.NickName,
		// Remark: t.Remark,
	}

	_, err = models.AddBlockchainTrans(tx, arrBlockchainTran) //store withdraw

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail to save blockchain_trans", err, arrBlockchainTran, true)
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: ""}
	}

	//store ewt_transfer_exchange
	arrEwtWithdraw := models.EwtTransferExchange{
		MemberId:        t.MemberId,
		DocNo:           docs,
		EwalletTypeId:   eWalletId,
		TransactionType: "TRANSFER_TO_EXCHANGE",
		Amount:          t.Amount,
		CryptoAddrTo:    t.WalletAddress,
		SigningKey:      t.SigningKey,
		TranHash:        return_hash,
		Remark:          t.Remark,
		Status:          "W",
		CreatedAt:       time.Now(),
		CreatedBy:       t.MemberId,
	}

	_, err = models.AddEwtTransferExchange(tx, arrEwtWithdraw) //store withdraw

	if err != nil {
		base.LogErrorLog("PostTransferExchange - fail to store to ewt_transfer_exchange", err, arrEwtWithdraw, true)
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: helpers.Translate("something_went_wrong", t.LangCode), Data: ""}
	}

	err = models.UpdateRunningDocNo("WT", tx) //update transfer doc no

	if err != nil {
		base.LogErrorLog("PostTransferExchange -fail to update WT doc no", err, t, true)
		return nil, err
	}

	//store transfer exchange log
	models.SaveEwtTransferExchangeLog(t.MemberId, t, data, res)

	arrData := make(map[string]interface{})
	arrData["tran_hash"] = return_hash
	arrData["ewallet_type"] = t.EwtTypeCode
	arrData["address_to"] = t.WalletAddress
	arrData["remark"] = t.Remark
	arrData["amount"] = t.Amount
	arrData["trans_time"] = time.Now().Format("2006-01-02 15:04:05")

	return arrData, nil

}

// TestProcessGetMemAddress
func TestProcessGetMemAddress(tx *gorm.DB, EntMemberID int, cryptoType string) (cryptoAddr string, err error) {
	cryptoAdd, err := member_service.ProcessGetMemAddress(tx, EntMemberID, cryptoType)
	if err != nil || cryptoAdd == "" {
		return cryptoAdd, err
	}
	return cryptoAdd, err
}

type GetDepositInfoStruct struct {
	EwtTypeCode string
	CryptoCode  string
	MemberID    int
}

type GetDepositInfoResponseStruct struct {
	WalletFrom *models.EwtSetup
	WalletTo   *models.EwtSetup
	Setting    *models.SysGeneralSetup
	Rate       float64
	CanLock    bool
}

type BlockchainDepositSettingStruct struct {
	EwtTypeCode          string `json:"ewallet_type_code"`
	EwtTypeName          string `json:"ewallet_type_name"`
	WalletTypeImageUrl   string `json:"wallet_type_image_url"`
	DispCryptoAddrStatus string `json:"disp_crypto_addr_status"`
	Lock                 int    `json:"lock"`
}

func GetDepositInfo(tx *gorm.DB, form GetDepositInfoStruct, getLiveRate bool) (GetDepositInfoResponseStruct, string) {
	var (
		response     GetDepositInfoResponseStruct
		bcSettingArr []BlockchainDepositSettingStruct
		// walletFrom BlockchainDepositSettingStruct
		Rate float64 = 1
	)

	if form.EwtTypeCode == "" {
		form.EwtTypeCode = "USDT"
	}

	// get wallet from
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: form.CryptoCode},
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
	)
	walletFrom, walFromErr := models.GetEwtSetupFn(arrCond, "", false)

	if walFromErr != nil {
		base.LogErrorLog("GetDepositInfo-GetEwtSetupFn_from_failed", walFromErr.Error(), arrCond, true)
		return response, "something_went_wrong"
	}

	if walletFrom == nil {
		base.LogErrorLog("wallet_service:GetDepositInfo()_from", "GetEwtSetupFn()_from", "wallet_is_not_available_to_perform_crypto_deposit", true)
		return response, "wallet_is_not_available_to_perform_crypto_deposit"
	}

	//get wallet to
	walletToArrCond := make([]models.WhereCondFn, 0)
	walletToArrCond = append(walletToArrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: form.EwtTypeCode},
		models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ewt_setup.blockchain_deposit_setting != ?", CondValue: ""},
	)
	walletTo, walToErr := models.GetEwtSetupFn(walletToArrCond, "", false)

	if walletTo == nil {
		// base.LogErrorLog("wallet_service:GetDepositInfo()_to", "GetEwtSetupFn()_to", "wallet_is_not_available_to_perform_crypto_deposit", true)
		// return response, "wallet_is_not_available_to_perform_crypto_deposit"
		return response, "deposit_is_closed"
	}

	if walToErr != nil {
		base.LogErrorLog("GetDepositInfo-GetEwtSetupFn_to_failed", walToErr.Error(), walletToArrCond, true)
		return response, "something_went_wrong"
	}

	json.Unmarshal([]byte(walletTo.BlockchainDepositSetting), &bcSettingArr)

	canLock := true
	for _, wallet := range bcSettingArr {
		if wallet.EwtTypeCode == form.CryptoCode && wallet.Lock == 0 {
			canLock = false
		}
	}

	// json decode BlockchainDepositSetting
	// json.Unmarshal([]byte(wallet.BlockchainDepositSetting), &arrDataReturn.Object)

	// get crypto live rate
	if getLiveRate && canLock {
		rate, liveRateErrMsg := GetLatestCryptoLiveRateFromTo(tx, form.CryptoCode, "", 1)

		if liveRateErrMsg != "" {
			return response, liveRateErrMsg
		}
		Rate = rate
	}

	// get deposit setup
	setting, generalSetupErr := models.GetSysGeneralSetupByID("crypto_purchase_setup")

	if generalSetupErr != nil {
		base.LogErrorLog("wallet_service:GetSysGeneralSetupByID()", generalSetupErr.Error(), "crypto_purchase_setup", true)
		return response, "something_went_wrong"
	}

	arrDataReturn := GetDepositInfoResponseStruct{
		WalletFrom: walletFrom,
		WalletTo:   walletTo,
		Setting:    setting,
		Rate:       Rate,
		CanLock:    canLock,
	}

	return arrDataReturn, ""
}

type CryptoPurchaseStruct struct {
	Amount        float64
	CryptoCode    string
	MemberID      int
	WalletFrom    *models.EwtSetup
	WalletTo      *models.EwtSetup
	Rate          float64
	CryptoAddress string
}

func AddCryptoPurchase(tx *gorm.DB, data CryptoPurchaseStruct) (*models.EwtTopupStruct, string) {
	expiryMin := 30
	expiryAt := base.GetCurrentDateTimeT().Add(time.Minute * time.Duration(expiryMin))
	totalAmountIn, _ := decimal.NewFromFloat(data.Amount).Div(decimal.NewFromFloat(data.Rate)).Float64()

	decimalLength := helpers.NumDecPlaces(totalAmountIn) //count decimal point

	if decimalLength < 2 {
		decimalLength = 2
	}

	if decimalLength > 8 {
		decimalLength = 8
	}

	totalAmountIn = float.RoundDown(totalAmountIn, decimalLength)

	runningDocNo, _ := models.GetRunningDocNo("WCT", tx)

	EwtTopupData := models.EwtTopupStruct{
		MemberID:      data.MemberID,
		EwalletTypeID: data.WalletTo.ID,
		DocNo:         runningDocNo,
		Status:        "W",
		Type:          "CRYPTO",
		CurrencyCode:  data.WalletTo.CurrencyCode,
		TransDate:     base.GetCurrentDateTimeT(),
		TotalIn:       data.Amount,
		// FromAddr:              data.CryptoAddress,
		Charges:               0,
		ConvertedCurrencyCode: data.WalletFrom.CurrencyCode,
		ConversionRate:        data.Rate,
		ConvertedTotalAmount:  totalAmountIn,
		Remark:                fmt.Sprint("#*", data.WalletFrom.EwtTypeCode, "*# = ", totalAmountIn),
		CreatedBy:             data.MemberID,
		ExpiryAt:              expiryAt,
	}

	// add crypto purchase
	ewtTopup, ewtTopupErr := models.AddEwtTopup(tx, EwtTopupData)

	if ewtTopupErr != nil {
		base.LogErrorLog("wallet_service:AddCryptoPurchase()", "AddEwtTopup()", ewtTopupErr.Error(), true)
		return ewtTopup, "something_went_wrong"
	}

	updateDocNoErr := models.UpdateRunningDocNo("WCT", tx)

	if updateDocNoErr != nil {
		base.LogErrorLog("wallet_service:AddCryptoPurchase()", "UpdateRunningDocNo()", updateDocNoErr.Error(), true)
		return ewtTopup, "something_went_wrong"
	}

	return ewtTopup, ""
}

type CheckCryptoPurchaseStruct struct {
	MemberID   int
	WalletFrom *models.EwtSetup
	WalletTo   *models.EwtSetup
}

func CheckCryptoPurchase(tx *gorm.DB, data CheckCryptoPurchaseStruct) (*models.EwtTopupStruct, string) {
	var EwtTopup *models.EwtTopupStruct

	//check setting
	arrTopupCond := make([]models.WhereCondFn, 0)
	arrTopupCond = append(arrTopupCond,
		models.WhereCondFn{Condition: "member_id = ?", CondValue: data.MemberID},
		models.WhereCondFn{Condition: "converted_currency_code = ?", CondValue: data.WalletFrom.CurrencyCode},
		models.WhereCondFn{Condition: "status = ?", CondValue: "W"},
		models.WhereCondFn{Condition: "ewallet_type_id = ?", CondValue: data.WalletTo.ID},
	)

	EwtTopup, _ = models.GetEwtTopupFn(arrTopupCond, "", false)

	// if no record found, return true
	if EwtTopup == nil {
		// base.LogErrorLog("wallet_service:CheckCryptoPurchase()", "GetEwtTopupFn()", topupErr.Error(), true)
		return EwtTopup, ""
	}

	// if expired
	if EwtTopup.ID > 0 && EwtTopup.ExpiryAt.Before(base.GetCurrentDateTimeT()) {
		// update expired row status to CNL
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "id = ?", CondValue: EwtTopup.ID},
		)
		updateColumn := map[string]interface{}{"status": "CNL"}
		err := models.UpdatesFnTx(tx, "ewt_topup", arrUpdCond, updateColumn, false)

		if err != nil {
			base.LogErrorLog("wallet_service:CheckCryptoPurchase()", "UpdatesFnTx()", err.Error(), true)
			return nil, "something_went_wrong"
		}

		return nil, ""
	}

	return EwtTopup, ""
}

type CancelCryptoPurchaseStruct struct {
	MemberID   int
	WalletFrom *models.EwtSetup
	WalletTo   *models.EwtSetup
}

func CancelCryptoPurchase(tx *gorm.DB, data CancelCryptoPurchaseStruct) string {
	//check setting
	arrTopupCond := make([]models.WhereCondFn, 0)
	arrTopupCond = append(arrTopupCond,
		models.WhereCondFn{Condition: "type = ?", CondValue: "CRYPTO"},
		models.WhereCondFn{Condition: "member_id = ?", CondValue: data.MemberID},
		models.WhereCondFn{Condition: "converted_currency_code = ?", CondValue: data.WalletFrom.CurrencyCode},
		models.WhereCondFn{Condition: "status = ?", CondValue: "W"},
		models.WhereCondFn{Condition: "ewallet_type_id = ?", CondValue: data.WalletTo.ID},
	)

	EwtTopup, _ := models.GetEwtTopupFn(arrTopupCond, "", false)

	// if exist
	if EwtTopup != nil {
		// update expired row status to CNL
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "id = ?", CondValue: EwtTopup.ID},
		)
		updateColumn := map[string]interface{}{"status": "CNL", "updated_by": data.MemberID}
		err := models.UpdatesFnTx(tx, "ewt_topup", arrUpdCond, updateColumn, false)

		if err != nil {
			base.LogErrorLog("wallet_service:CheckCryptoPurchase()", "UpdatesFnTx()", err.Error(), true)
			return "something_went_wrong"
		}
	}

	return ""
}

// SignedTransaction func
func SignedTransaction(transactionData string) (string, string) {
	var (
		response       app.ApiResponse
		domain, apiKey string
		err            error
	)

	secAPISetting, err := models.GetSysGeneralSetupByID("sec_api_setting")

	if err != nil {
		base.LogErrorLog("walletService:SignedTransaction()", "GetSysGeneralSetupByID():1", err.Error(), true)
		return "", "something_went_wrong"
	}
	if secAPISetting == nil {
		base.LogErrorLog("walletService:SignedTransaction()", "GetSysGeneralSetupByID():1", "sec_api_setting_not_found", true)
		return "", "something_went_wrong"
	}

	domain = secAPISetting.InputValue1
	apiKey = secAPISetting.InputValue2

	url := domain + "api/transaction/send/signedTransaction"
	header := map[string]string{
		"Content-Type":    "application/json",
		"X-Authorization": apiKey,
	}

	data := map[string]interface{}{"transaction_data": transactionData}

	res, err := base.RequestAPI("POST", url, header, data, &response)
	if err != nil { // api failed
		base.LogErrorLogV2("walletService:SignedTransaction()", "RequestAPI():1", map[string]interface{}{"err": err, "data": data}, true, "blockchain")
		return "", "something_went_wrong"
	}

	if res.StatusCode != 200 {
		errMsg, _ := json.Marshal(response.Msg)
		errMsgStr := string(errMsg)
		errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)

		base.LogErrorLogV2("walletService:SignedTransaction()", "RequestAPI():1", map[string]interface{}{"err": errMsgStr, "response_body": res.Body}, true, "blockchain")
		return "", "something_went_wrong"
	}

	returnHashValue, _ := json.Marshal(response.Data["hash"])
	returnHash := string(returnHashValue)
	returnHash = strings.Replace(returnHash, "\"", "", 2)

	return returnHash, ""
}

// GetTransactionNonceViaAPI func
func GetTransactionNonceViaAPI(cryptoAddr string) (int, error) {

	settingID := "nonce_api_setting"
	arrApiSetting, _ := models.GetSysGeneralSetupByID(settingID)

	if arrApiSetting.InputType1 == "1" {

		type apiRstStruct struct {
			Status     string `json:"status"`
			StatusCode string `json:"status_code"`
			Msg        string `json:"msg"`
			Data       struct {
				Nonce int `json:"nonce"`
			} `json:"data"`
		}

		header := map[string]string{
			"Content-Type":    "application/json",
			"X-Authorization": arrApiSetting.InputType2,
		}
		data := map[string]interface{}{
			"address": cryptoAddr,
		}
		res, err_api := base.RequestAPI(arrApiSetting.SettingValue1, arrApiSetting.InputValue1, header, data, nil)

		if err_api != nil {
			base.LogErrorLogV2("GetTransactionNonceViaAPI-error_in_api_call_before_call", err_api.Error(), nil, true, "blockchain")
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "nonce_is_not_available", Data: ""}
		}

		if res.StatusCode != 200 {
			base.LogErrorLogV2("GetTransactionNonceViaAPI-error_in_api_call_after_call", res.Body, nil, true, "blockchain")
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "nonce_is_not_available", Data: ""}
		}

		var apiRst apiRstStruct
		err := json.Unmarshal([]byte(res.Body), &apiRst)

		if err != nil {
			base.LogErrorLog("GetTransactionNonceViaAPI-error_in_json_decode_api_result", err_api.Error(), res.Body, true)
			return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "nonce_is_not_available", Data: ""}
		}
		return apiRst.Data.Nonce, nil

	}

	return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "nonce_is_not_available", Data: ""}
}

// GetCompanyAddress func
func GetCompanyAddress(cryptoType string) (string, string) {
	cryptoAddr, err := models.GetCustomMemberCryptoAddr(0, cryptoType, true, false)
	if err != nil {
		return "", err.Error()
	}

	return cryptoAddr, ""
}

func (s *WalletTransactionStructV2) WalletSummaryDetail() (app.ArrDataResponseList, error) {

	type WalletSummaryDetail struct {
		TransDate string `json:"trans_date"`
		Balance   string `json:"balance"`
	}

	ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, err := models.GetEwtDetailByWalletTypeForSummaryDetail(s.Page, s.MemberID, s.TransType, s.DateFrom, s.DateTo, s.WalletTypeCode)

	if s.Page == 0 {
		s.Page = 1
	}

	var (
		decimalPoint uint
	)

	arrWalletStatementList := make([]WalletSummaryDetail, 0)

	decimalPoint = 2
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: s.WalletTypeCode},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup != nil {
		decimalPoint = uint(ewtSetup.DecimalPoint)
	}

	for _, v := range ewt {
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ewt_detail.id = ? ", CondValue: v.ID},
		)
		arrEwtDetail, _ := models.GetEwtDetailFn(arrCond, false)

		if len(arrEwtDetail) > 0 {
			bal := helpers.CutOffDecimal(arrEwtDetail[0].Balance, decimalPoint, ".", ",")

			arrWalletStatementList = append(arrWalletStatementList,
				WalletSummaryDetail{
					TransDate: arrEwtDetail[0].TransDate.Format("2006-01-02"),
					Balance:   bal,
				})
		}
	}

	if err != nil {
		base.LogErrorLog("WalletSummaryDetail -failed to get record", err.Error(), map[string]interface{}{"err": err, "data": s}, true)
		return app.ArrDataResponseList{}, err
	}

	arrDataReturn := app.ArrDataResponseList{
		CurrentPage:           int(s.Page),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      arrWalletStatementList,
		TableHeaderList:       nil,
	}

	return arrDataReturn, nil
}

type WalletTransactionResultStructV3 struct {
	// ID                    string `json:"id"`
	TransDate             string `json:"trans_date"`
	TransType             string `json:"trans_type"`
	Amount                string `json:"amount"`
	CurrencyCode          string `json:"currency_code"`
	ConvertedAmount       string `json:"converted_amount"`
	ConvertedCurrencyCode string `json:"converted_currency_code"`
	Status                string `json:"status"`
	StatusColorCode       string `json:"status_color_code"`
}

type TransferTransaction struct {
	TransDate       string `json:"trans_date"`
	EwalletTypeName string `json:"ewallet_type_name"`
	DocNo           string `json:"doc_no"`
	Amount          string `json:"amount"`
	CryptoAddressTo string `json:"crypto_address_to"`
	Remark          string `json:"remark"`
	Status          string `json:"status"`
	StatusColorCode string `json:"status_color_code"`
}

type ExchangeTransaction struct {
	TransDate       string `json:"trans_date"`
	EwalletTypeName string `json:"ewallet_type_name"`
	DocNo           string `json:"doc_no"`
	Amount          string `json:"amount"`
	Payment         string `json:"payment"`
	Hash            string `json:"hash"`
	Remark          string `json:"remark"`
	Status          string `json:"status"`
	StatusColorCode string `json:"status_color_code"`
}

type ContractTransaction struct {
	TransDate       string `json:"trans_date"`
	EwalletTypeName string `json:"ewallet_type_name"`
	DocNo           string `json:"doc_no"`
	Amount          string `json:"amount"`
	Status          string `json:"status"`
	StatusColorCode string `json:"status_color_code"`
}

type WithdrawTransaction struct {
	TransDate       string `json:"trans_date"`
	EwalletTypeName string `json:"ewallet_type_name"`
	Address         string `json:"address"`
	DocNo           string `json:"doc_no"`
	Amount          string `json:"amount"`
	Payment         string `json:"payment"`
	Status          string `json:"status"`
	StatusColorCode string `json:"status_color_code"`
}

func (s *WalletTransactionStructV2) WalletStatementV3() (interface{}, error) {

	var (
		decimalPoint uint
		// decimalPointPay uint
	)

	arrWalletStatementList := make([]WalletTransactionResultStructV3, 0)
	arrWalletModuleStatementList := make([]interface{}, 0)
	status := helpers.Translate("completed", s.LangCode)
	statusColorCode := "#00A01F"

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: s.WalletTypeCode},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup != nil {
		decimalPoint = uint(ewtSetup.DecimalPoint)
	}

	switch strings.ToLower(s.TransType) {
	case "receive": //deposit
		if ewtSetup != nil {
			arrTopupCond := make([]models.WhereCondFn, 0)
			arrTopupCond = append(arrTopupCond,
				models.WhereCondFn{Condition: "a.member_id = ?", CondValue: s.MemberID},
				models.WhereCondFn{Condition: "a.status = ?", CondValue: "AP"},
				models.WhereCondFn{Condition: "a.ewallet_type_id = ?", CondValue: ewtSetup.ID},
			)

			EwtTopup, _ := models.GetEwtTopupArrayFn(arrTopupCond, false)

			if len(EwtTopup) > 0 {
				for _, v := range EwtTopup {
					// var decimalPointConv uint
					// arrConvCond := make([]models.WhereCondFn, 0)
					// arrConvCond = append(arrConvCond,
					// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: v.ConvertedCurrencyCode},
					// )
					// ewtSetupConv, _ := models.GetEwtSetupFn(arrConvCond, "", false)
					// if ewtSetupConv != nil {
					// 	decimalPointConv = uint(ewtSetupConv.DecimalPoint)
					// }

					if v.StatusDesc == "Approved" {
						v.StatusDesc = "Completed"
					}

					status = helpers.Translate(v.StatusDesc, s.LangCode)

					amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
					if v.TotalIn > 0 {
						amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
					}

					if ewtSetup.Control == "BLOCKCHAIN" {
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.ConvertedTotalAmount, decimalPoint, ".", ",")
							amount = "+" + amount
						}
					}

					remark := v.Remark

					if remark != "" {
						remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
					}

					arrWalletStatementList = append(arrWalletStatementList,
						WalletTransactionResultStructV3{
							TransDate:    v.TransDate.Format("2006-01-02 15:04:05"),
							TransType:    helpers.Translate(strings.ToLower(s.TransType), s.LangCode) + remark,
							Amount:       amount,
							CurrencyCode: v.CurrencyCode,
							// ConvertedAmount:       helpers.CutOffDecimal(v.ConvertedTotalAmount, decimalPointConv, ".", ","),
							ConvertedCurrencyCode: v.ConvertedCurrencyCode,
							Status:                status,
							StatusColorCode:       statusColorCode,
						})
				}
			}
		} else {
			// for module transaction history
			arrTopupCond := make([]models.WhereCondFn, 0)
			arrTopupCond = append(arrTopupCond,
				models.WhereCondFn{Condition: "a.member_id = ?", CondValue: s.MemberID},
				models.WhereCondFn{Condition: "a.status = ?", CondValue: "AP"},
				models.WhereCondFn{Condition: "a.ewallet_type_id = ?", CondValue: ewtSetup.ID},
			)

			EwtTopup, _ := models.GetEwtTopupArrayFn(arrTopupCond, false)

			if len(EwtTopup) > 0 {
				for _, v := range EwtTopup {
					var decimalPointConv uint
					arrConvCond := make([]models.WhereCondFn, 0)
					arrConvCond = append(arrConvCond,
						models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: v.ConvertedCurrencyCode},
					)
					ewtSetupConv, _ := models.GetEwtSetupFn(arrConvCond, "", false)
					if ewtSetupConv != nil {
						decimalPointConv = uint(ewtSetupConv.DecimalPoint)
					}

					if v.StatusDesc == "Approved" {
						v.StatusDesc = "Completed"
					}
					status = helpers.Translate(v.StatusDesc, s.LangCode)

					amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
					if v.TotalIn > 0 {
						amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
						amount = "+" + amount
					}

					if ewtSetup.Control == "BLOCKCHAIN" {
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.ConvertedTotalAmount, decimalPoint, ".", ",")
						}
					}

					remark := v.Remark

					if remark != "" {
						remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
					}

					arrWalletStatementList = append(arrWalletStatementList,
						WalletTransactionResultStructV3{
							TransDate:             v.TransDate.Format("2006-01-02 15:04:05"),
							TransType:             helpers.Translate(strings.ToLower(s.TransType), s.LangCode) + remark,
							Amount:                amount,
							CurrencyCode:          v.CurrencyCode,
							ConvertedAmount:       helpers.CutOffDecimal(v.ConvertedTotalAmount, decimalPointConv, ".", ","),
							ConvertedCurrencyCode: v.ConvertedCurrencyCode,
							Status:                status,
							StatusColorCode:       statusColorCode,
						})
				}
			}
		}

	case "transfer":
		if ewtSetup != nil {
			if ewtSetup.Control == "BLOCKCHAIN" {
				arrBlockCond := make([]models.WhereCondFn, 0)
				arrBlockCond = append(arrBlockCond,
					models.WhereCondFn{Condition: "blockchain_trans.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "TRANSFER_TO_EXCHANGE"},
				)
				BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

				if len(BlockchainTrans) > 0 {
					for _, v := range BlockchainTrans {
						transType := v.TransactionType

						if transType == "TRANSFER_TO_EXCHANGE" {
							transType = "TRANSFER"
						}

						remark := v.Remark
						if v.TotalIn > 0 {
							if remark != "" {
								remark = "-" + v.Remark
							}
						} else {
							if remark != "" {
								remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
							}
						}

						TransferExcDet, _ := models.GetEwtTransferExchangeDetailByDocNo(v.DocNo)

						if TransferExcDet != nil {
							status = helpers.Translate(TransferExcDet.StatusDesc, s.LangCode)
							if TransferExcDet.Status == "W" {
								status = helpers.Translate("pending", s.LangCode)
							}

							if TransferExcDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}

							if TransferExcDet.Status == "R" || TransferExcDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if TransferExcDet.Status == "P" || TransferExcDet.Status == "W" {
								statusColorCode = "#DBA000"
							} else if TransferExcDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			} else {
				//internal transfer
				arrEwtDetCond := make([]models.WhereCondFn, 0)
				arrEwtDetCond = append(arrEwtDetCond,
					models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "TRANSFER"},
				)
				EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

				if len(EwtDet) > 0 {
					for _, v := range EwtDet {
						transferDet, _ := models.GetEwtTransferDetailByDocNo(v.DocNo)

						if transferDet != nil {
							status = helpers.Translate(transferDet.StatusDesc, s.LangCode)
							if transferDet.Status == "W" {
								status = helpers.Translate("pending", s.LangCode)
							}

							if transferDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}

							if transferDet.Status == "R" || transferDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if transferDet.Status == "P" || transferDet.Status == "W" {
								statusColorCode = "#DBA000"
							} else if transferDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(v.TransactionType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			}

		} else {
			//for module transaction history
			arrBlockCond := make([]models.WhereCondFn, 0)
			arrBlockCond = append(arrBlockCond,
				models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
				models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "TRANSFER_TO_EXCHANGE"},
			)
			BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

			if len(BlockchainTrans) > 0 {
				for _, v := range BlockchainTrans {
					transType := v.TransactionType

					if transType == "TRANSFER_TO_EXCHANGE" {
						transType = "TRANSFER"
					}

					TransferExcDet, _ := models.GetEwtTransferExchangeDetailByDocNo(v.DocNo)

					if TransferExcDet != nil {
						status = helpers.Translate(TransferExcDet.StatusDesc, s.LangCode)
						if TransferExcDet.Status == "W" {
							status = helpers.Translate("pending", s.LangCode)
						}

						if TransferExcDet.Status == "AP" {
							status = helpers.Translate("completed", s.LangCode)
						}

						if TransferExcDet.Status == "R" || TransferExcDet.Status == "F" {
							statusColorCode = "#FD4343"
						} else if TransferExcDet.Status == "P" {
							statusColorCode = "#DBA000"
						} else if TransferExcDet.Status == "V" {
							statusColorCode = "#FD4343"
						} else {
							statusColorCode = "#00A01F"
						}

						arrCond := make([]models.WhereCondFn, 0)
						arrCond = append(arrCond,
							models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: TransferExcDet.EwalletTypeId},
						)
						ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
						if ewtSetup != nil {
							decimalPoint = uint(ewtSetup.DecimalPoint)
						} else {
							decimalPoint = uint(2)
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						arrWalletModuleStatementList = append(arrWalletModuleStatementList,
							TransferTransaction{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								EwalletTypeName: helpers.Translate(ewtSetup.EwtTypeName, s.LangCode),
								DocNo:           TransferExcDet.DocNo,
								Amount:          amount,
								CryptoAddressTo: TransferExcDet.CryptoAddrTo,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}

				}
			}

			//internal transfer
			arrEwtDetCond := make([]models.WhereCondFn, 0)
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
				models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "TRANSFER"},
			)
			EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

			if len(EwtDet) > 0 {
				for _, v := range EwtDet {
					transferDet, _ := models.GetEwtTransferDetailByDocNo(v.DocNo)

					if transferDet != nil {
						status = helpers.Translate(transferDet.StatusDesc, s.LangCode)
						if transferDet.Status == "W" {
							status = helpers.Translate("pending", s.LangCode)
						}

						if transferDet.Status == "AP" {
							status = helpers.Translate("completed", s.LangCode)
						}

						if transferDet.Status == "R" || transferDet.Status == "F" {
							statusColorCode = "#FD4343"
						} else if transferDet.Status == "P" || transferDet.Status == "W" {
							statusColorCode = "#DBA000"
						} else if transferDet.Status == "V" {
							statusColorCode = "#FD4343"
						} else {
							statusColorCode = "#00A01F"
						}
					}

					arrCond := make([]models.WhereCondFn, 0)
					arrCond = append(arrCond,
						models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: transferDet.EwtTypeFrom},
					)
					ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
					if ewtSetup != nil {
						decimalPoint = uint(ewtSetup.DecimalPoint)
					} else {
						decimalPoint = uint(2)
					}

					amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
					if v.TotalIn > 0 {
						amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
						amount = "+" + amount
					}

					if v.TotalOut > 0 {
						amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
						amount = "-" + amount
					}

					arrWalletModuleStatementList = append(arrWalletModuleStatementList,
						TransferTransaction{
							TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
							EwalletTypeName: helpers.Translate(ewtSetup.EwtTypeName, s.LangCode),
							DocNo:           transferDet.DocNo,
							Amount:          amount,
							CryptoAddressTo: transferDet.CryptoAddrTo,
							Status:          status,
							StatusColorCode: statusColorCode,
						})
				}
			}
		}

	// case "exchange":
	// 	if ewtSetup != nil {
	// 		if ewtSetup.Control == "BLOCKCHAIN" {
	// 			arrBlockCond := make([]models.WhereCondFn, 0)
	// 			arrBlockCond = append(arrBlockCond,
	// 				models.WhereCondFn{Condition: "blockchain_trans.ewallet_type_id = ?", CondValue: ewtSetup.ID},
	// 				models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
	// 				models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "EXCHANGE"},
	// 			)
	// 			BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

	// 			if len(BlockchainTrans) > 0 {
	// 				for _, v := range BlockchainTrans {
	// 					arrEwtExcCond := make([]models.WhereCondFn, 0)
	// 					arrEwtExcCond = append(arrEwtExcCond,
	// 						models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: v.DocNo},
	// 					)
	// 					EwtExcDet, _ := models.GetEwtExchange(arrEwtExcCond, "", false)

	// 					if EwtExcDet != nil {
	// 						status = helpers.Translate(EwtExcDet.Status, s.LangCode)

	// 						if EwtExcDet.Status == "PAID" {
	// 							status = helpers.Translate("completed", s.LangCode)
	// 						}

	// 						if EwtExcDet.Status == "REJECT" || EwtExcDet.Status == "FAILED" {
	// 							statusColorCode = "#FD4343"
	// 						} else if EwtExcDet.Status == "PENDING" {
	// 							statusColorCode = "#DBA000"
	// 						} else if EwtExcDet.Status == "VOID" {
	// 							statusColorCode = "#FD4343"
	// 						} else {
	// 							statusColorCode = "#00A01F"
	// 						}

	// 					}

	// 					amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
	// 					if v.TotalIn > 0 {
	// 						amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
	// 						amount = "+" + amount
	// 					}

	// 					if v.TotalOut > 0 {
	// 						amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
	// 						amount = "-" + amount
	// 					}

	// 					remark := v.Remark

	// 					if remark != "" {
	// 						remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
	// 					}

	// 					arrWalletStatementList = append(arrWalletStatementList,
	// 						WalletTransactionResultStructV3{
	// 							TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
	// 							TransType:       helpers.Translate(v.TransactionType, s.LangCode) + remark,
	// 							Amount:          amount,
	// 							Status:          status,
	// 							StatusColorCode: statusColorCode,
	// 						})

	// 				}
	// 			}
	// 		} else {
	// 			//for internal exchange
	// 			arrEwtDetCond := make([]models.WhereCondFn, 0)
	// 			arrEwtDetCond = append(arrEwtDetCond,
	// 				models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
	// 				models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
	// 				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "EXCHANGE"},
	// 			)
	// 			EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

	// 			if len(EwtDet) > 0 {
	// 				for _, v := range EwtDet {

	// 					amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")

	// 					if v.TotalIn > 0 {
	// 						amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
	// 						amount = "+" + amount
	// 					}

	// 					if v.TotalOut > 0 {
	// 						amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
	// 						amount = "-" + amount
	// 					}

	// 					arrEwtExcCond := make([]models.WhereCondFn, 0)
	// 					arrEwtExcCond = append(arrEwtExcCond,
	// 						models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: v.DocNo},
	// 					)
	// 					EwtExcDet, _ := models.GetEwtExchange(arrEwtExcCond, "", false)

	// 					if EwtExcDet != nil {
	// 						status = helpers.Translate(EwtExcDet.Status, s.LangCode)

	// 						if EwtExcDet.Status == "PAID" {
	// 							status = helpers.Translate("completed", s.LangCode)
	// 						}

	// 						if EwtExcDet.Status == "REJECT" || EwtExcDet.Status == "FAILED" {
	// 							statusColorCode = "#FD4343"
	// 						} else if EwtExcDet.Status == "PENDING" {
	// 							statusColorCode = "#DBA000"
	// 						} else if EwtExcDet.Status == "VOID" {
	// 							statusColorCode = "#FD4343"
	// 						} else {
	// 							statusColorCode = "#00A01F"
	// 						}

	// 						remark := v.Remark

	// 						if remark != "" {
	// 							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
	// 						}

	// 						arrWalletStatementList = append(arrWalletStatementList,
	// 							WalletTransactionResultStructV3{
	// 								TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
	// 								TransType:       helpers.Translate(v.TransactionType, s.LangCode) + remark,
	// 								Amount:          amount,
	// 								Status:          status,
	// 								StatusColorCode: statusColorCode,
	// 							})

	// 					}
	// 				}
	// 			}
	// 		}
	// 	} else {
	// 		//for module transaction history
	// 		type ExchangeData struct {
	// 			MemberId      int
	// 			EwalletTypeId int
	// 			TransDate     time.Time
	// 			DocNo         string
	// 			TotalIn       float64
	// 			TotalOut      float64
	// 			Remark        string
	// 		}

	// 		// //for blockchain exchange
	// 		// arrBlockCond := make([]models.WhereCondFn, 0)
	// 		// arrBlockCond = append(arrBlockCond,
	// 		// 	models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
	// 		// 	models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "EXCHANGE"},
	// 		// )
	// 		// BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

	// 		// if len(BlockchainTrans) > 0 {
	// 		// 	arrBlckExchangeData := make(map[string][]ExchangeData)

	// 		// 	for _, v := range BlockchainTrans {
	// 		// 		arrBlckExchangeData[v.DocNo] = append(arrBlckExchangeData[v.DocNo],
	// 		// 			ExchangeData{
	// 		// 				MemberId:      v.MemberID,
	// 		// 				EwalletTypeId: v.EwalletTypeID,
	// 		// 				TransDate:     v.DtTimestamp,
	// 		// 				DocNo:         v.DocNo,
	// 		// 				TotalIn:       v.TotalIn,
	// 		// 				TotalOut:      v.TotalOut,
	// 		// 				Remark:        v.Remark,
	// 		// 			})
	// 		// 	}

	// 		// 	for _, v2 := range arrBlckExchangeData {
	// 		// 		amount := helpers.CutOffDecimal(float64(0), 2, ".", ",")
	// 		// 		var paymentAmount string
	// 		// 		var docNo string
	// 		// 		var transDate string
	// 		// 		var ewalletTypeName string
	// 		// 		var remark string
	// 		// 		for _, v3 := range v2 {
	// 		// 			paymentAmt := helpers.CutOffDecimal(float64(0), 2, ".", ",")
	// 		// 			arrCond := make([]models.WhereCondFn, 0)
	// 		// 			arrCond = append(arrCond,
	// 		// 				models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: v3.EwalletTypeId},
	// 		// 			)
	// 		// 			ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
	// 		// 			if ewtSetup != nil {
	// 		// 				decimalPoint = uint(ewtSetup.DecimalPoint)
	// 		// 			} else {
	// 		// 				decimalPoint = uint(2)
	// 		// 			}

	// 		// 			if v3.TotalOut > 0 {
	// 		// 				paymentAmt = helpers.CutOffDecimal(v3.TotalOut, decimalPoint, ".", ",")
	// 		// 				paymentAmount += paymentAmt + " " + helpers.Translate(ewtSetup.EwtTypeName, s.LangCode) + ","
	// 		// 			}

	// 		// 			transDate = v3.TransDate.Format("2006-01-02 15:04:05")
	// 		// 			docNo = v3.DocNo
	// 		// 			remark = helpers.TransRemark(v3.Remark, s.LangCode)
	// 		// 		}

	// 		// 		//get ewt_detail record
	// 		// 		arrEwtDetCond := make([]models.WhereCondFn, 0)
	// 		// 		arrEwtDetCond = append(arrEwtDetCond,
	// 		// 			models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
	// 		// 			models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: docNo},
	// 		// 		)
	// 		// 		EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)
	// 		// 		if len(EwtDet) > 0 {
	// 		// 			arrPayCond := make([]models.WhereCondFn, 0)
	// 		// 			arrPayCond = append(arrPayCond,
	// 		// 				models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: EwtDet[0].EwalletTypeID},
	// 		// 			)
	// 		// 			ewtSetupPay, _ := models.GetEwtSetupFn(arrPayCond, "", false)
	// 		// 			if ewtSetupPay != nil {
	// 		// 				decimalPointPay = uint(ewtSetupPay.DecimalPoint)
	// 		// 			} else {
	// 		// 				decimalPointPay = uint(2)
	// 		// 			}
	// 		// 			paymentInterAmt := helpers.CutOffDecimal(EwtDet[0].TotalOut, decimalPointPay, ".", ",")
	// 		// 			paymentAmount = paymentInterAmt + " " + helpers.Translate(ewtSetupPay.EwtTypeName, s.LangCode) + "," + paymentAmount

	// 		// 			if len(EwtDet) > 1 {
	// 		// 				decimalPointPay2 := uint(2)
	// 		// 				arrPayCond2 := make([]models.WhereCondFn, 0)
	// 		// 				arrPayCond2 = append(arrPayCond2,
	// 		// 					models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: EwtDet[1].EwalletTypeID},
	// 		// 				)
	// 		// 				ewtSetupPay2, _ := models.GetEwtSetupFn(arrPayCond2, "", false)
	// 		// 				if ewtSetupPay2 != nil {
	// 		// 					decimalPointPay2 = uint(ewtSetupPay2.DecimalPoint)
	// 		// 				}
	// 		// 				paymentInterAmt2 := helpers.CutOffDecimal(EwtDet[1].TotalOut, decimalPointPay2, ".", ",")
	// 		// 				paymentAmount = paymentAmount + paymentInterAmt2 + " " + helpers.Translate(ewtSetupPay2.EwtTypeName, s.LangCode)
	// 		// 			}

	// 		// 		}

	// 		// 		paymentAmount = strings.TrimSuffix(paymentAmount, ",")

	// 		// 		arrEwtExcCond := make([]models.WhereCondFn, 0)
	// 		// 		arrEwtExcCond = append(arrEwtExcCond,
	// 		// 			models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: docNo},
	// 		// 		)
	// 		// 		EwtExcDet, _ := models.GetEwtExchange(arrEwtExcCond, "", false)

	// 		// 		if EwtExcDet != nil {
	// 		// 			status = helpers.Translate(EwtExcDet.Status, s.LangCode)

	// 		// 			if EwtExcDet.Status == "PAID" {
	// 		// 				status = helpers.Translate("completed", s.LangCode)
	// 		// 			}

	// 		// 			if EwtExcDet.Status == "REJECT" || EwtExcDet.Status == "FAILED" {
	// 		// 				statusColorCode = "#FD4343"
	// 		// 			} else if EwtExcDet.Status == "PENDING" {
	// 		// 				statusColorCode = "#DBA000"
	// 		// 			} else if EwtExcDet.Status == "VOID" {
	// 		// 				statusColorCode = "#FD4343"
	// 		// 			} else {
	// 		// 				statusColorCode = "#00A01F"
	// 		// 			}

	// 		// 			arrCond := make([]models.WhereCondFn, 0)
	// 		// 			arrCond = append(arrCond,
	// 		// 				models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: EwtExcDet.EwalletTypeID},
	// 		// 			)

	// 		// 			ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
	// 		// 			if ewtSetup != nil {
	// 		// 				decimalPoint = uint(ewtSetup.DecimalPoint)
	// 		// 			} else {
	// 		// 				decimalPoint = uint(2)
	// 		// 			}

	// 		// 			amount = helpers.CutOffDecimal(EwtExcDet.Amount, decimalPoint, ".", ",")
	// 		// 			amount = amount + " " + helpers.Translate(ewtSetup.EwtTypeName, s.LangCode)
	// 		// 			ewalletTypeName = helpers.Translate(ewtSetup.EwtTypeName, s.LangCode)

	// 		// 			arrWalletModuleStatementList = append(arrWalletModuleStatementList,
	// 		// 				ExchangeTransaction{
	// 		// 					TransDate:       transDate,
	// 		// 					EwalletTypeName: ewalletTypeName,
	// 		// 					DocNo:           docNo,
	// 		// 					Amount:          amount,
	// 		// 					Payment:         paymentAmount,
	// 		// 					Remark:          remark,
	// 		// 					Status:          status,
	// 		// 					StatusColorCode: statusColorCode,
	// 		// 				})
	// 		// 		}
	// 		// 	}
	// 		// }

	// 		//for internal exchange
	// 		arrEwtDetCond := make([]models.WhereCondFn, 0)
	// 		arrEwtDetCond = append(arrEwtDetCond,
	// 			models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
	// 			models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "EXCHANGE"},
	// 		)
	// 		EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

	// 		if len(EwtDet) > 0 {
	// 			arrExchangeData := make(map[string][]ExchangeData)
	// 			for _, v := range EwtDet {
	// 				arrExchangeData[v.DocNo] = append(arrExchangeData[v.DocNo],
	// 					ExchangeData{
	// 						MemberId:      v.MemberID,
	// 						EwalletTypeId: v.EwalletTypeID,
	// 						TransDate:     v.TransDate,
	// 						DocNo:         v.DocNo,
	// 						TotalIn:       v.TotalIn,
	// 						TotalOut:      v.TotalOut,
	// 						Remark:        v.Remark,
	// 					})
	// 			}

	// 			for _, v2 := range arrExchangeData {
	// 				amount := helpers.CutOffDecimal(float64(0), 2, ".", ",")
	// 				var paymentAmount string
	// 				var docNo string
	// 				var transDate string
	// 				var ewalletTypeName string
	// 				var remark string
	// 				if len(v2) > 1 {
	// 					for _, v3 := range v2 {
	// 						paymentAmt := helpers.CutOffDecimal(float64(0), 2, ".", ",")
	// 						arrCond := make([]models.WhereCondFn, 0)
	// 						arrCond = append(arrCond,
	// 							models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: v3.EwalletTypeId},
	// 						)
	// 						ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
	// 						if ewtSetup != nil {
	// 							decimalPoint = uint(ewtSetup.DecimalPoint)
	// 						} else {
	// 							decimalPoint = uint(2)
	// 						}

	// 						if v3.TotalOut > 0 {
	// 							paymentAmt = helpers.CutOffDecimal(v3.TotalOut, decimalPoint, ".", ",")
	// 							paymentAmount += paymentAmt + " " + helpers.Translate(ewtSetup.EwtTypeName, s.LangCode) + ","
	// 						}

	// 						transDate = v3.TransDate.Format("2006-01-02 15:04:05")
	// 						docNo = v3.DocNo
	// 						remark = helpers.TransRemark(v3.Remark, s.LangCode)
	// 					}
	// 				}

	// 				paymentAmount = strings.TrimSuffix(paymentAmount, ",")

	// 				arrEwtExcCond := make([]models.WhereCondFn, 0)
	// 				arrEwtExcCond = append(arrEwtExcCond,
	// 					models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: docNo},
	// 				)
	// 				EwtExcDet, _ := models.GetEwtExchange(arrEwtExcCond, "", false)

	// 				if EwtExcDet != nil {
	// 					status = helpers.Translate(EwtExcDet.Status, s.LangCode)

	// 					if EwtExcDet.Status == "PAID" {
	// 						status = helpers.Translate("completed", s.LangCode)
	// 					}

	// 					if EwtExcDet.Status == "REJECT" || EwtExcDet.Status == "FAILED" {
	// 						statusColorCode = "#FD4343"
	// 					} else if EwtExcDet.Status == "PENDING" {
	// 						statusColorCode = "#DBA000"
	// 					} else if EwtExcDet.Status == "VOID" {
	// 						statusColorCode = "#FD4343"
	// 					} else {
	// 						statusColorCode = "#00A01F"
	// 					}

	// 					arrCond := make([]models.WhereCondFn, 0)
	// 					arrCond = append(arrCond,
	// 						models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: EwtExcDet.EwalletTypeID},
	// 					)

	// 					ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
	// 					if ewtSetup != nil {
	// 						decimalPoint = uint(ewtSetup.DecimalPoint)
	// 					} else {
	// 						decimalPoint = uint(2)
	// 					}

	// 					amount = helpers.CutOffDecimal(EwtExcDet.Amount, decimalPoint, ".", ",")
	// 					amount = amount + " " + helpers.Translate(ewtSetup.EwtTypeName, s.LangCode)
	// 					ewalletTypeName = helpers.Translate(ewtSetup.EwtTypeName, s.LangCode)

	// 					arrWalletModuleStatementList = append(arrWalletModuleStatementList,
	// 						ExchangeTransaction{
	// 							TransDate:       transDate,
	// 							EwalletTypeName: ewalletTypeName,
	// 							DocNo:           docNo,
	// 							Amount:          amount,
	// 							Payment:         paymentAmount,
	// 							Remark:          remark,
	// 							Status:          status,
	// 							StatusColorCode: statusColorCode,
	// 						})
	// 				}
	// 			}
	// 		}
	// 	}
	case "contract":
		if ewtSetup != nil {
			if ewtSetup.Control == "BLOCKCHAIN" {
				arrBlockCond := make([]models.WhereCondFn, 0)
				arrBlockCond = append(arrBlockCond,
					models.WhereCondFn{Condition: "blockchain_trans.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "CONTRACT"},
				)
				BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

				if len(BlockchainTrans) > 0 {
					for _, v := range BlockchainTrans {
						SlsDet, _ := models.GetSlsMasterByDocNo(v.DocNo)

						if SlsDet != nil {
							status = helpers.Translate(SlsDet.StatusDesc, s.LangCode)

							if SlsDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}

							if SlsDet.Status == "R" || SlsDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if SlsDet.Status == "P" {
								statusColorCode = "#DBA000"
							} else if SlsDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(v.TransactionType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			} else {
				arrEwtDetCond := make([]models.WhereCondFn, 0)
				arrEwtDetCond = append(arrEwtDetCond,
					models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "CONTRACT"},
				)
				EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

				if len(EwtDet) > 0 {
					for _, v := range EwtDet {
						SlsDet, _ := models.GetSlsMasterByDocNo(v.DocNo)

						if SlsDet != nil {
							status = helpers.Translate(SlsDet.StatusDesc, s.LangCode)

							if SlsDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}

							if SlsDet.Status == "R" || SlsDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if SlsDet.Status == "P" {
								statusColorCode = "#DBA000"
							} else if SlsDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.CreatedAt.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(v.TransactionType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})

					}
				}
			}
		} else {
			//for module transaction history
			arrBlockCond := make([]models.WhereCondFn, 0)
			arrBlockCond = append(arrBlockCond,
				models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
				models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "CONTRACT"},
			)
			BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

			if len(BlockchainTrans) > 0 {
				for _, v := range BlockchainTrans {
					SlsDet, _ := models.GetSlsMasterByDocNo(v.DocNo)

					if SlsDet != nil {
						status = helpers.Translate(SlsDet.StatusDesc, s.LangCode)

						if SlsDet.Status == "AP" {
							status = helpers.Translate("completed", s.LangCode)
						}

						if SlsDet.Status == "R" || SlsDet.Status == "F" {
							statusColorCode = "#FD4343"
						} else if SlsDet.Status == "P" {
							statusColorCode = "#DBA000"
						} else if SlsDet.Status == "V" {
							statusColorCode = "#FD4343"
						} else {
							statusColorCode = "#00A01F"
						}

						// arrCond := make([]models.WhereCondFn, 0)
						// arrCond = append(arrCond,
						// 	models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: SlsDet.EwalletTypeID},
						// )
						// ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
						// if ewtSetup != nil {
						// 	decimalPoint = uint(ewtSetup.DecimalPoint)
						// } else {
						// 	decimalPoint = uint(2)
						// }

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						arrWalletModuleStatementList = append(arrWalletModuleStatementList,
							ContractTransaction{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								EwalletTypeName: helpers.Translate(ewtSetup.EwtTypeName, s.LangCode),
								DocNo:           SlsDet.DocNo,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			}
		}
	case "withdraw":
		if ewtSetup != nil {
			if ewtSetup.Control == "BLOCKCHAIN" {
				arrBlockCond := make([]models.WhereCondFn, 0)
				arrBlockCond = append(arrBlockCond,
					models.WhereCondFn{Condition: "blockchain_trans.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "WITHDRAW"},
				)
				BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

				if len(BlockchainTrans) > 0 {
					for _, v := range BlockchainTrans {
						transType := v.TransactionType

						withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

						if withdrawDet != nil {
							status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
							if withdrawDet.Status == "W" || withdrawDet.Status == "I" {
								status = helpers.Translate("pending", s.LangCode)
							}

							if withdrawDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}

							if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if withdrawDet.Status == "P" || withdrawDet.Status == "W" || withdrawDet.Status == "I" {
								statusColorCode = "#DBA000"
							} else if withdrawDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						if v.TotalIn > 0 {
							if v.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
								statusColorCode = "#00A01F"
							} else {
								status = helpers.Translate("failed", s.LangCode)
								statusColorCode = "#FD4343"
							}
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			} else {
				arrEwtDetCond := make([]models.WhereCondFn, 0)
				arrEwtDetCond = append(arrEwtDetCond,
					models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "WITHDRAW"},
				)
				EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

				if len(EwtDet) > 0 {
					for _, v := range EwtDet {
						withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

						if withdrawDet != nil {
							status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
							if withdrawDet.Status == "W" {
								status = helpers.Translate("pending", s.LangCode)
							}
							if withdrawDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}
							if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if withdrawDet.Status == "P" {
								statusColorCode = "#DBA000"
							} else if withdrawDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(v.TransactionType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			}

		} else {
			//for module transaction history
			arrEwtDetCond := make([]models.WhereCondFn, 0)
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "WITHDRAW"},
				models.WhereCondFn{Condition: "ewt_detail.total_out > ?", CondValue: 0},
			)
			EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

			if len(EwtDet) > 0 {
				for _, v := range EwtDet {
					arrCond := make([]models.WhereCondFn, 0)
					arrCond = append(arrCond,
						models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: v.EwalletTypeID},
					)
					ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
					if ewtSetup != nil {
						decimalPoint = uint(ewtSetup.DecimalPoint)
					} else {
						decimalPoint = uint(2)
					}

					amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
					payment := ""
					withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

					if withdrawDet != nil {
						status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
						if withdrawDet.Status == "W" || withdrawDet.Status == "I" {
							status = helpers.Translate("pending", s.LangCode)
						}
						if withdrawDet.Status == "AP" {
							status = helpers.Translate("completed", s.LangCode)
						}
						if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
							statusColorCode = "#FD4343"
						} else if withdrawDet.Status == "P" || withdrawDet.Status == "W" || withdrawDet.Status == "I" {
							statusColorCode = "#DBA000"
						} else if withdrawDet.Status == "V" {
							statusColorCode = "#FD4343"
						} else {
							statusColorCode = "#00A01F"
						}

						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",") + " " + withdrawDet.ConvertedCurrencyCode
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",") + " " + withdrawDet.ConvertedCurrencyCode
							amount = "-" + amount
						}

						if v.TotalIn > 0 {
							payment = helpers.CutOffDecimal(withdrawDet.ConvertedNetAmount, decimalPoint, ".", ",") + " " + withdrawDet.ConvertedCurrencyCode
							payment2 := helpers.CutOffDecimal(withdrawDet.GasFee, decimalPoint, ".", ",") + " " + withdrawDet.CurrencyCode

							if withdrawDet.CurrencyCode == withdrawDet.ConvertedCurrencyCode && withdrawDet.GasFee <= 0 {
								payment = "+" + payment
							} else {
								payment = "+" + payment + "," + payment2
							}
						}

						if v.TotalOut > 0 {
							payment = helpers.CutOffDecimal(withdrawDet.ConvertedNetAmount, decimalPoint, ".", ",") + " " + withdrawDet.ConvertedCurrencyCode
							payment2 := helpers.CutOffDecimal(withdrawDet.GasFee, decimalPoint, ".", ",") + " " + withdrawDet.CurrencyCode

							if withdrawDet.CurrencyCode == withdrawDet.ConvertedCurrencyCode && withdrawDet.GasFee <= 0 {
								payment = "-" + payment
							} else {
								payment = "-" + payment + "," + payment2
							}
						}
					}

					arrWalletModuleStatementList = append(arrWalletModuleStatementList,
						WithdrawTransaction{
							TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
							EwalletTypeName: helpers.Translate(ewtSetup.EwtTypeName, s.LangCode),
							Address:         withdrawDet.CryptoAddrTo,
							DocNo:           withdrawDet.DocNo,
							Amount:          amount,
							Payment:         payment,
							Status:          status,
							StatusColorCode: statusColorCode,
						})
				}
			}

			//for module transaction history
			// arrBlockCond := make([]models.WhereCondFn, 0)
			// arrBlockCond = append(arrBlockCond,
			// 	models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
			// 	models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "WITHDRAW"},
			// )
			// BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

			// if len(BlockchainTrans) > 0 {
			// 	for _, v := range BlockchainTrans {

			// 		withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

			// 		payment := ""
			// 		if withdrawDet != nil {
			// 			status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
			// 			if withdrawDet.Status == "W" || withdrawDet.Status == "I" {
			// 				status = helpers.Translate("pending", s.LangCode)
			// 			}

			// 			if withdrawDet.Status == "AP" {
			// 				status = helpers.Translate("completed", s.LangCode)
			// 			}

			// 			if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
			// 				statusColorCode = "#FD4343"
			// 			} else if withdrawDet.Status == "P" || withdrawDet.Status == "W" || withdrawDet.Status == "I" {
			// 				statusColorCode = "#DBA000"
			// 			} else if withdrawDet.Status == "V" {
			// 				statusColorCode = "#FD4343"
			// 			} else {
			// 				statusColorCode = "#00A01F"
			// 			}

			// 			arrCond := make([]models.WhereCondFn, 0)
			// 			arrCond = append(arrCond,
			// 				models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: withdrawDet.EwalletTypeId},
			// 			)
			// 			ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
			// 			if ewtSetup != nil {
			// 				decimalPoint = uint(ewtSetup.DecimalPoint)
			// 			} else {
			// 				decimalPoint = uint(2)
			// 			}

			// 			amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",") + " " + withdrawDet.CurrencyCode
			// 			if v.TotalIn > 0 {
			// 				amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",") + " " + withdrawDet.CurrencyCode
			// 				amount = "+" + amount
			// 			}

			// 			if v.TotalOut > 0 {
			// 				amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",") + " " + withdrawDet.CurrencyCode
			// 				amount = "-" + amount
			// 			}

			// 			convdecimalPoint := uint(2)
			// 			arrCond = make([]models.WhereCondFn, 0)
			// 			arrCond = append(arrCond,
			// 				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: withdrawDet.ConvertedCurrencyCode},
			// 			)
			// 			ewtSetup, _ = models.GetEwtSetupFn(arrCond, "", false)
			// 			if ewtSetup != nil {
			// 				convdecimalPoint = uint(ewtSetup.DecimalPoint)
			// 			}

			// 			payment = helpers.CutOffDecimal(withdrawDet.ConvertedNetAmount, convdecimalPoint, ".", ",") + " " + withdrawDet.ConvertedCurrencyCode + "," + helpers.CutOffDecimal(withdrawDet.Pool, decimalPoint, ".", ",") + " " + withdrawDet.CurrencyCode + "," + helpers.CutOffDecimal(withdrawDet.AdminFee, decimalPoint, ".", ",") + " " + withdrawDet.CurrencyCode

			// 			arrWalletModuleStatementList = append(arrWalletModuleStatementList,
			// 				WithdrawTransaction{
			// 					TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
			// 					EwalletTypeName: helpers.Translate(ewtSetup.EwtTypeName, s.LangCode),
			// 					DocNo:           withdrawDet.DocNo,
			// 					Amount:          amount,
			// 					Payment:         payment,
			// 					Address:         withdrawDet.CryptoAddrTo,
			// 					Status:          status,
			// 					StatusColorCode: statusColorCode,
			// 				})
			// 		}

			// 	}
			// }
		}
	case "bonus":
		if ewtSetup != nil {
			if ewtSetup.Control == "BLOCKCHAIN" {
				arrBlockCond := make([]models.WhereCondFn, 0)
				arrBlockCond = append(arrBlockCond,
					models.WhereCondFn{Condition: "blockchain_trans.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "BONUS"},
				)
				BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

				if len(BlockchainTrans) > 0 {
					for _, v := range BlockchainTrans {
						transType := v.TransactionType
						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						if v.Status == "R" {
							statusColorCode = "#FD4343"
							status = helpers.Translate("reject", s.LangCode)
						} else if v.Status == "F" {
							statusColorCode = "#FD4343"
							status = helpers.Translate("failed", s.LangCode)
						} else if v.Status == "P" {
							statusColorCode = "#DBA000"
							status = helpers.Translate("pending", s.LangCode)
						} else if v.Status == "V" {
							statusColorCode = "#FD4343"
							status = helpers.Translate("void", s.LangCode)
						} else {
							status = helpers.Translate("completed", s.LangCode)
							statusColorCode = "#00A01F"
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			} else {
				arrEwtDetCond := make([]models.WhereCondFn, 0)
				arrEwtDetCond = append(arrEwtDetCond,
					models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "ewt_detail.transaction_type != ?", CondValue: "TOPUP"},
				)
				EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

				if len(EwtDet) > 0 {
					for _, v := range EwtDet {
						transType := v.TransactionType
						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						status = helpers.Translate("completed", s.LangCode)
						statusColorCode = "#00A01F"

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.CreatedAt.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			}
		}
	case "staking":
		if ewtSetup != nil {
			if ewtSetup.Control == "BLOCKCHAIN" {
				arrBlockCond := make([]models.WhereCondFn, 0)
				arrBlockCond = append(arrBlockCond,
					models.WhereCondFn{Condition: "blockchain_trans.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type =?", CondValue: "STAKING"},
				)
				BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

				if len(BlockchainTrans) > 0 {
					for _, v := range BlockchainTrans {

						SlsDet, _ := models.GetSlsMasterByDocNo(v.DocNo)
						if SlsDet != nil {
							status = helpers.Translate(SlsDet.StatusDesc, s.LangCode)

							if SlsDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}

							if SlsDet.Status == "R" || SlsDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if SlsDet.Status == "P" {
								statusColorCode = "#DBA000"
							} else if SlsDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						transType := v.TransactionType
						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			} else {
				arrEwtDetCond := make([]models.WhereCondFn, 0)
				arrEwtDetCond = append(arrEwtDetCond,
					models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "ewt_detail.transaction_type != ?", CondValue: "TOPUP"},
				)
				EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

				if len(EwtDet) > 0 {
					for _, v := range EwtDet {
						SlsDet, _ := models.GetSlsMasterByDocNo(v.DocNo)
						if SlsDet != nil {
							status = helpers.Translate(SlsDet.StatusDesc, s.LangCode)

							if SlsDet.Status == "AP" {
								status = helpers.Translate("completed", s.LangCode)
							}

							if SlsDet.Status == "R" || SlsDet.Status == "F" {
								statusColorCode = "#FD4343"
							} else if SlsDet.Status == "P" {
								statusColorCode = "#DBA000"
							} else if SlsDet.Status == "V" {
								statusColorCode = "#FD4343"
							} else {
								statusColorCode = "#00A01F"
							}
						}

						transType := v.TransactionType
						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.CreatedAt.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}
			}
		}
	default: // all
		if ewtSetup != nil {
			if ewtSetup.Control == "BLOCKCHAIN" { //get blockchain_trans

				arrBlockCond := make([]models.WhereCondFn, 0)
				arrBlockCond = append(arrBlockCond,
					models.WhereCondFn{Condition: "blockchain_trans.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "blockchain_trans.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type != ?", CondValue: "STAKING-APPROVED"},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type != ?", CondValue: "WITHDRAW-POOL"},
					models.WhereCondFn{Condition: "blockchain_trans.transaction_type != ?", CondValue: "P2P_POOL"},
				)

				if s.TransType != "" {
					arrBlockCond = append(arrBlockCond,
						models.WhereCondFn{Condition: "blockchain_trans.transaction_type = ?", CondValue: strings.ToUpper(s.TransType)},
					)
				}
				BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

				if len(BlockchainTrans) > 0 {
					for _, v := range BlockchainTrans {
						transType := v.TransactionType

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")

						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						if transType == "TRANSFER_TO_EXCHANGE" {
							transType = "TRANSFER"
						}
						if v.TransactionType == "EXCHANGE" {
							// arrEwtExcCond := make([]models.WhereCondFn, 0)
							// arrEwtExcCond = append(arrEwtExcCond,
							// 	models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: v.DocNo},
							// )
							// EwtExcDet, _ := models.GetEwtExchange(arrEwtExcCond, "", false)

							// if EwtExcDet != nil {
							// 	status = helpers.Translate(EwtExcDet.Status, s.LangCode)

							// 	if EwtExcDet.Status == "PAID" {
							// 		status = helpers.Translate("completed", s.LangCode)
							// 	}
							// 	if EwtExcDet.Status == "REJECT" || EwtExcDet.Status == "FAILED" {
							// 		statusColorCode = "#FD4343"
							// 	} else if EwtExcDet.Status == "PENDING" {
							// 		statusColorCode = "#DBA000"
							// 	} else if EwtExcDet.Status == "VOID" {
							// 		statusColorCode = "#FD4343"
							// 	} else {
							// 		statusColorCode = "#00A01F"
							// 	}
							// }
						} else if v.TransactionType == "TRANSFER_TO_EXCHANGE" {
							if v.TotalIn > 0 {
								if remark != "" {
									remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
								}
							}

							TransferExcDet, _ := models.GetEwtTransferExchangeDetailByDocNo(v.DocNo)

							if TransferExcDet != nil {
								if v.TotalIn > 0 {
									if TransferExcDet.Remark != "" {
										remark = remark + " " + "(" + helpers.TransRemark(TransferExcDet.Remark, s.LangCode) + ")"
									}
								}
								status = helpers.Translate(TransferExcDet.StatusDesc, s.LangCode)
								if TransferExcDet.Status == "W" {
									status = helpers.Translate("pending", s.LangCode)
								}
								if TransferExcDet.Status == "AP" {
									status = helpers.Translate("completed", s.LangCode)
								}
								if TransferExcDet.Status == "R" || TransferExcDet.Status == "F" {
									statusColorCode = "#FD4343"
								} else if TransferExcDet.Status == "P" || TransferExcDet.Status == "W" {
									statusColorCode = "#DBA000"
								} else if TransferExcDet.Status == "V" {
									statusColorCode = "#FD4343"
								} else {
									statusColorCode = "#00A01F"
								}
							}
						} else if v.TransactionType == "CONTRACT" {
							SlsDet, _ := models.GetSlsMasterByDocNo(v.DocNo)

							if SlsDet != nil {
								status = helpers.Translate(SlsDet.StatusDesc, s.LangCode)

								if SlsDet.Status == "AP" {
									status = helpers.Translate("completed", s.LangCode)
								}
								if SlsDet.Status == "R" || SlsDet.Status == "F" {
									statusColorCode = "#FD4343"
								} else if SlsDet.Status == "P" {
									statusColorCode = "#DBA000"
								} else if SlsDet.Status == "V" {
									statusColorCode = "#FD4343"
								} else {
									statusColorCode = "#00A01F"
								}
							}
						} else if v.TransactionType == "WITHDRAW" {
							withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

							if withdrawDet != nil {
								status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
								if withdrawDet.Status == "AP" {
									status = helpers.Translate("completed", s.LangCode)
								}
								if withdrawDet.Status == "W" {
									status = helpers.Translate("pending", s.LangCode)
								}
								if withdrawDet.Status == "I" {
									status = helpers.Translate("pending", s.LangCode)
								}

								if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
									statusColorCode = "#FD4343"
								} else if withdrawDet.Status == "P" || withdrawDet.Status == "W" {
									statusColorCode = "#DBA000"
								} else if withdrawDet.Status == "V" {
									statusColorCode = "#FD4343"
								} else {
									statusColorCode = "#00A01F"
								}
							}

							if v.TotalIn > 0 {
								if v.Status == "AP" {
									status = helpers.Translate("completed", s.LangCode)
									statusColorCode = "#00A01F"
								} else {
									status = helpers.Translate("failed", s.LangCode)
									statusColorCode = "#FD4343"
								}
							}

						} else {
							if v.Status == "R" {
								statusColorCode = "#FD4343"
								status = helpers.Translate("reject", s.LangCode)
							} else if v.Status == "F" {
								statusColorCode = "#FD4343"
								status = helpers.Translate("failed", s.LangCode)
							} else if v.Status == "P" {
								statusColorCode = "#DBA000"
								status = helpers.Translate("pending", s.LangCode)
							} else if v.Status == "V" {
								statusColorCode = "#FD4343"
								status = helpers.Translate("void", s.LangCode)
							} else {
								status = helpers.Translate("completed", s.LangCode)
								statusColorCode = "#00A01F"
							}
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.DtTimestamp.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}

			} else { //get ewt_detail
				arrEwtDetCond := make([]models.WhereCondFn, 0)
				arrEwtDetCond = append(arrEwtDetCond,
					models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: ewtSetup.ID},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "ewt_detail.transaction_type != ?", CondValue: "TOPUP"},
				)
				if s.TransType != "" {
					arrEwtDetCond = append(arrEwtDetCond,
						models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: strings.ToUpper(s.TransType)},
					)
				}
				EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

				if len(EwtDet) > 0 {
					for _, v := range EwtDet {

						remark := v.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")

						if v.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if v.TotalOut > 0 {
							amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
							amount = "-" + amount
						}

						transType := v.TransactionType

						if v.TransactionType == "WITHDRAW" {
							withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

							if withdrawDet != nil {
								status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
								if withdrawDet.Status == "AP" {
									status = helpers.Translate("completed", s.LangCode)
								}
								if withdrawDet.Status == "W" {
									status = helpers.Translate("pending", s.LangCode)
								}
								if withdrawDet.Status == "I" {
									status = helpers.Translate("pending", s.LangCode)
								}
								if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
									statusColorCode = "#FD4343"
								} else if withdrawDet.Status == "P" || withdrawDet.Status == "W" {
									statusColorCode = "#DBA000"
								} else if withdrawDet.Status == "V" {
									statusColorCode = "#FD4343"
								} else {
									statusColorCode = "#00A01F"
								}
							}
						}

						if v.TransactionType == "TRANSFER" {
							transferDet, _ := models.GetEwtTransferDetailByDocNo(v.DocNo)

							if transferDet != nil {
								if v.TotalIn > 0 {
									if transferDet.Remark != "" {
										remark = remark + " " + "(" + helpers.TransRemark(transferDet.Remark, s.LangCode) + ")"
									}
								}
								status = helpers.Translate(transferDet.StatusDesc, s.LangCode)
								if transferDet.Status == "AP" {
									status = helpers.Translate("completed", s.LangCode)
								}
								if transferDet.Status == "W" {
									status = helpers.Translate("pending", s.LangCode)
								}
								if transferDet.Status == "R" || transferDet.Status == "F" {
									statusColorCode = "#FD4343"
								} else if transferDet.Status == "P" || transferDet.Status == "W" {
									statusColorCode = "#DBA000"
								} else if transferDet.Status == "V" {
									statusColorCode = "#FD4343"
								} else {
									statusColorCode = "#00A01F"
								}
							}
						}

						// if v.TransactionType == "EXCHANGE" {
						// 	arrEwtExcCond := make([]models.WhereCondFn, 0)
						// 	arrEwtExcCond = append(arrEwtExcCond,
						// 		models.WhereCondFn{Condition: "ewt_exchange.doc_no = ?", CondValue: v.DocNo},
						// 	)
						// 	EwtExcDet, _ := models.GetEwtExchange(arrEwtExcCond, "", false)

						// 	if EwtExcDet != nil {
						// 		status = helpers.Translate(EwtExcDet.Status, s.LangCode)

						// 		if EwtExcDet.Status == "PAID" {
						// 			status = helpers.Translate("completed", s.LangCode)
						// 		}
						// 		if EwtExcDet.Status == "REJECT" || EwtExcDet.Status == "FAILED" {
						// 			statusColorCode = "#FD4343"
						// 		} else if EwtExcDet.Status == "PENDING" {
						// 			statusColorCode = "#DBA000"
						// 		} else if EwtExcDet.Status == "VOID" {
						// 			statusColorCode = "#FD4343"
						// 		} else {
						// 			statusColorCode = "#00A01F"
						// 		}
						// 	}
						// }

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
								TransType:       helpers.Translate(transType, s.LangCode) + remark,
								Amount:          amount,
								Status:          status,
								StatusColorCode: statusColorCode,
							})
					}
				}

				//for topup
				arrTopupCond := make([]models.WhereCondFn, 0)
				arrTopupCond = append(arrTopupCond,
					models.WhereCondFn{Condition: "a.member_id = ?", CondValue: s.MemberID},
					models.WhereCondFn{Condition: "a.status = ?", CondValue: "AP"},
					models.WhereCondFn{Condition: "a.ewallet_type_id = ?", CondValue: ewtSetup.ID},
				)

				EwtTopup, _ := models.GetEwtTopupArrayFn(arrTopupCond, false)

				if len(EwtTopup) > 0 {
					for _, v2 := range EwtTopup {
						// var decimalPointConv uint
						// arrConvCond := make([]models.WhereCondFn, 0)
						// arrConvCond = append(arrConvCond,
						// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: v2.ConvertedCurrencyCode},
						// )
						// ewtSetupConv, _ := models.GetEwtSetupFn(arrConvCond, "", false)
						// if ewtSetupConv != nil {
						// 	decimalPointConv = uint(ewtSetupConv.DecimalPoint)
						// }

						status = helpers.Translate(v2.StatusDesc, s.LangCode)

						if v2.Status == "AP" {
							status = helpers.Translate("completed", s.LangCode)
						}
						if v2.Status == "R" || v2.Status == "F" {
							statusColorCode = "#FD4343"
						} else if v2.Status == "P" {
							statusColorCode = "#DBA000"
						} else if v2.Status == "V" {
							statusColorCode = "#FD4343"
						} else {
							statusColorCode = "#00A01F"
						}

						amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")
						if v2.TotalIn > 0 {
							amount = helpers.CutOffDecimal(v2.TotalIn, decimalPoint, ".", ",")
							amount = "+" + amount
						}

						if ewtSetup.Control == "BLOCKCHAIN" {
							if v2.TotalIn > 0 {
								amount = helpers.CutOffDecimal(v2.ConvertedTotalAmount, decimalPoint, ".", ",")
								amount = "+" + amount
							}
						}

						remark := v2.Remark

						if remark != "" {
							remark = "-" + helpers.TransRemark(v2.Remark, s.LangCode)
						}

						arrWalletStatementList = append(arrWalletStatementList,
							WalletTransactionResultStructV3{
								TransDate:    v2.TransDate.Format("2006-01-02 15:04:05"),
								TransType:    helpers.Translate("receive", s.LangCode) + remark,
								Amount:       amount,
								CurrencyCode: v2.CurrencyCode,
								// ConvertedAmount:       helpers.CutOffDecimal(v2.ConvertedTotalAmount, decimalPointConv, ".", ","),
								ConvertedCurrencyCode: v2.ConvertedCurrencyCode,
								Status:                status,
								StatusColorCode:       statusColorCode,
							})
					}
				}
			}
		}
	}

	//start paginate
	if len(arrWalletStatementList) > 0 {
		sort.Slice(arrWalletStatementList, func(p, q int) bool {
			return arrWalletStatementList[q].TransDate < arrWalletStatementList[p].TransDate
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

		totalRecord := len(arrWalletStatementList)

		totalPage := float64(totalRecord) / float64(limit)
		totalPage = math.Ceil(totalPage)

		pageStart, pageEnd := helpers.Paginate(int(s.Page), int(limit), totalRecord)

		processArr := arrWalletStatementList[pageStart:pageEnd]

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
	} else {

		sort.Slice(arrWalletModuleStatementList, func(i, j int) bool {
			commonID1 := reflect.ValueOf(arrWalletModuleStatementList[i]).FieldByName("TransDate").String()
			commonID2 := reflect.ValueOf(arrWalletModuleStatementList[j]).FieldByName("TransDate").String()
			return commonID1 > commonID2
		})

		page := base.Pagination{
			Page:    s.Page,
			DataArr: arrWalletModuleStatementList,
		}

		arrDataReturn := page.PaginationInterfaceV1()

		return arrDataReturn, nil
	}

}

type DataReceivedStruct struct {
	PrjConfigCode string `json:"prd_config_code"`
	UrlLink       string `json:"url_link"`
	ApiType       string `json:"api_type"`
	Method        string `json:"method"`
	DataReceived  string `json:"data_received"`
	ServerData    string `json:"server_data"`
}

type ProcessCryptoReturnDataStruct struct {
	Id               int     `json:"id"`
	ConfigCode       string  `json:"config_code"`
	TransStatus      string  `json:"trans_status"`
	CallbackStatus   string  `json:"callback_status"`
	CallbackStatusAt string  `json:"callback_status_at"`
	TxHash           string  `json:"tx_hash" valid:"Required"`
	FromAddr         string  `json:"from_addr" valid:"Required"`
	ToAddr           string  `json:"to_addr" valid:"Required"`
	AssetCode        string  `json:"asset_code" valid:"Required"`
	Amount           float64 `json:"from_addr" valid:"Required"`
	CreatedAt        string  `json:"created_at"`
}

func ProcessCryptoReturn(req *http.Request, data ProcessCryptoReturnDataStruct) {
	if setting.Cfg.Section("app").Key("Environment").String() == "LIVE" && req.Header.Get("X-Forwarded-For") != "172.104.59.144" {
		os.Exit(1)
	}

	// store in blockchain api log
	server := req.Header
	dataReceived, _ := json.Marshal(data)
	serverData, _ := json.Marshal(server)

	arrDataReceived := DataReceivedStruct{
		PrjConfigCode: "blockchain",
		UrlLink:       req.URL.Host + req.URL.RequestURI(),
		ApiType:       "crypto_return",
		Method:        "POST",
		DataReceived:  string(dataReceived),
		ServerData:    string(serverData),
	}

	bcTx := models.GetDB().Table("blockchain_api_log")

	models.SaveTx(bcTx, arrDataReceived)

	//begin
	// tx := models.Begin()

	// amount := helpers.CutOffDecimal(data.Amount, 10, ".", ",")

	// if strings.ToLower(data.AssetCode) == "eth" && amount <= 0.001 {
	// 	arrUpdBcApiCond := make([]models.WhereCondFn, 0)
	// arrUpdBcApiCond = append(arrUpdBcApiCond,
	// 	models.WhereCondFn{Condition: "prj_config_code = ?", CondValue: arrDataReceived.PrjConfigCode},
	// 	models.WhereCondFn{Condition: "id = ?", CondValue: arrData.EwalletTypeID},
	// )

	// arrExisingWallet, err := models.GetEwtSummaryFn(arrUpdBcApiCond, "", false)
	// if err != nil {
	// 	return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrData}
	// }

	// if len(arrExisingWallet) > 0 {
	// 	updateColumn := map[string]interface{}{"balance": latestbal, "updated_by": createdBy}
	// 	if arrData.TotalIn > 0 {
	// 		// updateColumn["total_in"] = float.Add(arrData.TotalIn, arrExisingWallet[0].TotalIn) // this will cut down decimal. can not use this. tested
	// 		updateColumn["total_in"] = arrData.TotalIn + arrExisingWallet[0].TotalIn
	// 	}
	// 	if arrData.TotalOut > 0 {
	// 		// updateColumn["total_out"] = float.Add(arrData.TotalOut, arrExisingWallet[0].TotalOut) // this will cut down decimal. can not use this. tested
	// 		updateColumn["total_out"] = arrData.TotalOut + arrExisingWallet[0].TotalOut
	// 	}

	// 	err := models.UpdatesFnTx(tx, "ewt_summary", arrUpdBcApiCond, updateColumn, false)
	// }

}

type SignTransactionViaApiStruct struct {
	TokenType  string
	PrivateKey string
	ToAddr     string
	Value      string
	Gas        string
	GasPrice   string
}

type SignTransactionRstStruct struct {
	MessageHash     string `json:"message_hash"`
	V               string `json:"v"`
	R               string `json:"r"`
	S               string `json:"s"`
	RawTransaction  string `json:"raw_transaction"`
	TransactionHash string `json:"transaction_hash"`
}

// func GenerateSignTransactionViaApi
func GenerateSignTransactionViaApi(arrData SignTransactionViaApiStruct) (*SignTransactionRstStruct, error) {

	apiSetting, err := models.GetSysGeneralSetupByID("sign_transaction_api_setting")

	if err != nil {
		base.LogErrorLog("GenerateSignedTransaction_GetSysGeneralSetupByID_failed", err.Error(), "sign_transaction_api_setting", true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sign_transaction_is_not_available", Data: ""}
	}
	if apiSetting == nil {
		base.LogErrorLog("GenerateSignedTransaction_no_sign_transaction_api_setting_setting", "sign_transaction_api_setting", nil, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sign_transaction_is_not_available", Data: ""}
	}

	header := map[string]string{
		"Content-Type":    "application/json",
		"X-Authorization": apiSetting.InputType2,
	}

	data := map[string]interface{}{
		"token_type":  arrData.TokenType,
		"private_key": arrData.PrivateKey,
		"to_addr":     arrData.ToAddr,
		"value":       arrData.Value,
		"gas":         arrData.Gas,
		"gas_price":   arrData.GasPrice,
	}

	res, err := base.RequestAPI(apiSetting.SettingValue1, apiSetting.InputValue1, header, data, nil)

	if err != nil {
		base.LogErrorLogV2("GenerateSignedTransaction_error_in_api_call_before_call", err.Error(), nil, true, "blockchain")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "sign_transaction_is_not_available", Data: ""}
	}

	if res.StatusCode != 200 {
		base.LogErrorLogV2("GenerateSignedTransaction_error_in_api_call_after_call", res.Body, nil, true, "blockchain")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "mnemonic_is_not_available", Data: ""}
	}

	type signTransactionApiRstStruct struct {
		Status     string `json:"status"`
		StatusCode string `json:"status_code"`
		Msg        string `json:"msg"`
		Data       struct {
			MessageHash     string `json:"message_hash"`
			V               string `json:"v"`
			R               string `json:"r"`
			S               string `json:"s"`
			RawTransaction  string `json:"raw_transaction"`
			TransactionHash string `json:"transaction_hash"`
		} `json:"data"`
	}

	var apiRst signTransactionApiRstStruct
	err = json.Unmarshal([]byte(res.Body), &apiRst)

	if err != nil {
		base.LogErrorLog("GenerateSignedTransaction_error_in_json_decode_api_result", err.Error(), res.Body, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "mnemonic_is_not_available", Data: ""}
	}

	arrDataReturn := SignTransactionRstStruct{
		MessageHash:     apiRst.Data.MessageHash,
		V:               apiRst.Data.V,
		R:               apiRst.Data.R,
		S:               apiRst.Data.S,
		RawTransaction:  apiRst.Data.RawTransaction,
		TransactionHash: apiRst.Data.TransactionHash,
	}

	return &arrDataReturn, nil
}

// GetLatestPriceMovementByEwtTypeCode func
func GetLatestPriceMovementByEwtTypeCode(ewtTypeCode string) (float64, string) {
	if strings.ToLower(ewtTypeCode) == "sec" {
		tokenRate, err := models.GetLatestSecPriceMovement()

		if err != nil {
			base.LogErrorLog("walletService:GetLatestPriceMovementByEwtTypeCode()", "GetLatestSecPriceMovement():1", err.Error(), true)
			return 0.00, "something_went_wrong"
		}

		return tokenRate, ""
	} else if strings.ToLower(ewtTypeCode) == "liga" {
		tokenRate, err := models.GetLatestLigaPriceMovement()

		if err != nil {
			base.LogErrorLog("walletService:GetLatestPriceMovementByEwtTypeCode()", "GetLatestSecPriceMovement():1", err.Error(), true)
			return 0.00, "something_went_wrong"
		}

		return tokenRate, ""
	} else {
		base.LogErrorLog("walletService:GetLatestPriceMovementByEwtTypeCode()", "invalid_ewt_type_code_"+ewtTypeCode, "", true)
		return 0.00, "something_went_wrong"
	}
}

// SaveMemberBlochchainWalletStruct struct
type SaveMemberBlochchainWalletStruct struct {
	EntMemberID, EwalletTypeID, LogOnly                                    int
	DocNo, TransactionType, TransactionData, Status, Remark                string
	TotalIn, TotalOut, ConversionRate, ConvertedTotalIn, ConvertedTotalOut float64
}

// SaveMemberBlochchainWallet func
func SaveMemberBlochchainWallet(saveMemberBlockchainWallet SaveMemberBlochchainWalletStruct) (string, map[string]string) {
	var arrData = make(map[string]string)
	// call blockchain transaction api
	hashValue, errMsg := SignedTransaction(saveMemberBlockchainWallet.TransactionData)

	if errMsg != "" {
		return errMsg, nil
	}

	// log blochchain transaction
	db := models.GetDB() // will save record regardless of begintransaction since api already called successfully
	_, err := models.AddBlockchainTrans(db, models.AddBlockchainTransStruct{
		MemberID:          saveMemberBlockchainWallet.EntMemberID,
		EwalletTypeID:     saveMemberBlockchainWallet.EwalletTypeID,
		DocNo:             saveMemberBlockchainWallet.DocNo,
		Status:            saveMemberBlockchainWallet.Status,
		TransactionType:   saveMemberBlockchainWallet.TransactionType,
		TotalIn:           saveMemberBlockchainWallet.TotalIn,
		TotalOut:          saveMemberBlockchainWallet.TotalOut,
		ConversionRate:    saveMemberBlockchainWallet.ConversionRate,
		ConvertedTotalIn:  saveMemberBlockchainWallet.ConvertedTotalIn,
		ConvertedTotalOut: saveMemberBlockchainWallet.ConvertedTotalOut,
		TransactionData:   saveMemberBlockchainWallet.TransactionData,
		HashValue:         hashValue,
		Remark:            saveMemberBlockchainWallet.Remark,
		LogOnly:           saveMemberBlockchainWallet.LogOnly,
	})

	if err != nil {
		base.LogErrorLog("walletService:SaveMemberBlochchainWallet()", "AddBlockchainTrans():1", err.Error(), true)
		return "something_went_wrong", nil
	}

	arrData["hashValue"] = hashValue

	return "", arrData
}

type GenerateSignTransactionStruct struct {
	TokenType       string
	PrivateKey      string
	ContractAddress string
	ChainID         int64
	Nonce           uint64
	ToAddr          string
	Amount          float64 // this is refer to amount for this transaction
	MaxGas          uint64
}

// GenerateSignTransaction func
func GenerateSignTransaction(arrData GenerateSignTransactionStruct) (string, error) {
	data := []byte("")
	// base.LogErrorLog("GenerateSignTransaction_net3", arrData, nil, true)
	var decimal int64 = 18 // default decimal for blockchain transaction amount

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.TokenType},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup != nil {
		if ewtSetup.BlockchainDecimalPoint > 0 {
			decimal = int64(ewtSetup.BlockchainDecimalPoint)
		}
	}

	if strings.ToLower(arrData.TokenType) == "sec" {
		signedTxSEC, err := util.SignTransaction(arrData.ChainID, arrData.PrivateKey, arrData.ToAddr, arrData.Nonce, arrData.MaxGas, arrData.Amount, data)
		if err != nil {
			base.LogErrorLog("GenerateSignTransaction_SignTransaction_sec_failed", err.Error(), arrData, true)
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}

		return signedTxSEC, nil
	} else if strings.ToLower(arrData.TokenType) == "liga" || strings.ToLower(arrData.TokenType) == "usds" {
		signedTxLIGA, err := util.SendERC20(arrData.ChainID, arrData.PrivateKey, arrData.ToAddr, arrData.ContractAddress, arrData.Nonce, arrData.MaxGas, arrData.Amount, decimal)
		if err != nil {
			base.LogErrorLog("GenerateSignTransaction_SignTransaction_SendERC20_failed", err.Error(), arrData, true)
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}
		return signedTxLIGA, nil
	}

	base.LogErrorLog("GenerateSignTransaction_SignTransaction_failed_invalid_token_type", "failed", arrData, true)
	return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_token_type"}
}

// ProcecssGenerateSignTransactionStruct struct
type ProcecssGenerateSignTransactionStruct struct {
	TokenType       string
	PrivateKey      string
	ContractAddress string
	ChainID         int64
	FromAddr        string
	ToAddr          string
	Amount          float64 // this is refer to amount for this transaction
	MaxGas          uint64
}

// ProcecssGenerateSignTransaction func
func ProcecssGenerateSignTransaction(arrData ProcecssGenerateSignTransactionStruct) (string, error) {

	nonce, err := GetTransactionNonceViaAPI(arrData.FromAddr)
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
	}

	arrGenerateSignTransaction := GenerateSignTransactionStruct{
		TokenType:       arrData.TokenType,
		PrivateKey:      arrData.PrivateKey,
		ContractAddress: arrData.ContractAddress,
		ChainID:         arrData.ChainID,
		Nonce:           uint64(nonce),
		ToAddr:          arrData.ToAddr,
		Amount:          arrData.Amount, // this is refer to amount for this transaction
		MaxGas:          arrData.MaxGas,
	}
	// base.LogErrorLog("ProcecssGenerateSignTransaction_net2", arrGenerateSignTransaction, nonce, true)
	signingKey, err := GenerateSignTransaction(arrGenerateSignTransaction)
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
	}

	return signingKey, nil
}

type TransactionCallbackData struct {
	Hash   string
	Status int
}

func UpdateBlockchainTransStatus(tx *gorm.DB, hash string, status string) string {
	// update blockchain trans record
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "blockchain_trans.hash_value = ?", CondValue: hash},
	)

	updateColumn := map[string]interface{}{"status": status}
	err := models.UpdatesFnTx(tx, "blockchain_trans", arrUpdCond, updateColumn, false)

	if err != nil {
		base.LogErrorLog("TransactionCallback", "UpdateBlockchainTrans", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

func UpdateTransferToExchangeStatus(tx *gorm.DB, docNo string, status string) string {
	// update ewt_transfer_exchange record
	arrSlsMasterUpdCond := make([]models.WhereCondFn, 0)
	arrSlsMasterUpdCond = append(arrSlsMasterUpdCond,
		models.WhereCondFn{Condition: "ewt_transfer_exchange.doc_no = ?", CondValue: docNo},
	)

	updateSlsMasterColumn := map[string]interface{}{"status": status}
	ewtTransferExchangeErr := models.UpdatesFnTx(tx, "ewt_transfer_exchange", arrSlsMasterUpdCond, updateSlsMasterColumn, false)

	if ewtTransferExchangeErr != nil {
		base.LogErrorLog("UpdateTransferToExchangeStatus", "UpdateEwtTransferExchange", ewtTransferExchangeErr.Error(), true)
		return "something_went_wrong"
	}

	if status == "AP" {
		// get crypto_addr_to from ewt_transfer_exchange
		ewtTransferExchangeCondFn := make([]models.WhereCondFn, 0)
		ewtTransferExchangeCondFn = append(ewtTransferExchangeCondFn,
			models.WhereCondFn{Condition: "ewt_transfer_exchange.doc_no = ? ", CondValue: docNo},
		)

		ewtTransferExchange, ewtTransferExchangeErr := models.GetEwtTransferExchangeFn(ewtTransferExchangeCondFn, false)

		if ewtTransferExchangeErr != nil || len(ewtTransferExchange) <= 0 {
			base.LogErrorLog("UpdateTransferToExchangeStatus", "GetEwtTransferExchangeFn", ewtTransferExchangeErr.Error(), true)
			return "something_went_wrong"
		}

		// get member info in ent_member_crypto
		entMemberCryptoCondFn := make([]models.WhereCondFn, 0)
		entMemberCryptoCondFn = append(entMemberCryptoCondFn,
			models.WhereCondFn{Condition: "ent_member_crypto.crypto_address = ? ", CondValue: ewtTransferExchange[0].CryptoAddrTo},
		)

		entMemberCrypto, entMemberCryptoErr := models.GetEntMemberCryptoFn(entMemberCryptoCondFn, false)

		if entMemberCryptoErr != nil {
			base.LogErrorLog("UpdateTransferToExchangeStatus", "GetEwtTransferExchangeFn", entMemberCryptoErr.Error(), true)
			return "something_went_wrong"
		}

		//get member info
		arrMemInfoCond := make([]models.WhereCondFn, 0)
		arrMemInfoCond = append(arrMemInfoCond,
			models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: ewtTransferExchange[0].MemberId},
		)
		memberInfo, memErr := models.GetEntMemberFn(arrMemInfoCond, "", false) //get member to details

		if memErr != nil {
			base.LogErrorLog("UpdateTransferToExchangeStatus", "arrMemInfoCond", memErr.Error(), true)
			return "something_went_wrong"
		}

		if memberInfo == nil {
			base.LogErrorLog("UpdateTransferToExchangeStatus", "emptyMemberInfo", arrMemInfoCond, true)
			return "something_went_wrong"
		}

		//insert new record in blockchain_trans
		_, err := models.AddBlockchainTrans(tx, models.AddBlockchainTransStruct{
			MemberID:         entMemberCrypto.MemberID,
			EwalletTypeID:    ewtTransferExchange[0].EwalletTypeId,
			DocNo:            docNo,
			Status:           status,
			TransactionType:  ewtTransferExchange[0].TransactionType,
			TotalIn:          ewtTransferExchange[0].Amount,
			ConversionRate:   1,
			ConvertedTotalIn: ewtTransferExchange[0].Amount,
			TransactionData:  ewtTransferExchange[0].SigningKey,
			HashValue:        ewtTransferExchange[0].TranHash,
			Remark:           "#*transfer_from*#" + " " + memberInfo.NickName,
		})

		if err != nil {
			base.LogErrorLog("wallet_service:UpdateTransferToExchangeStatus()", "AddBlockchainTrans()", err.Error(), true)
			return "something_went_wrong"
		}

	}
	return ""
}

// func TransactionCallback(tx *gorm.DB, data TransactionCallbackData) string {
// 	var bcStatus string

// 	if err != nil {
// 		base.LogErrorLog("TransactionCallback", "UpdateBlockchainTrans", err.Error(), true)
// 		return "something_went_wrong"
// 	}

// 	if len(BlockchainTrans) > 0 {
// 		for _, row := range BlockchainTrans {
// 			if row.TransactionType == "CONTRACT" || row.TransactionType == "STAKING" {
// 				// update sls master record
// 				arrSlsMasterUpdCond := make([]models.WhereCondFn, 0)
// 				arrSlsMasterUpdCond = append(arrSlsMasterUpdCond,
// 					models.WhereCondFn{Condition: "sls_master.doc_no = ?", CondValue: row.DocNo},
// 				)

// 				updateSlsMasterColumn := map[string]interface{}{"status": bcStatus}
// 				slsMasterErr := models.UpdatesFnTx(tx, "sls_master", arrSlsMasterUpdCond, updateSlsMasterColumn, false)

// 				if slsMasterErr != nil {
// 					base.LogErrorLog("TransactionCallback", "UpdateSlsMaster", err.Error(), true)
// 					return "something_went_wrong"
// 				}
// 			} else if row.TransactionType == "TRANSFER_TO_EXCHANGE" {
// 				// update ewt_transfer_exchange record
// 				arrSlsMasterUpdCond := make([]models.WhereCondFn, 0)
// 				arrSlsMasterUpdCond = append(arrSlsMasterUpdCond,
// 					models.WhereCondFn{Condition: "ewt_transfer_exchange.doc_no = ?", CondValue: row.DocNo},
// 				)

// 				updateSlsMasterColumn := map[string]interface{}{"status": bcStatus}
// 				slsMasterErr := models.UpdatesFnTx(tx, "ewt_transfer_exchange", arrSlsMasterUpdCond, updateSlsMasterColumn, false)

// 				if slsMasterErr != nil {
// 					base.LogErrorLog("TransactionCallback", "UpdateEwtTransferExchange", err.Error(), true)
// 					return "something_went_wrong"
// 				}
// 			} else if strings.ToLower(row.TransactionType) == "TRADING_MATCH" {
// 				arrTradMatchUpdCond := make([]models.WhereCondFn, 0)
// 				arrTradMatchUpdCond = append(arrTradMatchUpdCond,
// 					models.WhereCondFn{Condition: "trading_match.doc_no = ?", CondValue: row.DocNo},
// 				)

// 				updateTradMatchColumn := map[string]interface{}{"status": "AP"}
// 				tradMatchErr := models.UpdatesFnTx(tx, "trading_match", arrTradMatchUpdCond, updateTradMatchColumn, false)

// 				if tradMatchErr != nil {
// 					base.LogErrorLog("TransactionCallback", "UpdateTradingMatch", err.Error(), true)
// 					return "something_went_wrong"
// 				}
// 			} else if row.TransactionType == "UNSTAKE" {
// 				errMsg := ApproveUnstakeCallback(tx, bcStatus, row.DocNo, row.TotalIn)
// 				if errMsg != "" {
// 					base.LogErrorLog("TransactionCallback", "ApproveUnstakeCallback()", "[docNo:"+row.DocNo+", hashValue:"+data.Hash+"]"+errMsg, true)
// 					return "something_went_wrong"
// 				}
// 			}
// 		}

// 		return ""
// 	}

// 	return "invalid_hash"
// }

// func GetBlockchainWalletBalanceApiV1 just to get balance from api. extra step is cut off decimal tgt
func GetBlockchainWalletBalanceApiV1(walletTypeCode string, address string) (balance float64, err error) {

	var response app.ApiResponse
	walletTypeCode = strings.ToUpper(walletTypeCode)

	secApiSetting, _ := models.GetSysGeneralSetupByID("sec_api_setting")

	data := map[string]interface{}{
		"token_type": walletTypeCode,
		"address":    address,
	}

	api_key := secApiSetting.InputValue2

	url := secApiSetting.InputValue1 + "api/account/balance"
	header := map[string]string{
		"Content-Type":    "application/json",
		"X-Authorization": api_key,
	}

	res, err_api := base.RequestAPI("POST", url, header, data, &response)

	if err_api != nil {
		base.LogErrorLogV2("GetBlockchainWalletBalanceApiV1 -fail to call blockchain balance api", err_api.Error(), map[string]interface{}{"err": err_api, "data": data}, true, "blockchain")
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if res.StatusCode != 200 {
		errMsg, _ := json.Marshal(response.Msg)
		errMsgStr := string(errMsg)
		errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		base.LogErrorLogV2("GetBlockchainWalletBalanceApiV1 -fail to get blockchain wallet balance", errMsgStr, map[string]interface{}{"err": res.Body, "data": data}, true, "blockchain")
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}

	}

	if response.Data["balance"] == nil {
		base.LogErrorLog("GetBlockchainWalletBalanceApiV1 -empty balance", response.Data, res.Body, true)
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}

	}

	return_balance, _ := json.Marshal(response.Data["balance"])

	strBal := string(return_balance)
	strBal = strings.Replace(strBal, "\"", "", 2)

	var decimalPoint uint
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(walletTypeCode)},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup != nil {
		decimalPoint = uint(ewtSetup.DecimalPoint)
	}

	strBal = helpers.CutOffStringsDecimal(strBal, decimalPoint, '.')
	bal, _ := strconv.ParseFloat(strBal, 64)

	return bal, nil
}

func UpdateWithdrawStatus(tx *gorm.DB, docNo string, status string) string {
	updateStatus := status
	if status == "AP" {
		updateStatus = "W" // update status to W instead of AP
	}

	ewtWithdrawCond := make([]models.WhereCondFn, 0)
	ewtWithdrawCond = append(ewtWithdrawCond,
		models.WhereCondFn{Condition: " ewt_withdraw.doc_no = ? ", CondValue: docNo},
	)
	ewtWithdrawUpdateColumn := map[string]interface{}{
		"status": updateStatus,
	}

	ewtWithdrawErr := models.UpdatesFnTx(tx, "ewt_withdraw", ewtWithdrawCond, ewtWithdrawUpdateColumn, false)

	if ewtWithdrawErr != nil {
		base.LogErrorLog("wallet_service:UpdateWithdrawStatus()", "UpdateEwtWithdraw()", ewtWithdrawErr.Error(), true)
		return "something_went_wrong"
	}

	ewtWithdrawPoolCond := make([]models.WhereCondFn, 0)
	ewtWithdrawPoolCond = append(ewtWithdrawPoolCond,
		models.WhereCondFn{Condition: " ewt_withdraw_pool.doc_no = ? ", CondValue: docNo},
	)
	ewtWithdrawPoolUpdateColumn := map[string]interface{}{
		"status": updateStatus,
	}

	ewtWithdrawPoolErr := models.UpdatesFnTx(tx, "ewt_withdraw_pool", ewtWithdrawPoolCond, ewtWithdrawPoolUpdateColumn, false)

	if ewtWithdrawPoolErr != nil {
		base.LogErrorLog("wallet_service:UpdateWithdrawStatus()", "UpdateEwtWithdrawPool()", ewtWithdrawPoolErr.Error(), true)
		return "something_went_wrong"
	}
	return ""
}

func UpdateWithdrawPoolStatus(tx *gorm.DB, docNo string, status string) string {
	updateStatus := status

	ewtWithdrawCond := make([]models.WhereCondFn, 0)
	ewtWithdrawCond = append(ewtWithdrawCond,
		models.WhereCondFn{Condition: " ewt_withdraw.doc_no = ? ", CondValue: docNo},
	)
	ewtWithdrawUpdateColumn := map[string]interface{}{
		"status": updateStatus,
	}

	ewtWithdrawErr := models.UpdatesFnTx(tx, "ewt_withdraw", ewtWithdrawCond, ewtWithdrawUpdateColumn, false)

	if ewtWithdrawErr != nil {
		base.LogErrorLog("wallet_service:UpdateWithdrawStatus()", "UpdateEwtWithdrawPool()", ewtWithdrawErr.Error(), true)
		return "something_went_wrong"
	}

	ewtWithdrawPoolCond := make([]models.WhereCondFn, 0)
	ewtWithdrawPoolCond = append(ewtWithdrawPoolCond,
		models.WhereCondFn{Condition: " ewt_withdraw_pool.doc_no = ? ", CondValue: docNo},
	)
	ewtWithdrawPoolUpdateColumn := map[string]interface{}{
		"status": updateStatus,
	}

	ewtWithdrawPoolErr := models.UpdatesFnTx(tx, "ewt_withdraw_pool", ewtWithdrawPoolCond, ewtWithdrawPoolUpdateColumn, false)

	if ewtWithdrawPoolErr != nil {
		base.LogErrorLog("wallet_service:UpdateWithdrawStatus()", "UpdateEwtWithdrawPool()", ewtWithdrawPoolErr.Error(), true)
		return "something_went_wrong"
	}
	return ""
}

type MemberAccountTransferExchangeBatchSetupStruct struct {
	MemberID    int
	EntMemberID int
	LangCode    string
	EwalletType string
}

type MemberAccountTransferExchangeBatchSetupRstStruct struct {
	TransferSetupList []transferSetupListStruct `json:"account_list"`
	// Nonce             int                       `json:"nonce"`
}

type transferSetupListStruct struct {
	SigningKeySetting interface{} `json:"signing_key_setting"`
	AvailableBalance  float64     `json:"available_balance"`
	EwtTypeCode       string      `json:"ewallet_type_code"`
	EwtTypeName       string      `json:"ewallet_type_name"`
	AccountName       string      `json:"account_name"`
	To                string      `json:"to"`
	// WalletAddress     string      `json:"wallet_address"`
}

func GetMemberAccountTransferExchangeBatchSetupv1(arrData MemberAccountTransferExchangeBatchSetupStruct) (*MemberAccountTransferExchangeBatchSetupRstStruct, error) {
	langCode := arrData.LangCode

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.withdrawal_with_crypto = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.EwalletType},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup == nil {
		base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv1-field_in_GetEwtSetupFn", "invalid_form_EwalletType", arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member.main_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ent_member.tagged_member_id > ? ", CondValue: 0},
	)
	arrMemberAccountList, _ := models.GetEntMemberListFn(arrCond, false)

	arrTransferSetupList := make([]transferSetupListStruct, 0)
	arrDataReturn := MemberAccountTransferExchangeBatchSetupRstStruct{
		TransferSetupList: arrTransferSetupList,
	}

	if len(arrMemberAccountList) > 0 {
		for _, arrMemberAccountListV := range arrMemberAccountList {
			entMemberID := arrMemberAccountListV.ID
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
				models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
				models.WhereCondFn{Condition: "ewt_setup.withdrawal_with_crypto = ?", CondValue: 1},
				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.EwalletType},
			)

			result, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrCond, "", false)

			if err != nil {
				base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv1-failed_in_GetMemberEwtSetupBalanceFn", err.Error(), arrCond, true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
			}

			if len(result) > 0 {
				translatedWalletName := helpers.Translate(result[0].EwtTypeName, langCode)

				availBal := result[0].Balance
				var transferSigningKeySetting map[string]interface{}
				var signingKeySetting interface{}
				var cryptoAddress string
				var toCryptoAddress string

				if result[0].Control == "BLOCKCHAIN" {
					// start transfer blockchain is involved
					cryptoAddress, _ = models.GetCustomMemberCryptoAddr(arrMemberAccountListV.ID, result[0].EwtTypeCode, true, false)
					BlkCWalBal := GetBlockchainWalletBalanceByAddressV1(result[0].EwtTypeCode, cryptoAddress, arrMemberAccountListV.ID)
					if arrMemberAccountListV.TaggedMemberID > 0 {
						db := models.GetDB() // no need set begin transaction
						toCryptoAddress, err = member_service.ProcessGetMemAddress(db, arrMemberAccountListV.TaggedMemberID, result[0].EwtTypeCode)
						if err != nil {
							arrErr := map[string]interface{}{
								"TaggedMemberID": arrMemberAccountListV.TaggedMemberID,
								"EwtTypeCode":    result[0].EwtTypeCode,
							}
							base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv1-ProcessGetMemAddress_failed", err.Error(), arrErr, true)
							return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
						}
						// cryptoAddress = toCryptoAddress
					}
					availBal = BlkCWalBal.AvailableBalance
					transferSigningKeySetting, _ = GetSigningKeySettingByModule(result[0].EwtTypeCode, cryptoAddress, "TRANSFER")

					// append sign key setting get from other function
					if transferSigningKeySetting != nil {
						transferSigningKeySetting["decimal_point"] = result[0].BlockchainDecimalPoint
						transferSigningKeySetting["contract_address"] = result[0].ContractAddress
						transferSigningKeySetting["is_base"] = false
						transferSigningKeySetting["method"] = "transfer"

						if result[0].IsBase == 1 {
							transferSigningKeySetting["is_base"] = true
						}

						transferSigningKeySetting["to_address"] = toCryptoAddress

						signingKeySetting = transferSigningKeySetting
					}
				} else {
					// so far this is not involved
					// start transfer no blockchain is involved
					cryptoAddress, err = models.GetCustomMemberCryptoAddr(arrMemberAccountListV.ID, result[0].EwtTypeCode, true, false)

					if err != nil {
						cryptoAddress = ""
					}

					if result[0].EwtTypeCode == "USDT" {
						cryptoAddress = ""
					}
					fmt.Println("cryptoAddress:", cryptoAddress)
				}

				if availBal <= 0 {
					continue
				}

				arrTransferSetupList = append(arrTransferSetupList,
					transferSetupListStruct{
						AccountName:       arrMemberAccountListV.NickName,
						EwtTypeCode:       result[0].EwtTypeCode,
						EwtTypeName:       translatedWalletName,
						AvailableBalance:  availBal,
						SigningKeySetting: signingKeySetting,
					},
				)
			}
		}
		if len(arrTransferSetupList) > 0 {
			arrDataReturn.TransferSetupList = arrTransferSetupList
		} else {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "all_account_no_balance_to_transfer"}
		}
	}
	return &arrDataReturn, nil
}

func GetMemberAccountTransferExchangeBatchSetupv2(arrData MemberAccountTransferExchangeBatchSetupStruct) (*MemberAccountTransferExchangeBatchSetupRstStruct, error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_from.ewallet_type_code = ?", CondValue: arrData.EwalletType},
		models.WhereCondFn{Condition: "ewt_from.member_show = ?", CondValue: 1},
	)

	arrTransferSetupRst, _ := models.GetEwtTransferSetupFn(arrCond, "", false)

	if len(arrTransferSetupRst) < 1 {
		base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv2-failed_in_GetEwtTransferSetupFn", arrCond, arrData, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.EwalletType},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup == nil {
		base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv2-field_in_GetEwtSetupFn", "invalid_form_EwalletType", arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	arrTransferSetupList := make([]transferSetupListStruct, 0)
	if ewtSetup.Control == "BLOCKCHAIN" {
		result, err := GetTransferExchangeBatchInBlockchain(arrData, ewtSetup)
		if err != nil {
			arrErr := map[string]interface{}{
				"arrData":  arrData,
				"ewtSetup": ewtSetup,
			}
			if err.Error() != "all_account_no_balance_to_transfer" {
				base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv2-field_in_GetTransferExchangeBatchInBlockchain", err.Error(), arrErr, true)
			}
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}
		arrTransferSetupList = result
	} else {
		result, err := GetTransferExchangeBatchInInternal(arrData, ewtSetup)
		if err != nil {
			arrErr := map[string]interface{}{
				"arrData":  arrData,
				"ewtSetup": ewtSetup,
			}
			if err.Error() != "all_account_no_balance_to_transfer" {
				base.LogErrorLog("GetMemberAccountTransferExchangeBatchSetupv2-field_in_GetTransferExchangeBatchInInternal", err.Error(), arrErr, true)
			}
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
		}
		arrTransferSetupList = result
	}
	arrDataReturn := MemberAccountTransferExchangeBatchSetupRstStruct{
		TransferSetupList: arrTransferSetupList,
	}

	return &arrDataReturn, nil
}

func GetTransferExchangeBatchInBlockchain(arrData MemberAccountTransferExchangeBatchSetupStruct, ewtSetup *models.EwtSetup) ([]transferSetupListStruct, error) {

	var (
		cryptoAddrList []string
		cryptoAddress  string
	)
	type WalletInfo struct {
		EntMemberID      int
		CryptoAddr       string
		AvailableBalance float64
		Nonce            int
	}

	arrWalletInfo := make([]WalletInfo, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member.main_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ent_member.tagged_member_id > ? ", CondValue: 0},
	)
	arrMemberAccountList, _ := models.GetEntMemberListFn(arrCond, false)
	if len(arrMemberAccountList) > 0 {
		for _, arrMemberAccountListV := range arrMemberAccountList {
			entMemberID := arrMemberAccountListV.ID
			cryptoAddress, _ = models.GetCustomMemberCryptoAddr(entMemberID, ewtSetup.EwtTypeCode, true, false)
			if cryptoAddress != "" {
				cryptoAddrList = append(cryptoAddrList, cryptoAddress)
			}
		}

		arrTransferSetupList := make([]transferSetupListStruct, 0)
		if len(cryptoAddrList) > 0 {

			// start get GetBlockchainWalBalWithAddressByBatch
			blockchainWalletBatchReturn := BlockchainWalletBatchStruct{
				TokenType:  ewtSetup.EwtTypeCode,
				AddressArr: cryptoAddrList,
				Limit:      blockchainBatchLimitApi,
			}
			arrBlockchainWalBal := blockchainWalletBatchReturn.GetBlockchainWalBalWithAddressByBatch()
			// end get GetBlockchainWalBalWithAddressByBatch

			cryptoAddrList = cryptoAddrList[:len(cryptoAddrList)-1] // truncate all slice value
			if len(arrBlockchainWalBal) > 0 {
				for _, arrBlockchainWalBalV := range arrBlockchainWalBal {
					if arrBlockchainWalBalV.AvailableBalance > 0 { // process only available balance more than 0 - filteration
						arrWalletInfo = append(arrWalletInfo, WalletInfo{
							CryptoAddr:       arrBlockchainWalBalV.Address,
							AvailableBalance: arrBlockchainWalBalV.AvailableBalance,
						})
						cryptoAddrList = append(cryptoAddrList, arrBlockchainWalBalV.Address)
					}
				}
			}

			if len(cryptoAddrList) > 0 {
				// start get ProcessGetBatchTransactionNonceViaAPI
				arrBatchTransactionNonce, err := ProcessGetBatchTransactionNonceViaAPI(cryptoAddrList)
				if err != nil {
					return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
				}
				// end get ProcessGetBatchTransactionNonceViaAPI
				// start merge arrBatchTransactionNonce into arrWalletInfo
				if len(arrBatchTransactionNonce) > 0 && len(arrWalletInfo) > 0 {
					for arrWalletInfoK, arrWalletInfov := range arrWalletInfo {
						for _, arrBatchTransactionNonceV := range arrBatchTransactionNonce {
							if arrWalletInfov.CryptoAddr == arrBatchTransactionNonceV.CryptoAddr {
								arrWalletInfo[arrWalletInfoK].Nonce = arrBatchTransactionNonceV.Nonce
							}
						}
					}
				}
				// end merge arrBatchTransactionNonce into arrWalletInfo
			}

			// start build info return api data
			if len(arrWalletInfo) > 0 {
				translatedWalletName := helpers.Translate(ewtSetup.EwtTypeName, arrData.LangCode)
				for _, arrWalletInfoV := range arrWalletInfo {
					transferSigningKeySetting, _ := GetSigningKeySettingByModule(ewtSetup.EwtTypeCode, arrWalletInfoV.CryptoAddr, "TRANSFER_BATCH")
					cryptoType := ewtSetup.EwtTypeCode
					// if strings.ToLower(cryptoType) == "usdt" || strings.ToLower(cryptoType) == "eth" {
					// 	cryptoType = "ETH"
					// }
					if strings.ToLower(cryptoType) == "liga" || strings.ToLower(cryptoType) == "sec" || strings.ToLower(cryptoType) == "usds" {
						cryptoType = "SEC"
					}
					arrCond := make([]models.WhereCondFn, 0)
					arrCond = append(arrCond,
						models.WhereCondFn{Condition: " member_crypto.crypto_address = ? ", CondValue: arrWalletInfoV.CryptoAddr},
						models.WhereCondFn{Condition: " member_crypto.status = ? ", CondValue: "A"},
						// models.WhereCondFn{Condition: " tagged_member_crypto.crypto_type = ? ", CondValue: cryptoType},
						// models.WhereCondFn{Condition: " tagged_member_crypto.status = ? ", CondValue: "A"},
					)
					arrTaggedMemberCryptoAddr, _ := models.GetTaggedMemberCryptoAddrFn(arrCond, false)

					if len(arrTaggedMemberCryptoAddr) == 1 {
						taggedCryptoAddress, _ := models.GetCustomMemberCryptoAddr(arrTaggedMemberCryptoAddr[0].TaggedMemberID, ewtSetup.EwtTypeCode, true, false)
						if taggedCryptoAddress != "" {
							// append sign key setting get from other function
							if transferSigningKeySetting != nil {
								transferSigningKeySetting["decimal_point"] = ewtSetup.BlockchainDecimalPoint
								transferSigningKeySetting["contract_address"] = ewtSetup.ContractAddress
								transferSigningKeySetting["is_base"] = false
								transferSigningKeySetting["method"] = "transfer"

								if ewtSetup.IsBase == 1 {
									transferSigningKeySetting["is_base"] = true
								}

								transferSigningKeySetting["to_address"] = taggedCryptoAddress
								transferSigningKeySetting["nonce"] = arrWalletInfoV.Nonce

								arrTransferSetupList = append(arrTransferSetupList,
									transferSetupListStruct{
										AccountName:       arrTaggedMemberCryptoAddr[0].NickName,
										EwtTypeCode:       ewtSetup.EwtTypeCode,
										EwtTypeName:       translatedWalletName,
										AvailableBalance:  arrWalletInfoV.AvailableBalance,
										SigningKeySetting: transferSigningKeySetting,
									},
								)
							}
						}
					}
				}
			}
			// end build info return api data
		}
		if len(arrTransferSetupList) > 0 {
			return arrTransferSetupList, nil
		} else {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "all_account_no_balance_to_transfer"}
		}
	}
	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_tagged_account"}
}

type LigaHoldingMatchAumReturn struct {
	Balance          float64
	ConvertedBalance float64
	ContractBalance  float64
}

func CheckMemberLigaHoldingMatchAum(EwalletTypeCode string, entMemberID int) (LigaHoldingMatchAumReturn, error) {

	contractBal := float64(0)

	//get member total contract balance
	contractBalRst, err := models.GetMemberTotalContractAmount(entMemberID)

	if err != nil {
		// base.LogErrorLog("CheckMemberLigaHoldingMatchAum- fail_to_get_member_contract_amount", err.Error(), contractBal, true)
		return LigaHoldingMatchAumReturn{}, err
	}

	contractBal = contractBalRst.TotalContractAmount

	//get blockchain balance
	address, err := models.GetCustomMemberCryptoAddr(entMemberID, EwalletTypeCode, true, false)
	if err != nil {
		// base.LogErrorLog("CheckMemberLigaHoldingMatchAum- fail_to_get_member_address", err.Error(), address, true)
		return LigaHoldingMatchAumReturn{}, err
	}

	balance, err := GetBlockchainWalletBalanceApiV1(EwalletTypeCode, address)
	if err != nil {
		// base.LogErrorLog("CheckMemberLigaHoldingMatchAum- fail to get LIGA balance", err.Error(), balance, true)
		return LigaHoldingMatchAumReturn{}, err
	}

	rate, err := base.GetLatestPriceMovementByTokenType(EwalletTypeCode)

	if err != nil {
		// base.LogErrorLog("CheckMemberLigaHoldingMatchAum -get LIGA price movement error", err, rate, true)
		return LigaHoldingMatchAumReturn{}, err
	}

	convBal, _ := decimal.NewFromFloat(balance).Mul(decimal.NewFromFloat(rate)).Float64()

	//get liga holding wallet
	arrHoldCond := make([]models.WhereCondFn, 0)
	arrHoldCond = append(arrHoldCond,
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "LIGAH"},
	)

	HoldResult, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrHoldCond, "", false)

	if err != nil {
		// base.LogErrorLog("CheckMemberLigaHoldingMatchAum() -get LIGA Holding Wallet error", err, HoldResult, true)
		return LigaHoldingMatchAumReturn{}, err
	}

	if len(HoldResult) > 0 {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(HoldResult[0].Balance)).Float64()
		if balance < 0 {
			balance = float64(0)
		}
		convBal, _ = decimal.NewFromFloat(balance).Mul(decimal.NewFromFloat(rate)).Float64()
	}

	return LigaHoldingMatchAumReturn{
		Balance:          balance,
		ConvertedBalance: convBal,
		ContractBalance:  contractBal,
	}, nil
}

type BlockchainWalletBatchStruct struct {
	TokenType  string
	AddressArr []string
	Limit      int
}

type BlockchainWalletBatchReturnStruct struct {
	Address          string
	AvailableBalance float64
	Balance          float64
}

// get blockchain wallet balance by batch - pass in address array, wallet_type_code,limit (only for koo usage, dun call this func)
func (w *BlockchainWalletBatchStruct) GetBlockchainWalBalWithAddressByBatch() []BlockchainWalletBatchReturnStruct {

	type ApiResponseStruct struct {
		Status     string            `json:"status"`
		StatusCode string            `json:"status_code"`
		Msg        string            `json:"msg"`
		Data       map[string]string `json:"data"`
	}

	secApiSetting, _ := models.GetSysGeneralSetupByID("sec_api_setting")

	batch := 1

	limit := w.Limit

	if limit == 0 {
		limit = 100 //default 100
	}

	totalRecord := len(w.AddressArr)

	totalBatch := float64(totalRecord) / float64(limit)
	totalBatch = math.Ceil(totalBatch)
	arrData := make([]BlockchainWalletBatchReturnStruct, 0)

	for i := batch; i <= int(totalBatch); i++ {

		page := i
		curBatch := page

		if curBatch == 0 {
			curBatch = 1
		}

		if page != 0 {
			page--
		}

		pageStart, pageEnd := helpers.Paginate(int(page), int(limit), totalRecord)
		batchDataArr := w.AddressArr[pageStart:pageEnd]

		//call balance

		var (
			response     ApiResponseStruct
			decimalPoint uint
		)

		data := map[string]interface{}{
			"token_type": strings.ToUpper(w.TokenType),
			"addresses":  batchDataArr,
		}

		api_key := secApiSetting.InputValue2

		url := secApiSetting.InputValue1 + "api/account/batch/balance"
		header := map[string]string{
			"Content-Type":    "application/json",
			"X-Authorization": api_key,
		}

		res, err_api := base.RequestAPI("POST", url, header, data, &response)

		if err_api != nil {
			base.LogErrorLogV2("GetBlockchainWalBalWithAddressByBatch -fail to call blockchain balance api", err_api.Error(), map[string]interface{}{"err": err_api, "data": data}, true, "blockchain")
			return nil
		}

		if res.StatusCode != 200 {
			errMsg, _ := json.Marshal(response.Msg)
			errMsgStr := string(errMsg)
			errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
			base.LogErrorLogV2("GetBlockchainWalBalWithAddressByBatch -fail to get blockchain wallet balance", errMsgStr, map[string]interface{}{"err": res.Body, "data": data}, true, "blockchain")
			return nil
		}

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(w.TokenType)},
		)
		ewtSetup, err := models.GetEwtSetupFn(arrCond, "", false)

		if err != nil {
			base.LogErrorLog("GetBlockchainWalBalWithAddressByBatch -fail to get wallet setup", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
			return nil
		}

		decimalPoint = uint(8)
		if ewtSetup != nil {
			decimalPoint = uint(ewtSetup.DecimalPoint)
		}

		if len(response.Data) > 0 {
			for k, v := range response.Data {
				balance := float64(0)
				strBal := string(v)
				strBal = strings.Replace(strBal, "\"", "", 2)
				strBal = helpers.CutOffStringsDecimal(strBal, decimalPoint, '.')
				balance, _ = strconv.ParseFloat(strBal, 64)
				strBalance := helpers.CutOffDecimal(balance, decimalPoint, ".", "")
				balance, _ = strconv.ParseFloat(strBalance, 64)
				availBal := balance
				address := k

				//get address member
				arrMemCond := make([]models.WhereCondFn, 0)
				arrMemCond = append(arrMemCond,
					models.WhereCondFn{Condition: "ent_member_crypto.crypto_address = ?", CondValue: k},
					models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
				)
				arrExistingMemCrypto, err := models.GetEntMemberCryptoFn(arrMemCond, false)

				if err != nil {
					base.LogErrorLog("GetBlockchainWalBalWithAddressByBatch -get existing member crypto address error", err, map[string]interface{}{"err": err, "data": arrMemCond}, true)
					availBal = float64(0)
				}

				if arrExistingMemCrypto == nil {
					base.LogErrorLog("GetBlockchainWalBalWithAddressByBatch -fail to get address member", arrMemCond, arrExistingMemCrypto, true)
					availBal = float64(0)
				} else {
					entMemberID := arrExistingMemCrypto.MemberID

					//koo request add holding for available bal 20210602
					arrHoldCond := make([]models.WhereCondFn, 0)
					arrHoldCond = append(arrHoldCond,
						models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
						models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(w.TokenType) + "H"},
					)

					HoldResult, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrHoldCond, "", false)

					if err != nil {
						base.LogErrorLog("GetBlockchainWalBalWithAddressByBatch -get Holding Wallet error", err, arrHoldCond, true)
					}

					if len(HoldResult) > 0 {
						availBal, _ = decimal.NewFromFloat(availBal).Add(decimal.NewFromFloat(HoldResult[0].Balance)).Float64()
					}

					//get blockchain_trans -pending & log_only = 0
					PendingAmt, PendingAmtError := models.GetTotalPendingBlockchainAmount(entMemberID, strings.ToUpper(w.TokenType))
					if PendingAmtError != nil {
						base.LogErrorLog("GetBlockchainWalBalWithAddressByBatch -fail to get Blockchain Trans Pending Amount", PendingAmtError, map[string]interface{}{"err": PendingAmtError, "mem_id": entMemberID, "wallet_type": strings.ToUpper(w.TokenType)}, true)
					}

					if PendingAmt != nil {
						availBal, _ = decimal.NewFromFloat(availBal).Sub(decimal.NewFromFloat(PendingAmt.TotalPendingAmount)).Float64()
					}

					if availBal < 0 {
						availBal = float64(0)
					}
				}

				strAvailBal := helpers.CutOffDecimal(availBal, decimalPoint, ".", "")
				availBal, _ = strconv.ParseFloat(strAvailBal, 64)

				arrData = append(arrData,
					BlockchainWalletBatchReturnStruct{
						Address:          address,
						Balance:          balance,
						AvailableBalance: availBal,
					},
				)
			}
		}

	}

	return arrData
}

// GetBatchTransactionNonceViaAPI func
func GetBatchTransactionNonceViaAPI(cryptoAddrList []string) (map[string]int, error) {

	settingID := "batch_nonce_api_setting"
	arrApiSetting, _ := models.GetSysGeneralSetupByID(settingID)

	if arrApiSetting.InputType1 == "1" {

		type apiRstStruct struct {
			Status     string         `json:"status"`
			StatusCode string         `json:"status_code"`
			Msg        string         `json:"msg"`
			Data       map[string]int `json:"data"`
		}

		header := map[string]string{
			"Content-Type":    "application/json",
			"X-Authorization": arrApiSetting.InputType2,
		}
		data := map[string]interface{}{
			"addresses": cryptoAddrList,
		}
		res, err_api := base.RequestAPI(arrApiSetting.SettingValue1, arrApiSetting.InputValue1, header, data, nil)

		if err_api != nil {
			base.LogErrorLogV2("GetBatchTransactionNonceViaAPI-error_in_api_call_before_call", err_api.Error(), nil, true, "blockchain")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "batch_nonce_is_not_available", Data: ""}
		}

		if res.StatusCode != 200 {
			base.LogErrorLogV2("GetBatchTransactionNonceViaAPI-error_in_api_call_after_call", res.Body, nil, true, "blockchain")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "batch_nonce_is_not_available", Data: ""}
		}

		var apiRst apiRstStruct
		err := json.Unmarshal([]byte(res.Body), &apiRst)

		if err != nil {
			base.LogErrorLog("GetBatchTransactionNonceViaAPI-error_in_json_decode_api_result", err_api.Error(), res.Body, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "batch_nonce_is_not_available", Data: ""}
		}
		return apiRst.Data, nil

	}

	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "batch_nonce_is_not_available", Data: ""}
}

type ProcessGetBatchTransactionNonceViaAPIRst struct {
	CryptoAddr string
	Nonce      int
}

func ProcessGetBatchTransactionNonceViaAPI(cryptoAddrList []string) ([]ProcessGetBatchTransactionNonceViaAPIRst, error) {

	totalCryptoAddr := len(cryptoAddrList)
	if totalCryptoAddr == 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_crypto_address_in_list"}
	}

	numOfLooping := math.Ceil(float.Div(float64(totalCryptoAddr), float64(blockchainBatchLimitApi)))

	var startArrayIndex int
	arrDataReturn := make([]ProcessGetBatchTransactionNonceViaAPIRst, 0)
	for i := 1; i <= int(numOfLooping); i++ {

		lastArrayIndex := i * blockchainBatchLimitApi

		if int(numOfLooping) == 1 || i == int(numOfLooping) {
			// if only loop 1 times or current looping is last loop, set the last index equal to totalCryptoAddr
			lastArrayIndex = totalCryptoAddr
		}

		currentCryptoAddrList := cryptoAddrList[startArrayIndex:lastArrayIndex]

		batchTransactionNonceViaAPIRst, err := GetBatchTransactionNonceViaAPI(currentCryptoAddrList)
		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		for _, currentCryptoAddrListV := range currentCryptoAddrList {
			for batchTransactionNonceViaAPIRstK, batchTransactionNonceViaAPIRstV := range batchTransactionNonceViaAPIRst {
				if currentCryptoAddrListV == batchTransactionNonceViaAPIRstK {
					arrDataReturn = append(arrDataReturn,
						ProcessGetBatchTransactionNonceViaAPIRst{
							CryptoAddr: currentCryptoAddrListV,
							Nonce:      batchTransactionNonceViaAPIRstV,
						},
					)
				}
			}
		}
		startArrayIndex = lastArrayIndex
	}
	return arrDataReturn, nil
}

func GetPendingTransferIn(memID int) string {
	pendingTransactionsCond := make([]models.WhereCondFn, 0)
	pendingTransactionsCond = append(pendingTransactionsCond,
		models.WhereCondFn{Condition: "blockchain_adjust_out.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "blockchain_adjust_out.status = ?", CondValue: "P"},
		models.WhereCondFn{Condition: "blockchain_adjust_out.transaction_type = ?", CondValue: "ADJUST_IN"},
	)

	pendingTransactions, pendingTransactionsErr := models.GetBlockchainAdjustOutSumFn(pendingTransactionsCond, false)

	if pendingTransactionsErr != nil {
		base.LogErrorLog("GetPendingTransferOut()", "GetBlockchainAdjustOutSumFn()", pendingTransactionsErr, true)
		return "something_went_wrong"
	}

	if len(pendingTransactions) > 0 {
		// get company crypto address and private key
		arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
		arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
			models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: 0},
			models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
		)
		arrCompanyCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)
		if arrCompanyCrypto == nil {
			base.LogErrorLog("walletService:GetPendingTransferOut()", "GetEntMemberCryptoFn():1", "company_address_not_found", true)
			return "something_went_wrong"
		}
		companyAddress := arrCompanyCrypto.CryptoAddress
		companyPrivateKey := arrCompanyCrypto.PrivateKey

		// get member crypto address
		arrEntMemberCryptoFn = make([]models.WhereCondFn, 0)
		arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
			models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
		)
		memberCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)
		if memberCrypto == nil {
			base.LogErrorLog("walletService:GetPendingTransferOut()", "GetEntMemberCryptoFn():1", "company_address_not_found", true)
			return "something_went_wrong"
		}
		memberAddress := memberCrypto.CryptoAddress

		for _, tran := range pendingTransactions {

			signingKeySetting, signingKeySettingErr := GetSigningKeySettingByModule(tran.EwalletTypeCode, memberAddress, "TRANSFER")

			if signingKeySettingErr != "" {
				base.LogErrorLog("walletService:GetPendingTransferOut()", "GetEntMemberCryptoFn():1", signingKeySettingErr, true)
				return "something_went_wrong"
			}

			chainID, _ := helpers.ValueToInt(signingKeySetting["chain_id"].(string))
			maxGas, _ := helpers.ValueToInt(signingKeySetting["max_gas"].(string))
			// generate signing key
			signingKey, signingKeyErr := ProcecssGenerateSignTransaction(ProcecssGenerateSignTransactionStruct{
				TokenType:       tran.EwalletTypeCode,
				PrivateKey:      companyPrivateKey,
				ContractAddress: "",
				ChainID:         int64(chainID),
				FromAddr:        companyAddress,
				ToAddr:          memberAddress,
				Amount:          tran.TotalIn,
				MaxGas:          uint64(maxGas),
			})

			if signingKeyErr != nil {
				base.LogErrorLog("walletService:GetPendingTransferOut()", "ProcecssGenerateSignTransaction():1", signingKeyErr.Error(), true)
				return signingKeySettingErr
			}

			err := AdjustOut(memID, AdjustOutData{
				TransactionData: signingKey,
				TransactionType: tran.TransactionType,
				TokenType:       tran.EwalletTypeCode,
				TransactionIds:  tran.TransactionIds,
			})

			if err != "" {
				return err
			}
		}
	}

	return ""
}

type GetPendingTransferOutReturn struct {
	TokenType       string `json:"token_type"`
	TotalOut        string `json:"total_out"`
	ToAddress       string `json:"to_address"`
	TransactionType string `json:"transaction_type"`
	TransactionIds  string `json:"transaction_ids"`
}

func GetPendingTransferOut(memID int) ([]GetPendingTransferOutReturn, string) {
	// get pending adjust in
	pendingInErr := GetPendingTransferIn(memID)

	if pendingInErr != "" {
		return nil, pendingInErr
	}

	pendingTransactionsCond := make([]models.WhereCondFn, 0)
	pendingTransactionsCond = append(pendingTransactionsCond,
		models.WhereCondFn{Condition: "blockchain_adjust_out.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "blockchain_adjust_out.status = ?", CondValue: "P"},
		models.WhereCondFn{Condition: "blockchain_adjust_out.transaction_type = ?", CondValue: "ADJUST"},
	)

	pendingTransactions, pendingTransactionsErr := models.GetBlockchainAdjustOutSumFn(pendingTransactionsCond, false)

	if pendingTransactionsErr != nil {
		base.LogErrorLog("GetPendingTransferOut()", "GetBlockchainAdjustOutSumFn()", pendingTransactionsErr, true)
		return nil, "something_went_wrong"
	}

	// get company address
	companyAddr, companyAddrErr := models.GetMemberCryptoByMemID(0, "SEC")

	if companyAddrErr != nil {
		base.LogErrorLog("GetPendingTransferOut()", "GetMemberCryptoByMemID()", companyAddrErr, true)
		return nil, "something_went_wrong"
	}

	returnData := make([]GetPendingTransferOutReturn, 0)
	for _, trans := range pendingTransactions {
		returnData = append(returnData, GetPendingTransferOutReturn{
			TokenType:       trans.EwalletTypeCode,
			TotalOut:        helpers.CutOffDecimal(trans.TotalOut, 8, ".", ""),
			ToAddress:       companyAddr.CryptoAddress,
			TransactionType: trans.TransactionType,
			TransactionIds:  trans.TransactionIds,
		})
	}

	return returnData, ""
}

type AdjustOutData struct {
	TransactionData string `json:"transaction_data"`
	TransactionType string `json:"transaction_type"`
	TokenType       string `json:"token_type"`
	TransactionIds  string `json:"transaction_ids"`
}

func AdjustOut(memID int, data AdjustOutData) string {
	transactionIds := strings.Split(data.TransactionIds, ",")

	// check if have records in blockchain trans, cancel transaction
	condition := ""
	condValue := ""
	for i, transactionId := range transactionIds {
		if i == 0 {
			condition = "doc_no = ?"
			condValue = transactionId
		} else {
			condition += " OR doc_no = " + transactionId
		}
	}
	bcTransCond := make([]models.WhereCondFn, 0)
	bcTransCond = append(bcTransCond,
		models.WhereCondFn{Condition: condition, CondValue: condValue},
	)

	bcTrans, bcTransErr := models.GetBlockchainTransArrayFn(bcTransCond, false)

	if bcTransErr != nil {
		base.LogErrorLog("wallet_service:AdjustOut()", "GetBlockchainTransArrayFn()", bcTransErr, true)
		return "something_went_wrong"
	}

	if len(bcTrans) > 0 {
		return "blockchain_trans_record_exists"
	}

	hashValue, errMsg := SignedTransaction(data.TransactionData)

	if errMsg != "" {
		return errMsg
	}
	// hashValue := "000"

	for _, transactionId := range transactionIds {
		pendingTransactionsCond := make([]models.WhereCondFn, 0)
		pendingTransactionsCond = append(pendingTransactionsCond,
			models.WhereCondFn{Condition: "blockchain_adjust_out.id = ?", CondValue: transactionId},
		)

		pendingTransactions, pendingTransactionsErr := models.GetBlockchainAdjustOutFn(pendingTransactionsCond, false)

		if pendingTransactionsErr != nil {
			base.LogErrorLog("wallet_service:AdjustOut()", "GetBlockchainAdjustOutFn()", pendingTransactionsErr, true)
			return "something_went_wrong"
		}

		if len(pendingTransactions) <= 0 {
			base.LogErrorLog("wallet_service:AdjustOut()", "pending_transaction_not_found", transactionId, true)
			return "pending_transaction_not_found"
		}

		pendingTransaction := pendingTransactions[0]

		// log blochchain transaction
		db := models.GetDB() // will save record regardless of begintransaction since api already called successfully
		_, err := models.AddBlockchainTrans(db, models.AddBlockchainTransStruct{
			MemberID:          pendingTransaction.MemberID,
			EwalletTypeID:     pendingTransaction.EwalletTypeID,
			DocNo:             transactionId,
			Status:            pendingTransaction.Status,
			TransactionType:   pendingTransaction.TransactionType,
			TotalIn:           pendingTransaction.TotalIn,
			TotalOut:          pendingTransaction.TotalOut,
			ConversionRate:    pendingTransaction.ConversionRate,
			ConvertedTotalIn:  pendingTransaction.ConvertedTotalIn,
			ConvertedTotalOut: pendingTransaction.ConvertedTotalOut,
			TransactionData:   data.TransactionData,
			HashValue:         hashValue,
			Remark:            pendingTransaction.Remark,
		})

		if err != nil {
			base.LogErrorLog("walletService:AdjustOut()", "AddBlockchainTrans():1", err.Error(), true)
			return "something_went_wrong"
		}
	}

	return ""
}

func AdjustCallback(tx *gorm.DB, hash string, status string) string {
	// get blockchain trans record
	bcTransCond := make([]models.WhereCondFn, 0)
	bcTransCond = append(bcTransCond,
		models.WhereCondFn{Condition: "hash_value = ?", CondValue: hash},
	)

	bcTrans, bcTransErr := models.GetBlockchainTransArrayFn(bcTransCond, false)

	if bcTransErr != nil {
		base.LogErrorLog("wallet_service:AdjustCallback()", "GetBlockchainTransArrayFn()", bcTransErr, true)
		return "something_went_wrong"
	}

	for _, bcTran := range bcTrans {
		// update adjust out record
		AdjOutCond := make([]models.WhereCondFn, 0)
		AdjOutCond = append(AdjOutCond,
			models.WhereCondFn{Condition: "id = ?", CondValue: bcTran.DocNo},
		)
		updateColumn := map[string]interface{}{"status": status}
		err := models.UpdatesFnTx(tx, "blockchain_adjust_out", AdjOutCond, updateColumn, false)

		if err != nil {
			base.LogErrorLog("wallet_service:AdjustCallback()", "UpdateAdjustOut", err.Error(), true)
			return "something_went_wrong"
		}
	}

	return ""
}

// GetLatestExchangePriceMovementByEwtTypeCode func
func GetLatestExchangePriceMovementByEwtTypeCode(ewtTypeCode string) (float64, string) {
	if strings.ToLower(ewtTypeCode) == "sec" {
		tokenRate, err := models.GetLatestExchangePriceMovementSec()

		if err != nil {
			base.LogErrorLog("walletService:GetLatestExchangePriceMovementByEwtTypeCode()", "GetLatestExchangePriceMovementSec():1", err.Error(), true)
			return 0.00, "something_went_wrong"
		}

		return tokenRate, ""
	} else if strings.ToLower(ewtTypeCode) == "liga" {
		tokenRate, err := models.GetLatestExchangePriceMovementLiga()

		if err != nil {
			base.LogErrorLog("walletService:GetLatestExchangePriceMovementByEwtTypeCode()", "GetLatestExchangePriceMovementLiga():1", err.Error(), true)
			return 0.00, "something_went_wrong"
		}

		return tokenRate, ""
	} else {
		base.LogErrorLog("walletService:GetLatestExchangePriceMovementByEwtTypeCode()", "invalid_ewt_type_code_"+ewtTypeCode, "", true)
		return 0.00, "something_went_wrong"
	}
}

func GetMemberBlockchainWalletBalance(memID int, cryptoType, cryptoAddress string) (BlockchainWalletReturnStruct, string) {
	db := models.GetDB() // no need set begin transaction
	retrievedCryptoAddr, err := member_service.ProcessGetMemAddress(db, memID, cryptoType)
	if err != nil {
		base.LogErrorLog("walletService:GetMemberBlockchainWalletBalance()", "ProcessGetMemAddress():1", err.Error(), true)
		return BlockchainWalletReturnStruct{}, "something_went_wrong"
	}

	// validate retrieved crypto address with the provided crypto address
	if retrievedCryptoAddr != cryptoAddress {
		return BlockchainWalletReturnStruct{}, "invalid_crypto_address"
	}

	arrBlkCWalBal := GetBlockchainWalletBalanceByAddressV1(cryptoType, cryptoAddress, memID)

	return arrBlkCWalBal, ""
}

type WalletBalanceApi struct {
	CryptoCode   string  `json:"crypto_code"`
	TotalBalance float64 `json:"total_balance"`
}

type WalletBalanceApiRstStruct struct {
	WalletBalanceList []WalletBalanceApi `json:"wallet_balance_list"`
}

func GetWalletBalanceListApiv1(entMemberID int, ewtTypeCode string, langCode string) (*WalletBalanceApiRstStruct, error) {

	walletBalanceApiList := make([]WalletBalanceApi, 0)
	arrDataReturn := WalletBalanceApiRstStruct{WalletBalanceList: walletBalanceApiList}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: " ewallet_type_code IN (" + ewtTypeCode + ") AND ewt_setup.status = ?", CondValue: "A"},
	)

	result, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrCond, "", true)

	if err != nil {
		base.LogErrorLog("GetWalletBalanceListApiv1-GetMemberEwtSetupBalanceFn", err.Error(), arrCond, true)
		return &arrDataReturn, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if len(result) > 0 {
		for _, v := range result {
			totalBalance := v.Balance
			if v.Control == "BLOCKCHAIN" {
				address, err := models.GetCustomMemberCryptoAddr(entMemberID, v.EwtTypeCode, true, false)
				if err != nil {
					arrErr := map[string]interface{}{
						"entMemberID": entMemberID,
						"EwtTypeCode": v.EwtTypeCode,
					}
					base.LogErrorLog("GetWalletBalanceListApiv1-GetCustomMemberCryptoAddr_failed", err.Error(), arrErr, true)
					return &arrDataReturn, err
				}

				BlkCWalBal := GetBlockchainWalletBalanceByAddressV1(v.EwtTypeCode, address, entMemberID) // get liga & sec bal
				totalBalance = BlkCWalBal.AvailableBalance
			}

			walletBalanceApiList = append(walletBalanceApiList,
				WalletBalanceApi{
					CryptoCode:   v.EwtTypeCode,
					TotalBalance: totalBalance,
				},
			)
		}
	}
	arrDataReturn = WalletBalanceApiRstStruct{WalletBalanceList: walletBalanceApiList}
	return &arrDataReturn, nil
}

func GetPriceListApiv1() map[string]interface{} {

	secExchangePrice, _ := GetLatestExchangePriceMovementByEwtTypeCode("SEC")
	ligaExchangePrice, _ := GetLatestExchangePriceMovementByEwtTypeCode("LIGA")

	secSystemPrice, _ := models.GetLatestSecPriceMovement()
	ligaSystemPrice, _ := models.GetLatestLigaPriceMovement()

	systemPriceList := map[string]interface{}{
		"SECUSDT":  secSystemPrice,
		"LIGAUSDT": ligaSystemPrice,
	}

	exchangePriceList := map[string]interface{}{
		"SECUSDT":  secExchangePrice,
		"LIGAUSDT": ligaExchangePrice,
	}

	arrDataReturn := map[string]interface{}{
		"exchange_price": exchangePriceList,
		"system_price":   systemPriceList,
	}

	return arrDataReturn
}

func GetWalletSigningKeySetting(entMemberID int, ewtTypeCode string, method string) (map[string]interface{}, error) {

	var (
		err error
	)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: ewtTypeCode},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	)
	ewtSetup, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetWalletSigningKeySetting - get wallet setup fail", arrCond, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: arrCond}
	}

	var signingKeySetting map[string]interface{}
	if ewtSetup.Control == "BLOCKCHAIN" {

		address, err := models.GetCustomMemberCryptoAddr(entMemberID, ewtTypeCode, false, false)
		if err != nil {
			base.LogErrorLog("GetWalletSigningKeySetting - fail to get member address ", map[string]interface{}{"member_id": entMemberID, "wallet_type_code": ewtTypeCode}, err.Error(), true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: map[string]interface{}{"member_id": entMemberID, "wallet_type_code": ewtTypeCode}}
		}

		signingKeySetting, _ = GetSigningKeySettingByModule(ewtTypeCode, address, method)

		// append sign key setting get from other function
		if signingKeySetting != nil {
			signingKeySetting["decimal_point"] = ewtSetup.BlockchainDecimalPoint
			signingKeySetting["contract_address"] = ewtSetup.ContractAddress
			signingKeySetting["is_base"] = false
			signingKeySetting["method"] = "transfer"

			if ewtSetup.IsBase == 1 {
				signingKeySetting["is_base"] = true
			}
		}
	}

	return signingKeySetting, nil
}

type AllBlockchainWalletReturnStruct struct {
	WalletTypeCode string
	Balance        float64
}

func GetAllBlockchainWalletBalanceByAddressV1(address string) ([]AllBlockchainWalletReturnStruct, error) {
	var (
		err          error
		response     app.ApiArrayResponse
		decimalPoint uint
	)

	secApiSetting, _ := models.GetSysGeneralSetupByID("sec_api_setting")

	data := map[string]interface{}{
		"address": address,
	}

	api_key := secApiSetting.InputValue2

	url := secApiSetting.InputValue1 + "api/account/balance/all"
	header := map[string]string{
		"Content-Type":    "application/json",
		"X-Authorization": api_key,
	}

	res, err := base.RequestAPI("POST", url, header, data, &response)
	if err != nil {
		base.LogErrorLogV2("GetAllBlockchainWalletBalanceByAddressV1 -fail to call blockchain balance api", err.Error(), map[string]interface{}{"err": err, "data": data}, true, "blockchain")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "call_get_all_blockchain_wal_bal_api_failed", Data: ""}
	}

	if res.StatusCode != 200 {
		errMsg, _ := json.Marshal(response.Msg)
		errMsgStr := string(errMsg)
		errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		base.LogErrorLogV2("GetAllBlockchainWalletBalanceByAddressV1 -fail to get blockchain wallet balance", errMsgStr, map[string]interface{}{"err": res.Body, "data": data}, true, "blockchain")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "blockchain_wal_bal_api_return_error", Data: ""}
	}

	arrDataReturn := make([]AllBlockchainWalletReturnStruct, 0)
	decimalPoint = uint(8) //default 8
	if len(response.Data) > 0 {
		for _, v := range response.Data {

			walTypeCode, _ := json.Marshal(v["token_type"])
			strWalTypeCode := string(walTypeCode)
			strWalTypeCode = strings.Replace(strWalTypeCode, "\"", "", 2)

			//get wallet setup
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(strWalTypeCode)},
			)
			ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

			// if err != nil {
			// 	base.LogErrorLog("GetAllBlockchainWalletBalanceByAddressV1 -fail to get wallet setup", err.Error(), map[string]interface{}{"err": err, "data": v, "pass_data": data}, true)
			// }

			if ewtSetup != nil {
				decimalPoint = uint(ewtSetup.DecimalPoint)
			}

			return_balance, _ := json.Marshal(v["balance"])
			strBal := string(return_balance)
			strBal = strings.Replace(strBal, "\"", "", 2)
			strBal = helpers.CutOffStringsDecimal(strBal, decimalPoint, '.') //cut decimal to our sys decimal places

			bal := float64(0)
			bal, err = strconv.ParseFloat(strBal, 64)

			if err != nil {
				base.LogErrorLog("GetAllBlockchainWalletBalanceByAddressV1 -fail to convert balance to float64 after cut off strings", err.Error(), map[string]interface{}{"err": err, "data": v, "pass_data": data}, true)
			}

			strBalance := helpers.CutOffDecimal(bal, decimalPoint, ".", "")
			balance := float64(0)
			balance, err = strconv.ParseFloat(strBalance, 64)

			if err != nil {
				base.LogErrorLog("GetAllBlockchainWalletBalanceByAddressV1 -fail to convert balance to float64", err.Error(), map[string]interface{}{"err": err, "data": v, "pass_data": data}, true)
			}

			arrDataReturn = append(arrDataReturn, AllBlockchainWalletReturnStruct{
				WalletTypeCode: strWalTypeCode,
				Balance:        balance,
			})
		}

	}

	return arrDataReturn, nil
}

type AllBlockchainWalletWithOtherBalReturnStruct struct {
	WalletTypeCode            string
	Rate                      float64
	Balance                   float64
	ConvertedBalance          float64
	AvailableBalance          float64
	ConvertedAvailableBalance float64
	HoldingBalance            float64
	WithHoldingBalance        float64
}

func GetAllBlockchainWalletBalanceWithOtherBalV1(address string, entMemberID int) ([]AllBlockchainWalletWithOtherBalReturnStruct, error) {
	var (
		err          error
		decimalPoint uint
	)

	arrDataReturn := make([]AllBlockchainWalletWithOtherBalReturnStruct, 0)
	decimalPoint = uint(8)
	convertedBalance := float64(0)
	availBal := float64(0)
	convertedAvailBal := float64(0)
	holdingBal := float64(0)
	withHoldingBal := float64(0)
	rst, err := GetAllBlockchainWalletBalanceByAddressV1(address)

	if err != nil {
		return nil, err
	}

	if len(rst) > 0 {
		for _, v := range rst {
			holdingBal = float64(0)
			withHoldingBal = float64(0)

			//get wallet setup - Holding
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(v.WalletTypeCode) + "H"},
			)
			ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

			if ewtSetup != nil {
				decimalPoint = uint(ewtSetup.DecimalPoint)
			}

			token_rate, err := base.GetLatestPriceMovementByTokenType(v.WalletTypeCode)
			if err != nil {
				base.LogErrorLog("GetAllBlockchainWalletBalanceWithOtherBalV1 -get price movement error", err, map[string]interface{}{"wallet_type_code": v.WalletTypeCode}, true)
			}

			balance := v.Balance
			availBal = balance
			rate := token_rate
			convertedAvailBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
			convertedBalance, _ = decimal.NewFromFloat(balance).Mul(decimal.NewFromFloat(rate)).Float64()

			//get holding wallet
			arrHoldCond := make([]models.WhereCondFn, 0)
			arrHoldCond = append(arrHoldCond,
				models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(v.WalletTypeCode) + "H"},
			)

			HoldResult, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrHoldCond, "", false)

			if err != nil {
				base.LogErrorLog("GetAllBlockchainWalletBalanceByAddressV1 -get Holding Wallet error", err, arrHoldCond, true)
			}

			if len(HoldResult) > 0 {
				balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(HoldResult[0].Balance)).Float64()
				if balance < 0 {
					balance = float64(0)
				}
				convertedBalance, _ = decimal.NewFromFloat(balance).Mul(decimal.NewFromFloat(rate)).Float64()
				holdingBal = HoldResult[0].Balance
			}

			//get withholding wallet
			arrWithHoldCond := make([]models.WhereCondFn, 0)
			arrWithHoldCond = append(arrWithHoldCond,
				models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "W" + strings.ToUpper(v.WalletTypeCode)},
			)

			WithHoldResult, err := models.GetMemberEwtSetupBalanceFn(entMemberID, arrWithHoldCond, "", false)

			if err != nil {
				base.LogErrorLog("GetAllBlockchainWalletBalanceByAddressV1 -get WithHolding Wallet error", err, arrWithHoldCond, true)
			}

			if len(WithHoldResult) > 0 {
				balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(WithHoldResult[0].Balance)).Float64()
				if balance < 0 {
					balance = float64(0)
				}
				convertedBalance, _ = decimal.NewFromFloat(balance).Mul(decimal.NewFromFloat(rate)).Float64()
				withHoldingBal = WithHoldResult[0].Balance
			}

			//get blockchain_trans -pending & log_only = 0
			PendingAmt, err := models.GetTotalPendingBlockchainAmount(entMemberID, strings.ToUpper(v.WalletTypeCode))
			if err != nil {
				base.LogErrorLog("GetAllBlockchainWalletBalanceByAddressV1 -get Blockchain Trans Pending Amount error", err, v, true)
			}

			if PendingAmt != nil {
				if PendingAmt.TotalPendingAmount != 0 {
					availBal, _ = decimal.NewFromFloat(availBal).Sub(decimal.NewFromFloat(PendingAmt.TotalPendingAmount)).Float64()
					convertedAvailBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
				}
			}

			//get blockchain_adjust_out where pending
			PendingAdjOutAmt, err := models.GetTotalPendingBlockchainAdjustOutAmount(entMemberID, strings.ToUpper(v.WalletTypeCode))
			if err != nil {
				base.LogErrorLog("GetAllBlockchainWalletBalanceByAddressV1 -get Blockchain Adjust Out Pending Amount error", err, v, true)
			}

			if PendingAdjOutAmt != nil {
				if PendingAdjOutAmt.TotalPendingAmount != 0 {
					availBal, _ = decimal.NewFromFloat(availBal).Sub(decimal.NewFromFloat(PendingAdjOutAmt.TotalPendingAmount)).Float64()
					convertedAvailBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
				}
			}

			if availBal < 0 {
				availBal = float64(0)
				convertedAvailBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
			}

			if availBal > balance {
				availBal = balance
				convertedAvailBal, _ = decimal.NewFromFloat(availBal).Mul(decimal.NewFromFloat(rate)).Float64()
			}

			strBalance := helpers.CutOffDecimal(balance, decimalPoint, ".", "")
			balance, _ = strconv.ParseFloat(strBalance, 64)

			strConvBal := helpers.CutOffDecimal(convertedBalance, decimalPoint, ".", "")
			convertedBalance, _ = strconv.ParseFloat(strConvBal, 64)

			strAvailBal := helpers.CutOffDecimal(availBal, decimalPoint, ".", "")
			availBal, _ = strconv.ParseFloat(strAvailBal, 64)

			strAvailConvBal := helpers.CutOffDecimal(convertedAvailBal, decimalPoint, ".", "")
			convertedAvailBal, _ = strconv.ParseFloat(strAvailConvBal, 64)

			strHolding := helpers.CutOffDecimal(holdingBal, decimalPoint, ".", "")
			holdingBal, _ = strconv.ParseFloat(strHolding, 64)

			strWithHolding := helpers.CutOffDecimal(withHoldingBal, decimalPoint, ".", "")
			withHoldingBal, _ = strconv.ParseFloat(strWithHolding, 64)

			arrDataReturn = append(arrDataReturn, AllBlockchainWalletWithOtherBalReturnStruct{
				WalletTypeCode:            v.WalletTypeCode,
				Rate:                      rate,
				Balance:                   balance,
				ConvertedBalance:          convertedBalance,
				AvailableBalance:          availBal,
				ConvertedAvailableBalance: convertedAvailBal,
				HoldingBalance:            holdingBal,
				WithHoldingBalance:        withHoldingBal,
			})
		}
	}

	return arrDataReturn, nil
}

type SaveMemberBlockchainTransRecordsFromApiStruct struct {
	EntMemberID       int
	EwalletTypeID     int
	EwalletTypeCode   string
	DocNo             string
	Status            string
	TransactionType   string
	TotalIn           float64
	TotalOut          float64
	ConversionRate    float64
	ConvertedTotalIn  float64
	ConvertedTotalOut float64
	TransactionData   string
	HashValue         string
	Remark            string
	LogOnly           int
	CreatedAt         string
	Deduct            string
}

func SaveMemberBlockchainTransRecordsFromApi(tx *gorm.DB, arrData SaveMemberBlockchainTransRecordsFromApiStruct) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " hash_value = ?", CondValue: arrData.HashValue},
	)
	existingBlockchainTransRst, _ := models.GetBlockchainTransArrayFn(arrCond, false)

	if len(existingBlockchainTransRst) == 0 {
		arrCrtData := models.AddBlockchainTransV2Struct{
			MemberID:          arrData.EntMemberID,
			EwalletTypeID:     arrData.EwalletTypeID,
			DocNo:             arrData.DocNo,
			Status:            arrData.Status,
			TransactionType:   arrData.TransactionType,
			TotalIn:           arrData.TotalIn,
			TotalOut:          arrData.TotalOut,
			ConversionRate:    arrData.ConversionRate,
			ConvertedTotalIn:  arrData.ConvertedTotalIn,
			ConvertedTotalOut: arrData.ConvertedTotalOut,
			TransactionData:   arrData.TransactionData,
			HashValue:         arrData.HashValue,
			Remark:            arrData.Remark,
			LogOnly:           arrData.LogOnly,
			DtTimestamp:       arrData.CreatedAt,
		}
		_, err := models.AddBlockchainTransV2(tx, arrCrtData)

		if err != nil {
			base.LogErrorLog("SaveMemberBlockchainTransRecordsFromApi-failed_to_save_blockchain_trans", err.Error(), arrCrtData, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		// start deduct for holding wallet
		if arrData.Deduct == "1" {
			arrEwtSetupFn := make([]models.WhereCondFn, 0)
			arrEwtSetupFn = append(arrEwtSetupFn,
				models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: arrData.EwalletTypeCode + "H"},
				models.WhereCondFn{Condition: " control = ?", CondValue: "internal"},
				models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
			)
			ewtSetup, _ := models.GetEwtSetupFn(arrEwtSetupFn, "", false)

			if ewtSetup != nil {
				ewtIn := SaveMemberWalletStruct{
					EntMemberID:     arrData.EntMemberID,
					EwalletTypeID:   ewtSetup.ID,
					TotalIn:         arrData.TotalIn,
					TotalOut:        arrData.TotalOut,
					TransactionType: arrData.TransactionType,
					DocNo:           arrData.DocNo,
					Remark:          arrData.Remark,
					CreatedBy:       "AUTO",
				}

				// if arrData.TotalOut > 0 {
				ewtIn.AllowNegative = 1
				// }

				_, err := SaveMemberWallet(tx, ewtIn)
				if err != nil {
					base.LogErrorLog("SaveMemberBlockchainTransRecordsFromApi-save_wallet_failed", err.Error(), ewtIn, true)
					return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
				}
			}
		}
		// end deduct for holding wallet
	}

	return nil
}

type GenerateSigningKeyByModuleStruct struct {
	MemberId       int
	WalletTypeCode string
	Module         string
	Amount         float64
}

func (s *GenerateSigningKeyByModuleStruct) GenerateSigningKeyByModule() (string, error) {

	var (
		err        error
		signingKey string
	)

	// get company crypto address and private key
	arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
		models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: 0},
		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
	)
	arrCompanyCrypto, err := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)

	if err != nil {
		base.LogErrorLog("GenerateSigningKeyByModule - fail to get company address", err, arrEntMemberCryptoFn, true)
		return "", err
	}

	if arrCompanyCrypto == nil {
		base.LogErrorLog("GenerateSigningKeyByModule - empty company address", arrCompanyCrypto, arrEntMemberCryptoFn, true)
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "empty_company_address", Data: ""}
	}

	companyAddress := arrCompanyCrypto.CryptoAddress
	// companyPrivateKey := arrCompanyCrypto.PrivateKey

	memberCryptoInfo, err := models.GetCustomMemberCryptoInfov2(s.MemberId, s.WalletTypeCode, true, false)
	if err != nil {
		arrErrData := map[string]interface{}{
			"entMemberID": s.MemberId,
			"cryptoType":  s.WalletTypeCode,
		}
		base.LogErrorLog("GenerateSigningKeyByModule-GetCustomMemberCryptoInfov2 fail", err, arrErrData, true)
		return "", err
	}
	memberAddress := memberCryptoInfo.CryptoAddr
	memberPrivateKey := memberCryptoInfo.PrivateKey

	signingKeySetting, errMsg := GetSigningKeySettingByModule(s.WalletTypeCode, memberAddress, s.Module)
	arrErrSignData := map[string]interface{}{
		"walletCode": s.WalletTypeCode,
		"memberAddr": memberAddress,
		"module":     s.Module,
	}
	if errMsg != "" {
		base.LogErrorLog("GenerateSigningKeyByModule - GetSigningKeySettingByModule return fail", err, arrErrSignData, true)
		return "", err
	}

	chainID, _ := helpers.ValueToInt(signingKeySetting["chain_id"].(string))
	maxGas, _ := helpers.ValueToInt(signingKeySetting["max_gas"].(string))

	//get ewt_setup
	arrWalCond := make([]models.WhereCondFn, 0)
	arrWalCond = append(arrWalCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(s.WalletTypeCode)},
	)
	ewtSetup, err := models.GetEwtSetupFn(arrWalCond, "", false)

	if err != nil {
		base.LogErrorLog("GenerateSigningKeyByModule - GetEwtSetupFn return fail", err, arrWalCond, true)
		return "", nil
	}

	contractAddress := ewtSetup.ContractAddress
	// decimalPoint := ewtSetup.DecimalPoint

	switch s.Module {
	default:
		// generate signing key - from member wallet to company
		signingKey, err = ProcecssGenerateSignTransaction(ProcecssGenerateSignTransactionStruct{
			TokenType:       s.WalletTypeCode,
			PrivateKey:      memberPrivateKey,
			ContractAddress: contractAddress,
			ChainID:         int64(chainID),
			FromAddr:        memberAddress,
			ToAddr:          companyAddress,
			Amount:          s.Amount,
			MaxGas:          uint64(maxGas),
		})

		if err != nil {
			base.LogErrorLog("GenerateSigningKeyByModule-ProcecssGenerateSignTransaction return err", err.Error(), signingKey, true)
			return "", err
		}

	}

	return signingKey, nil
}

func GetTransferExchangeBatchInInternal(arrData MemberAccountTransferExchangeBatchSetupStruct, ewtSetup *models.EwtSetup) ([]transferSetupListStruct, error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member.main_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " ent_member.tagged_member_id > ? ", CondValue: 0},
	)
	arrMemberAccountList, _ := models.GetTaggedEntMemberListFn(arrCond, false)
	arrTransferSetupList := make([]transferSetupListStruct, 0)

	if len(arrMemberAccountList) > 0 {

		translatedWalletName := helpers.Translate(ewtSetup.EwtTypeName, arrData.LangCode)
		for _, arrMemberAccountListV := range arrMemberAccountList {
			entMemberID := arrMemberAccountListV.ID

			// start get GetWalletBalance Internal
			arrEwtBalData := GetWalletBalanceStruct{
				EntMemberID: entMemberID,
				EwtTypeCode: arrData.EwalletType,
			}

			arrBlockchainWalBal := GetWalletBalance(arrEwtBalData)
			if arrBlockchainWalBal.Balance > 0 {
				arrTransferSetupList = append(arrTransferSetupList,
					transferSetupListStruct{
						AccountName:      arrMemberAccountListV.NickName,
						EwtTypeCode:      ewtSetup.EwtTypeCode,
						EwtTypeName:      translatedWalletName,
						AvailableBalance: arrBlockchainWalBal.Balance,
						To:               arrMemberAccountListV.TaggedNickName,
					},
				)
			}
			// end get GetWalletBalance Internal
		}

		if len(arrTransferSetupList) > 0 {
			return arrTransferSetupList, nil
		} else {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "all_account_no_balance_to_transfer"}
		}
	}
	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_tagged_account"}

}

type WalletTransactionResultStructV4 struct {
	ID              string `json:"id"`
	TransDate       string `json:"trans_date"`
	TransType       string `json:"trans_type"`
	Amount          string `json:"amount"`
	Status          string `json:"status"`
	StatusColorCode string `json:"status_color_code"`
}

func (s *WalletTransactionStructV2) WalletStatementV4() (interface{}, error) {

	var (
		decimalPoint             uint
		currencyCode             string
		statusColorCode          string
		status                   string
		completedStatusColorCode = "#00A01F"
		completedStatus          = helpers.Translate("completed", s.LangCode)
		pendingStatus            = helpers.Translate("pending", s.LangCode)
		rejectStatusColorCode    = "#FD4343"
		pendingStatusColorCode   = "#DBA000"
		voidStatusColorCode      = "#FD4343"
	)

	arrWalletStatementList := make([]WalletTransactionResultStructV4, 0)

	arrEwtDetCond := make([]models.WhereCondFn, 0)
	arrEwtDetCond = append(arrEwtDetCond,
		models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
	)
	if s.TransType != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: strings.ToUpper(s.TransType)},
		)
	}

	if s.DateFrom != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(ewt_detail.created_at) >= ?", CondValue: s.DateFrom},
		)
	}

	if s.DateTo != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(ewt_detail.created_at) <= ?", CondValue: s.DateTo},
		)
	}

	if s.WalletTypeCode != "" {
		if s.WalletTypeCode == "EC" {
			s.WalletTypeCode = "CAP"
		}
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: s.WalletTypeCode},
		)
	}

	if s.RewardTypeCode != "" {
		if s.RewardTypeCode == "SPONSOR" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no LIKE ?", CondValue: "%bns_sponsor%"},
			)
		} else if s.RewardTypeCode == "PAIR" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no LIKE ?", CondValue: "%bns_pair%"},
			)
		} else if s.RewardTypeCode == "COMMUNITY" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no LIKE ?", CondValue: "%bns_community%"},
			)
		} else if s.RewardTypeCode == "GENERATION" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no LIKE ?", CondValue: "%bns_generation%"},
			)
		} else if s.RewardTypeCode == "CONTRACT" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "CONTRACT"},
			)
		}
	}

	EwtDet, _ := models.GetEwtDetailWithSetup(arrEwtDetCond, false)

	if len(EwtDet) > 0 {
		for _, v := range EwtDet {
			status = completedStatus
			statusColorCode = completedStatusColorCode
			decimalPoint = uint(v.DecimalPoint)
			currencyCode = v.CurrencyCode

			remark := v.Remark

			if remark != "" {
				remark = "-" + helpers.TransRemark(v.Remark, s.LangCode)
			}

			amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")

			if v.TotalIn > 0 {
				amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
				amount = "+" + amount + " " + currencyCode
			}

			if v.TotalOut > 0 {
				amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
				amount = "-" + amount + " " + currencyCode
			}

			transType := v.TransactionType

			if v.TransactionType == "WITHDRAW" {
				withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

				if withdrawDet != nil {
					status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
					if withdrawDet.Status == "AP" {
						status = completedStatus
					} else if withdrawDet.Status == "W" || withdrawDet.Status == "I" {
						status = pendingStatus
					}

					if withdrawDet.Status == "R" || withdrawDet.Status == "F" || withdrawDet.Status == "C" {
						statusColorCode = rejectStatusColorCode

						if v.TotalIn > 0 {
							statusColorCode = completedStatusColorCode
							status = completedStatus
						}
					} else if withdrawDet.Status == "P" || withdrawDet.Status == "W" {
						statusColorCode = pendingStatusColorCode
					} else if withdrawDet.Status == "V" {
						statusColorCode = voidStatusColorCode
					} else {
						statusColorCode = completedStatusColorCode
					}
				}
			}

			if v.TransactionType == "TRANSFER" {
				transferDet, _ := models.GetEwtTransferDetailByDocNo(v.DocNo)

				if transferDet != nil {
					if v.TotalIn > 0 {
						if transferDet.Remark != "" {
							remark = remark + " " + "(" + helpers.TransRemark(transferDet.Remark, s.LangCode) + ")"
						}
					}
					status = helpers.Translate(transferDet.StatusDesc, s.LangCode)
					if transferDet.Status == "AP" {
						status = completedStatus
					}
					if transferDet.Status == "W" {
						status = pendingStatus
					}
					if transferDet.Status == "R" || transferDet.Status == "F" {
						statusColorCode = rejectStatusColorCode
					} else if transferDet.Status == "P" || transferDet.Status == "W" {
						statusColorCode = pendingStatusColorCode
					} else if transferDet.Status == "V" {
						statusColorCode = voidStatusColorCode
					} else {
						statusColorCode = completedStatusColorCode
					}
				}
			}
			if v.TransactionType == "CONTRACT" {
				transType = "package_b_topup"
				//check whether is first purchase
				arrSlsMasterFn := make([]models.WhereCondFn, 0)
				arrSlsMasterFn = append(arrSlsMasterFn,
					models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: v.MemberID},
					models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
					models.WhereCondFn{Condition: "sls_master.status IN(?,'P') ", CondValue: "AP"},
					models.WhereCondFn{Condition: "sls_master.grp_type = ? ", CondValue: "0"},
				)
				arrSlsMaster, _ := models.GetSlsMasterAscFn(arrSlsMasterFn, "", false)

				if len(arrSlsMaster) > 0 {
					if arrSlsMaster[0].BatchNo == v.DocNo {
						transType = "package_b_subscription"
					}
				}
			}

			transTypeTrs := helpers.Translate(transType, s.LangCode)

			if v.TransactionType == "BONUS" {
				transTypeTrs = helpers.TransRemark(v.DocNo, s.LangCode)
			} else if v.TransactionType == "BOT" {
				transTypeTrs = helpers.TransRemark(v.Remark, s.LangCode)
			}

			arrWalletStatementList = append(arrWalletStatementList,
				WalletTransactionResultStructV4{
					ID:              strconv.Itoa(v.ID),
					TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
					TransType:       transTypeTrs,
					Amount:          amount,
					Status:          status,
					StatusColorCode: statusColorCode,
				})
		}
	}

	//start paginate

	sort.Slice(arrWalletStatementList, func(p, q int) bool {
		return arrWalletStatementList[q].TransDate < arrWalletStatementList[p].TransDate
	})

	arrDataReturn := app.ArrDataResponseDefaultList{}

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

	totalRecord := len(arrWalletStatementList)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(s.Page), int(limit), totalRecord)

	processArr := arrWalletStatementList[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(curPage),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      processArr,
	}

	return arrDataReturn, nil

}

func (s *WalletTransactionStructV2) WithdrawStatement() (interface{}, error) {

	var (
		decimalPoint   uint
		currencyCodeTo string
		ewtTypeName    string
	)

	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	var arrTableHeaderList []arrStatementListSettingListStruct

	arrWalletStatementList := make([]WalletTransactionResultStructV4, 0)
	arrWalletModuleStatementList := make([]interface{}, 0)
	statusColorCode := "#00A01F"

	arrEwtDetCond := make([]models.WhereCondFn, 0)
	arrEwtDetCond = append(arrEwtDetCond,
		models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
	)
	if s.TransType != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: strings.ToUpper(s.TransType)},
			models.WhereCondFn{Condition: "ewt_detail.total_out > ?", CondValue: 0},
		)

	}

	if s.DateFrom != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(ewt_detail.created_at) >= ?", CondValue: s.DateFrom},
		)
	}

	if s.DateTo != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(ewt_detail.created_at) <= ?", CondValue: s.DateTo},
		)
	}

	if s.WalletTypeCode != "" {
		arrCondWal := make([]models.WhereCondFn, 0)
		arrCondWal = append(arrCondWal,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: s.WalletTypeCode},
		)
		walrst, _ := models.GetEwtSetupFn(arrCondWal, "", false)

		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: walrst.ID},
		)
	}

	EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

	arrStatementListSetting, _ := models.GetSysGeneralSetupByID("withdraw_statement_api_setting")
	if arrStatementListSetting != nil {
		var arrStatementListSettingList map[string][]arrStatementListSettingListStruct
		json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
		arrTableHeaderList = arrStatementListSettingList["table_header_list"]
		for k, v1 := range arrStatementListSettingList["table_header_list"] {
			v1.Name = helpers.Translate(v1.Name, s.LangCode)
			arrTableHeaderList[k] = v1
		}
	}

	if len(EwtDet) > 0 {
		for _, v := range EwtDet {
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: v.EwalletTypeID},
				models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
			)
			ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
			if ewtSetup != nil {
				decimalPoint = uint(ewtSetup.DecimalPoint)
				ewtTypeName = helpers.Translate(ewtSetup.EwtTypeName, s.LangCode)
			}
			status := helpers.Translate("completed", s.LangCode)
			// transType := helpers.Translate(v.TransactionType, s.LangCode)

			withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

			if withdrawDet != nil {
				// withdrawType := helpers.Translate(withdrawDet.Type, s.LangCode)
				status = helpers.Translate(withdrawDet.StatusDesc, s.LangCode)
				// remark := helpers.TransRemark(withdrawDet.Remark, s.LangCode)
				remark := withdrawDet.Remark

				currencyCodeTo = withdrawDet.EwalletTo

				amount := helpers.CutOffDecimal(withdrawDet.NetAmount, decimalPoint, ".", ",")
				amount = amount + " " + helpers.Translate(currencyCodeTo, s.LangCode)

				cancelStatus := 0

				if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
					statusColorCode = "#FD4343"
				} else if withdrawDet.Status == "PR" || withdrawDet.Status == "W" {
					statusColorCode = "#DBA000"
				} else if withdrawDet.Status == "V" || withdrawDet.Status == "C" { //void /cancelled
					statusColorCode = "#FD4343"
				} else {
					statusColorCode = "#00A01F"
				}

				transType := v.TransactionType

				if withdrawDet.Status == "W" {
					cancelStatus = 1
				}

				arrWalletModuleStatementList = append(arrWalletModuleStatementList,
					WithdrawStatementListStruct{
						ID:              strconv.Itoa(withdrawDet.ID),
						DocNo:           withdrawDet.DocNo,
						TransDate:       withdrawDet.TransDate.Format("2006-01-02 15:04:05"),
						TransType:       helpers.Translate(transType, s.LangCode),
						EwalletTypeName: ewtTypeName,
						// WithdrawType:    withdrawType,
						// CreditOut:       fmt.Sprintf("%.2f", withdrawDet.TotalOut),
						// GasFee:          fmt.Sprintf("%.6f", withdrawDet.GasFee),
						// NetAmount:       fmt.Sprintf("%.6f", withdrawDet.NetAmount),
						Amount:          amount,
						Address:         withdrawDet.CryptoAddrTo,
						Hash:            withdrawDet.TranHash,
						Status:          status,
						StatusColorCode: statusColorCode,
						Remark:          remark,
						CancelStatus:    cancelStatus,
					})
			}
		}
	}

	//start paginate
	if len(arrWalletStatementList) > 0 {
		sort.Slice(arrWalletStatementList, func(p, q int) bool {
			return arrWalletStatementList[q].TransDate < arrWalletStatementList[p].TransDate
		})

		arrDataReturn := app.ArrDataResponseDefaultList{}

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

		totalRecord := len(arrWalletStatementList)

		totalPage := float64(totalRecord) / float64(limit)
		totalPage = math.Ceil(totalPage)

		pageStart, pageEnd := helpers.Paginate(int(s.Page), int(limit), totalRecord)

		processArr := arrWalletStatementList[pageStart:pageEnd]

		totalCurrentPageItems := len(processArr)

		perPage := int(limit)

		arrDataReturn = app.ArrDataResponseDefaultList{
			CurrentPage:           int(curPage),
			PerPage:               int(perPage),
			TotalCurrentPageItems: int(totalCurrentPageItems),
			TotalPage:             int(totalPage),
			TotalPageItems:        int(totalRecord),
			CurrentPageItems:      processArr,
		}

		return arrDataReturn, nil
	} else {

		sort.Slice(arrWalletModuleStatementList, func(i, j int) bool {
			commonID1 := reflect.ValueOf(arrWalletModuleStatementList[i]).FieldByName("TransDate").String()
			commonID2 := reflect.ValueOf(arrWalletModuleStatementList[j]).FieldByName("TransDate").String()
			return commonID1 > commonID2
		})

		page := base.Pagination{
			Page:    s.Page,
			DataArr: arrWalletModuleStatementList,
		}

		arrDataReturn := page.PaginationInterfaceV1()

		return arrDataReturn, nil
	}

}

type GetWithdrawSettingStruct struct {
	MemberID    int
	EwtTypeCode string
	LangCode    string
}

type WithdrawToSetup struct {
	EwalletTypeCode string  `json:"ewallet_type_code_to"`
	EwalletTypeName string  `json:"ewallet_type_name_to"`
	CurrencyCode    string  `json:"currency_code"`
	CryptoTypeCode  string  `json:"crypto_type_code"`
	Min             float64 `json:"min"`
	Max             float64 `json:"max"`
	// Rate         float64 `json:"rate"`
	GasFee float64 `json:"gas_fee"`
}

type WithdrawChargesSetup struct {
	ChargesType  string  `json:"charges_type"`
	DurationType string  `json:"duration_type"`
	Duration     string  `json:"duration"`
	AdminFee     float64 `json:"admin_fee"`
	AdminFeeDesc string  `json:"admin_fee_desc"`
	Qualify      int     `json:"qualify"`
}

type WithdrawSetup struct {
	EwtTypeCode               string                 `json:"ewallet_type_code"`
	EwtTypeName               string                 `json:"ewallet_type_name"`
	CurrencyCode              string                 `json:"currency_code"`
	MultipleOf                int                    `json:"multiple_of"`
	FreeWithdrawCountdownDays string                 `json:"free_withdraw_countdown_days"`
	WithdrawTo                []WithdrawToSetup      `json:"withdraw_to"`
	WithdrawCharges           []WithdrawChargesSetup `json:"withdraw_charges"`
}

func (w *GetWithdrawSettingStruct) GetMemberWithdrawSettingv1() (interface{}, error) {

	var (
		withdrawStatus int
		durationType   string
		withdrawSetup  WithdrawSetup
		expDateStr     string
		daysStr        string = "0"
		monthsStr      string = "0"
		multipleOf     int    = 0
	)

	//get wallet
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: w.EwtTypeCode},
	)

	ewtSetup, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberWithdrawSettingv1 - GetEwtSetupListFn", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	withdrawStatus = ewtSetup.Withdraw

	//check block all withdrawal
	checkBlkAll := member_service.VerifyIfInNetwork(w.MemberID, "WD_BLK_ALL")

	if checkBlkAll {
		withdrawStatus = 0
	}

	if withdrawStatus == 1 {
		//get withdraw setup
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_withdraw_setup.ewallet_type_id = ?", CondValue: ewtSetup.ID},
			models.WhereCondFn{Condition: "ewt_withdraw_setup.withdraw_type = ?", CondValue: "CRYPTO"},
		)

		arrWithdrawSetup, err := models.GetEwtWithdrawSetupFn(arrCond, "", false)

		if err != nil {
			base.LogErrorLog("GetMemberWithdrawSettingv1 - GetEwtWithdrawSetupFn", err.Error(), arrCond, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		//check member whether is first time withdraw
		// arrCond = make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: " ewt_withdraw.member_id = ? ", CondValue: w.MemberID},
		// 	models.WhereCondFn{Condition: " ewt_withdraw.status = ? ", CondValue: "AP"},
		// )
		// arrEwtWithdraw, err := models.GetEwtWithdrawFn(arrCond, false)

		// if err != nil {
		// 	base.LogErrorLog("GetMemberWithdrawSettingv1 - GetEwtWithdrawFn", err.Error(), arrCond, true)
		// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		// }

		//check countdown expire
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_withdraw.member_id = ?", CondValue: w.MemberID},
			models.WhereCondFn{Condition: "ewt_withdraw.charges_type = ?", CondValue: "FREE"},
			models.WhereCondFn{Condition: "ewt_withdraw.status = ?", CondValue: "AP"},
		)

		arrCheckCountdown, err := models.GetEwtWithdrawFn(arrCond, false) //get latest free record
		if err != nil {
			base.LogErrorLog("GetMemberWithdrawSettingv1 - check countdown status err", err.Error(), arrCond, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if len(arrCheckCountdown) > 0 {
			expDate := arrCheckCountdown[0].ExpiredAt
			currDate := base.GetCurrentDateTimeT()
			duration := expDate.Sub(currDate).Hours() / 24

			days := int(math.Ceil(duration))
			daysStr = strconv.Itoa(days)
		}

		if len(arrWithdrawSetup) > 0 {
			roundMonth := arrWithdrawSetup[0].CountdownDays / 30
			monthsStr = strconv.Itoa(roundMonth)
			multipleOf = arrWithdrawSetup[0].MultipleOf
		}
		expDateStr = helpers.TranslateV2("every_:0_months_you_are_entitled_to_1_free_withdrawal_next_free_withdrawal_available_in_:1_days", w.LangCode, map[string]string{"0": monthsStr, "1": daysStr})

		withdrawChargesArr := make([]WithdrawChargesSetup, 0)
		withdrawToArr := make([]WithdrawToSetup, 0)

		for _, withdrawSetupVal := range arrWithdrawSetup {
			gasFee := float64(0)
			if withdrawSetupVal.Main == 1 {
				if withdrawSetupVal.EwalletTypeNameTo == "USDT" {
					withdrawSetupVal.EwalletTypeNameTo = "USDT_TRC20"
				}
				if withdrawSetupVal.EwalletTypeNameTo == "USDC" {
					withdrawSetupVal.EwalletTypeNameTo = "USDC_TRC20"
				}

				if withdrawSetupVal.EwalletTypeCodeTo == "USDT_TRX" {
					gasFee, _ = models.GetLatestGasFeeMovementTron()
				} else if withdrawSetupVal.EwalletTypeCodeTo == "USDC_ERC20" || withdrawSetupVal.EwalletTypeCodeTo == "USDT_ERC20" {
					gasFee, _ = models.GetLatestGasFeeMovementErc20()
				}

				withdrawToArr = append(withdrawToArr, WithdrawToSetup{
					EwalletTypeCode: withdrawSetupVal.EwalletTypeCodeTo,
					EwalletTypeName: helpers.Translate(withdrawSetupVal.EwalletTypeNameTo, w.LangCode),
					CurrencyCode:    withdrawSetupVal.CurrencyCodeTo,
					CryptoTypeCode:  withdrawSetupVal.EwalletToBlkCCode,
					Min:             withdrawSetupVal.Min,
					Max:             withdrawSetupVal.Max,
					GasFee:          gasFee,
				})
			} else {
				if withdrawSetupVal.ChargesType != "" {
					qualifyStatus := 0
					if withdrawSetupVal.ChargesType == "FREE" {
						durationType = helpers.Translate("normal_duration", w.LangCode)
						//if is first time withdraw then enable
						// if len(arrEwtWithdraw) < 1 {
						// 	qualifyStatus = 1
						// }

						//if countdown end enable
						if len(arrCheckCountdown) > 0 {
							if helpers.CompareDateTime(time.Now(), ">", arrCheckCountdown[0].ExpiredAt) {
								qualifyStatus = 1
							}
						} else {
							qualifyStatus = 1
						}

					} else if withdrawSetupVal.ChargesType == "NORMAL" {
						durationType = helpers.Translate("normal_duration", w.LangCode)

						qualifyStatus = 1

						// if len(arrEwtWithdraw) > 0 {
						// 	qualifyStatus = 1
						// }

						//if countdown end disable
						// if len(arrCheckCountdown) > 0 {
						// 	if helpers.CompareDateTime(time.Now(), ">", arrCheckCountdown[0].ExpiredAt) {
						// 		qualifyStatus = 0
						// 	}
						// }

					} else if withdrawSetupVal.ChargesType == "EXPRESS" {
						durationType = helpers.Translate("express_duration", w.LangCode)

						qualifyStatus = 1

						// if len(arrEwtWithdraw) > 0 {
						// 	qualifyStatus = 1
						// }
						//if countdown end disable
						// if len(arrCheckCountdown) > 0 {
						// 	if helpers.CompareDateTime(time.Now(), ">", arrCheckCountdown[0].ExpiredAt) {
						// 		qualifyStatus = 0
						// 	}
						// }
					}

					withdrawChargesArr = append(withdrawChargesArr, WithdrawChargesSetup{
						ChargesType:  withdrawSetupVal.ChargesType,
						Duration:     helpers.Translate(withdrawSetupVal.Remark, w.LangCode),
						DurationType: durationType,
						AdminFee:     withdrawSetupVal.AdminFee,
						AdminFeeDesc: fmt.Sprintf("%.0f", withdrawSetupVal.AdminFee) + "%",
						Qualify:      qualifyStatus,
					})
				}
			}
		}

		withdrawSetup = WithdrawSetup{
			EwtTypeCode:               ewtSetup.EwtTypeCode,
			EwtTypeName:               helpers.Translate(ewtSetup.EwtTypeName, w.LangCode),
			CurrencyCode:              ewtSetup.CurrencyCode,
			MultipleOf:                multipleOf,
			FreeWithdrawCountdownDays: expDateStr,
			WithdrawTo:                withdrawToArr,
			WithdrawCharges:           withdrawChargesArr,
		}
	}
	return withdrawSetup, nil
}

type GetTransferSettingStruct struct {
	MemberID    int
	EwtTypeCode string
	LangCode    string
}

func (w *GetTransferSettingStruct) GetMemberTransferSettingv1() (interface{}, error) {

	var (
		// arrTransferSetup TransferSetup
		// arrTransferSetup []TransferSetup
		arrTransferSetup   map[string]interface{}
		toArr              []interface{}
		sameMemStatus      int = 0
		walletToDispStatus int = 0
	)
	//get wallet
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: w.EwtTypeCode},
	)

	ewtSetup, err := models.GetMemberEwtSetupBalanceFn(w.MemberID, arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberTransferSettingv1 - GetMemberEwtSetupBalanceFn", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	arrTransferSetupCond := make([]models.WhereCondFn, 0)
	arrTransferSetupCond = append(arrTransferSetupCond,
		models.WhereCondFn{Condition: "ewt_transfer_setup.ewallet_type_id_from = ?", CondValue: ewtSetup[0].ID},
		models.WhereCondFn{Condition: "ewt_transfer_setup.ewt_transfer_type = ?", CondValue: "Internal"},
		models.WhereCondFn{Condition: "ewt_transfer_setup.member_show = ?", CondValue: 1},
	)

	arrTransferSetupRst, _ := models.GetEwtTransferSetupFn(arrTransferSetupCond, "", false)
	if len(arrTransferSetupRst) > 0 {
		// translatedWalletFromName := helpers.Translate(arrTransferSetupRst[0].EwalletTypeNameFrom, w.LangCode)
		// translatedWalletToName := helpers.Translate(arrTransferSetupRst[0].EwalletTypeNameTo, w.LangCode)
		// arrTransferSetup = TransferSetup{
		// 	Min: arrTransferSetupRst[0].TransferMin,
		// 	Max: arrTransferSetupRst[0].TransferMax,
		// 	// AdminFee: arrTransferSetupRst[0].AdminFee,
		// 	EwtTypeCodeFrom: arrTransferSetupRst[0].EwalletTypeCodeFrom,
		// 	EwtTypeNameFrom: translatedWalletFromName,
		// 	EwtTypeCodeTo:   arrTransferSetupRst[0].EwalletTypeCodeTo,
		// 	EwtTypeNameTo:   translatedWalletToName,
		// }

		// arrTransferSetup := make([]TransferSetup, 0)
		// transferStatus = 1
		// for _, v2 := range arrTransferSetupRst {
		// 	translatedWalletFromName := helpers.Translate(v2.EwalletTypeNameFrom, w.LangCode)
		// 	translatedWalletToName := helpers.Translate(v2.EwalletTypeNameTo, w.LangCode)
		// 	arrTransferSetup = append(arrTransferSetup, TransferSetup{
		// 		EwtTypeCodeFrom: v2.EwalletTypeCodeFrom,
		// 		EwtTypeNameFrom: translatedWalletFromName,
		// 		EwtTypeCodeTo:   v2.EwalletTypeCodeTo,
		// 		EwtTypeNameTo:   translatedWalletToName,
		// 		Min:             v2.TransferMin,
		// 		Max:             v2.TransferMax,
		// 		// AdminFee:        v2.AdminFee,
		// 	})
		// }

		for _, v2 := range arrTransferSetupRst {
			translatedWalletFromName := helpers.Translate(v2.EwalletTypeNameFrom, w.LangCode)
			translatedWalletToName := helpers.Translate(v2.EwalletTypeNameTo, w.LangCode)
			walletToDispStatus = v2.ShowWalletTo
			sameMemStatus = v2.TransferSameMember
			toArr = append(toArr, map[string]interface{}{
				"ewallet_type_code_from": v2.EwalletTypeCodeFrom,
				"ewallet_type_name_from": translatedWalletFromName,
				"ewallet_type_code_to":   v2.EwalletTypeCodeTo,
				"ewallet_type_name_to":   translatedWalletToName,
			})
		}
	}
	arrTransferSetup = map[string]interface{}{
		"min":            arrTransferSetupRst[0].TransferMin,
		"max":            arrTransferSetupRst[0].TransferMax,
		"display_to":     walletToDispStatus, //whether to wallet to
		"same_member_to": sameMemStatus,      //whether need to let user key in username
		"to":             toArr,
	}

	return arrTransferSetup, nil
}

type GetTransferExchangeSettingStruct struct {
	MemberID    int
	EwtTypeCode string
	LangCode    string
}

func (w *GetTransferExchangeSettingStruct) GetMemberTransferExchangeSettingv1() (interface{}, error) {

	var (
		withdrawalWithCrypto WithdrawalCryptoSetup
	)

	//get wallet
	// arrCond := make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
	// 	models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
	// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: w.EwtTypeCode},
	// )

	// ewtSetup, err := models.GetMemberEwtSetupBalanceFn(w.MemberID, arrCond, "", false)

	// if err != nil {
	// 	base.LogErrorLog("GetMemberTransferExchangeSettingv1 - GetMemberEwtSetupBalanceFn", err.Error(), arrCond, true)
	// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	// }

	// if ewtSetup[0].WithdrawalWithCrypto == 1 {
	// 	withdrawalWithCrypto = WithdrawalCryptoSetup{
	// 		Min: ewtSetup[0].WithdrawMin,
	// 		Max: ewtSetup[0].WithdrawMax,
	// 	}
	// }

	return withdrawalWithCrypto, nil
}

type CancelWithdrawStruct struct {
	MemberId int
	DocNo    string
	LangCode string
}

func (w *CancelWithdrawStruct) CancelWithdraw(tx *gorm.DB) (interface{}, error) {
	var (
		err error
	)

	//get withdraw record with doc no
	withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(w.DocNo)

	if withdrawDet == nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_record", w.LangCode), Data: err}
	}

	//check whether record is same with member id with login
	if withdrawDet.MemberId != w.MemberId {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("unable_to_modify_this_record", w.LangCode), Data: err}
	}

	if withdrawDet.Status != "W" {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("cannot_cancel_this_withdraw_record", w.LangCode), Data: err}
	}

	//refund
	ewtIn := SaveMemberWalletStruct{
		EntMemberID:       w.MemberId,
		EwalletTypeID:     withdrawDet.EwalletTypeId,
		TotalIn:           withdrawDet.TotalOut,
		ConversionRate:    withdrawDet.ConversionRate,
		ConvertedTotalOut: withdrawDet.TotalOut,
		TransactionType:   "WITHDRAW-CANCEL",
		Remark:            withdrawDet.DocNo,
	}

	_, err = SaveMemberWallet(tx, ewtIn)

	if err != nil {
		base.LogErrorLog("CancelWithdraw - fail to save ewt_detail", err, ewtIn, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", w.LangCode), Data: err}
	}

	//update withdraw record
	arrUpdate := make([]models.WhereCondFn, 0)
	arrUpdate = append(arrUpdate,
		models.WhereCondFn{Condition: " ewt_withdraw.id = ? ", CondValue: withdrawDet.ID},
	)
	updateColumn := map[string]interface{}{
		"status":       "C",
		"cancelled_at": time.Now(),
		"cancelled_by": w.MemberId,
	}
	models.UpdatesFn("ewt_withdraw", arrUpdate, updateColumn, false)

	arrData := make(map[string]interface{})
	arrData["ewallet_type"] = helpers.Translate(withdrawDet.EwalletFrom, w.LangCode)
	arrData["amount"] = withdrawDet.TotalOut
	arrData["cancel_at"] = time.Now().Format("2006-01-02 15:04:05")

	return arrData, nil

}

type GetExchangeSettingStruct struct {
	MemberID    int
	EwtTypeCode string
	LangCode    string
}

type ExchangeToSetup struct {
	EwalletTypeCode string  `json:"ewallet_type_code_to"`
	EwalletTypeName string  `json:"ewallet_type_name_to"`
	CurrencyCode    string  `json:"currency_code"`
	Min             float64 `json:"min"`
	Max             float64 `json:"max"`
	Rate            float64 `json:"rate"`
}

type ExchangeSetup struct {
	EwtTypeCode  string            `json:"ewallet_type_code"`
	EwtTypeName  string            `json:"ewallet_type_name"`
	CurrencyCode string            `json:"currency_code"`
	ExchangeTo   []ExchangeToSetup `json:"exchange_to"`
}

func (ex *GetExchangeSettingStruct) GetMemberExchangeSettingv1() (interface{}, error) {

	var (
		exchangeStatus int
		exchangeSetup  ExchangeSetup
		rate           float64 = 1
	)

	//get wallet
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: ex.EwtTypeCode},
	)

	ewtSetup, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberExchangeSettingv1 - GetEwtSetupListFn", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	exchangeStatus = ewtSetup.Exchange

	//check block all exchange
	checkBlkAll := member_service.VerifyIfInNetwork(ex.MemberID, "EX_BLK_ALL")

	if checkBlkAll {
		exchangeStatus = 0
	}

	if exchangeStatus == 1 {
		//get exchange setup
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_exchange_setup.ewallet_type_id = ?", CondValue: ewtSetup.ID},
		)

		arrExchangeSetup, err := models.GetEwtExchangeSetupFn(arrCond, "", false)

		if err != nil {
			base.LogErrorLog("GetMemberExchangeSettingv1 - GetEwtExchangeSetupFn", err.Error(), arrCond, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		exchangeToArr := make([]ExchangeToSetup, 0)

		for _, exSetupVal := range arrExchangeSetup {
			if exSetupVal.Main == 1 {
				//get price movement if need. currently 1 to 1
				exchangeToArr = append(exchangeToArr, ExchangeToSetup{
					EwalletTypeCode: exSetupVal.EwalletTypeCodeTo,
					EwalletTypeName: helpers.Translate(exSetupVal.EwalletTypeNameTo, ex.LangCode) + " " + "(" + exSetupVal.EwalletTypeCodeTo + ")",
					CurrencyCode:    exSetupVal.CurrencyCodeTo,
					Min:             exSetupVal.Min,
					Max:             exSetupVal.Max,
					Rate:            rate,
				})
			} else {
				//for sub wallet setting (main = 0)
			}
		}

		exchangeSetup = ExchangeSetup{
			EwtTypeCode:  ewtSetup.EwtTypeCode,
			EwtTypeName:  helpers.Translate(ewtSetup.EwtTypeName, ex.LangCode),
			CurrencyCode: ewtSetup.CurrencyCode,
			ExchangeTo:   exchangeToArr,
		}
	}
	return exchangeSetup, nil
}

type PostExchangeStruct struct {
	MemberId          int
	LangCode          string
	Amount            float64
	EwalletTypeCode   string
	EwalletTypeCodeTo string
}

func (ex *PostExchangeStruct) PostExchange(tx *gorm.DB) (interface{}, error) {
	var (
		err                  error
		decimalPoint         uint    = 2
		adminFee             float64 = 0
		rate                 float64 = 1
		currency             string
		toWalletDecimalPoint uint    = 2
		netAmount            float64 = 0
		convertedNetAmount   float64 = 0
		// convertedTotalAmount float64 = 0
	)

	//check member wallet balance - save member wallet function will handle

	//get wallet setup
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(ex.EwalletTypeCode)},
	)
	WalSetup, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("PostExchange - fail to get wallet setup", err, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: err}
	}

	if WalSetup == nil {
		base.LogErrorLog("PostExchange - empty wallet setup", ex, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: ex}
	}

	decimalPoint = uint(WalSetup.DecimalPoint)
	eWalletId := WalSetup.ID
	currency = helpers.Translate(WalSetup.CurrencyCode, ex.LangCode)

	//get wallet to setup
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: strings.ToUpper(ex.EwalletTypeCodeTo)},
	)
	WalSetupTo, err := models.GetEwtSetupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("PostExchange - fail to get wallet setup", err, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: err}
	}

	if WalSetupTo == nil {
		base.LogErrorLog("PostExchange - empty wallet setup to", ex, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: ex}
	}

	toWalletDecimalPoint = uint(WalSetupTo.DecimalPoint)
	eWalletIdTo := WalSetupTo.ID

	//get exchange setup
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_exchange_setup.ewallet_type_id = ?", CondValue: WalSetup.ID},
		models.WhereCondFn{Condition: "ewt_exchange_setup.ewallet_type_id_to = ?", CondValue: WalSetupTo.ID},
	)

	arrExchangeSetup, err := models.GetEwtExchangeSetupFnV2(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("PostExchange - GetEwtExchangeSetupFnV2-General Setup", err.Error(), arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if arrExchangeSetup == nil {
		base.LogErrorLog("PostExchange - empty exchange setup", arrExchangeSetup, arrCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: ex}
	}

	adminFee = float64(arrExchangeSetup.AdminFee)

	//begin checking

	//check member wallet lock
	walletLock, err := models.GetEwtLockByMemberId(ex.MemberId, WalSetup.ID)

	if err != nil {
		base.LogErrorLog("PostExchange - fail to get ewtLock Setup", err, ex, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: ex}
	}

	if walletLock.Exchange == 1 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("wallet_is_being_locked_from_exchange", ex.LangCode), Data: ex}
	}

	//check amt
	if ex.Amount <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("amount_must_more_than_0", ex.LangCode), Data: ""}
	}

	// check min
	if ex.Amount < arrExchangeSetup.Min {
		strAmt := helpers.CutOffDecimal(arrExchangeSetup.Min, 2, ".", "")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("minimum_exchange_amount_is"+" "+strAmt, ex.LangCode), Data: ex}
	}

	//check multiple of
	if arrExchangeSetup.MultipleOf > 0 {
		multipleOf := float64(arrExchangeSetup.MultipleOf)
		if !helpers.IsMultipleOf(ex.Amount, multipleOf) {
			strAmt := helpers.CutOffDecimal(multipleOf, 2, ".", "")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("amount_must_be_multiple_of"+" "+strAmt, ex.LangCode), Data: ex}
		}
	}

	//end checking

	//get rate
	if WalSetup.CurrencyCode != WalSetupTo.CurrencyCode { //if to wallet currency not same get need rate
		rate, err = base.GetLatestPriceMovementByTokenType(ex.EwalletTypeCode)
		if err != nil {
			base.LogErrorLog("PostExchange - GetLatestPriceMovementByTokenType Error", err, ex, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: ex}
		}
	}

	netAmount = ex.Amount

	adminFeeAmount, _ := decimal.NewFromFloat(ex.Amount).Mul(decimal.NewFromFloat(adminFee)).Float64()
	adminFeeAmount, _ = decimal.NewFromFloat(adminFeeAmount).Div(decimal.NewFromFloat(100)).Float64()
	// convertedAdminFee, _ := decimal.NewFromFloat(adminFeeAmount).Mul(decimal.NewFromFloat(rate)).Float64()

	//deduct admin fee
	netAmount, _ = decimal.NewFromFloat(netAmount).Sub(decimal.NewFromFloat(adminFeeAmount)).Float64()
	convertedNetAmount, _ = decimal.NewFromFloat(netAmount).Mul(decimal.NewFromFloat(rate)).Float64()
	// convertedTotalAmount, _ = decimal.NewFromFloat(netAmount).Mul(decimal.NewFromFloat(rate)).Float64()

	if netAmount <= 0 || convertedNetAmount <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("not_enough_amount_to_deduct_admin_fee", ex.LangCode), Data: ex}
	}

	docs, err := models.GetRunningDocNo("EX", tx)

	if err != nil {
		base.LogErrorLog("PostExchange - fail to get WD doc no", err, ex, true) //store error log
		return nil, err
	}

	//deduct balance
	ConvertedTotalOut, _ := decimal.NewFromFloat(ex.Amount).Mul(decimal.NewFromFloat(rate)).Float64()
	ewtOut := SaveMemberWalletStruct{
		EntMemberID:       ex.MemberId,
		EwalletTypeID:     eWalletId,
		TotalOut:          ex.Amount,
		ConversionRate:    rate,
		ConvertedTotalOut: ConvertedTotalOut,
		TransactionType:   "EXCHANGE",
		DocNo:             docs,
		Remark:            docs,
	}

	_, err = SaveMemberWallet(tx, ewtOut)

	if err != nil {
		base.LogErrorLog("PostExchange - fail to save wallet out", err, ewtOut, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: err}
	}

	//add balance
	ewtIn := SaveMemberWalletStruct{
		EntMemberID:     ex.MemberId,
		EwalletTypeID:   eWalletIdTo,
		TotalIn:         convertedNetAmount,
		ConversionRate:  rate,
		TransactionType: "EXCHANGE",
		DocNo:           docs,
		Remark:          docs,
	}

	_, err = SaveMemberWallet(tx, ewtIn)

	if err != nil {
		base.LogErrorLog("PostExchange - fail to save wallet in", err, ewtIn, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: err}
	}

	//save ewt_exchange
	arrEwtExchange := models.EwtExchangeStruct{
		MemberID:             ex.MemberId,
		DocNo:                docs,
		EwalletTypeID:        eWalletId,
		EwalletTypeIDTo:      eWalletIdTo,
		Amount:               ex.Amount,
		AdminFee:             adminFeeAmount,
		NettAmount:           netAmount,
		Rate:                 rate,
		ConvertedTotalAmount: convertedNetAmount,
		Status:               "AP",
		CreatedAt:            time.Now(),
	}

	_, err = models.AddEwtExchange(tx, arrEwtExchange) //store exchange

	if err != nil {
		base.LogErrorLog("PostExchange - fail to save ewt_exchange", err, arrEwtExchange, true) //store error log
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: err}
	}

	err = models.UpdateRunningDocNo("EX", tx) //update exchange doc no

	if err != nil {
		base.LogErrorLog("PostExchange - fail to update EX doc no", err, ex, true) //store error log
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", ex.LangCode), Data: err}
	}

	convertedNetAmountStr := helpers.CutOffDecimal(convertedNetAmount, toWalletDecimalPoint, ".", ",")
	adminFeeAmountStr := helpers.CutOffDecimal(adminFeeAmount, decimalPoint, ".", ",")

	arrData := make(map[string]interface{})
	arrData["ewallet_type"] = helpers.Translate(ex.EwalletTypeCode, ex.LangCode)
	arrData["amount"] = ex.Amount
	arrData["payment"] = convertedNetAmountStr + " " + currency
	if adminFee > 0 {
		arrData["admin_fee"] = adminFeeAmountStr
	}
	arrData["trans_time"] = time.Now().Format("2006-01-02 15:04:05")
	arrData["ewallet_type_to"] = helpers.Translate(WalSetupTo.EwtTypeName, ex.LangCode)

	return arrData, nil

}

type WalletTransactionStrategyStruct struct {
	ID              string `json:"id"`
	TransDate       string `json:"trans_date"`
	Amount          string `json:"amount"`
	Status          string `json:"status"`
	StatusColorCode string `json:"status_color_code"`
	Remark          string `json:"remark"`
}

func (s *WalletTransactionStructV2) WalletStatementStrategyV1() (interface{}, error) {

	var (
		decimalPoint           uint
		statusColorCode        = "#00A01F"
		status                 = helpers.Translate("completed", s.LangCode)
		reloadForTranslated    = helpers.Translate("reload_for", s.LangCode)
		withdrawalToTranslated = helpers.Translate("withdrawal_to", s.LangCode)
		walletTranslated       = helpers.Translate("wallet", s.LangCode)
		walletOut              string
	)

	arrWalletStatementList := make([]WalletTransactionStrategyStruct, 0)

	arrEwtDetCond := make([]models.WhereCondFn, 0)
	arrEwtDetails := make([]models.WhereCondFn, 0)

	arrEwtDetCond = append(arrEwtDetCond,
		models.WhereCondFn{Condition: "a.member_id = ?", CondValue: s.MemberID},
	)
	if s.TransType != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "a.transaction_type = ?", CondValue: strings.ToUpper(s.TransType)},
		)
	}

	if s.DateFrom != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(a.created_at) >= ?", CondValue: s.DateFrom},
		)
	}

	if s.DateTo != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(a.created_at) <= ?", CondValue: s.DateTo},
		)
	}

	if s.WalletTypeCode != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "b.ewallet_type_code = ?", CondValue: s.WalletTypeCode},
		)
	}

	EwtDet, _ := models.GetEwtDetailStrategyWithSetup(arrEwtDetCond, false)

	if len(EwtDet) > 0 {

		for _, v := range EwtDet {
			decimalPoint = uint(v.DecimalPoint)

			if v.Remark != "" {
				v.Remark = helpers.TransRemark(v.Remark, s.LangCode)
			}

			amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")

			if v.TotalIn > 0 {
				amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
				amount = amount
			}

			if v.TotalOut > 0 {
				amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
				amount = amount
			}

			if v.TransactionType == "TRADING_DEPOSIT" {
				arrEwtDetails := append(arrEwtDetails,
					models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: v.DocNo},
					models.WhereCondFn{Condition: "ewt_detail.total_out > ?", CondValue: 0},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
				)

				EwtDetails, _ := models.GetEwtDetailWithSetup(arrEwtDetails, false)

				if len(EwtDetails) > 0 {
					walletOut = EwtDetails[0].EwalletTypeName
				}

				if v.Remark == "" {
					v.Remark = reloadForTranslated + " " + walletOut + " " + walletTranslated
				}
			} else if v.TransactionType == "TRADING_DEPOSIT_WITHDRAW" {
				arrEwtDetails := append(arrEwtDetails,
					models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: v.DocNo},
					models.WhereCondFn{Condition: "ewt_detail.total_in > ?", CondValue: 0},
					models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: s.MemberID},
				)

				EwtDetails, _ := models.GetEwtDetailWithSetup(arrEwtDetails, false)

				if len(EwtDetails) > 0 {
					walletOut = EwtDetails[0].EwalletTypeName
				}
				if v.Remark == "" {
					v.Remark = withdrawalToTranslated + " " + walletOut + " " + walletTranslated
				}
			}

			arrWalletStatementList = append(arrWalletStatementList,
				WalletTransactionStrategyStruct{
					ID:              strconv.Itoa(v.ID),
					TransDate:       v.TransDate.Format("2006-01-02 15:04:05"),
					Amount:          amount,
					Status:          status,
					StatusColorCode: statusColorCode,
					Remark:          v.Remark,
				})
		}
	}

	//start paginate

	sort.Slice(arrWalletStatementList, func(p, q int) bool {
		return arrWalletStatementList[q].TransDate < arrWalletStatementList[p].TransDate
	})

	arrDataReturn := app.ArrDataResponseDefaultList{}

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

	totalRecord := len(arrWalletStatementList)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(s.Page), int(limit), totalRecord)

	processArr := arrWalletStatementList[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(curPage),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      processArr,
	}

	return arrDataReturn, nil

}

func GetMemberWithdrawBalance(memID, walletTypeID int) (float64, error) {

	var (
		err               error
		totalBalance      float64
		bonusBalance      float64
		withdrawBalance   float64
		transferBalance   float64
		transferToBalance float64
		curDate           = time.Now().Format("2006-01-02")
	)

	//get tblq_bonus_payout
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "tblq_bonus_payout.paid_ewallet_type_id = ?", CondValue: walletTypeID},
		models.WhereCondFn{Condition: "tblq_bonus_payout.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "tblq_bonus_payout.t_bns_id != ?", CondValue: curDate},
	)
	BnsRst, err := models.GetSumTotalBonusPayoutFn(arrCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberWithdrawBalance - GetSumTotalBonusPayoutFn Return Err", err, arrCond, true)
		return totalBalance, err
	}

	bonusBalance = BnsRst.TotalBonusPayout

	//deduct alr withdraw amount
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_withdraw.ewallet_type_id = ?", CondValue: walletTypeID},
		models.WhereCondFn{Condition: "ewt_withdraw.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ewt_withdraw.status NOT IN (?,'C')", CondValue: "R"},
	)
	WithdrawRst, err := models.GetSumTotalWithdrawFn(arrCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberWithdrawBalance - GetSumTotalWithdrawFn Return Err", err, arrCond, true)
		return totalBalance, err
	}

	withdrawBalance = WithdrawRst.TotalWithdraw

	//deduct transfer to own amount
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_transfer.ewt_type_from = ?", CondValue: walletTypeID},
		models.WhereCondFn{Condition: "ewt_transfer.member_id_from = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ewt_transfer.member_id_to = ?", CondValue: memID},
	)
	transferRst, err := models.GetSumTotalTransferFn(arrCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberWithdrawBalance - GetSumTotalTransferFn Return Err", err, arrCond, true)
		return totalBalance, err
	}

	transferBalance = transferRst.TotalTransfer

	//add transfer to own amount
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_transfer.ewt_type_to = ?", CondValue: walletTypeID},
		models.WhereCondFn{Condition: "ewt_transfer.member_id_from = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ewt_transfer.member_id_to = ?", CondValue: memID},
	)
	transferToRst, err := models.GetSumTotalTransferFn(arrCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberWithdrawBalance - GetSumTotalTransferToFn Return Err", err, arrCond, true)
		return totalBalance, err
	}

	transferToBalance = transferToRst.TotalTransfer

	totalBalance, _ = decimal.NewFromFloat(bonusBalance).Sub(decimal.NewFromFloat(withdrawBalance)).Float64()
	totalBalance, _ = decimal.NewFromFloat(totalBalance).Sub(decimal.NewFromFloat(transferBalance)).Float64()
	totalBalance, _ = decimal.NewFromFloat(totalBalance).Add(decimal.NewFromFloat(transferToBalance)).Float64()

	//get current wallet balance
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: walletTypeID},
	)
	currBalRst, err := models.GetMemberEwtSetupBalanceFn(memID, arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberWithdrawBalance - fail to get current wallet balance", err, arrCond, true)
		return totalBalance, err
	}

	if len(currBalRst) < 0 {
		base.LogErrorLog("GetMemberWithdrawBalance - empty wallet balance", err, arrCond, true)
		return totalBalance, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "empty_wallet_balance", Data: arrCond}
	}

	currBal := currBalRst[0].Balance

	//admin limiter
	limiter, _ := models.GetMemAdminLimiterByEwtTypeID(memID, walletTypeID)
	limiterBal, _ := decimal.NewFromFloat(totalBalance).Add(decimal.NewFromFloat(limiter.TotalLimitAmount)).Float64()

	if currBal < limiterBal {
		totalBalance = currBal
	} else {
		totalBalance = limiterBal
	}

	if totalBalance < 0 {
		totalBalance = float64(0)
	}

	return totalBalance, nil
}
