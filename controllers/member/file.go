package member

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/app"
	"github.com/yapkah/go-api/pkg/setting"
	"github.com/yapkah/go-api/service/media_service"
)

// UploadMemberFileForm struct
type UploadMemberFileForm struct {
	Event      string `form:"event" json:"event" valid:"Required;MaxSize(20)"`
	TicketCode string `form:"ticket_code" json:"ticket_code"`
}

// UploadMemberFile function
func UploadMemberFile(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form UploadMemberFileForm
	)

	ok, msg := app.BindAndValid(c, &form)
	if ok == false {
		message := app.MsgStruct{
			Msg: msg[0],
		}
		appG.ResponseV2(0, http.StatusOK, message, nil)
		return
	}

	// get access user from middle ware
	u, ok := c.Get("access_user")
	if !ok {
		// user not found
		message := app.MsgStruct{
			Msg: "something_went_wrong",
		}
		appG.ResponseV2(0, http.StatusUnauthorized, message, nil)
		return
	}
	member := u.(*models.EntMemberMembers)
	envFolderPath := setting.Cfg.Section("server").Key("ENV").String()
	if envFolderPath != "" {
		envFolderPath = "/" + envFolderPath
	}
	if form.Event == "MEM_PROFILE" {
		file, header, err := c.Request.FormFile("avatar_img")
		if err != nil {
			message := app.MsgStruct{
				Msg: "please_upload_avatar_img",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		err = media_service.MediaValidation(file, header, "image")
		if err != nil {
			message := app.MsgStruct{
				Msg: err.Error() + "_for_avatar_img",
			}
			appG.ResponseV2(0, http.StatusOK, message, nil)
			return
		}
		settingID := "upload_image_setting"
		arrMediaSetting, _ := models.GetSysGeneralSetupByID(settingID)
		sizeLimit := arrMediaSetting.SettingValue2
		filename := "profile_upload_pic_" + member.NickName + "_" + strconv.Itoa(int(time.Now().Unix())) + filepath.Ext(header.Filename)
		module := "member/images/profile" + envFolderPath
		prefixName := "profile_upload_pic"
		mediaData, err := media_service.UploadMedia(file, filename, module, prefixName, sizeLimit, "")

		if err != nil {
			appG.ResponseError(err)
			return
		}
		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: " ent_member.id = ? ", CondValue: member.EntMemberID},
		)

		updateColumn := map[string]interface{}{"avatar": mediaData.FullURL, "path": mediaData.FileDirectory}
		err = models.UpdatesFn("ent_member", arrCond, updateColumn, false)

		if err != nil {
			appG.ResponseError(err)
			return
		}
	} else if form.Event == "MEM_KYC" {

		// UploadList struct
		type UploadList struct {
			FullURL       string
			FileDirectory string
		}
		// arrUploadList := map[string]UploadList{}
		// arrUploadList := make([]UploadList, 0)

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "ent_member_kyc.member_id = ?", CondValue: member.EntMemberID},
		)
		existingIC, _ := models.GetEntMemberKycFn(arrCond, false)

		// inputFormKeyList := []string{"self_ic_img"} // old code
		inputFormKeyList := []string{"front_ic_img", "back_ic_img", "self_ic_img"}
		inputFormKeyList2 := make([]string, 0)

		if len(existingIC) > 0 {
			for _, v1 := range inputFormKeyList {
				file, _, _ := c.Request.FormFile(v1)
				if file != nil {
					inputFormKeyList2 = append(inputFormKeyList2, v1)
				}
			}
			inputFormKeyList = inputFormKeyList2
		} else {
			models.ErrorLog("UploadMemberFile-MEM_KYC", "hacker_is_suspected_double_check_with_front_end", nil)
			message := app.MsgStruct{
				Msg: "successs",
			}
			appG.ResponseV2(1, http.StatusOK, message, "")
			return
		}

		if len(inputFormKeyList) > 0 {
			for _, v1 := range inputFormKeyList {
				file, header, err := c.Request.FormFile(v1)
				if err != nil {
					arrTransValue := make(map[string]string)
					arrTransValue["key"] = v1
					message := app.MsgStruct{
						Msg:    "please_upload_:key",
						Params: arrTransValue,
					}
					appG.ResponseV2(0, http.StatusOK, message, "")
					return
				}
				err = media_service.MediaValidation(file, header, "image")
				if err != nil {
					arrTransValue := make(map[string]string)
					arrTransValue["key"] = v1
					message := app.MsgStruct{
						Msg:    err.Error() + "_for_:key",
						Params: arrTransValue,
					}
					appG.ResponseV2(0, http.StatusOK, message, "")
					return
				}
			}
			settingID := "upload_image_setting"
			arrMediaSetting, _ := models.GetSysGeneralSetupByID(settingID)
			sizeLimit := arrMediaSetting.SettingValue2

			updateColumn := make(map[string]interface{}, 0)
			for _, v1 := range inputFormKeyList {
				file, header, err := c.Request.FormFile(v1)

				filename := "upload_kyc_pic_" + member.NickName + "_" + strconv.Itoa(int(time.Now().Unix())) + filepath.Ext(header.Filename)
				module := "member/images/kyc"
				prefixName := "upload_kyc_pic"
				mediaData, err := media_service.UploadMedia(file, filename, module, prefixName, sizeLimit, "")

				if err != nil {
					appG.ResponseError(err)
					return
				}

				// arrUploadList[v1] = UploadList{FullURL: mediaData.FullURL, FileDirectory: mediaData.FileDirectory}
				if v1 == "front_ic_img" {
					updateColumn["file_name_1"] = mediaData.FileDirectory
					updateColumn["file_url_1"] = mediaData.FullURL
				} else if v1 == "back_ic_img" {
					updateColumn["file_name_2"] = mediaData.FileDirectory
					updateColumn["file_url_2"] = mediaData.FullURL
				} else if v1 == "self_ic_img" {
					updateColumn["file_name_3"] = mediaData.FileDirectory
					updateColumn["file_url_3"] = mediaData.FullURL
				}
				// arrUploadList = append(arrUploadList,
				// UploadList{InputFormKey:v1, FullURL: mediaData.FullURL, FileDirectory: mediaData.FileDirectory},
				// )
			}

			if len(existingIC) > 0 {
				arrCond := make([]models.WhereCondFn, 0)
				arrCond = append(arrCond,
					models.WhereCondFn{Condition: " ent_member_kyc.member_id = ? ", CondValue: member.EntMemberID},
					models.WhereCondFn{Condition: " ent_member_kyc.id = ? ", CondValue: existingIC[0].ID},
				)
				// updateColumn := map[string]interface{}{
				// 	"file_name_1": arrUploadList["front_ic_img"].FileDirectory, "file_url_1": arrUploadList["front_ic_img"].FullURL,
				// 	"file_name_2": arrUploadList["back_ic_img"].FileDirectory, "file_url_2": arrUploadList["back_ic_img"].FullURL,
				// 	"file_name_3": arrUploadList["self_ic_img"].FileDirectory, "file_url_3": arrUploadList["self_ic_img"].FullURL,
				// 	// "file_name_1": arrUploadList["self_ic_img"].FileDirectory, "file_url_1": arrUploadList["self_ic_img"].FullURL,
				// 	"updated_by": member.EntMemberID,
				// }

				updateColumn["updated_by"] = member.EntMemberID // comment temporary
				// start this indicate the image upload is failed
				if existingIC[0].Status == "" {
					updateColumn["status"] = "P" // comment temporary
				}
				// end this indicate the image upload is failed

				err := models.UpdatesFn("ent_member_kyc", arrCond, updateColumn, false)

				if err != nil {
					appG.ResponseError(err)
					return
				}
			}
		}
	} else if form.Event == "MEM_SUPPORT_TICKET" {
		if form.TicketCode == "" {
			appG.ResponseV2(0, http.StatusOK, app.MsgStruct{Msg: "ticket_code_is_required"}, "")
			return
		}

		// UploadList struct
		type UploadList struct {
			FullURL       string
			FileDirectory string
		}

		arrCond := make([]models.WhereCondFn, 0)
		arrCond = append(arrCond,
			models.WhereCondFn{Condition: "support_ticket_mast.member_id = ?", CondValue: member.EntMemberID},
			models.WhereCondFn{Condition: "support_ticket_mast.ticket_code = ?", CondValue: form.TicketCode},
		)
		existingST, _ := models.GetSupportTicketMastFn(arrCond, 0, false)

		inputFormKeyList := []string{"issue_img"}
		inputFormKeyList2 := make([]string, 0)

		if len(existingST) > 0 {
			for _, v1 := range inputFormKeyList {
				file, _, _ := c.Request.FormFile(v1)
				if file != nil {
					inputFormKeyList2 = append(inputFormKeyList2, v1)
				}
			}
			inputFormKeyList = inputFormKeyList2
		} else {
			// models.ErrorLog("UploadMemberFile-MEM_SUPPORT_TICKET", "hacker_is_suspected_double_check_with_front_end", nil)
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "invalid_ticket_code"}, "")
			return
		}

		ticketID := existingST[0].ID

		if len(inputFormKeyList) > 0 {
			for _, v1 := range inputFormKeyList {

				form, _ := c.MultipartForm()

				files := form.File[v1]

				for i, v2 := range files {
					// file, header, err := c.Request.FormFile(v1)

					file, err := files[i].Open()

					if err != nil {
						arrTransValue := make(map[string]string)
						arrTransValue["key"] = v1
						message := app.MsgStruct{
							Msg:    "please_upload_:key",
							Params: arrTransValue,
						}
						appG.ResponseV2(0, http.StatusOK, message, "")
						return
					}

					contentType := v2.Header.Get("Content-Type")
					fileType := contentType[:strings.IndexByte(contentType, '/')]
					fileType = strings.Replace(fileType, " ", "", -1)
					fileType = strings.ToLower(fileType)

					if fileType != "image" {
						if fileType != "video" {
							message := app.MsgStruct{
								Msg: "invalid_file_type",
							}
							appG.ResponseV2(0, http.StatusOK, message, "")
							return
						}
					}

					err = media_service.MediaValidation(file, v2, fileType)
					if err != nil {
						arrTransValue := make(map[string]string)
						arrTransValue["key"] = v1
						message := app.MsgStruct{
							Msg:    err.Error() + "_for_:key",
							Params: arrTransValue,
						}
						appG.ResponseV2(0, http.StatusOK, message, "")
						return
					}
				}
			}

			var imgUrl string
			var vidUrl string

			updateColumn := make(map[string]interface{}, 0)
			for _, v1 := range inputFormKeyList {
				form, _ := c.MultipartForm()

				files := form.File[v1]

				for i, v2 := range files {
					// file, header, err := c.Request.FormFile(v1)
					file, _ := files[i].Open()
					header := v2

					contentType := v2.Header.Get("Content-Type")
					fileType := contentType[:strings.IndexByte(contentType, '/')]
					fileType = strings.Replace(fileType, " ", "", -1)
					fileType = strings.ToLower(fileType)

					if fileType == "image" {

						settingID := "upload_image_setting"
						arrMediaSetting, _ := models.GetSysGeneralSetupByID(settingID)
						sizeLimit := arrMediaSetting.SettingValue2

						filename := "upload_support_ticket_pic_" + member.NickName + "_" + strconv.Itoa(int(time.Now().Unix())) + filepath.Ext(header.Filename)
						module := "member/images/support-ticket"
						prefixName := "upload_support_ticket_pic"
						mediaData, err := media_service.UploadMedia(file, filename, module, prefixName, sizeLimit, "")

						if err != nil {
							appG.ResponseError(err)
							return
						}

						imgUrl = imgUrl + mediaData.FullURL + ","
					}

					if fileType == "video" {

						settingID := "upload_video_setting"
						arrMediaSetting, _ := models.GetSysGeneralSetupByID(settingID)
						sizeLimit := arrMediaSetting.SettingValue2

						filename := "upload_support_ticket_vid_" + member.NickName + "_" + strconv.Itoa(int(time.Now().Unix())) + filepath.Ext(header.Filename)
						module := "member/videos/support-ticket"
						prefixName := "upload_support_ticket_vid"
						mediaData, err := media_service.UploadMedia(file, filename, module, prefixName, sizeLimit, "")

						if err != nil {
							appG.ResponseError(err)
							return
						}

						vidUrl = vidUrl + mediaData.FullURL + ","
					}
				}
			}

			if len(existingST) > 0 {
				imgFile := strings.TrimSuffix(imgUrl, ",")
				vidFile := strings.TrimSuffix(vidUrl, ",")
				updateColumn["file_url_1"] = imgFile
				updateColumn["file_url_2"] = vidFile

				// add new record
				arrSupportTicketDet := models.SupportTicketDet{
					TicketID:  ticketID,
					FileURL1:  imgFile,
					FileURL2:  vidFile,
					CreatedBy: strconv.Itoa(member.EntMemberID),
					CreatedAt: time.Now(),
				}

				models.AddSupportTicketDetWithoutTx(arrSupportTicketDet)

				// arrCond = make([]models.WhereCondFn, 0)
				// arrCond = append(arrCond,
				// 	models.WhereCondFn{Condition: "support_ticket_det.ticket_id = ?", CondValue: existingST[0].ID},
				// 	models.WhereCondFn{Condition: "support_ticket_det.created_by = ?", CondValue: member.EntMemberID},
				// )
				// existingSTD, _ := models.GetSupportTicketDetFn(arrCond, 1, false)

				// if len(existingSTD) > 0 {
				// 	if existingSTD[0].FileURL1 != "" || existingSTD[0].FileURL2 != "" {
				// 		// add new record
				// 		arrSupportTicketDet := models.SupportTicketDet{
				// 			TicketID:  ticketID,
				// 			FileURL1:  imgFile,
				// 			FileURL2:  vidFile,
				// 			CreatedBy: strconv.Itoa(member.EntMemberID),
				// 			CreatedAt: time.Now(),
				// 		}

				// 		models.AddSupportTicketDetWithoutTx(arrSupportTicketDet)
				// 	} else {
				// 		arrCond := make([]models.WhereCondFn, 0)
				// 		arrCond = append(arrCond,
				// 			models.WhereCondFn{Condition: " support_ticket_det.id = ? ", CondValue: existingSTD[0].ID},
				// 			models.WhereCondFn{Condition: "support_ticket_det.created_by = ?", CondValue: member.EntMemberID},
				// 		)

				// 		err := models.UpdatesFn("support_ticket_det", arrCond, updateColumn, false)

				// 		if err != nil {
				// 			appG.ResponseError(err)
				// 			return
				// 		}
				// 	}

				// }
			}
		} else {
			appG.ResponseV2(1, http.StatusOK, app.MsgStruct{Msg: "please_upload_image"}, "")
			return
		}
	} else {
		message := app.MsgStruct{
			Msg: "success",
		}
		appG.ResponseV2(1, http.StatusOK, message, "")
		return
	}

	message := app.MsgStruct{
		Msg: "success",
	}
	appG.ResponseV2(1, http.StatusOK, message, "")
	return
}
