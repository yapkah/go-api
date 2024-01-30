package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/controllers/laliga"
	"github.com/yapkah/go-api/middleware/jwt"
)

// Laliga func
func Laliga(route *gin.RouterGroup) {
	auth := route.Group("/")
	auth.Use(jwt.JWT())
	// auth.Use(jwt.CheckScopeOr([]string{"ACCESS"})) // check if it is an access token
	{
		auth.POST("/member/wallet/trans/save", laliga.ProcessSaveTransv1) // save wallet trans
	}
}
