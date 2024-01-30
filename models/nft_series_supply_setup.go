package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// NftSeriesSupplySetup struct
type NftSeriesSupplySetup struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Type      int       `json:"country_id" gorm:"column:country_id"`
	Value     int       `json:"company_id" gorm:"column:company_id"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

type TotalNftSeriesSupplySetup struct {
	TotalValue int `json:"total_value" gorm:"column:total_value"`
}

func GetTotalNftSeriesSupplySetup(nftSeriesCode string) (*TotalNftSeriesSupplySetup, error) {
	var result TotalNftSeriesSupplySetup

	tx := db.Table("nft_series_supply_setup").
		Select("SUM(value) as total_value").
		Where("type LIKE ?", nftSeriesCode)

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
