package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
)

// CheckScopeOr is jwt scope middleware
// contain one of the scope in scope param
// p.s. This middleware only work with JWT() middlware
func CheckActiveMember() gin.HandlerFunc {
	return func(c *gin.Context) {

		u, _ := c.Get("access_user")
		members := u.(*models.EntMemberMembers)
		activeMember := strings.ToUpper(members.EntMemberStatus)

		if activeMember == "" {
			c.JSON(http.StatusOK, gin.H{
				"rst":  1000,
				"msg":  "please_create_an_account",
				"data": nil,
			})

			c.Abort()
			return
		} else if activeMember == "I" {
			c.JSON(http.StatusOK, gin.H{
				"rst":  1001,
				"msg":  "please_activate_your_account",
				"data": nil,
			})

			c.Abort()
			return
		} else if activeMember == "S" {
			c.JSON(http.StatusOK, gin.H{
				"rst":  0,
				"msg":  "invalid_member",
				"data": nil,
			})

			c.Abort()
			return
		}

		// if members.PrivateKey == "" {
		// 	c.JSON(http.StatusOK, gin.H{
		// 		"rst":  1000,
		// 		"msg":  "please_activate_your_account",
		// 		"data": "",
		// 	})

		// 	c.Abort()
		// 	return
		// }

		c.Next()
	}
}
