package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqAuctionRebate struct
type TblqAuctionRebate struct {
	BnsId          string    `json:"bns_id"`
	MemberId       int       `json:"member_id"`
	DocNo          string    `json:"doc_no"`
	HoldingNft     string    `json:"holding_nft"`
	HoldingNftType string    `json:"holding_nft_type"`
	FBv            float64   `json:"f_bv"`
	FPerc          float64   `json:"f_perc"`
	FBns           float64   `json:"f_bns"`
	FRate          float64   `json:"f_rate"`
	DtTimestamp    time.Time `json:"dt_timestamp"`
	DtPaid         time.Time `json:"dt_paid"`
}

type TblqAuctionRebateResult struct {
	TBnsId         string    `json:"t_bns_id"`
	NickName       string    `json:"nick_name"`
	DocNo          string    `json:"doc_no"`
	HoldingNftType string    `json:"holding_nft_type"`
	FBv            float64   `json:"f_bv"`
	FPerc          float64   `json:"f_perc"`
	FBns           float64   `json:"f_bns"`
	DtTimestamp    time.Time `json:"dt_timestamp"`
	DtPaid         time.Time `json:"dt_paid"`
}

// TblqBonusRebateFn get tblq_auction_rebates data with dynamic condition
func TblqAuctionRebateFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblqAuctionRebate, error) {
	var result []*TblqAuctionRebate
	tx := db.Table("tblq_auction_rebates")
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

//get auction rebate by memid
func GetAuctionRebateBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqAuctionRebateResult, error) {
	var (
		rwd []*TblqAuctionRebateResult
	)

	query := db.Table("tblq_auction_rebates as a").
		Select("a.bns_id as t_bns_id,b.nick_name,a.doc_no,a.holding_nft_type,a.f_bv,a.f_perc,a.f_bns,a.dt_timestamp,a.dt_paid").
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

type GlobalAuctionResult struct {
	TBnsId         string  `json:"t_bns_id"`
	AuctionRebates float64 `json:"auction_rebates"`
	AuctionLucky   float64 `json:"auction_lucky"`
}

func GetMemberAuctionBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*GlobalAuctionResult, error) {
	var (
		rwd []*GlobalAuctionResult
	)

	query := db.Table("rwd_period as a").
		Select("a.batch_code as t_bns_id,sum(b.f_bns) as auction_rebates").
		Joins("LEFT JOIN tblq_auction_rebates as b ON a.batch_code = b.bns_id and b.member_id = ?", mem_id)
		// Joins("LEFT JOIN tblq_auction_lucky as c ON a.batch_code = c.bns_id and c.member_id = ?", mem_id)

	if dateFrom != "" {
		query = query.Where("a.batch_code >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("a.batch_code <= ?", dateTo)
	}
	err := query.Group("a.batch_code").Order("a.batch_code desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}
