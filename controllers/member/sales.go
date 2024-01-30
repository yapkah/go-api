package member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/pkg/util"
	"github.com/yapkah/go-api/service/sales_service"
)

// MemberSalesForm struct
type MemberSalesForm struct {
	DocType  string `form:"doc_type" json:"doc_type"`
	PrdCode  string `form:"prd_code" json:"prd_code"`
	Page     int64  `form:"page" json:"page"`
	DateFrom string `form:"date_from" json:"date_from"`
	DateTo   string `form:"date_to" json:"date_to"`
}

//func GetMemberSalesListv1 function
func GetMemberSalesListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberSalesForm
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	if form.Page < 1 {
		form.Page = 1
	}

	// member.EntMemberID = 1
	arrData := sales_service.MemberSalesStruct{
		MemberID: member.EntMemberID,
		NickName: member.NickName,
		LangCode: langCode,
		Page:     form.Page,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		DocType:  form.DocType,
		PrdCode:  form.PrdCode,
	}

	arrMemberSalesList, errMsg := sales_service.GetMemberSalesListv1(arrData)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: errMsg,
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrMemberSalesList)
}

// MemberSalesTopupListForm struct
type MemberSalesTopupListForm struct {
	DocNo    string `form:"doc_no" json:"doc_no" valid:"Required;"`
	DateFrom string `form:"date_from" json:"date_from"`
	DateTo   string `form:"date_to" json:"date_to"`
	Page     int64  `form:"page" json:"page"`
}

//func GetMemberSalesTopupListv1 function
func GetMemberSalesTopupListv1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form MemberSalesTopupListForm
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	if form.Page < 1 {
		form.Page = 1
	}

	// member.EntMemberID = 1
	arrData := sales_service.MemberSalesTopupStruct{
		MemberID: member.EntMemberID,
		DocNo:    form.DocNo,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
		Page:     form.Page,
		LangCode: langCode,
	}

	arrMemberSalesList, errMsg := sales_service.GetMemberSalesTopupListv1(arrData)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: errMsg,
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrMemberSalesList)
}

// GetMemberMiningNodeListForm struct
type GetMemberMiningNodeListForm struct {
	Page int64 `form:"page" json:"page"`
}

//func GetMemberMiningNodeListV1 function
func GetMemberMiningNodeListV1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberMiningNodeListForm
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	if form.Page < 1 {
		form.Page = 1
	}

	// member.EntMemberID = 1
	arrData := sales_service.GetMemberMiningNodeList{
		MemberID: member.EntMemberID,
		Page:     form.Page,
		LangCode: langCode,
	}
	// if member.EntMemberID == 12246 {
	// 	base.LogErrorLogV2("before call GetMemberMiningNodeListV1:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
	// }
	arrMemberMiningNodeList, errMsg := sales_service.GetMemberMiningNodeListV1(arrData)
	// if member.EntMemberID == 12246 {
	// 	base.LogErrorLogV2("after call GetMemberMiningNodeListV1:", time.Now().Unix(), time.Now().UnixNano(), true, "koobot")
	// }
	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: errMsg,
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}

	appG.ResponseV2(1, http.StatusOK, message, arrMemberMiningNodeList)
}

// GetMemberMiningNodeListUpdateForm struct
type GetMemberMiningNodeListUpdateForm struct {
	NodeID int `form:"node_id" json:"node_id" valid:"Required;"`
}

//func GetMemberMiningNodeListUpdateV1 function
func GetMemberMiningNodeListUpdateV1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberMiningNodeListUpdateForm
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	var arrMemberMiningNodeList = sales_service.GetMemberMiningNodeListUpdateV1(member.EntMemberID, form.NodeID, langCode)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrMemberMiningNodeList)
}

// GetMemberMiningNodeTopupListForm struct
type GetMemberMiningNodeTopupListForm struct {
	NodeID int   `form:"node_id" json:"node_id" valid:"Required;"`
	Page   int64 `form:"page" json:"page"`
}

//func GetMemberMiningNodeTopupListV1 function
func GetMemberMiningNodeTopupListV1(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form GetMemberMiningNodeTopupListForm
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := sales_service.GetMemberMiningNodeTopupList{
		MemberID: member.EntMemberID,
		NickName: member.NickName,
		NodeID:   form.NodeID,
		Page:     form.Page,
		LangCode: langCode,
	}
	arrMemberMiningNodeTopupList, errMsg := sales_service.GetMemberMiningNodeTopupListV1(arrData)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: errMsg}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrMemberMiningNodeTopupList)
}

type AddBallotForm struct {
	Payments     string `form:"payments" json:"payments"`
	SecondaryPin string `form:"secondary_pin" json:"secondary_pin" valid:"Required;"`
}

func PostBallot(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AddBallotForm
		err  error
	)

	tx := models.Begin()

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	decryptedText, err := util.RsaDecryptPKCS1v15(form.SecondaryPin)
	if err != nil {
		base.LogErrorLog("PostBallot-RsaDecryptPKCS1v15_Failed", err.Error(), form.SecondaryPin, true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "invalid_secondary_pin"}, nil)
		return
	}

	form.SecondaryPin = decryptedText

	// check secondary password
	pinValidation := base.SecondaryPin{
		MemId:              entMemberID,
		SecondaryPin:       form.SecondaryPin,
		MemberSecondaryPin: member.SecondaryPin,
		LangCode:           langCode,
	}

	err = pinValidation.CheckSecondaryPin()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
		return
	}

	// // start check on member kyc
	// arrCond := make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: entMemberID},
	// )
	// arrEntMemberKyc, _ := models.GetEntMemberKycFn(arrCond, false)

	// kycStatus := false
	// if len(arrEntMemberKyc) > 0 {
	// 	if arrEntMemberKyc[0].Status == "AP" {
	// 		kycStatus = true
	// 	}
	// }

	// if !kycStatus {
	// 	appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "please_wait_for_kyc_to_be_approve"}, nil)
	// 	return
	// }

	ballot := sales_service.PostBallotStruct{
		MemberId: entMemberID,
		Payments: form.Payments,
	}
	arrData, err := ballot.PostBallot(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "ballot_successful"}, arrData)
	return
}

type MemberBallotForm struct {
	Page int64 `form:"page" json:"page"`
}

func GetMemberBallotList(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberBallotForm
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	ballotList := sales_service.BallotTransactionStruct{
		MemberID: entMemberID,
		LangCode: langCode,
		Page:     form.Page,
	}
	arrMemberBallotList, err := ballotList.GetMemberBallotListv1()

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: "fail_to_get_ballot_list",
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrMemberBallotList)
}

func GetBallotSetting(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	// get user info
	u, ok := c.Get("access_user")
	if !ok {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	// get lang code
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
	}

	arrData, err := sales_service.GetBallotSetting(member.EntMemberID, langCode)

	if err != nil {
		appG.ResponseError(err)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
	return
}

type AddBallotWinnerForm struct {
	Address string `form:"address" json:"address"`
}

func PostBallotWinner(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form AddBallotWinnerForm
		err  error
	)

	tx := models.Begin()

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.PostForm("lang_code") != "" {
		langCode = c.PostForm("lang_code")
	} else if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	u, ok := c.Get("access_user")
	if ok == false {
		appG.ResponseV2(0, http.StatusUnauthorized, app.MsgStruct{Msg: "invalid_member"}, nil)
		return
	}

	member := u.(*models.EntMemberMembers)
	entMemberID := member.EntMemberID

	ballot := sales_service.PostBallotWinnerStruct{
		MemberId: entMemberID,
		Address:  form.Address,
		LangCode: langCode,
	}
	arrData, err := ballot.PostBallotWinner(tx)

	if err != nil {
		tx.Rollback()
		appG.ResponseError(err)
		return
	}

	tx.Commit()

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "save_successfully"}, arrData)
	return
}

// MemberSalesListSummary struct
type MemberSalesListSummary struct {
	DocType string `form:"doc_type" json:"doc_type"`
}

//func GetMemberSalesListSummary function
func GetMemberSalesListSummary(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberSalesListSummary
	)
	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		langCode = c.GetHeader("Accept-Language")
	}

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, "")
		return
	}

	member := u.(*models.EntMemberMembers)

	// member.EntMemberID = 1
	arrData := sales_service.MemberSalesListSummaryStruct{
		MemberID: member.EntMemberID,
		DocType:  form.DocType,
		LangCode: langCode,
	}

	arrMemberSalesList, errMsg := sales_service.GetMemberSalesListSummary(arrData)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: errMsg,
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrMemberSalesList)
}
