package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/download_service"
)

// MemberDownloadForm struct
type MemberDownloadForm struct {
	CategoryCode string `form:"category_code" json:"category_code"`
	Type         string `form:"type" json:"type"`
}

//func GetMemberDownloadListv1 function
func GetMemberDownloadListv1(c *gin.Context) {

	var (
		appG          = app.Gin{C: c}
		form          MemberDownloadForm
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

	if form.Type == "" {
		form.Type = "MAIN"
	}

	download := download_service.GetMemberDownloadListStruct{
		MemberID:     member.EntMemberID,
		LangCode:     langCode,
		CategoryCode: form.CategoryCode,
		Type:         form.Type,
	}

	arrDataReturn = download.GetMemberDownloadListv1()

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}
