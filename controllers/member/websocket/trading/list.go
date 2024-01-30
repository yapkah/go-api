package trading

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	wspkg "github.com/yapkah/go-api/pkg/websocket"
	"github.com/yapkah/go-api/service/trading_service"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//func GetWSMemberAvailableTradingBuyListv1 function
// func GetWSMemberAvailableTradingBuyListv1(c *gin.Context) {

// 	w := c.Writer
// 	r := c.Request

// 	cors, _ := c.Get("cors_status")
// 	wsupgrader.CheckOrigin = func(r *http.Request) bool { return false }
// 	if cors == true {
// 		wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
// 	}

// 	conn, err := wsupgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		models.ErrorLog("GetWSMemberAvailableTradingBuyListv1-web_socket_failed", err.Error(), "Failed to set websocket upgrade")
// 		return
// 	}

// 	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
// 	if c.GetHeader("Accept-Language") != "" {
// 		langCode = c.GetHeader("Accept-Language")
// 	}

// 	// go wspkg.Writer(conn)

// 	type arrReq struct {
// 		Access     string `json:"access" form:"access"`
// 		CryptoCode string `json:"crypto_code" form:"crypto_code"`
// 	}

// 	msgDataByte, err := wspkg.ReadWSMsg(conn)

// 	if err != nil || msgDataByte == nil {
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_1", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	var arrInputReq arrReq
// 	err = json.Unmarshal(msgDataByte, &arrInputReq)
// 	if err != nil {
// 		models.ErrorLog("ReadWSMsg_error_in_json", err.Error(), nil)
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_2", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	if arrInputReq.Access == "" {
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_3", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	if arrInputReq.CryptoCode == "" {
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_4", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	// // accessToken := arrInputReq.Access
// 	// if arrInputReq.LangCode != "" {
// 	// 	langCode = arrInputReq.LangCode
// 	// }

// 	// u, ok := c.Get("access_user")
// 	// if !ok {
// 	// 	message := app.MsgStruct{
// 	// 		Msg: "invalid_member",
// 	// 	}
// 	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
// 	// 	return
// 	// }

// 	// member := u.(*models.EntMemberMembers)
// 	ch := time.Tick(1 * time.Second)
// 	go func(conn *websocket.Conn) {
// 		var skipRoomLive *trading_service.AvailableTradingPriceListRst
// 		var count int
// 		for range ch {
// 			// check user linked to clo
// 			// user, err := at.GetUser()
// 			// if err != nil || user == nil {
// 			// 	conn.WriteJSON(app.WSResponse{
// 			// 		Rst:  0,
// 			// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
// 			// 		Data: "",
// 			// 	})
// 			// 	conn.Close()
// 			// 	break
// 			// }
// 			// memberID = user.GetUserID()
// 			// platform := strings.ToLower(at.Platform)
// 			// start checking for htmlfive and app log log dt_expiry, token is expired
// 			// tokenRst := member_service.ProcessValidateToken(platform, ttoken, memberID)
// 			// if !tokenRst {
// 			// 	conn.WriteJSON(app.WSResponse{
// 			// 		Rst:  0,
// 			// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
// 			// 		Data: "",
// 			// 	})
// 			// 	conn.Close()
// 			// 	models.ErrorLog("ProcessValidateWSToken_ProcessValidateToken_UNAUTHORIZED2", "token_might_be_expired", nil) // need to comment out later
// 			// 	break
// 			// }
// 			// end checking for htmlfive and app log log dt_expiry, token is expired
// 			// fmt.Println(count, ". skipRoomLive", skipRoomLive)
// 			var result *trading_service.AvailableTradingPriceListRst
// 			// if strings.ToLower(arrInputReq.CryptoCode) == "sec" {
// 			arrData := trading_service.WSMemberAvailableTradingBuyListv1Struct{
// 				CryptoCode: arrInputReq.CryptoCode,
// 				LangCode:   langCode,
// 			}
// 			result = trading_service.GetWSMemberAvailableTradingBuyListv1(arrData)
// 			// } else if strings.ToLower(arrInputReq.CryptoCode) == "liga" {
// 			// 	result = trading_service.GetAvailableLigaTradingBuyList(arrInputReq.Quantitative, langCode)
// 			// }
// 			// result, err := room_service.GetRoomLiveStatusv1(memberID, roomTypeCode)

// 			// if len(result.AvailableTradingPriceList) > 0 {
// 			// 	sort.Slice(result.AvailableTradingPriceList, func(i, j int) bool {
// 			// 		return result.AvailableTradingPriceList[i].SeatNo < result.AvailableTradingPriceList[j].SeatNo
// 			// 	})
// 			// }

// 			var writeJSONStatus bool
// 			if reflect.DeepEqual(skipRoomLive, result) {
// 				// if !reflect.DeepEqual(skipRoomLive, result) {
// 				writeJSONStatus = true
// 			}

// 			if writeJSONStatus {
// 				conn.WriteJSON(app.WSResponse{
// 					Rst:  1,
// 					Msg:  helpers.TranslateV2("success", langCode, nil),
// 					Data: result,
// 				})
// 			}
// 			skipRoomLive = result
// 			count++
// 		}
// 	}(conn)
// }

//func GetWSMemberAvailableTradingSellListv1 function
// func GetWSMemberAvailableTradingSellListv1(c *gin.Context) {

// 	w := c.Writer
// 	r := c.Request

// 	cors, _ := c.Get("cors_status")
// 	wsupgrader.CheckOrigin = func(r *http.Request) bool { return false }
// 	if cors == true {
// 		wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
// 	}

// 	conn, err := wsupgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		models.ErrorLog("GetWSMemberAvailableTradingSellListv1-web_socket_failed", err.Error(), "Failed to set websocket upgrade")
// 		return
// 	}

// 	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
// 	if c.GetHeader("Accept-Language") != "" {
// 		langCode = c.GetHeader("Accept-Language")
// 	}

// 	// go wspkg.Writer(conn)

// 	type arrReq struct {
// 		Access     string `json:"access" form:"access"`
// 		CryptoCode string `json:"crypto_code" form:"crypto_code" `
// 	}

// 	msgDataByte, err := wspkg.ReadWSMsg(conn)

// 	if err != nil || msgDataByte == nil {
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_1", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	var arrInputReq arrReq
// 	err = json.Unmarshal(msgDataByte, &arrInputReq)
// 	if err != nil {
// 		models.ErrorLog("ReadWSMsg_error_in_json", err.Error(), nil)
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_2", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	if arrInputReq.Access == "" {
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_3", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	if arrInputReq.CryptoCode == "" {
// 		conn.WriteJSON(app.WSResponse{
// 			Rst:  0,
// 			Msg:  helpers.TranslateV2("failed_to_read_data_4", langCode, nil),
// 			Data: nil,
// 		})
// 		conn.Close()
// 		return
// 	}

// 	// // accessToken := arrInputReq.Access
// 	// if arrInputReq.LangCode != "" {
// 	// 	langCode = arrInputReq.LangCode
// 	// }

// 	// u, ok := c.Get("access_user")
// 	// if !ok {
// 	// 	message := app.MsgStruct{
// 	// 		Msg: "invalid_member",
// 	// 	}
// 	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
// 	// 	return
// 	// }

// 	// member := u.(*models.EntMemberMembers)
// 	ch := time.Tick(1 * time.Second)
// 	go func(conn *websocket.Conn) {
// 		var skipRoomLive *trading_service.AvailableTradingPriceListRst
// 		var count int
// 		for range ch {
// 			// check user linked to clo
// 			// user, err := at.GetUser()
// 			// if err != nil || user == nil {
// 			// 	conn.WriteJSON(app.WSResponse{
// 			// 		Rst:  0,
// 			// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
// 			// 		Data: "",
// 			// 	})
// 			// 	conn.Close()
// 			// 	break
// 			// }
// 			// memberID = user.GetUserID()
// 			// platform := strings.ToLower(at.Platform)
// 			// start checking for htmlfive and app log log dt_expiry, token is expired
// 			// tokenRst := member_service.ProcessValidateToken(platform, ttoken, memberID)
// 			// if !tokenRst {
// 			// 	conn.WriteJSON(app.WSResponse{
// 			// 		Rst:  0,
// 			// 		Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
// 			// 		Data: "",
// 			// 	})
// 			// 	conn.Close()
// 			// 	models.ErrorLog("ProcessValidateWSToken_ProcessValidateToken_UNAUTHORIZED2", "token_might_be_expired", nil) // need to comment out later
// 			// 	break
// 			// }
// 			// end checking for htmlfive and app log log dt_expiry, token is expired
// 			// fmt.Println(count, ". skipRoomLive", skipRoomLive)
// 			var result *trading_service.AvailableTradingPriceListRst
// 			// if strings.ToLower(arrInputReq.CryptoCode) == "sec" {
// 			arrData := trading_service.WSMemberAvailableTradingBuyListv1Struct{
// 				CryptoCode: arrInputReq.CryptoCode,
// 				LangCode:   langCode,
// 			}
// 			result = trading_service.GetWSMemberAvailableTradingSellListv1(arrData)
// 			// } else if strings.ToLower(arrInputReq.CryptoCode) == "liga" {
// 			// 	result = trading_service.GetAvailableLigaTradingBuyList(arrInputReq.Quantitative, langCode)
// 			// }
// 			// result, err := room_service.GetRoomLiveStatusv1(memberID, roomTypeCode)

// 			// if len(result.AvailableTradingPriceList) > 0 {
// 			// 	sort.Slice(result.AvailableTradingPriceList, func(i, j int) bool {
// 			// 		return result.AvailableTradingPriceList[i].SeatNo < result.AvailableTradingPriceList[j].SeatNo
// 			// 	})
// 			// }

// 			var writeJSONStatus bool
// 			if reflect.DeepEqual(skipRoomLive, result) {
// 				// if !reflect.DeepEqual(skipRoomLive, result) {
// 				writeJSONStatus = true
// 			}

// 			if writeJSONStatus {
// 				conn.WriteJSON(app.WSResponse{
// 					Rst:  1,
// 					Msg:  helpers.TranslateV2("success", langCode, nil),
// 					Data: result,
// 				})
// 			}
// 			skipRoomLive = result
// 			count++
// 		}
// 	}(conn)
// }

//func GetWSMemberTradingMarketListv1 function
func GetWSMemberTradingMarketListv1(c *gin.Context) {

	w := c.Writer
	r := c.Request

	cors, _ := c.Get("cors_status")
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return false }
	if cors == true {
		wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		models.ErrorLog("GetWSMemberTradingMarketListv1-web_socket_failed", err.Error(), "Failed to set websocket upgrade")
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

	claim, _ := util.ParseToken(arrInputReq.Access)

	if claim == nil {
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("unauthorized", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}

	// check token id in db
	at, err := models.GetAllStatusAccessTokenByID(claim.Id)
	if err != nil {
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("unauthorized", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}
	// check user linked to token
	user, err := at.GetUser()
	if err != nil {
		conn.WriteJSON(app.WSResponse{
			Rst:  0,
			Msg:  helpers.TranslateV2("unauthorized", langCode, nil),
			Data: nil,
		})
		conn.Close()
		return
	}
	memberID := user.GetMembersID()
	// u, ok := c.Get("access_user")
	// if !ok {
	// 	conn.WriteJSON(app.WSResponse{
	// 		Rst:  0,
	// 		Msg:  helpers.TranslateV2("invalid_member", langCode, nil),
	// 		Data: nil,
	// 	})
	// 	conn.Close()
	// 	return
	// }

	// member := u.(*models.EntMemberMembers)
	ch := time.Tick(5 * time.Second)
	go func(conn *websocket.Conn) {
		var skipRoomLive trading_service.WSMemberTradingMarketListRstStruct
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
			// // start checking for htmlfive and app log log dt_expiry, token is expired
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
			// fmt.Println(count, ". skipRoomLive", skipRoomLive)
			var result trading_service.WSMemberTradingMarketListRstStruct
			// if strings.ToLower(arrInputReq.CryptoCode) == "sec" {

			arrData := trading_service.WSMemberTradingMarketListStruct{
				MemberID: memberID,
				LangCode: langCode,
			}

			result = trading_service.GetWSMemberTradingMarketListv1(arrData)
			// } else if strings.ToLower(arrInputReq.CryptoCode) == "liga" {
			// 	result = trading_service.GetAvailableLigaTradingBuyList(arrInputReq.Quantitative, langCode)
			// }
			// result, err := room_service.GetRoomLiveStatusv1(memberID, roomTypeCode)

			// if len(result.AvailableTradingPriceList) > 0 {
			// 	sort.Slice(result.AvailableTradingPriceList, func(i, j int) bool {
			// 		return result.AvailableTradingPriceList[i].SeatNo < result.AvailableTradingPriceList[j].SeatNo
			// 	})
			// }

			var writeJSONStatus bool
			if reflect.DeepEqual(skipRoomLive, result) {
				// if !reflect.DeepEqual(skipRoomLive, result) {
				writeJSONStatus = true
			}

			if writeJSONStatus {
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			}
			skipRoomLive = result
			count++
		}
	}(conn)
}
