package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/setting"
)

// AppAppVersion struct
type AppAppVersion struct {
	ID             string `gorm:"primary_key" gorm:"column:id" json:"-"`
	AppName        string `gorm:"column:app_name" json:"app_name"`
	Platform       string `gorm:"column:platform" json:"platform"`
	AppVersion     string `gorm:"column:app_version" json:"app_version"`
	Maintenance    int    `gorm:"column:maintenance" json:"maintenance"`
	StoreURL       string `gorm:"column:store_url" json:"store_url"`
	WebsiteURL     string `gorm:"column:website_url" json:"website_url"`
	Path           string `gorm:"column:path" json:"-"`
	FolderPath     string `gorm:"column:folder_path" json:"-"`
	Custom1        string `gorm:"column:custom_1" json:"-"`
	BLatest        int    `gorm:"column:b_latest" json:"latest"`
	LowestVers     int    `gorm:"column:lowest_vers" json:"lowest_app_version"`
	FileURL        string `gorm:"column:file_url" json:"file_url"`
	RegisterStatus int    `gorm:"column:register_status" json:"register_status"`
}

// GetLatestAppAppVersionFn func
func GetLatestAppAppVersionFn(arrCond []WhereCondFn, debug bool) ([]*AppAppVersion, error) {
	var result []*AppAppVersion
	tx := db.Table("app_app_version")

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

// GetAppAppVersionFn func
func GetAppAppVersionFn(arrCond []WhereCondFn, debug bool) ([]*AppAppVersion, error) {
	memberServerDomain := setting.Cfg.Section("custom").Key("MemberServerDomain").String()
	var result []*AppAppVersion
	tx := db.Table("app_app_version").
		Select("app_app_version.id, app_name, platform, app_version, maintenance, store_url, path, folder_path, custom_1, b_latest, lowest_vers," +
			" CASE " +
			" WHEN app_app_version.website_url IS NOT NULL AND app_app_version.website_url != '' THEN website_url" +
			" ELSE '" + memberServerDomain + "/download-mobile-app'" +
			" END AS website_url," +
			" app_app_version.path AS 'file_url'").
		Order("dt_timestamp DESC")

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

// AddAppAppVersionStruct struct
type AddAppAppVersionStruct struct {
	ID          string `gorm:"primary_key" gorm:"column:id" json:"-"`
	AppName     string `gorm:"column:app_name" json:"app_name"`
	Platform    string `gorm:"column:platform" json:"platform"`
	AppVersion  string `gorm:"column:app_version" json:"app_version"`
	Maintenance int    `gorm:"column:maintenance" json:"maintenance"`
	StoreURL    string `gorm:"column:store_url" json:"store_url"`
	WebsiteURL  string `gorm:"column:website_url" json:"website_url"`
	Path        string `gorm:"column:path" json:"-"`
	FolderPath  string `gorm:"column:folder_path" json:"-"`
	Custom1     string `gorm:"column:custom_1" json:"-"`
	BLatest     int    `gorm:"column:b_latest" json:"latest"`
	LowestVers  int    `gorm:"column:lowest_vers" json:"lowest_app_version"`
}

// AddAppAppVersion add app_app_version
func AddAppAppVersion(arrData AddAppAppVersionStruct) (*AddAppAppVersionStruct, error) {
	if err := db.Table("app_app_version").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}
