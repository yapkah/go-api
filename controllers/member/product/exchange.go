package product

// GetExchangeSetting func
// func GetExchangeSetting(c *gin.Context) {
// 	var (
// 		appG   = app.Gin{C: c}
// 		errMsg string
// 	)

// 	// get user info
// 	u, ok := c.Get("access_user")
// 	if !ok {
// 		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
// 		return
// 	}

// 	member := u.(*models.EntMemberMembers)

// 	// get lang code
// 	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
// 	if c.GetHeader("Accept-Language") != "" {
// 		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
// 		if ok {
// 			langCode = c.GetHeader("Accept-Language")
// 		}
// 	}

// 	arrData, errMsg := product_service.GetExchangeSetting(member.EntMemberID, langCode)

// 	if errMsg != "" {
// 		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
// 		return
// 	}

// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
// 	return
// }

// PostExchangeForm struct
// type PostExchangeForm struct {
// 	Type         string  `form:"type" json:"type" valid:"Required"`
// 	Amount       float64 `form:"amount" json:"amount" valid:"Required"`
// 	Payments     string  `form:"payments" json:"payments" valid:"Required"`
// 	SecondaryPin string  `form:"secondary_pin" json:"secondary_pin"`
// }

// PostExchange function for verification without access token
// func PostExchange(c *gin.Context) {
// 	var (
// 		appG = app.Gin{C: c}
// 		form PostExchangeForm
// 		err  error
// 	)

// 	ok, msg := app.BindAndValid(c, &form)
// 	if ok == false {
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
// 		return
// 	}

// 	// get lang code
// 	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
// 	if c.GetHeader("Accept-Language") != "" {
// 		ok = models.ExistLangague(c.GetHeader("Accept-Language"))
// 		if ok {
// 			langCode = c.GetHeader("Accept-Language")
// 		}
// 	}

// 	// get member info from access token
// 	u, ok := c.Get("access_user")
// 	if !ok {
// 		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
// 		return
// 	}

// 	member := u.(*models.EntMemberMembers)

// 	entMemberID := member.EntMemberID

// 	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
// 	if err != nil {
// 		base.LogErrorLog("Register-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin_format"}, nil)
// 		return
// 	}

// 	wordCount := utf8.RuneCountInString(decryptedText)
// 	if wordCount < 6 {
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
// 		return
// 	}
// 	if wordCount > 6 {
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "transaction_pin_minimum_character_is_:0", Params: map[string]string{"0": "6"}}, nil)
// 		return
// 	}
// 	form.SecondaryPin = decryptedText

// 	// check secondary password
// 	pinValidation := base.SecondaryPin{
// 		MemId:              entMemberID,
// 		SecondaryPin:       form.SecondaryPin,
// 		MemberSecondaryPin: member.SecondaryPin,
// 		LangCode:           langCode,
// 	}

// 	err = pinValidation.CheckSecondaryPin()
// 	if err != nil {
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
// 		return
// 	}

// 	// being transaction
// 	tx := models.Begin()

// 	sourceInterface, _ := c.Get("sourceName")
// 	sourceName := sourceInterface.(string)

// 	genTranxDataStatus := false
// 	if strings.ToLower(sourceName) == "htmlfive" {
// 		genTranxDataStatus = true
// 	}

// 	// perform post exchange action
// 	postExchangeStruct := product_service.PostExchangeStruct{
// 		Type:               form.Type,
// 		MemberID:           entMemberID,
// 		Amount:             form.Amount,
// 		Payments:           form.Payments,
// 		GenTranxDataStatus: genTranxDataStatus,
// 	}

// 	msgStruct, arrData, exchangeCallback := product_service.PostExchange(tx, postExchangeStruct, langCode)
// 	if msgStruct.Msg != "" {
// 		models.Rollback(tx)
// 		appG.ResponseV2(0, http.StatusOK, msgStruct, nil)
// 		return
// 	}

// 	// commit transaction
// 	err = models.Commit(tx)
// 	if err != nil {
// 		models.ErrorLog("walletController:PostExchange()", "Commit():1", err.Error())
// 		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 		return
// 	}

// 	// if no use blockchain wallet, straight call ExchangeCallback() to generate and send exchange_debit signed transaction
// 	if exchangeCallback.Callback {
// 		db := models.GetDB() // no need set begin transaction
// 		errMsg := product_service.ExchangeCallback(db, exchangeCallback.DocNo)
// 		if errMsg != "" {
// 			base.LogErrorLog("walletController:PostExchange()", "ExchangeCallback():1", errMsg, true)
// 			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
// 			return
// 		}
// 	}

// 	// initialize the map before writing into it
// 	if arrData == nil {
// 		arrData = make(map[string]string)
// 	}

// 	arrData["receiving_summary"] = fmt.Sprintf("%s %s", helpers.CutOffDecimal(float64(form.Amount), 8, ".", ","), helpers.TranslateV2("USDS", langCode, make(map[string]string)))
// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
// 	return
// }
