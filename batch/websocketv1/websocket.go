package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	//"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/gredis"
	"github.com/yapkah/go-api/pkg/logging"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/translation"
	wspkg "github.com/yapkah/go-api/pkg/websocket"
	"github.com/yapkah/go-api/routers/websocket"
	"github.com/yapkah/go-api/service/trading_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

// const sleepTime = 1 * time.Second

func init() {
	setting.Setup()
	models.Setup()
	logging.Setup()
	gredis.Setup()
	// util.Setup()
	translation.Setup()
	wspkg.Setup()
}

// @title Golang Gin API
// @version 1.0
// @description An example of gin
// @termsOfService https://github.com/yapkah/go-api
// @license.name MIT
// @license.url https://github.com/yapkah/go-api/blob/master/LICENSE
func main() {
	gin.SetMode(setting.ServerSetting.RunMode)

	go LoopBatch()

	routersInit := websocket.InitRouter()
	readTimeout := setting.ServerSetting.ReadTimeout
	writeTimeout := setting.ServerSetting.WriteTimeout
	endPoint := fmt.Sprintf("%v:%d", setting.ServerSetting.Domain, setting.ServerSetting.WSHttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	log.Printf("[info] start http server listening %s", endPoint)

	server.ListenAndServe()

	// If you want Graceful Restart, you need a Unix system and download github.com/fvbock/endless
	//endless.DefaultReadTimeOut = readTimeout
	//endless.DefaultWriteTimeOut = writeTimeout
	//endless.DefaultMaxHeaderBytes = maxHeaderBytes
	//server := endless.NewServer(endPoint, routersInit)
	//server.BeforeBegin = func(add string) {
	//	log.Printf("Actual pid is %d", syscall.Getpid())
	//}
	//
	//err := server.ListenAndServe()
	//if err != nil {
	//	log.Printf("Server err: %v", err)
	//}
}

func LoopBatch() {
	for {
		// start exchange_price
		exchangePriceRst := wallet_service.GetWSExchangePriceRateList()
		arrDataReturn := map[string]interface{}{
			"rst":  1,
			"msg":  "success",
			"data": exchangePriceRst,
		}
		encodedArrDataReturn, _ := json.Marshal(arrDataReturn)
		wspkg.BroadcastData("exchange_price", encodedArrDataReturn)
		// end exchange_price

		// start market_price
		arrData := trading_service.WSMemberTradingMarketListStruct{
			MemberID: 1,
			LangCode: "en",
		}
		tradingMarketRst := trading_service.GetWSMemberTradingMarketListvv2(arrData)
		arrDataReturn["rst"] = 1
		arrDataReturn["rst"] = "success"
		arrDataReturn["data"] = tradingMarketRst
		encodedArrDataReturn, _ = json.Marshal(arrDataReturn)
		wspkg.BroadcastData("market_price", encodedArrDataReturn)
		// end market_price

		arrCrypto := []string{"LIGA", "SEC"}
		for _, arrCryptoV := range arrCrypto {
			// start available_buy_market_price
			arrAvailableSECTradingBuy := trading_service.WSMemberAvailableTradingBuyListv1Struct{
				CryptoCode: arrCryptoV,
				LangCode:   "en",
			}
			availableSECTradingBuyRst := trading_service.GetWSMemberAvailableTradingBuyListv1(arrAvailableSECTradingBuy)
			arrDataReturn["rst"] = 1
			arrDataReturn["rst"] = "success"
			arrDataReturn["data"] = availableSECTradingBuyRst
			encodedArrDataReturn, _ = json.Marshal(arrDataReturn)
			wspkg.BroadcastData("available_buy_market_price", encodedArrDataReturn)
			// end available_buy_market_price

			// start available_sell_market_price
			arrAvailableSECTradingSell := trading_service.WSMemberAvailableTradingBuyListv1Struct{
				CryptoCode: arrCryptoV,
				LangCode:   "en",
			}
			availableSECTradingSellRst := trading_service.GetWSMemberAvailableTradingSellListv1(arrAvailableSECTradingSell)
			arrDataReturn["rst"] = 1
			arrDataReturn["rst"] = "success"
			arrDataReturn["data"] = availableSECTradingSellRst
			encodedArrDataReturn, _ = json.Marshal(arrDataReturn)
			wspkg.BroadcastData("available_sell_market_price", encodedArrDataReturn)
			// end available_sell_market_price
		}

		time.Sleep(1 * time.Second)
	}
}
