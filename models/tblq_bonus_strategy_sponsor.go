package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblqBonusStrategySponsor struct
type TblqBonusStrategySponsor struct {
	ID               int       `gorm:"primary_key" json:"id"`
	BnsID            string    `json:"bns_id"`
	MemberID         int       `json:"member_id"`
	DownlineNickName string    `json:"downline_nick_name"`
	DocNo            string    `json:"doc_no"`
	Strategy         string    `json:"strategy"`
	CryptoPair       string    `json:"crypto_pair"`
	CryptoPairName   string    `json:"crypto_pair_name"`
	ILvl             int       `json:"i_lvl"`
	FBns             float64   `json:"f_bns"`
	FPerc            float64   `json:"f_perc"`
	Timestamp        time.Time `json:"timestamp"`
}

// GetTblqBonusStrategySponsorFn
func GetTblqBonusStrategySponsorFn(arrCond []WhereCondFn, debug bool) ([]*TblqBonusStrategySponsor, error) {
	var result []*TblqBonusStrategySponsor
	tx := db.Table("tblq_bonus_strategy_sponsor").
		Joins(" inner join (SELECT doc_no, MAX(strategy) as strategy, MAX(crypto_pair) as crypto_pair FROM `ent_member_trading_transaction` WHERE `status` = 'A' GROUP BY doc_no) AS ent_member_trading_transaction ON ent_member_trading_transaction.doc_no = tblq_bonus_strategy_sponsor.doc_no").
		Joins(" left join sys_trading_crypto_pair_setup ON sys_trading_crypto_pair_setup.code = ent_member_trading_transaction.crypto_pair").
		Joins(" inner join sls_master ON sls_master.doc_no = tblq_bonus_strategy_sponsor.doc_no").
		Joins(" inner join sls_master_bot_setting ON sls_master_bot_setting.sls_master_id = sls_master.id").
		Joins(" inner join prd_master ON prd_master.id = sls_master.prd_master_id").
		Joins(" inner join ent_member ON tblq_bonus_strategy_sponsor.downline_id = ent_member.id").
		Select("tblq_bonus_strategy_sponsor.*, ent_member.nick_name as downline_nick_name, ent_member_trading_transaction.strategy, ent_member_trading_transaction.crypto_pair, sys_trading_crypto_pair_setup.name as crypto_pair_name").
		Order("tblq_bonus_strategy_sponsor.bns_id DESC")

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

// GroupTblqBonusStrategySponsor struct
type GroupTblqBonusStrategySponsor struct {
	MemberID    int     `json:"member_id"`
	TotalAmount float64 `json:"total_amount"`
}

// GetGroupTblqBonusStrategySponsorFn
func GetGroupTblqBonusStrategySponsorFn(arrCond []WhereCondFn, debug bool) ([]*GroupTblqBonusStrategySponsor, error) {
	var result []*GroupTblqBonusStrategySponsor
	tx := db.Table("tblq_bonus_strategy_sponsor").
		Select("tblq_bonus_strategy_sponsor.member_id, SUM(tblq_bonus_strategy_sponsor.f_bns) as total_amount").
		Group("tblq_bonus_strategy_sponsor.member_id")

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
