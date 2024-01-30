package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"time"

	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/setting"
)

var db *gorm.DB

// Model struct
type Model struct {
	ID        int       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WhereCondFn
type WhereCondFn struct {
	Condition string
	CondValue interface{}
}

// OrderByFn
type OrderByFn struct {
	Condition string
}

// JoinFn
type JoinFn struct {
	JoinTable string
	JoinValue interface{}
}

// ArrModelFn
type ArrModelFn struct {
	Join  []JoinFn
	Where []WhereCondFn
}

// ArrUnionRawCondText
type ArrUnionRawCondText struct {
	Cond string
}

// Setup initializes the database instance
func Setup() {
	var err error
	db, err = gorm.Open(setting.DatabaseSetting.Type, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		setting.DatabaseSetting.User,
		setting.DatabaseSetting.Password,
		setting.DatabaseSetting.Host,
		setting.DatabaseSetting.Name))

	if err != nil {
		log.Fatalf("models.Setup err: %v", err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return setting.DatabaseSetting.TablePrefix + defaultTableName
	}

	db.SingularTable(true)
	// db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	// db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.DB().SetMaxIdleConns(0)
	db.DB().SetMaxOpenConns(30)
	db.DB().SetConnMaxLifetime((5 * time.Minute))
}

// CloseDB closes database connection (unnecessary)
func CloseDB() {
	defer db.Close()
}

// updateTimeStampForCreateCallback will set `CreatedAt`, `UpdatedAt` when creating
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now()
		if createTimeField, ok := scope.FieldByName("CreatedAt"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("UpdatedAt"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

// updateTimeStampForUpdateCallback will set `UpdatedAt` when updating
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("UpdatedAt", time.Now().Unix())
	}
}

// addExtraSpaceIfExist adds a separator
func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

// save function
func save(value interface{}) error {
	err := db.Save(value).Error
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// save function
func SaveTx(tx *gorm.DB, value interface{}) error {
	err := tx.Save(value).Error
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// transactions

// GetDB get db
func GetDB() *gorm.DB {
	return db
}

// Begin begin transactoin
func Begin() *gorm.DB {
	return db.Begin()
}

// BeginReadCommited begin read commited transaction
func BeginReadCommited() *gorm.DB {
	opt := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	return BeginTx(opt)
}

// BeginTx begins a transaction with options
// [sql.TxOptions] struct can refer to (https://godoc.org/database/sql#TxOptions)
// [sql.TxOptions.IsolationLevel] struct can refer to (https://godoc.org/database/sql#IsolationLevel)
func BeginTx(opts *sql.TxOptions) *gorm.DB {
	return db.BeginTx(context.Background(), opts)
}

// Commit commit transaction
func Commit(tx *gorm.DB) error {
	if err := tx.Commit().Error; err != nil {
		ErrorLog("transaction-commit", err.Error(), map[string]interface{}{"err": err})
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// Rollback rollback transaction
func Rollback(tx *gorm.DB) error {
	if err := tx.Rollback().Error; err != nil {
		ErrorLog("transaction-rollback", err.Error(), map[string]interface{}{"err": err})
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// Execute db execure
func Execute(tx *gorm.DB, query string, data []interface{}) error {
	if err := tx.Exec(query, data...).Error; err != nil {
		ErrorLog("db-execute", err.Error(), map[string]interface{}{"err": err})
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// Update Table Transaction together with db transaction records passing from prev function (usually tx contain of begin transaction feature)
func UpdatesFnTx(tx *gorm.DB, tableName string, arrCond []WhereCondFn, updateColumn map[string]interface{}, debug bool) error {
	tx = tx.Table(tableName)

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Updates(updateColumn).Error
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// Update Table Transaction without db transaction records passing from prev function
func UpdatesFn(tableName string, arrCond []WhereCondFn, updateColumn map[string]interface{}, debug bool) error {
	tx := db.Table(tableName)

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Updates(updateColumn).Error
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// Delete Table Transaction without db transaction records passing from prev function
func DeleteFn(tableName string, arrCond []WhereCondFn, debug bool) error {
	tx := db.Table(tableName)

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Delete(tx).Error
	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}
	return nil
}

// SQLDataPaginateStdReturn. use in standard return for sql pagination
type SQLPaginateStdReturn struct {
	CurrentPage           int64   `json:"current_page"`
	PerPage               int64   `json:"per_page"`
	TotalCurrentPageItems int64   `json:"total_current_page_items"`
	TotalPage             float64 `json:"total_page"`
	TotalPageItems        int64   `json:"total_page_items"`
}
