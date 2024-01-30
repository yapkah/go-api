package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// PrdGroupType struct
type PrdGroupType struct {
	ID            int     `gorm:"primary_key" json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	BnsType       string  `json:"bns_type"`
	DocType       string  `json:"doc_type"`
	CurrencyCode  string  `json:"currency_code"`
	DecimalPoint  float64 `json:"decimal_point"`
	Status        string  `json:"status"`
	PrincipleType string  `json:"principle_type"`
	Setting       string  `json:"setting"`
	RefundSetting string  `json:"refund_setting"`
}

// GetPrdGroupTypeFn
func GetPrdGroupTypeFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*PrdGroupType, error) {
	var result []*PrdGroupType
	tx := db.Table("prd_group_type")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
