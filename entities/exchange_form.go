package entities

import "github.com/shopspring/decimal"

type ExchangePreviewForm struct {
	FromCardID uint64           `json:"fromCardId,string" validate:"required"`
	ToCardID   uint64           `json:"toCardId,string"`
	ToCategory string           `json:"toCategory"`
	FromAmount *decimal.Decimal `json:"fromAmount"`
	ToAmount   *decimal.Decimal `json:"toAmount"`
}

type ExchangeConfirmForm struct {
	ExchangeKey string `json:"exchangeKey" validate:"required"`
	PinCode     string `json:"pinCode"`
}

type AutoExchangeForm struct {
}
