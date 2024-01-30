package file

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// type KooTestFrom struct {
// 	Crypto                             string  `form:"crypto" json:"crypto"`
// 	CryptoType                         string  `form:"crypto_type" json:"crypto_type"`
// 	RSAText                            string  `form:"rsa_text" json:"rsa_text"`
// 	EncryptedKey                       string  `form:"encrypted_key" json:"encrypted_key"`
// 	CutOffDecimalAmount                float64 `form:"cut_off_decimal_amount" json:"cut_off_decimal_amount"`
// 	Decimal                            uint    `form:"decimal" json:"decimal"`
// 	RoundDownAmount                    float64 `form:"round_down_amount" json:"round_down_amount"`
// 	EntMemberIDCryptoAddress           int     `form:"ent_member_id_crypto_address" json:"ent_member_id_crypto_address"`
// 	CompanyAddress                     string  `form:"company_address" json:"company_address"`
// 	AddMonitorAddr                     string  `form:"add_monitor_addr" json:"add_monitor_addr"`
// 	AddAllMonitorAddr                  string  `form:"add_all_monitor_addr" json:"add_all_monitor_addr"`
// 	ScryptText                         string  `form:"scrypt_text" json:"scrypt_text"`
// 	TestAutoMatchTrading               string  `form:"test_auto_match_trading" json:"test_auto_match_trading"`
// 	GetBlockchainWalletBalanceApiV1    string  `form:"get_blockchain_wallet_balanc_apiv1" json:"get_blockchain_wallet_balanc_apiv1"`
// 	TotalIn                            float64 `form:"total_in" json:"total_in"`
// 	TotalOut                           float64 `form:"total_out" json:"total_out"`
// 	PNOs                               string  `form:"pn_os" json:"pn_os"`
// 	PNGroupName                        string  `form:"pn_group_name" json:"pn_group_name"`
// 	ProcessUpdateMissingMemberCode     string  `form:"process_update_missing_member_code" json:"process_update_missing_member_code"`
// 	SigningKey                         string  `form:"signing_key" json:"signing_key"`
// 	EncryptedIDUsername                string  `form:"encrypted_id_username" json:"encrypted_id_username"`
// 	ProcessCryptoAddressChecking       string  `form:"process_crypto_address_checking" json:"process_crypto_address_checking"`
// 	KGraph                             string  `form:"k_graph" json:"k_graph"`
// 	TestProcessSendPushNotificationMsg string  `form:"test_process_send_pn_msg" json:"test_process_send_pn_msg"`
// 	ProcessPKChecking                  string  `form:"process_pk_checking" json:"process_pk_checking"`
// 	ProcessLaligaCallBack              string  `form:"process_laliga_callback" json:"process_laliga_callback"`
// 	EncryptOriMNValue                  string  `form:"encrypt_ori_mnvalue" json:"encrypt_ori_mnvalue"`
// 	TestPDF                            string  `form:"test_pdf" json:"test_pdf"`
// 	// TestValidation                  float64 `form:"test_validation" json:"test_validation" valid:"Required;Min(0);"`
// }

// func ServePDFFile
func ServePDFFile(c *gin.Context) {

	filename := strings.Replace(c.Param("filename"), "/", "", -1)
	prdGroup := strings.Replace(c.Param("prdGroup"), "/", "", -1)

	filePath := "./docs/member/sales/" + prdGroup + "/" + filename

	// c.File(filePath)
	// http.ServeFile(c.Writer, c.Request, filePath)

	//Seems this headers needed for some browsers (for example without this headers Chrome will download files as txt)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)
	//
	// apiServerDomain := setting.Cfg.Section("custom").Key("ApiServerDomain").String()
	// url := apiServerDomain + "/member/sales/contract/view/" + filename
	// fmt.Println("url:", url)
	// response, err := http.Get(url)
	// if err != nil || response.StatusCode != http.StatusOK {
	// 	fmt.Println("hihi")
	// 	fmt.Println("err:", err)
	// 	fmt.Println("response:", response.StatusCode)
	// 	c.Status(http.StatusServiceUnavailable)
	// 	return
	// }

	// reader := response.Body
	// defer reader.Close()
	// contentLength := response.ContentLength
	// contentType := response.Header.Get("Content-Type")

	// extraHeaders := map[string]string{
	// 	"Content-Disposition": `attachment; filename="` + filename + `"`,
	// }

	// c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)

}
