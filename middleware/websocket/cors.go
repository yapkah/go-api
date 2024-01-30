package websocket

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/setting"
)

// var wsupgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// p.s. This middleware only work with member htmlfive and app api middlware
func WSCorsChecking() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsStatus := false

		// if len(c.Request.Header["Sec-Websocket-Protocol"]) > 0 {
		// fmt.Println("header:", c.Request.Header)s
		// fmt.Println("")
		// fmt.Println("header Sec-Websocket-Protocol:", c.Request.Header["Sec-Websocket-Protocol"])
		// fmt.Printf("header Sec-Websocket-Key: type%T\n", c.Request.Header["Sec-Websocket-Key"][0])
		// webSocketKeyList := c.Request.Header["Sec-Websocket-Protocol"]
		// fmt.Println(webSocketKeyList)
		// fmt.Println(reflect.TypeOf(webSocketKeyList))
		// arrWebSocketKeyList := strings.Split(webSocketKeyList[0], ",")
		// for _, webSocketKeyListV := range webSocketKeyList {
		// 	// fmt.Println("arrWebSocketKeyListV:", arrWebSocketKeyListV)
		// 	// fmt.Println(webSocketKeyListK, "webSocketKeyListV:", webSocketKeyListV)
		// 	apiKey := helpers.SpaceStringsBuilder(arrWebSocketKeyListV)
		// 	arrCond := make([]models.WhereCondFn, 0)
		// 	arrCond = append(arrCond,
		// 		models.WhereCondFn{Condition: "api_keys.key = ?", CondValue: apiKey},
		// 		models.WhereCondFn{Condition: "api_keys.active = ?", CondValue: 1},
		// 	)
		// 	result, _ := models.GetApiKeysFn(arrCond, "", false)
		// 	if len(result) > 0 {
		// 		corsStatus = true
		// 		break
		// 	}
		// }
		// }

		if !corsStatus {
			fmt.Println("corsStatus:", corsStatus)
			// 	// return
		}
		corsDomainString := setting.Cfg.Section("cors").Key("CorsDomain").String()
		fmt.Println("header Origin:", c.Request.Header)
		if corsDomainString != "" {
			if len(c.Request.Header["Origin"]) > 0 {
				fmt.Println("header Origin:", c.Request.Header["Origin"][0])
				fmt.Printf("header Origin: type%T\n", c.Request.Header["Origin"][0])

				// perform bypass by Origin
				if !strings.Contains(corsDomainString, c.Request.Header["Origin"][0]) {
					models.ErrorLog("WSCorsChecking_invalid_cors", "cors_is_not_set", c.Request.Header["Origin"][0])
					corsStatus = false
				}
			}
		}
		// fmt.Println("corsStatus:", corsStatus)
		// c.Set("cors_status", corsStatus)
		c.Set("cors_status", true)
	}
}
