package member_service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
)

// Crypto struct
type Crypto struct {
	MemberID      int
	CryptoType    string
	CryptoAddress string
}

// Add crypto
func (c *Crypto) Add(tx *gorm.DB) string {
	// var ok bool
	var err error

	// verify crypto type
	if helpers.StringInSlice(c.CryptoType, []string{"USDT"}) == false {
		return "invalid_crypto_address_type"
	}

	// verify if address duplicated
	arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
		models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: c.CryptoType},
		models.WhereCondFn{Condition: "ent_member_crypto.crypto_address = ?", CondValue: c.CryptoAddress},
		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)
	arrEntMemberCrypto, err := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)

	if err != nil {
		models.ErrorLog("memberService|crypto:Add()", "GetEntMemberCryptoFn():1", err.Error())
		return "something_went_wrong"
	}

	if arrEntMemberCrypto != nil {
		return "crypto_address_already_taken"
	}

	// add crypto address
	entMemberCryptoModels := models.AddEntMemberCryptoStruct{
		MemberID:          c.MemberID,
		CryptoType:        c.CryptoType,
		CryptoAddress:     c.CryptoAddress,
		CryptoEncryptAddr: base.SHA256(strings.ToLower(c.CryptoAddress + "" + c.CryptoType + "WOD")),
		Status:            "A",
		CreatedBy:         c.MemberID,
	}

	entMemberCrypto, err := models.AddEntMemberCrypto(tx, entMemberCryptoModels)
	if err != nil {
		models.ErrorLog("memberService|crypto:Add()", "AddEntMemberCrypto():1", err.Error())
		return "something_went_wrong"
	}

	// deactivate old record status
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id != ?", CondValue: entMemberCrypto.ID},
		models.WhereCondFn{Condition: "member_id = ?", CondValue: c.MemberID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)
	updateColumn := map[string]interface{}{"status": "I"}
	err = models.UpdatesFnTx(tx, "ent_member_crypto", arrUpdCond, updateColumn, false)
	if err != nil {
		models.ErrorLog("memberService|crypto:Add()", "UpdatesFnTx():1", err.Error())
		return "something_went_wrong"
	}

	return ""
}

// func ProcessHCrypto
func ProcessHCrypto(crypto string) error {

	fmt.Println("crypto:", crypto)
	cryptoAddr, err := models.GetCustomMemberCryptoAddr(2, "SEC", true, false)
	fmt.Println("err:", err)
	fmt.Println("cryptoAddr:", cryptoAddr)
	// hCryptoRst := util.PHCry(crypto)
	// add1Byte := []byte(crypto)
	// cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt2").String()
	// hCryptoRst.AddrInfo1.Scrypted // this value will store in db as hashed value
	// err := models.CompareHashAndScryptedValue("e56763e7a5f51cc137adeec36ec0ca37b66fe5262da89d0f0221603063197b4e", add1Byte, cryptoSalt1)
	// if err != nil {
	// 	fmt.Println("error")
	// 	// base.LogErrorLog("ProcessHCrypto_failed_in_CompareHashAndScryptedValue", err.Error(), hCryptoRst, true)
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_bind"}
	// }
	// fmt.Println("err", err)

	// _ = CompareHashAndScryptedValue(generatedScryptedAdd1String, add1Byte)
	return nil
}

// func ProcessGetMemAddress. it can use either for ewt_setup.control = INTERNAL / BLOCKCHAIN. AS long as ewt_setup.status = 'A', ewt_setup.member_show = 1,
func ProcessGetMemAddress(tx *gorm.DB, entMemberID int, cryptoType string) (cryptoAddr string, err error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: cryptoType},
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		// models.WhereCondFn{Condition: "ewt_setup.member_show = ?", CondValue: 1},
	)
	wallet, _ := models.GetEwtSetupFn(arrCond, "", false)
	if wallet != nil {
		if wallet.CryptoAddr == 1 {
			// need to call redeem address api if the member does not have related crypto yet
			var callRedeemAddrApiStatus bool
			// start get member crypto
			cryptoAddr, err := models.GetCustomMemberCryptoAddr(entMemberID, wallet.EwtTypeCode, true, false)

			if err != nil {
				if err.Error() == "member_no_related_crypto_address" {
					callRedeemAddrApiStatus = true
				} else {
					arrErrData := map[string]interface{}{
						"entMemberID": entMemberID,
						"cryptoType":  wallet.EwtTypeCode,
					}
					base.LogErrorLog("ProcessGetMemAddress_GetCustomMemberCryptoAddr_failed", err, arrErrData, true)
					return "", err
				}
			}

			if cryptoAddr != "" && err == nil {
				return cryptoAddr, nil
			}
			if callRedeemAddrApiStatus {
				if strings.ToLower(cryptoType) == "usdt_erc20" || strings.ToLower(cryptoType) == "eth" || strings.ToLower(cryptoType) == "usdc_erc20" {
					cryptoType = "ETH"
				}
				if strings.ToLower(cryptoType) == "usdc" {
					cryptoType = "usdt"
				}
				if strings.ToLower(cryptoType) == "bep" {
					cryptoType = "BUSD"
				}
				cryptoType = strings.ToUpper(cryptoType)

				blockChainCryptoCode := wallet.BlockchainCryptoTypeCode
				// if strings.ToLower(blockChainCryptoCode) == "bsc" { // hard code if blockchain crypto code is bsc, use the same eth address. Requested from sam boss. Reason: X want too many address
				// 	blockChainCryptoCode = "ETH"
				// }
				// get nick_name
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: entMemberID},
					models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
				)
				member, _ := models.GetEntMemberFn(arrCond, "", false)
				if member == nil {
					base.LogErrorLog("ProcessGetMemAddress():unable_to_find_member_info", arrCond, nil, true)
					return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
				}
				cryptoAddrApiRst, err := GetCryptoAddrViaApi(member.NickName, blockChainCryptoCode)

				if err != nil {
					return "", err
				}

				if cryptoAddrApiRst.CryptoAddr == "" {
					return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
				}

				// start double check this crypto is not exist b4 in system for diff member
				arrCond = make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "ent_member_crypto.crypto_address = ?", CondValue: cryptoAddrApiRst.CryptoAddr},
					models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: cryptoType},
					models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
				)
				checkOthersCryptoAddr, err := models.GetEntMemberCryptoFn(arrCond, false)

				if checkOthersCryptoAddr != nil {
					if checkOthersCryptoAddr.MemberID != entMemberID { // address get from blockchain exist in system but diff member. need to return error. bcz address obtain from api is belongs to others member
						base.LogErrorLog("ProcessGetMemAddress_same_crypto_address_is_obtained", "crypto_address_obtained_is_belongs_to_member_id:"+strconv.Itoa(checkOthersCryptoAddr.MemberID), cryptoAddrApiRst.CryptoAddr, true)
						return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
					} else { // something is wrong with current hash function. plz check it ASAP
						// base.LogErrorLog("ProcessGetMemAddress_same_crypto_address_is_obtained", "crypto_address_obtained_is_retrieved_before_for_the_same_member_id:"+strconv.Itoa(checkOthersCryptoAddr.MemberID), cryptoAddrApiRst.CryptoAddr, true)
						// return checkOthersCryptoAddr.CryptoAddress, nil
						//usdc & usdt using same
						return cryptoAddrApiRst.CryptoAddr, nil
					}
				}
				// end double check this crypto is not exist b4 in system for diff member

				cryptoAddr := cryptoAddrApiRst.CryptoAddr
				hCryptoRst := util.PHCry(cryptoAddr)

				cryptoAddrByte := []byte(cryptoAddr)
				cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt1").String()
				generatedScryptedCryptoAddrByte, _ := util.GenerateScryptValue(cryptoAddrByte, cryptoSalt1)
				generatedScryptedCryptoAddrString := string(generatedScryptedCryptoAddrByte)

				// start perform save crypto
				entMemberCryptoModels := models.AddEntMemberCryptoStruct{
					MemberID:          entMemberID,
					CryptoType:        cryptoType,
					CryptoAddress:     cryptoAddr,
					CryptoEncryptAddr: generatedScryptedCryptoAddrString,
					// PrivateKey: generatedScryptedCryptoAddrString, // continue back this later
					Status:    "A",
					CreatedBy: entMemberID,
				}

				_, err = models.AddEntMemberCrypto(tx, entMemberCryptoModels)
				if err != nil {
					base.LogErrorLog("ProcessGetMemAddress_AddEntMemberCrypto_failed", err.Error(), entMemberCryptoModels, true)
					return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_save_ent_member_crypto"}
				}
				// end perform save crypto

				// start perform save reum addr
				arrCrtReumAddrData := models.ReumAddStruct{
					CryptoAddr:        hCryptoRst.AddrInfo1.Addr,
					CryptoEncryptAddr: hCryptoRst.AddrInfo1.Scrypted,
					MemberID:          strconv.Itoa(entMemberID),
					CryptoType:        cryptoType,
				}

				reumAddrRst, err := models.AddReumAddr(tx, arrCrtReumAddrData)
				if err != nil {
					base.LogErrorLog("ProcessGetMemAddress_AddReumAddr_failed", err.Error(), entMemberCryptoModels, true)
					return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_to_save_AddReumAddr_failed"}
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
					base.LogErrorLog("ProcessGetMemAddress_AddSCHash_failed", err.Error(), arrCrtSCHashData, true)
					return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "AddSCHash_failed"}
				}
				// end perform save sc_hash

				// start perform add crpyto address into monitoring list
				AddCryptoAddrToMonitor(wallet.BlockchainCryptoTypeCode, cryptoAddr)

				// end perform add crpyto address into monitoring list
				return cryptoAddr, nil
			}
			// end get member crypto
		} else { // no crypto address is needed

		}
	}

	return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_related_address"}
}

type CryptoAddrApiRst struct {
	CryptoAddr string
}

func GetCryptoAddrViaApi(username string, cryptoType string) (*CryptoAddrApiRst, error) {
	type apiRstStruct struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
		Data       string `json:"data"`
	}

	cryptoType = strings.ToUpper(cryptoType)
	settingID := "get_crypto_addr_api"
	arrApiSetting, _ := models.GetSysGeneralSetupByID(settingID)
	url := arrApiSetting.SettingValue2
	requestMethod := "POST"
	header := map[string]string{
		"X-Authorization": arrApiSetting.SettingValue1,
		"Content-Type":    "application/json",
	}
	arrPlatformConfig := map[string]interface{}{
		arrApiSetting.InputValue3: arrApiSetting.SettingValue3 + "_" + cryptoType,
		arrApiSetting.InputValue4: arrApiSetting.SettingValue4,
	}
	arrPlatformConfigEncoded, _ := json.Marshal(arrPlatformConfig)
	arrPlatformConfigEncodedString := string(arrPlatformConfigEncoded)
	usernameEncoded, _ := json.Marshal(map[string]interface{}{"redeem_by": username})
	usernameEncodedString := string(usernameEncoded)
	arrPostData := map[string]interface{}{
		"crypto_code":     cryptoType,
		"platform_config": arrPlatformConfigEncodedString,
		"params":          usernameEncodedString,
	}

	res, apiErr := base.RequestAPI(requestMethod, url, header, arrPostData, nil)

	if apiErr != nil {
		base.LogErrorLogV2("ProcessGetMemAddress_error_in_api_call_before_call", apiErr.Error(), nil, true, "blockchain")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
	}

	if res.StatusCode != 200 {
		base.LogErrorLogV2("ProcessGetMemAddress_error_in_api_call_after_call", res.Body, nil, true, "blockchain")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
	}

	var apiRst apiRstStruct
	err := json.Unmarshal([]byte(res.Body), &apiRst)

	if err != nil {
		base.LogErrorLog("ProcessGetMemAddress_error_in_json_decode_api_result", err.Error(), res.Body, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
	}

	arrDataReturn := CryptoAddrApiRst{
		CryptoAddr: apiRst.Data,
	}

	return &arrDataReturn, nil

}

// AddCryptoAddrToMonitor func
func AddCryptoAddrToMonitor(cryptoType, cryptoAddr string) error {

	if strings.ToLower(cryptoType) == "usdt" || strings.ToLower(cryptoType) == "usdc" {
		cryptoType = "TRX"
	}
	if strings.ToLower(cryptoType) == "usdt_erc20" || strings.ToLower(cryptoType) == "usdc_erc20" {
		cryptoType = "ETH"
	}
	if strings.ToLower(cryptoType) == "bep" {
		cryptoType = "BUSD"
	}
	cryptoType = strings.ToUpper(cryptoType)
	settingID := "add_monitor_crypto_addr_api"
	arrApiSetting, _ := models.GetSysGeneralSetupByID(settingID)

	if arrApiSetting == nil {
		base.LogErrorLog("AddCryptoAddrToMonitor_GetSysGeneralSetupByID_failed", "add_monitor_crypto_addr_api_setting_is_missing", settingID, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	url := arrApiSetting.SettingValue2
	configCodePart1 := arrApiSetting.SettingValue3
	configCode := strings.ToUpper(configCodePart1 + "_" + cryptoType)
	requestMethod := arrApiSetting.SettingValue1
	header := map[string]string{
		// "X-Authorization": arrApiSetting.SettingValue1,
		"Content-Type": "application/json",
	}
	arrPlatformConfig := map[string]interface{}{
		arrApiSetting.InputValue3: configCode,
		arrApiSetting.InputValue4: arrApiSetting.SettingValue4,
	}
	arrParams := make([]map[string]interface{}, 0)
	arrParams = append(arrParams,
		map[string]interface{}{"account": cryptoAddr, "status": "A"},
	)
	arrPlatformConfigEncoded, _ := json.Marshal(arrPlatformConfig)
	arrPlatformConfigEncodedString := string(arrPlatformConfigEncoded)
	entMemberIDEncoded, _ := json.Marshal(arrParams)
	entMemberIDEncodedString := string(entMemberIDEncoded)
	arrPostData := map[string]interface{}{
		"crypto_code":     cryptoType,
		"platform_config": arrPlatformConfigEncodedString,
		"params":          entMemberIDEncodedString,
	}

	res, apiErr := base.RequestAPI(requestMethod, url, header, arrPostData, nil)

	if apiErr != nil {
		base.LogErrorLogV2("AddCryptoAddrToMonitor_error_in_api_call_before_call", apiErr.Error(), nil, true, "blockchain")
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
	}

	if res.StatusCode != 200 {
		base.LogErrorLogV2("AddCryptoAddrToMonitor_error_in_api_call_after_call", res.Body, nil, true, "blockchain")
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
	}
	type apiRstStruct struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		Msg        string `json:"msg"`
		Data       []struct {
			Account string `json:"account"`
			Status  string `json:"status"`
		} `json:"data"`
	}

	var apiRst apiRstStruct
	err := json.Unmarshal([]byte(res.Body), &apiRst)

	if err != nil {
		base.LogErrorLog("AddCryptoAddrToMonitor_error_in_json_decode_api_result", err.Error(), res.Body, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
	}

	if apiRst.StatusCode != 200 {
		base.LogErrorLog("AddCryptoAddrToMonitor_error_in_api_call_after_call", res.Body, apiRst, true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong", Data: ""}
	}

	return nil
}

func ProcessCryptoAddressChecking() {
	arrCrypto := []string{"ETH", "USDT", "BUSD", "USDT_ERC20", "BNB"}
	timeStart := base.GetCurrentTime("2006-01-02 15:04:05")
	base.LogErrorLog("ProcessCryptoAddressChecking-start: "+timeStart, "", nil, true)
	// fmt.Println("time start: ", timeStart)
	for _, arrCryptoV := range arrCrypto {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: arrCryptoV},
		)
		activeMemCrypto, _ := models.GetEntMemberCryptoListFn(arrCond, false)
		// for _, activeMemCryptoV := range activeMemCrypto {
		for activeMemCryptoK, activeMemCryptoV := range activeMemCrypto {
			cryptoAddr, err := models.GetCustomMemberCryptoAddr(activeMemCryptoV.MemberID, activeMemCryptoV.CryptoType, true, false)
			if err != nil {
				arrErr := map[string]interface{}{
					"MemberID":   activeMemCryptoV.MemberID,
					"CryptoType": activeMemCryptoV.CryptoType,
				}
				msg := "ProcessCryptoAddressChecking-GetCustomMemberCryptoAddr_1: [" + strconv.Itoa(activeMemCryptoV.MemberID) + "-" + activeMemCryptoV.CryptoType + "]"
				base.LogErrorLog(msg, err.Error(), arrErr, true)
				os.Exit(0)
			} else if cryptoAddr == "" {
				arrErr := map[string]interface{}{
					"MemberID":   activeMemCryptoV.MemberID,
					"CryptoType": activeMemCryptoV.CryptoType,
				}
				msg1 := "ProcessCryptoAddressChecking-GetCustomMemberCryptoAddr_2: [" + strconv.Itoa(activeMemCryptoV.MemberID) + "-" + activeMemCryptoV.CryptoType + "]"
				base.LogErrorLog(msg1, "cryptoAddr_is_empty", arrErr, true)
				os.Exit(0)
			} else {
				msg := strconv.Itoa(activeMemCryptoK) + ".success - " + strconv.Itoa(activeMemCryptoV.MemberID) + " - " + activeMemCryptoV.CryptoType
				fmt.Println(msg)
			}
		}
	}
	timeEnd := base.GetCurrentTime("2006-01-02 15:04:05")
	fmt.Println("time end: ", timeEnd)
	base.LogErrorLog("ProcessCryptoAddressChecking-end: "+timeEnd, "", nil, true)
}

func ProcessPKChecking() {
	arrCrypto := []string{"SEC"}
	timeStart := base.GetCurrentTime("2006-01-02 15:04:05")
	base.LogErrorLog("ProcessPKChecking-start: "+timeStart, "", nil, true)
	fmt.Println("time start: ", timeStart)
	for _, arrCryptoV := range arrCrypto {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: arrCryptoV},
			models.WhereCondFn{Condition: "ent_member_crypto.member_id > ?", CondValue: 0},
			// models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: 11973}, // debug
		)
		activeMemCrypto, _ := models.GetEntMemberCryptoListFn(arrCond, false)
		for _, activeMemCryptoV := range activeMemCrypto {
			// for activeMemCryptoK, activeMemCryptoV := range activeMemCrypto {
			arrEntMemCond := make([]models.WhereCondFn, 0)
			arrEntMemCond = append(arrEntMemCond,
				models.WhereCondFn{Condition: " ent_member.id = ?", CondValue: activeMemCryptoV.MemberID},
			)
			arrEntMem, err := models.GetEntMemberFn(arrEntMemCond, "", false)

			if err != nil {
				arrErr := map[string]interface{}{
					"MemberID":   activeMemCryptoV.MemberID,
					"CryptoType": activeMemCryptoV.CryptoType,
				}
				msg := "ProcessPKChecking-GetEntMemberFn: [" + strconv.Itoa(activeMemCryptoV.MemberID) + "-" + activeMemCryptoV.CryptoType + "]"
				base.LogErrorLog(msg, err.Error(), arrErr, true)
				os.Exit(0)
			} else if arrEntMem == nil {
				// arrErr := map[string]interface{}{
				// 	"MemberID":   activeMemCryptoV.MemberID,
				// 	"CryptoType": activeMemCryptoV.CryptoType,
				// }
				// msg1 := "ProcessPKChecking-GetEntMemberFn: [" + strconv.Itoa(activeMemCryptoV.MemberID) + "-" + activeMemCryptoV.CryptoType + "]"
				// base.LogErrorLogV2(msg1, "ent_member_is_empty", arrErr, true, "koobot")
				// os.Exit(0)
			} else {
				if arrEntMem.DPK != "" && activeMemCryptoV.PrivateKey != "" {
					pkSalt := setting.Cfg.Section("custom").Key("PKSalt").String()
					key := activeMemCryptoV.PrivateKey + pkSalt
					keyByte := []byte(key)
					hasher := sha256.New()
					hasher.Write(keyByte)
					sha256Value := hex.EncodeToString(hasher.Sum(nil))

					if arrEntMem.DPK != sha256Value {
						arrErr := map[string]interface{}{
							"member_id":   activeMemCryptoV.MemberID,
							"CryptoType":  activeMemCryptoV.CryptoType,
							"d_pk":        arrEntMem.DPK,
							"sha256Value": sha256Value,
						}
						msg1 := "ProcessPKChecking-invalid_private_key_is_detected: [" + strconv.Itoa(activeMemCryptoV.MemberID) + "-" + activeMemCryptoV.CryptoType + "]"
						base.LogErrorLog(msg1, "pk_does_not_match", arrErr, true)
						os.Exit(0)
					} else if arrEntMem.DPK == sha256Value {
						msg1 := "ProcessPKChecking-success: [" + strconv.Itoa(activeMemCryptoV.MemberID) + "-" + activeMemCryptoV.CryptoType + "]"
						fmt.Println("success:", msg1)
					} else {
						fmt.Println("no action:")
					}
				}
			}
		}
	}
	timeEnd := base.GetCurrentTime("2006-01-02 15:04:05")
	fmt.Println("time end: ", timeEnd)
	base.LogErrorLog("ProcessPKChecking-end: "+timeEnd, "", nil, true)
}
