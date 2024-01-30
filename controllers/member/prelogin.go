package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/pkg/app"
)

// GetPreloginDocumentList function
func GetPreloginDocumentList(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	// get access user from middle ware
	// u, ok := c.Get("access_user")
	// if !ok {
	// 	// user not found
	// 	appG.Response(0, http.StatusUnauthorized, "something_went_wrong", "")
	// }
	// member := u.(*models.EntMemberMembers)

	arrDataReturn := map[string]interface{}{
		"term_of_service_url": "https://media02.securelayers.cloud/medias/WOD/PRELOGIN/DOCUMENT/wod_terms_of_service.pdf",
		"privacy_policy_url":  "https://media02.securelayers.cloud/medias/WOD/PRELOGIN/DOCUMENT/wod_privacy_policy.pdf",
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}
