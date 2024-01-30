package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/controllers/bcadmin"
)

// BCAdmin func
func BCAdmin(route *gin.RouterGroup) {
	route.POST("/member/search", bcadmin.SearchMemberInfoForBCAdmin) // save transaction from blockchain
}
