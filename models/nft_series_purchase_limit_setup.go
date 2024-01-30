package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// NftSeriesPurchaseLimitSetup struct
type NftSeriesPurchaseLimitSetup struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Type      int       `json:"country_id" gorm:"column:country_id"`
	Value     int       `json:"company_id" gorm:"column:company_id"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

type TotalNftSeriesPurchaseLimitSetup struct {
	TotalValue int `json:"total_value" gorm:"column:total_value"`
}

func GetTotalNftSeriesPurchaseLimitSetup(nftSeriesCode string) (*TotalNftSeriesPurchaseLimitSetup, error) {
	var result TotalNftSeriesPurchaseLimitSetup

	tx := db.Table("nft_series_purchase_limit_setup").
		Select("SUM(value) as total_value").
		Where("type LIKE ?", nftSeriesCode)

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
