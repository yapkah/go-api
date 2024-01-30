package util

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
)

// var jwtSecret []byte

type Claims struct {
	jwt.StandardClaims
	Scope []string    `json:"scope"`
	Data  interface{} `json:"data"`
}

// // TokenDetails struct
type TokenDetails struct {
	Token   string
	Scope   []string
	UUID    string
	Expires int64
}

// // SlotTokenData struct
// type SlotTokenData struct {
// 	MachCode string `json:"mach_code"`
// }

// GenerateAccessToken generate tokens used for auth
func GenerateAccessToken(scope []string, subID, memberCode, username string, expires time.Duration) (*TokenDetails, error) {
	var (
		err error
		sc  []string // scope
	)
	sa := false // scope access
	for _, v := range scope {
		if v == "ACCESS" {
			sa = true
			break
		}
	}

	if sa {
		sc = scope
	} else {
		// if "ACCESS" scope not exist (mainly prevent scope duplicate when refresh token)
		sc = append([]string{"ACCESS"}, scope...)
	}

	// time data
	exp := base.GetCurrentDateTimeT()
	if expires > 0 {
		exp = exp.Add(expires)
	} else {
		exp = exp.Add(setting.TokenSetting.AccessTokenExpire)
	}

	// generate access token unique id
	uuid, err := models.GenerateAccessTokenID()
	if err != nil {
		return nil, err
	}

	arrData := map[string]interface{}{
		"member_code": memberCode,
		"username":    username,
	}
	return GenerateToken(uuid, subID, exp, sc, "LOGIN", arrData)
}

// GenerateRefreshToken generate tokens used for auth
func GenerateRefreshToken(scope []string, subID string, expires time.Duration) (*TokenDetails, error) {
	var err error
	sc := append([]string{"REFRESH"}, scope...)

	// time data
	exp := time.Now()
	if expires > 0 {
		exp = exp.Add(expires)
	} else {
		exp = exp.Add(setting.TokenSetting.AccessTokenExpire)
	}

	// generate refresh token id
	uuid, err := models.GenerateRefreshTokenID()
	if err != nil {
		return nil, err
	}

	return GenerateToken(uuid, subID, exp, sc, "LOGIN", nil)
}

// // GenerateSlotAccessToken generate tokens used for slot
// func GenerateSlotAccessToken(scope []string, subID string, expires time.Duration, data interface{}) (*TokenDetails, error) {
// 	var (
// 		err error
// 		sc  []string // scope
// 	)
// 	sa := false // scope access
// 	for _, v := range scope {
// 		if v == "ACCESS" {
// 			sa = true
// 			break
// 		}
// 	}

// 	if sa {
// 		sc = scope
// 	} else {
// 		// if "ACCESS" scope not exist (mainly prevent scope duplicate when refresh token)
// 		sc = append([]string{"ACCESS"}, scope...)
// 	}

// 	// time data
// 	exp := time.Now()
// 	if expires > 0 {
// 		exp = exp.Add(expires)
// 	}

// 	// generate access token unique id
// 	uuid, err := models.GenerateSlotAccessTokenID()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return generateToken(uuid, subID, exp, sc, "SLOT", data)
// }

// // GenerateSlotRefreshToken generate tokens used for slot
// func GenerateSlotRefreshToken(scope []string, subID string, expires time.Duration) (*TokenDetails, error) {
// 	var err error
// 	sc := append([]string{"REFRESH"}, scope...)

// 	// time data
// 	exp := time.Now()
// 	if expires > 0 {
// 		exp = exp.Add(expires)
// 	}

// 	// generate refresh token id
// 	uuid, err := models.GenerateSlotRefreshTokenID()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return generateToken(uuid, subID, exp, sc, "SLOT", nil)
// }

// func GenerateToken
func GenerateToken(uuid string, subID string, expires time.Time, scope []string, tokenType string, data map[string]interface{}) (*TokenDetails, error) {

	var err error
	td := &TokenDetails{
		Scope:   scope,
		Expires: expires.Unix(),
		UUID:    uuid,
	}

	// get setting in config file
	appName := setting.Cfg.Section("custom").Key("AppName").String()

	// issue time
	nowTime := time.Now().Unix()

	var res []byte
	// get private key
	if tokenType == "LOGIN" {
		res, err = ioutil.ReadFile("storage/private.pem")
	} else if tokenType == "SLOT" {
		res, err = ioutil.ReadFile("storage/game_private.pem")
	} else {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.INVALID_JWT_TOKEN_TYPE}
	}

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.PRIVATE_KEY_MISSING}
	}
	jwtSecret, _ := jwt.ParseRSAPrivateKeyFromPEM(res)

	claims := Claims{
		jwt.StandardClaims{
			Id:        td.UUID,
			ExpiresAt: td.Expires,
			IssuedAt:  nowTime,
			// NotBefore: nowTime,
			Issuer:  appName,
			Subject: subID,
		},
		scope,
		data,
	}

	// generate access token
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	td.Token, err = tokenClaims.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	return td, nil
}

// // ParseToken parsing token
func ParseToken(token string) (*Claims, error) {
	// get public token
	res, err := ioutil.ReadFile("storage/public.pem")
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.PUBLIC_KEY_MISSING}
	}

	return parseToken(token, res)
}

// // ParseSlotToken parsing slot token
// func ParseSlotToken(token string) (*Claims, error) {
// 	// get public token
// 	res, err := ioutil.ReadFile("storage/game_public.pem")
// 	if err != nil {
// 		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.PUBLIC_KEY_MISSING}
// 	}

// 	return parseToken(token, res)
// }

func parseToken(token string, file []byte) (*Claims, error) {
	jwtSecret, _ := jwt.ParseRSAPublicKeyFromPEM(file)

	// parse token
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok {
			return claims, err
		}
	}

	return nil, err
}
