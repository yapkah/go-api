package sales_service

import (
	"encoding/json"
	"time"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/pkg/base"
)

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
	prdGroupTypePointer := &PrdGroupTypeSetup{}
	err := json.Unmarshal([]byte(rawPrdGroupTypeSetupInput), prdGroupTypePointer)
	if err != nil {
		base.LogErrorLog("salesService:GetPrdGroupTypeSetup():Unmarshal():1", err.Error(), map[string]interface{}{"rawPrdGroupTypeSetupInput": rawPrdGroupTypeSetupInput}, true)
		return PrdGroupTypeSetup{}, "something_went_wrong"
	}

	return *prdGroupTypePointer, ""
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

// IncomeCapSetup struct
type IncomeCapSetup struct {
	Status        bool
	EwalletTypeID int
}

// MapProductIncomeCapSetting func
func MapProductIncomeCapSetting(rawIncomeCapSettingInput string) (IncomeCapSetup, string) {
	var incomeCapSetup IncomeCapSetup

	// RawIncomeCapSetup struct
	type RawIncomeCapSetup struct {
		Status        string `json:"status"`
		EwalletTypeID string `json:"ewallet_type_id"`
	}

	if rawIncomeCapSettingInput == "" {
		return IncomeCapSetup{}, ""
	}

	// mapping product income cap setting into struct
	rawIncomeCapSetup := &RawIncomeCapSetup{}
	err := json.Unmarshal([]byte(rawIncomeCapSettingInput), rawIncomeCapSetup)
	if err != nil {
		base.LogErrorLog("salesService:MapProductIncomeCapSetting()", "Unmarshal():1", err.Error(), true)
		return IncomeCapSetup{}, "something_went_wrong"
	}

	// after map convert to relative  type
	status, err := helpers.ValueToBool(rawIncomeCapSetup.Status)
	if err != nil {
		base.LogErrorLog("salesService:MapProductRefundSetting()", "ValueToInt():1", err.Error(), true)
		return IncomeCapSetup{}, "something_went_wrong"
	}
	ewalletTypeID, err := helpers.ValueToInt(rawIncomeCapSetup.EwalletTypeID)
	if err != nil {
		base.LogErrorLog("salesService:GetPurchaseContractSetup()", "ValueToInt():1", err.Error(), true)
		return IncomeCapSetup{}, "something_went_wrong"
	}

	incomeCapSetup.Status = status
	incomeCapSetup.EwalletTypeID = ewalletTypeID

	return incomeCapSetup, ""
}

// ProductTopupSetting struct
type ProductTopupSetting struct {
	Status        bool
	MultipleOf    float64
	HistoryStatus bool
}

// GetProductTopupSetting func
func GetProductTopupSetting(rawProductTopupSettingStr string) (ProductTopupSetting, string) {
	var (
		productTopupSetup ProductTopupSetting
	)

	// RawProductTopupSetting struct
	type RawProductTopupSetting struct {
		Status        string `json:"status"`
		MultipleOf    string `json:"multiple_of"`
		HistoryStatus string `json:"history_status"`
	}

	// mapping purchase contract setting into struct
	rawProductTopupSetting := &RawProductTopupSetting{}
	err := json.Unmarshal([]byte(rawProductTopupSettingStr), rawProductTopupSetting)
	if err != nil {
		base.LogErrorLog("salesService:GetProductTopupSetting()", "Unmarshal():1", err.Error(), true)
		return ProductTopupSetting{}, "something_went_wrong"
	}

	// after map convert to relative type
	status, err := helpers.ValueToBool(rawProductTopupSetting.Status)
	if err != nil {
		base.LogErrorLog("salesService:GetProductTopupSetting()", "ValueToBool():1", err.Error(), true)
		return ProductTopupSetting{}, "something_went_wrong"
	}

	multipleOf, err := helpers.ValueToFloat(rawProductTopupSetting.MultipleOf)
	if err != nil {
		base.LogErrorLog("salesService:GetProductTopupSetting()", "ValueToFloat():1", err.Error(), true)
		return ProductTopupSetting{}, "something_went_wrong"
	}

	var historyStatus = false
	if rawProductTopupSetting.HistoryStatus != "" {
		historyStatus, err = helpers.ValueToBool(rawProductTopupSetting.HistoryStatus)
		if err != nil {
			base.LogErrorLog("salesService:GetProductTopupSetting()", "ValueToBool():2", err.Error(), true)
			return ProductTopupSetting{}, "something_went_wrong"
		}
	}

	productTopupSetup.Status = status
	productTopupSetup.MultipleOf = multipleOf
	productTopupSetup.HistoryStatus = historyStatus

	return productTopupSetup, ""
}
