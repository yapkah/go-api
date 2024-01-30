package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SlsMasterBotSetting struct
type SlsMasterBotSetting struct {
	ID            int       `gorm:"primary_key" json:"id"`
	MemberID      int       `json:"member_id" gorm:"column:member_id"`
	SlsMasterID   int       `json:"sls_master_id" gorm:"column:sls_master_id"`
	Platform      string    `json:"platform" gorm:"column:platform"`
	PrdMasterID   int       `json:"prd_master_id" gorm:"column:prd_master_id"`
	PrdMasterCode string    `json:"prd_master_code" gorm:"column:prd_master_code"`
	PrdMasterName string    `json:"prd_master_name" gorm:"column:prd_master_name"`
	DocNo         string    `json:"doc_no" gorm:"column:doc_no"`
	RefNo         string    `json:"ref_no" gorm:"column:ref_no"`
	SettingType   string    `json:"setting_type" gorm:"column:setting_type"`
	CryptoPair    string    `json:"crypto_pair" gorm:"column:crypto_pair"`
	Setting       string    `json:"setting" gorm:"column:setting"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
}

func GetSlsMasterBotSetting(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMasterBotSetting, error) {
	var result []*SlsMasterBotSetting
	tx := db.Table("sls_master_bot_setting").
		Joins("INNER JOIN sls_master ON sls_master.id = sls_master_bot_setting.sls_master_id").
		Joins("INNER JOIN prd_master ON prd_master.id = sls_master.prd_master_id").
		Select("sls_master_bot_setting.*, sls_master.member_id, sls_master.doc_no, sls_master.ref_no, prd_master.code as prd_master_code, prd_master.name as prd_master_name")

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

// AddSlsMasterBotSettingStruct struct
type AddSlsMasterBotSettingStruct struct {
	ID          int       `gorm:"primary_key" json:"id"`
	SlsMasterID int       `json:"sls_master_id" gorm:"column:sls_master_id"`
	Platform    string    `json:"platform" gorm:"column:platform"`
	SettingType string    `json:"setting_type" gorm:"column:setting_type"`
	CryptoPair  string    `json:"crypto_pair" gorm:"column:crypto_pair"`
	Setting     string    `json:"setting" gorm:"column:setting"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

// AddSlsMasterBotSetting func
func AddSlsMasterBotSetting(tx *gorm.DB, slsMasterBotSetting AddSlsMasterBotSettingStruct) (*AddSlsMasterBotSettingStruct, error) {
	if err := tx.Table("sls_master_bot_setting").Create(&slsMasterBotSetting).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMasterBotSetting, nil
}
