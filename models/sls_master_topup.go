package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SlsMasterTopupStruct struct
type SlsMasterTopupStruct struct {
	ID           int       `gorm:"primary_key" json:"id"`
	SlsMasterID  int       `json:"sls_master_id" gorm:"column:sls_master_id"`
	MemberID     int       `json:"member_id" gorm:"column:member_id"`
	DocNo        string    `json:"doc_no" gorm:"column:doc_no"`
	DocDate      string    `json:"doc_date" gorm:"column:doc_date"`
	BnsBatch     string    `json:"bns_batch" gorm:"column:bns_batch"`
	Status       string    `json:"status" gorm:"column:status"`
	StatusDesc   string    `json:"status_desc" gorm:"column:status_desc"`
	TotalAmount  float64   `json:"total_amount" gorm:"column:total_amount"`
	TotalBv      float64   `json:"total_bv" gorm:"column:total_bv"`
	CurrencyCode string    `json:"currency_code" gorm:"column:currency_code"`
	DecimalPoint int       `json:"decimal_point" gorm:"column:decimal_point"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetSlsMasterTopupFn get ent_member_crypto with dynamic condition
func GetSlsMasterTopupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMasterTopupStruct, error) {
	var result []*SlsMasterTopupStruct
	tx := db.Table("sls_master_topup").
		Select("sls_master_topup.*, sys_general.name as status_desc, prd_group_type.currency_code, prd_group_type.decimal_point").
		Joins("INNER JOIN sls_master on sls_master.id = sls_master_topup.sls_master_id").
		Joins("INNER JOIN prd_master on prd_master.id = sls_master.prd_master_id").
		Joins("INNER JOIN prd_group_type on prd_group_type.code = prd_master.prd_group").
		Joins("INNER JOIN sys_general on sls_master_topup.status = sys_general.code AND sys_general.type = ? ", "sales-status").
		Order("sls_master_topup.id desc")

	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}

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

// AddSlsMasterTopupStruct struct
type AddSlsMasterTopupStruct struct {
	ID          int       `gorm:"primary_key" json:"id"`
	SlsMasterID int       `json:"sls_master_id" gorm:"column:sls_master_id"`
	MemberID    int       `json:"member_id" gorm:"column:member_id"`
	DocNo       string    `json:"doc_no" gorm:"column:doc_no"`
	DocDate     string    `json:"doc_date" gorm:"column:doc_date"`
	BnsBatch    string    `json:"bns_batch" gorm:"column:bns_batch"`
	Status      string    `json:"status" gorm:"column:status"`
	TotalAmount float64   `json:"total_amount" gorm:"column:total_amount"`
	TotalBv     float64   `json:"total_bv" gorm:"column:total_bv"`
	Leverage    float64   `json:"leverage" gorm:"column:leverage"`
	CreatedBy   string    `json:"created_by"`
	ApprovedAt  time.Time `json:"approved_at"`
	ApprovedBy  string    `json:"approved_by"`
}

// AddSlsMasterTopup func
func AddSlsMasterTopup(tx *gorm.DB, slsMaster AddSlsMasterTopupStruct) (*AddSlsMasterTopupStruct, error) {
	if err := tx.Table("sls_master_topup").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}
