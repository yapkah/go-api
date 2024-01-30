package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberBallot struct
type EntMemberBallot struct {
	ID              int       `gorm:"primary_key" json:"id"`
	TicketNo        int       `json:"doc_no"`
	MemType         string    `json:"mem_type"`
	MemberId        int       `json:"member_id"`
	TransDate       time.Time `json:"trans_date"`
	TypeCode        string    `json:"type_code"`
	CurrencyFrom    string    `json:"currency_from"`
	Amount          float64   `json:"amount"`
	Price           float64   `json:"price"`
	ConvertedAmount float64   `json:"converted_amount"`
	CurrencyTo      string    `json:"currency_to"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       string    `json:"created_by"`
}

// GetEntMemberBallotFn get ent_member_ballot data with dynamic condition
func GetEntMemberBallotFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberBallot, error) {
	var result []*EntMemberBallot
	tx := db.Table("ent_member_ballot")
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

// AddEntMemberballot add ent_member_ballot records`
func AddEntMemberBallot(tx *gorm.DB, saveData EntMemberBallot) (*EntMemberBallot, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddEntMemberBallot-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

type LatestBallotTicketNoStruct struct {
	TicketNo int `json:"ticket_no"`
}

func GetLatestBallotTicketNo(arrCond []WhereCondFn, debug bool) (*LatestBallotTicketNoStruct, error) {
	var result LatestBallotTicketNoStruct
	tx := db.Table("ent_member_ballot")
	tx = tx.Select("MAX(ent_member_ballot.ticket_no) AS 'ticket_no'")

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

	return &result, nil
}

type LatestBallotVolumeStruct struct {
	Volume float64 `json:"volume"`
}

func GetLatestBallotVolume(arrCond []WhereCondFn, debug bool) (*LatestBallotVolumeStruct, error) {
	var result LatestBallotVolumeStruct
	tx := db.Table("ent_member_ballot")
	tx = tx.Select("SUM(ent_member_ballot.amount) AS 'volume'")

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

	return &result, nil
}
