package cors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	gincors "github.com/rs/cors/wrapper/gin"
	"github.com/yapkah/go-api/pkg/setting"
)

// Cors cors
func Cors() gin.HandlerFunc {
	config := cors.Options{
		AllowedOrigins: setting.CorsSetting.AllowOrigin,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}
	return gincors.New(config)
}
