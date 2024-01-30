package promo_code_service

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
)

// PromoCode struct
type PromoCode struct {
	MemberID       int
	Code           string
	Type           string
	PromotionValue float64
	ClaimedDocNo   string
}

// Validate func
func (promoCode *PromoCode) Validate(tx *gorm.DB) string {
	// validate promo code
	arrEntMemberPromoCodeFn := make([]models.WhereCondFn, 0)
	arrEntMemberPromoCodeFn = append(arrEntMemberPromoCodeFn,
		models.WhereCondFn{Condition: "member_id = ?", CondValue: promoCode.MemberID},
		models.WhereCondFn{Condition: "type = ?", CondValue: promoCode.Type},
		models.WhereCondFn{Condition: "code = ?", CondValue: promoCode.Code},
		models.WhereCondFn{Condition: "status = ?", CondValue: "ACTIVE"},
	)
	arrEntMemberPromoCode, err := models.GetEntMemberPromoCode(arrEntMemberPromoCodeFn, "", false)
	if err != nil {
		base.LogErrorLog("promoCodeService:Validate()", "GetEntMemberPromoCode():1", err.Error(), true)
		return "something_went_wrong"
	}
	if len(arrEntMemberPromoCode) <= 0 {
		return "invalid_promo_code"
	}

	// validate expiry date
	nowTime := time.Now()
	if nowTime.Equal(arrEntMemberPromoCode[0].ExpiryDate) || nowTime.After(arrEntMemberPromoCode[0].ExpiryDate) {
		return "promo_code_expired"
	}

	promoCode.PromotionValue = arrEntMemberPromoCode[0].PromotionValue

	return ""
}

// Use promo code
func (promoCode *PromoCode) Use(tx *gorm.DB) string {
	arrUpdateEntMemberPromoCodeFn := make([]models.WhereCondFn, 0)
	arrUpdateEntMemberPromoCodeFn = append(arrUpdateEntMemberPromoCodeFn,
		models.WhereCondFn{
			Condition: "code = ?",
			CondValue: promoCode.Code,
		},
		models.WhereCondFn{
			Condition: "status = ?",
			CondValue: "ACTIVE",
		})
	updateColumn := map[string]interface{}{
		"status":         "CLAIMED",
		"claimed_doc_no": promoCode.ClaimedDocNo,
		"updated_by":     promoCode.MemberID,
	}
	_ = models.UpdatesFnTx(tx, "ent_member_promo_code", arrUpdateEntMemberPromoCodeFn, updateColumn, false)
	return ""
}
