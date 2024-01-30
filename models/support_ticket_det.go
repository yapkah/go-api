package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SupportTicketDet struct
type SupportTicketDet struct {
	ID        int       `gorm:"primary_key" json:"id"`
	TicketID  int       `json:"ticket_id"`
	TicketMsg string    `json:"ticket_msg"`
	FileURL1  string    `gorm:"column:file_url_1" json:"file_url_1"`
	FileURL2  string    `gorm:"column:file_url_2" json:"file_url_2"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

type SupportTicketHisDet struct {
	ID                 int       `gorm:"primary_key" json:"id"`
	TicketID           int       `json:"ticket_id"`
	TicketTitle        string    `json:"ticket_title"`
	TicketMsg          string    `json:"ticket_msg"`
	TicketCategory     string    `json:"ticket_category"`
	TicketCategoryName string    `json:"ticket_category_name"`
	Status             string    `json:"status"`
	FileURL1           string    `gorm:"column:file_url_1" json:"file_url_1"`
	FileURL2           string    `gorm:"column:file_url_2" json:"file_url_2"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          string    `json:"created_by"`
	CreatedUser        string    `json:"created_user"`
}

// GetSysSupportTicketFn get sys_support_ticket data with dynamic condition
func GetSupportTicketDetFn(arrCond []WhereCondFn, limit int, order string, debug bool) ([]*SupportTicketDet, error) {
	var result []*SupportTicketDet
	tx := db.Table("support_ticket_det")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if order != "" {
		tx = tx.Order("support_ticket_det.id ASC")
	} else {
		tx = tx.Order("support_ticket_det.id DESC")
	}

	if limit > 0 {
		tx = tx.Limit(limit)
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

// AddSupportTicketDet add support_ticket_det records`
func AddSupportTicketDet(tx *gorm.DB, saveData SupportTicketDet) (*SupportTicketDet, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddSupportTicketDet-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

func AddSupportTicketDetWithoutTx(saveData SupportTicketDet) (*SupportTicketDet, error) {
	if err := db.Create(&saveData).Error; err != nil {
		ErrorLog("AddSupportTicketDetWithoutTx-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

func GetMemberSupportTicketDetailsByTicketCode(ticketCode string) ([]*SupportTicketHisDet, error) {
	var rst []*SupportTicketHisDet

	query := db.Table("support_ticket_det").
		Select("support_ticket_det.*, support_ticket_mast.ticket_title, support_ticket_mast.ticket_category, support_ticket_mast.status, support_ticket_category.name as ticket_category_name, CASE WHEN admins.nick_name IS NULL THEN '' ELSE admins.nick_name END AS created_user").
		Joins("join support_ticket_mast ON support_ticket_det.ticket_id = support_ticket_mast.id").
		Joins("join support_ticket_category ON support_ticket_category.code = support_ticket_mast.ticket_category").
		Joins("left join admins ON support_ticket_det.created_by = admins.nick_name")

	if ticketCode != "" {
		query = query.Where("support_ticket_mast.ticket_code = ?", ticketCode)
	}

	err := query.Order("support_ticket_det.id asc").Find(&rst).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return rst, nil
}
