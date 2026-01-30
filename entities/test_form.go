package entities

import (
	"shared-modules/common"

	"github.com/shopspring/decimal"
)

type AddAssetsForm struct {
	Email      string `json:"email" validate:"required"`
	Forwarding bool   `json:"forwarding"`
}

type GenerateSignatureForm struct {
	Data             string `json:"data" validate:"required"`
	PrivateKeyBase64 string `json:"privateKeyBase64"`
	PrivateKey       string `json:"privateKey"`
}

type MerchantSimulateAuthorizeForm struct {
	CardID  uint64          `json:"cardId,string"`
	Amount  decimal.Decimal `json:"amount" validate:"required"`
	OrderNO string          `json:"orderNo"`
}

type MerchantSimulateClearingForm struct {
	CardID  uint64          `json:"cardId,string"`
	Amount  decimal.Decimal `json:"amount"`
	OrderNO string          `json:"orderNo"`
}

type MerchantSimulateReversalForm struct {
	Amount  decimal.Decimal `json:"amount"`
	OrderNO string          `json:"orderNo" validate:"required"`
}

type MerchantSimulateRefundForm struct {
	Amount    decimal.Decimal `json:"amount"`
	OrderNO   string          `json:"orderNo" validate:"required"`
	Unrelated *bool           `json:"unrelated"`
}
type UploadS3Form struct {
	Purpose  common.FilePurpose `form:"purpose" binding:"required,min=1"`
	UserID   uint64             `form:"userId" binding:"required,min=1"`
	Filename string             `form:"filename" binding:"required,min=1,max=255"`
}
