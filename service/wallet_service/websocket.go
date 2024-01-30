package wallet_service

import (
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/pkg/base"
)

// WSExchangePrice struct
type WSExchangePrice struct {
	Name  string `json:"name"`
	Price string `json:"price"`
}
type WSExchangePriceRateListRst struct {
	ExchangePriceList []WSExchangePrice `json:"exchange_price_list"`
	Code              string            `json:"code"`
}

// GetWSExchangePriceRateList func
func GetWSExchangePriceRateList() WSExchangePriceRateListRst {

	exchangePriceList := make([]WSExchangePrice, 0)

	arrExchangeType := make([]string, 0)
	arrExchangeType = append(arrExchangeType, "SEC", "LIGA")

	for _, arrExchangeTypeV := range arrExchangeType {
		var exchangePrice string
		exchangePriceRst, _ := base.GetLatestExchangePriceMovementByTokenType(arrExchangeTypeV)
		if exchangePriceRst > 0 {
			exchangePrice = helpers.CutOffDecimalv2(exchangePriceRst, 2, ".", ",", true)
		}
		exchangePriceList = append(exchangePriceList,
			WSExchangePrice{
				Name:  arrExchangeTypeV,
				Price: exchangePrice,
			},
		)
	}
	arrDataReturn := WSExchangePriceRateListRst{
		ExchangePriceList: exchangePriceList,
		Code:              "exchange_price_list",
	}

	return arrDataReturn
}
