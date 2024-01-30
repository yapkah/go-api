package job_service

import (
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
)

// func ProcessTreeQJobService
func ProcessTreeQJobService() {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "api_keys.name = ?", CondValue: "admin"},
		models.WhereCondFn{Condition: "api_keys.active = ?", CondValue: 1},
	)
	result, err := models.GetApiKeysFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("job_service|ProcessTreeQJobService():failed_in_get_api_key", arrCond, err.Error(), true)
	}
	// start exclusive for admin
	if len(result) > 0 {
		adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()
		url := adminServerDomain + "/api/job/process-tree-q"
		header := map[string]string{
			"X-Authorization": result[0].Key,
		}

		extraSetting := base.ExtraSettingStruct{
			InsecureSkipVerify: true,
		}

		_, err := base.RequestAPIV2("GET", url, header, nil, nil, extraSetting)
		// _, err = base.RequestAPI("GET", url, header, nil, nil)

		if err != nil {
			base.LogErrorLog("job_service|ProcessTreeQJobService()", "failed_in_get_api", err.Error(), true)
			RebuildTreeJobService()
		}
	} else {
		base.LogErrorLog("job_service|ProcessTreeQJobService()", "missing_X_Authorization_api_key", arrCond, true)
	}
}

// func RebuildTreeJobService
func RebuildTreeJobService() {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "api_keys.name = ?", CondValue: "admin"},
		models.WhereCondFn{Condition: "api_keys.active = ?", CondValue: 1},
	)
	result, err := models.GetApiKeysFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("job_service|RebuildTreeJobService():failed_in_get_api_key", arrCond, err.Error(), true)
	}
	// start exclusive for admin
	if len(result) > 0 {
		adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()
		url := adminServerDomain + "/api/job/rebuild-tree"
		header := map[string]string{
			"X-Authorization": result[0].Key,
		}

		extraSetting := base.ExtraSettingStruct{
			InsecureSkipVerify: true,
		}

		_, err := base.RequestAPIV2("GET", url, header, nil, nil, extraSetting)
		// _, err = base.RequestAPI("GET", url, header, nil, nil)

		if err != nil {
			base.LogErrorLog("job_service|RebuildTreeJobService()", "failed_in_get_api", err.Error(), true)
		}
	} else {
		base.LogErrorLog("job_service|RebuildTreeJobService()", "missing_X_Authorization_api_key", arrCond, true)
	}
}
