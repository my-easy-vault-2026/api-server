package services

import (
	systemDao "api-server/dao/system"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
)

type QuoteService struct {
	parameterDao *systemDao.ParameterDao
	rateGetters  RateGetters
	logger       logger.Logger
}

func NewQuoteService(parameterDao *systemDao.ParameterDao, rateGetters RateGetters, logger logger.Logger) *QuoteService {
	return &QuoteService{
		parameterDao: parameterDao,
		rateGetters:  rateGetters,
		logger:       logger,
	}
}

// ListExchangeRate retrieves the real-time exchange rate for a given currency pair.
func (qs *QuoteService) ListExchangeRate(ctx context.Context, form *entities.ListExchangeRateForm) ([]*utils.ExchangeRate, error) {

	quoteCurrencies := make([]common.Currency, len(form.QuoteCurrencies))
	for i := range form.QuoteCurrencies {
		quoteCurrencies[i] = common.Currency(0).FromString(form.QuoteCurrencies[i])
	}

	rates, err := utils.ListExchangeRate(ctx, common.Currency(0).FromString(form.BaseCurrency), quoteCurrencies)
	if err != nil {
		logger.Warn("get exchange rate failed,", err)
		return nil, err
	}

	return rates, nil
}

// ListExchangeRate retrieves the real-time exchange rate for a given currency pair.
func (qs *QuoteService) GetExchangeRates(ctx context.Context, form *entities.GetExchangeRateForm) (*entities.GetExchangeRateVO, error) {
	return qs.rateGetters[form.Purpose].GetExchangeRates(ctx, form)
}
