package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/service/member_service"
)

// Check Custom Member Access for the Login action and Operation
func MemberAccessPermission() gin.HandlerFunc {
	return func(c *gin.Context) {

		u, _ := c.Get("access_user")
		members := u.(*models.EntMemberMembers)

		verificationRst := member_service.CheckMemberAccessPermission(members.EntMemberID)
		if !verificationRst {
			c.JSON(http.StatusUnauthorized, gin.H{
				"rst":  0,
				"msg":  e.GetMsg(e.UNAUTHORIZED),
				"data": nil,
			})

			c.Abort()
			return
		}

		c.Next()
	}
}
