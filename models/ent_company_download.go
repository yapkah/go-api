package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntCompanyDownloads struct
type EntCompanyDownloads struct {
	ID           int       `gorm:"primary_key" json:"id"`
	Type         string    `json:"type" gorm:"column:type"`
	Title        string    `json:"title" gorm:"column:title"`
	Desc         string    `json:"desc" gorm:"column:desc"`
	FileName     string    `json:"file_name" gorm:"column:file_name"`
	Path         string    `json:"path" gorm:"column:path"`
	BShow        uint      `json:"b_show" gorm:"column:b_show"`
	Status       string    `json:"status" gorm:"column:status"`
	CategoryCode string    `json:"category_code" gorm:"column:category_code"`
	CategoryName string    `json:"category_name" gorm:"column:category_name"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    string    `json:"updated_by"`
}

type ExtraEntCompDwnlDataStruct struct {
	LangID int
	Type   string
}

// GetEntCompanyDownloadsFn get ent_company_downloads with dynamic condition
func GetEntCompanyDownloadsFn(arrData ExtraEntCompDwnlDataStruct, arrCond []WhereCondFn, debug bool) ([]*EntCompanyDownloads, error) {
	var result []*EntCompanyDownloads
	tx := db.Table("ent_company_downloads").
		Joins("INNER JOIN dl_list_category ON ent_company_downloads.category_id = dl_list_category.id").
		Select("ent_company_downloads.*, dl_list_category.code AS 'category_code', dl_list_category.name AS 'category_name'").
		Order("ent_company_downloads.id DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if arrData.LangID > 0 {
		tx = tx.Where("FIND_IN_SET(?, ent_company_downloads.language_id)", arrData.LangID)
	}
	if arrData.Type != "" {
		tx = tx.Where("FIND_IN_SET(?, ent_company_downloads.type)", arrData.Type)
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
