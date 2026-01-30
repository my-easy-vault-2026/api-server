package entities

import (
	"shared-modules/common"
)

type ListWalletsForm struct {
	ID         uint64
	CategoryID uint64
}

type ListWalletAddressesForm struct {
	ID               uint64 `json:"id,string"`
	CategoryID       uint64 `json:"categoryId,string"`
	Mainnet          string `json:"mainnet"`
	Protocol         string `json:"protocol"`
	Currency         string `json:"currency"`
	CreateIfNotExist bool   `json:"createIfNotExist"`
}

type GetWalletAddressForm struct {
	ID               uint64 `json:"id,string"`
	CardID           uint64 `json:"cardId,string" validate:"required"`
	CategoryID       uint64 `json:"categoryId,string" validate:"required"`
	Mainnet          string `json:"mainnet" validate:"required"`
	Protocol         string `json:"protocol"`
	Currency         string `json:"currency" validate:"required"`
	CreateIfNotExist bool   `json:"createIfNotExist" validate:"required"`
}

type GetMainnetAddressForm struct {
	Mainnet string `json:"mainnet" validate:"required"`
}

type CreateWalletForm struct {
	Currency   string `json:"currency"`
	CategoryID uint64 `json:"categoryId,string"`
}

type CreatePointForm struct {
	CategoryID uint64 `json:"categoryId,string"`
}

type CreateWalletAddressForm struct {
	ID         uint64                     `json:"id,string"`
	CardID     uint64                     `json:"cardId,string"`
	UserID     uint64                     `json:"userId,string"`
	Address    string                     `json:"address"`
	CategoryID uint64                     `json:"categoryId,string"`
	Mainnet    common.Mainnet             `json:"mainnet"`
	Protocol   common.Protocol            `json:"protocol"`
	Currency   common.Currency            `json:"currency"`
	DeviceUUID string                     `json:"deviceUuid"`
	Flag       string                     `json:"flag"`
	Decimals   int                        `json:"decimals"`
	Type       common.CoinTokenType       `json:"type"`
	CoinsdoID  uint64                     `json:"coinsdoId,string"`
	Status     common.WalletAddressStatus `json:"status"`
}

type MerchantApplyGetAddressForm struct {
	CardID     uint64 `json:"cardId,string"`
	Currency   string `json:"currency"`
	CategoryID uint64 `json:"categoryId,string"`
	Mainnet    string `json:"mainnet" validate:"required"`
	Protocol   string `json:"protocol"`
}

type MerchantApplyGetAddressFromAdminForm struct {
	MerchantID uint64 `json:"merchantId,string" validate:"required"`
	MerchantApplyGetAddressForm
}

type CreateAddressFromPoolForm struct {
	CardID     uint64 `json:"cardId,string"`
	UserID     uint64 `json:"userId,string"`
	CategoryID uint64 `json:"categoryId,string"`
	Mainnet    string `json:"mainnet"`
	Protocol   string `json:"protocol"`
}

type CheckWalletAndPreviewForm struct {
	Currency       string            `json:"currency" validate:"required"`
	Mainnet        string            `json:"mainnet" validate:"required"`
	CardCategoryID uint64            `json:"cardCategoryId,string" validate:"required"`
	PromotionCode  string            `json:"promotionCode"`
	AddressLine1   string            `json:"addressLine1"`
	NationCode     common.NationCode `json:"nationCode"`
	PostalCode     string            `json:"postalCode"`
	City           string            `json:"city"`
	State          string            `json:"state"`
}
