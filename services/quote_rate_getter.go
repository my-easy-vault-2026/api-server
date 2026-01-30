package services

import (
	systemDao "api-server/dao/system"
	"api-server/lib"
	"context"
	"shared-modules/common"
	"shared-modules/entities"
	"shared-modules/logger"
	"shared-modules/utils"
)

type RateGetters map[common.RatePurpose]IRateGetter
type IRateGetter interface {
	GetExchangeRates(ctx context.Context, form *entities.GetExchangeRateForm) (*entities.GetExchangeRateVO, error)
}

type defaultRateGetter struct {
	parameterDao *systemDao.ParameterDao
	logger       lib.Logger
}

func NewRateGetters(parameterDao *systemDao.ParameterDao, logger lib.Logger) RateGetters {
	rg := RateGetters{
		0: &defaultRateGetter{
			parameterDao: parameterDao,
			logger:       logger,
		},
	}
	return rg
}

func (rg *defaultRateGetter) GetExchangeRates(ctx context.Context, form *entities.GetExchangeRateForm) (*entities.GetExchangeRateVO, error) {

	baseCurrencies := make([]common.Currency, len(form.BaseCurrencies))
	for i := range form.BaseCurrencies {
		baseCurrencies[i] = common.Currency(0).FromString(form.BaseCurrencies[i])
	}

	quoteCurrencies := make([]common.Currency, len(form.QuoteCurrencies))
	for i := range form.QuoteCurrencies {
		quoteCurrencies[i] = common.Currency(0).FromString(form.QuoteCurrencies[i])
	}

	res := &entities.GetExchangeRateVO{}
	resData := make(map[string]*entities.ExchangeRateVO)
	// 遍歷所有幣種，計算彼此之間的匯率
	for _, base := range baseCurrencies {
		for _, quote := range quoteCurrencies {
			rate, err := utils.GetUsdBaseExchangeRate(ctx, base, quote)
			if err != nil {
				logger.Warn("get exchange rate failed,", err)
				continue
			}
			key := rate.BaseCurrency.String() + "_" + rate.QuoteCurrency.String()

			if rate.Rate.IsZero() {
				logger.Errorf("exchange rate is zero: %v %v %v", rate.BaseCurrency, rate.QuoteCurrency, rate.Rate)
				return nil, utils.NewBusinessError(ctx, common.CODE_QUOTE_NO_SUCH_RATE)
			}

			resData[key] = &entities.ExchangeRateVO{
				BaseCurrency:  rate.QuoteCurrency.String(),
				QuoteCurrency: rate.BaseCurrency.String(),
				Rate:          rate.Rate,
				Timestamp:     rate.Timestamp.UnixMilli(),
			}
		}
	}

	res.Records = resData

	return res, nil
}
