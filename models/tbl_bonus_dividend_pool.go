package models

import (
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EwtDetail struct
type TblBonusDividendPool struct {
	ID         int       `gorm:"primary_key" json:"id"`
	TBnsId     string    `json:"t_bns_id"`
	TType      int       `json:"t_type"`
	TMemberId  int       `json:"t_member_id"`
	NDiamondId int       `json:"n_diamond_id"`
	NShare     float64   `json:"n_share"`
	FBns       float64   `json:"f_bns"`
	DtCreated  time.Time `json:"dt_created"`
}

type TblBonusDividendPoolResult struct {
	ID        int       `gorm:"primary_key" json:"id"`
	TBnsId    string    `json:"t_bns_id"`
	TType     int       `json:"t_type"`
	DtCreated time.Time `json:"dt_created"`
	NickName  string    `json:"nick_name"`
	NShare    string    `json:"n_share"`
	FBns      string    `json:"f_bns"`
}

// TblBonusDividendPoolFn get tbl_bonus_dividend_pool data with dynamic condition
func TblBonusDividendPoolFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblBonusDividendPool, error) {
	var result []*TblBonusDividendPool
	tx := db.Table("tbl_bonus_dividend_pool")
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

//get dividend pool by memid
func GetDividendPoolBnsByMemId(mem_id int, dateFrom string, dateTo string) ([]*TblBonusDividendPoolResult, error) {
	var (
		rwd []*TblBonusDividendPoolResult
	)

	query := db.Table("tbl_bonus_dividend_pool as a").
		Select("a.id,a.t_bns_id,a.t_type,a.dt_created,b.nick_name, a.n_share , a.f_bns").
		Joins("JOIN ent_member as b ON a.t_member_id = b.id").
		Joins("JOIN wod_member_diamond as c ON a.n_diamond_id = c.id").
		Joins("LEFT JOIN rwd_period as d ON a.t_bns_id = d.batch_code")

	if mem_id != 0 {
		query = query.Where("a.t_member_id = ?", mem_id)
	}

	if dateFrom != "" {
		dateFrom = strings.Replace(dateFrom, "-", "", 2) + "0000"
		query = query.Where("a.t_bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		dateTo = strings.Replace(dateTo, "-", "", 2) + "2359"
		query = query.Where("a.t_bns_id <= ?", dateTo)
	}

	err := query.Order("a.id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}
