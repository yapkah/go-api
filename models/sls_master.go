package models

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// SlsMaster struct
type SlsMaster struct {
	ID           int     `gorm:"primary_key" json:"id"`
	CountryID    int     `json:"country_id" gorm:"column:country_id"`
	CompanyID    int     `json:"company_id" gorm:"column:company_id"`
	MemberID     int     `json:"member_id" gorm:"column:member_id"`
	SponsorID    int     `json:"sponsor_id" gorm:"column:sponsor_id"`
	Status       string  `json:"status" gorm:"column:status"`
	Action       string  `json:"action" gorm:"column:action"`
	TotUnit      float64 `json:"total_unit" gorm:"column:total_unit"`
	PriceRate    float64 `json:"price_rate" gorm:"column:price_rate"`
	PrdMasterID  int     `json:"prd_master_id" gorm:"column:prd_master_id"`
	BatchNo      string  `json:"batch_no" gorm:"column:batch_no"`
	DocType      string  `json:"doc_type" gorm:"column:doc_type"`
	DocNo        string  `json:"doc_no" gorm:"column:doc_no"`
	DocDate      string  `json:"doc_date" gorm:"column:doc_date"`
	BnsBatch     string  `json:"bns_batch" gorm:"column:bns_batch"`
	BnsAction    string  `json:"bns_action" gorm:"column:bns_action"`
	TotalAmount  float64 `json:"total_amount" gorm:"column:total_amount"`
	SubTotal     float64 `json:"sub_total" gorm:"column:sub_total"`
	TotalPv      float64 `json:"total_pv" gorm:"column:total_pv"`
	TotalBv      float64 `json:"total_bv" gorm:"column:total_bv"`
	TotalSv      float64 `json:"total_sv" gorm:"column:total_sv"`
	TotalNv      float64 `json:"total_nv" gorm:"column:total_nv"`
	TotalTv      float64 `json:"total_tv" gorm:"column:total_tv"`
	CurrencyCode string  `json:"currency_code" gorm:"column:currency_code"`
	TokenRate    float64 `json:"token_rate" gorm:"column:token_rate"`
	ExchangeRate float64 `json:"exchange_rate" gorm:"column:exchange_rate"`
	MachineType  string  `json:"machine_type" gorm:"column:machine_type"`
	// Leverage    string    `json:"leverage" gorm:"column:leverage"`
	GrpType      string    `json:"grp_type" gorm:"column:grp_type"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    string    `json:"updated_by"`
	ApprovableAt time.Time `json:"payable_at"`
	ExpiredAt    time.Time `json:"expired_at"`
}

type SlsMasterPagination struct {
	ID            int     `gorm:"primary_key" json:"id"`
	CountryID     int     `json:"country_id" gorm:"column:country_id"`
	CompanyID     int     `json:"company_id" gorm:"column:company_id"`
	MemberID      int     `json:"member_id" gorm:"column:member_id"`
	SponsorID     int     `json:"sponsor_id" gorm:"column:sponsor_id"`
	Status        string  `json:"status" gorm:"column:status"`
	StatusDesc    string  `json:"status_desc" gorm:"column:status_desc"`
	Action        string  `json:"action" gorm:"column:action"`
	TotUnit       float64 `json:"total_unit" gorm:"column:total_unit"`
	PrdMasterID   int     `json:"prd_master_id" gorm:"column:prd_master_id"`
	PrdMasterCode string  `json:"prd_master_code" gorm:"column:prd_master_code"`
	PrdMasterName string  `json:"prd_master_name" gorm:"column:prd_master_name"`
	RebatePerc    float64 `json:"rebate_perc" gorm:"column:rebate_perc"`
	TopupSetting  string  `json:"topup_setting" gorm:"column:topup_setting"`
	RefundSetting string  `json:"refund_setting" gorm:"column:refund_setting"`
	DocType       string  `json:"doc_type" gorm:"column:doc_type"`
	DocNo         string  `json:"doc_no" gorm:"column:doc_no"`
	BatchNo       string  `json:"batch_no" gorm:"column:batch_no"`
	RefNo         string  `json:"ref_no" gorm:"column:ref_no"`
	DocDate       string  `json:"doc_date" gorm:"column:doc_date"`
	BnsBatch      string  `json:"bns_batch" gorm:"column:bns_batch"`
	TotalAmount   float64 `json:"total_amount" gorm:"column:total_amount"`
	SubTotal      float64 `json:"sub_total" gorm:"column:sub_total"`
	TotalPv       float64 `json:"total_pv" gorm:"column:total_pv"`
	TotalBv       float64 `json:"total_bv" gorm:"column:total_bv"`
	TotalSv       float64 `json:"total_sv" gorm:"column:total_sv"`
	TotalNv       float64 `json:"total_nv" gorm:"column:total_nv"`
	TotalTv       float64 `json:"total_tv" gorm:"column:total_tv"`
	AirdropRate   float64 `json:"airdrop_rate" gorm:"column:airdrop_rate"`
	// Leverage    string    `json:"leverage" gorm:"column:leverage"`
	// GrpType string `json:"grp_type" gorm:"column:grp_type"`
	// BnsAction    string    `json:"bns_action" gorm:"column:bns_action"`
	CreatedAt           time.Time                      `json:"created_at"`
	CreatedBy           string                         `json:"created_by"`
	UpdatedAt           time.Time                      `json:"updated_at"`
	UpdatedBy           string                         `json:"updated_by"`
	ApprovableAt        time.Time                      `json:"payable_at"`
	ExpiredAt           time.Time                      `json:"expired_at"`
	Payment             []MemberSalesListPaymentStruct `json:"payment"`
	MachineType         string                         `json:"machine_type"`
	TokenRate           float64                        `json:"token_rate"`
	ExchangeRate        float64                        `json:"exchange_rate"`
	DecimalPoint        int                            `json:"decimal_point"`
	CurrencyCode        string                         `json:"currency_code"`
	ProductCurrencyCode string                         `json:"product_currency_code"`
}

type SlsMasterByDocNo struct {
	ID           int       `gorm:"primary_key" json:"id"`
	CountryID    int       `json:"country_id" gorm:"column:country_id"`
	CompanyID    int       `json:"company_id" gorm:"column:company_id"`
	MemberID     int       `json:"member_id" gorm:"column:member_id"`
	SponsorID    int       `json:"sponsor_id" gorm:"column:sponsor_id"`
	Status       string    `json:"status" gorm:"column:status"`
	StatusDesc   string    `json:"status_desc" gorm:"column:status_desc"`
	Action       string    `json:"action" gorm:"column:action"`
	TotUnit      float64   `json:"total_unit" gorm:"column:total_unit"`
	PriceRate    float64   `json:"price_rate" gorm:"column:price_rate"`
	PrdMasterID  int       `json:"prd_master_id" gorm:"column:prd_master_id"`
	DocType      string    `json:"doc_type" gorm:"column:doc_type"`
	DocNo        string    `json:"doc_no" gorm:"column:doc_no"`
	DocDate      string    `json:"doc_date" gorm:"column:doc_date"`
	BnsBatch     string    `json:"bns_batch" gorm:"column:bns_batch"`
	TotalAmount  float64   `json:"total_amount" gorm:"column:total_amount"`
	SubTotal     float64   `json:"sub_total" gorm:"column:sub_total"`
	TotalPv      float64   `json:"total_pv" gorm:"column:total_pv"`
	TotalBv      float64   `json:"total_bv" gorm:"column:total_bv"`
	TotalSv      float64   `json:"total_sv" gorm:"column:total_sv"`
	TotalNv      float64   `json:"total_nv" gorm:"column:total_nv"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    string    `json:"updated_by"`
	ApprovableAt time.Time `json:"payable_at"`
}

type MemberSalesListPaymentStruct struct {
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currency_code"`
	DecimalPoint int     `json:"decimal_point"`
}

type GetTotalSalesAmountParams struct {
	MemID     int
	SponsorID int
	PrdCode   string
	PrdGroup  string
}

type TotalSalesAmount struct {
	TotalAmount          float64 `json:"total_amount"`
	ConvertedTotalAmount float64 `json:"converted_total_amount"`
	TotalBv              float64 `json:"total_bv"`
	TotalNv              float64 `json:"total_nv"`
	CurrencyCode         string  `json:"currency_code"`
	DecimalPoint         int     `json:"decimal_point"`
	TotalUnit            float64 `json:"total_unit"`
	RefundAmount         float64 `json:"refund_amount"`
}

type MemberTotalContractAmount struct {
	TotalContractAmount float64 `json:"total_contract_amount"`
}

// GetSlsMasterFn get ent_member_crypto with dynamic condition
func GetSlsMasterFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMaster, error) {
	var result []*SlsMaster
	tx := db.Table("sls_master").
		Joins("INNER JOIN ent_member ON ent_member.id = sls_master.member_id").
		Order("sls_master.id DESC")

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

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetSlsMasterAscFn get ent_member_crypto with dynamic condition
func GetSlsMasterAscFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*SlsMaster, error) {
	var result []*SlsMaster
	tx := db.Table("sls_master").
		Joins("INNER JOIN ent_member ON ent_member.id = sls_master.member_id").
		Order("sls_master.id ASC")

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

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// GetSlsMasterPaginateFn get ent_member_crypto with dynamic condition
func GetSlsMasterPaginateFn(arrCond []WhereCondFn, page int64, debug bool) (SQLPaginateStdReturn, []*SlsMasterPagination, error) {
	var (
		result                []*SlsMasterPagination
		perPage               int64
		totalPage             float64
		totalRecord           int64
		totalCurrentPageItems int64
		arrPaginateData       SQLPaginateStdReturn
	)
	tx := db.Table("sls_master").
		Joins("INNER JOIN prd_master ON prd_master.id = sls_master.prd_master_id").
		Joins("INNER JOIN prd_group_type ON prd_master.prd_group = prd_group_type.code").
		// Joins("LEFT JOIN blockchain_trans ON blockchain_trans.doc_no = sls_master.doc_no").
		// Joins("LEFT JOIN ewt_setup as sub_ewt_setup ON sub_ewt_setup.id = blockchain_trans.ewallet_type_id").
		// Joins("INNER JOIN ewt_detail ON ewt_detail.doc_no = sls_master.doc_no").
		// Joins("INNER JOIN ewt_setup as main_ewt_setup  ON main_ewt_setup.id = ewt_detail.ewallet_type_id").
		Joins("INNER JOIN sys_general ON sls_master.status = sys_general.code and sys_general.type='sales-status'").
		Select("sls_master.*, prd_master.code as prd_master_code, prd_master.name as prd_master_name, prd_master.rebate_perc, prd_group_type.refund_setting, sys_general.name as status_desc, prd_group_type.decimal_point, prd_group_type.topup_setting, prd_group_type.currency_code as product_currency_code").
		Order("sls_master.id DESC")

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

	for _, row := range result {
		var payment []MemberSalesListPaymentStruct

		// find payment in ewt_detail table
		batchNoCondition := ""
		if row.BatchNo != "" {
			batchNoCondition = "or ewt_detail.doc_no = '" + row.BatchNo + "'"
		}

		tx := db.Table("ewt_detail").
			Joins("INNER JOIN ewt_setup ON ewt_setup.id = ewt_detail.ewallet_type_id").
			Select("ewt_detail.total_out as amount, ewt_setup.currency_code as currency_code, ewt_setup.decimal_point as decimal_point").
			Where("ewt_detail.doc_no = ? "+batchNoCondition, row.DocNo).
			Where("ewt_detail.total_out > 0")

		err := tx.Find(&payment).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
		}

		row.Payment = append(row.Payment, payment...)

		// find payment in blockchain_trans table
		// batchNoCondition := ""
		// if row.BatchNo != "" {
		// 	batchNoCondition = "or blockchain_trans.doc_no = '" + row.BatchNo + "'"
		// }

		// bcTx := db.Table("blockchain_trans").
		// 	Joins("INNER JOIN ewt_setup ON ewt_setup.id = blockchain_trans.ewallet_type_id").
		// 	Select("blockchain_trans.total_out as amount, ewt_setup.currency_code as currency_code, ewt_setup.decimal_point as decimal_point").
		// 	Where("blockchain_trans.doc_no = ? "+batchNoCondition, row.DocNo).
		// 	Where("blockchain_trans.transaction_type != ?", "STAKING-APPROVED").
		// 	Where("blockchain_trans.log_only != ?", 1)

		// bcErr := bcTx.Find(&payment).Error
		// if bcErr != nil && bcErr != gorm.ErrRecordNotFound {
		// 	return arrPaginateData, nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: bcErr.Error(), Data: err}
		// }

		// row.Payment = append(row.Payment, payment...)
	}

	// return ewt, totalRecord, totalCurrentPageItems, perPage, totalPage, nil
	arrPaginateData = SQLPaginateStdReturn{
		CurrentPage:           oriPage,
		PerPage:               perPage,
		TotalCurrentPageItems: totalCurrentPageItems,
		TotalPage:             totalPage,
		TotalPageItems:        totalRecord,
	}
	return arrPaginateData, result, nil
}

// AddSlsMasterStruct struct
type AddSlsMasterStruct struct {
	ID           int     `gorm:"primary_key" json:"id"`
	CountryID    int     `json:"country_id" gorm:"column:country_id"`
	CompanyID    int     `json:"company_id" gorm:"column:company_id"`
	MemberID     int     `json:"member_id" gorm:"column:member_id"`
	SponsorID    int     `json:"sponsor_id" gorm:"column:sponsor_id"`
	Status       string  `json:"status" gorm:"column:status"`
	Action       string  `json:"action" gorm:"column:action"`
	TotUnit      float64 `json:"total_unit" gorm:"column:total_unit"`
	PriceRate    float64 `json:"price_rate" gorm:"column:price_rate"`
	PrdMasterID  int     `json:"prd_master_id" gorm:"column:prd_master_id"`
	BatchNo      string  `json:"batch_no" gorm:"column:batch_no"`
	DocType      string  `json:"doc_type" gorm:"column:doc_type"`
	DocNo        string  `json:"doc_no" gorm:"column:doc_no"`
	DocDate      string  `json:"doc_date" gorm:"column:doc_date"`
	BnsBatch     string  `json:"bns_batch" gorm:"column:bns_batch"`
	BnsAction    string  `json:"bns_action" gorm:"column:bns_action"`
	TotalAmount  float64 `json:"total_amount" gorm:"column:total_amount"`
	SubTotal     float64 `json:"sub_total" gorm:"column:sub_total"`
	TotalPv      float64 `json:"total_pv" gorm:"column:total_pv"`
	TotalBv      float64 `json:"total_bv" gorm:"column:total_bv"`
	TotalSv      float64 `json:"total_sv" gorm:"column:total_sv"`
	TotalNv      float64 `json:"total_nv" gorm:"column:total_nv"`
	CurrencyCode string  `json:"currency_code" gorm:"column:currency_code"`
	TokenRate    float64 `json:"token_rate" gorm:"column:token_rate"`
	ExchangeRate float64 `json:"exchange_rate" gorm:"column:exchange_rate"`
	Workflow     string  `json:"workflow" gorm:"column:workflow"`
	// Leverage    string    `json:"leverage" gorm:"column:leverage"`
	GrpType      string    `json:"grp_type" gorm:"column:grp_type"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    string    `json:"updated_by"`
	ApprovableAt time.Time `json:"payable_at"`
	ApprovedAt   time.Time `json:"approved_at"`
	ApprovedBy   string    `json:"approved_by"`
	ExpiredAt    time.Time `json:"expired_at"`
}

// AddSlsMaster func
func AddSlsMaster(tx *gorm.DB, slsMaster AddSlsMasterStruct) (*AddSlsMasterStruct, error) {
	if err := tx.Table("sls_master").Create(&slsMaster).Error; err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &slsMaster, nil
}

func GetSlsMasterByDocNo(docNo string) (*SlsMasterByDocNo, error) {
	var sls SlsMasterByDocNo

	query := db.Table("sls_master a").
		Select("a.*,b.name as status_desc").
		Joins("left join sys_general b ON a.status = b.code and b.type='sales-status'")

	if docNo != "" {
		query = query.Where("a.doc_no = ?", docNo)
	}

	err := query.Order("id desc").Find(&sls).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sls, nil
}

func GetTotalSalesAmount(arrCond []WhereCondFn, debug bool) (*TotalSalesAmount, error) {
	var sls TotalSalesAmount

	tx := db.Table("sls_master").
		Select("SUM(sls_master.total_unit) as total_unit, SUM(sls_master.total_amount) as total_amount, SUM(sls_master.total_amount * sls_master.token_rate) as converted_total_amount, SUM(sls_master.total_bv) as total_bv, SUM(sls_master.total_nv) as total_nv, prd_group_type.currency_code, prd_group_type.decimal_point, SUM(IFNULL(sls_master_refund.request_amount,0)) as refund_amount").
		Joins("inner join prd_master ON sls_master.prd_master_id = prd_master.id").
		Joins("inner join prd_group_type ON prd_master.prd_group = prd_group_type.code").
		Joins("inner join ent_member_tree_sponsor ON ent_member_tree_sponsor.member_id = sls_master.member_id").
		Joins("inner join ent_member ON ent_member_tree_sponsor.member_id = ent_member.id").
		Joins("left join sls_master_bot_setting ON sls_master_bot_setting.sls_master_id = sls_master.id").
		Joins("left join (SELECT SUM(request_amount) as request_amount, sls_master_id from sls_master_refund GROUP by member_id, sls_master_id) as sls_master_refund ON sls_master_refund.sls_master_id = sls_master.id").
		Where("sls_master.status = ?", "AP")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	tx = tx.Group("prd_group_type.currency_code,prd_group_type.decimal_point")

	err := tx.Find(&sls).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sls, nil
}

func GetTotalSalesAmountV2(arrCond []WhereCondFn, debug bool) (*TotalSalesAmount, error) {
	var sls TotalSalesAmount

	tx := db.Table("sls_master").
		Select("SUM(sls_master.total_unit) as total_unit, SUM(sls_master.total_amount) as total_amount, SUM(sls_master.total_amount * sls_master.token_rate) as converted_total_amount, SUM(sls_master.total_bv) as total_bv, SUM(sls_master.total_nv) as total_nv, prd_group_type.currency_code, prd_group_type.decimal_point, SUM(IFNULL(sls_master_refund.request_amount,0)) as refund_amount").
		Joins("inner join prd_master ON sls_master.prd_master_id = prd_master.id").
		Joins("inner join prd_group_type ON prd_master.prd_group = prd_group_type.code").
		Joins("inner join ent_member_tree_sponsor ON ent_member_tree_sponsor.member_id = sls_master.member_id").
		Joins("inner join ent_member ON ent_member_tree_sponsor.member_id = ent_member.id").
		Joins("left join sls_master_bot_setting ON sls_master_bot_setting.sls_master_id = sls_master.id").
		Joins("left join (SELECT SUM(request_amount) as request_amount, sls_master_id from sls_master_refund GROUP by member_id, sls_master_id) as sls_master_refund ON sls_master_refund.sls_master_id = sls_master.id")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}

	tx = tx.Group("prd_group_type.currency_code,prd_group_type.decimal_point")

	err := tx.Find(&sls).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sls, nil
}

// SlsMasterDetailsByDocNo struct
type SlsMasterDetailsByDocNo struct {
	ID                   int       `gorm:"primary_key" json:"id"`
	CountryID            int       `json:"country_id" gorm:"column:country_id"`
	CompanyID            int       `json:"company_id" gorm:"column:company_id"`
	MemberID             int       `json:"member_id" gorm:"column:member_id"`
	SponsorID            int       `json:"sponsor_id" gorm:"column:sponsor_id"`
	Status               string    `json:"status" gorm:"column:status"`
	StatusDesc           string    `json:"status_desc" gorm:"column:status_desc"`
	Action               string    `json:"action" gorm:"column:action"`
	TotUnit              float64   `json:"total_unit" gorm:"column:total_unit"`
	PriceRate            float64   `json:"price_rate" gorm:"column:price_rate"`
	PrdMasterID          int       `json:"prd_master_id" gorm:"column:prd_master_id"`
	PrdCode              string    `json:"prd_code" gorm:"column:prd_code"`
	Leverage             float64   `json:"prd_leverage" gorm:"column:prd_leverage"`
	IncomeCapSetting     string    `json:"income_cap_setting" gorm:"column:income_cap_setting"`
	DocType              string    `json:"doc_type" gorm:"column:doc_type"`
	DocNo                string    `json:"doc_no" gorm:"column:doc_no"`
	DocDate              string    `json:"doc_date" gorm:"column:doc_date"`
	BnsBatch             string    `json:"bns_batch" gorm:"column:bns_batch"`
	TotalAmount          float64   `json:"total_amount" gorm:"column:total_amount"`
	SubTotal             float64   `json:"sub_total" gorm:"column:sub_total"`
	TotalPv              float64   `json:"total_pv" gorm:"column:total_pv"`
	TotalBv              float64   `json:"total_bv" gorm:"column:total_bv"`
	TotalSv              float64   `json:"total_sv" gorm:"column:total_sv"`
	TotalNv              float64   `json:"total_nv" gorm:"column:total_nv"`
	TotalTv              float64   `json:"total_tv" gorm:"column:total_tv"`
	NftSeriesCode        string    `json:"nft_series_code" gorm:"column:nft_series_code"`
	TotalAirdrop         float64   `json:"total_airdrop" gorm:"column:total_airdrop"`
	TotalAirdropNft      float64   `json:"total_airdrop_nft" gorm:"column:total_airdrop_nft"`
	AirdropEwalletTypeID int       `json:"airdrop_ewallet_type_id" gorm:"column:airdrop_ewallet_type_id"`
	ExchangeRate         float64   `json:"exchange_rate" gorm:"column:exchange_rate"`
	TokenRate            float64   `json:"token_rate" gorm:"column:token_rate"`
	CreatedAt            time.Time `json:"created_at"`
	CreatedBy            string    `json:"created_by"`
	UpdatedAt            time.Time `json:"updated_at"`
	UpdatedBy            string    `json:"updated_by"`
	ApprovableAt         time.Time `json:"payable_at"`
}

// GetSlsMasterDetailsByDocNo func
func GetSlsMasterDetailsByDocNo(docNo string) (*SlsMasterDetailsByDocNo, error) {
	var sls SlsMasterDetailsByDocNo

	query := db.Table("sls_master").
		Select("sls_master.*, prd.code as prd_code, prd.leverage as prd_leverage, prd.income_cap_setting, status.name as status_desc").
		Joins("inner join sys_general status ON sls_master.status = status.code and status.type='sales-status'").
		Joins("inner join prd_master prd ON sls_master.prd_master_id = prd.id").
		Where("sls_master.doc_no = ?", docNo)

	err := query.Order("id desc").Find(&sls).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sls, nil
}

// GetSlsMasterDetailsByID func
func GetSlsMasterDetailsByID(slsMasterID int) (*SlsMasterDetailsByDocNo, error) {
	var sls SlsMasterDetailsByDocNo

	query := db.Table("sls_master").
		Select("sls_master.*, prd.code as prd_code, prd.leverage as prd_leverage, prd.income_cap_setting, status.name as status_desc").
		Joins("inner join sys_general status ON sls_master.status = status.code and status.type='sales-status'").
		Joins("inner join prd_master prd ON sls_master.prd_master_id = prd.id").
		Where("sls_master.id = ?", slsMasterID)

	err := query.Order("id desc").Find(&sls).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sls, nil
}

// GetSlsMasterDetailsByBatchNo func
func GetSlsMasterDetailsByBatchNo(batchNo string) ([]*SlsMasterDetailsByDocNo, error) {
	var sls []*SlsMasterDetailsByDocNo

	query := db.Table("sls_master").
		Select("sls_master.*, prd.code as prd_code, prd.leverage as prd_leverage, prd.income_cap_setting, status.name as status_desc").
		Joins("inner join sys_general status ON sls_master.status = status.code and status.type='sales-status'").
		Joins("inner join prd_master prd ON sls_master.prd_master_id = prd.id")

	if batchNo != "" {
		query = query.Where("sls_master.batch_no = ?", batchNo)
	}

	err := query.Order("id desc").Find(&sls).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return sls, nil
}

func GetMemberTotalContractAmount(memId int) (*MemberTotalContractAmount, error) {
	var sls MemberTotalContractAmount

	query := db.Table("sls_master a").
		Select("SUM(total_bv)as total_contract_amount").
		Where("member_id = ?", memId).
		Where("action = ?", "CONTRACT").
		Where("status = ?", "AP")

	err := query.Find(&sls).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sls, nil
}

type GetMembersWithSalesStruct struct {
	MemID int `json:"member_id" gorm:"column:member_id"`
}

func GetMembersWithSales(debug bool) ([]*GetMembersWithSalesStruct, error) {
	var sls []*GetMembersWithSalesStruct

	query := db.Table("sls_master").
		Select("sls_master.member_id").
		Where("sls_master.action = ?", "CONTRACT").
		Where("sls_master.status IN(?,?)", "AP", "P").
		Group("sls_master.member_id")

	if debug {
		query = query.Debug()
	}

	err := query.Find(&sls).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return sls, nil
}

type MemberTotalSpentUsdtAmount struct {
	TotalAmount float64 `json:"total_amount" gorm:"total_amount"`
}

func GetMemberTotalSpentUsdtAmount(memID int) (*MemberTotalSpentUsdtAmount, error) {
	var sls MemberTotalSpentUsdtAmount

	query := db.Table("sls_master").
		Select("SUM(sls_master.total_nv) as total_amount").
		Where("sls_master.member_id = ?", memID).
		Where("action = ?", "CONTRACT").
		Where("status IN(?,?)", "AP", "P")

	err := query.Find(&sls).Error
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &sls, nil
}

type MemberTotalSales struct {
	TotalAmount float64 `json:"total_amount" gorm:"total_amount"`
}

func GetMemberTotalSalesFn(arrCond []WhereCondFn, debug bool) (*MemberTotalSales, error) {
	var result MemberTotalSales

	tx := db.Table("sls_master").
		Select("SUM(sls_master.total_amount) as total_amount")

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

// EventSponsorRankingList struct
type EventSponsorRankingList struct {
	Username       string  `json:"username" gorm:"column:username"`
	TotalSponsored float64 `json:"total_sponsored" gorm:"column:total_sponsored"`
}

// GetEventSponsorRankingListFn
func GetEventSponsorRankingListFn(quota int, arrCond []WhereCondFn, debug bool) ([]*EventSponsorRankingList, error) {
	var result []*EventSponsorRankingList

	tx := db.Table("sls_master").
		Select("`sponsor`.`nick_name` as username, SUM(`sls_master`.`total_bv`) as `total_sponsored`").
		Joins("INNER JOIN `ent_member` on `ent_member`.`id` = `sls_master`.`member_id`").
		Joins("INNER JOIN `ent_member` as `sponsor` on `sponsor`.`id` = `sls_master`.`sponsor_id`")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	tx = tx.Group("`sls_master`.`sponsor_id`").
		Order("`total_sponsored` desc").
		Limit(quota)

	if debug {
		tx = tx.Debug()
	}

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// SlsMasterByMemberLot struct
type SlsMasterByMemberLot struct {
	MemberID     int       `json:"member_id" gorm:"column:member_id"`
	Username     string    `json:"username" gorm:"column:username"`
	DocNo        string    `json:"doc_no" gorm:"column:doc_no"`
	BatchNo      string    `json:"batch_no" gorm:"column:batch_no"`
	DocDate      time.Time `json:"doc_date" gorm:"column:doc_date"`
	Status       string    `json:"status" gorm:"column:status"`
	StatusCode   string    `json:"status_code" gorm:"column:status_code"`
	Action       string    `json:"action" gorm:"column:action"`
	PrdName      string    `json:"prd_name" gorm:"column:prd_name"`
	TotalAmount  float64   `json:"total_amount" gorm:"column:total_amount"`
	CurrencyCode string    `json:"currency_code" gorm:"column:currency_code"`
	ILevel       int       `json:"i_lvl" gorm:"column:i_lvl"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	ExpiredAt    time.Time `json:"expired_at" gorm:"column:expired_at"`
}

func GetSlsMasterByMemberLot(arrCond []WhereCondFn, debug bool) ([]*SlsMasterByMemberLot, error) {
	var result []*SlsMasterByMemberLot

	tx := db.Table("sls_master").
		Select("ent_member.id as member_id, ent_member.nick_name as username, sls_master.batch_no, sls_master.doc_no, sls_master.doc_date, sls_master.status as status_code, sys_general.name as status, sls_master.action, prd_master.name as prd_name, sls_master.total_amount, sls_master.currency_code, sls_master.created_at, sls_master.expired_at, ent_member_lot_sponsor.i_lvl").
		Joins("INNER JOIN ent_member ON sls_master.member_id = ent_member.id").
		Joins("INNER JOIN prd_master ON sls_master.prd_master_id = prd_master.id").
		Joins("INNER JOIN sys_general ON sls_master.status = sys_general.code AND sys_general.type = 'general-status'").
		Joins("INNER JOIN ent_member_lot_sponsor on sls_master.member_id = ent_member_lot_sponsor.member_id")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}

	if debug {
		tx = tx.Debug()
	}

	tx = tx.Order("sls_master.created_at desc")

	err := tx.Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

func GetMemberTotalNvFn(arrCond []WhereCondFn, debug bool) (*MemberTotalSales, error) {
	var result MemberTotalSales

	tx := db.Table("sls_master").
		Select("SUM(sls_master.total_nv) as total_amount")

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

func GetMemberTotalCapFn(arrCond []WhereCondFn, debug bool) (*MemberTotalSales, error) {
	var result MemberTotalSales

	tx := db.Table("sls_master").
		Select("SUM((sls_master.total_pv + sls_master.total_sv) * sls_master.leverage) as total_amount")

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
