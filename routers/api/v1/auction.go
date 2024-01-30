package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/controllers/auction"
)

// Auction func
func Auction(route *gin.RouterGroup) {
	route.POST("getMemberDetails", auction.GetMemberDetails) // get member basic info

	group := route.Group("/push-notification")
	{
		group.POST("msg/send", auction.ProcessSendIndPushNotification)         // send 1 push notification
		group.POST("msg/batch/send", auction.ProcessSendBatchPushNotification) // send batch push notification
		group.GET("list", auction.GetMemberPushNotificationList)               // get member push notification list
		group.POST("/msg/remove", auction.RemoveMemberPushNotification)        // remove member push notification list
	}
}
