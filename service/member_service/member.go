package member_service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/pkg/setting"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/job_service"
	"github.com/smartblock/gta-api/service/mobile_service"

	"github.com/jinzhu/gorm"
)

// Member struct
type Member struct {
	// Username, Email, MobilePrefix, MobileNo, Password, SecondaryPin, LangCode string
	Email, MobilePrefix, MobileNo, Password, SecondaryPin, LangCode string
	ReferralID                                                      int
}

// Add member func
func (m *Member) Add(tx *gorm.DB) (string, int) {
	var (
		ok  bool
		err error
	)

	// passsword checking
	ok = base.PasswordChecking(m.Password)
	if !ok {
		return e.GetMsg(e.PASSWORD_VALIDATION_ERROR), 0
	}

	// seconday pin checking
	ok = base.SecondaryPinChecking(m.SecondaryPin)
	if !ok {
		return e.GetMsg(e.SECONDARY_PIN_VALIDATION_ERROR), 0
	}

	// encrypt password
	password, err := base.Bcrypt(m.Password)
	if err != nil {
		base.LogErrorLog("memberService:Add()", "Bcrypt():1", err.Error(), true)
		return "something_went_wrong", 0
	}

	// encrypt secondary pin
	secondaryPin := util.EncodeMD5(m.SecondaryPin)

	// generate sub id
	subID, err := models.GenerateMemberSubID()
	if err != nil {
		base.LogErrorLog("memberService:Add()", "GenerateMemberSubID():1", err.Error(), true)
		return "something_went_wrong", 0
	}

	// add member
	arrMemberFn := models.Members{
		SubID:       subID,
		UserTypeID:  66,
		UserGroupID: 3,
		// Username:     m.Username,
		Email:        m.Email,
		MobilePrefix: m.MobilePrefix,
		MobileNo:     m.MobileNo,
		Status:       "A",
		Password:     password,
		SecondaryPin: secondaryPin,
	}

	arrMember, err := models.AddMember(tx, arrMemberFn)
	if err != nil {
		base.LogErrorLog("memberService:Add()", "AddMember():1", err.Error(), true)
		return "something_went_wrong", 0
	}

	// update members.created_by
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: arrMember.ID})
	updateColumn := map[string]interface{}{"created_by": arrMember.ID, "updated_by": arrMember.ID}
	err = models.UpdatesFnTx(tx, "members", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:Add()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong", 0
	}

	// reserve referral_code for this login's first profile
	arrEntMemberReservedSponsorFn := models.EntMemberReservedSponsor{
		MemberID:  arrMember.ID,
		SponsorID: m.ReferralID,
	}

	_, err = models.AddEntMemberReservedSponsor(tx, arrEntMemberReservedSponsorFn)
	if err != nil {
		base.LogErrorLog("memberService:Add()", "AddEntMemberReservedSponsor():1", err.Error(), true)
		return "something_went_wrong", 0
	}

	return "", arrMember.ID
}

// EntMember struct
type EntMember struct {
	MainID, CountryID                                   int
	Username, FirstName, ReferralCode, LangCode, Source string
}

// CreateProfile create account func
func (m *EntMember) CreateProfile(tx *gorm.DB) (errReturnMsg string, returnEntMemberID int) {
	var (
		err                                     error
		errMsg                                  string
		referralID, currentProfile, entMemberID int
	)
	// check username format
	username := strings.Trim(m.Username, " ")
	errMsg = base.UsernameChecking(username)
	if errMsg != "" {
		return errMsg, 0
	}

	// check if first profile
	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: m.MainID},
	)
	arrEntMember, err := models.GetFirstEntMemberFn(arrEntMemberFn, "", false)

	if err != nil {
		base.LogErrorLog("memberService:CreateProfile()", "GetFirstEntMemberFn():1", err.Error(), true)
		return "something_went_wrong", 0
	}

	if arrEntMember != nil && arrEntMember.Status != "I" && m.ReferralCode == "" {
		return "please_enter_referral_code", 0
	}

	if arrEntMember != nil && arrEntMember.Status == "I" && arrEntMember.NickName == username {
		// first profile already established before and nick_name or network does not require update
		return "", arrEntMember.ID
	}

	// validate if username is unique
	ok, err := ExistsMemberByUsername(username)
	if err != nil {
		base.LogErrorLog("memberService:CreateProfile()", "ExistsMemberByUsername():1", err.Error(), true)
		return "something_went_wrong", 0
	}

	if ok {
		// if not first profile
		if arrEntMember != nil && arrEntMember.Status != "I" {
			// get lastest created profile
			arrInactiveEntMemberFn := make([]models.WhereCondFn, 0)
			arrInactiveEntMemberFn = append(arrInactiveEntMemberFn,
				models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: m.MainID},
				models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "I"},
			)
			arrInactiveEntMember, err := models.GetEntMemberFn(arrInactiveEntMemberFn, "", false)
			if err != nil {
				base.LogErrorLog("memberService:CreateProfile()", "GetEntMemberFn():1", err.Error(), true)
				return "something_went_wrong", 0
			}

			// if latest created profile exist but not same username then return error
			if arrInactiveEntMember == nil || arrInactiveEntMember.NickName != username {
				return e.GetMsg(e.MEMBER_USERNAME_ALREADY_EXISTS), 0
			}
		} else {
			return e.GetMsg(e.MEMBER_USERNAME_ALREADY_EXISTS), 0
		}
	}

	if arrEntMember == nil || arrEntMember.Status != "I" {
		// establishing first profile or new profile
		referralID = 1 // default referral set to com

		if arrEntMember == nil { // if first profile
			// take reserved sponsor
			arrEntMemberReservedSponsorFn := make([]models.WhereCondFn, 0)
			arrEntMemberReservedSponsorFn = append(arrEntMemberReservedSponsorFn,
				models.WhereCondFn{Condition: "ent_member_reserved_sponsor.member_id = ?", CondValue: m.MainID},
			)
			arrEntMemberReservedSponsor, err := models.GetEntMemberReservedSponsorFn(arrEntMemberReservedSponsorFn, false)

			if err != nil {
				base.LogErrorLog("memberService:CreateProfile()", "GetEntMemberReservedSponsorFn():1", err.Error(), true)
				return "something_went_wrong", 0
			}

			if arrEntMemberReservedSponsor == nil {
				base.LogErrorLog("memberService:CreateProfile()", "reserved_sponsor_not_found_for_first_profile", "", true)
				return "something_went_wrong", 0
			}

			referralID = arrEntMemberReservedSponsor.SponsorID

			currentProfile = 1
		} else { // if no then take referralCode from input param
			// validate referral code
			if m.ReferralCode != "" {
				referralCode := strings.Trim(m.ReferralCode, " ")
				errMsg, entMemberSponsor := ValidateReferralCode(referralCode)

				if errMsg != "" {
					return errMsg, 0
				}

				referralID = entMemberSponsor.ID
			}
		}

		// get cur date
		entMemberID = 0
		curDate, err := base.GetCurrentTimeV2("yyyy-mm-dd")
		if err != nil {
			base.LogErrorLog("memberService:CreateProfile()", "GetCurrentTimeV2():1", err.Error(), true)
			return "something_went_wrong", 0
		}

		// get lastest created profile
		arrInactiveEntMemberFn := make([]models.WhereCondFn, 0)
		arrInactiveEntMemberFn = append(arrInactiveEntMemberFn,
			models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: m.MainID},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "I"},
		)
		arrInactiveEntMember, err := models.GetEntMemberFn(arrInactiveEntMemberFn, "", false)
		if err != nil {
			base.LogErrorLog("memberService:CreateProfile()", "GetEntMemberFn():1", err.Error(), true)
			return "something_went_wrong", 0
		}

		// if not first profile and checked got wasted ent_member.status = "I" slot, update instead of addEntMember
		if arrEntMember != nil && arrInactiveEntMember != nil {
			entMemberID = arrInactiveEntMember.ID

			// update ent_member.nick_name
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: entMemberID})
			updateColumn := map[string]interface{}{"nick_name": username, "updated_by": entMemberID}
			err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
			if err != nil {
				base.LogErrorLog("memberService:CreateProfile()", "UpdatesFnTx():2", err.Error(), true)
				return "something_went_wrong", 0
			}

			// delete sponsor tree
			arrEntMemberTreeSponsorDelFn := make([]models.WhereCondFn, 0)
			arrEntMemberTreeSponsorDelFn = append(arrEntMemberTreeSponsorDelFn,
				models.WhereCondFn{Condition: "member_id = ?", CondValue: entMemberID},
			)
			err = models.DeleteFn("ent_member_tree_sponsor", arrEntMemberTreeSponsorDelFn, false)
			if err != nil {
				base.LogErrorLog("memberService:CreateProfile()", "DeleteFn():1", err.Error(), true)
				return "something_went_wrong", 0
			}

		} else { // else create new profile
			memCode := GenRandomMemberCode()
			// memCode := "knjgntipxc" // for debug purposes
			arrAddEntMemberFn := models.AddEntMemberStruct{
				CountryID:      m.CountryID,
				CompanyID:      1,
				MainID:         m.MainID,
				MemberType:     "MEM",
				Source:         "APP",
				NickName:       username,
				Code:           memCode,
				CurrentProfile: currentProfile,
				Status:         "I",
				JoinDate:       curDate,
				// Avatar:             fmt.Sprintf("https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar%d.jpg", rand.Perm(10)[0]),
				PreferLanguageCode: m.LangCode,
			}

			entMemberFn, err := models.AddEntMember(tx, arrAddEntMemberFn)
			if err != nil {
				if strings.Contains(err.Error(), "Duplicate entry '"+memCode+"' for key 'code'") {
					for {
						memCode := GenRandomMemberCode()
						arrAddEntMemberFn.Code = memCode
						entMemberFn, err = models.AddEntMember(tx, arrAddEntMemberFn)
						if err == nil {
							break
						}
					}
				} else {
					base.LogErrorLog("memberService:CreateProfile()", "AddEntMember():1", err.Error(), true)
					return "something_went_wrong", 0
				}
			}

			entMemberID = entMemberFn.ID

			// update ent_member.created_by
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: entMemberID})
			updateColumn := map[string]interface{}{"created_by": entMemberID, "updated_by": entMemberID}
			err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
			if err != nil {
				base.LogErrorLog("memberService:CreateProfile()", "UpdatesFnTx():1", err.Error(), true)
				return "something_went_wrong", 0
			}
		}

		// establish tree network
		arrEntMemberTreeSponsorFn := models.EntMemberTreeSponsor{
			MemberLot:  "01",
			UplineID:   referralID,
			UplineLot:  "01",
			SponsorID:  referralID,
			SponsorLot: "01",
			Lvl:        2,
			CreatedBy:  1,
		}

		if referralID != 1 { // referral network validation
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: referralID},
				models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
			)

			sponsor, err := models.GetEntMemberEntMemberTreeSponsorFn(arrCond, false)

			if err != nil {
				base.LogErrorLog("memberService:CreateProfile()", "GetEntMemberEntMemberTreeSponsorFn():1", err.Error(), true)
				return "something_went_wrong", 0
			}

			if sponsor == nil {
				base.LogErrorLog("memberService:CreateProfile()", "referral_network_not_found", "", true)
				return "something_went_wrong", 0
			}

			arrEntMemberTreeSponsorFn.UplineID = sponsor.MemberID
			arrEntMemberTreeSponsorFn.SponsorID = sponsor.MemberID
			arrEntMemberTreeSponsorFn.Lvl = sponsor.Lvl + 1
		}

		arrEntMemberTreeSponsorFn.MemberID = entMemberID
		arrEntMemberTreeSponsorFn.CreatedBy = entMemberID

		_, err = models.AddEntMemberTreeSponsor(tx, arrEntMemberTreeSponsorFn)
		if err != nil {
			base.LogErrorLog("memberService:CreateProfile()", "AddEntMemberTreeSponsor():1", err.Error(), true)
			return "something_went_wrong", 0
		}
		curDateTimeString := base.GetCurrentTime("2006-01-02 15:04:05") // correct vers.
		arrAddEntMemberLotQueue := models.AddEntMemberLotQueueStruct{
			MemberID:   entMemberID,
			MemberLot:  "01",
			SponsorID:  arrEntMemberTreeSponsorFn.SponsorID,
			SponsorLot: "01",
			UplineID:   arrEntMemberTreeSponsorFn.UplineID,
			UplineLot:  "01",
			Type:       "REG",
			DtCreate:   curDateTimeString,
		}

		_, err = models.AddEntMemberLotQueue(tx, arrAddEntMemberLotQueue)
		if err != nil {
			base.LogErrorLog("memberService:CreateProfile()", "AddEntMemberLotQueue():1", err.Error(), true)
			return "something_went_wrong", 0
		}

		go job_service.ProcessTreeQJobService()
	} else { // change inactive established first profile username
		// mainID := arrEntMember.MainID
		entMemberID := arrEntMember.ID

		// update ent_member.nick_name
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: entMemberID})
		updateColumn := map[string]interface{}{"nick_name": username, "updated_by": entMemberID}
		err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
		if err != nil {
			base.LogErrorLog("memberService:CreateProfile()", "UpdatesFnTx():2", err.Error(), true)
			return "something_went_wrong", 0
		}

		// update member.username
		// arrUpdCond2 := make([]models.WhereCondFn, 0)
		// arrUpdCond2 = append(arrUpdCond2, models.WhereCondFn{Condition: "id = ?", CondValue: mainID})
		// updateColumn2 := map[string]interface{}{"username": username, "updated_by": mainID}
		// err = models.UpdatesFnTx(tx, "members", arrUpdCond2, updateColumn2, false)
		// if err != nil {
		// 	base.LogErrorLog("memberService:CreateProfile()", "UpdatesFnTx():3", err.Error(), true)
		// 	return "something_went_wrong"
		// }

	}

	return "", entMemberID
}

// CreateProfilev2 create account func
func (m *EntMember) CreateProfilev2(tx *gorm.DB) (errReturnMsg string, returnEntMemberID int) {

	if m.Source == "" {
		m.Source = "APP"
	}

	var (
		err                                                         error
		errMsg                                                      string
		referralID, placementID, legNo, currentProfile, entMemberID int
	)
	// check username format
	username := strings.Trim(m.Username, " ")
	errMsg = base.UsernameChecking(username)
	if errMsg != "" {
		return errMsg, 0
	}

	// check if first profile
	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: m.MainID},
	)
	arrEntMember, err := models.GetFirstEntMemberFn(arrEntMemberFn, "", false)

	if err != nil {
		base.LogErrorLog("CreateProfilev2-GetFirstEntMemberFn_failed", err.Error(), arrEntMemberFn, true)
		return "something_went_wrong", 0
	}

	if arrEntMember != nil && arrEntMember.Status != "I" && m.ReferralCode == "" {
		return "please_enter_referral_code", 0
	}

	if arrEntMember != nil && arrEntMember.Status == "I" && arrEntMember.NickName == username {
		// first profile already established before and nick_name or network does not require update
		return "", arrEntMember.ID
	}

	// validate if username is unique
	ok, err := ExistsMemberByUsername(username)
	if err != nil {
		base.LogErrorLog("CreateProfilev2-ExistsMemberByUsername_failed", err.Error(), username, true)
		return "something_went_wrong", 0
	}

	if ok {
		// if not first profile
		if arrEntMember != nil && arrEntMember.Status != "I" {
			// get lastest created profile
			arrInactiveEntMemberFn := make([]models.WhereCondFn, 0)
			arrInactiveEntMemberFn = append(arrInactiveEntMemberFn,
				models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: m.MainID},
				models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "I"},
			)
			arrInactiveEntMember, err := models.GetEntMemberFn(arrInactiveEntMemberFn, "", false)
			if err != nil {
				base.LogErrorLog("CreateProfilev2-GetEntMemberFn_failed", err.Error(), arrInactiveEntMemberFn, true)
				return "something_went_wrong", 0
			}

			// if latest created profile exist but not same username then return error
			if arrInactiveEntMember == nil || arrInactiveEntMember.NickName != username {
				return e.GetMsg(e.MEMBER_USERNAME_ALREADY_EXISTS), 0
			}
		} else {
			return e.GetMsg(e.MEMBER_USERNAME_ALREADY_EXISTS), 0
		}
	}

	if arrEntMember == nil || arrEntMember.Status != "I" {
		// establishing first profile or new profile
		referralID = 1  // default referral set to com
		placementID = 0 // default no placement
		legNo = 0

		if arrEntMember == nil { // if first profile
			// take reserved sponsor
			arrEntMemberReservedSponsorFn := make([]models.WhereCondFn, 0)
			arrEntMemberReservedSponsorFn = append(arrEntMemberReservedSponsorFn,
				models.WhereCondFn{Condition: "ent_member_reserved_sponsor.member_id = ?", CondValue: m.MainID},
			)
			arrEntMemberReservedSponsor, err := models.GetEntMemberReservedSponsorFn(arrEntMemberReservedSponsorFn, false)

			if err != nil {
				base.LogErrorLog("CreateProfilev2-GetEntMemberReservedSponsorFn_failed", err.Error(), arrEntMemberReservedSponsorFn, true)
				return "something_went_wrong", 0
			}

			if arrEntMemberReservedSponsor == nil {
				base.LogErrorLog("CreateProfilev2-arrEntMemberReservedSponsor_null", "reserved_sponsor_not_found_for_first_profile", arrEntMemberReservedSponsorFn, true)
				return "something_went_wrong", 0
			}

			referralID = arrEntMemberReservedSponsor.SponsorID
			placementID = arrEntMemberReservedSponsor.UplineID
			legNo = arrEntMemberReservedSponsor.LegNo

			currentProfile = 1
		} else { // if no then take referralCode from input param
			// validate referral code
			if m.ReferralCode != "" {
				referralCode := strings.Trim(m.ReferralCode, " ")
				errMsg, entMemberSponsor := ValidateReferralCode(referralCode)

				if errMsg != "" {
					return errMsg, 0
				}

				referralID = entMemberSponsor.ID
			}
		}

		// get cur date
		entMemberID = 0
		curDate, err := base.GetCurrentTimeV2("yyyy-mm-dd")
		if err != nil {
			base.LogErrorLog("CreateProfilev2-GetCurrentTimeV2", err.Error(), "GetCurrentTimeV2():1", true)
			return "something_went_wrong", 0
		}

		// get lastest created profile
		arrInactiveEntMemberFn := make([]models.WhereCondFn, 0)
		arrInactiveEntMemberFn = append(arrInactiveEntMemberFn,
			models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: m.MainID},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "I"},
		)
		arrInactiveEntMember, err := models.GetEntMemberFn(arrInactiveEntMemberFn, "", false)
		if err != nil {
			base.LogErrorLog("CreateProfilev2-GetEntMemberFn(get lastest created profile)", err.Error(), arrInactiveEntMemberFn, true)
			return "something_went_wrong", 0
		}

		// if not first profile and checked got wasted ent_member.status = "I" slot, update instead of addEntMember
		if arrEntMember != nil && arrInactiveEntMember != nil {
			entMemberID = arrInactiveEntMember.ID

			// update ent_member.nick_name
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: entMemberID})
			updateColumn := map[string]interface{}{"nick_name": username, "updated_by": entMemberID}
			err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
			if err != nil {
				base.LogErrorLog("CreateProfilev2-UpdatesFnTx_ent_member_failed", err.Error(), map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, true)
				return "something_went_wrong", 0
			}

			// delete sponsor tree
			arrEntMemberTreeSponsorDelFn := make([]models.WhereCondFn, 0)
			arrEntMemberTreeSponsorDelFn = append(arrEntMemberTreeSponsorDelFn,
				models.WhereCondFn{Condition: "member_id = ?", CondValue: entMemberID},
			)
			err = models.DeleteFn("ent_member_tree_sponsor", arrEntMemberTreeSponsorDelFn, false)
			if err != nil {
				base.LogErrorLog("CreateProfilev2-DeleteFn_ent_member_tree_sponsor", err.Error(), arrEntMemberTreeSponsorDelFn, true)
				return "something_went_wrong", 0
			}

		} else { // else create new profile
			memCode := GenRandomMemberCode()
			// memCode := "knjgntipxc" // for debug purposes
			arrAddEntMemberFn := models.AddEntMemberStruct{
				CountryID:      m.CountryID,
				CompanyID:      1,
				MainID:         m.MainID,
				MemberType:     "MEM",
				Source:         m.Source,
				NickName:       username,
				FirstName:      m.FirstName,
				Code:           memCode,
				CurrentProfile: currentProfile,
				Status:         "A",
				JoinDate:       curDate,
				// Avatar:             fmt.Sprintf("https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar%d.jpg", rand.Perm(10)[0]),
				PreferLanguageCode: m.LangCode,
			}

			entMemberFn, err := models.AddEntMember(tx, arrAddEntMemberFn)
			if err != nil {
				if strings.Contains(err.Error(), "Duplicate entry '"+memCode+"' for key 'code'") {
					for {
						memCode := GenRandomMemberCode()
						arrAddEntMemberFn.Code = memCode
						entMemberFn, err = models.AddEntMember(tx, arrAddEntMemberFn)
						if err == nil {
							break
						}
					}
				} else {
					base.LogErrorLog("CreateProfilev2-duplicate_member_code", err.Error(), arrAddEntMemberFn, true)
					return "something_went_wrong", 0
				}
			}

			entMemberID = entMemberFn.ID

			// update ent_member.created_by
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: entMemberID})
			updateColumn := map[string]interface{}{"created_by": entMemberID, "updated_by": entMemberID}
			err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
			if err != nil {
				base.LogErrorLog("CreateProfilev2-update_latest_ent_member", err.Error(), map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, true)
				return "something_went_wrong", 0
			}
		}

		// establish tree network
		arrEntMemberTreeSponsorFn := models.EntMemberTreeSponsor{
			MemberLot:  "01",
			SponsorID:  referralID,
			SponsorLot: "01",
			UplineID:   placementID,
			UplineLot:  "01",
			LegNo:      legNo,
			Lvl:        2,
			CreatedBy:  1,
		}

		if referralID != 1 { // referral network validation
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: referralID},
				models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
			)

			sponsor, err := models.GetEntMemberEntMemberTreeSponsorFn(arrCond, false)

			if err != nil {
				base.LogErrorLog("CreateProfilev2-GetEntMemberEntMemberTreeSponsorFn_failed", err.Error(), arrCond, true)
				return "something_went_wrong", 0
			}

			if sponsor == nil {
				base.LogErrorLog("CreateProfilev2-sponsor_null", "referral_network_not_found", arrCond, true)
				return "something_went_wrong", 0
			}

			// for gta project, placement will set onward, so default is still 0
			// if placementID == 0 {
			// 	arrEntMemberTreeSponsorFn.UplineID = sponsor.MemberID
			// }

			arrEntMemberTreeSponsorFn.SponsorID = sponsor.MemberID
			arrEntMemberTreeSponsorFn.Lvl = sponsor.Lvl + 1
		}

		arrEntMemberTreeSponsorFn.MemberID = entMemberID
		arrEntMemberTreeSponsorFn.CreatedBy = entMemberID

		_, err = models.AddEntMemberTreeSponsor(tx, arrEntMemberTreeSponsorFn)
		if err != nil {
			base.LogErrorLog("CreateProfilev2-AddEntMemberTreeSponsor_failed", err.Error(), arrEntMemberTreeSponsorFn, true)
			return "something_went_wrong", 0
		}
		curDateTimeString := base.GetCurrentTime("2006-01-02 15:04:05") // correct vers.
		arrAddEntMemberLotQueue := models.AddEntMemberLotQueueStruct{
			MemberID:   entMemberID,
			MemberLot:  "01",
			SponsorID:  arrEntMemberTreeSponsorFn.SponsorID,
			SponsorLot: "01",
			UplineID:   arrEntMemberTreeSponsorFn.UplineID,
			UplineLot:  "01",
			Type:       "REG",
			DtCreate:   curDateTimeString,
		}

		_, err = models.AddEntMemberLotQueue(tx, arrAddEntMemberLotQueue)
		if err != nil {
			base.LogErrorLog("CreateProfilev2-AddEntMemberLotQueue_failed", err.Error(), arrAddEntMemberLotQueue, true)
			return "something_went_wrong", 0
		}

		go job_service.ProcessTreeQJobService()
	} else { // change inactive established first profile username
		// mainID := arrEntMember.MainID
		entMemberID := arrEntMember.ID

		// update ent_member.nick_name
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: entMemberID})
		updateColumn := map[string]interface{}{"nick_name": username, "updated_by": entMemberID}
		err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
		if err != nil {
			base.LogErrorLog("CreateProfilev2-UpdatesFnTx_ent_member_nick_name", err.Error(), map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, true)
			return "something_went_wrong", 0
		}

		// update member.username
		// arrUpdCond2 := make([]models.WhereCondFn, 0)
		// arrUpdCond2 = append(arrUpdCond2, models.WhereCondFn{Condition: "id = ?", CondValue: mainID})
		// updateColumn2 := map[string]interface{}{"username": username, "updated_by": mainID}
		// err = models.UpdatesFnTx(tx, "members", arrUpdCond2, updateColumn2, false)
		// if err != nil {
		// 	base.LogErrorLog("memberService:CreateProfile()", "UpdatesFnTx():3", err.Error(), true)
		// 	return "something_went_wrong"
		// }

	}

	return "", entMemberID
}

// Profile struct
type Profile struct {
	Username                string `json:"username"`
	Nickname                string `json:"nick_name"`
	ReferralCode            string `json:"referral_code"`
	PlacementCode           string `json:"placement_code"`
	FirstName               string `json:"first_name"`
	ReferralName            string `json:"referral_name"`
	PlacementName           string `json:"placement_name"`
	PlacementLeg            string `json:"placement_leg"`
	HideInfo                int    `json:"hide_info"`
	CountryCode             string `json:"country_code"`
	Country                 string `json:"country"`
	MobilePrefix            string `json:"mobile_prefix"`
	MobileNo                string `json:"mobile_no"`
	Email                   string `json:"email"`
	Gender                  string `json:"gender"`
	GenderCode              string `json:"gender_code"`
	BirthDate               string `json:"birth_date"`
	JoinDate                string `json:"join_date"`
	CryptoAddr              string `json:"crypto_address"`
	KycStatusCode           string `json:"kyc_status_code"`
	KycStatusDesc           string `json:"kyc_status_desc"`
	InvitationInfo          string `json:"invitation_info"`
	InvitationPlacementInfo string `json:"invitation_placement_info"`
	ProfileImgURL           string `json:"profile_img_url"`
	ShareBtnInfo            string `json:"share_button_info"`
	EncryptedKey            string `json:"encrypted_id"`
	TotalSponsor            string `json:"total_sponsor"`
}

// GetProfile func
func GetProfile(username, langCode string) (*Profile, string) {
	var (
		profile Profile
	)

	// find member by username
	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: username},
	)
	arrEntMember, err := models.GetEntMemberFn(arrEntMemberFn, "", false)

	if err != nil {
		base.LogErrorLog("memberService:GetProfile()", "GetEntMemberFn():1", err.Error(), true)
		return nil, "something_went_wrong"
	}
	if arrEntMember == nil {
		return nil, e.GetMsg(e.INVALID_MEMBER)
	}

	// find member by ent_member.main_id
	arrMembersFn := make([]models.WhereCondFn, 0)
	arrMembersFn = append(arrMembersFn,
		models.WhereCondFn{Condition: "members.id = ?", CondValue: arrEntMember.MainID},
	)
	arrMembers, err := models.GetMembersFn(arrMembersFn, false)

	if err != nil {
		base.LogErrorLog("memberService:GetProfile()", "GetMembersFn():1", err.Error(), true)
		return nil, "something_went_wrong"
	}
	if arrMembers == nil {
		return nil, e.GetMsg(e.INVALID_MEMBER)
	}

	countryID := arrEntMember.CountryID

	// find member country
	arrSysTerritoryFn := make([]models.WhereCondFn, 0)
	arrSysTerritoryFn = append(arrSysTerritoryFn,
		models.WhereCondFn{Condition: "sys_territory.id = ?", CondValue: countryID},
	)
	arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

	if err != nil {
		base.LogErrorLog("memberService:GetProfile()", "GetSysTerritoryFn():1", err.Error(), true)
		return nil, "something_went_wrong"
	}
	if arrSysTerritory != nil {
		profile.CountryCode = arrSysTerritory.Code
		profile.Country = helpers.Translate(arrSysTerritory.Name, langCode)
	}

	// get referral name
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: arrEntMember.ID},
	)

	sponsor, err := models.GetMemberSponsorFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("memberService:GetProfile()", "GetMemberSponsor():1", err.Error(), true)
		return nil, "something_went_wrong"
	}

	if sponsor != nil {
		profile.ReferralName = sponsor.SponsorUsername
		profile.PlacementName = sponsor.UplineUsername
		profile.PlacementLeg = helpers.TranslateV2("leg_:0", langCode, map[string]string{"0": strconv.Itoa(sponsor.LegNo)})
	}

	//add by kahhou
	cryptoAddr := ""
	address, err := models.GetMemberCryptoByMemID(arrEntMember.ID, "")

	if err != nil {
		base.LogErrorLog("memberService:GetProfile()", "GetMemberCryptoByMemID():1", err.Error(), true)
		return nil, "something_went_wrong"
	}
	//end

	if address != nil {
		cryptoAddr = address.CryptoAddress
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: arrEntMember.ID},
	)
	arrEntMemberKyc, _ := models.GetEntMemberKycFn(arrCond, false)

	if len(arrEntMemberKyc) > 0 {
		if arrEntMemberKyc[0].Status == "P" {
			profile.KycStatusCode = arrEntMemberKyc[0].Status
			profile.KycStatusDesc = helpers.Translate("pending", langCode)
		} else if arrEntMemberKyc[0].Status == "AP" {
			profile.KycStatusCode = arrEntMemberKyc[0].Status
			profile.KycStatusDesc = helpers.Translate("approved", langCode)
		} else if arrEntMemberKyc[0].Status == "R" {
			profile.KycStatusCode = arrEntMemberKyc[0].Status
			profile.KycStatusDesc = helpers.Translate("rejected", langCode)
		}
	}

	// get total direct sponsor
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: arrEntMember.ID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	totalDirectSponsorRst, _ := models.GetTotalDirectSponsorFn(arrCond, false)
	totalDirectSponsor := "0"
	if totalDirectSponsorRst.TotalDirectSponsor > 0 {
		totalDirectSponsor = helpers.CutOffDecimal(totalDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
	}

	hideInfo := 0
	if arrEntMember.Source == "addr" {
		hideInfo = 1
	}

	// get current nft tier
	// db := models.GetDB() // no need set begin transaction
	// tier, err := GetMemberNftTier(db, arrEntMember.ID)
	// if err != nil {
	// 	return &profile, err.Error()
	// }

	// fmt.Println(tier)

	// get available leg
	placementCode := arrEntMember.NickName
	placement := GetPlacementLegOption(arrEntMember.NickName, langCode)
	if len(placement) > 0 {
		placementCode = fmt.Sprintf("%s-%d", placementCode, placement[0]["value"])
	}
	profile.Nickname = arrEntMember.NickName
	profile.Username = arrEntMember.NickName
	profile.ReferralCode = arrEntMember.NickName
	profile.PlacementCode = placementCode
	// profile.ReferralCode = arrEntMember.Code
	// profile.ReferralCode = util.EncodeBase64(arrEntMember.NickName)
	// profile.InvitationInfo = util.EncodeBase64(arrEntMember.NickName)
	profile.HideInfo = hideInfo
	profile.FirstName = arrEntMember.FirstName
	profile.MobilePrefix = "+" + arrMembers.MobilePrefix
	profile.MobileNo = arrMembers.MobileNo
	profile.Email = arrMembers.Email
	profile.JoinDate = arrEntMember.JoinDate
	profile.CryptoAddr = cryptoAddr
	profile.GenderCode = arrEntMember.GenderCode
	profile.BirthDate = arrEntMember.BirthDate
	profile.ProfileImgURL = arrEntMember.Avatar

	if arrEntMember.Gender != "" {
		profile.Gender = helpers.Translate(arrEntMember.Gender, langCode)
	}
	serverDomain := setting.Cfg.Section("custom").Key("MemberServerDomain").String()
	// url := serverDomain + "/register?r=" + arrEntMember.NickName
	// url := serverDomain + "/download-mobile-app"
	url := serverDomain
	arrShareInfoData := map[string]string{}
	arrShareInfoData["mobile_download_url"] = url
	arrShareInfoData["user_referral_code"] = arrEntMember.NickName
	profile.ShareBtnInfo = helpers.TranslateV2("share_info_msg", langCode, arrShareInfoData)
	profile.InvitationInfo = url + "?r=" + arrEntMember.NickName
	profile.InvitationPlacementInfo = url + "?p=" + arrEntMember.NickName
	profile.EncryptedKey = helpers.GetEncryptedID(arrEntMember.Code, arrEntMember.ID)
	profile.TotalSponsor = totalDirectSponsor
	return &profile, ""
}

// MemberMobile struct
type MemberMobile struct {
	MemberID                                int
	EntMemberStatus, MobilePrefix, MobileNo string
}

// UpdateMemberMobile func
func (m *MemberMobile) UpdateMemberMobile(tx *gorm.DB) string {
	var (
		err error
	)

	// check if member status allow to change mobile
	if m.EntMemberStatus != "I" {
		return "active_member_not_allow_to_change_mobile"
	}

	// check if mobile prefix exist
	arrSysTerritoryFn := make([]models.WhereCondFn, 0)
	arrSysTerritoryFn = append(arrSysTerritoryFn,
		models.WhereCondFn{Condition: "sys_territory.calling_no_prefix = ?", CondValue: m.MobilePrefix},
	)
	arrSysTerritory, err := models.GetSysTerritoryFn(arrSysTerritoryFn, "", false)

	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberMobile()", "GetSysTerritoryFn():1", err.Error(), true)
		return "something_went_wrong"
	}

	if arrSysTerritory == nil {
		return "invalid_mobile_prefix"
	}

	countryCode := arrSysTerritory.Code

	mobileNo := strings.Trim(m.MobileNo, " ")

	num, errMsg := mobile_service.ParseMobileNo(mobileNo, countryCode)
	if errMsg != "" {
		return errMsg
	}

	mobilePrefix := fmt.Sprintf("%v", *num.CountryCode)

	// check if unique
	ok, err := models.ExistsMemberByMobile(m.MobilePrefix, m.MobileNo)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberMobile()", "ExistsMemberByMobile():1", err.Error(), true)
		return "something_went_wrong"
	}
	if ok {
		return e.GetMsg(e.MEMBER_MOBILE_EXISTS)
	}

	// update mobile number
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: m.MemberID},
	)
	updateColumn := map[string]interface{}{"mobile_prefix": mobilePrefix, "mobile_no": mobileNo, "updated_by": m.MemberID}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberMobile()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

// MemberUsername struct
type MemberUsername struct {
	MemberID int
	Username string
}

// UpdateMemberUsername func
func (m *MemberUsername) UpdateMemberUsername(tx *gorm.DB) string {
	// validate if username is unique
	ok, err := ExistsMemberByUsername(m.Username)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberUsername()", "ExistsMemberByUsername():1", err.Error(), true)
		return "something_went_wrong"
	}

	if ok {
		return e.GetMsg(e.MEMBER_USERNAME_ALREADY_EXISTS)
	}

	// update member username
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: m.MemberID},
	)

	updateColumn := map[string]interface{}{"nick_name": m.Username}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberUsername()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

// MemberCountry struct
type MemberCountry struct {
	MemberID    int
	CountryCode string
}

// UpdateMemberCountry func
func (m *MemberCountry) UpdateMemberCountry(tx *gorm.DB) string {
	arrCountryData, err := models.GetCountryByCode(m.CountryCode)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberCountry()", "GetCountryByCode():1", err.Error(), true)
		return "something_went_wrong"
	}

	// update member country id
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: m.MemberID},
	)

	updateColumn := map[string]interface{}{"country_id": arrCountryData.ID}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberCountry()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

// MemberGender struct
type MemberGender struct {
	MemberID   int
	GenderCode string
}

// UpdateMemberGender func
func (m *MemberGender) UpdateMemberGender(tx *gorm.DB) string {
	arrGender, err := models.GetSysGeneralByCode("gender", m.GenderCode)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberGender()", "GetCountryByCode():1", err.Error(), true)
		return "something_went_wrong"
	}
	if arrGender == nil {
		return "invalid_gender_code"
	}

	// update member country id
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: m.MemberID},
	)

	updateColumn := map[string]interface{}{"gender_id": arrGender.ID}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberGender()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

// MemberBirthDate struct
type MemberBirthDate struct {
	MemberID  int
	BirthDate string
}

// UpdateMemberBirthDate func
func (m *MemberBirthDate) UpdateMemberBirthDate(tx *gorm.DB) string {
	// validate birth date format
	birthDate, ok := base.ValidateDateTimeFormat(m.BirthDate, "2006-01-02")
	if !ok {
		return "invalid_birth_date_format"
	}

	// update birth date
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: m.MemberID},
	)

	updateColumn := map[string]interface{}{"birth_date": birthDate}
	err := models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberBirthDate()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

// MemberFirstName struct
type MemberFirstName struct {
	MemberID  int
	FirstName string
}

// UpdateMemberFirstName func
func (m *MemberFirstName) UpdateMemberFirstName(tx *gorm.DB) string {
	// validate first name format
	var firstName = strings.Trim(m.FirstName, " ")
	var errMsg = base.FirstNameChecking(firstName)
	if errMsg != "" {
		return errMsg
	}

	// update first name
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: m.MemberID},
	)

	updateColumn := map[string]interface{}{"first_name": firstName}
	err := models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:UpdateMemberFirstName()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

// MemberSettingStatus struct
type MemberSettingStatus struct {
	Username             string `json:"username"`
	Avatar               string `json:"avatar"`
	AccountStatusCode    string `json:"account_status_code"`
	CryptoAddressStatus  int    `json:"crypto_address_status"`
	KycStatus            int    `json:"kyc_status"`
	PlacementStatus      int    `json:"placement_status"`
	PlacementWebviewPath string `json:"placement_webview_path"`
	TwoFAStatus          int    `json:"two_fa_status"`
	TwoFAKey             string `json:"two_fa_key"`
	EContractA           int    `json:"econtract_a"`
	EContractB           int    `json:"econtract_b"`
}

// GetMemberSettingStatus func
func GetMemberSettingStatus(memID int, langCode string) (*MemberSettingStatus, string) {
	var (
		memberSettingStatus MemberSettingStatus
	)

	// find member by username
	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: memID},
	)
	arrEntMember, err := models.GetEntMemberFn(arrEntMemberFn, "", false)

	if err != nil {
		base.LogErrorLog("memberService:GetMemberSettingStatus()", "GetEntMemberFn():1", err.Error(), true)
		return nil, "something_went_wrong"
	}
	if arrEntMember == nil {
		memberSettingStatus.AccountStatusCode = "I"
		return &memberSettingStatus, ""
	}

	memberSettingStatus.Username = arrEntMember.NickName
	memberSettingStatus.Avatar = arrEntMember.Avatar
	memberSettingStatus.AccountStatusCode = arrEntMember.Status

	// crypto address status
	arrEntMemberCryptoFn := make([]models.WhereCondFn, 0)
	arrEntMemberCryptoFn = append(arrEntMemberCryptoFn,
		models.WhereCondFn{Condition: "ent_member_crypto.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: "ent_member_crypto.status = ?", CondValue: "A"},
	)
	arrEntMemberCrypto, err := models.GetEntMemberCryptoFn(arrEntMemberCryptoFn, false)

	if err != nil {
		base.LogErrorLog("memberService:GetMemberSettingStatus()", "GetEntMemberCryptoFn():1", err.Error(), true)
		return nil, "something_went_wrong"
	}

	if arrEntMemberCrypto != nil {
		memberSettingStatus.CryptoAddressStatus = 1
	}

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: memID},
	)
	arrEntMemberKyc, _ := models.GetEntMemberKycFn(arrCond, false)

	if len(arrEntMemberKyc) > 0 {
		if arrEntMemberKyc[0].Status == "AP" {
			memberSettingStatus.KycStatus = 1
		} else if arrEntMemberKyc[0].Status == "R" {
			memberSettingStatus.KycStatus = 2
		}
	}

	// set placement webview path
	adminDomainSetting, _ := models.GetSysGeneralSetupByID("admin_domain")
	if adminDomainSetting != nil {
		memberSettingStatus.PlacementWebviewPath = adminDomainSetting.SettingValue1 + "/member/tree/placement"
	}

	// validate if got placement
	arrEntMemberSponsorFn := make([]models.WhereCondFn, 0)
	arrEntMemberSponsorFn = append(arrEntMemberSponsorFn,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.member_id = ?", CondValue: memID},
	)
	arrEntMemberSponsor, _ := models.GetMemberSponsorFn(arrEntMemberSponsorFn, false)
	if arrEntMemberSponsor.UplineID == 0 {
		memberSettingStatus.PlacementStatus = -1

		// check if only upline can place.
		networkStatus := VerifyIfInNetwork(memID, "NOT_ALLOW_TO_PLACE_OWN_PLACEMENT")

		if !networkStatus {
			memberSettingStatus.PlacementStatus = 1
		}
	}

	// 2fa setting
	var secret = ""

	// check if already got secret key
	arrEntMember2FaFn := make([]models.WhereCondFn, 0)
	arrEntMember2FaFn = append(arrEntMember2FaFn,
		models.WhereCondFn{Condition: "ent_member_2fa.member_id = ? ", CondValue: memID},
	)
	arrEntMember2Fa, _ := models.GetEntMember2FA(arrEntMember2FaFn, false)

	if len(arrEntMember2Fa) <= 0 {
		// generate new unique secret key
		charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		numChar := 10
		var validPassword bool
		for !validPassword {
			secret = base.GenerateRandomString(numChar, charSet)

			// validate if unique
			arrEntMember2FaFn2 := make([]models.WhereCondFn, 0)
			arrEntMember2FaFn2 = append(arrEntMember2FaFn2,
				models.WhereCondFn{Condition: "ent_member_2fa.secret = ? ", CondValue: secret},
			)
			arrEntMember2Fa2, _ := models.GetEntMember2FA(arrEntMember2FaFn2, false)
			if len(arrEntMember2Fa2) == 0 {
				validPassword = true
			}
		}

		// save secret key
		arrAddEntMember2FA := models.AddEntMember2FaParam{
			MemberID: memID,
			Secret:   secret,
			CodeUrl:  "",
			BEnable:  0,
		}
		db := models.GetDB() // no need set begin transaction
		_, err = models.AddEntMember2FA(db, arrAddEntMember2FA)
		if err != nil {
			secret = ""
		}

		memberSettingStatus.TwoFAStatus = 0
		memberSettingStatus.TwoFAKey = secret
	} else {
		if arrEntMember2Fa[0].BEnable == 1 {
			memberSettingStatus.TwoFAStatus = 1
		}

		memberSettingStatus.TwoFAKey = arrEntMember2Fa[0].Secret
	}

	if memberSettingStatus.TwoFAKey != "" {
		memberSettingStatus.TwoFAKey = util.EncodeBase32(memberSettingStatus.TwoFAKey)
	}

	// validate if member got buy contract
	// arrSlsMasterFn := []models.WhereCondFn{}
	// arrSlsMasterFn = append(arrSlsMasterFn,
	// 	models.WhereCondFn{Condition: " sls_master.member_id = ?", CondValue: memID},
	// 	models.WhereCondFn{Condition: " sls_master.doc_type = ?", CondValue: "CT"},
	// 	models.WhereCondFn{Condition: " sls_master.status = ?", CondValue: "AP"},
	// )
	// arrSlsMaster, _ := models.GetSlsMasterFn(arrSlsMasterFn, "", false)
	// if len(arrSlsMaster) > 0 {
	// 	memberSettingStatus.EContractB = 1
	// }

	// validate member got buy membership
	arrEntMemberMembershipFn := []models.WhereCondFn{}
	arrEntMemberMembershipFn = append(arrEntMemberMembershipFn,
		models.WhereCondFn{Condition: " ent_member_membership.member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " ent_member_membership.b_valid = ?", CondValue: 1},
		models.WhereCondFn{Condition: " ent_member_membership.expired_at >= ?", CondValue: time.Now()},
	)
	arrEntMemberMembership, _ := models.GetEntMemberMembership(arrEntMemberMembershipFn, "", false)
	if len(arrEntMemberMembership) > 0 {
		memberSettingStatus.EContractA = 1
	}

	return &memberSettingStatus, ""
}

// BindMemberMobileEmailStruct struct
type BindMemberMobileEmailStruct struct {
	BindType                      string // email/mobile
	MainID                        int
	MobilePrefix, MobileNo, Email string
}

// BindMemberMobileEmail func
func BindMemberMobileEmail(tx *gorm.DB, m BindMemberMobileEmailStruct, langCode string) string {
	// get member info
	arrMembersFn := make([]models.WhereCondFn, 0)
	arrMembersFn = append(arrMembersFn,
		models.WhereCondFn{Condition: "members.id = ?", CondValue: m.MainID},
	)
	arrMembers, err := models.GetMembersFn(arrMembersFn, false)
	if err != nil {
		base.LogErrorLog("memberService:BindMemberMobileEmail()", "GetMembersFn():1", err.Error(), true)
		return "something_went_wrong"
	}
	if arrMembers == nil {
		return e.GetMsg(e.INVALID_MEMBER)
	}

	// prepare update condition
	var updateColumn = map[string]interface{}{}

	if m.BindType == "MOBILE" {
		// check member already binded with mobile
		if arrMembers.MobilePrefix != "" || arrMembers.MobileNo != "" {
			return "user_already_binded_with_a_mobile_no"
		}

		// validate if mobile no is unique
		ok, err := models.ExistsMemberByMobile(m.MobilePrefix, m.MobileNo)
		if err != nil {
			base.LogErrorLog("memberService:BindMemberMobileEmail()", "ExistsMemberByMobile():1", err.Error(), true)
			return "something_went_wrong"
		}
		if ok {
			return e.GetMsg(e.MEMBER_MOBILE_EXISTS)
		}

		// append update col
		updateColumn = map[string]interface{}{"mobile_prefix": m.MobilePrefix, "mobile_no": m.MobileNo, "updated_by": m.MainID}
	} else if m.BindType == "EMAIL" {
		// check member already binded with email
		if arrMembers.Email != "" {
			return "user_already_binded_with_an_email"
		}

		// validate if email is unique
		ok, err := models.ExistsMemberByEmail(m.Email)
		if err != nil {
			base.LogErrorLog("memberService:BindMemberMobileEmail()", "ExistsMemberByEmail():1", err.Error(), true)
			return "something_went_wrong"
		}
		if ok {
			return e.GetMsg(e.MEMBER_EMAIL_EXISTS)
		}

		// append update col
		updateColumn = map[string]interface{}{"email": m.Email, "updated_by": m.MainID}
	} else {
		return "invalid_bind_type"
	}

	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "id = ?", CondValue: m.MainID})

	err = models.UpdatesFnTx(tx, "members", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("memberService:BindMemberMobileEmail()", "UpdatesFnTx():1", err.Error(), true)
		return "something_went_wrong"
	}

	return ""
}

// GetMemberCurRank func
func GetMemberCurRank(memID int, langCode string) string {
	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")
	curRankName := ""

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " DATE(tbl_bonus_rank_star.t_bns_fr) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " DATE(tbl_bonus_rank_star.t_bns_to) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " tbl_bonus_rank_star.b_latest = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " tbl_bonus_rank_star.t_member_id = ? ", CondValue: memID},
	)
	arrTblBonusRankStar, _ := models.GetTblBonusRankStarFn(arrCond, false)

	if len(arrTblBonusRankStar) > 0 {
		if arrTblBonusRankStar[0].TRankEff > 0 {
			// arrTransValue["num"] = fmt.Sprint(arrTblBonusRankStar[0].TRankEff)
			// curRankName = helpers.TranslateV2("v:num", langCode, map[string]string{})

			if arrTblBonusRankStar[0].TRankEff == 1 {
				curRankName = helpers.TranslateV2("associate", langCode, map[string]string{})
			} else if arrTblBonusRankStar[0].TRankEff == 2 {
				curRankName = helpers.TranslateV2("partner", langCode, map[string]string{})
			} else if arrTblBonusRankStar[0].TRankEff == 3 {
				curRankName = helpers.TranslateV2("senior_partner", langCode, map[string]string{})
			} else if arrTblBonusRankStar[0].TRankEff == 4 {
				curRankName = helpers.TranslateV2("regional_partner", langCode, map[string]string{})
			} else if arrTblBonusRankStar[0].TRankEff == 5 {
				curRankName = helpers.TranslateV2("global_partner", langCode, map[string]string{})
			} else if arrTblBonusRankStar[0].TRankEff == 6 {
				curRankName = helpers.TranslateV2("director", langCode, map[string]string{})
			}
		}
	}

	return curRankName
}

func GenRandomMemberCode() string {
	codeCharSet := "abcdefghijklmnopqrstuvwxyz"
	var memCode string
	for {
		memCode = base.GenerateRandomString(10, codeCharSet)
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member.code = ?", CondValue: memCode},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		arrExistingMemberCode, _ := models.GetEntMemberFn(arrCond, "", false)
		if arrExistingMemberCode == nil {
			return memCode
		}
	}
}

func ProcessUpdateMissingMemberCode() {
	// for i := 0; i < 20; i++ {
	// 	memCode := GenRandomMemberCode()
	// 	fmt.Println(strconv.Itoa(i)+"memCode:", memCode, len(memCode))
	// }
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member.code = '' AND ent_member.status = ?", CondValue: "I"},
	)
	arrEmptyCodeEntMember, _ := models.GetEntMemberListFn(arrCond, false)
	if len(arrEmptyCodeEntMember) > 0 {
		for _, arrEmptyCodeEntMemberV := range arrEmptyCodeEntMember {
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: " ent_member.id = ?", CondValue: arrEmptyCodeEntMemberV.ID},
			)
			updateColumn := make(map[string]interface{}, 0)
			memCode := GenRandomMemberCode()
			updateColumn["code"] = memCode
			// fmt.Println("memCode:", memCode, len(memCode))
			models.UpdatesFn("ent_member", arrUpdCond, updateColumn, false)
		}
	}
}

func GetMemberBigLegID(memID int) (int, string) {
	bigLegID := 0
	dateFormat := base.ConvertFormat("yyyy-mm-dd")
	ytdDate := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " tbl_bonus_rank_star_passup.t_member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " tbl_bonus_rank_star_passup.t_bns_id = ?", CondValue: ytdDate},
	)

	rst, err := models.GetTblBonusRankStarPassupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberBigLegID-GetTblBonusRankStarPassupFn_failed", err.Error(), arrCond, true)
		return bigLegID, "something_went_wrong"
	}

	if len(rst) > 0 {
		return rst[0].TDownlineID, ""
	}

	return bigLegID, ""
}

type RankingInfoStruct struct {
	CurrentRank  string  `json:"current_rank"`
	RankFrom     string  `json:"rank_from"`
	RankTo       string  `json:"rank_to"`
	RankPercent  float64 `json:"rank_percent"`
	RankAchieved string  `json:"rank_achieved"`
	RankTarget   string  `json:"rank_target"`
	RankLine     int     `json:"rank_line"`
}

func GetMemberRankingInfo(memID int) (*RankingInfoStruct, string) {
	var rankingInfo RankingInfoStruct
	//find big leg
	bigLegID, bigLegIDErr := GetMemberBigLegID(memID)

	if bigLegIDErr != "" {
		return nil, bigLegIDErr
	}

	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")
	// find member's current rank
	arrRankStarCond := make([]models.WhereCondFn, 0)
	arrRankStarCond = append(arrRankStarCond,
		models.WhereCondFn{Condition: "tbl_bonus_rank_star.t_member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " DATE(tbl_bonus_rank_star.t_bns_fr) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " DATE(tbl_bonus_rank_star.t_bns_to) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " tbl_bonus_rank_star.b_latest = ? ", CondValue: 1},
	)
	rankStarRst, rankStarErr := models.GetTblBonusRankStarFn(arrRankStarCond, false)

	if rankStarErr != nil {
		base.LogErrorLog("member_service:GetMemberRankingInfo()", "GetTblBonusRankStarFn", rankStarErr.Error(), true)
		return nil, "something_went_wrong"
	}

	currentRank := 0
	if len(rankStarRst) > 0 {
		currentRank = rankStarRst[0].TRankEff
	}
	rankTo := currentRank + 1
	if rankTo > 5 {
		rankTo = 5
	}
	rankingInfo.CurrentRank = "v" + strconv.Itoa(currentRank)
	rankingInfo.RankFrom = "v" + strconv.Itoa(currentRank)
	rankingInfo.RankTo = "v" + strconv.Itoa(rankTo)

	// get percent to reach next rank
	if currentRank == 0 {
		// check if member got active contract
		salesCond := make([]models.WhereCondFn, 0)
		salesCond = append(salesCond,
			models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		sales, salesErr := models.GetTotalSalesAmount(salesCond, false)
		if salesErr != nil {
			base.LogErrorLog("member_service:GetMemberRankingInfo()", "GetTotalSalesAmount()", salesErr.Error(), true)
			return nil, "something_went_wrong"
		}

		tbBonusRankStarPassupCond := make([]models.WhereCondFn, 0)
		tbBonusRankStarPassupCond = append(tbBonusRankStarPassupCond,
			models.WhereCondFn{Condition: "t_member_id = ?", CondValue: memID},
			// models.WhereCondFn{Condition: "t_direct_sponsor_id = ?", CondValue: arrData.EntMemberID},
			// models.WhereCondFn{Condition: "t_rank_qualify >= ?", CondValue: 1},
		)
		tbBonusRankStarPassupRst, tbBonusRankStarPassupErr := models.GetTblBonusRankStarPassupFn(tbBonusRankStarPassupCond, "", false)

		if tbBonusRankStarPassupErr != nil {
			base.LogErrorLog("member_service:GetMemberRankingInfo()", "GetTodayDirectSponsor", tbBonusRankStarPassupErr.Error(), true)
			return nil, "something_went_wrong"
		}
		passupValue := 0.00
		if len(tbBonusRankStarPassupRst) > 0 && sales.TotalAmount > 0 {
			passupValue = tbBonusRankStarPassupRst[0].FBvSmall
		}
		rankingInfo.RankAchieved = helpers.CutOffDecimal(passupValue, uint(2), ".", ",")
		rankingInfo.RankTarget = helpers.CutOffDecimal(50000.00, uint(2), ".", ",")
		rankingInfo.RankPercent = passupValue / 50000
		rankingInfo.RankLine = 0
	} else if currentRank < 5 {
		dateFormat := base.ConvertFormat("yyyy-mm-dd")
		ytdDate := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)

		numberOfDownlineCond := make([]models.WhereCondFn, 0)
		numberOfDownlineCond = append(numberOfDownlineCond,
			models.WhereCondFn{Condition: "t_direct_sponsor_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: ytdDate},
			models.WhereCondFn{Condition: "t_member_id <> ?", CondValue: bigLegID},
			// models.WhereCondFn{Condition: "t_rank_qualify >= ?", CondValue: currentRank},
		)
		numberOfDownlineRst, numberOfDownlineErr := models.GetTblBonusRankStarPassupFn(numberOfDownlineCond, "SUM(t_downline_rank"+strconv.Itoa(currentRank)+") as total_downline", false)

		if numberOfDownlineErr != nil {
			base.LogErrorLog("member_service:GetMemberRankingInfo()", "GetTblBonusRankStarPassupFn()", numberOfDownlineErr.Error(), true)
			return nil, "something_went_wrong"
		}

		target := 2
		if currentRank == 3 || currentRank == 4 {
			target = 3
		}
		totalDownline := 0
		if len(numberOfDownlineRst) > 0 {
			totalDownline = numberOfDownlineRst[0].TotalDownline
		}
		rankingInfo.RankPercent = float.Div(float64(totalDownline), float64(target))
		rankingInfo.RankLine = target - 1

		// max 100%
		if rankingInfo.RankPercent > 1 {
			rankingInfo.RankPercent = 1
		}
	} else {
		rankingInfo.RankPercent = 1
		rankingInfo.RankLine = 2
	}

	return &rankingInfo, ""
}

func GetMemberPoolRankingInfo(memID int) (*RankingInfoStruct, string) {
	var rankingInfo RankingInfoStruct
	//find big leg
	bigLegID, bigLegIDErr := GetMemberBigLegID(memID)

	if bigLegIDErr != "" {
		return nil, bigLegIDErr
	}

	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

	// find member's current rank
	arrRankStarCond := make([]models.WhereCondFn, 0)
	arrRankStarCond = append(arrRankStarCond,
		models.WhereCondFn{Condition: "tbl_p2p_bonus_rank_star.t_member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " DATE(tbl_p2p_bonus_rank_star.t_bns_fr) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " DATE(tbl_p2p_bonus_rank_star.t_bns_to) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " tbl_p2p_bonus_rank_star.b_latest = ? ", CondValue: 1},
	)
	rankStarRst, err := models.GetTblP2PBonusRankStarFn(arrRankStarCond, false)

	if err != nil {
		base.LogErrorLog("GetMemberPoolRankingInfo-GetTblP2PBonusRankStarFn_failed", err.Error(), arrRankStarCond, true)
		return nil, "something_went_wrong"
	}

	currentRank := 0
	if len(rankStarRst) > 0 {
		currentRank = rankStarRst[0].TRankEff
	}
	rankTo := currentRank + 1
	if rankTo > 5 {
		rankTo = 5
	}
	rankingInfo.CurrentRank = "p" + strconv.Itoa(currentRank)
	rankingInfo.RankFrom = "p" + strconv.Itoa(currentRank)
	rankingInfo.RankTo = "p" + strconv.Itoa(rankTo)

	// get percent to reach next rank
	if currentRank == 0 {
		// check if member got active contract
		salesCond := make([]models.WhereCondFn, 0)
		salesCond = append(salesCond,
			models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "P2P"},
			models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		sales, err := models.GetTotalSalesAmount(salesCond, false)
		if err != nil {
			base.LogErrorLog("GetMemberPoolRankingInfo-GetTotalSalesAmount_failed", err.Error(), salesCond, true)
			return nil, "something_went_wrong"
		}

		tbBonusRankStarPassupCond := make([]models.WhereCondFn, 0)
		tbBonusRankStarPassupCond = append(tbBonusRankStarPassupCond,
			models.WhereCondFn{Condition: "t_member_id = ?", CondValue: memID},
			// models.WhereCondFn{Condition: "t_direct_sponsor_id = ?", CondValue: arrData.EntMemberID},
			// models.WhereCondFn{Condition: "t_rank_qualify >= ?", CondValue: 1},
		)
		tbBonusRankStarPassupRst, err := models.GetTblP2PBonusRankStarPassupFn(tbBonusRankStarPassupCond, "", false)

		if err != nil {
			base.LogErrorLog("GetMemberPoolRankingInfo-GetTblP2PBonusRankStarPassupFn_1_failed", err.Error(), tbBonusRankStarPassupCond, true)
			return nil, "something_went_wrong"
		}
		passupValue := 0.00
		if len(tbBonusRankStarPassupRst) > 0 && sales.TotalAmount > 0 {
			passupValue = tbBonusRankStarPassupRst[0].FBvSmall
		}
		rankingInfo.RankAchieved = helpers.CutOffDecimal(passupValue, uint(2), ".", ",")
		rankingInfo.RankTarget = helpers.CutOffDecimal(3000.00, uint(2), ".", ",")
		rankingInfo.RankPercent = passupValue / 3000
		rankingInfo.RankLine = 0
	} else if currentRank < 5 {
		dateFormat := base.ConvertFormat("yyyy-mm-dd")
		ytdDate := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)

		numberOfDownlineCond := make([]models.WhereCondFn, 0)
		numberOfDownlineCond = append(numberOfDownlineCond,
			models.WhereCondFn{Condition: "t_direct_sponsor_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: ytdDate},
			models.WhereCondFn{Condition: "t_member_id <> ?", CondValue: bigLegID},
			// models.WhereCondFn{Condition: "t_rank_qualify >= ?", CondValue: currentRank},
		)
		numberOfDownlineRst, err := models.GetTblP2PBonusRankStarPassupFn(numberOfDownlineCond, "SUM(t_downline_rank"+strconv.Itoa(currentRank)+") as total_downline", false)

		if err != nil {
			base.LogErrorLog("GetMemberPoolRankingInfo-GetTblP2PBonusRankStarPassupFn_2_failed", err.Error(), numberOfDownlineCond, true)
			return nil, "something_went_wrong"
		}

		target := 2
		if currentRank == 3 || currentRank == 4 {
			target = 3
		}
		totalDownline := 0
		if len(numberOfDownlineRst) > 0 {
			totalDownline = numberOfDownlineRst[0].TotalDownline
		}
		rankingInfo.RankPercent = float.Div(float64(totalDownline), float64(target))
		rankingInfo.RankLine = target - 1

		// max 100%
		if rankingInfo.RankPercent > 1 {
			rankingInfo.RankPercent = 1
		}
	} else {
		rankingInfo.RankPercent = 1
		rankingInfo.RankLine = 2
	}

	return &rankingInfo, ""
}

func GetMemberPoolBigLegID(memID int) (int, string) {
	bigLegID := 0
	dateFormat := base.ConvertFormat("yyyy-mm-dd")
	ytdDate := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " tbl_p2p_bonus_rank_star_passup.t_member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " tbl_p2p_bonus_rank_star_passup.t_bns_id = ?", CondValue: ytdDate},
	)

	rst, err := models.GetTblP2PBonusRankStarPassupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberBigLegID-GetTblBonusRankStarPassupFn_failed", err.Error(), arrCond, true)
		return bigLegID, "something_went_wrong"
	}

	if len(rst) > 0 {
		return rst[0].TDownlineID, ""
	}

	return bigLegID, ""
}

func VerifyIfInNetwork(memberID int, setupCode string) bool {
	var networkType = "SPONSOR"

	// get blocked network setup by code
	arrEntMemberNetworkSetupFn := make([]models.WhereCondFn, 0)
	arrEntMemberNetworkSetupFn = append(arrEntMemberNetworkSetupFn,
		models.WhereCondFn{Condition: " ent_member_network_setup.code = ?", CondValue: setupCode},
	)

	arrEntMemberNetworkSetup, err := models.GetEntMemberNetworkSetupFn(arrEntMemberNetworkSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("memberService:VerifyIfInNetwork():GetEntMemberNetworkSetupFn()", err.Error(), map[string]interface{}{"arrCond": arrEntMemberNetworkSetup}, true)
		return false
	}

	if len(arrEntMemberNetworkSetup) > 0 {
		var (
			arrMemberIDRaw string = strings.ReplaceAll(arrEntMemberNetworkSetup[0].MemberID, " ", "")
			arrMemberID           = strings.Split(arrMemberIDRaw, ",")
		)

		// run get nearest upline
		nearestUpline := GetNearestUplineByMemberID(memberID, arrMemberID, networkType)
		if !nearestUpline.Status {
			base.LogErrorLog("memberService:VerifyIfInNetwork():GetNearestUplineByMemberID()", nearestUpline.ErrMsg, map[string]interface{}{"memberID": memberID, "arrTargetID": arrMemberID, "networkType": networkType}, true)
			return false
		}

		if nearestUpline.UplineID != 0 {
			fmt.Println("uplineID:", nearestUpline.UplineID, "uplineUsername:", nearestUpline.UplineUsername)
			return true
		}
	}

	return false
}

type NearestUpline struct {
	Status         bool
	UplineID       int
	UplineUsername string
	ErrMsg         string
}

func GetNearestUplineByMemberID(memberID int, arrTargetID []string, networkType string) NearestUpline {
	var (
		memID          string = fmt.Sprint(memberID)
		uplineID       int
		uplineUsername string
	)

	if helpers.StringInSlice(memID, arrTargetID) == true {
		arrEntMemberFn := make([]models.WhereCondFn, 0)
		arrEntMemberFn = append(arrEntMemberFn,
			models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: memID},
		)
		arrEntMember, err := models.GetEntMemberFn(arrEntMemberFn, "", false)
		if err != nil {
			return NearestUpline{Status: false, UplineID: 0, UplineUsername: "", ErrMsg: "GetEntMemberFn():1" + err.Error()}
		}

		uplineID = arrEntMember.ID
		uplineUsername = arrEntMember.NickName
		return NearestUpline{Status: true, UplineID: uplineID, UplineUsername: uplineUsername}
	}

	for !helpers.StringInSlice(memID, arrTargetID) {
		if memID == "1" { // if hit com, then uplineID and uplineUsername return empty
			return NearestUpline{Status: true, UplineID: 0, UplineUsername: ""}
		}

		// get current memID's sponsorID/uplineID
		arrEntMemberTreeSponsorFn := make([]models.WhereCondFn, 0)
		arrEntMemberTreeSponsorFn = append(arrEntMemberTreeSponsorFn,
			models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: memID},
		)
		arrEntMemberTreeSponsor, err := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberTreeSponsorFn, false)
		if err != nil {
			return NearestUpline{Status: false, UplineID: 0, UplineUsername: "", ErrMsg: "GetEntMemberEntMemberTreeSponsorFn():" + err.Error()}
		}

		if arrEntMemberTreeSponsor != nil {
			uplineID = arrEntMemberTreeSponsor.SponsorID
			memID = fmt.Sprint(uplineID)

			if networkType == "PLACEMENT" {
				uplineID = arrEntMemberTreeSponsor.UplineID
				memID = fmt.Sprint(uplineID)
			}
		} else {
			return NearestUpline{Status: false, UplineID: 0, UplineUsername: "", ErrMsg: "GetEntMemberEntMemberTreeSponsorFn():" + "no_record_found_for_member_id_" + memID}
		}
	}

	if uplineID != 0 && uplineID != 1 { // hit void or hit com
		arrEntMemberFn := make([]models.WhereCondFn, 0)
		arrEntMemberFn = append(arrEntMemberFn,
			models.WhereCondFn{Condition: "ent_member.id = ?", CondValue: uplineID},
		)
		arrEntMember, err := models.GetEntMemberFn(arrEntMemberFn, "", false)
		if err != nil {
			return NearestUpline{Status: false, UplineID: 0, UplineUsername: "", ErrMsg: "GetEntMemberFn():2" + err.Error()}
		}

		uplineUsername = arrEntMember.NickName
	}

	return NearestUpline{Status: true, UplineID: uplineID, UplineUsername: uplineUsername}
}

func GetMiningMemberBigLegID(memID int) (int, string) {
	bigLegID := 0
	dateFormat := base.ConvertFormat("yyyy-mm-dd")
	ytdDate := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " tbl_mm_bonus_rank_star_passup.t_member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " tbl_mm_bonus_rank_star_passup.t_bns_id = ?", CondValue: ytdDate},
	)
	rst, err := models.GetTblMiningBonusRankStarPassupFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("memberService:GetMemberBigLegID():GetTblMiningBonusRankStarPassupFn()", err.Error(), arrCond, true)
		return bigLegID, "something_went_wrong"
	}

	if len(rst) > 0 {
		return rst[0].TDownlineID, ""
	}

	return bigLegID, ""
}

func GetMemberMiningRankingInfo(memID int) (RankingInfoStruct, string) {
	var rankingInfo RankingInfoStruct
	//find big leg
	bigLegID, bigLegIDErr := GetMiningMemberBigLegID(memID)

	if bigLegIDErr != "" {
		return RankingInfoStruct{}, bigLegIDErr
	}

	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

	// find member's current rank
	arrRankStarCond := make([]models.WhereCondFn, 0)
	arrRankStarCond = append(arrRankStarCond,
		models.WhereCondFn{Condition: "tbl_mm_bonus_rank_star.t_member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " DATE(tbl_mm_bonus_rank_star.t_bns_fr) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " DATE(tbl_mm_bonus_rank_star.t_bns_to) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " tbl_mm_bonus_rank_star.b_latest = ? ", CondValue: 1},
	)
	rankStarRst, err := models.GetTblMiningBonusRankStarFn(arrRankStarCond, false)

	if err != nil {
		base.LogErrorLog("member_service:GetMemberMiningRankingInfo():GetTblMiningBonusRankStarFn", err.Error(), map[string]interface{}{"condition": arrRankStarCond}, true)
		return RankingInfoStruct{}, "something_went_wrong"
	}

	currentRank := 0
	if len(rankStarRst) > 0 {
		currentRank = rankStarRst[0].TRankEff
	}
	rankTo := currentRank + 1
	if rankTo > 5 {
		rankTo = 5
	}
	rankingInfo.CurrentRank = "V" + strconv.Itoa(currentRank)
	rankingInfo.RankFrom = "V" + strconv.Itoa(currentRank)
	rankingInfo.RankTo = "V" + strconv.Itoa(rankTo)

	// get percent to reach next rank
	if currentRank == 0 {
		// check if member got active contract
		salesCond := make([]models.WhereCondFn, 0)
		salesCond = append(salesCond,
			models.WhereCondFn{Condition: "sls_master.member_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "MINING"},
			models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		sales, err := models.GetTotalSalesAmount(salesCond, false)
		if err != nil {
			base.LogErrorLog("member_service:GetMemberMiningRankingInfo():GetTotalSalesAmount()", err.Error(), map[string]interface{}{"condition": salesCond}, true)
			return RankingInfoStruct{}, "something_went_wrong"
		}

		tbBonusRankStarPassupCond := make([]models.WhereCondFn, 0)
		tbBonusRankStarPassupCond = append(tbBonusRankStarPassupCond,
			models.WhereCondFn{Condition: "t_member_id = ?", CondValue: memID},
		)
		tbBonusRankStarPassupRst, err := models.GetTblMiningBonusRankStarPassupFn(tbBonusRankStarPassupCond, "", false)

		if err != nil {
			base.LogErrorLog("member_service:GetMemberMiningRankingInfo():GetTblMiningBonusRankStarPassupFn():1", err.Error(), map[string]interface{}{"condition": tbBonusRankStarPassupCond}, true)
			return RankingInfoStruct{}, "something_went_wrong"
		}
		passupValue := 0.00
		if len(tbBonusRankStarPassupRst) > 0 && sales.TotalAmount > 0 {
			passupValue = tbBonusRankStarPassupRst[0].FBvSmall
		}
		rankingInfo.RankAchieved = helpers.CutOffDecimal(passupValue, uint(2), ".", ",")
		rankingInfo.RankTarget = helpers.CutOffDecimal(50000.00, uint(2), ".", ",")
		rankingInfo.RankPercent = passupValue / 50000
		rankingInfo.RankLine = 0
	} else if currentRank < 5 {
		dateFormat := base.ConvertFormat("yyyy-mm-dd")
		ytdDate := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)

		tbBonusRankStarPassupCond := make([]models.WhereCondFn, 0)
		tbBonusRankStarPassupCond = append(tbBonusRankStarPassupCond,
			models.WhereCondFn{Condition: "t_direct_sponsor_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: ytdDate},
			models.WhereCondFn{Condition: "t_member_id <> ?", CondValue: bigLegID},
		)
		numberOfDownlineRst, numberOfDownlineErr := models.GetTblMiningBonusRankStarPassupFn(tbBonusRankStarPassupCond, "SUM(t_downline_rank"+strconv.Itoa(currentRank)+") as total_downline", false)

		if numberOfDownlineErr != nil {
			base.LogErrorLog("member_service:GetMemberMiningRankingInfo():GetTblMiningBonusRankStarPassupFn():2", numberOfDownlineErr.Error(), map[string]interface{}{"condition": tbBonusRankStarPassupCond}, true)
			return RankingInfoStruct{}, "something_went_wrong"
		}

		target := 2
		if currentRank == 3 || currentRank == 4 {
			target = 3
		}
		totalDownline := 0
		if len(numberOfDownlineRst) > 0 {
			totalDownline = numberOfDownlineRst[0].TotalDownline
		}
		rankingInfo.RankPercent = float.Div(float64(totalDownline), float64(target))
		rankingInfo.RankLine = target - 1

		// max 100%
		if rankingInfo.RankPercent > 1 {
			rankingInfo.RankPercent = 1
		}
	} else {
		rankingInfo.RankPercent = 1
		rankingInfo.RankLine = 2
	}

	return rankingInfo, ""
}

// GetMemberCurP2PRank func
func GetMemberCurP2PRank(memID int, langCode string) string {
	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " DATE(tbl_p2p_bonus_rank_star.t_bns_fr) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " DATE(tbl_p2p_bonus_rank_star.t_bns_to) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " tbl_p2p_bonus_rank_star.b_latest = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " tbl_p2p_bonus_rank_star.t_member_id = ? ", CondValue: memID},
	)
	arrTblP2PBonusRankStar, _ := models.GetTblP2PBonusRankStarFn(arrCond, false)

	curRankName := ""
	if len(arrTblP2PBonusRankStar) > 0 {
		fmt.Println(arrTblP2PBonusRankStar)
		arrTransValue := make(map[string]string)
		if arrTblP2PBonusRankStar[0].TRankEff > 0 { // more than 0 only got
			arrTransValue["num"] = fmt.Sprint(arrTblP2PBonusRankStar[0].TRankEff)
			curRankName = helpers.TranslateV2("p:num", langCode, arrTransValue)
		}

	}

	return curRankName
}

// GetMemberCurMMRank func
func GetMemberCurMMRank(memID int, langCode string) string {
	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " DATE(tbl_mm_bonus_rank_star.t_bns_fr) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " DATE(tbl_mm_bonus_rank_star.t_bns_to) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " tbl_mm_bonus_rank_star.b_latest = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " tbl_mm_bonus_rank_star.t_member_id = ? ", CondValue: memID},
	)
	arrTblMiningBonusRankStar, _ := models.GetTblMiningBonusRankStarFn(arrCond, false)

	curRankName := ""
	if len(arrTblMiningBonusRankStar) > 0 {
		arrTransValue := make(map[string]string)
		if arrTblMiningBonusRankStar[0].TRankEff > 0 {
			arrTransValue["num"] = fmt.Sprint(arrTblMiningBonusRankStar[0].TRankEff)
			curRankName = helpers.TranslateV2("v:num", langCode, arrTransValue)
		}
	}

	return curRankName
}

type RankingCriteriaStruct struct {
	CurrentRank  string               `json:"current_rank"`
	RankFrom     string               `json:"rank_from"`
	RankTo       string               `json:"rank_to"`
	CriteriaList []CriteriaListStruct `json:"criteria_list"`
}

type CriteriaListStruct struct {
	RankLabel    string  `json:"rank_label"`
	RankPercent  float64 `json:"rank_percent"`
	RankAchieved string  `json:"rank_achieved"`
	RankTarget   string  `json:"rank_target"`
	RankLine     int     `json:"rank_line"`
}

func GetMemberBZZMiningRankingInfo(memID int, langCode string) (*RankingCriteriaStruct, error) {
	var rankingInfo RankingCriteriaStruct
	//find big leg
	bigLegID, bigLegIDErr := GetMiningMemberBigLegID(memID)

	if bigLegIDErr != "" {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: bigLegIDErr}
	}

	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

	// find member's current rank
	arrRankStarCond := make([]models.WhereCondFn, 0)
	arrRankStarCond = append(arrRankStarCond,
		models.WhereCondFn{Condition: "tbl_mm_bonus_rank_star.t_member_id = ?", CondValue: memID},
		models.WhereCondFn{Condition: " DATE(tbl_mm_bonus_rank_star.t_bns_fr) <= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " DATE(tbl_mm_bonus_rank_star.t_bns_to) >= ? ", CondValue: curDate},
		models.WhereCondFn{Condition: " tbl_mm_bonus_rank_star.b_latest = ? ", CondValue: 1},
	)
	rankStarRst, err := models.GetTblMiningBonusRankStarFn(arrRankStarCond, false)

	if err != nil {
		base.LogErrorLog("GetMemberBZZMiningRankingInfo-GetTblMiningBonusRankStarFn_failed", err.Error(), arrRankStarCond, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	currentRank := 0
	if len(rankStarRst) > 0 {
		fmt.Println("rankStarRst[0].TRankEff:", rankStarRst[0].TRankEff)
		currentRank = rankStarRst[0].TRankEff
	}
	rankTo := currentRank + 1
	if rankTo > 5 {
		rankTo = 5
	}
	rankingInfo.CurrentRank = "V" + strconv.Itoa(currentRank)
	rankingInfo.RankFrom = "V" + strconv.Itoa(currentRank)
	rankingInfo.RankTo = "V" + strconv.Itoa(rankTo)

	personalTranslated := helpers.TranslateV2("ownself", langCode, nil)
	totalNetworkTranslated := helpers.TranslateV2("group", langCode, nil)
	minNetworkSalesTranslated := helpers.TranslateV2("min_network_sales", langCode, nil)
	numOfRankTranslated := rankingInfo.CurrentRank

	arrCriteriaList := make([]CriteriaListStruct, 0)
	if currentRank == 0 {
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member.id = ? ", CondValue: memID},
			models.WhereCondFn{Condition: " sls_master.status IN ('AP', 'EP') AND sls_master.action = ? ", CondValue: "MINING_BZZ"},
			models.WhereCondFn{Condition: " sls_master.total_bv > ? ", CondValue: 0},
		)
		totalPersonalSalesRst, _ := models.GetTotalBZZSalesFn(arrCond, false)
		totalPersonalNodeSales := float64(0)
		if totalPersonalSalesRst.TotalNodes > 0 {
			totalPersonalNodeSales = totalPersonalSalesRst.TotalNodes
		}

		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " sponsor_lot.member_id = ? ", CondValue: memID},
			models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
			models.WhereCondFn{Condition: " sls_master.status IN ('AP', 'EP') AND sls_master.action = ? ", CondValue: "MINING_BZZ"},
			models.WhereCondFn{Condition: " sls_master.total_bv > ? ", CondValue: 0},
		)
		totalNetworkSalesRst, _ := models.GetTotalNetworkBZZSalesFn(arrCond, false)
		totalNetworkNodeSales := float64(0)
		if totalNetworkSalesRst.TotalNodes > 0 {
			totalNetworkNodeSales = totalNetworkSalesRst.TotalNodes
		}
		// tbBonusRankStarPassupCond := make([]models.WhereCondFn, 0)
		// tbBonusRankStarPassupCond = append(tbBonusRankStarPassupCond,
		// 	models.WhereCondFn{Condition: "t_member_id = ?", CondValue: memID},
		// )
		// tbBonusRankStarPassupRst, err := models.GetTblMiningBonusRankStarPassupFn(tbBonusRankStarPassupCond, "", false)

		// if err != nil {
		// 	base.LogErrorLog("GetMemberBZZMiningRankingInfo-GetTblMiningBonusRankStarPassupFn():1", err.Error(), map[string]interface{}{"condition": tbBonusRankStarPassupCond}, true)
		// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		// }

		// fQtyAcc := 0.00
		// fqty := 0.00
		personalRankTarget := 8
		totalNetworkRankTarget := 16

		totalPersonalNodeSales2 := totalPersonalNodeSales
		if totalPersonalNodeSales > float64(personalRankTarget) {
			totalPersonalNodeSales2 = float64(personalRankTarget)
		}

		totalNetworkNodeSales2 := totalNetworkNodeSales
		if totalNetworkNodeSales > float64(totalNetworkRankTarget) {
			totalNetworkNodeSales2 = float64(totalNetworkRankTarget)
		}
		// if len(tbBonusRankStarPassupRst) > 0 {
		// 	fqty = tbBonusRankStarPassupRst[0].FQty
		// 	fQtyAcc = tbBonusRankStarPassupRst[0].FQtyAcc
		// }

		personalPurchaseRate := float.Div(totalPersonalNodeSales2, float64(personalRankTarget))
		personalRankAchieved := helpers.CutOffDecimal(totalPersonalNodeSales, 0, ".", ",")
		personalRankTargetString := strconv.Itoa(personalRankTarget)

		totalNetworkPurchaseRate := float.Div((totalNetworkNodeSales2), float64(totalNetworkRankTarget))
		totalNetworkRankAchieved := helpers.CutOffDecimal(totalNetworkNodeSales, 0, ".", ",")
		totalNetworkRankTargetString := strconv.Itoa(totalNetworkRankTarget)

		arrCriteriaList = append(arrCriteriaList,
			CriteriaListStruct{RankLabel: personalTranslated, RankPercent: personalPurchaseRate, RankAchieved: personalRankAchieved, RankTarget: personalRankTargetString},
			CriteriaListStruct{RankLabel: totalNetworkTranslated, RankPercent: totalNetworkPurchaseRate, RankAchieved: totalNetworkRankAchieved, RankTarget: totalNetworkRankTargetString},
		)
	} else if currentRank < 5 {
		dateFormat := base.ConvertFormat("yyyy-mm-dd")
		ytdDate := base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format(dateFormat)

		tbBonusRankStarPassupCond := make([]models.WhereCondFn, 0)
		tbBonusRankStarPassupCond = append(tbBonusRankStarPassupCond,
			models.WhereCondFn{Condition: "t_direct_sponsor_id = ?", CondValue: memID},
			models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: ytdDate},
			models.WhereCondFn{Condition: "t_member_id <> ?", CondValue: bigLegID},
		)
		numberOfDownlineRst, numberOfDownlineErr := models.GetTblMiningBonusRankStarPassupFn(tbBonusRankStarPassupCond, "SUM(t_downline_rank"+strconv.Itoa(currentRank)+") as total_downline", false)

		if numberOfDownlineErr != nil {
			base.LogErrorLog("GetMemberBZZMiningRankingInfo-GetTblMiningBonusRankStarPassupFn():2", numberOfDownlineErr.Error(), map[string]interface{}{"condition": tbBonusRankStarPassupCond}, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

		numOfRankTarget := 2
		minNetworkSalesTarget := 0

		if currentRank == 3 {
			minNetworkSalesTarget = 100
		} else if currentRank == 4 {
			minNetworkSalesTarget = 500
		} else if currentRank == 5 {
			minNetworkSalesTarget = 2000
		}

		if currentRank == 4 || currentRank == 5 {
			numOfRankTarget = 3
		}

		totalDownline := 0
		if len(numberOfDownlineRst) > 0 {
			totalDownline = numberOfDownlineRst[0].TotalDownline
		}
		numOfRankRate := float.Div(float64(totalDownline), float64(numOfRankTarget))
		numOfRankAchieved := helpers.CutOffDecimal(numOfRankRate, 0, ".", ",")
		numOfRankTargetString := strconv.Itoa(numOfRankTarget)

		// max 100%
		if numOfRankRate > 1 {
			numOfRankRate = 1
		}

		arrCriteriaList = append(arrCriteriaList,
			CriteriaListStruct{RankLabel: numOfRankTranslated, RankPercent: numOfRankRate, RankAchieved: numOfRankAchieved, RankTarget: numOfRankTargetString, RankLine: numOfRankTarget - 1},
		)

		if currentRank >= 3 {

			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " sponsor_lot.member_id = ? ", CondValue: memID},
				models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
				models.WhereCondFn{Condition: " sls_master.status IN ('AP', 'EP') AND sls_master.action = ? ", CondValue: "MINING_BZZ"},
				models.WhereCondFn{Condition: " sls_master.total_bv > ? ", CondValue: 0},
			)
			totalNetworkSalesRst, _ := models.GetTotalNetworkBZZSalesFn(arrCond, false)
			totalNetworkSales := float64(0)
			if totalNetworkSalesRst.TotalNodes > 0 {
				totalNetworkSales = totalNetworkSalesRst.TotalNodes
			}

			minNetworkSalesRate := float.Div(totalNetworkSales, float64(minNetworkSalesTarget))
			minNetworkSalesAchieved := helpers.CutOffDecimal(totalNetworkSales, 0, ".", ",")
			minNetworkSalesTargetString := strconv.Itoa(minNetworkSalesTarget)

			arrCriteriaList = append(arrCriteriaList,
				CriteriaListStruct{RankLabel: minNetworkSalesTranslated, RankPercent: minNetworkSalesRate, RankAchieved: minNetworkSalesAchieved, RankTarget: minNetworkSalesTargetString},
			)
		}
	} else {
		arrCriteriaList = append(arrCriteriaList,
			CriteriaListStruct{RankLabel: numOfRankTranslated, RankPercent: 1, RankAchieved: "3", RankTarget: "3", RankLine: 4},
			CriteriaListStruct{RankLabel: minNetworkSalesTranslated, RankPercent: 1, RankAchieved: "2000", RankTarget: "2000"},
		)
	}

	rankingInfo.CriteriaList = arrCriteriaList

	return &rankingInfo, nil
}

func CheckMemberAccessPermission(entMemberID int) bool {

	verificationRst := VerifyIfInNetwork(entMemberID, "BLK_LOGIN")
	if verificationRst {
		verificationRst = VerifyIfInNetwork(entMemberID, "UNBLK_LOGIN")
		if !verificationRst {
			return false
		}
	}

	return true
}

func GetTotalSubscriptionUser() int {
	var (
		totalSubscrUser int
		totalGhostUser  int
	)
	curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

	//get current registered member
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		models.WhereCondFn{Condition: "ent_member.join_date != ?", CondValue: curDate},
	)
	totalSubscrUserRst, _ := models.GetTotalActiveMemberFn(arrCond, false)
	totalSubscrUser = totalSubscrUserRst.TotalMember

	//get ghost member
	totalGhostUserRst, _ := models.GetTotalGhostMemberFn(nil, false)
	totalGhostUser = totalGhostUserRst.TotalMember

	totalSubscrUser = totalGhostUser + totalSubscrUser

	return totalSubscrUser
}

type SupportTicketStruct struct {
	MemberId     int
	CategoryCode string
	Title        string
	Msg          string
	LangCode     string
}

func (s *SupportTicketStruct) PostSupportTicket(tx *gorm.DB) (interface{}, error) {
	var (
		categoryCode string = s.CategoryCode
		ticketCode   string
		docType      string = "SPTK"
		err          error
	)

	// validate category code
	arrSupportTicketCategoryFn := []models.WhereCondFn{}
	arrSupportTicketCategoryFn = append(arrSupportTicketCategoryFn,
		models.WhereCondFn{Condition: " code = ? ", CondValue: categoryCode},
		models.WhereCondFn{Condition: " status = ? ", CondValue: "A"},
	)
	arrSupportTicketCategory, err := models.GetSupportTicketCategoryFn(arrSupportTicketCategoryFn, "", false)
	if err != nil {
		base.LogErrorLog("memberService:PostSupportTicket():GetSupportTicketCategoryFn():2", map[string]interface{}{"condition": arrSupportTicketCategoryFn}, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}
	if len(arrSupportTicketCategory) <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_category_code", s.LangCode), Data: err}
	}

	// get ticket code
	db := models.GetDB()
	ticketCode, err = models.GetRunningDocNo(docType, db) //get contract doc no
	if err != nil {
		base.LogErrorLog("memberService:PostSupportTicket():GetRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}
	err = models.UpdateRunningDocNo(docType, db) //update contract doc no
	if err != nil {
		base.LogErrorLog("memberService:PostSupportTicket():UpdateRunningDocNo():2", map[string]interface{}{"docType": docType}, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	arrSupportTicket := models.AddSupportTicketMastStruct{
		TicketCode:     ticketCode,
		TicketCategory: categoryCode,
		MemberID:       s.MemberId,
		TicketTitle:    s.Title,
		Status:         "P",
		AdminShow:      1,
		MemberShow:     1,
		CreatedBy:      strconv.Itoa(s.MemberId),
		CreatedAt:      time.Now(),
	}

	mastID, err := models.AddSupportTicketMast(tx, arrSupportTicket)
	if err != nil {
		base.LogErrorLog("PostSupportTicket-AddSupportTicketMast return Error", arrSupportTicket, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	arrSupportTicketDet := models.SupportTicketDet{
		TicketID:  mastID.ID,
		TicketMsg: s.Msg,
		CreatedBy: strconv.Itoa(s.MemberId),
		CreatedAt: time.Now(),
	}

	_, err = models.AddSupportTicketDet(tx, arrSupportTicketDet)
	if err != nil {
		base.LogErrorLog("PostSupportTicket-AddSupportTicketDet return Error", arrSupportTicketDet, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	arrData := make(map[string]interface{})
	arrData["ticket_no"] = ticketCode

	return arrData, nil

}

type SupportTicketListStruct struct {
	MemberID int    `json:"member_id"`
	Page     int64  `json:"page"`
	LangCode string `json:"lang_code"`
}

func (s *SupportTicketListStruct) GetMemberSupportTicketListv1() (interface{}, error) {

	type GetMemberSupportTicketListStructv1 struct {
		TicketCode  string `json:"ticket_code"`
		TicketTitle string `json:"ticket_title"`
		TicketMsg   string `json:"ticket_msg"`
		Category    string `json:"category"`
		// Address     string `json:"address"`
		StatusKey int    `json:"status_key"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	arrSysSupportTicketList := make([]GetMemberSupportTicketListStructv1, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "support_ticket_mast.member_id = ?", CondValue: s.MemberID},
	)

	arrSupportTicket, err := models.GetSupportTicketMastFn(arrCond, 0, false)

	if err != nil {
		base.LogErrorLog("GetMemberSupportTicketListv1-GetSupportTicketMastFn return Error", arrCond, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	if len(arrSupportTicket) > 0 {
		for _, arrSupportTicketV := range arrSupportTicket {
			var (
				firstMsg  = ""
				updatedAt = arrSupportTicketV.UpdatedAt.Format("2006-01-02 15:04:05")
			)

			// get first msg
			arrSupportTicketDetFn := []models.WhereCondFn{}
			arrSupportTicketDetFn = append(arrSupportTicketDetFn,
				models.WhereCondFn{Condition: "support_ticket_det.ticket_id = ?", CondValue: arrSupportTicketV.ID},
			)
			arrSupportTicketDet, _ := models.GetSupportTicketDetFn(arrSupportTicketDetFn, 1, "asc", false)
			if len(arrSupportTicketDet) > 0 {
				firstMsg = arrSupportTicketDet[0].TicketMsg
			}

			// set default updated at
			if updatedAt == "0001-01-01 00:00:00" {
				updatedAt = ""
			}

			status := helpers.Translate("closed", s.LangCode)
			statusKey := 0
			if arrSupportTicketV.Status != "AP" {
				status = helpers.Translate("open", s.LangCode)
				statusKey = 1
			}

			arrSysSupportTicketList = append(arrSysSupportTicketList,
				GetMemberSupportTicketListStructv1{
					TicketCode:  arrSupportTicketV.TicketCode,
					TicketTitle: arrSupportTicketV.TicketTitle,
					TicketMsg:   firstMsg,
					Category:    helpers.Translate(arrSupportTicketV.TicketCategoryName, s.LangCode),
					// Address:     arrSupportTicketV.Address,
					StatusKey: statusKey,
					Status:    status,
					CreatedAt: arrSupportTicketV.CreatedAt.Format("2006-01-02 15:04:05"),
					UpdatedAt: updatedAt,
				})
		}
	}

	//start paginate
	sort.Slice(arrSysSupportTicketList, func(p, q int) bool {
		return arrSysSupportTicketList[q].CreatedAt < arrSysSupportTicketList[p].CreatedAt
	})

	arrDataReturn := app.ArrDataResponseList{}

	//general setup default limit rows
	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	curPage := s.Page

	if curPage == 0 {
		curPage = 1
	}

	if s.Page != 0 {
		s.Page--
	}

	totalRecord := len(arrSysSupportTicketList)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(s.Page), int(limit), totalRecord)

	processArr := arrSysSupportTicketList[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(curPage),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      processArr,
		TableHeaderList:       nil,
	}

	return arrDataReturn, nil
}

type SupportTicketHistoryListStruct struct {
	TicketCode string `json:"ticket_code"`
	MemberID   int    `json:"member_id"`
	Page       int64  `json:"page"`
	LangCode   string `json:"lang_code"`
}

func (s *SupportTicketHistoryListStruct) GetMemberSupportTicketHistoryListv1() (interface{}, error) {

	type GetMemberSupportTicketHistoryListStructv1 struct {
		TicketTitle string      `json:"ticket_title"`
		TicketMsg   string      `json:"ticket_msg"`
		Category    string      `json:"category"`
		IssueImgUrl interface{} `json:"issue_img_url"`
		IssueVidUrl interface{} `json:"issue_vid_url"`
		Status      string      `json:"status"`
		StatusKey   int         `json:"status_key"`
		CreatedAt   string      `json:"created_at"`
		CreatedBy   string      `json:"created_by"`
	}

	arrSysSupportTicketList := make([]GetMemberSupportTicketHistoryListStructv1, 0)

	chatHistory, err := models.GetMemberSupportTicketDetailsByTicketCode(s.TicketCode)

	if err != nil {
		base.LogErrorLog("GetMemberSupportTicketHistoryListv1-GetMemberSupportTicketDetailsByTicketNo return Error", s.TicketCode, err.Error(), true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	if len(chatHistory) > 0 {
		for chatHistoryK, chatHistoryV := range chatHistory {

			splitUrl := strings.Split(chatHistoryV.FileURL1, ",")
			splitVid := strings.Split(chatHistoryV.FileURL2, ",")
			arrUrl := make([]interface{}, 0)
			arrVid := make([]interface{}, 0)

			for _, v := range splitUrl {
				if v != "" {
					arrUrl = append(arrUrl, v)
				}
			}

			for _, v := range splitVid {
				if v != "" {
					arrVid = append(arrVid, v)
				}
			}

			createdBy := helpers.Translate("you", s.LangCode)
			if chatHistoryV.CreatedUser != "" {
				// createdBy = helpers.Translate("admin", s.LangCode) + "-" + chatHistoryV.CreatedUser
				createdBy = helpers.Translate("admin", s.LangCode)
			}

			if chatHistoryK == 0 {
				status := helpers.Translate("closed", s.LangCode)
				statusKey := 0
				if chatHistoryV.Status != "AP" {
					status = helpers.Translate("open", s.LangCode)
					statusKey = 1
				}

				arrSysSupportTicketList = append(arrSysSupportTicketList,
					GetMemberSupportTicketHistoryListStructv1{
						TicketTitle: chatHistoryV.TicketTitle,
						TicketMsg:   chatHistoryV.TicketMsg,
						Category:    helpers.Translate(chatHistoryV.TicketCategoryName, s.LangCode),
						IssueImgUrl: arrUrl,
						IssueVidUrl: arrVid,
						Status:      status,
						StatusKey:   statusKey,
						CreatedAt:   chatHistoryV.CreatedAt.Format("2006-01-02 15:04:05"),
						CreatedBy:   createdBy,
					})
			} else {
				arrSysSupportTicketList = append(arrSysSupportTicketList,
					GetMemberSupportTicketHistoryListStructv1{
						TicketMsg:   chatHistoryV.TicketMsg,
						IssueImgUrl: arrUrl,
						IssueVidUrl: arrVid,
						CreatedAt:   chatHistoryV.CreatedAt.Format("2006-01-02 15:04:05"),
						CreatedBy:   createdBy,
					})
			}

		}
	}

	//start paginate
	sort.Slice(arrSysSupportTicketList, func(p, q int) bool {
		return arrSysSupportTicketList[q].CreatedAt > arrSysSupportTicketList[p].CreatedAt
	})

	arrDataReturn := app.ArrDataResponseList{}

	//general setup default limit rows
	arrLimitRowsSetting, _ := models.GetSysGeneralSetupByID("defaultlimitrow")

	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	curPage := s.Page

	if curPage == 0 {
		curPage = 1
	}

	if s.Page != 0 {
		s.Page--
	}

	totalRecord := len(arrSysSupportTicketList)

	totalPage := float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	pageStart, pageEnd := helpers.Paginate(int(s.Page), int(limit), totalRecord)

	processArr := arrSysSupportTicketList[pageStart:pageEnd]

	totalCurrentPageItems := len(processArr)

	perPage := int(limit)

	arrDataReturn = app.ArrDataResponseList{
		CurrentPage:           int(curPage),
		PerPage:               int(perPage),
		TotalCurrentPageItems: int(totalCurrentPageItems),
		TotalPage:             int(totalPage),
		TotalPageItems:        int(totalRecord),
		CurrentPageItems:      processArr,
		TableHeaderList:       nil,
	}

	return arrDataReturn, nil
}

func GetMemberSupportTicketCategoryList(langCode string) (interface{}, string) {
	var arrDataReturn = []map[string]interface{}{}

	arrSupportTicketCategoryFn := []models.WhereCondFn{}
	arrSupportTicketCategoryFn = append(arrSupportTicketCategoryFn, models.WhereCondFn{Condition: " status = ? ", CondValue: "A"})
	arrSupportTicketCategory, _ := models.GetSupportTicketCategoryFn(arrSupportTicketCategoryFn, "", false)

	if len(arrSupportTicketCategory) > 0 {
		for _, arrSupportTicketCategoryV := range arrSupportTicketCategory {
			arrDataReturn = append(arrDataReturn, map[string]interface{}{
				"code": arrSupportTicketCategoryV.Code,
				"name": helpers.TranslateV2(arrSupportTicketCategoryV.Name, langCode, map[string]string{}),
			})
		}
	}

	return arrDataReturn, ""
}

type SupportTicketReplyStruct struct {
	MemberId   int
	TicketCode string
	Msg        string
	LangCode   string
}

func (s *SupportTicketReplyStruct) PostSupportTicketReply(tx *gorm.DB) error {
	var (
		err error
	)

	//get mast with ticket no
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "support_ticket_mast.member_id = ?", CondValue: s.MemberId},
		models.WhereCondFn{Condition: "support_ticket_mast.ticket_code = ?", CondValue: s.TicketCode},
	)
	existingST, _ := models.GetSupportTicketMastFn(arrCond, 0, false)

	if len(existingST) < 1 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_ticket_code", s.LangCode), Data: err}
	}

	arrSupportTicketDet := models.SupportTicketDet{
		TicketID:  existingST[0].ID,
		TicketMsg: s.Msg,
		CreatedBy: strconv.Itoa(s.MemberId),
		CreatedAt: time.Now(),
	}

	_, err = models.AddSupportTicketDet(tx, arrSupportTicketDet)
	if err != nil {
		base.LogErrorLog("PostSupportTicket-AddSupportTicketDet return Error", arrSupportTicketDet, err.Error(), true)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	return nil

}

type SupportTicketCloseStruct struct {
	MemberId   int
	TicketCode string
	LangCode   string
}

func (s *SupportTicketCloseStruct) PostSupportTicketClose(tx *gorm.DB) error {
	var (
		err error
	)

	//get mast with ticket no
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "support_ticket_mast.member_id = ?", CondValue: s.MemberId},
		models.WhereCondFn{Condition: "support_ticket_mast.ticket_code = ?", CondValue: s.TicketCode},
	)
	existingST, _ := models.GetSupportTicketMastFn(arrCond, 0, false)

	if len(existingST) < 1 {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("invalid_ticket_no", s.LangCode), Data: err}
	}

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " support_ticket_mast.id = ? ", CondValue: existingST[0].ID},
	)

	updateColumn := make(map[string]interface{}, 0)
	updateColumn["status"] = "AP"

	err = models.UpdatesFn("support_ticket_mast", arrCond, updateColumn, false)

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	return nil

}

// share use by dashboard, dasahboard banner
type GetDashboardStruct struct {
	MemberID int
	LangCode string
	Type     string
}

func (d *GetDashboardStruct) GetMemberDashboard() (interface{}, error) {
	var (
		err           error
		arrDataReturn map[string]interface{}
	)

	//get dashboard banner
	banner := GetDashboardStruct{
		MemberID: d.MemberID,
		LangCode: d.LangCode,
		Type:     d.Type,
	}
	arrBanner, err := banner.GetMemberDashboardBanner()

	if err != nil {
		base.LogErrorLog("GetMemberDashboard - GetMemberDashboardBanner fail", err.Error(), map[string]interface{}{"err": err, "data": d}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", d.LangCode), Data: d}
	}

	arrDataReturn = map[string]interface{}{
		"banner_img": arrBanner,
	}

	return arrDataReturn, nil
}

func GetMemberTier(memID int) (string, error) {
	arrSlsTierFn := make([]models.WhereCondFn, 0)
	arrSlsTierFn = append(arrSlsTierFn,
		models.WhereCondFn{Condition: "sls_tier.member_id = ? ", CondValue: memID},
		models.WhereCondFn{Condition: "sls_tier.status = ? ", CondValue: "A"},
	)
	arrSlsTier, err := models.GetSlsTierFn(arrSlsTierFn, "", false)
	if err != nil {
		base.LogErrorLog("memberService:GetMemberNftTier()", "GetSlsTierFn():1", err.Error(), true)
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	var (
		tier = ""
	)
	if len(arrSlsTier) > 0 {
		tier = arrSlsTier[0].Tier
	}

	return tier, nil
}

func GetMemberTierAtSpecificTime(memID int, at string) (string, error) {
	arrSlsTierFn := make([]models.WhereCondFn, 0)
	arrSlsTierFn = append(arrSlsTierFn,
		models.WhereCondFn{Condition: "sls_tier.member_id = ? ", CondValue: memID},
		models.WhereCondFn{Condition: "sls_tier.created_at <= ? ", CondValue: at},
	)
	arrSlsTier, err := models.GetSlsTierFn(arrSlsTierFn, "", false)
	if err != nil {
		base.LogErrorLog("memberService:GetMemberNftTier()", "GetSlsTierFn():1", err.Error(), true)
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	var (
		tier = ""
	)
	if len(arrSlsTier) > 0 {
		tier = arrSlsTier[0].Tier
	}

	return tier, nil
}

func (d *GetDashboardStruct) GetMemberDashboardBanner() (interface{}, error) {
	var (
		arrReturn []map[string]interface{}
		err       error
		arrPopup  = make([]string, 0)
	)

	//get dashboard banner
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "sys_img.module = ?", CondValue: "DASHBOARD"},
		models.WhereCondFn{Condition: "FIND_IN_SET(?, sys_img.type)", CondValue: strings.ToUpper(d.Type)},
		models.WhereCondFn{Condition: "sys_img.status = ?", CondValue: "A"},
	)
	rst, err := models.GetSysImgFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetMemberDashboardBanner - GetSysImgFn fail", err.Error(), map[string]interface{}{"err": err, "data": d}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", d.LangCode), Data: d}
	}

	if rst != nil {
		for _, v := range rst {
			if v.PopupImg != "" {
				arrPopup = strings.Split(v.PopupImg, ",")
			}

			arrReturn = append(arrReturn, map[string]interface{}{
				"img_link":  v.ImgLink,
				"popup_img": arrPopup,
			})
		}
	}

	return arrReturn, nil
}

// Update2FAParam struct
type Update2FAParam struct {
	MemberID int
	Mode     string
}

// Update2FA func
func Update2FA(tx *gorm.DB, param Update2FAParam, langCode string) string {
	var (
		memID  int    = param.MemberID
		mode   string = param.Mode
		secret string = ""
	)

	if mode == "ON" {
		// check if already got secret key
		arrEntMember2FaFn := make([]models.WhereCondFn, 0)
		arrEntMember2FaFn = append(arrEntMember2FaFn,
			models.WhereCondFn{Condition: "ent_member_2fa.member_id = ? ", CondValue: memID},
		)
		arrEntMember2Fa, err := models.GetEntMember2FA(arrEntMember2FaFn, false)
		if err != nil {
			base.LogErrorLog("memberService:Update2FA():GetEntMember2FA():1", err.Error(), map[string]interface{}{"condition": arrEntMember2FaFn}, true)
			return "something_went_wrong"
		}
		if len(arrEntMember2Fa) > 0 {
			// update latest secret b_enable to 1
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: "id = ?", CondValue: arrEntMember2Fa[0].ID},
				models.WhereCondFn{Condition: "member_id = ?", CondValue: memID},
			)
			updateColumn := map[string]interface{}{"b_enable": 1}
			err = models.UpdatesFnTx(tx, "ent_member_2fa", arrUpdCond, updateColumn, false)
			if err != nil {
				base.LogErrorLog("memberService:Update2FA():UpdatesFnTx():1", err.Error(), map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, true)
				return "something_went_wrong"
			}
		} else {
			base.LogErrorLog("memberService:Update2FA()", "no_secret_available", map[string]interface{}{}, true)
			return "something_went_wrong"
		}

	} else if mode == "OFF" {
		// update b_enable to 0
		arrUpdCond := make([]models.WhereCondFn, 0)
		arrUpdCond = append(arrUpdCond, models.WhereCondFn{Condition: "member_id = ?", CondValue: memID})
		updateColumn := map[string]interface{}{"b_enable": 0}
		err := models.UpdatesFnTx(tx, "ent_member_2fa", arrUpdCond, updateColumn, false)
		if err != nil {
			base.LogErrorLog("memberService:Update2FA():UpdatesFnTx():2", err.Error(), map[string]interface{}{"arrUpdCond": arrUpdCond, "updateColumn": updateColumn}, true)
			return "something_went_wrong"
		}

		// generate new unique secret key
		charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		numChar := 10
		var validPassword bool
		for !validPassword {
			secret = base.GenerateRandomString(numChar, charSet)

			// validate if unique
			arrEntMember2FaFn2 := make([]models.WhereCondFn, 0)
			arrEntMember2FaFn2 = append(arrEntMember2FaFn2,
				models.WhereCondFn{Condition: "ent_member_2fa.secret = ? ", CondValue: secret},
			)
			arrEntMember2Fa2, err := models.GetEntMember2FA(arrEntMember2FaFn2, false)
			if err != nil {
				base.LogErrorLog("memberService:Update2FA():GetEntMember2FA():2", err.Error(), map[string]interface{}{"condition": arrEntMember2FaFn2}, true)
				return "something_went_wrong"
			}
			if len(arrEntMember2Fa2) == 0 {
				validPassword = true
			}
		}

		// save secret key
		arrAddEntMember2FA := models.AddEntMember2FaParam{
			MemberID: memID,
			Secret:   secret,
			CodeUrl:  "",
			BEnable:  0,
		}
		_, err = models.AddEntMember2FA(tx, arrAddEntMember2FA)
		if err != nil {
			base.LogErrorLog("memberService:Update2FA():GetEntMember2FA():1", err.Error(), map[string]interface{}{"data": arrAddEntMember2FA}, true)
			return "something_went_wrong"
		}
	} else {
		return "invalid_mode"
	}

	return ""
}

type GetStrategyRankingStruct struct {
	MemberID int
	LangCode string
	Type     string
	Market   string
}

func (s *GetStrategyRankingStruct) GetStrategyRanking() (interface{}, error) {
	var (
		// err           error
		arrDataReturn     map[string]interface{}
		arrResult         = make([]interface{}, 0)
		defaultProfilePic = "https://media02.securelayers.cloud/medias/GTA/MEMBER/IMAGES/PROFILE/default.png"
	)

	if strings.ToUpper(s.Type) == "MAIN" {
		//get from bot_leaderboard
		//0:费率,1:指数,2:网格,3:马丁格,4:反向马丁格

		var (
			Type0          string //crypto_funding_rates_arbitrage
			Type1          string //crypto_index_funding_rates_arbitrage
			Type2          string //spot_grid_trading
			Type3          string //martingale_trading
			Type4          string //reverse_martingale_trding
			translatedType string
			MarketType     int
		)

		//get prd_master
		arrPrdMasterFn := make([]models.WhereCondFn, 0)
		arrPrdMasterFn = append(arrPrdMasterFn,
			models.WhereCondFn{Condition: "prd_master.prd_group = ? ", CondValue: "BOT"},
		)
		arrPrdMaster, _ := models.GetPrdMasterFn(arrPrdMasterFn, "", false)

		if len(arrPrdMaster) > 0 {
			for _, v := range arrPrdMaster {
				switch strings.ToUpper(v.Code) {
				case "CFRA":
					Type0 = helpers.Translate(v.Name, s.LangCode)
				case "CIFRA":
					Type1 = helpers.Translate(v.Name, s.LangCode)
				case "SGT":
					Type2 = helpers.Translate(v.Name, s.LangCode)
				case "MT":
					Type3 = helpers.Translate(v.Name, s.LangCode)
				case "MTD":
					Type4 = helpers.Translate(v.Name, s.LangCode)
				}

			}
		}

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "bot_leaderboard.b_latest = ?", CondValue: 1},
		)

		if s.Market != "" {
			switch strings.ToUpper(s.Market) {
			case "CFRA":
				MarketType = 0
			case "CIFRA":
				MarketType = 1
			case "SGT":
				MarketType = 2
			case "MT":
				MarketType = 3
			case "MTD":
				MarketType = 4
			}
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "bot_leaderboard.type = ?", CondValue: MarketType},
			)

		}

		result, err := models.GetStrategyLeaderboardFn(arrCond, 10, false)

		if err != nil {
			base.LogErrorLog("GetStrategyRanking-GetStrategyLeaderboardFn", err.Error(), map[string]interface{}{"cond": arrCond}, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
		}

		if len(result) > 0 {
			for _, v := range result {
				switch v.Type {
				case "0":
					translatedType = Type0
				case "1":
					translatedType = Type1
				case "2":
					translatedType = Type2
				case "3":
					translatedType = Type3
				case "4":
					translatedType = Type4
				}

				arrResult = append(arrResult, map[string]interface{}{
					"type":           translatedType,
					"symbol":         v.Symbol,
					"total_earnings": helpers.CutOffDecimal(v.TotalEarnings, 10, ".", ","),
					"ratio":          strconv.Itoa(v.Ratio),
					"dt_timestamp":   v.DtTimestamp.Format("2006-01-02 15:04:05"),
				})
			}
		}

		arrDataReturn = map[string]interface{}{
			"status": 0,
			"list":   arrResult,
		}
	} else if strings.ToUpper(s.Type) == "A-LIST1" {
		//get from tblq_bonus_strategy_profit
		var (
			Type0          string //crypto_funding_rates_arbitrage
			Type1          string //crypto_index_funding_rates_arbitrage
			Type2          string //spot_grid_trading
			Type3          string //martingale_trading
			Type4          string //reverse_martingale_trding
			translatedType string
		)
		// curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

		//get prd_master
		arrPrdMasterFn := make([]models.WhereCondFn, 0)
		arrPrdMasterFn = append(arrPrdMasterFn,
			models.WhereCondFn{Condition: "prd_master.prd_group = ? ", CondValue: "BOT"},
		)
		arrPrdMaster, _ := models.GetPrdMasterFn(arrPrdMasterFn, "", false)

		if len(arrPrdMaster) > 0 {
			for _, v := range arrPrdMaster {
				switch strings.ToUpper(v.Code) {
				case "CFRA":
					Type0 = helpers.Translate(v.Name, s.LangCode)
				case "CIFRA":
					Type1 = helpers.Translate(v.Name, s.LangCode)
				case "SGT":
					Type2 = helpers.Translate(v.Name, s.LangCode)
				case "MT":
					Type3 = helpers.Translate(v.Name, s.LangCode)
				case "MTD":
					Type4 = helpers.Translate(v.Name, s.LangCode)
				}
			}
		}

		arrCond := make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: "tblq_bonus_strategy_profit.bns_id = ?", CondValue: curDate}, //get only today record
		// )

		if s.Market != "" {
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "tblq_bonus_strategy_profit.strategy = ?", CondValue: strings.ToUpper(s.Market)},
			)
		}

		result, err := models.GetLeaderboardTypeAFn(arrCond, 10, false)

		if err != nil {
			base.LogErrorLog("GetStrategyRanking-GetLeaderboardTypeAFn", err.Error(), map[string]interface{}{"cond": arrCond}, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
		}

		if len(result) > 0 {
			for _, v := range result {
				switch v.Strategy {
				case "CFRA":
					translatedType = Type0
				case "CIFRA":
					translatedType = Type1
				case "SGT":
					translatedType = Type2
				case "MT":
					translatedType = Type3
				case "MTD":
					translatedType = Type4
				default:
					translatedType = helpers.Translate(v.Strategy, s.LangCode)
				}

				if v.ProfilePic == "" {
					v.ProfilePic = defaultProfilePic
				}

				arrResult = append(arrResult, map[string]interface{}{
					"username":         v.Username,
					"profile_pic":      v.ProfilePic,
					"type":             translatedType,
					"symbol":           v.CryptoPair,
					"total_earnings":   helpers.CutOffDecimal(v.FProfit, 10, ".", ","),
					"referral_rewards": helpers.CutOffDecimal(v.SponsorReward, 10, ".", ","),
					"dt_timestamp":     v.DtTimestamp.Format("2006-01-02 15:04:05"),
				})
			}
		}

		arrDataReturn = map[string]interface{}{
			"status": 0,
			"list":   arrResult,
		}
	} else {
		type arrHeaderSettingStruct struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		}

		// show different event leaderboard
		arrStrategyEventsDetailsFn := make([]models.WhereCondFn, 0)
		arrStrategyEventsDetailsFn = append(arrStrategyEventsDetailsFn,
			models.WhereCondFn{Condition: "strategy_events.package = ?", CondValue: strings.ToUpper(s.Type)},
			models.WhereCondFn{Condition: "strategy_events.status = ?", CondValue: "A"},
			// models.WhereCondFn{Condition: "strategy_events.time_start <= ?", CondValue: time.Now()}, // always show unless status != 'A'
			// models.WhereCondFn{Condition: "strategy_events.time_end >= ?", CondValue: time.Now()}, // always show unless status != 'A'
		)
		arrStrategyEventsDetails, err := models.GetStrategyEventsDetailsFn(arrStrategyEventsDetailsFn, "", false)

		if err != nil {
			base.LogErrorLog("GetStrategyRanking-GetStrategyEventsFn", err.Error(), map[string]interface{}{"condition": arrStrategyEventsDetailsFn}, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
		}

		var (
			associateTranslated       = helpers.Translate("associate", s.LangCode)
			partnerTranslated         = helpers.Translate("partner", s.LangCode)
			seniorPartnerTranslated   = helpers.Translate("senior_partner", s.LangCode)
			regionalPartnerTranslated = helpers.Translate("regional_partner", s.LangCode)
			globalPartnerTranslated   = helpers.Translate("global_partner", s.LangCode)
			directorTranslated        = helpers.Translate("director", s.LangCode)
		)

		for _, arrStrategyEventsDetailsV := range arrStrategyEventsDetails {
			arrList := make([]interface{}, 0)

			var (
				image       string = ""
				title       string = ""
				description string = ""
			)

			// get header list
			var arrHeaderListSetting []arrHeaderSettingStruct
			json.Unmarshal([]byte(arrStrategyEventsDetailsV.TableHeader), &arrHeaderListSetting)

			if arrStrategyEventsDetailsV.SqlQuery != "" {
				// replace from_datetime and to_datetime
				sqlQuery := strings.Replace(arrStrategyEventsDetailsV.SqlQuery, "#from_datetime#", arrStrategyEventsDetailsV.TimeStart.Format("2006-01-02 15:04:05"), -1)
				sqlQuery = strings.Replace(sqlQuery, "#to_datetime#", arrStrategyEventsDetailsV.TimeEnd.Format("2006-01-02 15:04:05"), -1)

				arrStrategyEventsQuery, _ := models.GetStrategyEventsQuery(sqlQuery, false)

				if len(arrStrategyEventsQuery) > 0 {
					for _, arrStrategyEventsQueryV := range arrStrategyEventsQuery {
						var (
							rank            string = "-"
							packageTier     string
							packageTierLink string
							arrListData     = make([]interface{}, 0)
						)

						for _, arrHeaderListSettingV := range arrHeaderListSetting {
							headerTranslated := helpers.Translate(arrHeaderListSettingV.Name, s.LangCode)
							value := ""

							if arrHeaderListSettingV.Key == "nick_name" {
								value = arrStrategyEventsQueryV.NickName
								title = arrStrategyEventsQueryV.NickName
							}

							if arrHeaderListSettingV.Key == "avatar" {
								avatar := defaultProfilePic
								if arrStrategyEventsQueryV.Avatar != "" {
									value = arrStrategyEventsQueryV.Avatar
								}

								value = avatar
								image = avatar
							}

							if arrHeaderListSettingV.Key == "total_amount" {
								value = helpers.CutOffDecimal(arrStrategyEventsQueryV.TotalAmount, 2, ".", ",")
								description = helpers.CutOffDecimal(arrStrategyEventsQueryV.TotalAmount, 2, ".", ",")
							} else if arrHeaderListSettingV.Key == "total_amount_4dec" {
								value = helpers.CutOffDecimal(arrStrategyEventsQueryV.TotalAmount, 4, ".", ",")
								description = helpers.CutOffDecimal(arrStrategyEventsQueryV.TotalAmount, 4, ".", ",")
							} else if arrHeaderListSettingV.Key == "total_amount_6dec" {
								value = helpers.CutOffDecimal(arrStrategyEventsQueryV.TotalAmount, 6, ".", ",")
								description = helpers.CutOffDecimal(arrStrategyEventsQueryV.TotalAmount, 6, ".", ",")
							}

							if arrHeaderListSettingV.Key == "country" {
								// get country img
								adminServerDomain := setting.Cfg.Section("custom").Key("AdminServerDomain").String()
								value = adminServerDomain + "/assets/global/img/512_flags/" + strings.ToLower(arrStrategyEventsQueryV.CountryCode) + ".png"
							}

							if arrHeaderListSettingV.Key == "rank" {
								// get rank
								rankArr, _ := models.GetCommunityBonusByMemberId(arrStrategyEventsQueryV.MemberID, "", "")

								if len(rankArr) > 0 {
									if rankArr[0].FPerc <= 0.01 {
										rank = associateTranslated
									} else if rankArr[0].FPerc <= 0.03 {
										rank = partnerTranslated
									} else if rankArr[0].FPerc <= 0.05 {
										rank = seniorPartnerTranslated
									} else if rankArr[0].FPerc <= 0.07 {
										rank = regionalPartnerTranslated
									} else if rankArr[0].FPerc <= 0.10 {
										rank = globalPartnerTranslated
									} else {
										rank = directorTranslated
									}
								}

								value = rank
							}

							if arrHeaderListSettingV.Key == "total_package_amount" {
								// get total package amount
								arrMemberTotalSalesFn := make([]models.WhereCondFn, 0)
								arrMemberTotalSalesFn = append(arrMemberTotalSalesFn,
									models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: arrStrategyEventsQueryV.MemberID},
									models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
									// models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
									models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
								)
								arrMemberTotalSales, _ := models.GetMemberTotalSalesFn(arrMemberTotalSalesFn, false)

								value = helpers.CutOffDecimal(arrMemberTotalSales.TotalAmount, 0, ".", ",")
							}

							if arrHeaderListSettingV.Key == "total_reward_amount" {
								// get total reward amount
								memberRevenue, _ := models.GetMemberTotalBns(arrStrategyEventsQueryV.MemberID, "", 0, 0, 0)

								totalRewardAmount := 0.00
								if memberRevenue != nil {
									totalRewardAmount = memberRevenue.TotalBonus
								}

								value = helpers.CutOffDecimal(totalRewardAmount, 2, ".", ",")
							}

							if arrHeaderListSettingV.Key == "package_tier_name" {
								// get package tier
								packageTier, _ = GetMemberTier(arrStrategyEventsQueryV.MemberID)
								if packageTier == "" {
									// no tier
									packageTier = "B0"
								}

								value = packageTier
							}

							if arrHeaderListSettingV.Key == "package_tier" {
								if packageTier != "" {
									packageTierLink = "https://media02.securelayers.cloud/medias/GTA/PACKAGE/ICON/" + packageTier + ".jpg"
								}

								value = packageTierLink
							}

							if arrHeaderListSettingV.Key == "platform" {
								// get bot platform
								memberCurrentAPI := GetMemberCurrentAPI(arrStrategyEventsQueryV.MemberID)

								value = memberCurrentAPI.Platform
							}

							if arrHeaderListSettingV.Key == "total_today_bot_profit" {
								// get total bot profit
								arrGroupTblqBonusStrategyProfitFn := []models.WhereCondFn{}
								arrGroupTblqBonusStrategyProfitFn = append(arrGroupTblqBonusStrategyProfitFn,
									models.WhereCondFn{Condition: "tblq_bonus_strategy_profit.member_id = ?", CondValue: arrStrategyEventsQueryV.MemberID},
									models.WhereCondFn{Condition: "tblq_bonus_strategy_profit.bns_id = ?", CondValue: time.Now().AddDate(0, 0, -1).Format("2006-01-02")},
								)
								arrGroupTblqBonusStrategyProfit, _ := models.GetGroupTblqBonusStrategyProfitFn(arrGroupTblqBonusStrategyProfitFn, false)

								value = "0.00"
								if len(arrGroupTblqBonusStrategyProfit) > 0 {
									value = helpers.CutOffDecimalv2(arrGroupTblqBonusStrategyProfit[0].TotalAmount, 6, ".", ",", true)
								}
							}

							if arrHeaderListSettingV.Key == "total_today_bot_bonus" {
								// get total bot bonus
								arrGroupTblqBonusStrategySponsorFn := []models.WhereCondFn{}
								arrGroupTblqBonusStrategySponsorFn = append(arrGroupTblqBonusStrategySponsorFn,
									models.WhereCondFn{Condition: "tblq_bonus_strategy_sponsor.member_id = ?", CondValue: arrStrategyEventsQueryV.MemberID},
									models.WhereCondFn{Condition: "tblq_bonus_strategy_sponsor.bns_id = ?", CondValue: time.Now().AddDate(0, 0, -1).Format("2006-01-02")},
								)
								arrGroupTblqBonusStrategySponsor, _ := models.GetGroupTblqBonusStrategySponsorFn(arrGroupTblqBonusStrategySponsorFn, false)

								value = "0.00"
								if len(arrGroupTblqBonusStrategySponsor) > 0 {
									value = helpers.CutOffDecimalv2(arrGroupTblqBonusStrategySponsor[0].TotalAmount, 6, ".", ",", true)
								}
							}

							arrListData = append(arrListData, map[string]interface{}{
								"header": headerTranslated,
								"value":  value,
							})
						}

						arrList = append(arrList, map[string]interface{}{
							"image":       image,
							"title":       title,
							"description": description,
							"list":        arrListData,
						})
					}
				}
			}

			arrPrize := strings.Split(arrStrategyEventsDetailsV.Path, ",")

			arrResult = append(arrResult, map[string]interface{}{
				"event_title":          arrStrategyEventsDetailsV.Title,
				"event_description":    arrStrategyEventsDetailsV.Desc,
				"prize_img_url":        arrPrize,
				"data_arr":             arrList,
				"event_start_datetime": helpers.ConvertTimeToUnix(arrStrategyEventsDetailsV.TimeStart),
				"event_end_datetime":   helpers.ConvertTimeToUnix(arrStrategyEventsDetailsV.TimeEnd),
			})
		}

		arrDataReturn = map[string]interface{}{
			"status": 1,
			"events": arrResult,
		}
	}

	return arrDataReturn, nil
}

type GetCryptoPriceStruct struct {
	MemberID int
	LangCode string
}

func (s *GetCryptoPriceStruct) GetCryptoPrice() (interface{}, error) {
	var (
		// err           error
		arrDataReturn map[string]interface{}
		arrResult     = make([]interface{}, 0)
	)
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "crypto_price_movement.b_latest = ?", CondValue: 1},
	)

	result, err := models.GetCryptoPriceMovementFn(arrCond, "", false)

	if err != nil {
		base.LogErrorLog("GetCryptoPrice-GetCryptoPriceFn", err.Error(), map[string]interface{}{"cond": arrCond}, true)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", s.LangCode), Data: err}
	}

	if len(result) > 0 {
		for _, v := range result {

			lastDate := v.CreatedAt.AddDate(0, 0, -1)
			arrCond24h := make([]models.WhereCondFn, 0)
			arrCond24h = append(arrCond24h,
				models.WhereCondFn{Condition: "crypto_price_movement.b_latest = ?", CondValue: 0},
				models.WhereCondFn{Condition: "crypto_price_movement.created_at <= ?", CondValue: lastDate},
				models.WhereCondFn{Condition: "crypto_price_movement.code = ?", CondValue: v.Code},
			)

			arrResultPrice24h, _ := models.GetCryptoPriceMovementFn(arrCond24h, "", false)

			lastPrice := float64(0)
			if len(arrResultPrice24h) > 0 {
				lastPrice = arrResultPrice24h[0].Price
			}

			chargePerc := ((v.Price - lastPrice) / lastPrice) * 100
			chargeColor := "#ea4435"

			if chargePerc > 0 {
				chargeColor = "#5dba7c"
			}

			arrResult = append(arrResult, map[string]interface{}{
				"symbol":       strings.ToUpper(v.Code) + "/USDT",
				"price":        helpers.CutOffDecimal(v.Price, 10, ".", ","),
				"24hchange":    helpers.CutOffDecimal(chargePerc, 2, ".", ",") + "%",
				"24hcolor":     chargeColor,
				"dt_timestamp": v.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	arrDataReturn = map[string]interface{}{
		"list": arrResult,
	}

	return arrDataReturn, nil
}

type MemberPdf struct {
	PdfImg string `json:"pdf_img"`
}

func GetMemberPdf(memberID int, docType, langCode string) (MemberPdf, string) {
	var (
		data               = MemberPdf{}
		packageType string = "a"
		lang        string = "en"
		apiKeys     string = ""
	)

	if docType == "PACKAGE_B" {
		packageType = "b"
	}

	// api/genMemberContractPdf
	// lang = en / cn
	// package = a / b
	// memberid = xx

	// get admin domain
	arrAdminDomainSetting, _ := models.GetSysGeneralSetupByID("admin_domain")
	url := arrAdminDomainSetting.SettingValue1 + "/api/genMemberContractPdf"

	// api keys
	arrApiKeysFn := []models.WhereCondFn{}
	arrApiKeysFn = append(arrApiKeysFn, models.WhereCondFn{Condition: " name = ? ", CondValue: "admin"})
	arrApiKeys, err := models.GetApiKeysFn(arrApiKeysFn, "", false)
	if err != nil {
		base.LogErrorLog("memberService:GetMemberPdf():GetApiKeysFn()", err.Error(), map[string]interface{}{"condition": arrApiKeysFn}, true)
		return MemberPdf{}, "something_went_wrong"
	}
	if len(arrApiKeys) <= 0 {
		base.LogErrorLog("memberService:GetMemberPdf():GetApiKeysFn()", "api_keys_not_found", map[string]interface{}{"condition": arrApiKeysFn}, true)
		return MemberPdf{}, "something_went_wrong"
	}

	apiKeys = arrApiKeys[0].Key

	header := map[string]string{
		"Content-Type":    "application/json",
		"X-Authorization": apiKeys,
	}

	if langCode == "zh" {
		lang = "cn"
	}

	body := map[string]interface{}{
		"memberid": memberID,
		"package":  packageType,
		"lang":     lang,
	}

	response, err := base.RequestAPIV2("POST", url, header, body, nil, base.ExtraSettingStruct{})
	if err != nil {
		base.LogErrorLog("memberService:GetMemberPdf():RequestAPIV2()", err.Error(), map[string]interface{}{"url": url, "header": header, "body": body}, true)
		return MemberPdf{}, "something_went_wrong"
	}

	type PostGenerateMemberContractAPIResponse struct {
		Rst  int    `json:"rst"`
		Msg  string `json:"msg"`
		Path string `json:"path"`
	}
	var postGenerateMemberContractApiResponse = &PostGenerateMemberContractAPIResponse{}
	if response.Body == "" {
		return MemberPdf{}, "something_went_wrong"
	}

	err = json.Unmarshal([]byte(response.Body), postGenerateMemberContractApiResponse)
	if err != nil {
		base.LogErrorLog("memberService:GetMemberPdf():Unmarshal():1", err.Error(), map[string]interface{}{"input": response.Body}, true)
		return MemberPdf{}, "something_went_wrong"
	}

	if postGenerateMemberContractApiResponse.Rst != 1 {
		base.LogErrorLog("memberService:GetMemberPdf()", postGenerateMemberContractApiResponse.Msg, map[string]interface{}{"url": url, "header": header, "body": body, "response_body": postGenerateMemberContractApiResponse}, true)
		return MemberPdf{}, "something_went_wrong"
	}

	data.PdfImg = postGenerateMemberContractApiResponse.Path

	return data, ""
}

type MemberCurrentAPI struct {
	Platform     string
	PlatformCode string
	ApiDetails   []MemberCurrentAPIDetails
}

type MemberCurrentAPIDetails struct {
	Module        string
	ApiKey        string
	ApiSecret     string
	ApiPassphrase string
}

func GetMemberCurrentAPI(memberID int) MemberCurrentAPI {
	var (
		memberCurrentApi        = MemberCurrentAPI{}
		memberCurrentApiDetails = []MemberCurrentAPIDetails{}
	)

	// get current active limit
	arrEntMemberTradingApiFn := []models.WhereCondFn{}
	arrEntMemberTradingApiFn = append(arrEntMemberTradingApiFn,
		models.WhereCondFn{Condition: "ent_member_trading_api.member_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member_trading_api.status = ?", CondValue: "A"},
	)
	arrEntMemberTradingApi, _ := models.GetEntMemberTradingApiFn(arrEntMemberTradingApiFn, "", false)
	if len(arrEntMemberTradingApi) <= 0 {
		return memberCurrentApi
	}

	memberCurrentApi.Platform = arrEntMemberTradingApi[0].Platform
	memberCurrentApi.PlatformCode = arrEntMemberTradingApi[0].PlatformCode

	for _, arrEntMemberTradingApiV := range arrEntMemberTradingApi {
		memberCurrentApiDetails = append(memberCurrentApiDetails,
			MemberCurrentAPIDetails{
				Module:        arrEntMemberTradingApiV.Module,
				ApiKey:        arrEntMemberTradingApiV.ApiKey,
				ApiSecret:     arrEntMemberTradingApiV.ApiSecret,
				ApiPassphrase: arrEntMemberTradingApiV.ApiPassphrase,
			},
		)
	}

	memberCurrentApi.ApiDetails = memberCurrentApiDetails

	return memberCurrentApi
}
