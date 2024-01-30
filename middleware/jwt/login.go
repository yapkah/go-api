package jwt

import (
	"github.com/gin-gonic/gin"
)

// CheckDuplicateLogin check if there is duplicated login
// p.s. This middleware only work with JWT() middlware
func CheckDuplicateLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var code int

		// code = e.UNAUTHORIZED

		// tc, tok := c.Get("token_claim")
		// u, cok := c.Get("access_user")

		// if tok && cok {
		// 	claim := tc.(*util.Claims)
		// 	user := u.(models.User)
		// 	tokenID := user.GetLoginTokenID()

		// 	if claim.Id == tokenID {
		// 		code = e.SUCCESS
		// 	}
		// }

		// if code != e.SUCCESS { // error return
		// 	c.JSON(http.StatusUnauthorized, gin.H{
		// 		"rst":     0,
		// 		"msg":        trans(c, e.GetMsg(code)),
		// 		"data":       "",
		// 	})

		// 	c.Abort()
		// 	return
		// }

		c.Next()
	}
}
