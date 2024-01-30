package product

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/member_service"
	"github.com/yapkah/go-api/service/product_service"
	"github.com/yapkah/go-api/service/reward_service"
	"github.com/yapkah/go-api/service/wallet_service"
)

type MemberMiningActionListv1Form struct {
	Version string `form:"version" json:"version"`
}

//func GetMemberMiningActionListv1
func GetMemberMiningActionListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form MemberMiningActionListv1Form
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
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

	version := "old"
	if form.Version != "" {
		version = "new"
	}

	arrData := product_service.MemberMiningActionListv1ReqStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
		Version:     version,
	}
	arrDataReturn, err := product_service.GetMemberMiningActionListv1(arrData)

	if err != "" {
		message := app.MsgStruct{
			Msg: err,
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

type MemberContractMiningActionDetailStruct struct {
	MarketPrice             product_service.MarketPriceStruct `json:"market_price"`
	Ranking                 *member_service.RankingInfoStruct `json:"ranking"`
	TodaySponsor            string                            `json:"today_sponsor"`
	AccumulatedSalesAmount  string                            `json:"accumulated_amount"`
	TodaySponsorSalesAmount string                            `json:"today_sales_amount"`
	MatchingLevel           string                            `json:"matching_level"`
	TotalSponsor            string                            `json:"total_sponsor"`
	TotalSponsorSalesAmount string                            `json:"total_sponsor_sales_amount"`
	PurchaseContractStatus  int                               `json:"purchase_contract_status"`
	IncomeCap               struct {
		Total   string  `json:"total"`
		Balance string  `json:"balance"`
		Percent float64 `json:"percent"`
	} `json:"income_cap"`
	PoolAmount     string `json:"pool_amount"`
	LigaPoolAmount string `json:"liga_pool_amount"`
	UsdPoolAmount  string `json:"usd_pool_amount"`
}

//func GetMemberContractMiningActionDetailsv1
func GetMemberContractMiningActionDetailsv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
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

	arrData := product_service.MemberContractMiningActionDetailsv1ReqStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}
	miningDetails, miningDetailsErr := product_service.GetMemberContractMiningActionDetailsv1(arrData)
	if miningDetailsErr != "" {
		message := app.MsgStruct{
			Msg: miningDetailsErr,
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	// get member ranking info
	rankingInfo, rankingInfoErr := member_service.GetMemberRankingInfo(member.EntMemberID)
	if rankingInfoErr != "" {
		message := app.MsgStruct{
			Msg: rankingInfoErr,
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	arrDataReturn := MemberContractMiningActionDetailStruct{
		MarketPrice:             miningDetails.MarketPrice,
		Ranking:                 rankingInfo,
		TodaySponsor:            miningDetails.TodaySponsor,
		AccumulatedSalesAmount:  miningDetails.AccumulatedSalesAmount,
		TodaySponsorSalesAmount: miningDetails.TodaySponsorSalesAmount,
		MatchingLevel:           miningDetails.MatchingLevel,
		TotalSponsor:            miningDetails.TotalSponsor,
		TotalSponsorSalesAmount: miningDetails.TotalSponsorSalesAmount,
		PurchaseContractStatus:  1,
		IncomeCap:               miningDetails.IncomeCap,
		PoolAmount:              miningDetails.PoolAmount,
		LigaPoolAmount:          miningDetails.LigaPoolAmount,
		UsdPoolAmount:           miningDetails.UsdPoolAmount,
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
	return
}

//func GetMemberStakingMiningActionDetailsv1
func GetMemberStakingMiningActionDetailsv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
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

	arrData := product_service.MemberStakingMiningActionDetailsv1ReqStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}
	arrDataReturn, err := product_service.GetMemberStakingMiningActionDetailsv1(arrData)

	if err != "" {
		message := app.MsgStruct{
			Msg: err,
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

//func GetMemberPoolMiningActionDetailsv1
func GetMemberPoolMiningActionDetailsv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
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

	arrData := product_service.MemberPoolMiningActionDetailsv1ReqStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}
	miningDetails := product_service.GetMemberPoolMiningActionDetailsv1(arrData)
	// get member ranking info
	rankingInfo, rankingInfoErr := member_service.GetMemberPoolRankingInfo(member.EntMemberID)
	if rankingInfoErr != "" {
		message := app.MsgStruct{
			Msg: rankingInfoErr,
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member.nick_name = ?", CondValue: "secpool"},
	)
	arrSecPoolEntMem, _ := models.GetEntMemberFn(arrCond, "", false)
	secP2pWalletInfo, err := models.GetSECP2PPoolWalletInfo()
	if err != nil {
		base.LogErrorLog("GetMemberPoolMiningActionDetailsv1-GetSECP2PPoolWalletInfo_failed", err.Error(), nil, false)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}
	balRst := wallet_service.GetBlockchainWalletBalanceByAddressV1("SEC", secP2pWalletInfo.WalletAddress, arrSecPoolEntMem.ID)
	secPoolBal := "0.00"
	if balRst.AvailableBalance > 0 {
		secPoolBal = helpers.CutOffDecimal(balRst.AvailableBalance, 8, ".", ",")
	}

	type memberContractMiningActionPoolDetailStruct struct {
		MarketPrice             product_service.MarketPriceStruct `json:"market_price"`
		Ranking                 *member_service.RankingInfoStruct `json:"ranking"`
		TodaySponsor            string                            `json:"today_sponsor"`
		AccumulatedSalesAmount  string                            `json:"accumulated_amount"`
		TodaySponsorSalesAmount string                            `json:"today_sales_amount"`
		MatchingLevel           string                            `json:"matching_level"`
		TotalSponsor            string                            `json:"total_sponsor"`
		TotalSponsorSalesAmount string                            `json:"total_sponsor_sales_amount"`
		PurchaseContractStatus  int                               `json:"purchase_contract_status"`
		IncomeCap               struct {
			Total   string  `json:"total"`
			Balance string  `json:"balance"`
			Percent float64 `json:"percent"`
		} `json:"income_cap"`
		PoolAmount      string `json:"pool_amount"`
		TotalPoolAmount string `json:"total_pool_amount"`
	}

	arrDataReturn := memberContractMiningActionPoolDetailStruct{
		MarketPrice:             miningDetails.MarketPrice,
		Ranking:                 rankingInfo,
		TodaySponsor:            miningDetails.TodaySponsor,
		AccumulatedSalesAmount:  miningDetails.AccumulatedSalesAmount,
		TodaySponsorSalesAmount: miningDetails.TodaySponsorSalesAmount,
		MatchingLevel:           miningDetails.MatchingLevel,
		TotalSponsor:            miningDetails.TotalSponsor,
		TotalSponsorSalesAmount: miningDetails.TotalSponsorSalesAmount,
		PurchaseContractStatus:  1,
		IncomeCap:               miningDetails.IncomeCap,
		PoolAmount:              miningDetails.PoolAmount,
		TotalPoolAmount:         secPoolBal,
	}
	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrDataReturn)
}

// GetContractMiningActionHistoryListForm struct
type GetContractMiningActionHistoryListForm struct {
	Page     int64  `form:"page" json:"page"`
	DateFrom string `form:"date_from" json:"date_from"`
	DateTo   string `form:"date_to" json:"date_to"`
}

// GetContractMiningActionHistoryList function
func GetContractMiningActionHistoryList(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form GetContractMiningActionHistoryListForm
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
	arrData := reward_service.GetContractMiningActionHistoryListStruct{
		MemberID: member.EntMemberID,
		LangCode: langCode,
		Page:     form.Page,
		DateFrom: form.DateFrom,
		DateTo:   form.DateTo,
	}

	arrContractMiningActionHistoryList, errMsg := reward_service.GetContractMiningActionHistoryList(arrData)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: errMsg,
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrContractMiningActionHistoryList)
}

// GetContractMiningActionRankingListForm struct
type GetContractMiningActionRankingListForm struct {
	Page      int64  `form:"page" json:"page"`
	Date      string `form:"date" json:"date"`
	MaxNumber int    `form:"max_number" json:"max_number" valid:"Required"`
}

// GetContractMiningActionRankingList function
func GetContractMiningActionRankingList(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form GetContractMiningActionRankingListForm
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

	if form.Page < 1 {
		form.Page = 1
	}

	arrData := reward_service.GetContractMiningActionRankingListStruct{
		LangCode:  langCode,
		Page:      form.Page,
		Date:      form.Date,
		MaxNumber: form.MaxNumber,
	}

	arrContractMiningActionHistoryList, errMsg := reward_service.GetContractMiningActionRankingList(arrData)

	if errMsg != "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{
			Msg: errMsg,
		}, nil)
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, arrContractMiningActionHistoryList)
}

//func GetMemberMiningMiningActionDetailsv1
func GetMemberMiningMiningActionDetailsv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
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

	arrData := product_service.MemberMiningMiningActionDetailsv1ReqStruct{
		EntMemberID: member.EntMemberID,
		LangCode:    langCode,
	}
	arrDataReturn, err := product_service.GetMemberMiningMiningActionDetailsv1(arrData)

	if err != nil {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, arrDataReturn)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}

// GetContractMiningActionHistoryListForm struct
type GetMemberMiningMiningActionListv1Form struct {
	MiningCoinCode string `form:"mining_coin_code" json:"mining_coin_code" valid:"Required"`
	Page           int    `form:"page" json:"page"`
	// DateFrom       string `form:"date_from" json:"date_from"`
	// DateTo         string `form:"date_to" json:"date_to"`
}

//func GetMemberMiningMiningActionListv1
func GetMemberMiningMiningActionListv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form GetMemberMiningMiningActionListv1Form
	)

	langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	if c.GetHeader("Accept-Language") != "" {
		ok := models.ExistLangague(c.GetHeader("Accept-Language"))
		if ok {
			langCode = c.GetHeader("Accept-Language")
		}
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

	if form.Page < 1 {
		form.Page = 1
	}

	member := u.(*models.EntMemberMembers)

	arrData := product_service.MemberMiningMiningActionListv1ReqStruct{
		EntMemberID:    member.EntMemberID,
		MiningCoinCode: form.MiningCoinCode,
		Page:           form.Page,
		LangCode:       langCode,
	}
	arrDataReturn := product_service.GetMemberMiningMiningActionListv1(arrData)

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
	return
}
