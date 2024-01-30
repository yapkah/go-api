package wallet_service

import (
	"github.com/yapkah/go-api/models"
)

// AppSettingList struct
type AppSettingList struct {
	EwalletIconURL string `json:"wallet_type_image_url"`
}

// GetDecimalPlacesByEwalletTypeCode func
func GetDecimalPlacesByEwalletTypeCode(ewtTypeCode string) int {
	decimalPlaces := 2

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: ewtTypeCode},
	)
	arrEwtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if arrEwtSetup != nil {
		decimalPlaces = arrEwtSetup.DecimalPoint
	}

	return decimalPlaces
}
