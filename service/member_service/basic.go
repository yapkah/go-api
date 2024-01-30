package member_service

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
)

// GetRandUsername func
func GetRandUsername() (string, string) {
	var unique = false
	var username = ""
	var maxAttempts, counts = 50, 0

	for !unique { // loop until it is unique
		usernameTemp, errMsg := GenerateRandUsername()

		if errMsg != "" {
			base.LogErrorLog("memberService:GetRandUsername()", "GenerateRandUsername():1", errMsg, true) // log but still continue attempt to get random username
		} else {
			// check if unique
			arrEntMemberFn := make([]models.WhereCondFn, 0)
			arrEntMemberFn = append(arrEntMemberFn,
				models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: usernameTemp},
			)
			arrEntMember, _ := models.GetEntMemberFn(arrEntMemberFn, "", false)

			if arrEntMember == nil {
				unique = true
				username = usernameTemp
			}
		}

		counts++

		if counts > maxAttempts {
			base.LogErrorLog("memberService:GetRandUsername()", "", "reached_max_attempts_of_"+strconv.Itoa(maxAttempts), true)
			return "", "something_went_wrong"
		}
	}

	return username, ""
}

// GenerateRandUsername func
func GenerateRandUsername() (string, string) {
	var vowels = []string{"a", "e", "i", "o", "u"}
	var consonants = []string{"b", "c", "d", "f", "g", "h", "j", "k", "l", "m", "n", "p", "q", "r", "s", "t", "v", "w", "x", "y", "z"}
	var numbers = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	var minLength = 4
	var maxLength = 12 // current username maxlength is 19, put 12 is enuf to gen random username

	// pattern declaration
	var arrPattern [12]string
	arrPattern[7] = "CVCCVNN"
	arrPattern[8] = "CVCCVNNN"
	arrPattern[9] = "CVCCVVCNN"

	var length = rand.Intn(maxLength-minLength) + minLength
	var numStarFrom = int(length/2) + 1
	var username = ""

	// begin looping to draw username
	for i := 0; i < length; i++ {
		if arrPattern[length] != "" { // with certain pattern
			s := strings.Split(arrPattern[length], "")

			if s[i] == "C" {
				username += consonants[rand.Intn(len(consonants))]
			} else if s[i] == "V" {
				username += vowels[rand.Intn(len(vowels))]
			} else if s[i] == "N" {
				username += numbers[rand.Intn(len(numbers))]
			} else {
				return "", "invalid_pattern_" + arrPattern[length]
			}

		} else { // without certain pattern
			if i < numStarFrom {
				if i%2 == 0 { // even number [consonants]
					username += consonants[rand.Intn(len(consonants))]
				} else { // odd number [vowels]
					username += vowels[rand.Intn(len(vowels))]
				}
			} else {
				username += numbers[rand.Intn(len(numbers))]
			}
		}
	}

	return username, ""
}

type PlacementSetting struct {
	Status bool `json:"status"`
	MaxLeg int  `json:"max_leg"`
}

func GetPlacementLegOption(placementCode string, langCode string) []map[string]interface{} {
	arrSysGeneralSetup, _ := models.GetSysGeneralSetupByID("placement_setting")
	arrPlacementSetting := arrSysGeneralSetup.InputValue1

	arrPlacementSettingPointer := &PlacementSetting{}
	err := json.Unmarshal([]byte(arrPlacementSetting), arrPlacementSettingPointer)
	if err != nil {
		base.LogErrorLog("memberService:ValidatePlacementCode():Unmarshal():1", err.Error(), map[string]interface{}{"arrPlacementSetting": arrPlacementSetting}, true)
		return nil
	}

	var (
		placementID int
		arrData     = []map[string]interface{}{}
	)

	// get member id of placement_code
	arrEntMemberFn := make([]models.WhereCondFn, 0)
	arrEntMemberFn = append(arrEntMemberFn,
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: placementCode},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrEntMember, _ := models.GetEntMemberFn(arrEntMemberFn, "", false)
	if arrEntMember == nil {
		return arrData
	}

	placementID = arrEntMember.ID

	for i := 1; i <= arrPlacementSettingPointer.MaxLeg; i++ {
		var curLeg = i

		// find if this placement code leg no is occupied.
		arrEntMemberTreeSponsorFn := make([]models.WhereCondFn, 0)
		arrEntMemberTreeSponsorFn = append(arrEntMemberTreeSponsorFn,
			models.WhereCondFn{Condition: "ent_member_tree_sponsor.upline_id = ?", CondValue: placementID},
			models.WhereCondFn{Condition: "ent_member_tree_sponsor.leg_no = ?", CondValue: curLeg},
		)

		arrEntMemberTreeSponsor, _ := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberTreeSponsorFn, false)

		if arrEntMemberTreeSponsor != nil {
			continue
		}

		arrData = append(arrData, map[string]interface{}{
			"value": curLeg,
			"name":  helpers.TranslateV2("leg_:0", langCode, map[string]string{"0": strconv.Itoa(curLeg)}),
		})
	}

	return arrData
}
