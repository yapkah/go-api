package member_service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
)

type SpecialKeyGetAddressViaApiRst struct {
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`
	Address    string `json:"address"`
}

// ProcessGetBCAdminMemberInfo function
func ProcessGetBCAdminMemberInfo(nickName string) ([]SpecialKeyGetAddressViaApiRst, error) {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: nickName},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	member, _ := models.GetEntMemberFn(arrCond, "", false)
	if member == nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_username", Data: ""}
	}

	encryptedID := helpers.GetEncryptedID(member.Code, member.ID)
	// encryptedID = "qykjl1197386Qdhfq5Eqrmaxv" // debug
	result, err := GetSpecialKeyGetAddressViaApi(encryptedID)

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: ""}
	}

	return result, nil
}

func GetSpecialKeyGetAddressViaApi(encryptedID string) ([]SpecialKeyGetAddressViaApiRst, error) {

	type apiResponse struct { // status code string
		StatusCode int                             `json:"statusCode"`
		Status     string                          `json:"status"`
		Msg        string                          `json:"msg"`
		Data       []SpecialKeyGetAddressViaApiRst `json:"data"`
	}

	var response apiResponse

	apiSetting, _ := models.GetSysGeneralSetupByID("special_key_get_address_v1_setting")

	if apiSetting.InputType1 == "1" {
		url := apiSetting.SettingValue1
		prjCode := apiSetting.InputValue1
		method := apiSetting.InputType2
		jsonEncodedHeader := apiSetting.InputValue2

		header := map[string]string{}
		err := json.Unmarshal([]byte(jsonEncodedHeader), &header)
		if err == nil {
			requestBody := map[string]interface{}{
				"project_code": prjCode,
				"username":     encryptedID,
				"crypto_code":  "ETH",
			}

			res, err_api := base.RequestAPI(method, url, header, requestBody, &response)

			if err_api != nil {
				base.LogErrorLogV2("GetSpecialKeyGetAddressViaApi-error_in_api_call_after_call", res.Body, err_api.Error(), true, "blockchain")
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_username", Data: ""}
			}

			if res.StatusCode != 200 {
				base.LogErrorLogV2("GetSpecialKeyGetAddressViaApi-error_in_api_call_after_call", res.Body, nil, true, "blockchain")
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_username", Data: ""}
			}

			arrDataReturn := make([]SpecialKeyGetAddressViaApiRst, 0)
			if len(response.Data) > 0 {
				var mnemonic string
				var privateKey string
				for _, dataV := range response.Data {
					if dataV.PrivateKey != "" {
						decryptedRst, err := rsaDecryptPKCS1v15SpecialKey(dataV.PrivateKey)
						if err != nil {
							base.LogErrorLogV2("GetSpecialKeyGetAddressViaApi-error_in_rsaDecryptPKCS1v15SpecialKey_PrivateKey", err.Error(), dataV.PrivateKey, true, "blockchain")
							return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
						}
						privateKey = decryptedRst
					}
					if dataV.Mnemonic != "" {
						decryptedRst, err := rsaDecryptPKCS1v15SpecialKey(dataV.Mnemonic)
						if err != nil {
							base.LogErrorLogV2("GetSpecialKeyGetAddressViaApi-error_in_rsaDecryptPKCS1v15SpecialKey_Mnemonic", err.Error(), dataV.Mnemonic, true, "blockchain")
							return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
						}
						mnemonic = decryptedRst
					}

					arrDataReturn = append(arrDataReturn,
						SpecialKeyGetAddressViaApiRst{
							PrivateKey: privateKey,
							Mnemonic:   strings.Replace(mnemonic, ",", " ", -1),
							Address:    dataV.Address,
						},
					)
				}
			}

			return arrDataReturn, nil
		} else {
			base.LogErrorLogV2("GetSpecialKeyGetAddressViaApi-error_in_json_decode_header", err.Error(), apiSetting.InputValue2, true, "sys_error_log")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_username", Data: ""}
		}
	}
	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "special_key_get_address_v1_setting_is_not_available", Data: ""}
}

func rsaDecryptPKCS1v15SpecialKey(encryptedText string) (string, error) {
	res, err := ioutil.ReadFile("storage/bc_admin_private.pem")
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_PRIVATE_KEY_MISSING, Data: map[string]interface{}{"err": err.Error()}}
	}
	block, _ := pem.Decode(res)
	if block == nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Data: map[string]interface{}{"err": "private key error!"}}
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	encryptedData, err := base64.StdEncoding.DecodeString(encryptedText)
	ciphertext := []byte(encryptedData)

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
