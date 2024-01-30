package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AddWodMemberDividendTranStruct struct
type AddWodMemberDividendTranStruct struct {
	ID        int     `gorm:"primary_key" gorm:"column:id" json:"id"`
	MemberID  int     `json:"member_id" gorm:"column:member_id"`
	DiamondID int     `json:"diamond_id" gorm:"column:diamond_id"`
	IncomeCap float64 `json:"income_cap" gorm:"column:income_cap"`
	RoomBatch string  `json:"room_batch" gorm:"column:room_batch"`
	Remark    string  `json:"remark" gorm:"column:remark"`
}

// AddWodMemberDividendTran add wod member dividend transaction
func AddWodMemberDividendTran(arrSaveData AddWodMemberDividendTranStruct) (*AddWodMemberDividendTranStruct, error) {
	if err := db.Table("wod_member_dividend_tran").Create(&arrSaveData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrSaveData, nil
}

// WodMemberDividendTran struct
type WodMemberDividendTran struct {
	ID        int     `gorm:"primary_key" gorm:"column:id" json:"id"`
	MemberID  int     `json:"member_id" gorm:"column:member_id"`
	DiamondID int     `json:"diamond_id" gorm:"column:diamond_id"`
	IncomeCap float64 `json:"income_cap" gorm:"column:income_cap"`
	RoomBatch string  `json:"room_batch" gorm:"column:room_batch"`
	Remark    string  `json:"remark" gorm:"column:remark"`
}

// GetWodMemberDividendTran get wod member dividend transaction with dynamic condition
func GetWodMemberDividendTran(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*WodMemberDividendTran, error) {
	var result []*WodMemberDividendTran
	tx := db.Table("wod_member_dividend_tran")
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

// MemberTotalIncomeCap struct
type MemberTotalIncomeCap struct {
	TotalIncomeCap float64 `json:"total_income_cap" gorm:"column:total_income_cap"`
}

// GetMemberTotalIncomeCapByDiamondID get member total income cap by diamond id
func GetMemberTotalIncomeCapByDiamondID(diamondID int) (*MemberTotalIncomeCap, error) {
	var result MemberTotalIncomeCap
	tx := db.Table("wod_member_dividend_tran").
		Select("SUM(wod_member_dividend_tran.income_cap) as total_income_cap").
		Where("wod_member_dividend_tran.diamond_id = ?", diamondID)

	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
