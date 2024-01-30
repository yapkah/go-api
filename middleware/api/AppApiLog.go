package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
)

type appResponseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r appResponseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// appApiLogDisable route name
var appApiLogDisable = []string{
	// "/api/v1/member/profile/photo",
	// "/api/v1/admin/translation/updatebyfile",
	// "/api/v1/admin/app/version",
	"/api/v1/member/version/check",
}

// appApiLogDisableInputLog route name
var appApiLogDisableInputLog = []string{
	"/api/v1/member/profile/photo",
	"/api/v1/admin/translation/updatebyfile",
	"/api/html5/v1/member/file/upload",
	"/api/app/v1/member/file/upload",
	"/api/app/v1/app-version/process",
}

// Log api log
func LogAppApiLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start := time.Now().UnixNano() / int64(time.Millisecond)
		for _, url := range appApiLogDisable {
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
		for _, url := range appApiLogDisableInputLog {
			if url == c.Request.URL.String() {
				input = ""
				break
			}
		}
		log, _ := models.AddAppApiLog(route, method, string(header), ip, input)

		// replace response writer
		w := &appResponseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		if log != nil {
			c.Set("app_api_log", log)
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
			log.UpdateAppApiLog(output)

		}
	}
}

// ApplogOutput func
func ApplogOutput(c *gin.Context, output interface{}) error {
	l, ok := c.Get("app_api_log")
	if ok {
		log := l.(*models.AppApiLog)
		o, _ := json.Marshal(output)
		log.UpdateAppOutput(string(o))
	}
	return nil
}
