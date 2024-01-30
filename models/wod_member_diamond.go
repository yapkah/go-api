package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// WodMemberDiamondStruct struct
type WodMemberDiamondStruct struct {
	ID           int       `gorm:"primary_key" json:"id"`
	MemberID     int       `json:"member_id" gorm:"column:member_id"`
	RoomBatch    string    `json:"room_batch" gorm:"column:room_batch"`
	RoomType     string    `json:"room_type" gorm:"column:room_type"`
	DiamondCount int       `json:"diamond_count" gorm:"column:diamond_count"`
	LastStarID   int       `json:"last_star_id" gorm:"column:last_star_id"`
	GradeRank    string    `json:"grade_rank" gorm:"column:grade_rank"`
	Shares       string    `json:"shares" gorm:"column:shares"`
	IncomeCap    float64   `json:"income_cap" gorm:"column:income_cap"`
	Type         string    `json:"type" gorm:"column:type"`
	QualifyType  string    `json:"qualify_type" gorm:"column:qualify_type"`
	NSponsor     int       `json:"n_sponsor" gorm:"column:n_sponsor"`
	BDividend    int       `json:"b_dividend" gorm:"column:b_dividend"`
	BAdminFree   int       `json:"b_admin_free" gorm:"column:b_admin_free"`
	BPaid        int       `json:"b_paid" gorm:"column:b_paid"`
	BFreeze      int       `json:"b_freeze" gorm:"column:b_freeze"`
	DtCreated    time.Time `json:"dt_created" gorm:"column:dt_created"`
}

// GetWodMemberDiamondFn get wod_room_mast data with dynamic condition
func GetWodMemberDiamondFn(arrCond []WhereCondFn, debug bool) ([]*WodMemberDiamondStruct, error) {
	var result []*WodMemberDiamondStruct
	tx := db.Table("wod_member_diamond")

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

// AddWodMemberDiamondStruct struct
type AddWodMemberDiamondStruct struct {
	ID           int     `gorm:"primary_key" json:"id" gorm:"column:id"`
	MemberID     int     `json:"member_id" gorm:"column:member_id"`
	DiamondCount int     `json:"diamond_count" gorm:"column:diamond_count"`
	GradeRank    string  `json:"grade_rank" gorm:"column:grade_rank"`
	Shares       string  `json:"shares" gorm:"column:shares"`
	IncomeCap    float64 `json:"income_cap" gorm:"column:income_cap"`
	LastStarID   int     `json:"last_star_id" gorm:"column:last_star_id"`
	RoomBatch    string  `json:"room_batch" gorm:"column:room_batch"`
	RoomType     string  `json:"room_type" gorm:"column:room_type"`
}

// func AddWodMemberDiamond
func AddWodMemberDiamond(arrSaveData AddWodMemberDiamondStruct) (*AddWodMemberDiamondStruct, error) {
	if err := db.Table("wod_member_diamond").Create(&arrSaveData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrSaveData, nil
}

type MemberLastPlayGameStruct struct {
	SponsorNickName string `json:"sponsor_nick_name" gorm:"column:sponsor_nick_name"`
	// DownlineList    string `json:"downline_list" gorm:"column:downline_list"`
	ProfileName string `json:"profile_name" gorm:"column:profile_name"`
}

// func GetMemberLastPlayGame
func GetMemberLastPlayGame(arrCond []WhereCondFn, debug bool) ([]*MemberLastPlayGameStruct, error) {
	var result []*MemberLastPlayGameStruct
	tx := db.Table("wod_member_diamond").
		Joins("INNER JOIN tbl_bonus_diamond_star ON wod_member_diamond.id = tbl_bonus_diamond_star.t_diamond_id").
		Joins("INNER JOIN wod_member_rank ON wod_member_diamond.member_id = wod_member_rank.member_id").
		Joins("INNER JOIN wod_grade_setting ON wod_member_rank.grade_id = wod_grade_setting.id").
		Joins("INNER JOIN ent_member AS sponsor ON wod_member_diamond.member_id = sponsor.id").
		Joins("INNER JOIN ent_member AS downline ON tbl_bonus_diamond_star.t_downline_id = downline.id").
		Group("wod_member_diamond.member_id").
		// Select("sponsor.nick_name AS 'sponsor_nick_name' , GROUP_CONCAT(downline.nick_name ORDER BY downline.nick_name SEPARATOR ', ') AS 'downline_list'") // old code
		Select("sponsor.nick_name AS 'sponsor_nick_name' , wod_grade_setting.diamond_name AS 'profile_name'")

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
