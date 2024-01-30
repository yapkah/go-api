package member

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/service/member_service"
)

// Logout function
func Logout(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	sourceInterface, _ := c.Get("source")
	source := uint8(sourceInterface.(int))

	route := c.Request.URL.String()
	platformCheckingRst := strings.Contains(route, "/api/app")
	platform := "HTMLFIVE"
	if platformCheckingRst {
		platform = "APP"
	}

	// find member
	// get access user from middle ware
	u, ok := c.Get("access_user")

	if !ok {
		// user not found
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "logout_failed"}, nil)
	}

	mem := u.(*models.EntMemberMembers)

	tx := models.Begin()

	member_service.ProcessMemberLogout(tx, *mem, platform, source)

	_ = models.Commit(tx)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)

}
