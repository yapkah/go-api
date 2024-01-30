package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/controllers/member/websocket"
	wsmiddleware "github.com/smartblock/gta-api/middleware/websocket"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	r := gin.New()

	// v1 web socket
	r.Use(wsmiddleware.WSCorsChecking())
	r.GET("/ws/v1/member/connection", websocket.ProcessWSMemberConnection)
	r.GET("/ws/v2/member/connection", websocket.ProcessWSMemberConnectionV2)
	// webSocketv1 := r.Group("")
	// {
	// 	// webSocketv1.Use(wsmiddleware.WSCorsChecking())

	// 	memberGroupV2 := webSocketv1.Group("")
	// 	{
	// 		memberGroupV2.GET("/ws/v1/member/connection", websocket.ProcessWSMemberConnection)
	// 	}
	// }

	return r
}
