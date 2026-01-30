package web

import (
	"api-server/lib"
	"api-server/services"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
)

type QuoteHandler struct {
	quoteService *services.QuoteService
	logger       lib.Logger
}

func NewQuoteHandler(quoteService *services.QuoteService, logger lib.Logger) *QuoteHandler {
	return &QuoteHandler{
		quoteService: quoteService,
		logger:       logger,
	}
}

// @Param			request			body		entities.ListExchangeRateForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Success		0				{object}	entities.ListExchangeRateVO		"data"
// @Router			/web/quote/exchangeRates/list [post]
// @Description	List real-time exchange rate data.
// @Tags			web/quote
func (qh *QuoteHandler) ListExchangeRate(c *gin.Context) {
	form := &entities.ListExchangeRateForm{}

	err := c.ShouldBindJSON(form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	if !common.Currency(0).FromString(form.BaseCurrency).IsValid() {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_NO_SUCH_CURRENCY))
		return
	}

	for _, quoteCurrency := range form.QuoteCurrencies {
		if !common.Currency(0).FromString(quoteCurrency).IsValid() {
			utils.ReError(c, utils.NewBusinessError(c, common.CODE_NO_SUCH_CURRENCY))
			return
		}
	}

	exchangeRates, err := qh.quoteService.ListExchangeRate(c, form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	res := &entities.ListExchangeRateVO{
		Records: make([]*entities.ExchangeRateVO, len(exchangeRates)),
	}
	for i, exchangeRate := range exchangeRates {
		res.Records[i] = &entities.ExchangeRateVO{}
		err := copier.Copy(res.Records[i], exchangeRate)
		if err != nil {
			logger.Errorf("copy [%v] error, %v", exchangeRate, err)
			utils.ReError(c, err)
			return
		}
		res.Records[i].BaseCurrency = exchangeRate.BaseCurrency.String()
		res.Records[i].QuoteCurrency = exchangeRate.QuoteCurrency.String()
		res.Records[i].Timestamp = exchangeRate.Timestamp.UnixMilli()
	}

	utils.ReData(c, res)
}

// @Param			request			body		entities.GetExchangeRateForm	true	"body"
// @Param			X-Token			header		string							true	"User token"
// @Param			Accept-Language	header		string							false	"accept language"
// @Param			X-Extend		header		string							false	"Extend"
// @Param			X-Convert		header		string							false	"Convert"
// @Success		0				{object}	entities.ListExchangeRateVO		"data"
// @Router			/web/quote/exchangeRates/getRates [post]
// @Description	List real-time exchange rate data.
// @Tags			web/quote
func (qh *QuoteHandler) GetExchange(c *gin.Context) {
	form := &entities.GetExchangeRateForm{}

	err := c.ShouldBindJSON(form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(form); err != nil {
		utils.ReError(c, utils.NewBusinessError(c, common.CODE_REQUEST_BODY_INVALID_FORMAT, err.Error()))
		return
	}

	for _, baseCurrency := range form.BaseCurrencies {
		if !common.Currency(0).FromString(baseCurrency).IsValid() {
			utils.ReError(c, utils.NewBusinessError(c, common.CODE_NO_SUCH_CURRENCY))
			return
		}
	}

	for _, quoteCurrency := range form.QuoteCurrencies {
		if !common.Currency(0).FromString(quoteCurrency).IsValid() {
			utils.ReError(c, utils.NewBusinessError(c, common.CODE_NO_SUCH_CURRENCY))
			return
		}
	}

	exchangeRates, err := qh.quoteService.GetExchangeRates(c, form)
	if err != nil {
		utils.ReError(c, err)
		return
	}

	utils.ReData(c, exchangeRates)
}
