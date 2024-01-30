package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntCompanyAnnouncement struct
type EntCompanyAnnouncement struct {
	ID                       int       `gorm:"primary_key" json:"id"`
	Title                    string    `json:"title" gorm:"column:title"`
	Desc                     string    `json:"desc" gorm:"column:desc"`
	FolderPath               string    `json:"folder_path" gorm:"column:folder_path"`
	Path                     string    `json:"path" gorm:"column:path"`
	MemberShow               uint      `json:"member_show" gorm:"column:member_show"`
	Status                   string    `json:"status" gorm:"column:status"`
	SeqNo                    int       `json:"seq_no" gorm:"column:seq_no"`
	CategoryID               int       `json:"category_id" gorm:"column:category_id"`
	LanguageID               string    `json:"language_id" gorm:"column:language_id"`
	TUsername                string    `json:"t_username" gorm:"column:t_username"`
	TLeader                  string    `json:"t_leader" gorm:"column:t_leader"`
	OpenMemberGroup          string    `json:"open_member_group" gorm:"column:open_member_group"`
	PopUp                    uint      `json:"popup" gorm:"column:popup"`
	CreatedAt                time.Time `json:"created_at"`
	CreatedBy                string    `json:"created_by"`
	UpdatedAt                time.Time `json:"updated_at"`
	UpdatedBy                string    `json:"updated_by"`
	AnnouncementCategoryCode string    `json:"announcement_category_code" gorm:"column:announcement_category_code"`
	AnnouncementCategoryName string    `json:"announcement_category_name" gorm:"column:announcement_category_name"`
}

type ExtraEntCompAnnDataStruct struct {
	LangID int
	Type   string
}

// GetEntCompanyAnnouncementFn get ent_member_crypto with dynamic condition
func GetEntCompanyAnnouncementFn(arrData ExtraEntCompAnnDataStruct, arrCond []WhereCondFn, debug bool) ([]*EntCompanyAnnouncement, error) {
	var result []*EntCompanyAnnouncement
	tx := db.Table("ent_company_announcement").
		Joins("INNER JOIN dl_list_category ON ent_company_announcement.category_id = dl_list_category.id").
		Select("ent_company_announcement.*, dl_list_category.code AS 'announcement_category_code', dl_list_category.name AS 'announcement_category_name'").
		Order("ent_company_announcement.id DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if arrData.LangID > 0 {
		tx = tx.Where("FIND_IN_SET(?, ent_company_announcement.language_id)", arrData.LangID)
	}
	if arrData.Type != "" {
		tx = tx.Where("FIND_IN_SET(?, ent_company_announcement.type)", arrData.Type)
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

// GetEntCompanyAnnouncementPaginateFn get ent_member_crypto with dynamic condition
func GetEntCompanyAnnouncementPaginateFn(arrData ExtraEntCompAnnDataStruct, arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*EntCompanyAnnouncement, error) {
	var (
		result                []*EntCompanyAnnouncement
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("ent_company_announcement").
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
