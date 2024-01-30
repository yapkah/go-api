package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblqBonusP2PMatching struct
type TblqBonusP2PMatching struct {
	BnsId         string    `json:"t_bns_id"`
	MemberId      int       `json:"member_id"`
	DownlineId    int       `json:"t_downline_id"`
	WalletTyepeId int       `json:"wallet_type_id"`
	ILvl          int       `json:"i_lvl"`
	ILvlPaid      int       `json:"i_lvl_paid"`
	FBv           float64   `json:"f_bv"`
	FPerc         float64   `json:"f_perc"`
	FBns          float64   `json:"f_bns"`
	BurnBv        float64   `json:"burn_bv"`
	DtCreated     time.Time `json:"dt_created"`
	DtPaid        time.Time `json:"dt_paid"`
}

type TblqBonusP2PMatchingResult struct {
	TBnsId        string    `json:"t_bns_id"`
	Username      string    `json:"username"`
	DownlineId    string    `json:"downline_id"`
	WalletTyepeId int       `json:"wallet_type_id"`
	ILvl          string    `json:"i_lvl"`
	ILvlPaid      string    `json:"i_lvl_paid"`
	FBv           float64   `json:"f_bv"`
	FPerc         float64   `json:"f_perc"`
	FBns          float64   `json:"f_bns"`
	BurnBv        float64   `json:"burn_bv"`
	BurnBns       float64   `json:"burn_bns"`
	FRate         float64   `json:"f_rate"`
	DtCreated     time.Time `json:"dt_created"`
	DtPaid        time.Time `json:"dt_paid"`
}

// TblqBonusP2PMatchingFn get tblq_bonus_matching data with dynamic condition
func TblqBonusP2PMatchingFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqBonusP2PMatching, error) {
	var result []*TblqBonusP2PMatching
	tx := db.Table("tblq_bonus_p2p_matching")
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

//get p2p matching bonus by memid
func GetP2PMatchingBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusP2PMatchingResult, error) {
	var (
		rwd []*TblqBonusP2PMatchingResult
	)

	query := db.Table("tblq_bonus_p2p_matching as a").
		Select("a.bns_id as t_bns_id,b.nick_name as username,down.nick_name as downline_id,a.i_lvl,a.i_lvl_paid,a.f_bv,a.f_perc,a.f_bns,a.burn_bv,a.burn_bns,bonus.f_rate,a.dt_created,a.dt_paid").
		Joins("JOIN tbl_p2p_bonus as bonus ON a.bns_id = bonus.t_bns_id AND a.member_id = bonus.t_member_id").
		Joins("JOIN ent_member as b ON a.member_id = b.id").
		Joins("JOIN ent_member as down ON down.id = a.downline_id")

	if mem_id != 0 {
		query = query.Where("a.member_id = ?", mem_id)
	}

	if dateFrom != "" {
		// dateFrom = strings.Replace(dateFrom, "-", "", 2) + "0000"
		query = query.Where("a.bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		// dateTo = strings.Replace(dateTo, "-", "", 2) + "2359"
		query = query.Where("a.bns_id <= ?", dateTo)
	}

	err := query.Order("a.bns_id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}
