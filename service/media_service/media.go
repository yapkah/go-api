package media_service

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/setting"
)

// MediaData struct
type MediaData struct {
	FullURL       string `json:"fullURL"`
	Domain        string `json:"domain"`
	FileDirectory string `json:"fileDirectory"`
}

// SuccessResponse struct
type SuccessResponse struct {
	StatusCode int       `json:"statusCode"`
	Status     string    `json:"status"`
	Message    []string  `json:"message"`
	Data       MediaData `json:"data"`
}

// ErrorResponse struct
type ErrorResponse struct {
	StatusCode int         `json:"statusCode"`
	Status     string      `json:"status"`
	Message    []string    `json:"message"`
	Data       interface{} `json:"data"`
}

// UploadMedia func
func UploadMedia(file multipart.File, fileName string, module string, prefixName string, fileSize string, prjCode string) (*MediaData, error) {
	var err error
	mu, err := models.GetSysGeneralSetupByID("upload_file_url_setting")
	if err != nil {
		return nil, err
	}
	if mu == nil || mu.SettingValue1 == "" {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.MEDIA_URL_NOT_FOUND, Msg: "missing_upload_file_url_setting"}
	}

	url := mu.SettingValue1

	appName := prjCode
	if prjCode == "" {
		appName = setting.Cfg.Section("custom").Key("AppName").String()
	}
	data := map[string]string{
		"projectCode": appName,
		"replace":     "0",
		"module":      module,
		"prefixName":  prefixName,
		"fileMaxSize": fileSize,
		"debug":       "true",
	}
	header := map[string]string{
		"Authorization": mu.SettingValue3,
		// "Content-Type":  "application/json;multipart/form-data;",
	}

	fileData := map[string]base.FileStruct{
		"multiUpload": base.FileStruct{
			File:     file,
			FileName: fileName,
		},
	}
	data1 := data
	data1["fileName"] = fileName
	dataJson, _ := json.Marshal(data1)
	dataString := string(dataJson)
	arrLogData := models.AddGeneralApiLogStruct{
		PrjConfigCode: "upload_file_url_setting",
		URLLink:       url,
		ApiType:       "upload_file_url_setting",
		Method:        "POST",
		DataSent:      dataString,
	}
	AddGeneralApiLogRst, _ := models.AddGeneralApiLog(arrLogData)

	res, err := base.MultiPartPost(url, header, data, fileData)

	resJson, _ := json.Marshal(res)
	resString := string(resJson)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " general_api_log.id = ? ", CondValue: AddGeneralApiLogRst.ID},
	)
	updateColumn := map[string]interface{}{
		"data_received": resString,
	}
	_ = models.UpdatesFn("general_api_log", arrCond, updateColumn, false)

	if err != nil {
		return nil, err
	}

	if res.Body == "" {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.MEDIA_UPLOAD_FILE_ERROR, Data: map[string]interface{}{"return": res}}
	}

	if res.StatusCode != http.StatusOK {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.MEDIA_UPLOAD_FILE_ERROR, Data: map[string]interface{}{"response": res.Body}}
	}

	var response SuccessResponse
	err = json.Unmarshal([]byte(res.Body), &response)

	return &response.Data, nil
}

// MediaValidation func
func MediaValidation(file multipart.File, header *multipart.FileHeader, fileType string) error {
	settingID := "upload_" + fileType + "_setting"
	arrMediaSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil {
		base.LogErrorLog("MediaValidation-failed_to_get_setting_sql", err.Error(), settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "something_went_wrong"}
	}
	if arrMediaSetting.ID < 1 {
		base.LogErrorLog("MediaValidation-setting_is_missing ", settingID, fileType, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "something_went_wrong"}
	}

	extRst := false
	sliceOfMediaExtension := strings.Split(arrMediaSetting.SettingValue1, ",")
	for _, v1 := range sliceOfMediaExtension {
		mediaExt := strings.Contains(header.Header["Content-Type"][0], v1)
		if mediaExt {
			extRst = true
			break
		}
	}

	if extRst == false {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "please_upload_only_" + arrMediaSetting.SettingValue1 + "_" + fileType + "_" + header.Header["Content-Type"][0] + "_is_uploaded"}
	}

	mediaSize, err := strconv.ParseInt(arrMediaSetting.SettingValue2, 10, 64)
	if err != nil {
		base.LogErrorLog("MediaValidation-mediasize_problem ", err.Error(), settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "something_went_wrong"}
	}

	if header.Size > mediaSize {
		mediaFileSizeString := strings.Replace(arrMediaSetting.SettingValue2, "000000", "", -1)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "please_upload_less_than_" + mediaFileSizeString + "_mb"}
	}
	return nil
}

// UploadApp func
func UploadApp(file multipart.File, fileName string) (*MediaData, error) {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " name = ?", CondValue: "admin"},
	)
	arrAppKeys, err := models.GetApiKeysFn(arrCond, "", false)

	if len(arrAppKeys) < 1 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.MEDIA_UPLOAD_FILE_ERROR, Msg: "missing_admin_api_keys"}
	}

	header := map[string]string{
		"X-Authorization": arrAppKeys[0].Key,
		// "Content-Type":  "application/json;multipart/form-data;",
	}

	fileData := map[string]base.FileStruct{
		"file": base.FileStruct{
			File:     file,
			FileName: fileName,
		},
	}
	extraMultiPartPostSetting := base.ExtraSettingStruct{
		InsecureSkipVerify: true,
	}
	adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()
	url := adminServerDomain + "/api/app/upload"
	res, err := base.MultiPartPostV2(url, header, map[string]string{}, fileData, extraMultiPartPostSetting)

	if err != nil {
		return nil, err
	}

	if res.Body == "" {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.MEDIA_UPLOAD_FILE_ERROR, Data: map[string]interface{}{"return": res}}
	}

	if res.StatusCode != http.StatusOK {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.MEDIA_UPLOAD_FILE_ERROR, Data: map[string]interface{}{"response": res.Body}}
	}

	var response SuccessResponse
	err = json.Unmarshal([]byte(res.Body), &response)

	return &response.Data, nil
}
