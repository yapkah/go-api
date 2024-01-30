package member_service

import (
	"encoding/json"
	"fmt"

	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
)

// ExistsMemberByEmail func
func (m *Member) ExistsMemberByEmail() (bool, error) {
	return models.ExistsMemberByEmail(m.Email)
}

// ExistsMemberByMobile func
func (m *Member) ExistsMemberByMobile() (bool, error) {
	return models.ExistsMemberByMobile(m.MobilePrefix, m.MobileNo)
}

// ExistsMemberByUsername func
func ExistsMemberByUsername(username string) (bool, error) {
	return models.ExistsMemberByUsername(username)
}

// ValidateReferralCode func
func ValidateReferralCode(referralCode string) (string, *models.EntMember) {
	// referralNickName, err := util.DecodeBase64(referralCode)
	// if err != nil { // decode failed
	// 	return "invalid_referral_code", nil
	// }
	arrEntMemberCond := make([]models.WhereCondFn, 0)
	arrEntMemberCond = append(arrEntMemberCond,
		// models.WhereCondFn{Condition: "ent_member.code = ?", CondValue: referralCode},
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: referralCode},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrEntMember, err := models.GetEntMemberFn(arrEntMemberCond, "", false)
	if err != nil {
		base.LogErrorLog("memberService:ValidateReferralCode()", "GetEntMemberFn():1", err.Error(), true)
		return "something_went_wrong", nil
	}

	if arrEntMember == nil {
		return "invalid_referral_code", nil
	}

	return "", arrEntMember
}

// ValidatePlacementCode func
func ValidatePlacementCode(memID int, referralID int, placementCode string, legNo int) (string, *models.EntMember) {
	arrSysGeneralSetup, _ := models.GetSysGeneralSetupByID("placement_setting")
	arrPlacementSetting := arrSysGeneralSetup.InputValue1

	arrPlacementSettingPointer := &PlacementSetting{}
	err := json.Unmarshal([]byte(arrPlacementSetting), arrPlacementSettingPointer)
	if err != nil {
		base.LogErrorLog("memberService:ValidatePlacementCode():Unmarshal():1", err.Error(), map[string]interface{}{"arrPlacementSetting": arrPlacementSetting}, true)
		return "something_went_wrong", nil
	}

	// validate leg placement
	if arrPlacementSettingPointer.Status {
		if placementCode == "" {
			return "please_provide_placement_code", nil
		}

		// validate placement group
		if legNo <= 0 || legNo > arrPlacementSettingPointer.MaxLeg {
			return "please_provide_valid_placement_group", nil
		}

		// validate placement code
		arrEntMemberFn := make([]models.WhereCondFn, 0)
		arrEntMemberFn = append(arrEntMemberFn,
			models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: placementCode},
			models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
		)
		arrEntMember, err := models.GetEntMemberFn(arrEntMemberFn, "", false)
		if err != nil {
			base.LogErrorLog("memberService:ValidatePlacementCode():GetEntMemberFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberFn}, true)
			return "something_went_wrong", nil
		}

		if arrEntMember == nil {
			return "invalid_placement_code", nil
		}

		if arrEntMember.ID == memID {
			return "invalid_placement_code", nil
		}

		// validate if placement code is already placed
		arrEntMemberSponsorTreeFn := make([]models.WhereCondFn, 0)
		arrEntMemberSponsorTreeFn = append(arrEntMemberSponsorTreeFn,
			models.WhereCondFn{Condition: "ent_member_tree_sponsor.member_id = ?", CondValue: arrEntMember.ID},
		)
		arrEntMemberSponsorTree, err := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberSponsorTreeFn, false)
		if err != nil {
			base.LogErrorLog("memberService:ValidatePlacementCode():GetEntMemberEntMemberTreeSponsorFn():1", err.Error(), map[string]interface{}{"condition": arrEntMemberSponsorTreeFn}, true)
			return "something_went_wrong", nil
		}

		if arrEntMemberSponsorTree == nil {
			base.LogErrorLog("memberService:ValidatePlacementCode():GetEntMemberEntMemberTreeSponsorFn():1", "ent_member_tree_sponsor_not_found", map[string]interface{}{"condition": arrEntMemberSponsorTreeFn}, true)
			return "something_went_wrong", nil
		}

		if arrEntMember.ID != 1 && arrEntMemberSponsorTree.UplineID == 0 {
			return "placement_id_not_yet_bind_placement", nil
		}

		// validate placement code is in placement tree network
		nearestUpline := GetNearestUplineByMemberID(arrEntMember.ID, []string{fmt.Sprint(referralID)}, "PLACEMENT")
		if !nearestUpline.Status {
			base.LogErrorLog("memberService:ValidatePlacementCode():GetNearestUplineByMemberID():1", nearestUpline.ErrMsg, map[string]interface{}{"memberID": arrEntMember.ID, "arrTargetID": []string{fmt.Sprint(referralID)}, "networkType": "SPONSOR"}, true)
			return "something_went_wrong", nil
		}
		if nearestUpline.UplineID == 0 {
			return "placement_code_not_in_sponsor_network", nil
		}

		// validate if leg is taken
		arrEntMemberSponsorTreeFn2 := make([]models.WhereCondFn, 0)
		arrEntMemberSponsorTreeFn2 = append(arrEntMemberSponsorTreeFn2,
			models.WhereCondFn{Condition: "ent_member_tree_sponsor.upline_id = ?", CondValue: arrEntMember.ID},
			models.WhereCondFn{Condition: "ent_member_tree_sponsor.leg_no = ?", CondValue: legNo},
		)
		arrEntMemberSponsorTree2, err := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberSponsorTreeFn2, false)
		if err != nil {
			base.LogErrorLog("memberService:ValidatePlacementCode():GetEntMemberEntMemberTreeSponsorFn():2", err.Error(), map[string]interface{}{"condition": arrEntMemberSponsorTreeFn2}, true)
			return "something_went_wrong", nil
		}

		if arrEntMemberSponsorTree2 != nil {
			return "placement_already_taken", nil
		}

		return "", arrEntMember

	}

	return "", nil
}
