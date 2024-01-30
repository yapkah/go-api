package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqBonusStrategyProfit struct
type TblqBonusStrategyProfit struct {
	ID             int       `gorm:"primary_key" json:"id"`
	BnsID          string    `json:"bns_id"`
	MemberID       int       `json:"member_id"`
	DocNo          string    `json:"doc_no"`
	PrincipleValue float64   `json:"principle_value"`
	Platform       string    `json:"platform"`
	OrderId        string    `json:"order_id"`
	Strategy       string    `json:"strategy"`
	StrategyName   string    `json:"strategy_name"`
	CryptoPair     string    `json:"crypto_pair"`
	CryptoPairName string    `json:"crypto_pair_name"`
	FProfit        float64   `json:"f_profit"`
	DtTimestamp    time.Time `gorm:"column:dt_timestamp" json:"dt_timestamp"`
}

// GetTblqBonusStrategyProfitFn
func GetTblqBonusStrategyProfitFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusStrategyProfit, error) {
	var result []*TblqBonusStrategyProfit
	tx := db.Table("tblq_bonus_strategy_profit").
		Joins(" left join sys_trading_crypto_pair_setup ON sys_trading_crypto_pair_setup.code = tblq_bonus_strategy_profit.crypto_pair").
		Joins(" inner join prd_master ON prd_master.code = tblq_bonus_strategy_profit.strategy").
		Joins(" inner join sls_master ON sls_master.doc_no = tblq_bonus_strategy_profit.doc_no").
		Select("tblq_bonus_strategy_profit.*, sls_master.total_amount as principle_value, sys_trading_crypto_pair_setup.name as crypto_pair_name, prd_master.name as strategy_name").
		Order("tblq_bonus_strategy_profit.dt_timestamp DESC")

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

// TblqBonusStrategyProfit struct
type LeaderboardTypeAReturnStruct struct {
	ID            int       `gorm:"primary_key" json:"id"`
	BnsID         string    `json:"bns_id"`
	MemberID      int       `json:"member_id"`
	DocNo         string    `json:"doc_no"`
	Strategy      string    `json:"strategy"`
	CryptoPair    string    `json:"crypto_pair"`
	Username      string    `json:"username"`
	FProfit       float64   `json:"f_profit"`
	ProfilePic    string    `json:"profile_pic"`
	SponsorReward float64   `json:"sponsor_reward"`
	DtTimestamp   time.Time `json:"dt_timestamp"`
}

// GetLeaderboardTypeAFn
func GetLeaderboardTypeAFn(arrCond []WhereCondFn, limit int, debug bool) ([]*LeaderboardTypeAReturnStruct, error) {
	var result []*LeaderboardTypeAReturnStruct
	tx := db.Table("tblq_bonus_strategy_profit").
		Joins(" inner join ent_member ON ent_member.id = tblq_bonus_strategy_profit.member_id").
		Joins(" left join tbl_bonus ON tblq_bonus_strategy_profit.member_id = tbl_bonus.t_member_id AND tbl_bonus.t_bns_id = tblq_bonus_strategy_profit.bns_id").
		Select("tblq_bonus_strategy_profit.*, ent_member.nick_name as username,ent_member.avatar as profile_pic,tbl_bonus.f_bns_sponsor as sponsor_reward").
		Order("tblq_bonus_strategy_profit.f_profit DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
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

// GroupTblqBonusStrategyProfit struct
type GroupTblqBonusStrategyProfit struct {
	MemberID    int     `json:"member_id"`
	TotalAmount float64 `json:"total_amount"`
}

// GetGroupTblqBonusStrategyProfitFn
func GetGroupTblqBonusStrategyProfitFn(arrCond []WhereCondFn, debug bool) ([]*GroupTblqBonusStrategyProfit, error) {
	var result []*GroupTblqBonusStrategyProfit
	tx := db.Table("tblq_bonus_strategy_profit").
		Select("tblq_bonus_strategy_profit.member_id, SUM(tblq_bonus_strategy_profit.f_profit) as total_amount").
		Group("tblq_bonus_strategy_profit.member_id")

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
