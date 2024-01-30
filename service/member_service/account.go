package member_service

import (
	"encoding/json"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
)

type MemberAccountListStruct struct {
	MemberID    int
	EntMemberID int
	LangCode    string
}

// struct MemberAccountListRstStruct
type MemberAccountListRstStruct struct {
	Username             string `json:"username"`
	CurrentActiveAccount int    `json:"current_active_account"`
	TaggedUsername       string `json:"tagged_username"`
	// TransferSetup        []TransferSetupStruct `json:"tagged_crypto_addr"`
}

// func GetMemberAccountListv1
func GetMemberAccountListv1(arrData MemberAccountListStruct) []MemberAccountListRstStruct {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member.main_id = ? ", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
	)

	arrMemberAccountList, _ := models.GetEntMemberListFn(arrCond, false)

	arrDataReturn := make([]MemberAccountListRstStruct, 0)
	if len(arrMemberAccountList) > 0 {
		for _, arrMemberAccountListV := range arrMemberAccountList {
			var currentActiveMember int

			if arrMemberAccountListV.ID == arrData.EntMemberID {
				currentActiveMember = 1
			}

			var taggedUsername string
			// var translatedWalletTypeName string

			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ent_member.id = ? ", CondValue: arrMemberAccountListV.TaggedMemberID},
				models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
			)
			tagMemberRst, _ := models.GetEntMemberFn(arrCond, "", false)

			if tagMemberRst != nil {
				taggedUsername = tagMemberRst.NickName
			}

			// arrTransferSetup := make([]TransferSetupStruct, 0)
			// if taggedUsername != "" {
			// cryptoAddr, _ := models.GetCustomMemberCryptoAddr(tagMemberRst.ID, "SEC", true, false)

			// arrCond = make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: " ewt_setup.status = ? ", CondValue: "A"},
			// 	models.WhereCondFn{Condition: " ewt_setup.withdrawal_with_crypto = ? ", CondValue: 1},
			// )
			// arrEwtSetup, _ := models.GetEwtSetupListFn(arrCond, false)

			// if arrEwtSetup != nil {
			// 	for _, arrEwtSetupV := range arrEwtSetup {
			// 		translatedWalletTypeName = helpers.TranslateV2(arrEwtSetupV.EwtTypeName, arrData.LangCode, nil)
			// 		arrTransferSetup = append(arrTransferSetup,
			// 			TransferSetupStruct{
			// 				CryptoAddr:     cryptoAddr,
			// 				WalletTypeCode: arrEwtSetupV.EwtTypeCode,
			// 				WalletTypeName: translatedWalletTypeName,
			// 			},
			// 		)
			// 	}
			// }
			// }

			arrDataReturn = append(arrDataReturn,
				MemberAccountListRstStruct{
					Username:             arrMemberAccountListV.NickName,
					CurrentActiveAccount: currentActiveMember,
					TaggedUsername:       taggedUsername,
					// TransferSetup:        arrTransferSetup,
				},
			)
		}
	}

	return arrDataReturn
}

// struct SwitchCurrentActiveMemberAccountv1Struct
type SwitchCurrentActiveMemberAccountv1Struct struct {
	EntMemberID int
	MemberID    int
	UsernameTo  string
}

// SwitchCurrentActiveMemberAccountv1 switch current active member account to another acccount
func SwitchCurrentActiveMemberAccountv1(tx *gorm.DB, arrData SwitchCurrentActiveMemberAccountv1Struct) (logData bool, err error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: arrData.UsernameTo},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrTargetEntMember, _ := models.GetEntMemberFn(arrCond, "", false)
	if arrTargetEntMember == nil || arrTargetEntMember.ID < 1 {
		// base.LogErrorLog("SwitchCurrentActiveMemberAccountv1-GetEntMemberFn_failed", "get_target_account_info", arrCond, true)
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_switch_current_profile"}
	}

	if arrTargetEntMember.CurrentProfile == 1 || arrTargetEntMember.ID == arrData.EntMemberID {
		return false, nil
	}

	// update current login member current profile to inactive
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)

	updateColumn := map[string]interface{}{"current_profile": 0, "updated_by": arrData.MemberID}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("SwitchCurrentActiveMemberAccountv1-update_ent_member", "update_current_member_current_profile_to_0", err.Error(), true)
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	// update target profile to active
	arrUpdCond = make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrTargetEntMember.ID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)

	updateColumn = map[string]interface{}{"current_profile": 1, "updated_by": arrData.MemberID}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("SwitchCurrentActiveMemberAccountv1-update_ent_member", "update_target_current_profile_to_1", err.Error(), true)
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return true, nil
}

// DeactivateMemberAccountv1Struct struct
type DeactivateMemberAccountv1Struct struct {
	EntMemberID int
	MemberID    int
	UsernameTo  string
}

// DeactivateMemberAccountv1 inactivate target member account
func DeactivateMemberAccountv1(tx *gorm.DB, arrData DeactivateMemberAccountv1Struct) (logData bool, err error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: arrData.MemberID},
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: arrData.UsernameTo},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrTargetEntMember, _ := models.GetEntMemberFn(arrCond, "", false)
	if arrTargetEntMember == nil || arrTargetEntMember.ID < 1 {
		// base.LogErrorLog("DeactivateMemberAccountv1-GetEntMemberFn_failed", "get_target_account_info", arrCond, true)
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	if arrTargetEntMember.CurrentProfile == 1 {
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "current_active_account_cannot_be_delete_please_choose_another_account_if_available"}
	}

	// update target profile to active
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrTargetEntMember.ID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)
	curDateTimeString := base.GetCurrentTime("2006-01-02 15:04:05")

	updateColumn := map[string]interface{}{"status": "I", "updated_by": arrData.MemberID, "cancelled_by": arrData.MemberID, "cancelled_at": curDateTimeString}
	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("DeactivateMemberAccountv1-update_ent_member", "update_target_status_to_I", err.Error(), true)
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}

	return true, nil
}

// struct TagMemberAccountv1Struct
type TagMemberAccountv1Struct struct {
	CurrentLoginEntMemberID int
	EntMemberID             int
	TagEntMemberID          int
}

// TagMemberAccountv1 tag member account to another acccount
func TagMemberAccountv1(tx *gorm.DB, arrData TagMemberAccountv1Struct) (logData bool, err error) {

	// update member with tag member account
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: arrData.EntMemberID},
		models.WhereCondFn{Condition: "status = ?", CondValue: "A"},
	)

	updateColumn := map[string]interface{}{"tagged_member_id": arrData.TagEntMemberID, "updated_by": arrData.CurrentLoginEntMemberID}

	err = models.UpdatesFnTx(tx, "ent_member", arrUpdCond, updateColumn, false)
	if err != nil {
		base.LogErrorLog("TagMemberAccountv1-update_ent_member", "update_current_member_current_tag_account", err.Error(), true)
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
	}
	return true, nil
}

type MemberAccountTransferExchangeBatchAssetsStruct struct {
	EntMemberID int
	LangCode    string
}

type TransferSetupStruct struct {
	WalletTypeCode   string `json:"ewallet_type_code"`
	WalletTypeName   string `json:"ewallet_type_name"`
	WalletTypeImgURL string `json:"wallet_type_image_url"`
}

func GetMemberAccountTransferExchangeBatchAssetsv1(arrData MemberAccountTransferExchangeBatchAssetsStruct) []TransferSetupStruct {

	arrDataReturn := make([]TransferSetupStruct, 0)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ewt_from.member_show = ? ", CondValue: 1},
		models.WhereCondFn{Condition: " ewt_from.ewallet_type_code != ? ", CondValue: "PSSEC"},
	)

	arrTransferSetupRst, _ := models.GetDistinctEwtTransferFromFn(arrCond, false)

	if len(arrTransferSetupRst) > 0 {
		for _, arrTransferSetupRstV := range arrTransferSetupRst {

			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewt_setup.id = ? ", CondValue: arrTransferSetupRstV.EwalletTypeIdFrom},
			)
			arrEwtSetup, _ := models.GetEwtSetupListFn(arrCond, false)

			if arrEwtSetup != nil {
				for _, arrEwtSetupV := range arrEwtSetup {
					translatedWalletTypeName := helpers.TranslateV2(arrEwtSetupV.EwtTypeName, arrData.LangCode, nil)

					var appSettingList struct {
						BGGradBegin        string `json:"bg_grad_begin"`
						BGGradEnd          string `json:"bg_grad_end"`
						WalletTypeImageURL string `json:"wallet_type_image_url"`
					}

					var walletTypeImgURL string
					if arrEwtSetupV.AppSettingList != "" {
						json.Unmarshal([]byte(arrEwtSetupV.AppSettingList), &appSettingList)
						if appSettingList.WalletTypeImageURL != "" {
							walletTypeImgURL = appSettingList.WalletTypeImageURL
						}
					}

					arrDataReturn = append(arrDataReturn,
						TransferSetupStruct{
							WalletTypeCode:   arrEwtSetupV.EwtTypeCode,
							WalletTypeName:   translatedWalletTypeName,
							WalletTypeImgURL: walletTypeImgURL,
						},
					)
				}
			}
		}
	}

	return arrDataReturn
}

// struct SwitchCurrentActiveMemberAccountv2Struct
type SwitchCurrentActiveMemberAccountv2Struct struct {
	EntMemberID  int
	MemberMainID int
	SourceID     int
	UsernameTo   string
}

// SwitchCurrentActiveMemberAccountv2 switch current active member account to another acccount
func SwitchCurrentActiveMemberAccountv2(tx *gorm.DB, arrData SwitchCurrentActiveMemberAccountv2Struct) (logData bool, err error) {

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.main_id = ?", CondValue: arrData.MemberMainID},
		models.WhereCondFn{Condition: "ent_member.nick_name = ?", CondValue: arrData.UsernameTo},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrTargetEntMember, _ := models.GetEntMemberFn(arrCond, "", false)
	if arrTargetEntMember == nil || arrTargetEntMember.ID < 1 {
		return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "failed_switch_current_profile"}
	}

	// start get current active profile with main id
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "source_id = ?", CondValue: arrData.SourceID},
		models.WhereCondFn{Condition: "main_id = ?", CondValue: arrData.MemberMainID},
	)
	arrEntCurrentProfile, _ := models.GetEntCurrentProfileFn(arrCond, false)
	// end get current active profile with main id

	curDateTimeString := base.GetCurrentTime("2006-01-02 15:04:05")

	if len(arrEntCurrentProfile) < 1 {
		// no data. need to create.
		arrCrtData := models.AddEntCurrentProfileStruct{
			SourceID: arrData.SourceID,
			MainID:   arrData.MemberMainID,
			MemberID: arrTargetEntMember.ID,
		}

		_, err := models.AddEntCurrentProfile(tx, arrCrtData)

		if err != nil {
			base.LogErrorLog("SwitchCurrentActiveMemberAccountv2-create_ent_current_profile", err.Error(), arrCrtData, true)
			return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
		}

	} else { // current profile is exitsted.
		if arrEntCurrentProfile[0].MemberID == arrTargetEntMember.ID { // plan to switch same account. return success directly
			return false, nil
		} else {
			// start update ent_current_profile to target member id
			arrUpdCond := make([]models.WhereCondFn, 0)
			arrUpdCond = append(arrUpdCond,
				models.WhereCondFn{Condition: " main_id = ?", CondValue: arrData.MemberMainID},
				models.WhereCondFn{Condition: " source_id = ?", CondValue: arrData.SourceID},
			)

			updateColumn := map[string]interface{}{"member_id": arrTargetEntMember.ID, "updated_at": curDateTimeString}
			err = models.UpdatesFnTx(tx, "ent_current_profile", arrUpdCond, updateColumn, false)
			if err != nil {
				arrErr := map[string]interface{}{
					"arrUpdCond":   arrUpdCond,
					"updateColumn": updateColumn,
				}
				base.LogErrorLog("SwitchCurrentActiveMemberAccountv2-update_ent_current_profile", err.Error(), arrErr, true)
				return false, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "something_went_wrong"}
			}
			// end update ent_current_profile to target member id
		}
	}

	return true, nil
}
