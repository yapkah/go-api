package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// LaligaStake struct
type LaligaStake struct {
	ID            int       `gorm:"primary_key" json:"id"`
	MemberID      int       `json:"member_id" gorm:"column:member_id"`
	DocNo         string    `json:"doc_no" gorm:"column:doc_no"`
	CryptoCode    string    `json:"crypto_code" gorm:"column:crypto_code"`
	UnitPrice     float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalUnit     float64   `json:"total_unit" gorm:"column:total_unit"`
	TotalAmount   float64   `json:"total_amount" gorm:"column:total_amount"`
	BalanceUnit   float64   `json:"balance_unit" gorm:"column:balance_unit"`
	BalanceAmount float64   `json:"balance_amount" gorm:"column:balance_amount"`
	Remark        string    `json:"remark" gorm:"column:remark"`
	Status        string    `json:"status" gorm:"column:status"`
	SigningKey    string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash     string    `json:"trans_hash" gorm:"column:trans_hash"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
	ApprovedAt    time.Time `json:"approved_at"`
	ApprovedBy    string    `json:"approved_by"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     string    `json:"updated_by"`
}

// GetLaligaStakeFn get ent_member_crypto with dynamic condition
func GetLaligaStakeFn(arrCond []WhereCondFn, debug bool) ([]*LaligaStake, error) {
	var result []*LaligaStake
	tx := db.Table("laliga_stake").
		Joins("INNER JOIN ent_member ON laliga_stake.member_id = ent_member.id").
		Select("laliga_stake.*, ent_member.nick_name").
		Order("laliga_stake.created_at DESC")

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

// AddLaligaStakeStruct struct
type AddLaligaStakeStruct struct {
	ID            int       `gorm:"primary_key" json:"id"`
	MemberID      int       `json:"member_id" gorm:"column:member_id"`
	DocNo         string    `json:"doc_no" gorm:"column:doc_no"`
	CryptoCode    string    `json:"crypto_code" gorm:"column:crypto_code"`
	UnitPrice     float64   `json:"unit_price" gorm:"column:unit_price"`
	TotalUnit     float64   `json:"total_unit" gorm:"column:total_unit"`
	TotalAmount   float64   `json:"total_amount" gorm:"column:total_amount"`
	BalanceUnit   float64   `json:"balance_unit" gorm:"column:balance_unit"`
	BalanceAmount float64   `json:"balance_amount" gorm:"column:balance_amount"`
	Remark        string    `json:"remark" gorm:"column:remark"`
	Status        string    `json:"status" gorm:"column:status"`
	SigningKey    string    `json:"signing_key" gorm:"column:signing_key"`
	TransHash     string    `json:"trans_hash" gorm:"column:trans_hash"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
}

// AddLaligaStake func
func AddLaligaStake(tx *gorm.DB, arrData AddLaligaStakeStruct) (*AddLaligaStakeStruct, error) {
	if err := tx.Table("laliga_stake").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// AvailableLaligaStakeListStruct struct
type AvailableLaligaStakeListStruct struct {
	UnitPrice        float64 `json:"unit_price" gorm:"column:unit_price"`
	TotalBalanceUnit float64 `json:"total_balance_unit" gorm:"column:total_balance_unit"`
}

// GetAvailableLaligaStakeListFn get laliga_stake with dynamic condition
func GetAvailableLaligaStakeListFn(arrCond []WhereCondFn, limit uint, debug bool) ([]*AvailableLaligaStakeListStruct, error) {
	var (
		result []*AvailableLaligaStakeListStruct
	)
	tx := db.Table("laliga_stake").
		Select("laliga_stake.unit_price, SUM(laliga_stake.balance_unit) AS 'total_balance_unit'").
		Group("laliga_stake.crypto_code, laliga_stake.unit_price").
		Where("laliga_stake.status = 'P'").
		Order("laliga_stake.unit_price DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	// Pagination and limit
	err := tx.Limit(limit).Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
