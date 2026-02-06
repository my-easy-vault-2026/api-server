package entities

import "github.com/shopspring/decimal"

type TransferPreviewForm struct {
	FromWalletID uint64          `json:"fromWalletId,string" validate:"required"`
	ToUserID     uint64          `json:"toUserId,string"`
	ToEmail      string          `json:"toEmail" validate:"omitempty,email"`
	FromAmount   decimal.Decimal `json:"fromAmount,string" validate:"required"`
}

type TransferConfirmForm struct {
	Key     string `json:"transferKey" validate:"required"`
	PinCode string `json:"pinCode" validate:"required"`
}
