package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SwarmIP struct
type SwarmIP struct {
	ID               int       `gorm:"primary_key" json:"id"`
	IP               string    `json:"ip" gorm:"column:ip"`
	WalletAddress    string    `json:"wallet_address" gorm:"column:wallet_address"`
	ContractAddress  string    `json:"contract_address" gorm:"column:contract_address"`
	Status           string    `json:"status" gorm:"column:status"`
	TotalSettlements string    `json:"total_settlements" gorm:"column:total_settlements"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
	CreatedBy        string    `json:"created_by" gorm:"column:created_by"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at"`
	UpdatedBy        string    `json:"updated_by" gorm:"column:updated_by"`
}

// GetSwarmIPFn get ent_member_crypto with dynamic condition
func GetSwarmIPFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SwarmIP, error) {
	var result []*SwarmIP
	tx := db.Table("swarm_ip").
		Order("id desc")

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

// AddSwarmIP struct
type AddSwarmIP struct {
	ID               int       `gorm:"primary_key" json:"id"`
	IP               string    `json:"ip" gorm:"column:ip"`
	WalletAddress    string    `json:"wallet_address" gorm:"column:wallet_address"`
	ContractAddress  string    `json:"contract_address" gorm:"column:contract_address"`
	Status           string    `json:"status" gorm:"column:status"`
	TotalSettlements string    `json:"total_settlements" gorm:"column:total_settlements"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
	CreatedBy        string    `json:"created_by" gorm:"column:created_by"`
}

// AddSwarmIPFn func
func AddSwarmIPFn(tx *gorm.DB, swarmIP AddSwarmIP) (*AddSwarmIP, error) {
	if err := tx.Table("swarm_ip").Create(&swarmIP).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &swarmIP, nil
}

// GetSalesDetailsByWalletAddressStruct struct
type GetSalesDetailsByWalletAddressStruct struct {
	ID          int       `gorm:"primary_key" json:"id"`
	MemberID    int       `json:"member_id" gorm:"column:member_id"`
	SlsMasterID string    `json:"sls_master_id" gorm:"column:sls_master_id"`
	DocNo       string    `json:"doc_no" gorm:"column:doc_no"`
	Status      string    `json:"status" gorm:"column:status"`
	DocDate     time.Time `json:"doc_date" gorm:"column:doc_date"`
	IP          string    `json:"ip" gorm:"column:ip"`
}

// GetSalesDetailsByWalletAddress get ent_member_crypto with dynamic condition
func GetSalesDetailsByWalletAddress(walletAddress string, debug bool) ([]*GetSalesDetailsByWalletAddressStruct, error) {
	var result []*GetSalesDetailsByWalletAddressStruct
	tx := db.Table("swarm_ip").
		Select("sls_master.member_id, sls_master.id as sls_master_id, sls_master.doc_no, sls_master.status, sls_master.created_at as doc_date, swarm_ip.ip").
		Joins("INNER JOIN sls_master_mining_node ON sls_master_mining_node.ip = swarm_ip.ip").
		Joins("INNER JOIN sls_master_mining ON sls_master_mining.id = sls_master_mining_node.sls_master_mining_id").
		Joins("INNER JOIN sls_master ON sls_master.id = sls_master_mining.sls_master_id").
		Where("swarm_ip.wallet_address = ?", walletAddress).
		Order("swarm_ip.id desc")

	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
