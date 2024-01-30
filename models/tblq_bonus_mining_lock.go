package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

type BonusMiningLock struct {
	BnsID        time.Time `json:"bns_id"`
	MemberID     int       `json:"member_id"`
	WalletTypeID int       `json:"wallet_type_id"`
	MiningType   string    `json:"mining_type"`
	FBns         float64   `json:"f_bns"`
}

// GetBonusMiningLock fun
func GetBonusMiningLock(memberID int, cryptoType string) ([]*BonusMiningLock, error) {
	var rwd []*BonusMiningLock

	err := db.Table("tblq_bonus_mining_lock").
		Where("tblq_bonus_mining_lock.member_id = ?", memberID).
		Where("tblq_bonus_mining_lock.mining_type = ?", cryptoType).
		Where("tblq_bonus_mining_lock.dt_paid IS NULL").
		Order("tblq_bonus_mining_lock.bns_id desc").
		Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}
