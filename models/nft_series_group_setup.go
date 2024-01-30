package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// NftSeriesGroupSetup struct
type NftSeriesGroupSetup struct {
	ID          int    `gorm:"primary_key" json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// GetNftSeriesGroupSetupFn
func GetNftSeriesGroupSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*NftSeriesGroupSetup, error) {
	var result []*NftSeriesGroupSetup
	tx := db.Table("nft_series_group_setup")

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
