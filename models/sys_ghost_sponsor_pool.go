package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// SysGhostSponsorPool struct
type SysGhostSponsorPool struct {
	ID              int     `json:"id" gorm:"primary_key"`
	BnsDate         string  `json:"bns_date" gorm:"column:bns_date"`
	Username        string  `json:"username" gorm:"column:username"`
	TotalPoolAmount float64 `json:"total_pool_amount" gorm:"column:total_pool_amount"`
}

// GetSysGhostSponsorPoolListFn get ent_member_crypto with dynamic condition
func GetSysGhostSponsorPoolListFn(date string, maxNumber int, debug bool) ([]*SysGhostSponsorPool, error) {
	var result []*SysGhostSponsorPool
	tx := db.Raw("SELECT * FROM (SELECT nick_name AS username, bns_date, SUM(pool_amount) AS total_pool_amount "+
		"FROM sys_ghost_sponsor_pool "+
		"WHERE DATE(bns_date) = ? "+
		"AND pool_amount > ? "+
		"AND status = ? "+
		"GROUP BY nick_name, bns_date "+
		"ORDER BY total_pool_amount DESC "+
		"LIMIT ? ) as a", date, 0, "A", maxNumber)

	if debug {
		tx = tx.Debug()
	}

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
