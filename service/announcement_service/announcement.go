package announcement_service

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
)

type MemberAnnouncementListStruct struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Path  string `json:"path"`
	// BackgroundImgUrl string `json:"backgound_img_url"`
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name"`
	CreatedAt    string `json:"created_at"`
	PopUp        uint   `json:"pop_up"`
}

type MemberAnnouncementPaginateStruct struct {
	MemberID     int
	LangCode     string
	PopUp        string
	CategoryCode string
	Page         int64
	Type         string
}

// func GetMemberAnnouncementPaginateListv1
func GetMemberAnnouncementPaginateListv1(arrData MemberAnnouncementPaginateStruct) interface{} {

	// arrNewMemberAnnouncementList := make([]MemberAnnouncementListStruct, 0)
	arrDataReturn := app.ArrDataResponseList{
		CurrentPageItems: make([]MemberAnnouncementListStruct, 0),
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_company_announcement.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ent_company_announcement.status = ? ", CondValue: "A"},
	)
	if arrData.CategoryCode != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " dl_list_category.code = ? ", CondValue: arrData.CategoryCode},
		)
	}
	if arrData.PopUp != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_company_announcement.popup = ? ", CondValue: arrData.PopUp},
		)
	}
	sysLangCode, _ := models.GetLanguage(arrData.LangCode)

	if sysLangCode == nil || sysLangCode.ID == "" {
		return arrDataReturn
	}

	var sysLangID int
	if sysLangCode.ID != "" {
		sysLangIDInt, err := strconv.Atoi(sysLangCode.ID)
		if err != nil {
			base.LogErrorLog("GetMemberAnnouncementListv1_invalid_language_code", err.Error(), "invalid_langcode_is_received", true)
			return arrDataReturn
		}
		sysLangID = sysLangIDInt
	}

	arrEntUserLeaderCond := make([]models.WhereCondFn, 0)
	arrEntUserLeaderCond = append(arrEntUserLeaderCond,
		models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.MemberID},
	)
	entLeaderList, _ := models.GetEntUserLeaderFn(arrEntUserLeaderCond, false)
	leaderNickName := ""
	if len(entLeaderList) > 0 {
		leaderNickName = entLeaderList[0].LeaderName
	}
	arrExtraEntCompAnnData := models.ExtraEntCompAnnDataStruct{
		LangID: sysLangID,
		Type:   arrData.Type,
	}

	arrMemberAnnouncementList, _ := models.GetEntCompanyAnnouncementFn(arrExtraEntCompAnnData, arrCond, false)
	arrNewMemberAnnouncementList := make([]interface{}, 0)
	if len(arrMemberAnnouncementList) > 0 {
		for _, arrMemberAnnouncementListV := range arrMemberAnnouncementList {
			if arrMemberAnnouncementListV.OpenMemberGroup != "" {
				arrUsername := strings.Split(arrMemberAnnouncementListV.OpenMemberGroup, ",")
				strUsername := ""
				for _, arrUsernameV := range arrUsername {
					if strUsername != "" {
						strUsername = strUsername + ",'" + arrUsernameV + "'"
					} else {
						strUsername = "'" + arrUsernameV + "'"
					}
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ent_member.nick_name IN (" + strUsername + ") "},
				)

				arrEntMemberList, _ := models.GetEntMemberListFn(arrCond, false)

				if len(arrEntMemberList) > 0 {
					var arrTargetID []int
					for _, arrEntMemberListV := range arrEntMemberList {
						arrTargetID = append(arrTargetID, arrEntMemberListV.ID)
					}

					arrNearestUpline, err := models.GetNearestUplineByMemId(arrData.MemberID, arrTargetID, "", false)

					if err != nil {
						arrErr := map[string]interface{}{
							"MemberID":    arrData.MemberID,
							"arrTargetID": arrTargetID,
						}
						base.LogErrorLog("GetMemberAnnouncementLandv1-GetNearestUplineByMemId_OpenMemberGroup_failed", err.Error(), arrErr, true)
					}
					if arrNearestUpline == nil {
						continue
					}
				}
			}

			if arrMemberAnnouncementListV.TUsername != "" {
				arrUsername := strings.Split(arrMemberAnnouncementListV.TUsername, ",")
				strUsername := ""
				for _, arrUsernameV := range arrUsername {
					if strUsername != "" {
						strUsername = strUsername + ",'" + arrUsernameV + "'"
					} else {
						strUsername = "'" + arrUsernameV + "'"
					}
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ent_member.nick_name IN (" + strUsername + ") "},
				)

				arrEntMemberList, _ := models.GetEntMemberListFn(arrCond, false)

				if len(arrEntMemberList) > 0 {
					var arrTargetID []int
					for _, arrEntMemberListV := range arrEntMemberList {
						arrTargetID = append(arrTargetID, arrEntMemberListV.ID)
					}

					arrNearestUpline, err := models.GetNearestUplineByMemId(arrData.MemberID, arrTargetID, "", false)

					if err != nil {
						arrErr := map[string]interface{}{
							"MemberID":    arrData.MemberID,
							"arrTargetID": arrTargetID,
						}
						base.LogErrorLog("GetMemberAnnouncementLandv1-GetNearestUplineByMemId_failed", err.Error(), arrErr, true)
					}

					if arrNearestUpline != nil {
						continue
					}
				}
			}

			if arrMemberAnnouncementListV.TLeader != "" && leaderNickName != "" {
				arrLeaderUsername := strings.Split(arrMemberAnnouncementListV.TLeader, ",")
				stringInSliceRst := helpers.StringInSlice(leaderNickName, arrLeaderUsername)
				if stringInSliceRst {
					continue
				}
			}
			params := make(map[string]string)
			// adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()
			path := ""
			// if arrMemberAnnouncementListV.AnnouncementCategoryCode != "picture" {
			// path = adminServerDomain + arrMemberAnnouncementListV.FolderPath
			path = arrMemberAnnouncementListV.FolderPath
			// }
			createdAtString := arrMemberAnnouncementListV.CreatedAt.Format("2006-01-02 15:04:05")
			announcementCategoryNameTW := helpers.TranslateV2(arrMemberAnnouncementListV.AnnouncementCategoryName, arrData.LangCode, params)
			// backgroundImgURL := ""
			// if arrMemberAnnouncementListV.AnnouncementCategoryCode == "picture" {
			// 	backgroundImgURL = path
			// }
			arrNewMemberAnnouncementList = append(arrNewMemberAnnouncementList,
				MemberAnnouncementListStruct{
					ID:    arrMemberAnnouncementListV.ID,
					Title: arrMemberAnnouncementListV.Title,
					Desc:  arrMemberAnnouncementListV.Desc,
					Path:  path,
					// BackgroundImgUrl: backgroundImgURL,
					CategoryCode: arrMemberAnnouncementListV.AnnouncementCategoryCode,
					CategoryName: announcementCategoryNameTW,
					CreatedAt:    createdAtString,
					PopUp:        arrMemberAnnouncementListV.PopUp,
				},
			)
		}
	}

	type arrListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	var arrTableHeaderList []arrListSettingListStruct
	arrStatementListSetting, _ := models.GetSysGeneralSetupByID("announcement_list_api_setting")
	if arrStatementListSetting != nil {
		var arrStatementListSettingList map[string][]arrListSettingListStruct
		json.Unmarshal([]byte(arrStatementListSetting.InputType1), &arrStatementListSettingList)
		arrTableHeaderList = arrStatementListSettingList["table_header_list"]
		for k, v1 := range arrStatementListSettingList["table_header_list"] {
			v1.Name = helpers.Translate(v1.Name, arrData.LangCode)
			arrTableHeaderList[k] = v1
		}
	}

	page := base.Pagination{
		Page:      arrData.Page,
		DataArr:   arrNewMemberAnnouncementList,
		HeaderArr: arrTableHeaderList,
	}

	return page.PaginationInterfaceV1()
}

type MemberAnnouncementLandStruct struct {
	MemberID     int
	LangCode     string
	PopUp        string
	CategoryCode string
	Type         string
}

// func GetMemberAnnouncementLandv1
func GetMemberAnnouncementLandv1(arrData MemberAnnouncementLandStruct) []MemberAnnouncementListStruct {

	// arrNewMemberAnnouncementList := make([]MemberAnnouncementListStruct, 0)
	arrDataReturn := make([]MemberAnnouncementListStruct, 0)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_company_announcement.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ent_company_announcement.status = ? ", CondValue: "A"},
	)
	if arrData.CategoryCode != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " dl_list_category.code = ? ", CondValue: arrData.CategoryCode},
		)
	}
	if arrData.PopUp != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_company_announcement.popup = ? ", CondValue: arrData.PopUp},
		)
	}
	sysLangCode, _ := models.GetLanguage(arrData.LangCode)

	if sysLangCode == nil || sysLangCode.ID == "" {
		return arrDataReturn
	}

	var sysLangID int
	if sysLangCode.ID != "" {
		sysLangIDInt, err := strconv.Atoi(sysLangCode.ID)
		if err != nil {
			base.LogErrorLog("GetMemberAnnouncementListv1_invalid_language_code", err.Error(), "invalid_langcode_is_received", true)
			return arrDataReturn
		}
		sysLangID = sysLangIDInt
	}
	arrEntUserLeaderCond := make([]models.WhereCondFn, 0)
	arrEntUserLeaderCond = append(arrEntUserLeaderCond,
		models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.MemberID},
	)
	entLeaderList, _ := models.GetEntUserLeaderFn(arrEntUserLeaderCond, false)
	leaderNickName := ""
	if len(entLeaderList) > 0 {
		leaderNickName = entLeaderList[0].LeaderName
	}
	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")
	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue2, 10, 64)

	arrExtraEntCompAnnData := models.ExtraEntCompAnnDataStruct{
		LangID: sysLangID,
		Type:   arrData.Type,
	}
	arrMemberAnnouncementList, _ := models.GetEntCompanyAnnouncementFn(arrExtraEntCompAnnData, arrCond, false)
	arrNewMemberAnnouncementList := make([]MemberAnnouncementListStruct, 0)
	if len(arrMemberAnnouncementList) > 0 {
		for _, arrMemberAnnouncementListV := range arrMemberAnnouncementList {
			if arrMemberAnnouncementListV.OpenMemberGroup != "" {
				arrUsername := strings.Split(arrMemberAnnouncementListV.OpenMemberGroup, ",")
				strUsername := ""
				for _, arrUsernameV := range arrUsername {
					if strUsername != "" {
						strUsername = strUsername + ",'" + arrUsernameV + "'"
					} else {
						strUsername = "'" + arrUsernameV + "'"
					}
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ent_member.nick_name IN (" + strUsername + ") "},
				)

				arrEntMemberList, _ := models.GetEntMemberListFn(arrCond, false)

				if len(arrEntMemberList) > 0 {
					var arrTargetID []int
					for _, arrEntMemberListV := range arrEntMemberList {
						arrTargetID = append(arrTargetID, arrEntMemberListV.ID)
					}

					arrNearestUpline, err := models.GetNearestUplineByMemId(arrData.MemberID, arrTargetID, "", false)
					if err != nil {
						arrErr := map[string]interface{}{
							"MemberID":    arrData.MemberID,
							"arrTargetID": arrTargetID,
						}
						base.LogErrorLog("GetMemberAnnouncementLandv1-GetNearestUplineByMemId_OpenMemberGroup_failed", err.Error(), arrErr, true)
					}

					if arrNearestUpline == nil {
						continue
					}
				}
			}

			if arrMemberAnnouncementListV.TUsername != "" {
				arrUsername := strings.Split(arrMemberAnnouncementListV.TUsername, ",")
				strUsername := ""
				for _, arrUsernameV := range arrUsername {
					if strUsername != "" {
						strUsername = strUsername + ",'" + arrUsernameV + "'"
					} else {
						strUsername = "'" + arrUsernameV + "'"
					}
				}

				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ent_member.nick_name IN (" + strUsername + ") "},
				)

				arrEntMemberList, _ := models.GetEntMemberListFn(arrCond, false)

				if len(arrEntMemberList) > 0 {
					var arrTargetID []int
					for _, arrEntMemberListV := range arrEntMemberList {
						arrTargetID = append(arrTargetID, arrEntMemberListV.ID)
					}

					arrNearestUpline, err := models.GetNearestUplineByMemId(arrData.MemberID, arrTargetID, "", false)

					if err != nil {
						arrErr := map[string]interface{}{
							"MemberID":    arrData.MemberID,
							"arrTargetID": arrTargetID,
						}
						base.LogErrorLog("GetMemberAnnouncementLandv1-GetNearestUplineByMemId_failed", err.Error(), arrErr, true)
					}
					if arrNearestUpline != nil {
						continue
					}
				}
			}

			if arrMemberAnnouncementListV.TLeader != "" && leaderNickName != "" {
				arrLeaderUsername := strings.Split(arrMemberAnnouncementListV.TLeader, ",")
				stringInSliceRst := helpers.StringInSlice(leaderNickName, arrLeaderUsername)
				if stringInSliceRst {
					continue
				}
			}

			params := make(map[string]string)
			// adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()
			// path := adminServerDomain + arrMemberAnnouncementListV.FolderPath
			path := arrMemberAnnouncementListV.FolderPath
			createdAtString := arrMemberAnnouncementListV.CreatedAt.Format("2006-01-02 15:04:05")
			announcementCategoryNameTW := helpers.TranslateV2(arrMemberAnnouncementListV.AnnouncementCategoryName, arrData.LangCode, params)
			// backgroundImgURL := ""
			// if arrMemberAnnouncementListV.AnnouncementCategoryCode == "picture" {
			// 	backgroundImgURL = path
			// }
			arrNewMemberAnnouncementList = append(arrNewMemberAnnouncementList,
				MemberAnnouncementListStruct{
					ID:    arrMemberAnnouncementListV.ID,
					Title: arrMemberAnnouncementListV.Title,
					Desc:  arrMemberAnnouncementListV.Desc,
					Path:  path,
					// BackgroundImgUrl: backgroundImgURL,
					CategoryCode: arrMemberAnnouncementListV.AnnouncementCategoryCode,
					CategoryName: announcementCategoryNameTW,
					CreatedAt:    createdAtString,
					PopUp:        arrMemberAnnouncementListV.PopUp,
				},
			)

			if int(limit) == len(arrNewMemberAnnouncementList) {
				break
			}
		}
	}

	return arrNewMemberAnnouncementList
}
