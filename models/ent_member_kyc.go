package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AddEntMemberKYCStruct struct
type AddEntMemberKYCStruct struct {
	ID            int    `gorm:"primary_key" json:"id"`
	MemberID      int    `gorm:"column:member_id" json:"member_id"`
	FullName      string `gorm:"column:full_name" json:"full_name"`
	IdentityNo    string `gorm:"column:identity_no" json:"identity_no"`
	CountryID     int    `gorm:"column:country_id" json:"country_id"`
	WalletAddress string `gorm:"column:wallet_address" json:"wallet_address"`
	Email         string `gorm:"column:email" json:"email"`
	FileName1     string `gorm:"column:file_name_1" json:"file_name_1"`
	FileURL1      string `gorm:"column:file_url_1" json:"file_url_1"`
	FileName2     string `gorm:"column:file_name_2" json:"file_name_2"`
	FileURL2      string `gorm:"column:file_url_2" json:"file_url_2"`
	FileName3     string `gorm:"column:file_name_3" json:"file_name_3"`
	FileURL3      string `gorm:"column:file_url_3" json:"file_url_3"`
	Status        string `gorm:"column:status" json:"status"`
	CreatedBy     string `gorm:"column:created_by" json:"created_by"`
}

// AddEntMemberKYC add member
func AddEntMemberKYC(tx *gorm.DB, arrData AddEntMemberKYCStruct) (*AddEntMemberKYCStruct, error) {
	if err := tx.Table("ent_member_kyc").Create(&arrData).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &arrData, nil
}

// EntMemberKyc struct
type EntMemberKyc struct {
	ID            int    `gorm:"primary_key" json:"id"`
	MemberID      int    `gorm:"column:member_id" json:"member_id"`
	FullName      string `gorm:"column:full_name" json:"full_name"`
	IdentityNo    string `gorm:"column:identity_no" json:"identity_no"`
	CountryID     int    `gorm:"column:country_id" json:"country_id"`
	WalletAddress string `gorm:"column:wallet_address" json:"wallet_address"`
	Email         string `gorm:"column:email" json:"email"`
	// MobilePrefix string    `gorm:"column:mobile_prefix" json:"mobile_prefix"`
	// MobileNo     string    `gorm:"column:mobile_no" json:"mobile_no"`
	FileName1   string    `gorm:"column:file_name_1" json:"file_name_1"`
	FileURL1    string    `gorm:"column:file_url_1" json:"file_url_1"`
	FileName2   string    `gorm:"column:file_name_2" json:"file_name_2"`
	FileURL2    string    `gorm:"column:file_url_2" json:"file_url_2"`
	FileName3   string    `gorm:"column:file_name_3" json:"file_name_3"`
	FileURL3    string    `gorm:"column:file_url_3" json:"file_url_3"`
	Remark      string    `gorm:"column:remark" json:"remark"`
	Status      string    `gorm:"column:status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy   string    `gorm:"column:created_by" json:"created_by"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy   string    `gorm:"column:updated_by" json:"updated_by"`
	ApprovedAt  time.Time `gorm:"column:approved_at" json:"approved_at"`
	ApprovedBy  string    `gorm:"column:approved_by" json:"approved_by"`
	CancelledAt time.Time `gorm:"column:cancelled_at" json:"cancelled_at"`
	CancelledBy string    `gorm:"column:cancelled_by" json:"cancelled_by"`
	RejectedAt  time.Time `gorm:"column:rejected_at" json:"rejected_at"`
	RejectedBy  string    `gorm:"column:rejected_by" json:"rejected_by"`
}

// GetEntMemberKycFn get ent_member_kyc with dynamic condition
func GetEntMemberKycFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberKyc, error) {
	var result []*EntMemberKyc
	tx := db.Table("ent_member_kyc").
		Order("ent_member_kyc.id DESC")

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
