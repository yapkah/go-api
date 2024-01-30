package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusDiamondStarStruct struct
type TblBonusDiamondStarStruct struct {
	ID          int    `gorm:"primary_key" json:"id"`
	TBnsID      int    `json:"t_bns_id" gorm:"column:t_bns_id"`
	TDiamondID  int    `json:"t_diamond_id" gorm:"column:t_diamond_id"`
	TDownlineID int    `json:"t_downline_id" gorm:"column:t_downline_id"`
	NickName    string `json:"nick_name" gorm:"column:nick_name"`
}

// GetTblBonusDiamondStarFn get wod_room_mast data with dynamic condition
func GetTblBonusDiamondStarFn(arrCond []WhereCondFn, debug bool) ([]*TblBonusDiamondStarStruct, error) {
	var result []*TblBonusDiamondStarStruct
	tx := db.Table("tbl_bonus_diamond_star").
		Joins("INNER JOIN ent_member ON tbl_bonus_diamond_star.t_downline_id = ent_member.id").
		Select("tbl_bonus_diamond_star.*, ent_member.nick_name")

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
