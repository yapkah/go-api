package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SlsMasterMining struct
type SlsMasterMining struct {
	ID            int     `gorm:"primary_key" json:"id"`
	SlsMasterID   int     `json:"sls_master_id"`
	MemberID      int     `json:"member_id" gorm:"column:member_id"`
	SerialNumber  string  `json:"serial_number" gorm:"column:serial_number"`
	MachineType   string  `json:"machine_type"`
	FilPrice      float64 `json:"fil_price"`
	FilecoinPrice float64 `json:"filecoin_price"`
	FilTib        float64 `json:"fil_tib"`
	SecPrice      float64 `json:"sec_price"`
	SecTib        float64 `json:"sec_tib"`
	XchPrice      float64 `json:"xch_price"`
	XchTib        float64 `json:"xch_tib"`
	BzzPrice      float64 `json:"bzz_price"`
	BzzTib        float64 `json:"bzz_tib"`
}

// GetSlsMasterMiningFn
func GetSlsMasterMiningFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMasterMining, error) {
	var result []*SlsMasterMining
	tx := db.Table("sls_master_mining").
		Joins("INNER JOIN sls_master ON sls_master_mining.sls_master_id = sls_master.id").
		Select("sls_master_mining.*, sls_master.member_id")

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

// AddSlsMasterMiningStruct struct
type AddSlsMasterMiningStruct struct {
	ID            int     `gorm:"primary_key" json:"id"`
	SlsMasterID   int     `json:"sls_master_id" gorm:"column:sls_master_id"`
	MachineType   string  `json:"machine_type" gorm:"column:machine_type"`
	FilPrice      float64 `json:"fil_price" gorm:"column:fil_price"`
	FilecoinPrice float64 `json:"filecoin_price" gorm:"column:filecoin_price"`
	FilTib        float64 `json:"fil_tib" gorm:"column:fil_tib"`
	SecPrice      float64 `json:"sec_price" gorm:"column:sec_price"`
	SecTib        float64 `json:"sec_tib" gorm:"column:sec_tib"`
	XchPrice      float64 `json:"xch_price" gorm:"column:xch_price"`
	XchTib        float64 `json:"xch_tib" gorm:"column:xch_tib"`
	BzzPrice      float64 `json:"bzz_price" gorm:"column:bzz_price"`
	BzzTib        float64 `json:"bzz_tib" gorm:"column:bzz_tib"`
	CreatedBy     string  `json:"created_by"`
}

// AddSlsMasterMining func
func AddSlsMasterMining(tx *gorm.DB, slsMaster AddSlsMasterMiningStruct) (*AddSlsMasterMiningStruct, error) {
	if err := tx.Table("sls_master_mining").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}

// TotalBZZSalesStruct struct
type TotalBZZSalesStruct struct {
	TotalSales float64 `gorm:"column:total_sales" json:"total_sales"`
	TotalNodes float64 `gorm:"column:total_nodes" json:"total_nodes"`
}

// GetTotalBZZSalesFn get TotalNetworkBZZSalesStruct data with dynamic condition
func GetTotalBZZSalesFn(arrCond []WhereCondFn, debug bool) (*TotalBZZSalesStruct, error) {
	var result TotalBZZSalesStruct
	tx := db.Table("sls_master_mining").
		Joins("INNER JOIN sls_master ON sls_master_mining.sls_master_id = sls_master.id").
		Joins("INNER JOIN ent_member ON sls_master.member_id = ent_member.id AND ent_member.status = 'A'").
		Select("SUM(sls_master.total_amount) AS 'total_sales', SUM(sls_master_mining.bzz_tib) AS 'total_nodes'")

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

	return &result, nil
}
