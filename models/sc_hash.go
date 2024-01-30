package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SCHash struct
type SCHash struct {
	ID        int       `gorm:"primary_key" json:"id"`
	SCID      int       `json:"sc_id" gorm:"column:sc_id"`
	SCPart    string    `json:"sc_part" gorm:"column:sc_part"`
	SCEncrypt string    `json:"sc_encrypt" gorm:"column:sc_encrypt"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetSCHashFn get sc_hash data with dynamic condition
func GetSCHashFn(arrCond []WhereCondFn, debug bool) ([]*SCHash, error) {
	var result []*SCHash
	tx := db.Table("sc_hash")
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

// func AddSCHash add sc_hash records`
func AddSCHash(tx *gorm.DB, saveData SCHash) (*SCHash, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddSCHash-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}
