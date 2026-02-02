package services

import (
	"context"
	"shared-modules/common"
)

type IRateGetter interface {
	GetExchangeRate(ctx context.Context, quote common.Currency, base common.Currency) (*common.ExchangeRate, error)
}

type RateGetters struct {
	rateGetters       map[common.RatePurpose]IRateGetter
	defaultRateGetter *DefaultRateGetter
}

func NewRateGetters(defaultRateGetter *DefaultRateGetter) RateGetters {
	rg := RateGetters{
		rateGetters: map[common.RatePurpose]IRateGetter{
			0: defaultRateGetter,
		},
	}
	rg.defaultRateGetter = defaultRateGetter
	return rg
}

func (rg *RateGetters) Get(purpose common.RatePurpose) IRateGetter {
	if rg.rateGetters[purpose] == nil {
		return rg.defaultRateGetter
	}
	return rg.rateGetters[purpose]
}
