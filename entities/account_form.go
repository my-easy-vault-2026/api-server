package entities

import (
	"shared-modules/common"

	"shared-modules/utils"

	"github.com/shopspring/decimal"
)

type ListAssetsForm struct {
	ID           uint64             `json:"id,omitempty,string"`
	UserID       uint64             `json:"userId,omitempty,string"`
	CategoryID   uint64             `json:"categoryId,omitempty,string"`
	CategoryIDIn []uint64           `json:"categoryIdIn,omitempty,string"`
	IDIn         []uint64           `json:"idIn,omitempty,string"`
	TypeIn       []common.AssetType `json:"typeIn,omitempty"`
}

type ListAssetsByTypeForm struct {
	Type common.AssetType `json:"type"`
}

type AdminPageBillForm struct {
	AssetIDIn     []uint64        `json:"assetIdIn,omitempty,string"`
	AssetID       uint64          `json:"assetId,omitempty,string"`
	UserID        uint64          `json:"userId,omitempty,string"`
	CategoryID    uint64          `json:"categoryId,omitempty,string"`
	Currency      common.Currency `json:"currency,omitempty"`
	MerchantID    uint64          `json:"merchantId,omitempty,string"`
	CreatedAtFrom int64           `json:"createdAtFrom,string" validate:"required"`
	CreatedAtTo   int64           `json:"createdAtTo,string" validate:"required"`
	utils.Page
}

type PageBillForm struct {
	Email         string          `json:"email,omitempty"`
	AssetIDIn     []uint64        `json:"assetIdIn,omitempty,string"`
	AssetID       uint64          `json:"assetId,omitempty,string"`
	UserID        uint64          `json:"userId,omitempty,string"`
	CategoryID    uint64          `json:"categoryId,omitempty,string"`
	Currency      common.Currency `json:"currency,omitempty"`
	CreatedAtFrom int64           `json:"createdAtFrom,string" validate:"required"`
	CreatedAtTo   int64           `json:"createdAtTo,string" validate:"required"`
	utils.Page
}

type ListCategoriesForm struct {
	ID   uint64           `json:"id,string"`
	IDIn []uint64         `json:"idIn,string"`
	Type common.AssetType `json:"type"`
}

type GetCategoryForm struct {
	ID uint64 `json:"id,string"`
}

type GetAccountListForm struct {
}

type ManualTransferForm struct {
	AssetID        uint64                 `json:"assetId,string" validate:"required"`
	BusinessNO     string                 `json:"businessNo"`
	AgainstAssetID uint64                 `json:"againstAssetId,string"`
	Amount         decimal.Decimal        `json:"amount" validate:"required"`
	Memo           string                 `json:"memo"`
	By             string                 `json:"by" validate:"required"`
	Type           common.TransactionType `json:"type" validate:"required"`
}

type GetEquivalentValueForm struct {
	Currency string             `json:"currency" validate:"required"`
	Type     common.AssetType   `json:"type"`
	Purpose  common.RatePurpose `json:"purpose"`
}
