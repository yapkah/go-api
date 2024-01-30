package models

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// AccessToken struct
type AccessToken struct {
	TokenID   int       `gorm:"primary_key" json:"token_id"`
	ID        string    `gorm:"primary_key" json:"id"`
	SubID     string    `json:"sub_id"`
	Platform  string    `json:"platform"`
	TokenType string    `json:"token_type"`
	Source    uint8     `json:"source"`
	Scope     string    `json:"scope"`
	Status    string    `json:"status"` // A: active | R: revoked | D: duplicate login | RF: token refresh
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// User interface
// interface for getting user info when generating token
type User interface {
	GetUserType() string
	GetMembersID() int
	GetUserSubID() string
	GetUserCode() string
	// GetLoginTokenID() string
	// GetUserEmail() string
	GetUserName() string
	// GetHashedPassword() string
	GetAccessScope() []string
	// GetStatusScope() string
	// GetLanguage() string
	// UpdateLoginTokenID(*gorm.DB, string, string) error
}

// GetAccessTokenByID get access token by id
func GetAccessTokenByID(id string) (*AccessToken, error) {
	var token AccessToken
	err := db.Where("id = ? AND status = ?", id, "A").First(&token).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if token.TokenID <= 0 {
		return nil, nil
	}
	return &token, nil
}

// GetAllStatusAccessTokenByID get access token by id
func GetAllStatusAccessTokenByID(id string) (*AccessToken, error) {
	var token AccessToken
	err := db.Where("id = ?", id).First(&token).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if token.TokenID <= 0 {
		return nil, nil
	}
	return &token, nil
}

// // ExistAccessTokenByID check if token exist
// func ExistAccessTokenByID(id string) (bool, error) {
// 	var token AccessToken
// 	err := db.Select("token_id").Where("id = ? AND status = ?", id, "A").First(&token).Error
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}
// 	if token.TokenID > 0 {
// 		return true, nil
// 	}
// 	return false, nil
// }

// StoreAccessToken store token
func StoreAccessToken(tx *gorm.DB, id string, SubID string, userID int, tokenType string, scope string, status string, exp time.Time, platform string, source uint8) error {
	nowTime := time.Now()
	token := AccessToken{
		ID:        id,
		SubID:     SubID,
		Platform:  platform,
		TokenType: tokenType,
		Source:    source,
		Scope:     scope,
		Status:    status,
		CreatedAt: nowTime,
		UpdatedAt: nowTime,
		ExpiresAt: exp,
	}
	if err := tx.Create(&token).Error; err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GenerateAccessTokenID generate unique token id
func GenerateAccessTokenID() (string, error) {
	var count int
	for {
		var token AccessToken
		id := "AT-" + uuid.New().String()
		err := db.Select("token_id").Where("id = ?", id).First(&token).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if token.TokenID == 0 {
			return id, nil
		}

		if count >= 20 {
			ErrorLog("GenerateAccessTokenID", "error generate access token id", nil)
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.GENERATE_ACCESS_TOKEN_ID_ERROR}
		}
		count++
	}
}

// // Revoke revoke token
// func (at *AccessToken) Revoke(tx *gorm.DB) error {
// 	at.Status = GetAccesTokenRevokeStatus()
// 	return SaveTx(tx, at)
// }

// // Replace replace token for duplicate login
// func (at *AccessToken) Replace(tx *gorm.DB) error {
// 	at.Status = GetAccesTokenReplaceStatus()
// 	return SaveTx(tx, at)
// }

// // Refresh refresh token
// func (at *AccessToken) Refresh(tx *gorm.DB) error {
// 	at.Status = GetAccesTokenRefreshStatus()
// 	return SaveTx(tx, at)
// }

// // UpdateStatus udpate status
// func (at *AccessToken) UpdateStatus(tx *gorm.DB, status string) error {
// 	at.Status = status
// 	return SaveTx(tx, at)
// }

// GetAccesTokenRefreshStatus refresh status
func GetAccesTokenRefreshStatus() string {
	return "RF"
}

// GetAccesTokenReplaceStatus replace status for duplicate login
func GetAccesTokenReplaceStatus() string {
	return "D"
}

// GetAccesTokenRevokeStatus revoke status
func GetAccesTokenRevokeStatus() string {
	return "R"
}

// GetUser get user linked with this token
func (at *AccessToken) GetUser() (User, error) {
	var (
		err error
		usr User
	)

	switch at.TokenType {
	case "MEM":
		if at.Source == 1 {
			arrCond := make([]WhereCondFn, 0)
			arrCond = append(arrCond,
				WhereCondFn{Condition: "t_token = ?", CondValue: at.ID},
			)
			entMemTmpPwd, err := GetEntMemberTmpPwFn(arrCond, false)
			if err != nil {
				return nil, err
			}

			if entMemTmpPwd == nil {
				return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.INVALID_USER}
			}

			arrCond = make([]WhereCondFn, 0)
			arrCond = append(arrCond,
				WhereCondFn{Condition: "members.sub_id = ?", CondValue: at.SubID},
				WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
			)

			arrAdminEntMemberMemberData := map[string]string{
				"member_id": strconv.Itoa(entMemTmpPwd[0].MemberID),
			}
			usr, err = GetAdminEntMemberMemberFn(arrCond, arrAdminEntMemberMemberData, false)
		} else {
			arrCurrentActiveProfileMemberData := CurrentActiveProfileMemberStruct{
				SourceID: int(at.Source),
			}
			arrCond := make([]WhereCondFn, 0)
			arrCond = append(arrCond,
				WhereCondFn{Condition: "members.sub_id = ?", CondValue: at.SubID},
				WhereCondFn{Condition: "members.status = ?", CondValue: "A"},
			)
			usr, err = GetCurrentActiveProfileMemberFn(arrCond, arrCurrentActiveProfileMemberData, false)
		}

	case "ADM":
		// usr, err = GetAdminBySubID(at.SubID)
	default:
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_TOKEN_TYPE}
	}

	if err != nil {
		return nil, err
	}

	if usr == nil {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.INVALID_USER}
	}

	if usrID := usr.GetMembersID(); usrID <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.INVALID_USER}
	}

	return usr, nil
}

// // GetRefreshToken get refresh token linked
// func (at *AccessToken) GetRefreshToken() (*RefreshToken, error) {
// 	return GetRefreshTokenByAccessTokenID(at.ID)
// }

// // RevokeBothToken revoke both access and refresh token
// func (at *AccessToken) RevokeBothToken(tx *gorm.DB) error {
// 	err := at.Revoke(tx)
// 	if err != nil {
// 		return err
// 	}

// 	return at.RevokeRefreshToken(tx)
// }

// // RevokeBothTokenWithStatus revoke both access and refresh token
// func (at *AccessToken) RevokeBothTokenWithStatus(tx *gorm.DB, status string) error {
// 	err := at.UpdateStatus(tx, status)
// 	if err != nil {
// 		return err
// 	}

// 	return at.RevokeRefreshToken(tx)
// }

// // ReplaceLoginToken replace login token for duplicate login
// func (at *AccessToken) ReplaceLoginToken(tx *gorm.DB) error {
// 	err := at.Replace(tx)
// 	if err != nil {
// 		return err
// 	}

// 	return at.RevokeRefreshToken(tx)
// }

// // RefreshToken disable token for refresh token
// func (at *AccessToken) RefreshToken(tx *gorm.DB) error {
// 	err := at.Refresh(tx)
// 	if err != nil {
// 		return err
// 	}

// 	return at.RevokeRefreshToken(tx)
// }

// // RevokeRefreshToken revoke refresh token
// func (at *AccessToken) RevokeRefreshToken(tx *gorm.DB) error {
// 	// find refresh token
// 	rt, err := at.GetRefreshToken()
// 	if err != nil {
// 		return err
// 	}

// 	if rt != nil {
// 		err = rt.Revoke(tx)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// // GetScope get scope in array of string
// func (at *AccessToken) GetScope() ([]string, error) {
// 	var scope []string
// 	err := json.Unmarshal([]byte(at.Scope), &scope)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return scope, nil
// }

// GetAcccessTokenFn get access_token data with dynamic condition
func GetAcccessTokenFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*AccessToken, error) {
	var accessToken []*AccessToken
	tx := db.Table("access_token")
	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&accessToken).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return accessToken, nil
}
