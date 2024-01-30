package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/member_service"
)

// RequestMnemonic function
func RequestMnemonicv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	callGenerateMnemonicApiRst, err := member_service.CallGenerateMnemonicApi()

	if err != nil {
		message := app.MsgStruct{
			Msg: "something_went_wrong",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	if callGenerateMnemonicApiRst == nil {
		message := app.MsgStruct{
			Msg: "something_went_wrong",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}
	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, callGenerateMnemonicApiRst)
	return
}

// BindMnemonicv1Form struct
type BindMnemonicv1Form struct {
	Username   string `form:"username" json:"username" valid:"Required;MinSize(4);MaxSize(19)"`
	PrivateKey string `form:"private_key" json:"private_key" valid:"Required"`
	CryptoAddr string `form:"crypto_addr" json:"crypto_addr" valid:"Required"`
	PK         string `form:"pk" json:"pk"`
	Mn         string `form:"mn" json:"mn"`
}

// func BindMnemonicv1
func BindMnemonicv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form BindMnemonicv1Form
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// start checking on private key is existing or not
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_crypto.private_key = ?", CondValue: form.PrivateKey},
		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)
	arrExistingPrivateKey, _ := models.GetEntMemberCryptoFn(arrCond, false)

	if arrExistingPrivateKey != nil {
		message := app.MsgStruct{
			Msg: "this_mnemonic_is_not_available",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}
	// start checking on private key is existing or not

	// start checking username is valid or not
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "I"},
		models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: member.ID}, // to double checking this account is belong to this logined member or not
	)
	arrUnbindMnemonicEntMember, _ := models.GetEntMemberFn(arrCond, "", false)

	if arrUnbindMnemonicEntMember == nil {
		message := app.MsgStruct{
			Msg: "this_account_is_not_available",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}
	// end checking username is valid or not

	// start checking on crypto address is existing or not
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_crypto.crypto_address = ?", CondValue: form.CryptoAddr},
		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)
	arrExistingMemCrypto, _ := models.GetEntMemberCryptoFn(arrCond, false)

	if arrExistingMemCrypto != nil {
		message := app.MsgStruct{
			Msg: "this_crypto_address_is_not_available",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}
	// end checking on crypto address is existing or not

	// get profile by username
	arrData := member_service.BindMnemonicv1Struct{
		Username:        form.Username,
		CryptoAddress:   form.CryptoAddr,
		PrivateKey:      form.PrivateKey,
		EntMemberID:     member.EntMemberID,
		BindEntMemberID: arrUnbindMnemonicEntMember.ID,
		PK:              form.PK,
		Mn:              form.Mn,
	}

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

	arrDataReturn := map[string]interface{}{
		"encrypted_id": helpers.GetEncryptedID(member.Code, member.EntMemberID), // this is refer to ent_member.id
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
	return
}

// UpdatePKInfov1Form struct
type UpdatePKInfov1Form struct {
	PK string `form:"pk" json:"pk" valid:"Required"`
	Mn string `form:"mn" json:"mn"`
}

// func UpdatePKInfov1
func UpdatePKInfov1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form UpdatePKInfov1Form
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")

	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// start checking on private key is existing or not
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member_crypto.crypto_type = ? ", CondValue: "SEC"},
		models.WhereCondFn{Condition: " ent_member_crypto.member_id = ? ", CondValue: member.EntMemberID},
		models.WhereCondFn{Condition: " ent_member_crypto.status = ? ", CondValue: "A"},
	)
	arrExistingPrivateKey, _ := models.GetEntMemberCryptoFn(arrCond, false)

	if arrExistingPrivateKey == nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
		return
	} else {
		updateDBStatus := false
		arrUpdCond := arrCond
		updateColumn := map[string]interface{}{}

		if arrExistingPrivateKey.PrivateKey != "" {
			// start decrypt for PK
			decryptedPKString, err := util.RsaDecryptPKCS1v15(form.PK)
			if err != nil {
				base.LogErrorLog("UpdatePKInfov1-RsaDecryptPKCS1v15_PK_failed", err.Error(), form.PK, true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
			pk := decryptedPKString
			// end decrypt for Mn
			updateDBStatus = true
			updateColumn["private_key"] = pk
		}

		if arrExistingPrivateKey.Mnemonic != "" {
			if form.Mn != "" {
				// start decrypt for Mn
				decryptedMnString, err := util.RsaDecryptPKCS1v15(form.Mn)
				if err != nil {
					base.LogErrorLog("UpdatePKInfov1-RsaDecryptPKCS1v15_Mn_failed", err.Error(), form.Mn, true)
					appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
					return
				}
				mn := decryptedMnString
				// end decrypt for Mn

				// start encrypt
				encryptedMNText, err := util.RsaEncryptPKCS1v15(mn)
				if err != nil {
					appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
					return
				}
				// end encrypt
				updateDBStatus = true
				updateColumn["mn"] = encryptedMNText
			}
		}

		if updateDBStatus {
			tx := models.Begin()
			err := models.UpdatesFnTx(tx, "ent_member_crypto", arrUpdCond, updateColumn, false)

			if err != nil {
				models.Rollback(tx)
				arrErr := map[string]interface{}{
					"arrUpdCond":   arrUpdCond,
					"updateColumn": updateColumn,
				}
				base.LogErrorLog("UpdatePKInfov1-update_ent_member_crypto_failed", err.Error(), arrErr, true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}

			// commit transaction
			err = models.Commit(tx)

			if err != nil {
				base.LogErrorLog("UpdatePKInfov1-commit_ent_member_crypto_failed", err.Error(), nil, true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}
