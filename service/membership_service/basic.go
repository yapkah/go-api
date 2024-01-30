package membership_service

import (
	"encoding/json"

	"github.com/yapkah/go-api/pkg/base"
)

// EntMembershipSetupPeriod struct
type EntMembershipSetupPeriod struct {
	Years  int `json:"years"`
	Months int `json:"months"`
	Days   int `json:"days"`
}

// GetEntMembershipSetupPeriod func
func GetEntMembershipSetupPeriod(rawEntMembershipSetupInput string) (EntMembershipSetupPeriod, string) {
	prdGroupSetupPointer := &EntMembershipSetupPeriod{}
	if rawEntMembershipSetupInput == "" {
		return *prdGroupSetupPointer, ""
	}

	err := json.Unmarshal([]byte(rawEntMembershipSetupInput), prdGroupSetupPointer)
	if err != nil {
		base.LogErrorLog("productService:GetEntMembershipSetupPeriod():Unmarshal():1", err.Error(), map[string]interface{}{"rawEntMembershipSetupInput": rawEntMembershipSetupInput}, true)
		return EntMembershipSetupPeriod{}, "something_went_wrong"
	}

	return *prdGroupSetupPointer, ""
}
