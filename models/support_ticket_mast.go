package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SupportTicketMast struct
type SupportTicketMast struct {
	ID                 int       `gorm:"primary_key" json:"id"`
	TicketCode         string    `json:"ticket_code"`
	TicketCategory     string    `json:"ticket_category"`
	TicketCategoryName string    `json:"ticket_category_name"`
	MemberID           int       `json:"member_id"`
	TicketTitle        string    `json:"ticket_title"`
	Status             string    `json:"status"`
	MemberShow         int       `json:"member_show"`
	AdminShow          int       `json:"admin_show"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          string    `json:"created_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          string    `json:"updated_by"`
}

// GetSysSupportTicketFn get sys_support_ticket data with dynamic condition
func GetSupportTicketMastFn(arrCond []WhereCondFn, limit int, debug bool) ([]*SupportTicketMast, error) {
	var result []*SupportTicketMast
	tx := db.Table("support_ticket_mast").
		Joins("inner join support_ticket_category on support_ticket_category.code = support_ticket_mast.ticket_category").
		Select("support_ticket_mast.*, support_ticket_category.name as ticket_category_name").
		Order("support_ticket_mast.id DESC")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	if limit > 0 {
		tx = tx.Limit(limit)
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// AddSupportTicketMastStruct struct
type AddSupportTicketMastStruct struct {
	ID             int       `gorm:"primary_key" json:"id"`
	TicketCode     string    `json:"ticket_code"`
	TicketCategory string    `json:"ticket_category"`
	MemberID       int       `json:"member_id"`
	TicketTitle    string    `json:"ticket_title"`
	Status         string    `json:"status"`
	MemberShow     int       `json:"member_show"`
	AdminShow      int       `json:"admin_show"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      string    `json:"updated_by"`
}

// AddSupportTicketMast add support_ticket_mast records`
func AddSupportTicketMast(tx *gorm.DB, saveData AddSupportTicketMastStruct) (*AddSupportTicketMastStruct, error) {
	if err := tx.Table("support_ticket_mast").Create(&saveData).Error; err != nil {
		ErrorLog("AddSupportTicketMast-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

type GetSupportTicketNo struct {
	TicketNo int `gorm:"column:ticket_no" json:"ticket_no"`
}

func GetSupportTicketNoFn(arrCond []WhereCondFn, debug bool) (*GetSupportTicketNo, error) {
	var result GetSupportTicketNo
	tx := db.Table("support_ticket_mast").
		Select("MAX(ticket_no) AS ticket_no").
		Order("id desc")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			if v.CondValue != nil {
				tx = tx.Where(v.Condition, v.CondValue)
			} else {
				tx = tx.Where(v.Condition)
			}
		}
	}
	if debug {
		tx = tx.Debug()
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
