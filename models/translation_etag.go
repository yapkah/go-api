package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TranslationEtag struct
type TranslationEtag struct {
	ID            int       `gorm:"primary_key" json:"id"`
	Locale        string    `json:"locale" json:"locale"`
	EtagID        string    `json:"etag_id" json:"etag_id"`
	TotalRecord   int       `json:"total_record"`
	LastUpdatedAt time.Time `json:"last_updated_at" gorm:"column:last_updated_at"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
}

// GetTranslationEtagFn func
func GetTranslationEtagFn(arrWhereFn []WhereCondFn, debug bool) ([]*TranslationEtag, error) {
	var result []*TranslationEtag
	tx := db.Table("translation_etag")

	if len(arrWhereFn) > 0 {
		for _, v := range arrWhereFn {
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

// func AddTranslationEtag add translation_etag records`
func AddTranslationEtag(saveData TranslationEtag) (*TranslationEtag, error) {
	if err := db.Table("translation_etag").Create(&saveData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}
