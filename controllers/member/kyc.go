package member

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
)

// CreateMemberKYCForm struct
type CreateMemberKYCForm struct {
	Name          string `form:"name" json:"name" valid:"Required;MaxSize(200)"`
	IC            string `form:"ic" json:"ic" valid:"Required;MaxSize(30);"`
	WalletAddress string `form:"wallet_address" json:"wallet_address" valid:"Required"`
	Email         string `form:"email" json:"email" valid:"Required"`
	CountryCode   string `form:"country_code" json:"country_code" valid:"Required"`
	SecondaryPin  string `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

// CreateMemberKYCv1 function
func CreateMemberKYCv1(c *gin.Context) {
	var (
		appG      = app.Gin{C: c}
		form      CreateMemberKYCForm
		status    string
		fileName1 string
		fileURL1  string
		fileName2 string
		fileURL2  string
		fileName3 string
		fileURL3  string
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.PostForm("lang_code") != "" {
	// 	langCode = c.PostForm("lang_code")
	// } else if c.GetHeader("Accept-Language") != "" {
	// 	langCode = c.GetHeader("Accept-Language")
	// }

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	// get access user from middle ware
	u, ok := c.Get("access_user")
	if !ok {
		// user not found
		message := app.MsgStruct{
			Msg: "something_went_wrong",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	ok = models.ExistCountryCode(form.CountryCode)
	if !ok {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_country_code"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("CreateMemberKYCv1-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	secondaryPin := base.SecondaryPin{
		MemId:              member.EntMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
	}

	secondaryPinErr := secondaryPin.CheckSecondaryPin()

	if secondaryPinErr != nil {
		message := app.MsgStruct{
			Msg: secondaryPinErr.Error(),
		}
		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	arrCountryData, err := models.GetCountryByCode(strings.ToUpper(form.CountryCode))
	if err != nil {
		message := app.MsgStruct{
			Msg: "country_does_not_exists_in_system",
		}
		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	country_id := arrCountryData.ID

	// start checking on not allow to create / update if member ent_member_kyc.status = "AP" is exists
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: member.EntMemberID},
		models.WhereCondFn{Condition: "ent_member_kyc.status = ?", CondValue: "AP"},
	)
	existingApproveIC, _ := models.GetEntMemberKycFn(arrCond, false)

	if len(existingApproveIC) > 0 {
		message := app.MsgStruct{
			Msg: "not_allow_to_update_kyc_application",
		}
		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	//check pending
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: member.EntMemberID},
		models.WhereCondFn{Condition: "ent_member_kyc.status = ?", CondValue: "P"},
	)
	checkSubmitStatus, _ := models.GetEntMemberKycFn(arrCond, false)

	if len(checkSubmitStatus) > 0 {
		message := app.MsgStruct{
			Msg: "application_already_submitted_before",
		}
		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	// end checking on not allow to create / update if member ent_member_kyc.status = "AP" is exists

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_kyc.member_id != ?", CondValue: member.EntMemberID},
		models.WhereCondFn{Condition: "ent_member_kyc.identity_no = ?", CondValue: form.IC},
		models.WhereCondFn{Condition: "ent_member_kyc.status = ?", CondValue: "AP"},
	)
	existingIC, _ := models.GetEntMemberKycFn(arrCond, false)
	if len(existingIC) > 0 {
		message := app.MsgStruct{
			Msg: "identification_no_is_duplicated",
		}
		appG.ResponseV2(0, http.StatusOK, message, "")
		return
	}

	// start this is to replace update action. perform only save new record. prev record will remain as log.
	// step get back all the prev image records and insert back
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: member.EntMemberID},
	)
	existingIC, _ = models.GetEntMemberKycFn(arrCond, false)

	if len(existingIC) > 0 {
		if existingIC[0].Status != "R" {
			status = existingIC[0].Status
		} else if existingIC[0].Status == "R" {
			status = "P"
		}
		fileName1 = existingIC[0].FileName1
		fileURL1 = existingIC[0].FileURL1
		fileName2 = existingIC[0].FileName2
		fileURL2 = existingIC[0].FileURL2
		fileName3 = existingIC[0].FileName3
		fileURL3 = existingIC[0].FileURL3
	}
	// end this is to replace update action. perform only save new record. prev record will remain as log.

	arrData := models.AddEntMemberKYCStruct{
		MemberID:      member.EntMemberID,
		FullName:      form.Name,
		IdentityNo:    form.IC,
		CountryID:     country_id,
		WalletAddress: form.WalletAddress,
		Email:         form.Email,
		FileName1:     fileName1,
		FileURL1:      fileURL1,
		FileName2:     fileName2,
		FileURL2:      fileURL2,
		FileName3:     fileName3,
		FileURL3:      fileURL3,
		Status:        status, // this code can't be here bcz the file is not upload successfully. update it back in upload file action. unless no file need to upload
		CreatedBy:     strconv.Itoa(member.EntMemberID),
	}

	tx := models.Begin()

	// add ent_member_kyc record
	_, err = models.AddEntMemberKYC(tx, arrData)
	if err != nil {
		models.Rollback(tx)
		appG.ResponseError(err)
		return
	}

	err = models.Commit(tx)
	if err != nil {
		appG.ResponseError(err)
		return
	}

	// arrDataReturn := map[string]interface{}{
	// 	"id": result.ID,
	// }

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, "")
	return
}

// GetMemberKYCv1 function
func GetMemberKYCv1(c *gin.Context) {
	var (
		appG             = app.Gin{C: c}
		fullName         string
		identityNo       string
		country          string
		countryCode      string
		bscWalletAddress string
		email            string
		frontIcImg       string
		backIcImg        string
		selfIcImg        string
		kycStatus        string
		remark           string
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	// get access user from middle ware
	u, ok := c.Get("access_user")
	if !ok {
		// user not found
		message := app.MsgStruct{
			Msg: "something_went_wrong",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}
	member := u.(*models.EntMemberMembers)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: member.EntMemberID},
	)

	existingIC, _ := models.GetEntMemberKycFn(arrCond, false)
	if len(existingIC) > 0 {

		arrCountryData, _ := models.GetCountryByID(existingIC[0].CountryID)
		var countryName string

		if arrCountryData.Name != "" {
			countryName = helpers.Translate(arrCountryData.Name, langCode)
		}

		fullName = existingIC[0].FullName
		identityNo = existingIC[0].IdentityNo
		country = countryName
		countryCode = arrCountryData.Code
		bscWalletAddress = existingIC[0].WalletAddress
		email = existingIC[0].Email
		frontIcImg = existingIC[0].FileURL1
		backIcImg = existingIC[0].FileURL2
		selfIcImg = existingIC[0].FileURL3
		kycStatus = existingIC[0].Status
		remark = existingIC[0].Remark
	}

	arrDataReturn := map[string]interface{}{
		"full_name":          fullName,
		"ic":                 identityNo,
		"country":            country,
		"country_code":       countryCode,
		"bsc_wallet_address": bscWalletAddress,
		"email":              email,
		"front_ic_img":       frontIcImg,
		"back_ic_img":        backIcImg,
		"self_ic_img":        selfIcImg,
		"kyc_status":         kycStatus,
		"remark":             remark,
	}
	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
	return
}
