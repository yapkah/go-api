package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysSupportTicket struct
type SysSupportTicket struct {
	ID        int       `gorm:"primary_key" json:"id"`
	TicketNo  int       `json:"ticket_no`
	MemberId  int       `json:"member_id`
	Address   string    `json:"address"`
	Issue     string    `json:"issue"`
	FileName1 string    `gorm:"column:file_name_1" json:"file_name_1"`
	FileURL1  string    `gorm:"column:file_url_1" json:"file_url_1"`
	FileName2 string    `gorm:"column:file_name_2" json:"file_name_2"`
	FileURL2  string    `gorm:"column:file_url_2" json:"file_url_2"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

// GetSysSupportTicketFn get sys_support_ticket data with dynamic condition
func GetSysSupportTicketFn(arrCond []WhereCondFn, limit int, debug bool) ([]*SysSupportTicket, error) {
	var result []*SysSupportTicket
	tx := db.Table("sys_support_ticket").
		Order("sys_support_ticket.id DESC")
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

// AddSysSupportTicket add sys_support_ticket records`
func AddSysSupportTicket(tx *gorm.DB, saveData SysSupportTicket) (*SysSupportTicket, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddSysSupportTicket-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

type GetSysSupportTicketNo struct {
	TicketNo int `gorm:"column:ticket_no" json:"ticket_no"`
}

func GetSysSupportTicketNoFn(arrCond []WhereCondFn, debug bool) (*GetSysSupportTicketNo, error) {
	var result GetSysSupportTicketNo
	tx := db.Table("sys_support_ticket").
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
