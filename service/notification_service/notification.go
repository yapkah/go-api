package notification_service

import (
	"strings"

	"github.com/smartblock/gta-api/helpers"
)

type MemberNotificationListStruct struct {
	ID        int             `json:"id"`
	Message   string          `json:"message"`
	ExtraInfo ExtraInfoStruct `json:"extra_info"`
}

type ExtraInfoStruct struct {
	GameSession string `json:"game_session"`
}

type MemberNotificationStruct struct {
	MemberID int
	LangCode string
	PopUp    int
	Scenario string
}

// func GetMemberNotificationListv1
func GetMemberNotificationListv1(arrData MemberNotificationStruct) []MemberNotificationListStruct {

	arrDataReturn := make([]MemberNotificationListStruct, 0)
	scenario := strings.ToLower(arrData.Scenario)
	// fmt.Println("scenario", scenario)
	if scenario == "dashboard" {
		// hour := "24"
		// // filterSql := "AND IF(downline.d_last_game IS NOT NULL AND downline.d_last_game != '', downline.d_last_game <= DATE_SUB(NOW(), INTERVAL " + hour + " HOUR), downline.created_at <= DATE_SUB(NOW(), INTERVAL " + hour + " HOUR)) "
		// filterSql := "AND  downline.d_last_game <= DATE_SUB(NOW(), INTERVAL " + hour + " HOUR) "
		// arrCond := make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: " sponsor.id = ? " + filterSql, CondValue: arrData.MemberID},
		// 	models.WhereCondFn{Condition: " wod_member_diamond.type = ? ", CondValue: 2},
		// )
		// arrDownlineLastPlayGame, _ := models.GetMemberLastPlayGame(arrCond, true)
		// if len(arrDownlineLastPlayGame) > 0 {
		// 	if arrDownlineLastPlayGame[0].ProfileName != "" {
		// 		params := make(map[string]string)
		// 		params["profilename"] = helpers.TransDiamondName(arrDownlineLastPlayGame[0].ProfileName, arrData.LangCode)
		// 		translatedWord := helpers.TranslateV2("this_is_a_reminder_that_your_downline_accounts_has_not_been_active_in_the_game._your_:profilename_in_the_cullian_pool_will_be_moved_to_eternity_if_the_min_requirements_are_not_met.", arrData.LangCode, params)
		// 		arrDataReturn = append(arrDataReturn,
		// 			MemberNotificationListStruct{ID: len(arrDataReturn) + 1, Message: translatedWord},
		// 		)
		// 	}
		// }
	}
	params := make(map[string]string)
	translatedWord := helpers.TranslateV2("this_is_a_reminder_that_your_downline_accounts_has_not_been_active_in_the_game._your_:profilename_in_the_cullian_pool_will_be_moved_to_eternity_if_the_min_requirements_are_not_met.", arrData.LangCode, params)
	arrDataReturn = append(arrDataReturn,
		MemberNotificationListStruct{ID: len(arrDataReturn) + 1, Message: translatedWord},
	)

	return arrDataReturn
}
