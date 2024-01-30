package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// ApiKeys struct
type ApiKeys struct {
	ID          int       `gorm:"primary_key" json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	SaltKey     string    `json:"salt_key"`
	OauthPubKey string    `json:"oauth_pub_key"`
	OauthType   string    `json:"oauth_type"`
	SourceID    int       `json:"source_id"`
	Active      int       `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// GetApiKeysFn get api_keys data with dynamic condition
func GetApiKeysFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*ApiKeys, error) {
	var result []*ApiKeys
	tx := db.Table("api_keys")
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
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
