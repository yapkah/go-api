package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
)

// p.s. This middleware only work with member htmlfive and app api middlware
func CheckForMaintenanceMode() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var err error
		route := c.Request.URL.String()
		platformCheckingRst := strings.Contains(route, "/api/app")
		platform := "web"
		if platformCheckingRst {
			platform = "app"
		}

		arrReturnSuccessWithUnauthorize := []string{
			"/api/html5/v1/member/login",
			"/api/html5/v1/member/register/mobile",
			"/api/html5/v1/member/password",
		}
		// fmt.Println("ClientIP:", c.ClientIP())
		// fmt.Println("Host:", c.Request.Host)
		// fmt.Println("RemoteAddr:", c.Request.RemoteAddr)
		// fmt.Println("X-Forwarded-Host:", c.Request.Header["X-Forwarded-Host"])
		// fmt.Println("header:", c.Request.Header)

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " b_latest = ?", CondValue: 1},
			// models.WhereCondFn{Condition: " platform = ?", CondValue: platform},
		)

		arrLatestAppAppVersion, _ := models.GetLatestAppAppVersionFn(arrCond, false)

		if len(arrLatestAppAppVersion) < 1 {
			base.LogErrorLog("CheckForMaintenanceMode_no_version_is_set", "please_set_a_version_to_use_CheckForMaintenanceMode", arrCond, true)
			c.Next()
			return
		}

		arrMaintenancceSetting, _ := models.GetSysGeneralSetupByID("maintenance_setting")

		if arrMaintenancceSetting != nil {
			type ArrMaintenanceSettingStruct struct {
				NumOfAPIPlat int      `json:"numOfApiPlat"`
				SkipUsername []string `json:"skipUsername"`
				SkipMemberID []int    `json:"skipMemberID"` // this will change to member.id
				SkipURL      []string `json:"skipUrl"`
				APIPlatform  []struct {
					URL      string   `json:"url"`
					Platform []string `json:"platform"`
				} `json:"apiPlatform"`
			}
			var ArrMaintenanceSettingData ArrMaintenanceSettingStruct

			var encodedByPassDomainData []string

			_ = json.Unmarshal([]byte(arrMaintenancceSetting.InputType3), &encodedByPassDomainData)
			if len(encodedByPassDomainData) > 0 {
				if len(c.Request.Header["Origin"]) > 0 {
					// fmt.Println("header Origin:", c.Request.Header["Origin"][0])
					// fmt.Printf("header Origin: type%T\n", c.Request.Header["Origin"][0])

					// perform bypass by Origin
					skipOriginStatus := helpers.StringInSlice(c.Request.Header["Origin"][0], encodedByPassDomainData)
					if skipOriginStatus {
						c.Next()
						return
					}
				}
				if len(c.Request.Header["Referer"]) > 0 {
					// fmt.Printf("%T\n", c.Request.Header["Referer"])
					// fmt.Println("header Referer:", c.Request.Header["Referer"][0])
					// fmt.Printf("%T\n", c.Request.Header["Referer"][0])
					// perform bypass by Referer
					skipRefererStatus := helpers.StringInSlice(c.Request.Header["Referer"][0], encodedByPassDomainData)
					if skipRefererStatus {
						c.Next()
						return
					}
				}
			}

			err := json.Unmarshal([]byte(arrMaintenancceSetting.InputType1), &ArrMaintenanceSettingData)
			if err != nil {
				base.LogErrorLog("CheckForMaintenanceMode-error in decoding json format_apiMaintenanceSetting", arrMaintenancceSetting.InputType1, ArrMaintenanceSettingData, true)
				c.Next()
				return
			}

			var missingPlatformStatus bool
			if ArrMaintenanceSettingData.APIPlatform != nil {
				missingPlatformStatus = true
				for _, apiPlatformV := range ArrMaintenanceSettingData.APIPlatform {
					apiUrl := strings.Replace(apiPlatformV.URL, "/*", "", -1)
					platformCheckingRst := strings.Contains(route, apiUrl)
					if platformCheckingRst {
						for _, platformV := range apiPlatformV.Platform {
							for _, arrLatestAppAppVersionV := range arrLatestAppAppVersion {
								if strings.ToLower(platformV) == strings.ToLower(arrLatestAppAppVersionV.Platform) {
									missingPlatformStatus = false
									if arrLatestAppAppVersionV.Maintenance == 0 {
										c.Next()
										return
									}
								}
							}
						}
					}
				}
			}

			if missingPlatformStatus {
				base.LogErrorLog("CheckForMaintenanceMode-error please_set_a_version_to_use_CheckForMaintenanceMode", ArrMaintenanceSettingData.APIPlatform, platform, true)
			}

			// start perform bypass by SkipURL
			if arrReturnSuccessWithUnauthorize != nil {
				skipUrlStatus := helpers.StringInSlice(route, arrReturnSuccessWithUnauthorize)
				if skipUrlStatus {
					c.JSON(http.StatusOK, gin.H{
						"rst":  0,
						"msg":  helpers.TranslateV2("member_maintenance_msg", "en", make(map[string]string)),
						"data": nil,
					})
					c.Abort()
					return
				}
			}
			// end perform bypass by SkipURL

			// start perform bypass by SkipURL
			if ArrMaintenanceSettingData.SkipURL != nil {
				skipUrlStatus := helpers.StringInSlice(route, ArrMaintenanceSettingData.SkipURL)
				if skipUrlStatus {
					c.Next()
					return
				}
			}
			// end perform bypass by SkipURL

			// start perform bypass by username
			u, _ := c.Get("access_user")
			if u != nil {
				// perform bypass by member id
				if ArrMaintenanceSettingData.SkipMemberID != nil {
					members := u.(*models.EntMemberMembers)
					skipUsernameStatus := helpers.IntInSlice(members.ID, ArrMaintenanceSettingData.SkipMemberID)
					if skipUsernameStatus {
						c.Next()
						return
					}
				}
				// perform bypass by username
				if ArrMaintenanceSettingData.SkipUsername != nil {
					members := u.(*models.EntMemberMembers)
					skipUsernameStatus := helpers.StringInSlice(members.GetUserName(), ArrMaintenanceSettingData.SkipUsername)
					if skipUsernameStatus {
						c.Next()
						return
					}
				}
			}
			// end perform bypass by username

			c.JSON(http.StatusOK, gin.H{
				"rst":  0,
				"msg":  helpers.TranslateV2("member_maintenance_msg", "en", make(map[string]string)),
				"data": nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
