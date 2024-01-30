package event_service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
)

type MemberEventListStruct struct {
	ID         int    `json:"id"`
	EventTitle string `json:"event_title"`
	EventDesc  string `json:"event_desc"`
}

type MemberEventStruct struct {
	MemberID int
	LangCode string
}

// func GetMemberEventListv1
func GetMemberEventListv1(arrData MemberEventStruct) []MemberEventListStruct {

	arrDataReturn := make([]MemberEventListStruct, 0)

	settingID := "jackpot_ticket_setting"
	arrTicketSetting, _ := models.GetSysGeneralSetupByID(settingID)
	if arrTicketSetting.InputType1 == "1" {
		// totalJackpostTicket := "0"

		// arrCond = make([]models.WhereCondFn, 0)
		// arrCond = append(arrCond,
		// 	models.WhereCondFn{Condition: " wod_promo_event.member_id = ? ", CondValue: arrData.EntMemberID},
		// )
		// arrTotalWodPromoTicketBalRst, _ := models.GetTotalBalanceWodPromoEvent(arrCond, false)
		// if len(arrTotalWodPromoTicketBalRst) > 0 && arrTotalWodPromoTicketBalRst[0].TotalBalance > 0 {
		// 	totalJackpostTicket = strconv.Itoa(arrTotalWodPromoTicketBalRst[0].TotalBalance)
		// }

		// arrDataReturn.TotalJackpotTicket = totalJackpostTicket
		// params := make(map[string]string)
		// translatedJackpotEventTitle := helpers.TranslateV2("jackpot_event_title", arrData.LangCode, params)
		// translatedJackpotEventDesc := helpers.TranslateV2("jackpot_event_desc", arrData.LangCode, params)
		// arrDataReturn = append(arrDataReturn,
		// 	MemberEventListStruct{ID: len(arrDataReturn) + 1, EventTitle: translatedJackpotEventTitle, },
		// )
	}
	// hour := "24"
	// // filterSql := "AND IF(downline.d_last_game IS NOT NULL AND downline.d_last_game != '', downline.d_last_game <= DATE_SUB(NOW(), INTERVAL " + hour + " HOUR), downline.created_at <= DATE_SUB(NOW(), INTERVAL " + hour + " HOUR)) "
	// filterSql := "AND  downline.d_last_game <= DATE_SUB(NOW(), INTERVAL " + hour + " HOUR) "
	// arrCond := make([]models.WhereCondFn, 0)
	// arrCond = append(arrCond,
	// 	models.WhereCondFn{Condition: " sponsor.id = ? " + filterSql, CondValue: arrData.MemberID},
	// )
	// arrDownlineLastPlayGame, _ := models.getWodPromoEvent(arrCond, true)
	// fmt.Println("arrDownlineLastPlayGame:", arrDownlineLastPlayGame)
	// if len(arrDownlineLastPlayGame) > 0 {
	// 	if arrDownlineLastPlayGame[0].DownlineList != "" {
	// 		params := make(map[string]string)
	// 		params["downlinelist"] = arrDownlineLastPlayGame[0].DownlineList
	// 		translatedWord := helpers.TranslateV2("this_account_:downlinelist_no_play_game", arrData.LangCode, params)

	// 	}
	// }
	// arrDataReturn = append(arrDataReturn,
	// 	MemberNotificationPopUpListStruct{ID: len(arrDataReturn) + 1, Message: "testing hard code data"},
	// )

	return arrDataReturn
}

// GetEventSponsorRankingSetting func
func GetEventSponsorRankingSetting(eventType string, batchNo int, langCode string) (interface{}, string) {
	// get event setting by type and batch_no
	var arrGetSysEventsFn = make([]models.WhereCondFn, 0)

	if eventType == "ALL" {
		arrGetSysEventsFn = append(arrGetSysEventsFn, models.WhereCondFn{Condition: " type = ? ", CondValue: "TS"})
	} else if eventType == "ND" { // new downline
		arrGetSysEventsFn = append(arrGetSysEventsFn, models.WhereCondFn{Condition: " type = ? ", CondValue: "TS_ND"})
	} else {
		return nil, "invalid_type"
	}

	arrGetSysEventsFn = append(arrGetSysEventsFn, models.WhereCondFn{Condition: " batch_no = ? ", CondValue: batchNo})

	arrGetSysEvents, _ := models.GetSysEventsFn(arrGetSysEventsFn, "", false)
	if len(arrGetSysEvents) <= 0 {
		return nil, "invalid_batch_no"
	}

	var (
		sysEvents = arrGetSysEvents[0]
		status    = sysEvents.Status
		timeStart = sysEvents.TimeStart
		timeEnd   = sysEvents.TimeEnd
	)

	// validate input date range, if date range is emtpy, default get latest date range
	var arrWeeks = helpers.GetWeekStartAndEndDatesWithinDateRange(timeStart, timeEnd)

	if len(arrWeeks) <= 0 {
		base.LogErrorLog("eventService:GetEventSponsorRankingSetting()", "invalid_event_period", "", true)
		return nil, "something_went_wrong"
	}

	var (
		curDate     = time.Now()
		unlockWeeks = []map[string]string{}
	)

	for key, item := range arrWeeks {
		if helpers.CompareDateTime(curDate, ">=", item["week_start"]) {
			unlockWeeks = append(unlockWeeks, map[string]string{"label": helpers.TranslateV2("week_:0", langCode, map[string]string{"0": fmt.Sprint(key + 1)}), "week_start": item["week_start"].Format("2006-01-02"), "week_end": item["week_end"].Format("2006-01-02")})
		}
	}

	if status == 1 && helpers.CompareDateTime(curDate, "<", timeStart) {
		status = 0
	}

	var arrReturnData = map[string]interface{}{}
	arrReturnData["status"] = status
	arrReturnData["selection"] = unlockWeeks

	return arrReturnData, ""
}

// GetEventSponsorRankingListStruct struct
type GetEventSponsorRankingListStruct struct {
	Type     string
	BatchNo  int
	DateFrom string
	DateTo   string
	Page     int64
}

// EventSponsorRankingList struct
type EventSponsorRankingList struct {
	Rank           string  `json:"rank"`
	Username       string  `json:"username"`
	TotalSponsored float64 `json:"total_sponsored"`
}

// GetEventSponsorRankingList func
func GetEventSponsorRankingList(arrData GetEventSponsorRankingListStruct) (interface{}, string) {
	var arrEventSponsorRankingList = make([]EventSponsorRankingList, 0)

	var (
		eventType = ""
		batchNo   = arrData.BatchNo
	)

	if arrData.Type == "ALL" {
		eventType = "TS"
	} else if arrData.Type == "ND" { // new downline
		eventType = "TS_ND"
	} else {
		return nil, "invalid_type"
	}

	// get event setting by type and batch_no
	var arrGetSysEventsFn = make([]models.WhereCondFn, 0)
	arrGetSysEventsFn = append(arrGetSysEventsFn,
		models.WhereCondFn{Condition: " type = ? ", CondValue: eventType},
		models.WhereCondFn{Condition: " batch_no = ? ", CondValue: batchNo},
	)

	arrGetSysEvents, _ := models.GetSysEventsFn(arrGetSysEventsFn, "", false)
	if len(arrGetSysEvents) <= 0 {
		return nil, "invalid_batch_no"
	}

	var (
		sysEvents     = arrGetSysEvents[0]
		status        = sysEvents.Status
		timeStart     = sysEvents.TimeStart
		timeEnd       = sysEvents.TimeEnd
		rawEventSetup = sysEvents.Setting
	)

	// validate event status
	var curDate = time.Now()

	// if status == 0 || helpers.CompareDateTime(curDate, "<", timeStart) || helpers.CompareDateTime(curDate, ">", timeEnd) {
	if status == 0 {
		return nil, "event_not_available"
	}

	// validate input date range, if date range is emtpy, default get latest date range
	var arrWeeks = helpers.GetWeekStartAndEndDatesWithinDateRange(timeStart, timeEnd)

	if len(arrWeeks) <= 0 {
		base.LogErrorLog("eventService:GetEventSponsorRankingList()", "invalid_event_period", "", true)
		return nil, "something_went_wrong"
	}

	var (
		curWeekStart = arrWeeks[0]["week_start"]
		curWeekEnd   = arrWeeks[0]["week_end"]
		check        = false
	)

	eventSetup, errMsg := MapEventSetup(rawEventSetup)
	if errMsg != "" {
		base.LogErrorLog("eventService:GetEventSponsorRankingList()", "MapEventSetup()", errMsg, true)
		return nil, "something_went_wrong"
	}

	if arrData.DateFrom == "" || arrData.DateTo == "" {
		check = true
	}

	for _, item := range arrWeeks {
		if arrData.DateFrom != "" && arrData.DateTo != "" {
			if arrData.DateFrom == item["week_start"].Format("2006-01-02") && arrData.DateTo == item["week_end"].Format("2006-01-02") {
				curWeekStart = item["week_start"]
				curWeekEnd = item["week_end"]
				check = true

				break
			}
		} else {
			// get latest running week
			if helpers.CompareDateTime(curDate, ">=", item["week_start"]) {
				curWeekStart = item["week_start"]
				curWeekEnd = item["week_end"]
			}
		}
	}

	if !check {
		return nil, "invalid_date_from_or_date_to"
	}

	// get top x of sponsored.
	var arrGetEventSponsorRankingListFn = make([]models.WhereCondFn, 0)
	arrGetEventSponsorRankingListFn = append(arrGetEventSponsorRankingListFn,
		models.WhereCondFn{Condition: " sls_master.sponsor_id != ? ", CondValue: 1},
		models.WhereCondFn{Condition: " sls_master.status IN(?,'EP')", CondValue: "AP"}, // only accumulate expired and approved contract
		models.WhereCondFn{Condition: " sls_master.created_at >= ? ", CondValue: curWeekStart},
		models.WhereCondFn{Condition: " sls_master.created_at <= ? ", CondValue: curWeekEnd},
		models.WhereCondFn{Condition: " date(sls_master.doc_date) < ? ", CondValue: curDate.Format("2006-01-02")},
	)

	// set criteria to only accumulate sales from newly registered downline
	if arrData.Type == "ND" {
		arrGetEventSponsorRankingListFn = append(arrGetEventSponsorRankingListFn,
			models.WhereCondFn{Condition: " ent_member.created_at >= ? ", CondValue: timeStart},
		)
	}

	arrGetEventSponsorRankingList, _ := models.GetEventSponsorRankingListFn(eventSetup.Quota, arrGetEventSponsorRankingListFn, false)

	if len(arrGetEventSponsorRankingList) > 0 {
		for key, arrGetEventSponsorRankingListV := range arrGetEventSponsorRankingList {
			maskNum := 4
			if len(arrGetEventSponsorRankingListV.Username) <= 4 {
				maskNum = 2
			}

			arrEventSponsorRankingList = append(arrEventSponsorRankingList,
				EventSponsorRankingList{
					Rank:           fmt.Sprint(key + 1),
					Username:       helpers.MaskLeft(arrGetEventSponsorRankingListV.Username, maskNum),
					TotalSponsored: arrGetEventSponsorRankingListV.TotalSponsored,
				},
			)
		}
	}

	// get ghost data
	var arrGetEventGhostListFn = make([]models.WhereCondFn, 0)
	arrGetEventGhostListFn = append(arrGetEventGhostListFn,
		models.WhereCondFn{Condition: " dt_start >= ? ", CondValue: curWeekStart},
		models.WhereCondFn{Condition: " dt_start <= ? ", CondValue: curWeekEnd},
		models.WhereCondFn{Condition: " dt_start < ? ", CondValue: curDate}, // only display those ghost with affected day lesser than or equal to today
	)

	arrGetEventGhostList, _ := models.GetSysEventsGhostListFn(eventType, batchNo, eventSetup.Quota, arrGetEventGhostListFn, false)

	if len(arrGetEventGhostList) > 0 {
		for key, arrGetEventGhostListV := range arrGetEventGhostList {
			maskNum := 4
			if len(arrGetEventGhostListV.Username) <= 4 {
				maskNum = 2
			}

			arrEventSponsorRankingList = append(arrEventSponsorRankingList,
				EventSponsorRankingList{
					Rank:           fmt.Sprint(key + 1),
					Username:       helpers.MaskLeft(arrGetEventGhostListV.Username, maskNum),
					TotalSponsored: arrGetEventGhostListV.TotalAmount,
				},
			)
		}
	}

	// sort by sponsored amount then cut into size of quota
	sort.Slice(arrEventSponsorRankingList, func(p, q int) bool {
		return arrEventSponsorRankingList[q].TotalSponsored < arrEventSponsorRankingList[p].TotalSponsored
	})

	// convert to []interface{} for pagination
	var counter = 0
	var arrListingData []interface{}
	if len(arrEventSponsorRankingList) > 0 {
		for _, arrEventSponsorRankingListV := range arrEventSponsorRankingList {
			arrListingData = append(arrListingData,
				map[string]string{
					"rank":            arrEventSponsorRankingListV.Rank,
					"username":        arrEventSponsorRankingListV.Username,
					"total_sponsored": helpers.CutOffDecimal(arrEventSponsorRankingListV.TotalSponsored, 2, ".", ","),
				},
			)

			counter++
			if counter >= eventSetup.Quota {
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

// RarityList struct
type RarityList struct {
	Code, Name string
}

// GetRarityList
func GetRarityList() []RarityList {
	var arrRarityList = []RarityList{}

	auctionApiSetting, _ := models.GetSysGeneralSetupByID("auction_api_setting")

	var (
		domain   = auctionApiSetting.InputValue1
		url      = domain + "/api/v1/admin/prop/get"
		response app.ApiArrayResponse
	)

	header := map[string]string{
		"Content-Type": "application/json",
	}

	res, err := base.RequestAPI("POST", url, header, nil, &response)
	if err != nil {
		base.LogErrorLog("eventService:GetRarityList", "RequestAPI()", err.Error(), true)
		return arrRarityList
	}

	if res.StatusCode == 200 {
		if len(response.Data) > 0 {
			for _, data := range response.Data {
				rarityCodeByte, _ := json.Marshal(data["prop_code"])
				rarityCode := string(rarityCodeByte)
				rarityCode = strings.Replace(rarityCode, "\"", "", 2)

				rarityNameByte, _ := json.Marshal(data["prop_val"])
				rarityName := string(rarityNameByte)
				rarityName = strings.Replace(rarityName, "\"", "", 2)

				arrRarityList = append(arrRarityList, RarityList{
					Code: rarityCode,
					Name: rarityName,
				})
			}
		}
	} else {
		errMsg, _ := json.Marshal(response.Msg)
		errMsgStr := string(errMsg)
		errMsgStr = strings.Replace(errMsgStr, "\"", "", 2)
		base.LogErrorLog("eventService:GetRarityList", "RequestAPI()"+err.Error(), map[string]interface{}{"responseBody": response, "errMsg": errMsgStr}, true)
	}

	return arrRarityList
}

// GetAuctionLuckyNumberListStruct struct
type GetAuctionLuckyNumberListStruct struct {
	DateFrom string
	DateTo   string
	LangCode string
}

type LuckyNumber struct {
	Rarity      string `json:"rarity"`
	LuckyNumber int    `json:"lucky_number"`
	DateStart   string `json:"date_start"`
	DateEnd     string `json:"date_end"`
}

// GetAuctionLuckyNumberList func
func GetAuctionLuckyNumberList(arrData GetAuctionLuckyNumberListStruct) (string, interface{}) {
	var (
		curDate                         = time.Now().Format("2006-01-02")
		arrRarityList                   = GetRarityList() // get all rarity from auction site
		arrLuckyNumber                  = []LuckyNumber{}
		greatestDateStart, leastDateEnd time.Time
		title                           string = ""
	)

	// get one lucky number from each rarity
	for _, arrRarityListV := range arrRarityList {
		rarityCode := arrRarityListV.Code
		rarityName := helpers.TranslateV2(arrRarityListV.Name, arrData.LangCode, make(map[string]string))

		var arrGetAuctionLuckyNumberFn = make([]models.WhereCondFn, 0)
		arrGetAuctionLuckyNumberFn = append(arrGetAuctionLuckyNumberFn,
			models.WhereCondFn{Condition: " auction_lucky_number.rarity_code = ? ", CondValue: rarityCode},
			models.WhereCondFn{Condition: " auction_lucky_number.status = ? ", CondValue: 1}, // only get those active
		)

		// only get those lucky number within period
		if arrData.DateFrom != "" && arrData.DateTo != "" {
			dateFrom, err := base.StrToDateTime(arrData.DateFrom, "2006-01-02")
			if err == nil {
				arrGetAuctionLuckyNumberFn = append(arrGetAuctionLuckyNumberFn,
					models.WhereCondFn{Condition: " ? <= date(auction_lucky_number.date_end) ", CondValue: dateFrom.Format("2006-01-02")},
				)
			}

			dateTo, err := base.StrToDateTime(arrData.DateTo, "2006-01-02")
			if err == nil {
				arrGetAuctionLuckyNumberFn = append(arrGetAuctionLuckyNumberFn,
					models.WhereCondFn{Condition: " ? >= date(auction_lucky_number.date_start) ", CondValue: dateTo.Format("2006-01-02")},
				)
			}

		} else {
			arrGetAuctionLuckyNumberFn = append(arrGetAuctionLuckyNumberFn,
				models.WhereCondFn{Condition: " date(auction_lucky_number.date_start) <= ? ", CondValue: curDate},
				models.WhereCondFn{Condition: " date(auction_lucky_number.date_end) >= ? ", CondValue: curDate},
			)
		}

		arrGetAuctionLuckyNumber, _ := models.GetAuctionLuckyNumber(arrGetAuctionLuckyNumberFn, "", false)
		if len(arrGetAuctionLuckyNumber) > 0 {
			arrLuckyNumber = append(arrLuckyNumber, LuckyNumber{
				Rarity:      rarityName,
				LuckyNumber: arrGetAuctionLuckyNumber[0].LuckyNumber,
				DateStart:   arrGetAuctionLuckyNumber[0].DateStart.Format("2006-01-02"),
				DateEnd:     arrGetAuctionLuckyNumber[0].DateEnd.Format("2006-01-02"),
			})

			// take biggest date start
			if helpers.CompareDateTime(arrGetAuctionLuckyNumber[0].DateStart, ">", greatestDateStart) {
				greatestDateStart = arrGetAuctionLuckyNumber[0].DateStart
			}

			// take least date end
			if leastDateEnd.IsZero() || helpers.CompareDateTime(arrGetAuctionLuckyNumber[0].DateEnd, "<", leastDateEnd) {
				leastDateEnd = arrGetAuctionLuckyNumber[0].DateEnd
			}
		}
	}

	title = fmt.Sprint(greatestDateStart.Format("2006-01-02"), " - ", leastDateEnd.Format("2006-01-02"))

	return title, arrLuckyNumber
}
