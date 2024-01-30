package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// StrategyEvents struct
type StrategyEvents struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Package   string    `json:"package"`
	TimeStart time.Time `json:"time_start"`
	TimeEnd   time.Time `json:"time_end"`
	Title     string    `json:"title"`
	Desc      string    `json:"desc"`
	Path      string    `json:"path"`
	Status    string    `json:"status"`
	Seq       int       `json:"seq"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// GetStrategyEventsFn
func GetStrategyEventsFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*StrategyEvents, error) {
	var result []*StrategyEvents
	tx := db.Table("strategy_events a")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("seq asc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type StrategyEventsDetails struct {
	ID          int       `gorm:"primary_key" json:"id"`
	Title       string    `json:"title"`
	Desc        string    `json:"desc"`
	Package     string    `json:"package"`
	Status      string    `json:"status"`
	TimeStart   time.Time `json:"time_start"`
	TimeEnd     time.Time `json:"time_end"`
	Path        string    `json:"path"`
	SqlQuery    string    `json:"sql_query"`
	TableHeader string    `json:"table_header"`
	Seq         int       `json:"seq"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetStrategyEventsDetailsFn
func GetStrategyEventsDetailsFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*StrategyEventsDetails, error) {
	var result []*StrategyEventsDetails
	tx := db.Table("strategy_events").
		Select("strategy_events.*, strategy_events_query.sql_query, strategy_events_query.table_header").
		Joins("inner join strategy_events_query ON strategy_events.query_id = strategy_events_query.id")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("strategy_events.seq asc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
