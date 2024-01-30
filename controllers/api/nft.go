package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/base"
	"github.com/yapkah/go-api/pkg/setting"
)

func GetNftList(c *gin.Context) {
	var (
		appG    = app.Gin{C: c}
		arrData = []interface{}{}
		curTime = base.GetCurrentTime("2006-01-02 15:04:05")
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	langCode = c.GetHeader("Accept-Language")
	// }

	// get nft_series_group_setup
	nftSeriesGroupSetupFn := make([]models.WhereCondFn, 0)
	nftSeriesGroupSetupFn = append(nftSeriesGroupSetupFn,
		models.WhereCondFn{Condition: "nft_series_group_setup.status = ?", CondValue: "A"},
	)
	nftSeriesGroupSetup, err := models.GetNftSeriesGroupSetupFn(nftSeriesGroupSetupFn, "", false)
	if err != nil {
		base.LogErrorLog("apiController:GetNftList():GetNftSeriesGroupSetupFn()", map[string]interface{}{"condition": nftSeriesGroupSetupFn}, err.Error(), true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	if len(nftSeriesGroupSetup) > 0 {
		for _, nftSeriesGroupSetupV := range nftSeriesGroupSetup {
			arrNftSeriesSetupFn := make([]models.WhereCondFn, 0)
			arrNftSeriesSetupFn = append(arrNftSeriesSetupFn,
				models.WhereCondFn{Condition: "nft_series_setup.group_type = ?", CondValue: nftSeriesGroupSetupV.ID},
				models.WhereCondFn{Condition: "nft_series_setup.status = ?", CondValue: "A"},
				models.WhereCondFn{Condition: "nft_series_setup.start_date <= ?", CondValue: curTime},
				models.WhereCondFn{Condition: "nft_series_setup.end_date >= ?", CondValue: curTime},
			)
			arrNftSeriesSetup, err := models.GetNftSeriesSetupFn(arrNftSeriesSetupFn, "", false)
			if err != nil {
				base.LogErrorLog("apiController:GetNftList():GetNftSeriesSetupFn()", map[string]interface{}{"conndition": arrNftSeriesSetup}, err.Error(), true)
				appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}

			for _, arrNftSeriesSetupV := range arrNftSeriesSetup {
				var (
					arrDataValue = map[string]interface{}{}
					drawingUrl   string
				)

				// get drawing url
				arrNftImgFn := make([]models.WhereCondFn, 0)
				arrNftImgFn = append(arrNftImgFn,
					models.WhereCondFn{Condition: "nft_img.type = ?", CondValue: arrNftSeriesSetupV.Code},
					models.WhereCondFn{Condition: "nft_img.status = ?", CondValue: "A"},
				)
				arrNftImg, err := models.GetNftImgFn(arrNftImgFn, "", false)
				if err != nil {
					base.LogErrorLog("apiController:GetNftList():GetNftImgFn()", map[string]interface{}{"conndition": arrNftImgFn}, err.Error(), true)
					appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
					return
				}

				if len(arrNftImg) > 0 {
					drawingUrl = arrNftImg[0].ImgLink
				}

				arrDataValue["title"] = nftSeriesGroupSetupV.Name
				arrDataValue["description"] = nftSeriesGroupSetupV.Description
				arrDataValue["drawing_url"] = drawingUrl
				arrDataValue["opensea_url"] = "https://opensea.io/"
				arrDataValue["recolte_url"] = setting.Cfg.Section("custom").Key("MemberServerDomain").String()

				arrData = append(arrData, arrDataValue)
			}
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrData)
}
