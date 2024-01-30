package jwt

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/translation"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/member_service"
)

// JWT is jwt middleware
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {

		// sourceInterface, _ := c.Get("source")
		// source := sourceInterface.(int)

		arrUrlSkipJWTDecode := []string{
			"/api/html5/v1/member/login",
			"/api/html5/v1/member/register",
			"/api/html5/v1/member/username/random",
			"/api/html5/v1/member/reset/password",
			"/api/html5/v1/member/reset/password/key",
			"/api/html5/v1/member/reset/secondary-pin",
			"/api/html5/v1/member/mnemonic/request",
			"/api/html5/v1/member/reset/secondary-pin",
		}
		route := c.Request.URL.String()

		skipJwtDecodeUrlStatus := helpers.StringInSlice(route, arrUrlSkipJWTDecode)
		if skipJwtDecodeUrlStatus {
			c.Next()
			return
		}

		if strings.Contains(route, "/api/html5/v1/member/otp/request") {
			c.Next()
			return
		}

		var code int

		code = e.SUCCESS

		// get header token
		token := c.GetHeader("Authorization")

		if token == "" { // empty token
			code = e.UNAUTHORIZED
			// models.ErrorLog("JWT-1", token, nil)
		} else {
			// parse token
			claim, err := util.ParseToken(token)
			if err != nil {
				if err.Error() == "public_key_missing" {
					code = e.PUBLIC_KEY_MISSING
				} else if strings.Contains(err.Error(), "token is expired") {
					// so far skip. not using this. will use htmlloginlog and apploginlog for so far
				} else {
					// models.ErrorLog("JWT-2", token, nil)
					code = e.UNAUTHORIZED
				}
			}

			if claim != nil {
				// set token claim
				c.Set("token_claim", claim)

				// check token id in db
				at, err := models.GetAllStatusAccessTokenByID(claim.Id)
				if err != nil {
					// models.ErrorLog("JWT-3", claim, at)
					code = e.UNAUTHORIZED
				}

				if at != nil && at.Status == "A" && at.SubID == claim.Subject {
					// set access token
					c.Set("access_token", at)

					// check user linked to token
					user, err := at.GetUser()
					if err != nil {
						// models.ErrorLog("JWT-4", user, err)
						code = e.UNAUTHORIZED
					}

					if user != nil {
						// set access user
						c.Set("access_user", user)

						sourceInterface, _ := c.Get("sourceName")
						sourceName := sourceInterface.(string)

						route := c.Request.URL.String()
						platformCheckingRst := strings.Contains(route, "/api/app")
						platform := "htmlfive"
						if platformCheckingRst {
							platform = "app"
						}
						c.Set("token", at.ID)

						if strings.ToLower(sourceName) != "laliga" {
							// start checking for htmlfive and app log log dt_expiry, token is expired
							tokenRst := member_service.ProcessValidateToken(platform, at.ID, user.GetMembersID())
							// fmt.Println("tokenRst:", tokenRst)
							// models.ErrorLog("tokenRst3:", tokenRst, nil)
							if !tokenRst {
								// arrDebug3 := map[string]interface{}{
								// 	"platform":     platform,
								// 	"ID":           at.ID,
								// 	"GetMembersID": user.GetMembersID(),
								// }
								// models.ErrorLog("JWT-5", tokenRst, arrDebug3)
								code = e.UNAUTHORIZED
							}
							// end checking for htmlfive and app log log dt_expiry, token is expired
						}

						if code == e.SUCCESS && at.Source == 0 {
							// start decide the token need to b extend expiry or not
							member_service.ProcessExtendLoginPeriod(platform, at.ID, user.GetMembersID())
							// end decide the token need to b extend expiry or not
						}
					} else {
						// models.ErrorLog("JWT-6", "UNAUTHORIZED not user", nil)
						code = e.UNAUTHORIZED
					}
				} else {
					if at != nil && at.Status == models.GetAccesTokenReplaceStatus() {
						code = e.YOUR_ACCOUNT_IS_LOGIN_AT_ANOTHER_DEVICE
					} else {
						// models.ErrorLog("JWT-7", "else in GetAccesTokenReplaceStatus", nil)
						code = e.UNAUTHORIZED
					}
				}

			} else {
				// models.ErrorLog("JWT-8", "else in claim", nil)
				code = e.UNAUTHORIZED
			}
		}

		if code != e.SUCCESS { // error return

			urlReturnSuccessMap := []string{
				"/member/logout",
			}

			route := c.Request.URL.String()
			for _, v1 := range urlReturnSuccessMap {
				isReturnSuccessRst := strings.Contains(route, v1)
				if isReturnSuccessRst {
					c.JSON(http.StatusOK, gin.H{
						"rst":  1,
						"msg":  helpers.Translate("success", "en"),
						"data": nil,
					})
					c.Abort()
					return
				}
			}
			// models.ErrorLog("JWT-9", "last", nil)
			c.JSON(http.StatusUnauthorized, gin.H{
				"rst":  0,
				"msg":  trans(c, e.GetMsg(code)),
				"data": nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func trans(c *gin.Context, msg string) string {
	l, ok := c.Get("localizer")
	if !ok {
		return msg
	}

	localizer, ok := l.(*translation.Localizer)
	if !ok {
		return msg
	}

	return localizer.Trans(msg, nil)
}
