package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	//"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/gredis"
	"github.com/smartblock/gta-api/pkg/logging"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/translation"
	"github.com/smartblock/gta-api/routers"
)

func init() {
	setting.Setup()
	models.Setup()
	logging.Setup()
	gredis.Setup()
	// util.Setup()
	translation.Setup()
}

// @title Golang Gin API
// @version 1.0
// @description An example of gin
// @termsOfService https://github.com/smartblock/gta-api
// @license.name MIT
// @license.url https://github.com/smartblock/gta-api/blob/master/LICENSE
func main() {
	gin.SetMode(setting.ServerSetting.RunMode)

	routersInit := routers.InitRouter()
	readTimeout := setting.ServerSetting.ReadTimeout
	writeTimeout := setting.ServerSetting.WriteTimeout
	endPoint := fmt.Sprintf("%v:%d", setting.ServerSetting.Domain, setting.ServerSetting.HttpPort)
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
