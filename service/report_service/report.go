package report_service

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/service/member_service"
)

// func GetEntMemberReport
func GetEntMemberReport(langCode string, memID int) (interface{}, string) {
	var arrReturnData = []interface{}{}

	// get prd_group_type setting
	arrEntMemberReportFn := make([]models.WhereCondFn, 0)
	arrEntMemberReportFn = append(arrEntMemberReportFn, models.WhereCondFn{
		Condition: "ent_member_report.status = ?", CondValue: "A",
	})
	arrEntMemberReport, err := models.GetEntMemberReportFn(arrEntMemberReportFn, "", false)
	if err != nil {
		base.LogErrorLog("GetEntMemberReport:GetEntMemberReportFn()", map[string]interface{}{"condition": arrEntMemberReportFn}, err.Error(), true)
		return nil, "something_went_wrong"
	}

	for _, v := range arrEntMemberReport {
		var (
			reportName  string
			header      []EntMemberReportHeaderStruct
			filterParam []EntMemberReportFilterParamStruct
		)

		// translate report name
		if v.Name != "" {
			reportName = helpers.TranslateV2(v.Name, langCode, make(map[string]string))
		}

		if v.Header != "" {
			header, _ = GetEntMemberReportHeader(v.Header)
		}

		// translate header label
		for headerK, headerV := range header {
			if headerV.LabelName != "" {
				header[headerK].LabelName = helpers.TranslateV2(headerV.LabelName, langCode, make(map[string]string))
			}
		}

		if v.FilterParam != "" {
			filterParam, _ = GetEntMemberReportFilterParam(v.FilterParam)
		}

		// translate param name and select options
		for filterParamK, filterParamV := range filterParam {
			if filterParamV.Label != "" {
				filterParam[filterParamK].Label = helpers.TranslateV2(filterParamV.Label, langCode, make(map[string]string))
			}

			for filterParamOptionsK, filterParamOptionsV := range filterParamV.Options {
				filterParam[filterParamK].Options[filterParamOptionsK].Name = helpers.TranslateV2(filterParamOptionsV.Name, langCode, make(map[string]string))
			}
		}

		var arrReturnDataV = map[string]interface{}{
			"name":         reportName,
			"code":         v.Code,
			"header":       header,
			"filter_param": filterParam,
		}

		arrReturnData = append(arrReturnData, arrReturnDataV)
	}

	return arrReturnData, ""
}

// EntMemberReportFilterParamStruct struct
type EntMemberReportFilterParamStruct struct {
	Label     string                           `json:"label"`
	FieldName string                           `json:"field_name"`
	Type      string                           `json:"type"`
	Options   []FilterSelectOptionsParamStruct `json:"options"`
}

type FilterSelectOptionsParamStruct struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetEntMemberReportFilterParam func
func GetEntMemberReportFilterParam(rawEntMemberReportFilterParamInput string) ([]EntMemberReportFilterParamStruct, string) {
	memberReportFilterParamPointer := &[]EntMemberReportFilterParamStruct{}
	if rawEntMemberReportFilterParamInput == "" {
		return *memberReportFilterParamPointer, ""
	}

	err := json.Unmarshal([]byte(rawEntMemberReportFilterParamInput), memberReportFilterParamPointer)
	if err != nil {
		base.LogErrorLog("reportService:GetEntMemberReportFilterParam():Unmarshal():1", err.Error(), map[string]interface{}{"rawEntMemberReportFilterParamInput": rawEntMemberReportFilterParamInput}, true)
		return *memberReportFilterParamPointer, "something_went_wrong"
	}

	return *memberReportFilterParamPointer, ""
}

// EntMemberReportHeaderStruct struct
type EntMemberReportHeaderStruct struct {
	LabelName string `json:"label_name"`
	ParamName string `json:"param_name"`
	Show      struct {
		FilterParam string `json:"filter_param"`
		FilterValue string `json:"filter_value"`
	} `json:"show"`
}

// GetEntMemberReportHeader func
func GetEntMemberReportHeader(rawEntMemberReportHeaderInput string) ([]EntMemberReportHeaderStruct, string) {
	entMemberReportHeaderPointer := &[]EntMemberReportHeaderStruct{}
	if rawEntMemberReportHeaderInput == "" {
		return *entMemberReportHeaderPointer, ""
	}

	err := json.Unmarshal([]byte(rawEntMemberReportHeaderInput), entMemberReportHeaderPointer)
	if err != nil {
		base.LogErrorLog("reportService:GetEntMemberReportHeader():Unmarshal():1", err.Error(), map[string]interface{}{"rawEntMemberReportFilterParamInput": rawEntMemberReportHeaderInput}, true)
		return *entMemberReportHeaderPointer, "something_went_wrong"
	}

	return *entMemberReportHeaderPointer, ""
}

// GetReportListForm struct
type GetReportListForm struct {
	Code          string `json:"code" form:"code" valid:"Required"`
	Username      string `json:"username" form:"username"`
	Date          string `json:"date" form:"date"`
	DateFrom      string `json:"date_from" form:"date_from"`
	DateTo        string `json:"date_to" form:"date_to"`
	ReferralBonus string `json:"referral_bonus" form:"referral_bonus"`
	CreditType    string `json:"credit_type" form:"credit_type"`
	TransType     string `json:"trans_type" form:"trans_type"`
	MemberID      int
	LangCode      string
	Page          int `json:"page" form:"page"`
}

func ReportHandler(param GetReportListForm) interface{} {
	var (
		arrData       = []interface{}{}
		arrDataReturn = app.ArrDataResponseList{}
		arrSummary    = []interface{}{}
	)

	if param.Code == "SALES_INC_D_HISTORY" {
		arrData, arrSummary = param.GetDownlineSalesHistory()
	} else if param.Code == "WALLET_STATEMENT" {
		arrData = param.GetWalletStatement()
	} else if param.Code == "BONUS_HISTORY" {
		arrData, arrSummary = param.GetBonusHistory("")
	} else if param.Code == "SPA_BONUS_HISTORY" {
		arrData, arrSummary = param.GetBonusHistory("SPONSOR_A")
	} else if param.Code == "SPB_BONUS_HISTORY" {
		arrData, arrSummary = param.GetBonusHistory("SPONSOR_B")
	} else if param.Code == "RANKING_BONUS_HISTORY" {
		arrData, arrSummary = param.GetBonusHistory("RANKING")
	} else if param.Code == "STAKING_BONUS_HISTORY" {
		arrData, arrSummary = param.GetBonusHistory("STAKING")
	}

	// paginate arrData
	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	curPage := param.Page

	if curPage == 0 {
		curPage = 1
	}

	if param.Page != 0 {
		param.Page--
	}

	totalRecord := len(arrData)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(param.Page), int(limit), totalRecord)

	processArr := arrData[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(curPage),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      processArr,
		TableHeaderList:       nil,
		TableSummaryData:      arrSummary,
	}

	return arrDataReturn
}

func GetMemberFirstContractNo(memberID int) string {
	var docNo = ""

	// get members active normal contract sales (sort by asc)
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
		models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
	)
	arrSlsMaster, _ := models.GetSlsMasterAscFn(arrSlsMasterFn, "", false)
	if len(arrSlsMaster) > 0 {
		docNo = arrSlsMaster[0].DocNo
	}

	return docNo
}

func (param GetReportListForm) GetDownlineSalesHistory() ([]interface{}, []interface{}) {
	var arrReturnData = []interface{}{}

	// get member's lot info
	arrEntMemberLotSponsorFn := make([]models.WhereCondFn, 0)
	arrEntMemberLotSponsorFn = append(arrEntMemberLotSponsorFn,
		models.WhereCondFn{Condition: " member_id = ? ", CondValue: param.MemberID},
	)
	arrEntMemberLotSponsor, _ := models.GetEntMemberLotSponsorFn(arrEntMemberLotSponsorFn, false)

	var (
		iLft = 0
		iRgt = 0
		iLvl = 0
	)

	if len(arrEntMemberLotSponsor) > 0 {
		iLft = int(arrEntMemberLotSponsor[0].ILft)
		iRgt = int(arrEntMemberLotSponsor[0].IRgt)
		iLvl = int(arrEntMemberLotSponsor[0].Lvl)
	}

	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: "ent_member_lot_sponsor.i_lft > ? ", CondValue: iLft},
		models.WhereCondFn{Condition: "ent_member_lot_sponsor.i_rgt < ? ", CondValue: iRgt},
		models.WhereCondFn{Condition: "sls_master.status IN(?,'P') ", CondValue: "AP"},
		models.WhereCondFn{Condition: "sls_master.action IN(?) ", CondValue: "CONTRACT"},
	)

	if param.Username != "" {
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: "ent_member.nick_name = ? ", CondValue: param.Username},
		)
	}

	if param.DateFrom != "" {
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: "sls_master.doc_date >= ? ", CondValue: param.DateFrom},
		)
	}

	if param.DateTo != "" {
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: "sls_master.doc_date <= ? ", CondValue: param.DateTo},
		)
	}

	arrSlsMaster, err := models.GetSlsMasterByMemberLot(arrSlsMasterFn, false)
	if err != nil {
		base.LogErrorLog("GetDownlineSalesHistory():GetSlsMasterByMemberLot()", err.Error(), map[string]interface{}{"condition": arrSlsMasterFn}, true)
	}

	var (
		totalAmount   float64 = 0
		totalApAmount float64 = 0
	)

	for _, v := range arrSlsMaster {
		var (
			statusCode = v.StatusCode
			status     = v.Status
			lvl        = v.ILevel - iLvl
			actionDesc = v.Action
		)

		memberFirstContractNo := GetMemberFirstContractNo(v.MemberID)
		if memberFirstContractNo == v.DocNo {
			actionDesc = "purchase_package_b"
		} else {
			actionDesc = "topup_package_b"
		}

		// get member package tier
		packageTier, _ := member_service.GetMemberTierAtSpecificTime(v.MemberID, v.CreatedAt.Format("2006-01-02 15:04:05"))
		if packageTier == "" {
			packageTier = "-"
		}

		// get ap used
		var usedAP = ""
		arrEwtDetailFn := make([]models.WhereCondFn, 0)
		arrEwtDetailFn = append(arrEwtDetailFn,
			models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: v.BatchNo},
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "AP"},
		)
		arrEwtDetail, _ := models.GetEwtDetailWithSetup(arrEwtDetailFn, false)
		if len(arrEwtDetail) > 0 {
			totalApAmount += arrEwtDetail[0].TotalOut
			usedAP = helpers.CutOffDecimalv2(arrEwtDetail[0].TotalOut, 2, ".", ",", true)
		}
		var arrReturnDataV = map[string]interface{}{
			"id":               v.DocNo,
			"username":         v.Username,
			"doc_no":           v.DocNo,
			"doc_date":         v.DocDate.Format("2006-01-02"),
			"status_code":      statusCode,
			"status":           helpers.TranslateV2(status, param.LangCode, make(map[string]string)),
			"action":           v.Action,
			"action_desc":      helpers.TranslateV2(actionDesc, param.LangCode, make(map[string]string)),
			"prd_name":         helpers.TranslateV2(v.PrdName, param.LangCode, make(map[string]string)),
			"total_amount":     fmt.Sprintf("%s %s", helpers.CutOffDecimal(v.TotalAmount, 0, ".", ","), v.CurrencyCode),
			"currency_code":    v.CurrencyCode,
			"created_at":       v.CreatedAt.Format("2006-01-02 15:04:05"),
			"expired_at":       v.ExpiredAt.Format("2006-01-02 15:04:05"),
			"generation_level": lvl,
			"package_tier":     packageTier,
			"ap_used":          fmt.Sprintf("%sAP", usedAP),
		}

		totalAmount += v.TotalAmount

		arrReturnData = append(arrReturnData, arrReturnDataV)
	}

	var arrReturnDataHeader = []interface{}{}
	arrReturnDataHeader = append(arrReturnDataHeader,
		map[string]string{
			"label": helpers.TranslateV2("total_amount", param.LangCode, map[string]string{}),
			"value": helpers.CutOffDecimal(totalAmount, 2, ".", ",") + " " + helpers.TranslateV2("USDT", param.LangCode, map[string]string{}),
		},
		map[string]string{
			"label": helpers.TranslateV2("total_ap_amount", param.LangCode, map[string]string{}),
			"value": helpers.CutOffDecimal(totalApAmount, 2, ".", ",") + " " + helpers.TranslateV2("USDT", param.LangCode, map[string]string{}),
		})

	return arrReturnData, arrReturnDataHeader
}

func (param GetReportListForm) GetWalletStatement() []interface{} {
	var (
		decimalPoint    uint
		currencyCode    string
		arrReturnData   = []interface{}{}
		status          = helpers.Translate("completed", param.LangCode)
		statusColorCode = "#13B126"
		amountColorCode = "#13B126"
	)

	type arrStatementListSettingListStruct struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}

	arrEwtDetCond := make([]models.WhereCondFn, 0)
	arrEwtDetCond = append(arrEwtDetCond,
		models.WhereCondFn{Condition: "ewt_detail.member_id = ?", CondValue: param.MemberID},
	)
	if param.TransType != "" {
		if param.TransType == "STAKING_BNS" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: "#*bns_staking*#"},
			)
		} else if param.TransType == "SPONSOR_A_BNS" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: "#*bns_sponsor*#"},
			)
		} else if param.TransType == "SPONSOR_B_BNS" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: "#*bns_sponsor_annual*#"},
			)
		} else if param.TransType == "RANKING_BNS" {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: "BONUS"},
				models.WhereCondFn{Condition: "ewt_detail.doc_no = ?", CondValue: "#*bns_community*#"},
			)
		} else {
			arrEwtDetCond = append(arrEwtDetCond,
				models.WhereCondFn{Condition: "ewt_detail.transaction_type = ?", CondValue: strings.ToUpper(param.TransType)},
			)
		}
	}

	if param.DateFrom != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(ewt_detail.created_at) >= ?", CondValue: param.DateFrom},
		)
	}

	if param.DateTo != "" {
		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "date(ewt_detail.created_at) <= ?", CondValue: param.DateTo},
		)
	}

	if param.CreditType != "" {
		arrCondWal := make([]models.WhereCondFn, 0)
		arrCondWal = append(arrCondWal,
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: param.CreditType},
		)
		walrst, _ := models.GetEwtSetupFn(arrCondWal, "", false)

		arrEwtDetCond = append(arrEwtDetCond,
			models.WhereCondFn{Condition: "ewt_detail.ewallet_type_id = ?", CondValue: walrst.ID},
		)
	}

	EwtDet, _ := models.GetEwtDetailFn(arrEwtDetCond, false)

	if len(EwtDet) > 0 {
		for _, v := range EwtDet {
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ewt_setup.id = ?", CondValue: v.EwalletTypeID},
				models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
			)
			ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
			if ewtSetup != nil {
				decimalPoint = uint(ewtSetup.DecimalPoint)
				currencyCode = ewtSetup.CurrencyCode
			}

			remark := v.Remark

			if remark != "" {
				remark = "-" + helpers.TransRemark(v.Remark, param.LangCode)
			}

			amount := helpers.CutOffDecimal(float64(0), decimalPoint, ".", ",")

			if v.TotalIn > 0 {
				amount = helpers.CutOffDecimal(v.TotalIn, decimalPoint, ".", ",")
				amount = "+" + amount + " " + currencyCode
			}

			if v.TotalOut > 0 {
				amount = helpers.CutOffDecimal(v.TotalOut, decimalPoint, ".", ",")
				amount = "-" + amount + " " + currencyCode
				amountColorCode = "#F76464"
			}

			statusColorCode = "#13B126"
			status = helpers.Translate("completed", param.LangCode)

			transType := v.TransactionType
			transTypeTr := helpers.Translate(transType, param.LangCode)

			if v.TransactionType == "WITHDRAW" {
				withdrawDet, _ := models.GetEwtWithdrawDetailByDocNo(v.DocNo)

				if withdrawDet != nil {
					status = helpers.Translate(withdrawDet.StatusDesc, param.LangCode)
					if withdrawDet.Status == "AP" {
						status = helpers.Translate("completed", param.LangCode)
					}
					if withdrawDet.Status == "W" || withdrawDet.Status == "I" {
						status = helpers.Translate("pending", param.LangCode)
					}
					if withdrawDet.Status == "R" || withdrawDet.Status == "F" {
						statusColorCode = "#F76464"

						if v.TotalIn > 0 {
							statusColorCode = "#13B126"
							status = helpers.Translate("completed", param.LangCode)
						}
					} else if withdrawDet.Status == "P" || withdrawDet.Status == "W" {
						statusColorCode = "#FFA500"
					} else if withdrawDet.Status == "V" {
						statusColorCode = "#F76464"
					} else {
						statusColorCode = "#13B126"
					}
				}
			} else if v.TransactionType == "BONUS" {
				transTypeTr = helpers.TransRemark(v.DocNo, param.LangCode)
			} else if v.TransactionType == "NFT" {
				transTypeTr = helpers.Translate("nft_purchase", param.LangCode)
			}

			if param.TransType == "" {
				balance := helpers.CutOffDecimal(v.Balance, decimalPoint, ".", ",")
				arrReturnData = append(arrReturnData, map[string]interface{}{
					"id":                strconv.Itoa(v.ID),
					"trans_date":        v.TransDate.Format("2006-01-02 15:04:05"),
					"trans_type":        transTypeTr,
					"amount":            amount,
					"amount_color_code": amountColorCode,
					"balance":           balance,
					"status":            status,
					"status_color_code": statusColorCode,
				})
			} else {
				arrReturnData = append(arrReturnData, map[string]interface{}{
					"id":                strconv.Itoa(v.ID),
					"trans_date":        v.TransDate.Format("2006-01-02 15:04:05"),
					"trans_type":        transTypeTr,
					"amount":            amount,
					"amount_color_code": amountColorCode,
					"status":            status,
					"status_color_code": statusColorCode,
				})
			}

		}
	}

	return arrReturnData
}

func (param GetReportListForm) GetBonusHistory(rwdType string) ([]interface{}, []interface{}) {
	var (
		arrReturnData = []interface{}{}
		arrSummary    = []interface{}{}
		decimalPoint  uint
	)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "USDT"},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup != nil {
		decimalPoint = uint(ewtSetup.DecimalPoint)
	} else {
		decimalPoint = uint(2)
	}

	decimalPoint = uint(4)

	if param.DateFrom == "" && param.DateTo == "" {
		param.DateFrom = base.GetCurrentDateTimeT().AddDate(0, 0, -30).Format("2006-01-02")
		param.DateTo = base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02") //yesterday date
	}

	switch strings.ToUpper(rwdType) {
	case "RANKING": //community
		rst, err := models.GetTblQCommunityBonusByMemberId(param.MemberID, param.DateFrom, param.DateTo)

		if err != nil {
			base.LogErrorLog("GetBonusHistory - fail to get Community Bonus", err, param, true)
			return nil, nil
		}

		bnsTotal := float64(0)

		if len(rst) > 0 {
			for k, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				bnsTotal, _ = decimal.NewFromFloat(bnsTotal).Add(decimal.NewFromFloat(v.FBns)).Float64()

				arrReturnData = append(arrReturnData, map[string]string{
					"id":          strconv.Itoa(k + 1),
					"date":        v.TBnsID,
					"downline_id": v.DownlineId,
					"level":       v.ILvl,
					// "paid_level":  v.ILvlPaid,
					"sales_bv":   helpers.CutOffDecimal(v.FBv, 2, ".", ","),
					"percentage": helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					"bonus":      helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
					"created_at": v.DtCreated.Format("2006-01-02 15:04:05"),
				})

			}
		}

		arrSummary = append(arrSummary, map[string]string{
			"label": "bonus",
			"value": helpers.CutOffDecimal(bnsTotal, decimalPoint, ".", ",") + " " + helpers.Translate("USDT", param.LangCode),
		})

	case "SPONSOR_A":
		rst, err := models.GetSponsorBonusPassupByMemberId(param.MemberID, param.DateFrom, param.DateTo)

		if err != nil {
			base.LogErrorLog("GetBonusHistory - fail to get Sponsor Passup Bonus", err, param, true)
			return nil, nil
		}

		bnsTotal := float64(0)

		if len(rst) > 0 {
			for k, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				bnsTotal, _ = decimal.NewFromFloat(bnsTotal).Add(decimal.NewFromFloat(v.FBns)).Float64()

				arrReturnData = append(arrReturnData, map[string]string{
					"id":          strconv.Itoa(k + 1),
					"date":        v.TBnsID,
					"doc_no":      v.DocNo,
					"downline_id": v.DownlineID,
					"nodes":       helpers.CutOffDecimal(v.FBv, decimalPoint, ".", ","),
					"percentage":  helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					"bonus":       helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
					"created_at":  v.DtCreated.Format("2006-01-02 15:04:05"),
				})
			}
		}

		arrSummary = append(arrSummary, map[string]string{
			"label": "bonus",
			"value": helpers.CutOffDecimal(bnsTotal, decimalPoint, ".", ",") + " " + helpers.Translate("USDT", param.LangCode),
		})

	case "SPONSOR_B":
		rst, err := models.GetSponsorBonusPassupAnnualByMemberId(param.MemberID, param.DateFrom, param.DateTo)

		if err != nil {
			base.LogErrorLog("GetBonusHistory - fail to get Sponsor Passup Annual Bonus", err, param, true)
			return nil, nil
		}

		bnsTotal := float64(0)

		if len(rst) > 0 {
			for k, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				bnsTotal, _ = decimal.NewFromFloat(bnsTotal).Add(decimal.NewFromFloat(v.FBns)).Float64()

				arrReturnData = append(arrReturnData, map[string]string{
					"id":          strconv.Itoa(k + 1),
					"date":        v.TBnsID,
					"doc_no":      v.DocNo,
					"downline_id": v.DownlineID,
					"nodes":       helpers.CutOffDecimal(v.FBv, decimalPoint, ".", ","),
					"percentage":  helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					"bonus":       helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
					"created_at":  v.DtCreated.Format("2006-01-02 15:04:05"),
				})

			}
		}

		arrSummary = append(arrSummary, map[string]string{
			"label": "bonus",
			"value": helpers.CutOffDecimal(bnsTotal, decimalPoint, ".", ",") + " " + helpers.Translate("USDT", param.LangCode),
		})

	case "STAKING":
		rst, err := models.GetStakingBonusByMemberId(param.MemberID, param.DateFrom, param.DateTo)

		if err != nil {
			base.LogErrorLog("GetBonusHistory - fail to get Staking Bonus", err, param, true)
			return nil, nil
		}

		bnsTotal := float64(0)
		if len(rst) > 0 {
			for k, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				bnsTotal, _ = decimal.NewFromFloat(bnsTotal).Add(decimal.NewFromFloat(v.FBns)).Float64()

				arrReturnData = append(arrReturnData, map[string]string{
					"id":             strconv.Itoa(k + 1),
					"date":           v.TBnsID,
					"doc_no":         v.DocNo,
					"staking_date":   v.StakingDate.Format("2006-01-02 15:04:05"),
					"staking_value":  helpers.CutOffDecimal(v.StakingValue, decimalPoint, ".", ","),
					"staking_period": strconv.Itoa(v.StakingPeriod),
					// "bv":             helpers.CutOffDecimal(v.FBv, decimalPoint, ".", ","),
					"percentage": helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					"bonus":      helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
					"created_at": v.DtTimestamp.Format("2006-01-02 15:04:05"),
				})
			}
		}

		arrSummary = append(arrSummary, map[string]string{
			"label": "bonus",
			"value": helpers.CutOffDecimal(bnsTotal, decimalPoint, ".", ",") + " " + helpers.Translate("NFT", param.LangCode),
		})

	default:
		//get tbl_bonus
		rst, err := models.GetGroupedBnsIdRewardByMemId(param.MemberID, param.DateFrom, param.DateTo)

		if err != nil {
			base.LogErrorLog("GetBonusHistory - fail to get Total Bonus", err, param, true)
			return nil, nil
		}

		spn_a_total := float64(0)
		spn_b_total := float64(0)
		stak_total := float64(0)
		comm_total := float64(0)

		if len(rst) > 0 {
			for k, v := range rst {
				spn_a_total, _ = decimal.NewFromFloat(spn_a_total).Add(decimal.NewFromFloat(v.SponsorBns)).Float64()
				spn_b_total, _ = decimal.NewFromFloat(spn_b_total).Add(decimal.NewFromFloat(v.SponsorAnnualBns)).Float64()
				stak_total, _ = decimal.NewFromFloat(stak_total).Add(decimal.NewFromFloat(v.StakingBns)).Float64()
				comm_total, _ = decimal.NewFromFloat(comm_total).Add(decimal.NewFromFloat(v.CommunityBns)).Float64()

				arrReturnData = append(arrReturnData, map[string]string{
					"id":              strconv.Itoa(k + 1),
					"date":            v.TBnsId,
					"sponsor_bonus_a": helpers.CutOffDecimal(v.SponsorBns, decimalPoint, ".", ","),
					"sponsor_bonus_b": helpers.CutOffDecimal(v.SponsorAnnualBns, decimalPoint, ".", ","),
					"staking_bonus":   helpers.CutOffDecimal(v.StakingBns, decimalPoint, ".", ","),
					"community_bonus": helpers.CutOffDecimal(v.CommunityBns, decimalPoint, ".", ","),
				})
			}
		}

		arrSummary = append(arrSummary, map[string]string{
			"label": "sponsor_bonus_a",
			"value": helpers.CutOffDecimal(spn_a_total, decimalPoint, ".", ",") + " " + helpers.Translate("USDT", param.LangCode),
		})
		arrSummary = append(arrSummary, map[string]string{
			"label": "sponsor_bonus_b",
			"value": helpers.CutOffDecimal(spn_b_total, decimalPoint, ".", ",") + " " + helpers.Translate("USDT", param.LangCode),
		})
		arrSummary = append(arrSummary, map[string]string{
			"label": "staking_bonus",
			"value": helpers.CutOffDecimal(stak_total, decimalPoint, ".", ",") + " " + helpers.Translate("NFT", param.LangCode),
		})
		arrSummary = append(arrSummary, map[string]string{
			"label": "community_bonus",
			"value": helpers.CutOffDecimal(comm_total, decimalPoint, ".", ",") + " " + helpers.Translate("USDT", param.LangCode),
		})

	}

	return arrReturnData, arrSummary
}
