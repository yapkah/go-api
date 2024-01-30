package trans

import (
	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/pkg/translation"
)

// SetLocale func
func SetLocale() gin.HandlerFunc {
	return func(c *gin.Context) {
		var localizer *translation.Localizer
		accept := c.GetHeader("Accept-Language")
		if accept != "" {
			localizer = translation.NewLocalizer(accept)
		} else {
			localizer = translation.NewLocalizer("en")
		}
		c.Set("localizer", localizer)
		c.Next()
	}
}
