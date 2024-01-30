package product_service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/service/member_service"
)

// func GetProductsGroupSetting
func GetProductsGroupSetting(memID int, prdGroup, langCode string) (map[string]interface{}, string) {
	var (
		arrReturnData           = map[string]interface{}{}
		keyin           int     = 0
		keyinMin        float64 = 1
		keyinMultipleOf float64 = 1
		currencyCode    string  = ""
	)

	// get prd_group_type setting
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: prdGroup},
		models.WhereCondFn{Condition: "prd_group_type.status = ?", CondValue: "A"},
	)
	arrGetPrdGroupType, err := models.GetPrdGroupTypeFn(arrCond, "", false)
	if err != nil {
		base.LogErrorLog("GetProductsGroupSetting:GetPrdGroupTypeFn()", map[string]interface{}{"prdGroup": prdGroup, "arrCond": arrCond}, err.Error(), true)
		return nil, "something_went_wrong"
	}

	// process data
	if len(arrGetPrdGroupType) > 0 {
		currencyCode = arrGetPrdGroupType[0].CurrencyCode
		if arrGetPrdGroupType[0].PrincipleType == "KEYIN" {
			keyin = 1
		}

		if arrGetPrdGroupType[0].Setting != "" {
			arrPrdGroupTypeSetup, errMsg := GetPrdGroupTypeSetup(arrGetPrdGroupType[0].Setting)
			if errMsg != "" {
				return nil, errMsg
			}

			keyinMin = arrPrdGroupTypeSetup.KeyinMin
			keyinMultipleOf = arrPrdGroupTypeSetup.KeyinMultipleOf

			if arrGetPrdGroupType[0].Code == "CONTRACT" {
				// keyin min will be 50% of current value
				totalSalesFn := make([]models.WhereCondFn, 0)
				totalSalesFn = append(totalSalesFn,
					models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memID},
					models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "CONTRACT"},
					models.WhereCondFn{Condition: "sls_master.status IN(?,'EP')", CondValue: "AP"},
				)
				totalSales, _ := models.GetTotalSalesAmount(totalSalesFn, false)

				if totalSales.TotalBv > 0 {
					keyinMin = float.Div(totalSales.TotalBv, 2)
					keyinMin = math.Ceil(keyinMin)
				} else {
					keyinMin = 100
				}
			}

			if len(arrPrdGroupTypeSetup.Tiers) > 0 {
				arrReturnData["tiers"] = arrPrdGroupTypeSetup.Tiers
			}
		}
	}

	arrReturnData["keyin"] = keyin
	arrReturnData["keyin_min"] = keyinMin
	arrReturnData["keyin_multiple_of"] = keyinMultipleOf
	arrReturnData["currency_code"] = currencyCode
	arrReturnData["pdf_url"] = ""

	// if strings.ToLower(prdGroup) == "mining_bzz" {
	// 	serverDomain := setting.Cfg.Section("custom").Key("ApiServerDomain").String()
	// 	if strings.ToLower(langCode) == "zh" {
	// 		arrReturnData["pdf_url"] = serverDomain + "/templates/sales/node/view/sec_swarm_server_leasing_contract_node_zh.pdf"
	// 	} else {
	// 		arrReturnData["pdf_url"] = serverDomain + "/templates/sales/node/view/sec_swarm_server_leasing_contract_node_en.pdf"
	// 	}
	// } else if strings.ToLower(prdGroup) == "broadband" {
	// 	serverDomain := setting.Cfg.Section("custom").Key("ApiServerDomain").String()
	// 	if strings.ToLower(langCode) == "zh" {
	// 		arrReturnData["pdf_url"] = serverDomain + "/templates/sales/broadband/view/sec_swarm_server_leasing_contract_broadband_zh.pdf"
	// 	} else {
	// 		arrReturnData["pdf_url"] = serverDomain + "/templates/sales/broadband/view/sec_swarm_server_leasing_contract_broadband_en.pdf"
	// 	}
	// }

	return arrReturnData, ""
}

// func GetProductsv1
func GetProductsv1(memberID int, prdGroup string, langCode string) ([]map[string]interface{}, string) {
	var (
		// secMiningPrice, filecoinMiningPrice, chiaMiningPrice float64
		rebatePerc     string
		arrReturnData  []map[string]interface{}
		arrReturnDataV map[string]interface{}
	)

	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " prd_master.prd_group = ? ", CondValue: prdGroup},
		models.WhereCondFn{Condition: " prd_master.htmlfive_show = ? ", CondValue: 1},
	)
	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("product_service:GetProductsv1():GetPrdMasterFn():1", err.Error(), map[string]interface{}{"condition": arrPrdMasterFn}, true)
		return nil, "something_went_wrong"
	}

	if len(arrPrdMaster) > 0 {
		for _, arrPrdMasterV := range arrPrdMaster {
			var color []string
			json.Unmarshal([]byte(arrPrdMasterV.Color), &color)

			rebatePerc = fmt.Sprintf("%g", float.RoundUp(arrPrdMasterV.RebatePerc*(100), 2)) + "%"

			arrReturnDataV = map[string]interface{}{
				"code":          arrPrdMasterV.Code,
				"name":          helpers.TranslateV2(arrPrdMasterV.Name, langCode, nil),
				"path":          arrPrdMasterV.Path,
				"status":        arrPrdMasterV.Status,
				"color":         color,
				"amount":        arrPrdMasterV.Amount,
				"currency_code": arrPrdMasterV.CurrencyCode,
				"rebate_perc":   rebatePerc,
				"income_cap":    arrPrdMasterV.IncomeCap,
			}

			arrReturnData = append(arrReturnData, arrReturnDataV)
		}
	}
	return arrReturnData, ""
}

// type MemberMiningActionListv1ReqStruct
type MemberMiningActionListv1ReqStruct struct {
	EntMemberID int
	LangCode    string
	Version     string
}

type AmountArray struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}
type MemberMiningActionListv1RstStruct struct {
	MiningType string `json:"mining_type"`
	Name       string `json:"name"`
	ImageURL   string `json:"image_url"`
	// CryptoType     string        `json:"crypto_type"`
	CurrencyAmount string        `json:"currency_amount"`
	Show           bool          `json:"show"`
	AmountArray    []AmountArray `json:"amount_array"`
}

// func GetMemberMiningActionListv1
func GetMemberMiningActionListv1(arrData MemberMiningActionListv1ReqStruct) ([]MemberMiningActionListv1RstStruct, string) {
	// get current liga price
	ligaPrice, ligaPriceErr := base.GetLatestPriceMovementByTokenType("LIGA")
	if ligaPriceErr != nil {
		base.LogErrorLog("GetMemberMiningActionListv1()", "GetLatestLigaPriceMovement", ligaPriceErr.Error(), true)
		return nil, "something_went_wrong"
	}
	ligaEwtSetupCond := make([]models.WhereCondFn, 0)
	ligaEwtSetupCond = append(ligaEwtSetupCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "LIGA"},
	)
	ligaEwtSetup, ligaEwtSetupErr := models.GetEwtSetupFn(ligaEwtSetupCond, "", false)
	if ligaEwtSetupErr != nil {
		base.LogErrorLog("GetMemberMiningActionListv1()", "GetEwtSetupFn", ligaEwtSetupErr.Error(), true)
		return nil, "something_went_wrong"
	}

	// get setting from sys_general_setup
	generalSetupRst, generalSetupErr := models.GetSysGeneralSetupByID("mining_action_list_setting")

	if generalSetupErr != nil {
		base.LogErrorLog("GetMemberMiningActionListv1()", "GetSysGeneralSetupByID", generalSetupErr.Error(), true)
		return nil, "something_went_wrong"
	}
	var arrDataReturn []MemberMiningActionListv1RstStruct
	newArrDataReturn := make([]MemberMiningActionListv1RstStruct, 0)
	json.Unmarshal([]byte(generalSetupRst.SettingValue1), &arrDataReturn)

	latestExchangePrice, err := base.GetLatestExchangePriceMovementByTokenType("SEC")
	if err != nil {
		base.LogErrorLog("GetMemberMiningActionListv1-GetLatestExchangePriceMovementByTokenType_failed", err.Error(), "sec", true)
		return nil, "something_went_wrong"
	}
	secEwtSetupCond := make([]models.WhereCondFn, 0)
	secEwtSetupCond = append(secEwtSetupCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "SEC"},
	)
	secEwtSetup, err := models.GetEwtSetupFn(secEwtSetupCond, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberMiningActionListv1-GetEwtSetupFn", err.Error(), secEwtSetupCond, true)
		return nil, "something_went_wrong"
	}

	for i, action := range arrDataReturn {
		// if arrData.Version == "old" {
		// 	if strings.ToLower(arrDataReturn[i].MiningType) == "mining" {
		// 		continue
		// 	}
		// } else if arrData.Version == "new" {
		// 	if strings.ToLower(arrDataReturn[i].MiningType) == "staking" {
		// 		continue
		// 	}
		// }

		arrDataReturn[i].Name = helpers.TranslateV2(action.Name, arrData.LangCode, nil)
		// arrDataReturn[i].CryptoType = helpers.TranslateV2(action.CryptoType, arrData.LangCode, nil)
		arrDataReturn[i].CurrencyAmount = helpers.TranslateV2(action.CurrencyAmount, arrData.LangCode, nil)

		arrDataReturn[i].AmountArray = make([]AmountArray, 0)
		if action.MiningType == "STAKING" {
			currencyCodeArr := []string{
				"SEC", "LIGA",
			}
			for _, currencyCode := range currencyCodeArr {
				totalSalesCond := make([]models.WhereCondFn, 0)
				totalSalesCond = append(totalSalesCond,
					models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: arrData.EntMemberID},
					models.WhereCondFn{Condition: "sls_master.action = ?", CondValue: arrDataReturn[i].MiningType},
					models.WhereCondFn{Condition: "sls_master.currency_code = ?", CondValue: currencyCode})
				totalSales, totalSalesErr := models.GetTotalSalesAmount(totalSalesCond, false)

				if totalSalesErr != nil {
					base.LogErrorLog("product:GetProductsv1()", "GetTotalSalesAmount()", totalSalesErr.Error(), true)
					return nil, "something_went_wrong"
				}

				decimalPoint := uint(2)
				if totalSales.DecimalPoint > 0 {
					decimalPoint = uint(totalSales.DecimalPoint)
				}
				// fmt.Println(totalSales.TotalAmount, totalSales.RefundAmount)
				amount := helpers.CutOffDecimal(float.Sub(totalSales.TotalAmount, totalSales.RefundAmount), decimalPoint, ".", ",")
				arrDataReturn[i].AmountArray = append(arrDataReturn[i].AmountArray, AmountArray{
					Amount:       amount,
					CurrencyCode: currencyCode,
				})
			}
		} else if action.MiningType == "POOL" {
			totalSalesCond := make([]models.WhereCondFn, 0)
			totalSalesCond = append(totalSalesCond,
				models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: arrData.EntMemberID},
				models.WhereCondFn{Condition: "sls_master.action = ?", CondValue: "P2P"},
				models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"})
			totalSales, totalSalesErr := models.GetTotalSalesAmount(totalSalesCond, false)

			if totalSalesErr != nil {
				base.LogErrorLog("product:GetProductsv1()", "GetTotalSalesAmount()", totalSalesErr.Error(), true)
				return nil, "something_went_wrong"
			}

			decimalPoint := uint(2)
			if totalSales.DecimalPoint > 0 {
				decimalPoint = uint(totalSales.DecimalPoint)
			}

			amount := helpers.CutOffDecimal(totalSales.TotalAmount, decimalPoint, ".", ",")
			arrDataReturn[i].AmountArray = append(arrDataReturn[i].AmountArray, AmountArray{
				Amount:       amount,
				CurrencyCode: totalSales.CurrencyCode,
			})
			// assign liga price
			arrDataReturn[i].AmountArray = append(arrDataReturn[i].AmountArray, AmountArray{
				Amount:       helpers.CutOffDecimalv2(latestExchangePrice, uint(ligaEwtSetup.DecimalPoint), ".", ",", true),
				CurrencyCode: secEwtSetup.CurrencyCode,
			})
		} else if strings.ToLower(action.MiningType) == "mining" {

		} else if strings.ToLower(action.MiningType) == "contract" {

			arrSlsCond := make([]models.WhereCondFn, 0)
			arrSlsCond = append(arrSlsCond,
				models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: arrData.EntMemberID},
				models.WhereCondFn{Condition: "sls_master.status NOT IN ('V') AND sls_master.doc_type = ?", CondValue: "CT"},
			)
			arrCTSls, _ := models.GetSlsMasterFn(arrSlsCond, "", false)

			if len(arrCTSls) < 1 {
				continue
			}
		} else {
			totalSalesCond := make([]models.WhereCondFn, 0)
			totalSalesCond = append(totalSalesCond,
				models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: arrData.EntMemberID},
				models.WhereCondFn{Condition: "sls_master.action = ?", CondValue: arrDataReturn[i].MiningType},
				models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"})
			totalSales, totalSalesErr := models.GetTotalSalesAmount(totalSalesCond, false)

			if totalSalesErr != nil {
				base.LogErrorLog("product:GetProductsv1()", "GetTotalSalesAmount()", totalSalesErr.Error(), true)
				return nil, "something_went_wrong"
			}

			decimalPoint := uint(2)
			if totalSales.DecimalPoint > 0 {
				decimalPoint = uint(totalSales.DecimalPoint)
			}

			amount := helpers.CutOffDecimal(totalSales.TotalAmount, decimalPoint, ".", ",")
			arrDataReturn[i].AmountArray = append(arrDataReturn[i].AmountArray, AmountArray{
				Amount:       amount,
				CurrencyCode: totalSales.CurrencyCode,
			})
			// assign liga price
			arrDataReturn[i].AmountArray = append(arrDataReturn[i].AmountArray, AmountArray{
				Amount:       helpers.CutOffDecimalv2(ligaPrice, uint(ligaEwtSetup.DecimalPoint), ".", ",", true),
				CurrencyCode: ligaEwtSetup.CurrencyCode,
			})
		}

		newArrDataReturn = append(newArrDataReturn,
			arrDataReturn[i],
		)
	}

	// arrDataReturn = append(arrDataReturn,
	// 	MemberMiningActionListv1RstStruct{
	// 		MiningType:     "CONTRACT",
	// 		Name:           helpers.TranslateV2("contract", arrData.LangCode, nil),
	// 		ImageURL:       "https://media02.securelayers.cloud/medias/SEC/MINING_ACTION/IMAGES/icon-contract.png",
	// 		Amount:         123.21523542,
	// 		CryptoType:     helpers.TranslateV2("liga", arrData.LangCode, nil),
	// 		CurrencyAmount: helpers.TranslateV2("price", arrData.LangCode, nil),
	// 	},
	// )
	// arrDataReturn = append(arrDataReturn,
	// 	MemberMiningActionListv1RstStruct{
	// 		MiningType:     "STAKING",
	// 		Name:           helpers.TranslateV2("staking", arrData.LangCode, nil),
	// 		ImageURL:       "https://media02.securelayers.cloud/medias/SEC/MINING_ACTION/IMAGES/icon-staking.png",
	// 		Amount:         123.00350042,
	// 		CryptoType:     helpers.TranslateV2("sec", arrData.LangCode, nil),
	// 		CurrencyAmount: helpers.TranslateV2("hash", arrData.LangCode, nil),
	// 	},
	// )
	// arrDataReturn = append(arrDataReturn,
	// 	MemberMiningActionListv1RstStruct{
	// 		MiningType:     "POOL",
	// 		Name:           helpers.TranslateV2("pool", arrData.LangCode, nil),
	// 		ImageURL:       "https://media02.securelayers.cloud/medias/SEC/MINING_ACTION/IMAGES/icon-pool.png",
	// 		Amount:         3.21354200,
	// 		CryptoType:     helpers.TranslateV2("sec", arrData.LangCode, nil),
	// 		CurrencyAmount: helpers.TranslateV2("hash", arrData.LangCode, nil),
	// 	},
	// )
	return newArrDataReturn, ""
}

// type MemberContractMiningActionDetailsv1ReqStruct
type MemberContractMiningActionDetailsv1ReqStruct struct {
	EntMemberID int
	LangCode    string
}

// type MarketPriceListStruct
type MarketPriceListStruct struct {
	PriceRate string `json:"price_rate"`
	TransAt   string `json:"trans_at"`
}

type MarketPriceStruct struct {
	Min                float64                 `json:"min"`
	Max                float64                 `json:"max"`
	Interval           float64                 `json:"interval"`
	YAxisDecimalPoint  uint8                   `json:"y_axis_decimal_point"`
	MarketPriceList    []MarketPriceListStruct `json:"market_price_list"`
	CurrentMarketPrice string                  `json:"currenct_market_price"`
	CryptoType         string                  `json:"crypto_type"`
	CurrencyAmount     string                  `json:"currency_amount"`
}

// type MemberContractMiningActionDetailsv1RstStruct
type MemberContractMiningActionDetailsv1RstStruct struct {
	MarketPrice MarketPriceStruct `json:"market_price"`
	// MiningCap struct {
	// 	CurrentMiningAmount string  `json:"current_mining"`
	// 	MiningFrom          string  `json:"mining_from"`
	// 	MiningTo            string  `json:"mining_to"`
	// 	MiningPercent       float64 `json:"mining_percent"`
	// } `json:"mining_cap"`
	TodaySponsor            string `json:"today_sponsor"`
	AccumulatedSalesAmount  string `json:"accumulated_amount"`
	TodaySponsorSalesAmount string `json:"today_sales_amount"`
	MatchingLevel           string `json:"matching_level"`
	TotalSponsor            string `json:"total_sponsor"`
	TotalSponsorSalesAmount string `json:"total_sponsor_sales_amount"`
	PurchaseContractStatus  int    `json:"purchase_contract_status"`
	IncomeCap               struct {
		Total   string  `json:"total"`
		Balance string  `json:"balance"`
		Percent float64 `json:"percent"`
	} `json:"income_cap"`
	PoolAmount     string `json:"pool_amount"`
	LigaPoolAmount string `json:"liga_pool_amount"`
	UsdPoolAmount  string `json:"usd_pool_amount"`
}

// func GetMemberContractMiningActionDetailsv1
func GetMemberContractMiningActionDetailsv1(arrData MemberContractMiningActionDetailsv1ReqStruct) (*MemberContractMiningActionDetailsv1RstStruct, string) {

	// start get liga price
	var (
		decimalPoint       uint
		currentMarketPrice string
		minTP              float64
	)
	dtNow := base.GetCurrentTime("2006-01-02")
	totRecord := 7

	arrPriceMovementIndByTokenType := base.PriceMovementIndByTokenTypeStruct{
		MemberID:  arrData.EntMemberID,
		TokenType: "LIGA",
		Limit:     7,
		Date:      base.GetCurrentDateTimeT(),
	}
	arrMarketPriceRst, _ := base.GetPriceMovementIndByTokenTypeFn(arrPriceMovementIndByTokenType)

	// for _, rstV := range rst {
	// 	fmt.Println(rstV.CreatedAt, rstV.TokenPrice)
	// }

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		// models.WhereCondFn{Condition: " liga_price_movement.created_at <= (SELECT created_at FROM liga_price_movement where b_latest = 1)"},
		models.WhereCondFn{Condition: " DATE(liga_price_movement.created_at) <= '" + dtNow + "'"},
	)
	// arrMarketPriceRst, _ := models.GetLigaPriceMovementFn(arrCond, totRecord, false)
	arrMinMaxMarketPriceRst, _ := models.GetMinMaxLigaPriceMovementFn(arrCond, totRecord, false)
	ligaEwtSetupCond := make([]models.WhereCondFn, 0)
	ligaEwtSetupCond = append(ligaEwtSetupCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "LIGA"},
	)
	ligaEwtSetup, _ := models.GetEwtSetupFn(ligaEwtSetupCond, "", false)
	decimalPoint = 2
	blockChainDecimalPoint := uint(8)
	if ligaEwtSetup != nil {
		// decimalPoint = uint(ligaEwtSetup.DecimalPoint)
		blockChainDecimalPoint = uint(ligaEwtSetup.DecimalPoint)
	}
	marketPriceList := make([]MarketPriceListStruct, 0)
	// if len(arrMarketPriceRst) < totRecord {
	// balRec := totRecord - len(arrMarketPriceRst)
	for _, arrMarketPriceRstV := range arrMarketPriceRst {

		marketPrice1 := make([]MarketPriceListStruct, 1)
		marketPrice1[0].PriceRate = helpers.CutOffDecimal(arrMarketPriceRstV.TokenPrice, blockChainDecimalPoint, ".", ",")
		marketPrice1[0].TransAt = base.TimeFormat(arrMarketPriceRstV.CreatedAt, "2006-01-02 15:04:05")
		marketPriceList = append(marketPrice1, marketPriceList...)
	}
	ligaTokenRate, err := base.GetLatestPriceMovementByTokenType("LIGA")
	if err != nil {
		base.LogErrorLog("GetMemberContractMiningActionDetailsv1-GetLatestPriceMovementByTokenType_failed", err.Error(), ligaTokenRate, true)
		return nil, "something_went_wrong"
	}

	currentMarketPrice = helpers.CutOffDecimalv2(ligaTokenRate, blockChainDecimalPoint, ".", ",", true)
	// fmt.Println("decimalPoint:", decimalPoint)
	intervalPrice := float.RoundUp((arrMinMaxMarketPriceRst.MaxTokenPrice-arrMinMaxMarketPriceRst.MinTokenPrice)/float64(5), int(decimalPoint))
	// fmt.Println("MaxTokenPrice:", arrMinMaxMarketPriceRst.MaxTokenPrice)
	// fmt.Println("MinTokenPrice:", arrMinMaxMarketPriceRst.MinTokenPrice)
	// fmt.Println("intervalPrice:", intervalPrice)
	// for i := 0; i < balRec; i++ {
	// 	// priceRate := []string{float.RoundDown(arrMarketPriceRst[0].TokenPrice, 2)}
	// 	arrFirstMarketPrice := make([]MarketPriceListStruct, 0)

	// 	arrFirstMarketPrice = append(arrFirstMarketPrice,
	// 		MarketPriceListStruct{
	// 			PriceRate: helpers.CutOffDecimal(arrMarketPriceRst[0].TokenPrice, decimalPoint, ".", ","),
	// 			TransAt:   base.TimeFormat(arrMarketPriceRst[0].CreatedAt, "2006-01-02 15:04:05"),
	// 		},
	// 	)
	// 	// priceRate := []string{helpers.CutOffDecimal(arrMarketPriceRst[0].TokenPrice, decimalPoint, ".", ",")}
	// 	marketPriceList = append(arrFirstMarketPrice, marketPriceList...)
	// }
	// }

	if arrMinMaxMarketPriceRst.MinTokenPrice-intervalPrice > 0 {
		minTP = arrMinMaxMarketPriceRst.MinTokenPrice - intervalPrice
	}
	intervalPrice = 0.1 // hard code first
	minTP = 1.3         // hard code first
	// maxTP := arrMinMaxMarketPriceRst.MaxTokenPrice + intervalPrice
	maxTP := 1.8
	yAxisDecimalPoint := uint8(2)
	arrDataReturn := MemberContractMiningActionDetailsv1RstStruct{}
	arrDataReturn.MarketPrice.MarketPriceList = marketPriceList
	arrDataReturn.MarketPrice.CurrentMarketPrice = currentMarketPrice
	arrDataReturn.MarketPrice.CryptoType = helpers.TranslateV2("liga", arrData.LangCode, nil)
	arrDataReturn.MarketPrice.CurrencyAmount = helpers.TranslateV2("price", arrData.LangCode, nil)
	arrDataReturn.MarketPrice.Min = minTP
	arrDataReturn.MarketPrice.Max = maxTP
	arrDataReturn.MarketPrice.Interval = intervalPrice
	arrDataReturn.MarketPrice.YAxisDecimalPoint = yAxisDecimalPoint

	// arrDataReturn.MarketPrice.Interval = 0.002

	// arrDataReturn.Ranking.RankFromImageURL = ""
	// arrDataReturn.Ranking.RankToImageURL = ""
	// arrDataReturn.MiningCap.CurrentMiningAmount = "4432"
	// arrDataReturn.MiningCap.MiningFrom = "0"
	// arrDataReturn.MiningCap.MiningTo = "9000"
	// arrDataReturn.MiningCap.MiningPercent = 49
	// arrDataReturn.AccumulatedSalesAmount = "3,000 (3 day)"
	arrDataReturn.PurchaseContractStatus = 1

	// get income cap info
	ewtSummaryCond := make([]models.WhereCondFn, 0)
	ewtSummaryCond = append(ewtSummaryCond,
		models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: 7},
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: arrData.EntMemberID},
	)
	ewtSummaryRst, ewtSummaryErr := models.GetEwtSummarySetupFn(ewtSummaryCond, "", false)

	if ewtSummaryErr != nil {
		base.LogErrorLog("GetMemberContractMiningActionDetailsv1()", "GetEwtSummaryFn", ewtSummaryErr.Error(), true)
		return nil, "something_went_wrong"
	}

	if len(ewtSummaryRst) > 0 {
		arrDataReturn.IncomeCap.Balance = helpers.CutOffDecimal(ewtSummaryRst[0].Balance, uint(ewtSummaryRst[0].DecimalPoint), ".", ",")
		arrDataReturn.IncomeCap.Total = helpers.CutOffDecimal(ewtSummaryRst[0].TotalIn, uint(ewtSummaryRst[0].DecimalPoint), ".", ",")
		arrDataReturn.IncomeCap.Percent = (ewtSummaryRst[0].TotalIn - ewtSummaryRst[0].Balance) / (ewtSummaryRst[0].TotalIn)
	} else {
		arrDataReturn.IncomeCap.Balance = "0.00"
		arrDataReturn.IncomeCap.Total = "0.00"
	}

	// get today direct sponsor
	todayDirectSponsor := "0"
	todayDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")
	arrTodayDirectSponsorCond := make([]models.WhereCondFn, 0)
	arrTodayDirectSponsorCond = append(arrTodayDirectSponsorCond,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "date(ent_member_tree_sponsor.created_at) = ?", CondValue: todayDate},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	todayDirectSponsorRst, todayDirectSponsorErr := models.GetTotalDirectSponsorFn(arrTodayDirectSponsorCond, false)

	if todayDirectSponsorErr != nil {
		base.LogErrorLog("GetMemberContractMiningActionDetailsv1()", "GetTodayDirectSponsor", todayDirectSponsorErr.Error(), true)
		return nil, "something_went_wrong"
	}

	if todayDirectSponsorRst.TotalDirectSponsor > 0 {
		todayDirectSponsor = helpers.CutOffDecimal(todayDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
	}
	arrDataReturn.TodaySponsor = todayDirectSponsor

	// get total direct sponsor
	totalDirectSponsor := "0"
	arrTotalDirectSponsorCond := make([]models.WhereCondFn, 0)
	arrTotalDirectSponsorCond = append(arrTotalDirectSponsorCond,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	totalDirectSponsorRst, totalDirectSponsorErr := models.GetTotalDirectSponsorFn(arrTotalDirectSponsorCond, false)

	if totalDirectSponsorErr != nil {
		base.LogErrorLog("GetMemberContractMiningActionDetailsv1()", "GetTotalDirectSponsorFn", totalDirectSponsorErr.Error(), true)
		return nil, "something_went_wrong"
	}

	if totalDirectSponsorRst.TotalDirectSponsor > 0 {
		totalDirectSponsor = helpers.CutOffDecimal(totalDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
	}
	arrDataReturn.TotalSponsor = totalDirectSponsor
	arrDataReturn.MatchingLevel = totalDirectSponsor

	// get today direct sponsor sales amount

	// get today total bv from sls_master (without topup)
	todaySalesCond := make([]models.WhereCondFn, 0)
	todaySalesCond = append(todaySalesCond,
		models.WhereCondFn{Condition: "sls_master.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "sls_master.doc_type = ?", CondValue: "CT"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "date(sls_master.created_at) = ?", CondValue: todayDate},
	)
	todaySales, todaySalesErr := models.GetTotalSalesAmount(todaySalesCond, false)
	if todaySalesErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetTodaySalesAmount", todaySalesErr.Error(), true)
		return nil, "something_went_wrong"
	}
	todaySalesDecimalPoint := todaySales.DecimalPoint

	// get today topup amount from sls_master_topup
	todayTopupCond := make([]models.WhereCondFn, 0)
	todayTopupCond = append(todayTopupCond,
		models.WhereCondFn{Condition: "sls_master.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "sls_master.doc_type = ?", CondValue: "CT"},
		models.WhereCondFn{Condition: "sls_master_topup.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "date(sls_master_topup.created_at) = ?", CondValue: todayDate},
	)
	todayTopup, todayTopupErr := models.GetSlsMasterTopupFn(todayTopupCond, "", false)
	if todayTopupErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetTodaySalesTopup", todayTopupErr.Error(), true)
		return nil, "something_went_wrong"
	}

	todayTopupAmt := 0.00
	if len(todayTopup) > 0 {
		todayTopupAmt = todayTopup[0].TotalAmount
		todaySalesDecimalPoint = todayTopup[0].DecimalPoint
	}
	arrDataReturn.TodaySponsorSalesAmount = helpers.CutOffDecimal(float.Add(todaySales.TotalBv, todayTopupAmt), uint(todaySalesDecimalPoint), ".", ",")

	// get total direct sponsor sales amount
	totalSalesCond := make([]models.WhereCondFn, 0)
	totalSalesCond = append(totalSalesCond,
		models.WhereCondFn{Condition: "sls_master.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "sls_master.doc_type = ?", CondValue: "CT"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	totalSales, totalSalesErr := models.GetTotalSalesAmount(totalSalesCond, false)
	if totalSalesErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetTotalSalesAmount()", totalSalesErr.Error(), true)
		return nil, "something_went_wrong"
	}
	arrDataReturn.TotalSponsorSalesAmount = helpers.CutOffDecimal(totalSales.TotalAmount, uint(totalSales.DecimalPoint), ".", ",")

	// set matching level
	if totalDirectSponsorRst.TotalDirectSponsor < 15 && totalSales.TotalAmount > 10000 {
		arrDataReturn.MatchingLevel = "15"
	}
	// get accumulated amount
	accAmtCond := make([]models.WhereCondFn, 0)
	accAmtCond = append(accAmtCond,
		models.WhereCondFn{Condition: "t_member_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02")},
	)
	accAmtRst, accAmtErr := models.GetTblqBonusSponsorFn(accAmtCond, false)
	if accAmtErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetTblBonusSponsorFn()", accAmtErr.Error(), true)
		return nil, "something_went_wrong"
	}

	accAmt := 0.00
	accDay := 0
	if len(accAmtRst) > 0 {
		accAmt = float64(accAmtRst[0].NShare)
		accDay = accAmtRst[0].NDay
	}
	arrDataReturn.AccumulatedSalesAmount = helpers.CutOffDecimal(accAmt, 2, ".", ",") + " (" + strconv.Itoa(accDay) + helpers.TranslateV2("days", arrData.LangCode, nil) + ")"

	// get sponsor pool amount
	poolAmtCond := make([]models.WhereCondFn, 0)
	poolAmtCond = append(poolAmtCond,
		models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02")},
	)
	poolAmtRst, poolAmtErr := models.GetTblBonusSponsorPoolFn(poolAmtCond, "", false)
	if poolAmtErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetTblBonusSponsorFn()", poolAmtErr.Error(), true)
		return nil, "something_went_wrong"
	}

	poolAmt := 0.00
	if len(poolAmtRst) > 0 {
		poolAmt = float64(poolAmtRst[0].PoolCf)
	}

	todaySalesBvCond := make([]models.WhereCondFn, 0)
	todaySalesBvCond = append(todaySalesBvCond,
		models.WhereCondFn{Condition: "sls_master.doc_type = ?", CondValue: "CT"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "date(sls_master.doc_date) = ?", CondValue: todayDate},
	)
	todaySalesBv, todaySalesBvErr := models.GetTotalSalesAmount(todaySalesBvCond, false)
	if todaySalesBvErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetTotalSalesBv()", todaySalesBvErr.Error(), true)
		return nil, "something_went_wrong"
	}

	//get sponsor pool markup
	sponsorMarkupCond := make([]models.WhereCondFn, 0)
	sponsorMarkupCond = append(sponsorMarkupCond,
		models.WhereCondFn{Condition: "bns_date = ?", CondValue: todayDate},
	)
	sponsorMarkup, sponsorMarkupErr := models.GetSysSponsorPoolMarkupFn(sponsorMarkupCond, "SUM(pool_amount) as pool_amount", false)

	if sponsorMarkupErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetSysSponsorPoolMarkupFn()", sponsorMarkupErr.Error(), true)
		return nil, "something_went_wrong"
	}

	markupAmt := 0.00
	if len(sponsorMarkup) > 0 {
		markupAmt = sponsorMarkup[0].PoolAmount
	}

	// sponsor pool amount = yesterday pool cf + (today's sales bv x 10%) + markup
	arrDataReturn.PoolAmount = helpers.CutOffDecimal(float.Add(float.Add(poolAmt, (todaySalesBv.TotalAmount*0.1)), markupAmt), 2, ".", ",")

	// get liga pool amount and usds pool amount
	ligaPoolAmtCond := make([]models.WhereCondFn, 0)
	ligaPoolAmtCond = append(ligaPoolAmtCond,
		models.WhereCondFn{Condition: "status = ?", CondValue: "AP"},
	)
	ligaPoolAmtRst, ligaPoolAmtErr := models.GetEwtWithdrawPoolTotal(ligaPoolAmtCond, false)
	if ligaPoolAmtErr != nil {
		base.LogErrorLog("product_service:GetMemberContractMiningActionDetailsv1()", "GetEwtWithdrawPoolTotal()", ligaPoolAmtErr.Error(), true)
		return nil, "something_went_wrong"
	}
	// get usdt ewt setup
	usdtEwtSetupCond := make([]models.WhereCondFn, 0)
	usdtEwtSetupCond = append(usdtEwtSetupCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "USDT"},
	)
	usdtEwtSetup, _ := models.GetEwtSetupFn(usdtEwtSetupCond, "", false)

	ligaPoolAmt := ligaPoolAmtRst.LigaPool
	arrDataReturn.LigaPoolAmount = helpers.CutOffDecimalv2(ligaPoolAmt, uint(ligaEwtSetup.DecimalPoint), ".", ",", true) + " " + ligaEwtSetup.CurrencyCode
	usdPoolAmt := ligaPoolAmtRst.UsdPool
	arrDataReturn.UsdPoolAmount = helpers.CutOffDecimalv2(usdPoolAmt, uint(usdtEwtSetup.DecimalPoint), ".", ",", true) + " " + usdtEwtSetup.CurrencyCode

	return &arrDataReturn, ""
}

// type MemberStakingMiningActionDetailsv1ReqStruct
type MemberStakingMiningActionDetailsv1ReqStruct struct {
	EntMemberID int
	LangCode    string
}

// type CryptoStakingListStruct
type CryptoStakingListStruct struct {
	CryptoType        string  `json:"crypto_type"`
	StakeUnit         float64 `json:"stake_unit"`
	StakeUnitDisplay  string  `json:"stake_unit_display"`
	StakePrice        float64 `json:"stake_price"`
	StakePriceDisplay string  `json:"stake_price_display"`
}

// type MemberStakingMiningActionDetailsv1RstStruct
type MemberStakingMiningActionDetailsv1RstStruct struct {
	MarketPrice struct {
		Min                float64                 `json:"min"`
		Max                float64                 `json:"max"`
		Interval           float64                 `json:"interval"`
		MarketPriceList    []MarketPriceListStruct `json:"market_price_list"`
		CurrentMarketPrice string                  `json:"currenct_market_price"`
		CryptoType         string                  `json:"crypto_type"`
		CurrencyAmount     string                  `json:"currency_amount"`
	} `json:"market_price"`
	CryptoStakeList                 []CryptoStakingListStruct `json:"crypto_stake_list"`
	ComputingPower                  string                    `json:"computing_power"`
	ComputingPowerCurrency          string                    `json:"computing_power_currency"`
	ComputingPowerValuation         string                    `json:"computing_power_valuation"`
	ComputingPowerValuationCurrency string                    `json:"computing_power_valuation_currency"`
	PurchaseStakingStatus           int                       `json:"purchase_staking_status"`
}

// func GetMemberStakingMiningActionDetailsv1
func GetMemberStakingMiningActionDetailsv1(arrData MemberStakingMiningActionDetailsv1ReqStruct) (*MemberStakingMiningActionDetailsv1RstStruct, string) {

	decimalPoint := uint(8)
	totRecord := 7
	dateFormat := base.ConvertFormat("yyyy-mm-dd")
	endDateInString := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)
	startDateInString := base.GetCurrentDateTimeT().AddDate(0, 0, -7).Format(dateFormat)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " bns_id >= ? ", CondValue: startDateInString},
		models.WhereCondFn{Condition: " bns_id <= ? ", CondValue: endDateInString},
	)
	arrMarketPriceRst, _ := models.TblqBonusStakingRebateFn(arrCond, "", false)

	// arrange marketPriceList
	arrMarketPrice := make([]MarketPriceListStruct, totRecord)
	startDate, _ := time.Parse(dateFormat, startDateInString)
	for i, _ := range arrMarketPrice {
		priceDate := startDate.AddDate(0, 0, i)
		arrMarketPrice[i].PriceRate = helpers.CutOffDecimal(0.00, decimalPoint, ".", ",")
		arrMarketPrice[i].TransAt = priceDate.Format(dateFormat)

		for _, bnsRebate := range arrMarketPriceRst {
			bnsDate, _ := time.Parse(dateFormat, bnsRebate.BnsId)
			if priceDate.Equal(bnsDate) {
				arrMarketPrice[i].PriceRate = helpers.CutOffDecimal(bnsRebate.PersonalAsset, decimalPoint, ".", ",")
				arrMarketPrice[i].TransAt = bnsRebate.BnsId
			}
		}
	}

	// get min and max price, current price
	minPrice := 0.00
	maxPrice := 0.00
	currentPrice := "0.00"
	for i, bnsRebate := range arrMarketPriceRst {
		if i == 0 {
			minPrice = bnsRebate.PersonalAsset
			maxPrice = bnsRebate.PersonalAsset
		} else {
			if bnsRebate.PersonalAsset > maxPrice {
				maxPrice = bnsRebate.PersonalAsset
			}
			if bnsRebate.PersonalAsset < minPrice {
				minPrice = bnsRebate.PersonalAsset
			}
		}

		if i == len(arrMarketPriceRst)-1 {
			currentPrice = helpers.CutOffDecimal(arrMarketPriceRst[i].PersonalAsset, decimalPoint, ".", ",")
		}
	}

	if len(arrMarketPriceRst) < totRecord {
		minPrice = 0.00
	}
	intervalPrice := float.RoundUp((maxPrice-minPrice)/float64(5), int(decimalPoint))
	maxPrice = maxPrice + intervalPrice
	if minPrice-intervalPrice > 0 {
		minPrice = minPrice - intervalPrice
	}
	if maxPrice == 0 {
		intervalPrice = 0.5
		maxPrice = 1
	}

	// arrMinMaxMarketPriceRst, _ := models.GetMinMaxSecPriceMovementFn(arrCond, totRecord, false)

	// arrCond = make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "SEC"},
	// )
	// ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	// decimalPoint = 2
	// if ewtSetup != nil {
	// 	decimalPoint = uint(ewtSetup.DecimalPoint)
	// }
	// marketPriceList := make([]MarketPriceListStruct, 0)
	// // if len(arrMarketPriceRst) < totRecord {
	// // balRec := totRecord - len(arrMarketPriceRst)
	// for _, arrMarketPriceRstV := range arrMarketPriceRst {

	// 	marketPrice1 := make([]MarketPriceListStruct, 1)
	// 	marketPrice1[0].PriceRate = helpers.CutOffDecimal(arrMarketPriceRstV.TokenPrice, decimalPoint, ".", ",")
	// 	marketPrice1[0].TransAt = base.TimeFormat(arrMarketPriceRstV.CreatedAt, "2006-01-02 15:04:05")
	// 	marketPriceList = append(marketPrice1, marketPriceList...)

	// 	if arrMarketPriceRstV.BLatest == 1 {
	// 		currentMarketPrice = helpers.CutOffDecimal(arrMarketPriceRstV.TokenPrice, decimalPoint, ".", ",")
	// 	}
	// }

	// get sec total sales
	totalSecSalesCond := make([]models.WhereCondFn, 0)
	totalSecSalesCond = append(totalSecSalesCond,
		models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "prd_master.code = ?", CondValue: "SEC"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
	)
	totalSecSales, totalSecSalesErr := models.GetTotalSalesAmount(totalSecSalesCond, false)
	if totalSecSalesErr != nil {
		base.LogErrorLog("product_service:GetMemberStakingMiningActionDetailsv1()", "GetSecTotalSalesAmount()", totalSecSalesErr.Error(), true)
		return nil, "something_went_wrong"
	}

	// get liga total sales
	totalLigaSalesCond := make([]models.WhereCondFn, 0)
	totalLigaSalesCond = append(totalLigaSalesCond,
		models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "prd_master.code = ?", CondValue: "LIGA"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
	)
	totalLigaSales, totalLigaSalesErr := models.GetTotalSalesAmount(totalLigaSalesCond, false)
	if totalLigaSalesErr != nil {
		base.LogErrorLog("product_service:GetMemberStakingMiningActionDetailsv1()", "GetLigaTotalSalesAmount()", totalLigaSalesErr.Error(), true)
		return nil, "something_went_wrong"
	}

	// get sec current price
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "b_latest = ?", CondValue: 1},
	)
	arrSecPriceRst, _ := models.GetSecPriceMovementFn(arrCond, 1, false)

	currentSecPrice := 0.00
	if len(arrSecPriceRst) > 0 {
		currentSecPrice = arrSecPriceRst[0].TokenPrice
	}

	// get sec decimal point
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "SEC"},
	)
	ewtSetupSec, _ := models.GetEwtSetupFn(arrCond, "", false)
	secDecimalPoint := uint(2)
	if ewtSetupSec != nil {
		secDecimalPoint = uint(ewtSetupSec.DecimalPoint)
	}

	// get liga current price
	ligaTokenRate, err := base.GetLatestPriceMovementByTokenType("LIGA")
	if err != nil {
		base.LogErrorLog("GetMemberStakingMiningActionDetailsv1-GetLatestPriceMovementByTokenType", err.Error(), ligaTokenRate, true)
		return nil, "something_went_wrong"
	}

	currentLigaPrice := 0.00
	if ligaTokenRate > 0 {
		currentLigaPrice = ligaTokenRate
	}

	// get liga decimal point
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "LIGA"},
	)
	ewtSetupLiga, _ := models.GetEwtSetupFn(arrCond, "", false)
	ligaDecimalPoint := uint(2)
	if ewtSetupSec != nil {
		ligaDecimalPoint = uint(ewtSetupLiga.DecimalPoint)
	}

	arrCryptoStakingList := make([]CryptoStakingListStruct, 0)
	arrCryptoStakingList = append(arrCryptoStakingList,
		CryptoStakingListStruct{
			CryptoType:        helpers.TranslateV2("sec", arrData.LangCode, nil),
			StakeUnit:         float.Sub(totalSecSales.TotalAmount, totalSecSales.RefundAmount),
			StakeUnitDisplay:  helpers.CutOffDecimalv2(float.Sub(totalSecSales.TotalAmount, totalSecSales.RefundAmount), secDecimalPoint, ".", ",", true),
			StakePrice:        currentSecPrice,
			StakePriceDisplay: helpers.CutOffDecimalv2(currentSecPrice, secDecimalPoint, ".", ",", true),
		},
		CryptoStakingListStruct{
			CryptoType:        helpers.TranslateV2("liga", arrData.LangCode, nil),
			StakeUnit:         float.Sub(totalLigaSales.TotalAmount, totalLigaSales.RefundAmount),
			StakeUnitDisplay:  helpers.CutOffDecimalv2(float.Sub(totalLigaSales.TotalAmount, totalLigaSales.RefundAmount), ligaDecimalPoint, ".", ",", true),
			StakePrice:        currentLigaPrice,
			StakePriceDisplay: helpers.CutOffDecimalv2(currentLigaPrice, ligaDecimalPoint, ".", ",", true),
		},
	)

	arrDataReturn := MemberStakingMiningActionDetailsv1RstStruct{}
	arrDataReturn.MarketPrice.MarketPriceList = arrMarketPrice
	arrDataReturn.MarketPrice.CurrentMarketPrice = currentPrice
	arrDataReturn.MarketPrice.CryptoType = helpers.TranslateV2("sec", arrData.LangCode, nil)
	arrDataReturn.MarketPrice.CurrencyAmount = helpers.TranslateV2("hash", arrData.LangCode, nil)
	arrDataReturn.MarketPrice.Min = minPrice
	arrDataReturn.MarketPrice.Max = maxPrice
	arrDataReturn.MarketPrice.Interval = intervalPrice
	arrDataReturn.CryptoStakeList = arrCryptoStakingList
	arrDataReturn.ComputingPower = ""
	arrDataReturn.ComputingPowerCurrency = helpers.TranslateV2("hash", arrData.LangCode, nil)
	arrDataReturn.ComputingPowerValuation = ""
	arrDataReturn.ComputingPowerValuationCurrency = helpers.TranslateV2("hash", arrData.LangCode, nil)
	arrDataReturn.PurchaseStakingStatus = 1

	return &arrDataReturn, ""
}

// type MemberPoolMiningActionDetailsv1ReqStruct
type MemberPoolMiningActionDetailsv1ReqStruct struct {
	EntMemberID int
	LangCode    string
}

// type MemberPoolMiningActionDetailsv1RstStruct
type MemberPoolMiningActionDetailsv1RstStruct struct {
	MarketPrice             MarketPriceStruct `json:"market_price"`
	TodaySponsor            string            `json:"today_sponsor"`
	AccumulatedSalesAmount  string            `json:"accumulated_amount"`
	TodaySponsorSalesAmount string            `json:"today_sales_amount"`
	MatchingLevel           string            `json:"matching_level"`
	TotalSponsor            string            `json:"total_sponsor"`
	TotalSponsorSalesAmount string            `json:"total_sponsor_sales_amount"`
	PurchaseContractStatus  int               `json:"purchase_contract_status"`
	IncomeCap               struct {
		Total   string  `json:"total"`
		Balance string  `json:"balance"`
		Percent float64 `json:"percent"`
	} `json:"income_cap"`
	PoolAmount string `json:"pool_amount"`
}

// func GetMemberPoolMiningActionDetailsv1
func GetMemberPoolMiningActionDetailsv1(arrData MemberPoolMiningActionDetailsv1ReqStruct) *MemberPoolMiningActionDetailsv1RstStruct {

	// start get liga price
	var (
		decimalPoint       uint
		currentMarketPrice string
		minTP              float64
	)

	cryptoType := "SEC"
	totRecord := 7
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " created_at <= (SELECT created_at FROM exchange_price_movement_sec ORDER BY created_at DESC LIMIT 1)"},
	)
	arrCloseFn := models.ArrFnStruct{
		ArrCond: arrCond,
		Limit:   totRecord,
		OrderBy: " created_at DESC ",
	}
	arrMarketPriceRst, _ := models.GetExchangePriceMovementByTokenTypeFn(cryptoType, arrCloseFn, false)
	arrMinMaxMarketPriceRst, _ := models.GetMinMaxExchangePriceMovementByTokenTypeFn(cryptoType, arrCond, totRecord, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: cryptoType},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	decimalPoint = 2
	if ewtSetup != nil {
		decimalPoint = uint(ewtSetup.DecimalPoint)
	}
	marketPriceList := make([]MarketPriceListStruct, 0)
	// if len(arrMarketPriceRst) < totRecord {
	// balRec := totRecord - len(arrMarketPriceRst)
	for arrMarketPriceRstK, arrMarketPriceRstV := range arrMarketPriceRst {
		marketPrice1 := make([]MarketPriceListStruct, 1)
		marketPrice1[0].PriceRate = helpers.CutOffDecimal(arrMarketPriceRstV.TokenPrice, decimalPoint, ".", ",")
		marketPrice1[0].TransAt = base.TimeFormat(arrMarketPriceRstV.CreatedAt, "2006-01-02 15:04:05")
		marketPriceList = append(marketPrice1, marketPriceList...)
		if arrMarketPriceRstK == 0 { // take the first one as current latest price.
			currentMarketPrice = helpers.CutOffDecimal(arrMarketPriceRstV.TokenPrice, decimalPoint, ".", ",")
		}
	}
	// for i := 0; i < balRec; i++ {
	// 	// priceRate := []string{float.RoundDown(arrMarketPriceRst[0].TokenPrice, 2)}
	// 	arrFirstMarketPrice := make([]MarketPriceListStruct, 0)

	// 	arrFirstMarketPrice = append(arrFirstMarketPrice,
	// 		MarketPriceListStruct{
	// 			PriceRate: helpers.CutOffDecimal(arrMarketPriceRst[0].TokenPrice, decimalPoint, ".", ","),
	// 			TransAt:   base.TimeFormat(arrMarketPriceRst[0].CreatedAt, "2006-01-02 15:04:05"),
	// 		},
	// 	)
	// 	// priceRate := []string{helpers.CutOffDecimal(arrMarketPriceRst[0].TokenPrice, decimalPoint, ".", ",")}
	// 	marketPriceList = append(arrFirstMarketPrice, marketPriceList...)
	// }
	// }

	intervalPrice := float.RoundUp((arrMinMaxMarketPriceRst.MaxTokenPrice-arrMinMaxMarketPriceRst.MinTokenPrice)/float64(5), int(decimalPoint))

	if arrMinMaxMarketPriceRst.MinTokenPrice-intervalPrice > 0 {
		minTP = arrMinMaxMarketPriceRst.MinTokenPrice - intervalPrice
	}

	// start get income cap info
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewallet_type_code = ?", CondValue: "PCAP"},
	)
	pcapEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	if pcapEwtSetup == nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetEwtSetupFn_pcap_failed", "missing_pcap_ewt_setup", arrCond, true)
		return nil
	}

	ewtSummaryCond := make([]models.WhereCondFn, 0)
	ewtSummaryCond = append(ewtSummaryCond,
		models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: pcapEwtSetup.ID},
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: arrData.EntMemberID},
	)
	ewtSummaryRst, err := models.GetEwtSummarySetupFn(ewtSummaryCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetEwtSummarySetupFn_failed", err.Error(), ewtSummaryCond, true)
		return nil
	}

	incomeCapBalance := "0.00"
	totalIncomeCap := "0.00"
	incomeCapPercent := float64(0)
	if len(ewtSummaryRst) > 0 {
		incomeCapBalance = helpers.CutOffDecimal(ewtSummaryRst[0].Balance, uint(ewtSummaryRst[0].DecimalPoint), ".", ",")
		totalIncomeCap = helpers.CutOffDecimal(ewtSummaryRst[0].TotalIn, uint(ewtSummaryRst[0].DecimalPoint), ".", ",")
		incomeCapPercent = (ewtSummaryRst[0].TotalIn - ewtSummaryRst[0].Balance) / (ewtSummaryRst[0].TotalIn)
	}
	// end get income cap info

	// start get today direct sponsor
	todayDirectSponsor := "0"
	todayDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")
	arrTodayDirectSponsorCond := make([]models.WhereCondFn, 0)
	arrTodayDirectSponsorCond = append(arrTodayDirectSponsorCond,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "date(ent_member_tree_sponsor.created_at) = ?", CondValue: todayDate},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	todayDirectSponsorRst, err := models.GetTotalDirectSponsorFn(arrTodayDirectSponsorCond, false)

	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetTotalDirectSponsorFn_failed", err.Error(), arrTodayDirectSponsorCond, true)
		return nil
	}

	if todayDirectSponsorRst.TotalDirectSponsor > 0 {
		todayDirectSponsor = helpers.CutOffDecimal(todayDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
	}
	// end get today direct sponsor

	// start total direct sponsor
	totalDirectSponsor := "0"
	arrTotalDirectSponsorCond := make([]models.WhereCondFn, 0)
	arrTotalDirectSponsorCond = append(arrTotalDirectSponsorCond,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	totalDirectSponsorRst, err := models.GetTotalDirectSponsorFn(arrTotalDirectSponsorCond, false)

	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetTotalDirectSponsorFn_failed", err.Error(), arrTotalDirectSponsorCond, true)
		return nil
	}

	if totalDirectSponsorRst.TotalDirectSponsor > 0 {
		totalDirectSponsor = helpers.CutOffDecimal(totalDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
	}
	// end total direct sponsor

	// start today direct sponsor sales amount
	todaySponsorSalesAmount := "0.00"
	todaySalesCond := make([]models.WhereCondFn, 0)
	todaySalesCond = append(todaySalesCond,
		models.WhereCondFn{Condition: "sls_master.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "P2P"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "date(sls_master.created_at) = ?", CondValue: todayDate},
	)
	todaySales, err := models.GetTotalSalesAmount(todaySalesCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetTotalSalesAmount_1", err.Error(), todaySalesCond, true)
		return nil
	}
	todaySponsorSalesAmount = helpers.CutOffDecimal(todaySales.TotalAmount, uint(todaySales.DecimalPoint), ".", ",")
	// end today direct sponsor sales amount

	// start total direct sponsor sales amount
	totalSponsorSalesAmount := "0.00"
	totalSalesCond := make([]models.WhereCondFn, 0)
	totalSalesCond = append(totalSalesCond,
		models.WhereCondFn{Condition: "sls_master.sponsor_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "P2P"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	totalSales, err := models.GetTotalSalesAmount(totalSalesCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetTotalSalesAmount_2", err.Error(), totalSalesCond, true)
		return nil
	}
	totalSponsorSalesAmount = helpers.CutOffDecimal(totalSales.TotalAmount, uint(totalSales.DecimalPoint), ".", ",")
	// end total direct sponsor sales amount

	// start set matching level
	matchingLevel := totalDirectSponsor
	if totalDirectSponsorRst.TotalDirectSponsor < 15 && totalSales.TotalAmount > 1000 {
		matchingLevel = "15"
	}
	// end set matching level

	// start get accumulated amount
	accumulatedSalesAmount := "0.00"
	accAmtCond := make([]models.WhereCondFn, 0)
	accAmtCond = append(accAmtCond,
		models.WhereCondFn{Condition: "t_member_id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02")},
	)
	accAmtRst, err := models.GetTblP2PBonusSponsorFn(accAmtCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetTblP2PBonusSponsorFn_failed", err.Error(), accAmtCond, true)
		return nil
	}

	accAmt := 0.00
	accDay := 0
	if len(accAmtRst) > 0 {
		accAmt = float64(accAmtRst[0].NShare)
		accDay = accAmtRst[0].NDay
	}
	accumulatedSalesAmount = helpers.CutOffDecimal(accAmt, 2, ".", ",") + " (" + strconv.Itoa(accDay) + helpers.TranslateV2("days", arrData.LangCode, nil) + ")"
	// end get accumulated amount

	// start get sponsor pool amount
	sponsorPoolAmount := "0.00"
	poolAmtCond := make([]models.WhereCondFn, 0)
	poolAmtCond = append(poolAmtCond,
		models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02")},
	)
	poolAmtRst, err := models.GetTblP2PBonusSponsorPoolFn(poolAmtCond, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetTblP2PBonusSponsorPoolFn_failed", err.Error(), poolAmtCond, true)
		return nil
	}

	poolAmt := 0.00
	if len(poolAmtRst) > 0 {
		poolAmt = float64(poolAmtRst[0].PoolCf)
	}

	todaySalesBvCond := make([]models.WhereCondFn, 0)
	todaySalesBvCond = append(todaySalesBvCond,
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "P2P"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "date(sls_master.doc_date) = ?", CondValue: todayDate},
	)
	todaySalesBv, err := models.GetTotalSalesAmount(todaySalesBvCond, false)
	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetTotalSalesAmount_todaySalesBv_failed", err.Error(), todaySalesBvCond, true)
		return nil
	}

	//get sponsor pool markup
	sponsorMarkupCond := make([]models.WhereCondFn, 0)
	sponsorMarkupCond = append(sponsorMarkupCond,
		models.WhereCondFn{Condition: "bns_date = ?", CondValue: todayDate},
	)
	sponsorMarkup, err := models.GetSysSponsorPoolMarkupFn(sponsorMarkupCond, "SUM(pool_amount) as pool_amount", false)

	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetSysSponsorPoolMarkupFn_failed", err.Error(), sponsorMarkupCond, true)
		return nil
	}

	markupAmt := 0.00
	if len(sponsorMarkup) > 0 {
		markupAmt = sponsorMarkup[0].PoolAmount
	}

	// sponsor pool amount = yesterday pool cf + (today's sales bv x 10%) + markup
	sponsorPoolAmount = helpers.CutOffDecimal(float.Add(float.Add(poolAmt, (todaySalesBv.TotalAmount*0.1)), markupAmt), 2, ".", ",")
	// end get sponsor pool amount

	arrDataReturn := MemberPoolMiningActionDetailsv1RstStruct{}
	arrDataReturn.MarketPrice.MarketPriceList = marketPriceList
	arrDataReturn.MarketPrice.CurrentMarketPrice = currentMarketPrice
	arrDataReturn.MarketPrice.CryptoType = helpers.TranslateV2(cryptoType, arrData.LangCode, nil)
	arrDataReturn.MarketPrice.CurrencyAmount = helpers.TranslateV2("price", arrData.LangCode, nil)
	arrDataReturn.MarketPrice.Min = minTP
	arrDataReturn.MarketPrice.Max = arrMinMaxMarketPriceRst.MaxTokenPrice + intervalPrice
	arrDataReturn.MarketPrice.Interval = intervalPrice
	arrDataReturn.TodaySponsor = todayDirectSponsor
	arrDataReturn.MatchingLevel = matchingLevel
	arrDataReturn.TotalSponsor = totalDirectSponsor
	arrDataReturn.TodaySponsorSalesAmount = todaySponsorSalesAmount
	arrDataReturn.TotalSponsorSalesAmount = totalSponsorSalesAmount
	arrDataReturn.AccumulatedSalesAmount = accumulatedSalesAmount
	arrDataReturn.PoolAmount = sponsorPoolAmount
	arrDataReturn.PurchaseContractStatus = 1
	arrDataReturn.IncomeCap.Balance = incomeCapBalance
	arrDataReturn.IncomeCap.Total = totalIncomeCap
	arrDataReturn.IncomeCap.Percent = incomeCapPercent

	return &arrDataReturn
}

// type MemberMiningMiningActionDetailsv1ReqStruct
type MemberMiningMiningActionDetailsv1ReqStruct struct {
	EntMemberID int
	LangCode    string
}

// type MemberMiningMiningActionDetailsv1RstStruct
type MemberMiningMiningActionDetailsv1RstStruct struct {
	Ranking                 interface{}                        `json:"ranking"`
	PurchaseContractStatus  int                                `json:"purchase_contract_status"`
	PurchaseContractType    []MemberMiningPurchaseContractType `json:"purchase_contract_type"`
	MiningCoinListing       []MiningCoinListingStruct          `json:"mining_coin_listing"`
	MiningCoinDetailListing []MiningCoinDetailListingStruct    `json:"mining_coin_detail_listing"`
	RankingCriteria         interface{}                        `json:"ranking_criteria"`
}

type MemberMiningPurchaseContractType struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type MiningCoinListingStruct struct {
	MiningCoinCode string `json:"mining_coin_code"`
	MiningCoinName string `json:"mining_coin_name"`
	MiningCoinDesc string `json:"mining_coin_desc"`
}

type MiningCoinDetailListingStruct struct {
	MiningCoinCode     string                 `json:"mining_coin_code"`
	MiningCoinName     string                 `json:"mining_coin_name"`
	MiningCoinDesc     []string               `json:"mining_coin_desc"`
	ClickableDesc      string                 `json:"clickable_desc"`
	ClickableStatus    int                    `json:"clickable_status"`
	MiningCoinSubTitle string                 `json:"mining_coin_sub_title"`
	MiningCoinSubDesc  []string               `json:"mining_coin_sub_desc"`
	SubClickableDesc   string                 `json:"sub_clickable_desc"`
	SubClickableStatus int                    `json:"sub_clickable_status"`
	BroadbandStatus    int                    `json:"broadband_status"`
	Details            map[string]interface{} `json:"details"`
}

// func GetMemberMiningMiningActionDetailsv1
func GetMemberMiningMiningActionDetailsv1(arrData MemberMiningMiningActionDetailsv1ReqStruct) (*MemberMiningMiningActionDetailsv1RstStruct, error) {
	var (
		memberID                   int    = arrData.EntMemberID
		prdGroup                   string = "MINING"
		sysGeneralSetupID          string = "mining_coin_setting"
		arrMiningCoinListing              = []MiningCoinListingStruct{}
		arrMiningCoinDetailListing        = []MiningCoinDetailListingStruct{}
		miningRankingBoard         interface{}
		rankingCriteria            interface{}
	)

	// start to determine which ranking info to display
	setupRst, err := models.GetSysGeneralSetupByID("mining_ranking_display_setting")
	if err != nil {
		base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetSysGeneralSetupByID_failed_missing_ranking_display", err.Error(), map[string]interface{}{"settingID": sysGeneralSetupID}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
	}
	// end to determine which ranking info to display

	// start display mining ranking info
	if setupRst.InputType1 == "1" {
		// get ranking info
		miningRankingBoardRst, errMsg := member_service.GetMemberMiningRankingInfo(memberID)
		if errMsg != "" {
			base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetMemberMiningRankingInfo():1", errMsg, map[string]interface{}{"memberID": memberID}, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: errMsg}
		}
		miningRankingBoard = miningRankingBoardRst
	}
	// end display mining ranking info

	// start display mining bzz ranking info
	if setupRst.InputValue1 == "1" {
		// get ranking info
		bzzMiningRankingBoardRst, errMsg := member_service.GetMemberBZZMiningRankingInfo(memberID, arrData.LangCode)
		if errMsg != nil {
			base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetMemberMiningRankingInfo():1", errMsg.Error(), map[string]interface{}{"memberID": memberID}, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: errMsg.Error()}
		}
		rankingCriteria = bzzMiningRankingBoardRst
	}
	// end display mining bzz ranking info

	setupRst, err = models.GetSysGeneralSetupByID(sysGeneralSetupID)
	if err != nil {
		base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetSysGeneralSetupByID():1", err.Error(), map[string]interface{}{"settingID": sysGeneralSetupID}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
	}

	type ArrMiningCoinTypeStruct struct {
		Code            string `json:"code"`
		ClickableStatus int    `json:"clickable_status"`
		BroadbandStatus int    `json:"broadband_status"`
		ActionStatus    int    `json:"action_status"`
	}

	var arrMiningCoinType []ArrMiningCoinTypeStruct
	json.Unmarshal([]byte(setupRst.InputValue1), &arrMiningCoinType)

	// get prd_group_type setting
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: prdGroup},
		models.WhereCondFn{Condition: "prd_group_type.status = ?", CondValue: "A"},
	)
	arrGetPrdGroupType, err := models.GetPrdGroupTypeFn(arrCond, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetPrdGroupTypeFn()", err.Error(), map[string]interface{}{"prdGroup": prdGroup, "arrCond": arrCond}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
	}
	if len(arrGetPrdGroupType) <= 0 {
		base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetPrdGroupTypeFn()", "prd_group_not_found", map[string]interface{}{"prdGroup": prdGroup, "arrCond": arrCond}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
	}

	// get unit prices range of crypto - mining
	cryptoMiningDetails, errMsg := GetCryptoMiningDetails(memberID, 0, "MINING")
	if errMsg != "" {
		base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetCryptoMiningDetails():1", errMsg, map[string]interface{}{"memberID": memberID, "prdMasterID": 0, "prdGroupType": "MINING"}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
	}

	// get unit prices range of crypto - mining-bzz
	cryptoMiningDetailsBzz, errMsg := GetCryptoMiningDetails(memberID, 0, "MINING_BZZ")
	if errMsg != "" {
		base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetCryptoMiningDetails():2", errMsg, map[string]interface{}{"memberID": memberID, "prdMasterID": 0, "prdGroupType": "MINING_BZZ"}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: err}
	}

	if len(arrMiningCoinType) > 0 {
		for _, arrMiningCoinTypeV := range arrMiningCoinType {
			var (
				miningCoinCode     = arrMiningCoinTypeV.Code
				miningCoinName     string
				ewtCurrencyCode    string
				decimalPoint       int
				currencyCode       string = "TiB"
				miningCurrencyCode string = "TiB"
				miningCoinDesc     string
			)

			// get ewt_setup
			ewtSetupCond := make([]models.WhereCondFn, 0)
			ewtSetupCond = append(ewtSetupCond,
				models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: miningCoinCode},
			)
			ewtSetup, err := models.GetEwtSetupFn(ewtSetupCond, "", false)
			if err != nil {
				base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetEwtSetupFn():1", err.Error(), map[string]interface{}{"cond": ewtSetupCond}, true)
				continue
			}

			if miningCoinCode == "FIL" {
				miningCoinName = helpers.TranslateV2("IPFS", arrData.LangCode, nil)
			} else {
				miningCoinName = helpers.TranslateV2(ewtSetup.EwtTypeName, arrData.LangCode, nil)
			}

			if miningCoinCode == "BZZ" {
				currencyCode = helpers.TranslateV2("nodes", arrData.LangCode, nil)
				miningCurrencyCode = helpers.TranslateV2("nodes", arrData.LangCode, nil)
			}

			ewtCurrencyCode = ewtSetup.CurrencyCode
			decimalPoint = ewtSetup.DecimalPoint

			var (
				highestPriceRate float64 = 1
				lowestPriceRate  float64 = 1
			)

			if miningCoinCode == "FIL" {
				highestPriceRate = cryptoMiningDetails[0].FilecoinPrice * cryptoMiningDetails[0].FilPrice
				lowestPriceRate = cryptoMiningDetails[len(cryptoMiningDetails)-1].FilecoinPrice * cryptoMiningDetails[len(cryptoMiningDetails)-1].FilPrice
			} else if miningCoinCode == "SEC" {
				highestPriceRate = cryptoMiningDetails[0].SecPrice
				lowestPriceRate = cryptoMiningDetails[len(cryptoMiningDetails)-1].SecPrice
			} else if miningCoinCode == "XCH" {
				highestPriceRate = cryptoMiningDetails[0].XchPrice
				lowestPriceRate = cryptoMiningDetails[len(cryptoMiningDetails)-1].XchPrice
			} else if miningCoinCode == "BZZ" {
				highestPriceRate = cryptoMiningDetailsBzz[0].BzzPrice
				lowestPriceRate = cryptoMiningDetailsBzz[len(cryptoMiningDetailsBzz)-1].BzzPrice
			}

			miningCoinDesc = fmt.Sprintf("1%s = %sU - %sU", miningCurrencyCode, helpers.CutOffDecimal(lowestPriceRate, uint(0), ".", ","), helpers.CutOffDecimal(highestPriceRate, uint(0), ".", ","))

			var display bool

			if display {
				// append into arrMiningCoinListing
				arrMiningCoinListing = append(arrMiningCoinListing, MiningCoinListingStruct{
					MiningCoinCode: miningCoinCode,
					MiningCoinName: miningCoinName,
					MiningCoinDesc: miningCoinDesc,
				})
			}

			// process for arrMiningCoinDetailListing
			var (
				arrMiningCoinDesc = []string{}
				// arrMiningCoinSubDesc               = []string{}
				totalOwnedMiningAmountTitle string = "capacity"
				totalMiningAmount           float64
				totalMiningRewardAmount     float64
			)

			// get mining sales doc
			arrSlsMasterFn := make([]models.WhereCondFn, 0)
			arrSlsMasterFn = append(arrSlsMasterFn,
				models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memberID},
				models.WhereCondFn{Condition: "sls_master.action IN (?,'MINING_BZZ') ", CondValue: "MINING"},
				models.WhereCondFn{Condition: "sls_master.status = ? ", CondValue: "AP"},
			)
			if miningCoinCode == "SEC" {
				arrSlsMasterFn = append(arrSlsMasterFn,
					models.WhereCondFn{Condition: "sls_master.machine_type IN(?,'FIL','XCH','BZZ') ", CondValue: miningCoinCode},
				)
			} else if miningCoinCode == "BZZ" {
				arrSlsMasterFn = append(arrSlsMasterFn,
					models.WhereCondFn{Condition: "sls_master.machine_type IN(?,'BZZ_100') ", CondValue: miningCoinCode},
				)
			} else {
				arrSlsMasterFn = append(arrSlsMasterFn,
					models.WhereCondFn{Condition: "sls_master.machine_type = ? ", CondValue: miningCoinCode},
				)
			}
			arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
			if err != nil {
				base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetSlsMasterFn():1", err.Error(), map[string]interface{}{"cond": arrSlsMasterFn}, true)
				continue
			}

			// calculate total mining amount
			if len(arrSlsMaster) > 0 {
				for _, arrSlsMasterV := range arrSlsMaster {
					var miningPrice float64 = 0

					// get mining sales doc
					arrSlsMasterMiningFn := make([]models.WhereCondFn, 0)
					arrSlsMasterMiningFn = append(arrSlsMasterMiningFn,
						models.WhereCondFn{Condition: "sls_master_mining.sls_master_id = ? ", CondValue: arrSlsMasterV.ID},
					)
					arrSlsMasterMining, err := models.GetSlsMasterMiningFn(arrSlsMasterMiningFn, "", false)
					if err != nil {
						base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetSlsMasterMiningFn():1", err.Error(), map[string]interface{}{"cond": arrSlsMasterMiningFn}, true)
						continue
					}

					if len(arrSlsMasterMining) > 0 {
						if miningCoinCode == "FIL" {
							miningPrice = arrSlsMasterMining[0].FilTib
						} else if miningCoinCode == "SEC" {
							miningPrice = arrSlsMasterMining[0].SecTib
						} else if miningCoinCode == "XCH" {
							miningPrice = arrSlsMasterMining[0].XchTib
						} else if miningCoinCode == "BZZ" {
							miningPrice = arrSlsMasterMining[0].BzzTib
						}
					}

					miningPrice, err = helpers.ValueToFloat(helpers.CutOffDecimal(miningPrice, 4, ".", ""))
					if err != nil {
						base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-ValueToFloat():1", err.Error(), map[string]interface{}{"value": helpers.CutOffDecimal(miningPrice, 4, ".", "")}, true)
						continue
					}

					if miningPrice > 0 {
						totalMiningAmount = totalMiningAmount + miningPrice
					}
				}
			}

			// calculate total mining reward amount
			if miningCoinCode == "SEC" {
				totalBonusMiningRebate, err := models.GetTotalBonusMiningRebate(memberID)
				if err != nil {
					base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetTotalBonusMiningRebate():1", err.Error(), map[string]interface{}{"memberID": memberID}, true)
					continue
				}

				if totalBonusMiningRebate != nil {
					totalMiningRewardAmount = totalBonusMiningRebate.TotFBns
				}
			} else {
				totalBonusMiningRebateCrypto, err := models.GetTotalBonusMiningRebateCrypto(memberID, miningCoinCode)
				if err != nil {
					base.LogErrorLog("GetMemberMiningMiningActionDetailsv1-GetTotalBonusMiningRebateCrypto():1", err.Error(), map[string]interface{}{"memberID": memberID, "crypto_type": miningCoinCode}, true)
					continue
				}

				if totalBonusMiningRebateCrypto != nil {
					totalMiningRewardAmount = totalBonusMiningRebateCrypto.TotFBns
				}
			}

			if miningCoinCode == "FIL" {
				totalOwnedMiningAmountTitle = "power"
			}

			arrMiningCoinDesc = append(arrMiningCoinDesc, fmt.Sprintf("%s = %s %s/%s", helpers.TranslateV2(totalOwnedMiningAmountTitle, arrData.LangCode, nil), helpers.CutOffDecimal(totalMiningAmount, 4, ".", ","), ewtCurrencyCode, currencyCode)) // own mining amount
			arrMiningCoinDesc = append(arrMiningCoinDesc, fmt.Sprintf("%s = %s %s", ewtCurrencyCode, helpers.CutOffDecimal(totalMiningRewardAmount, uint(decimalPoint), ".", ","), ewtCurrencyCode))                                                   // own mining reward amount

			// append into arrMiningCoinDetailListing
			var miningCoinDetailListing = MiningCoinDetailListingStruct{
				MiningCoinCode:     miningCoinCode,
				MiningCoinName:     miningCoinName,
				MiningCoinDesc:     arrMiningCoinDesc,
				ClickableStatus:    arrMiningCoinTypeV.ClickableStatus,
				SubClickableStatus: arrMiningCoinTypeV.ActionStatus,
				BroadbandStatus:    arrMiningCoinTypeV.BroadbandStatus,
			}

			if arrMiningCoinTypeV.ClickableStatus == 1 {
				miningCoinDetailListing.ClickableDesc = helpers.TranslateV2("waiting_to_be_released", arrData.LangCode, nil)
			}

			if arrMiningCoinTypeV.ActionStatus == 1 {
				miningCoinDetailListing.SubClickableDesc = helpers.TranslateV2("withholding_sec", arrData.LangCode, nil)
			}

			arrMiningCoinDetailListing = append(arrMiningCoinDetailListing, miningCoinDetailListing)
		}
	}

	var arrMemberMiningPurchaseContractType = []MemberMiningPurchaseContractType{}
	arrMemberMiningPurchaseContractType = append(arrMemberMiningPurchaseContractType, MemberMiningPurchaseContractType{
		Code: "MINING_BZZ",
		Name: helpers.TranslateV2("bzz_node", arrData.LangCode, nil),
	})

	var arrDataReturn = MemberMiningMiningActionDetailsv1RstStruct{
		Ranking:                 miningRankingBoard,
		PurchaseContractStatus:  1,
		MiningCoinListing:       arrMiningCoinListing,
		MiningCoinDetailListing: arrMiningCoinDetailListing,
		PurchaseContractType:    arrMemberMiningPurchaseContractType,
		RankingCriteria:         rankingCriteria,
	}

	return &arrDataReturn, nil
}

// type MemberMiningMiningActionListv1ReqStruct
type MemberMiningMiningActionListv1ReqStruct struct {
	EntMemberID    int
	Page           int
	MiningCoinCode string
	LangCode       string
}

// MiningActionListv1Struct struct
type MiningActionListv1Struct struct {
	Amount string `json:"amount"`
	Date   string `json:"date"`
}

// func GetMemberMiningMiningActionListv1
func GetMemberMiningMiningActionListv1(arrData MemberMiningMiningActionListv1ReqStruct) interface{} {
	var (
		memberID           = arrData.EntMemberID
		cryptoType         = arrData.MiningCoinCode
		miningActionListv1 = []interface{}{}
	)

	arrBonusMiningLock, _ := models.GetBonusMiningLock(memberID, cryptoType)

	if len(arrBonusMiningLock) > 0 {
		for _, arrBonusMiningLockV := range arrBonusMiningLock {
			miningActionListv1 = append(miningActionListv1, map[string]interface{}{
				"date":   arrBonusMiningLockV.BnsID.Format("2006-01-02"),
				"amount": helpers.CutOffDecimal(arrBonusMiningLockV.FBns, uint(8), ".", ","),
			})
		}
	}

	page := base.Pagination{
		Page:    int64(arrData.Page),
		DataArr: miningActionListv1,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return &arrDataReturn
}

type CryptoMiningDetails struct {
	PrdMasterID             int
	FilPrice, FilecoinPrice float64
	SecPrice                float64
	XchPrice                float64
	BzzPrice                float64
}

// GetCryptoMiningDetails - prdMasterID default = 0
func GetCryptoMiningDetails(memberID, prdMasterID int, prdGroupType string) ([]CryptoMiningDetails, string) {
	var (
		cryptoMiningDetails        = []CryptoMiningDetails{}
		prdGroup            string = "MINING"
		curDate             string = base.GetCurrentDateTimeT().Format("2006-01-02")
		filMiningCode       string = "FIL"
		filPrice            float64
	)

	// get all mining prd
	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.prd_group IN(?,'MINING_BZZ') ", CondValue: prdGroup},
		models.WhereCondFn{Condition: " date(prd_master.date_start) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " date(prd_master.date_end) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " prd_master.status = ? ", CondValue: "A"},
	)
	if prdMasterID != 0 {
		arrPrdMasterFn = append(arrPrdMasterFn,
			models.WhereCondFn{Condition: " prd_master.id = ? ", CondValue: prdMasterID},
		)
	}
	if prdGroupType != "" {
		arrPrdMasterFn = append(arrPrdMasterFn,
			models.WhereCondFn{Condition: " prd_master.prd_group = ? ", CondValue: prdGroupType},
		)
	}

	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:GetCryptoMiningDetails():GetPrdMasterFn():1", map[string]interface{}{"condition": arrPrdMasterFn}, err.Error(), true)
		return nil, "something_went_wrong"
	}

	if len(arrPrdMaster) > 0 {
		filPrice, err = base.GetLatestPriceMovementByTokenType(filMiningCode)
		if err != nil {
			base.LogErrorLog("productService:GetCryptoMiningDetails():GetLatestPriceMovementByTokenType():1", err.Error(), map[string]interface{}{"token_type": filMiningCode}, true)
		}

		filPrice = float64(int(math.Round(filPrice))) // need to round up to whole number

		for _, arrPrdMasterV := range arrPrdMaster {
			prdMiningPrice, err := models.GetLatestPrdMiningPriceByPrdMasterID(arrPrdMasterV.ID, curDate)
			if err != nil {
				base.LogErrorLog("productService:GetCryptoMiningDetails():GetLatestPrdMiningPriceByPrdMasterID():1", map[string]interface{}{"prdMasterID": arrPrdMasterV.ID, "curDate": curDate}, err.Error(), true)
				return nil, "something_went_wrong"
			}

			if prdMiningPrice != nil {
				cryptoMiningDetails = append(cryptoMiningDetails, CryptoMiningDetails{
					FilPrice:      filPrice,
					FilecoinPrice: prdMiningPrice.FilPrice,
					SecPrice:      prdMiningPrice.SecPrice,
					XchPrice:      prdMiningPrice.XchPrice,
					BzzPrice:      prdMiningPrice.BzzPrice,
				})
			}
		}
	}

	return cryptoMiningDetails, ""
}

type DrawingsStruct struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func GetNftDrawingByMemberID(memberID int, seriesCode, langCode string) (DrawingsStruct, error) {
	var (
		title     string
		url       string
		arrReturn DrawingsStruct
	)

	//check this member whether assign before
	arrEntMemberNftImgFn := make([]models.WhereCondFn, 0)
	arrEntMemberNftImgFn = append(arrEntMemberNftImgFn,
		models.WhereCondFn{Condition: "ent_member_nft_img.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member_nft_img.type = ?", CondValue: seriesCode},
		models.WhereCondFn{Condition: "nft_img.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "FIND_IN_SET(?, nft_img.type) > 0", CondValue: seriesCode},
	)
	arrEntMemberNftImg, err := models.GetEwtMemberNftImgFn(arrEntMemberNftImgFn, false)
	if err != nil {
		base.LogErrorLog("GetNftDrawingByMemberID:GetEwtMemberNftImgFn()", arrEntMemberNftImgFn, err, true)
		return DrawingsStruct{}, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if len(arrEntMemberNftImg) < 1 {
		// find random image
		arrNftImgFn := make([]models.WhereCondFn, 0)
		arrNftImgFn = append(arrNftImgFn,
			models.WhereCondFn{Condition: "nft_img.status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "FIND_IN_SET(?, nft_img.type) > 0", CondValue: seriesCode},
		)
		arrNftImg, _ := models.GetRandomNftImgFn(arrNftImgFn, "", false)

		if len(arrNftImg) > 0 {
			title = arrNftImg[0].Title
			url = arrNftImg[0].ImgLink

			// save to ent_member_nft_img
			arrStore := models.EntMemberNftImg{
				Type:      seriesCode,
				MemberID:  memberID,
				ImgID:     arrNftImg[0].ID,
				CreatedAt: time.Now(),
			}

			models.AddEntMemberNftImg(arrStore)
		}
	} else {
		title = arrEntMemberNftImg[0].Title
		url = arrEntMemberNftImg[0].ImgLink
	}

	arrReturn = DrawingsStruct{
		Title: title,
		Url:   url,
	}

	return arrReturn, nil
}

// func GetNftSeries
func GetNftSeries(memberID int, langCode string) ([]interface{}, string) {
	var (
		arrReturnData = []interface{}{}
		curTime       = base.GetCurrentTime("2006-01-02 15:04:05")
	)

	// get nft_series_group_setup
	nftSeriesGroupSetupFn := make([]models.WhereCondFn, 0)
	nftSeriesGroupSetupFn = append(nftSeriesGroupSetupFn,
		models.WhereCondFn{Condition: "nft_series_group_setup.status = ?", CondValue: "A"},
	)
	nftSeriesGroupSetup, err := models.GetNftSeriesGroupSetupFn(nftSeriesGroupSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("GetNftSeries:GetNftSeriesGroupSetupFn()", map[string]interface{}{"condition": nftSeriesGroupSetupFn}, err.Error(), true)
		return nil, "something_went_wrong"
	}

	if len(nftSeriesGroupSetup) > 0 {
		for _, nftSeriesGroupSetupV := range nftSeriesGroupSetup {
			// get nft_series_setup setting
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "nft_series_setup.group_type = ?", CondValue: nftSeriesGroupSetupV.ID},
				models.WhereCondFn{Condition: "nft_series_setup.status = ?", CondValue: "A"},
				models.WhereCondFn{Condition: "nft_series_setup.purchase = ?", CondValue: 1},
				models.WhereCondFn{Condition: "nft_series_setup.start_date <= ?", CondValue: curTime},
				models.WhereCondFn{Condition: "nft_series_setup.end_date >= ?", CondValue: curTime},
			)
			arrNftSeriesSetup, err := models.GetNftSeriesSetupFn(arrCond, "", false)
			if err != nil {
				base.LogErrorLog("GetNftSeries:GetNftSeriesSetupFn()", map[string]interface{}{"condition": arrCond}, err.Error(), true)
				return nil, "something_went_wrong"
			}

			// process data
			if len(arrNftSeriesSetup) > 0 {
				var (
					seriesCode                 = arrNftSeriesSetup[0].Code
					purchasedAmount    float64 = 0
					distributedSupply  float64 = 0
					imgPath                    = ""
					arrReturnDataValue         = map[string]interface{}{}
				)

				// get purchased amount
				purchasedAmount, _ = GetCurrentPurchasedNftAmount(seriesCode)

				// get distributed supply
				distributedSupply, _ = GetCurrentDistributedNftSupply(seriesCode)

				// get drawings art
				drawingsRst, drawErr := GetNftDrawingByMemberID(memberID, seriesCode, langCode)
				if drawErr == nil {
					imgPath = drawingsRst.Url
				}

				arrReturnDataValue["code"] = seriesCode
				arrReturnDataValue["name"] = arrNftSeriesSetup[0].Name
				// arrReturnDataValue["name"] = helpers.TranslateV2(arrNftSeriesSetup[0].Name, langCode, nil)
				arrReturnDataValue["drawing_url"] = imgPath
				arrReturnDataValue["series_name"] = nftSeriesGroupSetupV.Name
				arrReturnDataValue["series_description"] = nftSeriesGroupSetupV.Description
				arrReturnDataValue["total_purchased"] = helpers.CutOffDecimalv2(purchasedAmount, 0, ".", ",", true)
				arrReturnDataValue["total_purchase_limit"] = helpers.CutOffDecimalv2(arrNftSeriesSetup[0].PurchaseLimit, 0, ".", ",", true)
				arrReturnDataValue["distributed_supply"] = helpers.CutOffDecimalv2(distributedSupply, 0, ".", ",", true)
				arrReturnDataValue["total_supply"] = helpers.CutOffDecimalv2(arrNftSeriesSetup[0].Supply, 0, ".", ",", true)
				arrReturnDataValue["start_date"] = arrNftSeriesSetup[0].StartDate.Format("2006-01-02 15:04:05")
				arrReturnDataValue["end_date"] = arrNftSeriesSetup[0].EndDate.Format("2006-01-02 15:04:05")
				arrReturnDataValue["price"] = arrNftSeriesSetup[0].Price

				arrReturnData = append(arrReturnData, arrReturnDataValue)
			}
		}
	}

	return arrReturnData, ""
}

// GetCurrentPurchasedNftAmount func
func GetCurrentPurchasedNftAmount(nftSeriesCode string) (float64, string) {
	var (
		purchasedNft            float64 = 0
		adjustedPurchasedAmount float64 = 0
		finalPurchasedAmount    float64 = 0
	)
	// total purchased by nft series type
	totalSalesFn := make([]models.WhereCondFn, 0)
	totalSalesFn = append(totalSalesFn,
		models.WhereCondFn{Condition: "sls_master.action = ?", CondValue: "NFT"},
		models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "sls_master.nft_series_code = ?", CondValue: nftSeriesCode},
	)
	totalSales, err := models.GetTotalSalesAmount(totalSalesFn, false)
	if err != nil {
		base.LogErrorLog("member_service:GetCurrentPurchasedNftAmount()", "GetTotalSalesAmount()", err.Error(), true)
		return purchasedNft, "something_went_wrong"
	}

	purchasedNft = totalSales.TotalBv

	// add admin adjust
	totalNftSeriesSupplySetup, err := models.GetTotalNftSeriesPurchaseLimitSetup(nftSeriesCode)
	if err != nil {
		base.LogErrorLog("member_service:GetCurrentPurchasedNftAmount()", "GetTotalNftSeriesPurchaseLimitSetup()", err.Error(), true)
		return 0, "something_went_wrong"
	}

	adjustedPurchasedAmount = float64(totalNftSeriesSupplySetup.TotalValue)

	finalPurchasedAmount = purchasedNft + adjustedPurchasedAmount

	// fmt.Println("nftSeriesCode:", nftSeriesCode, "purchasedNft:", purchasedNft, "adjustedPurchasedAmount:", adjustedPurchasedAmount, "finalPurchasedAmount:", finalPurchasedAmount)
	// set to 0 if supply fall to negative
	if finalPurchasedAmount < 0 {
		finalPurchasedAmount = 0
	}

	return finalPurchasedAmount, ""
}

// GetCurrentDistributedNftSupply func
func GetCurrentDistributedNftSupply(nftSeriesCode string) (float64, string) {
	var (
		distributedSupply         float64 = 0
		adjustedDistributedSupply float64 = 0
		reissueSupply             float64 = 0
		finalSupply               float64 = 0
	)

	// get nft_series_setup
	nftSeriesSetupFn := make([]models.WhereCondFn, 0)
	nftSeriesSetupFn = append(nftSeriesSetupFn,
		models.WhereCondFn{Condition: "nft_series_setup.code = ?", CondValue: nftSeriesCode},
		models.WhereCondFn{Condition: "nft_series_setup.status = ?", CondValue: "A"},
	)
	nftSeriesSetup, err := models.GetNftSeriesSetupFn(nftSeriesSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("member_service:GetCurrentDistributedNftSupply():GetNftSeriesSetupFn()", map[string]interface{}{"condition": nftSeriesSetupFn}, err.Error(), true)
		return 0, "something_went_wrong"
	}
	if len(nftSeriesSetup) <= 0 {
		base.LogErrorLog("member_service:GetCurrentDistributedNftSupply():GetNftSeriesSetupFn()", map[string]interface{}{"condition": nftSeriesSetupFn}, "nft_series_not_found", true)
		return 0, "something_went_wrong"
	}

	//total distributed airdrop amount by nft series type
	// totalSalesFn := make([]models.WhereCondFn, 0)
	// totalSalesFn = append(totalSalesFn,
	// 	models.WhereCondFn{Condition: "sls_master.action = ?", CondValue: "NFT"},
	// 	models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
	// 	models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	// 	models.WhereCondFn{Condition: "sls_master.airdrop_ewallet_type_id = ?", CondValue: nftSeriesSetup[0].EwalletTypeID},
	// )
	// totalSales, err := models.GetTotalSalesAmount(totalSalesFn, false)
	// if err != nil {
	// 	base.LogErrorLog("member_service:GetCurrentDistributedNftSupply()", "GetTotalSalesAmount()", err.Error(), true)
	// 	return distributedSupply, "something_went_wrong"
	// }

	// distributedSupply = totalSales.TotalAirdropNft

	// add admin adjust
	totalNftSeriesSupplySetup, err := models.GetTotalNftSeriesSupplySetup(nftSeriesCode)
	if err != nil {
		base.LogErrorLog("member_service:GetCurrentDistributedNftSupply()", "GetTotalNftSeriesSupplySetup()", err.Error(), true)
		return 0, "something_went_wrong"
	}

	adjustedDistributedSupply = float64(totalNftSeriesSupplySetup.TotalValue)

	// add admin payback for upaid airdrop
	var nftSeriesEwtTypeID = nftSeriesSetup[0].ID
	ewtDetailFn := make([]models.WhereCondFn, 0)
	ewtDetailFn = append(ewtDetailFn,
		models.WhereCondFn{Condition: " ewt_detail.ewallet_type_id = ? ", CondValue: nftSeriesEwtTypeID},
		models.WhereCondFn{Condition: " ewt_detail.transaction_type = ? ", CondValue: "ADJUSTMENT"},
		models.WhereCondFn{Condition: " ewt_detail.total_in > ? ", CondValue: 0},
	)
	ewtDetail, _ := models.GetEwtDetailFn(ewtDetailFn, false)

	for _, ewtDetailV := range ewtDetail {
		reissueSupply += ewtDetailV.TotalIn
	}

	finalSupply = distributedSupply + adjustedDistributedSupply
	finalSupply = finalSupply + reissueSupply

	// fmt.Println("nftSeriesCode:", nftSeriesCode, "distributedSupply:", distributedSupply, "adjustedDistributedSupply:", adjustedDistributedSupply, "reissueSupply:", reissueSupply, "finalSupply:", finalSupply)
	// set to 0 if supply fall to negative
	if finalSupply < 0 {
		finalSupply = 0
	}

	return finalSupply, ""
}
