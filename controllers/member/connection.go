package member

// "github.com/smartblock/gta-api/service/notification_service"

// UpdateConnectionForm struct
// type UpdateConnectionForm struct {
// 	Connection string `json:"connection" valid:"Required"`
// 	Device     string `json:"device" valid:"Required"`
// 	OS         string `json:"os" valid:"Required"`
// 	OsVersion  string `json:"os_version" valid:"Required"`
// 	AppVersion string `json:"app_version" valid:"Required"`
// 	// ReceiveNoti int    `json:"receive_noti" valid:"Min(0);Max(1)"`
// }

// // UpdateConnection function
// func UpdateConnection(c *gin.Context) {
// 	var (
// 		appG = app.Gin{C: c}
// 		form UpdateConnectionForm
// 		err  error
// 	)

// 	ok, msg := app.BindAndValid(c, &form)

// 	if ok == false {
// 		appG.ResponseV2("error", http.StatusBadRequest, e.INVALID_PARAMS, msg[0], "")
// 		return
// 	}

// 	u, ok := c.Get("access_user")
// 	if !ok {
// 		appG.Response("error", http.StatusUnauthorized, e.UNAUTHORIZED, nil)
// 		return
// 	}
// 	member := u.(*models.Members)

// 	// get access token from middle ware
// 	t, ok := c.Get("access_token")
// 	if !ok {
// 		// token not found
// 		appG.Response("error", http.StatusUnauthorized, e.UNAUTHORIZED, nil)
// 		return
// 	}
// 	at := t.(*models.AccessToken)

// 	// get connection
// 	var receiveNoti int = 1
// 	conn, err := models.GetNotiConnectionSetting(form.Connection)
// 	if err != nil {
// 		appG.ResponseError(err)
// 		return
// 	}

// 	if conn != nil {
// 		receiveNoti = conn.ReceiveNoti
// 	}

// 	tx := models.Begin()

// 	_, notiSet, err := member_service.UpdateMemberConnection(tx, member, at.ID, form.Connection, form.Device, form.OS, form.OsVersion, form.AppVersion, receiveNoti, "A")
// 	if err != nil {
// 		models.Rollback(tx)
// 		appG.ResponseError(err)
// 		return
// 	}

// 	err = models.Commit(tx)
// 	if err != nil {
// 		appG.ResponseError(err)
// 		return
// 	}

// 	tx = models.Begin()

// 	err = notification_service.UpdateMemberNotiStatus(tx, notiSet, member, receiveNoti)
// 	if err != nil {
// 		models.Rollback(tx)
// 		appG.ResponseError(err)
// 		return
// 	}

// 	err = models.Commit(tx)
// 	if err != nil {
// 		appG.ResponseError(err)
// 		return
// 	}

// 	appG.Response("success", http.StatusOK, e.SUCCESS, nil)
// }
