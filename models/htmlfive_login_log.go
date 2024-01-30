package models

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// HtmlfiveLoginLog struct
type HtmlfiveLoginLog struct {
	ID             int       `gorm:"primary_key" json:"id"`
	TUserID        int       `json:"t_user_id"`
	TNickName      string    `json:"t_nick_name"`
	TType          string    `json:"t_type"`
	Source         uint8     `json:"source"`
	LanguageID     string    `json:"language_id"`
	TToken         string    `json:"t_token"`
	BLogin         int       `json:"b_login"`
	BLogout        int       `json:"b_logout"`
	DtLogin        time.Time `json:"dt_login"`
	DtExpiry       time.Time `json:"dt_expiry"`
	TOs            string    `json:"t_os"`
	TModel         string    `json:"t_model"`
	TManufacturer  string    `json:"t_manufacturer"`
	TAppVersion    string    `json:"t_app_version"`
	TOsVersion     string    `json:"t_os_version"`
	TPushNotiToken string    `json:"t_push_noti_token"`
	DtTimestamp    time.Time `json:"dt_timestamp"`
}

// AddHtmlfiveLoginLog add api log
func AddHtmlfiveLoginLog(tx *gorm.DB, saveData HtmlfiveLoginLog) error {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddHtmlfiveLoginLog-AddHtmlfiveLoginLog", err.Error(), saveData)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GetHtmlfiveLoginLogFn get htmlfive_login_log data with dynamic condition
func GetHtmlfiveLoginLogFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*HtmlfiveLoginLog, error) {
	var htmlfiveLoginLog HtmlfiveLoginLog
	tx := db.Table("htmlfive_login_log")
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
	err := tx.Find(&htmlfiveLoginLog).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if htmlfiveLoginLog.ID <= 0 {
		return nil, nil
	}

	return &htmlfiveLoginLog, nil
}

// GetExistingActiveHtmlfiveLoginLog get htmlfive_login_log data  with dynamic condition
func GetExistingActiveHtmlfiveLoginLog(entMemberID int, source uint8, debug bool) ([]*HtmlfiveLoginLog, error) {
	var htmlfiveLoginLog []*HtmlfiveLoginLog
	tx := db.Raw("SELECT htmlfive_login_log.* " +
		"FROM htmlfive_login_log " +
		"LEFT JOIN (" +
		"SELECT * " +
		"FROM htmlfive_login_log " +
		"WHERE t_user_id = " + strconv.Itoa(entMemberID) + " " +
		"AND b_login = 0 " +
		"AND b_logout = 1 " +
		"AND source = " + strconv.Itoa(int(source)) + " " +
		") inactive_log ON htmlfive_login_log.t_token = inactive_log.t_token " +
		"WHERE htmlfive_login_log.t_user_id = " + strconv.Itoa(entMemberID) + " " +
		"AND htmlfive_login_log.b_login = 1 " +
		"AND htmlfive_login_log.b_logout = 0 " +
		"AND htmlfive_login_log.source = " + strconv.Itoa(int(source)) + " " +
		// "AND htmlfive_login_log.dt_expiry >= NOW() "+
		"AND inactive_log.id IS NULL ") // this one comment first. so far would happen yet also
	// "AND inactive_log.id IS NULL "+
	// "GROUP BY htmlfive_login_log.t_token", entMemberID, entMemberID) // this one comment first. so far would happen yet also
	if debug {
		tx = tx.Debug()
	}
	err := tx.Scan(&htmlfiveLoginLog).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return htmlfiveLoginLog, nil
}

// Update add api log
// func (a *HtmlfiveLoginLog) Update(output string, runtime int) error {
// 	a.Output = output
// 	a.RunningTime = runtime
// 	err := save(a)
// 	if err != nil {
// 		ErrorLog("HtmlfiveLoginLog-Update", err.Error(), map[string]interface{}{"output": output, "runtime": runtime})
// 		return err
// 	}
// 	return nil
// }

// UpdateUser update user data
// func (a *HtmlfiveLoginLog) UpdateUser(userid int, usertype, tokenid string) error {
// 	a.UserID = userid
// 	a.UserType = usertype
// 	a.TokenID = tokenid
// 	err := save(a)
// 	if err != nil {
// 		ErrorLog("HtmlfiveLoginLog-UpdateUser", err.Error(), map[string]interface{}{"userid": userid, "usertype": usertype, "tokenid": tokenid})
// 		return err
// 	}
// 	return nil
// }

// UpdateOutput update user data
// func (a *HtmlfiveLoginLog) UpdateOutput(output string) error {
// 	a.Output = output
// 	err := save(a)
// 	if err != nil {
// 		ErrorLog("HtmlfiveLoginLog-UpdateOutput", err.Error(), map[string]interface{}{"output": output})
// 		return err
// 	}
// 	return nil
// }
