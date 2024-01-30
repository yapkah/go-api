package download_service

import (
	"strconv"
	"strings"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
)

type GetMemberDownloadListStruct struct {
	MemberID     int    `json:"member_id"`
	LangCode     string `json:"lang_code"`
	CategoryCode string `json:"category_code"`
	Type         string `json:"type"`
	Path         string `json:"path"`
}

// func GetMemberDownloadListv1
func (arrData GetMemberDownloadListStruct) GetMemberDownloadListv1() interface{} {

	var (
		arrDataReturn []interface{}
		sysLangID     int
	)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_company_downloads.b_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ent_company_downloads.status = ? ", CondValue: "A"},
	)
	if arrData.CategoryCode != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " dl_list_category.code = ? ", CondValue: strings.ToLower(arrData.CategoryCode)},
		)
	}
	sysLangCode, _ := models.GetLanguage(arrData.LangCode)

	if sysLangCode == nil || sysLangCode.ID == "" {
		return arrDataReturn
	}
	if sysLangCode.ID != "" {
		sysLangIDInt, err := strconv.Atoi(sysLangCode.ID)
		if err != nil {
			base.LogErrorLog("GetMemberDownloadListv1-invalid_language_code", err.Error(), "invalid_langcode_is_received", true)
			return arrDataReturn
		}
		sysLangID = sysLangIDInt
	}

	arrExtraEntCompDwnlData := models.ExtraEntCompDwnlDataStruct{
		LangID: sysLangID,
		Type:   arrData.Type,
	}

	arrMemberDownloadList, _ := models.GetEntCompanyDownloadsFn(arrExtraEntCompDwnlData, arrCond, false)

	if len(arrMemberDownloadList) > 0 {
		for _, v := range arrMemberDownloadList {

			arrDataReturn = append(arrDataReturn, map[string]interface{}{
				"title":         v.Title,
				"description":   v.Desc,
				"category_code": v.CategoryCode,
				"category_name": helpers.Translate(v.CategoryName, arrData.LangCode),
				"path":          v.Path,
				"created_at":    v.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return arrDataReturn
}
