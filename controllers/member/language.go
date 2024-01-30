package member

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/language_service"
)

//get frontend translation function
func GetTranslation(c *gin.Context) {

	type FrontendTranslation struct {
		Lang      string `uri:"lang"`
		Namespace string `uri:"namespace"`
	}

	var (
		appG = app.Gin{C: c}
		form FrontendTranslation
	)

	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	// fmt.Println(form.Lang)
	// fmt.Println(form.Namespace)
	var namespace = strings.Replace(form.Namespace, ".json", "", 1)

	// fmt.Println(namespace)

	lang := models.GetFrontendTranslation(form.Lang, namespace)

	appG.C.JSON(http.StatusOK, lang)

	return
}

//Add Frontend Translation func
func AddTranslation(c *gin.Context) {

	type AddFrontendTranslation struct {
		Key string `form:"key" json:"key" valid:"Required"`
	}

	var (
		appG = app.Gin{C: c}
		form AddFrontendTranslation
	)

	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	key := form.Key

	processKey := strings.Split(key, ".")

	var group = processKey[0]
	var name = processKey[1]
	var name2 = strings.Split(name, "_")
	var value = ""

	for _, vname := range name2 {
		if len(vname) > 1 {
			vname = strings.Title(vname)
		}

		if value != "" {
			value = value + " "
		}

		if vname == strings.ToUpper(vname) {
			vname = "{{" + strings.ToLower(vname) + "}}"
		}
		value = value + vname
	}

	languages, _ := models.GetLanguageList()

	for _, v := range languages {
		check := models.GetFrontendTranslationByName(v.Locale, group, name)

		if check == nil {
			models.AddFrontendTranslation(v.Locale, group, name, value)
		}
	}

	appG.Response(1, http.StatusOK, "success", nil)
}

//get language list func
func LanguageList(c *gin.Context) {

	type LanguageReturnStruct struct {
		ID      string `json:"id"`
		Locale  string `json:"locale"`
		Name    string `json:"name"`
		FlagUrl string `json:"flag_url"`
	}

	var (
		appG      = app.Gin{C: c}
		arrReturn []LanguageReturnStruct
		langCode  string
	)

	lang, err := models.GetLanguageList()

	if err != nil {
		appG.ResponseError(err)
		return
	}

	adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()

	for _, v := range lang {

		langCode = strings.Replace(strings.ToLower(v.Locale), " ", "_", -1)

		flagUrl := adminServerDomain + "/assets/global/img/lang_flags/" + langCode + ".png"

		arrReturn = append(arrReturn, LanguageReturnStruct{
			ID:      v.ID,
			Locale:  v.Locale,
			Name:    v.Name,
			FlagUrl: flagUrl,
		})
	}

	appG.Response(1, http.StatusOK, "success", arrReturn)
}

//get app frontend translation function
func GetAppTranslation(c *gin.Context) {

	type AppTranslateReq struct {
		EtagID string `form:"etag_id" json:"etag_id"`
	}

	var (
		appG = app.Gin{C: c}
		form AppTranslateReq
	)

	formLangCode := strings.Replace(c.Param("lang"), "/", "", -1)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if formLangCode != "" {
		langCode = formLangCode
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}
	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], nil)
		return
	}

	arrData := language_service.AppTranslationStruct{
		EtagID:   form.EtagID,
		LangCode: langCode,
	}

	rst, err := language_service.GetAppTranslation(arrData)
	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	arrDataReturn := map[string]interface{}{
		"etag_id":         rst.EtagID,
		"translated_list": rst.TranslatedList,
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
	return
}

func UpdateAppTranslation(c *gin.Context) {

	arrCond := make([]models.WhereCondFn, 0)
	result, _ := models.GetTransFn(arrCond, false)

	if len(result) > 0 {
		for _, resultV := range result {
			fmt.Println("resultV:", resultV)
		}
	}
}

// UpdateMemberDeviceLanguagev1Form struct
type UpdateMemberDeviceLanguagev1Form struct {
	PushNotification string `form:"push_notification" json:"push_notification" valid:"Required;"`
	LangCode         string `form:"lang_code" json:"lang_code" valid:"Required;"`
}

// func update current device language
func UpdateMemberDeviceLanguagev1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form UpdateMemberDeviceLanguagev1Form
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}
	route := c.Request.URL.String()
	platformCheckingRst := strings.Contains(route, "/api/app")
	platform := "htmlfive"
	if platformCheckingRst {
		platform = "app"
	}

	sourceInterface, _ := c.Get("source")
	source := sourceInterface.(int)
	prjIDInterface, _ := c.Get("prjID")
	prjID := prjIDInterface.(int)
	tokenInterface, _ := c.Get("token")
	token := tokenInterface.(string)
	u, ok := c.Get("access_user")
	if !ok {
		appG.Response(0, http.StatusUnauthorized, "invalid_member", "")
		return
	}

	member := u.(*models.EntMemberMembers)

	entMemberID := member.EntMemberID

	arrData := language_service.ProcessUpdateMemberDeviceLanguagev1Struct{
		AccessToken: token,
		LangCode:    form.LangCode,
		SourceID:    source,
		PrjID:       prjID,
		Platform:    platform,
	}
	os, _, err := language_service.ProcessUpdateMemberDeviceLanguagev1(arrData)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	if os != "" {
		// begin transaction
		tx := models.Begin()

		// start process group subscription

		groupName := "LANG_" + strings.ToUpper(form.LangCode) + "-" + strconv.Itoa(arrData.PrjID)
		arrProcessPnData := base.ProcessMemberPushNotificationGroupStruct{
			GroupName: groupName,
			Os:        os,
			MemberID:  entMemberID,
			RegID:     form.PushNotification,
			PrjID:     prjID,
			SourceID:  source,
		}
		base.ProcessMemberPushNotificationGroup(tx, "removeAllIndPrevLangCodeRegID", arrProcessPnData)
		// end process group subscription

		models.Commit(tx)
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}
