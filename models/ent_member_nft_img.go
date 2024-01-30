package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntMemberNftImg struct
type EntMemberNftImg struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Type      string    `json:"type"`
	ImgID     int       `json:"img_id"`
	MemberID  int       `json:"member_id"`
	CreatedAt time.Time `json:"created_at"`
}

// EntMemberNftImgDetail struct
type EntMemberNftImgDetail struct {
	ID        int       `gorm:"primary_key" json:"id"`
	ImgID     int       `json:"img_id"`
	Title     string    `json:"title"`
	ImgLink   string    `json:"img_link"`
	MemberID  int       `json:"member_id"`
	CreatedAt time.Time `json:"created_at"`
}

// GetEwtMemberNftImgFn get ent_member_nft_img data with dynamic condition
func GetEwtMemberNftImgFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberNftImgDetail, error) {
	var result []*EntMemberNftImgDetail
	tx := db.Table("ent_member_nft_img").
		Select("ent_member_nft_img.id, ent_member_nft_img.member_id, ent_member_nft_img.img_id, nft_img.title, nft_img.img_link, ent_member_nft_img.created_at").
		Joins("inner join nft_img ON ent_member_nft_img.img_id = nft_img.id")

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

// AddEntMemberNftImg
func AddEntMemberNftImg(saveData EntMemberNftImg) (*EntMemberNftImg, error) {
	if err := db.Create(&saveData).Error; err != nil {
		ErrorLog("AddEntNftImg-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

type LatestMemberImgIdStruct struct {
	ImgNum int `json:"img_num"`
}

func GetLatestMemberNftImgID(arrCond []WhereCondFn, debug bool) (*LatestMemberImgIdStruct, error) {
	var result LatestMemberImgIdStruct
	tx := db.Table("ent_member_nft_img")
	tx = tx.Select("MAX(ent_member_nft_img.img_id) AS 'img_num'")

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

	return &result, nil
}
