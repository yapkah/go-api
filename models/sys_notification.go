package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SysNotification struct
type SysNotification struct {
	ID           int       `gorm:"primary_key" json:"id"`
	ApiKeyID     int       `json:"api_key_id" gorm:"column:api_key_id"`
	Type         string    `json:"type" gorm:"column:type"`
	PNType       string    `json:"pn_type" gorm:"column:pn_type"`
	MemberID     int       `json:"member_id" gorm:"column:member_id"`
	TBnsID       string    `json:"t_bns_id" gorm:"column:t_bns_id"`
	Icon         string    `json:"icon" gorm:"column:icon"`
	Label        string    `json:"label" gorm:"column:label"`
	Title        string    `json:"title" gorm:"column:title"`
	Msg          string    `json:"msg" gorm:"column:msg"`
	LangCode     string    `json:"lang_code" gorm:"column:lang_code"`
	CustMsg      string    `json:"cust_msg" gorm:"column:cust_msg"`
	BShow        int       `json:"b_show" gorm:"column:b_show"`
	PNSendStatus int       `json:"pn_send_status" gorm:"column:pn_send_status"`
	Status       string    `json:"status" gorm:"column:status"`
	RedirectURL  string    `json:"redirect_url" gorm:"column:redirect_url"`
	ViewedUsers  string    `json:"viewed_users" gorm:"column:viewed_users"`
	CountryID    int       `json:"country_id" gorm:"column:country_id"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    string    `json:"updated_by"`
}

// GetSysNotificationFn get sys_notification data with dynamic condition
func GetSysNotificationFn(arrCond []WhereCondFn, debug bool) (*SysNotification, error) {
	var result SysNotification
	tx := db.Table("sys_notification")

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

	if result.ID <= 0 {
		return nil, nil
	}

	return &result, nil
}

// AddSysNotificationStruct struct
type AddSysNotificationStruct struct {
	ID           int       `gorm:"primary_key" json:"id"`
	ApiKeyID     int       `json:"api_key_id" gorm:"column:api_key_id"`
	Type         string    `json:"type" gorm:"column:type"`
	PNType       string    `json:"pn_type" gorm:"column:pn_type"`
	MemberID     int       `json:"member_id" gorm:"column:member_id"`
	TBnsID       string    `json:"t_bns_id" gorm:"column:t_bns_id"`
	Icon         string    `json:"icon" gorm:"column:icon"`
	Label        string    `json:"label" gorm:"column:label"`
	Title        string    `json:"title" gorm:"column:title"`
	Msg          string    `json:"msg" gorm:"column:msg"`
	LangCode     string    `json:"lang_code" gorm:"column:lang_code"`
	CustMsg      string    `json:"cust_msg" gorm:"column:cust_msg"`
	BShow        int       `json:"b_show" gorm:"column:b_show"`
	PNSendStatus int       `json:"pn_send_status" gorm:"column:pn_send_status"`
	Status       string    `json:"status" gorm:"column:status"`
	RedirectURL  string    `json:"redirect_url" gorm:"column:redirect_url"`
	ViewedUsers  string    `json:"viewed_users" gorm:"column:viewed_users"`
	CountryID    int       `json:"country_id" gorm:"column:country_id"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
}

// AddSysNotification add sys_notification
func AddSysNotification(saveData AddSysNotificationStruct) (*AddSysNotificationStruct, error) {
	if err := db.Table("sys_notification").Create(&saveData).Error; err != nil {
		ErrorLog("AddSysNotification-AddSysNotification", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

// GetSysNotificationPaginateFn get sys_notification with dynamic condition
func GetSysNotificationPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*SysNotification, error) {
	var (
		result                []*SysNotification
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("sys_notification").
		Order("sys_notification.created_at DESC")

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

// GetSysNotificationLimitFn get sys_notification with dynamic condition
func GetSysNotificationLimitFn(arrCond []WhereCondFn, limit int, debug bool) ([]*SysNotification, error) {
	var (
		result []*SysNotification
	)
	tx := db.Table("sys_notification").
		Order("sys_notification.created_at DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	// Pagination and limit
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
