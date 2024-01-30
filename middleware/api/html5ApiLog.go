package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
)

type htmlfiveResponseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r htmlfiveResponseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// htmlfiveApiLogDisable route name
var htmlfiveApiLogDisable = []string{
	// "/api/v1/member/profile/photo",
	// "/api/v1/admin/translation/updatebyfile",
	// "/api/v1/admin/app/version",
	"/api/v1/member/version/check",
}

// htmlfiveApiLogDisableInputLog route name
var htmlfiveApiLogDisableInputLog = []string{
	"/api/v1/member/profile/photo",
	"/api/v1/admin/translation/updatebyfile",
	"/api/html5/v1/member/file/upload",
}

// Log api log
func LogHtml5ApiLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start := time.Now().UnixNano() / int64(time.Millisecond)
		for _, url := range htmlfiveApiLogDisable {
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
		for _, url := range htmlfiveApiLogDisableInputLog {
			if url == c.Request.URL.String() {
				input = ""
				break
			}
		}
		log, _ := models.AddHtmlfiveApiLog(route, method, string(header), ip, input)

		// replace response writer
		w := &htmlfiveResponseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		if log != nil {
			c.Set("htmlfive_api_log", log)
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
			log.UpdateHtmlfiveApiLog(output)

		}
	}
}

// HtmlfivelogOutput func
func HtmlfivelogOutput(c *gin.Context, output interface{}) error {
	l, ok := c.Get("htmlfive_api_log")
	if ok {
		log := l.(*models.HtmlfiveApiLog)
		o, _ := json.Marshal(output)
		log.UpdateHtmlfiveOutput(string(o))
	}
	return nil
}
