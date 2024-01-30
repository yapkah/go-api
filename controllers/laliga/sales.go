package laliga

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/service/sales_service"
)

// ProcessSaveTransv1Form struct
type ProcessSaveTransv1Form struct {
	// UnitPrice  string `form:"unit_price" json:"unit_price" valid:"Required;"`
	DocNo      string `form:"bizId" json:"bizId" valid:"Required;"`
	TransType  string `form:"trans_type" json:"trans_type" valid:"Required;"`
	TotalIn    string `form:"total_in" json:"total_in"`
	TotalOut   string `form:"total_out" json:"total_out"`
	SigningKey string `form:"signing_key" json:"signing_key"`
	Remark     string `form:"remark" json:"remark"`
}

//func ProcessSaveTransv1 function
func ProcessSaveTransv1(c *gin.Context) {

	var (
		appG = app.Gin{C: c}
		form ProcessSaveTransv1Form
	)

	// langCode := setting.Cfg.Section("app").Key("DefaultLangCode").String()
	// if c.GetHeader("Accept-Language") != "" {
	// 	langCode = c.GetHeader("Accept-Language")
	// }

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	u, ok := c.Get("access_user")
	if !ok {
		message := app.MsgStruct{
			Msg: "invalid_member",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}

	member := u.(*models.EntMemberMembers)

	if form.TotalIn == "" && form.TotalOut == "" {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "total_in_or_total_out_is_required_on_" + form.DocNo}, nil)
		return
	}

	totalIn := float64(0)
	if form.TotalIn != "" {
		restFloat, _ := strconv.ParseFloat(form.TotalIn, 64)
		cutoffString := helpers.CutOffDecimal(restFloat, 8, ".", ",")
		finalRst, _ := strconv.ParseFloat(cutoffString, 64)
		// totalInBigFloat, err := float.SetString(form.TotalIn)
		// if err != nil {
		// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "total_in_" + err.Error() + "_on_" + form.DocNo}, nil)
		// 	return
		// }
		// totalInRst := totalInBigFloat.Float64()
		totalIn = finalRst
	}

	totalOut := float64(0)
	if form.TotalOut != "" {
		restFloat, _ := strconv.ParseFloat(form.TotalOut, 64)
		cutoffString := helpers.CutOffDecimal(restFloat, 8, ".", ",")
		finalRst, _ := strconv.ParseFloat(cutoffString, 64)
		// totalOutBigFloat, err := float.SetString(form.TotalOut)
		// if err != nil {
		// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "total_out_" + err.Error() + "_on_" + form.DocNo}, nil)
		// 	return
		// }
		// totalOutRst := totalOutBigFloat.Float64()
		totalOut = finalRst
	}

	if totalIn == 0 && totalOut == 0 {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "only_total_in_or_total_out_is_required_on_" + form.DocNo}, nil)
		return
	}

	// begin transaction
	tx := models.Begin()
	var arrDataReturn map[string]interface{}
	if strings.ToLower(form.TransType) == "laliga_stake" {
		arrData := sales_service.ProcessSaveTransv1Struct{
			DocNo:       form.DocNo,
			EntMemberID: member.EntMemberID,
			CryptoType:  "LIGA",
			TransType:   form.TransType,
			TotalOut:    totalOut,
			SigningKey:  form.SigningKey,
			Remark:      form.Remark,
		}
		fmt.Println("laliga_stake", arrData)
		result, err := sales_service.ProcessSaveTransOutv1(tx, arrData)

		if err != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}
		arrDataReturn = result
	} else if strings.ToLower(form.TransType) == "laliga_unstake" || strings.ToLower(form.TransType) == "laliga_claim" {
		arrData := sales_service.ProcessSaveTransv1Struct{
			DocNo:       form.DocNo,
			EntMemberID: member.EntMemberID,
			CryptoType:  "LIGA",
			TransType:   form.TransType,
			TotalIn:     totalIn,
			Remark:      form.Remark,
		}
		fmt.Println(strings.ToLower(form.TransType), arrData)
		result, err := sales_service.ProcessSaveTransInv1(tx, arrData)

		if err != nil {
			models.Rollback(tx)
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}
		arrDataReturn = result
	}

	err := models.Commit(tx)
	if err != nil {
		models.Rollback(tx)
		base.LogErrorLog("ProcessSaveTransv1-Commit Failed", err.Error(), "", true)
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		return
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrDataReturn)
}
