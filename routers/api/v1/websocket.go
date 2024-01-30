package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/controllers/member/websocket/exchange"
	"github.com/smartblock/gta-api/controllers/member/websocket/trading"
)

// WebSocket func
func WebSocket(route *gin.RouterGroup) {

	// Trading
	memberGroup := route.Group("/member")
	{
		memberGroup.GET("/exchange-price/list", exchange.GetWSMemberExchangePriceListv1)
		// Trading
		tradingGroup := memberGroup.Group("/trading")
		{
			tradingGroup.GET("/market/list", trading.GetWSMemberTradingMarketListv1)
			// tradingGroup.GET("/available-market-price/buy/list", trading.GetWSMemberAvailableTradingSellListv1)
			// tradingGroup.GET("/available-market-price/sell/list", trading.GetWSMemberAvailableTradingBuyListv1)
		}
	}
	// Trading
	memberGroupV2 := route.Group("/v2/member")
	{
		memberGroupV2.GET("/exchange-price/list", exchange.GetWSMemberExchangePriceListv2)
	}
}
