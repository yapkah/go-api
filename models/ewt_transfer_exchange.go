package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EwtTransferExchanfe struct
type EwtTransferExchange struct {
	ID              int       `gorm:"primary_key" json:"id"`
	DocNo           string    `json:"doc_no"`
	MemberId        int       `json:"member_id`
	EwalletTypeId   int       `json:"ewallet_type_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	CryptoAddrTo    string    `json:"crypto_addr_to"`
	SigningKey      string    `json:"signing_key"`
	TranHash        string    `json:"tran_hash"`
	Remark          string    `json:"remark"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       int       `json:"created_by"`
}

type EwtTransferExchangeDetail struct {
	ID              int       `gorm:"primary_key" json:"id"`
	DocNo           string    `json:"doc_no"`
	MemberId        int       `json:"member_id`
	EwalletTypeId   int       `json:"ewallet_type_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	CryptoAddrTo    string    `json:"crypto_addr_to"`
	SigningKey      string    `json:"signing_key"`
	TranHash        string    `json:"tran_hash"`
	Remark          string    `json:"remark"`
	Status          string    `json:"status"`
	StatusDesc      string    `json:"status_desc"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       int       `json:"created_by"`
}

// GetEwtTransferExchangeFn get ewt_transfer_exchange data with dynamic condition
func GetEwtTransferExchangeFn(arrCond []WhereCondFn, debug bool) ([]*EwtTransferExchange, error) {
	var result []*EwtTransferExchange
	tx := db.Table("ewt_transfer_exchange")
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

func AddEwtTransferExchange(tx *gorm.DB, saveData EwtTransferExchange) (*EwtTransferExchange, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtTransferExchange-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

func GetEwtTransferExchangeDetailByDocNo(docNo string) (*EwtTransferExchangeDetail, error) {
	var ewt EwtTransferExchangeDetail

	query := db.Table("ewt_transfer_exchange a").
		Select("a.*,b.name as status_desc").
		Joins("left join sys_general b ON a.status = b.code and b.type='general-status'")

	if docNo != "" {
		query = query.Where("a.doc_no = ?", docNo)
	}

	err := query.Order("id desc").Find(&ewt).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &ewt, nil
}
