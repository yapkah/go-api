package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

type TblqBonusMiningSharingPassupResult struct {
	TBnsId     string    `json:"t_bns_id"`
	NickName   string    `json:"nick_name"`
	DownlineId string    `json:"downline_id"`
	ILvl       string    `json:"i_lvl"`
	ILvlPaid   string    `json:"i_lvl_paid"`
	FBv        float64   `json:"f_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	BurnBv     float64   `json:"burn_bv"`
	DtCreated  time.Time `json:"dt_created"`
	DtPaid     time.Time `json:"dt_paid"`
}

//get bonus mining sharing passup by memid
func GetMiningSharingPassupBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusMiningSharingPassupResult, error) {
	var (
		rwd []*TblqBonusMiningSharingPassupResult
	)

	query := db.Table("tblq_bonus_mining_sharing_passup as a").
		Select("a.bns_id as t_bns_id ,b.nick_name,b.nick_name as username , down.nick_name as downline_id , a.i_lvl , a.i_lvl_paid , a.f_bv , a.f_perc*100 as f_perc , a.f_bns , a.burn_bv").
		Joins("JOIN ent_member as b ON a.member_id = b.id").
		Joins("JOIN ent_member as down ON down.id = a.downline_id")

	if mem_id != 0 {
		query = query.Where("a.member_id = ?", mem_id)
	}

	if dateFrom != "" {
		// dateFrom = strings.Replace(dateFrom, "-", "", 2) + "0000"
		query = query.Where("a.bns_id >= ?", dateFrom)
	}

	if dateTo != "" {
		// dateTo = strings.Replace(dateTo, "-", "", 2) + "2359"
		query = query.Where("a.bns_id <= ?", dateTo)
	}

	err := query.Order("a.bns_id desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}
