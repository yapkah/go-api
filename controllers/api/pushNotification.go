package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartblock/gta-api/pkg/app"
	"github.com/smartblock/gta-api/pkg/base"
)

type ProcessPNFrom struct {
	Event       string `form:"event" json:"event" valid:"Required"` // createNewGroup, subscribeToGroup, unsubscribeFromGroup, sendPushNotificationInGroup, sendPushNotificationInd
	Os          string `form:"os" json:"os"`
	GroupName   string `form:"group_name" json:"group_name"`
	EntMemberID int    `form:"ent_member_id" json:"ent_member_id"`
	Subject     string `form:"subject" json:"subject"`
	Msg         string `form:"msg" json:"msg"`
	CusMsg      string `form:"cus_msg" json:"cus_msg"`
	RegID       string `form:"reg_id" json:"reg_id"`
}

// func ProcessPN
func ProcessPN(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form ProcessPNFrom
	)
	//validate input
	ok, msg := app.BindAndValid(c, &form)

	if ok == false {
		appG.Response(0, http.StatusOK, msg[0], "")
		return
	}

	switch strings.ToLower(form.Event) {
	case "createnewgroup":
		if form.Os != "" && form.GroupName != "" {
			arrCallCreateNewPushNotificationGroupApi := base.CreateNewPushNotificationGroupStruct{
				GroupName: form.GroupName,
				Os:        form.Os,
			}
			err := base.CallCreateNewPushNotificationGroupApi(arrCallCreateNewPushNotificationGroupApi)
			if err != nil {
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: err.Error()}, nil)
				return
			}
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
			return
		} else {
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "PNOs_or_PNGroupName_is_missing"}, nil)
			return
		}
	case "subscribetogroup":
		if form.RegID != "" && form.GroupName != "" {
			arrCallSubscribeToGroupApi := base.CallSubscribePushNotificationToGroupApiStruct{
				GroupName: form.GroupName,
				RegID:     form.RegID,
				Os:        form.Os,
			}
			err := base.CallSubscribePushNotificationToGroupApi(arrCallSubscribeToGroupApi)
			if err != nil {
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: err.Error()}, nil)
				return
			}
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
			return
		} else {
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "RegID_or_GroupName_or_Os_is_missing"}, nil)
			return
		}
	case "unsubscribefromgroup":
		if form.RegID != "" && form.GroupName != "" {
			arrCallUnsubscribeFromGroupApi := base.CallUnsubscribePushNotificationFromGroupApiStruct{
				GroupName: form.GroupName,
				RegID:     form.RegID,
				Os:        form.Os,
			}
			err := base.CallUnsubscribePushNotificationFromGroupApi(arrCallUnsubscribeFromGroupApi)
			if err != nil {
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: err.Error()}, nil)
				return
			}
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
			return
		} else {
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "RegID_or_GroupName_or_Os_is_missing"}, nil)
			return
		}
	case "sendpushnotificationind":
		if form.Subject != "" && form.RegID != "" {
			arrCallSendPushNotificationIndApi := base.CallSendPushNotificationIndApiStruct{
				RegID:   form.RegID,
				Os:      form.Os,
				Subject: form.Subject,
				Msg:     form.Msg,
				CusMsg:  form.CusMsg,
			}

			err := base.CallSendPushNotificationIndApi(arrCallSendPushNotificationIndApi)
			if err != nil {
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: err.Error()}, nil)
				return
			}
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
			return
		} else {
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "RegID_or_GroupName_or_Os_is_missing"}, nil)
			return
		}
	case "sendpushnotificationingroup":
		if form.Subject != "" && form.RegID != "" {
			arrCallSendPushNotificationIndApi := base.CallSendPushNotificationInGroupApiStruct{
				GroupName: form.GroupName,
				Subject:   form.Subject,
				Msg:       form.Msg,
				CusMsg:    form.CusMsg,
			}

			err := base.CallSendPushNotificationInGroupApi(arrCallSendPushNotificationIndApi)
			if err != nil {
				appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: err.Error()}, nil)
				return
			}
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "success"}, nil)
			return
		} else {
			appG.ResponseV2(0, http.StatusBadRequest, app.MsgStruct{Msg: "RegID_or_GroupName_or_Os_is_missing"}, nil)
			return
		}
	default:
		fmt.Printf("do nothing")
	}
}
