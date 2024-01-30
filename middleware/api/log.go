package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// apiLogDisable route name
var apiLogDisable = []string{
	// "/api/v1/member/profile/photo",
	// "/api/v1/admin/translation/updatebyfile",
	// "/api/v1/admin/app/version",
	"/api/v1/member/version/check",
}

// apiLogDisableInputLog route name
var apiLogDisableInputLog = []string{
	"/api/v1/member/profile/photo",
	"/api/v1/admin/translation/updatebyfile",
	"/api/v1/admin/app/version",
}

// Log api log
func Log() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UnixNano() / int64(time.Millisecond)
		for _, url := range apiLogDisable {
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
		for _, url := range apiLogDisableInputLog {
			if url == c.Request.URL.String() {
				input = ""
			}
		}

		log, _ := models.AddAPILog(route, method, string(header), ip, input)

		// replace response writer
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		if log != nil {
			c.Set("api_log", log)
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

			end := time.Now().UnixNano() / int64(time.Millisecond)
			runtime := end - start
			log.Update(output, int(runtime))

		}
	}
}

// LogOutput func
func LogOutput(c *gin.Context, output interface{}) error {
	l, ok := c.Get("api_log")
	if ok {
		log := l.(*models.ApiLog)
		o, _ := json.Marshal(output)
		log.UpdateOutput(string(o))
	}
	return nil
}
