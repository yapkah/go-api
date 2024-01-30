package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AddSlsMasterMiningContractStruct struct
type AddSlsMasterMiningContractStruct struct {
	ID           int       `gorm:"primary_key" json:"id"`
	MemberID     int       `gorm:"column:member_id" json:"member_id"`
	SerialNumber string    `gorm:"column:serial_number" json:"serial_number"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy    string    `gorm:"column:created_by" json:"created_by"`
}

// AddSlsMasterMiningContract add sls_master_mining_contract
func AddSlsMasterMiningContract(arrData AddSlsMasterMiningContractStruct) (*AddSlsMasterMiningContractStruct, error) {
	if err := db.Table("sls_master_mining_contract").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// SlsMasterMiningContract struct
type SlsMasterMiningContract struct {
	ID           int       `gorm:"primary_key" json:"id"`
	MemberID     int       `gorm:"column:member_id" json:"member_id"`
	SerialNumber string    `gorm:"column:serial_number" json:"serial_number"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy    string    `gorm:"column:created_by" json:"created_by"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy    string    `gorm:"column:updated_by" json:"updated_by"`
}

// GetSlsMasterMiningContract get sls_master_mining_contract with dynamic condition
func GetSlsMasterMiningContract(arrCond []WhereCondFn, debug bool) ([]*SlsMasterMiningContract, error) {
	var result []*SlsMasterMiningContract
	tx := db.Table("sls_master_mining_contract").
		Order("sls_master_mining_contract.id DESC")

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
