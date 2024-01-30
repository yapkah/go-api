package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	wspkg "github.com/yapkah/go-api/pkg/websocket"
	"github.com/yapkah/go-api/service/trading_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

const pingPeriod = 10 * time.Second

func ProcessWSMemberConnection(c *gin.Context) {
	w := c.Writer
	r := c.Request

	cors, _ := c.Get("cors_status")
	wspkg.Upgrader.CheckOrigin = func(r *http.Request) bool { return false }
	if cors == true {
		wspkg.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	conn, err := wspkg.Upgrader.Upgrade(w, r, nil)
	fmt.Println("err:", err)
	if err != nil {
		models.ErrorLog("GetWSMemberExchangePriceListv1-web_socket_failed", err.Error(), "Failed to set websocket upgrade")
		return
	}

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	langCode = c.GetHeader("Accept-Language")
	// }

	// type arrReq struct {
	// 	Access string `json:"access" form:"access"`
	// }

	// msgDataByte, err := wspkg.ReadWSMsg(conn)
	// if err != nil || msgDataByte == nil {
	// 	conn.WriteJSON(app.WSResponse{
	// 		Rst:  0,
	// 		Msg:  helpers.TranslateV2("failed_to_read_data_1", langCode, nil),
	// 		Data: nil,
	// 	})
	// 	conn.Close()
	// 	return
	// }

	// var arrInputReq arrReq
	// err = json.Unmarshal(msgDataByte, &arrInputReq)
	// if err != nil {
	// 	models.ErrorLog("ReadWSMsg_error_in_json", err.Error(), nil)
	// 	conn.WriteJSON(app.WSResponse{
	// 		Rst:  0,
	// 		Msg:  helpers.TranslateV2("failed_to_read_data_2", langCode, nil),
	// 		Data: nil,
	// 	})
	// 	conn.Close()
	// 	return
	// }

	// if arrInputReq.Access == "" {
	// 	conn.WriteJSON(app.WSResponse{
	// 		Rst:  0,
	// 		Msg:  helpers.TranslateV2("failed_to_read_data_3", langCode, nil),
	// 		Data: nil,
	// 	})
	// 	conn.Close()
	// 	return
	// }

	// wspkg.Connection([]string{"exchange_price", "market_price"}, conn)
	wspkg.Connection(conn)
	// wspkg.Connection("market_price", conn)
}

//func ProcessWSMemberConnectionV2 function
func ProcessWSMemberConnectionV2(c *gin.Context) {

	w := c.Writer
	r := c.Request

	cors, _ := c.Get("cors_status")
	// fmt.Println("cors:", cors)
	if cors == false {
		fmt.Println("here")
		return
	}
	wspkg.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := wspkg.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(r.Header)
		base.LogErrorLog("ProcessWSMemberConnectionV2-web_socket_failed", err.Error(), r.Header, true)
		return
	}
	fmt.Println("connected")
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	// go wspkg.Writer(conn)

	// fmt.Println(arrInputReq)

	// // accessToken := arrInputReq.Access
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
	// ch := time.Tick(1 * time.Second)
	go func(conn *websocket.Conn) {
		// 	var skipRoomLive *trading_service.AvailableTradingPriceListRst
		// 	var count int
		for {

			// _, msg, err := conn.ReadMessage()
			// if err != nil {
			// 	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			// 		log.Printf("error: %v", err)
			// 	}
			// 	break
			// }
			// code := string(msg)
			// fmt.Println("msg:", string(msg))

			type arrReq struct {
				Access     string `json:"access"`
				Code       string `json:"code"`
				CryptoCode string `json:"crypto_code"`
				PeriodCode string `json:"period_code"`
			}

			msgDataByte, err := wspkg.ReadWSMsg(conn)

			if err != nil || msgDataByte == nil {
				// base.LogErrorLog("ProcessWSMemberConnectionV2-ReadWSMsg_error_in_json", err.Error(), nil, true)
				conn.WriteJSON(app.WSResponse{
					Rst:  0,
					Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
					Data: nil,
				})
				conn.Close()
				return
			}
			// code := string(msgDataByte)
			var arrInputReq arrReq
			err = json.Unmarshal(msgDataByte, &arrInputReq)
			if err != nil {
				// base.LogErrorLog("ProcessWSMemberConnectionV2-Unmarshal_error_in_json", err.Error(), nil, true)
				conn.WriteJSON(app.WSResponse{
					Rst:  0,
					Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
					Data: nil,
				})
				conn.Close()
				return
			}

			if strings.ToLower(arrInputReq.Code) == "timer" {
				result := map[string]interface{}{
					"timer": map[string]interface{}{
						"interval": 2,
					},
					"code": "timer",
				}
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			}

			if arrInputReq.Access == "" {
				// base.LogErrorLog("ProcessWSMemberConnectionV2-missing_access_in_msg", arrInputReq, string(msgDataByte), true)
				conn.WriteJSON(app.WSResponse{
					Rst:  0,
					Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
					Data: nil,
				})
				conn.Close()
				return
			}

			if arrInputReq.Code == "" {
				// base.LogErrorLog("ProcessWSMemberConnectionV2-missing_code_in_msg", arrInputReq, string(msgDataByte), true)
				conn.WriteJSON(app.WSResponse{
					Rst:  0,
					Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
					Data: nil,
				})
				conn.Close()
				return
			}

			if strings.ToLower(arrInputReq.Code) == "exchange_price" {
				result := wallet_service.GetWSExchangePriceRateList()
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			} else if strings.ToLower(arrInputReq.Code) == "market_price" {
				arrData := trading_service.WSMemberTradingMarketListStruct{
					MemberID: 1,
					LangCode: "en",
				}
				result := trading_service.GetWSMemberTradingMarketListvv2(arrData)
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			} else if strings.ToLower(arrInputReq.Code) == "available_buy_market_price" {
				if arrInputReq.CryptoCode == "" {
					base.LogErrorLog("ProcessWSMemberConnectionV2-missing_crypto_code_available_buy_market_price_in_msg", arrInputReq, string(msgDataByte), true)
					conn.WriteJSON(app.WSResponse{
						Rst:  0,
						Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
						Data: nil,
					})
					conn.Close()
					return
				}
				arrData := trading_service.WSMemberAvailableTradingBuyListv1Struct{
					CryptoCode: arrInputReq.CryptoCode,
					LangCode:   langCode,
				}
				result := trading_service.GetWSMemberAvailableTradingSellListv2(arrData)
				result.Code = "available_buy_market_price"
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			} else if strings.ToLower(arrInputReq.Code) == "available_sell_market_price" {
				if arrInputReq.CryptoCode == "" {
					base.LogErrorLog("ProcessWSMemberConnectionV2-missing_crypto_code_available_sell_market_price_in_msg", arrInputReq, string(msgDataByte), true)
					conn.WriteJSON(app.WSResponse{
						Rst:  0,
						Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
						Data: nil,
					})
					conn.Close()
					return
				}
				arrData := trading_service.WSMemberAvailableTradingBuyListv1Struct{
					CryptoCode: arrInputReq.CryptoCode,
					LangCode:   langCode,
				}
				result := trading_service.GetWSMemberAvailableTradingBuyListv2(arrData)
				result.Code = "available_sell_market_price"
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			} else if strings.ToLower(arrInputReq.Code) == "trading_view" {
				if arrInputReq.CryptoCode == "" {
					base.LogErrorLog("ProcessWSMemberConnectionV2-missing_crypto_code_trading_view_in_msg", arrInputReq, string(msgDataByte), true)
					conn.WriteJSON(app.WSResponse{
						Rst:  0,
						Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
						Data: nil,
					})
					conn.Close()
					return
				}
				if arrInputReq.PeriodCode == "" {
					base.LogErrorLog("ProcessWSMemberConnectionV2-missing_PeriodCode_trading_view_in_msg", arrInputReq, string(msgDataByte), true)
					conn.WriteJSON(app.WSResponse{
						Rst:  0,
						Msg:  helpers.TranslateV2("something_went_wrong", langCode, nil),
						Data: nil,
					})
					conn.Close()
					return
				}
				arrData := trading_service.WSMemberExchangePriceTradingView{
					CryptoCode: arrInputReq.CryptoCode,
					PeriodCode: arrInputReq.PeriodCode,
					LangCode:   langCode,
				}
				result := trading_service.GetWSMemberExchangePriceTradingView(arrData)
				conn.WriteJSON(app.WSResponse{
					Rst:  1,
					Msg:  helpers.TranslateV2("success", langCode, nil),
					Data: result,
				})
			}
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
			// fmt.Println(count, ". skipRoomLive", skipRoomLive)
			// var result *trading_service.AvailableTradingPriceListRst
			// if strings.ToLower(arrInputReq.CryptoCode) == "sec" {
			// arrData := trading_service.WSMemberAvailableTradingBuyListv1Struct{
			// 	CryptoCode: arrInputReq.CryptoCode,
			// 	LangCode:   langCode,
			// }
			// result = trading_service.GetWSMemberAvailableTradingBuyListv1(arrData)
			// } else if strings.ToLower(arrInputReq.CryptoCode) == "liga" {
			// 	result = trading_service.GetAvailableLigaTradingBuyList(arrInputReq.Quantitative, langCode)
			// }
			// result, err := room_service.GetRoomLiveStatusv1(memberID, roomTypeCode)

			// if len(result.AvailableTradingPriceList) > 0 {
			// 	sort.Slice(result.AvailableTradingPriceList, func(i, j int) bool {
			// 		return result.AvailableTradingPriceList[i].SeatNo < result.AvailableTradingPriceList[j].SeatNo
			// 	})
			// }

			// var writeJSONStatus bool
			// if reflect.DeepEqual(skipRoomLive, result) {
			// if !reflect.DeepEqual(skipRoomLive, result) {
			// 	writeJSONStatus = true
			// }

			// if writeJSONStatus {
			// 	conn.WriteJSON(app.WSResponse{
			// 		Rst:  1,
			// 		Msg:  helpers.TranslateV2("success", langCode, nil),
			// 		Data: result,
			// 	})
			// }
			// skipRoomLive = result
			// count++
		}
	}(conn)

	dtNow := base.GetCurrentTime("2006-01-02 15:04:05")
	fmt.Println("dtNow:", dtNow)
	timer1 := time.NewTimer(60 * time.Second)
	select {
	case <-timer1.C:
		dtNow = base.GetCurrentTime("2006-01-02 15:04:05")
		fmt.Println("dtNow:", dtNow)
		// conn.Close()
	}
	// timer1 := time.NewTimer(5 * time.Second)
	// ticker := time.NewTicker(pingPeriod)
	// defer func() {
	// 	ticker.Stop()
	// 	conn.Close()
	// }()
}
