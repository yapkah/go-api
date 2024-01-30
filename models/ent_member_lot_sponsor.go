package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

// EntMemberLotSponsor struct
type EntMemberLotSponsor struct {
	ID        int     `gorm:"primary_key" json:"id"`
	MemberID  int     `gorm:"column:member_id" json:"member_id"`
	MemberLot string  `gorm:"column:member_lot" json:"member_lot"`
	ILft      float64 `gorm:"column:i_lft" json:"i_lft"`
	IRgt      float64 `gorm:"column:i_rgt" json:"i_rgt"`
	Lvl       int     `gorm:"column:i_lvl" json:"i_lvl"`
	NickName  string  `gorm:"column:nick_name" json:"nick_name"`
}

// GetEntMemberLotSponsorFn get ent_member_lot_sponsor data with dynamic condition
func GetEntMemberLotSponsorFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberLotSponsor, error) {
	var result []*EntMemberLotSponsor
	tx := db.Table("ent_member_lot_sponsor").
		Joins("JOIN ent_member ON ent_member_lot_sponsor.member_id = ent_member.id").
		Select("ent_member_lot_sponsor.id, ent_member_lot_sponsor.member_id, ent_member_lot_sponsor.member_lot, ent_member_lot_sponsor.i_lft, ent_member_lot_sponsor.i_rgt, ent_member_lot_sponsor.i_lvl, " +
			"ent_member.nick_name")

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

// TotalDownlineMemberStruct struct
type TotalDownlineMemberStruct struct {
	TotalDownline int `gorm:"column:total_downline" json:"total_downline"`
}

// GetTotalDownlineMemberFn get GetTotalDownlineMember data with dynamic condition
func GetTotalDownlineMemberFn(arrCond []WhereCondFn, debug bool) (*TotalDownlineMemberStruct, error) {
	var result TotalDownlineMemberStruct
	tx := db.Table("ent_member_lot_sponsor AS downline_lot").
		Joins("INNER JOIN ent_member_lot_sponsor AS sponsor_lot ON downline_lot.i_lft > sponsor_lot.i_lft AND downline_lot.i_rgt < sponsor_lot.i_rgt").
		Joins("INNER JOIN ent_member ON downline_lot.member_id = ent_member.id").
		Where("ent_member.status = 'A'").
		Select("COUNT(downline_lot.id) AS 'total_downline'")

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

// TotalDownlineMemberStruct struct
type TotalDownlineMemberSalesStruct struct {
	TotalSales float64 `gorm:"column:total_sales" json:"total_sales"`
}

// GetTotalDownlineMemberSalesFn get GetTotalDownlineMember data with dynamic condition
func GetTotalDownlineMemberSalesFn(arrCond []WhereCondFn, debug bool) (*TotalDownlineMemberSalesStruct, error) {
	var result TotalDownlineMemberSalesStruct
	tx := db.Table("ent_member_lot_sponsor AS downline_lot").
		Joins("INNER JOIN ent_member_lot_sponsor AS sponsor_lot ON downline_lot.i_lft > sponsor_lot.i_lft AND downline_lot.i_rgt < sponsor_lot.i_rgt").
		Joins("INNER JOIN ent_member ON downline_lot.member_id = ent_member.id AND ent_member.status = 'A'").
		Joins("INNER JOIN sls_master ON downline_lot.member_id = sls_master.member_id AND sls_master.status = 'AP'").
		Select("SUM(sls_master.total_amount) AS 'total_sales'")

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

// GetTotalNetworkSalesFn get GetTotalDownlineMember data with dynamic condition
func GetTotalNetworkSalesFn(arrCond []WhereCondFn, debug bool) (*TotalDownlineMemberSalesStruct, error) {
	var result TotalDownlineMemberSalesStruct
	tx := db.Table("ent_member_lot_sponsor AS downline_lot").
		Joins("INNER JOIN ent_member_lot_sponsor AS sponsor_lot ON downline_lot.i_lft > sponsor_lot.i_lft AND downline_lot.i_rgt < sponsor_lot.i_rgt"). // added on 20210402. ah peh request to remove own sales
		Joins("INNER JOIN ent_member ON downline_lot.member_id = ent_member.id AND ent_member.status = 'A'").
		Joins("INNER JOIN sls_master ON downline_lot.member_id = sls_master.member_id AND sls_master.status = 'AP'").
		Select("SUM(sls_master.total_amount) AS 'total_sales'")

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

//only return one record- by member
func GetEntMemberLotSponsorByMemberFn(arrCond []WhereCondFn, debug bool) (*EntMemberLotSponsor, error) {
	var result EntMemberLotSponsor
	tx := db.Table("ent_member_lot_sponsor").
		Joins("JOIN ent_member ON ent_member_lot_sponsor.member_id = ent_member.id").
		Select("ent_member_lot_sponsor.id, ent_member_lot_sponsor.member_id, ent_member_lot_sponsor.member_lot, ent_member_lot_sponsor.i_lft, ent_member_lot_sponsor.i_rgt, ent_member_lot_sponsor.i_lvl, " +
			"ent_member.nick_name")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}

// EntMemberLotSponsor struct
type EntMemberLotSponsorDetail struct {
	SponsorMemberID  int     `gorm:"column:sponsor_member_id" json:"sponsor_member_id"`
	SponsorNickName  string  `gorm:"column:sponsor_nick_name" json:"sponsor_nick_name"`
	DownlineMemberID int     `gorm:"column:downline_member_id" json:"downline_member_id"`
	DownlineNickName string  `gorm:"column:downline_nick_name" json:"downline_nick_name"`
	DownlineLot      string  `gorm:"column:downline_lot" json:"downline_lot"`
	DownlineLft      []uint8 `gorm:"column:downline_lft" json:"downline_lft"`
	DownlineRgt      []uint8 `gorm:"column:downline_rgt" json:"downline_rgt"`
	DownlineLvl      int     `gorm:"column:downline_lvl" json:"downline_lvl"`
	SponsorLot       string  `gorm:"column:sponsor_lot" json:"sponsor_lot"`
	SponsorLft       []uint8 `gorm:"column:sponsor_lft" json:"sponsor_lft"`
	SponsorRgt       []uint8 `gorm:"column:sponsor_rgt" json:"sponsor_rgt"`
	SponsorLvl       int     `gorm:"column:sponsor_lvl" json:"sponsor_lvl"`
}

// GetEntMemberLotSponsorDetailFn get ent_member_lot_sponsor detail with dynamic condition
func GetEntMemberLotSponsorDetailFn(arrCond []WhereCondFn, debug bool) ([]*EntMemberLotSponsorDetail, error) {
	var result []*EntMemberLotSponsorDetail
	tx := db.Table("ent_member AS sponsor").
		Joins("JOIN ent_member_lot_sponsor AS sponsor_lot ON sponsor.id = sponsor_lot.member_id").
		Joins("JOIN ent_member_lot_sponsor AS downline_lot ON downline_lot.i_lft >= sponsor_lot.i_lft AND downline_lot.i_rgt <= sponsor_lot.i_rgt").
		Joins("JOIN ent_member AS downline ON downline_lot.member_id = downline.id").
		Select("downline_lot.member_lot AS 'downline_lot', downline_lot.i_lft AS 'downline_lft', downline_lot.i_rgt AS 'downline_rgt', downline_lot.i_lvl AS 'downline_lvl', " +
			" sponsor_lot.member_lot AS 'sponsor_lot', sponsor_lot.i_lft AS 'sponsor_lft', sponsor_lot.i_rgt AS 'sponsor_rgt', sponsor_lot.i_lvl AS 'sponsor_lvl', " +
			"sponsor.id AS 'sponsor_member_id', sponsor.nick_name AS 'sponsor_nick_name', " +
			"downline.id AS 'downline_member_id', downline.nick_name AS 'downline_nick_name' ")

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

// TotalNetworkMemberStruct struct
type TotalNetworkMemberStruct struct {
	TotalMember float64 `gorm:"column:total_member" json:"total_member"`
}

// GetTotalNetworkMemberFn get GetTotalDownlineMember data with dynamic condition
func GetTotalNetworkMemberFn(arrCond []WhereCondFn, debug bool) (*TotalNetworkMemberStruct, error) {
	var result TotalNetworkMemberStruct
	tx := db.Table("ent_member_lot_sponsor AS downline_lot").
		Joins("INNER JOIN ent_member_lot_sponsor AS sponsor_lot ON downline_lot.i_lft > sponsor_lot.i_lft AND downline_lot.i_rgt < sponsor_lot.i_rgt").
		Joins("INNER JOIN ent_member ON downline_lot.member_id = ent_member.id AND ent_member.status = 'A'").
		Select("COUNT(*) AS 'total_member'")

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

// TotalNetworkBZZSalesStruct struct
type TotalNetworkBZZSalesStruct struct {
	TotalSales float64 `gorm:"column:total_sales" json:"total_sales"`
	TotalNodes float64 `gorm:"column:total_nodes" json:"total_nodes"`
}

// GetTotalNetworkBZZSalesFn get TotalNetworkBZZSalesStruct data with dynamic condition
func GetTotalNetworkBZZSalesFn(arrCond []WhereCondFn, debug bool) (*TotalNetworkBZZSalesStruct, error) {
	var result TotalNetworkBZZSalesStruct
	tx := db.Table("ent_member_lot_sponsor AS downline_lot").
		Joins("INNER JOIN ent_member_lot_sponsor AS sponsor_lot ON downline_lot.i_lft >= sponsor_lot.i_lft AND downline_lot.i_rgt <= sponsor_lot.i_rgt").
		Joins("INNER JOIN ent_member ON downline_lot.member_id = ent_member.id AND ent_member.status = 'A'").
		Joins("INNER JOIN sls_master ON downline_lot.member_id = sls_master.member_id AND sls_master.status = 'AP'").
		Joins("INNER JOIN sls_master_mining ON sls_master.id = sls_master_mining.sls_master_id").
		Select("SUM(sls_master.total_amount) AS 'total_sales', SUM(sls_master_mining.bzz_tib) AS 'total_nodes'")

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
