package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// WodMemberRank struct
type WodMemberRank struct {
	ID           int       `gorm:"primary_key" json:"id"`
	CurrentRank  string    `json:"current_rank"`
	TotalStar    int       `json:"total_star"`
	TotalDiamond int       `json:"total_diamond"`
	Grade        string    `json:"grade_rank"`
	Remark       string    `json:"remark"`
	ShareNumber  string    `json:"number_of_share"`
	AchievedAt   time.Time `json:"achieved_at"`
}

// GetWodMemberRankDetails get member rank details
func GetWodMemberRankDetails(arrCond []WhereCondFn, debug bool) (*WodMemberRank, error) {
	var result WodMemberRank
	tx := db.Table("wod_member_rank").
		Select("wod_member_rank.current_rank, wod_member_rank.total_star, wod_grade_setting.diamond_count as total_diamond, wod_grade_setting.grade_rank, wod_grade_setting.number_of_share").
		Joins("LEFT JOIN wod_grade_setting ON wod_grade_setting.id = wod_member_rank.grade_id").
		Order("wod_member_rank.grade_id DESC")

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

// WodMemberRankStruct struct
type WodMemberRankStruct struct {
	ID            int       `gorm:"primary_key" json:"id" gorm:"column:id"`
	MemberID      int       `json:"member_id" gorm:"column:member_id"`
	CurrentRank   string    `json:"current_rank" gorm:"column:current_rank"`
	GradeID       int       `json:"grade_id" gorm:"column:grade_id"`
	TotalStar     int       `json:"total_star" gorm:"column:total_star"`
	DLastStar     time.Time `json:"d_last_star" gorm:"column:d_last_star"`
	LastRoomBatch string    `json:"last_room_batch" gorm:"column:last_room_batch"`
	RoomType      string    `json:"room_type" gorm:"column:room_type"`
}

// GetWodMemberRankFn get wod_member_rank data with dynamic condition
func GetWodMemberRankFn(arrCond []WhereCondFn, debug bool) ([]*WodMemberRankStruct, error) {
	var result []*WodMemberRankStruct
	tx := db.Table("wod_member_rank")
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

// AddWodMemberRankStruct struct
type AddWodMemberRankStruct struct {
	ID            int       `gorm:"primary_key" json:"id" gorm:"column:id"`
	MemberID      int       `json:"member_id" gorm:"column:member_id"`
	CurrentRank   string    `json:"current_rank" gorm:"column:current_rank"`
	GradeID       int       `json:"grade_id" gorm:"column:grade_id"`
	TotalStar     int       `json:"total_star" gorm:"column:total_star"`
	DLastStar     time.Time `json:"d_last_star" gorm:"column:d_last_star"`
	LastRoomBatch string    `json:"last_room_batch" gorm:"column:last_room_batch"`
	RoomType      string    `json:"room_type" gorm:"column:room_type"`
}

// AddWodMemberRank function
func AddWodMemberRank(arrSaveData AddWodMemberRankStruct) (*AddWodMemberRankStruct, error) {
	if err := db.Table("wod_member_rank").Create(&arrSaveData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrSaveData, nil
}
