package sales_service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

type ProcessSaveTransv1Struct struct {
	EntMemberID int
	CryptoType  string
	DocNo       string
	TransType   string
	// UnitPrice   string
	TotalIn    float64
	TotalOut   float64
	SigningKey string
	Remark     string
}

func ProcessSaveTransOutv1(tx *gorm.DB, arrData ProcessSaveTransv1Struct) (map[string]interface{}, error) {
	ewtSetupCond := make([]models.WhereCondFn, 0)
	ewtSetupCond = append(ewtSetupCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.CryptoType},
	)
	ewtSetup, _ := models.GetEwtSetupFn(ewtSetupCond, "", false)

	if ewtSetup == nil {
		base.LogErrorLogV2("ProcessSaveTransv1-failed_to_get_ewt_setup", ewtSetupCond, arrData, true, "sys_error_log")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	// docType := "LALIGASTK"
	// docNo, err := models.GetRunningDocNo(docType, tx) //get withdraw doc no

	// if err != nil {
	// 	base.LogErrorLog("ProcessMemberSellTradingRequestv2-get_laliga_staking_doc_no_failed", err.Error(), nil, true)
	// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	// }
	var hashValue string
	if strings.ToLower(ewtSetup.Control) == "blockchain" {
		if arrData.TotalOut > 0 {
			if arrData.SigningKey == "" {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "signing_key_is_required"}
			}

			cryptoAddr, err := member_service.ProcessGetMemAddress(tx, arrData.EntMemberID, arrData.CryptoType)
			if err != nil {
				base.LogErrorLogV2("ProcessSaveTransv1-ProcessGetMemAddress_failed", err.Error(), arrData, true, "sys_error_log")
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}

			BlkCWalBal, _ := wallet_service.GetMemberBlockchainWalletBalance(arrData.EntMemberID, arrData.CryptoType, cryptoAddr)
			availableBalance := BlkCWalBal.AvailableBalance

			if arrData.TotalOut > availableBalance {
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "balance_not_enu_to_deduct"}
			}
		}

		// start send signing key to blockchain site
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:     arrData.EntMemberID,
			EwalletTypeID:   ewtSetup.ID,
			DocNo:           arrData.DocNo,
			Status:          "P",
			TransactionType: arrData.TransType,
			TransactionData: arrData.SigningKey,
			TotalOut:        arrData.TotalOut,
			LogOnly:         0,
			Remark:          arrData.Remark,
		}
		fmt.Println(arrSaveMemberBlockchainWallet.TotalOut)
		// errMsg, _ := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"} // debug
		}
		hashValue = saveMemberBlochchainWalletRst["hashValue"]
		// end send signing key to blockchain site

		// arrCrtData := models.AddLaligaStakeStruct{
		// 	MemberID:      arrData.EntMemberID,
		// 	DocNo:         arrData.DocNo,
		// 	CryptoCode:    arrData.CryptoType,
		// 	TotalUnit:     quantity,
		// 	UnitPrice:     unitPrice,
		// 	TotalAmount:   amount,
		// 	BalanceAmount: amount,
		// 	BalanceUnit:   quantity,
		// 	Status:        "P",
		// 	SigningKey:    arrData.SigningKey,
		// 	TransHash:     hashValue,
		// 	CreatedBy:     strconv.Itoa(arrData.EntMemberID),
		// }

		// _, err = models.AddLaligaStake(tx, arrCrtData)
		// if err != nil {
		// 	base.LogErrorLog("ProcessSaveTransv1-failed_to_save_laliga_stake", err.Error(), arrCrtData, true)
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// }

		// err = models.UpdateRunningDocNo(docType, tx) //update doc no
		// if err != nil {
		// 	base.LogErrorLog("ProcessSaveTransv1-failed_in_UpdateRunningDocNo", err.Error(), docType, true)
		// 	return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// }
	} else {
		// so far no this step yet
	}

	arrDataReturn := map[string]interface{}{
		"hash":       hashValue,
		"bizId":      arrData.DocNo,
		"trans_type": arrData.TransType,
	}
	return arrDataReturn, nil
}

func ProcessSaveTransInv1(tx *gorm.DB, arrData ProcessSaveTransv1Struct) (map[string]interface{}, error) {
	ewtSetupCond := make([]models.WhereCondFn, 0)
	ewtSetupCond = append(ewtSetupCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: arrData.CryptoType},
	)
	ewtSetup, _ := models.GetEwtSetupFn(ewtSetupCond, "", false)

	if ewtSetup == nil {
		base.LogErrorLogV2("ProcessSaveTransInv1-failed_to_get_ewt_setup", ewtSetupCond, arrData, true, "sys_error_log")
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	var hashValue string
	if strings.ToLower(ewtSetup.Control) == "blockchain" {

		cryptoAddr, err := member_service.ProcessGetMemAddress(tx, arrData.EntMemberID, arrData.CryptoType)
		if err != nil {
			base.LogErrorLogV2("ProcessSaveTransInv1-ProcessGetMemAddress_failed", err.Error(), arrData, true, "sys_error_log")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		contractAddress := ewtSetup.ContractAddress

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: 0},
			models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "ent_member_crypto.crypto_type = ?", CondValue: "SEC"},
		)
		arrCompanyCrypto, _ := models.GetEntMemberCryptoFn(arrCond, false)
		if arrCompanyCrypto == nil {
			base.LogErrorLogV2("ProcessSaveTransInv1-failed_to_get_company_info", ewtSetupCond, arrData, true, "sys_error_log")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		companyAddress := arrCompanyCrypto.CryptoAddress
		companyPrivateKey := arrCompanyCrypto.PrivateKey

		// start check company balance
		balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(arrData.CryptoType, companyAddress)
		if err != nil {
			arrErr := map[string]interface{}{
				"CryptoType":     arrData.CryptoType,
				"companyAddress": companyAddress,
			}
			base.LogErrorLogV2("ProcessSaveTransInv1-failed_GetBlockchainWalletBalanceApiV1", err.Error(), arrErr, true, "sys_error_log")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		if balance < arrData.TotalIn {
			arrErr := map[string]interface{}{
				"cryptoType":            arrData.CryptoType,
				"companyAddressBalance": balance,
				"TotalIn":               arrData.TotalIn,
			}
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			base.LogErrorLogV2("ProcessSaveTransInv1-company_wallet_balance_is_not_enough", errMsg, arrErr, true, "sys_error_log")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}
		// end check company balance

		// get member crypto address
		arrExchangeDebitSetting, errMsg := wallet_service.GetSigningKeySettingByModule(ewtSetup.EwtTypeCode, companyAddress, "UNSTAKELALIGA")
		if errMsg != "" {
			base.LogErrorLogV2("ProcessSaveTransInv1-failed_to_get_company_info", errMsg, arrData, true, "sys_error_log")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		chainID, _ := helpers.ValueToInt(arrExchangeDebitSetting["chain_id"].(string))
		maxGas, _ := helpers.ValueToInt(arrExchangeDebitSetting["max_gas"].(string))
		// send exchange-debit signing key
		arrProcecssGenerateSignTransaction := wallet_service.ProcecssGenerateSignTransactionStruct{
			TokenType:       ewtSetup.EwtTypeCode,
			PrivateKey:      companyPrivateKey,
			ContractAddress: contractAddress,
			ChainID:         int64(chainID),
			FromAddr:        companyAddress,
			ToAddr:          cryptoAddr, // this is refer to the member address
			Amount:          arrData.TotalIn,
			MaxGas:          uint64(maxGas),
		}

		signingKey, err := wallet_service.ProcecssGenerateSignTransaction(arrProcecssGenerateSignTransaction)
		if err != nil {
			base.LogErrorLogV2("ProcessSaveTransInv1-failed_to_ProcecssGenerateSignTransaction", err.Error(), arrProcecssGenerateSignTransaction, true, "sys_error_log")
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		// call sign transaction + insert blockchain_trans
		arrSaveMemberBlockchainWallet := wallet_service.SaveMemberBlochchainWalletStruct{
			EntMemberID:      arrData.EntMemberID,
			EwalletTypeID:    ewtSetup.ID,
			DocNo:            arrData.DocNo,
			Status:           "P",
			TransactionType:  arrData.TransType,
			TransactionData:  signingKey,
			TotalIn:          arrData.TotalIn,
			ConversionRate:   1,
			ConvertedTotalIn: arrData.TotalIn,
			LogOnly:          0,
			Remark:           arrData.Remark,
		}

		// errMsg, _ = wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		errMsg, saveMemberBlochchainWalletRst := wallet_service.SaveMemberBlochchainWallet(arrSaveMemberBlockchainWallet)
		if errMsg != "" {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"} // debug
		}
		hashValue = saveMemberBlochchainWalletRst["hashValue"]
	} else {
		// so far no this step yet
	}

	arrDataReturn := map[string]interface{}{
		"hash":       hashValue,
		"bizId":      arrData.DocNo,
		"trans_type": arrData.TransType,
	}
	return arrDataReturn, nil
}

func ProcessLaligaCallBack(manual bool) {
	settingID := "process_laliga_callback_setting"
	arrApiSetting, err := models.GetSysGeneralSetupByID(settingID)

	if err != nil || arrApiSetting == nil {
		fmt.Println("no process_process_laliga_callback_setting setting")
		return
	}

	if arrApiSetting.InputType1 != "1" && !manual {
		fmt.Println("process_laliga_callback_setting is off")
		return
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "status = ?", CondValue: "P"},
	)
	arrPendingLaligaProcessQ, _ := models.GetLaligaProcessQueueFn(arrCond, false)
	// base.LogErrorLog("ProcessLaligaCallBack-1", len(arrPendingLaligaProcessQ), arrCond, false)
	if len(arrPendingLaligaProcessQ) > 0 {
		header := map[string]string{
			"Content-Type": "application/json",
		}
		type apiRstStruct struct {
			Rst int    `json:"rst"`
			Msg string `json:"msg"`
		}
		secret := arrApiSetting.InputType2
		for _, arrPendingLaligaProcessQV := range arrPendingLaligaProcessQ {

			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " hash_value = ?", CondValue: arrPendingLaligaProcessQV.ProcessID},
			)
			arrTranx, _ := models.GetBlockchainTrans(arrCond, "", false)
			// base.LogErrorLog("ProcessLaligaCallBack-2", arrTranx, arrCond, false)

			if arrTranx != nil {
				transType := strings.ToUpper(arrTranx.TransactionType)
				signText := secret + arrTranx.DocNo + arrTranx.HashValue + transType + secret
				signedHash := md5.Sum([]byte(signText))
				signedHashHex := hex.EncodeToString(signedHash[:])
				arrDebug := map[string]interface{}{
					"secret":             secret,
					"arrTranx.DocNo":     arrTranx.DocNo,
					"arrTranx.HashValue": arrTranx.HashValue,
					"transType":          transType,
					"signText":           signText,
					"signedHashHex":      signedHashHex,
				}
				base.LogErrorLog("debug-signedHashHex-1", arrDebug, signedHashHex, false)

				data := map[string]interface{}{
					"bizId":      arrTranx.DocNo,
					"hash":       arrTranx.HashValue,
					"trans_type": transType,
					"sign":       signedHashHex,
				}
				method := arrApiSetting.SettingValue1
				url := arrApiSetting.InputValue1

				// base.LogErrorLog("ProcessLaligaCallBack-3", url, data, false)
				res, err_api := base.RequestAPI(method, url, header, data, nil)
				// base.LogErrorLog("ProcessLaligaCallBack-4", err_api, res, false)

				if err_api != nil {
					base.LogErrorLogV2("ProcessLaligaCallBack-error_in_api_call_before_call", err_api.Error(), nil, true, "sys_error_log")
					return
				}

				if res.StatusCode != 200 {
					base.LogErrorLogV2("ProcessLaligaCallBack-error_in_api_call_after_call", res.Body, nil, true, "sys_error_log")
					return
				}

				var apiRst apiRstStruct
				err := json.Unmarshal([]byte(res.Body), &apiRst)

				if err != nil {
					base.LogErrorLog("ProcessLaligaCallBack-error_in_json_decode_api_result", err.Error(), res.Body, true)
					return
				}

				if apiRst.Rst == 1 {
					arrUpdCond := make([]models.WhereCondFn, 0)
					arrUpdCond = append(arrUpdCond,
						models.WhereCondFn{Condition: " process_id = ?", CondValue: arrTranx.HashValue},
						models.WhereCondFn{Condition: " status = ?", CondValue: "P"},
					)
					arrUpdata := map[string]interface{}{"status": "AP"}
					models.UpdatesFn("laliga_process_queue", arrUpdCond, arrUpdata, false)
				}
			}
		}
	}
}
