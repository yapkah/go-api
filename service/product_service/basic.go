package product_service

import (
	"encoding/json"
	"time"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
)

// PrdMasterSetting struct
type PrdMasterSetting struct {
	Years  int `json:"years"`
	Months int `json:"months"`
	Days   int `json:"days"`
}

// GetPrdMasterSetup func
func GetPrdMasterSetup(rawPrdMasterSettingInput string) (PrdMasterSetting, string) {
	var prdMasterSetupPointer = &PrdMasterSetting{}

	if rawPrdMasterSettingInput == "" {
		return *prdMasterSetupPointer, ""
	}

	err := json.Unmarshal([]byte(rawPrdMasterSettingInput), prdMasterSetupPointer)
	if err != nil {
		base.LogErrorLog("productService:GetPrdMasterSetup():Unmarshal():1", err.Error(), map[string]interface{}{"rawPrdMasterSettingInput": rawPrdMasterSettingInput}, true)
		return PrdMasterSetting{}, "something_went_wrong"
	}

	return *prdMasterSetupPointer, ""
}

// PrdGroupTypeSetup struct
type PrdGroupTypeSetup struct {
	KeyinMin        float64       `json:"keyin_min"`
	KeyinMultipleOf float64       `json:"keyin_multiple_of"`
	SalesType       string        `json:"sales_type"`
	DelayTime       time.Duration `json:"delay_time"`
	Tiers           TiersStruct   `json:"tiers"`
}

type TiersStruct []struct {
	Tier     string  `json:"tier"`
	Title    string  `json:"title"`
	Min      float64 `json:"min"`
	Bonus    float64 `json:"bonus"`
	Leverage float64 `json:"leverage"`
}

// GetPrdGroupTypeSetup func
func GetPrdGroupTypeSetup(rawPrdGroupTypeSetupInput string) (PrdGroupTypeSetup, string) {
	prdGroupSetupPointer := &PrdGroupTypeSetup{}
	if rawPrdGroupTypeSetupInput == "" {
		return *prdGroupSetupPointer, ""
	}

	err := json.Unmarshal([]byte(rawPrdGroupTypeSetupInput), prdGroupSetupPointer)
	if err != nil {
		base.LogErrorLog("productService:GetPrdGroupTypeSetup():Unmarshal():1", err.Error(), map[string]interface{}{"rawPrdGroupTypeSetupInput": rawPrdGroupTypeSetupInput}, true)
		return PrdGroupTypeSetup{}, "something_went_wrong"
	}

	return *prdGroupSetupPointer, ""
}

// PrdGroupTypeRefundSetup struct
type PrdGroupTypeRefundSetup struct {
	Status                bool                             `json:"status"`
	Type                  string                           `json:"type"`
	RefundEwalletTypeCode string                           `json:"refund_ewallet_type_code"`
	PenaltyPercDef        int                              `json:"penalty_perc_def"`
	Penalty               []PrdGroupTypeRefundSetupPenalty `json:"penalty"`
}

type PrdGroupTypeRefundSetupPenalty struct {
	Label       string `json:"label"`
	Min         int    `json:"min"`
	PenaltyPerc int    `json:"penalty_perc"`
}

// GetPrdGroupTypeRefundSetup func
func GetPrdGroupTypeRefundSetup(rawPrdGroupTypeRefundSetupInput string) (PrdGroupTypeRefundSetup, string) {
	prdGroupRefundSetupPointer := &PrdGroupTypeRefundSetup{}
	if rawPrdGroupTypeRefundSetupInput == "" {
		return *prdGroupRefundSetupPointer, ""
	}

	err := json.Unmarshal([]byte(rawPrdGroupTypeRefundSetupInput), prdGroupRefundSetupPointer)
	if err != nil {
		base.LogErrorLog("productService:GetPrdGroupTypeRefundSetup():Unmarshal():1", err.Error(), map[string]interface{}{"rawPrdGroupTypeRefundSetupInput": rawPrdGroupTypeRefundSetupInput}, true)
		return PrdGroupTypeRefundSetup{}, "something_went_wrong"
	}

	return *prdGroupRefundSetupPointer, ""
}

// PostStakingSetup struct
type PostStakingSetup struct {
	MultipleOf float64
	DelayTime  time.Duration
}

// GetPostStakingSetup func
func GetPostStakingSetup() (PostStakingSetup, string) {
	var (
		postStakingSetup PostStakingSetup
	)

	// RawPostStakingSetup struct
	type RawPostStakingSetup struct {
		MultipleOf string `json:"multiple_of"`
		DelayTime  string `json:"delay_time"`
	}

	// get post staking setting
	arrGeneralSetup, err := models.GetSysGeneralSetupByID("post_staking_setting")
	if err != nil {
		base.LogErrorLog("productService:GetPostStakingSetup()", err.Error(), "post_staking_setting", true)
		return PostStakingSetup{}, "something_went_wrong"
	}
	if arrGeneralSetup == nil {
		base.LogErrorLog("productService:GetPostStakingSetup()", "GetSysGeneralSetupByID():1", "post_staking_setting_not_found", true)
		return PostStakingSetup{}, "something_went_wrong"
	}

	// mapping post staking setting into struct
	rawPostStakingSetup := &RawPostStakingSetup{}
	err = json.Unmarshal([]byte(arrGeneralSetup.SettingValue1), rawPostStakingSetup)
	if err != nil {
		base.LogErrorLog("productService:GetPostStakingSetup()", "Unmarshal():1", err.Error(), true)
		return PostStakingSetup{}, "something_went_wrong"
	}

	// after map convert to relative type
	multipleOf, err := helpers.ValueToFloat(rawPostStakingSetup.MultipleOf)
	if err != nil {
		base.LogErrorLog("productService:GetPostStakingSetup()", "ValueToFloat():1", err.Error(), true)
		return PostStakingSetup{}, "something_went_wrong"
	}

	postStakingSetup.MultipleOf = multipleOf

	// calculate payable_at for contract queue
	maxDelayTime, err := helpers.ValueToDuration(rawPostStakingSetup.DelayTime)
	if err != nil {
		base.LogErrorLog("productService:GetPostStakingSetup()", "ValueToDuration():1", err.Error(), true)
		return PostStakingSetup{}, "something_went_wrong"
	}
	postStakingSetup.DelayTime = maxDelayTime

	return postStakingSetup, ""
}

// ProductTopupSetting struct
type ProductTopupSetting struct {
	Status     bool
	MultipleOf float64
}

// GetProductTopupSetting func
func GetProductTopupSetting(rawProductTopupSettingStr string) (ProductTopupSetting, string) {
	var (
		productTopupSetup ProductTopupSetting
	)

	// RawProductTopupSetting struct
	type RawProductTopupSetting struct {
		Status     string `json:"status"`
		MultipleOf string `json:"multiple_of"`
	}

	// mapping purchase contract setting into struct
	rawProductTopupSetting := &RawProductTopupSetting{}
	err := json.Unmarshal([]byte(rawProductTopupSettingStr), rawProductTopupSetting)
	if err != nil {
		base.LogErrorLog("productService:GetProductTopupSetting()", "Unmarshal():1", err.Error(), true)
		return ProductTopupSetting{}, "something_went_wrong"
	}

	// after map convert to relative type
	status, err := helpers.ValueToBool(rawProductTopupSetting.Status)
	if err != nil {
		base.LogErrorLog("productService:GetProductTopupSetting()", "ValueToBool():1", err.Error(), true)
		return ProductTopupSetting{}, "something_went_wrong"
	}

	multipleOf, err := helpers.ValueToFloat(rawProductTopupSetting.MultipleOf)
	if err != nil {
		base.LogErrorLog("productService:GetProductTopupSetting()", "ValueToFloat():1", err.Error(), true)
		return ProductTopupSetting{}, "something_went_wrong"
	}

	productTopupSetup.Status = status
	productTopupSetup.MultipleOf = multipleOf

	return productTopupSetup, ""
}
