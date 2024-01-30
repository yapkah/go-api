package translation

import (
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/yapkah/go-api/models"
	"golang.org/x/text/language"
)

// Bundle var
var Bundle *i18n.Bundle

// Localizer struct
type Localizer struct {
	Localizer *i18n.Localizer
	Language  string
}

// Setup func
func Setup() {

	// Bundle = i18n.NewBundle(language.English)
	// Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	langs, err := models.GetLanguageList()
	if err != nil {
		log.Fatalf("translation.Setup [GetLanguageList] err: %v", err)
	}

	for _, lang := range langs {
		trans, err := models.GetTranslationByLocale("api", lang.Locale)
		if err != nil {
			log.Fatalf("translation.Setup [GetTranslationByLocale] err: %v", err)
		}

		var tranStr string

		for _, tran := range trans {
			tranStr = tranStr + tran.Name + " = " + "\"" + tran.Value + "\"\n"
		}

		// Bundle.MustParseMessageFileBytes([]byte(tranStr), lang.Locale+".toml")
	}

}

// NewLocalizer func
func NewLocalizer(langCode string) *Localizer {
	return &Localizer{
		Localizer: i18n.NewLocalizer(Bundle, langCode),
		Language:  langCode,
	}
}

// Trans func
func (l *Localizer) Trans(text string, template map[string]interface{}) string {
	tran, err := l.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    text,
		TemplateData: template,
	})

	if err != nil {
		return text
	}

	return tran
}

// AddMessage func
func AddMessage(locale string, messages []*i18n.Message) error {

	tag, err := language.Parse(locale)
	if err != nil {
		return err
	}

	return Bundle.AddMessages(tag, messages...)
}
