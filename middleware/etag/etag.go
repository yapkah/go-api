package etag

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/translation"
	"github.com/yapkah/go-api/pkg/util"
)

// Etag check if non match e-tag
func Etag() gin.HandlerFunc {
	return func(c *gin.Context) {

		code := e.SUCCESS
		etag := c.GetHeader("If-None-Match")
		var newEtag string

		d, ok := c.Get("response_data")
		if !ok {
			code := e.ETAG_RESPONSE_DATA_NOT_FOUND
			c.JSON(http.StatusNotFound, gin.H{
				"status":     "error",
				"statusCode": code,
				"msg":        trans(c, e.GetMsg(code)),
				"data":       nil,
			})
			c.Abort()
			return
		}

		response, ok := d.(app.Response)
		if !ok {
			code = e.ETAG_INVALID_RESPONSE_DATA
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":     "error",
				"statusCode": code,
				"msg":        trans(c, e.GetMsg(code)),
				"data":       nil,
			})
			c.Abort()
			return
		}

		// check data is same
		data, _ := json.Marshal(response.Data)
		newEtag = util.EncodeMD5(string(data))

		if etag != "" && newEtag == etag {
			c.JSON(http.StatusNotModified, gin.H{
				"status":     "error",
				"statusCode": http.StatusNotModified,
				"msg":        trans(c, e.GetMsg(http.StatusNotModified)),
				"data":       nil,
			})
			c.Abort()
			return
		}

		c.Header("ETag", "W/"+newEtag)
		c.JSON(http.StatusOK, response)
		c.Next()
	}
}

func trans(c *gin.Context, msg string) string {
	l, ok := c.Get("localizer")
	if !ok {
		return msg
	}

	localizer, ok := l.(*translation.Localizer)
	if !ok {
		return msg
	}

	return localizer.Trans(msg, nil)
}
