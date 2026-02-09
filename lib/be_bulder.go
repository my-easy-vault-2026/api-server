package lib

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"shared-modules/common"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type BEBuilder struct {
	i18    *I18N
	logger Logger
}

func NewBEBuilder(i18 *I18N, logger Logger) *BEBuilder {
	return &BEBuilder{
		i18:    i18,
		logger: logger,
	}
}

func (be *BEBuilder) WrapBusinessError(ctx context.Context, err error) *BusinessError {
	if err != nil {
		err1, ok := err.(*BusinessError)
		if ok {
			return err1
		} else {
			return be.NewBusinessError(ctx, common.CODE_SYSTEM_ERROR, err.Error())
		}
	}
	return nil
}

func (be *BEBuilder) GetErrorMsg(ctx context.Context, code int, lang string) (msg string) {
	if lang == "" {
		lang = strings.ToLower(fmt.Sprint(ctx.Value(common.CTX_KEY_LANGUAGE)))
	}
	bundle, ok := be.i18.Bundles[lang]
	if !ok || lang == "" {
		bundle = be.i18.Bundles["en"]
	}

	eMsg, err := i18n.NewLocalizer(
		bundle,
	).Localize(&i18n.LocalizeConfig{
		MessageID: "Code" + strconv.Itoa(code),
	})

	// 如果該語言找不到對應，則以英文為主
	if err != nil {
		be.logger.Warnf("text msg not set for code %d in bundle %v", code, err)
		bundle = be.i18.Bundles["en"]
		eMsg, err = i18n.NewLocalizer(
			bundle,
		).Localize(&i18n.LocalizeConfig{
			MessageID: "Code" + strconv.Itoa(code),
		})
		if err != nil {
			be.logger.Warnf("text msg not set for code %d in bundle %v", code, err)
			eMsg = "error message not set"
		}
		eMsg = "error message not set"
	}

	return eMsg
}

func (be *BEBuilder) NewBusinessError(ctx context.Context, code int, msg ...string) (ret *BusinessError) {
	if len(msg) != 0 {
		return NewBusinessError(code, msg[0])
	}

	lang := strings.ToLower(fmt.Sprint(ctx.Value(common.CTX_KEY_LANGUAGE)))
	bundle, ok := be.i18.Bundles[lang]
	if !ok || lang == "" {
		bundle = be.i18.Bundles["en"]
	}

	eMsg, err := i18n.NewLocalizer(
		bundle,
	).Localize(&i18n.LocalizeConfig{
		MessageID: "Code" + strconv.Itoa(code),
	})

	// 如果該語言找不到對應，則以英文為主
	if err != nil {
		be.logger.Warnf("text msg not set for code %d in bundle %v", code, err)
		bundle = be.i18.Bundles["en"]
		eMsg, err = i18n.NewLocalizer(
			bundle,
		).Localize(&i18n.LocalizeConfig{
			MessageID: "Code" + strconv.Itoa(code),
		})
		if err != nil {
			be.logger.Warnf("text msg not set for code %d in bundle %v", code, err)
			eMsg = "error message not set"
		}
		eMsg = "error message not set"
	}

	ret = NewBusinessError(
		code,
		eMsg,
	)
	return
}

func (be *BEBuilder) EqualBusinssError(code int, err error) bool {
	if bErr, ok := err.(*BusinessError); ok {
		return bErr.Code() == code
	}
	return false
}

func FormatErrorChainArrow(err error) string {
	return FormatErrorChain(err, " -> ")
}

func FormatErrorChainLink(err error) string {
	return FormatErrorChain(err, "🔗")
}

func FormatErrorChain(err error, separator string) string {
	var result strings.Builder

	for err != nil {
		result.WriteString(err.Error())
		err = errors.Unwrap(err)
		if err != nil {
			result.WriteString(separator)
		}
	}

	return result.String()
}

func FormatErrorChainHierarchical(err error) string {
	var result strings.Builder
	level := 0

	for err != nil {
		if level > 0 {
			result.WriteString("\n")
			result.WriteString(strings.Repeat("  ", level)) // 縮進
			result.WriteString("└─ ")
		}
		result.WriteString(err.Error())
		err = errors.Unwrap(err)
		level++
	}

	return result.String()
}

func GetRootCause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}
