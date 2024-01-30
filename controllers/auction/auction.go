package auction

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

// GetMemberDetailsForm struct
type GetMemberDetailsForm struct {
	Username      string `form:"username" json:"username" valid:"Required;"`
	CryptoType    string `form:"crypto_type" json:"crypto_type" valid:"Required;"`
	CryptoAddress string `form:"crypto_address" json:"crypto_address" valid:"Required;"`
}

func GetMemberDetails(c *gin.Context) {
	var (
		appG    = app.Gin{C: c}
		form    GetMemberDetailsForm
		arrData = map[string]interface{}{}
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// get member info
	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: form.Username},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrEntMember, _ := models.GetEntMemberFn(arrEntMemberFn, "", false)
	if arrEntMember == nil {
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "invalid_username"}, nil)
		return
	}

	var memID = arrEntMember.ID
	arrData["member_id"] = memID
	arrData["username"] = arrEntMember.NickName

	// get sponsor info
	arrEntMemberSponsorFn := make([]models.WhereCondFn, 0)
	arrEntMemberSponsorFn = append(arrEntMemberSponsorFn,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.member_id = ?", CondValue: memID},
	)
	arrEntMemberSponsor, _ := models.GetMemberSponsorFn(arrEntMemberSponsorFn, false)

	var sponsorID = 0
	var sponsorUsername = ""
	if arrEntMemberSponsor != nil {
		sponsorID = arrEntMemberSponsor.SponsorID
		sponsorUsername = arrEntMemberSponsor.SponsorUsername
	}

	arrData["sponsor_id"] = sponsorID
	arrData["sponsor_username"] = sponsorUsername

	// get sponsor wallet address
	db := models.GetDB() // no need set begin transaction
	sponsorCryptoAddr, err := member_service.ProcessGetMemAddress(db, sponsorID, form.CryptoType)
	if err != nil {
		base.LogErrorLog("auctionController:GetMemberDetails()", "ProcessGetMemAddress():1", err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, arrData)
		return
	}

	arrData["sponsor_wallet_address"] = sponsorCryptoAddr

	// get access token language code
	// arrData["lang_code"] = "en" // default put en

	// get wallet info
	arrMemberBlockchainWalletBalance, errMsg := wallet_service.GetMemberBlockchainWalletBalance(memID, form.CryptoType, form.CryptoAddress)
	if errMsg != "" {
		appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	arrData["balance"] = arrMemberBlockchainWalletBalance.Balance
	arrData["available_balance"] = arrMemberBlockchainWalletBalance.AvailableBalance

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}
