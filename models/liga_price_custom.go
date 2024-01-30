package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// LigaCustomNetworkPrice struct
type LigaCustomNetworkPrice struct {
	TokenPrice float64 `json:"token_price"`
	BNetwork   int     `json:"b_network"`
	MemberId   int     `json:"member_id"`
}

func GetLigaCustomNetworkPrice(memId int) (*LigaCustomNetworkPrice, error) {
	var price LigaCustomNetworkPrice

	query := db.Table("liga_price_custom a").
		Select("a.token_price, a.b_network").
		Joins("inner join ent_member_lot_sponsor b ON a.member_id = b.member_id").
		Joins("inner join ent_member_lot_sponsor c ON b.i_lft <= c.i_lft AND b.i_rgt >= c.i_rgt").
		Joins("inner join ent_member d ON c.member_id = d.id").
		Where("a.status = ? AND DATE(a.date) = DATE(CURDATE())", "A")

	if memId != 0 {
		query = query.Where("d.id = ?", memId)
	}

	err := query.Find(&price).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &price, nil
}

// LigaCustomNetworkPriceFnStruct struct
type LigaCustomNetworkPriceFnStruct struct {
	MemberId         int       `gorm:"column:member_id" json:"member_id"`
	DownlineMemberID int       `gorm:"column:downline_member_id" json:"downline_member_id"`
	DownlineNickName string    `gorm:"column:downline_nick_name" json:"downline_nick_name"`
	BNetwork         int       `gorm:"column:b_network" json:"b_network"`
	TokenPrice       float64   `gorm:"column:token_price" json:"token_price"`
	Status           string    `gorm:"column:status" json:"status"`
	Date             time.Time `gorm:"column:date" json:"date"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy        string    `gorm:"column:created_by" json:"created_by"`
}

// GetLigaCustomNetworkPriceFn get liga_price_custom data with dynamic condition
func GetLigaCustomNetworkPriceFn(arrCond []WhereCondFn, limit int, debug bool) ([]*LigaCustomNetworkPriceFnStruct, error) {
	var result []*LigaCustomNetworkPriceFnStruct
	tx := db.Table("liga_price_custom").
		Select("liga_price_custom.*, d.id AS 'downline_member_id', d.nick_name AS 'downline_nick_name'").
		Joins("inner join ent_member_lot_sponsor b ON liga_price_custom.member_id = b.member_id").
		Joins("inner join ent_member_lot_sponsor c ON b.i_lft <= c.i_lft AND b.i_rgt >= c.i_rgt").
		Joins("inner join ent_member d ON c.member_id = d.id").
		Order("liga_price_custom.created_at DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			if v.CondValue != nil {
				tx = tx.Where(v.Condition, v.CondValue)
			} else {
				tx = tx.Where(v.Condition)
			}
		}
	}
	if debug {
		tx = tx.Debug()
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
