package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/yapkah/go-api/pkg/e"
)

type EwtTransferSetup struct {
	PrjConfigCode         string  `gorm:"column:prj_config_code"`
	EwtTransferType       string  `gorm:"column:ewt_transfer_type"`
	CheckEwtTransferSetup int     `gorm:"column:check_ewt_transfer_setup" json:"code"`
	EwalletTypeIdFrom     int     `gorm:"column:ewallet_type_id_from" json:"ewallet_type_id_from"`
	EwalletTypeIdTo       int     `gorm:"column:transfer_type_id_to" json:"ewallet_type_id_to"`
	TransferSameMember    int     `gorm:"column:transfer_same_member"`
	TransferOtherMember   int     `gorm:"column:transfer_other_member"`
	TransferUpline        int     `gorm:"column:transfer_upline"`
	TransferSponsor       int     `gorm:"column:transfer_sponsor"`
	TransferUplineTree    int     `gorm:"column:transfer_upline_tree"`
	TransferSponsorTree   int     `gorm:"column:transfer_sponsor_tree"`
	TransferDownline      int     `gorm:"column:transfer_downline"`
	ShowWalletTo          int     `gorm:"column:show_wallet_to"`
	MemberShow            int     `gorm:"column:member_show"`
	TransferMin           float64 `gorm:"column:transfer_min"`
	TransferMax           float64 `gorm:"column:transfer_max"`
	// AvailableTransfer     float64 `gorm:"column:available_transfer"` // in %
	ProcessingFee       float64 `gorm:"column:processing_fee"`
	AdminFee            float64 `gorm:"column:admin_fee"`
	Charge              float64 `gorm:"column:charge"`
	Rate                float64 `gorm:"column:rate"`
	Tax                 float64 `gorm:"column:tax"`
	EwalletTypeCodeFrom string  `gorm:"column:ewt_type_code_from" json:"ewallet_type_code_from"`
	EwalletTypeNameFrom string  `gorm:"column:ewt_type_name_from" json:"ewallet_type_name_from"`
	EwalletTypeCodeTo   string  `gorm:"column:ewt_type_code_to" json:"ewallet_type_code_to"`
	EwalletTypeNameTo   string  `gorm:"column:ewt_type_name_to" json:"ewallet_type_name_to"`
}

type TransferSetupForm struct {
	ID                      int64   `gorm:"primary_key"`
	Type                    string  `gorm:"column:type"`
	TransferUpline          int     `gorm:"column:transfer_upline" json:"transfer_upline"`
	TransferDownline        int     `gorm:"column:transfer_downline" json:"transfer_downline"`
	TransferSameMemberType  string  `gorm:"column:transfer_same_member_type" json:"transfer_same_member_type"`
	TransferCrossMemberType string  `gorm:"column:transfer_cross_member_type" json:"transfer_cross_member_type"`
	TransferMin             float64 `gorm:"column:transfer_min" json:"transfer_min"`
	TransferMax             float64 `gorm:"column:transfer_max" json:"transfer_max"`
}

// GetEwtTransferSetupFn get ewt_transfer_setup data with dynamic condition
func GetEwtTransferSetupFn(arrCond []WhereCondFn, selectColumn string, debug bool) ([]*EwtTransferSetup, error) {
	var result []*EwtTransferSetup
	tx := db.Table("ewt_transfer_setup").
		Joins("INNER JOIN ewt_setup ewt_from ON ewt_transfer_setup.ewallet_type_id_from = ewt_from.id").
		Joins("INNER JOIN ewt_setup ewt_to ON ewt_transfer_setup.transfer_type_id_to = ewt_to.id").
		Select("ewt_transfer_setup.*, ewt_from.ewallet_type_code AS 'ewt_type_code_from', ewt_from.ewallet_type_name AS 'ewt_type_name_from', " +
			"ewt_to.ewallet_type_code AS 'ewt_type_code_to', ewt_to.ewallet_type_name AS 'ewt_type_name_to' " + selectColumn)
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

type DistinctEwtTransferFromStruct struct {
	EwalletTypeIdFrom int `gorm:"column:ewallet_type_id_from" json:"ewallet_type_id_from"`
}

// GetDistinctEwtTransferFromFn get distinct ewt_transfer_setup data with dynamic condition
func GetDistinctEwtTransferFromFn(arrCond []WhereCondFn, debug bool) ([]*DistinctEwtTransferFromStruct, error) {
	var result []*DistinctEwtTransferFromStruct
	tx := db.Table("ewt_transfer_setup").
		Joins("INNER JOIN ewt_setup ewt_from ON ewt_transfer_setup.ewallet_type_id_from = ewt_from.id AND ewt_from.status = 'A'").
		Group("ewallet_type_id_from").
		Select("ewt_transfer_setup.ewallet_type_id_from")

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

// func GetTransferSetupByFromId(from_id int, user_group string) ([]*TransferSetupStruct, error) {
// 	var wallets []*TransferSetupStruct
// 	query := db.Table("ewt_transfer_setup").
// 		Select("currency_code, b_display_code, wallet_id_from, wallet_id_to, transfer_same_member").
// 		Joins("JOIN ewt_setup ON ewt_setup.id = ewt_transfer_setup.wallet_id_from").
// 		Where("type = 'convert' AND ewt_transfer_setup.wallet_id_from = ?", from_id)

// 	if user_group != ""{
// 		query = query.Where("user_group = ?", user_group)
// 	}

// 	err := query.Scan(&wallets).Error

// 	if err != nil {
// 		if err != gorm.ErrRecordNotFound {
// 			return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: err.Error(), Data: err}
// 		}

// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}

// 	return wallets, nil
// }

// func GetTransferSetupByToId(from_id int, user_group string) ([]*TransferSetupStruct, error) {
// 	var wallets []*TransferSetupStruct
// 	query := db.Table("ewt_transfer_setup").
// 		Select("currency_code, b_display_code, wallet_id_to, transfer_same_member").
// 		Joins("JOIN ewt_setup ON ewt_setup.id = ewt_transfer_setup.wallet_id_from").
// 		Where("type = 'convert' AND ewt_transfer_setup.wallet_id_to = ?", from_id)

// 	if user_group != ""{
// 		query = query.Where("user_group = ?", user_group)
// 	}

// 	err := query.Scan(&wallets).Error

// 	if err != nil {
// 		if err != gorm.ErrRecordNotFound {
// 			return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: err.Error(), Data: err}
// 		}

// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}

// 	return wallets, nil
// }

// func GetTransferSetupByWalletId(from_id int, user_group string) (*TransferSetupForm, error) {
// 	var setup TransferSetupForm
// 	err := db.Table("ewt_transfer_setup").
// 		Where("user_group = ? AND wallet_id_from = ? AND wallet_id_to = ? AND type = 'Internal'", user_group, from_id, from_id).Scan(&setup).Error

// 	if err != nil {
// 		if err != gorm.ErrRecordNotFound {
// 			return nil, &e.CustomError{HTTPCode: http.StatusNotFound, Code: e.NOT_FOUND, Msg: err.Error(), Data: err}
// 		}

// 		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
// 	}

// 	return &setup, nil
// }

// CheckSuperUserPermission func
// func CheckSuperUserPermission(user_group_from string, user_group_to string, transfer_wallet_from int) (bool, error) {
// 	//var check bool
// 	var mem_type_list []string

// 	setup, err := GetTransferSetupByWalletId(transfer_wallet_from, user_group_from)

// 	if err != nil {
// 		return false, err
// 	}

// 	if setup.TransferSameMemberType == "1" && user_group_from == user_group_to {
// 		return true, nil
// 	}

// 	if setup.TransferCrossMemberType != "" {
// 		err := json.Unmarshal([]byte(setup.TransferCrossMemberType), &mem_type_list)

// 		if err != nil {
// 			return false, err
// 		}

// 		if helpers.Contains(mem_type_list, user_group_to) {
// 			return true, nil
// 		}
// 	}

// 	err = errors.New("no_transfer_permission")
// 	return false, err
// }
