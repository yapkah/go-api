package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// AddSlsMasterMiningNodeTopupStruct struct
type AddSlsMasterMiningNodeTopupStruct struct {
	ID                    int       `gorm:"primary_key" json:"id"`
	SlsMasterMiningNodeID int       `json:"sls_master_mining_node_id" gorm:"column:sls_master_mining_node_id"`
	MemberID              int       `json:"member_id" gorm:"column:member_id"`
	PrdMasterID           int       `json:"prd_master_id" gorm:"column:prd_master_id"`
	DocNo                 string    `json:"doc_no" gorm:"column:doc_no"`
	DocDate               string    `json:"doc_date" gorm:"column:doc_date"`
	Status                string    `json:"status" gorm:"column:status"`
	Months                int       `json:"months" gorm:"column:months"`
	CreatedBy             string    `json:"created_by"`
	ApprovedBy            string    `json:"approved_by"`
	ApprovedAt            time.Time `json:"approved_at"`
}

// AddSlsMasterMiningNodeTopup func
func AddSlsMasterMiningNodeTopup(tx *gorm.DB, slsMaster AddSlsMasterMiningNodeTopupStruct) (*AddSlsMasterMiningNodeTopupStruct, error) {
	if err := tx.Table("sls_master_mining_node_topup").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}

// SlsMasterMiningNodeTopupPaginateFn struct
type SlsMasterMiningNodeTopupPaginateFn struct {
	ID           int       `json:"id" gorm:"column:id"`
	DocNo        string    `json:"doc_no" gorm:"column:doc_no"`
	SerialNumber int       `json:"serial_number" gorm:"column:serial_number"`
	PrdName      string    `json:"prd_master_name" gorm:"column:prd_master_name"`
	Months       int       `json:"months" gorm:"column:months"`
	Status       string    `json:"status" gorm:"column:status"`
	StatusCode   string    `json:"status_code" gorm:"column:status_code"`
	DocDate      string    `json:"doc_date" gorm:"column:doc_date"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetSlsMasterMiningNodeTopupPaginateFn get ent_member_crypto with dynamic condition
func GetSlsMasterMiningNodeTopupPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*SlsMasterMiningNodeTopupPaginateFn, error) {
	var (
		result                []*SlsMasterMiningNodeTopupPaginateFn
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)

	tx := db.Table("sls_master_mining_node_topup").
		Select("sls_master_mining_node_topup.id, sls_master_mining_node_topup.doc_no, sls_master_mining_node_topup.serial_number, prd_master.name as prd_master_name, sys_general.code as status_code, sys_general.name as status, sls_master_mining_node_topup.doc_date, sls_master_mining_node_topup.months, sls_master_mining_node_topup.created_at").
		Joins("INNER JOIN prd_master on prd_master.id = sls_master_mining_node_topup.prd_master_id").
		Joins("INNER JOIN sys_general on sys_general.code = sls_master_mining_node_topup.status AND sys_general.type = 'sales-status'").
		Order("sls_master_mining_node_topup.id  DESC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	arrLimitRowsSetting, _ := GetSysGeneralSetupByID("defaultlimitrow")
	limit, _ := strconv.ParseInt(arrLimitRowsSetting.SettingValue1, 10, 64)

	// Total Records
	tx.Count(&totalRecord)
	oriPage := page
	if page != 0 {
		page--
	}

	newOffset := page * limit

	// Pagination and limit
	err := tx.Limit(limit).Offset(newOffset).Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	perPage = limit

	totalCurrentPageItems = int64(len(result))

	arrPaginateData = SQLPaginateStdReturn{
		CurrentPage:           oriPage,
		PerPage:               perPage,
		TotalCurrentPageItems: totalCurrentPageItems,
		TotalPage:             totalPage,
		TotalPageItems:        totalRecord,
	}
	return arrPaginateData, result, nil
}

// SlsMasterMiningNodeTopupStruct struct
type SlsMasterMiningNodeTopupStruct struct {
	ID                    int       `gorm:"primary_key" json:"id"`
	SlsMasterMiningNodeID int       `json:"sls_master_mining_node_id" gorm:"column:sls_master_mining_node_id"`
	MemberID              int       `json:"member_id" gorm:"column:member_id"`
	PrdMasterID           int       `json:"prd_master_id" gorm:"column:prd_master_id"`
	DocNo                 string    `json:"doc_no" gorm:"column:doc_no"`
	DocDate               string    `json:"doc_date" gorm:"column:doc_date"`
	Status                string    `json:"status" gorm:"column:status"`
	Months                int       `json:"months" gorm:"column:months"`
	CreatedBy             string    `json:"created_by"`
	ApprovedBy            string    `json:"approved_by"`
	ApprovedAt            time.Time `json:"approved_at"`
}

// GetSlsMasterMiningNodeTopupFn
func GetSlsMasterMiningNodeTopupFn(arrCond []WhereCondFn, debug bool) ([]*SlsMasterMiningNodeTopupStruct, error) {
	var result []*SlsMasterMiningNodeTopupStruct
	tx := db.Table("sls_master_mining_node_topup")

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
