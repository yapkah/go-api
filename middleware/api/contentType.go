package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/translation"
)

// disable route name
var disable = []string{
	"/api/v1/member/profile/photo",
	"/api/v1/admin/translation/updatebyfile",
	"/api/v1/admin/app/version",
}

// CheckContentType is jwt scope middleware
func CheckContentType() gin.HandlerFunc {
	return func(c *gin.Context) {

		for _, url := range disable {
			if url == c.Request.URL.String() {
				c.Next()
				return
			}
		}

		code := e.UNSUPPORTED_MEDIA_TYPE

		// get header token
		ctype := c.GetHeader("Content-Type")

		split := strings.SplitN(ctype, ";", 2)
		if split[0] != "" && split[0] == "application/json" {
			code = e.SUCCESS
		}

		if code != e.SUCCESS { // error return
			l, _ := c.Get("localizer")
			localizer, ok := l.(*translation.Localizer)
			var trans string = e.GetMsg(code)
			if ok {
				trans = localizer.Trans(trans, nil)
			}

			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"status":     "error",
				"statusCode": code,
				"msg":        trans,
				"data":       nil,
			})

			c.Abort()
			return
		}

		c.Next()
	}
}
