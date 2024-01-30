package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
)

//func GetMemberAnnouncementPopUpListv1 function
func GetMemberPoolListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	_, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	// member := u.(*models.EntMemberMembers)
	// fmt.Println("member:", member)
	// curDateString := base.GetCurrentTime("20060102") // correct vers.
	arrCond := make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: " t_bns_id LIKE ? ", CondValue: curDateString + "%"},
	// )

	arrTodayBonusPool, _ := models.GetTotalBonusPool(arrCond, false)
	eternityString := "0.00"
	cullinanString := "0.00"
	if arrTodayBonusPool.Eternity > 0 {
		// eternityString = helpers.Textify(arrTodayBonusPool.Eternity)
		eternityString = helpers.NumberFormatPhp(arrTodayBonusPool.Eternity, 2, ".", ",")
	}
	if arrTodayBonusPool.Cullinan > 0 {
		cullinanString = helpers.NumberFormatPhp(arrTodayBonusPool.Cullinan, 2, ".", ",")
	}

	arrDataReturn := make([]map[string]interface{}, 0)
	arrDataReturn = append(arrDataReturn,
		map[string]interface{}{
			"pool_name":   helpers.TranslateV2("eternity_pool", langCode, make(map[string]string)),
			"pool_amount": eternityString,
		},
		map[string]interface{}{
			"pool_name":   helpers.TranslateV2("cullinan_pool", langCode, make(map[string]string)),
			"pool_amount": cullinanString,
		})

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}
