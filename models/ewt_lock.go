package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EwtDetail struct
type EwtLock struct {
	ID                   int    `gorm:"primary_key" json:"id"`
	MemberID             int    `json:"member_id"`
	EwalletTypeID        int    `json:"ewallet_type_id"`
	Register             int    `json:"register"`
	Topup                int    `json:"topup"`
	PurchasePin          int    `json:"purchase_pin"`
	Refund               int    `json:"refund"`
	InternalTransfer     int    `json:"internal_transfer"`
	ExternalTransfer     int    `json:"external_transfer"`
	TransferFromExchange int    `json:"transfer_from_exchange"`
	Exchange             int    `json:"exchange"`
	Withdrawal           int    `json:"withdrawal"`
	CreatedBy            string `json:"created_by"`
	CreatedAt            string `json:"created_at"`
}

// GetEwtLockFn get ewt_lock data with dynamic condition
func GetEwtLockFn(arrCond []WhereCondFn, debug bool) ([]*EwtLock, error) {
	var result []*EwtLock
	tx := db.Table("ewt_lock")
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

//get latest ewt_lock record by member_id - will only return 1 record
func GetEwtLockByMemberId(member_id int, wallet_id int) (*EwtLock, error) {
	var result EwtLock
	tx := db.Table("ewt_lock").Where("member_id = ? AND ewallet_type_id = ?", member_id, wallet_id).Order("id desc").Limit(1)

	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
