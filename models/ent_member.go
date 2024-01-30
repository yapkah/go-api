package models

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

const memberUserType = "MEM"

// EntMember struct
type EntMember struct {
	ID                       int       `gorm:"primary_key" json:"id"`
	DPK                      string    `json:"-" gorm:"column:d_pk"`
	PrivateKey               string    `json:"-" gorm:"column:private_key"`
	CountryID                int       `json:"country_id"`
	CompanyID                int       `json:"company_id"`
	MainID                   int       `json:"main_id"`
	TaggedMemberID           int       `json:"tagged_member_id" gorm:"column:tagged_member_id"`
	MemberType               string    `json:"member_type"`
	Source                   string    `json:"source"`
	NickName                 string    `json:"nick_name"`
	CurrentProfile           int       `json:"current_profile"`
	FirstName                string    `json:"first_name"`
	LastName                 string    `json:"last_name"`
	Code                     string    `json:"code" gorm:"column:code"`
	Status                   string    `json:"status"` // A: active | I : inactive | T: terminate | S: suspend
	IdentityType             string    `json:"identity_type"`
	IdentityNo               string    `json:"identity_no"`
	Wechat                   string    `json:"wechat"`
	Avatar                   string    `json:"avatar"`
	Path                     string    `json:"path"`
	QrPath                   string    `json:"qr_path"`
	Gender                   string    `json:"gender"`
	GenderCode               string    `json:"gender_code"`
	GenderID                 int       `json:"gender_id"`
	RaceID                   int       `json:"race_id"`
	MaritalID                int       `json:"marital_id"`
	BirthDate                string    `json:"birth_date"`
	PreferLanguageCode       string    `json:"prefer_language_code"`
	JoinDate                 string    `json:"join_date"`
	Sms                      int       `json:"sms"`
	Remark                   string    `json:"remark"`
	DefaultAutoWithdrawal    string    `json:"default_auto_withdrawal"`
	UsdDefaultAutoWithdrawal string    `json:"usd_default_auto_withdrawal"`
	CreatedAt                time.Time `json:"created_at"`
	CreatedBy                string    `json:"created_by"`
	UpdatedAt                time.Time `json:"updated_at"`
	UpdatedBy                string    `json:"updated_by"`
	SuspendedAt              time.Time `json:"suspended_on"`
	SuspendedBy              string    `json:"suspended_by"`
	CancelledAt              time.Time `json:"cancelled_on"`
	CancelledBy              string    `json:"cancelled_by"`
	TerminatedAt             time.Time `json:"terminated_on"`
	TerminatedBy             string    `json:"terminated_by"`
}

// AddEntMemberStruct struct
type AddEntMemberStruct struct {
	ID                 int       `gorm:"primary_key" json:"id"`
	MainID             int       `json:"main_id"`
	CountryID          int       `json:"country_id"`
	CompanyID          int       `json:"company_id"`
	MemberType         string    `json:"member_type"`
	Source             string    `json:"source"`
	NickName           string    `json:"nick_name"`
	FirstName          string    `json:"first_name"`
	Code               string    `json:"code"`
	CurrentProfile     int       `json:"current_profile"`
	Status             string    `json:"status"` // A: active | I : inactive | T: terminate | S: suspend
	PreferLanguageCode string    `json:"prefer_language_code"`
	JoinDate           string    `json:"join_date"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          string    `json:"created_by"`
}

// GetEntMemberFn get ent_member data with dynamic condition
func GetEntMemberFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*EntMember, error) {
	var entMember EntMember
	tx := db.Table("ent_member").
		Joins("LEFT JOIN sys_general as gender ON ent_member.gender_id = gender.id").
		Select("ent_member.*, gender.name as gender, gender.code as gender_code")

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
	err := tx.Find(&entMember).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if entMember.ID <= 0 {
		return nil, nil
	}

	return &entMember, nil
}

// GetFirstEntMemberFn get first ent_member data with dynamic condition
func GetFirstEntMemberFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*EntMember, error) {
	var entMember EntMember
	tx := db.Table("ent_member")
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
	err := tx.First(&entMember).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if entMember.ID <= 0 {
		return nil, nil
	}

	return &entMember, nil
}

// AddEntMember add member
func AddEntMember(tx *gorm.DB, entMember AddEntMemberStruct) (*AddEntMemberStruct, error) {
	if err := tx.Table("ent_member").Create(&entMember).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &entMember, nil
}

// MemberAvailableDividendTicketStruct struct
type MemberAvailableDividendTicketStruct struct {
	SponsorID               int    `gorm:"column:sponsor_id" json:"sponsor_id"`
	AvailableDividendTicket int    `gorm:"column:available_dividend_ticket" json:"available_dividend_ticket"`
	NCount                  int    `gorm:"column:n_count" json:"n_count"`
	ADownline               string `gorm:"column:a_downline" json:"a_downline"`
}

// GetMemberAvailableDividendTicketFn func
func GetMemberAvailableDividendTicketFn(arrCond []WhereCondFn, debug bool) (*MemberAvailableDividendTicketStruct, error) {
	var result MemberAvailableDividendTicketStruct
	tx := db.Table("ent_member")
	tx = tx.Joins("JOIN ent_member_tree_sponsor ON ent_member.id = ent_member_tree_sponsor.member_id ").
		Joins("JOIN wod_member_star ON ent_member.id = wod_member_star.member_id ").
		Joins("LEFT JOIN tbl_bonus_diamond_star ON ent_member.id = tbl_bonus_diamond_star.t_downline_id ").
		Group("ent_member_tree_sponsor.sponsor_id").
		Select("ent_member_tree_sponsor.sponsor_id, FLOOR(COUNT(DISTINCT(ent_member.id))/2) AS 'available_dividend_ticket', COUNT(DISTINCT(ent_member.id)) AS 'n_count', GROUP_CONCAT(DISTINCT(ent_member.id)) AS 'a_downline' ")

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

type EntMemberSponsorTreeForNotification struct {
	SponsorNickName string `json:"sponsor_nick_name" gorm:"column:sponsor_nick_name"`
	DownlineList    string `json:"downline_list" gorm:"column:downline_list"`
}

// GetEntMemberSponsorTreeForNotification get member detail for notification with dynamic condition
func GetEntMemberSponsorTreeForNotificationFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberSponsorTreeForNotification, error) {
	var result []*EntMemberSponsorTreeForNotification
	tx := db.Table("ent_member downline")
	tx = tx.Joins(" JOIN ent_member_lot_sponsor AS downline_lot ON downline.id = downline_lot.member_id ").
		Joins("JOIN ent_member_lot_sponsor AS sponsor_lot ON downline_lot.i_lft > sponsor_lot.i_lft AND downline_lot.i_rgt < sponsor_lot.i_rgt ").
		Joins("JOIN ent_member AS sponsor ON sponsor_lot.member_id = sponsor.id ").
		// Select("downline.nick_name, downline.id, downline_lot.i_lvl , downline.d_last_game, downline.created_at "). // for debug
		Select("sponsor.nick_name AS 'sponsor_nick_name' , GROUP_CONCAT(downline.nick_name) AS 'downline_list'").
		Order("IF(downline.d_last_game IS NOT NULL AND downline.d_last_game != '', downline.d_last_game, downline.created_at) ASC")

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

// GetEntMemberID func get member id
func (m *EntMember) GetEntMemberID() int {
	return m.ID
}

// GetAllStatusMemberByUsername get member by nick_name
func GetAllStatusMemberByUsername(userid string) (*EntMember, error) {
	var entMember EntMember

	err := db.Where("nick_name = ? AND status != 'T'", userid).First(&entMember).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if entMember.ID <= 0 {
		return nil, nil
	}

	return &entMember, nil
}

// ExistsMemberByUsername get member by id
func ExistsMemberByUsername(userid string) (bool, error) {
	mem, err := GetAllStatusMemberByUsername(userid)

	if err != nil {
		return false, err
	}

	if mem != nil {
		return true, nil
	}
	return false, nil
}

// GetEntMemberListFn get ent_member data with dynamic condition
func GetEntMemberListFn(arrCond []WhereCondFn, debug bool) ([]*EntMember, error) {
	var entMember []*EntMember
	tx := db.Table("ent_member")

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
	err := tx.Find(&entMember).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return entMember, nil
}

func GetActiveMainAccEntMemberFn(arrCond []WhereCondFn, debug bool) (*EntMember, error) {

	var result EntMember
	entMember, err := GetEntMemberFn(arrCond, "", debug)

	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "MainAccountChecking_sql_db_problem", Data: err}
	}

	if entMember == nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member", Data: err}
	}

	tx := db.Table("ent_member").
		Where("ent_member.status = 'A'").
		Where("ent_member.main_id = " + strconv.Itoa(entMember.MainID) + "").
		Order("created_at ASC")

	if debug {
		tx = tx.Debug()
	}

	err = tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if result.ID <= 0 {
		return nil, nil
	}

	return &result, nil
}

func CheckTagMainAccount(taggedEntMemberID, entMemberID int) error {

	arrCond := make([]WhereCondFn, 0)
	arrCond = append(arrCond,
		WhereCondFn{Condition: " ent_member.id = ? ", CondValue: taggedEntMemberID},
		WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
	)
	taggedEntMember, err := GetActiveMainAccEntMemberFn(arrCond, false)

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if taggedEntMember == nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_tagged_member", Data: err}
	}

	arrCond = make([]WhereCondFn, 0)
	arrCond = append(arrCond,
		WhereCondFn{Condition: " ent_member.id = ? ", CondValue: entMemberID},
		WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
	)
	EntMember, err := GetActiveMainAccEntMemberFn(arrCond, false)
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if EntMember == nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member", Data: err}
	}

	if taggedEntMember.ID != EntMember.ID { // checking on is it under same main account
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "only_can_tag_to_main_account", Data: err}
	}

	if EntMember.ID != taggedEntMemberID { // only can tag with main account
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "only_can_tag_to_main_account", Data: err}
	}
	return nil
}

// GetLatestEntMemberFn get latest ent_member data with dynamic condition
func GetLatestEntMemberFn(arrCond []WhereCondFn, debug bool) (*EntMember, error) {
	var entMember EntMember
	tx := db.Table("ent_member").Order("id DESC")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.First(&entMember).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if entMember.ID <= 0 {
		return nil, nil
	}

	return &entMember, nil
}

// TaggedEntMember struct
type TaggedEntMember struct {
	ID             int    `gorm:"primary_key" json:"id"`
	MainID         int    `json:"main_id"`
	TaggedMemberID int    `json:"tagged_member_id" gorm:"column:tagged_member_id"`
	NickName       string `json:"nick_name"`
	FirstName      string `json:"first_name"`
	TaggedNickName string `json:"tagged_nick_name" gorm:"column:tagged_nick_name"` // A: active | I : inactive | T: terminate | S: suspend

}

// GetTaggedEntMemberListFn get ent_member data with dynamic condition
func GetTaggedEntMemberListFn(arrCond []WhereCondFn, debug bool) ([]*TaggedEntMember, error) {
	var entMember []*TaggedEntMember
	tx := db.Table("ent_member").
		Joins("LEFT JOIN ent_member AS tagged_member_id ON ent_member.tagged_member_id = tagged_member_id.id").
		Select("tagged_member_id.nick_name AS 'tagged_nick_name', ent_member.*")

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
	err := tx.Find(&entMember).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return entMember, nil
}

type TotalActiveMember struct {
	TotalMember int `json:"total_member"`
}

func GetTotalActiveMemberFn(arrCond []WhereCondFn, debug bool) (*TotalActiveMember, error) {
	var result TotalActiveMember
	tx := db.Table("ent_member")
	tx = tx.Select("COUNT(DISTINCT(ent_member.id)) AS 'total_member'")

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
