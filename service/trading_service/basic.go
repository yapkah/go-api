package trading_service

import (
	"encoding/json"

	"github.com/smartblock/gta-api/pkg/base"
)

// SysTradingApiPlatformSetting struct
type SysTradingApiPlatformSetting struct {
	Strategy []struct {
		Code   string `json:"code"`
		BgPath string `json:"bg_path"`
	} `json:"strategy"`
}

// GetSysTradingApiPlatformSetting func
func GetSysTradingApiPlatformSetting(rawSysTradingApiPlatformSettingStr string) (SysTradingApiPlatformSetting, string) {
	tradingApiPlatformSetting := SysTradingApiPlatformSetting{}
	err := json.Unmarshal([]byte(rawSysTradingApiPlatformSettingStr), &tradingApiPlatformSetting)
	if err != nil {
		base.LogErrorLog("tradingService:GetSysTradingApiPlatformSetting()", "Unmarshal():1", err.Error(), true)
		return SysTradingApiPlatformSetting{}, "something_went_wrong"
	}

	return tradingApiPlatformSetting, ""
}
