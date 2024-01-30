package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SysImg struct
type SysImg struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Module    string    `json:"module"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	ImgLink   string    `json:"img_link"`
	PopupImg  string    `json:"popup_img"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// GetSysImgFn
func GetSysImgFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SysImg, error) {
	var result []*SysImg
	tx := db.Table("sys_img")

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
