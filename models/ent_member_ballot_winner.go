package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberBallot struct
type EntMemberBallotWinner struct {
	ID         int       `gorm:"primary_key" json:"id"`
	TicketNo   int       `json:"doc_no"`
	BscAddress string    `json:"bsc_address"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
}

// GetEntMemberBallotWinnerFn get ent_member_ballot_winner data with dynamic condition
func GetEntMemberBallotWinnerFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberBallotWinner, error) {
	var result []*EntMemberBallotWinner
	tx := db.Table("ent_member_ballot_winner")
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
