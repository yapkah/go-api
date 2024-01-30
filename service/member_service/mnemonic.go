package member_service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
)

/*
private_key = encryptRSA(sha256(<<ori_private_key>> and <<secret_key>>))
pk = encryptRSA(<<ori_private_key>>)

eg:
ori_private_key = 123456
secret_key = abcd

private_key = encryptRSA(sha256(123456abcd))
pk = encryptRSA(123456)a
*/

// struct CallGenerateMnemonicApiRst
type CallGenerateMnemonicApiRst struct {
	Mnemonic      []string `json:"mnemonic"`
	CryptoAddress string   `json:"crypto_address"`
	PrivateKey    string   `json:"private_key"`
}

// GetMemberSettingStatus func
func CallGenerateMnemonicApi() (*CallGenerateMnemonicApiRst, error) {

	settingID := "mnemonic_api_setting"
	arrMnemonicApiSetting, _ := models.GetSysGeneralSetupByID(settingID)

	if arrMnemonicApiSetting.InputType1 == "1" {
		type mnemonicApiRstStruct struct {
			Status     string `json:"status"`
			StatusCode string `json:"status_code"`
			Msg        string `json:"msg"`
			Data       struct {
				Address    string   `json:"address"`
				PrivateKey string   `json:"private_key"`
				Mnemonic   []string `json:"mnemonic"`
			} `json:"data"`
		}
		url := arrMnemonicApiSetting.InputValue1
		requestMethod := arrMnemonicApiSetting.SettingValue1
		header := map[string]string{
			"X-Authorization": arrMnemonicApiSetting.InputType2,
		}

		res, err_api := base.RequestAPI(requestMethod, url, header, nil, nil)

		if err_api != nil {
			base.LogErrorLogV2("CallGenerateMnemonicApi_error_in_api_call_before_call", err_api.Error(), nil, true, "blockchain")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "mnemonic_is_not_available", Data: ""}
		}

		if res.StatusCode != 200 {
			base.LogErrorLogV2("CallGenerateMnemonicApi_error_in_api_call_after_call", res.Body, nil, true, "blockchain")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "mnemonic_is_not_available", Data: ""}
		}

		var mnemonicApiRst mnemonicApiRstStruct
		err := json.Unmarshal([]byte(res.Body), &mnemonicApiRst)

		if err != nil {
			base.LogErrorLog("CallGenerateMnemonicApi_error_in_json_decode_api_result", err_api.Error(), res.Body, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "mnemonic_is_not_available", Data: ""}
		}

		arrDataReturn := CallGenerateMnemonicApiRst{
			Mnemonic:      mnemonicApiRst.Data.Mnemonic,
			PrivateKey:    mnemonicApiRst.Data.PrivateKey,
			CryptoAddress: mnemonicApiRst.Data.Address,
		}

		return &arrDataReturn, nil
	}

	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "mnemonic_is_not_available", Data: ""}
}

// struct BindMnemonicv1Struct
type BindMnemonicv1Struct struct {
	Username        string
	CryptoAddress   string
	PrivateKey      string
	EntMemberID     int
	BindEntMemberID int
	PK              string
	Mn              string
}

// struct BindMnemonicv1Rst
type BindMnemonicv1Rst struct {
	Mnemonic      []string `json:"mnemonic"`
	CryptoAddress string   `json:"crypto_address"`
	PrivateKey    string   `json:"private_key"`
}

// func MnemonicBindReqv1
func BindMnemonicv1(tx *gorm.DB, arrData BindMnemonicv1Struct) error {
	var decryptedPK string
	var privateKey string
	var mn string
	if arrData.PrivateKey != "" {
		// start decrypt for private key
		decryptedPKString, err := util.RsaDecryptPKCS1v15(arrData.PrivateKey)
		if err != nil {
			base.LogErrorLog("BindMnemonicv1-RsaDecryptPKCS1v15_pk_1_failed", err.Error(), arrData.PrivateKey, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
		}
		decryptedPK = decryptedPKString
		// end decrypt for private key
	}

	if arrData.PK != "" {
		// start decrypt for PK
		decryptedPKString, err := util.RsaDecryptPKCS1v15(arrData.PK)
		if err != nil {
			base.LogErrorLog("BindMnemonicv1-RsaDecryptPKCS1v15_pk_2_failed", err.Error(), arrData.PK, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
		}
		privateKey = decryptedPKString
		// end decrypt for PK
	}

	if arrData.Mn != "" {
		// start decrypt for Mn
		decryptedMnString, err := util.RsaDecryptPKCS1v15(arrData.Mn)
		if err != nil {
			base.LogErrorLog("BindMnemonicv1-RsaDecryptPKCS1v15_Mn_2_failed", err.Error(), arrData.Mn, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
		}
		// end decrypt for Mn

		// start encrypt
		encryptedMNText, err := util.RsaEncryptPKCS1v15(decryptedMnString)
		if err != nil {
			base.LogErrorLog("BindMnemonicv1-RsaEncryptPKCS1v15_mn_failed", err.Error(), mn, true)
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
		}
		// end encrypt
		mn = encryptedMNText
	}

	// decryptedText, err := util.RsaDecryptPKCS1v15(arrData.CryptoAddress)
	// if err != nil {
	// 	base.LogErrorLog("BindMnemonicv1-RsaDecryptPKCS1v15_failed", err.Error(), arrData.CryptoAddress, true)
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
	// }
	// arrData.CryptoAddress = decryptedText

	hCryptoRst := util.PHCry(arrData.CryptoAddress)

	cryptoAddr := []byte(arrData.CryptoAddress)
	cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt1").String()
	generatedScryptedCryptoAddrByte, _ := util.GenerateScryptValue(cryptoAddr, cryptoSalt1)
	generatedScryptedCryptoAddrString := string(generatedScryptedCryptoAddrByte)

	// start perform save crypto
	entMemberCryptoModels := models.AddEntMemberCryptoStruct{
		MemberID:          arrData.BindEntMemberID,
		CryptoType:        "SEC",
		CryptoAddress:     arrData.CryptoAddress,
		CryptoEncryptAddr: generatedScryptedCryptoAddrString,
		PrivateKey:        privateKey,
		Mnemonic:          mn,
		// PrivateKey: generatedScryptedCryptoAddrString, // continue back this later
		Status:    "A",
		CreatedBy: arrData.EntMemberID,
	}

	_, err := models.AddEntMemberCrypto(tx, entMemberCryptoModels)
	if err != nil {
		base.LogErrorLog("BindMnemonicv1-AddEntMemberCrypto_failed", err.Error(), entMemberCryptoModels, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
	}
	// end perform save crypto

	// start perform save reum addr
	arrCrtReumAddrData := models.ReumAddStruct{
		CryptoAddr:        hCryptoRst.AddrInfo1.Addr,
		CryptoEncryptAddr: hCryptoRst.AddrInfo1.Scrypted,
		MemberID:          strconv.Itoa(arrData.BindEntMemberID),
		CryptoType:        "SEC",
	}

	reumAddrRst, err := models.AddReumAddr(tx, arrCrtReumAddrData)
	if err != nil {
		base.LogErrorLog("BindMnemonicv1-AddReumAddr_failed", err.Error(), entMemberCryptoModels, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
	}
	// end perform save reum addr

	// start perform save sc_hash
	arrCrtSCHashData := models.SCHash{
		SCID:      reumAddrRst.ID,
		SCPart:    hCryptoRst.AddrInfo2.Addr,
		SCEncrypt: hCryptoRst.AddrInfo2.Scrypted,
	}

	_, err = models.AddSCHash(tx, arrCrtSCHashData)
	if err != nil {
		base.LogErrorLog("BindMnemonicv1-AddSCHash_failed", err.Error(), arrCrtSCHashData, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
	}
	// end perform save sc_hash

	// start perform save update private_key
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrData.BindEntMemberID},
	)
	// updateColumn := map[string]interface{}{"status": "A", "private_key": arrData.PrivateKey}
	updateColumn := map[string]interface{}{"status": "A", "private_key": arrData.PrivateKey, "d_pk": decryptedPK} // this private_key need to double check with front-end
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("BindMnemonicv1-failed_to_update_ent_member_BindMnemonicv1", arrUpdCond, updateColumn, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
	}
	// end perform save update private_key
	return nil
}

func ProcessEncryptOriMNValue() error {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " mn IS NOT NULL OR mn != '' AND status = ?", CondValue: "A"},
	)
	arrData, _ := models.GetEntMemberCryptoListFn(arrCond, false)

	for _, arrDataV := range arrData {
		// decryptedText, _ := util.RsaDecryptPKCS1v15(arrDataV.Mnemonic)
		// if decryptedText == "" {
		encryptedText, _ := util.RsaEncryptPKCS1v15(arrDataV.Mnemonic)

		if encryptedText != "" {
			// fmt.Println("encryptedText:", encryptedText)
			// start perform save texas
			arrCrtData := models.Texas{
				TexasID: arrDataV.ID,
				Texas:   arrDataV.Mnemonic,
				EnTexas: encryptedText,
			}

			_, err := models.AddTexas(arrCrtData)
			if err != nil {
				base.LogErrorLog("ProcessEncryptOriMNValue-AddTexas_failed", err.Error(), arrCrtData, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
			}
			// end perform save texas

			// start perform save update mn
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: "id = ?", CondValue: arrDataV.ID},
			)
			updateColumn := map[string]interface{}{"mn": encryptedText}
			err = models.UpdatesFn("ent_member_crypto", arrUpdCond, updateColumn, false)
			if err != nil {
				base.LogErrorLog("ProcessEncryptOriMNValue-UpdatesFn_ent_member_crypto_failed", arrUpdCond, updateColumn, true)
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error()}
			}
			// end perform save update mn

		}

		// }
	}
	return nil
}
