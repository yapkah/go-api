package member_service

import (
	"net/http"
	"time"

	"github.com/yapkah/go-api/helpers"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/e"
)

// MemberTreeStruct struct
type MemberTreeReturnRstStruct struct {
	Username       string                      `json:"username"`
	ProfileImgURL  string                      `json:"profile_img_url"`
	TotalStar      int                         `json:"total_star"`
	TotalDiamond   int                         `json:"total_diamond"`
	ReferralBy     string                      `json:"referral_by"`
	DateJoined     string                      `json:"date_joined"`
	ChildrenStatus int                         `json:"children_status"`
	ChildrenList   []MemberTreeReturnRstStruct `json:"children_list"`
}

// MemberTreeFormStruct struct
type MemberTreeFormStruct struct {
	Level            int    `form:"level" json:"level"`
	DownlineUsername string `form:"downline_username" json:"downline_username" valid:"MaxSize(25)"`
	IncMem           int    `form:"inc_mem" json:"inc_mem"`
	IncDownMem       int    `form:"inc_down_mem" json:"inc_down_mem"`
	DataType         string `form:"data_type" json:"data_type"`
}

// ArrMemberTreeDataStruct struct
type ArrMemberTreeDataStruct struct {
	Level            int    `form:"level" json:"level"`
	DownlineMemberID int    `form:"downline_member_id" json:"downline_member_id"`
	IncMem           int    `form:"inc_mem" json:"inc_mem"`
	IncDownMem       int    `form:"inc_down_mem" json:"inc_down_mem"`
	EntMemberID      int    `form:"ent_member_id" json:"ent_member_id"`
	DataType         string `form:"data_type" json:"data_type"`
}

// GetMemberTreev1 func
func GetMemberTreev1(ArrMemberTreeDataStruct) ([]MemberTreeReturnRstStruct, error) {
	arrDataReturn := make([]MemberTreeReturnRstStruct, 0)
	arrEmpty := make([]MemberTreeReturnRstStruct, 0)

	arrChildrenList := make([]MemberTreeReturnRstStruct, 0)
	arrChildrenList2 := make([]MemberTreeReturnRstStruct, 0)
	arrSubChildrenList1 := make([]MemberTreeReturnRstStruct, 0)
	arrSubChildrenList2 := make([]MemberTreeReturnRstStruct, 0)
	// models.GetMember

	arrChildrenList1 := append(arrChildrenList,
		MemberTreeReturnRstStruct{Username: "aming1", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar4.jpg", TotalStar: 0, TotalDiamond: 0, ReferralBy: "aming", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "aming2", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar5.jpg", TotalStar: 0, TotalDiamond: 0, ReferralBy: "aming", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "aming3", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar6.jpg", TotalStar: 0, TotalDiamond: 0, ReferralBy: "aming", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
	)

	arrSubChildrenList2 = append(arrSubChildrenList2,
		MemberTreeReturnRstStruct{Username: "abu331", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar7.jpg", TotalStar: 0, TotalDiamond: 0, ReferralBy: "abu33", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "abu332", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar8.jpg", TotalStar: 5, TotalDiamond: 2, ReferralBy: "abu33", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "abu333", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar9.jpg", TotalStar: 5, TotalDiamond: 8, ReferralBy: "abu33", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
	)
	arrSubChildrenList1 = append(arrSubChildrenList1,
		MemberTreeReturnRstStruct{Username: "abu31", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar4.jpg", TotalStar: 3, TotalDiamond: 0, ReferralBy: "abu3", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "abu32", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar5.jpg", TotalStar: 5, TotalDiamond: 1, ReferralBy: "abu3", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "abu33", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar6.jpg", TotalStar: 5, TotalDiamond: 6, ReferralBy: "abu3", DateJoined: "27-10-2020", ChildrenStatus: 1, ChildrenList: arrSubChildrenList2},
	)
	arrChildrenList2 = append(arrChildrenList2,
		MemberTreeReturnRstStruct{Username: "abu1", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar4.jpg", TotalStar: 2, TotalDiamond: 0, ReferralBy: "abu", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "abu2", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar5.jpg", TotalStar: 1, TotalDiamond: 0, ReferralBy: "abu", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "abu3", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar6.jpg", TotalStar: 5, TotalDiamond: 5, ReferralBy: "abu", DateJoined: "27-10-2020", ChildrenStatus: 1, ChildrenList: arrSubChildrenList1},
	)

	arrDataReturn = append(arrDataReturn,
		MemberTreeReturnRstStruct{Username: "ali", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar1.jpg", TotalStar: 3, TotalDiamond: 0, ReferralBy: "COM", DateJoined: "27-10-2020", ChildrenStatus: 0, ChildrenList: arrEmpty},
		MemberTreeReturnRstStruct{Username: "aming", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar2.jpg", TotalStar: 5, TotalDiamond: 10, ReferralBy: "COM", DateJoined: "27-10-2020", ChildrenStatus: 1, ChildrenList: arrChildrenList1},
		MemberTreeReturnRstStruct{Username: "abu", ProfileImgURL: "https://media02.securelayers.cloud/medias/WOD/AVATAR/DEFAULT/avatar3.jpg", TotalStar: 5, TotalDiamond: 7, ReferralBy: "COM", DateJoined: "27-10-2020", ChildrenStatus: 1, ChildrenList: arrChildrenList2},
	)
	return arrDataReturn, nil
}

// ArrMemberTreeReturnRstStruct struct
type ArrMemberTreeReturnRstStruct struct {
	TotalDirectSponsor string                     `json:"total_direct_sponsor"`
	TotalLevel         string                     `json:"total_level"`
	TotalNetwork       int                        `json:"total_network"`
	DownlineList       []*models.MemberTreeStruct `json:"downline_list"`
}

func GetMemberTreev2(arrData ArrMemberTreeDataStruct, langCode string) (*ArrMemberTreeReturnRstStruct, error) {
	var arrDataReturn ArrMemberTreeReturnRstStruct

	memberID := arrData.EntMemberID

	if arrData.DownlineMemberID > 0 {
		memberID = arrData.DownlineMemberID
	}

	arrDownline := GetDownlineMemv1(arrData.EntMemberID, memberID, false, arrData.DataType, langCode)

	arrPersonalData := make([]*models.MemberTreeStruct, 0)
	if arrData.IncMem == 1 {
		arrPersonalData = GetDownlineMemv1(arrData.EntMemberID, arrData.EntMemberID, true, arrData.DataType, langCode)
		// fmt.Println("arrPersonalData:", arrPersonalData)
	}

	arrDownLineData := make([]*models.MemberTreeStruct, 0)
	if arrData.IncDownMem == 1 {
		arrDownLineData = GetDownlineMemv1(arrData.EntMemberID, memberID, true, arrData.DataType, langCode)
		// fmt.Println("arrDownLineData", arrDownLineData)
	}

	if arrData.IncMem == 1 && arrData.IncDownMem == 1 {
		arrDownLineData[0].ChildrenList = arrDownline
		arrPersonalData[0].ChildrenList = arrDownLineData
		arrDataReturn.DownlineList = arrPersonalData
	} else if arrData.IncMem == 1 {
		arrPersonalData[0].ChildrenList = arrDownline
		arrDataReturn.DownlineList = arrPersonalData
	} else if arrData.IncDownMem == 1 {
		arrDownLineData[0].ChildrenList = arrDownline
		arrDataReturn.DownlineList = arrDownLineData
	} else {
		arrDataReturn.DownlineList = arrDownline
	}
	totalNetwork := 0
	// directSponsorStar := ""
	// totalDirectDownline := 0
	// directDownlineWithStar := 0

	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sponsor_lot.member_id = ? ", CondValue: memberID},
	)
	totalDownlineMemberRst, _ := models.GetTotalDownlineMemberFn(arrCond, false)
	if totalDownlineMemberRst.TotalDownline > 0 {
		totalNetwork = totalDownlineMemberRst.TotalDownline
	}

	// arrDirectDownline := GetDownlineMemv1(arrData.EntMemberID, false)
	// if len(arrDirectDownline) > 0 {
	// 	for _, arrDirectDownlineV := range arrDirectDownline {
	// 		if arrDirectDownlineV.DownlineTotalStar > 0 {
	// 			directDownlineWithStar++
	// 		}
	// 		totalDirectDownline++
	// 	}
	// }
	// directSponsorStar = strconv.Itoa(directDownlineWithStar) + " / " + strconv.Itoa(totalDirectDownline)

	// get total direct sponsor
	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: memberID},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	totalDirectSponsorRst, _ := models.GetTotalDirectSponsorFn(arrCond, false)
	totalDirectSponsor := "0"
	if totalDirectSponsorRst.TotalDirectSponsor > 0 {
		totalDirectSponsor = helpers.CutOffDecimal(totalDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
	}

	// get cur date
	// curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")

	// start get total direct downline sales
	// arrCond = make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: memberID},
	// 	models.WhereCondFn{Condition: "sls_master.status = ?", CondValue: "AP"},
	// 	models.WhereCondFn{Condition: "sls_master.action = ?", CondValue: "CONTRACT"},
	// 	models.WhereCondFn{Condition: " sls_master.total_bv > ? ", CondValue: 0},
	// )
	// totalDirectDownlineSalesRst, _ := models.GetTotalDirectSalesFn(arrCond, false)
	// totalLevel := totalDirectSponsor
	// if totalDirectSponsorRst.TotalDirectSponsor < 15 && totalDirectDownlineSalesRst.TotalBV > 10000 {
	// 	totalLevel = "15"
	// }
	// end get total direct downline sales

	arrDataReturn.TotalDirectSponsor = totalDirectSponsor
	// arrDataReturn.TotalLevel = totalLevel
	arrDataReturn.TotalNetwork = totalNetwork

	return &arrDataReturn, nil
}

func GetMemberTreev3(arrData ArrMemberTreeDataStruct) (*ArrMemberTreeReturnRstStruct, error) {
	memberID := arrData.EntMemberID
	level := arrData.Level

	downlineLvl := 0
	if arrData.DownlineMemberID > 0 {
		memberID = arrData.DownlineMemberID
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.DownlineMemberID},
		)
		downlineRoot, _ := models.GetEntMemberLotSponsorFn(arrCond, false)
		if len(downlineRoot) < 1 {
			return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member"}
		}
		downlineLvl = downlineRoot[0].Lvl
	}
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " member_id = ? ", CondValue: arrData.EntMemberID},
	)
	root, _ := models.GetEntMemberLotSponsorFn(arrCond, false)
	if len(root) < 1 {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member"}
	}

	// arrCond := make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	// models.WhereCondFn{Condition: " ent_member_tree_sponsor.member_id = ? ", CondValue: memberID}, // level according to the search member and login member
	// 	models.WhereCondFn{Condition: " ent_member_tree_sponsor.member_id = ? ", CondValue: arrData.EntMemberID}, // level according to the login member
	// )
	// root, _ := models.GetEntMemberEntMemberTreeSponsorFn(arrCond, false)
	// if root.MemberID < 1 {
	// 	return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: "invalid_member"}
	// }

	searchDownlineStatus := true
	memberStartLevel := 0
	rootLvl := root[0].Lvl
	targetLvl := rootLvl + level - downlineLvl
	if arrData.DownlineMemberID > 0 {
		targetLvl = rootLvl + level - downlineLvl + root[0].Lvl
		if rootLvl+level == downlineLvl {
			// fmt.Println("same level")
			// if member search level is equal to max of level return only member search
			searchDownlineStatus = false
			memberStartLevel = level
		} else {
			memberStartLevel = downlineLvl - rootLvl
		}
	}
	// fmt.Println("rootLvl", rootLvl)
	// fmt.Println("level", level)
	// fmt.Println("downlineLvl", downlineLvl)
	// fmt.Println("targetLvl", targetLvl)
	arrDownline := make([]*models.MemberTreeStruct, 0)
	if searchDownlineStatus {
		arrDownline = GetDownlineMemberByLayerv2(rootLvl, memberID, rootLvl, targetLvl+1)
	}

	arrPersonalData := make([]*models.MemberTreeStruct, 0)
	if arrData.IncMem == 1 {
		arrPersonalData = GetDownlineMemv2(1, arrData.EntMemberID, true)
		// fmt.Println("arrPersonalData:", arrPersonalData)
	}

	arrDownLineData := make([]*models.MemberTreeStruct, 0)
	if arrData.IncDownMem == 1 {
		// arrDownLineData = GetDownlineMemv2(1, memberID, true) // level according to the search member and login member
		arrDownLineData = GetDownlineMemv2(1, memberID, true) // level according to the login member
		// fmt.Println("arrDownLineData", arrDownLineData)
	}

	if arrData.IncMem == 1 && arrData.IncDownMem == 1 {
		if len(arrPersonalData) > 0 {
			arrPersonalData[0].ChildrenList = arrDownline
			arrDownline = arrPersonalData
		}
	} else if arrData.IncMem == 1 {
		if len(arrPersonalData) > 0 {
			arrPersonalData[0].ChildrenList = arrDownline
			arrDownline = arrPersonalData
		}
	} else if arrData.IncDownMem == 1 {
		if len(arrDownLineData) > 0 {
			arrDownLineData[0].ChildrenList = arrDownline
			arrDownline = arrDownLineData
		}
	}
	// fmt.Println("memberStartLevel:", memberStartLevel)
	arrDownline = SetFriendListLayerRecursively(arrDownline, memberStartLevel)

	totalNetwork := 0

	arrCond = make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " sponsor_lot.member_id = ? ", CondValue: memberID},
		// models.WhereCondFn{Condition: " downline_lot.i_lvl <= ? ", CondValue: rootLvl + level},
		models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
	)
	totalDownlineMemberRst, _ := models.GetTotalDownlineMemberFn(arrCond, false)
	if totalDownlineMemberRst.TotalDownline > 0 {
		totalNetwork = totalDownlineMemberRst.TotalDownline
	}

	// arrCond = make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: " sponsor_lot.member_id = ? ", CondValue: memberID},
	// 	// models.WhereCondFn{Condition: " downline_lot.i_lvl <= ? ", CondValue: rootLvl + level},
	// 	models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
	// )
	// totalDownlineMemberSalesRst, _ := models.GetTotalDownlineMemberSalesFn(arrCond, true)
	// totalPerformance := "0"
	// if totalDownlineMemberSalesRst.TotalSales > 0 {
	// 	totalPerformance = helpers.CutOffDecimal(totalDownlineMemberSalesRst.TotalSales, 2, ".", ",")
	// }
	// arrDirectDownline := GetDownlineMemv1(arrData.EntMemberID, false)
	// if len(arrDirectDownline) > 0 {
	// 	for _, arrDirectDownlineV := range arrDirectDownline {
	// 		if arrDirectDownlineV.DownlineTotalStar > 0 {
	// 			directDownlineWithStar++
	// 		}
	// 		totalDirectDownline++
	// 	}
	// }
	// directSponsorStar = strconv.Itoa(directDownlineWithStar) + " / " + strconv.Itoa(totalDirectDownline)
	arrDataReturn := ArrMemberTreeReturnRstStruct{
		TotalNetwork: totalNetwork,
		// TotalPerformance: totalPerformance,
		DownlineList: arrDownline,
	}

	return &arrDataReturn, nil
}

func GetDownlineMemberByLayerv1(rootLvl int, sprMemID int, layer int, maxLayer int) []*models.MemberTreeStruct {
	arrEmpty := make([]*models.MemberTreeStruct, 0)
	for {
		if layer != maxLayer {
			arrChildren := GetDownlineMemv1(sprMemID, sprMemID, false, "", "en")
			layer++
			if len(arrChildren) > 0 {
				for arrChildrenK, arrChildrenV := range arrChildren {
					arrNewChildren := GetDownlineMemberByLayerv1(rootLvl, arrChildrenV.DownlineMemberID, layer, maxLayer)
					balLayer := maxLayer - layer
					if balLayer != 1 {
						arrChildren[arrChildrenK].ChildrenList = arrNewChildren
					} else {
						arrChildren[arrChildrenK].ChildrenList = arrEmpty
					}
				}
			}
			return arrChildren
		} else {
			return arrEmpty
		}
	}
}

func GetDownlineMemv1(rootMemberID, sprMemID int, incMem bool, dataType string, langCode string) []*models.MemberTreeStruct {
	arrCond := make([]models.WhereCondFn, 0)
	arrEmpty := make([]*models.MemberTreeStruct, 0)
	strSelectColumn := ""
	if incMem == true {
		// start get self data
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member_tree_sponsor.member_id = ? ", CondValue: sprMemID},
		)
	} else {
		// start get downline data
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member_tree_sponsor.sponsor_id = ? ", CondValue: sprMemID},
		)
	}
	result, _ := models.GetMemberTreeFn(arrCond, strSelectColumn, false)

	var rootLvl = 0
	arrEntMemberTreeRootFn := make([]models.WhereCondFn, 0)
	arrEntMemberTreeRootFn = append(arrEntMemberTreeRootFn,
		models.WhereCondFn{Condition: " ent_member_tree_sponsor.member_id = ? ", CondValue: rootMemberID},
	)

	arrEntMemberTreeRoot, err := models.GetEntMemberEntMemberTreeSponsorFn(arrEntMemberTreeRootFn, false)
	if err == nil {
		rootLvl = arrEntMemberTreeRoot.Lvl
	}

	if len(result) > 0 {
		// env := setting.Cfg.Section("server").Key("ENV").String()

		// curDate, _ := base.GetCurrentTimeV2("yyyy-mm-dd")
		translatedRank := helpers.TranslateV2("rank", langCode, nil)
		translatedPackageTypeTier := helpers.TranslateV2("package_type_tier", langCode, nil)
		translatedPackageAmount := helpers.TranslateV2("package_value", langCode, nil)
		translatedJoinedDate := helpers.TranslateV2("joined_date", langCode, nil)
		translatedCountry := helpers.TranslateV2("country", langCode, nil)
		translatedTotalSales := helpers.TranslateV2("total_sales", langCode, nil)
		translatedTotalSalesMonth := helpers.TranslateV2("this_month_total_sales", langCode, nil)
		translatedTotalSalesToday := helpers.TranslateV2("today_total_sales", langCode, nil)
		translatedCurrentDepositAmount := helpers.TranslateV2("current_deposit_amount", langCode, nil)
		translatedConnectedExchange := helpers.TranslateV2("connected_exchange", langCode, nil)

		for _, resultV := range result {
			// get total direct sponsor
			// arrCond = make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: "ent_member_tree_sponsor.sponsor_id = ?", CondValue: resultV.DownlineMemberID},
			// 	models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
			// )
			// totalDirectSponsorRst, _ := models.GetTotalDirectSponsorFn(arrCond, false)
			// totalDirectSponsor := "0"
			// if totalDirectSponsorRst.TotalDirectSponsor > 0 {
			// 	totalDirectSponsor = helpers.CutOffDecimal(totalDirectSponsorRst.TotalDirectSponsor, 0, ".", ",")
			// }

			// get total network
			// arrCond := make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: " sponsor_lot.member_id = ? ", CondValue: resultV.DownlineMemberID},
			// )
			// totalDownlineMemberRst, _ := models.GetTotalDownlineMemberFn(arrCond, false)
			// totalDirectSponsor := "0"
			// if totalDownlineMemberRst.TotalDownline > 0 {
			// 	totalDirectSponsor = fmt.Sprintf("%d", totalDownlineMemberRst.TotalDownline)
			// }

			// resultV.TotalDirectSponsor = totalDirectSponsor

			// start get today network total sales
			// arrCond = make([]models.WhereCondFn, 0)
			// arrCond = append(arrCond,
			// 	models.WhereCondFn{Condition: " sponsor_lot.member_id = ? ", CondValue: resultV.DownlineMemberID},
			// 	models.WhereCondFn{Condition: " ent_member.status = ? ", CondValue: "A"},
			// 	models.WhereCondFn{Condition: " sls_master.doc_date = ? ", CondValue: curDate},
			// 	models.WhereCondFn{Condition: " sls_master.status IN ('AP', 'EP') AND sls_master.action = ? ", CondValue: "CONTRACT"},
			// 	models.WhereCondFn{Condition: " sls_master.grp_type != ? ", CondValue: 1},
			// 	models.WhereCondFn{Condition: " sls_master.total_bv > ? ", CondValue: 0},
			// )
			// todayTotalNetworkSalesRst, _ := models.GetTotalNetworkSalesFn(arrCond, false)
			// todayTotalNetworkSales := "0"
			// if todayTotalNetworkSalesRst.TotalSales > 0 {
			// 	todayTotalNetworkSales = helpers.CutOffDecimal(todayTotalNetworkSalesRst.TotalSales, 2, ".", ",")
			// }
			// end get today network total sales

			arrExtraList := make([]models.ExtraListStruct, 0)

			if dataType == "TYPEA" {
				// get current deposit
				currentDeposit := "0"

				// get trading deposit wallet id by ewallet_type_code
				arrEwtSetupFn := make([]models.WhereCondFn, 0)
				arrEwtSetupFn = append(arrEwtSetupFn,
					models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: "TD"},
					models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
				)
				arrEwtSetup, _ := models.GetEwtSetupFn(arrEwtSetupFn, "", false)
				if arrEwtSetup != nil {
					arrEwtSummaryFn := make([]models.WhereCondFn, 0)
					arrEwtSummaryFn = append(arrEwtSummaryFn,
						models.WhereCondFn{Condition: "ewt_summary.member_id = ?", CondValue: resultV.DownlineMemberID},
						models.WhereCondFn{Condition: "ewt_summary.ewallet_type_id = ?", CondValue: arrEwtSetup.ID},
					)

					arrEwtSummary, _ := models.GetEwtSummaryFn(arrEwtSummaryFn, "", false)
					if len(arrEwtSummary) > 0 {
						currentDeposit = helpers.CutOffDecimal(arrEwtSummary[0].Balance, 2, ".", ",")
					}
				}

				arrExtraList = append(arrExtraList,
					models.ExtraListStruct{Key: "JD", TranslatedLabel: translatedJoinedDate, LabelValue: resultV.DownlineJoinDate},
					models.ExtraListStruct{Key: "CT", TranslatedLabel: translatedCountry, LabelValue: helpers.TranslateV2(resultV.DownlineCountry, langCode, map[string]string{})},
					models.ExtraListStruct{Key: "TD", TranslatedLabel: translatedCurrentDepositAmount, LabelValue: currentDeposit},
					models.ExtraListStruct{Key: "EXC", TranslatedLabel: translatedConnectedExchange, LabelValue: helpers.TranslateV2("binance", langCode, map[string]string{})},
				)
			} else {
				// get member tier
				v, err := GetMemberTier(resultV.DownlineMemberID)
				tier := "-"
				if err == nil && v != "" {
					tier = v
				}

				// get member package rank
				packageRank := GetMemberCurRank(resultV.DownlineMemberID, langCode)
				if packageRank == "" {
					packageRank = "-"
				}

				// get total sales
				arrMemberTotalSalesFn := make([]models.WhereCondFn, 0)
				arrMemberTotalSalesFn = append(arrMemberTotalSalesFn,
					models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: resultV.DownlineMemberID},
					models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
					models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
				)
				arrMemberTotalSales, _ := models.GetMemberTotalSalesFn(arrMemberTotalSalesFn, false)

				totalSales := helpers.CutOffDecimal(arrMemberTotalSales.TotalAmount, 0, ".", ",")

				// get this month total sales
				now := time.Now()
				currentYear, currentMonth, _ := now.Date()
				currentLocation := now.Location()

				firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
				lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

				arrMemberMonthTotalSalesFn := make([]models.WhereCondFn, 0)
				arrMemberMonthTotalSalesFn = append(arrMemberMonthTotalSalesFn,
					models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: resultV.DownlineMemberID},
					models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
					models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
					models.WhereCondFn{Condition: "sls_master.doc_date >= ? ", CondValue: firstOfMonth.Format("2006-01-02")},
					models.WhereCondFn{Condition: "sls_master.doc_date <= ? ", CondValue: lastOfMonth.Format("2006-01-02")},
				)
				arrMemberMonthTotalSales, _ := models.GetMemberTotalSalesFn(arrMemberMonthTotalSalesFn, false)

				totalSalesMonth := helpers.CutOffDecimal(arrMemberMonthTotalSales.TotalAmount, 0, ".", ",")

				// get today total sales
				arrMemberTodayTotalSalesFn := make([]models.WhereCondFn, 0)
				arrMemberTodayTotalSalesFn = append(arrMemberTodayTotalSalesFn,
					models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: resultV.DownlineMemberID},
					models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
					models.WhereCondFn{Condition: "sls_master.status IN(?,'EP') ", CondValue: "AP"},
					models.WhereCondFn{Condition: "sls_master.doc_date = CURDATE() AND 1=? ", CondValue: "1"},
				)
				arrMemberTodayTotalSales, _ := models.GetMemberTotalSalesFn(arrMemberTodayTotalSalesFn, false)

				totalSalesToday := helpers.CutOffDecimal(arrMemberTodayTotalSales.TotalAmount, 0, ".", ",")

				arrExtraList = append(arrExtraList,
					models.ExtraListStruct{Key: "PACKAGE_RANK", TranslatedLabel: translatedRank, LabelValue: packageRank},
					models.ExtraListStruct{Key: "TIER", TranslatedLabel: translatedPackageTypeTier, LabelValue: tier},
					models.ExtraListStruct{Key: "PV", TranslatedLabel: translatedPackageAmount, LabelValue: totalSales},
					models.ExtraListStruct{Key: "JD", TranslatedLabel: translatedJoinedDate, LabelValue: resultV.DownlineJoinDate},
					models.ExtraListStruct{Key: "CT", TranslatedLabel: translatedCountry, LabelValue: helpers.TranslateV2(resultV.DownlineCountry, langCode, map[string]string{})},
					models.ExtraListStruct{Key: "TPV", TranslatedLabel: translatedTotalSalesToday, LabelValue: totalSalesToday},
					models.ExtraListStruct{Key: "MPV", TranslatedLabel: translatedTotalSalesMonth, LabelValue: totalSalesMonth},
					models.ExtraListStruct{Key: "PV", TranslatedLabel: translatedTotalSales, LabelValue: totalSales},
				)
			}

			resultV.ExtraList = arrExtraList
			resultV.ChildrenList = make([]*models.MemberTreeStruct, 0)
		}

		for _, resultV := range result {
			resultV.Level = resultV.Level - rootLvl
		}

		return result
	}

	return arrEmpty
}

func GetDownlineMemberByLayerv2(rootLvl int, sprMemID int, layer int, maxLayer int) []*models.MemberTreeStruct {
	arrEmpty := make([]*models.MemberTreeStruct, 0)
	for {
		if layer != maxLayer {
			arrChildren := GetDownlineMemv2(rootLvl, sprMemID, false)
			layer++
			if len(arrChildren) > 0 {
				for arrChildrenK, arrChildrenV := range arrChildren {
					arrNewChildren := GetDownlineMemberByLayerv1(rootLvl, arrChildrenV.DownlineMemberID, layer, maxLayer)
					balLayer := maxLayer - layer
					if balLayer != 1 {
						arrChildren[arrChildrenK].ChildrenList = arrNewChildren
					} else {
						arrChildren[arrChildrenK].ChildrenList = arrEmpty
					}
				}
			}
			return arrChildren
		} else {
			return arrEmpty
		}
	}
}

func GetDownlineMemv2(rootLvl, sprMemID int, incMem bool) []*models.MemberTreeStruct {
	arrCond := make([]models.WhereCondFn, 0)
	arrEmpty := make([]*models.MemberTreeStruct, 0)
	strSelectColumn := ""
	if incMem == true {
		// start get self data
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member_tree_sponsor.member_id = ? ", CondValue: sprMemID},
		)
		strSelectColumn = " , 0 AS 'level' "
	} else {
		// start get downline data
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member_tree_sponsor.sponsor_id = ? ", CondValue: sprMemID},
		)
		strSelectColumn = " , 0 AS 'level' "
	}
	result, _ := models.GetMemberTreeFn(arrCond, strSelectColumn, false)
	if len(result) > 0 {
		return result
	}
	return arrEmpty
}

func SetFriendListLayerRecursively(arrData []*models.MemberTreeStruct, level int) []*models.MemberTreeStruct {
	// is in base array?
	if len(arrData) > 0 {
		for k1, _ := range arrData {
			// fmt.Println("set: ", level, arrData[k1].DownlineNickName)
			arrData[k1].Level = level
		}
	}

	// check arrays contained in this array
	if len(arrData) > 0 {
		for k2, v2 := range arrData {
			if len(v2.ChildrenList) > 0 {
				level = arrData[k2].Level
				level++
				arrData[k2].ChildrenList = SetFriendListLayerRecursively(v2.ChildrenList, level)
			} else {
				// arrData[k2].ChildrenStatus = 0
			}
		}
	}

	return arrData
}

// func CheckSponsorMember
func CheckSponsorMember(sponsorID, downlineID int) bool {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " ent_member.id = ? ", CondValue: downlineID},
	)
	arrEntMemberTreeSponsor, _ := models.GetEntMemberEntMemberTreeSponsorFn(arrCond, false)

	// var count int
	for {
		// fmt.Println("count: ", count, sponsorID, arrEntMemberTreeSponsor.SponsorID)
		if arrEntMemberTreeSponsor.SponsorID != 0 {
			if sponsorID != arrEntMemberTreeSponsor.SponsorID {
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ent_member.id = ? ", CondValue: arrEntMemberTreeSponsor.SponsorID},
				)
				arrEntMemberTreeSponsor, _ = models.GetEntMemberEntMemberTreeSponsorFn(arrCond, false)
			} else {
				return true
			}
			// if count == 10 {
			// 	return false
			// }
			// count++
		} else {
			return false
		}

	}
}
