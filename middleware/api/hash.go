package api

import (
	"github.com/gin-gonic/gin"
)

type List struct {
	Messages map[string][]string `key:"required"`
}

// CheckHash is params hash checking middleware
func CheckHash() gin.HandlerFunc {
	return func(c *gin.Context) {

		//var data map[string]interface{}
		////text, _ := json.Marshal(c.Request.Body)
		//byte, _ := ioutil.ReadAll(c.Request.Body)
		//json.Unmarshal(byte, &data)
		//
		//var array [20]string
		//
		//for k, v := range data{
		//}


	}
}