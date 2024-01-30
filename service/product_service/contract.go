package product_service

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/service/member_service"
	"github.com/smartblock/gta-api/service/sales_service"
	"github.com/smartblock/gta-api/service/wallet_service"
)

// PurchaseContractStruct struct
type PurchaseContractStruct struct {
	MemberID                                 int
	ContractCode, MachineType                string
	Unit, SecPrice, FilecoinPrice, ChiaPrice float64
	PaymentType, Payments                    string
	GenTranxDataStatus                       bool
}

// PurchaseContract func
func PurchaseContract(tx *gorm.DB, c PurchaseContractStruct, langCode string) (app.MsgStruct, map[string]string, string) {
	var (
		err                                            error
		errMsg                                         string
		batchNo                                        string
		batchDocType                                   string = "BT"
		docNo                                          string
		docType                                        string = "CT"
		memberID                                       int    = c.MemberID
		sponsorID                                      int
		prdMasterID                                    int
		prdCode                                        string = c.ContractCode
		prdGroup                                       string
		prdCurrencyCode                                string
		unit                                           float64 = c.Unit
		totalNv                                        float64
		unitPrice, payableAmt, tokenRate, exchangeRate float64
		action                                         string    = "CONTRACT"
		bnsAction                                      string    = "CONTRACT"
		grpType                                        string    = "0"
		curDate                                        string    = base.GetCurrentDateTimeT().Format("2006-01-02")
		curDateTime                                    time.Time = base.GetCurrentDateTimeT()
		approvableAt                                   time.Time
		expiredAt, _                                   = base.StrToDateTime("9999-01-01", "2006-01-02")
	)

	// validate contract code
	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.code = ? ", CondValue: prdCode},
		models.WhereCondFn{Condition: " date(prd_master.date_start) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " date(prd_master.date_end) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " prd_master.status = ? ", CondValue: "A"},
	)
	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:PurchaseContract():GetPrdMasterFn():1", map[string]interface{}{"condition": arrPrdMasterFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}
	if len(arrPrdMaster) <= 0 {
		return app.MsgStruct{Msg: "invalid_contract_code"}, nil, ""
	}

	prdMasterID = arrPrdMaster[0].ID
	prdCurrencyCode = arrPrdMaster[0].CurrencyCode
	unitPrice = arrPrdMaster[0].Amount
	prdGroup = arrPrdMaster[0].PrdGroup

	// validate prd_group_type
	arrPrdGroupTypeFn := make([]models.WhereCondFn, 0)
	arrPrdGroupTypeFn = append(arrPrdGroupTypeFn,
		models.WhereCondFn{Condition: " prd_group_type.code = ? ", CondValue: prdGroup},
		models.WhereCondFn{Condition: " prd_group_type.status = ? ", CondValue: "A"},
	)
	arrPrdGroupType, err := models.GetPrdGroupTypeFn(arrPrdGroupTypeFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:PurchaseContract():GetPrdGroupTypeFn():1", map[string]interface{}{"condition": arrPrdGroupTypeFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}
	if len(arrPrdGroupType) <= 0 {
		return app.MsgStruct{Msg: "invalid_contract_group_type"}, nil, ""
	}

	action = arrPrdGroupType[0].Code
	bnsAction = arrPrdGroupType[0].Code
	docType = arrPrdGroupType[0].DocType

	// validate if unit is positive
	if unit <= 0 {
		return app.MsgStruct{Msg: "please_enter_valid_amount"}, nil, ""
	}

	// get purchase contract setting
	if arrPrdMaster[0].PrdGroupSetting == "" {
		base.LogErrorLog("productService:PurchaseContract()", "product_group_setting_not_found", "", true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}

	arrPrdMasterSetup, errMsg := GetPrdMasterSetup(arrPrdMaster[0].Setting)
	if errMsg != "" {
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}

	arrPrdGroupTypeSetup, errMsg := GetPrdGroupTypeSetup(arrPrdMaster[0].PrdGroupSetting)
	if errMsg != "" {
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}

	var (
		keyinMin        = arrPrdGroupTypeSetup.KeyinMin
		keyinMultipleOf = arrPrdGroupTypeSetup.KeyinMultipleOf
		salesType       = arrPrdGroupTypeSetup.SalesType
		delayTime       = arrPrdGroupTypeSetup.DelayTime
		tiers           = arrPrdGroupTypeSetup.Tiers
		// exchangeSecPriceStatus = arrPrdGroupTypeSetup.ExchangeSecPriceStatus
	)

	// purchase amount must be more than or equal to keyinMin
	if unit < keyinMin {
		return app.MsgStruct{Msg: "minimum_purchase_unit_is_:0", Params: map[string]string{"0": helpers.CutOffDecimal(keyinMin, 2, ".", ",")}}, nil, ""
	}

	// purchased amount must be multiple of keyinMultipleOf
	if !helpers.IsMultipleOf(unit, keyinMultipleOf) {
		return app.MsgStruct{Msg: "purchase_unit_must_be_multiple_of_:0", Params: map[string]string{"0": helpers.CutOffDecimal(keyinMultipleOf, 2, ".", ",")}}, nil, ""
	}

	// if exchangeSecPriceStatus {
	// 	// get current exchange rate of blockchain coin
	// 	exchangePriceRate, errMsg := wallet_service.GetLatestExchangePriceMovementByEwtTypeCode("SEC")
	// 	if errMsg != "" {
	// 		return app.MsgStruct{Msg: errMsg}, nil, ""
	// 	}
	// 	if exchangePriceRate > 0 {
	// 		exchangeRate = exchangePriceRate
	// 	}
	// }

	// calculate payable amount
	payableAmt = unitPrice * unit

	// extraPaymentInfo := wallet_service.ExtraPaymentInfoStruct{
	// 	GenTranxDataStatus: c.GenTranxDataStatus,
	// 	EntMemberID:        memberID,
	// 	Module:             prdGroup,
	// }

	// map wallet payment structure format
	// paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStructv2(c.Payments, extraPaymentInfo)
	// if errMsg != "" {
	// 	return app.MsgStruct{Msg: errMsg}, nil, ""
	// }
	paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStruct(c.Payments)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil, ""
	}

	// get batch_no
	db := models.GetDB()
	batchNo, err = models.GetRunningDocNo(batchDocType, db) //get batch doc no
	if err != nil {
		base.LogErrorLog("productService:PurchaseContract():GetRunningDocNo():1", map[string]interface{}{"docType": batchDocType}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}
	err = models.UpdateRunningDocNo(batchDocType, db) //update batch doc no
	if err != nil {
		base.LogErrorLog("productService:PurchaseContract():UpdateRunningDocNo():1", map[string]interface{}{"docType": batchDocType}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}

	// validate payment with pay amount + deduct wallet
	msgStruct, arrData := wallet_service.PaymentProcess(tx, wallet_service.PaymentProcessStruct{
		MemberID:        memberID,
		PrdCurrencyCode: prdCurrencyCode,
		Module:          action,
		Type:            c.PaymentType,
		DocNo:           batchNo,
		Remark:          "#*batch no*#@" + batchNo,
		Amount:          payableAmt,
		Payments:        paymentStruct,
	}, 0, langCode)

	if msgStruct.Msg != "" {
		return msgStruct, nil, ""
	}

	// calculate approvable_at for contract queue
	approvableAt = base.GetCurrentDateTimeT().Add((delayTime * time.Minute))

	// get sponsor_id
	arrEntMemberTreeSponsorFn := make([]models.WhereCondFn, 0)
	arrEntMemberTreeSponsorFn = append(arrEntMemberTreeSponsorFn,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrEntMemberTreeSponsor, err := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberTreeSponsorFn, false)
	if err != nil {
		base.LogErrorLog("productService:PurchaseContract():GetEntMemberEntMemberTreeSponsorFn():1", map[string]interface{}{"condition": arrEntMemberTreeSponsorFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}
	sponsorID = arrEntMemberTreeSponsor.SponsorID

	if prdGroup == "CONTRACT" {
		// validate if got placement
		arrEntMemberSponsorFn := make([]models.WhereCondFn, 0)
		arrEntMemberSponsorFn = append(arrEntMemberSponsorFn,
			models.WhereCondFn{Condition: "ent_member_tree_sponsor.member_id = ?", CondValue: memberID},
		)
		arrEntMemberSponsor, _ := models.GetMemberSponsorFn(arrEntMemberSponsorFn, false)
		if arrEntMemberSponsor.UplineID == 0 {
			return app.MsgStruct{Msg: "please_set_placement_before_proceed_to_payment"}, nil, ""
		}

		// validate if got existing pending refund or not
		arrSlsMasterRefundBatchFn := make([]models.WhereCondFn, 0)
		arrSlsMasterRefundBatchFn = append(arrSlsMasterRefundBatchFn,
			models.WhereCondFn{Condition: "sls_master_refund_batch.member_id = ? ", CondValue: memberID},
			models.WhereCondFn{Condition: "sls_master_refund_batch.status != ?", CondValue: "R"}, // grab those pending/processing/completed request
		)
		arrSlsMasterRefundBatch, err := models.GetSlsMasterRefundBatchFn(arrSlsMasterRefundBatchFn, "", false)
		if err != nil {
			base.LogErrorLog("productService:PurchaseContract():GetSlsMasterRefundBatchFn():1", err.Error(), map[string]interface{}{"condition": arrSlsMasterRefundBatchFn}, true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
		}
		if len(arrSlsMasterRefundBatch) > 0 {
			return app.MsgStruct{Msg: "not_allowed_to_topup_when_having_active_refund_request"}, nil, ""
		}
	}

	if salesType == "SINGLE" {
		// generate in a single doc
		// tx passed into running doc need to be not affected by beginTransaction becoz blockchain_trans will still insert even if sales generate failed
		db := models.GetDB()
		docNo, err = models.GetRunningDocNo(docType, db) //get contract doc no
		if err != nil {
			base.LogErrorLog("productService:PurchaseContract():GetRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
		}
		err = models.UpdateRunningDocNo(docType, db) //update contract doc no
		if err != nil {
			base.LogErrorLog("productService:PurchaseContract():UpdateRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
		}

		var tier string = "B1"
		var workflow string = ""
		if prdGroup == "CONTRACT" {
			// validate if >= 50% of current purchase amount
			totalSalesFn := make([]models.WhereCondFn, 0)
			totalSalesFn = append(totalSalesFn,
				models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memberID},
				models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "CONTRACT"},
				models.WhereCondFn{Condition: "sls_master.status IN(?,'EP')", CondValue: "AP"},
			)
			totalSales, err := models.GetTotalSalesAmount(totalSalesFn, false)
			if err != nil {
				base.LogErrorLog("productService:PurchaseContract():GetTotalSalesAmount():1", map[string]interface{}{"condition": totalSalesFn}, err.Error(), true)
				return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
			}

			if totalSales.TotalBv > 0 {
				workflow = "TOPUP"
				var minTypeBPurchase = math.Ceil(float.Div(totalSales.TotalBv, 2)) // round up for float with decimal places

				if payableAmt < minTypeBPurchase {
					return app.MsgStruct{Msg: "minimum_topup_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(minTypeBPurchase, 0, ".", "", true)}}, nil, ""
				}
			} else {
				if payableAmt < 100 {
					return app.MsgStruct{Msg: "minimum_purchase_amount_is_:0", Params: map[string]string{"0": helpers.CutOffDecimalv2(100, 0, ".", "", true)}}, nil, ""
				}
			}

			// set expiry date to 365 days
			expiredAt = time.Now().AddDate(0, 0, 365)

			// previous owned package also update expiry date to 365 days later
			// if totalSales.TotalBv > 0 {
			// 	arrUpdCond := make([]models.WhereCondFn, 0)
			// 	arrUpdCond = append(arrUpdCond,
			// 		models.WhereCondFn{Condition: "member_id = ?", CondValue: memberID},
			// 		models.WhereCondFn{Condition: "action = ?", CondValue: "CONTRACT"},
			// 		models.WhereCondFn{Condition: "status IN(?,'EP')", CondValue: "AP"},
			// 	)
			// 	arrUpdateCols := map[string]interface{}{"status": "AP", "expired_at": expiredAt}
			// 	err = models.UpdatesFnTx(tx, "sls_master", arrUpdCond, arrUpdateCols, false)
			// 	if err != nil {
			// 		base.LogErrorLog("productService:PurchaseContract():UpdatesFnTx():1", err.Error(), map[string]interface{}{"condition": arrUpdCond, "updateCol": arrUpdateCols}, true)
			// 		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
			// 	}
			// }

			// find tiers
			for _, tierItem := range tiers {
				if payableAmt+totalSales.TotalAmount >= tierItem.Min {
					tier = tierItem.Tier
				}
			}
		}

		// calculate expired at for staking
		if prdGroup == "STAKING" {
			var (
				years, months, days int
			)

			if arrPrdMasterSetup.Years > 0 {
				years = arrPrdMasterSetup.Years
			}

			if arrPrdMasterSetup.Months > 0 {
				months = arrPrdMasterSetup.Months
			}

			if arrPrdMasterSetup.Days > 0 {
				days = arrPrdMasterSetup.Days
			}

			expiredAt = time.Now().AddDate(years, months, days)

		}

		// insert contract queue for pending contract
		var addSlsMasterParams = models.AddSlsMasterStruct{
			CountryID:   1,
			CompanyID:   1,
			MemberID:    memberID,
			SponsorID:   sponsorID,
			Status:      "AP",
			Action:      action,
			TotUnit:     unit,
			PriceRate:   unitPrice,
			PrdMasterID: prdMasterID,
			BatchNo:     batchNo,
			DocType:     docType,
			DocNo:       docNo,
			DocDate:     curDate,
			BnsBatch:    curDate,
			BnsAction:   bnsAction,
			TotalAmount: payableAmt,
			SubTotal:    payableAmt,
			// TotalPv:      payableAmt,
			TotalPv:      0.00, // quik requested to change
			TotalBv:      payableAmt,
			TotalSv:      payableAmt,
			TotalNv:      totalNv,
			TokenRate:    tokenRate,
			ExchangeRate: exchangeRate,
			Workflow:     workflow,
			CurrencyCode: prdCurrencyCode,
			GrpType:      grpType,
			CreatedAt:    curDateTime,
			CreatedBy:    fmt.Sprint(memberID),
			ApprovableAt: approvableAt,
			ApprovedAt:   time.Now(),
			ApprovedBy:   strconv.Itoa(memberID),
			ExpiredAt:    expiredAt,
		}

		_, err := models.AddSlsMaster(tx, addSlsMasterParams)
		if err != nil {
			base.LogErrorLog("productService:PurchaseContract():AddSlsMaster():1", map[string]interface{}{"param": addSlsMasterParams}, err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
		}

		if prdGroup == "CONTRACT" {
			// update previous record status to "I"
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: " member_id = ?", CondValue: memberID},
				models.WhereCondFn{Condition: " `force` != ?", CondValue: 1},
			)
			arrUpdateCols := map[string]interface{}{"status": "I"}
			err = models.UpdatesFnTx(tx, "sls_tier", arrUpdCond, arrUpdateCols, false)
			if err != nil {
				base.LogErrorLog("productService:PurchaseContract():UpdatesFnTx():1", err.Error(), map[string]interface{}{"condition": arrUpdCond, "updateCol": arrUpdateCols}, true)
				return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
			}

			// check if member got admin forced tier
			arrSlsTierFn := make([]models.WhereCondFn, 0)
			arrSlsTierFn = append(arrSlsTierFn,
				models.WhereCondFn{Condition: " sls_tier.member_id = ? ", CondValue: memberID},
				models.WhereCondFn{Condition: " sls_tier.status = ? ", CondValue: "A"},
				models.WhereCondFn{Condition: " sls_tier.force = ? ", CondValue: 1},
			)
			arrSlsTier, err := models.GetSlsTierFn(arrSlsTierFn, "", false)
			if err != nil {
				base.LogErrorLog("productService:PurchaseContract():GetMemberNftTier()", "GetSlsTierFn():1", err.Error(), true)
				return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
			}
			if len(arrSlsTier) <= 0 {
				// insert to sls_nft_tier
				var addSlsTierParams = models.SlsTier{
					MemberID:  memberID,
					Tier:      tier,
					Status:    "A",
					CreatedBy: docNo,
					CreatedAt: curDateTime,
				}

				_, err = models.AddSlsTier(tx, addSlsTierParams)
				if err != nil {
					base.LogErrorLog("productService:PurchaseContract():AddSlsTier():2", map[string]interface{}{"param": addSlsTierParams}, err.Error(), true)
					return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
				}
			}
		}

		// if prdGroup == "STAKING" {
		// 	for _, stakingWallet := range paymentStruct.MainWallet {
		// 		// insert to sls_nft
		// 		var addSlsNftParams = models.SlsNft{
		// 			SlsMasterID:   slsMaster.ID,
		// 			NftSeriesCode: stakingWallet.EwalletTypeCode,
		// 			Unit:          stakingWallet.Amount,
		// 			CreatedBy:     docNo,
		// 			CreatedAt:     time.Now(),
		// 		}

		// 		_, err := models.AddSlsNft(tx, addSlsNftParams)
		// 		if err != nil {
		// 			base.LogErrorLog("productService:PurchaseContract():AddSlsNft():2", map[string]interface{}{"param": addSlsNftParams}, err.Error(), true)
		// 			return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
		// 		}
		// 	}
		// }

	} else {
		// start to loop for x unit
		for x := 0.0; x < unit; x++ {
			// tx passed into running doc need to be not affected by beginTransaction becoz blockchain_trans will still insert even if sales generate failed
			db := models.GetDB()
			docNo, err = models.GetRunningDocNo(docType, db) //get contract doc no
			if err != nil {
				base.LogErrorLog("productService:PurchaseContract():GetRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
				return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
			}
			err = models.UpdateRunningDocNo(docType, db) //update contract doc no
			if err != nil {
				base.LogErrorLog("productService:PurchaseContract():UpdateRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
				return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
			}

			// insert contract queue for pending contract
			var addSlsMasterParams = models.AddSlsMasterStruct{
				CountryID:    1,
				CompanyID:    1,
				MemberID:     memberID,
				SponsorID:    sponsorID,
				Status:       "AP",
				Action:       action,
				TotUnit:      1, // doc is generated one by one in batch
				PriceRate:    unitPrice,
				PrdMasterID:  prdMasterID,
				BatchNo:      batchNo,
				DocType:      docType,
				DocNo:        docNo,
				DocDate:      curDate,
				BnsBatch:     curDate,
				BnsAction:    bnsAction,
				TotalAmount:  unitPrice,
				SubTotal:     unitPrice,
				TotalPv:      unitPrice,
				TotalBv:      unitPrice,
				TotalSv:      0.00,
				TotalNv:      totalNv,
				TokenRate:    tokenRate,
				ExchangeRate: exchangeRate,
				CurrencyCode: prdCurrencyCode,
				GrpType:      grpType,
				CreatedAt:    curDateTime,
				CreatedBy:    fmt.Sprint(memberID),
				ApprovableAt: approvableAt,
				ApprovedAt:   time.Now(),
				ApprovedBy:   strconv.Itoa(memberID),
			}

			_, err := models.AddSlsMaster(tx, addSlsMasterParams)
			if err != nil {
				base.LogErrorLog("productService:PurchaseContract():AddSlsMaster():2", map[string]interface{}{"param": addSlsMasterParams}, err.Error(), true)
				return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
			}
		}
	}

	// insert to sls_master_batch
	var addSlsMasterBatchParams = models.AddSlsMasterBatchStruct{
		BatchNo:   batchNo,
		Quantity:  unit,
		CreatedBy: fmt.Sprint(memberID),
	}

	_, err = models.AddSlsMasterBatch(tx, addSlsMasterBatchParams)
	if err != nil {
		base.LogErrorLog("productService:PurchaseContract():AddSlsMasterBatch():1", map[string]interface{}{"param": addSlsMasterBatchParams}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil, ""
	}

	return app.MsgStruct{Msg: ""}, arrData, docNo
}

// PostStakingStruct struct
type PostStakingStruct struct {
	GenTranxDataStatus      bool
	MemberID                int
	ProductCode             string
	Unit                    float64
	Payments                string
	ApprovedTransactionData string
}

// PostStaking func
func PostStaking(tx *gorm.DB, stk PostStakingStruct, langCode string) (app.MsgStruct, map[string]string) {
	var (
		err                   error
		errMsg                string
		docType               string = "STK"
		prdCode               string = stk.ProductCode
		prdMasterID           int
		unitPrice, payableAmt float64
	)

	// validate contract code
	curDate := base.GetCurrentDateTimeT().Format("2006-01-02")

	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.code = ? ", CondValue: prdCode},
		models.WhereCondFn{Condition: " date(prd_master.date_start) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " date(prd_master.date_end) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " prd_master.status = ? ", CondValue: "A"},
	)
	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)

	if err != nil {
		base.LogErrorLog("productService:PostStaking()", "GetPrdMasterFn():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	if len(arrPrdMaster) <= 0 {
		return app.MsgStruct{Msg: "invalid_product_code"}, nil
	}

	prdMasterID = arrPrdMaster[0].ID
	unitPrice = arrPrdMaster[0].Amount

	// validate if unit is positive
	unit := stk.Unit
	if unit <= 0 {
		return app.MsgStruct{Msg: "please_enter_valid_amount"}, nil
	}

	// get purchase contract setting
	postStakingSetting, errMsg := GetPostStakingSetup()
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// purchased amount must be multiple of x
	// ok, err := helpers.IsMultipleOf(unit, postStakingSetting.MultipleOf)
	// if err != nil {
	// 	base.LogErrorLog("productService:PostStaking()", "IsMultipleOf():1", err.Error(), true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}
	// }

	if !helpers.IsMultipleOf(unit, postStakingSetting.MultipleOf) {
		return app.MsgStruct{Msg: "amount_must_be_multiple_of_:0", Params: map[string]string{"0": helpers.CutOffDecimal(postStakingSetting.MultipleOf, 8, ".", ",")}}, nil
	}

	// calculate payable amount
	payableAmt = unitPrice * unit // in liga/sec

	// map wallet payment structure format
	paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStruct(stk.Payments)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// tx passed into running doc need to be not affected by beginTransaction becoz blockchain_trans will still insert even if sales generate failed
	db := models.GetDB()
	docNo, err := models.GetRunningDocNo(docType, db) //get transfer doc no
	if err != nil {
		base.LogErrorLog("productService:PostStaking()", "GetRunningDocNo():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	err = models.UpdateRunningDocNo(docType, db) //update transfer doc no
	if err != nil {
		base.LogErrorLog("productService:PostStaking()", "UpdateRunningDocNo():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// calculate approvable_at for contract queue
	approvableAt := base.GetCurrentDateTimeT().Add((postStakingSetting.DelayTime * time.Minute))

	// get sponsor_id
	arrEntMemberTreeSponsorFn := make([]models.WhereCondFn, 0)
	arrEntMemberTreeSponsorFn = append(arrEntMemberTreeSponsorFn,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: stk.MemberID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)

	arrEntMemberTreeSponsor, err := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberTreeSponsorFn, false)
	if err != nil {
		base.LogErrorLog("productService:PostStaking()", "GetEntMemberEntMemberTreeSponsorFn():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	var sponsorID = arrEntMemberTreeSponsor.SponsorID

	// get currency code of blockchain coin
	arrEwtSetupFn := make([]models.WhereCondFn, 0)
	arrEwtSetupFn = append(arrEwtSetupFn,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: prdCode},
	)
	arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:PostStaking()", "GetEwtSetupFn():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if arrEwtSetup == nil {
		base.LogErrorLog("productService:PostStaking()", "ewallet_type_code_"+prdCode+"_not_found", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// get current token rate of blockchain coin
	// conversionRate, errMsg := wallet_service.GetLatestPriceMovementByEwtTypeCode(prdCode)
	// if errMsg != "" {
	// 	return app.MsgStruct{Msg: errMsg}, nil
	// }

	// get current exchange rate of blockchain coin
	exchangePriceRate, errMsg := wallet_service.GetLatestExchangePriceMovementByEwtTypeCode(prdCode)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// validate converted amount
	// convertedUsdsAmt, err := helpers.ValueToFloat(helpers.CutOffDecimal(payableAmt*conversionRate, 8, ".", "")) // convert to usds price
	// if err != nil {
	// 	base.LogErrorLog("walletService:PostStaking()", "ValueToFloat():1", err.Error(), true)
	// 	return app.MsgStruct{Msg: "something_went_wrong"}, nil
	// }

	var conversionRate = 1.0

	// insert contract queue for pending contract
	_, err = models.AddSlsMaster(tx, models.AddSlsMasterStruct{
		CountryID:    1,
		CompanyID:    1,
		MemberID:     stk.MemberID,
		SponsorID:    sponsorID,
		Status:       "P",
		Action:       "STAKING",
		TotUnit:      payableAmt,
		PriceRate:    unitPrice,
		PrdMasterID:  prdMasterID,
		DocType:      docType,
		DocNo:        docNo,
		DocDate:      curDate,
		BnsBatch:     curDate,
		BnsAction:    "STAKING",
		TotalAmount:  payableAmt,
		SubTotal:     payableAmt,
		CurrencyCode: arrEwtSetup.CurrencyCode,
		TokenRate:    conversionRate,
		ExchangeRate: exchangePriceRate,
		TotalPv:      payableAmt,
		TotalBv:      payableAmt,
		TotalSv:      0.00,
		TotalNv:      0.00,
		GrpType:      "0",
		CreatedBy:    fmt.Sprint(stk.MemberID),
		ApprovableAt: approvableAt,
	})
	if err != nil {
		base.LogErrorLog("productService:PostStaking()", "AddSlsMaster():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// for liga staking required to save approved transaction first
	if prdCode == "LIGA" {
		if stk.ApprovedTransactionData == "" {
			return app.MsgStruct{Msg: "approved_transaction_data_cannot_be_empty"}, nil
		}

		approvedTransactionData, errMsg := wallet_service.ConvertTransactionDataToStruct(stk.ApprovedTransactionData)
		if errMsg != "" {
			return app.MsgStruct{Msg: errMsg}, nil
		}

		if payableAmt != approvedTransactionData.ConvertedAmount {
			base.LogErrorLog("walletService:PostStaking()", "invalid_converted_amount", stk, true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil
		}

		// call blockchain transaction api
		hashValue, errMsg := wallet_service.SignedTransaction(approvedTransactionData.TransactionData)
		if errMsg != "" {
			return app.MsgStruct{Msg: errMsg}, nil
		}

		// log blochchain transaction
		db := models.GetDB() // will save record even if other process after this failed
		_, err = models.AddBlockchainTrans(db, models.AddBlockchainTransStruct{
			MemberID:          stk.MemberID,
			EwalletTypeID:     arrEwtSetup.ID,
			DocNo:             docNo,
			Status:            "P",
			TransactionType:   "STAKING-APPROVED",
			TotalOut:          payableAmt,
			ConversionRate:    conversionRate,
			ConvertedTotalOut: payableAmt,
			TransactionData:   approvedTransactionData.TransactionData,
			HashValue:         hashValue,
			LogOnly:           1,
		})
		if err != nil {
			base.LogErrorLog("walletService:PostStaking()", "AddBlockchainTrans():1", err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil
		}
	}

	// validate payment with pay amount + deduct wallet
	msgStruct, arrData := wallet_service.PaymentProcess(tx, wallet_service.PaymentProcessStruct{
		MemberID: stk.MemberID,
		Module:   "STAKING",
		Type:     stk.ProductCode, // use product code as payment type
		DocNo:    docNo,
		// Remark:   "#*staking*#",
		Amount:   payableAmt, // pay in liga/sec price
		Payments: paymentStruct,
	}, 1, langCode)

	if msgStruct.Msg != "" {
		return msgStruct, nil
	}

	return app.MsgStruct{Msg: ""}, arrData
}

// PostUnstakeStruct struct
type PostUnstakeStruct struct {
	MemberID               int
	DocNo                  string
	Amount                 float64
	UnstakeTransactionData string
}

// PostUnstake func
func PostUnstake(tx *gorm.DB, stk PostUnstakeStruct, langCode string) (app.MsgStruct, map[string]string) {
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: stk.MemberID},
		models.WhereCondFn{Condition: "sls_master.doc_no = ? ", CondValue: stk.DocNo},
	)
	arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("salesService:PostUnstake()", "GetSlsMasterFn():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrSlsMaster) <= 0 {
		return app.MsgStruct{Msg: "doc_no_not_found"}, nil
	}

	// get doc's product
	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.id = ? ", CondValue: arrSlsMaster[0].PrdMasterID},
	)
	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("salesService:PostUnstake()", "GetPrdMasterFn():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrPrdMaster) <= 0 {
		return app.MsgStruct{Msg: "doc_product_not_found"}, nil
	}

	prdCode := arrPrdMaster[0].Code

	// get current token rate of blockchain coin
	// conversionRate, errMsg := wallet_service.GetLatestPriceMovementByEwtTypeCode(prdCode)
	// if errMsg != "" {
	// 	return app.MsgStruct{Msg: errMsg}, nil
	// }
	conversionRate := 1.0

	// get converted amount
	// convertedAmount, err := helpers.ValueToFloat(helpers.CutOffDecimal(stk.Amount/conversionRate, 8, ".", ""))
	refundAmount, err := helpers.ValueToFloat(helpers.CutOffDecimal(stk.Amount, 8, ".", ""))
	if err != nil {
		base.LogErrorLog("walletService:PostUnstake()", "ValueToFloat():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	errMsg, _ := sales_service.RefundSales(tx, sales_service.RefundSalesStruct{
		MemberID:      stk.MemberID,
		DocNo:         stk.DocNo,
		RequestAmount: refundAmount,
	})
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// prdCode := arrData["prdCode"]

	// get currency code of blockchain coin
	arrEwtSetupFn := make([]models.WhereCondFn, 0)
	arrEwtSetupFn = append(arrEwtSetupFn,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: prdCode},
	)
	arrEwtSetup, err := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:PostUnstake()", "GetEwtSetupFn():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if arrEwtSetup == nil {
		base.LogErrorLog("productService:PostUnstake()", "ewallet_type_code_"+prdCode+"_not_found", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	unstakeTransactionData, errMsg := wallet_service.ConvertTransactionDataToStruct(stk.UnstakeTransactionData)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// validate converted amount
	if refundAmount != unstakeTransactionData.ConvertedAmount {
		base.LogErrorLog("walletService:PostUnstake()", "invalid_converted_amount", stk, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// call unstake method sign transaction api
	hashValue, errMsg := wallet_service.SignedTransaction(unstakeTransactionData.TransactionData)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}
	// hashValue := "test"

	// log blochchain transaction
	db := models.GetDB() // will save record even if other process after this failed
	_, err = models.AddBlockchainTrans(db, models.AddBlockchainTransStruct{
		MemberID:         stk.MemberID,
		EwalletTypeID:    arrEwtSetup.ID,
		DocNo:            stk.DocNo,
		Status:           "P",
		TransactionType:  "UNSTAKE",
		TotalIn:          refundAmount,
		ConversionRate:   conversionRate,
		ConvertedTotalIn: refundAmount,
		TransactionData:  unstakeTransactionData.TransactionData,
		HashValue:        hashValue,
		LogOnly:          1, // refund bal will not affect instantly on available balance
	})
	if err != nil {
		base.LogErrorLog("walletService:PostUnstake()", "AddBlockchainTrans():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	arrReturnData := make(map[string]string)
	arrReturnData["hash_value"] = hashValue
	arrReturnData["unstake_amount"] = helpers.CutOffDecimal(refundAmount, 8, ".", ",")
	return app.MsgStruct{Msg: ""}, arrReturnData
}

func CalTotalAvailableExchangedUsdtAmt(memID int, docCreatedAt string) (float64, error) {
	var (
		totalExchangedUsdtAmount       float64
		totalSpentUsdtAmt              float64
		totalAvailableExchangedUsdtAmt float64
	)
	// get member's total exchanged usdt amount
	arrMemberTotalExchangedUsdtAmount, err := models.GetMemberTotalExchangedUsdtAmount(memID, docCreatedAt, false)
	if err != nil {
		return 0, err
	}

	totalExchangedUsdtAmount = arrMemberTotalExchangedUsdtAmount.TotalAmount

	// get total spent amount from sls_master
	arrMemberTotalSpentUsdtAmount, err := models.GetMemberTotalSpentUsdtAmount(memID)
	if err != nil {
		return 0, err
	}

	totalSpentUsdtAmt = arrMemberTotalSpentUsdtAmount.TotalAmount

	// handle if available exchange amount is negative, will set to 0 instead
	totalAvailableExchangedUsdtAmt = totalExchangedUsdtAmount - totalSpentUsdtAmt

	if totalAvailableExchangedUsdtAmt < 0 {
		totalAvailableExchangedUsdtAmt = 0
	}

	return totalAvailableExchangedUsdtAmt, nil
}

// TopupContractStruct struct
type TopupContractStruct struct {
	MemberID           int
	GenTranxDataStatus bool
	DocNo              string
	Amount             float64
	PaymentType        string
	Payments           string
}

// TopupContract func
func TopupContract(tx *gorm.DB, c TopupContractStruct, langCode string) (app.MsgStruct, map[string]string) {
	var (
		err             error
		errMsg          string
		docNo           string
		docType         string = "TP"
		curDocNo        string = c.DocNo
		memID           int    = c.MemberID
		paymentModule   string = "CONTRACT_TOPUP"
		curDate         string = base.GetCurrentDateTimeT().Format("2006-01-02")
		slsMasterID     int
		slsMasterStatus string
		slsPrdMasterID  int
		slsAction       string
		topupAmount     float64 = c.Amount
		leverage        float64
	)

	// validate doc no
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: memID},
		models.WhereCondFn{Condition: " sls_master.doc_no = ? ", CondValue: curDocNo},
	)
	arrSlsMaster, _ := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if len(arrSlsMaster) <= 0 {
		return app.MsgStruct{Msg: "invalid_doc_no"}, nil
	}

	slsMasterID = arrSlsMaster[0].ID
	slsMasterStatus = arrSlsMaster[0].Status
	slsPrdMasterID = arrSlsMaster[0].PrdMasterID
	slsAction = arrSlsMaster[0].Action

	// validate doc status
	if slsMasterStatus != "AP" {
		if slsMasterStatus == "P" {
			return app.MsgStruct{Msg: "doc_status_is_still_pending"}, nil
		} else {
			return app.MsgStruct{Msg: "invalid_doc_status"}, nil
		}
	}

	// validate if doc product can be topup
	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.id = ? ", CondValue: slsPrdMasterID},
	)
	arrPrdMaster, _ := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if len(arrPrdMaster) <= 0 {
		return app.MsgStruct{Msg: "invalid_prd_master_id"}, nil
	}
	var prdCurrencyCode = arrPrdMaster[0].CurrencyCode

	// validate topup setting (topup amount multiple of)
	if arrPrdMaster[0].TopupSetting == "" {
		return app.MsgStruct{Msg: "doc_type_not_allowed_to_topup"}, nil
	}

	productTopupSetting, errMsg := GetProductTopupSetting(arrPrdMaster[0].TopupSetting)
	if errMsg != "" {
		base.LogErrorLog("productService:TopupContract()", "GetProductTopupSetting():1", errMsg, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	if !productTopupSetting.Status {
		return app.MsgStruct{Msg: "doc_type_not_allowed_to_topup"}, nil
	}

	// topup amount must be multiple of x
	if !helpers.IsMultipleOf(topupAmount, productTopupSetting.MultipleOf) {
		return app.MsgStruct{Msg: "amount_must_be_multiple_of_:0", Params: map[string]string{"0": helpers.CutOffDecimal(productTopupSetting.MultipleOf, 2, ".", ",")}}, nil
	}

	extraPaymentInfo := wallet_service.ExtraPaymentInfoStruct{
		GenTranxDataStatus: c.GenTranxDataStatus,
		EntMemberID:        memID,
		Module:             paymentModule,
	}

	// map wallet payment structure format
	paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStructv2(c.Payments, extraPaymentInfo)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// tx passed into running doc need to be not affected by beginTransaction becoz blockchain_trans will still insert even if sales generate failed
	db := models.GetDB()
	docNo, err = models.GetRunningDocNo(docType, db) //get transfer doc no
	if err != nil {
		base.LogErrorLog("productService:TopupContract()", "GetRunningDocNo():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	err = models.UpdateRunningDocNo(docType, db) //update transfer doc no
	if err != nil {
		base.LogErrorLog("productService:TopupContract()", "UpdateRunningDocNo():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// calculate leverage
	// calculate income cap for contract topup
	arrMemberHighestPackageInfo, err := models.GetMemberHighestPackageInfo(memID, slsAction, docNo)
	if err != nil {
		base.LogErrorLog("productService:TopupContract():GetMemberHighestPackageInfo():1", err.Error(), map[string]interface{}{"memID": memID, "action": slsAction, "docNo": docNo}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	if arrMemberHighestPackageInfo != nil {
		leverage = arrMemberHighestPackageInfo.Leverage
	}

	// insert to income cap wallet if leverage > 0
	if leverage > 0 {
		incomeCapSetting, errMsg := sales_service.MapProductIncomeCapSetting(arrPrdMaster[0].IncomeCapSetting)
		if errMsg != "" {
			base.LogErrorLog("productService:TopupContract():MapProductIncomeCapSetting():1", err.Error(), map[string]interface{}{"income_cap_setting": arrPrdMaster[0].IncomeCapSetting}, true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil
		}

		if incomeCapSetting.Status {
			var (
				incomeCap            = float.Mul(topupAmount, leverage)
				saveMemberWalletData = wallet_service.SaveMemberWalletStruct{
					EntMemberID:     memID,
					EwalletTypeID:   incomeCapSetting.EwalletTypeID,
					TotalIn:         incomeCap,
					TransactionType: "INCOME_CAP",
					DocNo:           docNo,
					Remark:          fmt.Sprintf("#*income_cap*# %g @ %g", float.RoundUp(topupAmount, 0), leverage),
					CreatedBy:       "AUTO",
				}
			)
			_, err := wallet_service.SaveMemberWallet(tx, saveMemberWalletData)

			if err != nil {
				base.LogErrorLog("productService:TopupContract():SaveMemberWallet():1", err.Error(), map[string]interface{}{"saveMemberWalletData": saveMemberWalletData}, true)
				return app.MsgStruct{Msg: "something_went_wrong"}, nil
			}
		}
	}

	// save sls_master_topup
	_, err = models.AddSlsMasterTopup(tx, models.AddSlsMasterTopupStruct{
		SlsMasterID: slsMasterID,
		MemberID:    memID,
		DocNo:       docNo,
		DocDate:     curDate,
		BnsBatch:    curDate,
		Status:      "AP",
		TotalAmount: topupAmount,
		TotalBv:     topupAmount,
		CreatedBy:   fmt.Sprint(memID),
		Leverage:    leverage,
		ApprovedAt:  time.Now(),
		ApprovedBy:  fmt.Sprint(memID),
	})
	if err != nil {
		base.LogErrorLog("productService:TopupContract()", "AddSlsMasterTopup():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// validate payment with pay amount + deduct wallet
	msgStruct, arrData := wallet_service.PaymentProcess(tx, wallet_service.PaymentProcessStruct{
		MemberID:        memID,
		PrdCurrencyCode: prdCurrencyCode,
		Module:          paymentModule,
		Type:            c.PaymentType,
		DocNo:           docNo,
		Amount:          topupAmount,
		Payments:        paymentStruct,
	}, 0, langCode)

	if msgStruct.Msg != "" {
		return msgStruct, nil
	}

	return app.MsgStruct{Msg: ""}, arrData
}

// TopupMiningNodeStruct struct
type TopupMiningNodeStruct struct {
	MemberID           int
	NodeID             int
	ContractCode       string
	PaymentType        string
	Payments           string
	GenTranxDataStatus bool
}

// TopupMiningNode func
func TopupMiningNode(tx *gorm.DB, topupMiningNodeInput TopupMiningNodeStruct, langCode string) (app.MsgStruct, map[string]string) {
	var (
		err                                        error
		errMsg                                     string
		docNo                                      string
		docType                                    string
		nodeID                                     int    = topupMiningNodeInput.NodeID
		memID                                      int    = topupMiningNodeInput.MemberID
		prdCode                                    string = topupMiningNodeInput.ContractCode
		paymentModule                              string
		curDate                                    string = base.GetCurrentDateTimeT().Format("2006-01-02")
		prdMasterID                                int
		prdCurrencyCode, prdGroup, slsMasterStatus string
		unitPrice                                  float64
	)

	// validate contract code
	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.code = ? ", CondValue: prdCode},
		models.WhereCondFn{Condition: " date(prd_master.date_start) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " date(prd_master.date_end) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " prd_master.status = ? ", CondValue: "A"},
	)
	arrPrdMaster, err := models.GetPrdMasterFn(arrPrdMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:TopupMiningNode():GetPrdMasterFn():1", map[string]interface{}{"condition": arrPrdMasterFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrPrdMaster) <= 0 {
		return app.MsgStruct{Msg: "invalid_contract_code"}, nil
	}

	prdMasterID = arrPrdMaster[0].ID
	prdGroup = arrPrdMaster[0].PrdGroup
	prdCurrencyCode = arrPrdMaster[0].CurrencyCode
	unitPrice = arrPrdMaster[0].Amount

	// validate prd_group_type
	arrPrdGroupTypeFn := make([]models.WhereCondFn, 0)
	arrPrdGroupTypeFn = append(arrPrdGroupTypeFn,
		models.WhereCondFn{Condition: " prd_group_type.code = ? ", CondValue: prdGroup},
		models.WhereCondFn{Condition: " prd_group_type.status = ? ", CondValue: "A"},
	)
	arrPrdGroupType, err := models.GetPrdGroupTypeFn(arrPrdGroupTypeFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:TopupMiningNode():GetPrdGroupTypeFn():1", map[string]interface{}{"condition": arrPrdGroupTypeFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrPrdGroupType) <= 0 {
		return app.MsgStruct{Msg: "invalid_contract_group_type"}, nil
	}

	paymentModule = arrPrdGroupType[0].Code
	docType = arrPrdGroupType[0].DocType

	// validate node id + if node belong to bzz_100
	arrSlsMasterMiningNodeFn := make([]models.WhereCondFn, 0)
	arrSlsMasterMiningNodeFn = append(arrSlsMasterMiningNodeFn,
		models.WhereCondFn{Condition: " sls_master.member_id = ? ", CondValue: memID},
		models.WhereCondFn{Condition: " sls_master_mining.machine_type = ? ", CondValue: "BZZ_100"},
		models.WhereCondFn{Condition: " sls_master_mining_node.id = ? ", CondValue: nodeID},
	)
	arrSlsMasterMiningNode, _ := models.GetSlsMasterMiningNodeFn(arrSlsMasterMiningNodeFn, "", false)

	if len(arrSlsMasterMiningNode) <= 0 {
		return app.MsgStruct{Msg: "invalid_node_id"}, nil
	}

	slsMasterStatus = arrSlsMasterMiningNode[0].Status

	// validate doc status
	if slsMasterStatus != "AP" {
		if slsMasterStatus == "P" {
			return app.MsgStruct{Msg: "doc_status_is_still_pending"}, nil
		} else {
			return app.MsgStruct{Msg: "invalid_doc_status"}, nil
		}
	}

	extraPaymentInfo := wallet_service.ExtraPaymentInfoStruct{
		GenTranxDataStatus: topupMiningNodeInput.GenTranxDataStatus,
		EntMemberID:        memID,
		Module:             paymentModule,
	}

	// map wallet payment structure format
	paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStructv2(topupMiningNodeInput.Payments, extraPaymentInfo)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	var (
		year             int = 0
		month            int = 0
		day              int = 0
		startDate            = arrSlsMasterMiningNode[0].StartDate
		endDate              = arrSlsMasterMiningNode[0].EndDate
		broadbandSetting     = arrPrdMaster[0].BroadbandSetting
		curDateTime          = base.GetCurrentDateTimeT()
	)

	if broadbandSetting != "" {
		// update sls_master_mining_node.end_date
		type ArrBroadbandSettingStruct struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Day   int `json:"day"`
		}

		var arrBrodbandSetting ArrBroadbandSettingStruct
		json.Unmarshal([]byte(arrPrdMaster[0].BroadbandSetting), &arrBrodbandSetting)

		year = arrBrodbandSetting.Year
		month = arrBrodbandSetting.Month
		day = arrBrodbandSetting.Day
	}

	// tx passed into running doc need to be not affected by beginTransaction becoz blockchain_trans will still insert even if sales generate failed
	db := models.GetDB()
	docNo, err = models.GetRunningDocNo(docType, db) //get transfer doc no
	if err != nil {
		base.LogErrorLog("productService:TopupMiningNode()", "GetRunningDocNo():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	err = models.UpdateRunningDocNo(docType, db) //update transfer doc no
	if err != nil {
		base.LogErrorLog("productService:TopupMiningNode()", "UpdateRunningDocNo():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// save sls_master_mining_node_topup
	arrAddSlsMasterMiningNode := models.AddSlsMasterMiningNodeTopupStruct{
		SlsMasterMiningNodeID: nodeID,
		MemberID:              memID,
		PrdMasterID:           prdMasterID,
		DocNo:                 docNo,
		DocDate:               curDate,
		Status:                "AP",
		Months:                month,
		CreatedBy:             fmt.Sprint(memID),
		ApprovedBy:            "AUTO",
		ApprovedAt:            base.GetCurrentDateTimeT(),
	}
	arrSlsMasterMiningNodeTopup, err := models.AddSlsMasterMiningNodeTopup(tx, arrAddSlsMasterMiningNode)
	if err != nil {
		base.LogErrorLog("productService:TopupMiningNode():AddSlsMasterMiningNodeTopup():1", err.Error(), map[string]interface{}{"data": arrAddSlsMasterMiningNode}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	if helpers.CompareDateTime(endDate, "<", curDateTime) { // if today already passed end date
		startDate = curDateTime
		endDate = curDateTime
	}

	endDate = endDate.AddDate(year, month, day)

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: nodeID},
	)
	arrUpdateCols := map[string]interface{}{"start_date": startDate, "end_date": endDate}
	err = models.UpdatesFnTx(tx, "sls_master_mining_node", arrUpdCond, arrUpdateCols, false)
	if err != nil {
		base.LogErrorLog("productService:TopupMiningNode():UpdatesFnTx():1", err.Error(), map[string]interface{}{"condition": arrUpdCond, "updateCol": arrUpdateCols}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// validate payment with pay amount + deduct wallet
	msgStruct, arrData := wallet_service.PaymentProcess(tx, wallet_service.PaymentProcessStruct{
		MemberID:        memID,
		PrdCurrencyCode: prdCurrencyCode,
		Module:          paymentModule,
		Type:            topupMiningNodeInput.PaymentType,
		DocNo:           docNo,
		Amount:          unitPrice,
		Payments:        paymentStruct,
	}, 0, langCode)

	if msgStruct.Msg != "" {
		return msgStruct, nil
	}

	// updated previous sls_master_mining_node_topup status
	arrUpdCond2 := make([]models.WhereCondFn, 0)
	arrUpdCond2 = append(arrUpdCond2,
		models.WhereCondFn{Condition: "sls_master_mining_node_id = ?", CondValue: nodeID},
		models.WhereCondFn{Condition: "id != ?", CondValue: arrSlsMasterMiningNodeTopup.ID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "AP"},
	)
	arrUpdateCols2 := map[string]interface{}{"status": "EP", "expired_at": base.GetCurrentDateTimeT(), "expired_by": fmt.Sprint(memID)}
	err = models.UpdatesFnTx(tx, "sls_master_mining_node_topup", arrUpdCond2, arrUpdateCols2, false)
	if err != nil {
		base.LogErrorLog("productService:TopupMiningNode():UpdatesFnTx():1", err.Error(), map[string]interface{}{"condition": arrUpdCond, "updateCol": arrUpdateCols}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	return app.MsgStruct{Msg: ""}, arrData
}

func GetSubscriptionCancellationSetup(memberID int, langCode string) map[string]interface{} {
	var (
		packageValue      float64
		principleAmount   float64
		withdrawnAmount   float64
		penaltyAmount     float64
		totalRefundAmount float64
		prdCode           = "CONTRACT"
		statusCode        = "I"
		status            = "Inactive"
		arrReturnData     = map[string]interface{}{}
	)

	// get total package amount and cancellation fee
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: prdCode},
		models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
	)
	arrSlsMaster, _ := models.GetSlsMasterAscFn(arrSlsMasterFn, "", false)

	arrPrdMasterFn := make([]models.WhereCondFn, 0)
	arrPrdMasterFn = append(arrPrdMasterFn,
		models.WhereCondFn{Condition: " prd_master.prd_group = ? ", CondValue: prdCode},
	)
	arrPrdMaster, _ := models.GetPrdMasterFn(arrPrdMasterFn, "", false)

	arrCancellationFeeDetails := []map[string]interface{}{}

	if len(arrSlsMaster) > 0 {
		// validate prd_group_type.refund_setting
		var refundSetting, errMsg = GetPrdGroupTypeRefundSetup(arrPrdMaster[0].RefundSetting)
		if errMsg != "" {
			base.LogErrorLog("salesService:GetSubscriptionCancellationSetup():GetPrdGroupTypeRefundSetup():1", errMsg, map[string]interface{}{"value": arrPrdMaster[0].RefundSetting}, true)
			return map[string]interface{}{}
		}

		curDate, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
		latestSalesDate, _ := time.Parse("2006-01-02", arrSlsMaster[0].CreatedAt.Format("2006-01-02"))

		days := int(curDate.Sub(latestSalesDate).Hours() / 24)

		penaltyPerc := refundSetting.PenaltyPercDef
		for _, penalty := range refundSetting.Penalty {
			arrCancellationFeeDetails = append(arrCancellationFeeDetails, map[string]interface{}{
				"title": helpers.TranslateV2(penalty.Label, langCode, map[string]string{}),
				"value": fmt.Sprintf("%d%%", penalty.PenaltyPerc),
			})

			if days >= penalty.Min && penaltyPerc > penalty.PenaltyPerc {
				penaltyPerc = penalty.PenaltyPerc
			}
		}

		for _, arrSlsMasterV := range arrSlsMaster {
			principleAmount += arrSlsMasterV.TotalBv
			packageValue += arrSlsMasterV.TotalPv + arrSlsMasterV.TotalSv

			// validate penalty
			if penaltyPerc > 0 {
				// set penalty amount if doc not yet expired
				// if helpers.CompareDateTime(time.Now(), "<", arrSlsMaster[0].ExpiredAt) {
				penaltyAmount += float.Mul(arrSlsMasterV.TotalBv, float.Div(float64(penaltyPerc), 100))
				// }
			}
		}
	}

	// get withdrawn amount
	arrEwtWithdrawFn := make([]models.WhereCondFn, 0)
	arrEwtWithdrawFn = append(arrEwtWithdrawFn,
		models.WhereCondFn{Condition: " ewt_withdraw.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: " ewt_withdraw.status = ? ", CondValue: "AP"},
		models.WhereCondFn{Condition: " ewt_setup.ewallet_type_code In(?,'RPB') ", CondValue: "TP"}, // TP + RP-B
	)
	arrEwtWithdraw, _ := models.GetEwtWithdrawFn(arrEwtWithdrawFn, false)
	if len(arrEwtWithdraw) > 0 {
		for _, arrEwtWithdrawV := range arrEwtWithdraw {
			withdrawnAmount += arrEwtWithdrawV.TotalOut
		}
	}

	arrReturnData["cancellation_checkout_details"] = []map[string]interface{}{
		{
			"title": helpers.TranslateV2("package_amount", langCode, map[string]string{}),
			"value": helpers.CutOffDecimalv2(principleAmount, 2, ".", ",", true),
		},
		// {
		// 	"title": helpers.TranslateV2("withdrawn_amount", langCode, map[string]string{}),
		// 	"value": fmt.Sprintf("%s%s", "-", helpers.CutOffDecimalv2(withdrawnAmount, 2, ".", ",", true)),
		// },
		{
			"title": helpers.TranslateV2("cancellation_fee", langCode, map[string]string{}),
			"value": fmt.Sprintf("%s%s", "-", helpers.CutOffDecimalv2(penaltyAmount, 2, ".", ",", true)),
		},
	}

	totalRefundAmount = principleAmount - penaltyAmount - withdrawnAmount
	if totalRefundAmount < 0 {
		totalRefundAmount = 0
	}
	arrReturnData["package_amount"] = helpers.CutOffDecimalv2(packageValue, 2, ".", ",", true)
	arrReturnData["total_refund_amount"] = helpers.CutOffDecimalv2(totalRefundAmount, 2, ".", ",", true)

	// get cancellation status
	arrSlsMasterRefundBatchFn := make([]models.WhereCondFn, 0)
	arrSlsMasterRefundBatchFn = append(arrSlsMasterRefundBatchFn,
		models.WhereCondFn{Condition: "sls_master_refund_batch.member_id = ? ", CondValue: memberID},
	)
	arrSlsMasterRefundBatch, _ := models.GetSlsMasterRefundBatchFn(arrSlsMasterRefundBatchFn, "", false)

	if len(arrSlsMasterRefundBatch) > 0 {
		statusCode = arrSlsMasterRefundBatch[0].Status
		if arrSlsMasterRefundBatch[0].Status == "P" {
			status = "Pending"
		} else if arrSlsMasterRefundBatch[0].Status == "PR" {
			status = "Processing"
		} else if arrSlsMasterRefundBatch[0].Status == "R" {
			status = "Rejected"
		} else if arrSlsMasterRefundBatch[0].Status == "CP" {
			status = "Completed"
		}
	}

	arrReturnData["cancellation_status"] = helpers.TranslateV2(status, langCode, map[string]string{})
	arrReturnData["cancellation_status_code"] = statusCode
	arrReturnData["cancellation_fee_details"] = arrCancellationFeeDetails

	// validate if type b is refunded
	var typeBStatus = 1
	arrSlsMasterRefundBatchFn2 := make([]models.WhereCondFn, 0)
	arrSlsMasterRefundBatchFn2 = append(arrSlsMasterRefundBatchFn2,
		models.WhereCondFn{Condition: "sls_master_refund_batch.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master_refund_batch.status = ?", CondValue: "CP"}, // grab completed request
	)
	arrSlsMasterRefundBatch2, _ := models.GetSlsMasterRefundBatchFn(arrSlsMasterRefundBatchFn2, "", false)
	if len(arrSlsMasterRefundBatch2) > 0 {
		typeBStatus = 0
	}

	arrReturnData["type_b_status"] = typeBStatus

	// get package tier
	packageType, _ := member_service.GetMemberTier(memberID)

	arrReturnData["package_type"] = packageType

	// get expired at
	arrMemberTotalSalesFn := make([]models.WhereCondFn, 0)
	arrMemberTotalSalesFn = append(arrMemberTotalSalesFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
		models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
	)

	arrExp, _ := models.GetSlsMasterAscFn(arrMemberTotalSalesFn, "", false)
	expDateStr := ""
	if len(arrExp) > 0 {
		expDate := helpers.TruncateToDay(arrExp[0].ExpiredAt)
		currDate := helpers.TruncateToDay(base.GetCurrentDateTimeT())
		duration := expDate.Sub(currDate).Hours() / 24

		days := int(math.Ceil(duration))
		daysStr := strconv.Itoa(days)

		expDateStr = daysStr + " " + helpers.Translate("days", langCode)
	}

	arrReturnData["subsc_exp_days"] = expDateStr

	// earninng cap info
	var (
		capBalance         float64 = 0
		strCapBalance      string
		totalCapEarning    float64 = 0
		totalCapEarningStr string
		barPerc            float64 = 0
	)

	//get CAP Balance & setup
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "CAP"},
		models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: memberID},
	)
	bal, _ := models.GetMemberEwtSetupBalanceFn(memberID, arrCond, "", false)

	if len(bal) > 0 {
		capBalance = bal[0].Balance
	}

	strCapBalance = helpers.CutOffDecimal(capBalance, 0, ".", ",")

	//get totalCapEarning
	arrMemberTotalCapFn := make([]models.WhereCondFn, 0)
	arrMemberTotalCapFn = append(arrMemberTotalCapFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memberID},
		models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
		models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
	)
	arrMemberTotalCap, _ := models.GetMemberTotalCapFn(arrMemberTotalCapFn, false)

	totalCapEarning = arrMemberTotalCap.TotalAmount
	totalCapEarningStr = helpers.CutOffDecimal(totalCapEarning, 0, ".", ",")

	if totalCapEarning > 0 {
		barPerc, _ = decimal.NewFromFloat(capBalance).Div(decimal.NewFromFloat(totalCapEarning)).Float64()
	}

	arrReturnData["earning_cap_info"] = map[string]interface{}{
		"earning_cap_val":   strCapBalance,
		"earning_cap_total": totalCapEarningStr,
		"earning_cap_perc":  barPerc,
	}

	return arrReturnData
}

// PostSubscriptionCancellationStruct struct
type PostSubscriptionCancellationStruct struct {
	MemberID int
}

// PostSubscriptionCancellation func
func PostSubscriptionCancellation(tx *gorm.DB, data PostSubscriptionCancellationStruct, langCode string) (app.MsgStruct, map[string]string) {
	var (
		prdGroupType = "CONTRACT"
		batchDocType = "RFBT"
	)

	// validate if got existing pending refund or not
	arrSlsMasterRefundBatchFn := make([]models.WhereCondFn, 0)
	arrSlsMasterRefundBatchFn = append(arrSlsMasterRefundBatchFn,
		models.WhereCondFn{Condition: "sls_master_refund_batch.member_id = ? ", CondValue: data.MemberID},
		models.WhereCondFn{Condition: "sls_master_refund_batch.status != ?", CondValue: "R"}, // grab those pending/processing/completed request
	)
	arrSlsMasterRefundBatch, err := models.GetSlsMasterRefundBatchFn(arrSlsMasterRefundBatchFn, "", false)
	if err != nil {
		base.LogErrorLog("salesService:RefundSales():GetSlsMasterRefundBatchFn():1", err.Error(), map[string]interface{}{"condition": arrSlsMasterRefundBatchFn}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrSlsMasterRefundBatch) > 0 {
		return app.MsgStruct{Msg: "already_requested_refund"}, nil
	}

	// prd group type
	arrPrdGroupTypeFn := make([]models.WhereCondFn, 0)
	arrPrdGroupTypeFn = append(arrPrdGroupTypeFn,
		models.WhereCondFn{Condition: " prd_group_type.code = ? ", CondValue: prdGroupType},
		models.WhereCondFn{Condition: " prd_group_type.status = ? ", CondValue: "A"},
	)
	arrPrdGroupType, err := models.GetPrdGroupTypeFn(arrPrdGroupTypeFn, "", false)
	if err != nil {
		base.LogErrorLog("productService:PostSubscriptionCancellation():GetPrdGroupTypeFn():1", map[string]interface{}{"condition": arrPrdGroupTypeFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrPrdGroupType) <= 0 {
		base.LogErrorLog("productService:PostSubscriptionCancellation():GetPrdGroupTypeFn():1", map[string]interface{}{"condition": arrPrdGroupTypeFn}, "invalid_prd_group_type", true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	var prdGroupTypeRefundSetup, errMsg = GetPrdGroupTypeRefundSetup(arrPrdGroupType[0].RefundSetting)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	var (
		status = prdGroupTypeRefundSetup.Status
	)

	if !status {
		base.LogErrorLog("productService:PostSubscriptionCancellation()", map[string]interface{}{"condition": arrPrdGroupTypeFn}, "prd_group_type_does_not_support_refund", true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// get all related sls_master
	arrSlsMasterFn := make([]models.WhereCondFn, 0)
	arrSlsMasterFn = append(arrSlsMasterFn,
		models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: data.MemberID},
		models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: prdGroupType},
		models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
	)
	arrSlsMaster, err := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	if err != nil {
		base.LogErrorLog("salesService:PostSubscriptionCancellation()", "GetSlsMasterFn():1", err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrSlsMaster) <= 0 {
		return app.MsgStruct{Msg: "no_active_contract"}, nil
	}

	// get batch_no
	db := models.GetDB()
	batchNo, err := models.GetRunningDocNo(batchDocType, db) //get batch doc no
	if err != nil {
		base.LogErrorLog("salesService:PostSubscriptionCancellation():GetRunningDocNo():1", map[string]interface{}{"docType": batchDocType}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	err = models.UpdateRunningDocNo(batchDocType, db) //update batch doc no
	if err != nil {
		base.LogErrorLog("salesService:PostSubscriptionCancellation():UpdateRunningDocNo():1", map[string]interface{}{"docType": batchDocType}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// foreach and refund one by one
	for _, arrSlsMasterV := range arrSlsMaster {
		errMsg, _ := sales_service.RefundSales(tx, sales_service.RefundSalesStruct{
			BatchNo:       batchNo,
			MemberID:      data.MemberID,
			DocNo:         arrSlsMasterV.DocNo,
			RequestAmount: arrSlsMasterV.TotalBv,
		})

		if errMsg != "" {
			base.LogErrorLog("salesService:PostSubscriptionCancellation():RefundSales()", errMsg, map[string]interface{}{"member_id": data.MemberID, "doc_no": arrSlsMasterV.DocNo, "request_amount": arrSlsMasterV.TotalBv}, true)
			return app.MsgStruct{Msg: errMsg}, nil
		}
	}

	// insert to sls_master_refund
	var slsMasterRefundBatch = models.SlsMasterRefundBatch{
		DocNo:     batchNo,
		MemberID:  data.MemberID,
		Status:    "P",
		CreatedBy: fmt.Sprint(data.MemberID),
	}

	_, err = models.AddSlsMasterRefundBatch(tx, slsMasterRefundBatch)
	if err != nil {
		base.LogErrorLog("salesService:PostSubscriptionCancellation():AddSlsMasterRefund():1", err.Error(), map[string]interface{}{"condition": slsMasterRefundBatch}, true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	// arrReturnData := make(map[string]string)
	// arrReturnData["refunded_amount"] = helpers.CutOffDecimal(refundAmount, 8, ".", ",")

	return app.MsgStruct{Msg: ""}, nil
}
