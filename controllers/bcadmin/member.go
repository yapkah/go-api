package bcadmin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/service/member_service"
)

// SearchMemberInfoForBCAdminForm struct
type SearchMemberInfoForBCAdminForm struct {
	NickName string `form:"nick_name" json:"nick_name" valid:"Required;"`
}

//func SearchMemberInfoForBCAdmin function
func SearchMemberInfoForBCAdmin(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form SearchMemberInfoForBCAdminForm
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	langCode = c.GetHeader("Accept-Language")
	// }

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	result, err := member_service.ProcessGetBCAdminMemberInfo(form.NickName)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}
	arrDataReturn := map[string]interface{}{
		"member_pk_list": result, // this is refer to ent_member.id
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}
