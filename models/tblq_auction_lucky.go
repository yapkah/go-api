package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqAuctionLucky struct
type TblqAuctionLucky struct {
	BnsId             string    `json:"bns_id"`
	MemberId          int       `json:"member_id"`
	DocNo             string    `json:"doc_no"`
	HoldingLuckyNo    string    `json:"holding_lucky_no"`
	HoldingLuckyCount string    `json:"holding_lucky_count"`
	FBv               float64   `json:"f_bv"`
	FPerc             float64   `json:"f_perc"`
	FBns              float64   `json:"f_bns"`
	FRate             float64   `json:"f_rate"`
	DtTimestamp       time.Time `json:"dt_timestamp"`
	DtPaid            time.Time `json:"dt_paid"`
}

type TblqAuctionLuckyResult struct {
	TBnsId            string    `json:"t_bns_id"`
	NickName          string    `json:"nick_name"`
	HoldingLuckyNo    string    `json:"holding_lucky_no"`
	HoldingLuckyCount string    `json:"holding_lucky_count"`
	DocNo             string    `json:"doc_no"`
	FBv               float64   `json:"f_bv"`
	FPerc             float64   `json:"f_perc"`
	FBns              float64   `json:"f_bns"`
	DtTimestamp       time.Time `json:"dt_timestamp"`
	DtPaid            time.Time `json:"dt_paid"`
}

// TblqAuctionLuckyFn get tblq_auction_lucky data with dynamic condition
func TblqAuctionLuckyFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqAuctionLucky, error) {
	var result []*TblqAuctionLucky
	tx := db.Table("tblq_auction_lucky")
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

//get auction lucky by memid
func GetAuctionLuckyBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqAuctionLuckyResult, error) {
	var (
		rwd []*TblqAuctionLuckyResult
	)

	query := db.Table("tblq_auction_lucky as a").
		Select("a.bns_id as t_bns_id,b.nick_name,a.holding_lucky_no,a.holding_lucky_count,a.doc_no,a.f_bv,a.f_perc,a.f_bns,a.dt_timestamp,a.dt_paid").
		Joins("JOIN ent_member as b ON a.member_id = b.id")

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
