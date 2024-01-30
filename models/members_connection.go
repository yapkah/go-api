package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// MembersConnection struct
type MembersConnection struct {
	ID         int    `gorm:"primary_key" json:"id"`
	Connection string `json:"connection"`
	MemberID   int    `json:"member_id"`
	TokenID    string `json:"token_id"`
	Device     string `json:"device"`
	OS         string `json:"os"`
	OsVersion  string `json:"os_version"`
	AppVersion string `json:"app_version"`
	Status     string `json:"status"`
}

// AddConnection add member
func AddConnection(tx *gorm.DB, connection string, memberid int, tokenid, device, os, osVersion, appVersion, status string) (*MembersConnection, error) {
	var err error
	err = InactiveConnections(tx, connection)
	if err != nil {
		return nil, err
	}

	conn := MembersConnection{
		Connection: connection,
		MemberID:   memberid,
		TokenID:    tokenid,
		Device:     device,
		OS:         os,
		OsVersion:  osVersion,
		AppVersion: appVersion,
		Status:     status,
	}

	if err = tx.Create(&conn).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &conn, nil
}

// UpdateConnection update connection
func (c *MembersConnection) UpdateConnection(tx *gorm.DB, memberid int, tokenid, device, os, osVersion, appVersion, status string) error {
	if status == "A" {
		err := InactiveConnections(tx, c.Connection)
		if err != nil {
			return err
		}
	}

	c.TokenID = tokenid
	c.Device = device
	c.OS = os
	c.OsVersion = osVersion
	c.AppVersion = appVersion
	c.Status = status

	return SaveTx(tx, &c)
}

// InactiveConnections inactive connection
func InactiveConnections(tx *gorm.DB, connection string) error {
	err := tx.Table("members_connection").Where("connection = ?", connection).Where("status = ?", "A").Update("status", "I").Error
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// InactiveConnectionsByMember inactive connection
func InactiveConnectionsByMember(tx *gorm.DB, memberid int) error {
	err := tx.Table("members_connection").
		Where("member_id = ?", memberid).
		Where("status = ?", "A").
		Update("status", "I").Error

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GetConnection func
func GetConnection(connection string, memberid int) (*MembersConnection, error) {
	var conn MembersConnection
	err := db.Where("connection = ?", connection).
		Where("member_id = ?", memberid).
		Where("status != ?", "T").
		First(&conn).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &conn, nil
}

// ExistConnection func
func ExistConnection(connection string, memberid int) bool {
	conn, err := GetConnection(connection, memberid)

	if err != nil || conn == nil {
		return false
	}
	return true
}

// GetActiveConnections func
func GetActiveConnections() ([]*MembersConnection, error) {
	var conn []*MembersConnection
	err := db.Table("members_connection mc").
		Joins("JOIN members m ON mc.member_id = m.id").
		Joins("JOIN access_token a ON m.login_token_id = a.id").
		Joins("JOIN refresh_token r ON a.id = r.access_token_id").
		Where("mc.status = ?", "A").
		Where("m.status = ?", "A").
		Where("a.status = ?", "A").
		Where("r.status = ?", "A").
		Where("r.expires_at > ?", time.Now()).
		Select("mc.*").
		Find(&conn).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return conn, nil
}

// UpdateMemberConnectionByTokenID func
func UpdateMemberConnectionByTokenID(tx *gorm.DB, tokenid string, newtokenid string) (err error) {
	err = tx.Table("members_connection").
		Where("token_id = ?", tokenid).
		Where("status = ?", "A").
		Update("token_id", newtokenid).Error

	if err != nil {
		err = &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return
}

// GetMemberActiveConnection func
func GetMemberActiveConnection(memberid int) (*MembersConnection, error) {
	var conn MembersConnection
	err := db.Where("member_id = ?", memberid).
		Where("status = ?", "A").
		First(&conn).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &conn, nil
}

// GetAllMemberConnection func
func GetAllMemberConnection(memberid int) ([]*MembersConnection, error) {
	var conn []*MembersConnection
	err := db.Where("member_id = ?", memberid).
		Where("status != ?", "T").
		Find(&conn).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return conn, nil
}

// GetSendConnection func
func GetSendConnection(memberid int) (*MembersConnection, error) {
	var conn MembersConnection
	err := db.Table("members_connection").
		Joins("JOIN noti_connection_setting ncs ON members_connection.connection = ncs.connection").
		Where("members_connection.member_id = ?", memberid).
		Where("members_connection.status = ?", "A").
		Where("ncs.receive_noti = ?", "1").
		Select("members_connection.*").
		First(&conn).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &conn, nil
}
