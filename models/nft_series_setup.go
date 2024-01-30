package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// NftSeriesSetup struct
type NftSeriesSetup struct {
	ID              int       `gorm:"primary_key" json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Description     string    `json:"description"`
	EwalletTypeID   int       `json:"ewallet_type_id"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Purchase        int       `json:"purchase"`
	PurchaseLimit   float64   `json:"purchase_limit"`
	Staking         int       `json:"staking"`
	Supply          float64   `json:"supply"`
	Price           float64   `json:"price"`
	AirdropWalletID string    `json:"airdrop_wallet_id"`
}

// GetNftSeriesSetupFn
func GetNftSeriesSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*NftSeriesSetup, error) {
	var result []*NftSeriesSetup
	tx := db.Table("nft_series_setup")

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

// NftSeriesSetupDetail struct
type NftSeriesSetupDetail struct {
	ID              int       `gorm:"primary_key" json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Description     string    `json:"description"`
	EwalletTypeID   int       `json:"ewallet_type_id"`
	EwalletTypeCode string    `json:"ewallet_type_code"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Purchase        int       `json:"purchase"`
	PurchaseLimit   float64   `json:"purchase_limit"`
	Staking         int       `json:"staking"`
	Supply          float64   `json:"supply"`
	Price           float64   `json:"price"`
	AirdropWalletID string    `json:"airdrop_wallet_id"`
}

// GetNftSeriesSetupDetailFn
func GetNftSeriesSetupDetailFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*NftSeriesSetupDetail, error) {
	var result []*NftSeriesSetupDetail
	tx := db.Table("nft_series_setup").
		Select("nft_series_setup.*, ewt_setup.ewallet_type_code").
		Joins("inner join ewt_setup on ewt_setup.id = nft_series_setup.ewallet_type_id")

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
