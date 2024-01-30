package member

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/app_version_service"
	"github.com/smartblock/gta-api/service/media_service"
)

// GetAppVersListv1 function
func GetAppVersListv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form app_version_service.AppVersListForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	arrData := app_version_service.AppVersListForm{
		Platform: form.Platform,
		Latest:   form.Latest,
	}
	arrDataReturn := app_version_service.GetAppVersListv1(arrData)

	token := c.GetHeader("Authorization")

	var username string
	if token != "" {
		claim, _ := util.ParseToken(token)
		if claim != nil {
			at, _ := models.GetAllStatusAccessTokenByID(claim.Id)
			if at != nil {
				user, _ := at.GetUser()
				if user != nil {
					if user.GetUserName() != "" {
						username = user.GetUserName()
					}
				}
			}
		}
	}

	if username != "" {
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
			var arrMaintenanceSettingData ArrMaintenanceSettingStruct

			json.Unmarshal([]byte(arrMaintenancceSetting.InputType1), &arrMaintenanceSettingData)

			// perform bypass by username
			if arrMaintenanceSettingData.SkipUsername != nil {
				skipUsernameStatus := helpers.StringInSlice(username, arrMaintenanceSettingData.SkipUsername)
				if skipUsernameStatus {
					for arrDataReturnK, _ := range arrDataReturn {
						arrDataReturn[arrDataReturnK].Maintenance = 0
					}
				}
			}
		}
	}

	//check ballot session
	// ballotStatus := 1
	// arrGeneralSetup, _ := models.GetSysGeneralSetupByID("ballot_setting")

	// if arrGeneralSetup != nil {
	// 	type arrBallotGeneralSettingStructv2 struct {
	// 		StartTime string `json:"start_time"`
	// 		EndTime   string `json:"end_time"`
	// 	}
	// 	var arrGeneralSettingv2 arrBallotGeneralSettingStructv2

	// 	json.Unmarshal([]byte(arrGeneralSetup.InputValue2), &arrGeneralSettingv2)

	// 	if arrGeneralSettingv2.StartTime != "" && arrGeneralSettingv2.EndTime != "" {
	// 		currTime := time.Now().Format("2006-01-02 15:04:05")

	// 		if currTime > arrGeneralSettingv2.EndTime {
	// 			ballotStatus = 0
	// 		}

	// 	}
	// }

	for arrDataReturnK, _ := range arrDataReturn {
		// arrDataReturn[arrDataReturnK].RegisterStatus = ballotStatus
		arrDataReturn[arrDataReturnK].RegisterStatus = 1

	}

	// fmt.Println("result:", result)
	// // add ent_member_kyc record
	// _, err := models.AddEntMemberKYC(tx, arrData)
	// if err != nil {
	// 	models.Rollback(tx)
	// 	appG.ResponseError(err)
	// 	return
	// }

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
	return
}

// ProcessAppVersv1 func
func ProcessAppVersv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form app_version_service.ProcessAppVersionForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	file, header, err := c.Request.FormFile("apk_file")
	if file != nil {
		err = media_service.MediaValidation(file, header, "app")
		if err != nil {
			message := app.MsgStruct{
				Msg: err.Error() + "_for_apk_file",
			}
			appG.ResponseV2(0, http.StatusOK, message, "")
			return
		}

		filename := header.Filename
		mediaData, err := media_service.UploadApp(file, filename)

		if err != nil {
			message := app.MsgStruct{
				Msg: err.Error(),
			}
			appG.ResponseV2(0, http.StatusOK, message, "")
			return
		}
		form.Path = mediaData.FullURL
		form.FolderPath = mediaData.FileDirectory

	}

	result := app_version_service.ProcessAppVersion(form)

	if result != nil {
		message := app.MsgStruct{
			Msg: result.Error(),
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, nil)
	return
}

type CheckAppVersForm struct {
	Platform   string `form:"platform" json:"platform" valid:"Required;"`
	AppVersion string `form:"app_version" json:"app_version" valid:"Required;"`
}

// CheckAppVersv1 function
func CheckAppVersv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form CheckAppVersForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	sourceInterface, _ := c.Get("source")
	source := uint8(sourceInterface.(int))

	arrData := app_version_service.CheckAppVers{
		Platform:   form.Platform,
		AppVersion: form.AppVersion,
		LangCode:   langCode,
		Source:     source,
	}
	arrDataReturn, status, err := app_version_service.CheckAppVersv1(arrData)

	message := app.MsgStruct{
		Msg: err.Error(),
	}
	var rst int
	if status {
		rst = 1
	}
	if arrDataReturn != nil {
		appG.ResponseV2(rst, http.StatusOK, message, arrDataReturn)
		return
	}
	appG.ResponseV2(rst, http.StatusOK, message, nil)
	return
}
