package models

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// OTP struct
type OTP struct {
	ID         int       `gorm:"primary_key" json:"id"`
	SendType   string    `json:"send_type`
	ReceiverID string    `json:"receiver_id"`
	OtpType    string    `json:"otp_type"` // REG: register | RP: Reset Password
	Otp        string    `json:"otp"`
	Attempts   int       `json:"attempts"`
	BValid     int       `json:"b_valid"`
	ExpiredAt  time.Time `json:"expired_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetOtpByTimeCount get otp by time
func GetOtpByTimeCount(ReceiverID string, otpType string, date time.Time) (int, error) {
	type Count struct {
		Count int `json:"count"`
	}
	var count Count
	err := db.Table("otp").Select("COUNT(id) as count").Where("receiver_id = ? AND otp_type = ? AND created_at > ? ", ReceiverID, otpType, date).Group("receiver_id").First(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return count.Count, nil
}

// AddOTP add otp
func AddOTP(tx *gorm.DB, otp OTP) (*OTP, error) {
	if err := tx.Create(&otp).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &otp, nil
}

// GetOtpFn get sms otp data with dynamic condition
func GetOtpFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*OTP, error) {
	var otp OTP
	tx := db.Table("otp")
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
	err := tx.Find(&otp).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if otp.ID <= 0 {
		return nil, nil
	}

	return &otp, nil
}

// Validate otp
func (o *OTP) Validate(tx *gorm.DB, otp string) (bool, string) {
	var err error
	// increase attempt
	err = o.IncreaseAttempt(db) // use db because if failed also +1 to attempt after rollback
	if err != nil {
		ErrorLog("smsOtpModel:Validate()", "IncreaseAttempt():1", err.Error())
		return false, "something_went_wrong"
	}

	// if expired
	if !o.ValidateExpire() {
		return false, e.GetMsg(e.OTP_EXPIRED)
	}

	// if exceed attempt
	if !o.ValidateAttempts() {
		return false, e.GetMsg(e.OTP_EXCEED_MAX_ATTEMPTS)
	}

	// invalid otp code
	if !o.ValidateOTP(otp) {
		return false, e.GetMsg(e.INVALID_OTP)
	}

	// update otp status
	err = o.Use(tx)
	if err != nil {
		ErrorLog("smsOtpModel:Validate()", "Use():1", err.Error())
		return false, "something_went_wrong"
	}

	return true, ""
}

// IncreaseAttempt otp attempt +1
func (o *OTP) IncreaseAttempt(tx *gorm.DB) error {
	o.Attempts++
	return SaveTx(tx, o)
}

// Use otp
func (o *OTP) Use(tx *gorm.DB) error {
	o.BValid = 0
	return SaveTx(tx, o)
}

// ValidateExpire validate opt expire time
func (o *OTP) ValidateExpire() bool {
	nowTime := time.Now()
	// if expired
	if nowTime.Equal(o.ExpiredAt) || nowTime.After(o.ExpiredAt) {
		return false
	}
	return true
}

// OtpSetting struct
type OtpSetting struct {
	MaxAttempts string `json:"max_attempts"`
}

// ValidateAttempts validate opt attempts
func (o *OTP) ValidateAttempts() bool {
	arrGeneralSetup, err := GetSysGeneralSetupByID("otp_setting")
	if err != nil {
		return false
	}
	if arrGeneralSetup == nil {
		return false
	}

	otpSetting := &OtpSetting{}
	err = json.Unmarshal([]byte(arrGeneralSetup.SettingValue1), otpSetting)
	if err != nil {
		return false
	}

	maxReq, err := strconv.Atoi(otpSetting.MaxAttempts)
	if err != nil {
		return false
	}

	// if exceed attempt
	if o.Attempts > maxReq {
		return false
	}
	return true
}

// ValidateOTP validate opt
func (o *OTP) ValidateOTP(otp string) bool {
	if otp != o.Otp {
		return false
	}
	return true
}
