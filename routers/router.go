package routers

import (
	"github.com/gin-gonic/gin"
	_ "github.com/smartblock/gta-api/docs"
	"github.com/smartblock/gta-api/middleware/api"
	"github.com/smartblock/gta-api/middleware/cors"
	"github.com/smartblock/gta-api/middleware/websocket"
	v1 "github.com/smartblock/gta-api/routers/api/v1"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	// "github.com/smartblock/gta-api/controllers/token"

	"github.com/smartblock/gta-api/controllers/koo"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	r := gin.New()
	r.POST("/koo/test", koo.KooTest)
	r.GET("/quik/test", koo.Testing)
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// r.Use(cors.Test1())
	r.Use(cors.Cors())
	// r.Static("/templates/sales/node/view/", "./docs/templates/sales/node")           // to serve the contract file
	// r.Static("/templates/sales/broadband/view/", "./docs/templates/sales/broadband") // to serve the contract file
	// r.Static("/member/sales/view/node/", "./docs/member/sales/node")                 // to serve the node contract file
	// r.Static("/member/sales/view/broadband/", "./docs/member/sales/broadband")       // to serve the broadband contract file
	// r.GET("/koo/test2", koo.KooTestCors)
	// r.Use(trans.SetLocale())
	// r.Use(api.CheckContentType())
	// r.Use(api.CheckHash())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// r.GET("/member/sales/download/:prdGroup/:filename", file.ServePDFFile)
	// r.GET("/member/sales/download/broadband/:filename", file.ServePDFFile)

	// v1 app api
	apiAppv1 := r.Group("/api/app/v1")
	{
		apiAppv1.Use(api.ApiKey())
		apiAppv1.Use(api.RouteAccessPermission())
		apiAppv1.Use(api.LogAppApiLog())
		v1.App(apiAppv1) // member api
	}

	// v1 auction api
	// apiAuction1 := r.Group("/api/auction/v1")
	// {
	// 	apiAuction1.Use(api.ApiKey())
	// 	apiAuction1.Use(api.RouteAccessPermission())
	// 	apiAuction1.Use(api.LogAuctionApiLog())
	// 	v1.Auction(apiAuction1) // blockchain api
	// }

	// v1 api
	apiv1 := r.Group("/api")
	{
		// apiv1.Use(api.ApiKey())
		// apiv1.Use(api.RouteAccessPermission())
		apiv1.Use(api.Log())
		// apiv1.Use(api.CheckIPWhitelist())
		v1.Api(apiv1) // api
	}

	// v1 web socket
	webSocketv1 := r.Group("/ws")
	{
		webSocketv1.Use(websocket.WSCorsChecking())
		v1.WebSocket(webSocketv1)
	}

	// v1 bc-admin api
	apiBCAdmin := r.Group("/api/bc-admin/v1")
	{
		apiBCAdmin.Use(api.ApiKey())
		apiBCAdmin.Use(api.RouteAccessPermission())
		v1.BCAdmin(apiBCAdmin) // bc-admin api
	}

	// v1 bc-admin api
	// apiLaliga := r.Group("/api/laliga/v1")
	// {
	// 	apiLaliga.Use(api.ApiKey())
	// 	apiLaliga.Use(api.RouteAccessPermission())
	// 	apiLaliga.Use(api.Log())
	// 	v1.Laliga(apiLaliga) // bc-admin api
	// }

	// v1 htmlfive api
	apiHtml5v1 := r.Group("/api/html5/v1")
	{
		apiHtml5v1.Use(api.ApiKey())
		apiHtml5v1.Use(api.RouteAccessPermission())
		apiHtml5v1.Use(api.LogHtml5ApiLog())
		v1.Html5(apiHtml5v1) // member api
	}

	return r
}
