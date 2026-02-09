package middlewares

import (
	"api-server/infra"
	"api-server/lib"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type LanguageMiddleware struct {
	logger    lib.Logger
	i18n      *lib.I18N
	APIRouter infra.Router `name:"api"`
}

type LanguageMiddlewareParams struct {
	fx.In
	Logger    lib.Logger
	I18N      *lib.I18N
	APIRouter infra.Router `name:"api"`
}

func NewLanguageMiddleware(
	p LanguageMiddlewareParams,
) *LanguageMiddleware {
	return &LanguageMiddleware{
		logger:    p.Logger,
		i18n:      p.I18N,
		APIRouter: p.APIRouter,
	}
}

func (lm *LanguageMiddleware) Setup() {
	lm.APIRouter.Use(lm.Handle())
}

func (lm *LanguageMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptLanguage := c.Request.Header.Get("Accept-Language")
		if acceptLanguage == "" {
			acceptLanguage = "en"
		}
		matchLang, err := lm.i18n.MatchLang(acceptLanguage)
		if err != nil {
			lm.logger.Warnf("match lang failed: %v", err)
			matchLang = []string{"en"}
		}

		if len(matchLang) > 0 {
			c.Request.Header.Set("Accept-Language", matchLang[0])
		} else {
			c.Request.Header.Set("Accept-Language", "en")
		}
		c.Next()
	}
}
