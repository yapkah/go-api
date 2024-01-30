package member

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
)

// UpdateDeviceInfoForm struct
type UpdateDeviceInfoForm struct {
	AppVersion       string `form:"app_version" json:"app_version"`
	Manufacturer     string `form:"manufacturer" json:"manufacturer"`
	Model            string `form:"model" json:"model"`
	OS               string `form:"os" json:"os"`
	OSVersion        string `form:"os_version" json:"os_version"`
	PushNotification string `form:"push_notification" json:"push_notification"`
}

// UpdateDeviceInfo func
// Perform save device bind log and update member device info
func UpdateDeviceInfo(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form UpdateDeviceInfoForm
	)

	sourceInterface, _ := c.Get("source")
	source := uint8(sourceInterface.(int))

	if source == 1 {
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
		return
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	accessToken, ok := c.Get("access_token")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}
	at := accessToken.(*models.AccessToken)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "b_login = ?", CondValue: 1},
		models.WhereCondFn{Condition: "b_logout = ?", CondValue: 0},
		models.WhereCondFn{Condition: "t_token = ?", CondValue: at.ID},
	)
	appLoginLogRst, _ := models.GetAppLoginLogFn(arrCond, "", false)

	if appLoginLogRst != nil {
		arrCrtData := models.DeviceBindLog{
			MemberID:  member.EntMemberID,
			Bind:      1,
			CreatedBy: strconv.Itoa(member.EntMemberID),
		}
		updateColumn := map[string]interface{}{}
		if form.AppVersion != "" {
			updateColumn["t_app_version"] = form.AppVersion
			arrCrtData.TAppVersion = form.AppVersion
		}
		if form.Manufacturer != "" {
			updateColumn["t_manufacturer"] = form.Manufacturer
			arrCrtData.TManufacturer = form.Manufacturer
		}
		if form.Model != "" {
			updateColumn["t_model"] = form.Model
			arrCrtData.TModel = form.Model
		}
		if form.OS != "" {
			updateColumn["t_os"] = form.OS
			arrCrtData.TOs = form.OS
		}
		if form.OSVersion != "" {
			updateColumn["t_os_version"] = form.OSVersion
			arrCrtData.TOsVersion = form.OSVersion
		}
		if form.PushNotification != "" {
			updateColumn["t_push_noti_token"] = form.PushNotification
			arrCrtData.TPushNotiToken = form.PushNotification
		}

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " device_bind_log.t_push_noti_token = ? ", CondValue: appLoginLogRst.TPushNotiToken},
			models.WhereCondFn{Condition: " device_bind_log.member_id = ? ", CondValue: member.EntMemberID},
		)
		exsitingDeviceBindLogRst, _ := models.GetLatestDeviceBindLogFn(arrCond, false)

		if exsitingDeviceBindLogRst == nil || exsitingDeviceBindLogRst.Bind == 0 {
			_ = models.AddDeviceBindLog(arrCrtData)
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " app_login_log.id = ? ", CondValue: appLoginLogRst.ID},
		)

		_ = models.UpdatesFn("app_login_log", arrCond, updateColumn, false)
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)

	return
}
