package web

import (
	"api-server/lib"
	"api-server/services"
	"net/http"
	"shared-modules/common"
	"shared-modules/entities"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type QuoteHandler struct {
	quoteService *services.QuoteService
	logger       lib.Logger
	beBuilder    *lib.BEBuilder
	httpRes      *lib.HttpRes
}

func NewQuoteHandler(quoteService *services.QuoteService, logger lib.Logger, beBuilder *lib.BEBuilder, httpRes *lib.HttpRes) *QuoteHandler {
	return &QuoteHandler{
		quoteService: quoteService,
		logger:       logger,
		beBuilder:    beBuilder,
		httpRes:      httpRes,
	}
}

// @Param   quote   path   string   true   "Quote currency"
// @Param   base    path   string   true   "Base currency"
// @Param			request			body		entities.GetExchangeRateForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Success		0				{object}	entities.ListExchangeRateVO		"data"
// @Router			/web/quote/exchangeRate/:quote/:base [get]
// @Tags			web/quote
func (qh *QuoteHandler) GetExchange(c *gin.Context) {
	form := &entities.GetExchangeRateForm{}

	err := c.ShouldBindQuery(form)
	if err != nil {
		qh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	quoteString := c.Param("quote")
	quote := common.Currency(0).FromString(quoteString)
	baseString := c.Param("base")
	base := common.Currency(0).FromString(baseString)

	if !quote.IsValid() {
		qh.httpRes.ReError(c, http.StatusBadRequest, qh.beBuilder.NewBusinessError(c, common.CODE_NO_SUCH_CURRENCY))
		return
	}

	if !base.IsValid() {
		qh.httpRes.ReError(c, http.StatusBadRequest, qh.beBuilder.NewBusinessError(c, common.CODE_NO_SUCH_CURRENCY))
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		qh.httpRes.ReError(c, http.StatusBadRequest, qh.beBuilder.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	exchangeRates, err := qh.quoteService.GetExchangeRates(c, form.Purpose, quote, base)
	if err != nil {
		qh.httpRes.ReError(c, http.StatusBadRequest, err)
		return
	}

	res := &entities.ExchangeRateVO{
		Rate:      exchangeRates.Rate,
		Base:      exchangeRates.BaseCurrency.String(),
		Quote:     exchangeRates.QuoteCurrency.String(),
		Timestamp: exchangeRates.Timestamp.UnixMilli(),
		Purpose:   exchangeRates.Purpose.String(),
	}

	qh.httpRes.ReData(c, res)
}
