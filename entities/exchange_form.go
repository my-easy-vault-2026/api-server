package entities

import "github.com/shopspring/decimal"

type ExchangePreviewForm struct {
	FromWalletID uint64          `json:"fromWalletId,string" validate:"required"`
	ToCurrency   string          `json:"toCurrency" validate:"required"`
	FromAmount   decimal.Decimal `json:"fromAmount"`
}

type ExchangeConfirmForm struct {
	Key string `json:"key" validate:"required"`
}

type AutoExchangeForm struct {
}
