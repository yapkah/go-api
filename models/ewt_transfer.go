package models

import (
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EwtTransfer struct
type EwtTransfer struct {
	ID             int       `gorm:"primary_key" json:"id"`
	MemberIdFrom   int       `json:"member_id_from"`
	MemberIdTo     int       `json:"member_id_to"`
	DocNo          string    `json:"doc_no"`
	EwtTypeFrom    int       `json:"ewt_type_from"`
	EwtTypeTo      int       `json:"ewt_type_to"`
	TransferAmount float64   `json:"transfer_amount"`
	AdminFee       float64   `json:"admin_fee"`
	NettAmount     float64   `json:"nett_amount"`
	CryptoAddrTo   string    `json:"crypto_addr_to"`
	Remark         string    `json:"remark"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      int       `json:"created_by"`
	// ApprovableAt   time.Time `json:"approvable_at"`
}

type EwtTransferDetail struct {
	ID             int       `gorm:"primary_key" json:"id"`
	MemberIdFrom   int       `json:"member_id_from"`
	MemberIdTo     int       `json:"member_id_to"`
	MemberFrom     string    `json:"member_from"`
	MemberTo       string    `json:"member_to"`
	WalletFrom     string    `json:"wallet_from"`
	WalletTo       string    `json:"wallet_to"`
	DocNo          string    `json:"doc_no"`
	EwtTypeFrom    int       `json:"ewt_type_from"`
	EwtTypeTo      int       `json:"ewt_type_to"`
	TransferAmount float64   `json:"transfer_amount"`
	CryptoAddrTo   string    `json:"crypto_addr_to"`
	Remark         string    `json:"remark"`
	Reason         string    `json:"remark"`
	Status         string    `json:"status"`
	StatusDesc     string    `json:"status_desc"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      int       `json:"created_by"`
}

// GetEwtTransferFn get ewt_transfer data with dynamic condition
func GetEwtTransferFn(arrCond []WhereCondFn, selectColumn string, debug bool) (*EwtTransfer, error) {
	var result EwtTransfer
	tx := db.Table("ewt_transfer")
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
		os.Exit(1)
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

// AddEwtTransfer add ewt_transfer records`
func AddEwtTransfer(tx *gorm.DB, saveData EwtTransfer) (*EwtTransfer, error) {
	if err := tx.Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtTransfer-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

func GetEwtTransferDetailByDocNo(docNo string) (*EwtTransferDetail, error) {
	var ewt EwtTransferDetail

	query := db.Table("ewt_transfer a").
		Select("a.*,b.name as status_desc,c.nick_name as member_to,e.nick_name as member_from,d.ewallet_type_name as wallet_to,f.ewallet_type_name as wallet_from").
		Joins("inner join sys_general b ON a.status = b.code and b.type='general-status'").
		Joins("inner join ent_member c ON a.member_id_to = c.id").
		Joins("inner join ewt_setup d ON a.ewt_type_to = d.id").
		Joins("inner join ent_member e ON a.member_id_from = e.id").
		Joins("inner join ewt_setup f ON a.ewt_type_from = f.id")

	if docNo != "" {
		query = query.Where("a.doc_no = ?", docNo)
	}

	err := query.Order("id desc").Find(&ewt).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &ewt, nil
}

type SumTransferStruct struct {
	TotalTransfer float64 `json:"total_transfer"`
}

func GetSumTotalTransferFn(arrCond []WhereCondFn, debug bool) (*SumTransferStruct, error) {
	var result SumTransferStruct
	tx := db.Table("ewt_transfer").
		Select("SUM(ewt_transfer.transfer_amount) AS 'total_transfer'")

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

	return &result, nil
}
