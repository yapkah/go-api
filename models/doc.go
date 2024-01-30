package models

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// Docs struct
type Docs struct {
	Id          int    `gorm:"primary_key" json:"id"`
	DocNoPrefix string `gorm:"column(doc_no_prefix)" json:"doc_no_prefix"`
	RunningNo   int    `gorm:"column(running_no)" json:"running_no"`
	RunningType string `gorm:"column(running_type)" json:"running_type"`
	StartNo     int    `gorm:"column(start_no)" json:"start_no"`
	DocLength   int    `gorm:"column(doc_length)" json:"doc_length"`
	TableName   string `gorm:"column(table_name)" json:"table_name"`
}

// Docs struct
type GeneralDocs struct {
	Id     int    `gorm:"primary_key" json:"id"`
	Type   string `gorm:"column(type)" json:"type"`
	Code   string `gorm:"column(code)" json:"code"`
	Name   string `gorm:"column(name)" json:"name"`
	Status string `gorm:"column(status)" json:"status"`
}

// GetRunningDocNo get running doc no
func GetRunningDocNo(docType string, tx *gorm.DB) (string, error) {
	//type MemberEwtLockForm struct {
	//	ID              int       `gorm:"primary_key" json:"id"`
	//}
	var (
		docs   Docs
		doc_no string
	)

	//var lock MemberEwtLockForm
	//db.Exec("set transaction isolation level serializable") // this statement ensures synchronicity at the database level
	//tx := db.Begin()
	db.Table("sys_doc_no").Where("doc_no_prefix = ?", docType).First(&docs)

	err := tx.Table("sys_doc_no").Exec("SELECT * FROM sys_doc_no WHERE id = ? FOR UPDATE", docs.Id).
		First(&docs).Error

	//err = tx.Table("ewt_wallet_transaction").Exec("SELECT * FROM ewt_wallet_transaction WHERE id = ? FOR UPDATE", 200).
	//	First(&lock).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if docs.RunningType == "random" { // random
		repeated := true
		attemptCount := 0
		maxAttempt := 200
		doc_no = docs.DocNoPrefix

		// continue loop is repeated == true
		for repeated {
			for x := 0; x < docs.DocLength; x++ {
				randomDigits := RandomInt(0, 9)
				doc_no = fmt.Sprintf("%s%d", doc_no, randomDigits)
			}

			// validate if doc_no is repeated
			if docs.TableName == "sls_master" {
				arrSlsMasterFn := make([]WhereCondFn, 0)
				arrSlsMasterFn = append(arrSlsMasterFn,
					WhereCondFn{Condition: "? IN(sls_master.doc_no, sls_master.batch_no) ", CondValue: doc_no},
				)
				arrSlsMaster, _ := GetSlsMasterFn(arrSlsMasterFn, "", false)
				if len(arrSlsMaster) == 0 {
					repeated = false // break loop if not repeated
				}
			} else if docs.TableName == "ent_member_membership_log" {
				arrEntMemberMembershipLogFn := make([]WhereCondFn, 0)
				arrEntMemberMembershipLogFn = append(arrEntMemberMembershipLogFn,
					WhereCondFn{Condition: " ent_member_membership_log.doc_no = ? ", CondValue: doc_no},
				)
				arrEntMemberMembershipLog, _ := GetEntMemberMembershipLog(arrEntMemberMembershipLogFn, "", false)
				if len(arrEntMemberMembershipLog) == 0 {
					repeated = false // break loop if not repeated
				}
			} else if docs.TableName == "ent_member_trading_deposit" {
				arrEntMemberTradingDepositFn := make([]WhereCondFn, 0)
				arrEntMemberTradingDepositFn = append(arrEntMemberTradingDepositFn,
					WhereCondFn{Condition: " ent_member_trading_deposit.doc_no = ? ", CondValue: doc_no},
				)
				arrEntMemberTradingDeposit, _ := GetEntMemberTradingDeposit(arrEntMemberTradingDepositFn, "", false)
				if len(arrEntMemberTradingDeposit) == 0 {
					repeated = false // break loop if not repeated
				}
			} else if docs.TableName == "ent_member_trading_deposit_withdraw" {
				arrEntMemberTradingDepositWithdrawFn := make([]WhereCondFn, 0)
				arrEntMemberTradingDepositWithdrawFn = append(arrEntMemberTradingDepositWithdrawFn,
					WhereCondFn{Condition: " ent_member_trading_deposit_withdraw.doc_no = ? ", CondValue: doc_no},
				)
				arrEntMemberTradingDepositWithdraw, _ := GetEntMemberTradingDepositWithdraw(arrEntMemberTradingDepositWithdrawFn, "", false)
				if len(arrEntMemberTradingDepositWithdraw) == 0 {
					repeated = false // break loop if not repeated
				}
			} else {
				repeated = false // break loop without further checking if not repeated
			}

			attemptCount++

			if attemptCount >= maxAttempt {
				return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "generate_random_doc_no_reach_max_attempt", Data: nil}
			}
		}

	} else { // auto
		formatter := fmt.Sprintf("%08d", docs.StartNo)
		doc_no = docs.DocNoPrefix + formatter
	}

	return doc_no, nil
}

// UpdateRunningDocNo get running doc no
func UpdateRunningDocNo(docType string, tx *gorm.DB) error {
	var docs Docs
	err := db.Table("sys_doc_no").
		Where("doc_no_prefix = ?", docType).First(&docs).Error

	if err != nil {
		return errors.New("update_doc_prefix_not_found")
	}

	err = tx.Exec("UPDATE sys_doc_no SET start_no = ? WHERE doc_no_prefix = ? AND id = ?", docs.StartNo+1, docType, docs.Id).Error

	if err != nil {
		return err

	}
	return nil
}

// GetDocNoPrefix get doc prefix
func GetDocNoPrefix(TransType string) (string, error) {
	var docs Docs
	err := db.Table("sys_doc_no").
		Where("doc_type = ?", TransType).First(&docs).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return "", &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return docs.DocNoPrefix, nil
}

// Returns an int >= min, < max
func RandomInt(min, max int) int {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(max-min+1) + min
}
