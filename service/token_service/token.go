package token_service

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/util"
)

type Token struct {
	AccessToken  *util.TokenDetails
	RefreshToken *util.TokenDetails
}

// GenerateToken func
// func GenerateToken(tx *gorm.DB, user models.User) {
func GenerateToken(tx *gorm.DB, user models.User, platform string, source uint8) (*Token, error) {
	// find scope
	scope := user.GetAccessScope()
	subID := user.GetUserSubID()
	memberCode := user.GetUserCode()
	username := user.GetUserName()

	var err error

	exp, err := GetTokenExpireTime(source)
	if err != nil {
		base.LogErrorLog("GenerateToken-GetTokenExpireTime", err.Error(), "", true)
		return nil, err
	}

	// generate access token
	at, err := util.GenerateAccessToken(scope, subID, memberCode, username, exp.AccessToken)
	if err != nil {
		base.LogErrorLog("GenerateToken-GenerateAccessToken", err.Error(), "", true)
		return nil, err
	}

	// generate refresh token
	// rt, err := util.GenerateRefreshToken([]string{}, subID, exp.RefreshToken)
	// if err != nil {
	// 	models.ErrorLog("GenerateToken-GenerateRefreshToken", err.Error(), "")
	// 	return nil, err
	// }

	token := &Token{
		AccessToken: at,
		// RefreshToken: rt,
	}

	// store token to db
	err = storeGeneratedToken(tx, token, user, platform, source)
	if err != nil {
		base.LogErrorLog("GenerateToken-storeGeneratedToken", err.Error(), "", true)
		return nil, err
	}

	return token, nil
}

// GenerateAccessTokenWithScope func
// func GenerateAccessTokenWithScope(tx *gorm.DB, user models.User, scope []string, exp time.Duration) (*Token, error) {
// 	var err error

// 	subID := user.GetUserSubID()

// 	// generate access token
// 	at, err := util.GenerateAccessToken(scope, subID, exp)
// 	if err != nil {
// 		return nil, err
// 	}

// 	token := &Token{
// 		AccessToken: at,
// 	}

// 	// store token to db
// 	err = storeGeneratedToken(tx, token, user)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return token, nil
// }

// RefreshToken func
// func RefreshToken(tx *gorm.DB, tokenStr string) (*Token, error) {
// 	// parse token
// 	claim, err := util.ParseToken(tokenStr)
// 	if err != nil {
// 		return nil, &e.CustomError{HTTPCode: http.StatusBadRequest, Code: e.INVAID_REFRESH_TOKEN}
// 	}

// 	// check if scope is refresh token
// 	refreshCheck := false
// 	for _, v := range claim.Scope {
// 		if v == "REFRESH" {
// 			refreshCheck = true
// 			break
// 		}
// 	}

// 	if !refreshCheck {
// 		return nil, &e.CustomError{HTTPCode: http.StatusBadRequest, Code: e.INVAID_REFRESH_TOKEN}
// 	}

// 	// find refresh token
// 	rt, err := models.GetRefreshTokenByID(claim.Id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if rt == nil {
// 		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.INVAID_REFRESH_TOKEN}
// 	}

// 	// find access token by refresh token
// 	at, err := rt.GetAccessToken()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if at.SubID != claim.Subject {
// 		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.ACCESS_TOKEN_NOT_FOUND}
// 	}

// 	// find user by access token
// 	user, err := at.GetUser()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// find user access scope
// 	scope, err := at.GetScope()
// 	if err != nil {
// 		return nil, err
// 	}

// 	exp, err := GetTokenExpireTime()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// get user sub id
// 	subID := user.GetUserSubID()

// 	// generate access token
// 	atNew, err := util.GenerateAccessToken(scope, subID, exp.accessToken)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// generate refresh token
// 	rtNew, err := util.GenerateRefreshToken([]string{}, subID, exp.refreshToken)
// 	if err != nil {
// 		return nil, err
// 	}

// 	token := &Token{
// 		AccessToken:  atNew,
// 		RefreshToken: rtNew,
// 	}

// 	// store new token to db
// 	err = storeGeneratedToken(tx, token, user)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = user.UpdateLoginTokenID(tx, token.AccessToken.UUID, models.GetAccesTokenRefreshStatus())
// 	if err != nil {
// 		return nil, err
// 	}

// 	if user.GetUserType() == "MEM" {
// 		err = models.UpdateMemberConnectionByTokenID(tx, at.ID, token.AccessToken.UUID)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	return token, nil
// }

// storeGeneratedToken func
func storeGeneratedToken(tx *gorm.DB, token *Token, user models.User, platform string, source uint8) error {
	userID := user.GetMembersID()
	subID := user.GetUserSubID()
	tokenType := user.GetUserType()

	if token.AccessToken != nil {
		scope := token.AccessToken.Scope
		scopeStr, err := json.Marshal(scope)
		if err != nil {
			return err
		}

		expAT := time.Unix(token.AccessToken.Expires, 0)
		err = models.StoreAccessToken(tx, token.AccessToken.UUID, subID, userID, tokenType, string(scopeStr), "A", expAT, platform, source)
		if err != nil {
			return err
		}
	}

	if token.RefreshToken != nil {
		expRT := time.Unix(token.RefreshToken.Expires, 0)
		err := models.StoreRefreshToken(tx, token.RefreshToken.UUID, token.AccessToken.UUID, "A", expRT)
		if err != nil {
			return err
		}
	}

	return nil
}

type TokenExpire struct {
	AccessToken  time.Duration
	RefreshToken time.Duration
}

func GetTokenExpireTime(source uint8) (*TokenExpire, error) {
	set, err := models.GetSysGeneralSetupByID("jwt_token_setting")
	if err != nil {
		return nil, err
	}

	// access token expire
	var rte, ate time.Duration

	// refresh token expire
	if source == 1 {
		// access token expire
		accessTokenExpire, err := strconv.Atoi(set.InputType2)
		if err != nil {
			return nil, err
		}
		ate = time.Duration(accessTokenExpire)

		// refresh token expire
		refreshTokenExpire, err := strconv.Atoi(set.InputValue2)
		if err != nil {
			return nil, err
		}
		rte = time.Duration(refreshTokenExpire)
	} else {
		// access token expire
		accessTokenExpire, err := strconv.Atoi(set.InputValue1)
		if err != nil {
			return nil, err
		}
		ate = time.Duration(accessTokenExpire)

		// refresh token expire
		refreshTokenExpire, err := strconv.Atoi(set.SettingValue1)
		if err != nil {
			return nil, err
		}
		rte = time.Duration(refreshTokenExpire)
	}

	tke := &TokenExpire{
		AccessToken:  ate * time.Minute,
		RefreshToken: rte * time.Minute,
	}

	return tke, nil
}

// slot token

// GenerateSlotToken func
// func GenerateSlotToken(tx *gorm.DB, user models.User, slotMachConfig *models.SlotMachConfig) (*Token, error) {
// 	// find scope
// 	scope := user.GetAccessScope()
// 	subID := user.GetUserSubID()
// 	var err error

// 	exp, err := getSlotTokenExpireTime()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tokenData := util.SlotTokenData{MachCode: slotMachConfig.MachCode}

// 	// generate access token
// 	at, err := util.GenerateSlotAccessToken(scope, subID, exp.accessToken, tokenData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// generate refresh token
// 	rt, err := util.GenerateSlotRefreshToken([]string{}, subID, exp.refreshToken)
// 	if err != nil {
// 		return nil, err
// 	}

// 	token := &Token{
// 		AccessToken:  at,
// 		RefreshToken: rt,
// 	}

// 	// store token to db
// 	err = storeGeneratedSlotToken(tx, token, user, slotMachConfig.MachCode, &tokenData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return token, nil
// }

// RefreshSlotToken func
// func RefreshSlotToken(tx *gorm.DB, tokenStr string) (*Token, error) {
// parse token
// claim, err := util.ParseSlotToken(tokenStr)
// if err != nil {
// 	return nil, &e.CustomError{HTTPCode: http.StatusBadRequest, Code: e.INVAID_SLOT_REFRESH_TOKEN}
// }

// // check if scope is refresh token
// refreshCheck := false
// for _, v := range claim.Scope {
// 	if v == "REFRESH" {
// 		refreshCheck = true
// 		break
// 	}
// }

// if !refreshCheck {
// 	return nil, &e.CustomError{HTTPCode: http.StatusBadRequest, Code: e.INVAID_SLOT_REFRESH_TOKEN}
// }

// // find refresh token
// rt, err := models.GetSlotRefreshTokenByID(claim.Id)
// if err != nil {
// 	return nil, err
// }
// if rt == nil {
// 	return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.INVAID_SLOT_REFRESH_TOKEN}
// }

// // find access token by refresh token
// at, err := rt.GetAccessToken()
// if err != nil {
// 	return nil, err
// }

// if at.SubID != claim.Subject {
// 	return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.SLOT_ACCESS_TOKEN_NOT_FOUND}
// }

// // find user by access token
// user, err := at.GetUser()
// if err != nil {
// 	return nil, err
// }

// // find user access scope
// scope, err := at.GetScope()
// if err != nil {
// 	return nil, err
// }

// exp, err := getSlotTokenExpireTime()
// if err != nil {
// 	return nil, err
// }

// get user sub id
// subID := user.GetUserSubID()

// var tokenData util.SlotTokenData

// if at.TokenData != "" {
// 	err = json.Unmarshal([]byte(at.TokenData), &tokenData)

// 	if err != nil {
// 		return nil, err
// 	}
// }

// generate access token
// atNew, err := util.GenerateSlotAccessToken(scope, subID, exp.accessToken, tokenData)
// if err != nil {
// 	return nil, err
// }

// generate refresh token
// rtNew, err := util.GenerateSlotRefreshToken([]string{}, subID, exp.refreshToken)
// if err != nil {
// 	return nil, err
// }

// token := &Token{
// AccessToken:  atNew,
// RefreshToken: rtNew,
// }

// revoke old access token and refresh token
// err = at.RevokeBothToken(tx)
// if err != nil {
// 	return nil, err
// }

// store new token to db
// err = storeGeneratedSlotToken(tx, token, user, at.MachCode, &tokenData)
// if err != nil {
// 	return nil, err
// }

// err = user.UpdateSlotTokenID(tx, token.AccessToken.UUID, models.GetSlotAccesTokenRefreshStatus())
// if err != nil {
// 	return nil, err
// }

// return token, nil
// }

// getSlotTokenExpireTime func
// func getSlotTokenExpireTime() (*tokenExpire, error) {
// 	set, err := models.GetSysSettingByGroup("token")
// 	if err != nil {
// 		return nil, err
// 	}

// 	// access token expire
// 	ate, err := set["slot_access_token_expire"].ValueToDuration()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// refresh token expire
// 	rte, err := set["slot_refresh_token_expire"].ValueToDuration()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tke := &tokenExpire{
// 		accessToken:  ate * time.Minute,
// 		refreshToken: rte * time.Minute,
// 	}

// 	return tke, nil
// }

// storeGeneratedSlotToken func
// func storeGeneratedSlotToken(tx *gorm.DB, token *Token, user models.User, machCode string, data *util.SlotTokenData) error {
// userID := user.GetUserID()
// subID := user.GetUserSubID()

// if token.AccessToken != nil {
// 	scope := token.AccessToken.Scope
// 	scopeStr, err := json.Marshal(scope)
// 	if err != nil {
// 		return err
// 	}

// 	var tkData string
// 	if data != nil {
// 		d, err := json.Marshal(data)
// 		if err != nil {
// 			return err
// 		}
// 		tkData = string(d)
// 	}

// 	expAT := time.Unix(token.AccessToken.Expires, 0)
// 	err = models.StoreSlotAccessToken(tx, token.AccessToken.UUID, subID, userID, machCode, string(scopeStr), tkData, "A", expAT)
// 	if err != nil {
// 		return err
// 	}
// }

// if token.RefreshToken != nil {
// 	expRT := time.Unix(token.RefreshToken.Expires, 0)
// 	err := models.StoreSlotRefreshToken(tx, token.RefreshToken.UUID, token.AccessToken.UUID, "A", expRT)
// 	if err != nil {
// 		return err
// 	}
// }

// return nil
// }
