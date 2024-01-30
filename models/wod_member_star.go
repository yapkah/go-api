package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AddWodMemberStarStruct struct
type AddWodMemberStarStruct struct {
	ID           int    `json:"id" gorm:"column:id"`
	MemberID     int    `json:"member_id" gorm:"column:member_id"`
	RoomID       int    `json:"room_id" gorm:"column:room_id"`
	RoomBatch    string `json:"room_batch" gorm:"column:room_batch"`
	RoomTypeCode string `json:"room_type_code" gorm:"column:room_type_code"`
	Quantity     int    `json:"quantity" gorm:"column:quantity"`
}

// func AddWodMemberStar
func AddWodMemberStar(arrSaveData AddWodMemberStarStruct) (*AddWodMemberStarStruct, error) {
	if err := db.Table("wod_member_star").Create(&arrSaveData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrSaveData, nil
}

// WodMemberStarStruct struct
type WodMemberStarStruct struct {
	ID           int       `gorm:"primary_key" json:"id"`
	MemberID     int       `json:"member_id"`
	RoomID       int       `json:"room_id"`
	RoomBatch    string    `json:"room_batch"`
	RoomTypeCode string    `json:"room_type_code"`
	Quantity     int       `json:"quantity"`
	DtCreated    time.Time `json:"dt_created" gorm:"column:dt_created"`
}

// GetWodMemberStar get wod_member_star data with dynamic condition
func GetWodMemberStar(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*WodMemberStarStruct, error) {
	var result []*WodMemberStarStruct
	tx := db.Table("wod_member_star")
	if selectColumn != "" {
		tx = tx.Select(selectColumn)
	}
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

// CalWodMemberStarStruct struct
type CalWodMemberStarStruct struct {
	MemberID      int       `json:"member_id" gorm:"column:member_id"`
	RoomTypeCode  string    `json:"room_type_code" gorm:"column:room_type_code"`
	TotalStar     int       `json:"total_star" gorm:"column:total_star"`
	LastRoomBatch string    `json:"last_room_batch" gorm:"column:last_room_batch"`
	LastStar      time.Time `json:"last_star" gorm:"column:last_star"`
	LastStarID    int       `json:"last_star_id" gorm:"column:last_star_id"`
}

// GetCalWodMemberStarFn get data with dynamic condition
func GetCalWodMemberStarFn(arrCond []WhereCondFn, debug bool) ([]*CalWodMemberStarStruct, error) {
	var result []*CalWodMemberStarStruct
	tx := db.Table("wod_member_star").
		Joins("INNER JOIN ent_member ON wod_member_star.member_id = ent_member.id").
		Group("wod_member_star.member_id").
		Select("wod_member_star.member_id, wod_member_star.room_type_code, SUM(wod_member_star.quantity) 'total_star', MAX(wod_member_star.room_batch) 'last_room_batch', MAX(wod_member_star.dt_created) 'last_star', MAX(wod_member_star.id) 'last_star_id'")

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
