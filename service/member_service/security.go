package member_service

import (
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/e"
	"github.com/yapkah/go-api/pkg/util"

	"github.com/jinzhu/gorm"
)

// MemberPassword struct
type MemberPassword struct {
	MemberID            int
	Password            string
	CurrentPassword     string
	CurrentPasswordHash string
}

// UpdateMemberPassword function
func (p *MemberPassword) UpdateMemberPassword(tx *gorm.DB, byPassVerification bool) string {
	var (
		err error
	)

	// validate current password
	if byPassVerification == false {
		err = base.CheckBcryptPassword(p.CurrentPasswordHash, p.CurrentPassword)
		if err != nil {
			return "invalid_current_password"
		}
	}

	// new passsword format checking
	ok := base.PasswordChecking(p.Password)
	if !ok {
		return "password_must_contain_one_letter_one_number_and_one_special_character"
	}

	// encrypt password
	password, err := base.Bcrypt(p.Password)
	if err != nil {
		models.ErrorLog("memberService:UpdateMemberPassword()", "Bcrypt():1", err.Error())
		return "something_went_wrong"
	}

	// update password
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: p.MemberID},
	)
	updateColumn := map[string]interface{}{"password": password, "updated_by": p.MemberID}
	err = models.UpdatesFnTx(tx, "members", arrUpdCond, updateColumn, false)
	if err != nil {
		models.ErrorLog("memberservice:UpdateMemberPassword()", "UpdatesFnTx()", err.Error())
		return "something_went_wrong"
	}

	return ""
}

// MemberSecondaryPin struct
type MemberSecondaryPin struct {
	MemberID               int
	SecondaryPin           string
	CurrentSecondaryPin    string
	CurrentSecondaryPinMd5 string
	CurrentOldSecondaryPin string
}

// UpdateMemberSecondaryPin function
func (p *MemberSecondaryPin) UpdateMemberSecondaryPin(tx *gorm.DB, byPassVerification bool) string {
	var (
		err error
		ok  bool
	)

	// validate current secondary pin
	// if byPassVerification == false {
	// 	err = base.CheckMd5SecondaryPin(p.CurrentSecondaryPinMd5, p.CurrentSecondaryPin)
	// 	if err != nil {
	// 		return "invalid_current_secondary_pin"
	// 	}
	// }

	// validate current secondary pin
	if byPassVerification == false {
		pinValidation := base.SecondaryPin{
			MemId:              p.MemberID,
			SecondaryPin:       p.CurrentSecondaryPin,
			MemberSecondaryPin: p.CurrentSecondaryPinMd5,
			LangCode:           "en",
		}

		err = pinValidation.CheckSecondaryPin()

		if err != nil {
			return err.Error()
		}
	}

	// seconday pin checking
	ok = base.SecondaryPinChecking(p.SecondaryPin)
	if !ok {
		return e.GetMsg(e.SECONDARY_PIN_VALIDATION_ERROR)
	}

	// encrypt secondary pin
	secondaryPin := util.EncodeMD5(p.SecondaryPin)

	// update secondary pin
	arrUpdCond := make([]models.WhereCondFn, 0)
	arrUpdCond = append(arrUpdCond,
		models.WhereCondFn{Condition: "id = ?", CondValue: p.MemberID},
	)
	updateColumn := map[string]interface{}{"secondary_pin": secondaryPin, "updated_by": p.MemberID}
	err = models.UpdatesFnTx(tx, "members", arrUpdCond, updateColumn, false)
	if err != nil {
		models.ErrorLog("memberservice:UpdateMemberSecondaryPin()", "UpdatesFnTx()", err.Error())
		return "something_went_wrong"
	}

	return ""
}

func GenerateRandomLoginPassword() string {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+=!?_:.,"
	numChar := 8
	var validPassword bool
	for {
		if !validPassword {
			generatedPassword := base.GenerateRandomString(numChar, charSet)

			// new passsword format checking
			ok := base.PasswordChecking(generatedPassword)
			if ok {
				return generatedPassword
			}
		}
	}

}
