package models

import (
	"math"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// Translations struct
type Translations struct {
	ID     int    `gorm:"primary_key" json:"id"`
	Locale string `json:"locale"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

// Translations struct
type TranslationsV2 struct {
	ID         int       `gorm:"primary_key" json:"id"`
	Locale     string    `json:"locale"`
	Group      string    `json:"group"`
	Type       string    `json:"type"`
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

// TranslationsFrontend struct
type Translations_Frontend struct {
	ID         int       `gorm:"primary_key" json:"id"`
	Locale     string    `json:"locale"`
	Group      string    `json:"group"`
	Type       string    `json:"type"`
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

// AddTranslation add translation
func AddTranslation(tx *gorm.DB, locale, transType, name, value string) error {
	trans := Translations{
		Locale: locale,
		Type:   transType,
		Name:   name,
		Value:  value,
	}

	if err := tx.Create(&trans).Error; err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// GetTranslationByType func
func GetTranslationByType(transType string) ([]*Translations, error) {
	var trans []*Translations

	err := db.Where("type = ?", transType).Find(&trans).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return trans, nil
}

// GetTranslationByLocale func
func GetTranslationByLocale(transType, locale string) ([]*Translations, error) {
	var trans []*Translations

	if locale == "" {
		return GetTranslationByType(transType)
	}

	err := db.Where("type = ? AND locale = ?", transType, locale).Find(&trans).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return trans, nil
}

// GetTranslationByName func
func GetTranslationByName(locale, transType, name string) (*Translations, error) {
	var trans Translations

	err := db.Where("locale = ? AND type = ? AND NAME = ?", locale, transType, name).First(&trans).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &trans, nil
}

// DeleteTranslationsBytype func
func DeleteTranslationsBytype(tx *gorm.DB, transType string) error {

	err := tx.Where("type = ?", transType).Delete(&Translations{}).Error

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// ExistTranslationName func
func ExistTranslationName(locale, transType, name string) bool {
	trans, err := GetTranslationByName(locale, transType, name)
	if err != nil {
		return false
	}

	if trans == nil {
		return false
	}

	return true

}

// GetTranslationByID func
func GetTranslationByID(id int) (*Translations, error) {
	var trans Translations

	err := db.Where("id = ?", id).First(&trans).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &trans, nil
}

// ExistTranslationID func
func ExistTranslationID(id int) bool {
	trans, err := GetTranslationByID(id)
	if err != nil {
		return false
	}

	if trans == nil {
		return false
	}

	return true

}

// GetTransaltionList func
func GetTransaltionList(page int64, limit int64, transType string) ([]*Translations, int64, float64, int64, error) {
	var (
		trans       []*Translations
		totalPage   float64
		totalRecord int64
		endFlag     int64 = page
	)

	query := db.Table("translations")

	if transType != "" {
		query = query.Where("type = ?", transType)
	}

	// Total Records
	query.Count(&totalRecord)

	if page != 0 {
		page--
	}

	newOffset := page * limit

	// Pagination and limit
	err := query.Order("id").Limit(limit).Offset(newOffset).Find(&trans).Error
	if err != nil {
		return nil, 0, 0, 0, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	totalPage = float64(totalRecord) / float64(limit)
	totalPage = math.Ceil(totalPage)

	if int64(endFlag) >= int64(totalPage) {
		endFlag = 1
	} else {
		endFlag = 0
	}

	return trans, totalRecord, totalPage, endFlag, nil
}

// UpdateValue func
func (t *Translations) UpdateValue(tx *gorm.DB, value string) error {
	t.Value = value
	return SaveTx(tx, t)
}

// GetFrontendTranslation func
func GetFrontendTranslation(locale, group string) map[string]string {
	var trans []*Translations_Frontend

	err := db.Where("locale = ? AND translations_frontend.group = ?", locale, group).Find(&trans).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil
	}

	if err == gorm.ErrRecordNotFound {
		return nil
	}

	text := make(map[string]string)

	for _, v := range trans {
		// fmt.Println("name", v.Name, "value", v.Value)
		text[v.Name] = v.Value
	}

	return text
}

// GetFrontendTranslationByName func
func GetFrontendTranslationByName(locale, group, name string) *Translations_Frontend {
	var trans Translations_Frontend

	err := db.Where("locale = ? AND translations_frontend.group = ? AND name = ?", locale, group, name).First(&trans).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil
	}

	if err == gorm.ErrRecordNotFound {
		return nil
	}

	return &trans
}

//add frontend translation func
func AddFrontendTranslation(locale, group, name, value string) error {
	trans := Translations_Frontend{
		Locale:     locale,
		Group:      group,
		Type:       "label",
		Name:       name,
		Value:      value,
		Created_at: time.Now(),
		Updated_at: time.Now(),
	}

	if err := db.Create(&trans).Error; err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

//add translation func
func AddTranslationV2(locale, group, name, value string) error {
	trans := TranslationsV2{
		Locale:     locale,
		Group:      "all",
		Type:       group,
		Name:       name,
		Value:      value,
		Created_at: time.Now(),
		Updated_at: time.Now(),
	}

	if err := db.Table("translations").Create(&trans).Error; err != nil {
		return &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return nil
}

// AppTranslationsFrontend struct
type AppTranslationsFrontend struct {
	Key   string `json:"key" gorm:"column:name"`
	Value string `json:"value" gorm:"column:value"`
}

// func GetAppFrontendTranslationFn
func GetAppFrontendTranslationFn(arrCond []WhereCondFn, debug bool) ([]*AppTranslationsFrontend, error) {
	var result []*AppTranslationsFrontend
	tx := db.Table("translations_frontend")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}

// func GetAppFrontendTranslationFn
func GetLastUpdateAppFrontendTranslationFn(arrCond []WhereCondFn, debug bool) (*Translations_Frontend, error) {
	var result Translations_Frontend
	tx := db.Table("translations_frontend").
		Order("updated_at DESC")
	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return &result, nil
}
