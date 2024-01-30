package member

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
)

// get country list func
func CountryList(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()

	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	// find country list
	country, err := models.GetCountryList()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()

	for _, v := range country {
		v.Name = helpers.Translate(v.Name, langCode)
		countryCode := strings.Replace(strings.ToLower(v.Code), " ", "_", -1)
		v.CountryFlagUrl = adminServerDomain + "/assets/global/img/512_flags/" + countryCode + ".png"
	}

	appG.Response(1, http.StatusOK, "success", country)
}
