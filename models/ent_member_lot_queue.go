package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberLotQueue struct
type EntMemberLotQueue struct {
	ID          int       `gorm:"primary_key" json:"id"`
	MemberID    int       `gorm:"column:member_id" json:"member_id"`
	MemberLot   string    `gorm:"column:member_lot" json:"member_lot"`
	SponsorID   int       `gorm:"column:sponsor_id" json:"sponsor_id"`
	SponsorLot  string    `gorm:"column:sponsor_lot" json:"sponsor_lot"`
	UplineID    int       `gorm:"column:upline_id" json:"upline_id"`
	UplineLot   string    `gorm:"column:upline_lot" json:"upline_lot"`
	LegNo       int       `gorm:"column:leg_no" json:"leg_no"`
	Type        string    `gorm:"column:type" json:"type"`
	Status      string    `gorm:"column:status" json:"status"`
	Batch       string    `gorm:"column:batch" json:"batch"`
	DtProcess   time.Time `gorm:"column:dt_process" json:"dt_process"`
	DtCreate    time.Time `gorm:"column:dt_create" json:"dt_create"`
	DtTimestamp time.Time `gorm:"column:dt_timestamp" json:"dt_timestamp"`
}

// AddEntMemberLotQueueStruct struct
type AddEntMemberLotQueueStruct struct {
	ID         int    `gorm:"primary_key" json:"id"`
	MemberID   int    `gorm:"column:member_id" json:"member_id"`
	MemberLot  string `gorm:"column:member_lot" json:"member_lot"`
	SponsorID  int    `gorm:"column:sponsor_id" json:"sponsor_id"`
	SponsorLot string `gorm:"column:sponsor_lot" json:"sponsor_lot"`
	UplineID   int    `gorm:"column:upline_id" json:"upline_id"`
	UplineLot  string `gorm:"column:upline_lot" json:"upline_lot"`
	LegNo      int    `gorm:"column:leg_no" json:"leg_no"`
	Status     string `gorm:"column:status" json:"status"`
	Type       string `gorm:"column:type" json:"type"`
	DtCreate   string `gorm:"column:dt_create" json:"dt_create"`
}

// AddEntMemberLotQueue add ent_member_lot_queue
func AddEntMemberLotQueue(tx *gorm.DB, tree AddEntMemberLotQueueStruct) (*AddEntMemberLotQueueStruct, error) {
	if err := tx.Table("ent_member_lot_queue").Create(&tree).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &tree, nil
}
