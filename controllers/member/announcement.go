package member

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/service/announcement_service"
)

// MemberAnnouncementForm struct
type MemberAnnouncementForm struct {
	Popup        string `form:"popup" json:"popup"`
	CategoryCode string `form:"category_code" json:"category_code"`
	Page         int64  `form:"page" json:"page"`
	ViewType     string `form:"view_type" json:"view_type"`
	Type         string `form:"type" json:"type"`
}

//func GetMemberAnnouncementPopUpListv1 function
func GetMemberAnnouncementListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberAnnouncementForm
		arrDataReturn interface{}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	if form.Page < 1 {
		form.Page = 1
	}

	if form.Type == "" {
		form.Type = "MAIN"
	}

	if strings.ToLower(form.ViewType) == "land" {
		// member.EntMemberID = 7098
		arrData := announcement_service.MemberAnnouncementLandStruct{
			MemberID:     member.EntMemberID,
			LangCode:     langCode,
			PopUp:        form.Popup,
			CategoryCode: form.CategoryCode,
			Type:         form.Type,
		}

		arrDataReturn = announcement_service.GetMemberAnnouncementLandv1(arrData)

	} else {
		arrData := announcement_service.MemberAnnouncementPaginateStruct{
			MemberID:     member.EntMemberID,
			LangCode:     langCode,
			PopUp:        form.Popup,
			CategoryCode: form.CategoryCode,
			Page:         form.Page,
			Type:         form.Type,
		}

		arrDataReturn = announcement_service.GetMemberAnnouncementPaginateListv1(arrData)
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}
