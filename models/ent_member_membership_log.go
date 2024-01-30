package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberMembershipLog struct
type EntMemberMembershipLog struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"member_id" gorm:"column:member_id"`
	DocNo     string    `json:"doc_no" gorm:"column:doc_no"`
	Code      string    `json:"code" gorm:"column:code"`
	UnitPrice float64   `json:"unit_price" gorm:"column:unit_price"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

func GetEntMemberMembershipLog(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EntMemberMembershipLog, error) {
	var result []*EntMemberMembershipLog
	tx := db.Table("ent_member_membership_log")

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

// AddEntMemberMembershipLogStruct struct
type AddEntMemberMembershipLogStruct struct {
	ID             int       `gorm:"primary_key" json:"id"`
	MemberID       int       `json:"member_id" gorm:"column:member_id"`
	DocNo          string    `json:"doc_no" gorm:"column:doc_no"`
	Code           string    `json:"code" gorm:"column:code"`
	UnitPrice      float64   `json:"unit_price" gorm:"column:unit_price"`
	DiscountAmount float64   `json:"discount_amount" gorm:"column:discount_amount"`
	PaidAmount     float64   `json:"paid_amount" gorm:"column:paid_amount"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
}

// AddEntMemberMembershipLog func
func AddEntMemberMembershipLog(tx *gorm.DB, entMemberMembershipLog AddEntMemberMembershipLogStruct) (*AddEntMemberMembershipLogStruct, error) {
	if err := tx.Table("ent_member_membership_log").Create(&entMemberMembershipLog).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &entMemberMembershipLog, nil
}

// EntMemberMembershipLogPaginate struct
type EntMemberMembershipLogPaginate struct {
	ID        int       `gorm:"primary_key" json:"id"`
	MemberID  int       `json:"member_id" gorm:"column:member_id"`
	DocNo     string    `json:"doc_no" gorm:"column:doc_no"`
	Code      string    `json:"code" gorm:"column:code"`
	Name      string    `json:"name" gorm:"column:name"`
	UnitPrice float64   `json:"unit_price" gorm:"column:unit_price"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// GetEntMemberMembershipLogPaginateFn func
func GetEntMemberMembershipLogPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*EntMemberMembershipLogPaginate, error) {
	var (
		result                []*EntMemberMembershipLogPaginate
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("ent_member_membership_log").
		Joins("INNER JOIN ent_membership_setup ON ent_membership_setup.code = ent_member_membership_log.code").
		Select("ent_member_membership_log.*, ent_membership_setup.name").
		Order("ent_member_membership_log.id DESC")

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
