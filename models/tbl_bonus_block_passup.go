package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// TblBonusBlockPassup struct
type TblBonusBlockPassup struct {
	BnsId       string    `json:"t_bns_id"`
	MemberId    int       `json:"member_id"`
	MemberLot   int       `json:"t_member_lot"`
	DownlineId  int       `json:"t_downline_id"`
	DownlineLot int       `json:"t_downline_lot"`
	CenterId    int       `json:"t_center_id"`
	DocNo       string    `json:"t_doc_no"`
	ItemId      int       `json:"t_item_id"`
	ILvl        int       `json:"i_lvl"`
	ILvlPaid    int       `json:"i_lvl_paid"`
	FBv         float64   `json:"f_bv"`
	FPerc       float64   `json:"f_perc"`
	FBns        float64   `json:"f_bns"`
	FBnsBurn    float64   `json:"f_bns_burn"`
	DtCreated   time.Time `json:"dt_created"`
}

type TblBonusBlockPassupResult struct {
	TBnsId        string    `json:"t_bns_id"`
	Username      string    `json:"username"`
	DownlineId    string    `json:"downline_id"`
	TDocNo        string    `json:"t_doc_no"`
	WalletTyepeId int       `json:"wallet_type_id"`
	ILvl          string    `json:"i_lvl"`
	ILvlPaid      string    `json:"i_lvl_paid"`
	FBv           float64   `json:"f_bv"`
	FPerc         float64   `json:"f_perc"`
	FBns          float64   `json:"f_bns"`
	BurnBv        float64   `json:"burn_bv"`
	BurnBns       float64   `json:"burn_bns"`
	FRate         float64   `json:"f_rate"`
	DtCreated     time.Time `json:"dt_created"`
	DtPaid        time.Time `json:"dt_paid"`
}

// TblBonusBlockPassupFn get tblq_bonus_matching data with dynamic condition
func TblBonusBlockPassupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*TblBonusBlockPassup, error) {
	var result []*TblBonusBlockPassup
	tx := db.Table("tbl_bonus_block_passup")
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

//get generation bonus by memid
func GetGenerationBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblBonusBlockPassupResult, error) {
	var (
		rwd []*TblBonusBlockPassupResult
	)

	query := db.Table("tbl_bonus_block_passup as a").
		Select("a.t_bns_id as t_bns_id,b.nick_name as username,down.nick_name as downline_id,a.i_lvl,a.i_lvl_paid,a.f_bv,a.f_perc,a.f_bns,a.f_bns_burn,a.dt_created").
		Joins("JOIN ent_member as b ON a.t_member_id = b.id").
		Joins("JOIN ent_member as down ON down.id = a.t_downline_id")

	if mem_id != 0 {
		query = query.Where("a.t_member_id = ?", mem_id)
	}

	if dateFrom != "" {
		// dateFrom = strings.Replace(dateFrom, "-", "", 2) + "0000"
		query = query.Where("a.t_bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		// dateTo = strings.Replace(dateTo, "-", "", 2) + "2359"
		query = query.Where("a.t_bns_id <= ?", dateTo)
	}

	err := query.Limit(200).Order("a.t_bns_id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}
