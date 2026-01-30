package entities

import "github.com/shopspring/decimal"

type TransferPreviewForm struct {
	FromCardID uint64           `json:"fromCardId,string" validate:"required"`
	ToCardID   uint64           `json:"toCardId,string"`
	ToUserID   uint64           `json:"toUserId,string"`
	ToEmail    string           `json:"toEmail" validate:"omitempty,email"`
	ToAddress  string           `json:"toAddress"`
	Mainnet    string           `json:"mainnet"`
	Protocol   string           `json:"protocol"`
	FromAmount *decimal.Decimal `json:"fromAmount,string"`
	ToAmount   *decimal.Decimal `json:"toAmount,string"`
	Remark     *string          `json:"remark,omitempty"`
}

type TransferConfirmForm struct {
	ID       string `json:"id"`
	RawID    string `json:"rawId"`
	Type     string `json:"type"`
	Response struct {
		AuthenticatorData string `json:"authenticatorData"`
		ClientDataJSON    string `json:"clientDataJSON"`
		Signature         string `json:"signature"`
		UserHandle        string `json:"userHandle"`
	} `json:"response"`
	DeviceID    string `json:"deviceId"`
	TransferKey string `json:"transferKey" validate:"required"`
	PinCode     string `json:"pinCode"`
}
