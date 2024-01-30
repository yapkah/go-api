package main

import (
	"github.com/gin-gonic/gin"

	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/gredis"
	"github.com/smartblock/gta-api/pkg/logging"
	"github.com/smartblock/gta-api/pkg/setting"
	schedule "github.com/smartblock/gta-api/schedules"
)

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

// =====================================================
// SOLVE UNUNTU SERVER CAN'T RUN PROGRAM IN SECOND
// =====================================================
// ADVISE
// 1. Program run more than in second prefer to run server cronjob. To save server resource.
func main() {
	gin.SetMode(setting.ServerSetting.RunMode)

	// for {
	schedule.RunMainSchedule()
	// }
}
