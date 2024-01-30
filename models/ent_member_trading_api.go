package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberTradingApi struct
type EntMemberTradingApi struct {
	ID            int       `gorm:"primary_key" json:"id"`
	MemberID      int       `json:"member_id"`
	Platform      string    `json:"platform"`
	PlatformCode  string    `json:"platform_code"`
	Module        string    `json:"module"`
	ApiKey        string    `json:"api_key"`
	ApiSecret     string    `json:"api_secret"`
	ApiPassphrase string    `json:"api_passphrase"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetEntMemberTradingApiFn
func GetEntMemberTradingApiFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberTradingApi, error) {
	var result []*EntMemberTradingApi
	tx := db.Table("ent_member_trading_api").
		Select("ent_member_trading_api.member_id, sys_trading_api_platform.code as platform_code, sys_trading_api_platform.name as platform, ent_member_trading_api.module, ent_member_trading_api.api_key, ent_member_trading_api.api_secret, ent_member_trading_api.api_passphrase, ent_member_trading_api.status, ent_member_trading_api.created_at, ent_member_trading_api.updated_at").
		Joins("inner join sys_trading_api_platform ON sys_trading_api_platform.code = ent_member_trading_api.platform")

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

// AddEntMemberTradingApiStruct struct
type AddEntMemberTradingApiStruct struct {
	ID            int    `gorm:"primary_key" json:"id"`
	MemberID      int    `json:"member_id"`
	Platform      string `json:"platform"`
	Module        string `json:"module"`
	ApiKey        string `json:"api_key"`
	ApiSecret     string `json:"api_secret"`
	ApiPassphrase string `json:"api_passphrase"`
	Status        string `json:"status"`
	CreatedBy     string `json:"created_by"`
}

// AddEntMemberTradingApi add sms log
func AddEntMemberTradingApi(tx *gorm.DB, addEntMemberTradingApi AddEntMemberTradingApiStruct) (*AddEntMemberTradingApiStruct, error) {
	if err := tx.Table("ent_member_trading_api").Create(&addEntMemberTradingApi).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &addEntMemberTradingApi, nil
}
