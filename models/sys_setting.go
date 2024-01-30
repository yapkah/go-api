package models

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SysSetting struct
type SysSetting struct {
	GroupID   string `gorm:"primary_key" json:"group_id"`
	SettingID string `gorm:"primary_key" json:"setting_id"`
	Title     string `json:"title"`
	Desc      string `json:"description" gorm:"column:description"`
	Value     string `json:"value"`
}

// GetSysSettingByID func
func GetSysSettingByID(groupID, settingID string) (*SysSetting, error) {
	var sys SysSetting
	err := db.Where("group_id = ? AND setting_id = ?", groupID, settingID).First(&sys).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: err.Error(), Data: map[string]string{"groupID": groupID, "settingID": settingID}}
		}
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return &sys, nil
}

// GetSysSettingByGroup func
func GetSysSettingByGroup(groupID string) (map[string]*SysSetting, error) {
	var sys []*SysSetting
	err := db.Where("group_id = ?", groupID).Find(&sys).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	data := make(map[string]*SysSetting)
	for _, v := range sys {
		data[v.SettingID] = v
	}
	return data, nil
}

// ValueToInt convert value to Int
func (s *SysSetting) ValueToInt() (int, error) {
	data, err := strconv.Atoi(s.Value)
	if err != nil {
		return 0, err
	}
	return data, nil
}

// ValueToDuration convert value to duration
func (s *SysSetting) ValueToDuration() (time.Duration, error) {
	data, err := s.ValueToInt()
	if err != nil {
		return 0, err
	}
	return time.Duration(data), nil
}
