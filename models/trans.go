package models

import (
	"math"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TransStruct struct
type TransStruct struct {
	Name    string `gorm:"name" json:"name"`
	English string `json:"english"`
	Chinese string `json:"chinese"`
}

// func GetTransFn
func GetTransFn(arrCond []WhereCondFn, debug bool) ([]*TransStruct, error) {
	var result []*TransStruct
	tx := db.Table("trans")
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

// GetTransUpdatePaginateFn get trans with dynamic condition
func GetTransUpdatePaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*TransStruct, error) {
	var (
		result                []*TransStruct
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("trans").
		Joins("INNER JOIN dl_list_category ON ent_company_announcement.category_id = dl_list_category.id").
		Select("ent_company_announcement.*, dl_list_category.code AS 'announcement_category_code', dl_list_category.name AS 'announcement_category_name'").
		Order("ent_company_announcement.id DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	arrLimitRowsSetting, _ := GetSysGeneralSetupByID("defaultlimitrow")
	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	// Total Records
	tx.Count(&totalRecord)
	oriPage := page
	if page != 0 {
		page--
	}

	newOffset := page * limit

	// Pagination and limit
	err := tx.Limit(limit).Offset(newOffset).Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	perPage = limit

	totalCurrentPageItems = int64(len(result))

	// return ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, nil
	arrPaginateData = SQLPaginateStdReturn{
		CurrentPage:           oriPage,
		PerPage:               perPage,
		TotalCurrentPageItems: totalCurrentPageItems,
		TotalPage:             totalPage,
		TotalPageItems:        totalRecord,
	}
	return arrPaginateData, result, nil
}
