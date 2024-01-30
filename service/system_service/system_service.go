package system_service

import (
	"strconv"
	"time"

	"github.com/smartblock/gta-api/helpers"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/base"
)

type GetFaqListParam struct {
	Page int64
}

type FaqStruct struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func GetFaqList(param GetFaqListParam, langCode string) (interface{}, string) {
	var arrSysFaqFn = make([]models.WhereCondFn, 0)
	arrSysFaqFn = append(arrSysFaqFn,
		models.WhereCondFn{Condition: " sys_faq.locale = ?", CondValue: langCode},
		models.WhereCondFn{Condition: " sys_faq.status = ?", CondValue: "A"},
	)

	var arrSysFaq, err = models.GetSysFaqFn(arrSysFaqFn, false)
	if err != nil {
		base.LogErrorLog("systemService:GetFaqList():GetSysFaqFn():1", err.Error(), map[string]interface{}{"condition": arrSysFaqFn}, true)
		return nil, "something_went_wrong"
	}

	var arrListingData = []interface{}{}

	if len(arrSysFaq) > 0 {
		for _, arrSysFaqV := range arrSysFaq {
			arrListingData = append(arrListingData,
				FaqStruct{
					Title:   arrSysFaqV.Title,
					Content: arrSysFaqV.Content,
				},
			)
		}
	}

	page := base.Pagination{
		Page:    param.Page,
		DataArr: arrListingData,
	}

	arrDataReturn := page.PaginationInterfaceV1()

	return arrDataReturn, ""
}

type AboutUsDetails struct {
	SocialMedia []AboutUsSocialMedia `json:"social_media"`
}

type AboutUsSocialMedia struct {
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
	Url      string `json:"url"`
}

func GetAboutUsDetails(langCode string) (interface{}, string) {
	var arrEntSocialMediaFn = make([]models.WhereCondFn, 0)
	arrEntSocialMediaFn = append(arrEntSocialMediaFn,
		models.WhereCondFn{Condition: " ent_social_media.status = ?", CondValue: "A"},
	)

	var arrEntSocialMedia, err = models.GetEntSocialMediaFn(arrEntSocialMediaFn, false)
	if err != nil {
		base.LogErrorLog("systemService:GetAboutUsDetails():GetEntSocialMediaFn():1", err.Error(), map[string]interface{}{"condition": arrEntSocialMediaFn}, true)
		return nil, "something_went_wrong"
	}

	var (
		arrAboutUsDetails = AboutUsDetails{}
		arrSocialMedia    = []AboutUsSocialMedia{}
	)

	if len(arrEntSocialMedia) > 0 {
		for _, arrEntSocialMediaV := range arrEntSocialMedia {
			arrSocialMedia = append(arrSocialMedia,
				AboutUsSocialMedia{
					Name:     helpers.TranslateV2(arrEntSocialMediaV.Name, langCode, map[string]string{}),
					IconPath: arrEntSocialMediaV.IconPath,
					Url:      arrEntSocialMediaV.Url,
				},
			)
		}
	}

	arrAboutUsDetails.SocialMedia = arrSocialMedia

	return arrAboutUsDetails, ""
}

func GetCurrentServerTime() string {
	t := time.Now().Unix()
	return strconv.FormatInt(int64(t), 10)
}
