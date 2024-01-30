package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// UserGroup struct
type UserGroupStruc struct {
	ID        int    `gorm:"primary_key" json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ParentId  int    `json:"parent_id"`
	CreatedBy int    `json:"created_by"`
	UpdatedBy int    `json:"created_at"`
}

// AddUserGroup create new user group
func AddUserGroup(AdminId int, UserGroupCode string, UserGroupName string, ParentCode int) error {

	userGroup := UserGroupStruc{
		Code:      UserGroupCode,
		Name:      UserGroupName,
		ParentId:  ParentCode,
		CreatedBy: AdminId,
		UpdatedBy: AdminId,
	}

	if err := db.Table("sys_user_group").Create(&userGroup).Error; err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GetUserGroupId get parent Id
func GetUserGroupId(UserGroupCode string) (int, error) {

	var userGroup UserGroupStruc

	err := db.Table("sys_user_group").
		Where("code = ?", UserGroupCode).First(&userGroup).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return userGroup.ID, nil
}

// GetUserGroup get parent data
func GetUserGroupData(UserGroupCode string) (*UserGroupStruc, error) {

	var userGroup UserGroupStruc

	err := db.Table("sys_user_group").
		Where("id = ?", UserGroupCode).First(&userGroup).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &userGroup, nil
}

// GetUserGroupTree
func GetUserGroupTree() ([]*UserGroupStruc, error) {

	var userGroup []*UserGroupStruc

	err := db.Table("sys_user_group").Where("id >= 10").Order("parent_id").Scan(&userGroup).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return userGroup, nil
}
