package models

import (
	"net/http"

	"github.com/smartblock/gta-api/pkg/e"
)

// PrdMiningPrice struct
type PrdMiningPrice struct {
	ID       int     `gorm:"primary_key" json:"id"`
	FilPrice float64 `gorm:"fil_price" json:"fil_price"`
	SecPrice float64 `gorm:"sec_price" json:"sec_price"`
	XchPrice float64 `gorm:"xch_price" json:"xch_price"`
	BzzPrice float64 `gorm:"bzz_price" json:"bzz_price"`
}

// GetLatestPrdMiningPriceByPrdMasterID
func GetLatestPrdMiningPriceByPrdMasterID(prdMasterID int, curDate string) (*PrdMiningPrice, error) {
	var prdMiningPrice PrdMiningPrice

	err := db.Table("prd_mining_price").
		Where("prd_master_id = ?", prdMasterID).
		Where("status = ?", "A").
		Where("date <= ?", curDate).
		Order("created_at DESC").
		First(&prdMiningPrice).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &prdMiningPrice, nil
}
