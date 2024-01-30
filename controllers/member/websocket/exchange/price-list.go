package exchange

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	wspkg "github.com/smartblock/gta-api/pkg/websocket"
	"github.com/smartblock/gta-api/service/wallet_service"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//func GetWSMemberExchangePriceListv1 function
func GetWSMemberExchangePriceListv1(c *gin.Context) {
	w := c.Writer
	r := c.Request

	cors, _ := c.Get("cors_status")
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return false }
	if cors == true {
		wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		models.ErrorLog("GetWSMemberExchangePriceListv1-web_socket_failed", err.Error(), "Failed to set websocket upgrade")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	type arrReq struct {
		Access string `json:"access" form:"access"`
	}

	msgDataByte, err := wspkg.ReadWSMsg(conn)
	if err != nil || msgDataByte == nil {
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("failed_to_read_data_1", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}

	var arrInputReq arrReq
	err = json.Unmarshal(msgDataByte, &arrInputReq)
	if err != nil {
		models.ErrorLog("ReadWSMsg_error_in_json", err.Error(), nil)
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("failed_to_read_data_2", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}

	if arrInputReq.Access == "" {
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("failed_to_read_data_3", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}

	// accessToken := arrInputReq.Access
	// if arrInputReq.LangCode != "" {
	// 	langCode = arrInputReq.LangCode
	// }

	// u, ok := c.Get("access_user")
	// if !ok {
	// 	message := app.MsgStruct{
	// 		Msg: "invalid_member",
	// 	}
	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
	// 	return
	// }

	// member := u.(*models.EntMemberMembers)
	ch := time.Tick(5 * time.Second)
	go func(conn *websocket.Conn) {
		var skipLive *wallet_service.WSExchangePriceRateListRst
		var count int
		for range ch {
			// check user linked to clo
			// user, err := at.GetUser()
			// if err != nil || user == nil {
			// 	conn.WriteJSON(app.WSResponse{
			// 		Rst:  0,
			// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
			// 		Data: "",
			// 	})
			// 	conn.Close()
			// 	break
			// }
			// memberID = user.GetUserID()
			// platform := strings.ToLower(at.Platform)
			// start checking for htmlfive and app log log dt_expiry, token is expired
			// tokenRst := member_service.ProcessValidateToken(platform, ttoken, memberID)
			// if !tokenRst {
			// 	conn.WriteJSON(app.WSResponse{
			// 		Rst:  0,
			// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
			// 		Data: "",
			// 	})
			// 	conn.Close()
			// 	models.ErrorLog("ProcessValidateWSToken_ProcessValidateToken_UNAUTHORIZED2", "token_might_be_expired", nil) // need to comment out later
			// 	break
			// }
			// end checking for htmlfive and app log log dt_expiry, token is expired
			// fmt.Println(count, ". skipLive", skipLive)
			result := wallet_service.GetWSExchangePriceRateList()

			// if len(result.AvailableTradingPriceList) > 0 {
			// 	sort.Slice(result.AvailableTradingPriceList, func(i, j int) bool {
			// 		return result.AvailableTradingPriceList[i].SeatNo < result.AvailableTradingPriceList[j].SeatNo
			// 	})
			// }

			var writeJSONStatus bool
			if !reflect.DeepEqual(skipLive, result) {
				writeJSONStatus = true
			}

			if writeJSONStatus {
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			}
			count++
		}
	}(conn)
}

//func GetWSMemberExchangePriceListv1 function
func GetWSMemberExchangePriceListv2(c *gin.Context) {
	// hub := NewHub()
	go h.run()

	w := c.Writer
	r := c.Request

	cors, _ := c.Get("cors_status")
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return false }
	if cors == true {
		wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		models.ErrorLog("GetWSMemberExchangePriceListv1-web_socket_failed", err.Error(), "Failed to set websocket upgrade")
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	type arrReq struct {
		Access string `json:"access" form:"access"`
	}

	msgDataByte, err := wspkg.ReadWSMsg(conn)
	if err != nil || msgDataByte == nil {
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("failed_to_read_data_1", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}

	var arrInputReq arrReq
	err = json.Unmarshal(msgDataByte, &arrInputReq)
	if err != nil {
		models.ErrorLog("ReadWSMsg_error_in_json", err.Error(), nil)
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("failed_to_read_data_2", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}

	if arrInputReq.Access == "" {
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("failed_to_read_data_3", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}

	wsConn := &connection{
		send: make(chan []byte, 256),
		ws:   conn,
	}
	roomId := "1"
	s := subscription{wsConn, roomId}
	h.register <- s
	go s.writePump()
	go s.readPump()

	// accessToken := arrInputReq.Access
	// if arrInputReq.LangCode != "" {
	// 	langCode = arrInputReq.LangCode
	// }

	// u, ok := c.Get("access_user")
	// if !ok {
	// 	message := app.MsgStruct{
	// 		Msg: "invalid_member",
	// 	}
	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
	// 	return
	// }

	// member := u.(*models.EntMemberMembers)
	// ch := time.Tick(5 * time.Second)
	// go func(conn *websocket.Conn) {
	// 	var skipLive *wallet_service.WSExchangePriceRateListRst
	// 	var count int
	// 	for range ch {
	// check user linked to clo
	// user, err := at.GetUser()
	// if err != nil || user == nil {
	// 	conn.WriteJSON(app.WSResponse{
	// 		Rst:  0,
	// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
	// 		Data: "",
	// 	})
	// 	conn.Close()
	// 	break
	// }
	// memberID = user.GetUserID()
	// platform := strings.ToLower(at.Platform)
	// start checking for htmlfive and app log log dt_expiry, token is expired
	// tokenRst := member_service.ProcessValidateToken(platform, ttoken, memberID)
	// if !tokenRst {
	// 	conn.WriteJSON(app.WSResponse{
	// 		Rst:  0,
	// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
	// 		Data: "",
	// 	})
	// 	conn.Close()
	// 	models.ErrorLog("ProcessValidateWSToken_ProcessValidateToken_UNAUTHORIZED2", "token_might_be_expired", nil) // need to comment out later
	// 	break
	// }
	// end checking for htmlfive and app log log dt_expiry, token is expired
	// fmt.Println(count, ". skipLive", skipLive)
	// result := wallet_service.GetWSExchangePriceRateList()

	// if len(result.AvailableTradingPriceList) > 0 {
	// 	sort.Slice(result.AvailableTradingPriceList, func(i, j int) bool {
	// 		return result.AvailableTradingPriceList[i].SeatNo < result.AvailableTradingPriceList[j].SeatNo
	// 	})
	// }

	// 		var writeJSONStatus bool
	// 		if !reflect.DeepEqual(skipLive, result) {
	// 			writeJSONStatus = true
	// 		}

	// 		if writeJSONStatus {
	// 			conn.WriteJSON(app.WSResponse{
	// 				Rst:  1,
	// 				Msg:  helpers.TranslateV2("success", langCode, nil),
	// 				Data: result,
	// 			})
	// 		}
	// 		count++
	// 	}
	// }(conn)
}
