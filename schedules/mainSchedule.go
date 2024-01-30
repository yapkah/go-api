package schedule

import (
	"time"

	"github.com/smartblock/gta-api/pkg/base"
	"github.com/smartblock/gta-api/service/notification_service"
	"github.com/smartblock/gta-api/service/sales_service"
	"github.com/smartblock/gta-api/service/trading_service"
)

// const sleepTime = 1 * time.Second
// const errorSleepTime = 10 * time.Second
const panicSleepTime = 10 * time.Second

func RunMainSchedule() {
	go func() {
		// start run function in every 2 second
		for _ = range time.Tick(2 * time.Second) {
			ProcessAutoMatchTrading()
			ProcessLaLigaCallBack()
		}

		// end run function in every 2 second
	}()

	go func() {
		// start run function in every 10 second
		for _ = range time.Tick(10 * time.Second) {
			ProcessSendPushNotificationMsg()
		}

		// end run function in every 10 second
	}()

	// start key point to keep it run
	c := make(chan int)
	<-c
	// end key point to keep it run
}

func ProcessAutoMatchTrading() {
	defer handlepanic()
	trading_service.ProcessAutoMatchTrading(false)
}

func ProcessLaLigaCallBack() {
	defer handlepanic()
	sales_service.ProcessLaligaCallBack(false)
}

func handlepanic() {
	if a := recover(); a != nil {
		base.LogErrorLog("handlepanic error", "error recover", map[string]interface{}{"error": a}, true)
		time.Sleep(panicSleepTime)
	}
}

func ProcessSendPushNotificationMsg() {
	defer handlepanic()
	notification_service.ProcessSendPushNotificationMsg(false)
}
