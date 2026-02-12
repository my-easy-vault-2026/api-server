package entities

import "github.com/shopspring/decimal"

type ExchangePreviewForm struct {
	FromWalletID uint64          `form:"fromWalletId,string" validate:"required"`
	ToCurrency   string          `form:"toCurrency" validate:"required"`
	FromAmount   decimal.Decimal `form:"fromAmount"`
}

type ExchangeConfirmForm struct {
	Key string `form:"key" validate:"required"`
}

type AutoExchangeForm struct {
}
