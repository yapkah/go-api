package crypto_service

import (
	"strings"

	"github.com/smartblock/gta-api/pkg/base"
)

type HealthApiResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func GetHealthStatus(ip string) (*HealthApiResponse, string) {
	var response HealthApiResponse

	// arrGeneralSetup, err := models.GetSysGeneralSetupByID("swarm_api_setting")
	// if err != nil {
	// 	base.LogErrorLog("cryptoService:GetHealthStatus():GetSysGeneralSetupByID():1", err.Error(), map[string]interface{}{"settingID": "swarm_api_setting"}, true)
	// 	return nil, "something_went_wrong"
	// }
	// if arrGeneralSetup == nil {
	// 	base.LogErrorLog("cryptoService:GetHealthStatus():GetSysGeneralSetupByID():2", "swarm_api_setting_not_found", map[string]interface{}{"settingID": "swarm_api_setting"}, true)
	// 	return nil, "something_went_wrong"
	// }

	input := map[string]interface{}{}

	if !strings.Contains(ip, "http://") {
		ip = "http://" + ip
	}
	url := ip + "/health"
	header := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	res, err := base.RequestAPI("GET", url, header, input, &response)

	if err != nil {
		base.LogErrorLog("cryptoService:GetHealthStatus():RequestAPI():1", map[string]interface{}{"url": url, "header": header, "input": input}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("cryptoService:GetHealthStatus():RequestAPI():2", map[string]interface{}{"url": url, "header": header, "input": input, "response": res.Body}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	return &response, ""
}

type ContractAddressApiResponse struct {
	Address string `json:"chequebookAddress"`
}

func GetContractAddress(ip string) (*ContractAddressApiResponse, string) {
	var response ContractAddressApiResponse

	input := map[string]interface{}{}

	if !strings.Contains(ip, "http://") {
		ip = "http://" + ip
	}
	url := ip + "/chequebook/address"
	header := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	res, err := base.RequestAPI("GET", url, header, input, &response)

	if err != nil {
		base.LogErrorLog("cryptoService:GetContractAddress():RequestAPI():1", map[string]interface{}{"url": url, "header": header, "input": input}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("cryptoService:GetContractAddress():RequestAPI():2", map[string]interface{}{"url": url, "header": header, "input": input, "response": res.Body}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	return &response, ""
}

type WalletAddressApiResponse struct {
	Ethereum  string `json:"ethereum"`
	PublicKey string `json:"publicKey"`
}

func GetWalletAddress(ip string) (*WalletAddressApiResponse, string) {
	var response WalletAddressApiResponse

	input := map[string]interface{}{}

	if !strings.Contains(ip, "http://") {
		ip = "http://" + ip
	}
	url := ip + "/addresses"
	header := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	res, err := base.RequestAPI("GET", url, header, input, &response)

	if err != nil {
		base.LogErrorLog("cryptoService:GetWalletAddress():RequestAPI():1", map[string]interface{}{"url": url, "header": header, "input": input}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("cryptoService:GetWalletAddress():RequestAPI():2", map[string]interface{}{"url": url, "header": header, "input": input, "response": res.Body}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	return &response, ""
}

type SettlementsApiResponse struct {
	TotalReceived string `json:"totalReceived"`
	TotalSent     string `json:"totalSent"`
}

func GetSettlements(ip string) (*SettlementsApiResponse, string) {
	var response SettlementsApiResponse

	input := map[string]interface{}{}

	if !strings.Contains(ip, "http://") {
		ip = "http://" + ip
	}
	url := ip + "/settlements"
	header := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	res, err := base.RequestAPI("GET", url, header, input, &response)

	if err != nil {
		base.LogErrorLog("cryptoService:GetSettlements():RequestAPI():1", map[string]interface{}{"url": url, "header": header, "input": input}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	if res.StatusCode != 200 {
		base.LogErrorLog("cryptoService:GetSettlements():RequestAPI():2", map[string]interface{}{"url": url, "header": header, "input": input, "response": res.Body}, err.Error(), false)
		return nil, "something_went_wrong"
	}

	return &response, ""
}
