package models

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AppLoginLog struct
type AppLoginLog struct {
	ID             int       `gorm:"primary_key" gorm:"column:id" json:"id"`
	TUserID        int       `gorm:"column:t_user_id" json:"t_user_id"`
	TNickName      string    `gorm:"column:t_nick_name" json:"t_nick_name"`
	TType          string    `gorm:"column:t_type" json:"t_type"`
	Source         uint8     `gorm:"column:source" json:"source"`
	LanguageID     string    `gorm:"column:language_id" json:"language_id"`
	TToken         string    `gorm:"column:t_token" json:"t_token"`
	BLogin         int       `gorm:"column:b_login" json:"b_login"`
	BLogout        int       `gorm:"column:b_logout" json:"b_logout"`
	DtLogin        time.Time `gorm:"column:dt_login" json:"dt_login"`
	DtExpiry       time.Time `gorm:"column:dt_expiry" json:"dt_expiry"`
	TOs            string    `gorm:"column:t_os" json:"t_os"`
	TModel         string    `gorm:"column:t_model" json:"t_model"`
	TManufacturer  string    `gorm:"column:t_manufacturer" json:"t_manufacturer"`
	TAppVersion    string    `gorm:"column:t_app_version" json:"t_app_version"`
	TOsVersion     string    `gorm:"column:t_os_version" json:"t_os_version"`
	TPushNotiToken string    `gorm:"column:t_push_noti_token" json:"t_push_noti_token"`
	DtTimestamp    time.Time `gorm:"column:dt_timestamp" json:"dt_timestamp"`
}

// AddAppLoginLog add api log
func AddAppLoginLog(tx *gorm.DB, saveData AppLoginLog) error {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddAppLoginLog-AddAppLoginLog", err.Error(), saveData)
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GetAppLoginLogFn get app_login_log data with dynamic condition
func GetAppLoginLogFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*AppLoginLog, error) {
	var appLoginLog AppLoginLog
	tx := db.Table("app_login_log").
		Order("id DESC")
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
		ErrorLog("GetAppLoginLogFn-debug_app_login_log_sql", arrCond, tx.Debug())
	}
	err := tx.Find(&appLoginLog).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if appLoginLog.ID <= 0 {
		return nil, nil
	}

	return &appLoginLog, nil
}

// GetExistingActiveAppLoginLog get app_login_log data  with dynamic condition
func GetExistingActiveAppLoginLog(entMemberID int, source uint8, debug bool) ([]*AppLoginLog, error) {
	var appLoginLog []*AppLoginLog
	tx := db.Raw("SELECT app_login_log.* " +
		"FROM app_login_log " +
		"LEFT JOIN (" +
		"SELECT * " +
		"FROM app_login_log " +
		"WHERE t_user_id = " + strconv.Itoa(entMemberID) + " " +
		"AND b_login = 0 " +
		"AND b_logout = 1 " +
		"AND source = " + strconv.Itoa(int(source)) + " " +
		") inactive_log ON app_login_log.t_token = inactive_log.t_token " +
		"WHERE app_login_log.t_user_id = " + strconv.Itoa(entMemberID) + " " +
		"AND app_login_log.b_login = 1 " +
		"AND app_login_log.b_logout = 0 " +
		"AND app_login_log.source = " + strconv.Itoa(int(source)) + " " +
		// "AND app_login_log.dt_expiry >= NOW() "+
		"AND inactive_log.id IS NULL " +
		" ORDER BY app_login_log.dt_expiry DESC ") // this one comment first. so far would happen yet also
	// "AND inactive_log.id IS NULL "+
	// "GROUP BY app_login_log.t_token", entMemberID, entMemberID) // this one comment first. so far would happen yet also
	if debug {
		tx = tx.Debug()
	}
	err := tx.Scan(&appLoginLog).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return appLoginLog, nil
}

// Update add api log
// func (a *AppLoginLog) Update(output string, runtime int) error {
// 	a.Output = output
// 	a.RunningTime = runtime
// 	err := save(a)
// 	if err != nil {
// 		ErrorLog("AppLoginLog-Update", err.Error(), map[string]interface{}{"output": output, "runtime": runtime})
// 		return err
// 	}
// 	return nil
// }

// UpdateUser update user data
// func (a *AppLoginLog) UpdateUser(userid int, usertype, tokenid string) error {
// 	a.UserID = userid
// 	a.UserType = usertype
// 	a.TokenID = tokenid
// 	err := save(a)
// 	if err != nil {
// 		ErrorLog("AppLoginLog-UpdateUser", err.Error(), map[string]interface{}{"userid": userid, "usertype": usertype, "tokenid": tokenid})
// 		return err
// 	}
// 	return nil
// }

// UpdateOutput update user data
// func (a *AppLoginLog) UpdateOutput(output string) error {
// 	a.Output = output
// 	err := save(a)
// 	if err != nil {
// 		ErrorLog("AppLoginLog-UpdateOutput", err.Error(), map[string]interface{}{"output": output})
// 		return err
// 	}
// 	return nil
// }

// GetDistinctAppLoginLogFn get app_login_log data with dynamic condition
func GetDistinctAppLoginLogFn(arrCond []WhereCondFn, debug bool) ([]*AppLoginLog, error) {
	var result []*AppLoginLog
	tx := db.Table("app_login_log").
		Group("push_noti_token").
		Order("id DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
		ErrorLog("GetDistinctAppLoginLogFn-debug_app_login_log_sql", arrCond, tx.Debug())
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetAppLoginLogListFn get app_login_log data with dynamic condition
func GetAppLoginLogListFn(arrCond []WhereCondFn, debug bool) ([]*AppLoginLog, error) {
	var appLoginLog []*AppLoginLog
	tx := db.Table("app_login_log").
		Order("id DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&appLoginLog).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return appLoginLog, nil
}
