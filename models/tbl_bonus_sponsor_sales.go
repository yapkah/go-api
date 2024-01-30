package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqBonusSponsorPassup struct
type TblqBonusSponsorSales struct {
	BnsID         string    `gorm:"bns_id" json:"bns_id"`
	MemberId      int       `json:"member_id"`
	DocNo         string    `json:"doc_no"`
	DownlineID    int       `json:"downline_id"`
	ILvl          int       `json:"i_lvl"`
	TotalBv       float64   `json:"total_bv"`
	DtCreated     time.Time `json:"dt_created"`
	DtFlush       time.Time `json:"dt_flush"`
	DtFlushAnnual time.Time `json:"dt_flush_annual"`
}

type TblqBonusSponsorTotalSales struct {
	TotalSales float64 `json:"total_sales"`
}

//get GetTotalSponsorBonusSalesByMemberId
func GetTotalSponsorBonusSalesByMemberId(mem_id int, bns_type string) (*TblqBonusSponsorTotalSales, error) {
	var (
		rwd TblqBonusSponsorTotalSales
	)

	query := db.Table("tblq_bonus_sponsor_sales as a").
		Select("SUM(a.total_bv) as total_sales")

	if mem_id != 0 {
		query = query.Where("a.member_id = ?", mem_id)
	}

	if bns_type != "" {
		if bns_type == "A" {
			query = query.Where("a.dt_flush IS NULL")
		}

		if bns_type == "B" {
			query = query.Where("a.dt_flush_annual IS NULL")
		}
	}

	err := query.Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rwd, nil
}

// GetTblqBonusSponsorSalesFn get tblq_bonus_sponsor_sales data with dynamic condition
func GetTblqBonusSponsorSalesFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusSponsorPassup, error) {
	var result []*TblqBonusSponsorPassup
	tx := db.Table("tblq_bonus_sponsor_sales")
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

type SponsorABnsFlushDate struct {
	FlushDate string `json:"flush_date"`
}

func GetSponsorBonusAFlushDateByMemberId(mem_id int) (*SponsorABnsFlushDate, error) {
	var (
		rwd SponsorABnsFlushDate
	)

	query := db.Table("tblq_bonus_sponsor_sales as a").
		Select("MIN(a.bns_id) as flush_date")

	if mem_id != 0 {
		query = query.Where("a.member_id = ?", mem_id)
	}

	err := query.Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rwd, nil
}
