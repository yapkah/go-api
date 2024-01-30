package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SwarmData struct
type SwarmData struct {
	ID            int     `gorm:"primary_key" json:"id"`
	MemberID      int     `json:"member_id" gorm:"column:member_id"`
	WalletAddress string  `json:"wallet_address" gorm:"column:wallet_address"`
	TotalBalance  float64 `json:"total_balance" gorm:"column:total_balance"`
	TotalMined    float64 `json:"total_mined" gorm:"column:total_mined"`
}

// GetSwarmDataFn get ent_member_crypto with dynamic condition
func GetSwarmDataFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SwarmData, error) {
	var result []*SwarmData
	tx := db.Table("swarm_data").
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

// AddSwarmData struct
type AddSwarmData struct {
	ID            int     `gorm:"primary_key" json:"id"`
	MemberID      int     `json:"member_id" gorm:"column:member_id"`
	WalletAddress string  `json:"wallet_address" gorm:"column:wallet_address"`
	TotalBalance  float64 `json:"total_balance" gorm:"column:total_balance"`
	TotalMined    float64 `json:"total_mined" gorm:"column:total_mined"`
}

// AddSwarmDataFn func
func AddSwarmDataFn(tx *gorm.DB, swarmIP AddSwarmData) (*AddSwarmData, error) {
	if err := tx.Table("swarm_data").Create(&swarmIP).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &swarmIP, nil
}
