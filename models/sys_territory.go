package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysTerritory struct
type SysTerritory struct {
	ID              int    `gorm:"primary_key" json:"id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	CallingNoPrefix string `json:"calling_no_prefix"`
	Status          string `json:"status"`
	CountryFlagUrl  string `json:"country_flag_url"`
}

// GetSysTerritoryFn get ent_member data with dynamic condition
func GetSysTerritoryFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*SysTerritory, error) {
	var sysTerritory SysTerritory
	tx := db.Table("sys_territory")
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
	err := tx.Find(&sysTerritory).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if sysTerritory.ID <= 0 {
		return nil, nil
	}

	return &sysTerritory, nil
}

// GetCountryByCode func
func GetCountryByCode(code string) (*SysTerritory, error) {
	var sys SysTerritory
	err := db.Where("code = ? AND territory_type = 'country' AND status = 'A'", code).First(&sys).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &sys, nil
}

// GetCountryByCode func
func GetCountryByID(ID int) (*SysTerritory, error) {
	var sys SysTerritory
	err := db.Where("id = ? AND territory_type = 'country' AND status = 'A'", ID).First(&sys).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &sys, nil
}

// ExistCountryCode func
func ExistCountryCode(code string) bool {
	sys, err := GetCountryByCode(code)

	if err != nil {
		return false
	}

	if sys == nil {
		return false
	}

	return true
}

// GetCountryList func
func GetCountryList() ([]*SysTerritory, error) {
	var sys []*SysTerritory
	err := db.Where("status = 'A' AND territory_type='country'").Order("name").Find(&sys).Error

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return sys, nil
}
