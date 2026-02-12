package services

import (
	"context"

	"github.com/my-easy-vault-2026/shared-modules/common"

	systemDao "github.com/my-easy-vault-2026/api-server/dao/system"
	"github.com/my-easy-vault-2026/api-server/lib"
)

type QuoteService struct {
	parameterDao *systemDao.ParameterDao
	rateGetters  RateGetters
	logger       lib.Logger
	beBuilder    *lib.BEBuilder
}

func NewQuoteService(parameterDao *systemDao.ParameterDao, rateGetters RateGetters, logger lib.Logger, beBuilder *lib.BEBuilder) *QuoteService {
	return &QuoteService{
		parameterDao: parameterDao,
		rateGetters:  rateGetters,
		logger:       logger,
		beBuilder:    beBuilder,
	}
}

func (qs *QuoteService) GetExchangeRates(ctx context.Context, purpose common.RatePurpose, quote common.Currency, base common.Currency) (*common.ExchangeRate, error) {

	rate, err := qs.rateGetters.Get(purpose).GetExchangeRate(ctx, quote, base)
	if err != nil {
		qs.logger.Warn("get exchange rate failed,", err)
		return nil, err
	}

	return rate, nil
}
