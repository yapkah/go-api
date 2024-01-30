package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EwtWithdraw struct
type EwtWithdraw struct {
	ID              int    `gorm:"primary_key" json:"id"`
	DocNo           string `json:"doc_no"`
	MemberId        int    `json:"member_id`
	EwalletTypeId   int    `json:"ewallet_type_id"`
	EwalletTypeCode string `json:"ewallet_type_code"`
	EwalletTypeIdTo int    `json:"ewallet_type_id_to"`
	CurrencyCode    string `json:"currency_code"`
	Type            string `json:"type"`
	TransactionType string `json:"transaction_type"`
	// CryptoType            string    `json:"crypto_type"`
	TransDate             time.Time `json:"trans_date"`
	Markup                float64   `json:"markup"`
	NetAmount             float64   `json:"net_amount"`
	TotalOut              float64   `json:"total_out"`
	ChargesType           string    `json:"charges_type"`
	AdminFee              float64   `json:"admin_fee"`
	ConversionRate        float64   `json:"conversion_rate"`
	ConvertedTotalAmount  float64   `json:"converted_total_amount"`
	ConvertedNetAmount    float64   `json:"converted_net_amount"`
	ConvertedAdminFee     float64   `json:"converted_admin_fee"`
	ConvertedCurrencyCode string    `json:"converted_currency_code"`
	CryptoAddrTo          string    `json:"crypto_addr_to"`
	// CryptoAddrReturn      string    `json:"crypto_addr_return"`
	GasFee    float64   `json:"gas_fee"`
	GasPrice  string    `json:"gas_price"`
	TranHash  string    `json:"tran_hash"`
	Remark    string    `json:"remark"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int       `json:"created_by"`
	ExpiredAt time.Time `json:"expired_at"`
}

type EwtWithdrawDetail struct {
	ID              int    `gorm:"primary_key" json:"id"`
	DocNo           string `json:"doc_no"`
	MemberId        int    `json:"member_id_from"`
	EwalletTypeId   int    `json:"ewallet_type_id"`
	EwalletTypeIdTo int    `json:"ewallet_type_id_to"`
	EwalletFrom     string `json:"ewallet_from"`
	EwalletTo       string `json:"ewallet_to"`
	CurrencyCode    string `json:"currency_code"`
	Type            string `json:"type"`
	// CryptoType            string    `json:"crypto_type"`
	TransDate             time.Time `json:"trans_date"`
	NetAmount             float64   `json:"net_amount"`
	TotalOut              float64   `json:"total_out"`
	AdminFee              float64   `json:"admin_fee"`
	Markup                float64   `json:"markup"`
	ConvertedNetAmount    float64   `json:"converted_net_amount"`
	ConvertedAdminFee     float64   `json:"converted_admin_fee"`
	ConvertedCurrencyCode string    `json:"converted_currency_code"`
	ConversionRate        float64   `json:"conversion_rate"`
	CryptoAddrTo          string    `json:"crypto_addr_to"`
	CryptoAddId           string    `json:"crypto_addr_id"`
	GasFee                float64   `json:"gas_fee"`
	GasPrice              string    `json:"gas_price"`
	Remark                string    `json:"remark"`
	TranHash              string    `json:"tran_hash"`
	Status                string    `json:"status"`
	StatusDesc            string    `json:"status_desc"`
	CreatedAt             time.Time `json:"created_at"`
	CreatedBy             string    `json:"created_by"`
	ExpiredAt             time.Time `json:"expired_at"`
}

// GetEwtWithdrawFn get ewt_withdraw data with dynamic condition
func GetEwtWithdrawFn(arrCond []WhereCondFn, debug bool) ([]*EwtWithdraw, error) {
	var result []*EwtWithdraw
	tx := db.Table("ewt_withdraw").
		Select("ewt_withdraw.*, ewt_setup.ewallet_type_code").
		Joins("inner join ewt_setup ON ewt_setup.id = ewt_withdraw.ewallet_type_id")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Order("id desc").Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

type AddEwtWithdrawStruct struct {
	ID              int    `gorm:"primary_key" json:"id"`
	DocNo           string `json:"doc_no"`
	MemberId        int    `json:"member_id`
	EwalletTypeId   int    `json:"ewallet_type_id"`
	EwalletTypeIdTo int    `json:"ewallet_type_id_to"`
	CurrencyCode    string `json:"currency_code"`
	Type            string `json:"type"`
	TransactionType string `json:"transaction_type"`
	// CryptoType            string    `json:"crypto_type"`
	TransDate             time.Time `json:"trans_date"`
	Markup                float64   `json:"markup"`
	NetAmount             float64   `json:"net_amount"`
	TotalOut              float64   `json:"total_out"`
	ChargesType           string    `json:"charges_type"`
	AdminFee              float64   `json:"admin_fee"`
	ConversionRate        float64   `json:"conversion_rate"`
	ConvertedTotalAmount  float64   `json:"converted_total_amount"`
	ConvertedNetAmount    float64   `json:"converted_net_amount"`
	ConvertedAdminFee     float64   `json:"converted_admin_fee"`
	ConvertedCurrencyCode string    `json:"converted_currency_code"`
	CryptoAddrTo          string    `json:"crypto_addr_to"`
	// CryptoAddrReturn      string    `json:"crypto_addr_return"`
	GasFee    float64   `json:"gas_fee"`
	GasPrice  string    `json:"gas_price"`
	TranHash  string    `json:"tran_hash"`
	Remark    string    `json:"remark"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int       `json:"created_by"`
	ExpiredAt time.Time `json:"expired_at"`
}

// AddEwtDetail add ewt_detail records`
func AddEwtWithdraw(tx *gorm.DB, saveData AddEwtWithdrawStruct) (*AddEwtWithdrawStruct, error) {
	if err := tx.Table("ewt_withdraw").Create(&saveData).Error; err != nil {
		ErrorLog("AddEwtWithdraw-failed_to_save", err.Error(), saveData)
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &saveData, nil
}

func GetEwtWithdrawDetailByDocNo(docNo string) (*EwtWithdrawDetail, error) {
	var ewt EwtWithdrawDetail

	query := db.Table("ewt_withdraw a").
		Select("a.*,b.name as status_desc,c.ewallet_type_name as ewallet_from, d.ewallet_type_name as ewallet_to").
		Joins("left join sys_general b ON a.status = b.code and b.type='general-status'").
		Joins("inner join ewt_setup c ON a.ewallet_type_id = c.id").
		Joins("inner join ewt_setup d ON a.ewallet_type_id_to = d.id")

	if docNo != "" {
		query = query.Where("a.doc_no = ?", docNo)
	}

	err := query.Order("id desc").Find(&ewt).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &ewt, nil
}

func GetEwtWithdrawFnV2(arrCond []WhereCondFn, selectColumn string, debug bool) (*EwtWithdraw, error) {
	var result EwtWithdraw
	tx := db.Table("ewt_withdraw")

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
	}
	err := tx.Find(&result).Error

	if (err != nil && err != gorm.ErrRecordNotFound) || result.ID <= 0 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

type SumWithdrawStruct struct {
	TotalWithdraw float64 `json:"total_withdraw"`
}

func GetSumTotalWithdrawFn(arrCond []WhereCondFn, debug bool) (*SumWithdrawStruct, error) {
	var result SumWithdrawStruct
	tx := db.Table("ewt_withdraw").
		Select("SUM(ewt_withdraw.total_out) AS 'total_withdraw'")

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
