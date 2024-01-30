package base

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/setting"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// SplitData Separate data into several size
func SplitData(data []interface{}, size int) [][]interface{} {
	var chunkSet [][]interface{}
	var chunk []interface{}

	for len(data) > size {
		chunk, data = data[:size], data[size:]
		chunkSet = append(chunkSet, chunk)
	}
	if len(data) > 0 {
		chunkSet = append(chunkSet, data[:])
	}

	return chunkSet
}

// PasswordChecking func
func PasswordChecking(s string) bool {
	var number, letter, special bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsLetter(c):
			letter = true
		// if unicode.IsUpper(c) {
		// 	upper = true
		// }
		// if unicode.IsLower(c) {
		// 	lower = true
		// }
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
			// default:
			// 	return false
		}
	}

	if number && letter && special {
		return true
	}

	return false
}

// SecondaryPinChecking func
func SecondaryPinChecking(s string) bool {
	b := true
	for _, c := range s {
		if c < '0' || c > '9' {
			b = false
			break
		}
	}

	return b
}

// UsernameChecking func
func UsernameChecking(s string) string {
	var number, letter, other, space bool
	for _, c := range s {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			letter = true
		// case unicode.IsNumber(c):
		case unicode.IsDigit(c):
			number = true
		case unicode.IsSpace(c):
			space = true
		// case string(c) == "_":
		// case unicode.IsLetter(c):
		default:
			other = true
		}
	}

	if number && !letter {
		return e.GetMsg(e.USER_ID_MUST_CONTAIN_AT_LEAST_ONE_ALPHABET)
	}

	if space {
		return e.GetMsg(e.USER_ID_CANNOT_CONTAIN_SPACE)
	}

	if other {
		return e.GetMsg(e.USER_ID_ONLY_ACCEPT_ALPHABET_AND_NUMBER)
	}

	return ""
}

// FirstNameChecking func
func FirstNameChecking(s string) string {
	var number, letter, other, space bool
	for _, c := range s {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			letter = true
		case unicode.IsDigit(c):
			number = true
		case unicode.IsSpace(c):
			space = true
		case unicode.Is(unicode.Han, c):
			letter = true
		default:
			other = true
		}
	}

	if number && !letter {
		return "first_name_must_contain_at_least_one_alphabet"
	}

	if other {
		return "first_name_can_only_contain_alphabet_and_number"
	}

	if space {
		return ""
	}

	return ""
}

// AlphaNumericOnly func
func AlphaNumericOnly(s string) bool {
	var number, letter, space, other bool
	for _, c := range s {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			letter = true
		case unicode.IsDigit(c):
			number = true
		case unicode.IsSpace(c):
			space = true
		default:
			other = true
		}
	}

	if !letter && !number {
		return false
	}
	if other {
		return false
	}
	if space {
		return false
	}

	return true
}

// AlphaNumericCertainCharactersOnly func
func AlphaNumericCertainCharactersOnly(s string, character []string) bool {
	var number, letter, space, other bool
	for _, c := range s {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || helpers.StringInSlice(string(c), character):
			letter = true
		case unicode.IsDigit(c):
			number = true
		case unicode.IsSpace(c):
			space = true
		default:
			other = true
		}
	}

	if !letter && !number {
		return false
	}
	if other {
		return false
	}
	if space {
		return false
	}

	return true
}

// NoSpace func
func NoSpace(s string) bool {
	var space bool
	for _, c := range s {
		switch {
		case unicode.IsSpace(c):
			space = true
		}
	}

	if space {
		return false
	}

	return true
}

// Bcrypt func
func Bcrypt(str string) (string, error) {
	data, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CheckBcryptPassword func
// check bcrypt hashed password
func CheckBcryptPassword(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_PASSWORD}
	}
	return nil
}

// CheckMd5SecondaryPin func
// check md5 hashed secondary pin
func CheckMd5SecondaryPin(secondaryPinMd5 string, secondaryPin string) error {
	hash := md5.Sum([]byte(secondaryPin))
	if hex.EncodeToString(hash[:]) != secondaryPinMd5 {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_SECONDARY_PIN}
	}
	return nil
}

// GenerateRandomString func
func GenerateRandomString(length int, charSet string) string {
	if charSet == "" {
		charSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(b)
}

// CheckImageType func
func CheckImageType(mime string) bool {
	var supported = []string{"image/png", "image/jpeg"}
	for _, s := range supported {
		if s == mime {
			return true
		}
	}
	return false
}

// TemplateReplace func
func TemplateReplace(text string, data interface{}) (string, error) {
	var d bytes.Buffer
	tpl, err := template.New("text").Parse(text)
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.TEXT_PARSE_FAILED, Data: err}
	}

	err = tpl.Execute(&d, data)
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.TEXT_EXECUTE_FAILED, Data: err}
	}
	return d.String(), nil
}

// APIResponse struct
type APIResponse struct {
	Status     string
	StatusCode int
	Header     http.Header
	Body       string
}

// RequestAPI func
func RequestAPI(method, url string, header map[string]string, body map[string]interface{}, resStruct interface{}) (*APIResponse, error) {

	var (
		req     *http.Request
		err     error
		reqBody []byte
	)

	switch method {
	case "GET":
		if body != nil {
			reqBody, err = json.Marshal(body)
			if err != nil {
				models.ErrorLog("RequestAPI_GET_failed_in_json_Marshal_body", err.Error(), body)
				return nil, err
			}
		} else {
			reqBody = nil
		}

		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))

		// data add
		q := req.URL.Query()
		if body != nil {
			for k, d := range body {
				bd, ok := d.(string)
				if !ok {
					models.ErrorLog("RequestAPI_GET_failed_in_add_query_data_string_assertion", err.Error(), d)
					return nil, errors.New("invalid param " + k)
				}
				q.Add(k, bd)
			}
		}

		// query encode
		req.URL.RawQuery = q.Encode()

	case "POST":
		if body != nil {
			reqBody, err = json.Marshal(body)

			if err != nil {
				models.ErrorLog("RequestAPI_POST_failed_in_json_Marshal_body", err.Error(), body)
				return nil, err
			}
		} else {
			reqBody = nil
		}

		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))

	default:
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.REQUEST_API_INVALID_METHOD}
	}

	for k, h := range header {
		req.Header.Set(k, h)
	}

	if err != nil {
		models.ErrorLog("RequestAPI_failed_in_NewRequest", err.Error(), nil)
		return nil, err
	}

	// client := &http.Client{}
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	var headerDataString string
	if header != nil {
		headerDataJson, _ := json.Marshal(header)
		headerDataString = string(headerDataJson)
	}
	dataJson, _ := json.Marshal(body)
	dataString := string(dataJson)
	arrLogData := models.AddGeneralApiLogStruct{
		URLLink:  url,
		ApiType:  "RequestAPI_" + method,
		Method:   method,
		DataSent: headerDataString + dataString,
	}
	AddGeneralApiLogRst, _ := models.AddGeneralApiLog(arrLogData)

	resp, err := client.Do(req)
	if err != nil {
		models.ErrorLog("RequestAPI_failed_in_client_Do", err.Error(), req)
		return nil, err
	}

	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		models.ErrorLog("RequestAPI_failed_in_ioutil_ReadAll", err.Error(), resp.Body)
		return nil, err
	}

	if string(resBody) != "" && resStruct != nil {
		err = json.Unmarshal(resBody, resStruct)
		if err != nil {
			models.ErrorLog("RequestAPI_failed_in_json_Unmarshal_resBody", err.Error(), resBody)
			return nil, err
		}
	}

	response := &APIResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       string(resBody),
	}

	resJson, _ := json.Marshal(response)
	resString := string(resJson)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " general_api_log.id = ? ", CondValue: AddGeneralApiLogRst.ID},
	)
	updateColumn := map[string]interface{}{
		"data_received": resString,
	}
	_ = models.UpdatesFn("general_api_log", arrCond, updateColumn, false)

	return response, nil
}

// FileStruct MultiPartPost file type
type FileStruct struct {
	File     io.Reader
	FileName string
}

// MultiPartPost func
func MultiPartPost(url string, header map[string]string, body map[string]string, file map[string]FileStruct) (*APIResponse, error) {

	var (
		req *http.Request
		err error
		b   bytes.Buffer
	)

	w := multipart.NewWriter(&b)

	for k, d := range body {
		err = w.WriteField(k, d)
		if err != nil {
			models.ErrorLog("MultiPartPost-WriteField", err.Error(), body)
			return nil, err
		}
	}

	for k, f := range file {
		fw, err := w.CreateFormFile(k, f.FileName)
		if err != nil {
			models.ErrorLog("MultiPartPost-CreateFormFile", err.Error(), file)
			return nil, err
		}
		io.Copy(fw, f.File)
	}

	contentType := w.FormDataContentType()
	w.Close()
	req, err = http.NewRequest("POST", url, &b)

	for k, h := range header {
		req.Header.Set(k, h)
	}
	req.Header.Set("Content-Type", contentType)

	if err != nil {
		models.ErrorLog("MultiPartPost-NewRequest", err.Error(), req)
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		models.ErrorLog("MultiPartPost-clientDo", err.Error(), req)
		return nil, err
	}

	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		models.ErrorLog("MultiPartPost-ioutilReadAll", err.Error(), resp)
		return nil, err
	}

	response := &APIResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       string(resBody),
	}

	return response, nil
}

// PostFormRequestApi func
func PostFormRequestApi(url string, header map[string]string, body map[string]string, resStruct interface{}) (*APIResponse, error) {

	var (
		req *http.Request
		err error
		b   bytes.Buffer
	)

	w := multipart.NewWriter(&b)

	for k, d := range body {
		err = w.WriteField(k, d)
		if err != nil {
			models.ErrorLog("PostFormRequestApi-WriteField", err.Error(), body)
			return nil, err
		}
	}

	contentType := w.FormDataContentType()
	w.Close()
	req, err = http.NewRequest("POST", url, &b)

	for k, h := range header {
		req.Header.Set(k, h)
	}
	req.Header.Set("Content-Type", contentType)

	if err != nil {
		models.ErrorLog("PostFormRequestApi-NewRequest", err.Error(), req)
		return nil, err
	}

	client := &http.Client{}
	var headerDataString string
	if header != nil {
		headerDataJson, _ := json.Marshal(header)
		headerDataString = string(headerDataJson)
	}
	dataJson, _ := json.Marshal(body)
	dataString := string(dataJson)
	arrLogData := models.AddGeneralApiLogStruct{
		URLLink:  url,
		ApiType:  "RequestAPI_POST",
		Method:   "POST",
		DataSent: headerDataString + dataString,
	}
	AddGeneralApiLogRst, _ := models.AddGeneralApiLog(arrLogData)

	resp, err := client.Do(req)
	if err != nil {
		models.ErrorLog("PostFormRequestApi-clientDo", err.Error(), req)
		return nil, err
	}

	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		models.ErrorLog("PostFormRequestApi-ioutilReadAll", err.Error(), resp)
		return nil, err
	}

	if string(resBody) != "" && resStruct != nil {
		err = json.Unmarshal(resBody, resStruct)
		if err != nil {
			models.ErrorLog("PostFormRequestApi-failed_in_json_Unmarshal_resBody", err.Error(), string(resBody))
			return nil, err
		}
	}
	response := &APIResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       string(resBody),
	}

	resJSON, _ := json.Marshal(response)
	resString := string(resJSON)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " general_api_log.id = ? ", CondValue: AddGeneralApiLogRst.ID},
	)
	updateColumn := map[string]interface{}{
		"data_received": resString,
	}
	_ = models.UpdatesFn("general_api_log", arrCond, updateColumn, false)

	return response, nil
}

// CheckVersioNumberFormat func
func CheckVersioNumberFormat(version string) (err error) {
	re := regexp.MustCompile("^(\\d+)\\.(\\d+)\\.(\\d+)$")

	if re.MatchString(version) != true {
		err = &e.CustomError{HTTPCode: http.StatusOK, Code: e.INVALID_VERSION_NUMBER_FORMAT}
	}

	return
}

// SplitVersionNumber split version number to 3 int
func SplitVersionNumber(version string) (major int, minor int, build int, err error) {
	err = CheckVersioNumberFormat(version)

	if err != nil {
		return
	}

	// split version
	split := strings.SplitN(version, ".", 3)
	major, err = strconv.Atoi(split[0])
	minor, err = strconv.Atoi(split[1])
	build, err = strconv.Atoi(split[2])

	if err != nil {
		return 0, 0, 0, err
	}

	return
}

// DateRangeValidate func
func DateRangeValidate(daterange []string) error {
	if len(daterange) != 2 {
		return &e.CustomError{HTTPCode: http.StatusOK, Code: e.INVALID_DATE_RANGE}
	}
	if daterange[0] > daterange[1] {
		return &e.CustomError{HTTPCode: http.StatusOK, Code: e.INVALID_DATE_RANGE}
	}

	re := regexp.MustCompile("((19|20)\\d\\d)-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])")
	if !re.MatchString(daterange[0]) {
		code := e.INVALID_DATE_RANGE
		return &e.CustomError{HTTPCode: http.StatusOK, Code: code, Msg: e.GetMsg(code) + "yyyy-mm-dd " + daterange[0]}
	}
	if !re.MatchString(daterange[1]) {
		code := e.INVALID_DATE_RANGE
		return &e.CustomError{HTTPCode: http.StatusOK, Code: code, Msg: e.GetMsg(code) + "yyyy-mm-dd " + daterange[1]}
	}

	return nil
}

// SendSlack func
func SendSlack(text string) error {
	prefix := "🥺🥺【" + setting.ServerSetting.RunMode + "】🥺🥺\n"
	url := "https://hooks.slack.com/services/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"
	data := map[string]interface{}{"text": prefix + text}
	_, err := RequestAPI("POST", url, nil, data, nil)

	return err
}

//IntArrayToString func
func IntArrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

//added by kahhou - for api hash
// EncodeMD5 md5 encryption
func EncodeMD5(value string) string {
	m := md5.New()
	m.Write([]byte(value))

	return hex.EncodeToString(m.Sum(nil))
}

// SHA256 func
func SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	sha := hex.EncodeToString(hash[:])
	return sha
}

func HashInput(data map[string]interface{}, hash string) string {
	param, _ := json.Marshal(data)
	valueStr := EncodeMD5(string(param))
	valueStr += hash
	key := SHA256(valueStr)
	return key
}

//end added

func GetHashAlgo() []string {
	hashAlgo := []string{
		"md2",
		"md4",
		"md5",
		"sha1",
		"sha224",
		"sha256",
	}

	return hashAlgo
}

// CheckMigratedAccPassword func
func CheckMigratedAccPassword(hashed string, password string, debug bool) error {

	pieces := strings.Split(hashed, "$")
	if debug {
		fmt.Println("pieces:", pieces)
	}
	if len(pieces) != 4 {
		models.ErrorLog("CheckMigratedPasswordAcc-invalid_hashed_password_format_not_4_slice", hashed, password)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "invalid_password"}
	}
	header := pieces[0]
	iterString := pieces[1]
	salt := pieces[2]
	hash := pieces[3]

	iter, _ := strconv.Atoi(iterString)
	header1stPart := "pbkdf2_"
	charExits := strings.Contains(header, header1stPart)
	if !charExits {
		models.ErrorLog("CheckMigratedPasswordAcc-invalid_hashed_password_format_not_pbkdf2_", hashed, password)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "invalid_password"}
	}
	headerSecPart := strings.Replace(header, header1stPart, "", -1)

	hashAlgo := GetHashAlgo()
	hashMethodCheckingRst := helpers.StringInSlice(headerSecPart, hashAlgo)
	if !hashMethodCheckingRst {
		models.ErrorLog("CheckMigratedPasswordAcc-invalid_hashed_password_hashing_method", hashAlgo, headerSecPart)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "invalid_password"}
	}
	sliceofpasswordbyte := []byte(password)
	sliceofsaltbyte := []byte(salt)
	sDec, err := base64.StdEncoding.DecodeString(hash)

	if err != nil {
		models.ErrorLog("CheckMigratedPasswordAcc-invalid_hashed_base64_decodestring", err.Error(), hash)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "invalid_password"}
	}
	if debug {
		fmt.Println("sDec:", sDec)
	}

	calc := pbkdf2.Key(sliceofpasswordbyte, sliceofsaltbyte, iter, 32, sha256.New)
	if debug {
		fmt.Println("calc:", calc)
	}

	checkingRst := reflect.DeepEqual(sDec, calc)
	if !checkingRst {
		fmt.Println("checkingRst:", checkingRst)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Msg: "invalid_password"}
	}
	return nil
}

type SecondaryPin struct {
	MemId              int
	SecondaryPin       string
	MemberSecondaryPin string
	LangCode           string
}

// check hashed secondary pin--- form input secondary pin, current secondary pin
func (s *SecondaryPin) CheckSecondaryPin() error {

	// if s.LangCode == "" {
	// 	s.LangCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// }

	hash := md5.Sum([]byte(s.SecondaryPin))

	if hex.EncodeToString(hash[:]) != s.MemberSecondaryPin {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_secondary_pin", s.LangCode)}
	}

	return nil
}

// func TelegramMsgStruct
type TelegramMsgStruct struct {
	Group string
	Text  string
}

// func SendTelegramMsgViaBot
func SendTelegramMsgViaBot(arrData TelegramMsgStruct) {
	apiSetting, _ := models.GetSysGeneralSetupByID("telegram_api_setting") // use get request method https://api.telegram.org/bot<token>/getUpdates to get this info

	if apiSetting.InputType1 == "1" {
		var url = apiSetting.InputValue1
		url = url + apiSetting.InputType2 + "/sendMessage"

		var tableSetting map[string]string
		err := json.Unmarshal([]byte(apiSetting.InputValue2), &tableSetting)
		if err == nil {
			chatID := tableSetting[arrData.Group]
			if chatID != "" {
				dtNow := GetCurrentTime("2006-01-02 15:04:05")
				appName := setting.Cfg.Section("custom").Key("AppName").String()
				env := setting.Cfg.Section("telegram").Key("TELEGRAM_ENV").String() + "-GO"
				textTitle := appName + " (" + env + ") \n" + dtNow + "\n"
				text := textTitle + arrData.Text
				requestBody := map[string]interface{}{
					"text":    text,
					"chat_id": chatID,
				}
				res, err_api := RequestAPI("GET", url, nil, requestBody, nil)

				if err_api != nil {
					models.ErrorLog("SendTelegramMsgViaBot_error_in_api_call_before_call", err_api.Error(), nil)
				}

				if res.StatusCode != 200 {
					models.ErrorLog("SendTelegramMsgViaBot_error_in_api_call_after_call", res.Body, nil)
				}
			} else {
				models.ErrorLog("SendTelegramMsgViaBot", "chatID_not_found_"+arrData.Group, apiSetting.InputValue2)
			}
		}
	}
}

// func LogErrorLog. this included call telegram function
func LogErrorLog(data1 interface{}, data2 interface{}, data3 interface{}, callTelegramStatus bool) {
	arrErrorLog := models.ErrorLogV2(data1, data2, data3)

	if callTelegramStatus {
		a, err := json.Marshal(data1)
		var jdata1 string
		if err == nil {
			jdata1 = string(a)
		}

		textMsg := "sys_error_log.id:" + strconv.Itoa(arrErrorLog.ID) + " " + jdata1
		arrTelegramMsg := TelegramMsgStruct{
			Group: "sys_error_log",
			Text:  textMsg,
		}
		SendTelegramMsgViaBot(arrTelegramMsg)
	}
}

type Pagination struct {
	Page      int64
	DataArr   []interface{}
	HeaderArr interface{}
}

func (p *Pagination) PaginationInterfaceV1() interface{} {

	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	curPage := p.Page

	if curPage == 0 {
		curPage = 1
	}

	if p.Page != 0 {
		p.Page--
	}

	totalRecord := len(p.DataArr)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(p.Page), int(limit), totalRecord)

	processArr := p.DataArr[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn := app.ArrDataResponseList{}

	arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(curPage),
		PerPage:               perPage,
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        totalRecord,
		CurrentPageItems:      processArr,
		TableHeaderList:       p.HeaderArr,
	}

	return arrDataReturn
}

func (p *Pagination) PaginationInterfaceV2() interface{} { //without header

	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	curPage := p.Page

	if curPage == 0 {
		curPage = 1
	}

	if p.Page != 0 {
		p.Page--
	}

	totalRecord := len(p.DataArr)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(p.Page), int(limit), totalRecord)

	processArr := p.DataArr[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn := app.ArrDataResponseDefaultList{}

	arrDataReturn = app.ArrDataResponseDefaultList{
		CurrentPage:           int(curPage),
		PerPage:               perPage,
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        totalRecord,
		CurrentPageItems:      processArr,
	}

	return arrDataReturn
}

type ExtraSettingStruct struct {
	InsecureSkipVerify bool
}

// RequestAPI func
func RequestAPIV2(method, url string, header map[string]string, body map[string]interface{}, resStruct interface{}, extraSetting ExtraSettingStruct) (*APIResponse, error) {

	var (
		req     *http.Request
		err     error
		reqBody []byte
	)

	switch method {
	case "GET":
		if body != nil {
			reqBody, err = json.Marshal(body)
			if err != nil {
				models.ErrorLog("RequestAPI_GET_failed_in_json_Marshal_body", err.Error(), body)
				return nil, err
			}
		} else {
			reqBody = nil
		}

		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))

		// data add
		q := req.URL.Query()
		if body != nil {
			for k, d := range body {
				bd, ok := d.(string)
				if !ok {
					models.ErrorLog("RequestAPI_GET_failed_in_add_query_data_string_assertion", err.Error(), d)
					return nil, errors.New("invalid param " + k)
				}
				q.Add(k, bd)
			}
		}

		// query encode
		req.URL.RawQuery = q.Encode()

	case "POST":
		if body != nil {
			reqBody, err = json.Marshal(body)

			if err != nil {
				models.ErrorLog("RequestAPI_POST_failed_in_json_Marshal_body", err.Error(), body)
				return nil, err
			}
		} else {
			reqBody = nil
		}

		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))

	default:
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.REQUEST_API_INVALID_METHOD}
	}

	for k, h := range header {
		req.Header.Set(k, h)
	}

	if err != nil {
		models.ErrorLog("RequestAPI_failed_in_NewRequest", err.Error(), nil)
		return nil, err
	}

	insecureSkipVerifyStatus := false
	if extraSetting.InsecureSkipVerify {
		insecureSkipVerifyStatus = true
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerifyStatus,
			},
		},
	}

	var headerDataString string
	if header != nil {
		headerDataJson, _ := json.Marshal(header)
		headerDataString = string(headerDataJson)
	}
	dataJson, _ := json.Marshal(body)
	dataString := string(dataJson)
	arrLogData := models.AddGeneralApiLogStruct{
		URLLink:  url,
		ApiType:  "RequestAPI_" + method,
		Method:   method,
		DataSent: headerDataString + dataString,
	}
	AddGeneralApiLogRst, _ := models.AddGeneralApiLog(arrLogData)

	resp, err := client.Do(req)
	if err != nil {
		models.ErrorLog("RequestAPI_failed_in_client_Do", err.Error(), req)
		return nil, err
	}

	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		models.ErrorLog("RequestAPI_failed_in_ioutil_ReadAll", err.Error(), resp.Body)
		return nil, err
	}

	if string(resBody) != "" && resStruct != nil {
		err = json.Unmarshal(resBody, resStruct)
		if err != nil {
			models.ErrorLog("RequestAPI_failed_in_json_Unmarshal_resBody", err.Error(), resBody)
			return nil, err
		}
	}

	response := &APIResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       string(resBody),
	}

	resJson, _ := json.Marshal(response)
	resString := string(resJson)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " general_api_log.id = ? ", CondValue: AddGeneralApiLogRst.ID},
	)
	updateColumn := map[string]interface{}{
		"data_received": resString,
	}
	_ = models.UpdatesFn("general_api_log", arrCond, updateColumn, false)

	return response, nil
}

func GetLatestExchangePriceMovementByTokenType(tokenType string) (float64, error) {
	if strings.ToUpper(tokenType) != "USDS" && strings.ToUpper(tokenType) != "USDT" {
		tokenRate, err := models.GetLatestExchangePriceMovementByTokenType(tokenType)

		if err != nil {
			return float64(1), err
		}

		if tokenRate == 0 {
			tokenRate = float64(1)
		}

		return tokenRate, nil
	} else {
		return float64(1), nil
	}
}

// func LogErrorLogV2. this included call telegram function + diff group
func LogErrorLogV2(data1 interface{}, data2 interface{}, data3 interface{}, callTelegramStatus bool, groupName string) {
	arrErrorLog := models.ErrorLogV2(data1, data2, data3)

	if callTelegramStatus {
		a, err := json.Marshal(data1)
		var jdata1 string
		if err == nil {
			jdata1 = string(a)
		}

		textMsg := "sys_error_log.id:" + strconv.Itoa(arrErrorLog.ID) + " " + jdata1
		arrTelegramMsg := TelegramMsgStruct{
			Group: groupName,
			Text:  textMsg,
		}
		SendTelegramMsgViaBot(arrTelegramMsg)
	}
}

// MultiPartPostV2 func
func MultiPartPostV2(url string, header map[string]string, body map[string]string, file map[string]FileStruct, extraSetting ExtraSettingStruct) (*APIResponse, error) {

	var (
		req *http.Request
		err error
		b   bytes.Buffer
	)

	w := multipart.NewWriter(&b)

	for k, d := range body {
		err = w.WriteField(k, d)
		if err != nil {
			models.ErrorLog("MultiPartPost-WriteField", err.Error(), body)
			return nil, err
		}
	}

	for k, f := range file {
		fw, err := w.CreateFormFile(k, f.FileName)
		if err != nil {
			models.ErrorLog("MultiPartPost-CreateFormFile", err.Error(), file)
			return nil, err
		}
		io.Copy(fw, f.File)
	}

	contentType := w.FormDataContentType()
	w.Close()
	req, err = http.NewRequest("POST", url, &b)

	for k, h := range header {
		req.Header.Set(k, h)
	}
	req.Header.Set("Content-Type", contentType)

	if err != nil {
		models.ErrorLog("MultiPartPost-NewRequest", err.Error(), req)
		return nil, err
	}

	insecureSkipVerifyStatus := false
	if extraSetting.InsecureSkipVerify {
		insecureSkipVerifyStatus = true
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerifyStatus,
			},
		},
	}
	resp, err := client.Do(req)

	if err != nil {
		models.ErrorLog("MultiPartPost-clientDo", err.Error(), req)
		return nil, err
	}

	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		models.ErrorLog("MultiPartPost-ioutilReadAll", err.Error(), resp)
		return nil, err
	}

	response := &APIResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       string(resBody),
	}

	return response, nil
}

func GetLatestPriceMovementByTokenType(tokenType string) (float64, error) {

	var (
		tokenRate float64 = 1
		ETHCode   string  = "ETHUSDT"
		err       error
	)

	if strings.ToUpper(tokenType) != "USDT" && strings.ToUpper(tokenType) != "USDC" && strings.ToUpper(tokenType) != "AP" && strings.ToUpper(tokenType) != "TP" && strings.ToUpper(tokenType) != "RPA" && strings.ToUpper(tokenType) != "RPB" && strings.ToUpper(tokenType) != "PP" && strings.ToUpper(tokenType) != "TS" {

		if strings.ToUpper(tokenType) == "ETHUSDT" {
			arrCryptoMovement, err := models.GetLatestCryptoPriceMovementFn(ETHCode, false)
			if err != nil {
				LogErrorLog("GetLatestPriceMovementByTokenType - GetLatestCryptoPriceMovementFn err", err, arrCryptoMovement, true)
				return float64(1), err
			}

			tokenRate = arrCryptoMovement.Price
		} else {
			tokenRate, err = models.GetLatestPriceMovementByTokenType(tokenType) //default price movement
		}

		if err != nil {
			return float64(1), err
		}

		if tokenRate == 0 {
			tokenRate = float64(1)
		}

		return tokenRate, nil

	} else {
		return tokenRate, nil
	}

}

type PriceMovementIndByTokenTypeStruct struct {
	MemberID  int
	TokenType string
	Limit     int
	Date      time.Time
}

type PriceMovementIndByTokenTypeRst struct {
	// ID         int       `json:"id"`
	TokenPrice float64 `json:"token_price"`
	// BLatest    int       `json:"b_latest"`
	CreatedAt time.Time `json:"created_at"`
}

// GetPriceMovementIndByTokenTypeFn
func GetPriceMovementIndByTokenTypeFn(arrData PriceMovementIndByTokenTypeStruct) ([]PriceMovementIndByTokenTypeRst, error) {
	var arrDateList []string
	var arrDataReturn []PriceMovementIndByTokenTypeRst
	var tokenRate float64
	var transAt time.Time

	dateString := arrData.Date.Format("2006-01-02")
	lastDateT := arrData.Date
	yestDateString := lastDateT.Format("2006-01-02")
	arrDateList = append(arrDateList, yestDateString)

	for i := arrData.Limit; i > 1; i-- {
		yestDateTimeT := AddDurationInString(lastDateT, " -1 day")
		lastDateT = yestDateTimeT
		yestDateString := yestDateTimeT.Format("2006-01-02")
		arrDateList = append(arrDateList, yestDateString)
	}

	arrDefMarketPriceRst, _ := models.GetDefLigaPriceMovementFn(arrDateList, 0, false)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		// models.WhereCondFn{Condition: " liga_price_movement.created_at <= (SELECT created_at FROM liga_price_movement where b_latest = 1)"},
		models.WhereCondFn{Condition: " DATE(liga_price_custom.date) <= ? ", CondValue: dateString},
		models.WhereCondFn{Condition: " liga_price_custom.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " d.id = ? ", CondValue: arrData.MemberID},
	)
	customRate, err := models.GetLigaCustomNetworkPriceFn(arrCond, 7, false)

	if err != nil {
		LogErrorLog("GetPriceMovementIndByTokenTypeFn-GetLigaCustomNetworkPriceFn_failed", err.Error(), arrCond, false)
		return arrDataReturn, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	for arrDateListK, arrDateListV := range arrDateList {
		tokenRate = arrDefMarketPriceRst[arrDateListK].TokenPrice
		transAt = arrDefMarketPriceRst[arrDateListK].CreatedAt

		for _, customRateV := range customRate {
			if customRateV.Date.Format("2006-01-02") == arrDateListV {
				if customRateV.BNetwork == 1 { //include newtwork
					tokenRate = customRateV.TokenPrice
					transAt = customRateV.Date
				} else { //individual only
					if customRateV.MemberId == arrData.MemberID {
						tokenRate = customRateV.TokenPrice
						transAt = customRateV.Date
					}
				}
			}
		}
		//check custom
		arrDataReturn = append(arrDataReturn,
			PriceMovementIndByTokenTypeRst{
				TokenPrice: tokenRate,
				CreatedAt:  transAt,
			},
		)
	}

	return arrDataReturn, nil
}

type PushNotificationContentStruct struct {
	Msg    string            `json:"title"`
	Params map[string]string `json:"params"`
}

type CreateNewPushNotificationGroupStruct struct {
	GroupName string
	Os        string
}

// CallCreateNewPushNotificationGroupApi func
func CallCreateNewPushNotificationGroupApi(arrData CreateNewPushNotificationGroupStruct) error {
	/*
		http://ens.smartblock.pro/api/v2/subscribeToGroup
		ProjectCode:SEC
		DevMode:0
		GroupName:ZH-CN
		Service:JPUSH
	*/

	// start example of arrData
	// arrData['GroupName']
	// end example of arrData

	type apiRstStruct struct {
		Status     string      `json:"status"`
		StatusCode int         `json:"statusCode"`
		Data       interface{} `json:"data"`
		Message    string      `json:"message"`
	}
	var response apiRstStruct

	settingID := "pn_" + arrData.Os + "_create_new_group"
	arrSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil {
		LogErrorLog("CallCreateNewPushNotificationGroupApi-GetSysGeneralSetupByID_failed", err.Error(), settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting == nil {
		LogErrorLog("CallCreateNewPushNotificationGroupApi-GetSysGeneralSetupByID_is_not_set_in_db", settingID, arrSetting, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting.InputType1 == "1" {
		apiURL := arrSetting.InputValue2

		header := map[string]string{
			"x-api-key": arrSetting.InputType3,
		}

		data := map[string]string{
			"ProjectCode": arrSetting.SettingValue1,
			"DevMode":     arrSetting.InputValue1,
			"Service":     arrSetting.InputType2, // JPUSH;
			"GroupName":   arrData.GroupName,
		}

		_, err := PostFormRequestApi(apiURL, header, data, &response)
		if err != nil { // api failed
			LogErrorLog("CallCreateNewPushNotificationGroupApi-RequestAPI_failed", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		// if res.StatusCode != 200 {
		// fmt.Println("res:", res)
		// errMsg, _ := json.Marshal(response.Msg)
		// errMsgStr := string(errMsg)
		// errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		// LogErrorLog("CallCreateNewPushNotificationGroupApi-push_notification_service_is_down", response.Msg, map[string]interface{}{"err": errMsgStr, "response_body": res.Body}, true)
		// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// }

		// fmt.Println("res:", res)
	}

	return nil
}

type CallSubscribePushNotificationToGroupApiStruct struct {
	GroupName string
	Os        string
	RegID     string
}

// CallSubscribePushNotificationToGroupApi func
func CallSubscribePushNotificationToGroupApi(arrData CallSubscribePushNotificationToGroupApiStruct) error {
	/*
		start example of arrData
		arrData['GroupName']
		arrData['Os']
		arrData['RegID']
		end example of arrData

		start input require from api
		http://ens.smartblock.pro/api/v2/subscribeToGroup
		ProjectCode:SEC
		RegID:6c99893d4a37982ff54a0734769700d8e4c084269c8e97701761cd442012ee99
		DevMode:0
		GroupName:EN
		Os:IOS
		Service:JPUSH
		end input require from api
	*/

	// start example of arrData
	// arrData['GroupName']
	// end example of arrData
	type apiRstStruct struct {
		Status     string      `json:"status"`
		StatusCode int         `json:"statusCode"`
		Data       interface{} `json:"data"`
		Message    string      `json:"message"`
	}
	var response apiRstStruct

	settingID := "pn_" + arrData.Os + "_subscribe_to_group"
	arrSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil {
		LogErrorLog("CallSubscribePushNotificationToGroupApi-GetSysGeneralSetupByID_failed", err.Error(), settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting == nil {
		LogErrorLog("CallSubscribePushNotificationToGroupApi-GetSysGeneralSetupByID_is_not_set_in_db", settingID, arrSetting, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting.InputType1 == "1" {
		apiURL := arrSetting.InputValue2

		header := map[string]string{
			"x-api-key": arrSetting.InputType3,
		}

		data := map[string]string{
			"ProjectCode": arrSetting.SettingValue1,
			"DevMode":     arrSetting.InputValue1,
			"Os":          strings.ToUpper(arrData.Os), // IOS
			"Service":     arrSetting.InputType2,       // JPUSH;
			"RegID":       arrData.RegID,
			"GroupName":   arrData.GroupName,
		}

		rst, err := PostFormRequestApi(apiURL, header, data, &response)
		if err != nil { // api failed
			LogErrorLog("CallSubscribePushNotificationToGroupApi-RequestAPI_failed", err.Error(), map[string]interface{}{"err": err, "data": data, "rst": rst}, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// fmt.Println("rst:", rst)

		if rst.StatusCode != 200 {
			if response.Message != "inactive_regID" {
				LogErrorLog("CallSubscribePushNotificationToGroupApi-pn_subscribe_to_group_Api_down", response.Message, response, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		}
	}

	return nil
}

type CallUnsubscribePushNotificationFromGroupApiStruct struct {
	GroupName string
	Os        string
	RegID     string
}

// CallUnsubscribePushNotificationFromGroupApi func
func CallUnsubscribePushNotificationFromGroupApi(arrData CallUnsubscribePushNotificationFromGroupApiStruct) error {
	/*
		start example of arrData
		arrData['GroupName']
		arrData['Os']
		arrData['RegID']
		end example of arrData

		start input require from api
		http://ens.smartblock.pro/api/v2/unsubscribeFromGroup
		ProjectCode:BOD
		RegID:6c99893d4a37982ff54a0734769700d8e4c084269c8e97701761cd442012ee99
		DevMode:0
		GroupName:ZH-CN
		Os:IOS
		Service:JPUSH
		end input require from api
	*/

	// start example of arrData
	// arrData['GroupName']
	// end example of arrData

	type apiRstStruct struct {
		Status     string      `json:"status"`
		StatusCode int         `json:"statusCode"`
		Data       interface{} `json:"data"`
		Message    string      `json:"message"`
	}
	var response apiRstStruct

	settingID := "pn_" + arrData.Os + "_unsubscribe_from_group"
	arrSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil {
		LogErrorLog("CallUnsubscribePushNotificationFromGroupApi-GetSysGeneralSetupByID_failed", err.Error(), settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting == nil {
		LogErrorLog("CallUnsubscribePushNotificationFromGroupApi-GetSysGeneralSetupByID_is_not_set_in_db", settingID, arrSetting, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting.InputType1 == "1" {
		apiURL := arrSetting.InputValue2

		header := map[string]string{
			"x-api-key": arrSetting.InputType3,
		}

		data := map[string]string{
			"ProjectCode": arrSetting.SettingValue1,
			"DevMode":     arrSetting.InputValue1,
			"Os":          strings.ToUpper(arrData.Os), // IOS
			"Service":     arrSetting.InputType2,       // JPUSH;
			"RegID":       arrData.RegID,
			"GroupName":   arrData.GroupName,
		}

		rst, err := PostFormRequestApi(apiURL, header, data, &response)
		if err != nil { // api failed
			LogErrorLog("CallUnsubscribePushNotificationFromGroupApi-RequestAPI_failed", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		fmt.Println("rst:", rst)
		// if res.StatusCode != 200 {
		// fmt.Println("res:", res)
		// errMsg, _ := json.Marshal(response.Msg)
		// errMsgStr := string(errMsg)
		// errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		// LogErrorLog("CallUnsubscribePushNotificationFromGroupApi-push_notification_service_is_down", response.Msg, map[string]interface{}{"err": errMsgStr, "response_body": res.Body}, true)
		// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// }

		// fmt.Println("res:", res)
	}

	return nil
}

type CallSendPushNotificationIndApiStruct struct {
	Os      string
	RegID   string
	Subject string
	Msg     string
	CusMsg  string
}

// CallSendPushNotificationIndApi func
func CallSendPushNotificationIndApi(arrData CallSendPushNotificationIndApiStruct) error {
	/*
		start example of arrData
		arrData['os'] // required
		arrData['regID'] // required
		arrData['subject'] // required
		arrData['msg'] // required
		arrData['cusMsg'] // optional
		end example of arrData

		start input require from api
		http://ens.smartblock.pro/api/v2/SendPushNotification
		ProjectCode:SEC
		Service:JPUSH
		Os:ANDROID
		RegID:6c99893d4a37982ff54a0734769700d8e4c084269c8e97701761cd442012ee99
		Subject:Check out the latest info!
		Msg:Check out the latest info!
		DevMode:0
		cusMsg:`{"abc": "def"}`
		end input require from api
	*/

	// start example of arrData
	// arrData['GroupName']
	// end example of arrData

	type apiRstStruct struct {
		Status     string      `json:"status"`
		StatusCode int         `json:"statusCode"`
		Data       interface{} `json:"data"`
		Message    string      `json:"message"`
	}
	var response apiRstStruct

	settingID := "pn_" + arrData.Os + "_send_pn_individual"
	arrSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil {
		LogErrorLog("CallSendPushNotificationIndApi-GetSysGeneralSetupByID_failed", err.Error(), settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting == nil {
		LogErrorLog("CallSendPushNotificationIndApi-GetSysGeneralSetupByID_is_not_set_in_db", settingID, arrSetting, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting.InputType1 == "1" {
		apiURL := arrSetting.InputValue2

		header := map[string]string{
			"x-api-key": arrSetting.InputType3,
		}

		data := map[string]string{
			"ProjectCode": arrSetting.SettingValue1,
			"DevMode":     arrSetting.InputValue1,
			"Os":          strings.ToUpper(arrData.Os), // IOS
			"Service":     arrSetting.InputType2,       // JPUSH;
			"RegID":       arrData.RegID,
			"Subject":     arrData.Subject,
			"Msg":         arrData.Msg,
		}

		if arrData.CusMsg != "" {
			data["CusMsg"] = arrData.CusMsg
		}

		res, err := PostFormRequestApi(apiURL, header, data, &response)
		if err != nil { // api failed
			LogErrorLog("CallSendPushNotificationIndApi-RequestAPI_failed", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// fmt.Println(res)
		if res.StatusCode != 200 {
			if response.Message != "inactive_regID" {
				LogErrorLog("CallSendPushNotificationIndApi-PostFormRequestApi_down", response.Message, response, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
		}
	}

	return nil
}

type CallSendPushNotificationInGroupApiStruct struct {
	GroupName string
	Subject   string
	Msg       string
	CusMsg    string
}

// CallSendPushNotificationIndApi func
func CallSendPushNotificationInGroupApi(arrData CallSendPushNotificationInGroupApiStruct) error {
	/*
		start example of arrData
		arrData['groupName'] // required
		arrData['subject'] // required
		arrData['msg'] // required
		arrData['cusMsg'] // optional
		end example of arrData

		start input require from api
		http://ens.smartblock.pro/api/v2/SendPushNotificationByGroup
		ProjectCode:BOD
		Service:JPUSH
		GroupName:EN
		Subject:Checkout our latest info!
		Msg:Checkout our latest promotion!
		DevMode:
		cusMsg:`{"abc": "def"}`
		end input require from api
	*/

	// start example of arrData
	// arrData['GroupName']
	// end example of arrData

	type apiRstStruct struct {
		Status     string      `json:"status"`
		StatusCode int         `json:"statusCode"`
		Data       interface{} `json:"data"`
		Message    string      `json:"message"`
	}
	var response apiRstStruct

	settingID := "pn_send_pn_in_group"
	arrSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil {
		LogErrorLog("CallSendPushNotificationInGroupApi-GetSysGeneralSetupByID_failed", err.Error(), settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting == nil {
		LogErrorLog("CallSendPushNotificationInGroupApi-GetSysGeneralSetupByID_is_not_set_in_db", settingID, arrSetting, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrSetting.InputType1 == "1" {
		apiURL := arrSetting.InputValue2

		header := map[string]string{
			"x-api-key": arrSetting.InputType3,
		}

		data := map[string]interface{}{
			"ProjectCode": arrSetting.SettingValue1,
			"DevMode":     arrSetting.InputValue1,
			"Service":     arrSetting.InputType2, // JPUSH;
			"GroupName":   arrData.GroupName,
			"Subject":     arrData.Subject,
			"Msg":         arrData.Msg,
			"CusMsg":      arrData.CusMsg,
		}

		rst, err := RequestAPI(arrSetting.SettingValue2, apiURL, header, data, &response)
		if err != nil { // api failed
			LogErrorLog("CallSendPushNotificationInGroupApi-RequestAPI_failed", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		fmt.Println("rst:", rst)
		// if res.StatusCode != 200 {
		// fmt.Println("res:", res)
		// errMsg, _ := json.Marshal(response.Msg)
		// errMsgStr := string(errMsg)
		// errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		// LogErrorLog("CallSendPushNotificationInGroupApi-push_notification_service_is_down", response.Msg, map[string]interface{}{"err": errMsgStr, "response_body": res.Body}, true)
		// return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// }

		// fmt.Println("res:", res)
	}

	return nil
}

type ProcessMemberPushNotificationGroupStruct struct {
	GroupName string
	Os        string
	RegID     string
	MemberID  int
	PrjID     int
	SourceID  int
}

// func ProcessMemberPushNotificationGroup
func ProcessMemberPushNotificationGroup(tx *gorm.DB, action string, arrData ProcessMemberPushNotificationGroupStruct) {
	arrSybPNToGrpApi := CallSubscribePushNotificationToGroupApiStruct{
		GroupName: arrData.GroupName,
		Os:        arrData.Os,
		RegID:     arrData.RegID,
	}
	err := CallSubscribePushNotificationToGroupApi(arrSybPNToGrpApi)

	if err == nil {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "group_name = ?", CondValue: arrData.GroupName},
			models.WhereCondFn{Condition: "push_noti_token = ?", CondValue: arrData.RegID},
			models.WhereCondFn{Condition: "prj_id = ?", CondValue: arrData.PrjID},
		)
		arrExistingAppMemPnGrp, _ := models.GetAppMemberPnGroupFn(arrCond, false)

		if action != "logoutAction" {
			if len(arrExistingAppMemPnGrp) < 1 {
				// start save data in
				arrCrtSubPN := models.AppMemberPnGroup{
					PrjID:         arrData.PrjID,
					GroupName:     arrData.GroupName,
					OS:            arrData.Os,
					PushNotiToken: arrData.RegID,
				}

				if arrData.MemberID > 0 {
					arrCrtSubPN.MemberID = arrData.MemberID
				}
				models.AddAppMemberPnGroup(arrCrtSubPN)
			}
		}

		// start remove all prev lang code group record in app_member_pn_group with regID
		if action == "removeAllIndPrevLangCodeRegID" {
			// start retrieve all the prev lang code group record
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "push_noti_token = ?", CondValue: arrData.RegID},
				models.WhereCondFn{Condition: "os = ?", CondValue: arrData.Os},
				models.WhereCondFn{Condition: "group_name LIKE 'LANG_%' AND group_name != ?", CondValue: arrData.GroupName},
				models.WhereCondFn{Condition: "prj_id = ?", CondValue: arrData.PrjID},
			)
			arrAppMemPnGrp, _ := models.GetAppMemberPnGroupFn(arrCond, false)

			if len(arrAppMemPnGrp) > 0 {
				for _, arrAppMemPnGrpV := range arrAppMemPnGrp {
					arrUnsubPN := CallUnsubscribePushNotificationFromGroupApiStruct{
						GroupName: arrAppMemPnGrpV.GroupName,
						Os:        arrAppMemPnGrpV.OS,
						RegID:     arrAppMemPnGrpV.PushNotiToken,
					}
					err = CallUnsubscribePushNotificationFromGroupApi(arrUnsubPN)
					if err == nil {
						// start delete app_member_pn_group records
						arrUnsubPNDelFn := make([]models.WhereCondFn, 0)
						arrUnsubPNDelFn = append(arrUnsubPNDelFn,
							models.WhereCondFn{Condition: "id = ?", CondValue: arrAppMemPnGrpV.ID},
						)
						err = models.DeleteFn("app_member_pn_group", arrUnsubPNDelFn, false)
						if err != nil {
							LogErrorLog("ProcessMemberPushNotificationGroup-delete_app_member_pn_group_failed", err.Error(), arrUnsubPNDelFn, true)
						}
						// end delete app_member_pn_group records
					}
				}
			}
			// end retrieve all the prev lang code group record
		}
		// end remove all prev lang code group record in app_member_pn_group with regID
	}
}

type ProcessPushNotificationDataV1Struct struct {
	EntMemberID   int
	PnMsgID       string
	ApiKeysName   string
	SourceID      int
	MsgTitle      string
	MsgTitleValue map[string]string
	Msg           string
	MsgValue      map[string]string
	GroupName     string
	LangCode      string
	CustomData    string
	ArrFn         string
}

func ProcessPushNotificationDataV1(arrData ProcessPushNotificationDataV1Struct, translateTitle, translateMsg, saveSysNotification bool) {
	// support send push notification in group

	// start data needed
	//        arrData['member_id'] // optional. if this is empty, all member is send.
	//        arrData['pnMsgID'] // required
	//        arrData['msgTitle'] // required
	//        arrData['msg'] // required
	//        arrData['groupName'] // optional
	//        arrData['langCode'] // optional
	//        arrData['customData'] // optional
	//        arrData['arrFn'][] = ['delay', Carbon::now()->addMinutes(10)]  // optional
	// end data needed

	if arrData.ApiKeysName == "" {
		arrData.ApiKeysName = "app"
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "api_keys.name = ?", CondValue: arrData.ApiKeysName},
		models.WhereCondFn{Condition: "api_keys.active = ?", CondValue: 1},
	)
	apiKeyRst, _ := models.GetApiKeysFn(arrCond, "", false)

	if len(apiKeyRst) < 0 {
		LogErrorLog("ProcessPushNotificationDataV1-GetApiKeysFn_no_data", arrCond, nil, true)
		return
	}

	apiKeyID := apiKeyRst[0].ID

	arrLang, _ := models.GetLanguageList()
	if len(arrLang) < 1 {
		return
	}
	settingID := "default_app_notification_setting"
	arrSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil {
		LogErrorLog("ProcessPushNotificationDataV1-GetSysGeneralSetupByID_failed", err.Error(), settingID, true)
		return
	}

	if arrSetting == nil {
		LogErrorLog("ProcessPushNotificationDataV1-GetSysGeneralSetupByID_is_not_set_in_db", settingID, arrSetting, true)
		return
	}
	defLangCode := arrSetting.SettingValue2
	// arrOS := strings.Split(arrSetting.InputType3, ",")

	pnSub := PushNotificationContentStruct{
		Msg:    arrData.MsgTitle,
		Params: arrData.MsgTitleValue,
	}
	encodedPNSub, _ := json.Marshal(pnSub)

	pnMsg := PushNotificationContentStruct{
		Msg:    arrData.Msg,
		Params: arrData.MsgValue,
	}
	encodedPNMsg, _ := json.Marshal(pnMsg)

	if arrData.EntMemberID < 1 {
		if arrData.GroupName == "" {
			// this will send default group name (LANG)
			for _, arrLangV := range arrLang {
				msgTitle := arrData.MsgTitle
				msg := arrData.Msg

				if translateTitle {
					msgTitle = helpers.TranslateV2(arrData.MsgTitle, arrLangV.Locale, arrData.MsgTitleValue)
				}
				if translateMsg {
					msg = helpers.TranslateV2(arrData.Msg, arrLangV.Locale, arrData.MsgValue)
				}
				arrCallSendPushNotificationInGroupApiData := CallSendPushNotificationInGroupApiStruct{
					GroupName: strings.ToUpper("LANG_" + arrLangV.Locale),
					Subject:   msgTitle,
					Msg:       msg,
					CusMsg:    arrData.CustomData,
				}
				err = CallSendPushNotificationInGroupApi(arrCallSendPushNotificationInGroupApiData)
				if err != nil {
					LogErrorLog("ProcessPushNotificationDataV1-CallSendPushNotificationInGroupApi_member_id_group_name_is_given", err.Error(), arrCallSendPushNotificationInGroupApiData, true)
					return
				}

				if saveSysNotification {

					arrCrtSysNoti := models.AddSysNotificationStruct{
						ApiKeyID:     apiKeyID,
						Type:         "member",
						MemberID:     arrData.EntMemberID,
						Title:        string(encodedPNSub),
						Msg:          string(encodedPNMsg),
						LangCode:     arrLangV.Locale,
						CustMsg:      arrData.CustomData,
						BShow:        1,
						PNSendStatus: 0,
						Status:       "A",
						CreatedBy:    strconv.Itoa(arrData.EntMemberID),
					}
					_, err := models.AddSysNotification(arrCrtSysNoti)
					if err != nil {
						LogErrorLog("ProcessPushNotificationDataV1-AddSysNotification_member_id_group_name_is_given", err.Error(), arrCrtSysNoti, true)
						return
					}
				}
			}
		} else {
			// // this will send msg to specific group. - group name is passed from function caller.
			// // need to get current latest langCode and valid regID
			// // start get all valid and active push noti token
			// arrCond := make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: " b_login = ? ", CondValue: 1},
			// 	models.WhereCondFn{Condition: " t_push_noti_token != '' AND dt_login <= NOW() AND dt_expiry >= NOW() AND b_logout = ? ", CondValue: 0},
			// )
			// arrlatestActiveAppLoginToken, _ := models.GetDistinctAppLoginLogFn(arrCond, true)
			// // end get all valid and active push noti token
			// if len(arrlatestActiveAppLoginToken) > 0 {
			// 	for _, arrlatestActiveAppLoginTokenV := range arrlatestActiveAppLoginToken {
			// 		// start looping on filtering out the inactive push noti token
			// 		arrCond = make([]models.WhereCondFn, 0)
			// 		arrCond = append(arrCond,
			// 			models.WhereCondFn{Condition: " b_login = ? ", CondValue: 0},
			// 			models.WhereCondFn{Condition: " b_logout = ? ", CondValue: 1},
			// 			models.WhereCondFn{Condition: " t_push_noti_token = ? ", CondValue: arrlatestActiveAppLoginTokenV.TPushNotiToken},
			// 		)
			// 		arrLogoutAppLoginToken, _ := models.GetAppLoginLogFn(arrCond, "", true)
			// 		if arrLogoutAppLoginToken.TPushNotiToken == "" {
			// 			// start proccess valid active push noti token
			// 			langCode := defLangCode
			// 			if arrData.LangCode != "" {
			// 				langCode = arrData.LangCode
			// 			} else if arrLogoutAppLoginToken.LanguageID != "" {
			// 				langCode = arrLogoutAppLoginToken.LanguageID
			// 			}
			// 			msgTitle := arrData.MsgTitle
			// 			msg := arrData.Msg
			// 			if translateTitle {
			// 				msgTitle = helpers.TranslateV2(arrData.MsgTitle, langCode, arrData.MsgTitleValue)
			// 			}
			// 			if translateMsg {
			// 				msg = helpers.TranslateV2(arrData.Msg, langCode, arrData.MsgValue)
			// 			}
			// 			arrCallSendPushNotificationIndApiData := CallSendPushNotificationIndApiStruct{
			// 				Os:      arrlatestActiveAppLoginTokenV.TOs,
			// 				RegID:   arrlatestActiveAppLoginTokenV.TPushNotiToken,
			// 				Subject: msgTitle,
			// 				Msg:     msg,
			// 				CusMsg:  arrData.CustomData,
			// 			}
			// 			err = CallSendPushNotificationIndApi(arrCallSendPushNotificationIndApiData)
			// 			if err != nil {
			// 				LogErrorLog("ProcessPushNotificationDataV1-CallSendPushNotificationIndApi_groupName_is_given", err.Error(), arrCallSendPushNotificationIndApiData, true)
			// 				return
			// 			}
			// 			if saveSysNotification {
			// 				encodedMsgTitle, _ := json.Marshal(arrData.MsgTitle)
			// 				encodedMsg, _ := json.Marshal(arrData.Msg)
			// 				arrCrtSysNoti := models.AddSysNotificationStruct{
			// 					Type: "member",
			// 					// PNType      string    `json:"pn_type" gorm:"column:pn_type"`
			// 					MemberID:  arrData.EntMemberID,
			// 					Title:     string(encodedMsgTitle),
			// 					Msg:       string(encodedMsg),
			// 					BShow:     1,
			// 					Status:    "A",
			// 					CreatedBy: strconv.Itoa(arrData.EntMemberID),
			// 				}
			// 				_, err = models.AddSysNotification(arrCrtSysNoti)
			// 				if err != nil {
			// 					LogErrorLog("ProcessPushNotificationDataV1-AddSysNotification_groupName_is_given", err.Error(), arrCrtSysNoti, true)
			// 					return
			// 				}
			// 			}
			// 			// end proccess valid active push noti token
			// 		}
			// 	}
			// }
		}
	} else {
		// member_id_is_given
		// retrieve member token
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " b_login = ? ", CondValue: 1},
			models.WhereCondFn{Condition: " b_logout = ? ", CondValue: 0},
			models.WhereCondFn{Condition: " dt_login <= NOW() AND dt_expiry >= NOW() AND t_user_id = ? ", CondValue: arrData.EntMemberID},
		)
		latestCurrentActiveAppLoginToken, _ := models.GetExistingActiveAppLoginLog(arrData.EntMemberID, uint8(arrData.SourceID), true)
		fmt.Println("latestCurrentActiveAppLoginToken:", latestCurrentActiveAppLoginToken)

		if len(latestCurrentActiveAppLoginToken) > 0 {
			if latestCurrentActiveAppLoginToken[0].TPushNotiToken != "" && latestCurrentActiveAppLoginToken[0].TOs != "" {

				fmt.Println("inside")
				langCode := defLangCode
				if latestCurrentActiveAppLoginToken[0].LanguageID != "" {
					langCode = latestCurrentActiveAppLoginToken[0].LanguageID
				}

				msgTitle := arrData.MsgTitle
				msg := arrData.Msg

				if translateTitle {
					msgTitle = helpers.TranslateV2(arrData.MsgTitle, langCode, arrData.MsgTitleValue)
				}
				if translateMsg {
					msg = helpers.TranslateV2(arrData.Msg, langCode, arrData.MsgValue)
				}

				arrCallSendPushNotificationIndApiData := CallSendPushNotificationIndApiStruct{
					Os:      latestCurrentActiveAppLoginToken[0].TOs,
					RegID:   latestCurrentActiveAppLoginToken[0].TPushNotiToken,
					Subject: msgTitle,
					Msg:     msg,
					CusMsg:  arrData.CustomData,
				}
				err = CallSendPushNotificationIndApi(arrCallSendPushNotificationIndApiData)
				if err != nil {
					LogErrorLog("ProcessPushNotificationDataV1-CallSendPushNotificationIndApi_member_id_is_given", err.Error(), arrCallSendPushNotificationIndApiData, true)
					return
				}

				if saveSysNotification {
					arrCrtSysNoti := models.AddSysNotificationStruct{
						ApiKeyID:     apiKeyID,
						Type:         "member",
						MemberID:     arrData.EntMemberID,
						Title:        string(encodedPNSub),
						Msg:          string(encodedPNMsg),
						LangCode:     langCode,
						CustMsg:      arrData.CustomData,
						BShow:        1,
						PNSendStatus: 1,
						Status:       "A",
						CreatedBy:    strconv.Itoa(arrData.EntMemberID),
					}
					_, err := models.AddSysNotification(arrCrtSysNoti)
					if err != nil {
						LogErrorLog("ProcessPushNotificationDataV1-CallSendPushNotificationIndApi_AddSysNotification", err.Error(), arrCrtSysNoti, true)
						return
					}
				}
			}
		}
	}

	return
}

func IsValidETHAddress(v string) bool {
	re := regexp.MustCompile("0x[A-Fa-f0-9]{40}")
	return re.MatchString(v)
}

func IsValidTRXAddress(v string) bool {
	re := regexp.MustCompile("^T[A-Za-z0-9]{33}")
	return re.MatchString(v)
}

//CallSendMailApi struct
type CallSendMailApiStruct struct {
	Subject  string
	Message  string
	Type     string
	FromMail string
	FromName string
	ToEmail  []string
	ToName   []string
	CCEmail  []string // optional
	CCName   []string // optional
	BccEmail []string // optional
}

func (s CallSendMailApiStruct) CallSendMailApi() error {

	if s.Subject == "" {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "send_mail_subject_invalid"}
	}

	if s.Message == "" {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "send_mail_message_invalid"}
	}

	// to email and to name checking
	if len(s.ToEmail) != len(s.ToName) {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "to_email_must_same_length_with_to_name"}
	}

	// cc email and cc name checking
	if len(s.CCEmail) != len(s.CCName) {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "cc_email_must_same_length_with_cc_name"}
	}

	fromName := ""
	if s.FromName != "" {
		fromName = s.FromName
	} else {
		fromName = "No Reply"
	}

	fromMail := "noreply@gta.com"
	if s.FromMail != "" {
		fromMail = s.FromMail
	}

	settingID := "send_mail_api_setting"
	arrApiSetting, _ := models.GetSysGeneralSetupByID(settingID)

	if arrApiSetting == nil {
		LogErrorLog("CallSendMailApi_GetSysGeneralSetupByID_failed", "send_mail_api_setting_is_missing", settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrApiSetting.InputType1 != "1" {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "email_service_is_not_available"}
	}

	url := arrApiSetting.InputValue1
	header := map[string]string{
		"x-api-key": arrApiSetting.SettingValue1,
	}
	jToEmail, _ := json.Marshal(s.ToEmail)
	jToName, _ := json.Marshal(s.ToName)
	arrPostData := map[string]string{
		"ProjectCode": arrApiSetting.InputType2,
		"FromEmail":   fromMail,
		"FromName":    fromName,
		"ToEmail":     string(jToEmail),
		"ToName":      string(jToName),
		"Subject":     s.Subject,
		"Msg":         s.Message,
		"MsgType":     s.Type,
		// "CCEmail":          s.Type,
		// "CCName":          s.Type,
		// "BCCEmail":          s.Type,
	}

	res, err := PostFormRequestApi(url, header, arrPostData, nil)
	// res, err := RequestAPI(arrApiSetting.InputValue2, url, header, arrPostData, nil)

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: ""}
	}

	if res.StatusCode != 200 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "error_in_email_service", Data: ""}
	}
	jsonToEmail := strings.Join(s.ToEmail, ",")
	jsonArrPostData, _ := json.Marshal(arrPostData)
	emailLog := models.EmailLog{
		Email:    string(jsonToEmail),
		Provider: "SmartblockENS",
		Data:     string(jsonArrPostData),
	}

	db := models.GetDB() // no need transaction because if failed no need rollback

	models.AddEmailLog(db, emailLog)

	return nil
}

func RequestBinanceAPI(method, url string, header map[string]string, body map[string]interface{}, resStruct interface{}) (*APIResponse, error) {

	var (
		req     *http.Request
		err     error
		reqBody []byte
	)

	switch method {
	case "GET":
		if body != nil {
			reqBody, err = json.Marshal(body)
			if err != nil {
				models.ErrorLog("RequestBNAPI_GET_failed_in_json_Marshal_body", err.Error(), body)
				return nil, err
			}
		} else {
			reqBody = nil
		}

		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))

		// data add
		// q := req.URL.Query()
		// if body != nil {
		// 	for k, d := range body {
		// 		bd, ok := d.(string)
		// 		if !ok {
		// 			models.ErrorLog("RequestAPI_GET_failed_in_add_query_data_string_assertion", err.Error(), d)
		// 			return nil, errors.New("invalid param " + k)
		// 		}
		// 		q.Add(k, bd)
		// 	}
		// }

		// // query encode
		// req.URL.RawQuery = q.Encode()

	case "POST":
		if body != nil {
			reqBody, err = json.Marshal(body)

			if err != nil {
				models.ErrorLog("RequestBNAPI_POST_failed_in_json_Marshal_body", err.Error(), body)
				return nil, err
			}
		} else {
			reqBody = nil
		}

		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))

	default:
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.REQUEST_API_INVALID_METHOD}
	}

	for k, h := range header {
		req.Header.Set(k, h)
	}

	if err != nil {
		models.ErrorLog("RequestBNAPI_failed_in_NewRequest", err.Error(), nil)
		return nil, err
	}

	// client := &http.Client{}
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	var headerDataString string
	if header != nil {
		headerDataJson, _ := json.Marshal(header)
		headerDataString = string(headerDataJson)
	}
	dataJson, _ := json.Marshal(body)
	dataString := string(dataJson)
	arrLogData := models.AddGeneralApiLogStruct{
		URLLink:  url,
		ApiType:  "RequestBNAPI_" + method,
		Method:   method,
		DataSent: headerDataString + dataString,
	}
	AddGeneralApiLogRst, _ := models.AddGeneralApiLog(arrLogData)

	resp, err := client.Do(req)

	if err != nil {
		models.ErrorLog("RequestBNAPI_failed_in_client_Do", err.Error(), req)
		return nil, err
	}

	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		models.ErrorLog("RequestBNAPI_failed_in_ioutil_ReadAll", err.Error(), resp.Body)
		return nil, err
	}

	if string(resBody) != "" && resStruct != nil {
		err = json.Unmarshal(resBody, resStruct)
		if err != nil {
			models.ErrorLog("RequestBNAPI_failed_in_json_Unmarshal_resBody", err.Error(), resBody)
			return nil, err
		}
	}

	response := &APIResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       string(resBody),
	}

	resJson, _ := json.Marshal(response)
	resString := string(resJson)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " general_api_log.id = ? ", CondValue: AddGeneralApiLogRst.ID},
	)
	updateColumn := map[string]interface{}{
		"data_received": resString,
	}
	_ = models.UpdatesFn("general_api_log", arrCond, updateColumn, false)

	return response, nil
}
