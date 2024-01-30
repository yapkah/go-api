package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberTradingTransaction struct
type EntMemberTradingTransaction struct {
	ID               int       `gorm:"primary_key" json:"id"`
	MemberID         int       `json:"member_id"`
	DocNo            string    `json:"doc_no"`
	RefNo            string    `json:"ref_no"`
	DocDate          string    `json:"doc_date"`
	Platform         string    `json:"platform"`
	Strategy         string    `json:"strategy"`
	StrategyName     string    `json:"strategy_name"`
	CryptoPair       string    `json:"crypto_pair"`
	CryptoPairName   string    `json:"crypto_pair_name"`
	Type             string    `json:"type"`
	Num              float64   `json:"num"`
	Price            float64   `json:"price"`
	OrderId          string    `json:"order_id"`
	OrderType        string    `json:"order_type"`
	Remark2          string    `json:"remark2"`
	TotalBv          float64   `json:"total_bv"`
	TPrice           float64   `json:"t_price"`
	TQty             float64   `json:"t_qty"`
	TQuoteQty        float64   `json:"t_quote_qty"`
	TCommission      float64   `json:"t_commission"`
	TCommissionAsset string    `json:"t_commission_asset"`
	TTime            int64     `json:"t_time"`
	Timestamp        time.Time `json:"timestamp"`
}

// GetEntMemberTradingTransactionFn
func GetEntMemberTradingTransactionFn(arrCond []WhereCondFn, order string, debug bool) ([]*EntMemberTradingTransaction, error) {
	var result []*EntMemberTradingTransaction

	if order == "" {
		order = "DESC"
	}

	tx := db.Table("ent_member_trading_transaction").
		Select("ent_member_trading_transaction.*, sls_master.member_id, sls_master.total_bv, sls_master.ref_no, prd_master.name as strategy_name, sys_trading_crypto_pair_setup.name as crypto_pair_name").
		Joins("inner join sls_master ON sls_master.doc_no = ent_member_trading_transaction.doc_no").
		Joins("inner join prd_master ON prd_master.code = ent_member_trading_transaction.strategy").
		Joins("left join sys_trading_crypto_pair_setup ON sys_trading_crypto_pair_setup.code = ent_member_trading_transaction.crypto_pair").
		Order("ent_member_trading_transaction.timestamp " + order)

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

type EntMemberTradingDocDateFn struct {
	DocDate string `json:"doc_date"`
}

// GetEntMemberTradingDocDateFn
func GetEntMemberTradingDocDateFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberTradingDocDateFn, error) {
	var result []*EntMemberTradingDocDateFn

	tx := db.Table("ent_member_trading_transaction").
		Select("ent_member_trading_transaction.doc_date")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	tx = tx.Group("doc_date")

	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type TradingMaxQuoteQtyInGroup struct {
	DocNo       string  `json:"doc_no"`
	SumQuoteQty float64 `json:"sum_quote_qty"`
}

// GetTradingMaxQuoteQtyInGroupFn
func GetTradingMaxQuoteQtyInGroupFn(arrCond []WhereCondFn, order string, debug bool) ([]*TradingMaxQuoteQtyInGroup, error) {
	var result []*TradingMaxQuoteQtyInGroup

	tx := db.Table("ent_member_trading_transaction").
		Select("doc_no, sum(t_quote_qty) as sum_quote_qty").
		Group("doc_no").
		Order("sum_quote_qty DESC").
		Limit(1)

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
