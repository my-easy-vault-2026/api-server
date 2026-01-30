package entities

import "shared-modules/common"

type AutoYieldInfoForm struct {
	Currency string `json:"currency" validate:"required"`
}

type AutoYieldHistoryForm struct {
	Currency string `json:"currency" validate:"required"`
	Period   int    `json:"period" validate:"required,gt=0,lte=1000"`
}

type AutoYieldEnableForm struct {
	Enable *bool  `json:"enable" validate:"required"`
	CardID uint64 `json:"cardId,string" validate:"required"`
}

type AutoYieldSnapshotForm struct {
	Now        int64             `json:"now"`
	Type       common.AssetType  `json:"type"`
	Currencies []common.Currency `json:"currencies"`
}

type AutoYieldCheckSnapshotForm struct {
	Now        int64             `json:"now"`
	Type       common.AssetType  `json:"type"`
	Currencies []common.Currency `json:"currencies"`
}

type AutoYieldDistributeForm struct {
	Now        int64             `json:"now"`
	Type       common.AssetType  `json:"type"`
	Currencies []common.Currency `json:"currencies"`
}

type FinancialGetProductForm struct {
	Code common.FinancialCode `json:"code" validate:"required"`
}
