package entities

import "github.com/shopspring/decimal"

type ExchangePreviewForm struct {
	FromWalletID uint64          `json:"fromWalletId,string" validate:"required"`
	ToCurrency   string          `json:"toCurrency" validate:"required"`
	FromAmount   decimal.Decimal `json:"fromAmount"`
}

type ExchangeConfirmForm struct {
	ExchangeKey string `json:"exchangeKey" validate:"required"`
	PinCode     string `json:"pinCode"`
}

type AutoExchangeForm struct {
}
