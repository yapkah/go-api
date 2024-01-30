package reward_service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/e"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/service/member_service"
)

type RewardSummaryStruct struct {
	RwdType     string `json:"rwd_type"`
	RwdTypeName string `json:"rwd_type_name"`
}

// reward statement & reward summary share use
type RewardStatementPostStruct struct {
	MemberID int    `json:"member_id"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
	Page     int64  `json:"page"`
	LangCode string `json:"lang_code"`
	RwdType  string `json:"rwd_type"`
}

func (r *RewardStatementPostStruct) RewardSummary() (interface{}, error) {
	type arrBonusSummaryReturnStruct struct {
		RewardType     string        `json:"rwd_type"`
		RewardTypeName string        `json:"rwd_type_name"`
		BgGradStart    string        `json:"bg_grad_start"`
		BgGradEnd      string        `json:"bg_grad_end"`
		Info           []interface{} `json:"info"`
	}

	type arrPackageSetting struct {
		Tier string  `json:"tier"`
		Min  float64 `json:"min"`
		Max  float64 `json:"max"`
	}

	var (
		decimalPoint       uint = 0
		TotalIncome        float64
		TotalIncomeStr     string
		MonthlyIncome      float64
		MonthlyIncomeStr   string
		DailyIncome        float64
		DailyIncomeStr     string
		PackageAmountMin   float64 = 0
		PackageAmountMax   float64 = 100
		PackageAmountValue float64
		// Block                int
		QualifyLevel         string
		Rank                 int
		arrRewardSummaryList interface{}
		arrBonusSummaryList  []arrBonusSummaryReturnStruct
		arrBonusInfo         []interface{}
		arrPackageTiers      []arrPackageSetting
		arrExtra             interface{}
		capBalance           float64 = 0
		strCapBalance        string
		totalCapEarning      float64 = 0
		totalCapEarningStr   string
		earnBarPerc          float64 = 0
	)
	curDateTime := base.GetCurrentTime("2006-01-02")

	if r.RwdType != "" {
		TotalReward, err := models.GetMemberTotalBns(r.MemberID, strings.ToUpper(r.RwdType), 0, 0, 0)
		if err != nil {
			base.LogErrorLog("RewardSummary - fail to get total reward in rwdType", err, r, true)
			return nil, err
		}

		//daily income
		DailyRevenue, err := models.GetMemberTotalBns(r.MemberID, strings.ToUpper(r.RwdType), 0, 1, 0)
		if err != nil {
			base.LogErrorLog("RewardSummary - fail to get daily revenue in rwdType", err, r, true)
			return nil, err
		}

		if DailyRevenue != nil {
			DailyIncome = DailyRevenue.TotalBonus
		}

		DailyIncomeStr = helpers.CutOffDecimal(DailyIncome, decimalPoint, ".", ",")

		totalRwd := helpers.CutOffDecimal(0, decimalPoint, ".", ",")

		if TotalReward != nil {
			totalRwd = helpers.CutOffDecimal(TotalReward.TotalBonus, decimalPoint, ".", ",")
		}

		switch strings.ToUpper(r.RwdType) {
		case "SPONSOR":

			//get member tier
			// db := models.GetDB() // no need set begin transaction
			packageType, err := member_service.GetMemberTier(r.MemberID)
			if err != nil {
				base.LogErrorLog("RewardSummary - GetMemberTier() return err in rwdType", r, err.Error(), true)
				return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", r.LangCode), Data: r}
			}

			imgLink := "https://media02.securelayers.cloud/medias/GTA/PACKAGE/ICON/" + packageType + ".jpg"

			// arrCond := make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: "member_id = ?", CondValue: r.MemberID},
			// )
			// arrRst, _ := models.GetTblqBonusSponsorFn(arrCond, false)

			perc := float64(0)
			// if len(arrRst) > 0 {
			// 	perc = float.Mul(arrRst[0].FPerc, 100)
			// }

			// percStr := fmt.Sprintf("%.0f", perc) + "%"

			switch strings.ToUpper(packageType) {
			case "B1":
				perc = 3
			case "B2":
				perc = 4
			case "B3":
				perc = 6
			case "B4":
				perc = 8
			case "B5":
				perc = 10

			}
			percStr := fmt.Sprintf("%.0f", perc) + "%"

			arrExtra = map[string]interface{}{
				"package_img": imgLink,
				"header":      helpers.Translate("bonus", r.LangCode),
				"value":       percStr,
			}

			// get total direct sponsor
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: r.MemberID},
				models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
			)
			totalDirectSponsorRst, _ := models.GetTotalDirectSponsorFn(arrCond, false)
			totalDirectSponsor := "0"
			if totalDirectSponsorRst.TotalDirectSponsor > 0 {
				totalDirectSponsor = helpers.CutOffDecimal(totalDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
			}

			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("total_sponosor", r.LangCode),
				"value":  totalDirectSponsor,
			})

			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("yesterday", r.LangCode),
				"value":  DailyIncomeStr,
			})
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("total_reward", r.LangCode),
				"value":  totalRwd,
			})
		case "GENERATION":
			var (
				descMin string
				descMax string
				barDesc string
				min     string
				max     string
				// perc                float64
				rank int
				// noOfDirectSponsor   int
				// directSponsorLegAmt float64
			)
			// arrCond := make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_member_id = ?", CondValue: r.MemberID},
			// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.b_latest = ?", CondValue: 1},
			// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_bns_fr <= ?", CondValue: curDateTime},
			// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_bns_to >= ?", CondValue: curDateTime},
			// )
			// result, err := models.GetTblBonusRankBlockFn(arrCond, "", false)
			// if err != nil {
			// 	base.LogErrorLog("RewardSummary - fail to get GetTblBonusRankBlockFn in rwd type", err, r, true)
			// 	return nil, err
			// }

			// if len(result) > 0 {
			// 	QualifyLevel = result[0].TRankQualify
			// }

			// tbl_bonus_rank_star_passup changed to use tbl_bonus_rank_block
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				// models.WhereCondFn{Condition: "tbl_bonus_rank_star_passup.t_member_id = ?", CondValue: r.MemberID},
				models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_member_id = ?", CondValue: r.MemberID},
				models.WhereCondFn{Condition: "tbl_bonus_rank_block.b_latest = ?", CondValue: 1},
			)
			// rst, _ := models.GetTblBonusRankStarPassupFn(arrCond, "", false)
			rst, _ := models.GetTblBonusRankBlockFn(arrCond, "", false)

			if len(rst) > 0 {
				// rank = rst[0].TRankBlock
				// noOfDirectSponsor = rst[0].FQty
				// directSponsorLegAmt = rst[0].FBvDirect

				rank = rst[0].TRankEff
				// noOfDirectSponsor = rst[0].FQty
				// directSponsorLegAmt = rst[0].FBvDirect
			}

			if rank <= 0 {
				min = helpers.CutOffDecimal(0, 0, ".", ",")
				max = helpers.CutOffDecimal(4, 0, ".", ",")
				descMax = helpers.Translate("1st_block", r.LangCode)
				// perc = float64(noOfDirectSponsor) / float64(4)
				barDesc = helpers.Translate("no_of_direct_sponsor_with_higher_amount_than_you", r.LangCode)
				QualifyLevel = strconv.Itoa(0)
			} else if rank <= 1 {
				min = helpers.CutOffDecimal(4, 0, ".", ",")
				descMin = helpers.Translate("1st_block", r.LangCode)
				max = helpers.CutOffDecimal(8, 0, ".", ",")
				descMax = helpers.Translate("2nd_block", r.LangCode)
				// perc = float64(noOfDirectSponsor) / float64(8)
				barDesc = helpers.Translate("no_of_direct_sponsor_with_higher_amount_than_you", r.LangCode)
				QualifyLevel = strconv.Itoa(1)
			} else if rank <= 2 {
				min = helpers.CutOffDecimal(8, 0, ".", ",")
				descMin = helpers.Translate("2nd_block", r.LangCode)
				max = helpers.CutOffDecimal(16, 0, ".", ",")
				descMax = helpers.Translate("3rd_block", r.LangCode)
				// perc = float64(noOfDirectSponsor) / float64(16)
				barDesc = helpers.Translate("no_of_direct_sponsor_with_higher_amount_than_you", r.LangCode)
				QualifyLevel = strconv.Itoa(2)
			} else if rank <= 3 {
				min = helpers.CutOffDecimal(0, 0, ".", ",")
				descMin = helpers.Translate("3rd_block", r.LangCode)
				max = helpers.CutOffDecimal(150000, 0, ".", ",")
				descMax = helpers.Translate("4th_block", r.LangCode)
				// perc = float64(directSponsorLegAmt) / float64(150000)
				barDesc = helpers.Translate("amount_of_direct_referral", r.LangCode)
				QualifyLevel = strconv.Itoa(3)
			} else if rank <= 4 {
				min = helpers.CutOffDecimal(150000, 0, ".", ",")
				descMin = helpers.Translate("4th_block", r.LangCode)
				max = helpers.CutOffDecimal(200000, 0, ".", ",")
				descMax = helpers.Translate("5th_block", r.LangCode)
				// perc = directSponsorLegAmt / float64(200000)
				barDesc = helpers.Translate("amount_of_direct_referral", r.LangCode)
				QualifyLevel = strconv.Itoa(4) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(6)
			} else if rank <= 5 {
				min = helpers.CutOffDecimal(200000, 0, ".", ",")
				descMin = helpers.Translate("5th_block", r.LangCode)
				max = helpers.CutOffDecimal(300000, 0, ".", ",")
				descMax = helpers.Translate("6th_block", r.LangCode)
				// perc = directSponsorLegAmt / float64(300000)
				barDesc = helpers.Translate("amount_of_direct_referral", r.LangCode)
				QualifyLevel = strconv.Itoa(7) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(10)
			} else if rank <= 6 {
				min = helpers.CutOffDecimal(300000, 0, ".", ",")
				descMin = helpers.Translate("6th_block", r.LangCode)
				max = helpers.CutOffDecimal(400000, 0, ".", ",")
				descMax = helpers.Translate("7th_block", r.LangCode)
				// perc = directSponsorLegAmt / float64(400000)
				barDesc = helpers.Translate("amount_of_direct_referral", r.LangCode)
				QualifyLevel = strconv.Itoa(11) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(15)
			} else if rank <= 7 {
				min = helpers.CutOffDecimal(400000, 0, ".", ",")
				descMin = helpers.Translate("7th_block", r.LangCode)
				max = helpers.CutOffDecimal(500000, 0, ".", ",")
				descMax = helpers.Translate("8th_block", r.LangCode)
				// perc = directSponsorLegAmt / float64(500000)
				barDesc = helpers.Translate("amount_of_direct_referral", r.LangCode)
				QualifyLevel = strconv.Itoa(16) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(20)
			} else if rank <= 8 {
				min = helpers.CutOffDecimal(500000, 0, ".", ",")
				descMin = helpers.Translate("8th_block", r.LangCode)
				max = helpers.CutOffDecimal(600000, 0, ".", ",")
				descMax = helpers.Translate("9th_block", r.LangCode)
				// perc = directSponsorLegAmt / float64(600000)
				barDesc = helpers.Translate("amount_of_direct_referral", r.LangCode)
				QualifyLevel = strconv.Itoa(21) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(25)
			} else if rank <= 9 {
				min = helpers.CutOffDecimal(600000, 0, ".", ",")
				descMin = helpers.Translate("9th_block", r.LangCode)
				max = helpers.CutOffDecimal(600000, 0, ".", ",")
				// descMax = "9th" + " " + helpers.Translate("block", r.LangCode)
				// perc = directSponsorLegAmt / float64(600000)
				barDesc = helpers.Translate("amount_of_direct_referral", r.LangCode)
				QualifyLevel = strconv.Itoa(26) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(30)
			}

			// if perc > 1 {
			// 	perc = 1
			// }

			arrExtra = map[string]interface{}{
				"max":      max,
				"min":      min,
				"min_desc": descMin,
				"max_desc": descMax,
				"bar_desc": barDesc,
				// "perc":     perc,
			}

			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("qualify_level", r.LangCode),
				"value":  QualifyLevel,
			})
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("yesterday", r.LangCode),
				"value":  DailyIncomeStr,
			})
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("total_reward", r.LangCode),
				"value":  totalRwd,
			})
		case "REBATE":
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("yesterday", r.LangCode),
				"value":  DailyIncomeStr,
			})
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("total_reward", r.LangCode),
				"value":  totalRwd,
			})

			return arrRewardSummaryList, nil
		case "COMMUNITY": //RANKING
			var (
				descMin  string = helpers.Translate("member", r.LangCode)
				descMax  string = helpers.Translate("associate", r.LangCode)
				barDesc  string
				min      string = "0"
				max      string = "4"
				perc     float64
				rankDesc string = "-"
			)
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: "tbl_bonus_rank_star.t_member_id = ?", CondValue: r.MemberID},
				models.WhereCondFn{Condition: "tbl_bonus_rank_star.b_latest = ?", CondValue: 1},
				models.WhereCondFn{Condition: "tbl_bonus_rank_star.t_bns_fr <= ?", CondValue: curDateTime},
				models.WhereCondFn{Condition: "tbl_bonus_rank_star.t_bns_to >= ?", CondValue: curDateTime},
			)
			result, err := models.GetTblBonusRankStarFn(arrCond, false)
			if err != nil {
				base.LogErrorLog("RewardSummary - fail to get TblBonusRankBlockFn in rwdType", err, r, true)
				return nil, err
			}

			if len(result) > 0 {
				Rank = result[0].TRankEff
			}
			if Rank == 1 {
				rankDesc = helpers.Translate("associate", r.LangCode)
			} else if Rank == 2 {
				rankDesc = helpers.Translate("partner", r.LangCode)
			} else if Rank == 3 {
				rankDesc = helpers.Translate("senior_partner", r.LangCode)
			} else if Rank == 4 {
				rankDesc = helpers.Translate("regional_partner", r.LangCode)
			} else if Rank == 5 {
				rankDesc = helpers.Translate("global_partner", r.LangCode)
			} else if Rank == 6 {
				rankDesc = helpers.Translate("director", r.LangCode)
			}

			rst, _ := models.GetCommunityBonusByMemberId(r.MemberID, "", "")
			if len(rst) > 0 {
				if rst[0].FPerc <= 0.01 {
					descMin = helpers.Translate("associate", r.LangCode)
					descMax = helpers.Translate("partner", r.LangCode)
					barDesc = helpers.Translate("associate_in_4_different_network", r.LangCode)
					min = helpers.CutOffDecimal(0, 2, ".", ",")
					max = helpers.CutOffDecimal(0.03, 2, ".", ",")
					perc = rst[0].FPerc / 0.03
				} else if rst[0].FPerc <= 0.03 {
					descMin = helpers.Translate("partner", r.LangCode)
					descMax = helpers.Translate("senior_partner", r.LangCode)
					barDesc = helpers.Translate("partner_in_4_different_network", r.LangCode)
					min = helpers.CutOffDecimal(0.03, 2, ".", ",")
					max = helpers.CutOffDecimal(0.05, 2, ".", ",")
					perc = rst[0].FPerc / 0.05
				} else if rst[0].FPerc <= 0.05 {
					descMin = helpers.Translate("senior_partner", r.LangCode)
					descMax = helpers.Translate("regional_partner", r.LangCode)
					barDesc = helpers.Translate("senior_partner_in_4_different_network", r.LangCode)
					min = helpers.CutOffDecimal(0.05, 2, ".", ",")
					max = helpers.CutOffDecimal(0.07, 2, ".", ",")
					perc = rst[0].FPerc / 0.07
				} else if rst[0].FPerc <= 0.07 {
					descMin = helpers.Translate("regional_partner", r.LangCode)
					descMax = helpers.Translate("global_partner", r.LangCode)
					barDesc = helpers.Translate("regional_partner_in_4_different_network", r.LangCode)
					min = helpers.CutOffDecimal(0.07, 2, ".", ",")
					max = helpers.CutOffDecimal(0.10, 2, ".", ",")
					perc = rst[0].FPerc / 0.10
				} else if rst[0].FPerc <= 0.10 {
					descMin = helpers.Translate("global_partner", r.LangCode)
					descMax = helpers.Translate("director", r.LangCode)
					barDesc = helpers.Translate("global_partner_in_4_different_network", r.LangCode)
					perc = rst[0].FPerc / 0.10
				} else {
					descMin = helpers.Translate("director", r.LangCode)
					perc = 1
				}
			}

			if perc > 1 {
				perc = 1
			}

			arrExtra = map[string]interface{}{
				"min":      min,
				"max":      max,
				"min_desc": descMin,
				"max_desc": descMax,
				"bar_desc": barDesc,
				"perc":     perc,
			}

			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("rank", r.LangCode),
				"value":  rankDesc,
			})
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("current_month", r.LangCode),
				"value":  DailyIncomeStr,
			})
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("total_reward", r.LangCode),
				"value":  totalRwd,
			})
		case "PAIR":
			var (
				total1   string = "0"
				balance1 string = "0"
				total2   string = "0"
				balance2 string = "0"
				total3   string = "0"
				balance3 string = "0"
				total4   string = "0"
				balance4 string = "0"
				total5   string = "0"
				balance5 string = "0"
				total6   string = "0"
				balance6 string = "0"
			)
			rst, err := models.GetPairBonusByMemberId(r.MemberID, "", "")
			if err != nil {
				base.LogErrorLog("RewardSummary - fail to get GetPairBonusByMemberId in rwdType", err, r, true)
				return nil, err
			}

			if len(rst) > 0 {
				total1 = helpers.CutOffDecimal(rst[0].TotalPerLeg1, decimalPoint, ".", ",")
				balance1 = helpers.CutOffDecimal(rst[0].Leg1, decimalPoint, ".", ",")

				total2 = helpers.CutOffDecimal(rst[0].TotalPerLeg2, decimalPoint, ".", ",")
				balance2 = helpers.CutOffDecimal(rst[0].Leg2, decimalPoint, ".", ",")

				total3 = helpers.CutOffDecimal(rst[0].TotalPerLeg3, decimalPoint, ".", ",")
				balance3 = helpers.CutOffDecimal(rst[0].Leg3, decimalPoint, ".", ",")

				total4 = helpers.CutOffDecimal(rst[0].TotalPerLeg4, decimalPoint, ".", ",")
				balance4 = helpers.CutOffDecimal(rst[0].Leg4, decimalPoint, ".", ",")

				total5 = helpers.CutOffDecimal(rst[0].TotalPerLeg12, decimalPoint, ".", ",") //1+2
				balance5 = helpers.CutOffDecimal(rst[0].Leg12, decimalPoint, ".", ",")       //1+2

				total6 = helpers.CutOffDecimal(rst[0].TotalPerLeg34, decimalPoint, ".", ",") //3+4
				balance6 = helpers.CutOffDecimal(rst[0].Leg34, decimalPoint, ".", ",")       //3+4
			}

			arrExtra = map[string]interface{}{
				"1":          helpers.Translate("1", r.LangCode),
				"1total":     total1,
				"1balance":   balance1,
				"2":          helpers.Translate("2", r.LangCode),
				"2total":     total2,
				"2balance":   balance2,
				"3":          helpers.Translate("3", r.LangCode),
				"3total":     total3,
				"3balance":   balance3,
				"4":          helpers.Translate("4", r.LangCode),
				"4total":     total4,
				"4balance":   balance4,
				"1+2":        helpers.Translate("1+2", r.LangCode),
				"1+2total":   total5,
				"1+2balance": balance5,
				"3+4":        helpers.Translate("3+4", r.LangCode),
				"3+4total":   total6,
				"3+4balance": balance6,
			}

			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("yesterday", r.LangCode),
				"value":  DailyIncomeStr,
			})
			arrBonusInfo = append(arrBonusInfo, map[string]interface{}{
				"header": helpers.Translate("total_reward", r.LangCode),
				"value":  totalRwd,
			})
		}

		arrRewardSummaryList = map[string]interface{}{
			"info":  arrBonusInfo,
			"extra": arrExtra,
		}
	} else {

		// total revenue
		TotalRevenue, err := models.GetMemberTotalBns(r.MemberID, "", 0, 0, 0)

		if err != nil {
			base.LogErrorLog("RewardSummary - fail to get total revenue", err, r, true)
			return nil, err
		}

		if TotalRevenue != nil {
			TotalIncome = TotalRevenue.TotalBonus
		}

		// monthly income
		MonthlyRevenue, err := models.GetMemberTotalBns(r.MemberID, "", 1, 0, 0)
		if err != nil {
			base.LogErrorLog("RewardSummary - fail to get monthly revenue", err, r, true)
			return nil, err
		}

		if MonthlyRevenue != nil {
			MonthlyIncome = MonthlyRevenue.TotalBonus
		}

		//daily income
		DailyRevenue, err := models.GetMemberTotalBns(r.MemberID, "", 0, 1, 0)
		if err != nil {
			base.LogErrorLog("RewardSummary - fail to get daily revenue", err, r, true)
			return nil, err
		}

		if DailyRevenue != nil {
			DailyIncome = DailyRevenue.TotalBonus
		}

		TotalIncomeStr = helpers.CutOffDecimal(TotalIncome, decimalPoint, ".", ",")
		MonthlyIncomeStr = helpers.CutOffDecimal(MonthlyIncome, decimalPoint, ".", ",")
		DailyIncomeStr = helpers.CutOffDecimal(DailyIncome, decimalPoint, ".", ",")

		//get package amount
		arrMemberTotalSalesFn := make([]models.WhereCondFn, 0)
		arrMemberTotalSalesFn = append(arrMemberTotalSalesFn,
			models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: r.MemberID},
			models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
		)
		arrMemberTotalSales, _ := models.GetMemberTotalSalesFn(arrMemberTotalSalesFn, false)

		//get prd_group_type setting
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "prd_group_type.code = ?", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "prd_group_type.status = ?", CondValue: "A"},
		)
		arrGetPrdGroupType, _ := models.GetPrdGroupTypeFn(arrCond, "", false)

		var arrPackageList map[string][]arrPackageSetting
		json.Unmarshal([]byte(arrGetPrdGroupType[0].Setting), &arrPackageList)
		arrPackageTiers = arrPackageList["tiers"]

		//get member tier
		packageType, err := member_service.GetMemberTier(r.MemberID)
		if err != nil {
			base.LogErrorLog("RewardSummary - GetMemberTier() return err", r, err.Error(), true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", r.LangCode), Data: r}
		}

		if packageType == "" {
			//no tier
			packageType = "B0"
		}

		if len(arrPackageTiers) > 0 {
			for _, valPack := range arrPackageList["tiers"] {
				if packageType == valPack.Tier {
					PackageAmountMin = valPack.Min
					PackageAmountMax = valPack.Max
				}
			}
		}

		PackageAmountValue = arrMemberTotalSales.TotalAmount

		packagePerc := float64(0)
		if PackageAmountMax > 0 {
			packagePerc, _ = decimal.NewFromFloat(PackageAmountValue).Div(decimal.NewFromFloat(PackageAmountMax)).Float64()
		}

		if packagePerc > 1 {
			packagePerc = 1
		}

		PackageAmountMaxStr := helpers.CutOffDecimal(PackageAmountMax, 0, ".", ",")
		imgLink := "https://media02.securelayers.cloud/medias/GTA/PACKAGE/ICON/" + packageType + ".jpg"

		packageTypeMax := packageType

		switch strings.ToUpper(packageType) {
		case "B0":
			packageTypeMax = "B1"
		case "B1":
			packageTypeMax = "B2"
		case "B2":
			packageTypeMax = "B3"
		case "B3":
			packageTypeMax = "B4"
		case "B4":
			packageTypeMax = "B5"
		case "B5":
		}
		imgLinkMax := "https://media02.securelayers.cloud/medias/GTA/PACKAGE/ICON/" + packageTypeMax + ".jpg"
		if packageType == "B5" {
			PackageAmountMaxStr = ""
			imgLinkMax = ""
		}
		arrPackage := map[string]interface{}{
			"package_min_img":      imgLink,
			"package_amount_min":   helpers.CutOffDecimal(PackageAmountMin, 0, ".", ","),
			"package_amount_max":   PackageAmountMaxStr,
			"package_max_img":      imgLinkMax,
			"package_amount_value": helpers.CutOffDecimal(PackageAmountValue, 0, ".", ","),
			"package_perc":         packagePerc,
		}

		//get CAP Balance & setup
		arrCond = make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ewt_setup.status = ?", CondValue: "A"},
			models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "CAP"},
			models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: r.MemberID},
		)
		bal, err := models.GetMemberEwtSetupBalanceFn(r.MemberID, arrCond, "", false)

		if err != nil {
			base.LogErrorLog("RewardSummary - failed to get CAP balance", err.Error(), map[string]interface{}{"err": err, "data": r}, true)
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: helpers.Translate("something_went_wrong", r.LangCode), Data: r}
		}

		if len(bal) > 0 {
			capBalance = bal[0].Balance
		}

		strCapBalance = helpers.CutOffDecimal(capBalance, 0, ".", ",")

		//get totalCapEarning
		arrMemberTotalCapFn := make([]models.WhereCondFn, 0)
		arrMemberTotalCapFn = append(arrMemberTotalCapFn,
			models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: r.MemberID},
			models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
		)
		arrMemberTotalCap, _ := models.GetMemberTotalCapFn(arrMemberTotalCapFn, false)

		totalCapEarning = arrMemberTotalCap.TotalAmount
		totalCapEarningStr = helpers.CutOffDecimal(totalCapEarning, 0, ".", ",")

		if totalCapEarning > 0 {
			earnBarPerc, _ = decimal.NewFromFloat(capBalance).Div(decimal.NewFromFloat(totalCapEarning)).Float64()
		}

		if earnBarPerc > 1 {
			earnBarPerc = 1
		}

		//get earning cap
		arrEarningCap := map[string]interface{}{
			"earning_cap_value": strCapBalance,
			"earning_cap_total": totalCapEarningStr,
			"earning_cap_perc":  earnBarPerc,
		}

		arrGeneralBnsListSetting, err := models.GetSysGeneralSetupByID("reward_type_list_setting")
		var arrStatementListSettingList map[string][]arrBonusSummaryReturnStruct
		json.Unmarshal([]byte(arrGeneralBnsListSetting.InputType1), &arrStatementListSettingList)
		arrBonusSummaryList = arrStatementListSettingList["reward_type_list"]

		for k, v1 := range arrStatementListSettingList["reward_type_list"] {
			v1.RewardTypeName = helpers.Translate(v1.RewardTypeName, r.LangCode)
			TotalReward, _ := models.GetMemberTotalBns(r.MemberID, strings.ToUpper(v1.RewardType), 0, 0, 1)

			totalRwd := helpers.CutOffDecimal(0, decimalPoint, ".", ",")

			if TotalReward != nil {
				totalRwd = helpers.CutOffDecimal(TotalReward.TotalBonus, decimalPoint, ".", ",")
			}

			switch strings.ToUpper(v1.RewardType) {
			case "SPONSOR":
				// get total direct sponsor
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: r.MemberID},
					models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
				)
				totalDirectSponsorRst, _ := models.GetTotalDirectSponsorFn(arrCond, false)
				totalDirectSponsor := "0"
				if totalDirectSponsorRst.TotalDirectSponsor > 0 {
					totalDirectSponsor = helpers.CutOffDecimal(totalDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
				}

				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("total_referrals", r.LangCode),
					"value":  totalDirectSponsor,
				})

				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("total_reward", r.LangCode),
					"value":  totalRwd,
				})
			case "GENERATION":
				var (
					rank     int
					blockStr string = "-"
				)
				// arrCond := make([]models.WhereCondFn, 0)
				// arrCond = append(arrCond,
				// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_member_id = ?", CondValue: r.MemberID},
				// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.b_latest = ?", CondValue: 1},
				// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_bns_fr <= ?", CondValue: curDateTime},
				// 	models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_bns_to >= ?", CondValue: curDateTime},
				// )
				// result, err := models.GetTblBonusRankBlockFn(arrCond, "", false)
				// if err != nil {
				// 	base.LogErrorLog("RewardSummary - fail to get GetTblBonusRankBlockFn", err, r, true)
				// 	return nil, err
				// }

				// if len(result) > 0 {
				// 	Block = result[0].TRankEff
				// 	QualifyLevel = result[0].TRankQualify
				// }

				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_member_id = ?", CondValue: r.MemberID},
					models.WhereCondFn{Condition: "tbl_bonus_rank_block.t_member_id = ?", CondValue: r.MemberID},
					models.WhereCondFn{Condition: "tbl_bonus_rank_block.b_latest = ?", CondValue: 1},
				)
				// rst, _ := models.GetTblBonusRankStarPassupFn(arrCond, "", false)
				rst, _ := models.GetTblBonusRankBlockFn(arrCond, "", false)

				if len(rst) > 0 {
					rank = rst[0].TRankEff
				}
				if rank <= 0 {
					QualifyLevel = strconv.Itoa(0)
				} else if rank <= 1 {
					blockStr = helpers.Translate("1st_block", r.LangCode)
					QualifyLevel = strconv.Itoa(1)
				} else if rank <= 2 {
					blockStr = helpers.Translate("2nd_block", r.LangCode)
					QualifyLevel = strconv.Itoa(2)
				} else if rank <= 3 {
					blockStr = helpers.Translate("3rd_block", r.LangCode)
					QualifyLevel = strconv.Itoa(3)
				} else if rank <= 4 {
					blockStr = helpers.Translate("4th_block", r.LangCode)
					QualifyLevel = strconv.Itoa(4) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(6)
				} else if rank <= 5 {
					blockStr = helpers.Translate("5th_block", r.LangCode)
					QualifyLevel = strconv.Itoa(7) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(10)
				} else if rank <= 6 {
					blockStr = helpers.Translate("6th_block", r.LangCode)
					QualifyLevel = strconv.Itoa(11) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(15)
				} else if rank <= 7 {
					blockStr = helpers.Translate("7th_block", r.LangCode)
					QualifyLevel = strconv.Itoa(16) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(20)
				} else if rank <= 8 {
					blockStr = helpers.Translate("8th_block", r.LangCode)
					QualifyLevel = strconv.Itoa(21) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(25)
				} else if rank <= 9 {
					blockStr = helpers.Translate("9th_block", r.LangCode)
					QualifyLevel = strconv.Itoa(26) + " " + helpers.Translate("to", r.LangCode) + " " + strconv.Itoa(30)
				}
				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("block", r.LangCode),
					"value":  blockStr,
				})

				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("generation_unlocked", r.LangCode),
					"value":  QualifyLevel,
				})

				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("total_reward", r.LangCode),
					"value":  totalRwd,
				})
			case "REBATE":
				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("total_reward", r.LangCode),
					"value":  totalRwd,
				})
			case "PAIR":
				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("total_reward", r.LangCode),
					"value":  totalRwd,
				})
			case "COMMUNITY": //RANKING
				var (
					QualifyPercStr string = "-"
					rankDesc       string = "-"
				)
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: "tbl_bonus_rank_star.t_member_id = ?", CondValue: r.MemberID},
					models.WhereCondFn{Condition: "tbl_bonus_rank_star.b_latest = ?", CondValue: 1},
					models.WhereCondFn{Condition: "tbl_bonus_rank_star.t_bns_fr <= ?", CondValue: curDateTime},
					models.WhereCondFn{Condition: "tbl_bonus_rank_star.t_bns_to >= ?", CondValue: curDateTime},
				)
				result, err := models.GetTblBonusRankStarFn(arrCond, false)
				if err != nil {
					base.LogErrorLog("RewardSummary - fail to get TblBonusRankBlockFn", err, r, true)
					return nil, err
				}

				if len(result) > 0 {
					Rank = result[0].TRankEff
				}

				if Rank == 1 {
					rankDesc = helpers.Translate("associate", r.LangCode)
				} else if Rank == 2 {
					rankDesc = helpers.Translate("partner", r.LangCode)
				} else if Rank == 3 {
					rankDesc = helpers.Translate("senior_partner", r.LangCode)
				} else if Rank == 4 {
					rankDesc = helpers.Translate("regional_partner", r.LangCode)
				} else if Rank == 5 {
					rankDesc = helpers.Translate("global_partner", r.LangCode)
				} else if Rank == 6 {
					rankDesc = helpers.Translate("director", r.LangCode)
				}
				// rst, _ := models.GetCommunityBonusByMemberId(r.MemberID, "", "")

				// if len(rst) > 0 {
				// 	perc := float.Mul(rst[0].FPerc, 100)
				// 	QualifyPercStr = helpers.CutOffDecimal(perc, 0, ".", ",") + "%"

				// 	if rst[0].FPerc <= 0.01 {
				// 		rankDesc = helpers.Translate("associate", r.LangCode)
				// 	} else if rst[0].FPerc <= 0.03 {
				// 		rankDesc = helpers.Translate("partner", r.LangCode)
				// 	} else if rst[0].FPerc <= 0.05 {
				// 		rankDesc = helpers.Translate("senior_partner", r.LangCode)
				// 	} else if rst[0].FPerc <= 0.07 {
				// 		rankDesc = helpers.Translate("regional_partner", r.LangCode)
				// 	} else if rst[0].FPerc <= 0.10 {
				// 		rankDesc = helpers.Translate("global_partner", r.LangCode)
				// 	} else {
				// 		rankDesc = helpers.Translate("director", r.LangCode)
				// 	}
				// }

				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("rank", r.LangCode),
					"value":  rankDesc,
				})

				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("qualify", r.LangCode),
					"value":  QualifyPercStr,
				})

				v1.Info = append(v1.Info, map[string]interface{}{
					"header": helpers.Translate("total_reward", r.LangCode),
					"value":  totalRwd,
				})
			}

			arrBonusSummaryList[k] = v1
		}

		arrRewardSummaryList = map[string]interface{}{
			"total_revenue":    TotalIncomeStr,
			"monthly_income":   MonthlyIncomeStr,
			"daily_income":     DailyIncomeStr,
			"package_info":     arrPackage,
			"bonus_list":       arrBonusSummaryList,
			"earning_capacity": arrEarningCap,
		}
	}

	return arrRewardSummaryList, nil

}

type RewardReturnResultStruct struct {
	Header   string `json:"header"`
	Value    string `json:"value"`
	SubValue string `json:"sub_value"`
}

type RebateBonusDetailsStruct struct {
	// ID         string `gorm:"t_bns_id" json:"id"`
	Date     string `json:"date"`
	Username string `json:"username"`
	// DocNo string `json:"doc_no"`
	// Bv         string      `json:"bv"`
	// Percentage string      `json:"percentage"`
	// Bonus     string      `json:"bonus"`
	CreatedAt string      `json:"created_at"`
	List      interface{} `json:"list"`
}

type BonusDetailsStruct struct {
	Date      string      `json:"date"`
	Username  string      `json:"username"`
	CreatedAt string      `json:"created_at"`
	List      interface{} `json:"list"`
}

type GlobalBonusDetailsStruct struct {
	// ID         string `gorm:"t_bns_id" json:"id"`
	Date string `json:"date"`
	// Username string `json:"username"`
	// RebateBonus    string      `json:"rebate_bonus"`
	// MatchingBonus  string      `json:"matching_bonus"`
	// CommunityBonus string      `json:"community_bonus"`
	// SponsorBonus   string      `json:"sponsor_bonus"`
	// Bonus          string      `json:"bonus"`
	// CreatedAt string      `json:"created_at"`
	List interface{} `json:"list"`
}

// reward statement v2 with slice paginate
func (r *RewardStatementPostStruct) RewardStatement() (interface{}, error) {

	var (
		arrRewardStatement       []interface{}
		arrModuleRewardStatement []interface{}
		decimalPoint             uint = 0
		rwdType                       = r.RwdType
		tierTranslated                = helpers.Translate("tier", r.LangCode)
		nodesTranslated               = helpers.Translate("amount", r.LangCode)
		percentageTranslated          = helpers.Translate("percentage", r.LangCode)
		bonusTranslated               = helpers.Translate("bonus", r.LangCode)
		docNoTranslated               = helpers.Translate("doc_no", r.LangCode)
		// burnBvTranslated              = helpers.Translate("burn_bv", r.LangCode)
		downlineIDTranslated = helpers.Translate("downline_id", r.LangCode)
		levelTranslated      = helpers.Translate("level", r.LangCode)
		paidLevelTranslated  = helpers.Translate("paid_level", r.LangCode)
		uernameIDTranslated  = helpers.Translate("username_id", r.LangCode)
	)

	// if r.DateFrom == "" && r.DateTo == "" {
	// 	r.DateFrom = base.GetCurrentDateTimeT().AddDate(0, 0, -30).Format("2006-01-02")
	// 	r.DateTo = base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02") //yesterday date
	// }

	switch strings.ToUpper(rwdType) {

	case "REBATE":
		rst, err := models.GetRebateBonusByMemberId(r.MemberID, r.DateFrom, r.DateTo)

		if err != nil {
			base.LogErrorLog("RewardStatement - fail to get Rebate Bonus", err, r, true)
			return nil, err
		}

		if len(rst) > 0 {
			for _, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				arrModuleRewardStatement = make([]interface{}, 0)
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: tierTranslated,
					Value:  v.MemberTier,
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: nodesTranslated,
					Value:  helpers.CutOffDecimal(v.FBv, 2, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header:   percentageTranslated,
					Value:    helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					SubValue: "(%)",
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: bonusTranslated,
					Value:  helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
				})

				arrRewardStatement = append(arrRewardStatement, BonusDetailsStruct{
					Date:      v.TBnsId,
					Username:  v.NickName,
					CreatedAt: v.DtTimestamp.Format("2006-01-02 15:04:05"),
					List:      arrModuleRewardStatement,
				})
			}
		}

	case "SPONSOR":
		rst, err := models.GetSponsorBonusByMemberId(r.MemberID, r.DateFrom, r.DateTo)

		if err != nil {
			base.LogErrorLog("RewardStatement - fail to get Sponsor Bonus", err, r, true)
			return nil, err
		}

		if len(rst) > 0 {
			for _, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				arrModuleRewardStatement = make([]interface{}, 0)

				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: docNoTranslated,
					Value:  v.DocNo,
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: nodesTranslated,
					Value:  helpers.CutOffDecimal(v.FBv, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header:   percentageTranslated,
					Value:    helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					SubValue: "(%)",
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: bonusTranslated,
					Value:  helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
				})
				// arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
				// 	Header: burnBvTranslated,
				// 	Value:  helpers.CutOffDecimal(v.FBnsBurn, decimalPoint, ".", ","),
				// })
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: uernameIDTranslated,
					Value:  v.DownlineID,
				})

				arrRewardStatement = append(arrRewardStatement, BonusDetailsStruct{
					Date:      v.TBnsID,
					Username:  v.Username,
					CreatedAt: v.DtCreated.Format("2006-01-02 15:04:05"),
					List:      arrModuleRewardStatement,
				})
			}
		}
	case "COMMUNITY": //RANKING
		rst, err := models.GetCommunityBonusByMemberId(r.MemberID, r.DateFrom, r.DateTo)

		if err != nil {
			base.LogErrorLog("RewardStatement - fail to get Community Bonus", err, r, true)
			return nil, err
		}

		if len(rst) > 0 {
			for _, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				arrModuleRewardStatement = make([]interface{}, 0)

				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: downlineIDTranslated,
					Value:  v.DownlineId,
				})
				// arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
				// 	Header: levelTranslated,
				// 	Value:  v.ILvl,
				// })
				// arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
				// 	Header: paidLevelTranslated,
				// 	Value:  v.ILvlPaid,
				// })
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: nodesTranslated,
					Value:  helpers.CutOffDecimal(v.FBv, 2, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header:   percentageTranslated,
					Value:    helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					SubValue: "%",
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: bonusTranslated,
					Value:  helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
				})
				// arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
				// 	Header: burnBvTranslated,
				// 	Value:  helpers.CutOffDecimal(v.FBnsBurn, 2, ".", ","),
				// })

				arrRewardStatement = append(arrRewardStatement, BonusDetailsStruct{
					Date:      v.TBnsID,
					Username:  v.Username,
					CreatedAt: v.DtCreated.Format("2006-01-02 15:04:05"),
					List:      arrModuleRewardStatement,
				})
			}
		}

	case "GENERATION":
		rst, err := models.GetGenerationBonusByMemberId(r.MemberID, r.DateFrom, r.DateTo)

		if err != nil {
			base.LogErrorLog("RewardStatement - fail to get Generation Bonus", err, r, true)
			return nil, err
		}

		if len(rst) > 0 {
			for _, v := range rst {
				v.FPerc = float.Mul(v.FPerc, 100)
				arrModuleRewardStatement = make([]interface{}, 0)
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: docNoTranslated,
					Value:  v.TDocNo,
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: uernameIDTranslated,
					Value:  v.DownlineId,
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: levelTranslated,
					Value:  v.ILvl,
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: paidLevelTranslated,
					Value:  v.ILvlPaid,
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: nodesTranslated,
					Value:  helpers.CutOffDecimal(v.FBv, 2, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header:   percentageTranslated,
					Value:    helpers.CutOffDecimal(v.FPerc, 2, ".", ","),
					SubValue: "%",
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: bonusTranslated,
					Value:  helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
				})
				// arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
				// 	Header: burnBvTranslated,
				// 	Value:  helpers.CutOffDecimal(v.BurnBns, 2, ".", ","),
				// })

				arrRewardStatement = append(arrRewardStatement, BonusDetailsStruct{
					Date:      v.TBnsId,
					Username:  v.Username,
					CreatedAt: v.DtCreated.Format("2006-01-02 15:04:05"),
					List:      arrModuleRewardStatement,
				})
			}
		}

	case "PAIR":
		rst, err := models.GetPairBonusByMemberId(r.MemberID, r.DateFrom, r.DateTo)

		if err != nil {
			base.LogErrorLog("RewardStatement - fail to get Pair Bonus", err, r, true)
			return nil, err
		}

		if len(rst) > 0 {
			var (
				totalLeg1Translated  = helpers.Translate("total_leg1", r.LangCode)
				totalLeg2Translated  = helpers.Translate("total_leg2", r.LangCode)
				totalLeg3Translated  = helpers.Translate("total_leg3", r.LangCode)
				totalLeg4Translated  = helpers.Translate("total_leg4", r.LangCode)
				totalLeg12Translated = helpers.Translate("total_leg1+2", r.LangCode)
				totalLeg34Translated = helpers.Translate("total_leg3+4", r.LangCode)
				leg1Translated       = helpers.Translate("leg1", r.LangCode)
				leg2Translated       = helpers.Translate("leg2", r.LangCode)
				leg3Translated       = helpers.Translate("leg3", r.LangCode)
				leg4Translated       = helpers.Translate("leg4", r.LangCode)
				leg12Translated      = helpers.Translate("leg1+2", r.LangCode)
				leg34Translated      = helpers.Translate("leg3+4", r.LangCode)
			)
			for _, v := range rst {
				arrModuleRewardStatement = make([]interface{}, 0)

				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: totalLeg1Translated,
					Value:  helpers.CutOffDecimal(v.TotalPerLeg1, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: totalLeg2Translated,
					Value:  helpers.CutOffDecimal(v.TotalPerLeg2, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: totalLeg3Translated,
					Value:  helpers.CutOffDecimal(v.TotalPerLeg3, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: totalLeg4Translated,
					Value:  helpers.CutOffDecimal(v.TotalPerLeg4, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: totalLeg12Translated,
					Value:  helpers.CutOffDecimal(v.TotalPerLeg12, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: totalLeg34Translated,
					Value:  helpers.CutOffDecimal(v.TotalPerLeg34, decimalPoint, ".", ","),
				})

				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: leg1Translated,
					Value:  helpers.CutOffDecimal(v.Leg1, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: leg2Translated,
					Value:  helpers.CutOffDecimal(v.Leg2, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: leg3Translated,
					Value:  helpers.CutOffDecimal(v.Leg3, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: leg4Translated,
					Value:  helpers.CutOffDecimal(v.Leg4, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: leg12Translated,
					Value:  helpers.CutOffDecimal(v.Leg12, decimalPoint, ".", ","),
				})
				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: leg34Translated,
					Value:  helpers.CutOffDecimal(v.Leg34, decimalPoint, ".", ","),
				})

				arrModuleRewardStatement = append(arrModuleRewardStatement, RewardReturnResultStruct{
					Header: bonusTranslated,
					Value:  helpers.CutOffDecimal(v.FBns, decimalPoint, ".", ","),
				})
				arrRewardStatement = append(arrRewardStatement, BonusDetailsStruct{
					Date:      v.TBnsFr,
					Username:  v.Username,
					CreatedAt: v.DtCreated.Format("2006-01-02 15:04:05"),
					List:      arrModuleRewardStatement,
				})
			}
		}
	}

	//sort array
	// sort.Slice(arrRewardStatement, func(p, q int) bool {
	// 	return arrRewardStatement[q].Date < arrRewardStatement[p].Date
	// })

	page := base.Pagination{
		Page:    r.Page,
		DataArr: arrRewardStatement,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, nil
}

// GetContractMiningActionHistoryListStruct struct
type GetContractMiningActionHistoryListStruct struct {
	MemberID int
	LangCode string
	Page     int64
	DateFrom string
	DateTo   string
}

// ContractMiningActionHistoryList struct
type ContractMiningActionHistoryList struct {
	Date       string  `json:"date"`
	PoolAmount float64 `json:"pool_amount"`
}

// GetContractMiningActionHistoryList func
func GetContractMiningActionHistoryList(arrData GetContractMiningActionHistoryListStruct) (interface{}, string) {
	arrContractMiningHistoryList := make([]ContractMiningActionHistoryList, 0)

	// get cur date
	curDate, err := base.GetCurrentTimeV2("yyyy-mm-dd")
	if err != nil {
		base.LogErrorLog("rewardService:GetContractMiningActionHistoryList()", "GetCurrentTimeV2():1", err.Error(), true)
		return nil, "something_went_wrong"
	}

	dateFrom := "2021-01-01"
	dateTo := curDate
	// get sponsor pool history in date range
	arrTblBonusSponsorPoolHistoryFn := make([]models.WhereCondFn, 0)
	if arrData.DateFrom != "" {
		dateFrom = arrData.DateFrom
	}
	// if arrData.DateTo != "" {
	// 	dateTo = arrData.DateTo
	// }

	arrTblBonusSponsorPoolHistoryFn = append(arrTblBonusSponsorPoolHistoryFn,
		models.WhereCondFn{Condition: " date(tbl_bonus_sponsor_pool.t_bns_id) >= ? ", CondValue: dateFrom},
		models.WhereCondFn{Condition: " date(tbl_bonus_sponsor_pool.t_bns_id) <= ? ", CondValue: dateTo},
	)

	arrTblBonusSponsorPoolList, _ := models.GetTblBonusSponsorPoolFn(arrTblBonusSponsorPoolHistoryFn, "", false)
	if len(arrTblBonusSponsorPoolList) > 0 {

		for _, arrTblBonusSponsorPoolListV := range arrTblBonusSponsorPoolList {
			tBnsID := string(arrTblBonusSponsorPoolListV.TBnsId)

			// skip today's sponsor pool data get from tbl_bonus_sponsor_pool
			if tBnsID != curDate {
				arrContractMiningHistoryList = append(arrContractMiningHistoryList,
					ContractMiningActionHistoryList{
						Date:       tBnsID,
						PoolAmount: arrTblBonusSponsorPoolListV.TotalPool,
					},
				)
			}
		}
	}

	// add today's sponsor pool data get manually calculated from sls_master
	if curDate >= dateFrom && dateTo >= curDate {
		// get today total sales amount
		arrTotalSalesAmountFn := make([]models.WhereCondFn, 0)
		arrTotalSalesAmountFn = append(arrTotalSalesAmountFn,
			models.WhereCondFn{Condition: "sls_master.action = ?", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "sls_master.bns_batch = ?", CondValue: curDate},
			models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
		)
		arrTotalSalesAmount, err := models.GetTotalSalesAmount(arrTotalSalesAmountFn, false)

		if err != nil {
			base.LogErrorLog("rewardService:GetContractMiningActionHistoryList()", "GetTotalSalesAmount()", err.Error(), true)
			return nil, "something_went_wrong"
		}

		totalSalesAmount := arrTotalSalesAmount.TotalAmount

		// get ytd's carry forward pool amount
		arrTblBonusSponsorPoolFn := make([]models.WhereCondFn, 0)
		arrTblBonusSponsorPoolFn = append(arrTblBonusSponsorPoolFn,
			models.WhereCondFn{Condition: "t_bns_id = ?", CondValue: base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02")},
		)
		poolAmtRst, err := models.GetTblBonusSponsorPoolFn(arrTblBonusSponsorPoolFn, "", false)
		if err != nil {
			base.LogErrorLog("rewardService:GetContractMiningActionHistoryList()", "GetTblBonusSponsorFn()", err.Error(), true)
			return nil, "something_went_wrong"
		}
		carriedForwardPool := 0.0
		if len(poolAmtRst) > 0 {
			carriedForwardPool = float64(poolAmtRst[0].PoolCf)
		}

		// total_pool = today contract sales * 10%
		curPoolAmount := float.Mul(totalSalesAmount, 0.1) + carriedForwardPool

		arrContractMiningHistoryList = append(arrContractMiningHistoryList,
			ContractMiningActionHistoryList{
				Date:       curDate,
				PoolAmount: curPoolAmount,
			},
		)
	}

	// get sponsor pool markup
	arrSysSponsorPoolMarkupFn := make([]models.WhereCondFn, 0)
	arrSysSponsorPoolMarkupFn = append(arrSysSponsorPoolMarkupFn,
		models.WhereCondFn{Condition: " date(sys_sponsor_pool_markup.bns_date) >= ? ", CondValue: dateFrom},
		models.WhereCondFn{Condition: " date(sys_sponsor_pool_markup.bns_date) <= ? ", CondValue: dateTo},
	)

	arrSysSponsorPoolMarkup, _ := models.GetSysSponsorPoolMarkupFn(arrSysSponsorPoolMarkupFn, "", false)
	if len(arrSysSponsorPoolMarkup) > 0 {
		for _, arrSysSponsorPoolMarkupV := range arrSysSponsorPoolMarkup {
			paired := false
			for arrContractMiningHistoryListK, arrContractMiningHistoryListV := range arrContractMiningHistoryList {
				if arrContractMiningHistoryListV.Date == arrSysSponsorPoolMarkupV.BnsDate {
					arrContractMiningHistoryList[arrContractMiningHistoryListK].PoolAmount += arrSysSponsorPoolMarkupV.PoolAmount

					paired = true
				}
			}

			if !paired { // if not yet paired into arrContractMiningHistoryList, mean new bns_date, then append
				arrContractMiningHistoryList = append(arrContractMiningHistoryList,
					ContractMiningActionHistoryList{
						Date:       arrSysSponsorPoolMarkupV.BnsDate,
						PoolAmount: arrSysSponsorPoolMarkupV.PoolAmount,
					},
				)
			}
		}
	}

	//sort array
	sort.Slice(arrContractMiningHistoryList, func(p, q int) bool {
		return arrContractMiningHistoryList[q].Date < arrContractMiningHistoryList[p].Date
	})

	// convert to []interface{} for pagination
	var arrListingData []interface{}
	if len(arrContractMiningHistoryList) > 0 {
		for _, arrContractMiningHistoryListV := range arrContractMiningHistoryList {
			arrListingData = append(arrListingData,
				map[string]string{
					"date":        arrContractMiningHistoryListV.Date,
					"pool_amount": helpers.CutOffDecimal(arrContractMiningHistoryListV.PoolAmount, 2, ".", ","),
				},
			)
		}
	}

	page := base.Pagination{
		Page:    arrData.Page,
		DataArr: arrListingData,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, ""
}

// GetContractMiningActionRankingListStruct struct
type GetContractMiningActionRankingListStruct struct {
	MemberID  int
	LangCode  string
	Page      int64
	Date      string
	MaxNumber int
}

// ContractMiningActionRankingList struct
type ContractMiningActionRankingList struct {
	Username   string  `json:"username"`
	PoolAmount float64 `json:"pool_amount"`
}

// GetContractMiningActionRankingList func
func GetContractMiningActionRankingList(arrData GetContractMiningActionRankingListStruct) (interface{}, string) {
	date := base.TimeFormat(base.GetCurrentDateTimeT().AddDate(0, 0, -1), "2006-01-02")
	if arrData.Date != "" {
		date = arrData.Date
	}

	sysLangCode, _ := models.GetLanguage(arrData.LangCode)
	if sysLangCode == nil || sysLangCode.ID == "" {
		return nil, "something_is_wrong"
	}

	arrContractMiningRankingList := make([]ContractMiningActionRankingList, 0)

	arrTblBonusSponsorList, _ := models.GetContractMiningActionRankingListFn(date, arrData.MaxNumber, false)
	if len(arrTblBonusSponsorList) > 0 {
		for _, arrTblBonusSponsorListV := range arrTblBonusSponsorList {
			maskNum := 4
			if len(arrTblBonusSponsorListV.Username) <= 4 {
				maskNum = 2
			}

			arrContractMiningRankingList = append(arrContractMiningRankingList,
				ContractMiningActionRankingList{
					Username:   helpers.MaskLeft(arrTblBonusSponsorListV.Username, maskNum),
					PoolAmount: arrTblBonusSponsorListV.FBns,
					// Date:       string(arrTblBonusSponsorListV.TBnsID),
				},
			)
		}
	}

	arrSysGhostSponsorPool, _ := models.GetSysGhostSponsorPoolListFn(date, arrData.MaxNumber, false)
	if len(arrSysGhostSponsorPool) > 0 {
		for _, arrSysGhostSponsorPoolV := range arrSysGhostSponsorPool {
			maskNum := 4
			if len(arrSysGhostSponsorPoolV.Username) <= 4 {
				maskNum = 2
			}

			arrContractMiningRankingList = append(arrContractMiningRankingList,
				ContractMiningActionRankingList{
					Username:   helpers.MaskLeft(arrSysGhostSponsorPoolV.Username, maskNum),
					PoolAmount: arrSysGhostSponsorPoolV.TotalPoolAmount,
					// Date:       string(arrSysGhostSponsorPoolV.BnsDate),
				},
			)
		}
	}

	//sort array
	sort.Slice(arrContractMiningRankingList, func(p, q int) bool {
		return arrContractMiningRankingList[q].PoolAmount < arrContractMiningRankingList[p].PoolAmount
	})

	// convert to []interface{} for pagination + limit return data to top x rank
	var counter = 0
	var arrListingData []interface{}
	if len(arrContractMiningRankingList) > 0 {
		for _, arrContractMiningRankingListV := range arrContractMiningRankingList {
			arrListingData = append(arrListingData,
				map[string]string{
					"username":    arrContractMiningRankingListV.Username,
					"pool_amount": helpers.CutOffDecimal(arrContractMiningRankingListV.PoolAmount, 2, ".", ","),
				},
			)

			counter++
			if counter >= arrData.MaxNumber {
				break
			}
		}
	}

	page := base.Pagination{
		Page:    arrData.Page,
		DataArr: arrListingData,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, ""
}

type RewardHistoryPostStruct struct {
	MemberID int    `json:"member_id"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
	Page     int64  `json:"page"`
	LangCode string `json:"lang_code"`
}

type BonusHistoryStruct struct {
	Date   string `json:"date"`
	Bonus  string `json:"bonus"`
	Amount string `json:"amount"`
}

func (r *RewardHistoryPostStruct) RewardHistory() (interface{}, error) {

	var arrRewardHistory []interface{}

	var (
		decimalPoint uint
	)

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ewt_setup.ewallet_type_code = ?", CondValue: "USDT"},
	)
	ewtSetup, _ := models.GetEwtSetupFn(arrCond, "", false)
	if ewtSetup != nil {
		decimalPoint = uint(ewtSetup.DecimalPoint)
	} else {
		decimalPoint = uint(2)
	}

	if r.DateFrom == "" && r.DateTo == "" {
		r.DateFrom = base.GetCurrentDateTimeT().AddDate(0, 0, -30).Format("2006-01-02")
		r.DateTo = base.GetCurrentDateTimeT().AddDate(0, 0, -1).Format("2006-01-02") //yesterday date
	}

	//get tbl_bonus_payout
	rst, err := models.GetBonusPayoutByMemId(r.MemberID, r.DateFrom, r.DateTo)

	if err != nil {
		base.LogErrorLog("RewardHistory - fail to get Bonus Data", err, r, true)
		return nil, err
	}

	if len(rst) > 0 {
		for _, v := range rst {

			TotalBns := helpers.CutOffDecimal(v.PaidAmount, decimalPoint, ".", ",")
			BnsType := helpers.TransRemark(v.Remark, r.LangCode)
			CurrencyCode := helpers.Translate(v.PaidEwallet, r.LangCode)

			arrRewardHistory = append(arrRewardHistory, BonusHistoryStruct{
				Date:   v.TBnsId,
				Bonus:  BnsType,
				Amount: TotalBns + " " + CurrencyCode,
			})
		}
	}

	//sort array
	// sort.Slice(arrRewardStatement, func(p, q int) bool {
	// 	return arrRewardStatement[q].Date < arrRewardStatement[p].Date
	// })

	page := base.Pagination{
		Page:    r.Page,
		DataArr: arrRewardHistory,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, nil
}

func GetMemberBonusTier(memID int, bns_type string) (string, error) {

	var (
		tier       string = "-"
		totalSales float64
	)

	arrStatementListSetting, err := models.GetSysGeneralSetupByID("reward_type_list_setting")

	if err != nil {
		base.LogErrorLog("GetMemberBonusTier - fail to get reward_type_list_setting", err, "", true)
		return "", err
	}

	type arrSponsorBns struct {
		Balance float64 `json:"balance"`
		Tier    string  `json:"tier"`
	}
	var spnBnsSetting map[string][]arrSponsorBns
	json.Unmarshal([]byte(arrStatementListSetting.InputType2), &spnBnsSetting)

	spnSales, err := models.GetTotalSponsorBonusSalesByMemberId(memID, bns_type)

	if err != nil {
		base.LogErrorLog("GetMemberBonusTier - GetTotalSponsorBonusSalesByMemberId failed", err, map[string]interface{}{"memID": memID, "bns_type": bns_type}, true)
		return "", err
	}

	totalSales = spnSales.TotalSales

	if bns_type == "A" {
		for _, v1 := range spnBnsSetting["sponsor_a"] {
			if totalSales >= v1.Balance {
				tier = v1.Tier
				break
			}
		}
	}

	if bns_type == "B" {
		for _, v1 := range spnBnsSetting["sponsor_b"] {
			if totalSales >= v1.Balance {
				tier = v1.Tier
				break
			}
		}
	}

	return tier, nil
}

type RewardGraphPostStruct struct {
	MemberID int    `json:"member_id"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
	LangCode string `json:"lang_code"`
	RwdType  string `json:"rwd_type"`
	Type     string `json:"type"`
}

func (r *RewardGraphPostStruct) RewardGraph() (interface{}, error) {
	var (
		arrReturn []interface{}
	)

	//group by type to return different date
	switch strings.ToUpper(r.Type) {
	case "DAILY":
		bnsArr, err := models.GetMemberTotalBnsList(r.MemberID, strings.ToUpper(r.RwdType))
		if err != nil {
			base.LogErrorLog("RewardGraph - fail to get GetMemberTotalBnsList", err, r, true)
			return nil, err
		}

		// get days in current month
		arrDays := helpers.GetDaysInCurrentMonth()
		for _, arrDaysV := range arrDays {
			curDate := arrDaysV.Format("02/01/06")
			value := 0.00

			for _, bnsArrV := range bnsArr {
				dt, _ := time.Parse("2006-01-02", bnsArrV.TBnsID)
				bnsDate := dt.Format("02/01/06")

				if curDate == bnsDate {
					value = bnsArrV.TotalBonus
				}
			}

			arrReturn = append(arrReturn, map[string]interface{}{
				"date":  curDate,
				"value": helpers.CutOffDecimal(value, 2, ".", ""),
			})
		}

		// for _, v := range bnsArr {
		// 	dt, _ := time.Parse("2006-01-02", v.TBnsID)
		// 	date := dt.Format("02/01/06")

		// 	arrReturn = append(arrReturn, map[string]interface{}{
		// 		"date":  date,
		// 		"value": helpers.CutOffDecimal(v.TotalBonus, 2, ".", ""),
		// 	})
		// }
	case "WEEKLY":
		type WeeklyData struct {
			WeekStart string `json:"-"`
			Date      string `json:"date"`
			Value     string `json:"value"`
		}

		// get latest 15 weeks week start and end date
		arrLatestWeeks := helpers.GetLatestWeeks(15)
		for _, arrLatestWeeksV := range arrLatestWeeks {
			//convert from time.Time to string
			weekStStr := arrLatestWeeksV["week_start"].Format("2006-01-02")
			weekEndStr := arrLatestWeeksV["week_end"].Format("2006-01-02")

			//get sum from tbl_bonus
			sumRst, _ := models.GetMemberSumBnsByBnsType(r.MemberID, strings.ToUpper(r.RwdType), weekStStr, weekEndStr)

			weekStStr2 := arrLatestWeeksV["week_start"].Format("02/01/06")
			weekEndStr2 := arrLatestWeeksV["week_end"].Format("02/01/06")
			arrReturn = append(arrReturn, WeeklyData{
				WeekStart: arrLatestWeeksV["week_start"].Format("2006-01-02"),
				Date:      weekStStr2 + "-" + weekEndStr2,
				Value:     helpers.CutOffDecimal(sumRst.TotalBonus, 2, ".", ""),
			})
		}

		// //grab latest start & end date from tbl_bonus
		// arrCond := make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: "t_member_id = ?", CondValue: r.MemberID},
		// )
		// rstArr, err := models.GetTblBonusFn(arrCond, "MAX(t_bns_id) as bns_id_max, MIN(t_bns_id) as bns_id_min", false)
		// if err != nil {
		// 	base.LogErrorLog("RewardGraph - fail to get GetMemberTotalBnsWeek", err, r, true)
		// 	return nil, err
		// }
		// if len(rstArr) > 0 {
		// 	if rstArr[0].BnsIDMin != "" && rstArr[0].BnsIDMax != "" {
		// 		//convert string to time format
		// 		dtMin, _ := time.Parse("2006-01-02", rstArr[0].BnsIDMin)
		// 		dtMax, _ := time.Parse("2006-01-02", rstArr[0].BnsIDMax)

		// 		//get date start & end from
		// 		fmt.Println("dtMin:", dtMin, "dtMax:", dtMax)
		// 		var arrWeeks = helpers.GetWeekStartAndEndDatesWithinDateRange(dtMin, dtMax)
		// 		if len(arrWeeks) > 0 {
		// 			for _, val := range arrWeeks {

		// 				//convert from time.Time to string
		// 				weekStStr := val["week_start"].Format("2006-01-02")
		// 				weekEndStr := val["week_end"].Format("2006-01-02")

		// 				//get sum from tbl_bonus
		// 				sumRst, _ := models.GetMemberSumBnsByBnsType(r.MemberID, strings.ToUpper(r.RwdType), weekStStr, weekEndStr)

		// 				weekStStr2 := val["week_start"].Format("02/01/06")
		// 				weekEndStr2 := val["week_end"].Format("02/01/06")
		// 				arrReturn = append(arrReturn, WeeklyData{
		// 					WeekStart: val["week_start"].Format("2006-01-02"),
		// 					Date:      weekStStr2 + "-" + weekEndStr2,
		// 					Value:     helpers.CutOffDecimal(sumRst.TotalBonus, 2, ".", ""),
		// 				})
		// 			}

		// 			//sort desc
		// 			sort.Slice(arrReturn, func(i, j int) bool {
		// 				commonID1 := reflect.ValueOf(arrReturn[j]).FieldByName("WeekStart").String()
		// 				commonID2 := reflect.ValueOf(arrReturn[i]).FieldByName("WeekStart").String()
		// 				return commonID1 > commonID2
		// 			})

		// 		}
		// }
		// }
	case "MONTHLY":
		bnsArr, err := models.GetMemberTotalBnsTimeFrameList(r.MemberID, strings.ToUpper(r.RwdType), "MONTHLY")
		if err != nil {
			base.LogErrorLog("RewardGraph - fail to get GetMemberTotalBnsMonthList", err, r, true)
			return nil, err
		}

		// get latest 12 months month start and end date
		arrLatestMonths := helpers.GetLatestMonths(12)
		for _, arrLatestMonthsV := range arrLatestMonths {
			curMonth := arrLatestMonthsV["month_start"].Format("1/06")
			value := 0.00

			for _, v := range bnsArr {
				yr, _ := time.Parse("2006", v.Year)
				year := yr.Format("06")

				if curMonth == v.Month+"/"+year {
					value = v.TotalBonus
				}
			}

			arrReturn = append(arrReturn, map[string]interface{}{
				"date":  curMonth,
				"value": helpers.CutOffDecimal(value, 2, ".", ""),
			})
		}

		// for _, v := range bnsArr {
		// 	yr, _ := time.Parse("2006", v.Year)
		// 	year := yr.Format("06")

		// 	arrReturn = append(arrReturn, map[string]interface{}{
		// 		"date":  v.Month + "/" + year,
		// 		"value": helpers.CutOffDecimal(v.TotalBonus, 2, ".", ""),
		// 	})
		// }
	case "YEARLY":
		bnsArr, err := models.GetMemberTotalBnsTimeFrameList(r.MemberID, strings.ToUpper(r.RwdType), "YEARLY")
		if err != nil {
			base.LogErrorLog("RewardGraph - fail to get GetMemberTotalBnsYearList", err, r, true)
			return nil, err
		}

		// get latest 5 years year start and end date
		arrLatestYears := helpers.GetLatestYears(5)
		for _, arrLatestYearsV := range arrLatestYears {
			curYear := arrLatestYearsV["year_start"].Format("2006")
			value := 0.00

			for _, v := range bnsArr {
				if curYear == v.TBnsID {
					value = v.TotalBonus
				}
			}

			arrReturn = append(arrReturn, map[string]interface{}{
				"date":  curYear,
				"value": helpers.CutOffDecimal(value, 2, ".", ""),
			})
		}

		// for _, v := range bnsArr {
		// 	arrReturn = append(arrReturn, map[string]interface{}{
		// 		"date":  v.TBnsID,
		// 		"value": helpers.CutOffDecimal(v.TotalBonus, 2, ".", ""),
		// 	})
		// }
	}

	return arrReturn, nil
}
