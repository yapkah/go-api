package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// NftImg struct
type NftImg struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Title     string    `json:"title"`
	ImgLink   string    `json:"img_link"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// GetNftImgFn
func GetNftImgFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*NftImg, error) {
	var result []*NftImg
	tx := db.Table("nft_img")

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
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetRandomNftImgFn
func GetRandomNftImgFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*NftImg, error) {
	var result []*NftImg
	tx := db.Table("nft_img")

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

	tx = tx.Order("RAND()")

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
