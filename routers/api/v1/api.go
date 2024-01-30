package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/controllers/api"
	middleware "github.com/smartblock/gta-api/middleware/api"
)

// Api func
func Api(route *gin.RouterGroup) {
	routeAccessPermissionGroup := route.Group("/")
	routeAccessPermissionGroup.Use(middleware.RouteAccessPermission())
	routeAccessPermissionGroup.Use(middleware.CheckIPWhitelist())
	{
		apiKeyGroup := routeAccessPermissionGroup.Group("/")
		apiKeyGroup.Use(middleware.ApiKey())
		{
			apiKeyGroup.POST("/transactionCallback", api.TransactionCallback)                              // transaction callback from blockchain
			apiKeyGroup.POST("/transactionCallbackBatch", api.TransactionCallbackBatch)                    // transaction callback from blockchain
			apiKeyGroup.POST("/blockchain/trans/save", api.ProcessSaveMemberBlockchainTransRecordsFromApi) // save transaction from blockchain

			// Trading
			v1Group := apiKeyGroup.Group("v1")
			{
				// trading
				tradingGroup := v1Group.Group("/trading")
				{
					tradingGroup.POST("/buy/request", api.ProcessAutoTradingBuyRequestv1)
					tradingGroup.POST("/sell/request", api.ProcessAutoTradingSellRequestv1)
					tradingGroup.GET("/price/list", api.GetPriceListApiv1)
					tradingGroup.GET("/open-order/list", api.GetOpenOrderListv1)
					tradingGroup.GET("/list", api.GetAutoTradingListv1)
					tradingGroup.POST("/request/cancel", api.ProcessCancelAutoTradingRequestv1)
				}

				// wallet
				walletGroup := v1Group.Group("/wallet")
				{
					walletGroup.GET("/balance", api.GetWalletBalanceListApiv1)
				}
			}

			// swarm
			// swarmGroup := apiKeyGroup.Group("/swarm")
			// {
			// 	swarmV1Group := swarmGroup.Group("/v1")
			// 	{
			// 		swarmV1Group.POST("/wallet/data", api.UpdateWalletDataApi)
			// 	}
			// }

			// apiKeyGroup.POST("decryptPrivateKey", api.DecryptMemberPrivateKey) // decrypt member private key
			// apiKeyGroup.POST("recalSlsMasterSpentUsdtAmt", api.RecalSlsMasterSpentUsdtAmt) // recalculate sls_master spent usdt
			// apiKeyGroup.POST("recalEwtExchangeSpentUsdtAmt", api.RecalEwtExchangeSpentUsdtAmt) // recalculate ewt_exchange spent usdt
		}

		route.POST("/callback-crypto-withdrawal/update", api.ProcessUpdateCryptoWithdrawalv1) // add frontend translation
		route.POST("/processCryptoReturn", api.ProcessCryptoReturn)
		route.POST("/processUpdateTranslationFrontendCSVFile", api.ProcessUpdateTranslationFrontendCSVFile) // update front-end translation word via csv

		route.POST("/processPN", api.ProcessPN) // update front-end translation word via csv
	}

	// nft
	route.GET("/nft/list", api.GetNftList)
}
