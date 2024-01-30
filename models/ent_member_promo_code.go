package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberPromoCode struct
type EntMemberPromoCode struct {
	ID             int       `gorm:"primary_key" json:"id"`
	MemberID       int       `json:"member_id"`
	Type           string    `json:"type"`
	Code           string    `json:"code"`
	Status         string    `json:"status"`
	PromotionValue float64   `json:"promotion_value"`
	ExpiryDate     time.Time `json:"expiry_date"`
	ClaimedDocNo   string    `json:"claimed_doc_no"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      string    `json:"updated_by"`
}

func GetEntMemberPromoCode(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberPromoCode, error) {
	var result []*EntMemberPromoCode
	tx := db.Table("ent_member_promo_code")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("ent_member_promo_code.id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
