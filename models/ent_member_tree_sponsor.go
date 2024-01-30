package models

import (
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberTreeSponsor struct
type EntMemberTreeSponsor struct {
	ID         int    `gorm:"primary_key" json:"id"`
	MemberID   int    `json:"member_id"`
	MemberLot  string `json:"member_lot"`
	SponsorID  int    `json:"sponsor_id"`
	SponsorLot string `json:"sponsor_lot"`
	UplineID   int    `json:"upline_id"`
	UplineLot  string `json:"upline_lot"`
	LegNo      int    `json:"leg_no"`
	Lvl        int    `json:"lvl"`
	CreatedBy  int    `json:"created_at"`
}

// AddEntMemberTreeSponsor add member tree sponsor
func AddEntMemberTreeSponsor(tx *gorm.DB, tree EntMemberTreeSponsor) (*EntMemberTreeSponsor, error) {
	if err := tx.Table("ent_member_tree_sponsor").Create(&tree).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &tree, nil
}

// GetEntMemberEntMemberTreeSponsorFn get ent_member data with dynamic condition
func GetEntMemberEntMemberTreeSponsorFn(arrCond []WhereCondFn, debug bool) (*EntMemberTreeSponsor, error) {
	var entMemberTreeSponsor EntMemberTreeSponsor
	tx := db.Table("ent_member").
		Joins("JOIN ent_member_tree_sponsor on ent_member_tree_sponsor.member_id = ent_member.id").
		Select("ent_member_tree_sponsor.id, ent_member.id as member_id, ent_member_tree_sponsor.member_lot, ent_member_tree_sponsor.upline_id, ent_member_tree_sponsor.upline_lot, ent_member_tree_sponsor.sponsor_id, ent_member_tree_sponsor.sponsor_lot, ent_member_tree_sponsor.lvl")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&entMemberTreeSponsor).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if entMemberTreeSponsor.ID <= 0 {
		return nil, nil
	}

	return &entMemberTreeSponsor, nil
}

type ExtraListStruct struct {
	Key             string `json:"key"` // for app usage
	TranslatedLabel string `json:"translated_label"`
	LabelValue      string `json:"label_value"`
}

// MemberTreeStruct struct
type MemberTreeStruct struct {
	ID               int    `gorm:"primary_key" json:"-"`
	DownlineLot      string `gorm:"column:member_lot" json:"-"`
	SponsorLot       string `gorm:"column:sponsor_lot" json:"-"`
	Level            int    `gorm:"column:level" json:"level"`
	RootLvl          int    `gorm:"column:root_lvl" json:"-"`
	DownlineMemberID int    `gorm:"column:downline_member_id" json:"-"`
	DownlineNickName string `gorm:"column:downline_nick_name" json:"downline_username"`
	DownlineJoinDate string `gorm:"column:downline_join_date" json:"-"`
	DownlineCountry  string `gorm:"column:downline_country" json:"-"`
	// Rank             string `gorm:"column:rank" json:"rank"`
	ProfileImgURL string `gorm:"column:profile_img_url" json:"profile_img_url"`
	// DownlineTotalStar    int                 `gorm:"column:downline_total_star" json:"total_star"`
	// DownlineGradeRank    string              `gorm:"column:downline_grade_rank" json:"downline_grade_rank"`
	ChildrenStatus     int                 `gorm:"column:children_status" json:"children_status"`
	SponsorMemberID    int                 `gorm:"column:sponsor_member_id" json:"-"`
	SponsorNickName    string              `gorm:"column:sponsor_nick_name" json:"sponsor_nick_name"`
	TotalDirectSponsor string              `gorm:"column:total_direct_sponsor" json:"total_direct_sponsor"`
	ExtraList          []ExtraListStruct   `json:"extra_list"`
	ChildrenList       []*MemberTreeStruct `json:"children"`
}

// GetMemberTreeFn get ent_member data with dynamic condition
func GetMemberTreeFn(arrCond []WhereCondFn, strSelectColumn string, debug bool) ([]*MemberTreeStruct, error) {
	var result []*MemberTreeStruct
	tx := db.Table("ent_member_tree_sponsor").
		Joins("JOIN ent_member AS downline on ent_member_tree_sponsor.member_id = downline.id AND downline.status IN ('A')").
		Joins("JOIN sys_territory on sys_territory.id = downline.country_id").
		Joins("JOIN ent_member AS sponsor on ent_member_tree_sponsor.sponsor_id = sponsor.id AND sponsor.status IN ('A')").
		Joins("LEFT JOIN (SELECT IFNULL(count(id),0) as total, sponsor_id FROM ent_member_tree_sponsor GROUP BY sponsor_id) AS d on ent_member_tree_sponsor.member_id = d.sponsor_id").
		Select("ent_member_tree_sponsor.id, ent_member_tree_sponsor.member_lot, ent_member_tree_sponsor.sponsor_lot, ent_member_tree_sponsor.lvl as level, " +
			"downline.id as 'downline_member_id', downline.nick_name as 'downline_nick_name', DATE_FORMAT(downline.join_date, '%d/%m/%Y') as 'downline_join_date', sys_territory.name as downline_country, " +
			"downline.avatar as 'profile_img_url', " +
			"IF(IFNULL(d.total,0) > 0, '1', '0') as 'children_status', " +
			"sponsor.id as 'sponsor_member_id', sponsor.nick_name as 'sponsor_nick_name'" + strSelectColumn)

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

// EntMemberSponsor struct
type EntMemberSponsor struct {
	ID              int    `gorm:"primary_key" json:"id"`
	MemberID        int    `json:"member_id"`
	MemberLot       string `json:"member_lot"`
	SponsorID       int    `json:"sponsor_id"`
	SponsorUsername string `json:"sponsor_username"`
	SponsorLot      string `json:"sponsor_lot"`
	UplineID        int    `json:"upline_id"`
	UplineUsername  string `json:"upline_username"`
	LegNo           int    `json:"leg_no"`
	Lvl             int    `json:"lvl"`
}

// GetMemberSponsor get ent_member data with dynamic condition
func GetMemberSponsorFn(arrCond []WhereCondFn, debug bool) (*EntMemberSponsor, error) {
	var result EntMemberSponsor
	tx := db.Table("ent_member").
		Joins("JOIN ent_member_tree_sponsor on ent_member_tree_sponsor.member_id = ent_member.id").
		Joins("JOIN ent_member as sponsor on sponsor.id = ent_member_tree_sponsor.sponsor_id").
		Joins("LEFT JOIN ent_member as upline on upline.id = ent_member_tree_sponsor.upline_id").
		Select("ent_member_tree_sponsor.id, ent_member.id as member_id, ent_member_tree_sponsor.member_lot, ent_member_tree_sponsor.sponsor_id, sponsor.nick_name as sponsor_username, upline.nick_name as upline_username, ent_member_tree_sponsor.leg_no, ent_member_tree_sponsor.sponsor_lot, ent_member_tree_sponsor.lvl, ent_member_tree_sponsor.upline_id")

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

	if result.ID <= 0 {
		return nil, nil
	}

	return &result, nil
}

// GetTotalDirectSponsorStruct struct
type GetTotalDirectSponsorSalesStruct struct {
	TotalDirectSponsor float64 `gorm:"column:total_direct_sponsor" json:"total_direct_sponsor"`
}

// GetTotalDirectSponsorFn func
func GetTotalDirectSponsorFn(arrCond []WhereCondFn, debug bool) (*GetTotalDirectSponsorSalesStruct, error) {
	var result GetTotalDirectSponsorSalesStruct
	tx := db.Table("ent_member_tree_sponsor").
		Joins("INNER JOIN ent_member ON ent_member.id = ent_member_tree_sponsor.member_id").
		Select("COUNT(*) AS 'total_direct_sponsor'")

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

type NearestUplineStruct struct {
	EntMemberID int
	NickName    string
}

// GetNearestUplineByMemId func
func GetNearestUplineByMemId(entMemberID int, arrTargetID []int, sponsorType string, debug bool) (*NearestUplineStruct, error) {
	field := "sponsor_id"
	if sponsorType == "placement" {
		field = "upline_id"
	}

	withinSlice := false
	for _, b := range arrTargetID {
		if b == entMemberID {
			withinSlice = true
			break
		}
	}
	arrDataReturn := NearestUplineStruct{}

	arrCond := make([]WhereCondFn, 0)
	if withinSlice {
		arrCond = append(arrCond,
			WhereCondFn{Condition: "ent_member.id = ?", CondValue: entMemberID},
			WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		entMember, err := GetEntMemberFn(arrCond, "", debug)
		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if entMember == nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member", Data: err}
		}
		arrDataReturn.EntMemberID = entMember.ID
		arrDataReturn.NickName = entMember.NickName

		return &arrDataReturn, nil
	}

	nickName := ""
	for {
		for _, b := range arrTargetID {
			if b == entMemberID {
				withinSlice = true
				break
			}
		}

		if withinSlice {
			break
		}

		if entMemberID == 1 {
			// "neither_one_target_id_is_" + field + "_of_this_member"
			return nil, nil
			// return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "neither_one_target_id_is_" + field + "_of_this_member"}
		}

		arrCond = make([]WhereCondFn, 0)
		arrCond = append(arrCond,
			WhereCondFn{Condition: "ent_member_tree_sponsor.member_id = ?", CondValue: entMemberID},
		)
		treeArr, err := GetMemberSponsorFn(arrCond, debug)

		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if treeArr == nil {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "no_data_found_for_member_id_" + strconv.Itoa(entMemberID), Data: err}
		}

		entMemberID = treeArr.SponsorID
		nickName = treeArr.SponsorUsername
		if field == "upline_id" {
			entMemberID = treeArr.UplineID
		}
	}

	arrDataReturn.EntMemberID = entMemberID
	arrDataReturn.NickName = nickName

	return &arrDataReturn, nil
}

// TotalDirectSalesStruct struct
type TotalDirectSalesStruct struct {
	TotalBV    float64 `gorm:"column:total_bv" json:"total_bv"`
	TotalSales float64 `gorm:"column:total_sales" json:"total_sales"`
}

// GetTotalDirectSponsorFn func
func GetTotalDirectSalesFn(arrCond []WhereCondFn, debug bool) (*TotalDirectSalesStruct, error) {
	var result TotalDirectSalesStruct
	tx := db.Table("ent_member_tree_sponsor").
		Joins("INNER JOIN sls_master ON ent_member_tree_sponsor.member_id = sls_master.member_id").
		Select("SUM(total_bv) AS 'total_bv', SUM(total_amount) AS 'total_amount'")

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
