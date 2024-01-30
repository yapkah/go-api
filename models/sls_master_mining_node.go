package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SlsMasterMiningNodeStruct struct
type SlsMasterMiningNodeStruct struct {
	ID          int       `gorm:"primary_key" json:"id"`
	SlsMasterID string    `json:"sls_master_id" gorm:"column:sls_master_id"`
	DocNo       string    `json:"doc_no" gorm:"column:doc_no"`
	Status      string    `json:"status" gorm:"column:status"`
	DocDate     time.Time `json:"doc_date" gorm:"column:doc_date"`
	StartDate   time.Time `json:"start_date" gorm:"column:start_date"`
	EndDate     time.Time `json:"end_date" gorm:"column:end_date"`
	IP          string    `json:"ip" gorm:"column:ip"`
}

// GetSlsMasterMiningNodeFn get ent_member_crypto with dynamic condition
func GetSlsMasterMiningNodeFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMasterMiningNodeStruct, error) {
	var result []*SlsMasterMiningNodeStruct
	tx := db.Table("sls_master_mining_node").
		Select("sls_master_mining_node.*, sls_master.id as sls_master_id, sls_master.created_at as doc_date, sls_master.doc_no, sls_master.status").
		Joins("INNER JOIN sls_master_mining ON sls_master_mining.id = sls_master_mining_node.sls_master_mining_id").
		Joins("INNER JOIN sls_master ON sls_master.id = sls_master_mining.sls_master_id").
		Order("sls_master_mining_node.id desc")

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

// AddSlsMasterMiningNodeStruct struct
type AddSlsMasterMiningNodeStruct struct {
	ID                int    `gorm:"primary_key" json:"id"`
	SlsMasterMiningID int    `json:"sls_master_mining_id" gorm:"column:sls_master_mining_id"`
	StartDate         string `json:"start_date" gorm:"column:start_date"`
	EndDate           string `json:"end_date" gorm:"column:end_date"`
	CreatedBy         string `json:"created_by"`
}

// AddSlsMasterMiningNode func
func AddSlsMasterMiningNode(tx *gorm.DB, slsMaster AddSlsMasterMiningNodeStruct) (*AddSlsMasterMiningNodeStruct, error) {
	if err := tx.Table("sls_master_mining_node").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}
