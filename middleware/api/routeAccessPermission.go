package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
)

// p.s. This middleware check route access permission
func RouteAccessPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int

		code = e.UNAUTHORIZED
		apiKey := c.GetHeader("X-Authorization")

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"rst":  0,
				"msg":  e.GetMsg(code),
				"data": nil,
			})

			c.Abort()
			return
		}

		route := c.Request.URL.String()
		arrSetting, err := models.GetSysGeneralSetupByID("route_access_setting")

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		if arrSetting == nil || errMsg != "" {
			base.LogErrorLogV2("RouteAccessPermission-general_setup_missing_route_access_setting", "route_access_setting", errMsg, true, "koobot")
		}

		if arrSetting != nil && arrSetting.InputType1 == "1" {

			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " api_keys.key = ?", CondValue: apiKey},
				models.WhereCondFn{Condition: " api_keys.active = ?", CondValue: 1},
			)
			apiKeyRst, _ := models.GetApiKeysFn(arrCond, "", false)

			checkUnblockRouteAccess := false
			if arrSetting.InputValue1 != "" {
				arrBlockRouteAccess := make(map[string][]string, 0)
				err := json.Unmarshal([]byte(arrSetting.InputValue1), &arrBlockRouteAccess)
				if err != nil {
					base.LogErrorLog("RouteAccessPermission-json_decode_arrBlockRouteAccessRoute_failed", arrSetting.InputValue1, nil, true)
					c.JSON(http.StatusOK, gin.H{
						"rst":  0,
						"msg":  helpers.TranslateV2("something_went_wrong", "en", nil),
						"data": nil,
					})

					c.Abort()
					return
				}

				// start checking on block route list
				if len(apiKeyRst) > 0 {
					if len(arrBlockRouteAccess[apiKeyRst[0].Name]) > 0 {
						for _, routeNameV := range arrBlockRouteAccess[apiKeyRst[0].Name] {
							if routeNameV != "" {
								blockRouteAccessRst := strings.Contains(route, routeNameV)
								if blockRouteAccessRst {
									checkUnblockRouteAccess = true
									break
									// c.JSON(http.StatusOK, gin.H{
									// 	"rst":  0,
									// 	"msg":  helpers.TranslateV2("no_permission_to_process", "en", nil),
									// 	"data": nil,
									// })

									// c.Abort()
									// return
								}
							}
						}
					}
				}
				// end checking on block route list
			}

			if arrSetting.SettingValue1 != "" && checkUnblockRouteAccess {
				arrUnblockAccessRoute := make(map[string][]string, 0)
				err := json.Unmarshal([]byte(arrSetting.SettingValue1), &arrUnblockAccessRoute)
				if err != nil {
					base.LogErrorLog("RouteAccessPermission-json_decode_arrBlockRouteAccessRoute_failed", arrSetting.SettingValue1, nil, true)
					c.JSON(http.StatusOK, gin.H{
						"rst":  0,
						"msg":  helpers.TranslateV2("something_went_wrong", "en", nil),
						"data": nil,
					})

					c.Abort()
					return
				}

				// start checking on unblock route list
				if len(apiKeyRst) > 0 {
					if len(arrUnblockAccessRoute[apiKeyRst[0].Name]) > 0 {
						unblockRouteAccess := false
						for _, routeNameV := range arrUnblockAccessRoute[apiKeyRst[0].Name] {
							if routeNameV != "" {
								unblockRouteAccessRst := strings.Contains(route, routeNameV)
								if unblockRouteAccessRst {
									unblockRouteAccess = true
									break
								}
							}
						}

						if !unblockRouteAccess {
							c.JSON(http.StatusUnauthorized, gin.H{
								"rst":  0,
								"msg":  e.GetMsg(code),
								"data": nil,
							})

							c.Abort()
							return
						}
					}
				}
				// end checking on unblock route list
			}
		}
		c.Next()
	}
}
