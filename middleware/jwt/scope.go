package jwt

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/util"
)

// CheckScopeOr is jwt scope middleware
// contain one of the scope in scope param
// p.s. This middleware only work with JWT() middlware
func CheckScopeOr(scope []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int

		code = e.UNAUTHORIZED
		tc, ok := c.Get("token_claim")
		if ok {
			claim := tc.(*util.Claims)
			if claim != nil {
				// check scope (in_array)
				for _, v := range claim.Scope {
					for _, s := range scope {
						if v == s {
							code = e.SUCCESS
							break
						}
					}
				}
			}
		}

		if code != e.SUCCESS { // error return
			c.JSON(http.StatusUnauthorized, gin.H{
				"rst":  0,
				"msg":  trans(c, e.GetMsg(code)),
				"data": nil,
			})

			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckScopeAnd is jwt scope middleware
// token must contain all of the scope provided
// p.s. This middleware only work with JWT() middlware
func CheckScopeAnd(scope []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		count := 0

		code = e.UNAUTHORIZED
		tc, ok := c.Get("token_claim")
		if ok {
			claim := tc.(*util.Claims)
			if claim != nil {
				// check scope (in_array)
				for _, v := range claim.Scope {
					for _, s := range scope {
						if v == s {
							count++
						}
					}
				}
			}
		}

		if len(scope) == count {
			code = e.SUCCESS
		}

		if code != e.SUCCESS { // error return
			c.JSON(http.StatusUnauthorized, gin.H{
				"rst":  0,
				"msg":  e.GetMsg(code),
				"data": nil,
			})

			c.Abort()
			return
		}

		c.Next()
	}
}
