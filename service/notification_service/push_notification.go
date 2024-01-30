package notification_service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/setting"
)

// ProcessSendIndPushNotificationForm struct
type ProcessSendIndPushNotificationForm struct {
	CryptoAddress string         `form:"crypto_address" json:"crypto_address" valid:"Required;"`
	Subject       string         `form:"subject" json:"subject" valid:"Required;"`
	SubjectParams []ParamsStruct `form:"subject_params" json:"subject_params"`
	Msg           string         `form:"msg" json:"msg" valid:"Required;"`
	MsgParams     []ParamsStruct `form:"msg_params" json:"msg_params"`
	CustMsg       string         `form:"cust_msg" json:"cust_msg"`
	LangCode      string         `form:"lang_code" json:"lang_code"`
}
type ParamsStruct struct {
	Key       string `form:"key" json:"key"`
	Value     string `form:"value" json:"value"`
	Translate bool   `form:"translate" json:"translate"`
}

type ProcessSendPushNotificationIndFromApiReqStruct struct {
	CryptoAddress string
	Subject       string
	SubjectParams []ParamsStruct
	Msg           string
	MsgParams     []ParamsStruct
	CustMsg       string
	LangCode      string
	Source        uint8
	PrjID         uint8
}

func ProcessSendPushNotificationIndFromApiReq(tx *gorm.DB, arrData ProcessSendPushNotificationIndFromApiReqStruct) error {
	langCode := arrData.LangCode

	subjectParams := make(map[string]string, 0)
	if len(arrData.SubjectParams) > 0 {
		for _, subjectParamsV := range arrData.SubjectParams {
			value := subjectParamsV.Value
			if subjectParamsV.Translate {
				value = helpers.TranslateV2(subjectParamsV.Value, langCode, nil)
			}
			subjectParams[subjectParamsV.Key] = value
		}
	}
	msgParams := make(map[string]string, 0)
	if len(arrData.MsgParams) > 0 {
		for _, msgParamsV := range arrData.MsgParams {
			value := msgParamsV.Value
			if msgParamsV.Translate {
				value = helpers.TranslateV2(msgParamsV.Value, langCode, nil)
			}
			msgParams[msgParamsV.Key] = value
		}
	}

	translatedSubject := helpers.TranslateV2(arrData.Subject, langCode, subjectParams)
	translatedMsg := helpers.TranslateV2(arrData.Msg, langCode, msgParams)

	arrEntMemCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemCryptoFn = append(arrEntMemCryptoFn,
		models.WhereCondFn{Condition: " crypto_address = ?", CondValue: arrData.CryptoAddress},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEntMemCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemCryptoFn, false)

	if arrEntMemCrypto == nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_crypto_address"}
	}

	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: arrEntMemCrypto.MemberID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrEntMember, _ := models.GetEntMemberFn(arrEntMemberFn, "", false)
	if arrEntMember == nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_crypto_address"}
	}

	pnSub := base.PushNotificationContentStruct{
		Msg:    arrData.Subject,
		Params: subjectParams,
	}
	encodedPNSub, _ := json.Marshal(pnSub)

	pnMsg := base.PushNotificationContentStruct{
		Msg:    arrData.Msg,
		Params: msgParams,
	}
	encodedPNMsg, _ := json.Marshal(pnMsg)

	arrCrtSysNoti := models.AddSysNotificationStruct{
		ApiKeyID: int(arrData.PrjID),
		Type:     "member",
		// PNType      string    `json:"pn_type" gorm:"column:pn_type"`
		MemberID:     arrEntMember.ID,
		Title:        string(encodedPNSub),
		Msg:          string(encodedPNMsg),
		LangCode:     langCode,
		CustMsg:      arrData.CustMsg,
		BShow:        1,
		PNSendStatus: 0,
		Status:       "A",
		CreatedBy:    strconv.Itoa(arrEntMember.ID),
	}
	sysNoti, _ := models.AddSysNotification(arrCrtSysNoti)

	existingActiveAppLoginLog, _ := models.GetExistingActiveAppLoginLog(arrEntMember.MainID, arrData.Source, false)
	// begin transaction
	if len(existingActiveAppLoginLog) < 1 {
		return nil
	}

	var os string
	var pushNotiToken string
	for _, existingActiveAppLoginLogV := range existingActiveAppLoginLog {
		if existingActiveAppLoginLogV.TPushNotiToken != "" {
			os = existingActiveAppLoginLogV.TOs
			pushNotiToken = existingActiveAppLoginLogV.TPushNotiToken
			break
		}
	}

	if pushNotiToken == "" {
		return nil
	}

	arrPN := base.CallSendPushNotificationIndApiStruct{
		Os:      os,
		RegID:   pushNotiToken,
		Subject: translatedSubject,
		Msg:     translatedMsg,
	}

	if arrData.CustMsg != "" {
		arrPN.CusMsg = arrData.CustMsg
	}

	err := base.CallSendPushNotificationIndApi(arrPN)

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
	}

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: " id = ?", CondValue: sysNoti.ID},
	)

	updateColumn := map[string]interface{}{"pn_send_status": 1, "updated_by": 0}
	err = models.UpdatesFn("sys_notification", arrUpdCond, updateColumn, false)

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
	}

	return nil
}

type ProcessPushNotificationBatchFromApiReqStruct struct {
	PNList []ProcessSendIndPushNotificationForm
	Source uint8
	PrjID  uint8
}

func ProcessPushNotificationBatchFromApiReq(tx *gorm.DB, arrData ProcessPushNotificationBatchFromApiReqStruct) error {

	if len(arrData.PNList) > 0 {
		for _, pnListV := range arrData.PNList {

			langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
			if pnListV.LangCode != "" {
				langCode = pnListV.LangCode
			}

			subjectParams := make(map[string]string, 0)
			if len(pnListV.SubjectParams) > 0 {
				for _, subjectParamsV := range pnListV.SubjectParams {
					subjectParams[subjectParamsV.Key] = subjectParamsV.Value
				}
			}
			msgParams := make(map[string]string, 0)
			if len(pnListV.MsgParams) > 0 {
				for _, msgParamsV := range pnListV.MsgParams {
					msgParams[msgParamsV.Key] = msgParamsV.Value
				}
			}

			arrEntMemCryptoFn := make([]models.WhereCondFn, 0)
			arrEntMemCryptoFn = append(arrEntMemCryptoFn,
				models.WhereCondFn{Condition: " crypto_address = ?", CondValue: pnListV.CryptoAddress},
				models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
			)
			arrEntMemCrypto, _ := models.GetEntMemberCryptoFn(arrEntMemCryptoFn, false)

			if arrEntMemCrypto == nil {
				continue
			}

			pnSub := base.PushNotificationContentStruct{
				Msg:    pnListV.Subject,
				Params: subjectParams,
			}
			encodedPNSub, _ := json.Marshal(pnSub)

			pnMsg := base.PushNotificationContentStruct{
				Msg:    pnListV.Msg,
				Params: msgParams,
			}
			encodedPNMsg, _ := json.Marshal(pnMsg)

			arrCrtSysNoti := models.AddSysNotificationStruct{
				ApiKeyID: int(arrData.PrjID),
				Type:     "member",
				// PNType      string    `json:"pn_type" gorm:"column:pn_type"`
				MemberID:     arrEntMemCrypto.MemberID,
				Title:        string(encodedPNSub),
				Msg:          string(encodedPNMsg),
				LangCode:     langCode,
				CustMsg:      pnListV.CustMsg,
				BShow:        1,
				PNSendStatus: 0,
				Status:       "A",
				CreatedBy:    strconv.Itoa(arrEntMemCrypto.MemberID),
			}
			models.AddSysNotification(arrCrtSysNoti)
		}
	}

	return nil
}

type MemberSysNotificationListStruct struct {
	ID        int    `json:"id"`
	Subject   string `json:"subject"`
	Msg       string `json:"msg"`
	CreatedAt string `json:"created_at"`
}

type MemberSysNotificationPaginateStruct struct {
	MemberID int
	LangCode string
	Page     int64
	ApiKeyID int
}

// func GetMemberSysNotificationLandListv1
func GetMemberSysNotificationLandListv1(arrData MemberSysNotificationPaginateStruct) interface{} {

	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")
	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue2, 10, 64)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sys_notification.api_key_id = ? ", CondValue: arrData.ApiKeyID},
		models.WhereCondFn{Condition: " sys_notification.member_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " sys_notification.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " sys_notification.b_show = ? ", CondValue: 1},
	)

	arrDataList, _ := models.GetSysNotificationLimitFn(arrCond, int(limit), false)
	arrNewList := make([]MemberSysNotificationListStruct, 0)
	if len(arrDataList) > 0 {
		for _, arrDataListV := range arrDataList {
			createdAtString := arrDataListV.CreatedAt.Format("2006-01-02 15:04:05")

			subject := ""
			if arrDataListV.Title != "" {
				var pnSubject base.PushNotificationContentStruct
				err := json.Unmarshal([]byte(arrDataListV.Title), &pnSubject)
				if err != nil {
					subject = pnSubject.Msg
				} else {
					subject = helpers.TranslateV2(pnSubject.Msg, arrData.LangCode, pnSubject.Params)
				}
			}

			msg := ""
			if arrDataListV.Msg != "" {
				var pnMsg base.PushNotificationContentStruct
				err := json.Unmarshal([]byte(arrDataListV.Msg), &pnMsg)
				if err != nil {
					msg = pnMsg.Msg
				} else {
					msg = helpers.TranslateV2(pnMsg.Msg, arrData.LangCode, pnMsg.Params)
				}
			}

			arrNewList = append(arrNewList,
				MemberSysNotificationListStruct{
					ID:        arrDataListV.ID,
					Subject:   subject,
					Msg:       msg,
					CreatedAt: createdAtString,
				},
			)
		}
	}

	return arrNewList
}

// func GetMemberSysNotificationPaginateListv1
func GetMemberSysNotificationPaginateListv1(arrData MemberSysNotificationPaginateStruct) interface{} {

	// arrNewMemberAnnouncementList := make([]MemberSysNotificationListStruct, 0)
	arrDataReturn := app.ArrDataResponseDefaultList{
		CurrentPageItems: make([]MemberSysNotificationListStruct, 0),
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sys_notification.api_key_id = ? ", CondValue: arrData.ApiKeyID},
		models.WhereCondFn{Condition: " sys_notification.member_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " sys_notification.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " sys_notification.b_show = ? ", CondValue: 1},
	)

	arrPaginateData, arrDataList, _ := models.GetSysNotificationPaginateFn(arrCond, arrData.Page, false)
	arrNewList := make([]MemberSysNotificationListStruct, 0)
	if len(arrDataList) > 0 {
		for _, arrDataListV := range arrDataList {
			createdAtString := arrDataListV.CreatedAt.Format("2006-01-02 15:04:05")

			subject := ""
			if arrDataListV.Title != "" {
				var pnSubject base.PushNotificationContentStruct
				err := json.Unmarshal([]byte(arrDataListV.Title), &pnSubject)
				if err != nil {
					subject = pnSubject.Msg
				} else {
					subject = helpers.TranslateV2(pnSubject.Msg, arrData.LangCode, pnSubject.Params)
				}
			}

			msg := ""
			if arrDataListV.Msg != "" {
				var pnMsg base.PushNotificationContentStruct
				err := json.Unmarshal([]byte(arrDataListV.Msg), &pnMsg)
				if err != nil {
					msg = pnMsg.Msg
				} else {
					msg = helpers.TranslateV2(pnMsg.Msg, arrData.LangCode, pnMsg.Params)
				}
			}

			arrNewList = append(arrNewList,
				MemberSysNotificationListStruct{
					ID:        arrDataListV.ID,
					Subject:   subject,
					Msg:       msg,
					CreatedAt: createdAtString,
				},
			)
		}
	}

	arrDataReturn.CurrentPage = int(arrPaginateData.CurrentPage)
	arrDataReturn.PerPage = int(arrPaginateData.PerPage)
	arrDataReturn.TotalCurrentPageItems = int(arrPaginateData.TotalCurrentPageItems)
	arrDataReturn.TotalPage = int(arrPaginateData.TotalPage)
	arrDataReturn.TotalPageItems = int(arrPaginateData.TotalPageItems)
	arrDataReturn.CurrentPageItems = arrNewList

	return arrDataReturn
}

// func ProcessSendPushNotificationMsg
func ProcessSendPushNotificationMsg(manual bool) {
	settingID := "pn_process_background_setting"
	arrSettingRst, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil || arrSettingRst == nil {
		fmt.Println("no pn_process_background_setting setting")
		return
	}

	if arrSettingRst.InputType1 != "1" && !manual {
		fmt.Println("pn_process_background_setting is off")
		return
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sys_notification.status = ? ", CondValue: "A"},
		models.WhereCondFn{Condition: " sys_notification.b_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " sys_notification.pn_send_status = ? ", CondValue: 0},
	)
	limitString := arrSettingRst.InputValue1
	limitInt, _ := strconv.Atoi(limitString)
	arrSysNotificationList, _ := models.GetSysNotificationLimitFn(arrCond, limitInt, false)

	if len(arrSysNotificationList) > 0 {
		for _, arrSysNotificationListV := range arrSysNotificationList {
			arrEntMemberFn := make([]models.WhereCondFn, 0)
			arrEntMemberFn = append(arrEntMemberFn,
				models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: arrSysNotificationListV.MemberID},
				models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
			)
			arrEntMember, _ := models.GetEntMemberFn(arrEntMemberFn, "", false)
			if arrEntMember == nil {
				// &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_crypto_address"}
				continue
			}

			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " id = ? ", CondValue: arrSysNotificationListV.ApiKeyID},
			)
			arrApiKey, _ := models.GetApiKeysFn(arrCond, "", false)

			if len(arrApiKey) < 1 {
				continue
			}

			existingActiveAppLoginLog, _ := models.GetExistingActiveAppLoginLog(arrEntMember.MainID, uint8(arrApiKey[0].SourceID), false)
			// begin transaction
			if len(existingActiveAppLoginLog) < 1 {
				continue
			}

			var os string
			var pushNotiToken string
			for _, existingActiveAppLoginLogV := range existingActiveAppLoginLog {
				if existingActiveAppLoginLogV.TPushNotiToken != "" {
					os = existingActiveAppLoginLogV.TOs
					pushNotiToken = existingActiveAppLoginLogV.TPushNotiToken
					break
				}
			}

			if pushNotiToken == "" {
				continue
			}

			subject := ""
			if arrSysNotificationListV.Title != "" {
				var pnSubject base.PushNotificationContentStruct
				err := json.Unmarshal([]byte(arrSysNotificationListV.Title), &pnSubject)
				if err != nil {
					subject = pnSubject.Msg
				} else {
					subject = helpers.TranslateV2(pnSubject.Msg, arrSysNotificationListV.LangCode, pnSubject.Params)
				}
			}

			msg := ""
			if arrSysNotificationListV.Msg != "" {
				var pnMsg base.PushNotificationContentStruct
				err := json.Unmarshal([]byte(arrSysNotificationListV.Msg), &pnMsg)
				if err != nil {
					msg = pnMsg.Msg
				} else {
					msg = helpers.TranslateV2(pnMsg.Msg, arrSysNotificationListV.LangCode, pnMsg.Params)
				}
			}
			arrPN := base.CallSendPushNotificationIndApiStruct{
				Os:      os,
				RegID:   pushNotiToken,
				Subject: subject,
				Msg:     msg,
			}

			if arrSysNotificationListV.CustMsg != "" {
				arrPN.CusMsg = arrSysNotificationListV.CustMsg
			}
			// fmt.Println("arrPN:", arrPN)
			err := base.CallSendPushNotificationIndApi(arrPN)

			if err != nil {
				continue
			}

			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: " id = ?", CondValue: arrSysNotificationListV.ID},
			)

			updateColumn := map[string]interface{}{"pn_send_status": 1, "updated_by": 0}
			err = models.UpdatesFn("sys_notification", arrUpdCond, updateColumn, false)

			if err != nil {
				continue
			}

			fmt.Println("success-id: ", arrSysNotificationListV.ID)
		}
	}
}
