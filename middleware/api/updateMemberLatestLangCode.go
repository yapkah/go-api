package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/service/language_service"
)

// p.s. This middleware check manage the member language by device
func UpdateMemberLatestLangCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		var langCode string

		prjIDInterface, _ := c.Get("prjID")
		prjID := prjIDInterface.(int)
		tokenInterface, _ := c.Get("token")
		token := tokenInterface.(string)
		sourceInterface, _ := c.Get("source")
		source := sourceInterface.(int)

		if prjID > 0 {
			if c.GetHeader("Accept-Language") != "" {
				ok := models.ExistLangague(c.GetHeader("Accept-Language"))
				if ok {
					langCode = c.GetHeader("Accept-Language")
				}
			}

			route := c.Request.URL.String()
			platformCheckingRst := strings.Contains(route, "/api/app")
			platform := "htmlfive"
			if platformCheckingRst {
				platform = "app"
			}

			u, ok := c.Get("access_user")
			if !ok {
				fmt.Println("here")
				base.LogErrorLog("UpdateMemberLatestLangCode-failed_to_get_access_user_globally", "invalid_member", "no_process_latest_langCode_must_use_update_langcode_api", true)
				c.Next()
				return
			}

			member := u.(*models.EntMemberMembers)

			entMemberID := member.EntMemberID

			// fmt.Println("langCode:", langCode)
			// fmt.Println("token:", token)
			if langCode != "" {

				// start update member language code in login log table
				if token != "" {
					arrData := language_service.ProcessUpdateMemberDeviceLanguagev1Struct{
						AccessToken: token,
						LangCode:    langCode,
						SourceID:    source,
						PrjID:       prjID,
						Platform:    platform,
					}
					// fmt.Println("arrData:", arrData)
					os, pushNotiToken, _ := language_service.ProcessUpdateMemberDeviceLanguagev1(arrData)
					// fmt.Println("os:", os)
					if os != "" && pushNotiToken != "" {
						// begin transaction
						tx := models.Begin()

						// start process group subscription

						groupName := "LANG_" + strings.ToUpper(langCode) + "-" + strconv.Itoa(prjID)
						arrProcessPnData := base.ProcessMemberPushNotificationGroupStruct{
							GroupName: groupName,
							Os:        os,
							MemberID:  entMemberID,
							RegID:     pushNotiToken,
							PrjID:     prjID,
							SourceID:  source,
						}
						base.ProcessMemberPushNotificationGroup(tx, "removeAllIndPrevLangCodeRegID", arrProcessPnData)
						// end process group subscription

						models.Commit(tx)
					}
				}
				// end update member language code in login log table

				// langGroup := "LANG_" + strings.ToUpper(langCode)
				// arrCond := make([]models.WhereCondFn, 0)
				// arrCond = append(arrCond,
				// 	models.WhereCondFn{Condition: "group_name = ?", CondValue: langGroup},
				// 	models.WhereCondFn{Condition: "push_noti_token = ?", CondValue: arrData.RegID},
				// )
				// arrExistingAppMemPnGrp, _ := models.GetAppMemberPnGroupFn(arrCond, false)

				// start insert new pn group into app_member_pn_group
				// arrCrtSubPN := models.AppMemberPnGroup{
				// 	GroupName: langGroup,
				// 	PrjID:     prjID,
				// OS:            arrData.Os,
				// PushNotiToken: arrData.RegID,
				// }

				// if arrData.MemberID > 0 {
				// 	arrCrtSubPN.MemberID = arrData.MemberID
				// }
				// _, err := models.AddAppMemberPnGroup(arrCrtSubPN)
				// if err != nil {
				// 	models.Rollback(tx)
				// 	base.LogErrorLog("UpdateMemberLatestLangCode-AddAppMemberPnGroup_failed", err.Error(), arrCrtSubPN, true)
				// }
				// models.Commit(tx)
				// end insert new pn group into app_member_pn_group
			}
		}

		c.Next()
	}
}
