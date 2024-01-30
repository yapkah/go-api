package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
)

// p.s. This middleware check IP White List
func CheckIPWhitelist() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int

		code = e.UNAUTHORIZED
		apiKey := c.GetHeader("X-Authorization")
		ip := c.ClientIP()

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"rst":  0,
				"msg":  e.GetMsg(code),
				"data": nil,
			})

			c.Abort()
			return
		}

		arrSetting, _ := models.GetSysGeneralSetupByID("ip_whitelist_setting")
		if arrSetting == nil {
			base.LogErrorLog("CheckIPWhitelist-general_setup_missing_ip_whitelist_setting", "ip_whitelist_setting", nil, true)
		}
		if arrSetting != nil && arrSetting.InputType1 == "1" {

			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " api_keys.key = ?", CondValue: apiKey},
				models.WhereCondFn{Condition: " api_keys.active = ?", CondValue: 1},
			)
			apiKeyRst, _ := models.GetApiKeysFn(arrCond, "", false)
			if arrSetting.InputValue1 != "" {
				arrIPWhitelist := make(map[string][]string, 0)
				err := json.Unmarshal([]byte(arrSetting.InputValue1), &arrIPWhitelist)
				if err != nil {
					base.LogErrorLog("CheckIPWhitelist-json_decode_arrIPWhitelist_failed", arrSetting.InputValue1, nil, true)
					c.JSON(http.StatusOK, gin.H{
						"rst":  0,
						"msg":  helpers.TranslateV2("something_went_wrong", "en", nil),
						"data": nil,
					})

					c.Abort()
					return
				}

				checkIPWhiteStatus := false
				// start checking on ip white list
				if len(apiKeyRst) > 0 {
					if len(arrIPWhitelist[apiKeyRst[0].Name]) > 0 {
						for _, routeNameV := range arrIPWhitelist[apiKeyRst[0].Name] {
							if routeNameV != "" {
								ipWhiteRst := strings.Contains(ip, routeNameV)
								if ipWhiteRst {
									checkIPWhiteStatus = true
									break
								}
							}
						}

						if !checkIPWhiteStatus {
							c.JSON(http.StatusUnauthorized, gin.H{
								"rst":  0,
								"msg":  helpers.TranslateV2("unauthorized", "en", nil),
								"data": nil,
							})

							c.Abort()
							return
						}
					}
				}
				// end checking on ip white list
			}
		}
		c.Next()
	}
}
