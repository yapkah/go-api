package language_service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
)

type AppTranslationStruct struct {
	LangCode string
	EtagID   string
}

type AppTranslationRstStruct struct {
	EtagID         string                            `json:"etag_id"`
	TranslatedList []*models.AppTranslationsFrontend `json:"translated_list"`
}

func GetAppTranslation(arrData AppTranslationStruct) (AppTranslationRstStruct, error) {

	var (
		returnTranslatedWord, crtTranslationTag, updTranslationTag bool
		etag, lastUpdateAt                                         string
	)

	langCode := arrData.LangCode
	etagID := arrData.EtagID
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " translation_etag.locale = ? ", CondValue: langCode},
	)
	arrTranslationEtag, _ := models.GetTranslationEtagFn(arrCond, false)

	arrReturnTranslated := make([]*models.AppTranslationsFrontend, 0)
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " translations_frontend.locale = ? ", CondValue: langCode},
	)
	arrTranslated, _ := models.GetAppFrontendTranslationFn(arrCond, false)

	if len(arrTranslationEtag) > 0 {
		if etagID != "" {
			if arrTranslationEtag[0].EtagID != etagID { // etag from front-end tally with back-end.
				etag = arrTranslationEtag[0].EtagID
				returnTranslatedWord = true
			} else { // etag from front-end tally with back-end.

				// start check with total record
				if arrTranslationEtag[0].TotalRecord != len(arrTranslated) {
					updTranslationTag = true
					returnTranslatedWord = true
				}
				// end check with total record
				// start check with last update time
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " translations_frontend.locale = ? ", CondValue: langCode},
				)
				arrLastUpdatedTranslation, _ := models.GetLastUpdateAppFrontendTranslationFn(arrCond, false)
				// fmt.Println("arrLastUpdatedTranslation.Updated_at:", arrLastUpdatedTranslation.Updated_at)
				if arrLastUpdatedTranslation.Updated_at.Year() != 0001 {
					etagLastUpdatedAtString := arrTranslationEtag[0].LastUpdatedAt.Format("2006-01-02 15:04:05")
					lastUpdatedTranslationString := arrLastUpdatedTranslation.Updated_at.Format("2006-01-02 15:04:05")
					// fmt.Println("etagLastUpdatedAtString:", etagLastUpdatedAtString)
					// fmt.Println("lastUpdatedTranslationString:", lastUpdatedTranslationString)
					// fmt.Println("arrTranslationEtag[0].LastUpdatedAt:", arrTranslationEtag[0].LastUpdatedAt)
					if lastUpdatedTranslationString > etagLastUpdatedAtString || arrTranslationEtag[0].LastUpdatedAt.Year() == 0001 {
						updTranslationTag = true
						returnTranslatedWord = true
						lastUpdateAt = arrLastUpdatedTranslation.Updated_at.Format("2006-01-02 15:04:05")
					}
				}
				// end check with last update time
			}
		} else {
			etag = arrTranslationEtag[0].EtagID
			returnTranslatedWord = true
		}
	} else {
		crtTranslationTag = true
		returnTranslatedWord = true
	}

	if returnTranslatedWord {
		arrReturnTranslated = arrTranslated
	}
	if crtTranslationTag {
		etag = strconv.Itoa(len(arrTranslated)) + "_" + base.GetCurrentTime("20060102150405")
		arrCrtData := models.TranslationEtag{
			EtagID:      etag,
			Locale:      langCode,
			TotalRecord: len(arrTranslated),
		}
		_, _ = models.AddTranslationEtag(arrCrtData)
	}
	if updTranslationTag {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " translation_etag.locale = ? ", CondValue: langCode},
		)
		etag = strconv.Itoa(len(arrTranslated)) + "_" + base.GetCurrentTime("20060102150405")
		updateColumn := map[string]interface{}{
			"etag_id":      etag,
			"total_record": len(arrTranslated),
		}
		if lastUpdateAt != "" {
			updateColumn["last_updated_at"] = lastUpdateAt
		}
		_ = models.UpdatesFn("translation_etag", arrCond, updateColumn, false)
	}

	arrDataReturn := AppTranslationRstStruct{
		EtagID:         etag,
		TranslatedList: arrReturnTranslated,
	}
	return arrDataReturn, nil
}

type ProcessUpdateMemberDeviceLanguagev1Struct struct {
	AccessToken string
	LangCode    string
	Platform    string
	SourceID    int
	PrjID       int
}

func ProcessUpdateMemberDeviceLanguagev1(arrData ProcessUpdateMemberDeviceLanguagev1Struct) (string, string, error) {
	var os, pushNotiToken string

	arrCond := make([]models.WhereCondFn, 0)
	if strings.ToLower(arrData.Platform) == "app" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "b_login = ?", CondValue: 1},
			models.WhereCondFn{Condition: "b_logout = ?", CondValue: 0},
			models.WhereCondFn{Condition: "t_token = ?", CondValue: arrData.AccessToken},
			models.WhereCondFn{Condition: "source = ?", CondValue: arrData.SourceID},
		)
		loginLogRst, _ := models.GetAppLoginLogFn(arrCond, "", false)
		if loginLogRst != nil {
			// start update current device lang code
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: "id = ?", CondValue: loginLogRst.ID},
			)

			updateColumn := map[string]interface{}{"language_id": arrData.LangCode}
			err := models.UpdatesFn("app_login_log", arrUpdCond, updateColumn, false)
			if err != nil {
				arrErr := map[string]interface{}{
					"upd_cond":   arrUpdCond,
					"upd_column": updateColumn,
				}
				base.LogErrorLog("ProcessUpdateMemberDeviceLanguagev1-update_app_login_log_language_id_failed", err.Error(), arrErr, true)
				return "", "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrErr}
			}
			// end update current device lang code

			if loginLogRst.TOs != "" {
				os = loginLogRst.TOs
			}
			if loginLogRst.TPushNotiToken != "" {
				pushNotiToken = loginLogRst.TPushNotiToken
			}
		}
	} else if strings.ToLower(arrData.Platform) == "htmlfive" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "b_login = ?", CondValue: 1},
			models.WhereCondFn{Condition: "b_logout = ?", CondValue: 0},
			models.WhereCondFn{Condition: "t_token = ?", CondValue: arrData.AccessToken},
			models.WhereCondFn{Condition: "source = ?", CondValue: arrData.SourceID},
		)
		loginLogRst, _ := models.GetHtmlfiveLoginLogFn(arrCond, "", false)
		if loginLogRst != nil {
			// start update current device lang code
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: "id = ?", CondValue: loginLogRst.ID},
			)

			updateColumn := map[string]interface{}{"language_id": arrData.LangCode}
			err := models.UpdatesFn("htmlfive_login_log", arrUpdCond, updateColumn, false)
			if err != nil {
				arrErr := map[string]interface{}{
					"upd_cond":   arrUpdCond,
					"upd_column": updateColumn,
				}
				base.LogErrorLog("ProcessUpdateMemberDeviceLanguagev1-update_htmlfive_login_log_language_id_failed", err.Error(), arrErr, true)
				return "", "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: arrErr}
			}
			// end update current device lang code

			if loginLogRst.TOs != "" {
				os = loginLogRst.TOs
			}
			if loginLogRst.TPushNotiToken != "" {
				pushNotiToken = loginLogRst.TPushNotiToken
			}
		}
	}

	return os, pushNotiToken, nil
}
