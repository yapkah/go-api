package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
)

type auctionResponseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r auctionResponseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// auctionApiLogDisable route name
var auctionApiLogDisable = []string{
	// "/api/v1/member/profile/photo",
	// "/api/v1/admin/translation/updatebyfile",
	// "/api/v1/admin/app/version",
	"/api/v1/member/version/check",
}

// auctionApiLogDisableInputLog route name
var auctionApiLogDisableInputLog = []string{
	"/api/v1/member/profile/photo",
	"/api/v1/admin/translation/updatebyfile",
	"/api/html5/v1/member/file/upload",
	"/api/app/v1/member/file/upload",
	"/api/app/v1/app-version/process",
}

// Log api log
func LogAuctionApiLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start := time.Now().UnixNano() / int64(time.Millisecond)
		for _, url := range auctionApiLogDisable {
			if url == c.Request.URL.String() {
				c.Next()
				return
			}
		}

		// var err error
		route := c.Request.URL.String()
		method := c.Request.Method
		header, _ := json.Marshal(c.Request.Header)
		ip := c.ClientIP()
		r, _ := c.GetRawData()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(r))) // Write body back
		input := string(r)

		// disable input log
		for _, url := range auctionApiLogDisableInputLog {
			if url == c.Request.URL.String() {
				input = ""
				break
			}
		}
		log, _ := models.AddAuctionApiLog(route, method, string(header), ip, input)

		// replace response writer
		w := &auctionResponseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		if log != nil {
			c.Set("auction_api_log", log)
		}
		c.Next()

		if log != nil {
			output := w.body.String()

			type response struct {
				Status string
			}
			var res response

			json.Unmarshal([]byte(w.body.String()), &res)

			// if res.Status == "" || res.Status == "success" {
			// 	output = ""
			// }

			// end := time.Now().UnixNano() / int64(time.Millisecond)
			// runtime := end - start
			log.UpdateAuctionApiLog(output)

		}
	}
}
