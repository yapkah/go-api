package main

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smartblock/gta-api/jobs"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/pkg/gredis"
	"github.com/smartblock/gta-api/pkg/logging"
	"github.com/smartblock/gta-api/pkg/setting"
)

const sleepTime = 1 * time.Second
const errorSleepTime = 10 * time.Second
const panicSleepTime = 10 * time.Second

func init() {
	setting.Setup()
	models.Setup()
	logging.Setup()
	gredis.Setup()
}

// @title Golang Gin API
// @version 1.0
// @description An example of gin
// @termsOfService https://github.com/smartblock/gta-api
// @license.name MIT
// @license.url https://github.com/smartblock/gta-api/blob/master/LICENSE
func main() {
	gin.SetMode(setting.ServerSetting.RunMode)
	for {
		batch()
		time.Sleep(sleepTime)
	}
}

func batch() {
	arrCond := make([]models.WhereCondFn, 0)
	arrCond = append(arrCond,
		models.WhereCondFn{Condition: " UNIX_TIMESTAMP() >= available_at AND queue = ?", CondValue: "default"},
		models.WhereCondFn{Condition: " attempts < ?", CondValue: 10},
	)
	arrGolangJob, err := models.GetGolangJobsFn(arrCond, false)

	if err != nil {
		base.LogErrorLog("golang job sql error", err.Error(), arrCond, true)
	}

	if len(arrGolangJob) > 0 {
		for _, v1 := range arrGolangJob {
			jobs.RunMainJobs(v1)
		}
	}

	defer handlepanic() // to make it keep running when panic
	// run ur batch
}

func handlepanic() {
	if a := recover(); a != nil {
		base.LogErrorLog("handlepanic error", "error recover", map[string]interface{}{"error": a}, true)
		time.Sleep(panicSleepTime)
	}
}
