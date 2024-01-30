package models

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
	"golang.org/x/crypto/bcrypt"
)

// Members struct
type Members struct {
	ID    int    `gorm:"primary_key" json:"id"`
	SubID string `json:"sub_id" gorm:"column:sub_id"`
	// Username      string    `json:"username" gorm:"column:username"`
	Email         string    `json:"email" gorm:"column:email"`
	MobilePrefix  string    `json:"mobile_prefix" gorm:"column:mobile_prefix"`
	MobileNo      string    `json:"mobile_no" gorm:"column:mobile_no"`
	Password      string    `json:"-" gorm:"column:password"`      // hide password when return with json format
	SecondaryPin  string    `json:"-" gorm:"column:secondary_pin"` // hide SecondaryPin when return with json format
	UserTypeID    int       `json:"user_type_id" gorm:"column:user_type_id"`
	UserGroupID   int       `json:"user_group_id" gorm:"column:user_group_id"`
	Status        string    `json:"status"` // A: active | I : inactive | T: terminate | S: suspend
	ResetPassword int       `json:"-"`      // hide password when return with json format
	RememberToken string    `json:"remember_token" gorm:"column:remember_token"`
	ForceLogout   int       `json:"force_logout"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     string    `json:"updated_by"`
}

// ProfileV2 struct
type ProfileV2 struct {
	Members
	ReferralLink string `json:"referral_link"`
}

// AddMember add member
func AddMember(tx *gorm.DB, member Members) (*Members, error) {
	if err := tx.Table("members").Create(&member).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &member, nil
}

// UpdateMember update member
func UpdateMember(userId string, email string, password string, sponsorID int, status string, contact_no string) error {
	//var member Members
	//db.Table("members").Where("email = ? ", email).First(&member)
	if email != "" {
		if password != "" {
			if err := GetDB().Table("members").
				Where("email = ?", email).
				Update("password", password).Update("user_id", userId).Update("status", status).Error; err != nil {
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
			}

			return nil
		}
		if err := GetDB().Table("members").
			Where("email = ?", email).
			Update("status", status).Update("user_id", userId).Error; err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
	}
	if contact_no != "" {
		if password != "" {
			if err := GetDB().Table("members").
				Where("mobile_no = ?", contact_no).
				Update("password", password).Update("user_id", userId).Update("status", status).Error; err != nil {
				return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
			}

			return nil
		}
		if err := GetDB().Table("members").
			Where("mobile_no = ?", contact_no).
			Update("status", status).Update("user_id", userId).Error; err != nil {
			return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
	}

	return nil
}

// GenerateMemberSubID generate sub id for member
func GenerateMemberSubID() (string, error) {
	var count int
	for {
		var mem Members
		id := memberUserType + "-" + uuid.New().String()
		err := db.Select("id").Where("sub_id = ?", id).First(&mem).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		if mem.ID == 0 {
			return id, nil
		}

		if count >= 20 {
			ErrorLog("GenerateMemberSubID", "generate member sub id error", nil)
			return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.GENERATE_MEMBER_SUB_ID_ERROR}
		}
		count++
	}
}

// GetAllStatusMemberByEmail get member by email
// status (when status is empty string means find all status)
func GetAllStatusMemberByEmail(email string) (*Members, error) {
	var member Members

	err := db.Where("email = ? AND status in('A','I')", email).First(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if member.ID <= 0 {
		return nil, nil
	}

	return &member, nil
}

// ExistsMemberByEmail get member by email
func ExistsMemberByEmail(email string) (bool, error) {
	mem, err := GetAllStatusMemberByEmail(email)
	if err != nil {
		return false, err
	}

	if mem != nil {
		return true, nil
	}
	return false, nil
}

// GetDetailByUserID get member by user id
func GetDetailByUserID(userid string) (*Members, error) {
	var member Members

	err := db.Select("members.*, ent_member.sub_id").Joins("JOIN ent_member ON members.ent_member_id = ent_member.id").Where("members.nick_name = ?", userid).First(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if member.ID <= 0 {
		return nil, nil
	}
	return &member, nil
}

// GetMemberByID get member by id
func GetMemberByID(id int) (*Members, error) {
	var member Members
	err := db.Where("id = ?", id).First(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if member.ID <= 0 {
		return nil, nil
	}

	return &member, nil
}

// GetMemberBySubID get member by id
func GetMemberBySubID(id string) (*Members, error) {
	var result Members
	err := db.Where("members.sub_id = ?", id).First(&result).Error
	if err != nil || err == gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

// model functions

// Activate activate member
func (m *Members) Activate(tx *gorm.DB) error {
	m.Status = "A"
	return SaveTx(tx, m)
}

// UpdatePassword update member password
func (m *Members) UpdatePassword(tx *gorm.DB, password string) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.Password = string(pass)

	return SaveTx(tx, m)
}

// functions for token generate

// GetUserType get member token type for access token
func (m *EntMemberMembers) GetUserType() string {
	return memberUserType
}

// GetUserSubID func get member sub id
func (m *EntMemberMembers) GetUserSubID() string {
	return m.SubID
}

// GetHashedPassword get member hashed password
func (m *Members) GetHashedPassword() string {
	return m.Password
}

// GetUserName get member nickname
func (m *EntMemberMembers) GetUserName() string {
	return m.NickName
}

// GetStatusScope get member status scope (for token use) [login]
func (m *EntMemberMembers) GetStatusScope() string {
	return "STAT-" + m.Status
}

// GetAccessScope get member access scope
func (m *EntMemberMembers) GetAccessScope() []string {
	basic := m.GetUserType()
	status := m.GetStatusScope()
	return []string{basic, status}
}

// GetEntMemberID func get member id
func (m *EntMemberMembers) GetMembersID() int {
	return m.ID
}

// GetMembersBySponsorID get member by sponsor id
func GetMembersBySponsorID(sponsorID int) ([]*Members, error) {
	var members []*Members

	err := db.Where("sponsor_id = ? AND status != 'T'", sponsorID).Find(&members).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return members, nil
}

// GetMemberCOM get member by sponsor id
func GetMemberCOM() (*Members, error) {
	var member Members

	err := db.Where("user_id = 'com' AND status = 'A'").First(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.ERROR}
	}

	return &member, nil
}

// MemberTree struct
type MemberTree struct {
	ID            int    `json:"_"`
	MemberID      int    `json:"_"`
	UserID        string `json:"user_id"`
	Parent        int    `json:"_"`
	ParentUserID  string `json:"parent"`
	Avatar        string `json:"avatar"`
	Level         int    `json:"level"`
	DownlineCount int    `json:"downline_count"`
	Status        string `json:"status"`
}

// GetMemberSponsorTree func
func GetMemberSponsorTree(memberID, level int) ([]*MemberTree, error) {
	var (
		tree   []*MemberTree
		member MemberTree
	)

	err := db.Table("members").
		Joins("JOIN geneology_tree on geneology_tree.member_id = members.id").
		Joins("LEFT JOIN members m2 on members.id = m2.sponsor_id").
		Where("geneology_tree.member_id = ? AND geneology_tree.level = ?", memberID, 1).
		Select("members.id, geneology_tree.member_id, members.user_id, members.status, 0 as parent, members.avatar, 0 AS level, count(m2.id) as downline_count").
		Group("members.id").
		First(&member).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	query := db.Table("members").
		Joins("JOIN geneology_tree on geneology_tree.member_id = members.id").
		Joins("JOIN members m2 on m2.id = members.sponsor_id").
		Joins("LEFT JOIN members m3 on members.id = m3.sponsor_id").
		Where("geneology_tree.sponsor_id = ?", memberID)

	if level > 0 {
		query = query.Where("geneology_tree.level <= ?", level)
	}

	err = query.Select("members.id, members.status, geneology_tree.member_id, members.user_id, m2.id as parent, m2.user_id as parent_user_id, members.avatar, geneology_tree.level, COUNT(m3.id) as downline_count").
		Order("geneology_tree.level").
		Order("members.created_at").
		Group("members.id").
		Scan(&tree).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return append([]*MemberTree{&member}, tree...), nil
}

// GetAllMembers get member by sponsor id
func GetAllMembers() ([]*Members, error) {
	var member []*Members

	err := db.Where("status != 'T'").Find(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.ERROR}
	}

	return member, nil
}

// GetAllMembersv2 get member by sponsor id
func GetAllMembersv2() ([]*Members, error) {
	var member []*Members

	err := db.Where("status != 'D'").Find(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.ERROR}
	}

	return member, nil
}

// EntMemberMembers struct
type EntMemberMembers struct {
	ID              int    `json:"id" gorm:"column:id"`
	EntMemberID     int    `json:"ent_member_id" gorm:"column:ent_member_id"`
	MainID          int    `json:"main_id" gorm:"column:main_id"`
	CountryID       int    `json:"country_id" gorm:"column:country_id"`
	NickName        string `json:"nick_name" gorm:"column:nick_name"`
	Status          string `json:"status"`                   // A: active | I : inactive | T: terminate | S: suspend
	Password        string `json:"-" gorm:"column:password"` // hide password when return with json format
	SubID           string `json:"sub_id" gorm:"column:sub_id"`
	MobilePrefix    string `json:"mobile_prefix" gorm:"column:mobile_prefix"`
	MobileNo        string `json:"mobile_no" gorm:"column:mobile_no"`
	Email           string `json:"email" gorm:"column:email"`
	EntMemberStatus string `json:"ent_member_status" gorm:"column:ent_member_status"`
	SecondaryPin    string `json:"-" gorm:"column:secondary_pin"` // hide password when return with json format
	CurrentProfile  int    `json:"current_profile" gorm:"column:current_profile"`
	PrivateKey      string `json:"-" gorm:"column:private_key"`
	Code            string `json:"ent_member_code" gorm:"column:ent_member_code"`
}

// GetEntMemberMemberFn get members and ent_member info with dynamically condition
func GetEntMemberMemberFn(arrCond []WhereCondFn, debug bool) (*EntMemberMembers, error) {
	var result EntMemberMembers

	tx := db.Table("members").
		Select("members.id, ent_member.id AS 'ent_member_id', ent_member.main_id, ent_member.country_id, ent_member.nick_name, ent_member.current_profile, members.status, members.password," +
			"members.sub_id, members.mobile_prefix, members.mobile_no, members.email, ent_member.status AS 'ent_member_status', members.secondary_pin AS 'secondary_pin', ent_member.private_key, ent_member.code AS 'ent_member_code'").
		Joins("INNER JOIN ent_member ON members.id = ent_member.main_id AND ent_member.current_profile = 1")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if debug {
		tx = tx.Debug()
	}

	err := tx.First(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if result.ID <= 0 {
		tx = db.Table("members").
			Select("members.id, ent_member.id AS 'ent_member_id', ent_member.main_id, ent_member.country_id, ent_member.nick_name, ent_member.current_profile, members.status, members.password," +
				"members.sub_id, members.mobile_prefix, members.mobile_no, members.email, ent_member.status AS 'ent_member_status', members.secondary_pin AS 'secondary_pin'").
			Joins("LEFT JOIN ent_member ON members.id = ent_member.main_id AND ent_member.current_profile = 1")

		if len(arrCond) > 0 {
			for _, v := range arrCond {
				tx = tx.Where(v.Condition, v.CondValue)
			}
		}

		if debug {
			tx = tx.Debug()
		}

		err := tx.First(&result).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
	}

	if result.ID <= 0 {
		return nil, nil
	}
	return &result, nil
}

// GetAllStatusMemberByMobile get all status member by mobile
func GetAllStatusMemberByMobile(mobilePrefix, mobileNo string) (*Members, error) {
	var member Members

	err := db.Where("mobile_prefix = ? AND mobile_no = ? AND status in('A','I')", mobilePrefix, mobileNo).First(&member).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if member.ID <= 0 {
		return nil, nil
	}

	return &member, nil
}

// ExistsMemberByMobile get member by email
func ExistsMemberByMobile(mobilePrefix, mobileNo string) (bool, error) {
	mem, err := GetAllStatusMemberByMobile(mobilePrefix, mobileNo)
	if err != nil {
		return false, err
	}

	if mem != nil {
		return true, nil
	}
	return false, nil
}

// GetMembersFn get members info with dynamically condition
func GetMembersFn(arrCond []WhereCondFn, debug bool) (*Members, error) {
	var member Members

	tx := db.Table("members")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if debug {
		tx = tx.Debug()
	}

	err := tx.First(&member).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if member.ID <= 0 {
		return nil, nil
	}
	return &member, nil
}

// GetAdminEntMemberMemberFn get members and ent_member info with dynamically condition
func GetAdminEntMemberMemberFn(arrCond []WhereCondFn, arrData map[string]string, debug bool) (*EntMemberMembers, error) {
	var result EntMemberMembers

	tx := db.Table("members").
		Select("members.id, ent_member.id AS 'ent_member_id', ent_member.main_id, ent_member.country_id, ent_member.nick_name, ent_member.current_profile, members.status, members.password," +
			"members.sub_id, members.mobile_prefix, members.mobile_no, members.email, ent_member.status AS 'ent_member_status', members.secondary_pin AS 'secondary_pin', ent_member.private_key,ent_member.code AS 'ent_member_code'")

	if arrData["member_id"] != "" {
		tx = tx.Joins("INNER JOIN ent_member ON members.id = ent_member.main_id AND ent_member.id = " + arrData["member_id"])
	}
	if arrData["nick_name"] != "" {
		tx = tx.Joins("INNER JOIN ent_member ON members.id = ent_member.main_id AND ent_member.nick_name = '" + arrData["nick_name"] + "'")
	}

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if debug {
		tx = tx.Debug()
	}

	err := tx.First(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if result.ID <= 0 {
		tx = db.Table("members").
			Select("members.id, ent_member.id AS 'ent_member_id', ent_member.main_id, ent_member.country_id, ent_member.nick_name, ent_member.current_profile, members.status, members.password," +
				"members.sub_id, members.mobile_prefix, members.mobile_no, members.email, ent_member.status AS 'ent_member_status', members.secondary_pin AS 'secondary_pin'")

		if arrData["member_id"] != "" {
			tx = tx.Joins("LEFT JOIN ent_member ON members.id = ent_member.main_id AND ent_member.id = " + arrData["member_id"])
		}
		if arrData["nick_name"] != "" {
			tx = tx.Joins("LEFT JOIN ent_member ON members.id = ent_member.main_id AND ent_member.nick_name = '" + arrData["nick_name"] + "'")
		}

		if len(arrCond) > 0 {
			for _, v := range arrCond {
				tx = tx.Where(v.Condition, v.CondValue)
			}
		}

		if debug {
			tx = tx.Debug()
		}

		err := tx.First(&result).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
	}

	if result.ID <= 0 {
		return nil, nil
	}
	return &result, nil
}

// GetUserCode func get member code
func (m *EntMemberMembers) GetUserCode() string {
	return m.Code
}

type CurrentActiveProfileMemberStruct struct {
	SourceID int
}

func GetCurrentActiveProfileMemberFn(arrCond []WhereCondFn, arrData CurrentActiveProfileMemberStruct, debug bool) (*EntMemberMembers, error) {
	var result EntMemberMembers

	tx := db.Table("members").
		Select("members.id, ent_member.id AS 'ent_member_id', ent_member.main_id, ent_member.country_id, ent_member.nick_name, ent_member.current_profile, members.status, members.password," +
			"members.sub_id, members.mobile_prefix, members.mobile_no, members.email, ent_member.status AS 'ent_member_status', members.secondary_pin AS 'secondary_pin', ent_member.private_key, ent_member.code AS 'ent_member_code'").
		Joins("INNER JOIN ent_member ON members.id = ent_member.main_id").
		Joins("INNER JOIN ent_current_profile ON ent_member.main_id = ent_current_profile.main_id AND ent_member.id = ent_current_profile.member_id AND ent_current_profile.source_id = " + strconv.Itoa(arrData.SourceID))

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if debug {
		tx = tx.Debug()
	}

	err := tx.First(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	crtEntCurrentProfile := false
	if result.ID < 1 {
		tx = db.Table("members").
			Select("members.id, ent_member.id AS 'ent_member_id', ent_member.main_id, ent_member.country_id, ent_member.nick_name, ent_member.current_profile, members.status, members.password," +
				"members.sub_id, members.mobile_prefix, members.mobile_no, members.email, ent_member.status AS 'ent_member_status', members.secondary_pin AS 'secondary_pin', ent_member.private_key, ent_member.code AS 'ent_member_code'").
			Joins("LEFT JOIN ent_member ON members.id = ent_member.main_id AND ent_member.current_profile = 1")

		if len(arrCond) > 0 {
			for _, v := range arrCond {
				tx = tx.Where(v.Condition, v.CondValue)
			}
		}

		if debug {
			tx = tx.Debug()
		}

		err := tx.First(&result).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}
		crtEntCurrentProfile = true
	}

	if result.ID <= 0 {
		return nil, nil
	}

	if crtEntCurrentProfile {
		arrCond := make([]WhereCondFn, 0)
		arrCond = append(arrCond,
			WhereCondFn{Condition: "source_id = ?", CondValue: arrData.SourceID},
			WhereCondFn{Condition: "main_id = ?", CondValue: result.ID},
		)
		arrEntCurrentProfile, _ := GetEntCurrentProfileFn(arrCond, false)

		if len(arrEntCurrentProfile) < 1 { // no existing records is existed.
			// start perform save new records
			arrCrtData := AddEntCurrentProfileStruct{
				SourceID: arrData.SourceID,
				MainID:   result.ID,
				MemberID: result.EntMemberID,
			}

			AddEntCurrentProfile(GetDB(), arrCrtData)
			// end perform save new records
		} else {
			// start perform update prev records

			updateColumn := map[string]interface{}{}
			updateColumn["member_id"] = result.EntMemberID

			arrCond := make([]WhereCondFn, 0)
			arrCond = append(arrCond,
				WhereCondFn{Condition: " source_id = ? ", CondValue: arrData.SourceID},
				WhereCondFn{Condition: " main_id = ? ", CondValue: result.ID},
			)
			_ = UpdatesFn("ent_current_profile", arrCond, updateColumn, false)
			// end perform update prev records
		}
	}

	return &result, nil
}
