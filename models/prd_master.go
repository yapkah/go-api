package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// PrdMaster struct
type PrdMaster struct {
	ID               int       `gorm:"primary_key" json:"id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	PrdGroup         string    `json:"prd_group"`
	Path             string    `json:"path"`
	Color            string    `json:"color"`
	DtStart          time.Time `json:"dt_start"`
	DtEnd            time.Time `json:"dt_end"`
	Status           string    `json:"status"`
	Amount           float64   `json:"amount"`
	DocType          string    `json:"doc_type"`
	CurrencyCode     string    `json:"currency_code"`
	GasFeeSetting    string    `json:"gas_fee_setting"`
	RefundSetting    string    `json:"refund_setting"`
	RebatePerc       float64   `json:"rebate_perc"`
	PrincipleType    string    `json:"principle_type"`
	IncomeCap        float64   `json:"income_cap"`
	IncomeCapSetting string    `json:"income_cap_setting"`
	PrdGroupSetting  string    `json:"prd_group_setting"`
	Leverage         float64   `json:"leverage"`
	TopupSetting     string    `json:"topup_setting"`
	BroadbandSetting string    `json:"broadband_setting"`
	Setting          string    `json:"setting_setting"`
	CreatedAt        time.Time `json:"created_at"`
}

// GetPrdMasterFn get wod_room_type data with dynamic condition
func GetPrdMasterFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*PrdMaster, error) {
	var result []*PrdMaster
	tx := db.Table("prd_master").
		Select("prd_master.*, prd_master.leverage as income_cap, prd_price.unit_price as amount, prd_group_type.doc_type, prd_group_type.setting as prd_group_setting, prd_group_type.refund_setting, prd_group_type.currency_code, prd_group_type.principle_type, prd_group_type.topup_setting" + selectColumn).
		Joins("INNER JOIN prd_price ON prd_master.id = prd_price.prd_master_id").
		Joins("INNER JOIN prd_group_type ON prd_master.prd_group = prd_group_type.code")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("prd_master.seq_no").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// MemberHighestPackageInfo struct
type MemberHighestPackageInfo struct {
	PrdCode   string  `json:"prd_code" gorm:"column:prd_code"`
	Leverage  float64 `json:"leverage" gorm:"column:leverage"`
	UnitPrice float64 `json:"unit_price" gorm:"column:unit_price"`
}

// GetMemberHighestPackageInfo func
func GetMemberHighestPackageInfo(memID int, docAction, docNo string) (*MemberHighestPackageInfo, error) {
	var result MemberHighestPackageInfo
	tx := db.Table("sls_master").
		Select("prd_master.code as prd_code, prd_master.leverage, MAX(prd_price.unit_price)").
		Joins("INNER JOIN prd_master ON sls_master.prd_master_id = prd_master.id").
		Joins("INNER JOIN prd_price on sls_master.prd_master_id = prd_price.prd_master_id").
		Where("sls_master.member_id = ? AND sls_master.action = ? AND (sls_master.doc_no = ? OR sls_master.status = ?)", memID, docAction, docNo, "AP").
		Group("sls_master.prd_master_id").
		Order("MAX(prd_price.unit_price) DESC").
		Limit(1)

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
