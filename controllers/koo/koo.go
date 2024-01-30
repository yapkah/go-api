package koo

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/sales_service"
	"github.com/yapkah/go-api/service/trading_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

type KooTestFrom struct {
	Crypto                             string  `form:"crypto" json:"crypto"`
	CryptoType                         string  `form:"crypto_type" json:"crypto_type"`
	RSAText                            string  `form:"rsa_text" json:"rsa_text"`
	EncryptedKey                       string  `form:"encrypted_key" json:"encrypted_key"`
	CutOffDecimalAmount                float64 `form:"cut_off_decimal_amount" json:"cut_off_decimal_amount"`
	Decimal                            uint    `form:"decimal" json:"decimal"`
	RoundDownAmount                    float64 `form:"round_down_amount" json:"round_down_amount"`
	EntMemberIDCryptoAddress           int     `form:"ent_member_id_crypto_address" json:"ent_member_id_crypto_address"`
	CompanyAddress                     string  `form:"company_address" json:"company_address"`
	AddMonitorAddr                     string  `form:"add_monitor_addr" json:"add_monitor_addr"`
	AddAllMonitorAddr                  string  `form:"add_all_monitor_addr" json:"add_all_monitor_addr"`
	ScryptText                         string  `form:"scrypt_text" json:"scrypt_text"`
	TestAutoMatchTrading               string  `form:"test_auto_match_trading" json:"test_auto_match_trading"`
	GetBlockchainWalletBalanceApiV1    string  `form:"get_blockchain_wallet_balanc_apiv1" json:"get_blockchain_wallet_balanc_apiv1"`
	TotalIn                            float64 `form:"total_in" json:"total_in"`
	TotalOut                           float64 `form:"total_out" json:"total_out"`
	PNOs                               string  `form:"pn_os" json:"pn_os"`
	PNGroupName                        string  `form:"pn_group_name" json:"pn_group_name"`
	ProcessUpdateMissingMemberCode     string  `form:"process_update_missing_member_code" json:"process_update_missing_member_code"`
	SigningKey                         string  `form:"signing_key" json:"signing_key"`
	EncryptedIDUsername                string  `form:"encrypted_id_username" json:"encrypted_id_username"`
	ProcessCryptoAddressChecking       string  `form:"process_crypto_address_checking" json:"process_crypto_address_checking"`
	KGraph                             string  `form:"k_graph" json:"k_graph"`
	TestProcessSendPushNotificationMsg string  `form:"test_process_send_pn_msg" json:"test_process_send_pn_msg"`
	ProcessPKChecking                  string  `form:"process_pk_checking" json:"process_pk_checking"`
	ProcessLaligaCallBack              string  `form:"process_laliga_callback" json:"process_laliga_callback"`
	EncryptOriMNValue                  string  `form:"encrypt_ori_mnvalue" json:"encrypt_ori_mnvalue"`
	TestPDF                            string  `form:"test_pdf" json:"test_pdf"`
	TestCallEmailApi                   string  `form:"test_call_email_api" json:"test_call_email_api"`
	TestCommitOld                      string  `form:"test_commit_old" json:"test_commit_old"`
	TestCommitNew                      string  `form:"test_commit_new" json:"test_commit_new"`
	// TestValidation                  float64 `form:"test_validation" json:"test_validation" valid:"Required;Min(0);"`
}

// func KooTest
func KooTest(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form KooTestFrom
	)
	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	if form.Crypto != "" {
		member_service.ProcessHCrypto(form.Crypto)
	}

	if form.CryptoType != "" && form.EntMemberIDCryptoAddress >= 0 {
		tx := models.Begin()
		cryptoAddr, err := wallet_service.TestProcessGetMemAddress(tx, form.EntMemberIDCryptoAddress, form.CryptoType)
		if err != nil {
			tx.Rollback()
			fmt.Println("err:", err)
			return
		}
		tx.Commit()

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
			"EntMemberIDCryptoAddress": form.EntMemberIDCryptoAddress,
			"CryptoType":               form.CryptoType,
			"crypto_address":           cryptoAddr,
		})
		fmt.Println("success cryptoAddr:", cryptoAddr)
		return
	}

	// startTimeT := base.GetCurrentTimeT()
	// loc := base.GetTimeZone()
	// startTimeT, _ := time.Parse("2006-01-02 15:04:05", curDateTimeString)
	// startTimeT = startTimeT.In(loc)
	// fmt.Println("startTimeT:", startTimeT)

	if form.CompanyAddress != "" {
		fmt.Println("form.CompanyAddress:", form.CompanyAddress)
		arrAddInfo := strings.Split(form.CompanyAddress, ",")
		if len(arrAddInfo) != 2 {
			message := app.MsgStruct{
				Msg: "please_give_nick_name_and_crypto_address",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: arrAddInfo[1]},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		member, _ := models.GetEntMemberFn(arrCond, "", false)
		if member == nil {
			message := app.MsgStruct{
				Msg: "invalid_nick_name",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		arrData := member_service.BindMnemonicv1Struct{
			Username:      member.NickName,
			CryptoAddress: arrAddInfo[0],
			EntMemberID:   member.ID,
		}

		fmt.Println("arrData:", arrData)
		tx := models.Begin()
		err := member_service.BindMnemonicv1(tx, arrData)
		if err != nil {
			models.Rollback(tx)
			message := app.MsgStruct{
				Msg: "something_went_wrong",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		// commit transaction
		models.Commit(tx)
		fmt.Println("success")
	}

	if form.RSAText != "" {
		// 	b := []byte(form.RSAText)

		// fmt.Println("form.RSAText:", form.RSAText)
		// fmt.Println("=== start encrypting ===")
		encryptedText, err := util.RsaEncryptPKCS1v15(form.RSAText)
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
			"encrypted_text": encryptedText,
		})

		// fmt.Println("encryptedText:", encryptedText)
		// decryptedText, err := util.RsaDecryptPKCS1v15(encryptedText)
		// fmt.Println("decryptedText:", decryptedText)
		// fmt.Println("err:", err)
		// fmt.Println("=== start decrypting ===")
		// fmt.Println("err:", err)
		return
	}

	if form.EncryptedKey != "" {
		decryptedText, err := util.RsaDecryptPKCS1v15(form.EncryptedKey)
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
			"decrypted_text": decryptedText,
		})
		return
	}

	if form.CutOffDecimalAmount != 0 {
		y := helpers.CutOffDecimalv2(form.CutOffDecimalAmount, form.Decimal, ".", ",", true)
		s, err := strconv.ParseFloat(y, 64)
		if err != nil {
			fmt.Println(err.Error()) // 3.14159265
		}
		arrDataReturn := map[string]interface{}{
			"input":         form.CutOffDecimalAmount,
			"decimal_point": form.Decimal,
			"string_input":  y,
			"float_input":   s,
		}
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
		return
	}

	if form.AddMonitorAddr != "" {
		arrAddInfo := strings.Split(form.AddMonitorAddr, ",")
		if len(arrAddInfo) != 2 {
			message := app.MsgStruct{
				Msg: "please_give_crypto_type_and_crypto_address",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		err := member_service.AddCryptoAddrToMonitor(arrAddInfo[0], arrAddInfo[1])

		fmt.Println("err:", err)
	}

	if form.AddAllMonitorAddr != "" {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "crypto_type != ?", CondValue: "SEC"},
		)
		entMemberCrypto, _ := models.GetEntMemberCryptoListFn(arrCond, true)
		for _, entMemberCryptoV := range entMemberCrypto {
			err := member_service.AddCryptoAddrToMonitor(entMemberCryptoV.CryptoType, entMemberCryptoV.CryptoAddress)
			if err == nil {
				fmt.Println("success", entMemberCryptoV.CryptoType, entMemberCryptoV.CryptoAddress)
			} else {
				fmt.Println("err", err.Error(), entMemberCryptoV.CryptoType, entMemberCryptoV.CryptoAddress)
			}
		}
	}

	if form.ScryptText != "" {
		// hCryptoRst := util.PHCry(form.ScryptText)

		cryptoAddr := []byte(form.ScryptText)
		cryptoSalt1 := setting.Cfg.Section("custom").Key("CryptoSalt1").String()
		generatedScryptedCryptoAddrByte, _ := util.GenerateScryptValue(cryptoAddr, cryptoSalt1)
		generatedScryptedCryptoAddrString := string(generatedScryptedCryptoAddrByte)
		// fmt.Println("hCryptoRst:", hCryptoRst)

		addrByte := []byte(form.ScryptText)
		err := models.CompareHashAndScryptedValue(generatedScryptedCryptoAddrString, addrByte, cryptoSalt1)
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_scrypted_text"}, nil)
			return
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
			"scrypt_text":   form.ScryptText,
			"scrypted_text": generatedScryptedCryptoAddrString,
		})
	}

	if strings.ToLower(form.TestAutoMatchTrading) == "yes" {
		trading_service.ProcessAutoMatchTrading(true)

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{})
		return
	}

	if strings.ToLower(form.GetBlockchainWalletBalanceApiV1) != "" {
		arrAddInfo := strings.Split(form.GetBlockchainWalletBalanceApiV1, ",")
		if len(arrAddInfo) != 2 {
			message := app.MsgStruct{
				Msg: "please_give_crypto_and_crypto_address",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		crypto := arrAddInfo[0]
		address := arrAddInfo[1]
		balance, err := wallet_service.GetBlockchainWalletBalanceApiV1(crypto, address)
		if err != nil {
			message := app.MsgStruct{
				Msg: err.Error(),
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{"balance": balance})
		return
	}

	if form.PNOs != "" && form.PNGroupName != "" {
		arrCallCreateNewPushNotificationGroupApi := base.CreateNewPushNotificationGroupStruct{
			GroupName: form.PNGroupName,
			Os:        form.PNOs,
		}
		err := base.CallCreateNewPushNotificationGroupApi(arrCallCreateNewPushNotificationGroupApi)
		fmt.Println("err", err)
	}

	if form.ProcessUpdateMissingMemberCode == "yes" {
		member_service.ProcessUpdateMissingMemberCode()
	}

	if form.SigningKey != "" {
		result, err := util.DecodeSigningKey(form.SigningKey)

		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
			"SigningKey": form.SigningKey,
			"result":     result,
		})
	}

	if form.EncryptedIDUsername != "" {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.EncryptedIDUsername},
		)
		member, _ := models.GetEntMemberFn(arrCond, "", false)
		if member == nil {
			message := app.MsgStruct{
				Msg: "invalid_nick_name",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}

		result := helpers.GetEncryptedID(member.Code, member.ID)

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{
			"encrypted_id_username": form.EncryptedIDUsername,
			"result":                result,
		})
	}

	if form.ProcessCryptoAddressChecking == "yes" {
		member_service.ProcessCryptoAddressChecking()
	}

	if form.KGraph == "yes" {
		arrData := trading_service.WSMemberExchangePriceTradingView{
			CryptoCode: "SEC",
			PeriodCode: "15MIN",
			LangCode:   "en",
		}
		result := trading_service.GetWSMemberExchangePriceTradingView(arrData)
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, result)
		return
	}

	if form.ProcessPKChecking == "yes" {
		member_service.ProcessPKChecking()
	}

	if strings.ToLower(form.ProcessLaligaCallBack) == "yes" {
		sales_service.ProcessLaligaCallBack(true)

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{})
		return
	}

	if strings.ToLower(form.EncryptOriMNValue) == "yes" {
		err := member_service.ProcessEncryptOriMNValue()

		msg := "success"
		if err != nil {
			msg = err.Error()
		}
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: msg}, map[string]interface{}{})
		return
	}

	// secP2PPoolWalletInfo, err := models.GetSECP2PPoolWalletInfo()
	// if err != nil {
	// 	fmt.Println("err msg", err)
	// }
	// fmt.Println("secP2PPoolWalletInfo:", secP2PPoolWalletInfo)

	// if strings.ToLower(form.TestProcessSendPushNotificationMsg) == "yes" {
	// 	notification_service.ProcessSendPushNotificationMsg(true)

	// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, map[string]interface{}{})
	// 	return
	// }

	if strings.ToLower(form.TestProcessSendPushNotificationMsg) == "yes" {

		msgTitleValue := map[string]string{
			"date":  "2021-05-17 16:00",
			"title": "Happy Niu Year",
		}

		msgValue := map[string]string{
			"amount": "RM50.50",
			"title":  "Happy Niu Year",
		}

		arrProcessPushNotificationDataV1 := base.ProcessPushNotificationDataV1Struct{
			EntMemberID: 11976,
			// PnMsgID: ,
			ApiKeysName:   "app",
			SourceID:      0,
			MsgTitle:      ":title_on_:date",
			MsgTitleValue: msgTitleValue,
			Msg:           "you_will_receive_:amount_for_:title",
			MsgValue:      msgValue,
			GroupName:     "",
			LangCode:      "en",
			// CustomData:"",
			// ArrFn         "",
		}

		base.ProcessPushNotificationDataV1(arrProcessPushNotificationDataV1, true, true, true)
	}

	if strings.ToLower(form.TestPDF) == "yes" {
		arrBZZContractPDF := sales_service.BZZContractPDFStruct{}
		arrBZZContractPDF.NickName = "안녕하세요"
		arrBZZContractPDF.SerialNumber = "4948916437"
		arrBZZContractPDF.LangCode = "kr"
		arrBZZContractPDF.DocNo = "MM00000385"
		arrBZZContractPDF.TotalAmount = "2000"
		arrBZZContractPDF.TotalNodes = "10"
		sales_service.GenerateBZZContractPDF(arrBZZContractPDF)
	}

	if form.TestCallEmailApi != "" {
		sendmailData := base.CallSendMailApiStruct{
			Subject: "重设密码请求",
			Message: "<h1>testing 123</h1><h3>Good Morning</h3>",
			Type:    "HTML",
			ToEmail: []string{form.TestCallEmailApi},
			ToName:  []string{"testuser1"},
		}

		err := sendmailData.CallSendMailApi()
		if err != nil {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}

		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
		return
	}

	if form.TestCommitOld == "yes" {
		tx := models.Begin()
		for i := 0; i < 3; i++ {
			ewtOut := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     2,
				EwalletTypeID:   1,
				TotalIn:         3,
				TransactionType: "TEST COMMIT " + strconv.Itoa(i),
				DocNo:           strconv.Itoa(i),
				Remark:          "TEST COMMIT " + strconv.Itoa(i),
				CreatedBy:       strconv.Itoa(2),
			}

			rst, err := wallet_service.SaveMemberWallet(tx, ewtOut)
			fmt.Println("rst:", rst)
			if err != nil {
				fmt.Println("TestCommit error", err.Error())
				os.Exit(0)
			}

			arrEwtBal := wallet_service.GetWalletBalanceStruct{
				Tx:          tx,
				EntMemberID: 2,
				EwtTypeID:   1,
			}
			walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
			fmt.Println(strconv.Itoa(i)+" balance:", walletBalance.Balance)
		}
		tx.Commit()
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: 2,
			EwtTypeID:   1,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		fmt.Println("last balance:", walletBalance.Balance)
	}

	if form.TestCommitNew == "yes" {
		tx := models.Begin()
		for i := 0; i < 3; i++ {
			ewtOut := wallet_service.SaveMemberWalletStruct{
				EntMemberID:     2,
				EwalletTypeID:   1,
				TotalIn:         3,
				TransactionType: "TEST COMMIT " + strconv.Itoa(i),
				DocNo:           strconv.Itoa(i),
				Remark:          "TEST COMMIT " + strconv.Itoa(i),
				CreatedBy:       strconv.Itoa(2),
			}

			rst, err := wallet_service.SaveMemberWalletTx(tx, ewtOut)
			fmt.Println("rst:", rst)
			if err != nil {
				fmt.Println("TestCommit error", err.Error())
				os.Exit(0)
			}

			arrEwtBal := wallet_service.GetWalletBalanceStruct{
				Tx:          tx,
				EntMemberID: 2,
				EwtTypeID:   1,
			}
			walletBalance := wallet_service.GetWalletBalanceTx(arrEwtBal)
			fmt.Println(strconv.Itoa(i)+" balance:", walletBalance.Balance)
		}
		tx.Commit()
		arrEwtBal := wallet_service.GetWalletBalanceStruct{
			EntMemberID: 2,
			EwtTypeID:   1,
		}
		walletBalance := wallet_service.GetWalletBalance(arrEwtBal)
		fmt.Println("last balance:", walletBalance.Balance)
	}
	// base.LogErrorLog("tes456", nil, nil, true)
	// url := "https://admin.sechain.io/api/job/rebuild-tree"
	// header := map[string]string{
	// 	"X-Authorization": "ziHBLv2rZaQlJRGdC3JUEBErti4pM25bPHuZuMdF6hAD2zzOTP4Wcqfc0vcy6ZF5",
	// }
	// extraSetting := base.ExtraSettingStruct{
	// 	InsecureSkipVerify: true,
	// }

	// rst, err := base.RequestAPIV2("GET", url, header, nil, nil, extraSetting)

	// if err != nil {
	// 	fmt.Println("err:", err.Error())
	// }
	// fmt.Println("rst:", rst)
	// addressList := []string{
	// 	"0xb0018ae2df59F09f7ad0e84F18A2c2150Df4136f",
	// }

	// result, err := wallet_service.ProcessGetBatchTransactionNonceViaAPI(addressList)
	// fmt.Println("err:", err)
	// fmt.Println("result:", err)
	// for resultK, resultV := range result {
	// 	fmt.Println("count:", resultK)
	// 	fmt.Println("CryptoAddr:", resultV.CryptoAddr)
	// 	fmt.Println("Nonce:", resultV.Nonce)
	// }
	// trading_service.TestDBTranx()
	// nonce, err := wallet_service.GetTransactionNonceViaAPI("0x3dab29a5fe3d346fc7f505026498122bd828bd02")
	// fmt.Println(err)
	// fmt.Println(nonce)
	// x := 80.00000001677
	// fmt.Println("Substr:", helpers.Substr("abcd1234.001232", 0, 4))
	// u, _ := c.Get("access_user")
	// members := u.(*models.EntMemberMembers)
	// ewtTypeCode = strings.Replace(c.Param("ewallet_type_code"), "/", "", -1)

	// if ewtTypeCode != "" {
	// 	ewtTypeCode = strings.Trim(ewtTypeCode, " ")
	// }
	// fmt.Println("WalletTypeID:", form.WalletTypeID)
	// fmt.Println("WalletTypeID:", c.PostForm("wallet_type_id"))
	// fmt.Printf("%T\n", c.PostForm("wallet_type_id"))
	// fmt.Printf("%T\n", form.total_in)
	// walletTypeID, _ := strconv.ParseInt(c.PostForm("ewallet_type_id"), 64)
	// totalIn, _ := strconv.ParseFloat(c.PostForm("total_in"), 64)
	// totalOut, _ := strconv.ParseFloat(c.PostForm("total_out"), 64)

	/* start test SaveMemberWallet function
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	member, _ := models.GetEntMemberFn(arrCond, "", false)

	if member != nil {
	tx := models.Begin()
	ewtIn := wallet_service.SaveMemberWalletStruct{
		EntMemberID:     member.ID,
		//EntMemberID: 11973,
		EwalletTypeID:   form.WalletTypeID,
		// EwalletTypeID:   10,
		TotalIn:         form.TotalIn,  // if this is pass (float64), TotalOut must empty / 0
		TotalOut:        form.TotalOut, // if this is pass (float64), TotalIn must empty / 0
		TransactionType: "ADJUSTMENT",
		DocNo:           "",
		Remark:          "ADJUSMENT FOR TESTING",
		CreatedBy:       "6",
	}
	fmt.Println("ewtIn:", ewtIn)
	_, err := wallet_service.SaveMemberWallet(tx, ewtIn)
	// models.Rollback(tx)
	if err != nil {
		fmt.Println(err)
		fmt.Println("error kootest")
		models.Rollback(tx)
		appG.Response(0, http.StatusOK, err.Error(), ewtIn)
		return
	}
	// fmt.Println(ewtDetailID)

	err = models.Commit(tx)
	if err != nil {
		models.Rollback(tx)

	}
	appG.Response(1, http.StatusOK, "success", "")
	return
	// }
	*/ // end here
	/* start test for processRoomQueue  */
	// room_service.ProcessRoomQueue()
	// room_service.ProcessRoomWinner(10)
	// room_service.CalWodMemberStarRanking(6)
	// room_service.ProcessBringForwardPendingRoom()
	// room_service.ProcessRoomLive()
	// fmt.Println("err in koo controller:", err)
	// appG.Response(1, http.StatusOK, "success", "")
}

// func QuikTest
func Testing(c *gin.Context) {

	// fmt.Println("a")
}

func KooTestCors(c *gin.Context) {
	fmt.Println("hihihi")
}
