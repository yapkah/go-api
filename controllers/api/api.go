package api

import (
	"encoding/csv"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/pkg/util"
	"github.com/smartblock/gta-api/service/product_service"
	"github.com/smartblock/gta-api/service/sales_service"
	"github.com/smartblock/gta-api/service/trading_service"
	"github.com/smartblock/gta-api/service/wallet_service"
)

// ProcessUpdateCryptoWithdrawalv1Form struct
type ProcessUpdateCryptoWithdrawalv1Form struct {
	ID           int    `form:"id" json:"id" valid:"Required;"`
	BatchID      int    `form:"batch_id" json:"batch_id" valid:"Required;"`
	ConfigCode   string `form:"config_code" json:"config_code"`
	AssetCode    string `form:"asset_code" json:"asset_code" valid:"Required;"`
	DocNo        string `form:"doc_no" json:"doc_no" valid:"Required;"`
	ToAddress    string `form:"to_address" json:"to_address" valid:"Required;"`
	TxHash       string `form:"tx_hash" json:"tx_hash"`
	Value        string `form:"value" json:"value" valid:"Required;"`
	Gas          string `form:"gas" json:"gas"`
	GasPrice     string `form:"gas_price" json:"gas_price"`
	RunStatus    string `form:"run_status" json:"run_status"` // either "COMPLETE" or "REJECTED"
	RunAttempts  int    `form:"run_attempts" json:"run_attempts"`
	RunStatusAt  string `form:"run_status_at" json:"run_status_at"`
	ApproverName string `form:"approver_name" json:"approver_name"`
	ApproveAt    string `form:"approve_at" json:"approve_at"`
	Remark       string `form:"remark" json:"remark"`
}

// func ProcessUpdateCryptoWithdrawalv1
func ProcessUpdateCryptoWithdrawalv1(c *gin.Context) {

	var (
		form ProcessUpdateCryptoWithdrawalv1Form
	)

	// validate input
	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		arrDataReturn := map[string]interface{}{
			"rst": 0,
			"msg": msg,
		}
		c.JSON(200, arrDataReturn)
		return
	}

	tx := models.Begin()

	arrData := wallet_service.ProcessUpdateCryptoWithdrawalv1Struct{
		BatchID:      form.BatchID,
		DocNo:        form.DocNo,
		ToAddress:    form.ToAddress,
		TxHash:       form.TxHash,
		Value:        form.Value,
		Gas:          form.Gas,
		GasPrice:     form.GasPrice,
		RunStatus:    form.RunStatus, // either "COMPLETE" or "REJECTED"
		RunAttempts:  form.RunAttempts,
		RunStatusAt:  form.RunStatusAt,
		ApproverName: form.ApproverName,
		ApproveAt:    form.ApproveAt,
		Remark:       form.Remark,
	}

	err := wallet_service.ProcessUpdateCryptoWithdrawalv1(tx, arrData)

	if err != nil {
		models.Rollback(tx)
		arrDataReturn := map[string]interface{}{
			"rst": 0,
			"msg": err.Error(),
		}
		c.JSON(200, arrDataReturn)
		return
	}

	err = models.Commit(tx)

	c.JSON(200, form)
	return

}

type ProcessCryptoReturnDataStruct struct {
	Id               int     `json:"id"`
	ConfigCode       string  `json:"config_code"`
	TransStatus      string  `json:"trans_status"`
	CallbackStatus   string  `json:"callback_status"`
	CallbackStatusAt string  `json:"callback_status_at"`
	TxHash           string  `json:"tx_hash" valid:"Required"`
	FromAddr         string  `json:"from_addr" valid:"Required"`
	ToAddr           string  `json:"to_addr" valid:"Required"`
	AssetCode        string  `json:"asset_code" valid:"Required"`
	Amount           float64 `json:"from_addr" valid:"Required"`
	CreatedAt        string  `json:"created_at"`
}

func ProcessCryptoReturn(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form ProcessCryptoReturnDataStruct
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// wallet_service.ProcessCryptoReturn(c.Request, form)
	return
}

type TransactionCallbackData struct {
	Hash   string `json:"hash" form:"hash"`
	Status bool   `json:"status" form:"status"`
}

func TransactionCallback(c *gin.Context) {
	var (
		appG     = app.Gin{C: c}
		form     TransactionCallbackData
		bcStatus string
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	tx := models.Begin()

	if form.Status {
		bcStatus = "AP"
	} else {
		bcStatus = "F"
	}

	// get blockchain trans record
	arrBlockCond := make([]models.WhereCondFn, 0)
	arrBlockCond = append(arrBlockCond,
		models.WhereCondFn{Condition: "blockchain_trans.hash_value = ?", CondValue: form.Hash},
		models.WhereCondFn{Condition: "blockchain_trans.transaction_type != 'TRANSFER_TO_EXCHANGE' OR blockchain_trans.log_only = ? ", CondValue: 0},
		models.WhereCondFn{Condition: "blockchain_trans.status = ?", CondValue: "P"},
	)
	BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

	if len(BlockchainTrans) > 0 {
		row := BlockchainTrans[0]

		// if row.TransactionType == "CONTRACT" || row.TransactionType == "P2P" || row.TransactionType == "MINING" || row.TransactionType == "MINING_BZZ" {
		// 	errMsg := sales_service.SlsMasterCallback(tx, row.DocNo, form.Hash, row.TransactionType, bcStatus)
		// 	if errMsg != "" {
		// 		base.LogErrorLog("TransactionCallback", "SlsMasterCallback()", "[docNo:"+row.DocNo+", hashValue:"+form.Hash+"]"+errMsg, true)
		// 		tx.Rollback()
		// 		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		// 		return
		// 	}
		// } else if row.TransactionType == "CONTRACT_TOPUP" {
		// 	errMsg := sales_service.SlsMasterTopupCallback(tx, row.DocNo, form.Hash, row.TransactionType, bcStatus)
		// 	if errMsg != "" {
		// 		tx.Rollback()
		// 		base.LogErrorLog("TransactionCallback-SlsMasterTopupCallback()", errMsg, map[string]interface{}{"input": form, "doc_no": row.DocNo}, true)
		// 		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
		// 		return
		// 	}
		// }

		if row.TransactionType == "TRANSFER_TO_EXCHANGE" {
			// Update TransferToExchange Status
			ewtTransferExchangeErr := wallet_service.UpdateTransferToExchangeStatus(tx, row.DocNo, bcStatus)

			if ewtTransferExchangeErr != "" {
				tx.Rollback()
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: ewtTransferExchangeErr}, nil)
				return
			}
		} else if strings.ToLower(row.TransactionType) == "trading_match" {
			err := trading_service.UpdateTradingMatchTranxCallback(tx, row.DocNo, form.Status)
			if err != nil {
				tx.Rollback()
				arrErr := map[string]interface{}{
					"doc_no": row.DocNo,
					"form":   form,
				}
				base.LogErrorLog("TransactionCallback-UpdateTradingMatchTranxCallback_failed", err.Error(), arrErr, true)
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		} else if strings.ToLower(row.TransactionType) == "trading_after_match" {
			err := trading_service.UpdateTradingAfterMatchTranxCallback(tx, row.DocNo, form.Status)
			if err != nil {
				tx.Rollback()
				arrErr := map[string]interface{}{
					"doc_no": row.DocNo,
					"form":   form,
				}
				base.LogErrorLog("TransactionCallback-UpdateTradingAfterMatchTranxCallback", err.Error(), arrErr, true)
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		} else if strings.ToLower(row.TransactionType) == "trading_sell" {
			err := trading_service.UpdateTradingSellTranxCallback(tx, row.DocNo, form.Status)
			if err != nil {
				tx.Rollback()
				arrErr := map[string]interface{}{
					"doc_no": row.DocNo,
					"form":   form,
				}
				base.LogErrorLog("TransactionCallback-UpdateTradingSellTranxCallback_failed", err.Error(), arrErr, true)
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		} else if strings.ToLower(row.TransactionType) == "trading_buy" {
			err := trading_service.UpdateTradingBuyTranxCallback(tx, row.DocNo, form.Status)
			if err != nil {
				tx.Rollback()
				arrErr := map[string]interface{}{
					"doc_no": row.DocNo,
					"form":   form,
				}
				base.LogErrorLog("TransactionCallback-UpdateTradingBuyTranxCallback_failed", err.Error(), arrErr, true)
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		} else if strings.ToLower(row.TransactionType) == "trading_cancel" {
			err := trading_service.UpdateTradingCancelTranxCallback(tx, row.DocNo, form.Status)
			if err != nil {
				tx.Rollback()
				arrErr := map[string]interface{}{
					"doc_no": row.DocNo,
					"form":   form,
				}
				base.LogErrorLog("TransactionCallback-UpdateTradingCancelTranxCallback_failed", err.Error(), arrErr, true)
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		} else if row.TransactionType == "UNSTAKE" {
			errMsg := sales_service.ApproveUnstakeCallback(tx, bcStatus, row.DocNo, row.ConvertedTotalIn)
			if errMsg != "" {
				base.LogErrorLog("TransactionCallback", "ApproveUnstakeCallback()", "[docNo:"+row.DocNo+", hashValue:"+form.Hash+"]"+errMsg, true)
				tx.Rollback()
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		} else if row.TransactionType == "WITHDRAW" {
			// Update ewt_withdraw Status
			ewtWithdrawStatusErr := wallet_service.UpdateWithdrawStatus(tx, row.DocNo, bcStatus)

			if ewtWithdrawStatusErr != "" {
				tx.Rollback()
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: ewtWithdrawStatusErr}, nil)
				return
			}
		} else if row.TransactionType == "EXCHANGE" {
			if row.TotalOut > 0 && form.Status { // deduct wallet tranasction approved - call exchange debit action
				// errMsg := product_service.ExchangeCallback(tx, row.DocNo)
				// if errMsg != "" {
				// 	base.LogErrorLog("TransactionCallback", "ExchangeCallback()", "[docNo:"+row.DocNo+", hashValue:"+form.Hash+"]"+errMsg, true)
				// 	tx.Rollback()
				// 	appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				// 	return
				// }
			} else if row.TotalIn > 0 { // debit to usds callback - approve ewt_exchange
				errMsg := product_service.ExchangeApproveCallback(tx, row.DocNo, form.Status)
				if errMsg != "" {
					base.LogErrorLog("TransactionCallback", "ExchangeApproveCallback()", "[docNo:"+row.DocNo+", hashValue:"+form.Hash+"]"+errMsg, true)
					tx.Rollback()
					appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
					return
				}
			}
		} else if row.TransactionType == "ADJUST" || row.TransactionType == "ADJUST_IN" {
			errMsg := wallet_service.AdjustCallback(tx, form.Hash, bcStatus)
			if errMsg != "" {
				base.LogErrorLog("TransactionCallback", "AdjustCallback()", "[hashValue:"+form.Hash+"]"+errMsg, true)
				tx.Rollback()
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
				return
			}
		} else if row.TransactionType == "WITHDRAW-POOL" {
			// Update ewt_withdraw_pool Status
			ewtWithdrawPoolStatusErr := wallet_service.UpdateWithdrawPoolStatus(tx, row.DocNo, bcStatus)

			if ewtWithdrawPoolStatusErr != "" {
				tx.Rollback()
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: ewtWithdrawPoolStatusErr}, nil)
				return
			}
		} else if strings.ToLower(row.TransactionType) == "laliga_stake" || strings.ToLower(row.TransactionType) == "laliga_unstake" || strings.ToLower(row.TransactionType) == "laliga_claim" {
			// processID := strings.ToUpper(row.TransactionType) + "-" + strconv.Itoa(row.ID)
			if form.Status {
				processID := row.HashValue
				arrLaligaProcessQ := make([]models.WhereCondFn, 0)
				arrLaligaProcessQ = append(arrLaligaProcessQ,
					models.WhereCondFn{Condition: "process_id = ?", CondValue: processID},
				)
				existingLaligaProcessQ, _ := models.GetLaligaProcessQueueFn(arrLaligaProcessQ, false)

				if len(existingLaligaProcessQ) < 1 {
					arrCrtData := models.AddLaligaProcessQueueStruct{
						ProcessID: processID,
						Status:    "P",
					}
					models.AddLaligaProcessQueue(tx, arrCrtData)
				}
			}
		}

		// update blockchain tran
		updateBlockchainTransErr := wallet_service.UpdateBlockchainTransStatus(tx, form.Hash, bcStatus)

		if updateBlockchainTransErr != "" {
			tx.Rollback()
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: updateBlockchainTransErr}, nil)
			return
		}

	} else {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "invalid_hash"}, nil)
		return
	}

	tx.Commit()
	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

type TransactionCallbackBatchData struct {
	TxData []TransactionCallbackData `json:"tx_data"`
}

func TransactionCallbackBatch(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form TransactionCallbackBatchData
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	var arrRst = []map[string]string{}

	arrTransactionCallback := form.TxData
	if len(arrTransactionCallback) > 0 {
		for _, arrTransactionCallbackV := range arrTransactionCallback {
			hashValue := arrTransactionCallbackV.Hash
			bcStatus := "AP"

			if !arrTransactionCallbackV.Status {
				bcStatus = "F"
			}

			// get blockchain trans record
			arrBlockCond := make([]models.WhereCondFn, 0)
			arrBlockCond = append(arrBlockCond,
				models.WhereCondFn{Condition: "blockchain_trans.hash_value = ?", CondValue: hashValue},
				models.WhereCondFn{Condition: "blockchain_trans.transaction_type != 'TRANSFER_TO_EXCHANGE' OR blockchain_trans.log_only = ? ", CondValue: 0},
				models.WhereCondFn{Condition: "blockchain_trans.status = ?", CondValue: "P"},
			)
			BlockchainTrans, _ := models.GetBlockchainTransArrayFn(arrBlockCond, false)

			if len(BlockchainTrans) > 0 {
				row := BlockchainTrans[0]

				tx := models.Begin()
				// if row.TransactionType == "CONTRACT" || row.TransactionType == "P2P" || row.TransactionType == "MINING" || row.TransactionType == "MINING_BZZ" {
				// 	errMsg := sales_service.SlsMasterCallback(tx, row.DocNo, hashValue, row.TransactionType, bcStatus)
				// 	if errMsg != "" {
				// 		tx.Rollback()
				// 		base.LogErrorLog("TransactionCallbackBatch-SlsMasterCallback()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
				// 		arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
				// 		continue
				// 	}
				// } else if row.TransactionType == "CONTRACT_TOPUP" {
				// 	errMsg := sales_service.SlsMasterTopupCallback(tx, row.DocNo, hashValue, row.TransactionType, bcStatus)
				// 	if errMsg != "" {
				// 		tx.Rollback()
				// 		base.LogErrorLog("TransactionCallbackBatch-SlsMasterTopupCallback()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
				// 		arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
				// 		continue
				// 	}
				// }

				if row.TransactionType == "TRANSFER_TO_EXCHANGE" {
					// Update TransferToExchange Status
					errMsg := wallet_service.UpdateTransferToExchangeStatus(tx, row.DocNo, bcStatus)
					if errMsg != "" {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-UpdateTransferToExchangeStatus()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if strings.ToLower(row.TransactionType) == "trading_match" {
					err := trading_service.UpdateTradingMatchTranxCallback(tx, row.DocNo, arrTransactionCallbackV.Status)
					if err != nil {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-UpdateTradingMatchTranxCallback()", err.Error(), map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if strings.ToLower(row.TransactionType) == "trading_after_match" {
					err := trading_service.UpdateTradingAfterMatchTranxCallback(tx, row.DocNo, arrTransactionCallbackV.Status)
					if err != nil {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-UpdateTradingAfterMatchTranxCallback()", err.Error(), map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if strings.ToLower(row.TransactionType) == "trading_sell" {
					err := trading_service.UpdateTradingSellTranxCallback(tx, row.DocNo, arrTransactionCallbackV.Status)
					if err != nil {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-UpdateTradingSellTranxCallback()", err.Error(), map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if strings.ToLower(row.TransactionType) == "trading_buy" {
					err := trading_service.UpdateTradingBuyTranxCallback(tx, row.DocNo, arrTransactionCallbackV.Status)
					if err != nil {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-UpdateTradingBuyTranxCallback()", err.Error(), map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if strings.ToLower(row.TransactionType) == "trading_cancel" {
					err := trading_service.UpdateTradingCancelTranxCallback(tx, row.DocNo, arrTransactionCallbackV.Status)
					if err != nil {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-UpdateTradingCancelTranxCallback()", err.Error(), map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if row.TransactionType == "UNSTAKE" {
					errMsg := sales_service.ApproveUnstakeCallback(tx, bcStatus, row.DocNo, row.ConvertedTotalIn)
					if errMsg != "" {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-ApproveUnstakeCallback()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if row.TransactionType == "WITHDRAW" {
					// Update ewt_withdraw Status
					errMsg := wallet_service.UpdateWithdrawStatus(tx, row.DocNo, bcStatus)
					if errMsg != "" {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-UpdateWithdrawStatus()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if row.TransactionType == "EXCHANGE" {
					if row.TotalOut > 0 && arrTransactionCallbackV.Status { // deduct wallet transaction approved - call exchange debit action
						// errMsg := product_service.ExchangeCallback(tx, row.DocNo)
						// if errMsg != "" {
						// 	tx.Rollback()
						// 	base.LogErrorLog("TransactionCallbackBatch-ExchangeCallback()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						// 	arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						// 	continue
						// }
					} else if row.TotalIn > 0 { // debit to usds callback - approve ewt_exchange
						errMsg := product_service.ExchangeApproveCallback(tx, row.DocNo, arrTransactionCallbackV.Status)
						if errMsg != "" {
							tx.Rollback()
							base.LogErrorLog("TransactionCallbackBatch-ExchangeApproveCallback()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
							arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
							continue
						}
					}
				} else if row.TransactionType == "ADJUST" || row.TransactionType == "ADJUST_IN" {
					errMsg := wallet_service.AdjustCallback(tx, hashValue, bcStatus)
					if errMsg != "" {
						tx.Rollback()
						base.LogErrorLog("TransactionCallbackBatch-AdjustCallback()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, true)
						arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
						continue
					}
				} else if row.TransactionType == "WITHDRAW-POOL" {
					// Update ewt_withdraw_pool Status
					ewtWithdrawPoolStatusErr := wallet_service.UpdateWithdrawPoolStatus(tx, row.DocNo, bcStatus)

					if ewtWithdrawPoolStatusErr != "" {
						tx.Rollback()
						appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: ewtWithdrawPoolStatusErr}, nil)
						return
					}
				} else if strings.ToLower(row.TransactionType) == "laliga_stake" || strings.ToLower(row.TransactionType) == "laliga_unstake" || strings.ToLower(row.TransactionType) == "laliga_claim" {
					// processID := strings.ToUpper(row.TransactionType) + "-" + strconv.Itoa(row.ID)
					if arrTransactionCallbackV.Status {
						processID := row.HashValue
						arrLaligaProcessQ := make([]models.WhereCondFn, 0)
						arrLaligaProcessQ = append(arrLaligaProcessQ,
							models.WhereCondFn{Condition: "process_id = ?", CondValue: processID},
						)
						existingLaligaProcessQ, _ := models.GetLaligaProcessQueueFn(arrLaligaProcessQ, false)

						if len(existingLaligaProcessQ) < 1 {
							arrCrtData := models.AddLaligaProcessQueueStruct{
								ProcessID: processID,
								Status:    "P",
							}
							_, err := models.AddLaligaProcessQueue(tx, arrCrtData)
							if err != nil {
								tx.Rollback()
								base.LogErrorLog("TransactionCallbackBatch-"+strings.ToLower(row.TransactionType), err.Error(), arrCrtData, true)
								appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "something_went_wrong"}, nil)
								return
							}
						}
					}
				}

				// update blockchain tran
				errMsg := wallet_service.UpdateBlockchainTransStatus(tx, hashValue, bcStatus)
				if errMsg != "" {
					tx.Rollback()
					base.LogErrorLog("TransactionCallbackBatch-UpdateBlockchainTransStatus()", errMsg, map[string]interface{}{"input": arrTransactionCallbackV, "doc_no": row.DocNo}, false)
					arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
					continue
				}

				tx.Commit()
				arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "1"})

			} else {
				base.LogErrorLog("TransactionCallbackBatch()", "invalid_hash", map[string]interface{}{"input": arrTransactionCallbackV}, false)
				arrRst = append(arrRst, map[string]string{"hash_value": hashValue, "status": "0"})
				continue
			}

		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, arrRst)
	return
}

// UpdateTranslationFrontendCSVFileForm struct
type UpdateTranslationFrontendCSVFileForm struct {
	DataFile *multipart.FileHeader `form:"data_file" json:"data_file"`
	Type     string                `form:"type" json:"type" valid:"MaxSize(100)"`
}

// UpdateTranslationByFile get translation list
func ProcessUpdateTranslationFrontendCSVFile(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form UpdateTranslationFrontendCSVFileForm
		err  error
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	file, header, err := c.Request.FormFile("data_file")
	if err != nil {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "data_file_is_required_1"}, nil)
		return
	}

	if file == nil || header == nil {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "data_file_is_required_2"}, nil)
		return
	}

	if header.Header["Content-Type"][0] != "text/csv" {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "only_csv_file_is_allowed"}, nil)
		return
	}

	r := csv.NewReader(file)

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "date_file_read_failed_2"}, nil)
			return
		}

		tx := models.Begin()
		for i := 1; i < len(record); i++ {
			locale := "en"
			if i == 2 {
				locale = "zh"
			} else if i == 3 {
				locale = "ko"
			} else if i == 4 {
				locale = "id"
			} else if i == 5 {
				locale = "th"
			}
			// fmt.Println("i:", i)
			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " translations_frontend.group = ?", CondValue: "common"},
				models.WhereCondFn{Condition: " translations_frontend.type = ?", CondValue: "label"},
				models.WhereCondFn{Condition: " translations_frontend.name = ?", CondValue: record[0]},
				models.WhereCondFn{Condition: " translations_frontend.locale = ?", CondValue: locale},
			)
			result, _ := models.GetAppFrontendTranslationFn(arrCond, false)
			if len(result) < 1 {
				// start perform insert action if the data is not exists in current db
				// fmt.Println(locale, record[0], record[i])
				err = models.AddFrontendTranslation(locale, "common", record[0], record[i])
				if err != nil {
					models.Rollback(tx)
					appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "AddFrontendTranslation_fail"}, nil)
					return
				}
				// end perform insert action if the data is not exists in current db
			}
			// 	err := models.AddTranslation(tx, fileHeader[i], form.Type, record[0], record[i])
			// 	if err != nil {
			// 		models.Rollback(tx)
			// 		appG.ResponseError(err)
			// 		return
			// 	}
		}

		err = models.Commit(tx)
		if err != nil {
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: err.Error()}, nil)
			return
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
}

// ExchangeCallbackForm struct
type ExchangeCallbackForm struct {
	DocNo string `form:"doc_no" json:"doc_no" valid:"Required;"`
}

func DecryptMemberPrivateKey(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	// get ent_member with d_pk is null
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: "ent_member.d_pk IS NULL OR ent_member.d_pk = ?", CondValue: ""},
		models.WhereCondFn{Condition: "ent_member.status = ?", CondValue: "A"},
	)
	arrEntMember, _ := models.GetEntMemberListFn(arrCond, false)
	if len(arrEntMember) > 0 {
		for _, v := range arrEntMember {
			if v.PrivateKey != "" {
				// decrypt private key
				decryptedPrivateKey, err := util.RsaDecryptPKCS1v15(v.PrivateKey)
				if err != nil {
					continue
				}

				// update decrypted private key into ent_member
				arrUpdateEntMember := make([]models.WhereCondFn, 0)
				arrUpdateEntMember = append(arrUpdateEntMember,
					models.WhereCondFn{Condition: " ent_member.id = ? ", CondValue: v.ID},
				)
				updateColumn := map[string]interface{}{
					"d_pk": decryptedPrivateKey,
				}
				_ = models.UpdatesFn("ent_member", arrUpdateEntMember, updateColumn, false)
			}
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

func RecalSlsMasterSpentUsdtAmt(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	// reset total_nv to 0
	arrUpdateEntMember := make([]models.WhereCondFn, 0)
	updateColumn := map[string]interface{}{
		"total_nv": 0,
	}
	_ = models.UpdatesFn("sls_master", arrUpdateEntMember, updateColumn, false)

	arrMembers, _ := models.GetMembersWithSales(false)

	for _, arrMembersV := range arrMembers {
		var memID = arrMembersV.MemID
		// get members active normal contract sales (sort by asc)
		arrSlsMasterFn := make([]models.WhereCondFn, 0)
		arrSlsMasterFn = append(arrSlsMasterFn,
			models.WhereCondFn{Condition: "sls_master.member_id = ? ", CondValue: memID},
			models.WhereCondFn{Condition: "sls_master.action = ? ", CondValue: "CONTRACT"},
			models.WhereCondFn{Condition: "sls_master.status IN(?,'P') ", CondValue: "AP"},
			models.WhereCondFn{Condition: "sls_master.grp_type = ? ", CondValue: "0"}, //double confirm with quik
		)
		arrSlsMaster, err := models.GetSlsMasterAscFn(arrSlsMasterFn, "", false)
		if err != nil {
			base.LogErrorLog("apiController:RecalSlsMasterSpentUsdtAmt()", "GetSlsMasterAscFn():1", err.Error(), true)
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "something_went_wrong"}, nil)
			return
		}

		// loop member's sales
		if len(arrSlsMaster) > 0 {
			for _, arrSlsMasterV := range arrSlsMaster {
				// calculate total spent amount
				createdAtDt := arrSlsMasterV.CreatedAt
				createdAt := createdAtDt.Format("2006-01-02 15:04:05")
				totAvailableExchangedUsdtAmt, _ := product_service.CalTotalAvailableExchangedUsdtAmt(memID, createdAt)

				var curSpentUsdtAmount = 0.00
				if totAvailableExchangedUsdtAmt < arrSlsMasterV.TotalBv {
					// if available exchange amount is sufficient, then spending amount will be contract full amount
					curSpentUsdtAmount = totAvailableExchangedUsdtAmt
				} else {
					// if available exchange amount is insufficient, then spending amount will be available exchange amount
					curSpentUsdtAmount = arrSlsMasterV.TotalBv
				}

				// update total_nv
				arrUpdateEntMember2 := make([]models.WhereCondFn, 0)
				arrUpdateEntMember2 = append(arrUpdateEntMember2,
					models.WhereCondFn{Condition: " sls_master.id = ? ", CondValue: arrSlsMasterV.ID},
				)
				updateColumn2 := map[string]interface{}{
					"total_nv": curSpentUsdtAmount,
				}
				_ = models.UpdatesFn("sls_master", arrUpdateEntMember2, updateColumn2, false)
			}
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

func RecalEwtExchangeSpentUsdtAmt(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	// reset total_usdt_spent to 0
	arrUpdateEwtExchange := make([]models.WhereCondFn, 0)
	updateColumn := map[string]interface{}{
		"total_usdt_spent": 0,
	}
	_ = models.UpdatesFn("ewt_exchange", arrUpdateEwtExchange, updateColumn, false)

	arrCond := make([]models.WhereCondFn, 0)
	arrEwtExchange, _ := models.GetEwtExchangeFn(arrCond, "", false)

	if len(arrEwtExchange) > 0 {
		for _, arrEwtExchangeV := range arrEwtExchange {
			var totUsdtSpentAmount = 0.00

			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewt_detail.doc_no = ? ", CondValue: arrEwtExchangeV.DocNo},
				models.WhereCondFn{Condition: " ewt_detail.ewallet_type_id = ? ", CondValue: 1}, // USDT
				models.WhereCondFn{Condition: " ewt_detail.total_out > ? ", CondValue: 0},
			)
			arrEwtDetail, _ := models.GetEwtDetailFn(arrCond, false)

			if len(arrEwtDetail) > 0 {
				for _, arrEwtDetailV := range arrEwtDetail {
					totUsdtSpentAmount += arrEwtDetailV.TotalOut
				}
			}

			if totUsdtSpentAmount > 0 {
				// update total_nv
				arrUpdateEwtExchange2 := make([]models.WhereCondFn, 0)
				arrUpdateEwtExchange2 = append(arrUpdateEwtExchange2,
					models.WhereCondFn{Condition: " ewt_exchange.id = ? ", CondValue: arrEwtExchangeV.ID},
				)
				updateColumn2 := map[string]interface{}{
					"total_usdt_spent": totUsdtSpentAmount,
				}
				_ = models.UpdatesFn("ewt_exchange", arrUpdateEwtExchange2, updateColumn2, false)
			}
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

type BlockchainTransStruct struct {
	To         string `json:"to"`
	CryptoType string `json:"crypto_type"`
	TransType  string `json:"trans_type"`
	Hash       string `json:"hash"`
	TotalIn    string `json:"total_in"`
	TotalOut   string `json:"total_out"`
	Remark     string `json:"remark"`
	CreatedAt  string `json:"created_at"`
	Deduct     string `json:"deduct"`
}
type ProcessSaveMemberBlockchainTransRecordsFromApiForm struct {
	TxData []BlockchainTransStruct `json:"tx_data"`
}

func ProcessSaveMemberBlockchainTransRecordsFromApi(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form ProcessSaveMemberBlockchainTransRecordsFromApiForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	// fmt.Println(form.TxData)
	// var arrBlockchainTransList []arrBlockchainTransStruct
	// err := json.Unmarshal([]byte(form.TxData), &arrBlockchainTransList)

	// if err != nil {
	// 	// base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-failed_to_decode_TxData", err.Error(), form.TxData, true)
	// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: err.Error()}, nil)
	// 	return
	// }
	arrBlockchainTransList := form.TxData
	if len(arrBlockchainTransList) > 0 {
		for _, arrBlockchainTransListV := range arrBlockchainTransList {
			tx := models.Begin()
			if arrBlockchainTransListV.Hash == "" {
				tx.Rollback()
				// base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-TxData_to_is_required", err.Error(), form.TxData, true)
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "hash_is_required"}, nil)
				return
			}
			if arrBlockchainTransListV.To == "" {
				tx.Rollback()
				// base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-TxData_to_is_required", err.Error(), form.TxData, true)
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "to_is_required_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}
			if arrBlockchainTransListV.CryptoType == "" {
				tx.Rollback()
				// base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-TxData_CryptoType_is_required", err.Error(), form.TxData, true)
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "to_is_required_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}
			if arrBlockchainTransListV.TransType == "" {
				tx.Rollback()
				// base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-TxData_to_is_required", err.Error(), form.TxData, true)
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "trans_type_is_required_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}

			if arrBlockchainTransListV.TotalIn == "" && arrBlockchainTransListV.TotalOut == "" {
				tx.Rollback()
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "total_in_or_total_out_is_required_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}

			if arrBlockchainTransListV.TotalIn != "" && arrBlockchainTransListV.TotalOut != "" {
				tx.Rollback()
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "only_total_in_or_total_out_is_required_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}

			totalIn := float64(0)
			if arrBlockchainTransListV.TotalIn != "" {
				totalInBigFloat, err := float.SetString(arrBlockchainTransListV.TotalIn)
				if err != nil {
					tx.Rollback()
					appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "total_in_" + err.Error() + "_on_" + arrBlockchainTransListV.Hash}, nil)
					return
				}
				totalInRst := totalInBigFloat.Float64()
				totalIn = totalInRst
			}

			totalOut := float64(0)
			if arrBlockchainTransListV.TotalOut != "" {
				totalOutBigFloat, err := float.SetString(arrBlockchainTransListV.TotalOut)
				if err != nil {
					tx.Rollback()
					appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "total_out_" + err.Error() + "_on_" + arrBlockchainTransListV.Hash}, nil)
					return
				}
				totalOutRst := totalOutBigFloat.Float64()
				totalOut = totalOutRst
			}

			arrCond := make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " crypto_address = ?", CondValue: arrBlockchainTransListV.To},
				models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
			)
			entMemberCrypto, err := models.GetEntMemberCryptoFn(arrCond, false)
			if err != nil {
				tx.Rollback()
				base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-GetEntMemberCryptoFn_failed", err.Error(), arrCond, true)
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "invalid_to_" + err.Error() + "_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}
			if entMemberCrypto == nil {
				tx.Rollback()
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "invalid_to_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}

			arrCond = make([]models.WhereCondFn, 0)
			arrCond = append(arrCond,
				models.WhereCondFn{Condition: " ewallet_type_code = ?", CondValue: arrBlockchainTransListV.CryptoType},
				models.WhereCondFn{Condition: " control = ?", CondValue: "blockchain"},
				models.WhereCondFn{Condition: " status = ?", CondValue: "A"},
			)
			ewtSetup, err := models.GetEwtSetupFn(arrCond, "", false)
			if err != nil {
				tx.Rollback()
				base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-GetEwtSetupFn_failed", err.Error(), arrCond, true)
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "invalid_crypto_type_" + err.Error() + "_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}
			if ewtSetup == nil {
				tx.Rollback()
				appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "invalid_crypto_type_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}
			// createdAt := base.GetCurrentDateTimeT()
			createdAt := base.GetCurrentTime("2006-01-02 15:04:05")
			if arrBlockchainTransListV.CreatedAt != "" {
				// timeStampT, err := time.Parse("2006-01-02 15:04:05", arrBlockchainTransListV.CreatedAt)
				// if err != nil {
				// 	tx.Rollback()
				// 	base.LogErrorLog("ProcessSaveMemberBlockchainTransRecordsFromApi-time_parse_failed", err.Error(), arrBlockchainTransListV.CreatedAt, true)
				// 	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "invalid_created_at_" + err.Error() + "_on_" + arrBlockchainTransListV.Hash}, nil)
				// 	return
				// }
				createdAt = arrBlockchainTransListV.CreatedAt
			}

			arrData := wallet_service.SaveMemberBlockchainTransRecordsFromApiStruct{
				EntMemberID:     entMemberCrypto.MemberID,
				EwalletTypeID:   ewtSetup.ID,
				Status:          "AP",
				TransactionType: arrBlockchainTransListV.TransType,
				TotalIn:         totalIn,
				TotalOut:        totalOut,
				ConversionRate:  1,
				HashValue:       arrBlockchainTransListV.Hash,
				Remark:          arrBlockchainTransListV.Remark,
				LogOnly:         0,
				CreatedAt:       createdAt,
				EwalletTypeCode: arrBlockchainTransListV.CryptoType,
				Deduct:          arrBlockchainTransListV.Deduct,
				// ConvertedTotalIn:  totalIn,
				// ConvertedTotalOut: totalOut,
			}
			err = wallet_service.SaveMemberBlockchainTransRecordsFromApi(tx, arrData)

			if err != nil {
				tx.Rollback()
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "save_tx_failed_on_" + arrBlockchainTransListV.Hash}, nil)
				return
			}
			tx.Commit()
		}
	}

	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}

type UpdateWalletDataParam struct {
	WalletAddress    string  `json:"wallet_address" form:"wallet_address" valid:"Required;"`
	TotalBalance     float64 `json:"total_balance" form:"total_balance" valid:"Required;"`
	TotalMinedAmount float64 `json:"total_mined_amount" form:"total_mined_amount" valid:"Required;"`
}

func UpdateWalletDataApi(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form UpdateWalletDataParam
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: msg[0]}, nil)
		return
	}

	tx := models.Begin()

	rst := sales_service.UpdateWalletData(tx, form.WalletAddress, sales_service.WalletData{
		TotalBalance:     form.TotalBalance,
		TotalMinedAmount: form.TotalMinedAmount,
	})
	if rst != "" {
		tx.Rollback()
		appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: rst}, nil)
		return
	}

	tx.Commit()
	appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
	return
}
