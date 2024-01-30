package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// AppMemberPnGroup struct
type AppMemberPnGroup struct {
	ID            int       `gorm:"primary_key" json:"id"`
	PrjID         int       `json:"prj_id" gorm:"column:prj_id"`
	GroupName     string    `json:"group_name" gorm:"column:group_name"`
	MemberID      int       `json:"member_id" gorm:"column:member_id"`
	PushNotiToken string    `json:"push_noti_token" gorm:"column:push_noti_token"`
	OS            string    `json:"os" gorm:"column:os"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
}

// GetAppMemberPnGroupFn get app_member_pn_group with dynamic condition
func GetAppMemberPnGroupFn(arrCond []WhereCondFn, debug bool) ([]*AppMemberPnGroup, error) {
	var result []*AppMemberPnGroup
	tx := db.Table("app_member_pn_group").
		Select("app_member_pn_group.*")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// AddAppMemberPnGroup func
func AddAppMemberPnGroup(arrData AppMemberPnGroup) (*AppMemberPnGroup, error) {
	if err := db.Table("app_member_pn_group").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}
