package models

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// RefreshToken struct
type RefreshToken struct {
	RefreshTokenID int       `gorm:"primary_key" json:"refresh_token_id"`
	ID             string    `gorm:"primary_key" json:"id"`
	AccessTokenID  string    `json:"access_token_id"`
	Status         string    `json:"status"` // A: active | R: revoked
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// GetRefreshTokenByID get token by id
// func GetRefreshTokenByID(id string) (*RefreshToken, error) {
// 	var token RefreshToken
// 	err := db.Where("id = ? AND status = ?", id, "A").First(&token).Error
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}
// 	if token.ID == "" {
// 		return nil, nil
// 	}
// 	return &token, nil
// }

// GetRefreshTokenByAccessTokenID get token by id
func GetRefreshTokenByAccessTokenID(id string) (*RefreshToken, error) {
	var token RefreshToken
	err := db.Where("access_token_id = ? AND status = ?", id, "A").First(&token).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if token.ID == "" {
		return nil, nil
	}
	return &token, nil
}

// ExistRefreshTokenByID check if token exists
// func ExistRefreshTokenByID(id string) (bool, error) {
// 	var token RefreshToken
// 	err := db.Select("id").Where("id = ? AND status = ?", id, "A").First(&token).Error
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}
// 	if token.ID != "" {
// 		return true, nil
// 	}
// 	return false, nil
// }

// StoreRefreshToken store refresh token
func StoreRefreshToken(tx *gorm.DB, id string, accessTokenID string, status string, exp time.Time) error {
	nowTime := time.Now()
	token := RefreshToken{
		ID:            id,
		AccessTokenID: accessTokenID,
		Status:        status,
		CreatedAt:     nowTime,
		UpdatedAt:     nowTime,
		ExpiresAt:     exp,
	}
	if err := tx.Create(&token).Error; err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GenerateRefreshTokenID generate unique token id
func GenerateRefreshTokenID() (string, error) {
	var count int
	for {
		var token RefreshToken
		id := "RT-" + uuid.New().String()
		err := db.Select("id").Where("id = ?", id).First(&token).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if token.ID == "" {
			return id, nil
		}

		if count >= 20 {
			ErrorLog("GenerateRefreshTokenID", "generate refresh token id error", nil)
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.GENERATE_REFRESH_TOKEN_ID_ERROR}
		}
		count++
	}
}

// // GetAccessToken get access token linked with this refresh token
// func (rt *RefreshToken) GetAccessToken() (*AccessToken, error) {
// 	at, err := GetAccessTokenByID(rt.AccessTokenID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if at == nil { // refresh token must link with an access token
// 		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.ACCESS_TOKEN_NOT_FOUND}
// 	}
// 	return at, nil
// }

// // Revoke revoke token
// func (rt *RefreshToken) Revoke(tx *gorm.DB) error {
// 	rt.Status = "R"
// 	return SaveTx(tx, rt)
// }

// // RevokeBothToken revoke access token and refresh token
// func (rt *RefreshToken) RevokeBothToken(tx *gorm.DB) error {
// 	at, err := rt.GetAccessToken()
// 	if err != nil {
// 		return err
// 	}

// 	err = at.Revoke(tx)
// 	if err != nil {
// 		return err
// 	}

// 	return rt.Revoke(tx)
// }

// GetRefreshTokenFn get access_token data with dynamic condition
func GetRefreshTokenFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*RefreshToken, error) {
	var refreshToken []*RefreshToken
	tx := db.Table("refresh_token")
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
	err := tx.Find(&refreshToken).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return refreshToken, nil
}
