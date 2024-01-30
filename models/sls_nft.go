package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SlsNft struct
type SlsNft struct {
	ID            int       `gorm:"primary_key" json:"id"`
	SlsMasterID   int       `json:"sls_master_id"`
	NftSeriesCode string    `json:"nft_series_code"`
	Unit          float64   `json:"unit"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// GetSlsNftFn
func GetSlsNftFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsNft, error) {
	var result []*SlsNft
	tx := db.Table("sls_nft").Order("sls_nft.created_at DESC")

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

// AddSlsNft func
func AddSlsNft(tx *gorm.DB, slsNftTier SlsNft) (*SlsNft, error) {
	if err := tx.Table("sls_nft").Create(&slsNftTier).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsNftTier, nil
}
