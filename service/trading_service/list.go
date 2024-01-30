package trading_service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/float"
)

const orangeColorCode = "#FFA500"
const greenColorCode = "#13B126"
const redColorCode = "#F76464"
const whiteColorCode = "#ffffff"

type MemberTradingBuyListStruct struct {
	ID                     int     `json:"id"`
	CoinPair               string  `json:"coin_pair"`
	TransDateTime          string  `json:"trans_date_time"`
	MatchedRate            float64 `json:"matched_rate"`
	MatchedRateColorCode   string  `json:"matched_rate_color_code"`
	MatchedQuantityDisplay string  `json:"matched_quantity_display"`
	TotalQuantityDisplay   string  `json:"total_quantity_display"`
	BalanceQuantity        float64 `json:"balance_quantity"`
	UnitPrice              float64 `json:"unit_price"`
	UnitPriceDisplay       string  `json:"unit_price_display"`
	StatusDesc             string  `json:"status_desc"`
	DisplayCancelButton    int     `json:"display_cancel_button"`
	StatusColorCode        string  `json:"status_color_code"`
}

type MemberTradingBuyListPaginateStruct struct {
	EntMemberID int
	LangCode    string
	CryptoCode  string
	Page        int64
}

// func GetMemberTradingBuyPaginateListv1
func GetMemberTradingBuyPaginateListv1(arrData MemberTradingBuyListPaginateStruct) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingBuyListStruct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " trading_buy.status IN (?, '') ", CondValue: "P"},
	)

	arrPaginateData, arrDataList, _ := models.GetTradingBuyPaginateFn(arrCond, arrData.Page, false)

	if len(arrDataList) > 0 {
		pendingTranslated := helpers.TranslateV2("pending", arrData.LangCode, nil)
		processingTranslated := helpers.TranslateV2("processing", arrData.LangCode, nil)
		for _, arrDataListV := range arrDataList {
			matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)
			var displayCancelButton int
			statusDesc := processingTranslated
			statusColorCode := orangeColorCode
			if arrDataListV.Status != "" {
				displayCancelButton = 1
				statusDesc = pendingTranslated
				// statusColorCode = greenColorCode
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrDataListV.CryptoToDecimalPoint), ".", ",")
			}
			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingBuyListStruct{
					ID:                     arrDataListV.ID,
					CoinPair:               helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil),
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   greenColorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrDataListV.CryptoToDecimalPoint), ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              arrDataListV.UnitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(arrDataListV.UnitPrice, uint(arrDataListV.CryptoFromDecimalPoint), ".", ","),
					DisplayCancelButton:    displayCancelButton,
					StatusDesc:             statusDesc,
					StatusColorCode:        statusColorCode,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

type MemberTradingSellListStruct struct {
	ID                     int     `json:"id"`
	CoinPair               string  `json:"coin_pair"`
	TransDateTime          string  `json:"trans_date_time"`
	MatchedRate            float64 `json:"matched_rate"`
	MatchedRateColorCode   string  `json:"matched_rate_color_code"`
	MatchedQuantityDisplay string  `json:"matched_quantity_display"`
	TotalQuantityDisplay   string  `json:"total_quantity_display"`
	BalanceQuantity        float64 `json:"balance_quantity"`
	UnitPrice              float64 `json:"unit_price"`
	UnitPriceDisplay       string  `json:"unit_price_display"`
	StatusDesc             string  `json:"status_desc"`
	DisplayCancelButton    int     `json:"display_cancel_button"`
	StatusColorCode        string  `json:"status_color_code"`
}

type MemberTradingSellListPaginateStruct struct {
	EntMemberID int
	LangCode    string
	CryptoCode  string
	Page        int64
}

// func GetMemberTradingSellPaginateListv1
func GetMemberTradingSellPaginateListv1(arrData MemberTradingSellListPaginateStruct) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingSellListStruct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " trading_sell.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_sell.status IN (?, '') ", CondValue: "P"},
	)

	arrPaginateData, arrDataList, _ := models.GetTradingSellPaginateFn(arrCond, arrData.Page, false)

	if len(arrDataList) > 0 {
		pendingTranslated := helpers.TranslateV2("pending", arrData.LangCode, nil)
		processingTranslated := helpers.TranslateV2("processing", arrData.LangCode, nil)
		for _, arrDataListV := range arrDataList {
			matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)
			var displayCancelButton int
			statusDesc := processingTranslated
			statusColorCode := orangeColorCode
			if arrDataListV.Status != "" {
				displayCancelButton = 1
				statusDesc = pendingTranslated
				// statusColorCode = greenColorCode
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrDataListV.CryptoFromDecimalPoint), ".", ",")
			}

			// fmt.Println("arrDataListV.UnitPrice:", arrDataListV.UnitPrice)
			// unitPrice, ok := arrDataListV.UnitPrice.(float64) // for interface{}
			// fmt.Println("ok:", ok)
			// unitPrice, _ := strconv.ParseFloat(arrDataListV.UnitPrice, 64)
			unitPrice := arrDataListV.UnitPrice
			// fmt.Println("unitPrice:", unitPrice)

			// arrDataListV.UnitPrice = float.TrailZeroFloat(unitPrice, 9)

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingSellListStruct{
					ID:                     arrDataListV.ID,
					CoinPair:               helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil),
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   redColorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrDataListV.CryptoFromDecimalPoint), ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              unitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(unitPrice, uint(arrDataListV.CryptoToDecimalPoint), ".", ","),
					DisplayCancelButton:    displayCancelButton,
					StatusDesc:             statusDesc,
					StatusColorCode:        statusColorCode,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

type MemberTradingTransListPaginateStruct struct {
	MemberID int
	LangCode string
	Type     string
	Page     int64
}

type MemberTradingTransListStruct struct {
	ID                          int     `json:"id"`
	Username                    string  `json:"username"`
	Quantity                    float64 `json:"quantity"`
	QuantityDisplay             string  `json:"quantity_display"`
	QuantityCurrencyCodeName    string  `json:"quantity_currency_code_name"`
	UnitPrice                   float64 `json:"unit_price"`
	UnitPriceDisplay            string  `json:"unit_price_display"`
	UnitPriceCurrencyCodeName   string  `json:"unit_price_currency_code_name"`
	TotalAmount                 float64 `json:"total_amount"`
	TotalAmountDisplay          string  `json:"total_amount_display"`
	TotalAmountCurrencyCodeName string  `json:"total_amount_currency_code_name"`
}

// func GetMemberTradingTransPaginateListv1
func GetMemberTradingTransPaginateListv1(arrData MemberTradingTransListPaginateStruct) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingTransListStruct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPageItems: arrMemberTradingList,
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_match.status = ? ", CondValue: arrData.Type},
	)

	arrPaginateData, arrDataList, _ := models.GetTradingMatchPaginateFn(arrCond, arrData.Page, false)
	fmt.Println("arrDataList:", arrDataList)
	if len(arrDataList) > 0 {
		// for _, arrDataListV := range arrDataList {
		// arrMemberTradingList = append(arrMemberTradingList,
		// 	MemberTradingSellListStruct{
		// 		ID:                          arrDataListV.ID,
		// 		Username:                    arrDataListV.NickName,
		// 		Quantity:                    arrDataListV.TotalUnit,
		// 		QuantityDisplay:             helpers.CutOffDecimal(arrDataListV.TotalUnit, 0, ".", ","),
		// 		QuantityCurrencyCodeName:    helpers.TranslateV2(arrData.CryptoCode, arrData.LangCode, nil),
		// 		UnitPrice:                   arrDataListV.UnitPrice,
		// 		UnitPriceDisplay:            helpers.CutOffDecimal(arrDataListV.UnitPrice, 4, ".", ","),
		// 		UnitPriceCurrencyCodeName:   helpers.TranslateV2("usdt", arrData.LangCode, nil),
		// 		TotalAmount:                 arrDataListV.TotalAmount,
		// 		TotalAmountDisplay:          helpers.CutOffDecimal(arrDataListV.TotalAmount, 4, ".", ","),
		// 		TotalAmountCurrencyCodeName: helpers.TranslateV2("usdt", arrData.LangCode, nil),
		// 	},
		// )
		// }
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

type MemberTradingMarketListPaginateStruct struct {
	MemberID int
	LangCode string
	Page     int64
}

type MemberTradingMarketListStruct struct {
	CryptoCodeFrom                string `json:"crypto_code_from"`
	CryptoCodeTo                  string `json:"crypto_code_to"`
	QuantityDisplay               string `json:"quantity_display"`
	ChangesPeriodUnit             string `json:"changes_period_unit"`
	UnitPriceDisplay              string `json:"unit_price_display"`
	UnitPriceCurrencyCodeName     string `json:"unit_price_currency_code_name"`
	ChangesPeriodPercentDisplay   string `json:"changes_period_percent_display"`
	ChangesPeriodPercentColorCode string `json:"changes_period_percent_color_code"`
	TotalAmountDisplay            string `json:"total_amount_display"`
	TotalAmountColorCode          string `json:"total_amount_color_code"`
}

// func GetMemberTradingMarketPaginateListv1
func GetMemberTradingMarketPaginateListv1(arrData MemberTradingMarketListPaginateStruct) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingMarketListStruct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPageItems: arrMemberTradingList,
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
	)

	arrPaginateData, arrDataList, _ := models.GetTradingSetupPaginateFn(arrCond, arrData.Page, false)

	if len(arrDataList) > 0 {
		curDateTimeT := base.GetCurrentDateTimeT()
		yestDateTimeT := base.AddDurationInString(curDateTimeT, " -1 day")
		// yestDateTimeString := yestDateTimeT.Format("2006-01-02 15:04:05")
		yestDateString := yestDateTimeT.Format("2006-01-02")

		for _, arrDataListV := range arrDataList {
			unitPrice := float64(1)
			yestUnitPrice := float64(1)
			var quantity float64
			// fmt.Println("CodeFrom:", arrDataListV.CodeFrom)
			arrMarketPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrDataListV.CodeFrom)
			if arrMarketPrice > 0 {
				unitPrice = arrMarketPrice
			}
			if strings.ToLower(arrDataListV.CodeFrom) == "liga" {
				// arrMarketPrice, _ := models.GetLatestLigaPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_liga.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementLigaFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	fmt.Println("arrMarketPriceMovementListV", arrMarketPriceMovementListV)
					// 	totRecord := len(arrMarketPriceMovementList)
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			} else if strings.ToLower(arrDataListV.CodeFrom) == "sec" {
				// arrMarketPrice, _ := models.GetLatestSecPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_sec.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementSecFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// totRecord := len(arrMarketPriceMovementList)
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			}
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_match.created_at) = ? ", CondValue: yestDateString},
				models.WhereCondFn{Condition: " trading_match.status = ? ", CondValue: "M"},
				models.WhereCondFn{Condition: " trading_match.crypto_code = ? ", CondValue: arrDataListV.CodeFrom},
			)
			arrTradingMatchMarketDetails, _ := models.GetTradingMatchMarketDetails(arrCond, false)

			if arrTradingMatchMarketDetails.Volume > 0 {
				quantity = arrTradingMatchMarketDetails.Volume
			}

			diffUnitPrice := unitPrice - yestUnitPrice
			// diffUnitPriceRate := diffUnitPrice / unitPrice * 100
			ratio := float.Div(diffUnitPrice, unitPrice)
			diffUnitPriceRate := float.Mul(ratio, 100)
			// fmt.Println("unitPrice:", unitPrice)
			// fmt.Println("yestUnitPrice:", yestUnitPrice)
			// fmt.Println("diffUnitPrice:", diffUnitPrice)
			// fmt.Println("diffUnitPriceRate:", diffUnitPriceRate)

			changesPeriodPercentColorCode := greenColorCode
			if diffUnitPrice < 0 {
				changesPeriodPercentColorCode = redColorCode
			}

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingMarketListStruct{
					CryptoCodeFrom:  helpers.TranslateV2(arrDataListV.NameFrom, arrData.LangCode, nil),
					CryptoCodeTo:    helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					QuantityDisplay: fmt.Sprintf("%f", quantity),
					// QuantityDisplay:               helpers.CutOffDecimal(quantity, uint(arrDataListV.DecimalPointFrom), ".", ","),
					ChangesPeriodUnit: "24h",
					// UnitPriceDisplay:              helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					UnitPriceDisplay:              fmt.Sprintf("%f", unitPrice),
					UnitPriceCurrencyCodeName:     helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					ChangesPeriodPercentDisplay:   helpers.CutOffDecimal(diffUnitPriceRate, 2, ".", ",") + "%",
					ChangesPeriodPercentColorCode: changesPeriodPercentColorCode,
					// TotalAmountDisplay:            helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					TotalAmountDisplay:   fmt.Sprintf("%f", unitPrice),
					TotalAmountColorCode: changesPeriodPercentColorCode,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

type AvailableTradingPriceListRst struct {
	AvailableTradingPriceList []AvailableTradingPrice `json:"available_trading_price_list"`
	UnitPriceDisplay          string                  `json:"unit_price_display"`
	UnitPrice                 string                  `json:"unit_price"`
	QuantityDisplay           string                  `json:"quantity_display"`
}

type AvailableTradingPrice struct {
	UnitPriceDisplay string `json:"unit_price_display"`
	QuantityDisplay  string `json:"quantity_display"`
}

func GetAvailableSecTradingBuyList(quantitative string, langCode string) *AvailableTradingPriceListRst {
	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType("SEC")
	var arrDataReturn AvailableTradingPriceListRst
	arrAvailableTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: "SEC"},
	)

	if strings.ToLower(quantitative) == "low" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.unit_price > ? ", CondValue: latestPrice},
		)
	} else if strings.ToLower(quantitative) == "high" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.unit_price < ? ", CondValue: latestPrice},
		)
	}

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingBuyListFn(arrCond, 14, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: "SEC"},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			arrAvailableTradingPriceList = append(arrAvailableTradingPriceList,
				AvailableTradingPrice{
					QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
					UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
				},
			)
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: "SEC"},
		models.WhereCondFn{Condition: " trading_buy.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingBuyListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.QuantityDisplay = quantityDisplay
	arrDataReturn.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", langCode, nil)
	arrDataReturn.AvailableTradingPriceList = arrAvailableTradingPriceList

	return &arrDataReturn
}

func GetAvailableLigaTradingBuyList(quantitative string, langCode string) *AvailableTradingPriceListRst {
	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType("LIGA")
	var arrDataReturn AvailableTradingPriceListRst
	arrAvailableTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: "LIGA"},
	)

	if strings.ToLower(quantitative) == "low" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.unit_price > ? ", CondValue: latestPrice},
		)
	} else if strings.ToLower(quantitative) == "high" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.unit_price < ? ", CondValue: latestPrice},
		)
	}

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingBuyListFn(arrCond, 14, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: "LIGA"},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			arrAvailableTradingPriceList = append(arrAvailableTradingPriceList,
				AvailableTradingPrice{
					QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
					UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
				},
			)
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: "LIGA"},
		models.WhereCondFn{Condition: " trading_buy.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingBuyListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.QuantityDisplay = quantityDisplay
	arrDataReturn.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", langCode, nil)
	arrDataReturn.AvailableTradingPriceList = arrAvailableTradingPriceList

	return &arrDataReturn
}

func GetAvailableSecTradingSellList(quantitative string, langCode string) *AvailableTradingPriceListRst {
	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType("SEC")
	var arrDataReturn AvailableTradingPriceListRst
	arrAvailableTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: "SEC"},
	)

	if strings.ToLower(quantitative) == "low" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.unit_price > ? ", CondValue: latestPrice},
		)
	} else if strings.ToLower(quantitative) == "high" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.unit_price < ? ", CondValue: latestPrice},
		)
	}

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingSellListFn(arrCond, 14, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: "SEC"},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			arrAvailableTradingPriceList = append(arrAvailableTradingPriceList,
				AvailableTradingPrice{
					QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
					UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
				},
			)
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: "SEC"},
		models.WhereCondFn{Condition: " trading_sell.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingSellListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.QuantityDisplay = quantityDisplay
	arrDataReturn.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", langCode, nil)
	arrDataReturn.AvailableTradingPriceList = arrAvailableTradingPriceList

	return &arrDataReturn
}

func GetAvailableLigaTradingSellList(quantitative string, langCode string) *AvailableTradingPriceListRst {
	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType("LIGA")
	var arrDataReturn AvailableTradingPriceListRst
	arrAvailableTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: "LIGA"},
	)

	if strings.ToLower(quantitative) == "low" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.unit_price > ? ", CondValue: latestPrice},
		)
	} else if strings.ToLower(quantitative) == "high" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.unit_price < ? ", CondValue: latestPrice},
		)
	}

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingSellListFn(arrCond, 14, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: "LIGA"},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			arrAvailableTradingPriceList = append(arrAvailableTradingPriceList,
				AvailableTradingPrice{
					QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
					UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
				},
			)
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: "LIGA"},
		models.WhereCondFn{Condition: " trading_sell.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingSellListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.QuantityDisplay = quantityDisplay
	arrDataReturn.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", langCode, nil)
	arrDataReturn.AvailableTradingPriceList = arrAvailableTradingPriceList

	return &arrDataReturn
}

type MemberTradingHistoryTransListv1 struct {
	EntMemberID int
	LangCode    string
	Page        int
	DateFrom    string
	DateTo      string
}

type MemberTradingOpenOrderTransListv1Struct struct {
	ID                     int     `json:"id"`
	ActionType             string  `json:"action_type"`
	CoinPair               string  `json:"coin_pair"`
	TransDateTime          string  `json:"trans_date_time"`
	MatchedRate            float64 `json:"matched_rate"`
	MatchedRateColorCode   string  `json:"matched_rate_color_code"`
	MatchedQuantityDisplay string  `json:"matched_quantity_display"`
	TotalQuantityDisplay   string  `json:"total_quantity_display"`
	BalanceQuantity        float64 `json:"balance_quantity"`
	UnitPrice              float64 `json:"unit_price"`
	UnitPriceDisplay       string  `json:"unit_price_display"`
	StatusDesc             string  `json:"status_desc"`
	DisplayCancelButton    int     `json:"display_cancel_button"`
	StatusColorCode        string  `json:"status_color_code"`
}

// func GetMemberTradingBuyOpenOrderTransListv1
func GetMemberTradingBuyOpenOrderTransListv1(arrData MemberTradingHistoryTransListv1) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingOpenOrderTransListv1Struct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_buy.status IN (?, '') ", CondValue: "P"},
	)

	if arrData.DateFrom != "" {
		dateFromT, _ := time.Parse("02-01-2006", arrData.DateFrom)
		dateFromString := dateFromT.Format("2006-01-02")
		if dateFromString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_buy.created_at) >= ? ", CondValue: dateFromString},
			)
		}
	}
	if arrData.DateTo != "" {
		dateToT, _ := time.Parse("02-01-2006", arrData.DateTo)
		dateToString := dateToT.Format("2006-01-02")
		if dateToString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_buy.created_at) <= ? ", CondValue: dateToString},
			)
		}
	}

	arrPaginateData, arrDataList, _ := models.GetTradingBuyPaginateFn(arrCond, int64(arrData.Page), false)

	if len(arrDataList) > 0 {
		pendingTranslated := helpers.TranslateV2("pending", arrData.LangCode, nil)
		processingTranslated := helpers.TranslateV2("processing", arrData.LangCode, nil)
		for _, arrDataListV := range arrDataList {
			matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)
			var displayCancelButton int
			statusDesc := processingTranslated
			statusColorCode := orangeColorCode
			if arrDataListV.Status != "" {
				displayCancelButton = 1
				statusDesc = pendingTranslated
				// statusColorCode = greenColorCode
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrDataListV.CryptoToDecimalPoint), ".", ",")
			}
			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingOpenOrderTransListv1Struct{
					ID:                     arrDataListV.ID,
					ActionType:             "BUY",
					CoinPair:               helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil),
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   greenColorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrDataListV.CryptoToDecimalPoint), ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              arrDataListV.UnitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(arrDataListV.UnitPrice, uint(arrDataListV.CryptoFromDecimalPoint), ".", ","),
					DisplayCancelButton:    displayCancelButton,
					StatusDesc:             statusDesc,
					StatusColorCode:        statusColorCode,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}

	return arrDataReturn
}

// func GetMemberTradingSellOpenOrderTransListv1
func GetMemberTradingSellOpenOrderTransListv1(arrData MemberTradingHistoryTransListv1) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingOpenOrderTransListv1Struct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_sell.status IN (?, '') ", CondValue: "P"},
	)

	if arrData.DateFrom != "" {
		dateFromT, _ := time.Parse("02-01-2006", arrData.DateFrom)
		dateFromString := dateFromT.Format("2006-01-02")
		if dateFromString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_sell.created_at) >= ? ", CondValue: dateFromString},
			)
		}
	}

	if arrData.DateTo != "" {
		dateToT, _ := time.Parse("02-01-2006", arrData.DateTo)
		dateToString := dateToT.Format("2006-01-02")
		if dateToString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_sell.created_at) <= ? ", CondValue: dateToString},
			)
		}
	}

	arrPaginateData, arrDataList, _ := models.GetTradingSellPaginateFn(arrCond, int64(arrData.Page), false)

	if len(arrDataList) > 0 {
		pendingTranslated := helpers.TranslateV2("pending", arrData.LangCode, nil)
		processingTranslated := helpers.TranslateV2("processing", arrData.LangCode, nil)
		for _, arrDataListV := range arrDataList {
			matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)
			var displayCancelButton int
			statusDesc := processingTranslated
			statusColorCode := orangeColorCode
			if arrDataListV.Status != "" {
				displayCancelButton = 1
				statusDesc = pendingTranslated
				// statusColorCode = greenColorCode
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrDataListV.CryptoFromDecimalPoint), ".", ",")
			}

			// unitPrice, ok := arrDataListV.UnitPrice.(float64) // for interface{}
			// fmt.Println("ok:", ok)
			// unitPrice, _ := strconv.ParseFloat(arrDataListV.UnitPrice, 64)
			unitPrice := arrDataListV.UnitPrice

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingOpenOrderTransListv1Struct{
					ID:                     arrDataListV.ID,
					ActionType:             "SELL",
					CoinPair:               helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil),
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   redColorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrDataListV.CryptoFromDecimalPoint), ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              unitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(unitPrice, uint(arrDataListV.CryptoToDecimalPoint), ".", ","),
					DisplayCancelButton:    displayCancelButton,
					StatusDesc:             statusDesc,
					StatusColorCode:        statusColorCode,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

// func GetMemberTradingOpenOrderTransListv1
func GetMemberTradingOpenOrderTransListv1(arrData MemberTradingHistoryTransListv1) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingOpenOrderTransListv1Struct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrTradSellRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrTradBuyRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrUnionCond := make(map[string][]models.ArrUnionRawCondText, 0)

	arrTradSellRawSQL = append(arrTradSellRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_sell.member_id = " + strconv.Itoa(arrData.EntMemberID)},
		models.ArrUnionRawCondText{Cond: " AND trading_sell.status IN ('P', '')"},
	)
	arrTradBuyRawSQL = append(arrTradBuyRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_buy.member_id = " + strconv.Itoa(arrData.EntMemberID)},
		models.ArrUnionRawCondText{Cond: " AND trading_buy.status IN ('P', '')"},
	)

	if arrData.DateFrom != "" {
		dateFromT, _ := time.Parse("02-01-2006", arrData.DateFrom)
		dateFromString := dateFromT.Format("2006-01-02")
		if dateFromString != "0001-01-01" {
			arrTradSellRawSQL = append(arrTradSellRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_sell.created_at) >= '" + dateFromString + "'"},
			)
			arrTradBuyRawSQL = append(arrTradBuyRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_buy.created_at) >= '" + dateFromString + "'"},
			)
		}
	}
	if arrData.DateTo != "" {
		dateToT, _ := time.Parse("02-01-2006", arrData.DateTo)
		dateToString := dateToT.Format("2006-01-02")
		if dateToString != "0001-01-01" {
			arrTradSellRawSQL = append(arrTradSellRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_sell.created_at) <= '" + dateToString + "'"},
			)
			arrTradBuyRawSQL = append(arrTradBuyRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_buy.created_at) <= '" + dateToString + "'"},
			)
		}
	}

	arrUnionCond["trading_sell"] = arrTradSellRawSQL
	arrUnionCond["trading_buy"] = arrTradBuyRawSQL

	arrPaginateData, arrDataList, _ := models.GetTradingBuySellPaginateFn(arrUnionCond, int64(arrData.Page), false)

	if len(arrDataList) > 0 {
		pendingTranslated := helpers.TranslateV2("pending", arrData.LangCode, nil)
		processingTranslated := helpers.TranslateV2("processing", arrData.LangCode, nil)
		for _, arrDataListV := range arrDataList {
			matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)

			var displayCancelButton int
			statusDesc := processingTranslated
			statusColorCode := orangeColorCode
			if arrDataListV.Status != "" {
				displayCancelButton = 1
				statusDesc = pendingTranslated
				// statusColorCode = greenColorCode
			}

			coinPair := helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil)
			colorCode := redColorCode
			if strings.ToLower(arrDataListV.ActionType) == "buy" {
				coinPair = helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil)
				colorCode = greenColorCode
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, 8, ".", ",")
			}

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingOpenOrderTransListv1Struct{
					ID:                     arrDataListV.ID,
					ActionType:             arrDataListV.ActionType,
					CoinPair:               coinPair,
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   colorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, 8, ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              arrDataListV.UnitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(arrDataListV.UnitPrice, 8, ".", ","),
					DisplayCancelButton:    displayCancelButton,
					StatusDesc:             statusDesc,
					StatusColorCode:        statusColorCode,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

type MemberTradingOrderHistoryTransListv1Struct struct {
	ID                     int     `json:"id"`
	ActionType             string  `json:"action_type"`
	CoinPair               string  `json:"coin_pair"`
	TransDateTime          string  `json:"trans_date_time"`
	MatchedRate            float64 `json:"matched_rate"`
	MatchedRateColorCode   string  `json:"matched_rate_color_code"`
	MatchedQuantityDisplay string  `json:"matched_quantity_display"`
	TotalQuantityDisplay   string  `json:"total_quantity_display"`
	BalanceQuantity        float64 `json:"balance_quantity"`
	UnitPrice              float64 `json:"unit_price"`
	UnitPriceDisplay       string  `json:"unit_price_display"`
	CancelQuantityDisplay  string  `json:"cancel_quantity_display"`
}

// func GetMemberTradingBuyOrderHistoryTransListv1
func GetMemberTradingBuyOrderHistoryTransListv1(arrData MemberTradingHistoryTransListv1) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingOrderHistoryTransListv1Struct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.status IN ('P', 'AP', 'M', 'C') AND trading_buy.member_id = ? ", CondValue: arrData.EntMemberID},
	)

	if arrData.DateFrom != "" {
		dateFromT, _ := time.Parse("02-01-2006", arrData.DateFrom)
		dateFromString := dateFromT.Format("2006-01-02")
		if dateFromString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_buy.created_at) >= ? ", CondValue: dateFromString},
			)
		}
	}

	if arrData.DateTo != "" {
		dateToT, _ := time.Parse("02-01-2006", arrData.DateTo)
		dateToString := dateToT.Format("2006-01-02")
		if dateToString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_buy.created_at) <= ? ", CondValue: dateToString},
			)
		}
	}

	arrPaginateData, arrDataList, _ := models.GetTradingBuyPaginateFn(arrCond, int64(arrData.Page), false)

	if len(arrDataList) > 0 {
		for _, arrDataListV := range arrDataList {

			var cancelQuantity float64
			if strings.ToLower(arrDataListV.Status) == "c" {
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " trading_cancel.status IN (?)", CondValue: "AP"},
					models.WhereCondFn{Condition: " trading_cancel.transaction_type = ?", CondValue: "BUY"},
					models.WhereCondFn{Condition: " trading_cancel.trading_id = ?", CondValue: arrDataListV.ID},
				)

				arrTradCancel, _ := models.GetTotalTradingCancelFn(arrCond, false)
				if arrTradCancel != nil {
					if arrTradCancel.TotalCancelUnit > 0 {
						cancelQuantity = arrTradCancel.TotalCancelUnit
					}
				}
			}
			// start get matched unit
			matchedUnit := float64(0)
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " trading_match.buy_id = ? AND trading_match.status IN ('M', 'AP')", CondValue: arrDataListV.ID},
			)
			totalTradingMatchRst, _ := models.GetTotalTradingMatchFn(arrCond, false)
			if totalTradingMatchRst.TotalUnit > 0 {
				matchedUnit = totalTradingMatchRst.TotalUnit
			}
			// end get matched unit
			// matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit - cancelQuantity
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)

			if strings.ToLower(arrDataListV.Status) == "c" { // this is bcz fully cancel. in listing, this need to be in 100% used.
				matchedRate = 1
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrDataListV.CryptoToDecimalPoint), ".", ",")
			}

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingOrderHistoryTransListv1Struct{
					ID:                     arrDataListV.ID,
					ActionType:             "BUY",
					CoinPair:               helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil),
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   greenColorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrDataListV.CryptoToDecimalPoint), ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              arrDataListV.UnitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(arrDataListV.UnitPrice, uint(arrDataListV.CryptoFromDecimalPoint), ".", ","),
					CancelQuantityDisplay:  helpers.CutOffDecimal(cancelQuantity, uint(arrDataListV.CryptoToDecimalPoint), ".", ","),
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}

	return arrDataReturn
}

// func GetMemberTradingSellOrderHistoryTransListv1
func GetMemberTradingSellOrderHistoryTransListv1(arrData MemberTradingHistoryTransListv1) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingOrderHistoryTransListv1Struct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.status IN ('P', 'AP', 'M', 'C') AND trading_sell.member_id = ? ", CondValue: arrData.EntMemberID},
	)

	if arrData.DateFrom != "" {
		dateFromT, _ := time.Parse("02-01-2006", arrData.DateFrom)
		dateFromString := dateFromT.Format("2006-01-02")
		if dateFromString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_sell.created_at) >= ? ", CondValue: dateFromString},
			)
		}
	}

	if arrData.DateTo != "" {
		dateToT, _ := time.Parse("02-01-2006", arrData.DateTo)
		dateToString := dateToT.Format("2006-01-02")
		if dateToString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_sell.created_at) <= ? ", CondValue: dateToString},
			)
		}
	}

	arrPaginateData, arrDataList, _ := models.GetTradingSellPaginateFn(arrCond, int64(arrData.Page), false)

	if len(arrDataList) > 0 {
		for _, arrDataListV := range arrDataList {

			var cancelQuantity float64
			if strings.ToLower(arrDataListV.Status) == "c" {
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " trading_cancel.status IN (?)", CondValue: "AP"},
					models.WhereCondFn{Condition: " trading_cancel.transaction_type = ?", CondValue: "SELL"},
					models.WhereCondFn{Condition: " trading_cancel.trading_id = ?", CondValue: arrDataListV.ID},
				)

				arrTradCancel, _ := models.GetTotalTradingCancelFn(arrCond, false)
				if arrTradCancel != nil {
					if arrTradCancel.TotalCancelUnit > 0 {
						cancelQuantity = arrTradCancel.TotalCancelUnit
					}
				}
			}

			// start get matched unit
			matchedUnit := float64(0)
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " trading_match.sell_id = ? AND trading_match.status IN ('M', 'AP')", CondValue: arrDataListV.ID},
			)
			totalTradingMatchRst, _ := models.GetTotalTradingMatchFn(arrCond, false)
			if totalTradingMatchRst.TotalUnit > 0 {
				matchedUnit = totalTradingMatchRst.TotalUnit
			}
			// end get matched unit
			// matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit - cancelQuantity
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)

			if strings.ToLower(arrDataListV.Status) == "c" { // this is bcz fully cancel. in listing, this need to be in 100% used.
				matchedRate = 1
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrDataListV.CryptoFromDecimalPoint), ".", ",")
			}

			// unitPrice, ok := arrDataListV.UnitPrice.(float64) // for interface{}
			// fmt.Println("ok:", ok)
			// unitPrice, _ := strconv.ParseFloat(arrDataListV.UnitPrice, 64)
			unitPrice := arrDataListV.UnitPrice

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingOrderHistoryTransListv1Struct{
					ID:                     arrDataListV.ID,
					ActionType:             "SELL",
					CoinPair:               helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil),
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   redColorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrDataListV.CryptoFromDecimalPoint), ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              unitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(unitPrice, uint(arrDataListV.CryptoToDecimalPoint), ".", ","),
					CancelQuantityDisplay:  helpers.CutOffDecimal(cancelQuantity, uint(arrDataListV.CryptoFromDecimalPoint), ".", ","),
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

// func GetMemberTradingOrderHistoryTransListv1
func GetMemberTradingOrderHistoryTransListv1(arrData MemberTradingHistoryTransListv1) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingOrderHistoryTransListv1Struct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrTradSellRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrTradBuyRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrUnionCond := make(map[string][]models.ArrUnionRawCondText, 0)

	arrTradSellRawSQL = append(arrTradSellRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_sell.member_id = " + strconv.Itoa(arrData.EntMemberID)},
		models.ArrUnionRawCondText{Cond: " AND trading_sell.status IN ('P', 'AP', 'M', 'C')"},
	)
	arrTradBuyRawSQL = append(arrTradBuyRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_buy.member_id = " + strconv.Itoa(arrData.EntMemberID)},
		models.ArrUnionRawCondText{Cond: " AND trading_buy.status IN ('P', 'AP', 'M', 'C')"},
	)

	if arrData.DateFrom != "" {
		dateFromT, _ := time.Parse("02-01-2006", arrData.DateFrom)
		dateFromString := dateFromT.Format("2006-01-02")
		if dateFromString != "0001-01-01" {
			arrTradSellRawSQL = append(arrTradSellRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_sell.created_at) >= '" + dateFromString + "'"},
			)
			arrTradBuyRawSQL = append(arrTradBuyRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_buy.created_at) >= '" + dateFromString + "'"},
			)
		}

	}

	if arrData.DateTo != "" {
		dateToT, _ := time.Parse("02-01-2006", arrData.DateTo)
		dateToString := dateToT.Format("2006-01-02")
		if dateToString != "0001-01-01" {
			arrTradSellRawSQL = append(arrTradSellRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_sell.created_at) <= '" + dateToString + "'"},
			)
			arrTradBuyRawSQL = append(arrTradBuyRawSQL,
				models.ArrUnionRawCondText{Cond: " AND DATE(trading_buy.created_at) <= '" + dateToString + "'"},
			)
		}
	}

	arrUnionCond["trading_sell"] = arrTradSellRawSQL
	arrUnionCond["trading_buy"] = arrTradBuyRawSQL

	arrPaginateData, arrDataList, _ := models.GetTradingBuySellPaginateFn(arrUnionCond, int64(arrData.Page), false)
	if len(arrDataList) > 0 {
		for _, arrDataListV := range arrDataList {

			coinPair := helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil)
			colorCode := redColorCode
			if strings.ToLower(arrDataListV.ActionType) == "buy" {
				coinPair = helpers.TranslateV2(arrDataListV.CryptoToName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoFromName, arrData.LangCode, nil)
				colorCode = greenColorCode
			}

			var cancelQuantity float64
			if strings.ToLower(arrDataListV.Status) == "c" {
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " trading_cancel.status IN (?)", CondValue: "AP"},
					models.WhereCondFn{Condition: " trading_cancel.transaction_type = ?", CondValue: arrDataListV.ActionType},
					models.WhereCondFn{Condition: " trading_cancel.trading_id = ?", CondValue: arrDataListV.ID},
				)

				arrTradCancel, _ := models.GetTotalTradingCancelFn(arrCond, false)
				if arrTradCancel != nil {
					if arrTradCancel.TotalCancelUnit > 0 {
						cancelQuantity = arrTradCancel.TotalCancelUnit
					}
				}
			}

			// start get matched unit
			matchedUnit := float64(0)
			arrCond := make([]models.WhereCondFn, 0)
			if strings.ToLower(arrDataListV.ActionType) == "buy" {
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " trading_match.buy_id = ? AND trading_match.status IN ('M', 'AP')", CondValue: arrDataListV.ID},
				)
			} else if strings.ToLower(arrDataListV.ActionType) == "sell" {
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " trading_match.sell_id = ? AND trading_match.status IN ('M', 'AP')", CondValue: arrDataListV.ID},
				)
			}
			totalTradingMatchRst, _ := models.GetTotalTradingMatchFn(arrCond, false)
			if totalTradingMatchRst.TotalUnit > 0 {
				matchedUnit = totalTradingMatchRst.TotalUnit
			}
			// end get matched unit

			// matchedUnit := arrDataListV.TotalUnit - arrDataListV.BalanceUnit - cancelQuantity
			// matchedRate := matchedUnit / arrDataListV.TotalUnit
			matchedRate := float.Div(matchedUnit, arrDataListV.TotalUnit)
			if strings.ToLower(arrDataListV.Status) == "c" { // this is bcz fully cancel. in listing, this need to be in 100% used.
				matchedRate = 1
			}
			matchedUnitString := "0"
			if matchedUnit > 0 {
				matchedUnitString = helpers.CutOffDecimal(matchedUnit, 8, ".", ",")
			}

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingOrderHistoryTransListv1Struct{
					ID:                     arrDataListV.ID,
					ActionType:             arrDataListV.ActionType,
					CoinPair:               coinPair,
					TransDateTime:          base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					MatchedRate:            matchedRate,
					MatchedQuantityDisplay: matchedUnitString,
					MatchedRateColorCode:   colorCode,
					TotalQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, 8, ".", ","),
					BalanceQuantity:        arrDataListV.BalanceUnit,
					UnitPrice:              arrDataListV.UnitPrice,
					UnitPriceDisplay:       helpers.CutOffDecimal(arrDataListV.UnitPrice, 8, ".", ","),
					CancelQuantityDisplay:  helpers.CutOffDecimal(cancelQuantity, 8, ".", ","),
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

type MemberTradingHistoryTransListv1Struct struct {
	ActionType               string `json:"action_type"`
	ActionTypeDesc           string `json:"action_type_desc"`
	ActionTypeColorCode      string `json:"action_type_color_code"`
	CoinPair                 string `json:"coin_pair"`
	TransDateTime            string `json:"trans_date_time"`
	UnitPriceDisplay         string `json:"unit_price_display"`
	MatchedQuantityDisplay   string `json:"matched_quantity_display"`
	MatchedQuantityColorCode string `json:"matched_quantity_color_code"`
	TotalQuantityDisplay     string `json:"total_quantity_display"`
}

// func GetMemberTradingHistoryTransListv1
func GetMemberTradingHistoryTransListv1(arrData MemberTradingHistoryTransListv1) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingHistoryTransListv1Struct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrData.Page),
		PerPage:               30,
		TotalCurrentPageItems: 0,
		TotalPage:             1,
		TotalPageItems:        0,
		CurrentPageItems:      arrMemberTradingList,
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_match.buyer_member_id = " + strconv.Itoa(arrData.EntMemberID) + " OR trading_match.seller_member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_match.status = ? ", CondValue: "M"},
	)

	if arrData.DateFrom != "" {
		dateFromT, _ := time.Parse("02-01-2006", arrData.DateFrom)
		dateFromString := dateFromT.Format("2006-01-02")
		if dateFromString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_match.created_at) >= ? ", CondValue: dateFromString},
			)
		}
	}
	if arrData.DateTo != "" {
		dateToT, _ := time.Parse("02-01-2006", arrData.DateTo)
		dateToString := dateToT.Format("2006-01-02")
		if dateToString != "0001-01-01" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_match.created_at) <= ? ", CondValue: dateToString},
			)
		}
	}

	arrPaginateData, arrDataList, _ := models.GetTradingMatchPaginateFn(arrCond, int64(arrData.Page), false)

	if len(arrDataList) > 0 {
		buyTranslated := helpers.TranslateV2("buy", arrData.LangCode, nil)
		sellTranslated := helpers.TranslateV2("sell", arrData.LangCode, nil)
		for _, arrDataListV := range arrDataList {
			actionTypeCode := "SELL"
			actionTypeTranslated := sellTranslated
			actionTypeColorCode := redColorCode
			matchedQuantityColorCode := redColorCode
			if arrData.EntMemberID == arrDataListV.BuyerEntMemberID {
				actionTypeCode = "BUY"
				actionTypeTranslated = buyTranslated
				actionTypeColorCode = greenColorCode
				matchedQuantityColorCode = greenColorCode
			}
			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingHistoryTransListv1Struct{
					ActionType:               actionTypeCode,
					ActionTypeDesc:           actionTypeTranslated,
					ActionTypeColorCode:      actionTypeColorCode,
					CoinPair:                 helpers.TranslateV2(arrDataListV.CryptoNameFrom, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrDataListV.CryptoNameTo, arrData.LangCode, nil),
					TransDateTime:            base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					UnitPriceDisplay:         helpers.CutOffDecimal(arrDataListV.UnitPrice, 8, ".", ","),
					MatchedQuantityDisplay:   helpers.CutOffDecimal(arrDataListV.TotalUnit, 8, ".", ","),
					MatchedQuantityColorCode: matchedQuantityColorCode,
					TotalQuantityDisplay:     helpers.CutOffDecimal(arrDataListV.TotalAmount, 8, ".", ","),
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}

	return arrDataReturn
}

type MemberTradingOrderHistoryTransDetailsv1 struct {
	EntMemberID int
	LangCode    string
	ID          int
}

type TradeDetailsStruct struct {
	TransDateTime      string `json:"trans_date_time"`
	QuantityDisplay    string `json:"total_quantity_display"`
	UnitPriceDisplay   string `json:"unit_price_display"`
	UnitPriceColorCode string `json:"matched_rate_color_code"`
	ActionType         string `json:"action_type"`
	ActionTypeDesc     string `json:"action_type_desc"`
}

type MemberTradingOrderHistoryTransDetailsv1Struct struct {
	CoinPair                   string               `json:"coin_pair"`
	FilledRate                 string               `json:"filled_rate"`
	ActionType                 string               `json:"action_type"`
	ActionTypeDesc             string               `json:"action_type_desc"`
	ActionTypeColorCode        string               `json:"action_type_color_code"`
	MatchedQuantityDisplay     string               `json:"matched_quantity_display"`
	TotalCancelQuantityDisplay string               `json:"total_cancel_quantity_display"`
	TotalQuantityDisplay       string               `json:"total_quantity_display"`
	UnitPriceDisplay           string               `json:"unit_price_display"`
	TotalAmountDisplay         string               `json:"total_amount_display"`
	TradeDetail                []TradeDetailsStruct `json:"trade_detail"`
}

// func GetMemberTradingBuyOrderHistoryTransListv1
func GetMemberTradingBuyOrderHistoryTransDetailsv1(arrData MemberTradingOrderHistoryTransDetailsv1) MemberTradingOrderHistoryTransDetailsv1Struct {

	arrDataReturn := MemberTradingOrderHistoryTransDetailsv1Struct{}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_buy.id = ? ", CondValue: arrData.ID},
	)

	arrTrading, _ := models.GetTradingBuyFn(arrCond, false)

	if len(arrTrading) < 1 {
		return arrDataReturn
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: arrTrading[0].CryptoCode},
	)
	arrTradingFrom, _ := models.GetEwtSetupFn(arrCond, "", false)
	if arrTradingFrom == nil {
		return arrDataReturn
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: arrTrading[0].CryptoCodeTo},
	)
	arrTradingTo, _ := models.GetEwtSetupFn(arrCond, "", false)
	if arrTradingTo == nil {
		return arrDataReturn
	}

	arrTradeDetail := make([]TradeDetailsStruct, 0)
	arrTradMatchRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrTradCancelRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrUnionCond := make(map[string][]models.ArrUnionRawCondText, 0)

	arrTradMatchRawSQL = append(arrTradMatchRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_match.buy_id = '" + strconv.Itoa(arrData.ID) + "'"},
	)

	arrTradCancelRawSQL = append(arrTradCancelRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_cancel.transaction_type = 'BUY'"},
		models.ArrUnionRawCondText{Cond: " AND trading_cancel.trading_id = '" + strconv.Itoa(arrData.ID) + "'"},
	)

	arrUnionCond["trading_match"] = arrTradMatchRawSQL
	arrUnionCond["trading_cancel"] = arrTradCancelRawSQL

	arrTradingDetailList, _ := models.GetTradingDetailsListFn(arrUnionCond, false)

	var totalCancelQuantity float64
	if len(arrTradingDetailList) > 0 {
		for _, arrDataListV := range arrTradingDetailList {
			colorCode := greenColorCode
			if strings.ToLower(arrDataListV.ActionType) == "cancel" {
				colorCode = redColorCode
				totalCancelQuantity = totalCancelQuantity + arrDataListV.TotalUnit
			}
			arrTradeDetail = append(arrTradeDetail,
				TradeDetailsStruct{
					TransDateTime:      base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					QuantityDisplay:    helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrTradingTo.DecimalPoint), ".", ","),
					UnitPriceDisplay:   helpers.CutOffDecimal(arrDataListV.UnitPrice, uint(arrTradingFrom.DecimalPoint), ".", ","),
					UnitPriceColorCode: colorCode,
					ActionType:         arrDataListV.ActionType,
					ActionTypeDesc:     helpers.TranslateV2(arrDataListV.ActionType, arrData.LangCode, nil),
				},
			)
		}
	}

	matchedUnit := arrTrading[0].TotalUnit - arrTrading[0].BalanceUnit - totalCancelQuantity
	// matchedUnitRate := matchedUnit / arrTrading[0].TotalUnit * 100
	ratio := float.Div(matchedUnit, arrTrading[0].TotalUnit)
	matchedUnitRate := float.Mul(ratio, 100)
	matchedUnitString := "0"
	if matchedUnit > 0 {
		matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrTradingTo.DecimalPoint), ".", ",")
	}
	arrDataReturn.CoinPair = helpers.TranslateV2(arrTradingTo.EwtTypeName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrTradingFrom.EwtTypeName, arrData.LangCode, nil)
	arrDataReturn.FilledRate = helpers.CutOffDecimal(matchedUnitRate, 2, ".", ",") + "%"
	arrDataReturn.ActionType = "BUY"
	arrDataReturn.ActionTypeDesc = helpers.TranslateV2("buy", arrData.LangCode, nil)
	arrDataReturn.ActionTypeColorCode = greenColorCode
	arrDataReturn.MatchedQuantityDisplay = matchedUnitString
	arrDataReturn.TotalQuantityDisplay = helpers.CutOffDecimal(arrTrading[0].TotalUnit, uint(arrTradingTo.DecimalPoint), ".", ",")
	arrDataReturn.UnitPriceDisplay = helpers.CutOffDecimal(arrTrading[0].UnitPrice, uint(arrTradingTo.DecimalPoint), ".", ",")
	arrDataReturn.TotalAmountDisplay = helpers.CutOffDecimal(arrTrading[0].TotalAmount, uint(arrTradingTo.DecimalPoint), ".", ",")
	arrDataReturn.TotalCancelQuantityDisplay = helpers.CutOffDecimal(totalCancelQuantity, uint(arrTradingTo.DecimalPoint), ".", ",")

	arrDataReturn.TradeDetail = arrTradeDetail

	return arrDataReturn
}

// func GetMemberTradingBuyOrderHistoryTransListv1
func GetMemberTradingSellOrderHistoryTransDetailsv1(arrData MemberTradingOrderHistoryTransDetailsv1) MemberTradingOrderHistoryTransDetailsv1Struct {

	arrDataReturn := MemberTradingOrderHistoryTransDetailsv1Struct{}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.member_id = ? ", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: " trading_sell.id = ? ", CondValue: arrData.ID},
	)

	arrTrading, _ := models.GetTradingSellFn(arrCond, false)

	if len(arrTrading) < 1 {
		return arrDataReturn
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: arrTrading[0].CryptoCode},
	)
	arrTradingFrom, _ := models.GetEwtSetupFn(arrCond, "", false)
	if arrTradingFrom == nil {
		return arrDataReturn
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code = ? ", CondValue: arrTrading[0].CryptoCodeTo},
	)
	arrTradingTo, _ := models.GetEwtSetupFn(arrCond, "", false)
	if arrTradingTo == nil {
		return arrDataReturn
	}

	arrTradeDetail := make([]TradeDetailsStruct, 0)

	arrTradMatchRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrTradCancelRawSQL := make([]models.ArrUnionRawCondText, 0)
	arrUnionCond := make(map[string][]models.ArrUnionRawCondText, 0)

	arrTradMatchRawSQL = append(arrTradMatchRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_match.sell_id = '" + strconv.Itoa(arrData.ID) + "'"},
	)

	arrTradCancelRawSQL = append(arrTradCancelRawSQL,
		models.ArrUnionRawCondText{Cond: " AND trading_cancel.transaction_type = 'SELL'"},
		models.ArrUnionRawCondText{Cond: " AND trading_cancel.trading_id = '" + strconv.Itoa(arrData.ID) + "'"},
	)

	arrUnionCond["trading_match"] = arrTradMatchRawSQL
	arrUnionCond["trading_cancel"] = arrTradCancelRawSQL

	arrTradingDetailList, _ := models.GetTradingDetailsListFn(arrUnionCond, false)

	var totalCancelQuantity float64
	if len(arrTradingDetailList) > 0 {
		for _, arrDataListV := range arrTradingDetailList {
			colorCode := greenColorCode
			if strings.ToLower(arrDataListV.ActionType) == "cancel" {
				colorCode = redColorCode
				totalCancelQuantity = totalCancelQuantity + arrDataListV.TotalUnit
			}
			arrTradeDetail = append(arrTradeDetail,
				TradeDetailsStruct{
					TransDateTime:      base.TimeFormat(arrDataListV.CreatedAt, "2006-01-02 15:04:05"),
					QuantityDisplay:    helpers.CutOffDecimal(arrDataListV.TotalUnit, uint(arrTradingFrom.DecimalPoint), ".", ","),
					UnitPriceDisplay:   helpers.CutOffDecimal(arrDataListV.UnitPrice, uint(arrTradingTo.DecimalPoint), ".", ","),
					UnitPriceColorCode: colorCode,
					ActionType:         arrDataListV.ActionType,
					ActionTypeDesc:     helpers.TranslateV2(arrDataListV.ActionType, arrData.LangCode, nil),
				},
			)
		}
	}

	matchedUnit := arrTrading[0].TotalUnit - arrTrading[0].BalanceUnit - totalCancelQuantity
	// matchedUnitRate := matchedUnit / arrTrading[0].TotalUnit * 100
	ratio := float.Div(matchedUnit, arrTrading[0].TotalUnit)
	matchedUnitRate := float.Mul(ratio, 100)
	matchedUnitString := "0"
	if matchedUnit > 0 {
		matchedUnitString = helpers.CutOffDecimal(matchedUnit, uint(arrTradingTo.DecimalPoint), ".", ",")
	}
	arrDataReturn.CoinPair = helpers.TranslateV2(arrTradingFrom.EwtTypeName, arrData.LangCode, nil) + " / " + helpers.TranslateV2(arrTradingTo.EwtTypeName, arrData.LangCode, nil)
	arrDataReturn.FilledRate = helpers.CutOffDecimal(matchedUnitRate, 2, ".", ",") + "%"
	arrDataReturn.ActionType = "SELL"
	arrDataReturn.ActionTypeDesc = helpers.TranslateV2("sell", arrData.LangCode, nil)
	arrDataReturn.ActionTypeColorCode = redColorCode
	arrDataReturn.MatchedQuantityDisplay = matchedUnitString
	arrDataReturn.TotalQuantityDisplay = helpers.CutOffDecimal(arrTrading[0].TotalUnit, uint(arrTradingTo.DecimalPoint), ".", ",")
	arrDataReturn.UnitPriceDisplay = helpers.CutOffDecimal(arrTrading[0].UnitPrice, uint(arrTradingTo.DecimalPoint), ".", ",")
	arrDataReturn.TotalAmountDisplay = helpers.CutOffDecimal(arrTrading[0].TotalAmount, uint(arrTradingTo.DecimalPoint), ".", ",")
	arrDataReturn.TotalCancelQuantityDisplay = helpers.CutOffDecimal(totalCancelQuantity, uint(arrTradingTo.DecimalPoint), ".", ",")

	arrDataReturn.TradeDetail = arrTradeDetail

	return arrDataReturn
}

// func GetMemberTradingMarketPaginateListv2
func GetMemberTradingMarketPaginateListv2(arrData MemberTradingMarketListPaginateStruct) app.ArrDataResponseDefaultList {

	arrMemberTradingList := make([]MemberTradingMarketListStruct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPageItems: arrMemberTradingList,
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
	)

	arrPaginateData, arrDataList, _ := models.GetTradingSetupPaginateFn(arrCond, arrData.Page, false)

	if len(arrDataList) > 0 {
		curDateTimeT := base.GetCurrentDateTimeT()
		yestDateTimeT := base.AddDurationInString(curDateTimeT, " -1 day")
		// yestDateTimeString := yestDateTimeT.Format("2006-01-02 15:04:05")
		yestDateString := yestDateTimeT.Format("2006-01-02")

		for _, arrDataListV := range arrDataList {
			unitPrice := float64(1)
			yestUnitPrice := float64(1)
			var quantity float64
			// fmt.Println("CodeFrom:", arrDataListV.CodeFrom)
			arrMarketPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrDataListV.CodeFrom)
			if arrMarketPrice > 0 {
				unitPrice = arrMarketPrice
			}
			if strings.ToLower(arrDataListV.CodeFrom) == "liga" {
				// arrMarketPrice, _ := models.GetLatestLigaPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_liga.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementLigaFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	fmt.Println("arrMarketPriceMovementListV", arrMarketPriceMovementListV)
					// 	totRecord := len(arrMarketPriceMovementList)
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			} else if strings.ToLower(arrDataListV.CodeFrom) == "sec" {
				// arrMarketPrice, _ := models.GetLatestSecPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_sec.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementSecFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// totRecord := len(arrMarketPriceMovementList)
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			}
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_match.created_at) = ? ", CondValue: yestDateString},
				models.WhereCondFn{Condition: " trading_match.status = ? ", CondValue: "M"},
				models.WhereCondFn{Condition: " trading_match.crypto_code = ? ", CondValue: arrDataListV.CodeFrom},
			)
			arrTradingMatchMarketDetails, _ := models.GetTradingMatchMarketDetails(arrCond, false)

			if arrTradingMatchMarketDetails.Volume > 0 {
				quantity = arrTradingMatchMarketDetails.Volume
			}

			diffUnitPrice := unitPrice - yestUnitPrice
			// diffUnitPriceRate := diffUnitPrice / unitPrice * 100
			ratio := float.Div(diffUnitPrice, unitPrice)
			diffUnitPriceRate := float.Mul(ratio, 100)
			// fmt.Println("unitPrice:", unitPrice)
			// fmt.Println("yestUnitPrice:", yestUnitPrice)
			// fmt.Println("diffUnitPrice:", diffUnitPrice)
			// fmt.Println("diffUnitPriceRate:", diffUnitPriceRate)

			changesPeriodPercentColorCode := greenColorCode
			if diffUnitPrice < 0 {
				changesPeriodPercentColorCode = redColorCode
			}

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingMarketListStruct{
					CryptoCodeFrom:  helpers.TranslateV2(arrDataListV.NameFrom, arrData.LangCode, nil),
					CryptoCodeTo:    helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					QuantityDisplay: fmt.Sprintf("%f", quantity),
					// QuantityDisplay:               helpers.CutOffDecimal(quantity, uint(arrDataListV.DecimalPointFrom), ".", ","),
					ChangesPeriodUnit: "24h",
					// UnitPriceDisplay:              helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					UnitPriceDisplay:              fmt.Sprintf("%f", unitPrice),
					UnitPriceCurrencyCodeName:     helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					ChangesPeriodPercentDisplay:   helpers.CutOffDecimal(diffUnitPriceRate, 2, ".", ",") + "%",
					ChangesPeriodPercentColorCode: changesPeriodPercentColorCode,
					// TotalAmountDisplay:            helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					TotalAmountDisplay:   fmt.Sprintf("%f", unitPrice),
					TotalAmountColorCode: changesPeriodPercentColorCode,
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrMemberTradingList,
	}
	return arrDataReturn
}

type WSMemberTradingMarketListStruct struct {
	MemberID int
	LangCode string
}

type WSMemberTradingMarketListRstStruct struct {
	ExchangePriceList []MemberTradingMarketListStruct `json:"exchange_price_list"`
}

// func GetWSMemberTradingMarketListv1
func GetWSMemberTradingMarketListv1(arrData WSMemberTradingMarketListStruct) WSMemberTradingMarketListRstStruct {

	arrMemberTradingList := make([]MemberTradingMarketListStruct, 0)
	arrDataReturn := WSMemberTradingMarketListRstStruct{
		ExchangePriceList: arrMemberTradingList,
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
	)

	arrDataList, _ := models.GetTradingSetupFn(arrCond, false)

	if len(arrDataList) > 0 {
		curDateTimeT := base.GetCurrentDateTimeT()
		yestDateTimeT := base.AddDurationInString(curDateTimeT, " -1 day")
		// yestDateTimeString := yestDateTimeT.Format("2006-01-02 15:04:05")
		yestDateString := yestDateTimeT.Format("2006-01-02")

		for _, arrDataListV := range arrDataList {
			unitPrice := float64(1)
			yestUnitPrice := float64(1)
			var quantity float64
			// fmt.Println("CodeFrom:", arrDataListV.CodeFrom)
			arrMarketPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrDataListV.CodeFrom)
			if arrMarketPrice > 0 {
				unitPrice = arrMarketPrice
			}
			if strings.ToLower(arrDataListV.CodeFrom) == "liga" {
				// arrMarketPrice, _ := models.GetLatestLigaPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_liga.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementLigaFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	fmt.Println("arrMarketPriceMovementListV", arrMarketPriceMovementListV)
					// 	totRecord := len(arrMarketPriceMovementList)
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			} else if strings.ToLower(arrDataListV.CodeFrom) == "sec" {
				// arrMarketPrice, _ := models.GetLatestSecPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_sec.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementSecFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// totRecord := len(arrMarketPriceMovementList)
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			}
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_match.created_at) = ? ", CondValue: yestDateString},
				models.WhereCondFn{Condition: " trading_match.status = ? ", CondValue: "M"},
				models.WhereCondFn{Condition: " trading_match.crypto_code = ? ", CondValue: arrDataListV.CodeFrom},
			)
			arrTradingMatchMarketDetails, _ := models.GetTradingMatchMarketDetails(arrCond, false)

			if arrTradingMatchMarketDetails.Volume > 0 {
				quantity = arrTradingMatchMarketDetails.Volume
			}

			diffUnitPrice := unitPrice - yestUnitPrice
			// diffUnitPriceRate := diffUnitPrice / unitPrice * 100
			ratio := float.Div(diffUnitPrice, unitPrice)
			diffUnitPriceRate := float.Mul(ratio, 100)
			// fmt.Println("unitPrice:", unitPrice)
			// fmt.Println("yestUnitPrice:", yestUnitPrice)
			// fmt.Println("diffUnitPrice:", diffUnitPrice)
			// fmt.Println("diffUnitPriceRate:", diffUnitPriceRate)

			changesPeriodPercentColorCode := greenColorCode
			if diffUnitPrice < 0 {
				changesPeriodPercentColorCode = redColorCode
			}

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingMarketListStruct{
					CryptoCodeFrom:  helpers.TranslateV2(arrDataListV.NameFrom, arrData.LangCode, nil),
					CryptoCodeTo:    helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					QuantityDisplay: fmt.Sprintf("%f", quantity),
					// QuantityDisplay:               helpers.CutOffDecimal(quantity, uint(arrDataListV.DecimalPointFrom), ".", ","),
					ChangesPeriodUnit: "24h",
					// UnitPriceDisplay:              helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					UnitPriceDisplay:              fmt.Sprintf("%f", unitPrice),
					UnitPriceCurrencyCodeName:     helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					ChangesPeriodPercentDisplay:   helpers.CutOffDecimal(diffUnitPriceRate, 2, ".", ",") + "%",
					ChangesPeriodPercentColorCode: changesPeriodPercentColorCode,
					// TotalAmountDisplay:            helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					TotalAmountDisplay:   fmt.Sprintf("%f", unitPrice),
					TotalAmountColorCode: changesPeriodPercentColorCode,
				},
			)
		}
	}

	arrDataReturn = WSMemberTradingMarketListRstStruct{
		ExchangePriceList: arrMemberTradingList,
	}
	return arrDataReturn
}

type WSMemberTradingMarketListRstStructv2 struct {
	ExchangePriceList []MemberTradingMarketListStruct `json:"trading_market_price_list"`
	Code              string                          `json:"code"`
}

// func GetWSMemberTradingMarketListv1
func GetWSMemberTradingMarketListvv2(arrData WSMemberTradingMarketListStruct) WSMemberTradingMarketListRstStructv2 {

	arrMemberTradingList := make([]MemberTradingMarketListStruct, 0)
	arrDataReturn := WSMemberTradingMarketListRstStructv2{
		ExchangePriceList: arrMemberTradingList,
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_setup.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " trading_setup.member_show = ? ", CondValue: 1},
	)

	arrDataList, _ := models.GetTradingSetupFn(arrCond, false)

	if len(arrDataList) > 0 {
		curDateTimeT := base.GetCurrentDateTimeT()
		yestDateTimeT := base.AddDurationInString(curDateTimeT, " -1 day")
		// yestDateTimeString := yestDateTimeT.Format("2006-01-02 15:04:05")
		yestDateString := yestDateTimeT.Format("2006-01-02")

		for _, arrDataListV := range arrDataList {
			unitPrice := float64(1)
			yestUnitPrice := float64(1)
			var quantity float64
			// fmt.Println("CodeFrom:", arrDataListV.CodeFrom)
			arrMarketPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrDataListV.CodeFrom)
			if arrMarketPrice > 0 {
				unitPrice = arrMarketPrice
			}
			if strings.ToLower(arrDataListV.CodeFrom) == "liga" {
				// arrMarketPrice, _ := models.GetLatestLigaPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_liga.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementLigaFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	fmt.Println("arrMarketPriceMovementListV", arrMarketPriceMovementListV)
					// 	totRecord := len(arrMarketPriceMovementList)
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			} else if strings.ToLower(arrDataListV.CodeFrom) == "sec" {
				// arrMarketPrice, _ := models.GetLatestSecPriceMovement()
				// if arrMarketPrice > 0 {
				// 	unitPrice = arrMarketPrice
				// }
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " DATE(exchange_price_movement_sec.created_at) <= ? ", CondValue: yestDateString},
				)

				arrMarketPriceMovementList, _ := models.GetExchangePriceMovementSecFn(arrCond, 4, false)
				if len(arrMarketPriceMovementList) > 0 {
					yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// totRecord := len(arrMarketPriceMovementList)
					// for arrMarketPriceMovementListK, arrMarketPriceMovementListV := range arrMarketPriceMovementList {
					// 	if arrMarketPriceMovementListV.BLatest == 1 {
					// 		if totRecord-1 == arrMarketPriceMovementListK { // no last price before the latest price
					// 			yestUnitPrice = arrMarketPriceMovementList[0].TokenPrice
					// 			break
					// 		} else { // next array is existing [take it as last]
					// 			yestUnitPrice = arrMarketPriceMovementList[arrMarketPriceMovementListK+1].TokenPrice
					// 			break
					// 		}
					// 	}
					// }
				}
			}
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " DATE(trading_match.created_at) = ? ", CondValue: yestDateString},
				models.WhereCondFn{Condition: " trading_match.status = ? ", CondValue: "M"},
				models.WhereCondFn{Condition: " trading_match.crypto_code = ? ", CondValue: arrDataListV.CodeFrom},
			)
			arrTradingMatchMarketDetails, _ := models.GetTradingMatchMarketDetails(arrCond, false)

			if arrTradingMatchMarketDetails.Volume > 0 {
				quantity = arrTradingMatchMarketDetails.Volume
			}

			diffUnitPrice := unitPrice - yestUnitPrice
			// diffUnitPriceRate := diffUnitPrice / unitPrice * 100
			ratio := float.Div(diffUnitPrice, unitPrice)
			diffUnitPriceRate := float.Mul(ratio, 100)
			// fmt.Println("unitPrice:", unitPrice)
			// fmt.Println("yestUnitPrice:", yestUnitPrice)
			// fmt.Println("diffUnitPrice:", diffUnitPrice)
			// fmt.Println("diffUnitPriceRate:", diffUnitPriceRate)

			changesPeriodPercentColorCode := greenColorCode
			if diffUnitPrice < 0 {
				changesPeriodPercentColorCode = redColorCode
			}

			arrMemberTradingList = append(arrMemberTradingList,
				MemberTradingMarketListStruct{
					CryptoCodeFrom:  helpers.TranslateV2(arrDataListV.NameFrom, arrData.LangCode, nil),
					CryptoCodeTo:    helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					QuantityDisplay: fmt.Sprintf("%f", quantity),
					// QuantityDisplay:               helpers.CutOffDecimal(quantity, uint(arrDataListV.DecimalPointFrom), ".", ","),
					ChangesPeriodUnit: "24h",
					// UnitPriceDisplay:              helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					UnitPriceDisplay:              fmt.Sprintf("%f", unitPrice),
					UnitPriceCurrencyCodeName:     helpers.TranslateV2(arrDataListV.NameTo, arrData.LangCode, nil),
					ChangesPeriodPercentDisplay:   helpers.CutOffDecimal(diffUnitPriceRate, 2, ".", ",") + "%",
					ChangesPeriodPercentColorCode: changesPeriodPercentColorCode,
					// TotalAmountDisplay:            helpers.CutOffDecimal(unitPrice, uint(arrDataListV.DecimalPointFrom), ".", ","),
					TotalAmountDisplay:   fmt.Sprintf("%f", unitPrice),
					TotalAmountColorCode: changesPeriodPercentColorCode,
				},
			)
		}
	}

	arrDataReturn = WSMemberTradingMarketListRstStructv2{
		Code:              "trading_market_price_list",
		ExchangePriceList: arrMemberTradingList,
	}
	return arrDataReturn
}

type WSMemberAvailableTradingBuyListv1Struct struct {
	CryptoCode string
	LangCode   string
}

type AvailableTradingPriceStruct struct {
	AvailableTradingPriceList []AvailableTradingPrice `json:"available_trading_price_list"`
	CryptoCode                string                  `json:"crypto_code"`
	UnitPriceDisplay          string                  `json:"unit_price_display"`
	UnitPrice                 string                  `json:"unit_price"`
	QuantityDisplay           string                  `json:"quantity_display"`
}

type WSMemberAvailableTradingPriceListRst struct {
	Code                  string                      `json:"code"`
	AvailableTradingPrice AvailableTradingPriceStruct `json:"available_trading_price"`
}

// func GetWSMemberAvailableTradingBuyListv1
func GetWSMemberAvailableTradingBuyListv1(arrData WSMemberAvailableTradingBuyListv1Struct) *WSMemberAvailableTradingPriceListRst {

	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	var arrDataReturn WSMemberAvailableTradingPriceListRst
	arrAvailableTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code = ? ", CondValue: arrData.CryptoCode},
	)

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingBuyListFn(arrCond, 20, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			arrAvailableTradingPriceList = append(arrAvailableTradingPriceList,
				AvailableTradingPrice{
					QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
					UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
				},
			)
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " trading_buy.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingBuyListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.Code = "available_buy_market_price"
	arrDataReturn.AvailableTradingPrice.QuantityDisplay = quantityDisplay
	arrDataReturn.AvailableTradingPrice.CryptoCode = arrData.CryptoCode
	arrDataReturn.AvailableTradingPrice.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", arrData.LangCode, nil)
	arrDataReturn.AvailableTradingPrice.UnitPrice = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",")
	arrDataReturn.AvailableTradingPrice.AvailableTradingPriceList = arrAvailableTradingPriceList

	return &arrDataReturn
}

// func GetWSMemberAvailableTradingSellListv1
func GetWSMemberAvailableTradingSellListv1(arrData WSMemberAvailableTradingBuyListv1Struct) *WSMemberAvailableTradingPriceListRst {

	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	var arrDataReturn WSMemberAvailableTradingPriceListRst
	arrAvailableTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: arrData.CryptoCode},
	)

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingSellListFn(arrCond, 20, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			arrAvailableTradingPriceList = append(arrAvailableTradingPriceList,
				AvailableTradingPrice{
					QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
					UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
				},
			)
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " trading_sell.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingSellListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.Code = "available_sell_market_price"
	arrDataReturn.AvailableTradingPrice.QuantityDisplay = quantityDisplay
	arrDataReturn.AvailableTradingPrice.CryptoCode = arrData.CryptoCode
	arrDataReturn.AvailableTradingPrice.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", arrData.LangCode, nil)
	arrDataReturn.AvailableTradingPrice.UnitPrice = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",")
	arrDataReturn.AvailableTradingPrice.AvailableTradingPriceList = arrAvailableTradingPriceList

	return &arrDataReturn
}

type WSMemberAvailableTradingBuyPriceListV2Rst struct {
	Code                    string                        `json:"code"`
	AvailableBuyMarketPrice AvailableTradingPriceV2Struct `json:"available_sell_market_price"`
}

type AvailableTradingPriceV2Struct struct {
	AvailableHighTradingPriceList []AvailableTradingPrice `json:"available_high_trading_price_list"`
	AvailableLowTradingPriceList  []AvailableTradingPrice `json:"available_low_trading_price_list"`
	CryptoCode                    string                  `json:"crypto_code"`
	UnitPriceDisplay              string                  `json:"unit_price_display"`
	UnitPrice                     string                  `json:"unit_price"`
	QuantityDisplay               string                  `json:"quantity_display"`
}

// func GetWSMemberAvailableTradingBuyListv2
func GetWSMemberAvailableTradingBuyListv2(arrData WSMemberAvailableTradingBuyListv1Struct) *WSMemberAvailableTradingBuyPriceListV2Rst {

	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	var arrDataReturn WSMemberAvailableTradingBuyPriceListV2Rst
	arrAvailableLowTradingPriceList := make([]AvailableTradingPrice, 0)
	arrAvailableHighTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: arrData.CryptoCode},
	)

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingBuyListFn(arrCond, 20, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			if arrAvailableTradingPriceListRstV.UnitPrice < latestPrice {
				arrAvailableLowTradingPriceList = append(arrAvailableLowTradingPriceList,
					AvailableTradingPrice{
						QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
						UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
					},
				)
			} else if arrAvailableTradingPriceListRstV.UnitPrice > latestPrice {
				arrAvailableHighTradingPriceList = append(arrAvailableHighTradingPriceList,
					AvailableTradingPrice{
						QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
						UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
					},
				)
			}
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " trading_buy.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingBuyListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.Code = "available_buy_market_price"
	arrDataReturn.AvailableBuyMarketPrice.QuantityDisplay = quantityDisplay
	arrDataReturn.AvailableBuyMarketPrice.CryptoCode = arrData.CryptoCode
	arrDataReturn.AvailableBuyMarketPrice.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", arrData.LangCode, nil)
	arrDataReturn.AvailableBuyMarketPrice.UnitPrice = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",")
	arrDataReturn.AvailableBuyMarketPrice.AvailableHighTradingPriceList = arrAvailableHighTradingPriceList
	arrDataReturn.AvailableBuyMarketPrice.AvailableLowTradingPriceList = arrAvailableLowTradingPriceList

	return &arrDataReturn
}

type WSMemberAvailableTradingSellPriceListV2Rst struct {
	Code                     string                        `json:"code"`
	AvailableSellMarketPrice AvailableTradingPriceV2Struct `json:"available_buy_market_price"`
}

// func GetWSMemberAvailableTradingSellListv2
func GetWSMemberAvailableTradingSellListv2(arrData WSMemberAvailableTradingBuyListv1Struct) *WSMemberAvailableTradingSellPriceListV2Rst {

	latestPrice, _ := base.GetLatestExchangePriceMovementByTokenType(arrData.CryptoCode)

	var arrDataReturn WSMemberAvailableTradingSellPriceListV2Rst
	arrAvailableLowTradingPriceList := make([]AvailableTradingPrice, 0)
	arrAvailableHighTradingPriceList := make([]AvailableTradingPrice, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: arrData.CryptoCode},
	)

	arrAvailableTradingPriceListRst, _ := models.GetAvailableTradingSellListFn(arrCond, 20, false)

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewallet_type_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)

	decimalPoint := 2
	if arrEwtSetup != nil {
		decimalPoint = arrEwtSetup.DecimalPoint
	}

	if len(arrAvailableTradingPriceListRst) > 0 {
		for _, arrAvailableTradingPriceListRstV := range arrAvailableTradingPriceListRst {
			if arrAvailableTradingPriceListRstV.UnitPrice < latestPrice {
				arrAvailableLowTradingPriceList = append(arrAvailableLowTradingPriceList,
					AvailableTradingPrice{
						QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
						UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
					},
				)
			} else if arrAvailableTradingPriceListRstV.UnitPrice > latestPrice {
				arrAvailableHighTradingPriceList = append(arrAvailableHighTradingPriceList,
					AvailableTradingPrice{
						QuantityDisplay:  helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.TotalBalanceUnit, uint(decimalPoint), ".", ","),
						UnitPriceDisplay: helpers.CutOffDecimal(arrAvailableTradingPriceListRstV.UnitPrice, uint(decimalPoint), ".", ","),
					},
				)
			}
		}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: arrData.CryptoCode},
		models.WhereCondFn{Condition: " trading_sell.unit_price = ? ", CondValue: latestPrice},
	)
	arrCurrentTradingPrice, _ := models.GetAvailableTradingSellListFn(arrCond, 1, false)

	quantityDisplay := helpers.CutOffDecimal(0, uint(decimalPoint), ".", ",")
	if len(arrCurrentTradingPrice) > 0 {
		if arrCurrentTradingPrice[0].TotalBalanceUnit > 0 {
			quantityDisplay = helpers.CutOffDecimal(arrCurrentTradingPrice[0].TotalBalanceUnit, uint(decimalPoint), ".", ",")
		}
	}

	arrDataReturn.Code = "available_sell_market_price"
	arrDataReturn.AvailableSellMarketPrice.QuantityDisplay = quantityDisplay
	arrDataReturn.AvailableSellMarketPrice.CryptoCode = arrData.CryptoCode
	arrDataReturn.AvailableSellMarketPrice.UnitPriceDisplay = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",") + " " + helpers.TranslateV2("usdt", arrData.LangCode, nil)
	arrDataReturn.AvailableSellMarketPrice.UnitPrice = helpers.CutOffDecimal(latestPrice, uint(decimalPoint), ".", ",")
	arrDataReturn.AvailableSellMarketPrice.AvailableHighTradingPriceList = arrAvailableHighTradingPriceList
	arrDataReturn.AvailableSellMarketPrice.AvailableLowTradingPriceList = arrAvailableLowTradingPriceList

	return &arrDataReturn
}

type WSMemberExchangePriceTradingView struct {
	CryptoCode string
	LangCode   string
	PeriodCode string
}
type WSMemberTradingView struct {
	ID     float64 `json:"id"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Vol    float64 `json:"vol"`
	Amount float64 `json:"amount"`
}

type WSMemberTradingViewRst struct {
	Code        string                `json:"code"`
	TradingView []WSMemberTradingView `json:"trading_view"`
}

// func GetWSMemberExchangePriceTradingView
func GetWSMemberExchangePriceTradingView(arrData WSMemberExchangePriceTradingView) WSMemberTradingViewRst {

	var arrDataReturn WSMemberTradingViewRst
	var tradingView []WSMemberTradingView
	if strings.ToLower(arrData.PeriodCode) == "1day" {
		tradingView = GetExchangePriceTradingViewPerDay(arrData)
	} else if strings.Contains(strings.ToLower(arrData.PeriodCode), "min") {
		tradingView = GetExchangePriceTradingViewByMinute(arrData)
	}
	arrDataReturn.Code = "trading_view"
	arrDataReturn.TradingView = tradingView

	return arrDataReturn
}

func GetExchangePriceTradingViewPerDay(arrData WSMemberExchangePriceTradingView) []WSMemberTradingView {
	arrDataReturn := make([]WSMemberTradingView, 0)
	arrCond := make([]models.WhereCondFn, 0)
	minMaxExchangePriceList := make([]*models.MinMaxExchangePriceMovementPerDay, 0)

	minMaxExchangePriceRst, _ := models.GetMinMaxExchangePriceMovementByTokenTypePerDayFn(arrData.CryptoCode, arrCond, false)
	minMaxExchangePriceList = minMaxExchangePriceRst

	for _, minMaxExchangePriceListV := range minMaxExchangePriceList {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " DATE(created_at) = ? ", CondValue: minMaxExchangePriceListV.TimeSlice},
		)
		// start get open price
		arrOpenFn := models.ArrFnStruct{
			ArrCond: arrCond,
			Limit:   1,
			OrderBy: " created_at ASC ",
		}
		openPrice := float64(0)
		openPriceRst, _ := models.GetExchangePriceMovementByTokenTypeFn(arrData.CryptoCode, arrOpenFn, false)
		if len(openPriceRst) > 0 && openPriceRst[0].TokenPrice > 0 {
			openPrice = openPriceRst[0].TokenPrice
			openPriceString := helpers.CutOffDecimalv2(openPrice, 2, ".", ",", true)
			openPrice, _ = strconv.ParseFloat(openPriceString, 64)
		}
		// end get open price

		// start get close price
		arrCloseFn := models.ArrFnStruct{
			ArrCond: arrCond,
			Limit:   1,
			OrderBy: " created_at DESC ",
		}
		closePrice := float64(0)
		closePriceRst, _ := models.GetExchangePriceMovementByTokenTypeFn(arrData.CryptoCode, arrCloseFn, false)
		if len(closePriceRst) > 0 && closePriceRst[0].TokenPrice > 0 {
			closePrice = closePriceRst[0].TokenPrice
			closePriceString := helpers.CutOffDecimalv2(closePrice, 2, ".", ",", true)
			closePrice, _ = strconv.ParseFloat(closePriceString, 64)
		}
		// start get close price

		// start get total amount and total volume
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " crypto_code = ? ", CondValue: arrData.CryptoCode},
			models.WhereCondFn{Condition: " DATE(created_at) = ? ", CondValue: minMaxExchangePriceListV.TimeSlice},
		)
		totalAmount := float64(0)
		totalVolume := float64(0)
		totalTradingRst, _ := models.GetTotalTradingMatchFn(arrCond, false)
		if totalTradingRst.TotalUnit > 0 {
			totalVolume = totalTradingRst.TotalUnit
		}
		if totalTradingRst.TotalAmount > 0 {
			totalAmount = totalTradingRst.TotalAmount
		}
		// start get total amount and total volume
		minPriceString := helpers.CutOffDecimalv2(minMaxExchangePriceListV.MinPrice, 2, ".", ",", true)
		minPriceFloat, _ := strconv.ParseFloat(minPriceString, 64)
		maxPriceString := helpers.CutOffDecimalv2(minMaxExchangePriceListV.MaxPrice, 2, ".", ",", true)
		maxPriceFloat, _ := strconv.ParseFloat(maxPriceString, 64)

		arrDataReturn = append(arrDataReturn,
			WSMemberTradingView{
				ID:     float64(minMaxExchangePriceListV.DTUnix) * 1000,
				Min:    minPriceFloat,
				Max:    maxPriceFloat,
				Open:   openPrice,
				Close:  closePrice,
				Vol:    totalVolume,
				Amount: totalAmount,
			},
		)
	}

	return arrDataReturn
}

func GetExchangePriceTradingViewByMinute(arrData WSMemberExchangePriceTradingView) []WSMemberTradingView {
	arrDataReturn := make([]WSMemberTradingView, 0)
	arrCond := make([]models.WhereCondFn, 0)
	minMaxExchangePriceList := make([]*models.MinMaxExchangePriceMovementByMinute, 0)
	min, _ := strconv.Atoi(strings.Replace(strings.ToLower(arrData.PeriodCode), "min", "", -1))
	minMaxExchangePriceRst, _ := models.GetMinMaxExchangePriceMovementByTokenTypeByMinuteFn(arrData.CryptoCode, min, arrCond, false)
	minMaxExchangePriceList = minMaxExchangePriceRst

	for _, minMaxExchangePriceListV := range minMaxExchangePriceList {
		// fmt.Println(minMaxExchangePriceListV.MinPrice, minMaxExchangePriceListV.MaxPrice, minMaxExchangePriceListV.TimeSlice)
		timeSlideAddedT := minMaxExchangePriceListV.TimeSlice.Add(time.Minute * time.Duration(min))
		timeSlideString := minMaxExchangePriceListV.TimeSlice.Format("2006-01-02 15:04:05")
		timeSlideAddedString := timeSlideAddedT.Format("2006-01-02 15:04:05")
		// fmt.Println(timeSlideAddedString)

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " created_at > ? ", CondValue: timeSlideString},
			models.WhereCondFn{Condition: " created_at < ? ", CondValue: timeSlideAddedString},
		)
		// start get open price
		arrOpenFn := models.ArrFnStruct{
			ArrCond: arrCond,
			Limit:   1,
			OrderBy: " created_at ASC ",
		}
		openPrice := float64(0)
		openPriceRst, _ := models.GetExchangePriceMovementByTokenTypeFn(arrData.CryptoCode, arrOpenFn, false)
		if len(openPriceRst) > 0 && openPriceRst[0].TokenPrice > 0 {
			openPrice = openPriceRst[0].TokenPrice
			openPriceString := helpers.CutOffDecimalv2(openPrice, 2, ".", ",", true)
			openPrice, _ = strconv.ParseFloat(openPriceString, 64)
		}

		// end get open price

		// start get close price
		arrCloseFn := models.ArrFnStruct{
			ArrCond: arrCond,
			Limit:   1,
			OrderBy: " created_at DESC ",
		}
		closePrice := float64(0)
		closePriceRst, _ := models.GetExchangePriceMovementByTokenTypeFn(arrData.CryptoCode, arrCloseFn, false)
		if len(closePriceRst) > 0 && closePriceRst[0].TokenPrice > 0 {
			closePrice = closePriceRst[0].TokenPrice
			closePriceString := helpers.CutOffDecimalv2(closePrice, 2, ".", ",", true)
			closePrice, _ = strconv.ParseFloat(closePriceString, 64)
		}
		// start get close price

		// start get total amount and total volume
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " crypto_code = ? ", CondValue: arrData.CryptoCode},
			models.WhereCondFn{Condition: " created_at > ? ", CondValue: timeSlideString},
			models.WhereCondFn{Condition: " created_at < ? ", CondValue: timeSlideAddedString},
		)
		totalAmount := float64(0)
		totalVolume := float64(0)
		totalTradingRst, _ := models.GetTotalTradingMatchFn(arrCond, false)
		if totalTradingRst.TotalUnit > 0 {
			totalVolume = totalTradingRst.TotalUnit
		}
		if totalTradingRst.TotalAmount > 0 {
			totalAmount = totalTradingRst.TotalAmount
		}
		// start get total amount and total volume
		minPriceString := helpers.CutOffDecimalv2(minMaxExchangePriceListV.MinPrice, 2, ".", ",", true)
		minPriceFloat, _ := strconv.ParseFloat(minPriceString, 64)
		maxPriceString := helpers.CutOffDecimalv2(minMaxExchangePriceListV.MaxPrice, 2, ".", ",", true)
		maxPriceFloat, _ := strconv.ParseFloat(maxPriceString, 64)

		arrDataReturn = append(arrDataReturn,
			WSMemberTradingView{
				ID:     float64(minMaxExchangePriceListV.DTUnix) * 1000,
				Min:    minPriceFloat,
				Max:    maxPriceFloat,
				Open:   openPrice,
				Close:  closePrice,
				Vol:    totalVolume,
				Amount: totalAmount,
			},
		)
	}

	return arrDataReturn
}

// AutoTradingListForm struct
type AutoTradingListForm struct {
	CryptoType string `form:"crypto_type" json:"crypto_type" valid:"Required;MaxSize(5)"`
	Action     string `form:"action" json:"action" valid:"Required;MaxSize(4)"`
	Status     string `form:"status" json:"status" valid:"Required;MaxSize(4)"`
	Username   string `form:"username" json:"username"`
}

// AutoTradingListRst struct
type AutoTradingListRst struct {
	List []AutoTradingList `json:"list"`
}
type AutoTradingList struct {
	DocNo            string  `json:"doc_no"`
	TotalUnit        float64 `json:"total_unit"`
	UnitPrice        float64 `json:"unit_price"`
	TotalBalanceUnit float64 `json:"total_balance_unit"`
}

// func AutoTradingListForm
func GetAutoTradingListv1(arrData AutoTradingListForm) AutoTradingListRst {

	if strings.ToLower(arrData.Action) == "buy" {
		username := "trader_buy"
		if arrData.Username != "" {
			username = arrData.Username
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_buy.crypto_code_to = ? ", CondValue: arrData.CryptoType},
			models.WhereCondFn{Condition: " ent_member.nick_name = ? ", CondValue: username},
		)

		if arrData.Status != "" {
			arrCond = append(arrCond, models.WhereCondFn{Condition: " trading_buy.status = ? ", CondValue: arrData.Status})
		}

		arrTradingListRst, _ := models.GetTradingBuyFn(arrCond, false)

		autoTradingList := make([]AutoTradingList, 0)
		if len(arrTradingListRst) > 0 {
			for _, arrTradingListV := range arrTradingListRst {
				autoTradingList = append(autoTradingList,
					AutoTradingList{
						DocNo:            arrTradingListV.DocNo,
						TotalUnit:        arrTradingListV.TotalUnit,
						UnitPrice:        arrTradingListV.UnitPrice,
						TotalBalanceUnit: arrTradingListV.BalanceUnit,
					},
				)
			}
		}

		arrDataReturn := AutoTradingListRst{
			List: autoTradingList,
		}

		return arrDataReturn

	} else {
		username := "trader_sell"
		if arrData.Username != "" {
			username = arrData.Username
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_sell.crypto_code = ? ", CondValue: arrData.CryptoType},
			models.WhereCondFn{Condition: " ent_member.nick_name = ? ", CondValue: username},
		)
		if arrData.Status != "" {
			arrCond = append(arrCond, models.WhereCondFn{Condition: " trading_sell.status = ? ", CondValue: arrData.Status})
		}
		arrTradingListRst, _ := models.GetTradingSellFn(arrCond, false)

		autoTradingList := make([]AutoTradingList, 0)
		if len(arrTradingListRst) > 0 {
			for _, arrTradingListV := range arrTradingListRst {
				autoTradingList = append(autoTradingList,
					AutoTradingList{
						DocNo:            arrTradingListV.DocNo,
						TotalUnit:        arrTradingListV.TotalUnit,
						UnitPrice:        arrTradingListV.UnitPrice,
						TotalBalanceUnit: arrTradingListV.BalanceUnit,
					},
				)
			}
		}

		arrDataReturn := AutoTradingListRst{
			List: autoTradingList,
		}

		return arrDataReturn
	}
}
