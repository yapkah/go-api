package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

type TblqBonusMiningSharingResult struct {
	TBnsId     string    `json:"t_bns_id"`
	NickName   string    `json:"nick_name"`
	DocNo      string    `json:"doc_no"`
	PersonalBv float64   `json:"personal_bv"`
	TotalBv    float64   `json:"total_bv"`
	ReleaseBv  float64   `json:"release_bv"`
	FPerc      float64   `json:"f_perc"`
	FBns       float64   `json:"f_bns"`
	DtCreated  time.Time `json:"dt_created"`
	DtPaid     time.Time `json:"dt_paid"`
}

//get bonus mining sharing by memid
func GetMiningSharingBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblqBonusMiningSharingResult, error) {
	var (
		rwd []*TblqBonusMiningSharingResult
	)

	query := db.Table("tblq_bonus_mining_sharing as a").
		Select("a.bns_id as t_bns_id ,b.nick_name, a.personal_bv , a.total_bv , a.release_bv , f_perc, a.f_bns,a.dt_created,a.dt_paid").
		Joins("JOIN ent_member as b ON a.member_id = b.id")

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
