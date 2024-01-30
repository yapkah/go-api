package app_version_service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
)

type AppVersListForm struct {
	Platform string `form:"platform" json:"platform" valid:"MaxSize(15)"`
	Latest   string `form:"latest" json:"latest"`
}

// func GetAppVersListv1
func GetAppVersListv1(arrData AppVersListForm) []*models.AppAppVersion {

	arrCond := make([]models.WhereCondFn, 0)

	if arrData.Latest != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "b_latest = ?", CondValue: arrData.Latest},
		)
	}
	if arrData.Platform != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "platform = ?", CondValue: arrData.Platform},
		)
	}
	result, err := models.GetAppAppVersionFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("GetAppVersListv1-:failed_to_get_app_app_version", err.Error(), arrCond, true)
	}

	return result
}

// ProcessAppVersionForm struct
type ProcessAppVersionForm struct {
	AppName          string `form:"app_name" json:"app_name"`
	Platform         string `form:"platform" json:"platform" valid:"MaxSize(15)"`
	AppVersion       string `form:"app_version" json:"app_version" valid:"Required"`
	Maintenance      int    `form:"maintenance" json:"maintenance"`
	StoreURL         string `form:"store_url" json:"store_url"`
	WebsiteURL       string `form:"website_url" json:"website_url"`
	Path             string `form:"path" json:"path"`
	FolderPath       string `form:"folder_path" json:"folder_path"`
	Latest           int    `form:"latest" json:"latest"`
	LowestAppVersion int    `form:"lowest_app_version" json:"lowest_app_version"`
}

// func ProcessAppVersion
func ProcessAppVersion(arrData ProcessAppVersionForm) error {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " app_version = ?", CondValue: arrData.AppVersion},
		models.WhereCondFn{Condition: "platform = ?", CondValue: arrData.Platform},
	)
	result, err := models.GetAppAppVersionFn(arrCond, false)

	if len(result) > 0 {
		// perform update action
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond,
			models.WhereCondFn{Condition: " app_version = ?", CondValue: arrData.AppVersion},
			models.WhereCondFn{Condition: " platform = ?", CondValue: arrData.Platform},
		)
		updateColumn := make(map[string]interface{}, 0)
		updateColumn["maintenance"] = arrData.Maintenance
		updateColumn["b_latest"] = arrData.Latest

		if arrData.AppName != "" {
			if arrData.AppName == `""` {
				updateColumn["app_name"] = ""
			} else {
				updateColumn["app_name"] = arrData.AppName
			}
		}
		if arrData.StoreURL != "" {
			if arrData.StoreURL == `""` {
				updateColumn["store_url"] = ""
			} else {
				updateColumn["store_url"] = arrData.StoreURL
			}
		}
		if arrData.LowestAppVersion > 0 {
			updateColumn["lowest_vers"] = arrData.LowestAppVersion
		}
		if arrData.Path != "" {
			updateColumn["path"] = arrData.Path
		}
		if arrData.WebsiteURL != "" {
			if arrData.WebsiteURL == `""` {
				updateColumn["website_url"] = ""
			} else {
				updateColumn["website_url"] = arrData.WebsiteURL
			}
			// start need this bcz no file is upload to server
			// if strings.ToLower(arrData.Platform) == "ios" {
			// 	if arrData.Path == "" {
			// 		updateColumn["path"] = updateColumn["website_url"]
			// 	}
			// }
			// end need this bcz no file is upload to server
		}
		if arrData.FolderPath != "" {
			updateColumn["folder_path"] = arrData.FolderPath
		}
		err := models.UpdatesFn("app_app_version", arrUpdCond, updateColumn, false)

		if err != nil {
			base.LogErrorLog("ProcessAppVersion-update_app_app_version_failed", err.Error(), updateColumn, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_update_app_version"}
		}
	} else {
		// perform create action
		arrProcessData := models.AddAppAppVersionStruct{}
		arrProcessData.Maintenance = arrData.Maintenance
		arrProcessData.BLatest = arrData.Latest

		if arrData.AppName != "" && arrData.AppName != `""` {
			arrProcessData.AppName = arrData.AppName
		}
		if arrData.StoreURL != "" && arrData.StoreURL != `""` {
			arrProcessData.StoreURL = arrData.StoreURL
		}
		if arrData.LowestAppVersion > 0 {
			arrProcessData.LowestVers = arrData.LowestAppVersion
		}
		if arrData.Path != "" {
			arrProcessData.Path = arrData.Path
		}
		if arrData.WebsiteURL != "" && arrData.WebsiteURL != `""` {
			arrProcessData.WebsiteURL = arrData.WebsiteURL
			// start need this bcz no file is upload to server
			if strings.ToLower(arrData.Platform) == "ios" {
				if arrData.Path == "" {
					arrProcessData.Path = arrProcessData.WebsiteURL
				}
			}
			// end need this bcz no file is upload to server
		}
		if arrData.FolderPath != "" {
			arrProcessData.FolderPath = arrData.FolderPath
		}
		arrProcessData.Platform = arrData.Platform
		arrProcessData.AppVersion = arrData.AppVersion

		_, err := models.AddAppAppVersion(arrProcessData)
		if err != nil {
			base.LogErrorLog("ProcessAppVersion-AddAppAppVersion_failed", err.Error(), arrProcessData, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_create_app_version"}
		}
	}

	if err != nil {
		base.LogErrorLog("ProcessAppVersion-failed_to_process_app_version", err.Error(), arrData, true)
	}

	return nil
}

type CheckAppVers struct {
	Platform   string
	AppVersion string
	LangCode   string
	Source     uint8
}

type AppVersRst struct {
	AppName        string `json:"app_name"`
	Platform       string `json:"platform"`
	AppVersion     string `json:"app_version"`
	Maintenance    int    `json:"maintenance"`
	StoreURL       string `json:"store_url"`
	WebsiteURL     string `json:"website_url"`
	BLatest        int    `json:"latest"`
	LowestVers     int    `json:"lowest_app_version"`
	FileURL        string `json:"file_url"`
	Update         int    `json:"update"`
	MaintenanceMsg string `json:"maintenance_msg"`
}

// func CheckAppVersv1
func CheckAppVersv1(arrData CheckAppVers) (*AppVersRst, bool, error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "b_latest = ?", CondValue: 1},
	)

	if arrData.Platform != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "platform = ?", CondValue: arrData.Platform},
		)
	}
	result, err := models.GetAppAppVersionFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("CheckAppVersv1-failed_to_get_app_app_version", err.Error(), arrCond, true)
		return nil, false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_related_records"}
	}

	if len(result) > 1 {
		return nil, false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "more_than_1_latest_app_version_is_set"}
	}
	if len(result) < 1 {
		return nil, false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_related_records"}
	}

	arrInputAppVers := strings.Split(arrData.AppVersion, ".")
	arrLatAppVers := strings.Split(result[0].AppVersion, ".")

	arrDataReturn := AppVersRst{
		AppName:     result[0].AppName,
		Platform:    result[0].Platform,
		AppVersion:  result[0].AppVersion,
		Maintenance: result[0].Maintenance,
		StoreURL:    result[0].StoreURL,
		WebsiteURL:  result[0].WebsiteURL,
		BLatest:     result[0].BLatest,
		LowestVers:  result[0].LowestVers,
		FileURL:     result[0].FileURL,
	}

	if arrData.Source == 1 {
		arrDataReturn.Update = 0
		return &arrDataReturn, true, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "latest_app_version"}
	}

	if len(arrLatAppVers) != len(arrInputAppVers) {
		return &arrDataReturn, false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_app_version"}
	}

	errMsg := ""
	var errStatus bool
	for i := 0; i < len(arrInputAppVers); i++ {
		maintenanceMsg := ""
		if result[0].Maintenance == 1 {
			maintenanceMsg = "app_is_under_maintenance"
		}
		arrDataReturn.MaintenanceMsg = maintenanceMsg
		latAppVersInt, _ := strconv.Atoi(arrLatAppVers[i])
		inputAppVersInt, _ := strconv.Atoi(arrInputAppVers[i])
		if latAppVersInt > inputAppVersInt {
			arrDataReturn.Update = 1
			errMsg = "old_app_version"
			errStatus = true
			break
		} else if inputAppVersInt > latAppVersInt {
			arrDataReturn.Update = 0
			errMsg = "latest_app_version"
			errStatus = true
			break
		} else {
			arrDataReturn.Update = 0
			errMsg = "latest_app_version"
			errStatus = true
		}
	}

	return &arrDataReturn, errStatus, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: errMsg}
}
