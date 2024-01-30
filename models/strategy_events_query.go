package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// StrategyEventsQuery struct
type StrategyEventsQuery struct {
	ID                int       `gorm:"primary_key" json:"id"`
	StarategyEventsID string    `json:"strategy_events_id"`
	SqlQuery          string    `json:"sql_query"`
	TableHeader       string    `json:"table_header"`
	CreatedBy         string    `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
}

// GetStrategyEventsQueryFn
func GetStrategyEventsQueryFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*StrategyEventsQuery, error) {
	var result []*StrategyEventsQuery
	tx := db.Table("strategy_events_query")

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

type GetStrategyEventsQueryStruct struct {
	MemberID    int     `json:"member_id"`
	NickName    string  `json:"nick_name"`
	Avatar      string  `json:"avatar"`
	CountryCode string  `json:"country_code"`
	TotalAmount float64 `json:"total_amount"`
}

func GetStrategyEventsQuery(query string, debug bool) ([]*GetStrategyEventsQueryStruct, error) {
	var result []*GetStrategyEventsQueryStruct
	tx := db.Raw(query)

	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
