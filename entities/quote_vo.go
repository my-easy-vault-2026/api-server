package entities

import "github.com/shopspring/decimal"

type ListExchangeRateVO struct {
	Records        []*ExchangeRateVO            `json:"records"`
	EffectiveRates map[string][]*ExchangeRateVO `json:"effectiveRates,omitempty"`
	Fees           map[string]decimal.Decimal   `json:"fees,omitempty"`
}

type GetExchangeRateVO struct {
	Records map[string]*ExchangeRateVO `json:"records"`
}
