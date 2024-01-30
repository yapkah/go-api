package member

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
)

//func GetMemberBlockChainExplorerListv1 function
func GetMemberBlockChainExplorerListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	// u, ok := c.Get("access_user")
	// if !ok {
	// 	message := app.MsgStruct{
	// 		Msg: "invalid_member",
	// 	}
	// 	appG.ResponseV2(0, http.StatusUnauthorized, message, "")
	// 	return
	// }

	// member := u.(*models.EntMemberMembers)

	// member.EntMemberID = 7098
	// arrData := notification_service.MemberNotificationStruct{
	// 	// MemberID: member.EntMemberID,
	// 	LangCode: langCode,
	// }
	// arrMemberNotificationPopUpList := notification_service.GetMemberNotificationListv1(arrData)

	settingID := "blockchain_explorer_list"
	arrSettingRst, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil || arrSettingRst == nil {
		message := app.MsgStruct{
			Msg: "success",
		}
		appG.ResponseV2(1, http.StatusOK, message, make([]map[string]string, 0))
		return
	}

	if arrSettingRst.InputType1 != "1" {
		message := app.MsgStruct{
			Msg: "success",
		}
		appG.ResponseV2(1, http.StatusOK, message, make([]map[string]string, 0))
		return
	}

	type arrBlockchainExplorerListStruct []struct {
		BlockchainURL string `json:"blockchain_url"`
		Name          string `json:"name"`
		ImageURL      string `json:"image_url"`
		Desc          string `json:"desc"`
	}

	var arrBlockchainExplorerList arrBlockchainExplorerListStruct

	if arrSettingRst.InputValue1 != "" {
		err = json.Unmarshal([]byte(arrSettingRst.InputValue1), &arrBlockchainExplorerList)

		if err != nil {
			base.LogErrorLog("GetMemberBlockChainExplorerListv1-failed_to_decode_arrBlockchainExplorerList", err, arrSettingRst.InputValue1, true)
			message := app.MsgStruct{
				Msg: "success",
			}
			appG.ResponseV2(1, http.StatusOK, message, make([]map[string]string, 0))
			return
		}

		if len(arrBlockchainExplorerList) > 0 {
			for arrBlockchainExplorerListK, arrBlockchainExplorerListV := range arrBlockchainExplorerList {
				arrBlockchainExplorerList[arrBlockchainExplorerListK].Name = helpers.TranslateV2(arrBlockchainExplorerListV.Name, langCode, make(map[string]string))
				arrBlockchainExplorerList[arrBlockchainExplorerListK].Desc = helpers.TranslateV2(arrBlockchainExplorerListV.Desc, langCode, make(map[string]string))
			}
		}
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrBlockchainExplorerList)
}
