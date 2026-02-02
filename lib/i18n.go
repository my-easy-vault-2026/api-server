package lib

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"shared-modules/common"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type I18N struct {
	Bundles map[string]*i18n.Bundle
	env     *Env
	logger  Logger
}

func NewI18N(env *Env, logger Logger) *I18N {
	i18 := &I18N{
		Bundles: make(map[string]*i18n.Bundle, 0),
		env:     env,
		logger:  logger,
	}

	i18.Bundles = make(map[string]*i18n.Bundle, 0)

	configPath := env.I18NConfigPath

	// TODO: i18n setting
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "en.toml")
	i18.Bundles["en"] = bundle
	i18.Bundles["en_us"] = bundle
	i18.Bundles["en-us"] = bundle
	i18.Bundles["en_uk"] = bundle
	i18.Bundles["en-uk"] = bundle

	bundle = i18n.NewBundle(language.SimplifiedChinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "zh_hans.toml")
	i18.Bundles["zh_cn"] = bundle
	i18.Bundles["zh-cn"] = bundle
	i18.Bundles["zh"] = bundle
	i18.Bundles["zh_hans"] = bundle
	i18.Bundles["zh-hans"] = bundle

	bundle = i18n.NewBundle(language.TraditionalChinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "zh_hant.toml")
	i18.Bundles["zh_tw"] = bundle
	i18.Bundles["zh-tw"] = bundle
	i18.Bundles["zh_hant"] = bundle
	i18.Bundles["zh-hant"] = bundle

	bundle = i18n.NewBundle(language.Japanese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "ja.toml")
	i18.Bundles["ja"] = bundle

	bundle = i18n.NewBundle(language.Korean)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "ko.toml")
	i18.Bundles["ko"] = bundle

	bundle = i18n.NewBundle(language.French)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "fr.toml")
	i18.Bundles["fr"] = bundle

	bundle = i18n.NewBundle(language.German)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "de.toml")
	i18.Bundles["de"] = bundle

	bundle = i18n.NewBundle(language.Spanish)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "es.toml")
	i18.Bundles["es"] = bundle

	bundle = i18n.NewBundle(language.Portuguese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "pt.toml")
	i18.Bundles["pt"] = bundle

	bundle = i18n.NewBundle(language.Vietnamese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "vi.toml")
	i18.Bundles["vi"] = bundle

	bundle = i18n.NewBundle(language.Russian)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "ru.toml")
	i18.Bundles["ru"] = bundle

	bundle = i18n.NewBundle(language.Arabic)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(configPath + "ar.toml")
	i18.Bundles["ar"] = bundle
	return nil
}

func (i *I18N) Translate(ctx context.Context, msg common.TranslateMsg, value ...string) string {

	templateData := map[string]interface{}{}

	if len(value) > 0 {
		for i, v := range value {
			templateData[fmt.Sprintf("value%d", i+1)] = v
		}
	}

	bundle := i.Bundles["en"]
	c, ok := ctx.(*gin.Context)
	if ok {
		al := strings.ToLower(c.Request.Header.Get("Accept-Language"))
		if b, ok := i.Bundles[al]; ok {
			bundle = b
		}
	}

	str, err := i18n.NewLocalizer(
		bundle,
		language.English.String(),
	).Localize(&i18n.LocalizeConfig{
		MessageID:    string(msg),
		TemplateData: templateData,
	})

	if err != nil {
		i.logger.Errorf("text msg not set: [%s], ", string(msg), err)
		return string(msg)
	}

	return str
}

func (i *I18N) TranslateByCode(ctx context.Context, language common.Language, code int, value ...string) string {

	templateData := map[string]interface{}{}

	if len(value) > 0 {
		for i, v := range value {
			templateData[fmt.Sprintf("value%d", i+1)] = v
		}
	}

	bundle := i.Bundles["en"]

	al := strings.ToLower(language.String())
	if b, ok := i.Bundles[al]; ok {
		bundle = b
	}

	str, err := i18n.NewLocalizer(
		bundle,
	).Localize(&i18n.LocalizeConfig{
		MessageID:    "Code" + strconv.Itoa(code),
		TemplateData: templateData,
	})

	if err != nil {
		// logger.Errorf("text msg not set: [%s], ", string(code), err)
		// return string(code)
	}

	return str
}

func (i *I18N) MatchLang(header string) ([]string, error) {
	var matcher = language.NewMatcher([]language.Tag{
		language.English,
		language.AmericanEnglish,
		language.BritishEnglish,
		language.SimplifiedChinese,
		language.TraditionalChinese,
		language.Japanese,
		language.Korean,
		language.French,
		language.German,
		language.Spanish,
		language.Portuguese,
		language.Russian,
		language.Vietnamese,
		language.Arabic,
	})

	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, len(tags))
	for _, tag := range tags {
		if _, _, c := matcher.Match(tag); c == language.Exact {
			res = append(res, tag.String())
		}
	}
	return res, nil
}
