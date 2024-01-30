package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AddEntUserLeaderStruct struct
type AddEntUserLeaderStruct struct {
	ID         int       `gorm:"primary_key" json:"id"`
	MemberID   int       `gorm:"column:member_id" json:"member_id"`
	NickName   string    `gorm:"column:nick_name" json:"nick_name"`
	LeaderID   int       `gorm:"column:leader_id" json:"leader_id"`
	LeaderName string    `gorm:"column:leader_name" json:"leader_name"`
	BrokerID   int       `gorm:"column:broker_id" json:"broker_id"`
	BrokerName string    `gorm:"column:broker_name" json:"broker_name"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

// AddEntUserLeader add member
func AddEntUserLeader(tx *gorm.DB, arrData AddEntUserLeaderStruct) (*AddEntUserLeaderStruct, error) {
	if err := tx.Table("ent_user_leader").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// EntUserLeader struct
type EntUserLeader struct {
	ID         int       `gorm:"primary_key" json:"id"`
	MemberID   int       `gorm:"column:member_id" json:"member_id"`
	NickName   string    `gorm:"column:nick_name" json:"nick_name"`
	LeaderID   int       `gorm:"column:leader_id" json:"leader_id"`
	LeaderName string    `gorm:"column:leader_name" json:"leader_name"`
	BrokerID   int       `gorm:"column:broker_id" json:"broker_id"`
	BrokerName string    `gorm:"column:broker_name" json:"broker_name"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

// GetEntUserLeaderFn get ent_user_leader with dynamic condition
func GetEntUserLeaderFn(arrCond []WhereCondFn, debug bool) ([]*EntUserLeader, error) {
	var result []*EntUserLeader
	tx := db.Table("ent_user_leader")

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
