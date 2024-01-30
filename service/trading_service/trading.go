package trading_service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/media_service"
	"github.com/smartblock/gta-api/service/product_service"
	"github.com/smartblock/gta-api/service/wallet_service"
)

type MemberTradingStrategyList struct {
	Code       string `json:"code"`
	Status     int    `json:"status"`
	Name       string `json:"name"`
	ComingSoon int    `json:"coming_soon"`
}

func GetMemberCurrentTradingStrategyList(memID int, username, langCode string) ([]MemberTradingStrategyList, string) {
	var (
		data                    = []MemberTradingStrategyList{}
		memberCurrentAPI        = GetMemberCurrentAPI(memID)
		comingSoonExceptionalID = []string{"sammy", "zoro"}
		comingSoon              = 1
	)

	if helpers.StringInSlice(username, comingSoonExceptionalID) {
		comingSoon = 0
	}

	if memberCurrentAPI.PlatformCode == "KC" {
		data = append(data,
			MemberTradingStrategyList{
				Code:       "CFRA",
				Status:     1,
				Name:       helpers.TranslateV2("crypto_funding_rates_arbitage", langCode, map[string]string{}),
				ComingSoon: 0,
			},
			MemberTradingStrategyList{
				Code:       "SGT",
				Status:     1,
				Name:       helpers.TranslateV2("spot_grid_trading", langCode, map[string]string{}),
				ComingSoon: 0,
			},
			MemberTradingStrategyList{
				Code:       "MT",
				Status:     1,
				Name:       helpers.TranslateV2("martingale_trading", langCode, map[string]string{}),
				ComingSoon: comingSoon,
			},
			MemberTradingStrategyList{
				Code:       "MTD",
				Status:     1,
				Name:       helpers.TranslateV2("reverse_martingale_trading", langCode, map[string]string{}),
				ComingSoon: comingSoon,
			},
		)
	} else {
		data = append(data,
			MemberTradingStrategyList{
				Code:       "CFRA",
				Status:     1,
				Name:       helpers.TranslateV2("crypto_funding_rates_arbitage", langCode, map[string]string{}),
				ComingSoon: 0,
			},
			MemberTradingStrategyList{
				Code:       "CIFRA",
				Status:     1,
				Name:       helpers.TranslateV2("crypto_index_funding_rates_arbitrage", langCode, map[string]string{}),
				ComingSoon: 0,
			},
			MemberTradingStrategyList{
				Code:       "SGT",
				Status:     1,
				Name:       helpers.TranslateV2("spot_grid_trading", langCode, map[string]string{}),
				ComingSoon: 0,
			},
			MemberTradingStrategyList{
				Code:       "MT",
				Status:     1,
				Name:       helpers.TranslateV2("martingale_trading", langCode, map[string]string{}),
				ComingSoon: comingSoon,
			},
			MemberTradingStrategyList{
				Code:       "MTD",
				Status:     1,
				Name:       helpers.TranslateV2("reverse_martingale_trading", langCode, map[string]string{}),
				ComingSoon: comingSoon,
			},
		)
	}

	return data, ""
}

type MemberTradingTncStatus struct {
	Status    int
	Signature string
	TncUrl    string
}

func GetMemberCurrentTradingTncStatus(memID int, langCode string) (MemberTradingTncStatus, string) {
	var (
		data             = MemberTradingTncStatus{}
		tncStatus int    = 0
		signature string = ""
	)

	// retrieve member trading tnc signature
	arrEntMemberTradingTncFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingTncFn = append(arrEntMemberTradingTncFn,
		models.WhereCondFn{Condition: "ent_member_trading_tnc.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ent_member_trading_tnc.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingTnc, err := models.GetEntMemberTradingTncFn(arrEntMemberTradingTncFn, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberCurrentTradingTncStatus:GetEntMemberTradingTncFn()", map[string]interface{}{"condition": arrEntMemberTradingTncFn}, err.Error(), true)
		return MemberTradingTncStatus{}, "something_went_wrong"
	}

	if len(arrEntMemberTradingTnc) > 0 {
		tncStatus = 1
		signature = arrEntMemberTradingTnc[0].Signature
	}

	// retrieve system trading tnc path
	arrGeneralSetup, _ := models.GetSysGeneralSetupByID("trading_tnc_path")

	if arrGeneralSetup == nil {
		base.LogErrorLog("GetMemberCurrentTradingTncStatus:GetSysGeneralSetupByID()", map[string]interface{}{"setting_id": "trading_tnc_path"}, "setting_not_found", true)
		return MemberTradingTncStatus{}, "something_went_wrong"
	}

	data.Status = tncStatus
	data.Signature = signature
	data.TncUrl = arrGeneralSetup.InputValue1

	return data, ""
}

// MemberTradingTnc struct
type MemberTradingTnc struct {
	MemberID        int
	SignatureFile   multipart.File
	SignatureHeader *multipart.FileHeader
	LangCode        string
}

func (input *MemberTradingTnc) UpdateMemberTradingTnc(tx *gorm.DB) string {

	//***** follow sammy format return string msg, not error ********

	//check member whether have active record
	arrEntMemberTradingTncFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingTncFn = append(arrEntMemberTradingTncFn,
		models.WhereCondFn{Condition: "ent_member_trading_tnc.member_id = ?", CondValue: input.MemberID},
		models.WhereCondFn{Condition: "ent_member_trading_tnc.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingTnc, err := models.GetEntMemberTradingTncFn(arrEntMemberTradingTncFn, "", false)
	if err != nil {
		base.LogErrorLog("UpdateMemberTradingTnc:GetEntMemberTradingTncFn()", map[string]interface{}{"condition": arrEntMemberTradingTncFn}, err.Error(), true)
		return "something_went_wrong"
	}

	if len(arrEntMemberTradingTnc) > 0 {
		return "record_already_exists"
	}

	//validate media type
	err = media_service.MediaValidation(input.SignatureFile, input.SignatureHeader, "image")
	if err != nil {
		return "unsupport_format"
	}

	// get image upload setting
	settingID := "upload_image_setting"
	arrMediaSetting, _ := models.GetSysGeneralSetupByID(settingID)
	sizeLimit := arrMediaSetting.SettingValue2
	filename := "tnc_upload_pic_" + "_" + strconv.Itoa(int(time.Now().Unix())) + filepath.Ext(input.SignatureHeader.Filename)
	module := "member/images/tnc"
	prefixName := "tnc_upload_pic"
	mediaData, err := media_service.UploadMedia(input.SignatureFile, filename, module, prefixName, sizeLimit, "")

	if err != nil {
		base.LogErrorLog("UpdateMemberTradingTnc:UploadMedia()", map[string]interface{}{"type": "upload_image_setting"}, err.Error(), true)
		return "something_went_wrong"
	}

	//save to ent_member_trading_tnc
	arrSaveEntMemberTradingTnc := models.EntMemberTradingTnc{
		MemberID:  input.MemberID,
		Signature: mediaData.FullURL,
		Status:    "A",
		CreatedAt: time.Now(),
		CreatedBy: strconv.Itoa(input.MemberID),
	}

	_, err = models.AddEntMemberTradingTnc(tx, arrSaveEntMemberTradingTnc)

	if err != nil {
		base.LogErrorLog("UpdateMemberTradingTnc:AddEntMemberTradingTnc", err, map[string]interface{}{"arrSave": arrSaveEntMemberTradingTnc, "err": err}, true)
		return "something_went_wrong"
	}

	return ""
}

type MemberTradingApiStatus struct {
	Status           int
	ResetStatus      int
	PopupReminder    string
	MemberTradingApi MemberTradingApi
}

type MemberTradingApi struct {
	Platform                string                    `json:"platform"`
	PlatformCode            string                    `json:"platform_code"`
	MemberTradingApiDetails []MemberTradingApiDetails `json:"api_management_details"`
	Strategy                []map[string]interface{}  `json:"api_strategy"`
}

type MemberTradingApiDetails struct {
	Module     string `json:"module"`
	ModuleCode string `json:"module_code"`
	ApiKey     string `json:"api_key"`
}

func GetMemberCurrentTradingApiStatus(memID int, langCode string) (MemberTradingApiStatus, string) {
	var (
		data             = MemberTradingApiStatus{}
		memberCurrentAPI = GetMemberCurrentAPI(memID)
	)

	if memberCurrentAPI.PlatformCode != "" {
		data.MemberTradingApi.Platform = memberCurrentAPI.Platform
		data.MemberTradingApi.PlatformCode = memberCurrentAPI.PlatformCode

		if len(memberCurrentAPI.ApiDetails) == 2 {
			data.Status = 1
		}

		date := time.Now()

		// process api list data
		for _, apiDetailsV := range memberCurrentAPI.ApiDetails {
			var memberTradingApiDetails = MemberTradingApiDetails{}
			memberTradingApiDetails.ApiKey = apiDetailsV.ApiKey
			memberTradingApiDetails.Module = helpers.TranslateV2(apiDetailsV.Module, langCode, nil)
			memberTradingApiDetails.ModuleCode = apiDetailsV.Module

			diff := date.Sub(apiDetailsV.CreatedAt)
			numberOfDays := int(diff.Hours() / 24)
			if numberOfDays >= 90 {
				data.PopupReminder = helpers.TranslateV2("api_key_already_set_for_:0_days_and_is_already_expired_please_reset_api_key_and_reopen_existing_trade", langCode, map[string]string{"0": strconv.Itoa(numberOfDays)})
			} else if numberOfDays >= 80 {
				data.PopupReminder = helpers.TranslateV2("api_key_already_set_for_:0_days_and_is_about_to_expire_please_reset_api_key_and_reopen_existing_trade", langCode, map[string]string{"0": strconv.Itoa(numberOfDays)})
			}

			data.MemberTradingApi.MemberTradingApiDetails = append(data.MemberTradingApi.MemberTradingApiDetails, memberTradingApiDetails)
		}

		// get available strategy
		arrSysTradingApiPlatformFn := []models.WhereCondFn{}
		arrSysTradingApiPlatformFn = append(arrSysTradingApiPlatformFn,
			models.WhereCondFn{Condition: " sys_trading_api_platform.code = ?", CondValue: memberCurrentAPI.PlatformCode},
		)
		arrSysTradingApiPlatform, _ := models.GetSysTradingApiPlatformFn(arrSysTradingApiPlatformFn, "", false)
		if len(arrSysTradingApiPlatform) <= 0 {
			base.LogErrorLog("tradingService:GetMemberCurrentTradingApiStatus():GetSysTradingApiPlatformFn():1", map[string]interface{}{"condition": arrSysTradingApiPlatformFn}, "sys_trading_api_platform_not_found", true)
			return MemberTradingApiStatus{}, "something_went_wrong"
		}

		arrSysTradingApiPlatformSetting, errMsg := GetSysTradingApiPlatformSetting(arrSysTradingApiPlatform[0].Setting)
		if errMsg != "" {
			base.LogErrorLog("tradingService:GetMemberCurrentTradingApiStatus():GetSysTradingApiPlatformSetting():1", map[string]interface{}{"value": arrSysTradingApiPlatform[0].Setting}, errMsg, true)
			return MemberTradingApiStatus{}, "something_went_wrong"
		}

		arrStrategy := []map[string]interface{}{}
		if len(arrSysTradingApiPlatformSetting.Strategy) > 0 {
			for _, arrSysTradingApiPlatformSettingV := range arrSysTradingApiPlatformSetting.Strategy {
				strategyCode := arrSysTradingApiPlatformSettingV.Code
				strategyName := ""

				arrPrdMasterFn := []models.WhereCondFn{}
				arrPrdMasterFn = append(arrPrdMasterFn, models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: strategyCode})
				arrPrdMaster, _ := models.GetPrdMasterFn(arrPrdMasterFn, "", false)

				if len(arrPrdMaster) > 0 {
					strategyName = arrPrdMaster[0].Name
				}

				comingSoon := 0
				// if helpers.StringInSlice(strategyCode, []string{"MT", "MTD"}) && !helpers.IntInSlice(memID, []int{64, 52, 39, 38, 116, 472}) {
				// 	comingSoon = 1
				// }

				arrStrategy = append(arrStrategy,
					map[string]interface{}{
						"code":        strategyCode,
						"name":        helpers.TranslateV2(strategyName, langCode, map[string]string{}),
						"status":      1,
						"coming_soon": comingSoon,
						"bg_path":     arrSysTradingApiPlatformSettingV.BgPath},
				)
			}
		}

		data.MemberTradingApi.Strategy = arrStrategy
	}

	// will only be able to reset api if there is no active auto bot running
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: memID},
		models.WhereCondFn{Condition: " sls_master.action = ? ", CondValue: "BOT"},
		models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "AP"},
	)
	arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberCurrentTradingApiStatus():GetSlsMasterFn():1", map[string]interface{}{"condition": arrSlsMasterFn}, err.Error(), true)
		return MemberTradingApiStatus{}, "something_went_wrong"
	}
	if len(arrSlsMaster) <= 0 {
		data.ResetStatus = 1 // set reset flag to 1
	}

	return data, ""
}

// func GetTradingApiPlatform
func GetTradingApiPlatform(langCode string) ([]map[string]interface{}, string) {
	var (
		arrReturnData  []map[string]interface{}
		arrReturnDataV map[string]interface{}
	)

	arrSysTradingApiPlatformFn := make([]models.WhereCondFn, 0)
	arrSysTradingApiPlatformFn = append(arrSysTradingApiPlatformFn,
		models.WhereCondFn{Condition: "sys_trading_api_platform.status = ?", CondValue: "A"},
	)
	arrSysTradingApiPlatform, err := models.GetSysTradingApiPlatformFn(arrSysTradingApiPlatformFn, "", false)
	if err != nil {
		base.LogErrorLog("GetTradingApiPlatform:GetSysTradingApiPlatformFn()", map[string]interface{}{"condition": arrSysTradingApiPlatformFn}, err.Error(), true)
		return nil, "something_went_wrong"
	}

	// process data
	if len(arrSysTradingApiPlatform) > 0 {
		for _, arrSysTradingApiPlatformV := range arrSysTradingApiPlatform {
			arrReturnDataV = map[string]interface{}{
				"code":    arrSysTradingApiPlatformV.Code,
				"name":    helpers.TranslateV2(arrSysTradingApiPlatformV.Name, langCode, nil),
				"img_url": arrSysTradingApiPlatformV.ImgUrl,
			}

			arrReturnData = append(arrReturnData, arrReturnDataV)
		}
	}

	return arrReturnData, ""
}

// UpdateMemberTradingApiParam struct
type UpdateMemberTradingApiParam struct {
	MemberID                int
	PlatformCode            string
	MemberTradingApiDetails []UpdateMemberTradingApiDetails
}

type UpdateMemberTradingApiDetails struct {
	Module     string
	ApiKey     string
	Secret     string
	Passphrase string
}

func UpdateMemberTradingApi(tx *gorm.DB, input UpdateMemberTradingApiParam) string {
	var (
		memID          = input.MemberID
		switchPlatform = false
	)

	// check if got active record
	arrEntMemberTradingApiFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingApiFn = append(arrEntMemberTradingApiFn,
		models.WhereCondFn{Condition: "ent_member_trading_api.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ent_member_trading_api.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "sys_trading_api_platform.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingApi, err := models.GetEntMemberTradingApiFn(arrEntMemberTradingApiFn, "", false)
	if err != nil {
		base.LogErrorLog("UpdateMemberTradingApi:GetEntMemberTradingApiFn()", map[string]interface{}{"condition": arrEntMemberTradingApiFn}, err.Error(), true)
		return "something_went_wrong"
	}

	if len(arrEntMemberTradingApi) > 0 {
		// check if got active running bot
		arrSlsMasterFn := make([]models.WhereCondFn, 0)
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: memID},
			models.WhereCondFn{Condition: " sls_master.action = ? ", CondValue: "BOT"},
			models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "AP"},
		)
		arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetEntMemberTradingApiFn():GetSlsMasterFn():1", map[string]interface{}{"condition": arrSlsMasterFn}, err.Error(), true)
			return "something_went_wrong"
		}
		if len(arrSlsMaster) > 0 {
			return "unable_to_reset_api_if_there_is_active_strategy_running"
		}

		// update current trading api status to I, and will not return error
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "member_id = ?", CondValue: memID},
		)
		updateColumn := map[string]interface{}{"status": "I", "updated_by": fmt.Sprint(memID)}
		err = models.UpdatesFnTx(tx, "ent_member_trading_api", arrUpdCond, updateColumn, false)
		if err != nil {
			base.LogErrorLog("tradingService:GetEntMemberTradingApiFn():UpdatesFnTx():1", map[string]interface{}{"arrCond": arrUpdCond, "updateColumn": updateColumn}, err.Error(), true)
			return "something_went_wrong"
		}

		// set switchPlatform
		if arrEntMemberTradingApi[0].PlatformCode != input.PlatformCode {
			switchPlatform = true
		}
	}

	// validate platform code
	arrSysTradingApiPlatformFn := make([]models.WhereCondFn, 0)
	arrSysTradingApiPlatformFn = append(arrSysTradingApiPlatformFn,
		models.WhereCondFn{Condition: "sys_trading_api_platform.code = ?", CondValue: input.PlatformCode},
		models.WhereCondFn{Condition: "sys_trading_api_platform.status = ?", CondValue: "A"},
	)
	arrSysTradingApiPlatform, err := models.GetSysTradingApiPlatformFn(arrSysTradingApiPlatformFn, "", false)
	if err != nil {
		base.LogErrorLog("UpdateMemberTradingApi:GetSysTradingApiPlatformFn()", map[string]interface{}{"condition": arrSysTradingApiPlatformFn}, err.Error(), true)
		return "something_went_wrong"
	}
	if len(arrSysTradingApiPlatform) <= 0 {
		return "invalid_platform_code"
	}

	var (
		curApiKey        = ""
		curApiSecret     = ""
		curApiPassphrase = ""
	)

	for _, memberTradingApiDetailsV := range input.MemberTradingApiDetails {
		if memberTradingApiDetailsV.ApiKey != curApiKey || memberTradingApiDetailsV.Secret != curApiSecret || (input.PlatformCode == "KC" && memberTradingApiDetailsV.Passphrase != curApiPassphrase) {
			// validate api key and secret key format
			bStatus := base.AlphaNumericOnly(memberTradingApiDetailsV.ApiKey)
			if !bStatus {
				return "api_key_can_only_contain_alphanumeric_characters"
			}

			if input.PlatformCode == "BN" {
				// validate secret
				bStatus = base.AlphaNumericOnly(memberTradingApiDetailsV.Secret)
				if !bStatus {
					return "api_secret_can_only_contain_alphanumeric_characters"
				}

				// validate api key and secret key status
				currentUnixTimestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
				data := fmt.Sprintf("timestamp=%v", currentUnixTimestamp)
				signature := util.GenerateHmacSHA256(memberTradingApiDetailsV.Secret, data, "")

				bnAccTradingStatusParam := GetMemberBinanceAccountApiTradingStatusParam{
					Timestamp: currentUnixTimestamp,
					Signature: signature,
					ApiKey:    memberTradingApiDetailsV.ApiKey,
				}
				bnAccTradingStatusRst, err := bnAccTradingStatusParam.GetBinanceAccountApiTradingStatus()
				if err != nil {
					base.LogErrorLog("UpdateMemberTradingApi():GetBinanceAccountApiTradingStatus()", err.Error(), map[string]interface{}{"param": bnAccTradingStatusParam}, true)
					return "invalid_api_key_or_secret"
				}

				if bnAccTradingStatusRst.Data.IsLocked {
					return "api_account_trading_function_is_locked"
				}
			} else if input.PlatformCode == "KC" {
				// validate secret
				bStatus = base.AlphaNumericCertainCharactersOnly(memberTradingApiDetailsV.Secret, []string{"-"})
				if !bStatus {
					return strings.ToLower(memberTradingApiDetailsV.Module) + "api_secret_can_only_contain_alphanumeric_characters_and_minus_symbol"
				}

				// validate passphrase
				bStatus = base.NoSpace(memberTradingApiDetailsV.Passphrase)
				if !bStatus {
					return strings.ToLower(memberTradingApiDetailsV.Module) + "_api_passphrase_cannot_contain_space"
				}

				// validate api key, secret key and passphrasestatus
				if memberTradingApiDetailsV.Module == "FUTURE" {
					kcAccTradingStatusParam := GetMemberKucoinFutureAccountApiTradingStatusParam{
						ApiKey:     memberTradingApiDetailsV.ApiKey,
						Secret:     memberTradingApiDetailsV.Secret,
						Passphrase: memberTradingApiDetailsV.Passphrase,
					}
					kcAccTradingStatusRst, err := kcAccTradingStatusParam.GetKucoinFutureAccountApiTradingStatus()
					if err != nil {
						base.LogErrorLog("UpdateMemberTradingApi():GetKucoinAccountApiTradingStatus()", err.Error(), map[string]interface{}{"param": kcAccTradingStatusParam}, false)
						return "invalid_future_api_key_secret_or_passphrase"
					}

					if kcAccTradingStatusRst.Code != "200000" {
						return "invalid_future_api_key_secret_or_passphrase"
					}
				} else {
					kcAccTradingStatusParam := GetMemberKucoinSpotAccountApiTradingStatusParam{
						ApiKey:     memberTradingApiDetailsV.ApiKey,
						Secret:     memberTradingApiDetailsV.Secret,
						Passphrase: memberTradingApiDetailsV.Passphrase,
					}
					kcAccTradingStatusRst, err := kcAccTradingStatusParam.GetKucoinSpotAccountApiTradingStatus()
					if err != nil {
						base.LogErrorLog("UpdateMemberTradingApi():GetKucoinAccountApiTradingStatus()", err.Error(), map[string]interface{}{"param": kcAccTradingStatusParam}, false)
						return "invalid_spot_api_key_secret_or_passphrase"
					}

					if kcAccTradingStatusRst.Code != "200000" {
						return "invalid_spot_api_key_secret_or_passphrase"
					}
				}

				// update prev api to "I"
				// arrUpdCond := make([]models.WhereCondFn, 0)
				// arrUpdCond = append(arrUpdCond,
				// 	models.WhereCondFn{Condition: "member_id = ?", CondValue: memID},
				// 	models.WhereCondFn{Condition: "platform != 'KC' OR module = ?", CondValue: memberTradingApiDetailsV.Module},
				// )
				// updateColumn := map[string]interface{}{"status": "I", "updated_by": fmt.Sprint(memID)}
				// err = models.UpdatesFnTx(tx, "ent_member_trading_api", arrUpdCond, updateColumn, false)
				// if err != nil {
				// 	base.LogErrorLog("tradingService:GetEntMemberTradingApiFn():UpdatesFnTx():1", map[string]interface{}{"arrCond": arrUpdCond, "updateColumn": updateColumn}, err.Error(), true)
				// 	return "something_went_wrong"
				// }
			}
		}

		curApiKey = memberTradingApiDetailsV.ApiKey
		curApiSecret = memberTradingApiDetailsV.Secret
		curApiPassphrase = memberTradingApiDetailsV.Passphrase

		// encrypt secret
		encryptedSecret := util.EncodeAscii85(curApiSecret)

		// store to ent_member_trading_api
		entMemberTradingApi := models.AddEntMemberTradingApiStruct{
			MemberID:      memID,
			Platform:      input.PlatformCode,
			Module:        memberTradingApiDetailsV.Module,
			ApiKey:        curApiKey,
			ApiSecret:     encryptedSecret,
			ApiPassphrase: curApiPassphrase,
			Status:        "A",
			CreatedBy:     fmt.Sprint(memID),
		}

		_, err = models.AddEntMemberTradingApi(tx, entMemberTradingApi)
		if err != nil {
			base.LogErrorLog("UpdateMemberTradingApi:AddEntMemberTradingApi()", map[string]interface{}{"input": entMemberTradingApi}, err.Error(), true)
			return "something_went_wrong"
		}
	}

	// clear wallet limit if switchPlatform is set to true
	if switchPlatform {
		arrUpdCond2 := make([]models.WhereCondFn, 0)
		arrUpdCond2 = append(arrUpdCond2,
			models.WhereCondFn{Condition: "member_id = ?", CondValue: memID},
		)
		updateColumn2 := map[string]interface{}{"status": "I", "updated_by": fmt.Sprint(memID)}
		err = models.UpdatesFnTx(tx, "ent_member_trading_wallet_limit", arrUpdCond2, updateColumn2, false)
		if err != nil {
			base.LogErrorLog("tradingService:GetEntMemberTradingApiFn():UpdatesFnTx():2", map[string]interface{}{"arrCond": arrUpdCond2, "updateColumn": updateColumn2}, err.Error(), true)
			return "something_went_wrong"
		}
	}

	return ""
}

type MemberCurrentAPI struct {
	Platform     string
	PlatformCode string
	ApiDetails   []MemberCurrentAPIDetails
}

type MemberCurrentAPIDetails struct {
	Module        string
	ApiKey        string
	ApiSecret     string
	ApiPassphrase string
	CreatedAt     time.Time
}

func GetMemberCurrentAPI(memberID int) MemberCurrentAPI {
	var (
		memberCurrentApi        = MemberCurrentAPI{}
		memberCurrentApiDetails = []MemberCurrentAPIDetails{}
	)

	// get current active limit
	arrEntMemberTradingApiFn := []models.WhereCondFn{}
	arrEntMemberTradingApiFn = append(arrEntMemberTradingApiFn,
		models.WhereCondFn{Condition: "ent_member_trading_api.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member_trading_api.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingApi, _ := models.GetEntMemberTradingApiFn(arrEntMemberTradingApiFn, "", false)
	if len(arrEntMemberTradingApi) <= 0 {
		return memberCurrentApi
	}

	memberCurrentApi.Platform = arrEntMemberTradingApi[0].Platform
	memberCurrentApi.PlatformCode = arrEntMemberTradingApi[0].PlatformCode

	for _, arrEntMemberTradingApiV := range arrEntMemberTradingApi {
		memberCurrentApiDetails = append(memberCurrentApiDetails,
			MemberCurrentAPIDetails{
				Module:        arrEntMemberTradingApiV.Module,
				ApiKey:        arrEntMemberTradingApiV.ApiKey,
				ApiSecret:     arrEntMemberTradingApiV.ApiSecret,
				ApiPassphrase: arrEntMemberTradingApiV.ApiPassphrase,
				CreatedAt:     arrEntMemberTradingApiV.CreatedAt,
			},
		)
	}

	memberCurrentApi.ApiDetails = memberCurrentApiDetails

	return memberCurrentApi
}

type BinanceAccountApiTradingStatus struct {
	Data struct {
		IsLocked           bool `json:"isLocked"`
		PlannedRecoverTime int  `json:"plannedRecoverTime"`
		TriggerCondition   struct {
			GCR  int `json:"GCR"`
			IFER int `json:"IFER"`
			UFR  int `json:"UFR"`
		} `json:"triggerCondition"`
		UpdateTime int `json:"updateTime"`
	} `json:"data"`
}

type GetMemberBinanceAccountApiTradingStatusParam struct {
	Timestamp string
	Signature string
	ApiKey    string
}

func (b *GetMemberBinanceAccountApiTradingStatusParam) GetBinanceAccountApiTradingStatus() (*BinanceAccountApiTradingStatus, error) {
	var (
		err      error
		response BinanceAccountApiTradingStatus
	)

	data := map[string]interface{}{
		"timestamp": b.Timestamp,
		"signature": b.Signature,
	}

	url := fmt.Sprintf("http://api.binance.com/sapi/v1/account/apiTradingStatus?timestamp=%v&signature=%v", b.Timestamp, b.Signature)
	header := map[string]string{
		"Content-Type": "application/json",
		"X-MBX-APIKEY": b.ApiKey,
	}

	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetBinanceAccountApiTradingStatus:RequestBinanceAPI", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetBinanceAccountApiTradingStatus:RequestBinanceAPI", res.Body, map[string]interface{}{"res": res, "data": data}, true)
		return &response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return &response, nil
}

type KucoinSpotAccountApiTradingStatus struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Id        string `json:"id"`
		Currency  string `json:"currency"`
		Type      string `json:"type"`
		Balance   string `json:"balance"`
		Available string `json:"available"`
		Holds     string `json:"holds"`
	} `json:"data"`
}

type GetMemberKucoinSpotAccountApiTradingStatusParam struct {
	ApiKey     string
	Secret     string
	Passphrase string
}

func (b *GetMemberKucoinSpotAccountApiTradingStatusParam) GetKucoinSpotAccountApiTradingStatus() (*KucoinSpotAccountApiTradingStatus, error) {
	var (
		err      error
		response KucoinSpotAccountApiTradingStatus
	)

	data := map[string]interface{}{
		"type":     "trade",
		"currency": "USDT",
	}
	encodedData, err := json.Marshal(data)
	if err != nil {
		base.LogErrorLog("GetKucoinSpotAccountApiTradingStatus:Marshal()", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	currentUnixTimestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	stringToSign := fmt.Sprintf("%v%s%s%s", currentUnixTimestamp, "GET", "/api/v1/accounts", string(encodedData)) // {timestamp+method+endpoint+body}

	// sha256 hmac encode
	signature := util.GenerateHmacSHA256(b.Secret, stringToSign, "base64")

	url := "https://api.kucoin.com/api/v1/accounts"

	header := map[string]string{
		"Content-Type":      "application/json",
		"KC-API-KEY":        b.ApiKey,
		"KC-API-SIGN":       signature,
		"KC-API-PASSPHRASE": b.Passphrase,
		"KC-API-TIMESTAMP":  currentUnixTimestamp,
	}

	res, err := base.RequestBinanceAPI("GET", url, header, data, &response)

	if err != nil {
		base.LogErrorLog("GetKucoinSpotAccountApiTradingStatus:RequestBinanceAPI", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetKucoinSpotAccountApiTradingStatus:RequestBinanceAPI", res.Body, map[string]interface{}{"res": res, "data": data}, true)
		return &response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return &response, nil
}

type KucoinFutureAccountApiTradingStatus struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccountEnquiry   float64 `json:"accountEquity"`
		UnrealisedPnl    float64 `json:"unrealisedPNL"`
		MarginBalance    float64 `json:"marginBalance"`
		PositionMargin   float64 `json:"positionMargin"`
		OrderMargin      float64 `json:"orderMargin"`
		FrozenFunds      float64 `json:"frozenFunds"`
		AvailableBalance float64 `json:"availableBalance"`
		Currency         string  `json:"currency"`
	} `json:"data"`
}

type GetMemberKucoinFutureAccountApiTradingStatusParam struct {
	ApiKey     string
	Secret     string
	Passphrase string
}

func (b *GetMemberKucoinFutureAccountApiTradingStatusParam) GetKucoinFutureAccountApiTradingStatus() (*KucoinFutureAccountApiTradingStatus, error) {
	var (
		err      error
		response KucoinFutureAccountApiTradingStatus
	)

	data := map[string]interface{}{}
	encodedData, err := json.Marshal(data)
	if err != nil {
		base.LogErrorLog("GetKucoinFutureAccountApiTradingStatus:Marshal()", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	currentUnixTimestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	stringToSign := fmt.Sprintf("%v%s%s%s", currentUnixTimestamp, "GET", "/api/v1/account-overview?currency=USDT", string(encodedData)) // {timestamp+method+endpoint+body}

	// sha256 hmac encode
	signature := util.GenerateHmacSHA256(b.Secret, stringToSign, "base64")

	url := "https://api-futures.kucoin.com/api/v1/account-overview?currency=USDT"

	header := map[string]string{
		"Content-Type":      "application/json",
		"KC-API-KEY":        b.ApiKey,
		"KC-API-SIGN":       signature,
		"KC-API-PASSPHRASE": b.Passphrase,
		"KC-API-TIMESTAMP":  currentUnixTimestamp,
	}

	res, err := base.RequestBinanceAPI("GET", url, header, data, &response)

	if err != nil {
		base.LogErrorLog("GetKucoinFutureAccountApiTradingStatus:RequestBinanceAPI", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetKucoinFutureAccountApiTradingStatus:RequestBinanceAPI", res.Body, map[string]interface{}{"res": res, "data": data}, true)
		return &response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return &response, nil
}

type KucoinFutureAccountWithdrawalLimit struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AvailableAmount float64 `json:"availableAmount"`
		Currency        string  `json:"currency"`
	} `json:"data"`
}
type GetMemberKucoinFutureAccountWithdrawalLimitParam struct {
	ApiKey     string
	Secret     string
	Passphrase string
}

func (b *GetMemberKucoinFutureAccountWithdrawalLimitParam) GetKucoinFutureAccountWithdrawalLimit() (*KucoinFutureAccountWithdrawalLimit, error) {
	var (
		err      error
		response KucoinFutureAccountWithdrawalLimit
	)

	data := map[string]interface{}{
		"currency": "USDT",
	}
	encodedData, err := json.Marshal(data)
	if err != nil {
		base.LogErrorLog("GetKucoinFutureAccountWithdrawalLimit:Marshal()", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	currentUnixTimestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	stringToSign := fmt.Sprintf("%v%s%s%s", currentUnixTimestamp, "GET", "/api/v1/withdrawals/quotas", string(encodedData)) // {timestamp+method+endpoint+body}

	// sha256 hmac encode
	signature := util.GenerateHmacSHA256(b.Secret, stringToSign, "base64")

	url := "https://api-futures.kucoin.com/api/v1/withdrawals/quotas" //future

	header := map[string]string{
		"Content-Type":      "application/json",
		"KC-API-KEY":        b.ApiKey,
		"KC-API-SIGN":       signature,
		"KC-API-PASSPHRASE": b.Passphrase,
		"KC-API-TIMESTAMP":  currentUnixTimestamp,
	}

	res, err := base.RequestBinanceAPI("GET", url, header, data, &response)

	if err != nil {
		base.LogErrorLog("GetKucoinFutureAccountWithdrawalLimit:RequestBinanceAPI", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetKucoinFutureAccountWithdrawalLimit:RequestBinanceAPI", res.Body, map[string]interface{}{"res": res, "data": data}, true)
		return &response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return &response, nil
}

type MemberMembershipStatus struct {
	Status             int
	RenewStatus        int
	MembershipExpiring int
	ExpiryDate         string
}

func GetMemberMembershipStatus(memID int, langCode string) (MemberMembershipStatus, string) {
	var (
		data                         = MemberMembershipStatus{}
		membershipStatus      int    = 0
		renewMembershipStatus int    = 0 // will only be 1 if membership_status = 1
		expireDate            string = ""
		membershipExpiring    int    = 1
	)

	arrEntMemberMembershipFn := make([]models.WhereCondFn, 0)
	arrEntMemberMembershipFn = append(arrEntMemberMembershipFn,
		models.WhereCondFn{Condition: " ent_member_membership.member_id = ? ", CondValue: memID},
		models.WhereCondFn{Condition: " ent_member_membership.b_valid = ? ", CondValue: 1},
	)
	arrEntMemberMembership, _ := models.GetEntMemberMembership(arrEntMemberMembershipFn, "", false)

	if len(arrEntMemberMembership) > 0 {
		if helpers.CompareDateTime(time.Now(), "<=", arrEntMemberMembership[0].ExpiredAt) {
			membershipStatus = 1
			renewMembershipStatus = 1
			expireDate = arrEntMemberMembership[0].ExpiredAt.Format("2006-01-02")
			// expireDate = arrEntMemberMembership[0].ExpiredAt.Format("2006-01-02 15:04:05")

			// calculate total number of days
			duration := arrEntMemberMembership[0].ExpiredAt.Sub(time.Now())
			days := int(duration.Hours() / 24)

			if days > 7 {
				membershipExpiring = 0
			}
		}
	}

	data.Status = membershipStatus
	data.RenewStatus = renewMembershipStatus
	data.ExpiryDate = expireDate
	data.MembershipExpiring = membershipExpiring // if membership about to expire in a week, renew button will turn red

	return data, ""
}

type MemberDepositStatus struct {
	Status               int
	DepositMin           float64
	DepositMax           float64
	DepositOptions       []float64
	DepositLow           int
	DepositLowAndWithBot int
	CurrentDepositAmount float64
}

func GetMemberDepositStatus(memID int, langCode string) (MemberDepositStatus, string) {
	var (
		data                                  = MemberDepositStatus{}
		depositStatus                 int     = 0
		currentDepositAmount          float64 = 0.00
		tradingDepositEwalletTypeCode string  = "TD"
		depositLow                    int     = 0
		depositLowAndWithBot          int     = -1
	)

	// get deposit options
	tradingGeneralSetting, errMsg := GetTradingGeneralSetting()
	if errMsg != "" {
		return MemberDepositStatus{}, errMsg
	}

	// get trading deposit wallet id by ewallet_type_code
	arrEwtSetupFn := make([]models.WhereCondFn, 0)
	arrEwtSetupFn = append(arrEwtSetupFn,
		models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: tradingDepositEwalletTypeCode},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberDepositStatus():GetEwtSetupFn():1", err.Error(), arrEwtSetupFn, true)
		return MemberDepositStatus{}, "something_went_wrong"
	}
	if arrEwtSetup == nil {
		base.LogErrorLog("tradingService:GetMemberDepositStatus():GetEwtSetupFn():1", "ewallet_setup_not_found", arrEwtSetupFn, true)
		return MemberDepositStatus{}, "something_went_wrong"
	}

	arrEwtSummaryFn := make([]models.WhereCondFn, 0)
	arrEwtSummaryFn = append(arrEwtSummaryFn,
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrEwtSetup.ID},
	)

	arrEwtSummary, err := models.GetEwtSummaryFn(arrEwtSummaryFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberDepositStatus():GetEwtSummaryFn():1", err.Error(), map[string]interface{}{"condition": arrEwtSummaryFn}, true)
		return MemberDepositStatus{}, "something_went_wrong"
	}

	if len(arrEwtSummary) > 0 {
		currentDepositAmount = arrEwtSummary[0].Balance

		if currentDepositAmount > 0 {
			depositStatus = 1
		}
	}

	if currentDepositAmount < 100 {
		depositLow = 1
	}

	// deposit low and with active bot
	depositLowAndWithBot = -1

	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " sls_master.action = ?", CondValue: "BOT"},
		models.WhereCondFn{Condition: " sls_master.status = ?", CondValue: "AP"},
	)
	arrSlsMaster, _ := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if len(arrSlsMaster) > 0 {
		if currentDepositAmount <= 0 {
			depositLowAndWithBot = 1
		} else {
			depositLowAndWithBot = 0
		}
	}

	data.Status = depositStatus
	data.DepositMin = 100
	data.DepositMax = 100000 // app site will validate with this and wallet balance
	data.CurrentDepositAmount = currentDepositAmount
	data.DepositLow = depositLow                     // if deposit low, deposit button will turn red
	data.DepositLowAndWithBot = depositLowAndWithBot // -1: no active bot, 0: got active bot and deposit not low, 1: got active bot and deposit is low
	data.DepositOptions = tradingGeneralSetting.DepositOptions

	return data, ""
}

type MemberTradingLimitStatus struct {
	Status                   int
	SpotWalletLimitMin       float64
	SpotWalletLimitMax       float64
	FutureWalletLimitMin     float64
	FutureWalletLimitMax     float64
	CurrentSpotWalletLimit   float64
	CurrentFutureWalletLimit float64
	WalletLimitOptions       []float64
}

func GetMemberTradingLimitStatus(memID int, langCode string) (MemberTradingLimitStatus, string) {
	var (
		data                             = MemberTradingLimitStatus{}
		limitStatus              int     = 0
		currentSpotWalletLimit   float64 = 0.00
		currentFutureWalletLimit float64 = 0.00
	)

	// get deposit options
	tradingGeneralSetting, errMsg := GetTradingGeneralSetting()
	if errMsg != "" {
		return MemberTradingLimitStatus{}, errMsg
	}

	// get member current spot wallet limit
	arrEntMemberTradingSpotWalletLimitFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingSpotWalletLimitFn = append(arrEntMemberTradingSpotWalletLimitFn,
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.module = ?", CondValue: "SPOT"},
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.status = ?", CondValue: "A"},
	)

	arrEntMemberTradingSpotWalletLimit, err := models.GetEntMemberTradingWalletLimit(arrEntMemberTradingSpotWalletLimitFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberTradingLimitStatus():GetEntMemberTradingWalletLimit():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingSpotWalletLimitFn}, true)
		return MemberTradingLimitStatus{}, "something_went_wrong"
	}

	if len(arrEntMemberTradingSpotWalletLimit) > 0 {
		currentSpotWalletLimit = arrEntMemberTradingSpotWalletLimit[0].TotalAmount
	}

	// get member current future wallet limit
	arrEntMemberTradingFutureWalletLimitFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingFutureWalletLimitFn = append(arrEntMemberTradingFutureWalletLimitFn,
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.module = ?", CondValue: "FUTURE"},
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.status = ?", CondValue: "A"},
	)

	arrEntMemberTradingFutureWalletLimit, err := models.GetEntMemberTradingWalletLimit(arrEntMemberTradingFutureWalletLimitFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberTradingLimitStatus():GetEntMemberTradingWalletLimit():2", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingFutureWalletLimitFn}, true)
		return MemberTradingLimitStatus{}, "something_went_wrong"
	}

	if len(arrEntMemberTradingFutureWalletLimit) > 0 {
		currentFutureWalletLimit = arrEntMemberTradingFutureWalletLimit[0].TotalAmount
	}

	// define current limit status
	if currentSpotWalletLimit > 0 && currentFutureWalletLimit > 0 {
		limitStatus = 1
	}

	// get spot wallet limit min
	var spotWalletLimitMin float64 = 0
	totalSalesFn := make([]models.WhereCondFn, 0)
	totalSalesFn = append(totalSalesFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "BOT"},
		models.WhereCondFn{Condition: "prd_master.code IN(?,'MT') ", CondValue: "SGT"},
		// models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"}, // include liquitationed record
	)
	totalSales, err := models.GetTotalSalesAmount(totalSalesFn, false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberTradingLimitStatus():GetTotalSalesAmount():1", map[string]interface{}{"condition": totalSalesFn}, err.Error(), true)
		return MemberTradingLimitStatus{}, "something_went_wrong"
	}
	if totalSales.TotalAmount > 0 {
		spotWalletLimitMin = totalSales.TotalAmount
	}

	data.SpotWalletLimitMin = spotWalletLimitMin // min must be at least equal to current active auto-bot amount
	data.SpotWalletLimitMax = 100000             // app site will validate with this and binance balance

	// get future wallet limit min
	var futureWalletLimitMin float64 = 0
	totalSales2Fn := make([]models.WhereCondFn, 0)
	totalSales2Fn = append(totalSales2Fn,
		models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "BOT"},
		models.WhereCondFn{Condition: "prd_master.code IN(?,'CIFRA') ", CondValue: "CFRA"},
		// models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"}, // include liquitationed record
	)
	totalSales2, err := models.GetTotalSalesAmount(totalSales2Fn, false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberTradingLimitStatus():GetTotalSalesAmount():2", map[string]interface{}{"condition": totalSales2Fn}, err.Error(), true)
		return MemberTradingLimitStatus{}, "something_went_wrong"
	}
	if totalSales2.TotalAmount > 0 {
		futureWalletLimitMin = totalSales2.TotalAmount
	}

	data.FutureWalletLimitMin = futureWalletLimitMin // min must be at least equal to current active auto-bot amount
	data.FutureWalletLimitMax = 100000               // app site will validate with this and binance balance

	data.Status = limitStatus
	data.CurrentSpotWalletLimit = currentSpotWalletLimit
	data.CurrentFutureWalletLimit = currentFutureWalletLimit
	data.WalletLimitOptions = tradingGeneralSetting.WalletLimitOptions

	return data, ""
}

// TradingGeneralSetting struct
type TradingGeneralSetting struct {
	DepositOptions     []float64 `json:"deposit_options"`
	WalletLimitOptions []float64 `json:"wallet_limit_options"`
}

// GetTradingGeneralSetting func
func GetTradingGeneralSetting() (TradingGeneralSetting, string) {
	sysGeneralSetup, _ := models.GetSysGeneralSetupByID("trading_general_setting")
	tradingGeneralSetting := sysGeneralSetup.InputValue1

	tradingGeneralSettingPointer := &TradingGeneralSetting{}
	if tradingGeneralSetting == "" {
		return *tradingGeneralSettingPointer, ""
	}

	err := json.Unmarshal([]byte(tradingGeneralSetting), tradingGeneralSettingPointer)
	if err != nil {
		base.LogErrorLog("tradingService:GetTradingGeneralSetting():Unmarshal():1", err.Error(), map[string]interface{}{"tradingGeneralSetting": tradingGeneralSetting}, true)
		return TradingGeneralSetting{}, "something_went_wrong"
	}

	return *tradingGeneralSettingPointer, ""
}

// AddTradingDepositParam struct
type AddTradingDepositParam struct {
	MemberID int
	Amount   float64
	Payments string
}

// AddTradingDeposit func
func AddTradingDeposit(tx *gorm.DB, addTradingDepositParam AddTradingDepositParam, langCode string) (app.MsgStruct, map[string]string) {
	var (
		docNo                         string
		docType                       string  = "TD"
		memberID                      int     = addTradingDepositParam.MemberID
		payableAmt                    float64 = addTradingDepositParam.Amount
		module                        string  = "TRADING_DEPOSIT"
		prdCurrencyCode               string  = "USDT"
		tradingDepositEwalletTypeCode string  = "TD"
		min                           float64 = 100
		max                           float64 = 100000
	)

	// validate min max
	if payableAmt < min {
		return app.MsgStruct{Msg: "minimum_deposit_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(min, 0, ".", ",", true)}}, nil
	}

	if payableAmt > max {
		return app.MsgStruct{Msg: "maximum_deposit_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(max, 0, ".", ",", true)}}, nil
	}

	// get trading general setting
	// tradingGeneralSetting, errMsg := GetTradingGeneralSetting()
	// if errMsg != "" {
	// 	return app.MsgStruct{Msg: errMsg}, nil
	// }

	// validate amount
	// if !helpers.Float64InSlice(payableAmt, tradingGeneralSetting.DepositOptions) {
	// 	return app.MsgStruct{Msg: "invalid_amount"}, nil
	// }

	// get doc_no
	db := models.GetDB()
	docNo, err := models.GetRunningDocNo(docType, db) //get doc no
	if err != nil {
		base.LogErrorLog("tradingService:AddTradingDeposit():GetRunningDocNo():1", err.Error(), map[string]interface{}{"docType": docType}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	err = models.UpdateRunningDocNo(docType, db) //update doc no
	if err != nil {
		base.LogErrorLog("tradingService:AddTradingDeposit():UpdateRunningDocNo():1", err.Error(), map[string]interface{}{"docType": docType}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// validate payment
	paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStruct(addTradingDepositParam.Payments)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// validate payment with pay amount + deduct wallet
	msgStruct, arrData := wallet_service.PaymentProcess(tx, wallet_service.PaymentProcessStruct{
		MemberID:        memberID,
		PrdCurrencyCode: prdCurrencyCode,
		Module:          module,
		Type:            "DEFAULT",
		DocNo:           docNo,
		Remark:          "",
		Amount:          payableAmt,
		Payments:        paymentStruct,
	}, 0, langCode)

	if msgStruct.Msg != "" {
		return msgStruct, nil
	}

	// insert to ent_member_trading_deposit
	var addEntMemberTradingDepositParam = models.EntMemberTradingDeposit{
		MemberID:    memberID,
		DocNo:       docNo,
		TotalAmount: payableAmt,
		CreatedBy:   fmt.Sprint(memberID),
	}

	_, err = models.AddEntMemberTradingDeposit(tx, addEntMemberTradingDepositParam)
	if err != nil {
		base.LogErrorLog("tradingService:AddTradingDeposit():AddEntMemberTradingDeposit():1", err.Error(), map[string]interface{}{"param": addEntMemberTradingDepositParam}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// get trading deposit wallet id by ewallet_type_code
	arrEwtSetupFn := make([]models.WhereCondFn, 0)
	arrEwtSetupFn = append(arrEwtSetupFn,
		models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: tradingDepositEwalletTypeCode},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:AddTradingDeposit():GetEwtSetupFn():1", err.Error(), arrEwtSetupFn, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if arrEwtSetup == nil {
		base.LogErrorLog("tradingService:AddTradingDeposit():GetEwtSetupFn():1", "ewallet_setup_not_found", arrEwtSetupFn, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// transfer in amount to TD wallet
	ewtIn := wallet_service.SaveMemberWalletStruct{
		EntMemberID:     memberID,
		EwalletTypeID:   arrEwtSetup.ID,
		TotalIn:         payableAmt,
		TransactionType: "TRADING_DEPOSIT",
		DocNo:           docNo,
		CreatedBy:       fmt.Sprint(memberID),
	}

	_, err = wallet_service.SaveMemberWallet(tx, ewtIn)
	if err != nil {
		base.LogErrorLog("tradingService:AddTradingDeposit():SaveMemberWallet():1", err.Error(), ewtIn, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	return app.MsgStruct{Msg: ""}, arrData
}

// WithdrawTradingDepositParam struct
type WithdrawTradingDepositParam struct {
	MemberID        int
	Amount          float64
	EwalletTypeCode string
}

// WithdrawTradingDeposit func
func WithdrawTradingDeposit(tx *gorm.DB, addTradingDepositParam WithdrawTradingDepositParam, langCode string) app.MsgStruct {
	var (
		docNo                         string
		docType                       string  = "TDW"
		memberID                      int     = addTradingDepositParam.MemberID
		withdrawAmount                float64 = addTradingDepositParam.Amount
		withdrawToEwtTypeCode         string  = addTradingDepositParam.EwalletTypeCode
		withdrawToEwtTypeID           int
		tradingDepositEwalletTypeCode string = "TD"
		tradingDepositEwalletTypeID   int
		min                           float64 = 10
		max                           float64 = 100000
		multipleOf                    float64 = 10
	)

	// get trading deposit wallet id by ewallet_type_code
	arrEwtSetupFn := make([]models.WhereCondFn, 0)
	arrEwtSetupFn = append(arrEwtSetupFn,
		models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: tradingDepositEwalletTypeCode},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():GetEwtSetupFn():1", err.Error(), arrEwtSetupFn, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}
	if arrEwtSetup == nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():GetEwtSetupFn():1", "ewallet_setup_not_found", arrEwtSetupFn, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	tradingDepositEwalletTypeID = arrEwtSetup.ID

	// validate withdraw amount
	arrEwtSummaryFn := make([]models.WhereCondFn, 0)
	arrEwtSummaryFn = append(arrEwtSummaryFn,
		models.WhereCondFn{Condition: " member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: " ewallet_type_id = ?", CondValue: tradingDepositEwalletTypeID},
	)
	arrEwtSummary, err := models.GetEwtSummaryFn(arrEwtSummaryFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():GetEwtSummaryFn():1", err.Error(), arrEwtSummary, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}
	if len(arrEwtSummary) <= 0 || arrEwtSummary[0].Balance < withdrawAmount {
		return app.MsgStruct{Msg: "insufficient_deposit_balance"}
	}

	// validate min max
	if withdrawAmount < min {
		return app.MsgStruct{Msg: "minimum_deposit_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(min, 0, ".", ",", true)}}
	}

	if withdrawAmount > max {
		return app.MsgStruct{Msg: "maximum_deposit_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(max, 0, ".", ",", true)}}
	}

	// validate if is multiple of
	if !helpers.IsMultipleOf(withdrawAmount, multipleOf) {
		return app.MsgStruct{Msg: "deposit_amount_must_be_multiple_of_:0", Params: map[string]string{"0": helpers.CutOffDecimal(multipleOf, 2, ".", ",")}}
	}

	// validate ewallet type code
	if !helpers.StringInSlice(withdrawToEwtTypeCode, []string{"USDT", "USDC"}) {
		return app.MsgStruct{Msg: "invalid_ewallet_type_code"}
	}

	// get doc_no
	db := models.GetDB()
	docNo, err = models.GetRunningDocNo(docType, db) //get doc no
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():GetRunningDocNo():1", err.Error(), map[string]interface{}{"docType": docType}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}
	err = models.UpdateRunningDocNo(docType, db) //update doc no
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():UpdateRunningDocNo():1", err.Error(), map[string]interface{}{"docType": docType}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// insert to ent_member_trading_deposit_withdraw
	var addEntMemberTradingDepositWithdrawParam = models.EntMemberTradingDepositWithdraw{
		MemberID:    memberID,
		DocNo:       docNo,
		TotalAmount: withdrawAmount,
		CreatedBy:   fmt.Sprint(memberID),
	}

	_, err = models.AddEntMemberTradingDepositWithdraw(tx, addEntMemberTradingDepositWithdrawParam)
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():GetEntMemberTradingDepositWithdraw():1", err.Error(), map[string]interface{}{"param": addEntMemberTradingDepositWithdrawParam}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// deduct from deposit TD wallet
	ewtOut := wallet_service.SaveMemberWalletStruct{
		EntMemberID:     memberID,
		EwalletTypeID:   tradingDepositEwalletTypeID,
		TotalOut:        withdrawAmount,
		TransactionType: "TRADING_DEPOSIT_WITHDRAW",
		DocNo:           docNo,
		CreatedBy:       fmt.Sprint(memberID),
	}

	_, err = wallet_service.SaveMemberWallet(tx, ewtOut)
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():SaveMemberWallet():1", err.Error(), ewtOut, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// transfer in amount to USDT/USDC wallet
	arrEwtSetupFn2 := make([]models.WhereCondFn, 0)
	arrEwtSetupFn2 = append(arrEwtSetupFn2,
		models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: withdrawToEwtTypeCode},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEwtSetup2, err := models.GetEwtSetupFn(arrEwtSetupFn2, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():GetEwtSetupFn():2", err.Error(), arrEwtSetupFn2, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}
	if arrEwtSetup2 == nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():GetEwtSetupFn():2", "ewallet_setup_not_found", arrEwtSetupFn2, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	withdrawToEwtTypeID = arrEwtSetup2.ID

	ewtIn := wallet_service.SaveMemberWalletStruct{
		EntMemberID:     memberID,
		EwalletTypeID:   withdrawToEwtTypeID,
		TotalIn:         withdrawAmount,
		TransactionType: "TRADING_DEPOSIT_WITHDRAW",
		DocNo:           docNo,
		CreatedBy:       fmt.Sprint(memberID),
	}

	_, err = wallet_service.SaveMemberWallet(tx, ewtIn)
	if err != nil {
		base.LogErrorLog("tradingService:WithdrawTradingDeposit():SaveMemberWallet():1", err.Error(), ewtIn, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}
}

// UpdateTradingWalletLimitParam struct
type UpdateTradingWalletLimitParam struct {
	MemberID int
	Module   string
	Amount   float64
}

// UpdateTradingWalletLimit func
func UpdateTradingWalletLimit(tx *gorm.DB, updateTradingWalletLimitParam UpdateTradingWalletLimitParam, langCode string) app.MsgStruct {
	var (
		min          float64 = 0
		max          float64 = 100000
		memberID     int     = updateTradingWalletLimitParam.MemberID
		module       string  = updateTradingWalletLimitParam.Module
		limitAmount  float64 = updateTradingWalletLimitParam.Amount
		platform     string
		platformCode string
	)

	// validate module
	if !helpers.StringInSlice(module, []string{"SPOT", "FUTURE"}) {
		return app.MsgStruct{Msg: "invalid_module"}
	}

	// validate min max
	if limitAmount <= 0 {
		return app.MsgStruct{Msg: "invalid_amount"}
	}

	// get current active limit
	memberCurrentApi := GetMemberCurrentAPI(memberID)
	if memberCurrentApi.PlatformCode == "" {
		return app.MsgStruct{Msg: "please_setup_api_to_proceed"}
	}

	platform = memberCurrentApi.Platform
	platformCode = memberCurrentApi.PlatformCode

	// validate min active bot
	var totalActiveBotAmount float64 = 0
	totalSalesFn := make([]models.WhereCondFn, 0)
	totalSalesFn = append(totalSalesFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master_bot_setting.platform = ?", CondValue: platformCode},
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "BOT"},
		// models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"}, // include liquitationed record
	)
	if module == "SPOT" {
		// SGT + MT
		totalSalesFn = append(totalSalesFn,
			models.WhereCondFn{Condition: "prd_master.code IN(?,'MT')", CondValue: "SGT"},
		)

	} else if module == "FUTURE" {
		// CFRA + CIFRA + MTD
		totalSalesFn = append(totalSalesFn,
			models.WhereCondFn{Condition: "prd_master.code IN(?,'CIFRA','MTD')", CondValue: "CFRA"},
		)
	}
	totalSales, err := models.GetTotalSalesAmount(totalSalesFn, false)
	if err != nil {
		base.LogErrorLog("tradingService:UpdateTradingWalletLimit():GetTotalSalesAmount():1", map[string]interface{}{"condition": totalSalesFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}
	if totalSales.TotalAmount > 0 {
		totalActiveBotAmount = totalSales.TotalAmount
	}
	if totalActiveBotAmount > min {
		min = totalActiveBotAmount
	}

	if limitAmount < min {
		return app.MsgStruct{Msg: "minimum_limit_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(min, 0, ".", ",", true)}}
	}

	if limitAmount > max {
		return app.MsgStruct{Msg: "maximum_limit_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(max, 0, ".", ",", true)}}
	}

	// get trading general setting
	// tradingGeneralSetting, errMsg := GetTradingGeneralSetting()
	// if errMsg != "" {
	// 	return app.MsgStruct{Msg: errMsg}
	// }

	// // validate amount with options - off since member allowed to keyin custom amount
	// if !helpers.Float64InSlice(limitAmount, tradingGeneralSetting.WalletLimitOptions) {
	// 	return app.MsgStruct{Msg: "invalid_amount"}
	// }

	// validate binance balance - start sammmy offed for testing efficiency
	var memberPlatformBalanceParam = GetMemberStrategyBalanceStruct{
		MemberID: memberID,
		Platform: platformCode,
	}

	if module == "SPOT" {
		var memberPlatformBalance, errMsg = memberPlatformBalanceParam.GetMemberStrategyBalancev1()
		if errMsg != "" {
			return app.MsgStruct{Msg: errMsg}
		}
		var binanceBalance = memberPlatformBalance.Balance
		if limitAmount > binanceBalance {
			return app.MsgStruct{Msg: "insufficient_spot_balance"}
		}
	} else if module == "FUTURE" {
		var memberPlatformBalance, errMsg = memberPlatformBalanceParam.GetMemberStrategyFuturesBalancev1()
		if errMsg != "" {
			return app.MsgStruct{Msg: errMsg}
		}
		var binanceBalance = memberPlatformBalance.Balance
		if limitAmount > binanceBalance {
			return app.MsgStruct{Msg: "insufficient_future_balance"}
		}
	}
	// end sammmy offed for testing efficiency

	// update current to inactive
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "module = ?", CondValue: module},
	)
	updateColumn := map[string]interface{}{"status": "I", "updated_by": memberID}
	err = models.UpdatesFnTx(tx, "ent_member_trading_wallet_limit", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("tradingService:UpdateTradingWalletLimit()", "UpdatesFnTx():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// insert to ent_member_trading_wallet_limit
	var addEntMemberTradingWalletLimitParam = models.EntMemberTradingWalletLimit{
		MemberID:    memberID,
		Module:      module,
		TotalAmount: limitAmount,
		Status:      "A",
		CreatedBy:   fmt.Sprint(memberID),
	}

	_, err = models.AddEntMemberTradingWalletLimit(tx, addEntMemberTradingWalletLimitParam)
	if err != nil {
		base.LogErrorLog("tradingService:UpdateTradingWalletLimit():AddEntMemberTradingWalletLimit():1", err.Error(), map[string]interface{}{"param": addEntMemberTradingWalletLimitParam}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// add to sls_master_bot_log
	var remark = "#*updated_" + strings.ToLower(platform) + "_spot_wallet_limit_to*# "
	if module == "FUTURE" {
		remark = "#*updated_" + strings.ToLower(platform) + "_future_wallet_limit_to*# "
	}
	var addSlsMasterBotLog = models.SlsMasterBotLog{
		MemberID:   memberID,
		Status:     "A",
		RemarkType: "S",
		Remark:     remark + helpers.CutOffDecimalv2(limitAmount, 0, ".", ",", true),
		CreatedAt:  time.Now(),
		CreatedBy:  "AUTO",
	}

	_, err = models.AddSlsMasterBotLog(tx, addSlsMasterBotLog)
	if err != nil {
		base.LogErrorLog("tradingService:UpdateTradingWalletLimit():AddSlsMasterBotLog():1", err.Error(), map[string]interface{}{"param": addSlsMasterBotLog}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}
}

func GetMemberAutoTradingSetup(memID int, strategyCode string, langCode string) (map[string]interface{}, string) {
	var (
		data = map[string]interface{}{}
	)

	// validate strategy code
	memberCurrentTradingApiStatus, errMsg := GetMemberCurrentTradingApiStatus(memID, langCode)
	if errMsg != "" {
		return nil, errMsg
	}

	if memberCurrentTradingApiStatus.MemberTradingApi.PlatformCode == "" {
		return nil, "please_setup_api_to_proceed"
	}

	arrStategy := []string{}
	for _, strategyV := range memberCurrentTradingApiStatus.MemberTradingApi.Strategy {
		arrStategy = append(arrStategy, fmt.Sprint(strategyV["code"]))
	}
	if !helpers.StringInSlice(strategyCode, arrStategy) {
		return nil, "invalid_strategy_code"
	}

	data["keyin_min"] = 100
	data["currency_code"] = "USDT"
	data["color_code"] = helpers.AutoTradingColorCode(strategyCode)

	// get member available wallet limit
	data["available_wallet_limit"] = GetCurrentMemberAvailableWalletLimit(memID, strategyCode)

	if strategyCode != "CIFRA" {
		data["crypto_pair"] = GetCryptoPair(memID, memberCurrentTradingApiStatus.MemberTradingApi.PlatformCode, strategyCode)
	} else {
		data["keyin_min"] = 2000
		data["crypto_pair"] = []map[string]string{
			0: {
				"code": "DEFIUSDT",
				"name": "DEFI/USDT",
			},
		}
	}

	if strategyCode == "SGT" {
		var arrMode = []map[string]string{}
		arrMode = append(arrMode, map[string]string{
			"code": "ARITHMETIC",
			"name": helpers.TranslateV2("arithmetic_mode", langCode, map[string]string{}),
		},
			map[string]string{
				"code": "GEOMETRIC",
				"name": helpers.TranslateV2("geometric_mode", langCode, map[string]string{}),
			},
		)

		data["mode"] = arrMode
		data["default_mode"] = helpers.TranslateV2("arithmetic_mode", langCode, map[string]string{})
		data["default_mode_descriptions"] = helpers.TranslateV2(":0_days", langCode, map[string]string{"0": "7"})
	}

	if strategyCode == "MT" || strategyCode == "MTD" {
		data["min_first_order_amount"] = 11
	}

	return data, ""
}

type CryptoPair struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func GetCryptoPair(memID int, platform, strategyCode string) []CryptoPair {
	var (
		cryptoPair      = []CryptoPair{}
		cifraCryptoPair = "DEFIUSDT"
	)

	if platform == "KC" {
		cifraCryptoPair = "DEFIUSDTM"
	}

	arrSysTradingCryptoPairSetupFn := make([]models.WhereCondFn, 0)
	arrSysTradingCryptoPairSetupFn = append(arrSysTradingCryptoPairSetupFn,
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrSysTradingCryptoPairSetup, err := models.GetSysTradingCryptoPairSetupByPlatformFn(arrSysTradingCryptoPairSetupFn, "", platform, false)
	if err != nil {
		base.LogErrorLog("tradingService:GetCryptoPair():GetSysTradingCryptoPairSetupByPlatformFn():1", err.Error(), map[string]interface{}{"param": arrSysTradingCryptoPairSetupFn, "platform": platform}, true)
	}

	if len(arrSysTradingCryptoPairSetup) > 0 {
		for _, arrSysTradingCryptoPairSetupV := range arrSysTradingCryptoPairSetup {
			if strategyCode == "CIFRA" {
				if arrSysTradingCryptoPairSetupV.Code != cifraCryptoPair {
					continue
				}
			} else if arrSysTradingCryptoPairSetupV.Code == cifraCryptoPair {
				continue
			}

			// do checking to filter out used crypto_pair
			arrSlsMasterBotSettingFn := make([]models.WhereCondFn, 0)
			arrSlsMasterBotSettingFn = append(arrSlsMasterBotSettingFn,
				models.WhereCondFn{Condition: " sls_master_bot_setting.crypto_pair = ?", CondValue: arrSysTradingCryptoPairSetupV.Code},
				models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
				models.WhereCondFn{Condition: " sls_master.status = ?", CondValue: "AP"},
			)
			arrSlsMasterBotSetting, err := models.GetSlsMasterBotSetting(arrSlsMasterBotSettingFn, "", false)
			if err != nil {
				base.LogErrorLog("tradingService:GetCryptoPair():GetSlsMasterBotSetting():1", err.Error(), arrSlsMasterBotSettingFn, true)
			}
			if len(arrSlsMasterBotSetting) <= 0 {
				cryptoPair = append(cryptoPair, CryptoPair{
					Code: arrSysTradingCryptoPairSetupV.Code,
					Name: arrSysTradingCryptoPairSetupV.Name,
				})
			}
		}
	}

	return cryptoPair
}

type MemberAutoTradingFundingRate struct {
	FundingRate string `json:"funding_rate"`
}

func GetMemberAutoTradingFundingRate(memID int, cryptoPair string, langCode string) (MemberAutoTradingFundingRate, string) {
	var (
		data         = MemberAutoTradingFundingRate{}
		platformCode = ""
	)

	// validate strategy code
	memberCurrentTradingApiStatus, errMsg := GetMemberCurrentTradingApiStatus(memID, langCode)
	if errMsg != "" {
		return MemberAutoTradingFundingRate{}, "something_went_wrong"
	}
	if memberCurrentTradingApiStatus.MemberTradingApi.PlatformCode == "" {
		return MemberAutoTradingFundingRate{}, "please_setup_api_to_proceed"
	}

	platformCode = memberCurrentTradingApiStatus.MemberTradingApi.PlatformCode

	if platformCode == "BN" {
		// get funding rate
		arrBotPremiumIndexFn := make([]models.WhereCondFn, 0)
		arrBotPremiumIndexFn = append(arrBotPremiumIndexFn,
			models.WhereCondFn{Condition: "bot_premium_index.symbol = ?", CondValue: cryptoPair},
			models.WhereCondFn{Condition: "bot_premium_index.b_latest = ?", CondValue: 1},
		)
		arrBotPremiumIndex, err := models.GetBotPremiumIndexFn(arrBotPremiumIndexFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingFundingRate:GetBotPremiumIndexFn()", map[string]interface{}{"condition": arrBotPremiumIndexFn}, err.Error(), true)
			return MemberAutoTradingFundingRate{}, "something_went_wrong"
		}
		if len(arrBotPremiumIndex) > 0 {
			data.FundingRate = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(float.Mul(arrBotPremiumIndex[0].LastFundingRate, 100), 2, ".", ",", true))
		}
	} else if platformCode == "KC" {
		// get funding rate
		arrBotPremiumIndexKcFn := make([]models.WhereCondFn, 0)
		arrBotPremiumIndexKcFn = append(arrBotPremiumIndexKcFn,
			models.WhereCondFn{Condition: "bot_premium_index_kc.symbol = ?", CondValue: cryptoPair},
			models.WhereCondFn{Condition: "bot_premium_index_kc.b_latest = ?", CondValue: 1},
		)
		arrBotPremiumIndexKc, err := models.GetBotPremiumIndexKcFn(arrBotPremiumIndexKcFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingFundingRate:GetBotPremiumIndexKcFn()", map[string]interface{}{"condition": arrBotPremiumIndexKcFn}, err.Error(), true)
			return MemberAutoTradingFundingRate{}, "something_went_wrong"
		}
		if len(arrBotPremiumIndexKc) > 0 {
			data.FundingRate = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(float.Mul(arrBotPremiumIndexKc[0].LastFundingRate, 100), 2, ".", ",", true))
		}
	}

	// arrBotFundingRateFn := make([]models.WhereCondFn, 0)
	// arrBotFundingRateFn = append(arrBotFundingRateFn,
	// 	models.WhereCondFn{Condition: "bot_funding_rate.symbol = ?", CondValue: cryptoPair},
	// 	models.WhereCondFn{Condition: "bot_funding_rate.b_latest = ?", CondValue: 1},
	// )
	// arrBotFundingRate, err := models.GetBotFundingRateFn(arrBotFundingRateFn, "", false)
	// if err != nil {
	// 	base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetBotFundingRateFn()", map[string]interface{}{"condition": arrBotFundingRateFn}, err.Error(), true)
	// 	return MemberAutoTradingFundingRate{}, "something_went_wrong"
	// }

	// if len(arrBotFundingRate) > 0 {
	// 	data.FundingRate = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(arrBotFundingRate[0].FundingRate, 6, ".", ",", true))
	// }

	return data, ""
}

type MemberAutoTradingGridData struct {
	LowerPrice               float64 `json:"lower_price"`
	LowerPriceStr            string  `json:"lower_price_str"`
	UpperPrice               float64 `json:"upper_price"`
	UpperPriceStr            string  `json:"upper_price_str"`
	MarketPrice              float64 `json:"market_price"`
	CryptoPricePercentage    float64 `json:"crypto_pair_percentage"`
	CalCryptoPricePercentage float64 `json:"cal_crypto_pair_percentage"`
	GridQuantity             float64 `json:"grid_quantity"`
	TakerRate                float64 `json:"taker_rate"`
	LowerSingleGridProfit    float64 `json:"lower_single_grid_profit"`
	UpperSingleGridProfit    float64 `json:"upper_single_grid_profit"`
	SingleGridProfit         float64 `json:"single_grid_profit"`
	SingleGridProfitStr      string  `json:"single_grid_profit_str"`
}

func GetMemberAutoTradingGridData(memID int, cryptoPair string, mode string, langCode string) (MemberAutoTradingGridData, string) {
	var (
		data                             = MemberAutoTradingGridData{}
		uppestPrice              float64 = 0
		lowestPrice              float64 = 0
		marketPrice              float64 = 0
		cryptoPricePercentage    float64 = 0
		calCryptoPricePercentage float64 = 0
		gridQuantity             float64 = 0
		takerRate                float64 = 0
		platform                 string  = ""
		// date7DaysBefore                  = time.Now().AddDate(0, 0, -6).Format("2006-01-02")
		date7DaysBefore = time.Now().AddDate(0, 0, -29).Format("2006-01-02")
	)

	if !helpers.StringInSlice(mode, []string{"ARITHMETIC", "GEOMETRIC"}) {
		return MemberAutoTradingGridData{}, "invalid_mode"
	}

	memberCurrentAPI := GetMemberCurrentAPI(memID)
	if memberCurrentAPI.PlatformCode == "" {
		return MemberAutoTradingGridData{}, "please_setup_api_to_proceed"
	}

	platform = memberCurrentAPI.PlatformCode

	// validate crypto pair
	var cryptoPairList = GetCryptoPair(memID, platform, "SGT")
	var cryptoPairStatus = false
	for _, cryptoPairListV := range cryptoPairList {
		if cryptoPairListV.Code == cryptoPair {
			cryptoPairStatus = true
		}
	}

	if !cryptoPairStatus {
		return MemberAutoTradingGridData{}, "invalid_crypto_pair"
	}

	if platform != "KC" {
		// get lower and upper price
		arrBotPremiumIndexFn := make([]models.WhereCondFn, 0)
		arrBotPremiumIndexFn = append(arrBotPremiumIndexFn,
			models.WhereCondFn{Condition: "bot_premium_index.symbol = ?", CondValue: cryptoPair},
			models.WhereCondFn{Condition: "bot_premium_index.dt_timestamp >= ?", CondValue: date7DaysBefore},
		)
		arrBotPremiumIndex, err := models.GetBotPremiumIndexFn(arrBotPremiumIndexFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetBotPremiumIndexFn()", map[string]interface{}{"condition": arrBotPremiumIndexFn}, err.Error(), true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}

		if len(arrBotPremiumIndex) <= 1 {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetBotPremiumIndexFn()", map[string]interface{}{"condition": arrBotPremiumIndexFn}, "insufficient_data_to_get_lower_and_upper_price", true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}

		for arrBotPremiumIndexK, arrBotPremiumIndexV := range arrBotPremiumIndex {
			if arrBotPremiumIndexK == 0 { // first loop direct insert
				lowestPrice = arrBotPremiumIndexV.IndexPrice
				uppestPrice = arrBotPremiumIndexV.IndexPrice
			} else {
				if arrBotPremiumIndexV.IndexPrice < lowestPrice {
					lowestPrice = arrBotPremiumIndexV.IndexPrice
				} else if arrBotPremiumIndexV.IndexPrice > uppestPrice {
					uppestPrice = arrBotPremiumIndexV.IndexPrice
				}
			}
		}
	} else {
		// get lower and upper price
		arrBotPremiumIndexKcFn := make([]models.WhereCondFn, 0)
		arrBotPremiumIndexKcFn = append(arrBotPremiumIndexKcFn,
			models.WhereCondFn{Condition: "bot_premium_index_kc.symbol = ?", CondValue: cryptoPair},
			models.WhereCondFn{Condition: "bot_premium_index_kc.dt_timestamp >= ?", CondValue: date7DaysBefore},
		)
		arrBotPremiumIndexKc, err := models.GetBotPremiumIndexKcFn(arrBotPremiumIndexKcFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetBotPremiumIndexKcFn()", map[string]interface{}{"condition": arrBotPremiumIndexKcFn}, err.Error(), true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}

		if len(arrBotPremiumIndexKc) <= 1 {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetBotPremiumIndexKcFn()", map[string]interface{}{"condition": arrBotPremiumIndexKcFn}, "insufficient_data_to_get_lower_and_upper_price", true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}

		for arrBotPremiumIndexKcK, arrBotPremiumIndexKcV := range arrBotPremiumIndexKc {
			if arrBotPremiumIndexKcK == 0 { // first loop direct insert
				lowestPrice = arrBotPremiumIndexKcV.IndexPrice
				uppestPrice = arrBotPremiumIndexKcV.IndexPrice
			} else {
				if arrBotPremiumIndexKcV.IndexPrice < lowestPrice {
					lowestPrice = arrBotPremiumIndexKcV.IndexPrice
				} else if arrBotPremiumIndexKcV.IndexPrice > uppestPrice {
					uppestPrice = arrBotPremiumIndexKcV.IndexPrice
				}
			}
		}
	}

	data.LowerPrice = lowestPrice
	data.UpperPrice = uppestPrice

	if lowestPrice <= 0 || uppestPrice <= 0 {
		base.LogErrorLog("tradingService:GetMemberAutoTradingGridData()", map[string]interface{}{"lowestPrice": lowestPrice, "uppestPrice": uppestPrice}, "lowestPrice_or_uppestPrice_invalid_value", true)
		return MemberAutoTradingGridData{}, "something_went_wrong"
	}

	// data.LowerPrice = fmt.Sprintf("%.6f", 10000000.123456789)
	data.LowerPriceStr = helpers.CutOffDecimalv2(lowestPrice, 6, ".", ",", true)
	data.UpperPriceStr = helpers.CutOffDecimalv2(uppestPrice, 6, ".", ",", true)

	// get grid_quantity and admin_fee from sys_trading_crypto_pair_setup
	arrSysTradingCryptoPairSetupFn := make([]models.WhereCondFn, 0)
	arrSysTradingCryptoPairSetupFn = append(arrSysTradingCryptoPairSetupFn,
		models.WhereCondFn{Condition: "code = ?", CondValue: cryptoPair},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)
	arrSysTradingCryptoPairSetup, err := models.GetSysTradingCryptoPairSetupByPlatformFn(arrSysTradingCryptoPairSetupFn, "", platform, false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetSysTradingCryptoPairSetupByPlatformFn()", map[string]interface{}{"condition": arrSysTradingCryptoPairSetupFn, "platform": platform}, err.Error(), true)
		return MemberAutoTradingGridData{}, "something_went_wrong"
	}
	if len(arrSysTradingCryptoPairSetup) <= 0 {
		return MemberAutoTradingGridData{}, "invalid_crypto_pair"
	}

	// gridQuantity = float64(arrSysTradingCryptoPairSetup[0].GridQuantity) // removed since 5th april 2022 and is calculated instead
	takerRate = arrSysTradingCryptoPairSetup[0].TakerRate

	// get marketPrice
	if platform != "KC" {
		arrBinancePrice, err := GetBinanceCryptoPrice(cryptoPair)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}

		marketPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:ParseFloat()", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}
	} else {
		arrKucoinPrice, err := GetKucoinCryptoPrice(cryptoPair)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}

		if arrKucoinPrice.Code != "200000" {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair, "response": arrKucoinPrice}, arrKucoinPrice.Msg, true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}

		// marketPrice = arrKucoinPrice.Data.Price

		// kucoin api changed return data type of price from string to float
		marketPrice, err = strconv.ParseFloat(arrKucoinPrice.Data.Price, 64)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingGridData:ParseFloat()", map[string]interface{}{"value": arrKucoinPrice.Data.Price}, err.Error(), true)
			return MemberAutoTradingGridData{}, "something_went_wrong"
		}
	}

	// calculate grid quantity -> uppestPrice - lowestPrice / (1% of currennt coin price)
	cryptoPricePercentage = 2
	calCryptoPricePercentage = float.Mul(marketPrice, float.Div(cryptoPricePercentage, 100))
	gridQuantity = float.Div((uppestPrice - lowestPrice), calCryptoPricePercentage)

	// round up gridQuantity to whole number
	gridQuantity, err = strconv.ParseFloat(helpers.CutOffDecimalv2(gridQuantity, 0, ".", "", true), 64)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberAutoTradingGridData():strconv.ParseFloat():1", map[string]interface{}{"gridQuantity": gridQuantity}, err.Error(), true)
		return MemberAutoTradingGridData{}, "something_went_wrong"
	}

	// return relevant data
	data.MarketPrice = marketPrice
	data.CryptoPricePercentage = cryptoPricePercentage
	data.CalCryptoPricePercentage = calCryptoPricePercentage
	data.GridQuantity = gridQuantity
	data.TakerRate = takerRate

	if gridQuantity <= 1 {
		base.LogErrorLog("tradingService:GetMemberAutoTradingGridData()", map[string]interface{}{"uppestPrice": uppestPrice, "lowestPrice": lowestPrice, "marketPrice": marketPrice, "cryptoPricePercentage": cryptoPricePercentage, "calCryptoPricePercentage": calCryptoPricePercentage, "gridQuantity": gridQuantity}, "invalid_gridQuantity_will_cause_panic_in_calculation", true)
		return MemberAutoTradingGridData{}, "something_went_wrong"
	}

	if takerRate <= 0 {
		base.LogErrorLog("tradingService:GetMemberAutoTradingGridData()", map[string]interface{}{"takerRate": takerRate}, "invalid_takerRate_will_cause_panic_in_calculation", true)
		return MemberAutoTradingGridData{}, "something_went_wrong"
	}

	// calculate single grid profit
	if mode == "ARITHMETIC" {
		var lowestProfit = float.Div(float.Div((uppestPrice-lowestPrice), (gridQuantity-1)), uppestPrice) - float.Mul(2, float.Div(takerRate, 100))
		lowestProfit = float.Mul(lowestProfit, 100)

		var highestProfit = float.Div(float.Div((uppestPrice-lowestPrice), (gridQuantity-1)), lowestPrice) - float.Mul(2, float.Div(takerRate, 100))
		highestProfit = float.Mul(highestProfit, 100)

		data.LowerSingleGridProfit = lowestProfit
		data.UpperSingleGridProfit = highestProfit

		data.SingleGridProfitStr = fmt.Sprintf("%s%% - %s%% (%d %s)",
			helpers.CutOffDecimalv2(lowestProfit, 6, ".", ",", true),
			helpers.CutOffDecimalv2(highestProfit, 6, ".", ",", true),
			int(gridQuantity),
			helpers.TranslateV2("Grids", langCode, map[string]string{}),
		)
	} else if mode == "GEOMETRIC" {
		var singleGridProfit = ((math.Pow(float.Div(uppestPrice, lowestPrice), float.Div(1, (gridQuantity-1))) - 1) - float.Mul(2, float.Div(takerRate, 100)))
		singleGridProfit = float.Mul(singleGridProfit, 100)

		data.SingleGridProfit = singleGridProfit
		data.SingleGridProfitStr = fmt.Sprintf("%s%% (%d %s)",
			helpers.CutOffDecimalv2(singleGridProfit, 2, ".", ",", true),
			int(gridQuantity),
			helpers.TranslateV2("Grids", langCode, map[string]string{}),
		)
	}

	return data, ""
}

type MemberAutoTradingMartingaleData struct {
	FirstOrderPrice        string  `json:"first_order_price"`
	FirstOrderCurrencyCode string  `json:"first_order_currency_code"`
	PriceScale             string  `json:"price_scale"`
	TakeProfitCallback     string  `json:"take_profit_callback"`
	TakeProfit             string  `json:"take_profit"`
	AddShares              string  `json:"add_shares"`
	MinFirstOrderAmount    float64 `json:"min_first_order_amount"`
}

func GetMemberAutoTradingMartingaleData(memID int, cryptoPair string, langCode string) (MemberAutoTradingMartingaleData, string) {
	var (
		data                          = MemberAutoTradingMartingaleData{}
		platform               string = ""
		platformCryptoPairCode string = cryptoPair
		firstOrderPrice        float64
	)

	memberCurrentAPI := GetMemberCurrentAPI(memID)
	if memberCurrentAPI.PlatformCode == "" {
		return MemberAutoTradingMartingaleData{}, "please_setup_api_to_proceed"
	}

	platform = memberCurrentAPI.PlatformCode

	var cryptoPairList = GetCryptoPair(memID, platform, "MT")
	var cryptoPairStatus = false
	for _, cryptoPairListV := range cryptoPairList {
		if cryptoPairListV.Code == cryptoPair {
			cryptoPairStatus = true
		}
	}

	if !cryptoPairStatus {
		return MemberAutoTradingMartingaleData{}, "invalid_crypto_pair"
	}

	// get price_scale, take_profit_callback, and take_profit from sys_trading_crypto_pair_setup
	arrSysTradingCryptoPairSetupFn := make([]models.WhereCondFn, 0)
	arrSysTradingCryptoPairSetupFn = append(arrSysTradingCryptoPairSetupFn,
		models.WhereCondFn{Condition: "code = ?", CondValue: cryptoPair},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)
	arrSysTradingCryptoPairSetup, err := models.GetSysTradingCryptoPairSetupByPlatformFn(arrSysTradingCryptoPairSetupFn, "", platform, false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberAutoTradingMartingaleData:GetSysTradingCryptoPairSetupByPlatformFn()", map[string]interface{}{"condition": arrSysTradingCryptoPairSetupFn, "platform": platform}, err.Error(), true)
		return MemberAutoTradingMartingaleData{}, "something_went_wrong"
	}
	if len(arrSysTradingCryptoPairSetup) <= 0 {
		return MemberAutoTradingMartingaleData{}, "invalid_crypto_pair"
	}

	data.PriceScale = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(arrSysTradingCryptoPairSetup[0].PriceScale, 2, ".", ",", true))
	data.TakeProfitCallback = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(arrSysTradingCryptoPairSetup[0].TakeProfitRatio, 6, ".", ",", true))
	data.TakeProfit = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(arrSysTradingCryptoPairSetup[0].TakeProfitAdjustment, 6, ".", ",", true))
	data.AddShares = fmt.Sprintf("x%s", helpers.CutOffDecimalv2(float64(arrSysTradingCryptoPairSetup[0].AddShares), 0, ".", ",", true))

	// get first order price
	if platform != "KC" {
		arrBinancePrice, err := GetBinanceCryptoPrice(platformCryptoPairCode)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingMartingaleData:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": platformCryptoPairCode}, err.Error(), true)
			return MemberAutoTradingMartingaleData{}, "something_went_wrong"
		}

		firstOrderPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingMartingaleData:ParseFloat():1", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
			return MemberAutoTradingMartingaleData{}, "something_went_wrong"
		}
	} else {
		arrBotPremiumIndexKcFn := make([]models.WhereCondFn, 0)
		arrBotPremiumIndexKcFn = append(arrBotPremiumIndexKcFn,
			models.WhereCondFn{Condition: "bot_premium_index_kc.symbol = ?", CondValue: cryptoPair},
			models.WhereCondFn{Condition: "bot_premium_index_kc.b_latest = ?", CondValue: 1},
		)
		arrBotPremiumIndexKc, err := models.GetBotPremiumIndexKcFn(arrBotPremiumIndexKcFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingMartingaleData:GetBotPremiumIndexKcFn()", map[string]interface{}{"condition": arrBotPremiumIndexKcFn}, err.Error(), true)
			return MemberAutoTradingMartingaleData{}, "something_went_wrong"
		}

		if len(arrBotPremiumIndexKc) <= 0 {
			base.LogErrorLog("tradingService:GetMemberAutoTradingMartingaleData:GetBotPremiumIndexKcFn()", map[string]interface{}{"condition": arrBotPremiumIndexKcFn}, "emmpty_data", true)
			return MemberAutoTradingMartingaleData{}, "something_went_wrong"
		}

		platformCryptoPairCode = arrBotPremiumIndexKc[0].Symbol

		arrBinancePrice, err := GetKucoinCryptoPrice(platformCryptoPairCode)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingMartingaleData:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": platformCryptoPairCode}, err.Error(), true)
			return MemberAutoTradingMartingaleData{}, "something_went_wrong"
		}

		// firstOrderPrice = arrBinancePrice.Data.Price

		// kucoin api changed return data type of price from string to float
		firstOrderPrice, err = strconv.ParseFloat(arrBinancePrice.Data.Price, 64)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingMartingaleData:ParseFloat():2", map[string]interface{}{"value": arrBinancePrice.Data.Price}, err.Error(), true)
			return MemberAutoTradingMartingaleData{}, "something_went_wrong"
		}
	}

	// data.FirstOrderPrice = fmt.Sprintf("%s USDT", helpers.CutOffDecimalv2(firstOrderPrice, 6, ".", ",", true))
	data.FirstOrderPrice = helpers.CutOffDecimalv2(firstOrderPrice, 6, ".", "", true)
	data.FirstOrderCurrencyCode = "USDT"
	data.MinFirstOrderAmount = 11
	return data, ""
}

type MemberAutoTradingReverseMartingaleData struct {
	FirstOrderPrice        string  `json:"first_order_price"`
	FirstOrderCurrencyCode string  `json:"first_order_currency_code"`
	PriceScale             string  `json:"price_scale"`
	TakeProfitCallback     string  `json:"take_profit_callback"`
	TakeProfit             string  `json:"take_profit"`
	AddShares              string  `json:"add_shares"`
	MinFirstOrderAmount    float64 `json:"min_first_order_amount"`
}

func GetMemberAutoTradingReverseMartingaleData(memID int, cryptoPair string, langCode string) (MemberAutoTradingReverseMartingaleData, string) {
	var (
		data                          = MemberAutoTradingReverseMartingaleData{}
		platform               string = ""
		platformCryptoPairCode string = cryptoPair
		firstOrderPrice        float64
	)

	memberCurrentAPI := GetMemberCurrentAPI(memID)
	if memberCurrentAPI.PlatformCode == "" {
		return MemberAutoTradingReverseMartingaleData{}, "please_setup_api_to_proceed"
	}

	platform = memberCurrentAPI.PlatformCode

	var cryptoPairList = GetCryptoPair(memID, platform, "MTD")
	var cryptoPairStatus = false
	for _, cryptoPairListV := range cryptoPairList {
		if cryptoPairListV.Code == cryptoPair {
			cryptoPairStatus = true
		}
	}

	if !cryptoPairStatus {
		return MemberAutoTradingReverseMartingaleData{}, "invalid_crypto_pair"
	}

	// get price_scale, take_profit_callback and take_profit from sys_trading_crypto_pair_setup
	arrSysTradingCryptoPairSetupFn := make([]models.WhereCondFn, 0)
	arrSysTradingCryptoPairSetupFn = append(arrSysTradingCryptoPairSetupFn,
		models.WhereCondFn{Condition: "code = ?", CondValue: cryptoPair},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)
	arrSysTradingCryptoPairSetup, err := models.GetSysTradingCryptoPairSetupByPlatformFn(arrSysTradingCryptoPairSetupFn, "", platform, false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberAutoTradingReverseMartingaleData:GetSysTradingCryptoPairSetupByPlatformFn()", map[string]interface{}{"condition": arrSysTradingCryptoPairSetupFn, "platform": platform}, err.Error(), true)
		return MemberAutoTradingReverseMartingaleData{}, "something_went_wrong"
	}
	if len(arrSysTradingCryptoPairSetup) <= 0 {
		return MemberAutoTradingReverseMartingaleData{}, "invalid_crypto_pair"
	}

	data.PriceScale = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(arrSysTradingCryptoPairSetup[0].MtdPriceScale, 2, ".", ",", true))
	data.TakeProfitCallback = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(arrSysTradingCryptoPairSetup[0].MtdTakeProfitRatio, 6, ".", ",", true))
	data.TakeProfit = fmt.Sprintf("%s%%", helpers.CutOffDecimalv2(arrSysTradingCryptoPairSetup[0].TakeProfitAdjustment, 6, ".", ",", true))
	data.AddShares = fmt.Sprintf("x%s", helpers.CutOffDecimalv2(float64(arrSysTradingCryptoPairSetup[0].AddShares), 0, ".", ",", true))

	// get first order price
	if platform != "KC" {
		arrBinancePrice, err := GetBinanceCryptoPrice(platformCryptoPairCode)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingReverseMartingaleData:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": platformCryptoPairCode}, err.Error(), true)
			return MemberAutoTradingReverseMartingaleData{}, "something_went_wrong"
		}

		firstOrderPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingReverseMartingaleData:ParseFloat():1", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
			return MemberAutoTradingReverseMartingaleData{}, "something_went_wrong"
		}
	} else {
		arrBotPremiumIndexKcFn := make([]models.WhereCondFn, 0)
		arrBotPremiumIndexKcFn = append(arrBotPremiumIndexKcFn,
			models.WhereCondFn{Condition: "bot_premium_index_kc.symbol = ?", CondValue: cryptoPair},
			models.WhereCondFn{Condition: "bot_premium_index_kc.b_latest = ?", CondValue: 1},
		)
		arrBotPremiumIndexKc, err := models.GetBotPremiumIndexKcFn(arrBotPremiumIndexKcFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingReverseMartingaleData:GetBotPremiumIndexKcFn()", map[string]interface{}{"condition": arrBotPremiumIndexKcFn}, err.Error(), true)
			return MemberAutoTradingReverseMartingaleData{}, "something_went_wrong"
		}

		if len(arrBotPremiumIndexKc) <= 0 {
			base.LogErrorLog("tradingService:GetMemberAutoTradingReverseMartingaleData:GetBotPremiumIndexKcFn()", map[string]interface{}{"condition": arrBotPremiumIndexKcFn}, "emmpty_data", true)
			return MemberAutoTradingReverseMartingaleData{}, "something_went_wrong"
		}

		platformCryptoPairCode = arrBotPremiumIndexKc[0].Symbol

		arrBinancePrice, err := GetKucoinCryptoPrice(platformCryptoPairCode)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingReverseMartingaleData:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": platformCryptoPairCode}, err.Error(), true)
			return MemberAutoTradingReverseMartingaleData{}, "something_went_wrong"
		}

		// firstOrderPrice = arrBinancePrice.Data.Price

		// kucoin api changed return data type of price from string to float
		firstOrderPrice, err = strconv.ParseFloat(arrBinancePrice.Data.Price, 64)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingReverseMartingaleData:ParseFloat():2", map[string]interface{}{"value": arrBinancePrice.Data.Price}, err.Error(), true)
			return MemberAutoTradingReverseMartingaleData{}, "something_went_wrong"
		}
	}

	// data.FirstOrderPrice = fmt.Sprintf("%s USDT", helpers.CutOffDecimalv2(firstOrderPrice, 6, ".", ",", true))
	data.FirstOrderPrice = helpers.CutOffDecimalv2(firstOrderPrice, 6, ".", "", true)
	data.FirstOrderCurrencyCode = "USDT"

	minFirstOrderAmount := 0.00
	if platform == "KC" {
		// minFirstOrderAmount need to calculate from data grab in kucoin
		minFirstOrderAmount, _ = GetKucoinMinOrderAmount(cryptoPair)
	} else if platform == "BN" {
		// minFirstOrderAmount need to calculate from data grab in binace
		minFirstOrderAmount, _ = GetBinanceMinOrderAmount(cryptoPair)
	}

	data.MinFirstOrderAmount = minFirstOrderAmount
	return data, ""
}

type BinanceCryptoPriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func GetBinanceCryptoPrice(symbol string) (*BinanceCryptoPriceResponse, error) {

	var (
		err      error
		response *BinanceCryptoPriceResponse
	)

	apiSetting, _ := models.GetSysGeneralSetupByID("binance_api_setting")

	if apiSetting == nil {
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "invalid_setting"}
	}

	domain := apiSetting.InputValue1
	url := fmt.Sprintf("%sapi/v3/ticker/price?symbol=%s", domain, symbol)
	header := map[string]string{
		"Content-Type": "application/json",
	}
	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetBinanceCryptoPrice:RequestBinanceAPI()", err.Error(), map[string]interface{}{"err": err, "symbol": symbol}, true)
		return response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetBinanceCryptoPrice:RequestBinanceAPI()", res.Body, map[string]interface{}{"res": res, "symbol": symbol}, true)
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return response, nil
}

func GetBinanceFutureCryptoPrice(symbol string) (*BinanceCryptoPriceResponse, error) {

	var (
		err      error
		response *BinanceCryptoPriceResponse
	)

	apiSetting, _ := models.GetSysGeneralSetupByID("binance_api_setting")

	if apiSetting == nil {
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "invalid_setting"}
	}

	domain := apiSetting.InputValue2
	url := fmt.Sprintf("%sfapi/v1/ticker/price?symbol=%s", domain, symbol)
	header := map[string]string{
		"Content-Type": "application/json",
	}
	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetBinanceCryptoPrice:RequestBinanceAPI()", err.Error(), map[string]interface{}{"err": err, "symbol": symbol}, true)
		return response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetBinanceCryptoPrice:RequestBinanceAPI()", res.Body, map[string]interface{}{"res": res, "symbol": symbol}, true)
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return response, nil
}

type KucoinCryptoPriceResponse struct {
	Code string `json:"code"`
	Data struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	} `json:"data"`
	Msg string `json:"msg"`
}

func GetKucoinCryptoPrice(symbol string) (*KucoinCryptoPriceResponse, error) {

	var (
		err      error
		response *KucoinCryptoPriceResponse
	)

	apiSetting, _ := models.GetSysGeneralSetupByID("kucoin_api_setting")

	if apiSetting == nil {
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "invalid_setting"}
	}

	domain := apiSetting.InputValue1
	url := fmt.Sprintf("%sapi/v1/ticker?symbol=%s", domain, symbol)
	header := map[string]string{
		"Content-Type": "application/json",
	}
	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetKucoinCryptoPrice:RequestBinanceAPI()", err.Error(), map[string]interface{}{"err": err, "symbol": symbol}, true)
		return response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetKucoinCryptoPrice:RequestBinanceAPI()", res.Body, map[string]interface{}{"res": res, "symbol": symbol}, true)
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return response, nil
}

// ValidateMemberAutoTradingStatus func to validate on member api management, tnc signature, membership, deposit, wallet_limit status
func ValidateMemberAutoTradingStatus(memberID int) string {
	// validate on tnc signature
	arrEntMemberTradingTncFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingTncFn = append(arrEntMemberTradingTncFn,
		models.WhereCondFn{Condition: "ent_member_trading_tnc.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member_trading_tnc.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingTnc, err := models.GetEntMemberTradingTncFn(arrEntMemberTradingTncFn, "", false)
	if err != nil {
		base.LogErrorLog("ValidateMemberAutoTradingStatus:GetEntMemberTradingTncFn()", map[string]interface{}{"condition": arrEntMemberTradingTncFn}, err.Error(), true)
		return "something_went_wrong"
	}

	if len(arrEntMemberTradingTnc) <= 0 {
		return "please_agree_with_terms_and_condition_before_proceed"
	}

	// validate on api management
	arrEntMemberTradingApiFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingApiFn = append(arrEntMemberTradingApiFn,
		models.WhereCondFn{Condition: "ent_member_trading_api.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member_trading_api.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "sys_trading_api_platform.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingApi, err := models.GetEntMemberTradingApiFn(arrEntMemberTradingApiFn, "", false)
	if err != nil {
		base.LogErrorLog("ValidateMemberAutoTradingStatus:GetEntMemberTradingApiFn()", map[string]interface{}{"condition": arrEntMemberTradingApiFn}, err.Error(), true)
		return "something_went_wrong"
	}

	if len(arrEntMemberTradingApi) <= 0 {
		return "please_set_your_api_management_before_proceed"
	}

	// validate on membership
	arrEntMemberMembershipFn := make([]models.WhereCondFn, 0)
	arrEntMemberMembershipFn = append(arrEntMemberMembershipFn,
		models.WhereCondFn{Condition: " ent_member_membership.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: " ent_member_membership.b_valid = ? ", CondValue: 1},
	)
	arrEntMemberMembership, _ := models.GetEntMemberMembership(arrEntMemberMembershipFn, "", false)

	if len(arrEntMemberMembership) > 0 {
		if helpers.CompareDateTime(time.Now(), ">", arrEntMemberMembership[0].ExpiredAt) {
			return "please_purchase_membership_before_proceed"
		}
	}

	// validate on deposit
	var tradingDepositEwalletTypeCode string = "TD"

	// get trading deposit wallet id by ewallet_type_code
	arrEwtSetupFn := make([]models.WhereCondFn, 0)
	arrEwtSetupFn = append(arrEwtSetupFn,
		models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: tradingDepositEwalletTypeCode},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
	)
	arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:ValidateMemberAutoTradingStatus():GetEwtSetupFn():1", err.Error(), arrEwtSetupFn, true)
		return "something_went_wrong"
	}
	if arrEwtSetup == nil {
		base.LogErrorLog("tradingService:ValidateMemberAutoTradingStatus():GetEwtSetupFn():1", "ewallet_setup_not_found", arrEwtSetupFn, true)
		return "something_went_wrong"
	}

	arrEwtSummaryFn := make([]models.WhereCondFn, 0)
	arrEwtSummaryFn = append(arrEwtSummaryFn,
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrEwtSetup.ID},
	)

	arrEwtSummary, err := models.GetEwtSummaryFn(arrEwtSummaryFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:ValidateMemberAutoTradingStatus():GetEwtSummaryFn():1", err.Error(), map[string]interface{}{"condition": arrEwtSummaryFn}, true)
		return "something_went_wrong"
	}

	if len(arrEwtSummary) <= 0 || arrEwtSummary[0].Balance <= 0 {
		return "please_make_deposit_before_proceed"
	}

	if arrEwtSummary[0].Balance < 30 { // deposit less than 30 will not allow to add bot
		return "deposit_running_low_please_make_deposit_before_proceed"
	}

	// validate on wallet_limit
	arrEntMemberTradingWalletLimitFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.status = ?", CondValue: "A"},
	)

	arrEntMemberTradingWalletLimit, err := models.GetEntMemberTradingWalletLimit(arrEntMemberTradingWalletLimitFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:ValidateMemberAutoTradingStatus():GetEntMemberTradingWalletLimit():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingWalletLimitFn}, true)
		return "something_went_wrong"
	}

	if len(arrEntMemberTradingWalletLimit) <= 0 || arrEntMemberTradingWalletLimit[0].TotalAmount <= 0 {
		return "please_set_wallet_limit_before_proceed"
	}

	return ""
}

// MemberAutoTrading struct
type MemberAutoTrading struct {
	MemberID   int
	Platform   string
	PrdCode    string
	Type       string
	Amount     float64
	CryptoPair string
	LangCode   string
}

func AddMemberAutoTrading(tx *gorm.DB, input MemberAutoTrading) (int, string, app.MsgStruct) {
	var (
		platform        string = input.Platform
		settingType     string = input.Type
		cryptoPair      string = input.CryptoPair
		batchNo         string
		batchDocType    string = "BT"
		docNo           string
		docType         string = "BOT"
		memberID        int    = input.MemberID
		sponsorID       int
		prdMasterID     int
		prdCode         string = input.PrdCode
		prdGroup        string
		prdCurrencyCode string
		amount          float64 = input.Amount
		unitPrice       float64
		action          string    = "BOT"
		bnsAction       string    = "BOT"
		curDate         string    = base.GetCurrentDateTimeT().Format("2006-01-02")
		curDateTime     time.Time = base.GetCurrentDateTimeT()
		expiredAt, _              = base.StrToDateTime("9999-01-01", "2006-01-02")
	)

	// validate type
	if !helpers.StringInSlice(settingType, []string{"AI", "PROF"}) {
		return 0, "", app.MsgStruct{Msg: "invalid_type"}
	}

	// get prd_master setup
	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.code = ? ", CondValue: prdCode},
		models.WhereCondFn{Condition: " date(prd_master.date_start) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " date(prd_master.date_end) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " prd_master.status = ? ", CondValue: "A"},
	)
	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():GetPrdMasterFn():1", map[string]interface{}{"condition": arrPrdMasterFn}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}
	if len(arrPrdMaster) <= 0 {
		return 0, "", app.MsgStruct{Msg: "invalid_contract_code"}
	}

	prdMasterID = arrPrdMaster[0].ID
	prdCurrencyCode = arrPrdMaster[0].CurrencyCode
	unitPrice = arrPrdMaster[0].Amount
	prdGroup = arrPrdMaster[0].PrdGroup

	// validate if got same strategy active
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: " sls_master.prd_master_id = ? ", CondValue: prdMasterID},
		models.WhereCondFn{Condition: " sls_master.status = ? ", CondValue: "AP"},
	)
	arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():GetSlsMasterFn():1", map[string]interface{}{"condition": arrSlsMasterFn}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}
	if len(arrSlsMaster) >= 10 {
		return 0, "", app.MsgStruct{Msg: "can_only_have_:0_active_order_running_per_each_strategy", Params: map[string]string{"0": "10"}}
	}

	// validate prd_group_type
	arrPrdGroupTypeFn := make([]models.WhereCondFn, 0)
	arrPrdGroupTypeFn = append(arrPrdGroupTypeFn,
		models.WhereCondFn{Condition: " prd_group_type.code = ? ", CondValue: prdGroup},
		models.WhereCondFn{Condition: " prd_group_type.status = ? ", CondValue: "A"},
	)
	arrPrdGroupType, err := models.GetPrdGroupTypeFn(arrPrdGroupTypeFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():GetPrdGroupTypeFn():1", map[string]interface{}{"condition": arrPrdGroupTypeFn}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}
	if len(arrPrdGroupType) <= 0 {
		return 0, "", app.MsgStruct{Msg: "invalid_contract_group_type"}
	}

	action = arrPrdGroupType[0].Code
	bnsAction = arrPrdGroupType[0].Code
	docType = arrPrdGroupType[0].DocType

	// validate if amount is positive
	if amount <= 0 {
		return 0, "", app.MsgStruct{Msg: "please_enter_valid_amount"}
	}

	// get purchase contract setting
	if arrPrdMaster[0].PrdGroupSetting == "" {
		base.LogErrorLog("tradingService:AddMemberAutoTrading()", "product_group_setting_not_found", "", true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}

	arrPrdGroupTypeSetup, errMsg := product_service.GetPrdGroupTypeSetup(arrPrdMaster[0].PrdGroupSetting)
	if errMsg != "" {
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}

	var (
		keyinMin        = arrPrdGroupTypeSetup.KeyinMin
		keyinMultipleOf = arrPrdGroupTypeSetup.KeyinMultipleOf
	)

	if prdCode == "CIFRA" {
		keyinMin = 2000
	}

	// amount must be more than or equal to keyinMin
	if amount < keyinMin {
		return 0, "", app.MsgStruct{Msg: "minimum_purchase_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimal(keyinMin, 2, ".", ",")}}
	}

	// amount must be multiple of keyinMultipleOf
	if !helpers.IsMultipleOf(amount, keyinMultipleOf) {
		return 0, "", app.MsgStruct{Msg: "purchase_amount_must_be_multiple_of_:0", Params: map[string]string{"0": helpers.CutOffDecimal(keyinMultipleOf, 2, ".", ",")}}
	}

	// validate crypto_pair, CFRA-AI crypto pair will be automatically picked by system algorithm
	if prdCode != "CFRA" || settingType != "AI" {
		var cryptoPairValid = 0
		var arrCryptoPair = GetCryptoPair(memberID, platform, prdCode)
		for _, arrCryptoPairV := range arrCryptoPair {
			if arrCryptoPairV.Code == cryptoPair {
				cryptoPairValid = 1
			}
		}

		if cryptoPairValid != 1 {
			return 0, "", app.MsgStruct{Msg: "invalid_crypto_pair"}
		}
	}

	// validate available wallet limit
	var availableWalletLimit = GetCurrentMemberAvailableWalletLimit(memberID, prdCode)
	if amount > availableWalletLimit {
		return 0, "", app.MsgStruct{Msg: "cannot_exceed_available_wallet_limit_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(availableWalletLimit, 0, ".", ",", true)}}
	}

	// binance/kucoin must have sufficient balance to afford the trading amount - eddie
	var memberPlatformBalanceParam = GetMemberStrategyBalanceStruct{
		MemberID: memberID,
		Platform: input.Platform,
	}

	if prdCode == "CFRA" || prdCode == "CIFRA" || prdCode == "MTD" { // get platform future balance
		var memberPlatformBalance, errMsg = memberPlatformBalanceParam.GetMemberStrategyFuturesBalancev1()
		if errMsg != "" {
			return 0, "", app.MsgStruct{Msg: errMsg}
		}
		var binanceBalance = memberPlatformBalance.Balance
		if amount > binanceBalance {
			return 0, "", app.MsgStruct{Msg: "insufficient balance "}
		}
	} else { // get platform spot balance
		var memberPlatformBalance, errMsg = memberPlatformBalanceParam.GetMemberStrategyBalancev1()
		if errMsg != "" {
			return 0, "", app.MsgStruct{Msg: errMsg}
		}
		var binanceBalance = memberPlatformBalance.Balance
		if amount > binanceBalance {
			return 0, "", app.MsgStruct{Msg: "insufficient balance "}
		}
	}

	// validate crypto usdt value
	var marketPrice = 0.00
	if platform != "KC" {
		arrBinancePrice, err := GetBinanceCryptoPrice(cryptoPair)
		if err != nil {
			base.LogErrorLog("tradingService:AddMemberAutoTrading:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)

			// get market price from bot_premium_index instead
			arrBotPremiumIndexFn := make([]models.WhereCondFn, 0)
			arrBotPremiumIndexFn = append(arrBotPremiumIndexFn,
				models.WhereCondFn{Condition: "bot_premium_index.symbol = ?", CondValue: cryptoPair},
				models.WhereCondFn{Condition: "bot_premium_index.b_latest = ?", CondValue: 1},
			)
			arrBotPremiumIndex, err := models.GetBotPremiumIndexFn(arrBotPremiumIndexFn, "", false)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:GetBotPremiumIndexFn():2", map[string]interface{}{"condition": arrBotPremiumIndexFn}, err.Error(), true)
				return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
			}
			if len(arrBotPremiumIndex) <= 0 {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:GetBotPremiumIndexFn():2", map[string]interface{}{"condition": arrBotPremiumIndexFn}, "attemp_to_get_market_price_from_previous_data_fail", true)
				return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
			}

			marketPrice = arrBotPremiumIndex[0].MarkPrice
		} else {
			marketPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:ParseFloat():2", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
				return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
			}
		}
	} else {
		arrKucoinPrice, err := GetKucoinCryptoPrice(cryptoPair)
		if err != nil || arrKucoinPrice.Code != "200000" {
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
			} else {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair, "response": arrKucoinPrice}, arrKucoinPrice.Msg, true)
			}

			// return 0, "", app.MsgStruct{Msg: "something_went_wrong"}

			// get market price from bot_premium_index_kc instead
			arrBotPremiumIndexFn := make([]models.WhereCondFn, 0)
			arrBotPremiumIndexFn = append(arrBotPremiumIndexFn,
				models.WhereCondFn{Condition: "bot_premium_index_kc.symbol = ?", CondValue: cryptoPair},
				models.WhereCondFn{Condition: "bot_premium_index_kc.b_latest = ?", CondValue: 1},
			)
			arrBotPremiumIndex, err := models.GetBotPremiumIndexKcFn(arrBotPremiumIndexFn, "", false)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:GetBotPremiumIndexKcFn():1", map[string]interface{}{"condition": arrBotPremiumIndexFn}, err.Error(), true)
				return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
			}
			if len(arrBotPremiumIndex) <= 0 {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:GetBotPremiumIndexKcFn():1", map[string]interface{}{"condition": arrBotPremiumIndexFn}, "attemp_to_get_market_price_from_previous_data_fail", true)
				return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
			}

			marketPrice = arrBotPremiumIndex[0].MarkPrice
		} else {
			// marketPrice = arrKucoinPrice.Data.Price

			// kucoin api changed return data type of price from string to float
			marketPrice, err = strconv.ParseFloat(arrKucoinPrice.Data.Price, 64)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTrading:ParseFloat():1", map[string]interface{}{"value": arrKucoinPrice.Data.Price}, err.Error(), true)
				return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
			}
		}
	}

	// if platform == "KC" && (prdCode == "CFRA" || prdCode == "CIFRA" || prdCode == "MTD") { // for kucoin need
	// 	// minOrderAmount need to calculate from data grab in kucoin
	// 	minTradeAmount, errMsg := GetKucoinMinOrderAmount(cryptoPair)
	// 	if errMsg != "" {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTrading:GetKucoinMinOrderAmount()", map[string]interface{}{"cryptoPair": cryptoPair}, errMsg, true)
	// 		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	// 	}

	// 	if amount < minTradeAmount {
	// 		return 0, "", app.MsgStruct{Msg: "min_trade_amount_is_:0", Params: map[string]string{"0": fmt.Sprintf("%.3f", minTradeAmount)}}
	// 	}
	// } else {
	// 	if float.Mul(amount, marketPrice) < 10 {
	// 		return 0, "", app.MsgStruct{Msg: "crypto_pair_usdt_value_is_too_low_please_try_to_increase_your_trading_amount"}
	// 	}
	// }

	if prdCode != "CFRA" && float.Mul(amount, marketPrice) < 10 {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():crypto_pair_usdt_value_is_too_low_please_try_to_increase_your_trading_amount():1", map[string]interface{}{"platform": platform, "cryptoPair": cryptoPair, "amount": amount, "marketPrice": marketPrice}, "", false)
		return 0, "", app.MsgStruct{Msg: "crypto_pair_usdt_value_is_too_low_please_try_to_increase_your_trading_amount"}
	}

	// get batch_no
	db := models.GetDB()
	batchNo, err = models.GetRunningDocNo(batchDocType, db) //get batch doc no
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():GetRunningDocNo():1", map[string]interface{}{"docType": batchDocType}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}
	err = models.UpdateRunningDocNo(batchDocType, db) //update batch doc no
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():UpdateRunningDocNo():1", map[string]interface{}{"docType": batchDocType}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}

	// get doc_no
	docNo, err = models.GetRunningDocNo(docType, db) //get contract doc no
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():GetRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}
	err = models.UpdateRunningDocNo(docType, db) //update contract doc no
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():UpdateRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}

	// get sponsor_id
	arrEntMemberTreeSponsorFn := make([]models.WhereCondFn, 0)
	arrEntMemberTreeSponsorFn = append(arrEntMemberTreeSponsorFn,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrEntMemberTreeSponsor, err := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberTreeSponsorFn, false)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():GetEntMemberEntMemberTreeSponsorFn():1", map[string]interface{}{"condition": arrEntMemberTreeSponsorFn}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}
	sponsorID = arrEntMemberTreeSponsor.SponsorID

	// save to sls_master
	var addSlsMasterParams = models.AddSlsMasterStruct{
		CountryID:    1,
		CompanyID:    1,
		MemberID:     memberID,
		SponsorID:    sponsorID,
		Status:       "AP",
		Action:       action,
		TotUnit:      amount,
		PriceRate:    unitPrice,
		PrdMasterID:  prdMasterID,
		BatchNo:      batchNo,
		DocType:      docType,
		DocNo:        docNo,
		DocDate:      curDate,
		BnsBatch:     curDate,
		BnsAction:    bnsAction,
		TotalAmount:  amount,
		SubTotal:     amount,
		TotalPv:      amount,
		TotalBv:      amount,
		TotalSv:      0.00,
		TotalNv:      0.00,
		TokenRate:    marketPrice,
		ExchangeRate: 1,
		CurrencyCode: prdCurrencyCode,
		GrpType:      "0",
		CreatedAt:    curDateTime,
		CreatedBy:    fmt.Sprint(memberID),
		ApprovableAt: curDateTime,
		ApprovedAt:   time.Now(),
		ApprovedBy:   strconv.Itoa(memberID),
		ExpiredAt:    expiredAt,
	}

	slsMaster, err := models.AddSlsMaster(tx, addSlsMasterParams)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():AddSlsMaster():1", map[string]interface{}{"param": addSlsMasterParams}, err.Error(), true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}

	// add to sls_master_bot_log
	var addSlsMasterBotLog = models.SlsMasterBotLog{
		MemberID:   memberID,
		DocNo:      docNo,
		Status:     "A",
		RemarkType: "S",
		Remark:     strings.Replace(cryptoPair, "USDT", "/USDT", 1) + " #*order_created_under_strategy*# #*" + arrPrdMaster[0].Name + "*# - #*trading_amount*#: " + helpers.CutOffDecimalv2(amount, 0, ".", ",", true) + ", #*doc_no*#: " + docNo,
		CreatedAt:  time.Now(),
		CreatedBy:  "AUTO",
	}

	_, err = models.AddSlsMasterBotLog(tx, addSlsMasterBotLog)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTrading():AddSlsMasterBotLog():1", err.Error(), map[string]interface{}{"param": addSlsMasterBotLog}, true)
		return 0, "", app.MsgStruct{Msg: "something_went_wrong"}
	}

	return slsMaster.ID, docNo, app.MsgStruct{Msg: ""}
}

func GetCurrentMemberAvailableWalletLimit(memberID int, strategyCode string) float64 {
	var (
		walletLimitSet       float64 = 0
		spentLimit           float64 = 0
		availableWalletLimit float64 = 0
		module                       = ""
	)

	// get module
	if strategyCode == "CFRA" || strategyCode == "CIFRA" || strategyCode == "MTD" {
		module = "FUTURE"
	} else if strategyCode == "SGT" || strategyCode == "MT" {
		module = "SPOT"
	} else {
		base.LogErrorLog("tradingService:GetCurrentMemberAvailableWalletLimit():GetEntMemberTradingWalletLimit():1", "strategy_code_does_not_have_matching_module", map[string]interface{}{"strategyCode": strategyCode}, true)
		return 0
	}

	// take wallet limit amount set
	arrEntMemberTradingWalletLimitFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.module = ?", CondValue: module},
		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingWalletLimit, err := models.GetEntMemberTradingWalletLimit(arrEntMemberTradingWalletLimitFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetCurrentMemberAvailableWalletLimit():GetEntMemberTradingWalletLimit():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingWalletLimitFn}, true)
		return 0
	}

	if len(arrEntMemberTradingWalletLimit) > 0 {
		walletLimitSet = arrEntMemberTradingWalletLimit[0].TotalAmount
	}

	if strategyCode == "CFRA" || strategyCode == "CIFRA" {
		// take wallet limit amount set
		arrEntMemberTradingWalletLimitFn := make([]models.WhereCondFn, 0)
		arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
			models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.member_id = ?", CondValue: memberID},
			models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.module = ?", CondValue: "SPOT"},
			models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.status = ?", CondValue: "A"},
		)
		arrEntMemberTradingWalletLimit, err := models.GetEntMemberTradingWalletLimit(arrEntMemberTradingWalletLimitFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetCurrentMemberAvailableWalletLimit():GetEntMemberTradingWalletLimit():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingWalletLimitFn}, true)
			return 0
		}

		if len(arrEntMemberTradingWalletLimit) > 0 {
			walletLimitSet += arrEntMemberTradingWalletLimit[0].TotalAmount
		}
	}

	memberCurrentApi := GetMemberCurrentAPI(memberID)
	if memberCurrentApi.PlatformCode == "" {
		base.LogErrorLog("tradingService:GetCurrentMemberAvailableWalletLimit():GetMemberCurrentAPI():1", map[string]interface{}{"condition": memberCurrentApi}, "api_not_setup_before", true)
		return 0
	}

	// take spent wallet limit
	totalSalesFn := make([]models.WhereCondFn, 0)
	totalSalesFn = append(totalSalesFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "BOT"},
		models.WhereCondFn{Condition: "sls_master_bot_setting.platform = ?", CondValue: memberCurrentApi.PlatformCode},
		models.WhereCondFn{Condition: "sls_master.ref_no is null AND 1=?", CondValue: "1"},
		// models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"}, // include liquitationed record
	)
	if strategyCode == "CFRA" || strategyCode == "CIFRA" || strategyCode == "MTD" {
		totalSalesFn = append(totalSalesFn,
			models.WhereCondFn{Condition: "prd_master.code IN(?,'CIFRA','MTD')", CondValue: "CFRA"},
		)
	} else if strategyCode == "SGT" || strategyCode == "MT" {
		totalSalesFn = append(totalSalesFn,
			models.WhereCondFn{Condition: "prd_master.code IN(?,'MT')", CondValue: "SGT"},
		)
	}

	totalSales, err := models.GetTotalSalesAmount(totalSalesFn, false)
	if err != nil {
		base.LogErrorLog("tradingService:GetCurrentMemberAvailableWalletLimit():GetTotalSalesAmount():1", map[string]interface{}{"condition": totalSalesFn}, err.Error(), true)
		return 0
	}
	if totalSales.TotalAmount > 0 {
		spentLimit = totalSales.TotalAmount
	}

	availableWalletLimit = walletLimitSet - spentLimit
	if availableWalletLimit < 0 {
		availableWalletLimit = 0
	}

	return availableWalletLimit
}

// MemberAutoTradingCFRA struct
type MemberAutoTradingCFRA struct {
	MemberID   int
	Type       string
	Amount     float64
	CryptoPair string
	LangCode   string
}

func AddMemberAutoTradingCFRA(tx *gorm.DB, input MemberAutoTradingCFRA) app.MsgStruct {
	var (
		strategyCode           = "CFRA"
		memberID     int       = input.MemberID
		settingType  string    = input.Type
		cryptoPair   string    = input.CryptoPair
		amount       float64   = input.Amount
		curDateTime  time.Time = base.GetCurrentDateTimeT()
		langCode     string    = input.LangCode
		errMsg       string    = ""
	)

	var (
		memberCurrentAPI = GetMemberCurrentAPI(memberID)
		platform         = memberCurrentAPI.PlatformCode
		apiKey           = ""
		secret           = ""
		passphrase       = ""
		apiKey2          = ""
		secret2          = ""
		passphrase2      = ""
	)
	if platform == "" {
		return app.MsgStruct{Msg: "please_setup_api_to_proceed"}
	}

	for _, apiDetails := range memberCurrentAPI.ApiDetails {
		if apiDetails.Module == "SPOT" {
			apiKey = apiDetails.ApiKey
			secret = apiDetails.ApiSecret
			passphrase = apiDetails.ApiPassphrase
		} else if apiDetails.Module == "FUTURE" {
			apiKey2 = apiDetails.ApiKey
			secret2 = apiDetails.ApiSecret
			passphrase2 = apiDetails.ApiPassphrase
		}
	}

	if settingType == "AI" {
		cryptoPair = ""

		// get system auto picked crypto_pair
		cryptoPair, errMsg = AutoCryptoPairPicker(memberID, platform)
		if errMsg != "" {
			base.LogErrorLog("tradingService:AddMemberAutoTradingCFRA:AutoCryptoPairPicker()", map[string]interface{}{"memberID": memberID}, errMsg, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	}

	// basic add auto trading flow which include validation + insert to sls_master
	var addMemberAutoTradingParam = MemberAutoTrading{
		MemberID:   memberID,
		Platform:   platform,
		PrdCode:    strategyCode,
		Type:       settingType,
		Amount:     amount,
		CryptoPair: cryptoPair,
		LangCode:   langCode,
	}

	slsMasterID, docNo, msgStruct := AddMemberAutoTrading(tx, addMemberAutoTradingParam)
	if msgStruct.Msg != "" {
		return msgStruct
	}

	// 费率ai版和专业版：增加判断（币对规则），交易金额/2（如：交易金额100/2，现货50，合约50），是否满足交易所合约规则的最低买入额和现货最低11U
	var (
		halfOfTradingAmount = float.Div(amount, 2)
		minOrderAmount      = 0.00
	)

	if platform == "KC" {
		// minOrderAmount need to calculate from data grab in kucoin
		minOrderAmount, errMsg = GetKucoinMinOrderAmount(cryptoPair)
		if errMsg != "" {
			base.LogErrorLog("tradingService:AddMemberAutoTradingCFRA:GetKucoinMinOrderAmount()", map[string]interface{}{"cryptoPair": cryptoPair}, errMsg, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	} else if platform == "BN" {
		// minOrderAmount need to calculate from data grab in binance
		minOrderAmount, errMsg = GetBinanceMinOrderAmount(cryptoPair)
		if errMsg != "" {
			base.LogErrorLog("tradingService:AddMemberAutoTradingCFRA:GetBinanceMinOrderAmount()", map[string]interface{}{"cryptoPair": cryptoPair}, errMsg, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	}

	if halfOfTradingAmount < minOrderAmount {
		return app.MsgStruct{Msg: "minimum_contracts_order_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(float.Mul(minOrderAmount, 2), 8, ".", ",", true)}}
	}

	if halfOfTradingAmount < 11 {
		return app.MsgStruct{Msg: "minimum_spot_order_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(float.Mul(11, 2), 8, ".", ",", true)}}
	}

	// insert to sls_master_bot_setting
	var setting string = "{}"
	var addSlsMasterBotSettingParams = models.AddSlsMasterBotSettingStruct{
		SlsMasterID: slsMasterID,
		Platform:    platform,
		SettingType: settingType,
		CryptoPair:  cryptoPair,
		Setting:     setting,
		CreatedAt:   curDateTime,
		CreatedBy:   strconv.Itoa(memberID),
	}

	_, err := models.AddSlsMasterBotSetting(tx, addSlsMasterBotSettingParams)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingCFRA():AddSlsMasterBotSetting():1", map[string]interface{}{"param": addSlsMasterBotSettingParams}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// call china api to add bot - always call api at last step
	var memberAutoTradingApiParam = MemberAutoTradingApi{
		DocNo:      docNo,
		Platform:   platform,
		AppID:      apiKey,
		Secret:     secret,
		AppPwd:     passphrase,
		AppID2:     apiKey2,
		Secret2:    secret2,
		AppPwd2:    passphrase2,
		Strategy:   strategyCode,
		CryptoPair: cryptoPair,
		Amount:     amount,
		Setting:    setting,
	}
	errMsg = PostMemberAutoTradingApi(memberAutoTradingApiParam)
	if errMsg != "" {
		base.LogErrorLog("tradingService:AddMemberAutoTradingCFRA():PostMemberAutoTradingApi():1", map[string]interface{}{"param": memberAutoTradingApiParam}, errMsg, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}
}

func AutoCryptoPairPicker(memberID int, platform string) (string, string) {
	var topCyrptoPairRank = 30
	// get top 30 from bot_premium_index where b_latest = 1
	arrBotPremiumIndex, err := models.GetBotPremiumIndexRankedFundingRate(topCyrptoPairRank, platform, false)
	if err != nil {
		base.LogErrorLog("tradingService:AutoCryptoPairPicker():GetBotPremiumIndexRankedFundingRate():1", map[string]interface{}{"rankNo": topCyrptoPairRank, "platform": platform}, err.Error(), true)
		return "", "something_went_wrong"
	}

	if len(arrBotPremiumIndex) <= 0 {
		base.LogErrorLog("tradingService:AutoCryptoPairPicker():GetBotPremiumIndexRankedFundingRate():1", map[string]interface{}{"rankNo": topCyrptoPairRank, "platform": platform, "rst": arrBotPremiumIndex}, "empty_array_arrBotPremiumIndex", true)
		return "", "something_went_wrong"
	}

	var cryptoPair = ""
	// loop each crypto pair
	for _, arrBotPremiumIndexV := range arrBotPremiumIndex {
		var (
			curCryptoPair      = arrBotPremiumIndexV.Symbol
			limit              = 300
			totalRecord        = 0.00
			totalPositive      = 0.00
			positivePercentage = 0.00
		)

		// validate if crypto pair is used in previous order
		arrSlsMasterBotSettingFn := make([]models.WhereCondFn, 0)
		arrSlsMasterBotSettingFn = append(arrSlsMasterBotSettingFn,
			models.WhereCondFn{Condition: " sls_master_bot_setting.crypto_pair = ?", CondValue: curCryptoPair},
			models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memberID},
			models.WhereCondFn{Condition: " sls_master.status = ?", CondValue: "AP"},
		)
		arrSlsMasterBotSetting, err := models.GetSlsMasterBotSetting(arrSlsMasterBotSettingFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:AutoCryptoPairPicker():GetSlsMasterBotSetting():1", err.Error(), arrSlsMasterBotSettingFn, true)
		}
		if len(arrSlsMasterBotSetting) > 0 {
			continue // jump to next crypto pair
		}

		// get latest 300 history of that crypto pair  /fapi/v1/fundingRate
		arrBotPremiumIndexFundingRateHistory, err := models.GetBotPremiumIndexFundingRateHistory(curCryptoPair, platform, limit, false)
		if err != nil {
			base.LogErrorLog("tradingService:AutoCryptoPairPicker():GetBotPremiumIndexFundingRateHistory():1", map[string]interface{}{"crypto_pair": curCryptoPair, "limit": limit}, err.Error(), true)
			return "", "something_went_wrong"
		}

		totalRecord = float64(len(arrBotPremiumIndexFundingRateHistory))

		if totalRecord <= 0 {
			base.LogErrorLog("tradingService:AutoCryptoPairPicker():GetBotPremiumIndexFundingRateHistory():1", map[string]interface{}{"rst": arrBotPremiumIndexFundingRateHistory}, "empty_array_arrBotPremiumIndex", true)
			return "", "something_went_wrong"
		}

		// calculate and validate if funding rate 90% is positive
		for _, arrBotPremiumIndexFundingRateHistoryV := range arrBotPremiumIndexFundingRateHistory {
			if arrBotPremiumIndexFundingRateHistoryV.LastFundingRate >= 0 {
				totalPositive += 1
			}
		}

		positivePercentage = float.Mul(float.Div(totalPositive, totalRecord), 100)
		// fmt.Println("curCryptoPair:", curCryptoPair, "positivePercentage:", positivePercentage, "totalRecord:", totalRecord, "totalPositive:", totalPositive)
		if positivePercentage > 90 {
			cryptoPair = curCryptoPair
			break // break loop and use this crypto pair
		}
	}

	if cryptoPair == "" {
		base.LogErrorLog("tradingService:AutoCryptoPairPicker()", map[string]interface{}{"memberID": memberID, "cryptoPair": cryptoPair}, "crypto_pair_not_found", true)
		return "", "something_went_wrong"
	}

	return cryptoPair, ""
}

// MemberAutoTradingCIFRA struct
type MemberAutoTradingCIFRA struct {
	MemberID   int
	Type       string
	Amount     float64
	CryptoPair string
	LangCode   string
}

func AddMemberAutoTradingCIFRA(tx *gorm.DB, input MemberAutoTradingCIFRA) app.MsgStruct {
	var (
		strategyCode           = "CIFRA"
		memberID     int       = input.MemberID
		settingType  string    = input.Type
		cryptoPair   string    = input.CryptoPair
		amount       float64   = input.Amount
		curDateTime  time.Time = base.GetCurrentDateTimeT()
		errMsg       string    = ""
		langCode     string    = input.LangCode
	)

	var (
		memberCurrentAPI = GetMemberCurrentAPI(memberID)
		platform         = memberCurrentAPI.PlatformCode
		apiKey           = ""
		secret           = ""
	)
	if platform == "" {
		return app.MsgStruct{Msg: "please_setup_api_to_proceed"}
	}
	if platform == "KC" {
		return app.MsgStruct{Msg: "strategy_is_unavailable_for_chosen_api"}
	}

	apiKey = memberCurrentAPI.ApiDetails[0].ApiKey
	secret = memberCurrentAPI.ApiDetails[0].ApiSecret

	// basic add auto trading flow which include validation + insert to sls_master
	var addMemberAutoTradingParam = MemberAutoTrading{
		MemberID:   memberID,
		Platform:   platform,
		PrdCode:    strategyCode,
		Type:       settingType,
		Amount:     amount,
		CryptoPair: cryptoPair,
		LangCode:   langCode,
	}

	slsMasterID, docNo, msgStruct := AddMemberAutoTrading(tx, addMemberAutoTradingParam)
	if msgStruct.Msg != "" {
		return msgStruct
	}

	// 指数费率：增加判断完整金额必须大于2000U
	if amount < 2000 {
		return app.MsgStruct{Msg: "minimum_order_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(2000, 8, ".", ",", true)}}
	}

	// 指数费率：增加判断（币对规则），交易金额/2，是否满足交易所合约规则最低买入额。
	var (
		halfOfTradingAmount = float.Div(amount, 2)
		minOrderAmount      = 0.00
	)

	if platform == "KC" {
		// minOrderAmount need to calculate from data grab in kucoin
		minOrderAmount, errMsg = GetKucoinMinOrderAmount(cryptoPair)
		if errMsg != "" {
			base.LogErrorLog("tradingService:AddMemberAutoTradingCIFRA:GetKucoinMinOrderAmount()", map[string]interface{}{"cryptoPair": cryptoPair}, errMsg, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	} else if platform == "BN" {
		// minOrderAmount need to calculate from data grab in binance
		minOrderAmount, errMsg = GetBinanceMinOrderAmount(cryptoPair)
		if errMsg != "" {
			base.LogErrorLog("tradingService:AddMemberAutoTradingCIFRA:GetBinanceMinOrderAmount()", map[string]interface{}{"cryptoPair": cryptoPair}, errMsg, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	}

	if halfOfTradingAmount < minOrderAmount {
		return app.MsgStruct{Msg: "minimum_contracts_order_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(float.Mul(minOrderAmount, 2), 8, ".", ",", true)}}
	}

	// insert to sls_master_bot_setting
	var setting string = "{}"
	var addSlsMasterBotSettingParams = models.AddSlsMasterBotSettingStruct{
		SlsMasterID: slsMasterID,
		Platform:    platform,
		SettingType: settingType,
		CryptoPair:  cryptoPair,
		Setting:     setting,
		CreatedAt:   curDateTime,
		CreatedBy:   strconv.Itoa(memberID),
	}

	_, err := models.AddSlsMasterBotSetting(tx, addSlsMasterBotSettingParams)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingCIFRA():AddSlsMasterBotSetting():1", map[string]interface{}{"param": addSlsMasterBotSettingParams}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// call china api to add bot - always call api at last step
	var memberAutoTradingApiParam = MemberAutoTradingApi{
		DocNo:      docNo,
		Platform:   platform,
		AppID:      apiKey,
		Secret:     secret,
		Strategy:   strategyCode,
		CryptoPair: cryptoPair,
		Amount:     amount,
		Setting:    setting,
	}
	errMsg = PostMemberAutoTradingApi(memberAutoTradingApiParam)
	if errMsg != "" {
		base.LogErrorLog("tradingService:AddMemberAutoTradingCIFRA():PostMemberAutoTradingApi():1", map[string]interface{}{"param": memberAutoTradingApiParam}, errMsg, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}
}

// MemberAutoTradingSGT struct
type MemberAutoTradingSGT struct {
	MemberID                 int
	Type                     string
	Amount                   float64
	CryptoPair               string
	UpperPrice               float64 // required if type = PROF
	LowerPrice               float64 // required if type = PROF
	CryptoPricePercentage    float64 // required if type = PROF | mode = GEOMETRIC only
	CalCryptoPricePercentage float64 // required if type = PROF | mode = ARITHMETIC only
	Mode                     string
	LangCode                 string
}

func AddMemberAutoTradingSGT(tx *gorm.DB, input MemberAutoTradingSGT) app.MsgStruct {
	var (
		strategyCode           = "SGT"
		memberID     int       = input.MemberID
		settingType  string    = input.Type
		cryptoPair   string    = input.CryptoPair
		amount       float64   = input.Amount
		curDateTime  time.Time = base.GetCurrentDateTimeT()
		langCode     string    = input.LangCode
	)

	var (
		memberCurrentAPI = GetMemberCurrentAPI(memberID)
		platform         = memberCurrentAPI.PlatformCode
		apiKey           = ""
		secret           = ""
		passphrase       = ""
		apiKey2          = ""
		secret2          = ""
		passphrase2      = ""
	)
	if platform == "" {
		return app.MsgStruct{Msg: "please_setup_api_to_proceed"}
	}

	for _, apiDetails := range memberCurrentAPI.ApiDetails {
		if apiDetails.Module == "SPOT" {
			apiKey = apiDetails.ApiKey
			secret = apiDetails.ApiSecret
			passphrase = apiDetails.ApiPassphrase
		} else if apiDetails.Module == "FUTURE" {
			apiKey2 = apiDetails.ApiKey
			secret2 = apiDetails.ApiSecret
			passphrase2 = apiDetails.ApiPassphrase
		}
	}

	// basic add auto trading flow which include validation + insert to sls_master
	var addMemberAutoTradingParam = MemberAutoTrading{
		MemberID:   memberID,
		Platform:   platform,
		PrdCode:    strategyCode,
		Type:       settingType,
		Amount:     amount,
		CryptoPair: cryptoPair,
		LangCode:   langCode,
	}

	slsMasterID, docNo, msgStruct := AddMemberAutoTrading(tx, addMemberAutoTradingParam)
	if msgStruct.Msg != "" {
		return msgStruct
	}

	// validate setting value
	var (
		lowerPrice               float64
		upperPrice               float64
		gridQuantity             float64
		takerRate                float64
		readPrice                float64
		buyQuantity              float64                                  // only ai need to send to china - "Spot grid edit 220422.docx"
		cryptoPricePercentage    float64 = input.CryptoPricePercentage    // member keyin % for current coin price | GEOMETRIC
		calCryptoPricePercentage float64 = input.CalCryptoPricePercentage // member keyin 1% of current coin price | ARITHMETIC
		mode                     string  = input.Mode
	)

	if settingType == "AI" {
		mode = "ARITHMETIC" // default will be ARITHMETIC mode
	}

	// get taker rate
	memberAutoTradingGridData, errMsg := GetMemberAutoTradingGridData(memberID, cryptoPair, mode, langCode)
	if errMsg != "" {
		base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():GetMemberAutoTradingGridData():1", map[string]interface{}{"memberID": memberID, "cryptoPair": cryptoPair, "mode": mode, "langCode": langCode}, errMsg, true)
		return app.MsgStruct{Msg: errMsg}
	}

	takerRate = memberAutoTradingGridData.TakerRate

	// get read price
	// arrBinancePrice, err := GetBinanceCryptoPrice(cryptoPair)
	// if err != nil {
	// 	base.LogErrorLog("tradingService:AddMemberAutoTradingSGT:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}
	// }

	// readPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
	// if err != nil {
	// 	base.LogErrorLog("tradingService:AddMemberAutoTradingSGT:ParseFloat()", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}
	// }

	readPrice = memberAutoTradingGridData.MarketPrice

	// get default value for AI
	if settingType == "AI" {
		upperPrice = memberAutoTradingGridData.UpperPrice
		lowerPrice = memberAutoTradingGridData.LowerPrice
		gridQuantity = memberAutoTradingGridData.GridQuantity
		cryptoPricePercentage = memberAutoTradingGridData.CryptoPricePercentage
		calCryptoPricePercentage = memberAutoTradingGridData.CalCryptoPricePercentage

		// buy quantity = trading_amount / grid_quantity /latest_price - "Spot grid edit 220422.docx"
		buyQuantity = float.Div(float.Div(amount, gridQuantity), readPrice)

		// to recommend member lowest trading amount for checking reqiurement to be above 10USDT - "Spot grid edit 220422.docx"
		var lowestTotalAmount = float.Mul(lowerPrice, buyQuantity)
		if lowestTotalAmount < 10 {
			return app.MsgStruct{Msg: "lowest_total_amount_must_be_above_:0", Params: map[string]string{"0": "10"}}
		}

		// validate upper_single_grid_profit and lower_single_grid_profit
		// if input.UpperSingleGridProfit == 0 {
		// 	return app.MsgStruct{Msg: "upper_single_grid_profit_is_required"}
		// }
		// if input.LowerSingleGridProfit == 0 {
		// 	return app.MsgStruct{Msg: "lower_single_grid_profit_is_required"}
		// }
		// if input.UpperSingleGridProfit != memberAutoTradingGridData.UpperSingleGridProfit || input.LowerSingleGridProfit != memberAutoTradingGridData.LowerSingleGridProfit {
		// 	base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():GetMemberAutoTradingGridData():1", map[string]interface{}{"inputLowerGrid": input.LowerSingleGridProfit, "calLowerGrid": memberAutoTradingGridData.LowerSingleGridProfit, "inputUpperGrid": input.UpperSingleGridProfit, "calUpperrGrid": memberAutoTradingGridData.UpperSingleGridProfit, "type": "AI"}, "data_is_outdated_please_refresh_to_try_again", true)
		// 	return app.MsgStruct{Msg: "data_is_outdated_please_refresh_to_try_again"}
		// }
	} else if settingType == "PROF" { // validate and use member input
		upperPrice = input.UpperPrice
		lowerPrice = input.LowerPrice

		// arithmetic will receive 1% of current price, where as geometric will receive x% instead - "Spot grid edit 220422.docx"
		if mode == "ARITHMETIC" {
			cryptoPricePercentage = 1
			if !(calCryptoPricePercentage > 0) {
				return app.MsgStruct{Msg: "calculated_crypto_price_percentage_must_be_positive_number"}
			}
		} else if mode == "GEOMETRIC" {
			if !(cryptoPricePercentage > 0) {
				return app.MsgStruct{Msg: "crypto_price_percentage_must_be_positive_number"}
			}

			// validate cryptoPricePercentage can only up to 1 decimal places
			cutCryptoPricePercentageStr := helpers.CutOffDecimalv2(cryptoPricePercentage, 1, ".", "", true)
			cutCryptoPricePercentage, err := strconv.ParseFloat(cutCryptoPricePercentageStr, 64)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():strconv.ParseFloat():1", map[string]interface{}{"string": cutCryptoPricePercentageStr}, err.Error(), true)
				return app.MsgStruct{Msg: "something_went_wrong"}
			}

			if cryptoPricePercentage != cutCryptoPricePercentage {
				return app.MsgStruct{Msg: "crypto_price_percentage_max_:0_decimal_place", Params: map[string]string{"0": "1"}}
			}

			// only up to 1 decimal place, eg:1.2%
			calCryptoPricePercentage = float.Mul(readPrice, float.Div(cryptoPricePercentage, 100))
		}

		// calculate grid quantity -> uppestPrice - lowestPrice / (1% of currennt coin price)
		gridQuantity = float.Div((upperPrice - lowerPrice), calCryptoPricePercentage)

		// round up gridQuantity to whole number
		var err error
		gridQuantity, err = strconv.ParseFloat(helpers.CutOffDecimalv2(gridQuantity, 0, ".", "", true), 64)
		if err != nil {
			base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():strconv.ParseFloat():2", map[string]interface{}{"gridQuantity": gridQuantity}, err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}

		// validate lower price
		if !(lowerPrice > 0) {
			return app.MsgStruct{Msg: "lower_price_must_be_positive_number"}
		}

		// validate upper price
		if upperPrice <= lowerPrice {
			return app.MsgStruct{Msg: "upper_price_must_greater_than_lower_price"}
		}

		// validate grid quantity
		if !(gridQuantity > 1) || gridQuantity != float64(int(gridQuantity)) {
			base.LogErrorLog("tradingService:AddMemberAutoTradingSGT()", map[string]interface{}{"gridQuantity": gridQuantity}, "invalid_grid_quantity.grid_quantity_must_be_a_positive_whole_number_bigger_than_1", true)
			return app.MsgStruct{Msg: "calculated_grid_quantity_must_be_number_bigger_than_1"}
		}

		// validate mode
		if mode == "" {
			return app.MsgStruct{Msg: "mode_is_required"}
		}
		if !helpers.StringInSlice(mode, []string{"ARITHMETIC", "GEOMETRIC"}) {
			return app.MsgStruct{Msg: "invalid_mode"}
		}

		if mode == "ARITHMETIC" {
			// validate upper_single_grid_profit and lower_single_grid_profit
			// if input.UpperSingleGridProfit == 0 {
			// 	return app.MsgStruct{Msg: "upper_single_grid_profit_is_required"}
			// }
			// if input.LowerSingleGridProfit == 0 {
			// 	return app.MsgStruct{Msg: "lower_single_grid_profit_is_required"}
			// }

			var (
				upperSingleGridProfit float64 = 0
				lowerSingleGridProfit float64 = 0
			)

			// calculate upper_single_grid_profit and lower_single_grid_profit
			lowerSingleGridProfit = float.Div(float.Div((upperPrice-lowerPrice), (gridQuantity-1)), upperPrice) - float.Mul(2, float.Div(takerRate, 100))
			lowerSingleGridProfit = float.Mul(lowerSingleGridProfit, 100)
			lowerSingleGridProfit, err := strconv.ParseFloat(helpers.CutOffDecimalv2(lowerSingleGridProfit, 2, ".", "", true), 64)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():ParseFloat():2", map[string]interface{}{"value": lowerSingleGridProfit}, errMsg, true)
				return app.MsgStruct{Msg: errMsg}
			}

			upperSingleGridProfit = float.Div(float.Div((upperPrice-lowerPrice), (gridQuantity-1)), lowerPrice) - float.Mul(2, float.Div(takerRate, 100))
			upperSingleGridProfit = float.Mul(upperSingleGridProfit, 100)
			upperSingleGridProfit, err = strconv.ParseFloat(helpers.CutOffDecimalv2(upperSingleGridProfit, 2, ".", "", true), 64)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():ParseFloat():3", map[string]interface{}{"value": upperSingleGridProfit}, errMsg, true)
				return app.MsgStruct{Msg: errMsg}
			}

			// if !helpers.FloatEquality(input.UpperSingleGridProfit, upperSingleGridProfit) || !helpers.FloatEquality(input.LowerSingleGridProfit, lowerSingleGridProfit) {
			// 	base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():GetMemberAutoTradingGridData():1", map[string]interface{}{"inputLowerGrid": input.LowerSingleGridProfit, "calLowerGrid": lowerSingleGridProfit, "inputUpperGrid": input.UpperSingleGridProfit, "calUpperrGrid": upperSingleGridProfit, "type": "PROF"}, "data_is_outdated_please_refresh_to_try_again", true)
			// 	return app.MsgStruct{Msg: "data_is_outdated_please_refresh_to_try_again"}
			// }

		} else if mode == "GEOMETRIC" {
			// validate single_grid_profit
			// if input.SingleGridProfit == 0 {
			// 	return app.MsgStruct{Msg: "single_grid_profit_is_required"}
			// }

			var singleGridProfit = ((math.Pow(float.Div(upperPrice, lowerPrice), float.Div(1, (gridQuantity-1))) - 1) - float.Mul(2, float.Div(takerRate, 100)))
			singleGridProfit = float.Mul(singleGridProfit, 100)
			singleGridProfit, err := strconv.ParseFloat(helpers.CutOffDecimalv2(singleGridProfit, 2, ".", "", true), 64)
			if err != nil {
				base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():ParseFloat():4", map[string]interface{}{"value": singleGridProfit}, errMsg, true)
				return app.MsgStruct{Msg: errMsg}
			}

			// if !helpers.FloatEquality(input.SingleGridProfit, singleGridProfit) {
			// 	base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():GetMemberAutoTradingGridData():1", map[string]interface{}{"inputSingleGridProfit": input.SingleGridProfit, "calSingleGridProfit": singleGridProfit, "type": "PROF"}, "data_is_outdated_please_refresh_to_try_again", true)
			// 	return app.MsgStruct{Msg: "data_is_outdated_please_refresh_to_try_again"}
			// }
		}
	} else {
		return app.MsgStruct{Msg: "invalid_type"}
	}

	// 网格ai版和专业版：增加判断，每格交易金额必须大于11。大于11可以发送到策略服务，小于11提示不发送策略服务。
	// 2023-03-30 alan 换成必须大于20u
	if float.Div(amount, gridQuantity) < 20 {
		return app.MsgStruct{Msg: "minimum_of_trading_amount_per_grid_is_:0", Params: map[string]string{"0": "20"}}
	}

	// china asked to change takerRate to return figure based on below logic.
	// ARITHMETIC: takerRate = calCryptoPricePercentage/readPrice, GEOMETRIC: takerRate = calCryptoPricePercentage
	if mode == "ARITHMETIC" {
		takerRate = float.Div(calCryptoPricePercentage, readPrice)
	} else if mode == "GEOMETRIC" {
		takerRate = calCryptoPricePercentage
	}

	// insert to sls_master_bot_setting
	var arrSetting = map[string]interface{}{
		"lowerPrice":               lowerPrice,
		"upperPrice":               upperPrice,
		"calCryptoPricePercentage": calCryptoPricePercentage,
		"gridQuantity":             gridQuantity,
		"mode":                     mode,
		"takerRate":                takerRate,
		"readPrice":                readPrice,
	}

	if settingType == "AI" {
		arrSetting["buyQuantity"] = buyQuantity
	}

	c, err := json.Marshal(arrSetting)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():Marshal():1", map[string]interface{}{"param": arrSetting}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	var setting string = string(c)

	arrSetting2 := arrSetting
	if settingType == "PROF" {
		arrSetting2["cryptoPricePercentage"] = cryptoPricePercentage
	}

	c2, err := json.Marshal(arrSetting2)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():Marshal():2", map[string]interface{}{"param": arrSetting2}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	var setting2 string = string(c2)
	var addSlsMasterBotSettingParams = models.AddSlsMasterBotSettingStruct{
		SlsMasterID: slsMasterID,
		Platform:    platform,
		SettingType: settingType,
		CryptoPair:  cryptoPair,
		Setting:     setting2,
		CreatedAt:   curDateTime,
		CreatedBy:   strconv.Itoa(memberID),
	}

	_, err = models.AddSlsMasterBotSetting(tx, addSlsMasterBotSettingParams)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():AddSlsMasterBotSetting():1", map[string]interface{}{"param": addSlsMasterBotSettingParams}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// call china api to add bot - always call api at last step
	var memberAutoTradingApiParam = MemberAutoTradingApi{
		DocNo:      docNo,
		Platform:   platform,
		AppID:      apiKey,
		Secret:     secret,
		AppPwd:     passphrase,
		AppID2:     apiKey2,
		Secret2:    secret2,
		AppPwd2:    passphrase2,
		Strategy:   strategyCode,
		CryptoPair: cryptoPair,
		Amount:     amount,
		Setting:    setting,
	}
	errMsg = PostMemberAutoTradingApi(memberAutoTradingApiParam)
	if errMsg != "" {
		base.LogErrorLog("tradingService:AddMemberAutoTradingSGT():PostMemberAutoTradingApi():1", map[string]interface{}{"param": memberAutoTradingApiParam}, errMsg, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}
}

// MemberAutoTradingMT struct
type MemberAutoTradingMT struct {
	MemberID           int
	Type               string
	Amount             float64
	CryptoPair         string
	FirstOrderAmount   float64
	FirstOrderPrice    float64
	PriceScale         float64
	TakeProfitCallback float64
	TakeProfit         float64
	// AddShares          float64
	LangCode string
}

func AddMemberAutoTradingMT(tx *gorm.DB, input MemberAutoTradingMT) app.MsgStruct {
	var (
		strategyCode           = "MT"
		memberID     int       = input.MemberID
		settingType  string    = input.Type
		cryptoPair   string    = input.CryptoPair
		amount       float64   = input.Amount
		curDateTime  time.Time = base.GetCurrentDateTimeT()
		langCode     string    = input.LangCode
	)

	var (
		memberCurrentAPI = GetMemberCurrentAPI(memberID)
		platform         = memberCurrentAPI.PlatformCode
		apiKey           = ""
		secret           = ""
		passphrase       = ""
		apiKey2          = ""
		secret2          = ""
		passphrase2      = ""
	)
	if platform == "" {
		return app.MsgStruct{Msg: "please_setup_api_to_proceed"}
	}

	for _, apiDetails := range memberCurrentAPI.ApiDetails {
		if apiDetails.Module == "SPOT" {
			apiKey = apiDetails.ApiKey
			secret = apiDetails.ApiSecret
			passphrase = apiDetails.ApiPassphrase
		} else if apiDetails.Module == "FUTURE" {
			apiKey2 = apiDetails.ApiKey
			secret2 = apiDetails.ApiSecret
			passphrase2 = apiDetails.ApiPassphrase
		}
	}

	// basic add auto trading flow which include validation + insert to sls_master
	var addMemberAutoTradingParam = MemberAutoTrading{
		MemberID:   memberID,
		Platform:   platform,
		PrdCode:    strategyCode,
		Type:       settingType,
		Amount:     amount,
		CryptoPair: cryptoPair,
		LangCode:   langCode,
	}

	slsMasterID, docNo, msgStruct := AddMemberAutoTrading(tx, addMemberAutoTradingParam)
	if msgStruct.Msg != "" {
		return msgStruct
	}

	var (
		firstOrderPrice     float64
		firstOrderAmount    float64 = input.FirstOrderAmount
		minFirstOrderAmount float64 = 11
		priceScale          float64
		takeProfitCallback  float64
		takeProfit          float64 // take profit callback
		safetyOrders        float64
		addShares           float64 = 2 // add shares fixed at 2
		circularTrans       int         // 1/0
		errMsg              string
	)

	// 马丁ai版和专业版：增加判断（币对规则），首单金额是否大于11U。
	if firstOrderAmount < minFirstOrderAmount {
		return app.MsgStruct{Msg: "minimum_first_order_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(minFirstOrderAmount, 8, ".", ",", true)}}
	}

	if firstOrderAmount > amount {
		return app.MsgStruct{Msg: "first_order_amount_cannot_more_than_trading_amount"}
	}

	// get first_order_price, take_profit_adjust and add_shares from sys_trading_crypto_pair_setup
	arrSysTradingCryptoPairSetupFn := make([]models.WhereCondFn, 0)
	arrSysTradingCryptoPairSetupFn = append(arrSysTradingCryptoPairSetupFn,
		models.WhereCondFn{Condition: "code = ?", CondValue: cryptoPair},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)
	arrSysTradingCryptoPairSetup, err := models.GetSysTradingCryptoPairSetupByPlatformFn(arrSysTradingCryptoPairSetupFn, "", platform, false)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMT:GetSysTradingCryptoPairSetupByPlatformFn()", map[string]interface{}{"condition": arrSysTradingCryptoPairSetupFn, "platform": platform}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}
	if len(arrSysTradingCryptoPairSetup) <= 0 {
		return app.MsgStruct{Msg: "invalid_crypto_pair"}
	}

	circularTrans = arrSysTradingCryptoPairSetup[0].CircularTrans

	// get first order price - first order hardcode after discussion on 2022-08-08 meeting
	// if platform != "KC" {
	// 	arrBinancePrice, err := GetBinanceCryptoPrice(cryptoPair)
	// 	if err != nil {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTradingMT:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}

	// 	firstOrderPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
	// 	if err != nil {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTradingMT:ParseFloat()", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}
	// } else {
	// 	arrKucoinPrice, err := GetKucoinCryptoPrice(cryptoPair)
	// 	if err != nil {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTradingMT:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}

	// 	if arrKucoinPrice.Code != "200000" {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTradingMT:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair, "response": arrKucoinPrice}, arrKucoinPrice.Msg, true)
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}

	// 	firstOrderPrice, err = strconv.ParseFloat(arrKucoinPrice.Data.Price, 64)
	// 	if err != nil {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTradingMT:ParseFloat()", map[string]interface{}{"value": arrKucoinPrice.Data.Price}, err.Error(), true)
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}
	// }

	// firstOrderPrice = 100000 // after discussion on 2022-08-08 meeting
	firstOrderPrice = 200000 // after discussion on 2023-11-27 meeting

	if settingType == "AI" { // use default value
		priceScale = arrSysTradingCryptoPairSetup[0].PriceScale
		takeProfitCallback = arrSysTradingCryptoPairSetup[0].TakeProfitRatio
		takeProfit = arrSysTradingCryptoPairSetup[0].TakeProfitAdjustment
		// addShares = float64(arrSysTradingCryptoPairSetup[0].AddShares)

		// get safety orders
		safetyOrders = GetSafetyOrders(firstOrderAmount, amount, addShares)

	} else if settingType == "PROF" { // validate setting value
		// first order price
		if input.FirstOrderPrice == 0 {
			return app.MsgStruct{Msg: "first_order_price_is_required"}
		}

		// price scale
		if input.PriceScale == 0 {
			return app.MsgStruct{Msg: "price_scale_is_required"}
		}

		// // subsequent price scale
		// if !(input.SubsequentPriceScale > 0) || input.SubsequentPriceScale != float64(int(input.SubsequentPriceScale)) {
		// 	return app.MsgStruct{Msg: "subsequent_price_scale_must_be_a_positive_whole_number"}
		// }

		// // subsequent add shares
		// if !(input.SubsequentAddShares > 0) || input.SubsequentAddShares != float64(int(input.SubsequentAddShares)) {
		// 	return app.MsgStruct{Msg: "subsequent_add_shares_must_be_a_positive_whole_number"}
		// }

		// take profit callback
		if input.TakeProfitCallback == 0 {
			return app.MsgStruct{Msg: "take_profit_callback_is_required"}
		}

		// take profit
		if input.TakeProfit == 0 {
			return app.MsgStruct{Msg: "take_profit_is_required"}
		}

		// safety orders must be integer
		// if input.SafetyOrders == 0 {
		// 	return app.MsgStruct{Msg: "safety_orders_is_required"}
		// }

		// if !(input.SafetyOrders > 0) || input.SafetyOrders != float64(int(input.SafetyOrders)) {
		// 	return app.MsgStruct{Msg: "safety_orders_must_be_a_positive_whole_number"}
		// }

		// safety orders must be integer
		// if input.AddShares == 0 {
		// 	return app.MsgStruct{Msg: "add_shares_is_required"}
		// }

		// if !(input.AddShares > 0) || input.AddShares != float64(int(input.AddShares)) {
		// 	return app.MsgStruct{Msg: "add_shares_must_be_a_positive_whole_number"}
		// }

		priceScale = input.PriceScale
		firstOrderPrice = input.FirstOrderPrice
		takeProfitCallback = input.TakeProfitCallback
		takeProfit = input.TakeProfit
		// addShares = input.AddShares

		// get safety orders
		safetyOrders = GetSafetyOrders(firstOrderAmount, amount, addShares)
	} else {
		return app.MsgStruct{Msg: "invalid_type"}
	}

	if safetyOrders <= 0 {
		return app.MsgStruct{Msg: "invalid_safety_orders"}
	}

	// convert price scale from 1% to 0.01
	// priceScale = float.Div(priceScale, 100)

	// insert to sls_master_bot_log
	var addSlsMasterBotLog = models.SlsMasterBotLog{
		MemberID:   memberID,
		DocNo:      docNo,
		Status:     "A",
		RemarkType: "S",
		Remark:     strings.Replace(cryptoPair, "USDT", "/USDT", -1) + " #*first_order_price*#:" + helpers.CutOffDecimalv2(firstOrderPrice, 2, ".", ",", true) + " #*doc_no*#:" + docNo,
		CreatedAt:  time.Now(),
		CreatedBy:  "AUTO",
	}
	_, err = models.AddSlsMasterBotLog(tx, addSlsMasterBotLog)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMT():AddSlsMasterBotLog():1", err.Error(), map[string]interface{}{"param": addSlsMasterBotLog}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// insert to sls_master_bot_setting
	var arrSetting = map[string]interface{}{
		"firstOrderPrice":  firstOrderPrice,
		"firstOrderAmount": firstOrderAmount,
		// "priceScale":           fmt.Sprintf("%f%%", priceScale),
		// "takeProfitRatio":      fmt.Sprintf("%f%%", takeProfitCallback),
		// "takeProfitAdjust":     fmt.Sprintf("%f%%", takeProfit),
		"priceScale":           float.Div(priceScale, 100),
		"takeProfitRatio":      float.Div(takeProfitCallback, 100),
		"takeProfitAdjust":     float.Div(takeProfit, 100),
		"safetyOrders":         safetyOrders,
		"addShares":            addShares,
		"circularTransactions": circularTrans, // 1/0
	}

	c, err := json.Marshal(arrSetting)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMT():Marshal():1", map[string]interface{}{"param": arrSetting}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	var setting string = string(c)
	var addSlsMasterBotSettingParams = models.AddSlsMasterBotSettingStruct{
		SlsMasterID: slsMasterID,
		Platform:    platform,
		SettingType: settingType,
		CryptoPair:  cryptoPair,
		Setting:     setting,
		CreatedAt:   curDateTime,
		CreatedBy:   strconv.Itoa(memberID),
	}

	_, err = models.AddSlsMasterBotSetting(tx, addSlsMasterBotSettingParams)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMT():AddSlsMasterBotSetting():1", map[string]interface{}{"param": addSlsMasterBotSettingParams}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// call china api to add bot - always call api at last step
	var memberAutoTradingApiParam = MemberAutoTradingApi{
		DocNo:      docNo,
		Platform:   platform,
		AppID:      apiKey,
		Secret:     secret,
		AppPwd:     passphrase,
		AppID2:     apiKey2,
		Secret2:    secret2,
		AppPwd2:    passphrase2,
		Strategy:   strategyCode,
		CryptoPair: cryptoPair,
		Amount:     amount,
		Setting:    setting,
	}
	errMsg = PostMemberAutoTradingApi(memberAutoTradingApiParam)
	if errMsg != "" {
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}
}

// MemberAutoTradingMTD struct
type MemberAutoTradingMTD struct {
	MemberID           int
	Type               string
	Amount             float64
	CryptoPair         string
	FirstOrderAmount   float64
	FirstOrderPrice    float64
	PriceScale         float64
	TakeProfitCallback float64
	TakeProfit         float64
	// AddShares          float64
	LangCode string
}

func AddMemberAutoTradingMTD(tx *gorm.DB, input MemberAutoTradingMTD) app.MsgStruct {
	var (
		strategyCode           = "MTD"
		memberID     int       = input.MemberID
		settingType  string    = input.Type
		cryptoPair   string    = input.CryptoPair
		amount       float64   = input.Amount
		curDateTime  time.Time = base.GetCurrentDateTimeT()
		langCode     string    = input.LangCode
	)

	var (
		memberCurrentAPI = GetMemberCurrentAPI(memberID)
		platform         = memberCurrentAPI.PlatformCode
		apiKey           = ""
		secret           = ""
		passphrase       = ""
		apiKey2          = ""
		secret2          = ""
		passphrase2      = ""
	)
	if platform == "" {
		return app.MsgStruct{Msg: "please_setup_api_to_proceed"}
	}

	for _, apiDetails := range memberCurrentAPI.ApiDetails {
		if apiDetails.Module == "SPOT" {
			apiKey = apiDetails.ApiKey
			secret = apiDetails.ApiSecret
			passphrase = apiDetails.ApiPassphrase
		} else if apiDetails.Module == "FUTURE" {
			apiKey2 = apiDetails.ApiKey
			secret2 = apiDetails.ApiSecret
			passphrase2 = apiDetails.ApiPassphrase
		}
	}

	// basic add auto trading flow which include validation + insert to sls_master
	var addMemberAutoTradingParam = MemberAutoTrading{
		MemberID:   memberID,
		Platform:   platform,
		PrdCode:    strategyCode,
		Type:       settingType,
		Amount:     amount,
		CryptoPair: cryptoPair,
		LangCode:   langCode,
	}

	slsMasterID, docNo, msgStruct := AddMemberAutoTrading(tx, addMemberAutoTradingParam)
	if msgStruct.Msg != "" {
		return msgStruct
	}

	var (
		firstOrderPrice     float64
		firstOrderAmount    float64 = input.FirstOrderAmount
		minFirstOrderAmount float64 = 0
		priceScale          float64
		takeProfitCallback  float64
		takeProfit          float64 // take profit callback
		safetyOrders        float64
		addShares           float64 = 2 // add shares fixed at 2
		circularTrans       int         // 1/0
		errMsg              string
	)

	// 反向马丁ai版和专业版：增加判断（币对合约规则），首单金额是否满足交易所合约规则最低买入金额。
	if platform == "KC" {
		// minFirstOrderAmount need to calculate from data grab in kucoin
		minFirstOrderAmount, errMsg = GetKucoinMinOrderAmount(cryptoPair)
		if errMsg != "" {
			base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:GetKucoinMinOrderAmount()", map[string]interface{}{"cryptoPair": cryptoPair}, errMsg, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	} else if platform == "BN" {
		// minFirstOrderAmount need to calculate from data grab in binance
		minFirstOrderAmount, errMsg = GetBinanceMinOrderAmount(cryptoPair)
		if errMsg != "" {
			base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:GetBinanceMinOrderAmount()", map[string]interface{}{"cryptoPair": cryptoPair}, errMsg, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	}

	if firstOrderAmount < minFirstOrderAmount {
		return app.MsgStruct{Msg: "minimum_first_order_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(minFirstOrderAmount, 8, ".", ",", true)}}
	}

	if firstOrderAmount > amount {
		return app.MsgStruct{Msg: "first_order_amount_cannot_more_than_trading_amount"}
	}

	// get first_order_price, take_profit_adjust and add_shares from sys_trading_crypto_pair_setup
	arrSysTradingCryptoPairSetupFn := make([]models.WhereCondFn, 0)
	arrSysTradingCryptoPairSetupFn = append(arrSysTradingCryptoPairSetupFn,
		models.WhereCondFn{Condition: "code = ?", CondValue: cryptoPair},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)
	arrSysTradingCryptoPairSetup, err := models.GetSysTradingCryptoPairSetupByPlatformFn(arrSysTradingCryptoPairSetupFn, "", platform, false)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:GetSysTradingCryptoPairSetupByPlatformFn()", map[string]interface{}{"condition": arrSysTradingCryptoPairSetupFn, "platform": platform}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}
	if len(arrSysTradingCryptoPairSetup) <= 0 {
		return app.MsgStruct{Msg: "invalid_crypto_pair"}
	}

	circularTrans = arrSysTradingCryptoPairSetup[0].MtdCircularTrans

	// get first order price
	// if platform != "KC" {
	// 	arrBinancePrice, err := GetBinanceCryptoPrice(cryptoPair)
	// 	if err != nil {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}

	// 	firstOrderPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
	// 	if err != nil {
	// 		base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:ParseFloat()", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}
	// } else {
	// arrKucoinPrice, err := GetKucoinCryptoPrice(cryptoPair)
	// if err != nil {
	// 	base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}
	// }

	// if arrKucoinPrice.Code != "200000" {
	// 	base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:GetKucoinCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair, "response": arrKucoinPrice}, arrKucoinPrice.Msg, true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}
	// }

	// firstOrderPrice, err = strconv.ParseFloat(arrKucoinPrice.Data.Price, 64)
	// if err != nil {
	// 	base.LogErrorLog("tradingService:AddMemberAutoTradingMTD:ParseFloat()", map[string]interface{}{"value": arrKucoinPrice.Data.Price}, err.Error(), true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}
	// }
	// }

	firstOrderPrice = 0 // only AI setting firstOrderPrice need to be set to 0

	if settingType == "AI" { // use default value
		priceScale = arrSysTradingCryptoPairSetup[0].MtdPriceScale
		takeProfitCallback = arrSysTradingCryptoPairSetup[0].MtdTakeProfitRatio
		takeProfit = arrSysTradingCryptoPairSetup[0].MtdTakeProfitAdjustment
		// addShares = float64(arrSysTradingCryptoPairSetup[0].MtdAddShares)

		// get safety orders
		safetyOrders = GetSafetyOrders(firstOrderAmount, amount, addShares)

	} else if settingType == "PROF" { // validate setting value
		// first order price
		if input.FirstOrderPrice == 0 {
			return app.MsgStruct{Msg: "first_order_price_is_required"}
		}

		// price scale
		if input.PriceScale == 0 {
			return app.MsgStruct{Msg: "price_scale_is_required"}
		}

		// // subsequent price scale
		// if !(input.SubsequentPriceScale > 0) || input.SubsequentPriceScale != float64(int(input.SubsequentPriceScale)) {
		// 	return app.MsgStruct{Msg: "subsequent_price_scale_must_be_a_positive_whole_number"}
		// }

		// // subsequent add shares
		// if !(input.SubsequentAddShares > 0) || input.SubsequentAddShares != float64(int(input.SubsequentAddShares)) {
		// 	return app.MsgStruct{Msg: "subsequent_add_shares_must_be_a_positive_whole_number"}
		// }

		// take profit callback
		if input.TakeProfitCallback == 0 {
			return app.MsgStruct{Msg: "take_profit_callback_is_required"}
		}

		// take profit
		if input.TakeProfit == 0 {
			return app.MsgStruct{Msg: "take_profit_is_required"}
		}

		// safety orders must be integer
		// if input.SafetyOrders == 0 {
		// 	return app.MsgStruct{Msg: "safety_orders_is_required"}
		// }

		// if !(input.SafetyOrders > 0) || input.SafetyOrders != float64(int(input.SafetyOrders)) {
		// 	return app.MsgStruct{Msg: "safety_orders_must_be_a_positive_whole_number"}
		// }

		// safety orders must be integer
		// if input.AddShares == 0 {
		// 	return app.MsgStruct{Msg: "add_shares_is_required"}
		// }

		// if !(input.AddShares > 0) || input.AddShares != float64(int(input.AddShares)) {
		// 	return app.MsgStruct{Msg: "add_shares_must_be_a_positive_whole_number"}
		// }

		priceScale = input.PriceScale
		firstOrderPrice = input.FirstOrderPrice
		takeProfitCallback = input.TakeProfitCallback
		takeProfit = input.TakeProfit
		// addShares = input.AddShares

		// get safety orders
		safetyOrders = GetSafetyOrders(firstOrderAmount, amount, addShares)
	} else {
		return app.MsgStruct{Msg: "invalid_type"}
	}

	if safetyOrders <= 0 {
		return app.MsgStruct{Msg: "invalid_safety_orders"}
	}

	// convert price scale from 1% to 0.01
	// priceScale = float.Div(priceScale, 100)

	// insert to sls_master_bot_log
	var addSlsMasterBotLog = models.SlsMasterBotLog{
		MemberID:   memberID,
		DocNo:      docNo,
		Status:     "A",
		RemarkType: "S",
		Remark:     strings.Replace(cryptoPair, "USDT", "/USDT", -1) + " #*first_order_price*#:" + helpers.CutOffDecimalv2(firstOrderPrice, 2, ".", ",", true) + " #*doc_no*#:" + docNo,
		CreatedAt:  time.Now(),
		CreatedBy:  "AUTO",
	}
	_, err = models.AddSlsMasterBotLog(tx, addSlsMasterBotLog)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMTD():AddSlsMasterBotLog():1", err.Error(), map[string]interface{}{"param": addSlsMasterBotLog}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// insert to sls_master_bot_setting
	var arrSetting = map[string]interface{}{
		"firstOrderPrice":  firstOrderPrice,
		"firstOrderAmount": firstOrderAmount,
		// "priceScale":           fmt.Sprintf("%f%%", priceScale),
		// "takeProfitRatio":      fmt.Sprintf("%f%%", takeProfitCallback),
		// "takeProfitAdjust":     fmt.Sprintf("%f%%", takeProfit),
		"priceScale":           float.Div(priceScale, 100),
		"takeProfitRatio":      float.Div(takeProfitCallback, 100),
		"takeProfitAdjust":     float.Div(takeProfit, 100),
		"safetyOrders":         safetyOrders,
		"addShares":            addShares,
		"circularTransactions": circularTrans, // 1/0
	}

	c, err := json.Marshal(arrSetting)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMTD():Marshal():1", map[string]interface{}{"param": arrSetting}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	var setting string = string(c)
	var addSlsMasterBotSettingParams = models.AddSlsMasterBotSettingStruct{
		SlsMasterID: slsMasterID,
		Platform:    platform,
		SettingType: settingType,
		CryptoPair:  cryptoPair,
		Setting:     setting,
		CreatedAt:   curDateTime,
		CreatedBy:   strconv.Itoa(memberID),
	}

	_, err = models.AddSlsMasterBotSetting(tx, addSlsMasterBotSettingParams)
	if err != nil {
		base.LogErrorLog("tradingService:AddMemberAutoTradingMTD():AddSlsMasterBotSetting():1", map[string]interface{}{"param": addSlsMasterBotSettingParams}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	// call china api to add bot - always call api at last step
	var memberAutoTradingApiParam = MemberAutoTradingApi{
		DocNo:      docNo,
		Platform:   platform,
		AppID:      apiKey,
		Secret:     secret,
		AppPwd:     passphrase,
		AppID2:     apiKey2,
		Secret2:    secret2,
		AppPwd2:    passphrase2,
		Strategy:   strategyCode,
		CryptoPair: cryptoPair,
		Amount:     amount,
		Setting:    setting,
	}
	errMsg = PostMemberAutoTradingApi(memberAutoTradingApiParam)
	if errMsg != "" {
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}
}

func GetSafetyOrders(firstOrderAmount, tradingAmount float64, addShares float64) float64 {
	var (
		safetyOrders       float64 = 1
		currentOrderAmount float64 = firstOrderAmount
		ordersSequence             = []float64{firstOrderAmount}
	)

	// for currentOrderAmount <= tradingAmount {
	// 	safetyOrders++

	// 	// if current total invested amount more than trading amount then stop.
	// 	currentOrderAmount = currentOrderAmount + float.Mul(currentOrderAmount, addShares)
	// 	if currentOrderAmount > tradingAmount {
	// 		break
	// 	}
	// }

	// alan version - 2023/04/09
	for currentOrderAmount <= tradingAmount {
		if safetyOrders == 1 {
			ordersSequence = append(ordersSequence, firstOrderAmount*2)
		} else {
			nextOrderAmount := 0.00
			for _, curOrdersAmount := range ordersSequence {
				nextOrderAmount += curOrdersAmount
			}

			ordersSequence = append(ordersSequence, nextOrderAmount)
		}

		// sum total order amount see if exceed trading amount
		totalOrdersAmount := 0.00
		for _, curOrdersAmount := range ordersSequence {
			totalOrdersAmount += curOrdersAmount
		}

		// fmt.Println("safetyOrders:", safetyOrders+1, "totalOrdersAmount: ", totalOrdersAmount, "ordersSequence: ", ordersSequence)
		if totalOrdersAmount > tradingAmount {
			break
		}

		safetyOrders++
	}

	return safetyOrders
}

// MemberAutoTradingApi struct
type MemberAutoTradingApi struct {
	DocNo      string  `json:"docNo"`
	Platform   string  `json:"platform"`
	AppID      string  `json:"appId"`
	Secret     string  `json:"secret"`
	AppPwd     string  `json:"apiPwd"`
	AppID2     string  `json:"appId2"`
	Secret2    string  `json:"secret2"`
	AppPwd2    string  `json:"apiPwd2"`
	Strategy   string  `json:"strategy"`
	CryptoPair string  `json:"cryptoPair"`
	Amount     float64 `json:"amount"`
	Setting    string  `json:"setting"`
}

func PostMemberAutoTradingApi(input MemberAutoTradingApi) string {
	var (
		platform    string = input.Platform
		body               = map[string]interface{}{}
		mergedValue string = ""
		hashValue   string = ""
		salt        string = ""
	)

	// double checking in case member do not have required apiKey, seret or passphrase
	if platform == "BN" && (input.AppID == "" || input.Secret == "") {
		base.LogErrorLog("tradingService:PostMemberAutoTradingApi()", "invalid_api_management_setup", map[string]interface{}{"param": input}, true)
		return "something_went_wrong"
	}

	if platform == "KC" && (input.AppID == "" || input.Secret == "" || input.AppPwd == "" || input.AppID2 == "" || input.Secret2 == "" || input.AppPwd2 == "") {
		base.LogErrorLog("tradingService:PostMemberAutoTradingApi()", "invalid_api_management_setup", map[string]interface{}{"param": input}, true)
		return "something_went_wrong"
	}

	if platform == "BN" {
		input.Platform = "Binance"
	}
	if platform == "KC" {
		input.Platform = "kucoin"

		input.CryptoPair = strings.Replace(input.CryptoPair, "USDTM", "-USDT", -1)
		input.CryptoPair = strings.Replace(input.CryptoPair, "XBT", "BTC", -1)

		// if input.Strategy == "SGT" || input.Strategy == "MT" { // spot
		// 	// spot - BTC-USDT, future - XBT-USDT
		// 	input.CryptoPair = strings.Replace(input.CryptoPair, "XBT", "BTC", -1)
		// }
	}

	// convert MemberAutoTradingApi struct into map[string]interface{}
	var inputByte, _ = json.Marshal(input)
	err := json.Unmarshal(inputByte, &body)
	if err != nil {
		return err.Error()
	}

	// sort by keys
	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// loop annd merge value
	for _, k := range keys {
		mergedValue = fmt.Sprintf("%s%v", mergedValue, body[k])
	}

	// get and apply salt
	arrMediaSetting, _ := models.GetSysGeneralSetupByID("trading_salt")
	salt = arrMediaSetting.InputValue1
	mergedValue = fmt.Sprintf("%s%s", mergedValue, salt)

	keyByte := []byte(mergedValue)
	hasher := sha256.New()
	hasher.Write(keyByte)
	hashValue = hex.EncodeToString(hasher.Sum(nil))

	body["hashValue"] = hashValue

	apiSetting, _ := models.GetSysGeneralSetupByID("auto_trading_setting")
	if apiSetting.InputValue2 != "1" {
		return ""
	}

	url := apiSetting.InputValue1 + "/account/report"
	header := map[string]string{
		"Content-Type": "application/json",
	}

	response, err := base.RequestAPIV2("POST", url, header, body, nil, base.ExtraSettingStruct{})
	if err != nil {
		base.LogErrorLog("tradingService:PostMemberAutoTradingApi():RequestAPIV2()", err.Error(), map[string]interface{}{"url": url, "header": header, "body": body}, true)
		return "something_went_wrong"
	}

	type PostAutoTradingAccountReportAPIResponse struct {
		Rst  int    `json:"rst"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	var postAutoTradingAccountReportApiResponse = &PostAutoTradingAccountReportAPIResponse{}
	if response.Body == "" {
		return "something_went_wrong"
	}

	err = json.Unmarshal([]byte(response.Body), postAutoTradingAccountReportApiResponse)
	if err != nil {
		base.LogErrorLog("tradingService:PostMemberAutoTradingApi():Unmarshal():1", err.Error(), map[string]interface{}{"input": response.Body}, true)
		return "something_went_wrong"
	}

	if postAutoTradingAccountReportApiResponse.Rst != 1 {
		base.LogErrorLog("tradingService:PostMemberAutoTradingApi()", postAutoTradingAccountReportApiResponse.Msg, map[string]interface{}{"url": url, "header": header, "body": body, "response_body": postAutoTradingAccountReportApiResponse}, true)
		return "something_went_wrong"
	}

	return ""
}

type GetMemberStrategyBalanceStruct struct {
	MemberID int
	Platform string
	LangCode string
}

type MemberStrategyBalanceReturnStruct struct {
	Coin       string  `json:"coin"`
	Balance    float64 `json:"balance"`
	BalanceStr string  `json:"balance_str"`
	Desc       string  `json:"description"`
}

func (b *GetMemberStrategyBalanceStruct) GetMemberStrategyBalancev1() (MemberStrategyBalanceReturnStruct, string) {
	var (
		balance     float64
		balanceStr  string
		coin        string
		EwtTypeCode string = "USDT"
		descTrans   string = helpers.Translate("for_spot_grid_trading_and_martingale_trading_strategies", b.LangCode)
	)

	balanceStr = helpers.CutOffDecimal(balance, 8, ".", ",")

	// get api_keys & secret
	arrEntMemberTradingApiFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingApiFn = append(arrEntMemberTradingApiFn,
		models.WhereCondFn{Condition: "ent_member_trading_api.member_id = ?", CondValue: b.MemberID},
		models.WhereCondFn{Condition: "ent_member_trading_api.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ent_member_trading_api.platform = ?", CondValue: strings.ToUpper(b.Platform)},
		models.WhereCondFn{Condition: "ent_member_trading_api.module = ?", CondValue: "SPOT"},
		models.WhereCondFn{Condition: "sys_trading_api_platform.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingApi, err := models.GetEntMemberTradingApiFn(arrEntMemberTradingApiFn, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberStrategyBalancev1-GetEntMemberTradingApiFn", arrEntMemberTradingApiFn, err.Error(), true)
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "something_went_wrong"
	}

	if len(arrEntMemberTradingApi) < 1 {
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "please_setup_api_to_proceed"
	}

	// decrypt secret key
	decryptedScrtKey, err := util.DecodeAscii85(arrEntMemberTradingApi[0].ApiSecret)
	if err != nil {
		base.LogErrorLog("GetMemberStrategyBalancev1-DecodeAscii85", map[string]interface{}{"decryptedScrtKey": decryptedScrtKey, "input": arrEntMemberTradingApi[0].ApiSecret}, err.Error(), true)
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "something_went_wrong"
	}

	// check ewt_summary_strategy record
	arrEwtSummaryStrategyFn := make([]models.WhereCondFn, 0)
	arrEwtSummaryStrategyFn = append(arrEwtSummaryStrategyFn,
		models.WhereCondFn{Condition: "ewt_summary_strategy.member_id = ?", CondValue: b.MemberID},
		models.WhereCondFn{Condition: "ewt_summary_strategy.coin = ?", CondValue: strings.ToUpper(EwtTypeCode)},
		models.WhereCondFn{Condition: "ewt_summary_strategy.platform = ?", CondValue: strings.ToUpper(b.Platform)},
	)
	arrEwtSummaryStrategy, err := models.GetEwtSummaryStrategyFn(arrEwtSummaryStrategyFn, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberStrategyBalancev1-GetEwtSummaryStrategyFn", map[string]interface{}{"condition": arrEwtSummaryStrategyFn}, err.Error(), true)
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "something_went_wrong"
	}

	switch strings.ToUpper(b.Platform) {
	case "BN":
		currentUnixTimestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
		data := fmt.Sprintf("timestamp=%v", currentUnixTimestamp)
		signature := util.GenerateHmacSHA256(decryptedScrtKey, data, "")

		// begin call binance balance api
		binanceApi := GetMemberBinanceBalanceApiStruct{
			ApiKey:    arrEntMemberTradingApi[0].ApiKey,
			Timestamp: currentUnixTimestamp,
			Signature: signature,
		}
		rst, err := binanceApi.GetBinanceBalanceApiv1()

		if err != nil {
			models.ErrorLog("GetMemberStrategyBalancev1-GetBinanceBalanceApiv1 Error", map[string]interface{}{"memID": b.MemberID, "api_key": arrEntMemberTradingApi[0].ApiKey, "scrt_key": decryptedScrtKey, "timestamp": currentUnixTimestamp}, nil)
			//if api down grab from ewt_summary_bn record
			if len(arrEwtSummaryStrategy) > 0 {
				coin = arrEwtSummaryStrategy[0].Coin
				balance = arrEwtSummaryStrategy[0].Balance
				balanceStr = helpers.CutOffDecimal(balance, 8, ".", ",")
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: helpers.CutOffDecimal(balance, 8, ".", ","),
					Desc:       descTrans,
				}, ""
			} else {
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: balanceStr,
					Desc:       descTrans,
				}, ""
			}

		}

		for _, v := range rst {
			if v.Coin == strings.ToUpper(EwtTypeCode) {
				coin = v.Coin
				balance, _ = strconv.ParseFloat(v.Free, 64)
				balanceStr = v.Free
			}
		}

	case "KC":
		// begin call kucoin balance api
		kucoinAccountParam := GetMemberKucoinSpotAccountApiTradingStatusParam{
			ApiKey:     arrEntMemberTradingApi[0].ApiKey,
			Secret:     decryptedScrtKey,
			Passphrase: arrEntMemberTradingApi[0].ApiPassphrase,
		}
		rst, err := kucoinAccountParam.GetKucoinSpotAccountApiTradingStatus()

		if err != nil {
			models.ErrorLog("GetMemberStrategyBalancev1():GetKucoinSpotAccountApiTradingStatus", map[string]interface{}{"param": kucoinAccountParam}, nil)

			// if api down grab from ewt_summary_bn record
			if len(arrEwtSummaryStrategy) > 0 {
				coin = arrEwtSummaryStrategy[0].Coin
				balance = arrEwtSummaryStrategy[0].Balance
				balanceStr = helpers.CutOffDecimal(balance, 8, ".", ",")
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: helpers.CutOffDecimal(balance, 8, ".", ","),
					Desc:       descTrans,
				}, ""
			} else {
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: balanceStr,
					Desc:       descTrans,
				}, ""
			}

		}

		if len(rst.Data) > 0 {
			coin = rst.Data[0].Currency
			balance, _ = strconv.ParseFloat(rst.Data[0].Available, 64)
			balanceStr = helpers.CutOffDecimalv2(balance, 6, ".", ",", true)
		}
	}

	if len(arrEwtSummaryStrategy) > 0 {
		//update record
		arrUpdateEwtSummaryStrategy := make([]models.WhereCondFn, 0)
		arrUpdateEwtSummaryStrategy = append(arrUpdateEwtSummaryStrategy,
			models.WhereCondFn{Condition: " ewt_summary_strategy.member_id = ? ", CondValue: b.MemberID},
		)
		updateColumn := map[string]interface{}{
			"balance":    balance,
			"updated_at": time.Now(),
		}
		models.UpdatesFn("ewt_summary_strategy", arrUpdateEwtSummaryStrategy, updateColumn, false)

	} else {
		//store record
		arrStoreEwtSummaryStrategy := models.AddEwtSummaryStrategyStruct{
			MemberID:  b.MemberID,
			Platform:  strings.ToUpper(b.Platform),
			Coin:      strings.ToUpper(EwtTypeCode),
			Balance:   balance,
			CreatedBy: "AUTO",
			CreatedAt: time.Now(),
		}
		models.AddEwtSummaryStrategy(arrStoreEwtSummaryStrategy)
	}

	arrDataReturn := MemberStrategyBalanceReturnStruct{
		Coin:       coin,
		Balance:    balance,
		BalanceStr: balanceStr,
		Desc:       descTrans,
	}

	return arrDataReturn, ""
}

type GetMemberBinanceBalanceApiStruct struct {
	Timestamp string
	Signature string
	ApiKey    string
}

type BinanceBalanceResponse struct {
	Coin string `json:"coin"`
	Free string `json:"free"`
}

func (b *GetMemberBinanceBalanceApiStruct) GetBinanceBalanceApiv1() ([]*BinanceBalanceResponse, error) {

	var (
		err      error
		response []*BinanceBalanceResponse
	)

	apiSetting, _ := models.GetSysGeneralSetupByID("binance_api_setting")

	if apiSetting == nil {
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "invalid_setting"}
	}

	data := map[string]interface{}{
		"timestamp": b.Timestamp,
		"signature": b.Signature,
	}

	url := apiSetting.InputValue1 + fmt.Sprintf("sapi/v1/capital/config/getall?timestamp=%v&signature=%v", b.Timestamp, b.Signature)
	header := map[string]string{
		"Content-Type": "application/json",
		"X-MBX-APIKEY": b.ApiKey,
	}

	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetBinanceBalanceApiv1-GetBinanceSpotBalanceApi failed", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetBinanceBalanceApiv1-BinanceSpotBalanceApiReturnErr", res.Body, map[string]interface{}{"res": res, "data": data}, true)
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return response, nil
}

// AutoTradingLiquidationParam struct
type AutoTradingLiquidationParam struct {
	MemberID     int
	DocNo        string
	StrategyCode string
	LangCode     string
}

func PostAutoTradingLiquidation(tx *gorm.DB, input AutoTradingLiquidationParam) app.MsgStruct {
	var (
		memberID               = input.MemberID
		docNo                  = input.DocNo
		strategyCode           = input.StrategyCode
		curDateTime  time.Time = base.GetCurrentDateTimeT()
	)

	// check if got active auto bot for this strategy code
	arrSlsMasterBotSettingFn := make([]models.WhereCondFn, 0)
	arrSlsMasterBotSettingFn = append(arrSlsMasterBotSettingFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: " sls_master.action = ?", CondValue: "BOT"},
		models.WhereCondFn{Condition: " sls_master.status = ?", CondValue: "AP"},
		// models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: strategyCode},
	)

	if strategyCode != "ALL" {
		arrSlsMasterBotSettingFn = append(arrSlsMasterBotSettingFn,
			models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: strategyCode},
		)
	}

	if docNo != "" {
		arrSlsMasterBotSettingFn = append(arrSlsMasterBotSettingFn,
			models.WhereCondFn{Condition: " (sls_master.doc_no = ? OR sls_master.ref_no = '" + docNo + "')", CondValue: docNo},
		)
	}

	arrSlsMasterBotSetting, err := models.GetSlsMasterBotSetting(arrSlsMasterBotSettingFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:PostAutoTradingLiquidation():GetSlsMasterBotSetting():1", err.Error(), arrSlsMasterBotSettingFn, true)
	}
	if len(arrSlsMasterBotSetting) <= 0 {
		return app.MsgStruct{Msg: "invalid_doc_no_or_strategy_code"}
	}

	var arrActiveStrategyCode = []string{}
	for _, arrSlsMasterBotSettingV := range arrSlsMasterBotSetting {
		// update sls_master.status and expired_at
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: "id = ?", CondValue: arrSlsMasterBotSettingV.SlsMasterID},
		)
		updateColumn := map[string]interface{}{"status": "EP", "expired_at": curDateTime, "updated_at": curDateTime, "updated_by": fmt.Sprint(memberID)}
		err = models.UpdatesFnTx(tx, "sls_master", arrUpdCond, updateColumn, false)
		if err != nil {
			base.LogErrorLog("tradingService:PostAutoTradingLiquidation():UpdatesFnTx():1", err.Error(), map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}

		if !helpers.StringInSlice(arrSlsMasterBotSettingV.PrdMasterCode, arrActiveStrategyCode) {
			arrActiveStrategyCode = append(arrActiveStrategyCode, arrSlsMasterBotSettingV.PrdMasterCode)
		}

		// call api
		curDocNo := arrSlsMasterBotSettingV.DocNo
		if arrSlsMasterBotSettingV.RefNo != "" {
			curDocNo = arrSlsMasterBotSettingV.RefNo
		}

		var memberAutoTradingLiquidationApi = MemberAutoTradingLiquidationApi{
			DocNo:    curDocNo,
			Strategy: arrSlsMasterBotSettingV.PrdMasterCode,
		}
		errMsg := PostMemberAutoTradingLiquidationApi(memberAutoTradingLiquidationApi)
		if errMsg != "" {
			return app.MsgStruct{Msg: "something_went_wrong"}
		}

		// add to sls_master_bot_log
		var addSlsMasterBotLog = models.SlsMasterBotLog{
			MemberID:   memberID,
			DocNo:      arrSlsMasterBotSettingV.DocNo,
			Status:     "A",
			RemarkType: "S",
			Remark:     strings.Replace(arrSlsMasterBotSettingV.CryptoPair, "USDT", "/USDT", 1) + " #*order_liquidated*# - #*doc_no*#: " + docNo,
			CreatedAt:  time.Now(),
			CreatedBy:  "AUTO",
		}

		_, err = models.AddSlsMasterBotLog(tx, addSlsMasterBotLog)
		if err != nil {
			base.LogErrorLog("tradingService:PostAutoTradingLiquidation():AddSlsMasterBotLog():1", err.Error(), map[string]interface{}{"param": addSlsMasterBotLog}, true)
			return app.MsgStruct{Msg: "something_went_wrong"}
		}
	}

	// get member api key
	// arrEntMemberTradingApiFn := make([]models.WhereCondFn, 0)
	// arrEntMemberTradingApiFn = append(arrEntMemberTradingApiFn,
	// 	models.WhereCondFn{Condition: "ent_member_trading_api.member_id = ?", CondValue: memberID},
	// 	models.WhereCondFn{Condition: "ent_member_trading_api.status = ?", CondValue: "A"},
	// 	models.WhereCondFn{Condition: "sys_trading_api_platform.status = ?", CondValue: "A"},
	// )
	// arrEntMemberTradingApi, err := models.GetEntMemberTradingApiFn(arrEntMemberTradingApiFn, "", false)
	// if err != nil {
	// 	base.LogErrorLog("tradingService:PostAutoTradingLiquidation:GetEntMemberTradingApiFn()", map[string]interface{}{"condition": arrEntMemberTradingApiFn}, err.Error(), true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}
	// }
	// if len(arrEntMemberTradingApi) <= 0 {
	// 	return app.MsgStruct{Msg: "member_trading_api_not_found"}
	// }

	// for _, arrActiveStrategyCodeV := range arrActiveStrategyCode {
	// 	// call api
	// 	var memberAutoTradingLiquidationApi = MemberAutoTradingLiquidationApi{
	// 		DocNo:    arrEntMemberTradingApi[0].ApiKey,
	// 		Strategy: arrActiveStrategyCodeV,
	// 	}
	// 	errMsg := PostMemberAutoTradingLiquidationApi(memberAutoTradingLiquidationApi)
	// 	if errMsg != "" {
	// 		return app.MsgStruct{Msg: "something_went_wrong"}
	// 	}
	// }

	return app.MsgStruct{Msg: ""}
}

// MemberAutoTradingLiquidationApi struct
type MemberAutoTradingLiquidationApi struct {
	DocNo    string `json:"docNo"` // change to doc_no
	Strategy string `json:"strategy"`
}

func PostMemberAutoTradingLiquidationApi(input MemberAutoTradingLiquidationApi) string {
	var (
		body               = map[string]interface{}{}
		mergedValue string = ""
		hashValue   string = ""
		salt        string = ""
	)

	// convert MemberAutoTradingLiquidationApi struct into map[string]interface{}
	var inputByte, _ = json.Marshal(input)
	err := json.Unmarshal(inputByte, &body)
	if err != nil {
		return err.Error()
	}

	// sort by keys
	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// loop annd merge value
	for _, k := range keys {
		mergedValue = fmt.Sprintf("%s%v", mergedValue, body[k])
	}

	// get and apply salt
	arrMediaSetting, _ := models.GetSysGeneralSetupByID("trading_salt")
	salt = arrMediaSetting.InputValue1
	mergedValue = fmt.Sprintf("%s%s", mergedValue, salt)

	keyByte := []byte(mergedValue)
	hasher := sha256.New()
	hasher.Write(keyByte)
	hashValue = hex.EncodeToString(hasher.Sum(nil))

	body["hashValue"] = hashValue

	apiSetting, _ := models.GetSysGeneralSetupByID("auto_trading_setting")
	if apiSetting.InputValue2 != "1" {
		return ""
	}

	url := apiSetting.InputValue1 + "/account/stop"
	header := map[string]string{
		"Content-Type": "application/json",
	}

	response, err := base.RequestAPIV2("POST", url, header, body, nil, base.ExtraSettingStruct{})
	if err != nil {
		base.LogErrorLog("tradingService:PostMemberAutoTradingLiquidationApi():RequestAPIV2()", err.Error(), map[string]interface{}{"url": url, "header": header, "body": body}, true)
		return "something_went_wrong"
	}

	type PostAutoTradingLiquidationAPIResponse struct {
		Rst  int    `json:"rst"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}

	var autoTradingLiquidationPointer = &PostAutoTradingLiquidationAPIResponse{}
	if response.Body == "" {
		base.LogErrorLog("tradingService:PostMemberAutoTradingLiquidationApi():RequestAPIV2()", "return_empty_body", map[string]interface{}{"url": url, "header": header, "body": body}, true)
		return "something_went_wrong"
	}

	err = json.Unmarshal([]byte(response.Body), autoTradingLiquidationPointer)
	if err != nil {
		base.LogErrorLog("tradingService:PostMemberAutoTradingLiquidationApi():Unmarshal():1", err.Error(), map[string]interface{}{"input": response.Body}, true)
		return "something_went_wrong"
	}

	if autoTradingLiquidationPointer.Rst != 1 {
		base.LogErrorLog("tradingService:PostMemberAutoTradingLiquidationApi()", autoTradingLiquidationPointer.Msg, map[string]interface{}{"url": url, "header": header, "body": body, "response_body": autoTradingLiquidationPointer}, true)
		return "something_went_wrong"
	}

	return ""
}

type GetMemberAutoTradingTransactionParam struct {
	Strategy   string
	CryptoPair string
	DateFrom   string
	DateTo     string
	Page       int64
}

type TradingTransactionStruct struct {
	Strategy           string      `json:"strategy"`
	CryptoPair         string      `json:"crypto_pair"`
	OrderNo            string      `json:"order_no"`
	AutoTradingApiIcon string      `json:"auto_trading_api_icon"`
	CreatedAt          string      `json:"created_at"`
	Details            interface{} `json:"details"`
}

func GetMemberAutoTradingTransaction(memID int, param GetMemberAutoTradingTransactionParam, langCode string) (interface{}, string) {
	var (
		arrListingData  = []interface{}{}
		arrListingData1 = []interface{}{}
		arrListingData2 = []interface{}{}
		arrListingData3 = []interface{}{}
		errMsg          string
	)

	arrListingData1, errMsg = GetMemberCryptoFundingTransaction(memID, param, langCode)
	if errMsg != "" {
		return nil, errMsg
	}
	for _, arrListingData1V := range arrListingData1 {
		arrListingData = append(arrListingData, arrListingData1V)
	}

	arrListingData2, errMsg = GetMemberSpotAndMartingaleTransaction(memID, param, langCode)
	if errMsg != "" {
		return nil, errMsg
	}

	for _, arrListingData2V := range arrListingData2 {
		arrListingData = append(arrListingData, arrListingData2V)
	}

	arrListingData3, errMsg = GetMemberReverseMartingaleTransaction(memID, param, langCode)
	if errMsg != "" {
		return nil, errMsg
	}

	for _, arrListingData3V := range arrListingData3 {
		arrListingData = append(arrListingData, arrListingData3V)
	}

	sort.Slice(arrListingData, func(i, j int) bool {
		commonID1 := reflect.ValueOf(arrListingData[i]).FieldByName("CreatedAt").String()
		commonID2 := reflect.ValueOf(arrListingData[j]).FieldByName("CreatedAt").String()
		return commonID1 > commonID2
	})

	page := base.Pagination{
		Page:    param.Page,
		DataArr: arrListingData,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, ""
}

func GetMemberCryptoFundingTransaction(memID int, param GetMemberAutoTradingTransactionParam, langCode string) ([]interface{}, string) {
	var (
		currencyCode       = "USDT"
		orderQtyLabel      = helpers.TranslateV2("order_qty", langCode, map[string]string{})
		buyQuantityLabel   = helpers.TranslateV2("buy_quantity", langCode, map[string]string{})
		buyPriceLabel      = helpers.TranslateV2("buy_price", langCode, map[string]string{})
		tradePriceLabel    = helpers.TranslateV2("trade_price", langCode, map[string]string{})
		tradeQuoteQtyLabel = helpers.TranslateV2("trade_quote_qty", langCode, map[string]string{})
		tradeQtyLabel      = helpers.TranslateV2("total_trade_qty", langCode, map[string]string{})
		profitLabel        = helpers.TranslateV2("profit", langCode, map[string]string{})
		orderTypeLabel     = helpers.TranslateV2("order_type", langCode, map[string]string{})
		spotLabel          = helpers.TranslateV2("spot", langCode, map[string]string{})
		futureLabel        = helpers.TranslateV2("future", langCode, map[string]string{})
	)

	// get member transaction record
	var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
	arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " prd_master.code IN(?,'CIFRA')", CondValue: "CFRA"},
		models.WhereCondFn{Condition: " ent_member_trading_transaction.status = ?", CondValue: "A"},
		// models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "BUY"},
	)

	if param.Strategy != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: param.Strategy},
		)
	}

	if param.CryptoPair != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " ent_member_trading_transaction.crypto_pair = ?", CondValue: param.CryptoPair},
		)
	}

	if param.DateFrom != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " date(ent_member_trading_transaction.timestamp) >= ?", CondValue: param.DateFrom},
		)
	}

	if param.DateTo != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " date(ent_member_trading_transaction.timestamp) <= ?", CondValue: param.DateTo},
		)
	}

	var arrEntMemberTradingTransaction, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberCryptoFundingTransaction():GetEntMemberTradingTransactionFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn}, true)
		return nil, "something_went_wrong"
	}

	var arrListingData = []interface{}{}

	if len(arrEntMemberTradingTransaction) > 0 {
		for _, arrEntMemberTradingTransactionV := range arrEntMemberTradingTransaction {
			var (
				details                    = []interface{}{}
				tradeCurrencyCode          = strings.Replace(arrEntMemberTradingTransactionV.CryptoPair, "USDT", "", -1)
				profit             float64 = 0
				profitValue        string  = ""
				autoTradingApiIcon         = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/Binance_Logo_x3.png"
			)

			if arrEntMemberTradingTransactionV.Platform == "KC" {
				autoTradingApiIcon = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/KUCoin.png"
			}

			// get profit
			var arrTblqBonusStrategyProfitFn = make([]models.WhereCondFn, 0)
			arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
				models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.doc_no = ?", CondValue: arrEntMemberTradingTransactionV.DocNo},
				models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.order_id = ?", CondValue: arrEntMemberTradingTransactionV.OrderId},
				models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.member_id = ?", CondValue: memID},
			)

			var arrTblqBonusStrategyProfit, err = models.GetTblqBonusStrategyProfitFn(arrTblqBonusStrategyProfitFn, false)
			if err != nil {
				base.LogErrorLog("tradingService:GetMemberCryptoFundingTransaction():GetTblqBonusStrategyProfitFn():1", err.Error(), map[string]interface{}{"condition": arrTblqBonusStrategyProfitFn}, true)
				return nil, "something_went_wrong"
			}
			if len(arrTblqBonusStrategyProfit) > 0 {
				for _, arrTblqBonusStrategyProfitV := range arrTblqBonusStrategyProfit {
					profit += arrTblqBonusStrategyProfitV.FProfit
				}
			}

			profitValue = "-"
			if profit > 0 {
				profitValue = helpers.CutOffDecimalv2(profit, 6, ".", ",", true)
			}

			tPrice := arrEntMemberTradingTransactionV.TPrice
			if arrEntMemberTradingTransactionV.Platform == "KC" && arrEntMemberTradingTransactionV.TQuoteQty > 0 && arrEntMemberTradingTransactionV.TQty > 0 {
				tPrice = float.Div(arrEntMemberTradingTransactionV.TQuoteQty, arrEntMemberTradingTransactionV.TQty)
			}

			orderType := ""
			if arrEntMemberTradingTransactionV.OrderType == "SPOT" {
				orderType = spotLabel
			} else if arrEntMemberTradingTransactionV.OrderType == "FUTURE" {
				orderType = futureLabel
			}

			if arrEntMemberTradingTransactionV.Strategy == "CFRA" {
				details = append(details,
					map[string]string{
						"label": fmt.Sprintf("%s (%s)", orderQtyLabel, currencyCode),
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TotalBv, 2, ".", ",", true),
					},
					map[string]string{
						"label": buyQuantityLabel,
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.Num, 2, ".", ",", true),
					},
					map[string]string{
						"label": buyPriceLabel,
						"value": helpers.CutOffDecimalv2(tPrice, 2, ".", ",", true),
					},
					map[string]string{
						"label": tradeQtyLabel,
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TQty, 2, ".", ",", true),
					},
					map[string]string{
						"label": tradePriceLabel,
						"value": helpers.CutOffDecimalv2(tPrice, 2, ".", ",", true),
					},
					map[string]string{
						"label": fmt.Sprintf("%s (%s)", tradeQuoteQtyLabel, tradeCurrencyCode),
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TQuoteQty, 2, ".", ",", true),
					},
					map[string]string{
						"label": orderTypeLabel,
						"value": orderType,
					},
				)
			} else {
				details = append(details,
					map[string]string{
						"label": fmt.Sprintf("%s (%s)", orderQtyLabel, currencyCode),
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TotalBv, 2, ".", ",", true),
					},
					map[string]string{
						"label": buyQuantityLabel,
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.Num, 2, ".", ",", true),
					},
					map[string]string{
						"label": buyPriceLabel,
						"value": helpers.CutOffDecimalv2(tPrice, 2, ".", ",", true),
					},
					map[string]string{
						"label": tradeQtyLabel,
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TQty, 2, ".", ",", true),
					},
					map[string]string{
						"label": tradePriceLabel,
						"value": helpers.CutOffDecimalv2(tPrice, 2, ".", ",", true),
					},
					map[string]string{
						"label": fmt.Sprintf("%s (%s)", tradeQuoteQtyLabel, tradeCurrencyCode),
						"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TQuoteQty, 2, ".", ",", true),
					},
					map[string]string{
						"label": profitLabel,
						"value": profitValue,
					},
					map[string]string{
						"label": orderTypeLabel,
						"value": orderType,
					},
				)
			}

			var cryptoPair = arrEntMemberTradingTransactionV.CryptoPairName
			if cryptoPair == "" {
				cryptoPair = strings.Replace(arrEntMemberTradingTransactionV.CryptoPair, "USDT", "/USDT", 1)
				cryptoPair = strings.Replace(cryptoPair, "USDTM", "USDT", 1)
				cryptoPair = strings.Replace(cryptoPair, "XBT", "BTC", 1)
				cryptoPair = strings.Replace(cryptoPair, "-", "", 1)
			}

			createdAt := arrEntMemberTradingTransactionV.Timestamp.Format("2006-01-02 15:04:05")
			if arrEntMemberTradingTransactionV.TTime != 0 {
				createdAt = time.Unix(0, arrEntMemberTradingTransactionV.TTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
			}

			arrListingData = append(arrListingData,
				TradingTransactionStruct{
					Strategy:           arrEntMemberTradingTransactionV.StrategyName,
					CryptoPair:         cryptoPair,
					OrderNo:            arrEntMemberTradingTransactionV.DocNo,
					AutoTradingApiIcon: autoTradingApiIcon,
					CreatedAt:          createdAt,
					Details:            details,
				},
			)
		}
	}

	return arrListingData, ""
}

func GetMemberSpotAndMartingaleTransaction(memID int, param GetMemberAutoTradingTransactionParam, langCode string) ([]interface{}, string) {
	var (
		currencyCode       = "USDT"
		orderQtyLabel      = helpers.TranslateV2("order_qty", langCode, map[string]string{})
		orderPriceLabel    = helpers.TranslateV2("order_price", langCode, map[string]string{})
		tradeQtyLabel      = helpers.TranslateV2("trade_qty", langCode, map[string]string{})
		tradePriceLabel    = helpers.TranslateV2("trade_price", langCode, map[string]string{})
		tradeQuoteQtyLabel = helpers.TranslateV2("trade_quote_qty", langCode, map[string]string{})
		adminFeeLabel      = helpers.TranslateV2("admin_fee", langCode, map[string]string{})
		profitLabel        = helpers.TranslateV2("profit", langCode, map[string]string{})
		orderProgressLabel = helpers.TranslateV2("order_progress", langCode, map[string]string{})
		progressLabel      = helpers.TranslateV2("progress", langCode, map[string]string{})
		buyLabel           = helpers.TranslateV2("buy", langCode, map[string]string{})
		sellLabel          = helpers.TranslateV2("sell", langCode, map[string]string{})
	)

	// get member transaction record
	var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
	arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
		// models.WhereCondFn{Condition: " sls_master.status = ?", CondValue: "EP"},
		models.WhereCondFn{Condition: " prd_master.code IN(?,'MT')", CondValue: "SGT"},
		models.WhereCondFn{Condition: " ent_member_trading_transaction.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "BUY"},
		// models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "SELL"},
	)

	if param.Strategy != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: param.Strategy},
		)
	}

	if param.CryptoPair != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " ent_member_trading_transaction.crypto_pair = ?", CondValue: param.CryptoPair},
		)
	}

	if param.DateFrom != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " date(ent_member_trading_transaction.timestamp) >= ?", CondValue: param.DateFrom},
		)
	}

	if param.DateTo != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " date(ent_member_trading_transaction.timestamp) <= ?", CondValue: param.DateTo},
		)
	}

	var arrEntMemberTradingTransaction, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberSpotAndMartingaleTransaction():GetEntMemberTradingTransactionFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn}, false)
		return nil, "something_went_wrong"
	}

	var arrListingData = []interface{}{}

	if len(arrEntMemberTradingTransaction) > 0 {
		for _, arrEntMemberTradingTransactionV := range arrEntMemberTradingTransaction {
			var details = []interface{}{}
			var tradeCurrencyCode = strings.Replace(arrEntMemberTradingTransactionV.CryptoPair, "USDT", "", -1)

			var (
				profit             float64 = 0
				profitValue        string  = ""
				orderProgress      float64 = 0
				tQty                       = arrEntMemberTradingTransactionV.TQty
				tPrice                     = arrEntMemberTradingTransactionV.TPrice
				tQuoteQty                  = arrEntMemberTradingTransactionV.TQuoteQty
				tCommission                = arrEntMemberTradingTransactionV.TCommission
				autoTradingApiIcon         = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/Binance_Logo_x3.png"
			)

			if arrEntMemberTradingTransactionV.Platform == "KC" {
				autoTradingApiIcon = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/KUCoin.png"
			}

			// get sold transaction if got
			var arrEntMemberTradingTransactionFn2 = make([]models.WhereCondFn, 0)
			arrEntMemberTradingTransactionFn2 = append(arrEntMemberTradingTransactionFn2,
				models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
				models.WhereCondFn{Condition: " sls_master.doc_no = ? ", CondValue: arrEntMemberTradingTransactionV.DocNo},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.status = ?", CondValue: "A"},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "SELL"},
				// models.WhereCondFn{Condition: " ent_member_trading_transaction.remark2 = ?", CondValue: arrEntMemberTradingTransactionV.Remark2},
				// models.WhereCondFn{Condition: " ent_member_trading_transaction.timestamp >= ?", CondValue: arrEntMemberTradingTransactionV.Timestamp},
			)
			var arrEntMemberTradingTransaction2, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn2, "ASC", false)
			if err != nil {
				base.LogErrorLog("tradingService:GetMemberSpotAndMartingaleTransaction():GetEntMemberTradingTransactionFn():2", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn2}, true)
				return nil, "something_went_wrong"
			}

			progress := buyLabel
			if len(arrEntMemberTradingTransaction2) > 0 {
				progress = sellLabel

				tQty = arrEntMemberTradingTransaction2[0].TQty
				tPrice = arrEntMemberTradingTransaction2[0].TPrice

				if arrEntMemberTradingTransaction2[0].Platform == "KC" {
					tPrice = float.Div(arrEntMemberTradingTransaction2[0].TQuoteQty, arrEntMemberTradingTransaction2[0].TQty)
				}

				tQuoteQty = arrEntMemberTradingTransaction2[0].TQuoteQty
				tCommission = arrEntMemberTradingTransaction2[0].TCommission

				// get profit
				var arrTblqBonusStrategyProfitFn = make([]models.WhereCondFn, 0)
				arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
					models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.doc_no = ?", CondValue: arrEntMemberTradingTransactionV.DocNo},
					models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.order_id = ?", CondValue: arrEntMemberTradingTransaction2[0].OrderId},
					models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.member_id = ?", CondValue: memID},
				)

				var arrTblqBonusStrategyProfit, err = models.GetTblqBonusStrategyProfitFn(arrTblqBonusStrategyProfitFn, false)
				if err != nil {
					base.LogErrorLog("tradingService:GetMemberSpotAndMartingaleTransaction():GetTblqBonusStrategyProfitFn():1", err.Error(), map[string]interface{}{"condition": arrTblqBonusStrategyProfitFn}, true)
					return nil, "something_went_wrong"
				}
				if len(arrTblqBonusStrategyProfit) > 0 {
					for _, arrTblqBonusStrategyProfitV := range arrTblqBonusStrategyProfit {
						profit += arrTblqBonusStrategyProfitV.FProfit
					}
				}
			}

			// get order progress
			orderDocNo := arrEntMemberTradingTransactionV.RefNo
			if orderDocNo == "" {
				orderDocNo = arrEntMemberTradingTransactionV.DocNo
			}
			var arrEntMemberTradingTransactionFn3 = make([]models.WhereCondFn, 0)
			arrEntMemberTradingTransactionFn3 = append(arrEntMemberTradingTransactionFn3,
				models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.id <= ?", CondValue: arrEntMemberTradingTransactionV.ID},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_no like ?", CondValue: fmt.Sprintf("%s%%", orderDocNo)},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "BUY"},
			)

			arrEntMemberTradingTransaction3, err := models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn3, "", false)
			if err != nil {
				base.LogErrorLog("tradingService:GetMemberSpotAndMartingaleTransaction():GetEntMemberTradingTransactionFn():3", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn3}, true)
				return nil, "something_went_wrong"
			}
			orderProgress = float64(len(arrEntMemberTradingTransaction3))

			profitValue = "-"
			if profit != 0 {
				profitValue = helpers.CutOffDecimalv2(profit, 6, ".", ",", true)
			}

			details = append(details,
				// map[string]string{
				// 	"label": fmt.Sprintf("%s (%s)", orderQtyLabel, currencyCode),
				// 	"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TotalBv, 2, ".", ",", true),
				// },
				// map[string]string{
				// 	"label": tradePriceLabel,
				// 	"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TPrice, 2, ".", ",", true),
				// },
				// map[string]string{
				// 	"label": fmt.Sprintf("%s (%s)", tradeQuoteQtyLabel, tradeCurrencyCode),
				// 	"value": helpers.CutOffDecimalv2(tQty, 2, ".", ",", true),
				// },
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", orderQtyLabel, currencyCode),
					"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TotalBv, 2, ".", ",", true),
				},
				map[string]string{
					"label": orderPriceLabel,
					"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.Price, 2, ".", ",", true),
				},
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", tradeQtyLabel, currencyCode),
					"value": helpers.CutOffDecimalv2(tQuoteQty, 8, ".", ",", true),
				},
				map[string]string{
					"label": tradePriceLabel,
					"value": helpers.CutOffDecimalv2(tPrice, 2, ".", ",", true),
				},
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", tradeQuoteQtyLabel, tradeCurrencyCode),
					"value": helpers.CutOffDecimalv2(tQty, 8, ".", ",", true),
				},
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", adminFeeLabel, currencyCode),
					"value": helpers.CutOffDecimalv2(tCommission, 8, ".", ",", true),
				},
				map[string]string{
					"label": profitLabel,
					"value": profitValue,
				},
				map[string]string{
					"label": progressLabel,
					"value": progress,
				},
				map[string]string{
					"label": orderProgressLabel,
					"value": helpers.CutOffDecimalv2(orderProgress, 0, ".", ",", true),
				},
			)

			var cryptoPair = arrEntMemberTradingTransactionV.CryptoPairName
			if cryptoPair == "" {
				cryptoPair = strings.Replace(arrEntMemberTradingTransactionV.CryptoPair, "USDT", "/USDT", 1)
				cryptoPair = strings.Replace(cryptoPair, "USDTM", "USDT", 1)
				cryptoPair = strings.Replace(cryptoPair, "XBT", "BTC", 1)
			}

			createdAt := arrEntMemberTradingTransactionV.Timestamp.Format("2006-01-02 15:04:05")
			if arrEntMemberTradingTransactionV.TTime != 0 {
				createdAt = time.Unix(0, arrEntMemberTradingTransactionV.TTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
			}
			arrListingData = append(arrListingData,
				TradingTransactionStruct{
					Strategy:           arrEntMemberTradingTransactionV.StrategyName,
					CryptoPair:         cryptoPair,
					OrderNo:            arrEntMemberTradingTransactionV.DocNo,
					AutoTradingApiIcon: autoTradingApiIcon,
					CreatedAt:          createdAt,
					Details:            details,
				},
			)
		}
	}

	return arrListingData, ""
}

func GetMemberReverseMartingaleTransaction(memID int, param GetMemberAutoTradingTransactionParam, langCode string) ([]interface{}, string) {
	var (
		currencyCode       = "USDT"
		orderQtyLabel      = helpers.TranslateV2("order_qty", langCode, map[string]string{})
		orderPriceLabel    = helpers.TranslateV2("order_price", langCode, map[string]string{})
		tradeQtyLabel      = helpers.TranslateV2("trade_qty", langCode, map[string]string{})
		tradePriceLabel    = helpers.TranslateV2("trade_price", langCode, map[string]string{})
		tradeQuoteQtyLabel = helpers.TranslateV2("trade_quote_qty", langCode, map[string]string{})
		adminFeeLabel      = helpers.TranslateV2("admin_fee", langCode, map[string]string{})
		profitLabel        = helpers.TranslateV2("profit", langCode, map[string]string{})
		orderProgressLabel = helpers.TranslateV2("order_progress", langCode, map[string]string{})
		progressLabel      = helpers.TranslateV2("progress", langCode, map[string]string{})
		buyLabel           = helpers.TranslateV2("buy", langCode, map[string]string{})
		sellLabel          = helpers.TranslateV2("sell", langCode, map[string]string{})
	)

	// get member transaction record
	var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
	arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " prd_master.code IN(?)", CondValue: "MTD"},
		models.WhereCondFn{Condition: " ent_member_trading_transaction.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "SELL"},
	)

	if param.Strategy != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: param.Strategy},
		)
	}

	if param.CryptoPair != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " ent_member_trading_transaction.crypto_pair = ?", CondValue: param.CryptoPair},
		)
	}

	if param.DateFrom != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " date(ent_member_trading_transaction.timestamp) >= ?", CondValue: param.DateFrom},
		)
	}

	if param.DateTo != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " date(ent_member_trading_transaction.timestamp) <= ?", CondValue: param.DateTo},
		)
	}

	var arrEntMemberTradingTransaction, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
	if err != nil {
		base.LogErrorLog("tradingService:GetMemberReverseMartingaleTransaction():GetEntMemberTradingTransactionFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn}, false)
		return nil, "something_went_wrong"
	}

	var arrListingData = []interface{}{}

	if len(arrEntMemberTradingTransaction) > 0 {
		for _, arrEntMemberTradingTransactionV := range arrEntMemberTradingTransaction {
			var details = []interface{}{}
			var tradeCurrencyCode = strings.Replace(arrEntMemberTradingTransactionV.CryptoPair, "USDT", "", -1)

			var (
				profit             float64 = 0
				profitValue        string  = ""
				orderProgress      float64 = 0
				tQty                       = arrEntMemberTradingTransactionV.TQty
				tPrice                     = arrEntMemberTradingTransactionV.TPrice
				tQuoteQty                  = arrEntMemberTradingTransactionV.TQuoteQty
				tCommission                = arrEntMemberTradingTransactionV.TCommission
				autoTradingApiIcon         = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/Binance_Logo_x3.png"
			)

			if arrEntMemberTradingTransactionV.Platform == "KC" {
				autoTradingApiIcon = "https://media02.securelayers.cloud/medias/GTA/TRADING/API/KUCoin.png"
			}

			// get sold transaction if got
			var arrEntMemberTradingTransactionFn2 = make([]models.WhereCondFn, 0)
			arrEntMemberTradingTransactionFn2 = append(arrEntMemberTradingTransactionFn2,
				models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
				models.WhereCondFn{Condition: " sls_master.doc_no = ? ", CondValue: arrEntMemberTradingTransactionV.DocNo},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.status = ?", CondValue: "A"},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "BUY"},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.remark2 = ?", CondValue: arrEntMemberTradingTransactionV.Remark2},
				// models.WhereCondFn{Condition: " ent_member_trading_transaction.timestamp >= ?", CondValue: arrEntMemberTradingTransactionV.Timestamp},
			)
			var arrEntMemberTradingTransaction2, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn2, "ASC", false)
			if err != nil {
				base.LogErrorLog("tradingService:GetMemberReverseMartingaleTransaction():GetEntMemberTradingTransactionFn():2", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn2}, true)
				return nil, "something_went_wrong"
			}

			if len(arrEntMemberTradingTransaction2) > 0 {
				tQty = arrEntMemberTradingTransaction2[0].TQty
				tPrice = arrEntMemberTradingTransaction2[0].TPrice

				if arrEntMemberTradingTransaction2[0].Platform == "KC" {
					tPrice = float.Div(arrEntMemberTradingTransaction2[0].TQuoteQty, arrEntMemberTradingTransaction2[0].TQty)
				}

				tQuoteQty = arrEntMemberTradingTransaction2[0].TQuoteQty
				tCommission = arrEntMemberTradingTransaction2[0].TCommission

				// get profit
				var arrTblqBonusStrategyProfitFn = make([]models.WhereCondFn, 0)
				arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
					models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.doc_no = ?", CondValue: arrEntMemberTradingTransactionV.DocNo},
					models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.order_id = ?", CondValue: arrEntMemberTradingTransaction2[0].OrderId},
					models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.member_id = ?", CondValue: memID},
				)

				var arrTblqBonusStrategyProfit, err = models.GetTblqBonusStrategyProfitFn(arrTblqBonusStrategyProfitFn, false)
				if err != nil {
					base.LogErrorLog("tradingService:GetMemberReverseMartingaleTransaction():GetTblqBonusStrategyProfitFn():1", err.Error(), map[string]interface{}{"condition": arrTblqBonusStrategyProfitFn}, true)
					return nil, "something_went_wrong"
				}
				if len(arrTblqBonusStrategyProfit) > 0 {
					for _, arrTblqBonusStrategyProfitV := range arrTblqBonusStrategyProfit {
						profit += arrTblqBonusStrategyProfitV.FProfit
					}
				}
			}

			// get order progress
			orderDocNo := arrEntMemberTradingTransactionV.RefNo
			if orderDocNo == "" {
				orderDocNo = arrEntMemberTradingTransactionV.DocNo
			}
			var arrEntMemberTradingTransactionFn3 = make([]models.WhereCondFn, 0)
			arrEntMemberTradingTransactionFn3 = append(arrEntMemberTradingTransactionFn3,
				models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.id <= ?", CondValue: arrEntMemberTradingTransactionV.ID},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_no like ?", CondValue: fmt.Sprintf("%s%%", orderDocNo)},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.type = ?", CondValue: "SELL"},
			)

			arrEntMemberTradingTransaction3, err := models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn3, "", false)
			if err != nil {
				base.LogErrorLog("tradingService:GetMemberReverseMartingaleTransaction():GetEntMemberTradingTransactionFn():3", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn3}, true)
				return nil, "something_went_wrong"
			}
			orderProgress = float64(len(arrEntMemberTradingTransaction3))

			profitValue = "-"
			progress := sellLabel
			if profit != 0 {
				profitValue = helpers.CutOffDecimalv2(profit, 6, ".", ",", true)
				progress = buyLabel
			}

			details = append(details,
				// map[string]string{
				// 	"label": fmt.Sprintf("%s (%s)", orderQtyLabel, currencyCode),
				// 	"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TotalBv, 2, ".", ",", true),
				// },
				// map[string]string{
				// 	"label": tradePriceLabel,
				// 	"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TPrice, 2, ".", ",", true),
				// },
				// map[string]string{
				// 	"label": fmt.Sprintf("%s (%s)", tradeQuoteQtyLabel, tradeCurrencyCode),
				// 	"value": helpers.CutOffDecimalv2(tQty, 2, ".", ",", true),
				// },
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", orderQtyLabel, currencyCode),
					"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.TotalBv, 2, ".", ",", true),
				},
				map[string]string{
					"label": orderPriceLabel,
					"value": helpers.CutOffDecimalv2(arrEntMemberTradingTransactionV.Price, 2, ".", ",", true),
				},
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", tradeQtyLabel, currencyCode),
					"value": helpers.CutOffDecimalv2(tQuoteQty, 8, ".", ",", true),
				},
				map[string]string{
					"label": tradePriceLabel,
					"value": helpers.CutOffDecimalv2(tPrice, 2, ".", ",", true),
				},
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", tradeQuoteQtyLabel, tradeCurrencyCode),
					"value": helpers.CutOffDecimalv2(tQty, 8, ".", ",", true),
				},
				map[string]string{
					"label": fmt.Sprintf("%s (%s)", adminFeeLabel, currencyCode),
					"value": helpers.CutOffDecimalv2(tCommission, 8, ".", ",", true),
				},
				map[string]string{
					"label": profitLabel,
					"value": profitValue,
				},
				map[string]string{
					"label": progressLabel,
					"value": progress,
				},
				map[string]string{
					"label": orderProgressLabel,
					"value": helpers.CutOffDecimalv2(orderProgress, 0, ".", ",", true),
				},
			)

			var cryptoPair = arrEntMemberTradingTransactionV.CryptoPairName
			if cryptoPair == "" {
				cryptoPair = strings.Replace(arrEntMemberTradingTransactionV.CryptoPair, "USDT", "/USDT", 1)
				cryptoPair = strings.Replace(cryptoPair, "USDTM", "USDT", 1)
				cryptoPair = strings.Replace(cryptoPair, "XBT", "BTC", 1)
			}

			createdAt := arrEntMemberTradingTransactionV.Timestamp.Format("2006-01-02 15:04:05")
			if arrEntMemberTradingTransactionV.TTime != 0 {
				createdAt = time.Unix(0, arrEntMemberTradingTransactionV.TTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
			}
			arrListingData = append(arrListingData,
				TradingTransactionStruct{
					Strategy:           arrEntMemberTradingTransactionV.StrategyName,
					CryptoPair:         cryptoPair,
					OrderNo:            arrEntMemberTradingTransactionV.DocNo,
					AutoTradingApiIcon: autoTradingApiIcon,
					CreatedAt:          createdAt,
					Details:            details,
				},
			)
		}
	}

	return arrListingData, ""
}

func GetMemberAutoTradingProfit(memID int, dataType, strategy, cryptoPair, dateFrom, dateTo, downlineUsername string, page int64, langCode string) (interface{}, string) {
	var (
		arrProfitSummary = map[string]interface{}{}
		arrProfitDetails = []interface{}{}
	)

	if dataType == "PROFIT" {
		var (
			todayProfit              float64
			accumulatedProfit        float64
			totalFundsUtilized       float64
			arrCFRADocNo             []string // get from order's princple value
			arrMTDocNo               []string // grouped total quote_qty, type = BUY
			arrMTDDocNo              []string // grouped total quote_qty, type = SELL
			arrSGTDocNo              []string
			totalSpotFundsUtilized   float64 = 0
			totalFutureFundsUtilized float64 = 0
		)

		// get profit details list
		var arrTblqBonusStrategyProfitFn = make([]models.WhereCondFn, 0)
		arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
			models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.member_id = ?", CondValue: memID},
		)

		if strategy != "" {
			arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
				models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: strategy},
			)
		}

		if cryptoPair != "" {
			arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
				models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.crypto_pair LIKE '" + cryptoPair + "%' AND 1=?", CondValue: 1},
			)
		}

		if dateFrom != "" {
			arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
				models.WhereCondFn{Condition: " DATE(tblq_bonus_strategy_profit.bns_id) >= ?", CondValue: dateFrom},
			)
		}

		if dateTo != "" {
			arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
				models.WhereCondFn{Condition: " DATE(tblq_bonus_strategy_profit.bns_id) <= ?", CondValue: dateTo},
			)
		}

		var arrTblqBonusStrategyProfit, err = models.GetTblqBonusStrategyProfitFn(arrTblqBonusStrategyProfitFn, false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingProfit():GetTblqBonusStrategyProfitFn():1", err.Error(), map[string]interface{}{"condition": arrTblqBonusStrategyProfitFn}, true)
			return nil, "something_went_wrong"
		}
		if len(arrTblqBonusStrategyProfit) > 0 {
			var (
				cfraLabel  = helpers.TranslateV2("Crypto Funding Rates Arbitrage", langCode, map[string]string{})
				cifraLabel = helpers.TranslateV2("Crypto Index Funding Rates Arbitrage", langCode, map[string]string{})
				sgtLabel   = helpers.TranslateV2("Spot Grid Trading", langCode, map[string]string{})
				mtLabel    = helpers.TranslateV2("Martingale Trading", langCode, map[string]string{})
				mtdLabel   = helpers.TranslateV2("Martingale Trading Reverse", langCode, map[string]string{})

				profitLabel       = helpers.TranslateV2("profit", langCode, map[string]string{})
				buyQuantityLabel  = helpers.TranslateV2("buy_quantity", langCode, map[string]string{})
				buyPriceLabel     = helpers.TranslateV2("buy_price", langCode, map[string]string{})
				totalAmountLabel  = helpers.TranslateV2("total_amount", langCode, map[string]string{})
				sellQuantityLabel = helpers.TranslateV2("sell_quantity", langCode, map[string]string{})
				sellPriceLabel    = helpers.TranslateV2("sell_price", langCode, map[string]string{})
				dateLabel         = helpers.TranslateV2("date", langCode, map[string]string{})
				curDate           = time.Now().Format("2006-01-02")
			)

			for _, arrTblqBonusStrategyProfitV := range arrTblqBonusStrategyProfit {
				var profit = arrTblqBonusStrategyProfitV.FProfit
				var profitDetails = []map[string]interface{}{}

				profitDetails = append(profitDetails,
					map[string]interface{}{
						"label": profitLabel,
						"value": helpers.CutOffDecimalv2(profit, 8, ".", ",", true),
					},
				)

				str := strings.Split(arrTblqBonusStrategyProfitV.DocNo, "-")
				if arrTblqBonusStrategyProfitV.Strategy == "CFRA" || arrTblqBonusStrategyProfitV.Strategy == "CIFRA" {
					if !helpers.StringInSlice(str[0], arrCFRADocNo) {
						arrCFRADocNo = append(arrCFRADocNo, str[0])
						totalFutureFundsUtilized += arrTblqBonusStrategyProfitV.PrincipleValue // direct grab from order.total_amount
					}
				} else if arrTblqBonusStrategyProfitV.Strategy == "MT" {
					if !helpers.StringInSlice(str[0], arrMTDocNo) {
						arrMTDocNo = append(arrMTDocNo, str[0])
					}
				} else if arrTblqBonusStrategyProfitV.Strategy == "MTD" {
					if !helpers.StringInSlice(str[0], arrMTDDocNo) {
						arrMTDDocNo = append(arrMTDDocNo, str[0])
					}
				} else if arrTblqBonusStrategyProfitV.Strategy == "SGT" {
					if !helpers.StringInSlice(str[0], arrSGTDocNo) {
						arrSGTDocNo = append(arrSGTDocNo, str[0])
					}
				}

				// get sell details
				var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
				arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
					models.WhereCondFn{Condition: " ent_member_trading_transaction.order_id = ?", CondValue: arrTblqBonusStrategyProfitV.OrderId},
				)
				var arrEntMemberTradingTransaction, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
				if err != nil {
					base.LogErrorLog("tradingService:GetMemberAutoTradingProfit():GetEntMemberTradingTransactionFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn}, true)
					return nil, "something_went_wrong"
				}
				// if !strings.Contains(arrTblqBonusStrategyProfitV.DocNo, "-") {
				// 	if len(arrEntMemberTradingTransaction) > 0 {
				// 		totalFundsUtilized += arrEntMemberTradingTransaction[0].TQuoteQty
				// 	}
				// }

				if arrTblqBonusStrategyProfitV.Strategy == "MT" || arrTblqBonusStrategyProfitV.Strategy == "MTD" || arrTblqBonusStrategyProfitV.Strategy == "SGT" {
					var (
						buyQuantity       = "-"
						buyPrice          = "-"
						totalAmount       = "-"
						sellQuantity      = "-"
						sellPrice         = "-"
						soldTransactionId = 0
					)

					if len(arrEntMemberTradingTransaction) > 0 {
						sellQuantity = helpers.CutOffDecimalv2(arrEntMemberTradingTransaction[0].Num, 8, ".", ",", true)
						sellPrice = helpers.CutOffDecimalv2(arrEntMemberTradingTransaction[0].Price, 2, ".", ",", true)
						totalAmount = helpers.CutOffDecimalv2(arrEntMemberTradingTransaction[0].TotalBv, 2, ".", ",", true)
						soldTransactionId = arrEntMemberTradingTransaction[0].ID

						// get buy details
						var (
							totalBuyQuantity = 0.00
							totalBuyPrice    = 0.00
							tQuoteQty        = 0.00
						)

						// if arrTblqBonusStrategyProfitV.Strategy == "SGT" && arrTblqBonusStrategyProfitV.Platform == "KC" {
						if arrTblqBonusStrategyProfitV.Strategy == "SGT" {
							var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
							arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
								models.WhereCondFn{Condition: " ent_member_trading_transaction.id != ?", CondValue: soldTransactionId},
								models.WhereCondFn{Condition: " ent_member_trading_transaction.remark2 = ?", CondValue: soldTransactionId},
							)
							var arrEntMemberTradingTransaction, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
							if err != nil {
								base.LogErrorLog("tradingService:GetMemberAutoTradingProfit():GetEntMemberTradingTransactionFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn}, true)
								return nil, "something_went_wrong"
							}

							if len(arrEntMemberTradingTransaction) > 0 {
								totalBuyQuantity = arrEntMemberTradingTransaction[0].Num
								totalBuyPrice = arrEntMemberTradingTransaction[0].Price
							}
						} else {
							var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
							arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
								models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_no = ?", CondValue: arrTblqBonusStrategyProfitV.DocNo},
								models.WhereCondFn{Condition: " ent_member_trading_transaction.id < ?", CondValue: soldTransactionId},
							)
							var arrEntMemberTradingTransaction, err = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
							if err != nil {
								base.LogErrorLog("tradingService:GetMemberAutoTradingProfit():GetEntMemberTradingTransactionFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingTransactionFn}, true)
								return nil, "something_went_wrong"
							}

							for _, arrEntMemberTradingTransactionV := range arrEntMemberTradingTransaction {
								if arrEntMemberTradingTransactionV.Type == "SELL" {
									break
								}

								totalBuyQuantity += arrEntMemberTradingTransactionV.Num
								totalBuyPrice += arrEntMemberTradingTransactionV.Price
								tQuoteQty += arrEntMemberTradingTransactionV.TQuoteQty
							}
						}

						buyQuantity = helpers.CutOffDecimalv2(totalBuyQuantity, 8, ".", ",", true)
						// buyPrice = helpers.CutOffDecimalv2(totalBuyPrice, 2, ".", ",", true)
						buyPrice = helpers.CutOffDecimalv2(totalBuyPrice, 2, ".", ",", true)
						if tQuoteQty > 0 {
							buyPrice = helpers.CutOffDecimalv2(float.Div(tQuoteQty, totalBuyQuantity), 2, ".", ",", true)
						}
					}

					profitDetails = append(profitDetails,
						map[string]interface{}{
							"label": buyQuantityLabel,
							"value": buyQuantity,
						},
						map[string]interface{}{
							"label": buyPriceLabel,
							"value": buyPrice,
						},
						map[string]interface{}{
							"label": totalAmountLabel,
							"value": totalAmount,
						},
						map[string]interface{}{
							"label": sellQuantityLabel,
							"value": sellQuantity,
						},
						map[string]interface{}{
							"label": sellPriceLabel,
							"value": sellPrice,
						},
					)
				}

				profitDetails = append(profitDetails,
					map[string]interface{}{
						"label": dateLabel,
						"value": arrTblqBonusStrategyProfitV.DtTimestamp.Format("2006-01-02 15:04:05"),
					},
				)

				cryptoPairName := arrTblqBonusStrategyProfitV.CryptoPairName
				if cryptoPairName == "" {
					cryptoPairName = strings.Replace(arrTblqBonusStrategyProfitV.CryptoPair, "USDTM", "USDT", -1)
					cryptoPairName = strings.Replace(cryptoPairName, "USDT", "/USDT", -1)
				}

				strategyName := ""
				if arrTblqBonusStrategyProfitV.Strategy == "CFRA" {
					strategyName = cfraLabel
				} else if arrTblqBonusStrategyProfitV.Strategy == "CIFRA" {
					strategyName = cifraLabel
				} else if arrTblqBonusStrategyProfitV.Strategy == "SGT" {
					strategyName = sgtLabel
				} else if arrTblqBonusStrategyProfitV.Strategy == "MT" {
					strategyName = mtLabel
				} else if arrTblqBonusStrategyProfitV.Strategy == "MTD" {
					strategyName = mtdLabel
				}

				arrProfitDetails = append(arrProfitDetails, map[string]interface{}{
					"strategy": strategyName,
					// "strategy":   helpers.TranslateV2(arrTblqBonusStrategyProfitV.StrategyName, langCode, map[string]string{}),
					"order_no":   arrTblqBonusStrategyProfitV.DocNo,
					"token_sold": cryptoPairName,
					"profit":     helpers.CutOffDecimalv2(profit, 8, ".", ",", true),
					"details":    profitDetails,
				})

				// get today profit
				if arrTblqBonusStrategyProfitV.BnsID == curDate {
					todayProfit += profit
				}

				// get accumulated profit
				accumulatedProfit += profit
			}
		}

		// total funds utilized cannot exceed current member total wallet limit
		arrEntMemberTradingWalletLimitFn := make([]models.WhereCondFn, 0)
		arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
			models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.member_id = ?", CondValue: memID},
			// models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.status = ?", CondValue: "A"},
		)

		if strategy == "SGT" || strategy == "MT" {
			arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
				models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.module = ?", CondValue: "SPOT"},
			)
		} else if strategy == "CFRA" || strategy == "CIFRA" || strategy == "MTD" {
			arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
				models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.module = ?", CondValue: "FUTURE"},
			)
		}

		// if dateFrom != "" {
		// 	arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
		// 		models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.created_at >= ?", CondValue: dateFrom},
		// 	)
		// }

		if dateTo != "" {
			arrEntMemberTradingWalletLimitFn = append(arrEntMemberTradingWalletLimitFn,
				models.WhereCondFn{Condition: "ent_member_trading_wallet_limit.created_at <= ?", CondValue: dateTo},
			)
		}

		arrEntMemberTradingWalletLimit, err := models.GetEntMemberTradingWalletLimit(arrEntMemberTradingWalletLimitFn, "", false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingProfit():GetEntMemberTradingWalletLimit():1", err.Error(), map[string]interface{}{"condition": arrEntMemberTradingWalletLimitFn}, true)
			return nil, "something_went_wrong"
		}

		var totalSpotWalletLimit float64 = 0
		var totalFutureWalletLimit float64 = 0
		for _, walletLimit := range arrEntMemberTradingWalletLimit {
			if walletLimit.Module == "FUTURE" {
				totalFutureWalletLimit += walletLimit.TotalAmount
			} else {
				totalSpotWalletLimit += walletLimit.TotalAmount
			}
		}

		// get total funds utilized
		// if len(arrEntMemberTradingTransaction) > 0 {
		// 	if arrEntMemberTradingTransaction[0].Strategy == "SGT" || arrEntMemberTradingTransaction[0].Strategy == "MT" {
		// 		totalSpotFundsUtilized += arrEntMemberTradingTransaction[0].TQuoteQty
		// 	} else if arrEntMemberTradingTransaction[0].Strategy == "CFRA" || arrEntMemberTradingTransaction[0].Strategy == "CIFRA" || arrEntMemberTradingTransaction[0].Strategy == "MTD" {
		// 		totalFutureFundsUtilized += arrEntMemberTradingTransaction[0].TQuoteQty
		// 	}
		// }

		if len(arrMTDocNo) > 0 {
			for _, docNo := range arrMTDocNo {
				totalSpotFundsUtilized += GetMartingaleTradingFundsUtilized(docNo, dateFrom, dateTo)
			}
		}

		if len(arrMTDDocNo) > 0 {
			for _, docNo := range arrMTDDocNo {
				totalFutureFundsUtilized += GetReverseMartingaleTradingFundsUtilized(docNo, dateFrom, dateTo)
			}
		}

		if len(arrSGTDocNo) > 0 {
			// to be enhance
			// for _, docNo := range arrSGTDocNo {
			// 	totalSpotFundsUtilized += GetSpotGridTradingFundsUtilized(docNo, dateFrom, dateTo)
			// }
		}

		if len(arrEntMemberTradingWalletLimit) <= 0 {
			totalFundsUtilized = 0.00
		} else {
			if totalSpotWalletLimit < totalSpotFundsUtilized {
				totalSpotFundsUtilized = totalSpotWalletLimit
			}

			if totalFutureWalletLimit < totalFutureFundsUtilized {
				totalFutureFundsUtilized = totalFutureWalletLimit
			}

			totalFundsUtilized = totalFutureFundsUtilized + totalSpotFundsUtilized
		}

		// get profit summary
		arrProfitSummary["today_profit"] = helpers.CutOffDecimalv2(todayProfit, 2, ".", ",", true)
		arrProfitSummary["accumulated_profit"] = helpers.CutOffDecimalv2(accumulatedProfit, 2, ".", ",", true)
		arrProfitSummary["total_funds_utilized"] = helpers.CutOffDecimalv2(totalFundsUtilized, 2, ".", ",", true)

		totalProfitPercentage := 0.00
		if totalFundsUtilized != 0 {
			totalProfitPercentage = float.Mul(float.Div(accumulatedProfit, totalFundsUtilized), 100)
		}
		arrProfitSummary["total_profit_percentage"] = helpers.CutOffDecimalv2(totalProfitPercentage, 6, ".", ",", true) + "%"

	} else if dataType == "BONUS" {
		var (
			todayBonus       float64
			accumulatedBonus float64
		)

		// get bonus details list
		var arrTblqBonusStrategySponsorFn = make([]models.WhereCondFn, 0)
		arrTblqBonusStrategySponsorFn = append(arrTblqBonusStrategySponsorFn,
			models.WhereCondFn{Condition: " tblq_bonus_strategy_sponsor.member_id = ?", CondValue: memID},
		)

		if strategy != "" {
			arrTblqBonusStrategySponsorFn = append(arrTblqBonusStrategySponsorFn,
				models.WhereCondFn{Condition: " prd_master.code = ?", CondValue: strategy},
			)
		}

		if cryptoPair != "" {
			arrTblqBonusStrategySponsorFn = append(arrTblqBonusStrategySponsorFn,
				models.WhereCondFn{Condition: " sls_master_bot_setting.crypto_pair LIKE '" + cryptoPair + "%' AND 1=?", CondValue: 1},
			)
		}

		if dateFrom != "" {
			arrTblqBonusStrategySponsorFn = append(arrTblqBonusStrategySponsorFn,
				models.WhereCondFn{Condition: " DATE(tblq_bonus_strategy_sponsor.bns_id) >= ?", CondValue: dateFrom},
			)
		}

		if dateTo != "" {
			arrTblqBonusStrategySponsorFn = append(arrTblqBonusStrategySponsorFn,
				models.WhereCondFn{Condition: " DATE(tblq_bonus_strategy_sponsor.bns_id) <= ?", CondValue: dateTo},
			)
		}

		if downlineUsername != "" {
			arrTblqBonusStrategySponsorFn = append(arrTblqBonusStrategySponsorFn,
				models.WhereCondFn{Condition: " ent_member.nick_name = ?", CondValue: downlineUsername},
			)
		}

		var arrTblqBonusStrategySponsor, err = models.GetTblqBonusStrategySponsorFn(arrTblqBonusStrategySponsorFn, false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingProfit():GetTblqBonusStrategySponsorFn():1", err.Error(), map[string]interface{}{"condition": arrTblqBonusStrategySponsorFn}, true)
			return nil, "something_went_wrong"
		}
		if len(arrTblqBonusStrategySponsor) > 0 {
			var (
				dateLabel               = helpers.TranslateV2("date", langCode, map[string]string{})
				downlineUsernameLabel   = helpers.TranslateV2("downline_username", langCode, map[string]string{})
				levelLabel              = helpers.TranslateV2("level", langCode, map[string]string{})
				referralUserProfitLabel = helpers.TranslateV2("referral_user_profit", langCode, map[string]string{})
			)

			for _, arrTblqBonusStrategySponsorV := range arrTblqBonusStrategySponsor {
				var bonus = arrTblqBonusStrategySponsorV.FBns
				var bonusDetails = []map[string]interface{}{}

				bonusDetails = append(bonusDetails,
					map[string]interface{}{
						"label": dateLabel,
						"value": arrTblqBonusStrategySponsorV.BnsID,
					},
					map[string]interface{}{
						"label": downlineUsernameLabel,
						"value": arrTblqBonusStrategySponsorV.DownlineNickName,
					},
					map[string]interface{}{
						"label": levelLabel,
						"value": helpers.TranslateV2("level_:0_:1", langCode, map[string]string{"0": strconv.Itoa(arrTblqBonusStrategySponsorV.ILvl), "1": fmt.Sprintf("%.0f", float.Mul(arrTblqBonusStrategySponsorV.FPerc, 100))}),
					},
					map[string]interface{}{
						"label": referralUserProfitLabel,
						"value": helpers.CutOffDecimalv2(bonus, 8, ".", ",", true),
					},
				)

				cryptoPairName := arrTblqBonusStrategySponsorV.CryptoPairName
				if cryptoPairName == "" {
					cryptoPairName = strings.Replace(arrTblqBonusStrategySponsorV.CryptoPair, "USDTM", "USDT", -1)
					cryptoPairName = strings.Replace(cryptoPairName, "USDT", "/USDT", -1)
				}

				arrProfitDetails = append(arrProfitDetails,
					map[string]interface{}{
						"order_no":   arrTblqBonusStrategySponsorV.DocNo,
						"token_sold": cryptoPairName,
						"details":    bonusDetails,
					},
				)

				// get today bonus
				if arrTblqBonusStrategySponsorV.BnsID == time.Now().Format("2006-01-02") {
					todayBonus += bonus
				}

				// get accumulated bonus
				accumulatedBonus += bonus
			}
		}

		// get profit summary
		arrProfitSummary["today_bonus"] = helpers.CutOffDecimalv2(todayBonus, 2, ".", ",", true)
		arrProfitSummary["accumulated_bonus"] = helpers.CutOffDecimalv2(accumulatedBonus, 2, ".", ",", true)
	} else {
		return nil, "invalid_type"
	}

	pagination := base.Pagination{
		Page:      page,
		DataArr:   arrProfitDetails,
		HeaderArr: arrProfitSummary,
	}
	arrProfitDetailsPaginated := pagination.PaginationInterfaceV1()

	return arrProfitDetailsPaginated, ""
}

func GetMartingaleTradingFundsUtilized(docNo, dateFrom, dateTo string) float64 {
	var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
	if dateFrom != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " doc_date >= ?", CondValue: dateFrom},
		)
	}

	if dateTo != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " doc_date <= ?", CondValue: dateTo},
		)
	}

	arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
		models.WhereCondFn{Condition: " doc_no like ?", CondValue: fmt.Sprintf("%v%%", docNo)},
		models.WhereCondFn{Condition: " type like ?", CondValue: "BUY"},
	)
	var arrEntMemberTradingTransaction, _ = models.GetTradingMaxQuoteQtyInGroupFn(arrEntMemberTradingTransactionFn, "", false)

	if len(arrEntMemberTradingTransaction) > 0 {
		return arrEntMemberTradingTransaction[0].SumQuoteQty
	}

	return 0.00
}

func GetReverseMartingaleTradingFundsUtilized(docNo, dateFrom, dateTo string) float64 {
	var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
	if dateFrom != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " doc_date >= ?", CondValue: dateFrom},
		)
	}

	if dateTo != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " doc_date <= ?", CondValue: dateTo},
		)
	}

	arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
		models.WhereCondFn{Condition: " doc_no like ?", CondValue: fmt.Sprintf("%v%%", docNo)},
		models.WhereCondFn{Condition: " type like ?", CondValue: "SELL"},
	)
	var arrEntMemberTradingTransaction, _ = models.GetTradingMaxQuoteQtyInGroupFn(arrEntMemberTradingTransactionFn, "", false)

	if len(arrEntMemberTradingTransaction) > 0 {
		return arrEntMemberTradingTransaction[0].SumQuoteQty
	}

	return 0.00
}

func GetSpotGridTradingFundsUtilized(docNo, dateFrom, dateTo string) float64 {
	var maxPosition float64 = 0 // 最大持仓

	// get all sell transaction first
	var arrEntMemberTradingDocDateFn = make([]models.WhereCondFn, 0)
	if dateFrom != "" {
		arrEntMemberTradingDocDateFn = append(arrEntMemberTradingDocDateFn,
			models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date >= ?", CondValue: dateFrom},
		)
	}

	if dateTo != "" {
		arrEntMemberTradingDocDateFn = append(arrEntMemberTradingDocDateFn,
			models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date <= ?", CondValue: dateTo},
		)
	}

	arrEntMemberTradingDocDateFn = append(arrEntMemberTradingDocDateFn,
		models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_no like ?", CondValue: docNo},
	)
	var arrEntMemberTradingTransaction, _ = models.GetEntMemberTradingDocDateFn(arrEntMemberTradingDocDateFn, false)
	if len(arrEntMemberTradingTransaction) > 0 {
		for _, arrEntMemberTradingTransactionV := range arrEntMemberTradingTransaction {
			curPosition := GetSpotGridTradingPosition(docNo, dateFrom, arrEntMemberTradingTransactionV.DocDate)

			if curPosition > maxPosition {
				maxPosition = curPosition
			}

		}
	}

	return maxPosition
}

func GetSpotGridTradingPosition(docNo, dateFrom, dateTo string) float64 {
	// get all sell transaction first
	var arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
	if dateFrom != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date >= ?", CondValue: dateFrom},
		)
	}

	if dateTo != "" {
		arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
			models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date <= ?", CondValue: dateTo},
		)
	}

	arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
		models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_no like ?", CondValue: docNo},
		models.WhereCondFn{Condition: " ent_member_trading_transaction.type like ?", CondValue: "SELL"},
	)
	var arrEntMemberTradingTransaction, _ = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "ASC", false)
	if len(arrEntMemberTradingTransaction) > 0 {
		var matchedID = []int{}
		for _, arrEntMemberTradingTransactionV := range arrEntMemberTradingTransaction {
			arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
			arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
				models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_no like ?", CondValue: docNo},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.timestamp < ?", CondValue: arrEntMemberTradingTransactionV.Timestamp},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.type like ?", CondValue: "BUY"},
			)

			if dateFrom != "" {
				arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
					models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date >= ?", CondValue: dateFrom},
				)
			}

			if dateTo != "" {
				arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
					models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date <= ?", CondValue: dateTo},
				)
			}

			if len(matchedID) > 0 {
				arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
					models.WhereCondFn{Condition: " ent_member_trading_transaction.id not in(?)", CondValue: matchedID},
				)
			}

			var arrEntMemberTradingTransaction2, _ = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "DESC", false)
			if len(arrEntMemberTradingTransaction2) > 0 {
				matchedID = append(matchedID, arrEntMemberTradingTransaction2[0].ID)
			}
		}

		// find those 持仓
		if len(matchedID) > 0 {
			arrEntMemberTradingTransactionFn = make([]models.WhereCondFn, 0)
			if dateFrom != "" {
				arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
					models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date >= ?", CondValue: dateFrom},
				)
			}

			if dateTo != "" {
				arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
					models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_date <= ?", CondValue: dateTo},
				)
			}

			arrEntMemberTradingTransactionFn = append(arrEntMemberTradingTransactionFn,
				models.WhereCondFn{Condition: " ent_member_trading_transaction.doc_no like ?", CondValue: docNo},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.type like ?", CondValue: "BUY"},
				models.WhereCondFn{Condition: " ent_member_trading_transaction.id not in(?)", CondValue: matchedID},
			)
			var arrEntMemberTradingTransaction3, _ = models.GetEntMemberTradingTransactionFn(arrEntMemberTradingTransactionFn, "", false)
			if len(arrEntMemberTradingTransaction3) > 0 {
				var totalFundUtilized float64
				for _, arrEntMemberTradingTransaction3V := range arrEntMemberTradingTransaction3 {
					totalFundUtilized += arrEntMemberTradingTransaction3V.TQuoteQty
				}

				return totalFundUtilized
			}
		}
	}

	return 0.00
}

type MemberAutoTradingProfitGraph struct {
	Strategy     string  `json:"strategy"`
	StrategyCode string  `json:"strategy_code"`
	Value        float64 `json:"value"`
	ValueStr     string  `json:"value_str"`
	ColorCode    string  `json:"color_code"`
}

func GetMemberAutoTradingProfitGraph(memID int, profitType string, dataType, langCode string) ([]MemberAutoTradingProfitGraph, string) {
	var (
		data                     = []MemberAutoTradingProfitGraph{}
		fromDate          string = ""
		toDate            string = ""
		curTime                  = time.Now()
		graphValueStorage struct {
			CFRA  float64
			CIFRA float64
			SGT   float64
			MT    float64
			MTD   float64
		}
	)

	// determine from and to date by dataType
	if dataType == "DAILY" {
		fromDate = curTime.Format("2006-01-02")
		toDate = curTime.Format("2006-01-02")
	} else if dataType == "WEEKLY" {
		fromDate = helpers.WeekStartDate(curTime).Format("2006-01-02")
		toDate = helpers.WeekEndDate(curTime).Format("2006-01-02")
	} else if dataType == "MONTHLY" {
		currentYear, currentMonth, _ := curTime.Date()
		currentLocation := curTime.Location()

		fromDate = time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation).Format("2006-01-02")
		toDate = time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation).AddDate(0, 1, -1).Format("2006-01-02")
	} else if dataType == "YEARLY" {
		currentYear, _, _ := curTime.Date()
		currentLocation := curTime.Location()

		fromDate = time.Date(currentYear, 1, 1, 0, 0, 0, 0, currentLocation).Format("2006-01-02")
		toDate = time.Date(currentYear, 1, 1, 0, 0, 0, 0, currentLocation).AddDate(1, 0, -1).Format("2006-01-02")
	} else {
		return nil, "invalid_data_type"
	}

	// fmt.Println("fromDate:", fromDate, "toDate:", toDate)

	if profitType == "PROFIT" {
		// get profit details list
		var arrTblqBonusStrategyProfitFn = make([]models.WhereCondFn, 0)
		arrTblqBonusStrategyProfitFn = append(arrTblqBonusStrategyProfitFn,
			models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.member_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.bns_id >= ?", CondValue: fromDate},
			models.WhereCondFn{Condition: " tblq_bonus_strategy_profit.bns_id <= ?", CondValue: toDate},
		)

		var arrTblqBonusStrategyProfit, err = models.GetTblqBonusStrategyProfitFn(arrTblqBonusStrategyProfitFn, false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingProfitGraph():GetTblqBonusStrategyProfitFn():1", err.Error(), map[string]interface{}{"condition": arrTblqBonusStrategyProfitFn}, true)
			return nil, "something_went_wrong"
		}
		if len(arrTblqBonusStrategyProfit) > 0 {
			for _, arrTblqBonusStrategyProfitV := range arrTblqBonusStrategyProfit {
				if arrTblqBonusStrategyProfitV.Strategy == "CFRA" {
					graphValueStorage.CFRA += arrTblqBonusStrategyProfitV.FProfit
				} else if arrTblqBonusStrategyProfitV.Strategy == "CIFRA" {
					graphValueStorage.CIFRA += arrTblqBonusStrategyProfitV.FProfit
				} else if arrTblqBonusStrategyProfitV.Strategy == "SGT" {
					graphValueStorage.SGT += arrTblqBonusStrategyProfitV.FProfit
				} else if arrTblqBonusStrategyProfitV.Strategy == "MT" {
					graphValueStorage.MT += arrTblqBonusStrategyProfitV.FProfit
				} else if arrTblqBonusStrategyProfitV.Strategy == "MTD" {
					graphValueStorage.MTD += arrTblqBonusStrategyProfitV.FProfit
				}
			}
		}
	} else if profitType == "BONUS" {
		// get bonus details list
		var arrTblqBonusStrategySponsorFn = make([]models.WhereCondFn, 0)
		arrTblqBonusStrategySponsorFn = append(arrTblqBonusStrategySponsorFn,
			models.WhereCondFn{Condition: " tblq_bonus_strategy_sponsor.member_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: " tblq_bonus_strategy_sponsor.bns_id >= ?", CondValue: fromDate},
			models.WhereCondFn{Condition: " tblq_bonus_strategy_sponsor.bns_id <= ?", CondValue: toDate},
		)

		var arrTblqBonusStrategySponsor, err = models.GetTblqBonusStrategySponsorFn(arrTblqBonusStrategySponsorFn, false)
		if err != nil {
			base.LogErrorLog("tradingService:GetMemberAutoTradingProfitGraph():GetTblqBonusStrategySponsorFn():1", err.Error(), map[string]interface{}{"condition": arrTblqBonusStrategySponsorFn}, true)
			return nil, "something_went_wrong"
		}
		if len(arrTblqBonusStrategySponsor) > 0 {
			for _, arrTblqBonusStrategySponsorV := range arrTblqBonusStrategySponsor {
				if arrTblqBonusStrategySponsorV.Strategy == "CFRA" {
					graphValueStorage.CFRA += arrTblqBonusStrategySponsorV.FBns
				} else if arrTblqBonusStrategySponsorV.Strategy == "CIFRA" {
					graphValueStorage.CIFRA += arrTblqBonusStrategySponsorV.FBns
				} else if arrTblqBonusStrategySponsorV.Strategy == "SGT" {
					graphValueStorage.SGT += arrTblqBonusStrategySponsorV.FBns
				} else if arrTblqBonusStrategySponsorV.Strategy == "MT" {
					graphValueStorage.MT += arrTblqBonusStrategySponsorV.FBns
				} else if arrTblqBonusStrategySponsorV.Strategy == "MTD" {
					graphValueStorage.MTD += arrTblqBonusStrategySponsorV.FBns
				}
			}
		}
	} else {
		return nil, "invalid_type"
	}

	data = append(data,
		MemberAutoTradingProfitGraph{
			Strategy:     helpers.TranslateV2("Crypto Funding Rates Arbitage", langCode, map[string]string{}),
			StrategyCode: "CFRA",
			Value:        graphValueStorage.CFRA,
			ValueStr:     helpers.CutOffDecimalv2(graphValueStorage.CFRA, 6, ".", ",", true),
			ColorCode:    helpers.AutoTradingColorCode("CFRA"),
		},
		MemberAutoTradingProfitGraph{
			Strategy:     helpers.TranslateV2("Crypto Index Funding Rates Arbitage", langCode, map[string]string{}),
			StrategyCode: "CIFRA",
			Value:        graphValueStorage.CIFRA,
			ValueStr:     helpers.CutOffDecimalv2(graphValueStorage.CIFRA, 6, ".", ",", true),
			ColorCode:    helpers.AutoTradingColorCode("CIFRA"),
		},
		MemberAutoTradingProfitGraph{
			Strategy:     helpers.TranslateV2("Spot Grid Trading", langCode, map[string]string{}),
			StrategyCode: "SGT",
			Value:        graphValueStorage.SGT,
			ValueStr:     helpers.CutOffDecimalv2(graphValueStorage.SGT, 6, ".", ",", true),
			ColorCode:    helpers.AutoTradingColorCode("SGT"),
		},
		MemberAutoTradingProfitGraph{
			Strategy:     helpers.TranslateV2("Martingale Trading", langCode, map[string]string{}),
			StrategyCode: "MT",
			Value:        graphValueStorage.MT,
			ValueStr:     helpers.CutOffDecimalv2(graphValueStorage.MT, 6, ".", ",", true),
			ColorCode:    helpers.AutoTradingColorCode("MT"),
		},
		MemberAutoTradingProfitGraph{
			Strategy:     helpers.TranslateV2("Martingale Trading Reverse", langCode, map[string]string{}),
			StrategyCode: "MTD",
			Value:        graphValueStorage.MTD,
			ValueStr:     helpers.CutOffDecimalv2(graphValueStorage.MTD, 6, ".", ",", true),
			ColorCode:    helpers.AutoTradingColorCode("MTD"),
		},
	)

	return data, ""
}

type GetMemberAutoTradingReportsParam struct {
	PoolType   string
	Strategy   string
	CryptoPair string
	DateFrom   string
	DateTo     string
	Page       int64
}

type TradingPoolStatistics struct {
	Bid        string
	Ask        string
	CryptoPair string
	Profit     string
	Timestamp  string
}

func GetMemberAutoTradingReports(memID int, param GetMemberAutoTradingReportsParam, langCode string) (interface{}, string) {
	var arrTradingPoolStatistics = []interface{}{}

	var arrCond = []models.WhereCondFn{}
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " trading_pool_statistics.b_status = ?", CondValue: 1},
	)

	if param.Strategy != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_pool_statistics.strategy = ?", CondValue: param.Strategy},
		)
	}

	if param.DateFrom != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(trading_pool_statistics.trade_at) >= ?", CondValue: param.DateFrom},
		)
	}

	if param.DateTo != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(trading_pool_statistics.trade_at) <= ?", CondValue: param.DateTo},
		)
	}

	if param.PoolType == "B1" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_pool_statistics.symbol IN(?, 'ETHUSDT', 'XRPUSDT')", CondValue: "BTCUSDT"},
		)
	} else if param.PoolType == "B2" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_pool_statistics.symbol IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT')", CondValue: "BTCUSDT"},
		)
	} else if param.PoolType == "B3" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_pool_statistics.symbol IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT', 'LTCUSDT', 'GOGEUSDT', 'DOTUSDT')", CondValue: "BTCUSDT"},
		)
	} else if param.PoolType == "B4" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_pool_statistics.symbol IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT', 'LTCUSDT', 'GOGEUSDT', 'DOTUSDT', 'BHCUSDT', 'LINKUSDT')", CondValue: "BTCUSDT"},
		)
	} else if param.PoolType == "B5" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " trading_pool_statistics.symbol IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT', 'LTCUSDT', 'GOGEUSDT', 'DOTUSDT', 'BHCUSDT', 'LINKUSDT', 'DEFIUSDT', 'BTCDOMUSDT')", CondValue: "BTCUSDT"},
		)
	}

	arrTradingPoolStatistic, _ := models.GetTradingPoolStatisticFn(arrCond, false)
	// arrTradingPoolStatisticGroup, _ := models.GetTradingPoolStatisticGroupFn(arrCond, false)

	var (
		totalDailyProfit         float64 = 0
		totalDailyProfitMember   float64 = 0
		totalDailyNumberOfTrades float64 = 0
	)

	var (
		// pool1Label = helpers.TranslateV2("b1", langCode, map[string]string{})
		// pool2Label = helpers.TranslateV2("b2", langCode, map[string]string{})
		// pool3Label = helpers.TranslateV2("b3", langCode, map[string]string{})
		// pool4Label = helpers.TranslateV2("b4", langCode, map[string]string{})
		// pool5Label = helpers.TranslateV2("b5", langCode, map[string]string{})
		cfraLabel  = helpers.TranslateV2("crypto_funding_rates_arbitrage", langCode, map[string]string{})
		cifraLabel = helpers.TranslateV2("crypto_index_funding_rates_arbitrage", langCode, map[string]string{})
		sgtLabel   = helpers.TranslateV2("spot_grid_trading", langCode, map[string]string{})
		mtLabel    = helpers.TranslateV2("martingale_trading", langCode, map[string]string{})
		mtdLabel   = helpers.TranslateV2("reverse_martingale_trading", langCode, map[string]string{})
	)

	for _, arrTradingPoolStatisticV := range arrTradingPoolStatistic {
		var (
			// pool            = ""
			strategy        = ""
			profitColorCode = "#309304"
		)

		if arrTradingPoolStatisticV.Strategy == "CFRA" {
			// pool = pool1Label
			strategy = cfraLabel
		} else if arrTradingPoolStatisticV.Strategy == "CIFRA" {
			// pool = pool2Label
			strategy = cifraLabel
		} else if arrTradingPoolStatisticV.Strategy == "SGT" {
			// pool = pool3Label
			strategy = sgtLabel
		} else if arrTradingPoolStatisticV.Strategy == "MT" {
			// pool = pool4Label
			strategy = mtLabel
		} else if arrTradingPoolStatisticV.Strategy == "MTD" {
			// pool = pool5Label
			strategy = mtdLabel
		}

		if arrTradingPoolStatisticV.EarningsRatio < 0 {
			profitColorCode = "#F76464"
		}

		arrTradingPoolStatistics = append(arrTradingPoolStatistics,
			map[string]interface{}{
				// "pool":              pool,
				"pool":              param.PoolType,
				"strategy":          strategy,
				"bid":               helpers.CutOffDecimalv2(arrTradingPoolStatisticV.SellPrice, 2, ".", ",", true),
				"ask":               helpers.CutOffDecimalv2(arrTradingPoolStatisticV.BuyPrice, 2, ".", ",", true),
				"crypto_pair":       strings.Replace(arrTradingPoolStatisticV.Symbol, "USDT", "/USDT", -1),
				"profit":            helpers.CutOffDecimalv2(arrTradingPoolStatisticV.EarningsRatio, 6, ".", ",", true) + "%",
				"profit_color_code": profitColorCode,
				"created_at":        arrTradingPoolStatisticV.TradeAt.Format("2006-01-02 15:04:05"),
			},
		)

		if arrTradingPoolStatisticV.TradeAt.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			totalDailyProfit += arrTradingPoolStatisticV.EarningsRatio
			totalDailyNumberOfTrades++
		}
	}

	// get total daily profit member
	arrTblBonusRebateGroupFn := []models.WhereCondFn{}
	arrTblBonusRebateGroupFn = append(arrTblBonusRebateGroupFn, models.WhereCondFn{
		Condition: " bns_id = ?",
		CondValue: time.Now().Format("2006-01-02"),
		// CondValue: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
	})

	if param.Strategy != "" {
		// var prdMasterID = 0

		// if param.Strategy == "CFRA" {
		// 	prdMasterID = 2
		// } else if param.Strategy == "CIFRA" {
		// 	prdMasterID = 3
		// } else if param.Strategy == "SGT" {
		// 	prdMasterID = 4
		// } else if param.Strategy == "MT" {
		// 	prdMasterID = 5
		// } else if param.Strategy == "MTD" {
		// 	prdMasterID = 6
		// }

		arrTblBonusRebateGroupFn = append(arrTblBonusRebateGroupFn, models.WhereCondFn{
			Condition: " tblq_bonus_strategy_profit.strategy = ?",
			CondValue: param.Strategy,
		})

		if param.PoolType == "B1" {
			arrTblBonusRebateGroupFn = append(arrTblBonusRebateGroupFn,
				models.WhereCondFn{Condition: " REPLACE(REPLACE(tblq_bonus_strategy_profit.crypto_pair, '-', ''), 'USDTM', 'USDT') IN(?, 'ETHUSDT', 'XRPUSDT')", CondValue: "BTCUSDT"},
			)
		} else if param.PoolType == "B2" {
			arrTblBonusRebateGroupFn = append(arrTblBonusRebateGroupFn,
				models.WhereCondFn{Condition: " REPLACE(REPLACE(tblq_bonus_strategy_profit.crypto_pair, '-', ''), 'USDTM', 'USDT') IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT')", CondValue: "BTCUSDT"},
			)
		} else if param.PoolType == "B3" {
			arrTblBonusRebateGroupFn = append(arrTblBonusRebateGroupFn,
				models.WhereCondFn{Condition: " REPLACE(REPLACE(tblq_bonus_strategy_profit.crypto_pair, '-', ''), 'USDTM', 'USDT') IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT', 'LTCUSDT', 'GOGEUSDT', 'DOTUSDT')", CondValue: "BTCUSDT"},
			)
		} else if param.PoolType == "B4" {
			arrTblBonusRebateGroupFn = append(arrTblBonusRebateGroupFn,
				models.WhereCondFn{Condition: " REPLACE(REPLACE(tblq_bonus_strategy_profit.crypto_pair, '-', ''), 'USDTM', 'USDT') IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT', 'LTCUSDT', 'GOGEUSDT', 'DOTUSDT', 'BHCUSDT', 'LINKUSDT')", CondValue: "BTCUSDT"},
			)
		} else if param.PoolType == "B5" {
			arrTblBonusRebateGroupFn = append(arrTblBonusRebateGroupFn,
				models.WhereCondFn{Condition: " REPLACE(REPLACE(tblq_bonus_strategy_profit.crypto_pair, '-', ''), 'USDTM', 'USDT') IN(?, 'ETHUSDT', 'XRPUSDT', 'ADAUSDT', 'BNBUSDT', 'LTCUSDT', 'GOGEUSDT', 'DOTUSDT', 'BHCUSDT', 'LINKUSDT', 'DEFIUSDT', 'BTCDOMUSDT')", CondValue: "BTCUSDT"},
			)
		}
	}

	arrBonusRebateGroup, _ := models.GetTblqBonusStrategyProfitFn(arrTblBonusRebateGroupFn, false)
	// arrBonusRebateGroup, _ := models.TblqBonusRebateGroupFn(arrTblBonusRebateGroupFn, "", false)

	if len(arrBonusRebateGroup) > 0 {
		for _, arrBonusRebateGroupV := range arrBonusRebateGroup {
			totalDailyProfitMember += arrBonusRebateGroupV.FProfit
		}
	}

	var arrReportsSummary = map[string]interface{}{}
	arrReportsSummary["trading_pool"] = param.PoolType
	arrReportsSummary["total_daily_profit"] = helpers.CutOffDecimalv2(totalDailyProfit, 6, ".", ",", true) + "%"
	arrReportsSummary["total_daily_profit_member"] = helpers.CutOffDecimalv2(totalDailyProfitMember, 6, ".", ",", true) + "%"
	arrReportsSummary["total_daily_number_of_trades"] = helpers.CutOffDecimalv2(totalDailyNumberOfTrades, 0, ".", ",", true)

	page := base.Pagination{
		Page:      param.Page,
		DataArr:   arrTradingPoolStatistics,
		HeaderArr: arrReportsSummary,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, ""
}

type GetMemberBinanceIncomeHistoryApiStruct struct {
	Timestamp  string
	Signature  string
	ApiKey     string
	DateFrom   string
	DateTo     string
	CryptoPair string
}
type BinanceIncomeHistoryResponse struct {
	Symbol     string `json:"symbol"`
	IncomeType string `json:"incomeType"`
	Income     string `json:"income"`
	Asset      string `json:"asset"`
	Time       int64  `json:"time"`
	Info       string `json:"info"`
	TranId     int    `json:"tranId"`
	TradeID    string `json:"tradeId"`
}

func (b *GetMemberBinanceIncomeHistoryApiStruct) GetBinanceIncomeHistoryApiv1() ([]*BinanceIncomeHistoryResponse, error) {

	var (
		err      error
		response []*BinanceIncomeHistoryResponse
	)

	data := map[string]interface{}{
		"timestamp": b.Timestamp,
		"signature": b.Signature,
	}

	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/income?timestamp=%v&signature=%v", b.Timestamp, b.Signature)

	if b.DateFrom != "" && b.DateTo != "" && b.CryptoPair != "" {
		dateFrom := b.DateFrom
		dateTo := b.DateTo
		url = fmt.Sprintf("https://fapi.binance.com/fapi/v1/income?symbol=%vstartTime=%v&endTime=%v&timestamp=%v&signature=%v", b.CryptoPair, dateFrom, dateTo, b.Timestamp, b.Signature)
	} else if b.DateFrom != "" && b.DateTo != "" {
		dateFrom := b.DateFrom
		dateTo := b.DateTo
		url = fmt.Sprintf("https://fapi.binance.com/fapi/v1/income?startTime=%v&endTime=%v&timestamp=%v&signature=%v", dateFrom, dateTo, b.Timestamp, b.Signature)
	} else if b.CryptoPair != "" {
		url = fmt.Sprintf("https://fapi.binance.com/fapi/v1/income?symbol=%v&timestamp=%v&signature=%v", b.CryptoPair, b.Timestamp, b.Signature)
	}

	header := map[string]string{
		"Content-Type": "application/json",
		"X-MBX-APIKEY": b.ApiKey,
	}

	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetBinanceIncomeHistoryApiv1 failed", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetBinanceIncomeHistoryApiv1 ReturnErr", res.Body, map[string]interface{}{"res": res, "data": data}, true)
		return response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return response, nil
}

func (b *GetMemberStrategyBalanceStruct) GetMemberStrategyFuturesBalancev1() (MemberStrategyBalanceReturnStruct, string) {
	//  **** return string, not error *****

	var (
		balance     float64
		balanceStr  string
		coin        string
		EwtTypeCode string = "USDT"
		descTrans   string = helpers.Translate("for_crypto_funding_rates_arbitrage_and_crypto_index_funding_rate_arbitrage_strategies", b.LangCode)
	)

	balanceStr = helpers.CutOffDecimal(balance, 8, ".", ",")

	//get api_keys & secret
	arrEntMemberTradingApiFn := make([]models.WhereCondFn, 0)
	arrEntMemberTradingApiFn = append(arrEntMemberTradingApiFn,
		models.WhereCondFn{Condition: "ent_member_trading_api.member_id = ?", CondValue: b.MemberID},
		models.WhereCondFn{Condition: "ent_member_trading_api.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ent_member_trading_api.platform = ?", CondValue: strings.ToUpper(b.Platform)},
		models.WhereCondFn{Condition: "ent_member_trading_api.module = ?", CondValue: "FUTURE"},
		models.WhereCondFn{Condition: "sys_trading_api_platform.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingApi, err := models.GetEntMemberTradingApiFn(arrEntMemberTradingApiFn, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberStrategyFuturesBalancev1-GetEntMemberTradingApiFn", arrEntMemberTradingApiFn, err.Error(), true)
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "something_went_wrong"
	}

	if len(arrEntMemberTradingApi) < 1 {
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "invalid_record"
	}

	//decrypt secret key
	decryptedScrtKey, err := util.DecodeAscii85(arrEntMemberTradingApi[0].ApiSecret)
	if err != nil {
		base.LogErrorLog("GetMemberStrategyFuturesBalancev1-DecodeAscii85", map[string]interface{}{"decryptedScrtKey": decryptedScrtKey, "input": arrEntMemberTradingApi[0].ApiSecret}, err.Error(), true)
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "something_went_wrong"
	}

	currentUnixTimestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	data := fmt.Sprintf("timestamp=%v", currentUnixTimestamp)

	signature := util.GenerateHmacSHA256(decryptedScrtKey, data, "")
	//check ewt_summary_strategy_futures record
	arrEwtSummaryStrategyFn := make([]models.WhereCondFn, 0)
	arrEwtSummaryStrategyFn = append(arrEwtSummaryStrategyFn,
		models.WhereCondFn{Condition: "ewt_summary_strategy_futures.member_id = ?", CondValue: b.MemberID},
		models.WhereCondFn{Condition: "ewt_summary_strategy_futures.coin = ?", CondValue: strings.ToUpper(EwtTypeCode)},
		models.WhereCondFn{Condition: "ewt_summary_strategy_futures.platform = ?", CondValue: strings.ToUpper(b.Platform)},
	)
	arrEwtSummaryStrategy, err := models.GetEwtSummaryStrategyFuturesFn(arrEwtSummaryStrategyFn, "", false)
	if err != nil {
		base.LogErrorLog("GetMemberStrategyFuturesBalancev1-GetEwtSummaryStrategyFuturesFn", map[string]interface{}{"condition": arrEwtSummaryStrategyFn}, err.Error(), true)
		return MemberStrategyBalanceReturnStruct{
			Coin:       coin,
			Balance:    balance,
			BalanceStr: balanceStr,
			Desc:       descTrans,
		}, "something_went_wrong"
	}

	switch strings.ToUpper(b.Platform) {
	case "BN":
		// begin call binance futures balance api
		binanceApi := GetMemberBinanceBalanceApiStruct{
			Timestamp: currentUnixTimestamp,
			Signature: signature,
			ApiKey:    arrEntMemberTradingApi[0].ApiKey,
		}
		rst, err := binanceApi.GetBinanceFuturesBalanceApiv1()

		if err != nil {
			models.ErrorLog("GetMemberStrategyFuturesBalancev1-GetBinanceFuturesBalanceApiv1 Error", map[string]interface{}{"memID": b.MemberID, "api_key": arrEntMemberTradingApi[0].ApiKey, "scrt_key": decryptedScrtKey, "timestamp": currentUnixTimestamp}, nil)
			// if api down grab from ewt_summary_strategy_futures record
			if len(arrEwtSummaryStrategy) > 0 {
				coin = arrEwtSummaryStrategy[0].Coin
				balance = arrEwtSummaryStrategy[0].Balance
				balanceStr = helpers.CutOffDecimal(balance, 8, ".", ",")
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: helpers.CutOffDecimal(balance, 8, ".", ","),
					Desc:       descTrans,
				}, ""
			} else {
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: balanceStr,
					Desc:       descTrans,
				}, ""
			}
		}

		for _, v := range rst.Assets {
			if v.Asset == strings.ToUpper(EwtTypeCode) {
				coin = v.Asset
				balance, _ = strconv.ParseFloat(v.AvailableBalance, 64)
				balanceStr = v.AvailableBalance
			}
		}

	case "KC":
		// begin call kucoin balance api
		kucoinAccountParam := GetMemberKucoinFutureAccountApiTradingStatusParam{
			ApiKey:     arrEntMemberTradingApi[0].ApiKey,
			Secret:     decryptedScrtKey,
			Passphrase: arrEntMemberTradingApi[0].ApiPassphrase,
		}
		rst, err := kucoinAccountParam.GetKucoinFutureAccountApiTradingStatus()

		if err != nil {
			models.ErrorLog("GetMemberStrategyFuturesBalancev1():GetKucoinFutureAccountApiTradingStatus Error", map[string]interface{}{"param": kucoinAccountParam}, nil)
			// if api down grab from ewt_summary_strategy_futures record
			if len(arrEwtSummaryStrategy) > 0 {
				coin = arrEwtSummaryStrategy[0].Coin
				balance = arrEwtSummaryStrategy[0].Balance
				balanceStr = helpers.CutOffDecimal(balance, 8, ".", ",")
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: helpers.CutOffDecimal(balance, 8, ".", ","),
					Desc:       descTrans,
				}, ""
			} else {
				return MemberStrategyBalanceReturnStruct{
					Coin:       coin,
					Balance:    balance,
					BalanceStr: balanceStr,
					Desc:       descTrans,
				}, ""
			}
		}

		coin = rst.Data.Currency
		balance = rst.Data.AvailableBalance
		balanceStr = helpers.CutOffDecimalv2(balance, 6, ".", ",", true)

		fmt.Println("final:", rst, "coin:", coin)
	}

	if len(arrEwtSummaryStrategy) > 0 {
		//update record
		arrUpdateEwtSummaryStrategy := make([]models.WhereCondFn, 0)
		arrUpdateEwtSummaryStrategy = append(arrUpdateEwtSummaryStrategy,
			models.WhereCondFn{Condition: " ewt_summary_strategy_futures.member_id = ? ", CondValue: b.MemberID},
		)
		updateColumn := map[string]interface{}{
			"balance":    balance,
			"updated_at": time.Now(),
		}
		models.UpdatesFn("ewt_summary_strategy_futures", arrUpdateEwtSummaryStrategy, updateColumn, false)

	} else {
		//store record
		arrStoreEwtSummaryStrategyFutures := models.AddEwtSummaryStrategyFuturesStruct{
			MemberID:  b.MemberID,
			Platform:  strings.ToUpper(b.Platform),
			Coin:      strings.ToUpper(EwtTypeCode),
			Balance:   balance,
			CreatedBy: "AUTO",
			CreatedAt: time.Now(),
		}
		models.AddEwtSummaryStrategyFutures(arrStoreEwtSummaryStrategyFutures)
	}

	arrDataReturn := MemberStrategyBalanceReturnStruct{
		Coin:       coin,
		Balance:    balance,
		BalanceStr: balanceStr,
		Desc:       descTrans,
	}

	return arrDataReturn, ""
}

type BinanceFuturesAssetsStruct struct {
	Asset            string `json:"asset"`
	AvailableBalance string `json:"availableBalance"`
}
type BinanceFuturesBalanceResponse struct {
	Assets []BinanceFuturesAssetsStruct `json:"assets"`
}

func (b *GetMemberBinanceBalanceApiStruct) GetBinanceFuturesBalanceApiv1() (*BinanceFuturesBalanceResponse, error) {

	var (
		err      error
		response BinanceFuturesBalanceResponse
	)

	data := map[string]interface{}{
		"timestamp": b.Timestamp,
		"signature": b.Signature,
	}

	url := fmt.Sprintf("https://fapi.binance.com/fapi/v2/account?timestamp=%v&signature=%v", b.Timestamp, b.Signature)
	header := map[string]string{
		"Content-Type": "application/json",
		"X-MBX-APIKEY": b.ApiKey,
	}

	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetBinanceFuturesBalanceApiv1-GetBinanceFuturesBalanceApi failed", err.Error(), map[string]interface{}{"err": err, "data": data}, true)
		return &response, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetBinanceFuturesBalanceApiv1-GetBinanceFuturesBalanceApi", res.Body, map[string]interface{}{"res": res, "data": data}, true)
		return &response, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	return &response, nil
}

type GetMemberAutoTradingLogsParam struct {
	Page int64
}

func GetMemberAutoTradingLogs(memID int, param GetMemberAutoTradingLogsParam, langCode string) (interface{}, string) {
	var arrListingData = []interface{}{}

	// get member data from sls_master_bot_log
	var arrSlsMasterBotLogFn = []models.WhereCondFn{}
	arrSlsMasterBotLogFn = append(arrSlsMasterBotLogFn,
		models.WhereCondFn{Condition: " member_id = ? ", CondValue: memID},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	var arrSlsMasterBotLog, err = models.GetSlsMasterBotLog(arrSlsMasterBotLogFn, false)
	if err != nil {
		base.LogErrorLog("GetMemberAutoTradingLogs():GetSlsMasterBotLog", err.Error(), map[string]interface{}{"condition": arrSlsMasterBotLogFn}, true)
		return nil, "something_went_wrong"
	}

	if len(arrSlsMasterBotLog) > 0 {
		var date = ""
		var logsListing = []interface{}{}
		for arrSlsMasterBotLogK, arrSlsMasterBotLogV := range arrSlsMasterBotLog {
			if date != arrSlsMasterBotLogV.CreatedAt.Format("2006-01-02") {
				if date != "" {
					// finishing date group if current loop data belong to other date
					arrListingData = append(arrListingData,
						map[string]interface{}{
							"date":    date,
							"listing": logsListing,
						},
					)
				}

				// clear logs listing
				logsListing = []interface{}{}

				// update date
				date = arrSlsMasterBotLogV.CreatedAt.Format("2006-01-02")
			}

			// get strategy platform
			var arrSlsMasterBotSettingFn = []models.WhereCondFn{}
			arrSlsMasterBotSettingFn = append(arrSlsMasterBotSettingFn,
				models.WhereCondFn{Condition: " sls_master.doc_no = ? ", CondValue: arrSlsMasterBotLogV.DocNo},
			)
			var arrSlsMasterBotSetting, _ = models.GetSlsMasterBotSetting(arrSlsMasterBotSettingFn, "", false)
			platform := ""
			if len(arrSlsMasterBotSetting) > 0 {
				platform = helpers.TranslateV2(arrSlsMasterBotSetting[0].Platform, langCode, map[string]string{})
			}

			remark := arrSlsMasterBotLogV.Remark

			if strings.Contains(remark, "{") && strings.Contains(remark, "}") {
				extractedStr := helpers.GetStringInBetween(remark, "{", "}")

				if extractedStr != "" {
					extractedStr = "{" + extractedStr + "}"

					// LogErrorMsg struct
					type LogErrorMsg struct {
						Code int    `json:"code"`
						Msg  string `json:"msg"`
					}

					// mapping purchase contract setting into struct
					logErrorMsg := &LogErrorMsg{}
					json.Unmarshal([]byte(extractedStr), logErrorMsg)

					extractedContent := helpers.TranslateV2("error_code", langCode, map[string]string{}) + ": " + strconv.Itoa(logErrorMsg.Code) + " - " + logErrorMsg.Msg
					remark = strings.Replace(remark, extractedStr, extractedContent, -1)
					remark = strings.Replace(remark, "{", "", -1)
					remark = strings.Replace(remark, "}", "", -1)
				}
			}

			remark = strings.Replace(remark, "USDTM", "USDT", -1)
			remark = strings.Replace(remark, "XBT", "BTC", -1)

			remark = BotLogTrans(helpers.TransRemark(remark, langCode), langCode)
			if platform != "" {
				remark = platform + " - " + remark
			}
			// append logs data into logs listing
			logsListing = append(logsListing,
				map[string]string{
					"created_at": arrSlsMasterBotLogV.CreatedAt.Format("2006-01-02 15:04:05"),
					"time":       arrSlsMasterBotLogV.CreatedAt.Format("15:04:05"),
					"type":       arrSlsMasterBotLogV.RemarkType,
					"remark":     remark,
				},
			)

			// finishing date group if current loop is last record
			if arrSlsMasterBotLogK == len(arrSlsMasterBotLog)-1 {
				arrListingData = append(arrListingData,
					map[string]interface{}{
						"date":    date,
						"listing": logsListing,
					},
				)
			}
		}
	}

	page := base.Pagination{
		Page:    param.Page,
		DataArr: arrListingData,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, ""
}

func BotLogTrans(remark, locale string) string {
	remark = strings.ToLower(remark)
	remark = strings.Replace(remark, " ", "_", -1)

	arrFn := []models.WhereCondFn{}
	arrFn = append(arrFn,
		// models.WhereCondFn{Condition: " `key` LIKE ?", CondValue: remark},
		models.WhereCondFn{Condition: " `locale` = ?", CondValue: locale},
	)
	arr, _ := models.GetSlsMasterBotLogTrans(arrFn, false)
	if len(arr) > 0 {
		for _, value := range arr {
			if strings.Contains(remark, value.Key) {
				remark = strings.Replace(remark, value.Key, value.Value, -1)
			}
		}
	}

	remark = strings.Title(strings.Replace(remark, "_", " ", -1))

	return remark
}

func GetKucoinMinOrderAmount(cryptoPair string) (float64, string) {
	var minFirstOrderAmount = 0.00

	// get crypto pair multiplier
	activeContracts, err := GetKucoinActiveContracts(cryptoPair)
	if err != nil {
		base.LogErrorLog("GetKucoinMinOrderAmount():GetKucoinActiveContracts()", err.Error(), map[string]interface{}{"cryptoPair": cryptoPair}, false)
		return minFirstOrderAmount, "something_went_wrong"
	}

	multiplier := activeContracts.Multiplier

	// get crypto pair price
	cryptoPrice, err := GetKucoinCryptoPrice(cryptoPair)
	if err != nil {
		base.LogErrorLog("GetKucoinMinOrderAmount():GetKucoinCryptoPrice()", err.Error(), map[string]interface{}{"cryptoPair": cryptoPair}, false)
		return minFirstOrderAmount, "something_went_wrong"
	}

	priceStr := cryptoPrice.Data.Price
	// price := priceStr

	// kucoin api changed return data type of price from string to float
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		base.LogErrorLog("GetKucoinMinOrderAmount():ParseFloat()", err.Error(), map[string]interface{}{"priceStr": priceStr}, false)
		return minFirstOrderAmount, "something_went_wrong"
	}

	minFirstOrderAmount = float.Mul(multiplier, price)

	return minFirstOrderAmount, ""
}

type KucoinActiveContractsResponse struct {
	Code string                  `json:"code"`
	Data []KucoinActiveContracts `json:"data"`
}

type KucoinActiveContracts struct {
	Symbol         string  `json:"symbol"`
	RootSymbol     string  `json:"rootSymbol"`
	BaseCurrency   string  `json:"baseCurrency"`
	QuoteCurrency  string  `json:"quoteCurrency"`
	SettleCurrency string  `json:"settleCurrency"`
	Multiplier     float64 `json:"multiplier"`
}

func GetKucoinActiveContracts(cryptoPair string) (KucoinActiveContracts, error) {

	var (
		response    *KucoinActiveContractsResponse
		responseVal = KucoinActiveContracts{}
		err         error
	)

	data := map[string]interface{}{}

	url := fmt.Sprintf("https://api-futures.kucoin.com/api/v1/contracts/active")

	header := map[string]string{
		"Content-Type": "application/json",
	}

	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetKucoinActiveContracts():RequestBinanceAPI()", err.Error(), map[string]interface{}{"data": data, "err": err}, false)
		return responseVal, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetKucoinActiveContracts():ReturnErr", res.Body, map[string]interface{}{"data": data, "response": res}, false)
		return responseVal, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	for _, value := range response.Data {
		if value.Symbol == cryptoPair {
			responseVal = value
		}
	}

	return responseVal, nil
}

type BinanceExchangeInfoResponse struct {
	Symbols []BinanceExchangeInfo `json:"symbols"`
}

type BinanceExchangeInfo struct {
	Symbol  string                       `json:"symbol"`
	Filters []BinanceExchangeInfoFilters `json:"filters"`
}

type BinanceExchangeInfoFilters struct {
	MinPrice   string `json:"minPrice"`
	MaxPrice   string `json:"maxPrice"`
	MinQty     string `json:"minQty"`
	MaxQty     string `json:"maxQty"`
	StepSize   string `json:"stepSize"`
	FilterType string `json:"filterType"`
	TickSize   string `json:"tickSize"`
	Notional   string `json:"notional"`
}

func GetBinanceExchangeInfo(cryptoPair string) (BinanceExchangeInfo, error) {

	var (
		response    *BinanceExchangeInfoResponse
		responseVal = BinanceExchangeInfo{}
		err         error
	)

	data := map[string]interface{}{}

	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/exchangeInfo")

	header := map[string]string{
		"Content-Type": "application/json",
	}

	res, err := base.RequestBinanceAPI("GET", url, header, nil, &response)

	if err != nil {
		base.LogErrorLog("GetBinanceExchangeInfo():RequestBinanceAPI()", err.Error(), map[string]interface{}{"data": data, "err": err}, false)
		return responseVal, err
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("GetBinanceExchangeInfo():ReturnErr", res.Body, map[string]interface{}{"data": data, "response": res}, false)
		return responseVal, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Msg: "something_went_wrong"}
	}

	for _, value := range response.Symbols {
		if value.Symbol == cryptoPair {
			responseVal = value
		}
	}

	return responseVal, nil
}

func GetBinanceMinOrderAmount(cryptoPair string) (float64, string) {
	var minFirstOrderAmount = 0.00

	exchangeInfo, err := GetBinanceExchangeInfo(cryptoPair)
	if err != nil {
		base.LogErrorLog("tradingService:GetBinanceMinOrderAmount():GetBinanceExchangeInfo()", err.Error(), map[string]interface{}{"cryptoPair": cryptoPair}, false)
		return minFirstOrderAmount, "something_went_wrong"
	}

	var minQty = 0.00
	var notional = 0.00
	for _, exchangeInfoV := range exchangeInfo.Filters {
		if exchangeInfoV.FilterType == "LOT_SIZE" {
			minQty, err = strconv.ParseFloat(exchangeInfoV.MinQty, 64)
			if err != nil {
				base.LogErrorLog("tradingService:GetBinanceMinOrderAmount():Atoi():1", err.Error(), map[string]interface{}{"value": exchangeInfoV.MinQty}, false)
				return minFirstOrderAmount, "something_went_wrong"
			}
		} else if exchangeInfoV.FilterType == "MIN_NOTIONAL" {
			notional, err = strconv.ParseFloat(exchangeInfoV.Notional, 64)
			if err != nil {
				base.LogErrorLog("tradingService:GetBinanceMinOrderAmount():Atoi():2", err.Error(), map[string]interface{}{"value": exchangeInfoV.Notional}, false)
				return minFirstOrderAmount, "something_went_wrong"
			}
		}
	}

	var marketPrice = 0.00
	arrBinancePrice, err := GetBinanceCryptoPrice(cryptoPair)
	if err != nil {
		arrBinancePrice, err = GetBinanceFutureCryptoPrice(cryptoPair)
		if err != nil {
			base.LogErrorLog("tradingService:GetBinanceMinOrderAmount:GetBinanceCryptoPrice()", map[string]interface{}{"cryptoPair": cryptoPair}, err.Error(), true)
			return minFirstOrderAmount, "something_went_wrong"
		}

		marketPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
	} else {
		marketPrice, err = strconv.ParseFloat(arrBinancePrice.Price, 64)
	}

	if err != nil {
		base.LogErrorLog("tradingService:GetBinanceMinOrderAmount:ParseFloat():2", map[string]interface{}{"value": arrBinancePrice.Price}, err.Error(), true)
		return minFirstOrderAmount, "something_went_wrong"
	}

	minFirstOrderAmount = float.Add(float.Mul(marketPrice, minQty), notional)
	// minFirstOrderAmount = float.Add(minFirstOrderAmount, float.Mul(minFirstOrderAmount, 0.01)) markup 1% to avoid price changes when china there start trade

	return minFirstOrderAmount, ""
}

type MemberAutoTradingSafetyOrdersData struct {
	SafetyOrders float64 `json:"safety_orders"`
}

func GetMemberAutoTradingSafetyOrders(memID int, firstOrderAmount, amount float64, langCode string) (MemberAutoTradingSafetyOrdersData, string) {
	var data = MemberAutoTradingSafetyOrdersData{}

	// get safety orders
	data.SafetyOrders = GetSafetyOrders(firstOrderAmount, amount, 2)

	return data, ""
}
