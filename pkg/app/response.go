package app

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/translation"
)

type Gin struct {
	C *gin.Context
}

type Response struct {
	Rst  int         `json:"rst"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data"`
}

type ArrDataResponseList struct {
	CurrentPage           int         `json:"current_page"`
	PerPage               int         `json:"per_page"`
	TotalCurrentPageItems int         `json:"total_current_page_items"`
	TotalPage             int         `json:"total_page"`
	TotalPageItems        int         `json:"total_page_items"`
	CurrentPageItems      interface{} `json:"current_page_items"`
	TableHeaderList       interface{} `json:"table_header_list"`
	TableSummaryData      interface{} `json:"table_summary_data"`
}

type ArrDataResponseDefaultList struct {
	CurrentPage           int         `json:"current_page"`
	PerPage               int         `json:"per_page"`
	TotalCurrentPageItems int         `json:"total_current_page_items"`
	TotalPage             int         `json:"total_page"`
	TotalPageItems        int         `json:"total_page_items"`
	CurrentPageItems      interface{} `json:"current_page_items"`
}

type ApiResponse struct { // status code string
	StatusCode int                    `json:"statusCode"`
	Status     string                 `json:"status"`
	Message    []string               `json:"message"`
	Msg        string                 `json:"msg"`
	Data       map[string]interface{} `json:"data"`
}

type ApiArrayResponse struct { // status code string
	StatusCode int                      `json:"statusCode"`
	Status     string                   `json:"status"`
	Message    []string                 `json:"message"`
	Msg        string                   `json:"msg"`
	Data       []map[string]interface{} `json:"data"`
}

// type ResponseList struct {
// 	Status       string      `json:"status"`
// 	StatusCode   int         `json:"statusCode"`
// 	Page         int         `json:"page"`
// 	TotalPages   int64       `json:"total_pages"`
// 	TotalRecords int64       `json:"total_records"`
// 	EndFlag      int64       `json:"end_flag"`
// 	Data         interface{} `json:"data"`
// }
// type ResponseList2 struct {
// 	Status          string      `json:"status"`
// 	StatusCode      int         `json:"statusCode"`
// 	Page            int         `json:"page"`
// 	TotalPages      int64       `json:"total_pages"`
// 	TotalRecords    int64       `json:"total_records"`
// 	EndFlag         int64       `json:"end_flag"`
// 	TotalAmount     float64     `json:"total_amount"`
// 	TotalPageAmount float64     `json:"total_page_amount"`
// 	Data            interface{} `json:"data"`
// }

// type JsonResponse struct {
// 	ResponseCode    int         `json:"-"`
// 	ResponseMsg     string      `json:"-"`
// 	Page            int         `json:"page"`
// 	TotalAmount     float64     `json:"total_amount"`
// 	TotalPageAmount float64     `json:"total_page_amount"`
// 	TotalPages      int64       `json:"total_pages"`
// 	TotalRecords    int64       `json:"total_records"`
// 	EndFlag         int64       `json:"end_flag"`
// 	Data            interface{} `json:"data"`
// }

// Response setting gin.JSON
func (g *Gin) Response(errCode int, httpCode int, message string, data interface{}) {

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if g.C.PostForm("lang_code") != "" {
		langCode = g.C.PostForm("lang_code")
	} else if g.C.GetHeader("Accept-Language") != "" {
		langCode = g.C.GetHeader("Accept-Language")
	}

	// check language
	ok := models.ExistLangague(langCode)
	if !ok {
		langCode = "zh"
	}

	message = strings.Replace(strings.ToLower(message), " ", "_", -1)
	translatedWord := helpers.TranslateV2(message, langCode, make(map[string]string))

	g.C.JSON(httpCode, Response{
		Rst:  errCode,
		Msg:  translatedWord,
		Data: data,
	})
	return
}

type MsgStruct struct {
	Msg      string
	LangCode string
	Params   map[string]string
}

// Response setting gin.JSON
func (g *Gin) ResponseV2(errCode int, httpCode int, message MsgStruct, data interface{}) {

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if g.C.PostForm("lang_code") != "" {
		langCode = g.C.PostForm("lang_code")
	} else if g.C.GetHeader("Accept-Language") != "" {
		langCode = g.C.GetHeader("Accept-Language")
	}

	// check language
	ok := models.ExistLangague(langCode)
	if !ok {
		langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	}

	translatedWord := helpers.TranslateV2(message.Msg, langCode, message.Params)

	g.C.JSON(httpCode, Response{
		Rst:  errCode,
		Msg:  translatedWord,
		Data: data,
	})
	return
}

// ResponseError response with custom error
func (g *Gin) ResponseError(err error) {

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if g.C.PostForm("lang_code") != "" {
		langCode = g.C.PostForm("lang_code")
	} else if g.C.GetHeader("Accept-Language") != "" {
		langCode = g.C.GetHeader("Accept-Language")
	}

	// check language
	ok := models.ExistLangague(langCode)
	if !ok {
		langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	}

	switch err.(type) {

	case *e.CustomError:
		c := err.(*e.CustomError)
		var (
			httpCode int
			code     int
		)

		httpCode = http.StatusOK

		if code = c.Code; code == 0 {
			code = e.UNPROCESSABLE_ENTITY
		}

		g.C.JSON(httpCode, Response{
			Rst: 0,
			// Msg:  g.Trans(c.Error(), c.TemplateData),
			Msg:  helpers.TranslateV2(c.Error(), langCode, make(map[string]string)),
			Data: nil,
		})

	default:
		g.C.JSON(http.StatusOK, Response{
			Rst:  0,
			Msg:  helpers.TranslateV2(err.Error(), langCode, make(map[string]string)),
			Data: nil,
		})
	}
	return
}

// JSONResponse setting gin.JSON
// func (g *Gin) JSONResponse(httpCode int, data interface{}) {
// 	g.C.JSON(httpCode, data)
// 	return
// }

// func (g *Gin) ResponseList(httpCode, errCode int, data interface{}, page int, totalPages int64, totalRecords int64, endFlag int64, response interface{}) Response {
// 	var status string
// 	if status = "success"; httpCode != 200 {
// 		status = "error"
// 	}

// 	//msg := [1]string{e.GetMsg(errCode)}

// 	return_json, _ := g.C.Get("return_json")
// 	//Marshal will return []byte, []byte is marshalled as base64 encoded string
// 	res2B, _ := json.Marshal(response)

// 	if return_json == 1 {
// 		//Cast res2B to string will display the response in string form
// 		return Response{Status: status, StatusCode: httpCode, Data: string(res2B)}
// 	}

// 	g.C.JSON(httpCode, ResponseList{
// 		Status:       status,
// 		StatusCode:   httpCode,
// 		Page:         page,
// 		TotalPages:   totalPages,
// 		TotalRecords: totalRecords,
// 		EndFlag:      endFlag,
// 		Data:         data,
// 	})

// 	return Response{}
// }
// func (g *Gin) ResponseList2(httpCode, errCode int, data interface{}, page int, totalPages int64, totalRecords int64, endFlag int64, totalAmount float64, totalPageAmount float64, response interface{}) Response {
// 	var status string
// 	if status = "success"; httpCode != 200 {
// 		status = "error"
// 	}

// 	//msg := [1]string{e.GetMsg(errCode)}

// 	return_json, _ := g.C.Get("return_json")
// 	//Marshal will return []byte, []byte is marshalled as base64 encoded string
// 	res2B, _ := json.Marshal(response)

// 	if return_json == 1 {
// 		//Cast res2B to string will display the response in string form
// 		return Response{Status: status, StatusCode: httpCode, Data: string(res2B)}
// 	}

// 	g.C.JSON(httpCode, ResponseList2{
// 		Status:          status,
// 		StatusCode:      httpCode,
// 		Page:            page,
// 		TotalPages:      totalPages,
// 		TotalRecords:    totalRecords,
// 		EndFlag:         endFlag,
// 		Data:            data,
// 		TotalPageAmount: totalPageAmount,
// 		TotalAmount:     totalAmount,
// 	})

// 	return Response{}
// }

// ResponseEtag response will pass and check data at etag middleware
// func (g *Gin) ResponseEtag(status string, httpCode, errCode int, data interface{}) {
// 	g.C.Set("response_data", Response{
// 		Status:     status,
// 		StatusCode: errCode,
// 		Msg:        g.Trans(e.GetMsg(errCode), nil),
// 		Data:       data,
// 	})
// 	return
// }

// Trans func
func (g *Gin) Trans(text string, template map[string]interface{}) string {
	l, ok := g.C.Get("localizer")

	if !ok {
		return text
	}

	localizer, ok := l.(*translation.Localizer)

	if !ok {
		return text
	}

	return localizer.Trans(text, template)
}

// RemarkTrans func
func (g *Gin) RemarkTrans(text string) string {
	l, ok := g.C.Get("localizer")

	if !ok {
		return text
	}

	localizer, ok := l.(*translation.Localizer)

	if !ok {
		return text
	}

	regex := regexp.MustCompile(`\#\*.*?\*\#`)

	loop := true
	var trans string
	trans = text
	for loop {
		// regular expression pattern
		t := regex.FindString(trans)

		if t == "" {
			loop = false
		} else {
			extract := t

			t = strings.Trim(t, "#*")
			t = strings.Trim(t, "*#")

			res := localizer.Trans(t, nil)

			trans = strings.Replace(trans, extract, res, -1)
		}
	}

	return trans
}

// AbortHTTPCode func
func (g *Gin) AbortHTTPCode(httpCode int) {
	g.C.JSON(httpCode, nil)
	return
}

type WSResponse struct {
	Rst  int         `json:"rst"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data"`
}
