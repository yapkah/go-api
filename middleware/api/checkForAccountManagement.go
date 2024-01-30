package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
)

// p.s. This middleware checking on devices vs acc management
func CheckForAccountManagement() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var err error
		route := c.Request.URL.String()
		platformCheckingRst := strings.Contains(route, "/api/app")
		var appPlatform bool
		if platformCheckingRst {
			appPlatform = true
		}

		if appPlatform {

			arrValidationURLList := []string{
				"/api/app/v1/member/withdrawal",
				"/api/app/v1/member/exchange",
				"/api/app/v1/member/product",
			}

			// arrCond := make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: " b_latest = ?", CondValue: 1},
			// 	models.WhereCondFn{Condition: " platform = ?", CondValue: platform},
			// )

			// arrLatestAppAppVersion, _ := models.GetLatestAppAppVersionFn(arrCond, false)

			// start perform checking the url need to monitor the account is valid or not
			if arrValidationURLList != nil {
				skipUrlStatus := helpers.StringInSlice(route, arrValidationURLList)
				if skipUrlStatus {
					// start checking on monitor the account
					// c.JSON(http.StatusOK, gin.H{
					// 	"rst":  0,
					// 	"msg":  helpers.TranslateV2("member_maintenance_msg", "en", make(map[string]string)),
					// 	"data": "",
					// })
					// c.Abort()
					// return
					// end checking on monitor the account
				}
			}
			// end perform checking the url need to monitor the account is valid or not
		}

		c.Next()
	}
}
