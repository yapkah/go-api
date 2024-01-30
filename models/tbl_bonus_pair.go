package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// TblBonusPair struct
type TblBonusPair struct {
	TBnsFr      string    `json:"t_bns_fr"`
	TBnsTo      string    `json:"t_bns_to"`
	TMemberId   int       `json:"t_member_id"`
	TMemberLot  int       `json:"t_member_lot"`
	TPairGroup  int       `json:"t_pair_group"`
	TRankEff    string    `json:"t_rank_eff"`
	TStatus     string    `json:"t_status"`
	ILft        int       `json:"i_lft"`
	IRgt        int       `json:"i_rgt"`
	ILvl        int       `json:"i_lvl"`
	FBBf1       float64   `json:"f_b_bf_1"`
	FBBf2       float64   `json:"f_b_bf_2"`
	FBBf3       float64   `json:"f_b_bf_3"`
	FBBf4       float64   `json:"f_b_bf_4"`
	FBBf5       float64   `json:"f_b_bf_5"`
	FBBf6       float64   `json:"f_b_bf_6"`
	FBCur1      float64   `json:"f_b_cur_1"`
	FBCur2      float64   `json:"f_b_cur_2"`
	FBCur3      float64   `json:"f_b_cur_3"`
	FBCur4      float64   `json:"f_b_cur_4"`
	FBCur5      float64   `json:"f_b_cur_5"`
	FBCur6      float64   `json:"f_b_cur_6"`
	FBTot1      float64   `json:"f_b_tot_1"`
	FBTot2      float64   `json:"f_b_tot_2"`
	FBTot3      float64   `json:"f_b_tot_3"`
	FBTot4      float64   `json:"f_b_tot_4"`
	FBTot5      float64   `json:"f_b_tot_5"`
	FBTot6      float64   `json:"f_b_tot_6"`
	FBAcc       float64   `json:"f_b_acc"`
	FBFlush1    float64   `json:"f_b_flush_1"`
	FBFlush2    float64   `json:"f_b_flush_2"`
	FBFlush3    float64   `json:"f_b_flush_3"`
	FBFlush4    float64   `json:"f_b_flush_4"`
	FBFlush5    float64   `json:"f_b_flush_5"`
	FBFlush6    float64   `json:"f_b_flush_6"`
	FBCF1       float64   `json:"f_b_cf_1"`
	FBCF2       float64   `json:"f_b_cf_2"`
	FBCF3       float64   `json:"f_b_cf_3"`
	FBCF4       float64   `json:"f_b_cf_4"`
	FBCF5       float64   `json:"f_b_cf_5"`
	FBCF6       float64   `json:"f_b_cf_6"`
	FBMatch     float64   `json:"f_b_match"`
	FBMatch2    float64   `json:"f_b_match_2"`
	FBMatch3    float64   `json:"f_b_match_3"`
	IBPair      int       `json:"i_b_pair"`
	IBPair2     int       `json:"i_b_pair_2"`
	IBPair3     int       `json:"i_b_pair_3"`
	IBPairAcc   int       `json:"i_b_pair_acc"`
	IBPairAcc2  int       `json:"i_b_pair_acc_2"`
	IBPairAcc3  int       `json:"i_b_pair_acc_3"`
	FBPairPerc  float64   `json:"f_b_pair_perc"`
	FBPairPerc2 float64   `json:"f_b_pair_perc_2"`
	FBPairPerc3 float64   `json:"f_b_pair_perc_3"`
	FBnsPair    float64   `json:"f_bns_pair"`
	FBnsPair2   float64   `json:"f_bns_pair_2"`
	FBnsPair3   float64   `json:"f_bns_pair_3"`
	FBns        float64   `json:"f_bns"`
	BCap        int       `json:"b_cap"`
	BLatest     int       `json:"b_latest"`
	DtCreated   time.Time `json:"dt_created"`
}

type TblBonusPairResult struct {
	TBnsFr        string    `json:"t_bns_fr"`
	TBnsTo        string    `json:"t_bns_to"`
	TMemberId     int       `json:"t_member_id"`
	Username      string    `json:"username"`
	TMemberLot    int       `json:"t_member_lot"`
	TPairGroup    int       `json:"t_pair_group"`
	TRankEff      string    `json:"t_rank_eff"`
	TStatus       string    `json:"t_status"`
	ILft          int       `json:"i_lft"`
	IRgt          int       `json:"i_rgt"`
	ILvl          int       `json:"i_lvl"`
	FBBf1         float64   `json:"f_b_bf_1"`
	FBBf2         float64   `json:"f_b_bf_2"`
	FBBf3         float64   `json:"f_b_bf_3"`
	FBBf4         float64   `json:"f_b_bf_4"`
	FBBf5         float64   `json:"f_b_bf_5"`
	FBBf6         float64   `json:"f_b_bf_6"`
	FBCur1        float64   `json:"f_b_cur_1"`
	FBCur2        float64   `json:"f_b_cur_2"`
	FBCur3        float64   `json:"f_b_cur_3"`
	FBCur4        float64   `json:"f_b_cur_4"`
	FBCur5        float64   `json:"f_b_cur_5"`
	FBCur6        float64   `json:"f_b_cur_6"`
	FBTot1        float64   `json:"f_b_tot_1"`
	FBTot2        float64   `json:"f_b_tot_2"`
	FBTot3        float64   `json:"f_b_tot_3"`
	FBTot4        float64   `json:"f_b_tot_4"`
	FBTot5        float64   `json:"f_b_tot_5"`
	FBTot6        float64   `json:"f_b_tot_6"`
	FBAcc         float64   `json:"f_b_acc"`
	FBFlush1      float64   `json:"f_b_flush_1"`
	FBFlush2      float64   `json:"f_b_flush_2"`
	FBFlush3      float64   `json:"f_b_flush_3"`
	FBFlush4      float64   `json:"f_b_flush_4"`
	FBFlush5      float64   `json:"f_b_flush_5"`
	FBFlush6      float64   `json:"f_b_flush_6"`
	Fbcf1         float64   `json:"f_b_cf_1"`
	Fbcf2         float64   `json:"f_b_cf_2"`
	Fbcf3         float64   `json:"f_b_cf_3"`
	Fbcf4         float64   `json:"f_b_cf_4"`
	Fbcf5         float64   `json:"f_b_cf_5"`
	Fbcf6         float64   `json:"f_b_cf_6"`
	FBMatch       float64   `json:"f_b_match"`
	FBMatch2      float64   `json:"f_b_match_2"`
	FBMatch3      float64   `json:"f_b_match_3"`
	IBPair        int       `json:"i_b_pair"`
	IBPair2       int       `json:"i_b_pair_2"`
	IBPair3       int       `json:"i_b_pair_3"`
	IBPairAcc     int       `json:"i_b_pair_acc"`
	IBPairAcc2    int       `json:"i_b_pair_acc_2"`
	IBPairAcc3    int       `json:"i_b_pair_acc_3"`
	FBPairPerc    float64   `json:"f_b_pair_perc"`
	FBPairPerc2   float64   `json:"f_b_pair_perc_2"`
	FBPairPerc3   float64   `json:"f_b_pair_perc_3"`
	FBnsPair      float64   `json:"f_bns_pair"`
	FBnsPair2     float64   `json:"f_bns_pair_2"`
	FBnsPair3     float64   `json:"f_bns_pair_3"`
	FBns          float64   `json:"f_bns"`
	BCap          int       `json:"b_cap"`
	BLatest       int       `json:"b_latest"`
	DtCreated     time.Time `json:"dt_created"`
	TotalPerLeg1  float64   `json:"total_per_leg1"`
	TotalPerLeg2  float64   `json:"total_per_leg2"`
	TotalPerLeg3  float64   `json:"total_per_leg2"`
	TotalPerLeg4  float64   `json:"total_per_leg4"`
	TotalPerLeg12 float64   `json:"total_per_leg12"`
	TotalPerLeg34 float64   `json:"total_per_leg34"`
	Leg1          float64   `json:"leg1"`
	Leg2          float64   `json:"leg2"`
	Leg3          float64   `json:"leg3"`
	Leg4          float64   `json:"leg4"`
	Leg12         float64   `json:"leg12"`
	Leg34         float64   `json:"leg34"`
}

//get Pair bonus by memid
func GetPairBonusByMemberId(mem_id int, dateFrom string, dateTo string) ([]*TblBonusPairResult, error) {
	var (
		rwd []*TblBonusPairResult
	)

	query := db.Table("tbl_bonus_pair as a").
		Select("a.*,b.nick_name as username,a.dt_created,a.f_b_cur_1 as total_per_leg1,a.f_b_cur_2 as total_per_leg2,a.f_b_cur_3 as total_per_leg3,a.f_b_cur_4 as total_per_leg4,a.f_b_cur_5 as total_per_leg12,a.f_b_cur_6 as total_per_leg34,a.f_b_cf_1 as leg1,a.f_b_cf_2 as leg2,a.f_b_cf_3 as leg3,a.f_b_cf_4 as leg4,a.f_b_cf_5 as leg12,a.f_b_cf_6 as leg34").
		Joins("JOIN ent_member as b ON a.t_member_id = b.id")

	if mem_id != 0 {
		query = query.Where("a.t_member_id = ?", mem_id)
	}

	if dateFrom != "" {
		query = query.Where("a.t_bns_fr >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("a.t_bns_fr <= ?", dateTo)
	}

	err := query.Order("a.t_bns_fr desc").Find(&rwd).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return rwd, nil
}

// GetTblBonusPairFn get tbl_bonus_pair data with dynamic condition
func GetTblBonusPairFn(arrCond []WhereCondFn, debug bool) ([]*TblBonusPair, error) {
	var result []*TblBonusPair
	tx := db.Table("tbl_bonus_pair")
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
