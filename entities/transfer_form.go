package entities

import "github.com/shopspring/decimal"

type TransferPreviewForm struct {
	FromWalletID uint64          `form:"fromWalletId,string" validate:"required"`
	ToUserID     uint64          `form:"toUserId,string"`
	ToEmail      string          `form:"toEmail" validate:"omitempty,email"`
	FromAmount   decimal.Decimal `form:"fromAmount,string" validate:"required"`
}

type TransferConfirmForm struct {
	Key     string `form:"key" validate:"required"`
	PinCode string `form:"pinCode" validate:"required"`
}
