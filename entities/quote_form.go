package entities

import "shared-modules/common"

type ListExchangeRateForm struct {
	BaseCurrency    string             `json:"baseCurrency" validate:"required"`
	QuoteCurrencies []string           `json:"quoteCurrencies" validate:"required"`
	Purpose         common.RatePurpose `json:"purpose"` //
}

type GetExchangeRateForm struct {
	BaseCurrencies  []string           `json:"baseCurrencies" validate:"required"`
	QuoteCurrencies []string           `json:"quoteCurrencies" validate:"required"`
	Purpose         common.RatePurpose `json:"purpose"` //
}
