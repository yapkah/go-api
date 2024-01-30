package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/service/event_service"
)

//func GetMemberNotificationPopUpListv1 function
func GetMemberEventListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	arrData := event_service.MemberEventStruct{
		MemberID: member.ID,
		LangCode: langCode,
	}
	arrMemberEventList := event_service.GetMemberEventListv1(arrData)

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrMemberEventList)
}
