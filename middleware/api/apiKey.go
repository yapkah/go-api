package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/e"
)

// CheckScopeOr is jwt scope middleware
// contain one of the scope in scope param
// p.s. This middleware only work with JWT() middlware
func ApiKey() gin.HandlerFunc {
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

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "api_keys.key = ?", CondValue: apiKey},
			models.WhereCondFn{Condition: "api_keys.active = ?", CondValue: 1},
		)
		result, err := models.GetApiKeysFn(arrCond, "", false)

		// start exclusive for admin
		// if len(result) < 1 {
		// 	arrCond := make([]models.WhereCondFn, 0)
		// 	arrCond = append(arrCond,
		// 		models.WhereCondFn{Condition: "api_keys.name = 'admin-app' OR api_keys.name = ? ", CondValue: "admin"},
		// 		models.WhereCondFn{Condition: "api_keys.key = ?", CondValue: apiKey},
		// 		models.WhereCondFn{Condition: "api_keys.active = ?", CondValue: 1},
		// 	)
		// 	result, _ = models.GetApiKeysFn(arrCond, "", false)
		// }
		// end exclusive for admin
		if err != nil || len(result) < 1 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"rst":  0,
				"msg":  e.GetMsg(code),
				"data": nil,
			})

			c.Abort()
			return
		}
		c.Set("source", result[0].SourceID)
		c.Set("sourceName", result[0].Name)
		c.Set("prjID", result[0].ID)
		if result[0].Name == "admin" || result[0].Name == "admin-app" { // admin call api
			ip := c.ClientIP()

			arrData := models.AddApiKeyAdminEventsStruct{
				ApiKeyID:  result[0].ID,
				IpAddress: ip,
				Event:     route,
			}
			models.AddApiKeyAdminEvents(arrData)
		} else { // others platform call api
			ip := c.ClientIP()

			arrData := models.ApiKeyAccessEvents{
				ApiKeyID:  result[0].ID,
				IpAddress: ip,
				Url:       route,
			}
			models.AddApiKeyAccessEvents(arrData)
		}

		c.Next()
	}
}
