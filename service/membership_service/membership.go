package membership_service

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/service/promo_code_service"
	"github.com/smartblock/gta-api/service/wallet_service"
)

// func GetMembershipSetup
func GetMembershipSetup(memberID int, langCode string) ([]map[string]interface{}, string) {
	var (
		arrReturnData  []map[string]interface{}
		arrReturnDataV map[string]interface{}
	)

	arrEntMembershipSetupFn := make([]models.WhereCondFn, 0)
	arrEntMembershipSetupFn = append(arrEntMembershipSetupFn,
		models.WhereCondFn{Condition: " ent_membership_setup.status = ? ", CondValue: "A"},
	)
	arrEntMembershipSetup, err := models.GetEntMembershipSetup(arrEntMembershipSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("member_service:GetMembershipSetup():GetEntMembershipSetup():1", err.Error(), map[string]interface{}{"condition": arrEntMembershipSetupFn}, true)
		return nil, "something_went_wrong"
	}

	if len(arrEntMembershipSetup) > 0 {
		for _, arrEntMembershipSetupV := range arrEntMembershipSetup {
			var membershipCode = arrEntMembershipSetupV.Code

			arrEntMembershipSetupPriceFn := make([]models.WhereCondFn, 0)
			arrEntMembershipSetupPriceFn = append(arrEntMembershipSetupPriceFn,
				models.WhereCondFn{Condition: " ent_membership_setup_price.code = ? ", CondValue: membershipCode},
				models.WhereCondFn{Condition: " ent_membership_setup_price.status = ? ", CondValue: "A"},
			)
			arrEntMembershipSetupPrice, err := models.GetEntMembershipSetupPrice(arrEntMembershipSetupPriceFn, "", false)
			if err != nil {
				base.LogErrorLog("member_service:GetMembershipSetup():GetEntMembershipSetupPrice():1", err.Error(), map[string]interface{}{"condition": arrEntMembershipSetupPriceFn}, true)
				return nil, "something_went_wrong"
			}

			if len(membershipCode) <= 0 {
				continue
			}

			arrReturnDataV = map[string]interface{}{
				"code":   membershipCode,
				"name":   helpers.TranslateV2(arrEntMembershipSetupV.Name, langCode, nil),
				"amount": arrEntMembershipSetupPrice[0].UnitPrice,
			}

			arrReturnData = append(arrReturnData, arrReturnDataV)
		}
	}
	return arrReturnData, ""
}

type MembershipUpdateHistoryListFilter struct {
	MemberID int
	LangCode string
	DateFrom string
	DateTo   string
	Page     int64
}

type MembershipUpdateHistory struct {
	DocNo          string `json:"doc_no"`
	Membership     string `json:"membership"`
	MembershipCode string `json:"membership_code"`
	UnitPrice      string `json:"unit_price"`
	CreatedAt      string `json:"created_at"`
}

// func GetMembershipUpdateHistoryList
func GetMembershipUpdateHistoryList(arrData MembershipUpdateHistoryListFilter) (*app.ArrDataResponseList, string) {
	arrDataReturn := app.ArrDataResponseList{
		CurrentPageItems: make([]MembershipUpdateHistory, 0),
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member_membership_log.member_id = ? ", CondValue: arrData.MemberID},
	)

	if arrData.DateFrom != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(ent_member_membership_log.created_at) >= ? ", CondValue: arrData.DateFrom},
		)
	}

	if arrData.DateTo != "" {
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " date(ent_member_membership_log.created_at) <= ? ", CondValue: arrData.DateTo},
		)
	}

	arrPaginateData, arrEntMemberMembershipLog, _ := models.GetEntMemberMembershipLogPaginateFn(arrCond, arrData.Page, false)
	arrNewMemberSalesList := make([]MembershipUpdateHistory, 0)
	if len(arrEntMemberMembershipLog) > 0 {
		for _, arrEntMemberMembershipLogV := range arrEntMemberMembershipLog {
			arrNewMemberSalesList = append(arrNewMemberSalesList,
				MembershipUpdateHistory{
					DocNo:          arrEntMemberMembershipLogV.DocNo,
					Membership:     helpers.TranslateV2(arrEntMemberMembershipLogV.Name, arrData.LangCode, nil),
					MembershipCode: arrEntMemberMembershipLogV.Code,
					UnitPrice:      helpers.CutOffDecimal(arrEntMemberMembershipLogV.UnitPrice, 4, ".", ","),
					CreatedAt:      arrEntMemberMembershipLogV.CreatedAt.Format("2006-01-02 15:04:05"),
				},
			)
		}
	}

	arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(arrPaginateData.CurrentPage),
		PerPage:               int(arrPaginateData.PerPage),
		TotalCurrentPageItems: int(arrPaginateData.TotalCurrentPageItems),
		TotalPage:             int(arrPaginateData.TotalPage),
		TotalPageItems:        int(arrPaginateData.TotalPageItems),
		CurrentPageItems:      arrNewMemberSalesList,
	}
	return &arrDataReturn, ""
}

// PurchaseMembershipParam struct
type PurchaseMembershipParam struct {
	MemberID                                  int
	PackageCode, Payments, PinCode, PromoCode string
}

// PurchaseMembership func
func PurchaseMembership(tx *gorm.DB, purchaseMembershipParam PurchaseMembershipParam, langCode string) (app.MsgStruct, map[string]string) {
	var (
		docNo           string
		docType         string  = "MBS"
		memberID        int     = purchaseMembershipParam.MemberID
		packageCode     string  = purchaseMembershipParam.PackageCode
		module          string  = "MEMBERSHIP"
		prdCurrencyCode string  = "USDT"
		totalAmount     float64 = 0
		payableAmount   float64 = 0
		discountAmount  float64 = 0
		// curDate      string    = base.GetCurrentDateTimeT().Format("2006-01-02")
		// curDateTime  time.Time = base.GetCurrentDateTimeT()
		// approvableAt time.Time
		expiredAt, _ = base.StrToDateTime("9999-01-01", "2006-01-02")
	)

	// validate package code
	arrEntMembershipSetupFn := make([]models.WhereCondFn, 0)
	arrEntMembershipSetupFn = append(arrEntMembershipSetupFn,
		models.WhereCondFn{Condition: " ent_membership_setup.code = ? ", CondValue: packageCode},
		models.WhereCondFn{Condition: " ent_membership_setup.status = ? ", CondValue: "A"},
	)
	arrEntMembershipSetup, err := models.GetEntMembershipSetup(arrEntMembershipSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("membershipService:PurchaseMembership():GetEntMembershipSetup():1", map[string]interface{}{"condition": arrEntMembershipSetupFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrEntMembershipSetup) <= 0 {
		return app.MsgStruct{Msg: "invalid_package_code"}, nil
	}

	// get price
	arrEntMembershipSetupPriceFn := make([]models.WhereCondFn, 0)
	arrEntMembershipSetupPriceFn = append(arrEntMembershipSetupPriceFn,
		models.WhereCondFn{Condition: " ent_membership_setup_price.code = ? ", CondValue: packageCode},
		models.WhereCondFn{Condition: " ent_membership_setup_price.status = ? ", CondValue: "A"},
	)
	arrEntMembershipSetupPrice, err := models.GetEntMembershipSetupPrice(arrEntMembershipSetupPriceFn, "", false)
	if err != nil {
		base.LogErrorLog("membershipService:PurchaseMembership():GetEntMembershipSetupPrice():1", map[string]interface{}{"condition": arrEntMembershipSetupPriceFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrEntMembershipSetupPrice) <= 0 {
		base.LogErrorLog("membershipService:PurchaseMembership():GetEntMembershipSetupPrice():1", map[string]interface{}{"condition": arrEntMembershipSetupPriceFn}, "membership_price_not_found", true)
		return app.MsgStruct{Msg: "invalid_package_code"}, nil
	}

	totalAmount = arrEntMembershipSetupPrice[0].UnitPrice
	payableAmount = totalAmount

	// get doc_no
	db := models.GetDB()
	docNo, err = models.GetRunningDocNo(docType, db) //get doc no
	if err != nil {
		base.LogErrorLog("membershipService:PurchaseMembership():GetRunningDocNo():1", map[string]interface{}{"docType": docType}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	err = models.UpdateRunningDocNo(docType, db) //update doc no
	if err != nil {
		base.LogErrorLog("membershipService:PurchaseMembership():UpdateRunningDocNo():1", map[string]interface{}{"docType": docType}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	if purchaseMembershipParam.PinCode != "" { // pay with pin code
		// validate pin code
		errMsg, membershipPinCode := ValidateMembershipPinCode(memberID, purchaseMembershipParam.PinCode, packageCode)
		if errMsg != "" {
			return app.MsgStruct{Msg: errMsg}, nil
		}

		// set pin code status to "used"
		msgStruct := ProcessMembershipPinCode(tx, memberID, membershipPinCode.ID)
		if msgStruct.Msg != "" {
			return msgStruct, nil
		}
	} else { // pay with ewallet
		// validate payment
		if purchaseMembershipParam.Payments == "" {
			return app.MsgStruct{Msg: "please_provide_payment"}, nil
		}

		paymentStruct, errMsg := wallet_service.ConvertPaymentInputToStruct(purchaseMembershipParam.Payments)
		if errMsg != "" {
			return app.MsgStruct{Msg: errMsg}, nil
		}

		if purchaseMembershipParam.PromoCode != "" {
			// validate promo code if is povided
			promoCode := promo_code_service.PromoCode{
				MemberID: memberID,
				Code:     purchaseMembershipParam.PromoCode,
				Type:     "MEMBERSHIP",
			}
			errMsg := promoCode.Validate(tx)
			if errMsg != "" {
				return app.MsgStruct{Msg: errMsg}, nil
			}

			// calculate payable amount
			discountAmount = float.Mul(totalAmount, float.Div(promoCode.PromotionValue, 100))
			payableAmount = float.Sub(payableAmount, discountAmount)

			// claim promo code
			promoCode.ClaimedDocNo = docNo
			errMsg = promoCode.Use(tx)
			if errMsg != "" {
				return app.MsgStruct{Msg: errMsg}, nil
			}
		}

		// validate payment with pay amount + deduct wallet
		msgStruct, _ := wallet_service.PaymentProcess(tx, wallet_service.PaymentProcessStruct{
			MemberID:        memberID,
			PrdCurrencyCode: prdCurrencyCode,
			Module:          module,
			Type:            "DEFAULT",
			DocNo:           docNo,
			Remark:          "",
			Amount:          payableAmount,
			Payments:        paymentStruct,
		}, 0, langCode)

		if msgStruct.Msg != "" {
			return msgStruct, nil
		}
	}

	// get period setting and calculate expired date
	if arrEntMembershipSetup[0].PeriodSetting == "" {
		base.LogErrorLog("membershipService:PurchaseMembership()", map[string]interface{}{"arrEntMembershipSetup": arrEntMembershipSetup}, "invalid_period_setting", true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	periodSetting, errMsg := GetEntMembershipSetupPeriod(arrEntMembershipSetup[0].PeriodSetting)
	if errMsg != "" {
		return app.MsgStruct{Msg: errMsg}, nil
	}

	// calculate expired at for staking
	var (
		days, months, years int
	)

	if periodSetting.Days > 0 {
		days = periodSetting.Days
	}

	if periodSetting.Months > 0 {
		months = periodSetting.Months
	}

	if periodSetting.Years > 0 {
		years = periodSetting.Years
	}

	// default expired at count from current timestamp
	expiredAt = time.Now().AddDate(years, months, days)

	// update/insert to ent_member_membership
	arrEntMemberMembershipFn := make([]models.WhereCondFn, 0)
	arrEntMemberMembershipFn = append(arrEntMemberMembershipFn,
		models.WhereCondFn{Condition: " ent_member_membership.member_id = ? ", CondValue: memberID},
	)
	arrEntMemberMembership, err := models.GetEntMemberMembership(arrEntMemberMembershipFn, "", false)
	if err != nil {
		base.LogErrorLog("membershipService:PurchaseMembership():GetEntMemberMembership():1", map[string]interface{}{"condition": arrEntMemberMembershipFn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}
	if len(arrEntMemberMembership) <= 0 {
		// insert to ent_member_membership
		var addEntMemberMembershipParam = models.AddEntMemberMembershipStruct{
			MemberID:  memberID,
			BValid:    1,
			CreatedBy: fmt.Sprint(memberID),
			ExpiredAt: expiredAt,
		}

		_, err = models.AddEntMemberMembership(tx, addEntMemberMembershipParam)
		if err != nil {
			base.LogErrorLog("membershipService:PurchaseMembership():AddEntMemberMembership():1", map[string]interface{}{"param": addEntMemberMembershipParam}, err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil
		}
	} else {
		// recalculate expired at from current expire at if it is not expired yet
		if helpers.CompareDateTime(time.Now(), "<=", arrEntMemberMembership[0].ExpiredAt) {
			expiredAt = arrEntMemberMembership[0].ExpiredAt.AddDate(years, months, days)
		}

		// update ent_member_membership
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: arrEntMemberMembership[0].ID})
		updateColumn := map[string]interface{}{"b_valid": 1, "expired_at": expiredAt}
		err = models.UpdatesFnTx(tx, "ent_member_membership", arrUpdCond, updateColumn, false)
		if err != nil {
			base.LogErrorLog("membershipService:PurchaseMembership():UpdatesFnTx():1", map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, err.Error(), true)
			return app.MsgStruct{Msg: "something_went_wrong"}, nil
		}
	}

	// insert to ent_member_membership_log
	var addEntMemberMembershipLogParam = models.AddEntMemberMembershipLogStruct{
		MemberID:       memberID,
		DocNo:          docNo,
		Code:           packageCode,
		UnitPrice:      totalAmount,
		DiscountAmount: discountAmount,
		PaidAmount:     payableAmount,
		CreatedBy:      fmt.Sprint(memberID),
	}

	_, err = models.AddEntMemberMembershipLog(tx, addEntMemberMembershipLogParam)
	if err != nil {
		base.LogErrorLog("membershipService:PurchaseMembership():AddEntMemberMembershipLog():1", map[string]interface{}{"param": addEntMemberMembershipLogParam}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}, nil
	}

	return app.MsgStruct{Msg: ""}, nil
}

func ValidateMembershipPinCode(memberID int, pinCode, packageCode string) (string, *models.EntMemberMembershipPin) {
	arrEntMemberMembershipPinFn := []models.WhereCondFn{}
	arrEntMemberMembershipPinFn = append(arrEntMemberMembershipPinFn,
		models.WhereCondFn{Condition: " pin_code = ?", CondValue: pinCode},
		models.WhereCondFn{Condition: " membership_type = ?", CondValue: packageCode},
		models.WhereCondFn{Condition: " status = ?", CondValue: "A"})
	arrEntMemberMembershipPin, _ := models.GetEntMemberMembershipPin(arrEntMemberMembershipPinFn, "", false)
	if len(arrEntMemberMembershipPin) > 0 {
		entMemberMembershipPin := arrEntMemberMembershipPin[0]
		// validate if under same network. - to be enhance
		// if arrEntMemberMembershipPin[0].MemberID == memberID {
		// 	return app.MsgStruct{}, entMemberMembershipPin
		// }

		return "", entMemberMembershipPin
	}

	return "invalid_pin_code", nil
}

func ProcessMembershipPinCode(tx *gorm.DB, memberID int, pinID int) app.MsgStruct {
	// update ent_member_membership
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: pinID})
	updateColumn := map[string]interface{}{"status": "R", "redeemed_by": memberID, "redeemed_at": time.Now()}
	err := models.UpdatesFnTx(tx, "ent_member_membership_pin", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("membershipService:ProcessMembershipPinCode():UpdatesFnTx():1", map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, err.Error(), true)
		return app.MsgStruct{Msg: "something_went_wrong"}
	}

	return app.MsgStruct{Msg: ""}

}
