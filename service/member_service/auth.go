package member_service

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/service/token_service"

	"github.com/jinzhu/gorm"
)

// GetLoginUserByUsername get user by using username
// func GetLoginUserByUsername(username string) (*models.Members, error) {
// 	var (
// 		errMsg string
// 		mem    *models.Members
// 	)

// 	// user id check
// 	errMsg = base.UsernameChecking(username)
// 	if errMsg == "" {
// 		arrCond := make([]models.WhereCondFn, 0)
// 		arrCond = append(arrCond,
// 			models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: username},
// 			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
// 		)
// 		mem, _ = models.GetMembersFn(arrCond, false)
// 		if mem != nil && mem.Status == "T" {
// 			return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ACCOUNT_TERMINATED}
// 		}

// 		if mem != nil {
// 			return mem, nil
// 		}
// 	}
// 	return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_USERNAME}
// }

// MemLoginData struct
type MemLoginData struct {
	AccessToken string
	ATUUID      string
	// ExpiredAtUnix int64
	// ExpiredAt     string
}

// ProcessMemberLogin process member login (gen tokens (access & refresh) and add log)
func ProcessMemberLogin(tx *gorm.DB, user models.User, langCode string, platform string, source uint8) (*MemLoginData, error) {
	dtNow := base.GetCurrentDateTimeT().Add(-25 * time.Second) // must put it here, otherwise will result in different in a few milli-seconds
	subID := user.GetUserSubID()
	membersID := user.GetMembersID()

	// inactivate all the existing login token (access token & refresh token) to prevent multiple active connection need to inactive all connection
	InactivateExistingLoginToken(tx, subID, platform, source)

	// generate token
	token, err := token_service.GenerateToken(tx, user, platform, source)
	if err != nil {
		arrErr := map[string]interface{}{
			"user":     user,
			"platform": platform,
			"source":   source,
		}
		base.LogErrorLog("ProcessMemberLogin-GenerateToken", err.Error(), arrErr, true)

		return nil, err
	}
	expAT := time.Unix(token.AccessToken.Expires, 0)

	// inactivate last active account'
	InactivateExistingLoginLog(tx, membersID, platform, source)

	// save AddHtmlfiveLoginLog
	if strings.ToLower(platform) == "htmlfive" {
		activeHtmlLoginLog := models.HtmlfiveLoginLog{
			TUserID: membersID,
			// TNickName:   username,
			TType:       "MEM",
			Source:      source,
			LanguageID:  langCode,
			TToken:      token.AccessToken.UUID,
			BLogin:      1,
			BLogout:     0,
			DtLogin:     dtNow,
			DtExpiry:    expAT,
			DtTimestamp: dtNow,
		}
		err = models.AddHtmlfiveLoginLog(tx, activeHtmlLoginLog)
		if err != nil {
			return nil, err
		}
	} else if strings.ToLower(platform) == "app" {
		activeAppLoginLog := models.AppLoginLog{
			TUserID: membersID,
			// TNickName:   username,
			TType:       "MEM",
			Source:      source,
			LanguageID:  langCode,
			TToken:      token.AccessToken.UUID,
			BLogin:      1,
			BLogout:     0,
			DtLogin:     dtNow,
			DtExpiry:    expAT,
			DtTimestamp: dtNow,
		}
		err = models.AddAppLoginLog(tx, activeAppLoginLog)
		if err != nil {
			return nil, err
		}
	}

	// dT := base.TimeFormat(expAT, "2006-01-02 15:04:05")
	arrDataReturn := MemLoginData{
		AccessToken: token.AccessToken.Token,
		// RefreshToken:  token.RefreshToken.Token,
		// ExpiredAtUnix: token.AccessToken.Expires,
		ATUUID: token.AccessToken.UUID,
	}

	return &arrDataReturn, nil
}

// InactivateExistingLoginLog Inactivate all the existing login log (htmlfive_login_log & app_login_log)
func InactivateExistingLoginLog(tx *gorm.DB, entMemberID int, platform string, source uint8) {
	if strings.ToLower(platform) == "htmlfive" {
		data, err := models.GetExistingActiveHtmlfiveLoginLog(entMemberID, source, false)
		if err == nil {
			if len(data) > 0 {
				dtNow := base.GetCurrentDateTimeT()
				for _, v := range data {
					activeHtmlLoginLog := models.HtmlfiveLoginLog{
						TUserID:     entMemberID,
						TNickName:   v.TNickName,
						TType:       v.TType,
						Source:      v.Source,
						LanguageID:  v.LanguageID,
						TToken:      v.TToken,
						BLogin:      0,
						BLogout:     1,
						DtLogin:     v.DtLogin,
						DtExpiry:    v.DtExpiry,
						DtTimestamp: dtNow,
					}
					err = models.AddHtmlfiveLoginLog(tx, activeHtmlLoginLog)
					if err != nil {
						base.LogErrorLog("InactivateExistingLoginLog-failed_to_save_de-activate_htmlfive_login_log", err, activeHtmlLoginLog, true)
					}
				}
			}
		}
	} else if strings.ToLower(platform) == "app" {
		data, err := models.GetExistingActiveAppLoginLog(entMemberID, source, false)
		if err == nil {
			if len(data) > 0 {
				dtNow := base.GetCurrentDateTimeT()
				for _, v := range data {
					activeAppLoginLog := models.AppLoginLog{
						TUserID:     entMemberID,
						TNickName:   v.TNickName,
						TType:       v.TType,
						Source:      v.Source,
						LanguageID:  v.LanguageID,
						TToken:      v.TToken,
						BLogin:      0,
						BLogout:     1,
						DtLogin:     v.DtLogin,
						DtExpiry:    v.DtExpiry,
						DtTimestamp: dtNow,
					}
					err = models.AddAppLoginLog(tx, activeAppLoginLog)
					if err != nil {
						base.LogErrorLog("InactivateExistingLoginLog-failed_to_save_de-activate_app_login_log", err, activeAppLoginLog, true)
					}
				}
			}
		}
	}
}

// InactivateExistingLoginToken Inactivate all the existing login token (access_token & refresh_token)
func InactivateExistingLoginToken(tx *gorm.DB, subID string, platform string, source uint8) {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "sub_id = ?", CondValue: subID},
		models.WhereCondFn{Condition: "platform = ?", CondValue: platform},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "source = ?", CondValue: source},
	)
	// in here: id = access_token, token_id = id (auto increment)
	accToken, err := models.GetAcccessTokenFn(arrCond, "id, token_id", false)

	if err == nil {
		if len(accToken) > 0 {
			var arrTokenID []int
			for _, v := range accToken {
				arrTokenID = append(arrTokenID, v.TokenID)
			}

			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "token_id IN (?)", CondValue: arrTokenID})
			updateColumn := map[string]interface{}{"status": models.GetAccesTokenRevokeStatus()}
			err = models.UpdatesFnTx(tx, "access_token", arrUpdCond, updateColumn, false)
			if err != nil {
				base.LogErrorLog("InactivateExistingLoginToken-Update access_token to R", arrUpdCond, updateColumn, true)
			}

			for _, v2 := range accToken {
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "access_token_id = ?", CondValue: v2.ID},
					models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
				)

				refreshToken, err := models.GetRefreshTokenFn(arrCond, "", false)
				if err == nil {
					if len(refreshToken) > 0 {
						for _, v3 := range refreshToken {
							arrUpdCond = make([]models.WhereCondFn, 0)
							arrUpdCond = append(arrUpdCond,
								models.WhereCondFn{Condition: "refresh_token_id = ?", CondValue: v3.RefreshTokenID},
								models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
							)
							updateColumn := map[string]interface{}{"status": "R"}
							err = models.UpdatesFnTx(tx, "refresh_token", arrUpdCond, updateColumn, false)
							if err != nil {
								base.LogErrorLog("InactivateExistingLoginToken-Update refresh_token to R", arrUpdCond, updateColumn, true)
							}
						}
					}
				} else {
					base.LogErrorLog("InactivateExistingLoginToken-failed_to_get_refresh_token", err, arrCond, true)
				}
			}
		}
	}
}

// ProcessMemberLogout process member logout
func ProcessMemberLogout(tx *gorm.DB, member models.EntMemberMembers, platform string, source uint8) {
	InactivateExistingLoginLog(tx, member.ID, platform, source)
	InactivateExistingLoginToken(tx, member.SubID, platform, source)
}

type LoginAttemptsLogData struct {
	MemberID int
	ClientIP string
}

// func LoginAttemptsLog
func LoginAttemptsLog(arrData LoginAttemptsLogData, loginType string, status string) {

	arrLoginSetting, _ := models.GetSysGeneralSetupByID("login_setting")
	if arrLoginSetting == nil {
		base.LogErrorLog("LoginAttemptsLog-general_setup_missing_login_setting", "sys_general_setup:login_setting", nil, true)
		return
	}
	loginSettingSwitch := arrLoginSetting.SettingValue2
	maxAttemptsString := arrLoginSetting.SettingValue3

	if loginSettingSwitch != "1" {
		return
	}

	var arrMaxAttempts []int
	if err := json.Unmarshal([]byte(maxAttemptsString), &arrMaxAttempts); maxAttemptsString == "" || err != nil {
		base.LogErrorLog("LoginAttemptsLog_failed_in_decoding_max_attempts", "login_setting", maxAttemptsString, true)
		return
	}
	maxAttempts := arrMaxAttempts[len(arrMaxAttempts)-1]

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sys_login_attempts_log.member_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " sys_login_attempts_log.login_type = ?", CondValue: "member"},
	)
	arrSysLoginAttemptsLog, _ := models.GetSysLoginAttemptsLogFn(arrCond, false)

	// if login failed
	if strings.ToLower(status) == "f" {
		ip := "no_ip"
		if arrData.ClientIP != "" {
			ip = arrData.ClientIP
		}

		if len(arrSysLoginAttemptsLog) > 0 {
			if arrSysLoginAttemptsLog[0].Attempts >= (maxAttempts - 1) {
				// stored into locked account log table
				var arrClientIP []string
				arrClientIP = append(arrClientIP, ip)
				arrClientIPString := ""
				arrClientIPJ, err := json.Marshal(arrClientIP)
				if err == nil {
					arrClientIPString = string(arrClientIPJ)
				}
				arrSysLoginLockedAccountLogData := models.AddSysLoginLockedAccountLogStruct{
					MemberID:  arrData.MemberID,
					LoginType: loginType,
					ClientIP:  arrClientIPString,
				}
				err = models.AddSysLoginLockedAccountLog(arrSysLoginLockedAccountLogData)
				if err != nil {
					base.LogErrorLog("LoginAttemptsLog-failed_to_save_arrSysLoginLockedAccountLogData", err, arrSysLoginLockedAccountLogData, true)
					return
				}
			}

			var arrClientIP []string
			clientIPString := arrSysLoginAttemptsLog[0].ClientIP
			if err := json.Unmarshal([]byte(clientIPString), &arrClientIP); err != nil {
				base.LogErrorLog("LoginAttemptsLog-failed_in_decoding_client_ip", "login_setting", clientIPString, true)
				return
			}

			// check if client ip contains $ip
			clientIPExits := helpers.StringInSlice(ip, arrClientIP)
			if !clientIPExits {
				arrClientIP = append(arrClientIP, ip)
			}

			// increase attempts by 1 and add diff ip address into client ip
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " sys_login_attempts_log.id = ? ", CondValue: arrSysLoginAttemptsLog[0].ID},
			)
			arrClientIPString := ""
			arrClientIPJ, err := json.Marshal(arrClientIP)
			if err == nil {
				arrClientIPString = string(arrClientIPJ)
			}
			updateColumn := map[string]interface{}{"attempts": arrSysLoginAttemptsLog[0].Attempts + 1, "client_ip": arrClientIPString}
			err = models.UpdatesFn("sys_login_attempts_log", arrCond, updateColumn, false)

			if err != nil {
				base.LogErrorLog("LoginAttemptsLog-failed_in_update_sys_login_attempts_log", arrCond, updateColumn, true)
				return
			}
		} else {
			// stored into attempts log table
			var arrClientIP []string
			arrClientIP = append(arrClientIP, ip)
			arrClientIPString := ""
			arrClientIPJ, err := json.Marshal(arrClientIP)
			if err == nil {
				arrClientIPString = string(arrClientIPJ)
			}
			arrSysLoginAttemptsLogData := models.AddSysLoginAttemptsLogStruct{
				MemberID:  arrData.MemberID,
				LoginType: loginType,
				ClientIP:  arrClientIPString,
				Attempts:  1,
			}
			err = models.AddSysLoginAttemptsLog(arrSysLoginAttemptsLogData)
			if err != nil {
				base.LogErrorLog("AddSysLoginAttemptsLog-failed_to_save_arrSysLoginAttemptsLogData", err, arrSysLoginAttemptsLogData, true)
				return
			}
		}
	} else {
		if len(arrSysLoginAttemptsLog) > 0 {
			// set attempts to 0
			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " sys_login_attempts_log.member_id = ? ", CondValue: arrData.MemberID},
				models.WhereCondFn{Condition: " sys_login_attempts_log.login_type = ? ", CondValue: loginType},
			)

			updateColumn := map[string]interface{}{"attempts": 0}
			err := models.UpdatesFn("sys_login_attempts_log", arrCond, updateColumn, false)

			if err != nil {
				base.LogErrorLog("LoginAttemptsLog-failed_in_update_sys_login_attempts_log_reset_attemps", arrCond, updateColumn, true)
				return
			}
		}
	}
}

type IsLockedAccountDataReturn struct {
	IsLockedAccount bool
	Hours           float64
	Minutes         float64
	Seconds         float64
}

func IsLockedAccount(arrData LoginAttemptsLogData, loginType string) (IsLockedAccountDataReturn, error) {
	arrLoginSetting, _ := models.GetSysGeneralSetupByID("login_setting")

	arrDataReturn := IsLockedAccountDataReturn{
		Hours:   0,
		Minutes: 0,
		Seconds: 0,
	}

	if arrLoginSetting == nil {
		base.LogErrorLog("IsLockedAccount-login_setting_missing", "setting_id:login_setting", nil, true)
		return arrDataReturn, nil
	}

	loginSettingSwitch := arrLoginSetting.SettingValue2
	maxAttemptsString := arrLoginSetting.SettingValue3
	lockTimeString := arrLoginSetting.SettingValue4
	if loginSettingSwitch != "1" {
		return arrDataReturn, nil
	}

	var arrMaxAttempts []int
	if err := json.Unmarshal([]byte(maxAttemptsString), &arrMaxAttempts); maxAttemptsString == "" || err != nil {
		base.LogErrorLog("LoginAttemptsLog-failed_in_decoding_max_attempts", "login_setting", maxAttemptsString, true)
		return arrDataReturn, nil
	}
	// maxAttempts := arrMaxAttempts[len(arrMaxAttempts)-1]

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sys_login_attempts_log.member_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " sys_login_attempts_log.login_type = ?", CondValue: "member"},
	)
	arrSysLoginAttemptsLog, _ := models.GetSysLoginAttemptsLogFn(arrCond, false)

	curDateTimeString := base.GetCurrentTime("2006-01-02 15:04:05")
	curDateTimeT := base.GetCurrentDateTimeT()

	if len(arrSysLoginAttemptsLog) > 0 {
		latestDtT := arrSysLoginAttemptsLog[0].CreatedAt
		if arrSysLoginAttemptsLog[0].UpdatedAt.After(arrSysLoginAttemptsLog[0].CreatedAt) {
			latestDtT = arrSysLoginAttemptsLog[0].UpdatedAt
		}

		var arrLockTime []string
		if err := json.Unmarshal([]byte(lockTimeString), &arrLockTime); lockTimeString == "" || err != nil {
			base.LogErrorLog("LoginAttemptsLog-failed_in_decoding_lock_time", "login_setting", lockTimeString, true)
			return arrDataReturn, nil
		}

		for k1 := len(arrMaxAttempts) - 1; k1 >= 0; k1-- {
			// get the time that will be unlocked
			unlockTimeT := latestDtT
			lockTime := arrLockTime[k1]
			unlockTimeT = base.AddDurationInString(latestDtT, lockTime)
			unlockTimeString := unlockTimeT.Format("2006-01-02 15:04:05")
			maxAttempt := arrMaxAttempts[k1]
			if arrSysLoginAttemptsLog[0].Attempts >= maxAttempt {
				// if havent reach unlock time
				if curDateTimeString < unlockTimeString {
					timeLeft := unlockTimeT.Sub(curDateTimeT)
					arrHrMinSec := base.ConvertSecToHrMinSec(timeLeft)
					arrDataReturn.IsLockedAccount = true
					arrDataReturn.Hours = arrHrMinSec.Hours
					arrDataReturn.Minutes = arrHrMinSec.Minutes
					arrDataReturn.Seconds = arrHrMinSec.Seconds
					break
				} else {
					// set attempts to 0
					arrUpdCond := make([]models.WhereCondFn, 0)
					arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.MemberID})
					arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: " login_type = ? ", CondValue: loginType})
					updateColumn := map[string]interface{}{"attempts": 0}
					err := models.UpdatesFn("sys_login_attempts_log", arrUpdCond, updateColumn, false)
					if err != nil {
						base.LogErrorLog("IsLockedAccount-update_sys_login_attempts_log_attempts_to_0_1", arrUpdCond, updateColumn, true)
						return arrDataReturn, nil
					}
				}
			}
		}
	}
	return arrDataReturn, nil
}

// func ProcessValidateToken
func ProcessValidateToken(platform string, token string, memberID int) bool {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "b_login = ?", CondValue: 0},
		models.WhereCondFn{Condition: "b_logout = ?", CondValue: 1},
		models.WhereCondFn{Condition: "t_token = ?", CondValue: token},
	)

	// base.LogErrorLog("ProcessValidateToken_app", platform, arrCond, true)
	if strings.ToLower(platform) == "htmlfive" {
		tokenRst, err := models.GetHtmlfiveLoginLogFn(arrCond, "", false)
		if err != nil {
			base.LogErrorLog("ProcessValidateToken_failed_to_get_inactive_HtmlfiveLoginLog", err.Error(), arrCond, true)
			return true
		}
		if tokenRst != nil {
			return false
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "b_login = ?", CondValue: 1},
			models.WhereCondFn{Condition: "b_logout = ?", CondValue: 0},
			models.WhereCondFn{Condition: " dt_login <= NOW() AND dt_expiry >= NOW() AND t_token = ?", CondValue: token},
		)

		tokenRst, err = models.GetHtmlfiveLoginLogFn(arrCond, "", false)

		if err != nil {
			base.LogErrorLog("ProcessValidateToken_failed_to_get_active_HtmlfiveLoginLog", err.Error(), arrCond, true)
			return true
		}

		if tokenRst == nil {
			return false
		}
	} else if strings.ToLower(platform) == "app" {
		tokenRst, err := models.GetAppLoginLogFn(arrCond, "", false)
		// base.LogErrorLog("ProcessValidateToken_app_1", arrCond, tokenRst, true)
		if err != nil {
			// base.LogErrorLog("ProcessValidateToken_failed_to_get_inactive_AppLoginLog", err.Error(), arrCond, true)
			return true
		}
		// base.LogErrorLog("ProcessValidateToken_tokenRst_1", tokenRst, arrCond, true)
		if tokenRst != nil {
			return false
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "b_login = ?", CondValue: 1},
			models.WhereCondFn{Condition: "b_logout = ?", CondValue: 0},
			models.WhereCondFn{Condition: " dt_login <= NOW() AND dt_expiry >= NOW() AND t_token = ?", CondValue: token},
		)

		tokenRst, err = models.GetAppLoginLogFn(arrCond, "", false)
		// base.LogErrorLog("ProcessValidateToken_app_2", arrCond, tokenRst, true)

		if err != nil {
			// base.LogErrorLog("ProcessValidateToken_failed_to_get_active_AppLoginLog", err.Error(), arrCond, true)
			return true
		}

		// base.LogErrorLog("ProcessValidateToken_tokenRst_2", tokenRst, arrCond, true)
		if tokenRst == nil {
			return false
		}
	}
	// base.LogErrorLog("outside", arrCond, nil, true)
	return true
}

// func ProcessExtendLoginPeriod
func ProcessExtendLoginPeriod(platform string, token string, memberID int) {

	settingID := strings.Replace(strings.ToLower(platform+"_login_setting"), " ", "_", -1)
	loginSetting, err := models.GetSysGeneralSetupByID(settingID)
	if err != nil {
		base.LogErrorLog("JWT-failed_to_get_"+settingID, err.Error(), settingID, true)
	}
	if loginSetting.InputValue1 == "1" {
		if err == nil {
			if loginSetting.InputType1 != "" {
				dtStart := base.GetCurrentDateTimeT()
				// dtNowString := base.GetCurrentTime("2006-01-02 15:04:05") // incorrect vers.
				// dtNowT, _ := time.Parse("2006-01-02 15:04:05", dtNowString) // incorrect vers.
				dTexp := base.AddDurationInString(dtStart, loginSetting.InputType1)

				arrUpdCond := make([]models.WhereCondFn, 0)
				arrUpdCond = append(arrUpdCond,
					models.WhereCondFn{Condition: " t_user_id = ? ", CondValue: memberID},
					models.WhereCondFn{Condition: " t_token = ? ", CondValue: token},
				)
				updateColumn := map[string]interface{}{"dt_expiry": dTexp}
				err = models.UpdatesFn(platform+"_login_log", arrUpdCond, updateColumn, false)
				if err != nil {
					base.LogErrorLog("ProcessExtendLoginPeriod_update_dt_expiry_failed", err.Error(), updateColumn, true)
				}
			}
		}
	}
}

func UpdateCurrentProfileWithLoginMember(tx *gorm.DB, entMemberMainID, currentLoginMember int, sourceID int) error {
	// arrUpdCond := make([]models.WhereCondFn, 0)
	// arrUpdCond = append(arrUpdCond,
	// 	models.WhereCondFn{Condition: " ent_member.main_id = ? ", CondValue: entMemberMainID},
	// )
	// updateColumn := map[string]interface{}{"current_profile": 0}
	// err := models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	// if err != nil {
	// 	base.LogErrorLog("UpdateCurrentProfileWithLoginMember-update_current_profile_0_failed", err.Error(), updateColumn, true)
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	// }

	// arrUpdCond = make([]models.WhereCondFn, 0)
	// arrUpdCond = append(arrUpdCond,
	// 	models.WhereCondFn{Condition: " ent_member.id = ? ", CondValue: currentLoginMember},
	// )
	// updateColumn = map[string]interface{}{"current_profile": 1}
	// err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	// if err != nil {
	// 	base.LogErrorLog("UpdateCurrentProfileWithLoginMember-update_current_profile_1_failed", err.Error(), updateColumn, true)
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	// }
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: " source_id = ? ", CondValue: sourceID},
		models.WhereCondFn{Condition: " main_id = ? ", CondValue: entMemberMainID},
	)
	updateColumn := map[string]interface{}{"member_id": currentLoginMember}
	err := models.UpdatesFnTx(tx, "ent_current_profile", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("UpdateCurrentProfileWithLoginMember-update_current_profile_1_failed", err.Error(), updateColumn, false)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return nil
}
