package wallet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/service/wallet_service"
)

// GetWSExchangePriceRateList func
func GetWSExchangePriceRateList(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	ok := models.ExistLangague(c.GetHeader("Accept-Language"))
	// 	if ok {
	// 		langCode = c.GetHeader("Accept-Language")
	// 	}
	// }

	// u, ok := c.Get("access_user")
	// if !ok {
	// 	message := app.MsgStruct{
	// 		Msg: "invalid_member",
	// 	}
	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, "")
	// 	return
	// }

	// member := u.(*models.EntMemberMembers)

	arrWSExchangePriceRateList := wallet_service.GetWSExchangePriceRateList()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrWSExchangePriceRateList)
}
